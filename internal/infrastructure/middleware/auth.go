package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

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

// JWTAuthMiddleware quản lý JWT authentication
type JWTAuthMiddleware struct {
	secretKey       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	enabled         bool
}

// NewJWTAuth tạo JWTAuthMiddleware mới
func NewJWTAuth(cfg config.JWTConfig) *JWTAuthMiddleware {
	accessTTL, _ := time.ParseDuration(cfg.AccessTokenTTL)
	if accessTTL == 0 {
		accessTTL = 15 * time.Minute
	}

	refreshTTL, _ := time.ParseDuration(cfg.RefreshTokenTTL)
	if refreshTTL == 0 {
		refreshTTL = 168 * time.Hour // 7 days
	}

	return &JWTAuthMiddleware{
		secretKey:       []byte(cfg.SecretKey),
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
		enabled:         cfg.Enabled,
	}
}

// GenerateAccessToken tạo access token mới
func (j *JWTAuthMiddleware) GenerateAccessToken(userID, role, email string) (string, error) {
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "restaurant-api",
		},
		UserID: userID,
		Role:   role,
		Email:  email,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// GenerateRefreshToken tạo refresh token mới
func (j *JWTAuthMiddleware) GenerateRefreshToken(userID string) (string, error) {
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.refreshTokenTTL)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Subject:   userID,
		Issuer:    "restaurant-api",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
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
