// Package handler chứa HTTP Handlers
package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"restaurant_project/internal/application/usecase"
	"restaurant_project/internal/infrastructure/middleware"
	"restaurant_project/internal/presentation/http/dto"
)

// AuthHandler xử lý các HTTP request liên quan đến Authentication
type AuthHandler struct {
	useCase *usecase.AuthUseCase
}

// NewAuthHandler tạo mới AuthHandler
func NewAuthHandler(uc *usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{
		useCase: uc,
	}
}

// Register xử lý POST /api/auth/register - Đăng ký tài khoản mới
// @Summary Đăng ký tài khoản
// @Description Đăng ký tài khoản mới (Customer)
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Thông tin đăng ký"
// @Success 201 {object} dto.APIResponse{data=dto.AuthResponse}
// @Failure 400 {object} dto.APIResponse
// @Router /api/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest,
			dto.NewErrorResponse("Dữ liệu không hợp lệ", err))
		return
	}

	input := usecase.RegisterInput{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	}

	result, err := h.useCase.Register(c.Request.Context(), input)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err == usecase.ErrUsernameExists || err == usecase.ErrEmailExists {
			statusCode = http.StatusConflict
		}
		c.JSON(statusCode,
			dto.NewErrorResponse("Đăng ký thất bại", err))
		return
	}

	response := dto.AuthResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    result.ExpiresIn,
		User:         dto.ToUserResponse(result.User),
	}

	c.JSON(http.StatusCreated,
		dto.NewSuccessResponse("Đăng ký thành công", response))
}

// Login xử lý POST /api/auth/login - Đăng nhập
// @Summary Đăng nhập
// @Description Đăng nhập và nhận tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Thông tin đăng nhập"
// @Success 200 {object} dto.APIResponse{data=dto.AuthResponse}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest,
			dto.NewErrorResponse("Dữ liệu không hợp lệ", err))
		return
	}

	input := usecase.LoginInput{
		Username: req.Username,
		Password: req.Password,
	}

	result, err := h.useCase.Login(c.Request.Context(), input)
	if err != nil {
		statusCode := http.StatusBadRequest

		// Check if account is locked
		var lockedErr *usecase.AccountLockedError
		if errors.As(err, &lockedErr) {
			c.JSON(http.StatusLocked, gin.H{
				"success":           false,
				"message":           "Tài khoản tạm thời bị khóa",
				"error":             err.Error(),
				"remaining_seconds": lockedErr.RemainingSeconds,
			})
			return
		}

		if err == usecase.ErrInvalidCredentials {
			statusCode = http.StatusUnauthorized
		}
		if err == usecase.ErrUserInactive {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode,
			dto.NewErrorResponse("Đăng nhập thất bại", err))
		return
	}

	response := dto.AuthResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    result.ExpiresIn,
		User:         dto.ToUserResponse(result.User),
	}

	c.JSON(http.StatusOK,
		dto.NewSuccessResponse("Đăng nhập thành công", response))
}

// RefreshToken xử lý POST /api/auth/refresh - Làm mới access token
// @Summary Làm mới access token
// @Description Làm mới access token từ refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} dto.APIResponse{data=dto.RefreshTokenResponse}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Router /api/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest,
			dto.NewErrorResponse("Dữ liệu không hợp lệ", err))
		return
	}

	result, err := h.useCase.RefreshToken(c.Request.Context(), req.RefreshToken, req.AccessToken)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err == usecase.ErrInvalidToken {
			statusCode = http.StatusUnauthorized
		}
		if err == usecase.ErrUserInactive {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode,
			dto.NewErrorResponse("Làm mới token thất bại", err))
		return
	}

	response := dto.RefreshTokenResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    result.ExpiresIn,
	}

	c.JSON(http.StatusOK,
		dto.NewSuccessResponse("Làm mới token thành công", response))
}

// Logout xử lý POST /api/auth/logout - Đăng xuất (revoke token)
// @Summary Đăng xuất
// @Description Đăng xuất và revoke access token hiện tại
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Router /api/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Lấy token từ header
	tokenString, ok := middleware.ExtractTokenFromHeader(c)
	if !ok {
		c.JSON(http.StatusUnauthorized,
			dto.NewErrorResponse("Token không tìm thấy", nil))
		return
	}

	// Validate token và lấy claims
	jwtAuth := h.useCase.GetJWTAuth()
	claims, err := jwtAuth.ValidateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized,
			dto.NewErrorResponse("Token không hợp lệ", err))
		return
	}

	// Blacklist token
	if err := jwtAuth.BlacklistToken(c, claims); err != nil {
		c.JSON(http.StatusInternalServerError,
			dto.NewErrorResponse("Không thể đăng xuất", err))
		return
	}

	c.JSON(http.StatusOK,
		dto.NewSuccessResponse("Đăng xuất thành công", nil))
}

