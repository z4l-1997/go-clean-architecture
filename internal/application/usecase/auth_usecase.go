// Package usecase chứa Application Use Cases
package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"restaurant_project/internal/domain/entity"
	"restaurant_project/internal/domain/repository"
	"restaurant_project/internal/domain/service"
	"restaurant_project/internal/infrastructure/middleware"
	"restaurant_project/pkg/password"
)

// Auth use case errors
var (
	ErrInvalidCredentials        = errors.New("username hoặc password không đúng")
	ErrUserInactive              = errors.New("tài khoản đã bị vô hiệu hóa")
	ErrInvalidToken              = errors.New("token không hợp lệ hoặc đã hết hạn")
	ErrAccountLocked             = errors.New("tài khoản tạm thời bị khóa do đăng nhập sai nhiều lần")
	ErrInvalidVerificationToken  = errors.New("token xác thực không hợp lệ hoặc đã hết hạn")
	ErrEmailAlreadyVerified      = errors.New("email đã được xác thực")
	ErrResendCooldown            = errors.New("vui lòng đợi trước khi gửi lại email xác thực")
)

// RegisterInput là input để đăng ký tài khoản mới
type RegisterInput struct {
	Username string
	Email    string
	Password string
}

// LoginInput là input để đăng nhập
type LoginInput struct {
	Username string
	Password string
}

// AuthResult là kết quả sau khi đăng nhập/đăng ký thành công
type AuthResult struct {
	User         *entity.User
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64 // seconds
}

// AuthUseCase xử lý business logic liên quan đến Authentication
type AuthUseCase struct {
	userRepo                 repository.IUserRepository
	jwtAuth                  *middleware.JWTAuthMiddleware
	loginAttemptService      service.LoginAttemptService
	emailVerificationService service.EmailVerificationService
	emailService             service.EmailService
}

// NewAuthUseCase tạo mới AuthUseCase
func NewAuthUseCase(
	userRepo repository.IUserRepository,
	jwtAuth *middleware.JWTAuthMiddleware,
	loginAttemptService service.LoginAttemptService,
	emailVerificationService service.EmailVerificationService,
	emailService service.EmailService,
) *AuthUseCase {
	return &AuthUseCase{
		userRepo:                 userRepo,
		jwtAuth:                  jwtAuth,
		loginAttemptService:      loginAttemptService,
		emailVerificationService: emailVerificationService,
		emailService:             emailService,
	}
}

// Register đăng ký tài khoản mới (Public - chỉ tạo được Customer)
func (uc *AuthUseCase) Register(ctx context.Context, input RegisterInput) (*AuthResult, error) {
	// Check username exists
	exists, err := uc.userRepo.ExistsByUsername(ctx, input.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUsernameExists
	}

	// Check email exists
	exists, err = uc.userRepo.ExistsByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailExists
	}

	// Hash password
	passwordHash, err := password.Hash(input.Password)
	if err != nil {
		return nil, err
	}

	// Create user entity với role Customer
	user, err := entity.NewUser(
		uuid.New().String(),
		input.Username,
		input.Email,
		passwordHash,
		entity.RoleCustomer, // Public register chỉ tạo được Customer
	)
	if err != nil {
		return nil, err
	}

	// Save to repository
	if err := uc.userRepo.Save(ctx, user); err != nil {
		return nil, err
	}

	// Generate email verification token và gửi email
	if uc.emailVerificationService != nil && uc.emailService != nil {
		token, err := uc.emailVerificationService.GenerateToken(ctx, user.ID)
		if err == nil && token != "" {
			// Set cooldown trước khi gửi email
			_ = uc.emailVerificationService.SetResendCooldown(ctx, user.ID)
			// Gửi email (hoặc log ra console trong dev mode)
			_ = uc.emailService.SendVerificationEmail(ctx, user.Email, token)
		}
	}

	// Generate tokens
	return uc.generateAuthResult(ctx, user)
}

