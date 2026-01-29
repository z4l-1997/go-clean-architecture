// Package dto chứa Data Transfer Objects
package dto

// ============================================
// AUTH REQUEST DTOs
// ============================================

// RegisterRequest là dữ liệu để đăng ký tài khoản mới (Public - Customer only)
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50" example:"john_doe"`
	Email    string `json:"email" binding:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required,min=6,max=100" example:"password123"`
}

// LoginRequest là dữ liệu để đăng nhập
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"john_doe"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// RefreshTokenRequest là dữ liệu để refresh access token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIs..."`
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIs..."` // Optional: gửi kèm access token cũ để cleanup
}

// ============================================
// AUTH RESPONSE DTOs
// ============================================

// AuthResponse là dữ liệu trả về sau khi đăng nhập/đăng ký thành công
type AuthResponse struct {
	AccessToken  string       `json:"access_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	RefreshToken string       `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	TokenType    string       `json:"token_type" example:"Bearer"`
	ExpiresIn    int64        `json:"expires_in" example:"900"` // seconds
	User         UserResponse `json:"user"`
}

// RefreshTokenResponse là dữ liệu trả về sau khi refresh token thành công
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	TokenType    string `json:"token_type" example:"Bearer"`
	ExpiresIn    int64  `json:"expires_in" example:"900"` // seconds
}

// LogoutAllResponse là dữ liệu trả về sau khi logout tất cả thiết bị
type LogoutAllResponse struct {
	RevokedSessions int64  `json:"revoked_sessions" example:"3"`
	Message         string `json:"message" example:"Đã đăng xuất khỏi tất cả thiết bị"`
}

// ActiveSessionsResponse là dữ liệu trả về số session active
type ActiveSessionsResponse struct {
	ActiveSessions int64 `json:"active_sessions" example:"3"`
}

// ============================================
// EMAIL VERIFICATION DTOs
// ============================================

// VerifyEmailRequest là dữ liệu để xác thực email
type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required" example:"abc123def456..."`
}

// VerifyEmailResponse là dữ liệu trả về sau khi xác thực email thành công
type VerifyEmailResponse struct {
	Message string `json:"message" example:"Email đã được xác thực thành công"`
}

// ResendVerificationResponse là dữ liệu trả về sau khi gửi lại email xác thực
type ResendVerificationResponse struct {
	Message string `json:"message" example:"Email xác thực đã được gửi lại"`
}

// ResendCooldownResponse là dữ liệu trả về khi đang trong thời gian chờ
type ResendCooldownResponse struct {
	Message          string `json:"message" example:"Vui lòng đợi trước khi gửi lại"`
	RemainingSeconds int64  `json:"remaining_seconds" example:"45"`
}