// LogoutAllDevices xử lý POST /api/auth/logout-all - Đăng xuất khỏi tất cả thiết bị
// @Summary Đăng xuất tất cả thiết bị
// @Description Đăng xuất và revoke tất cả tokens của user (tất cả devices)
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.APIResponse{data=dto.LogoutAllResponse}
// @Failure 401 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/auth/logout-all [post]
func (h *AuthHandler) LogoutAllDevices(c *gin.Context) {
	// Lấy userID từ context (đã được set bởi JWT middleware)
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized,
			dto.NewErrorResponse("Không tìm thấy thông tin user", nil))
		return
	}

	// Lấy số session trước khi revoke (để thông báo cho user)
	sessionCount, _ := h.useCase.GetActiveSessionCount(c.Request.Context(), userID)

	// Revoke tất cả tokens
	if err := h.useCase.LogoutAllDevices(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError,
			dto.NewErrorResponse("Không thể đăng xuất tất cả thiết bị", err))
		return
	}

	response := dto.LogoutAllResponse{
		RevokedSessions: sessionCount,
		Message:         "Đã đăng xuất khỏi tất cả thiết bị",
	}

	c.JSON(http.StatusOK,
		dto.NewSuccessResponse("Đăng xuất tất cả thiết bị thành công", response))
}

// GetActiveSessions xử lý GET /api/auth/sessions - Lấy số lượng session active
// @Summary Lấy số session active
// @Description Lấy số lượng thiết bị đang đăng nhập
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.APIResponse{data=dto.ActiveSessionsResponse}
// @Failure 401 {object} dto.APIResponse
// @Router /api/auth/sessions [get]
func (h *AuthHandler) GetActiveSessions(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized,
			dto.NewErrorResponse("Không tìm thấy thông tin user", nil))
		return
	}

	count, err := h.useCase.GetActiveSessionCount(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			dto.NewErrorResponse("Không thể lấy thông tin session", err))
		return
	}

	response := dto.ActiveSessionsResponse{
		ActiveSessions: count,
	}

	c.JSON(http.StatusOK,
		dto.NewSuccessResponse("Lấy thông tin session thành công", response))
}

// VerifyEmail xử lý POST /api/auth/verify-email - Xác thực email
// @Summary Xác thực email
// @Description Xác thực email của user với verification token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.VerifyEmailRequest true "Verification token"
// @Success 200 {object} dto.APIResponse{data=dto.VerifyEmailResponse}
// @Failure 400 {object} dto.APIResponse
// @Router /api/auth/verify-email [post]
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req dto.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest,
			dto.NewErrorResponse("Dữ liệu không hợp lệ", err))
		return
	}

	err := h.useCase.VerifyEmail(c.Request.Context(), req.Token)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err == usecase.ErrInvalidVerificationToken {
			statusCode = http.StatusBadRequest
		}
		if err == usecase.ErrEmailAlreadyVerified {
			statusCode = http.StatusConflict
		}
		c.JSON(statusCode,
			dto.NewErrorResponse("Xác thực email thất bại", err))
		return
	}

	response := dto.VerifyEmailResponse{
		Message: "Email đã được xác thực thành công",
	}

	c.JSON(http.StatusOK,
		dto.NewSuccessResponse("Xác thực email thành công", response))
}

// ResendVerification xử lý POST /api/auth/resend-verification - Gửi lại email xác thực
// @Summary Gửi lại email xác thực
// @Description Gửi lại email xác thực cho user (cần đăng nhập)
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.APIResponse{data=dto.ResendVerificationResponse}
// @Failure 401 {object} dto.APIResponse
// @Failure 409 {object} dto.APIResponse
// @Failure 429 {object} dto.APIResponse{data=dto.ResendCooldownResponse}
// @Router /api/auth/resend-verification [post]
func (h *AuthHandler) ResendVerification(c *gin.Context) {
	// Lấy userID từ context (đã được set bởi JWT middleware)
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized,
			dto.NewErrorResponse("Không tìm thấy thông tin user", nil))
		return
	}

	err := h.useCase.ResendVerificationEmail(c.Request.Context(), userID)
	if err != nil {
		// Check if it's a cooldown error
		var cooldownErr *usecase.ResendCooldownError
		if errors.As(err, &cooldownErr) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success":           false,
				"message":           "Vui lòng đợi trước khi gửi lại email xác thực",
				"error":             err.Error(),
				"remaining_seconds": cooldownErr.RemainingSeconds,
			})
			return
		}

		statusCode := http.StatusBadRequest
		if err == usecase.ErrEmailAlreadyVerified {
			statusCode = http.StatusConflict
		}
		c.JSON(statusCode,
			dto.NewErrorResponse("Gửi lại email xác thực thất bại", err))
		return
	}

	response := dto.ResendVerificationResponse{
		Message: "Email xác thực đã được gửi lại",
	}

	c.JSON(http.StatusOK,
		dto.NewSuccessResponse("Gửi lại email xác thực thành công", response))
}

// ============================================================
// RouteRegistrar Interface Implementation
// ============================================================

// BasePath trả về base path cho Auth module
func (h *AuthHandler) BasePath() string {
	return "/auth"
}

// RegisterRoutes đăng ký tất cả routes của Auth module
// Note: register, login, refresh, verify-email là PUBLIC - không cần JWT
// logout, resend-verification cần JWT authentication
func (h *AuthHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/register", h.Register)
	rg.POST("/login", h.Login)
	rg.POST("/refresh", h.RefreshToken)
	rg.POST("/verify-email", h.VerifyEmail)
}

// RegisterProtectedRoutes đăng ký routes cần JWT authentication
func (h *AuthHandler) RegisterProtectedRoutes(rg *gin.RouterGroup) {
	rg.POST("/logout", h.Logout)
	rg.POST("/logout-all", h.LogoutAllDevices)
	rg.GET("/sessions", h.GetActiveSessions)
	rg.POST("/resend-verification", h.ResendVerification)
}