// Login đăng nhập và trả về tokens
func (uc *AuthUseCase) Login(ctx context.Context, input LoginInput) (*AuthResult, error) {
	// Check if account is locked
	if uc.loginAttemptService != nil {
		isLocked, err := uc.loginAttemptService.IsLocked(ctx, input.Username)
		if err == nil && isLocked {
			remainingTime, _ := uc.loginAttemptService.GetRemainingLockTime(ctx, input.Username)
			return nil, &AccountLockedError{
				RemainingSeconds: remainingTime,
			}
		}
	}

	// Find user by username
	user, err := uc.userRepo.FindByUsername(ctx, input.Username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		// Increment failed attempts even for non-existent users (để không leak thông tin)
		if uc.loginAttemptService != nil {
			_, _ = uc.loginAttemptService.IncrementAttempts(ctx, input.Username)
		}
		return nil, ErrInvalidCredentials
	}

	// Check if user is active
	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// Verify password
	if !password.Verify(input.Password, user.PasswordHash) {
		// Increment failed attempts
		if uc.loginAttemptService != nil {
			_, _ = uc.loginAttemptService.IncrementAttempts(ctx, input.Username)
		}
		return nil, ErrInvalidCredentials
	}

	// Login success - reset failed attempts
	if uc.loginAttemptService != nil {
		_ = uc.loginAttemptService.ResetAttempts(ctx, input.Username)
	}

	// Generate tokens
	return uc.generateAuthResult(ctx, user)
}

// AccountLockedError chứa thông tin về tài khoản bị khóa
type AccountLockedError struct {
	RemainingSeconds int64
}

func (e *AccountLockedError) Error() string {
	return "tài khoản tạm thời bị khóa do đăng nhập sai nhiều lần"
}

