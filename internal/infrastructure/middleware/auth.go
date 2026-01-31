package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"go.uber.org/zap"

	"restaurant_project/internal/domain/service"
	"restaurant_project/internal/infrastructure/config"
	"restaurant_project/pkg/logger"
)

// Context keys cho JWT claims
const (
	ContextKeyUserID   = "user_id"
	ContextKeyUserRole = "user_role"
	ContextKeyClaims   = "jwt_claims"
)

// UserClaims chứa thông tin user trong JWT
type UserClaims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	Email  string `json:"email,omitempty"`
}

// TokenInfo chứa thông tin token được tạo (dùng để track)
type TokenInfo struct {
	Token string        // JWT token string
	JTI   string        // JWT ID (unique identifier)
	TTL   time.Duration // Time to live
}

// JWTAuthMiddleware quản lý JWT authentication
type JWTAuthMiddleware struct {
	secretKey        []byte
	accessTokenTTL   time.Duration
	refreshTokenTTL  time.Duration
	enabled          bool
	blacklistService service.TokenBlacklistService
}

// NewJWTAuth tạo JWTAuthMiddleware mới
func NewJWTAuth(cfg config.JWTConfig, blacklistService service.TokenBlacklistService) *JWTAuthMiddleware {
	accessTTL, _ := time.ParseDuration(cfg.AccessTokenTTL)
	if accessTTL == 0 {
		accessTTL = 15 * time.Minute
	}

	refreshTTL, _ := time.ParseDuration(cfg.RefreshTokenTTL)
	if refreshTTL == 0 {
		refreshTTL = 2 * time.Hour // 2 hours (rotation enabled)
	}

	return &JWTAuthMiddleware{
		secretKey:        []byte(cfg.SecretKey),
		accessTokenTTL:   accessTTL,
		refreshTokenTTL:  refreshTTL,
		enabled:          cfg.Enabled,
		blacklistService: blacklistService,
	}
}

// GenerateAccessToken tạo access token mới với JTI để support blacklist
func (j *JWTAuthMiddleware) GenerateAccessToken(userID, role, email string) (string, error) {
	tokenInfo, err := j.GenerateAccessTokenWithInfo(userID, role, email)
	if err != nil {
		return "", err
	}
	return tokenInfo.Token, nil
}

// GenerateAccessTokenWithInfo tạo access token và trả về đầy đủ thông tin (bao gồm JTI)
func (j *JWTAuthMiddleware) GenerateAccessTokenWithInfo(userID, role, email string) (*TokenInfo, error) {
	jti := uuid.New().String()

	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti, // JTI cho token blacklist
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "restaurant-api",
		},
		UserID: userID,
		Role:   role,
		Email:  email,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secretKey)
	if err != nil {
		return nil, err
	}

	return &TokenInfo{
		Token: tokenString,
		JTI:   jti,
		TTL:   j.accessTokenTTL,
	}, nil
}

// GenerateRefreshToken tạo refresh token mới với JTI
func (j *JWTAuthMiddleware) GenerateRefreshToken(userID string) (string, error) {
	tokenInfo, err := j.GenerateRefreshTokenWithInfo(userID)
	if err != nil {
		return "", err
	}
	return tokenInfo.Token, nil
}

// GenerateRefreshTokenWithInfo tạo refresh token và trả về đầy đủ thông tin (bao gồm JTI)
func (j *JWTAuthMiddleware) GenerateRefreshTokenWithInfo(userID string) (*TokenInfo, error) {
	jti := uuid.New().String()

	claims := jwt.RegisteredClaims{
		ID:        jti, // JTI cho token blacklist
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.refreshTokenTTL)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Subject:   userID,
		Issuer:    "restaurant-api",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secretKey)
	if err != nil {
		return nil, err
	}

	return &TokenInfo{
		Token: tokenString,
		JTI:   jti,
		TTL:   j.refreshTokenTTL,
	}, nil
}