// RefreshToken làm mới access token từ refresh token
// oldAccessToken (optional): access token cũ để cleanup JTI khỏi Redis
func (uc *AuthUseCase) RefreshToken(ctx context.Context, refreshToken string, oldAccessToken string) (*AuthResult, error) {
	// Validate refresh token và lấy user ID
	claims, err := uc.jwtAuth.ValidateToken(refreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Check refresh token đã bị blacklist chưa (chống reuse attack)
	if claims.ID != "" {
		if blacklistService := uc.jwtAuth.GetBlacklistService(); blacklistService != nil {
			isBlacklisted, err := blacklistService.IsBlacklisted(ctx, claims.ID)
			if err == nil && isBlacklisted {
				return nil, ErrInvalidToken
			}
		}
	}

	// Lấy user từ claims.Subject (refresh token chỉ có Subject là userID)
	userID := claims.Subject
	if userID == "" {
		// Thử lấy từ UserID field nếu Subject rỗng
		userID = claims.UserID
	}

	if userID == "" {
		return nil, ErrInvalidToken
	}

	// Revoke refresh token cũ: blacklist + untrack
	if claims.ID != "" {
		if blacklistService := uc.jwtAuth.GetBlacklistService(); blacklistService != nil {
			remainingTTL := uc.jwtAuth.GetTokenRemainingTime(claims)
			if remainingTTL > 0 {
				_ = blacklistService.Blacklist(ctx, claims.ID, remainingTTL)
			}
			_ = blacklistService.UntrackUserToken(ctx, userID, claims.ID)
		}
	}

	// Blacklist + untrack access token cũ để không thể sử dụng lại
	if oldAccessToken != "" {
		if oldClaims, err := uc.jwtAuth.ParseTokenIgnoreExpiry(oldAccessToken); err == nil && oldClaims.ID != "" {
			if blacklistService := uc.jwtAuth.GetBlacklistService(); blacklistService != nil {
				remainingTTL := uc.jwtAuth.GetTokenRemainingTime(oldClaims)
				if remainingTTL > 0 {
					_ = blacklistService.Blacklist(ctx, oldClaims.ID, remainingTTL)
				}
				_ = blacklistService.UntrackUserToken(ctx, userID, oldClaims.ID)
			}
		}
	}

	// Find user
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Check if user is still active
	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// Generate new tokens (cả access + refresh mới)
	return uc.generateAuthResult(ctx, user)
}

// generateAuthResult tạo AuthResult với tokens và track tokens cho user
func (uc *AuthUseCase) generateAuthResult(ctx context.Context, user *entity.User) (*AuthResult, error) {
	// Generate access token with info
	accessTokenInfo, err := uc.jwtAuth.GenerateAccessTokenWithInfo(user.ID, string(user.Role), user.Email)
	if err != nil {
		return nil, err
	}

	// Generate refresh token with info
	refreshTokenInfo, err := uc.jwtAuth.GenerateRefreshTokenWithInfo(user.ID)
	if err != nil {
		return nil, err
	}

	// Track tokens cho user (để support logout all devices)
	blacklistService := uc.jwtAuth.GetBlacklistService()
	if blacklistService != nil {
		// Track access token
		_ = blacklistService.TrackUserToken(ctx, user.ID, accessTokenInfo.JTI, accessTokenInfo.TTL)
		// Track refresh token
		_ = blacklistService.TrackUserToken(ctx, user.ID, refreshTokenInfo.JTI, refreshTokenInfo.TTL)
	}

	return &AuthResult{
		User:         user,
		AccessToken:  accessTokenInfo.Token,
		RefreshToken: refreshTokenInfo.Token,
		ExpiresIn:    int64(accessTokenInfo.TTL.Seconds()),
	}, nil
}

// GetJWTAuth trả về JWTAuthMiddleware (dùng cho logout)
func (uc *AuthUseCase) GetJWTAuth() *middleware.JWTAuthMiddleware {
	return uc.jwtAuth
}

// LogoutAllDevices đăng xuất khỏi tất cả thiết bị
// Revoke tất cả tokens của user
func (uc *AuthUseCase) LogoutAllDevices(ctx context.Context, userID string) error {
	blacklistService := uc.jwtAuth.GetBlacklistService()
	if blacklistService == nil {
		return nil // Blacklist không được bật
	}

	return blacklistService.RevokeAllUserTokens(ctx, userID)
}

// GetActiveSessionCount lấy số lượng session active của user
func (uc *AuthUseCase) GetActiveSessionCount(ctx context.Context, userID string) (int64, error) {
	blacklistService := uc.jwtAuth.GetBlacklistService()
	if blacklistService == nil {
		return 0, nil
	}

	return blacklistService.GetUserActiveTokenCount(ctx, userID)
}

// VerifyEmail xác thực email của user với token
func (uc *AuthUseCase) VerifyEmail(ctx context.Context, token string) error {
	if uc.emailVerificationService == nil {
		return ErrInvalidVerificationToken
	}

	// Validate token và lấy userID
	userID, err := uc.emailVerificationService.ValidateToken(ctx, token)
	if err != nil {
		return ErrInvalidVerificationToken
	}

	// Lấy user
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Check if email already verified
	if user.IsEmailVerified {
		return ErrEmailAlreadyVerified
	}

	// Update user's email verification status
	user.VerifyEmail()
	if err := uc.userRepo.Save(ctx, user); err != nil {
		return err
	}

	// Invalidate all verification tokens của user
	_ = uc.emailVerificationService.InvalidateAllUserTokens(ctx, userID)

	return nil
}

// ResendCooldownError chứa thông tin về cooldown khi gửi lại email
type ResendCooldownError struct {
	RemainingSeconds int64
}

func (e *ResendCooldownError) Error() string {
	return "vui lòng đợi trước khi gửi lại email xác thực"
}

// ResendVerificationEmail gửi lại email xác thực
func (uc *AuthUseCase) ResendVerificationEmail(ctx context.Context, userID string) error {
	if uc.emailVerificationService == nil || uc.emailService == nil {
		return errors.New("email verification service không khả dụng")
	}

	// Lấy user
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Check if email already verified
	if user.IsEmailVerified {
		return ErrEmailAlreadyVerified
	}

	// Check cooldown
	canResend, remainingSeconds, err := uc.emailVerificationService.CanResend(ctx, userID)
	if err != nil {
		return err
	}
	if !canResend {
		return &ResendCooldownError{RemainingSeconds: remainingSeconds}
	}

	// Invalidate old tokens
	_ = uc.emailVerificationService.InvalidateAllUserTokens(ctx, userID)

	// Generate new token
	token, err := uc.emailVerificationService.GenerateToken(ctx, userID)
	if err != nil {
		return err
	}

	// Set cooldown
	_ = uc.emailVerificationService.SetResendCooldown(ctx, userID)

	// Send email
	return uc.emailService.SendVerificationEmail(ctx, user.Email, token)
}