// ValidateToken xác thực và parse JWT token
func (j *JWTAuthMiddleware) ValidateToken(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// ParseTokenUnverifiedExpiry parse JWT token mà không check expiry
// Dùng khi cần lấy JTI từ access token đã hết hạn (ví dụ: khi refresh)
func (j *JWTAuthMiddleware) ParseTokenIgnoreExpiry(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secretKey, nil
	}, jwt.WithoutClaimsValidation())

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// Middleware trả về gin middleware để xác thực JWT
// Sử dụng: router.Use(jwtAuth.Middleware())
func (j *JWTAuthMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !j.enabled {
			c.Next()
			return
		}

		// Lấy token từ Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":      "Authorization header required",
				"code":       "AUTH_HEADER_MISSING",
				"request_id": logger.GetRequestID(c),
			})
			return
		}

		// Kiểm tra format "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":      "Invalid authorization format. Use: Bearer <token>",
				"code":       "AUTH_FORMAT_INVALID",
				"request_id": logger.GetRequestID(c),
			})
			return
		}

		// Validate token
		claims, err := j.ValidateToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":      "Invalid or expired token",
				"code":       "TOKEN_INVALID",
				"request_id": logger.GetRequestID(c),
			})
			return
		}

		// Check token blacklist (nếu token đã logout)
		if j.blacklistService != nil && claims.ID != "" {
			isBlacklisted, err := j.blacklistService.IsBlacklisted(c.Request.Context(), claims.ID)
			if err != nil {
				// Fail-closed: từ chối request nếu không thể kiểm tra blacklist
				logger.Error("Failed to check token blacklist", zap.Error(err))
				c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
					"error":      "Unable to verify token status",
					"code":       "BLACKLIST_CHECK_FAILED",
					"request_id": logger.GetRequestID(c),
				})
				return
			}
			if isBlacklisted {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error":      "Token has been revoked",
					"code":       "TOKEN_REVOKED",
					"request_id": logger.GetRequestID(c),
				})
				return
			}
		}

		// Lưu claims vào context
		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyUserRole, claims.Role)
		c.Set(ContextKeyClaims, claims)

		c.Next()
	}
}

// OptionalMiddleware cho phép request không có token đi qua
// Nhưng nếu có token thì vẫn parse và lưu vào context
func (j *JWTAuthMiddleware) OptionalMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !j.enabled {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.Next()
			return
		}

		claims, err := j.ValidateToken(parts[1])
		if err != nil {
			c.Next()
			return
		}

		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyUserRole, claims.Role)
		c.Set(ContextKeyClaims, claims)

		c.Next()
	}
}

// GetUserID lấy user ID từ context
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get(ContextKeyUserID)
	if !exists {
		return "", false
	}
	return userID.(string), true
}

// GetUserRole lấy role từ context
func GetUserRole(c *gin.Context) (string, bool) {
	role, exists := c.Get(ContextKeyUserRole)
	if !exists {
		return "", false
	}
	return role.(string), true
}

// GetClaims lấy toàn bộ claims từ context
func GetClaims(c *gin.Context) (*UserClaims, bool) {
	claims, exists := c.Get(ContextKeyClaims)
	if !exists {
		return nil, false
	}
	return claims.(*UserClaims), true
}

// ExtractTokenFromHeader lấy token string từ Authorization header
func ExtractTokenFromHeader(c *gin.Context) (string, bool) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", false
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", false
	}

	return parts[1], true
}

// GetTokenRemainingTime tính thời gian còn lại của token (cho blacklist TTL)
func (j *JWTAuthMiddleware) GetTokenRemainingTime(claims *UserClaims) time.Duration {
	if claims.ExpiresAt == nil {
		return 0
	}

	remaining := time.Until(claims.ExpiresAt.Time)
	if remaining < 0 {
		return 0
	}

	return remaining
}

// BlacklistToken thêm token vào blacklist và xóa khỏi tracking
func (j *JWTAuthMiddleware) BlacklistToken(c *gin.Context, claims *UserClaims) error {
	if j.blacklistService == nil {
		return nil
	}

	if claims.ID == "" {
		return errors.New("token has no JTI")
	}

	ttl := j.GetTokenRemainingTime(claims)
	if ttl <= 0 {
		// Token đã hết hạn, không cần blacklist
		return nil
	}

	ctx := c.Request.Context()

	if err := j.blacklistService.Blacklist(ctx, claims.ID, ttl); err != nil {
		return err
	}

	// Xóa token khỏi user_tokens SET và token_user mapping
	if claims.UserID != "" {
		_ = j.blacklistService.UntrackUserToken(ctx, claims.UserID, claims.ID)
	}

	return nil
}

// GetBlacklistService trả về TokenBlacklistService (dùng cho AuthUseCase)
func (j *JWTAuthMiddleware) GetBlacklistService() service.TokenBlacklistService {
	return j.blacklistService
}
