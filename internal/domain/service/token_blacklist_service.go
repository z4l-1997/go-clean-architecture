// Package service chứa các Domain Service interfaces
package service

import (
	"context"
	"time"
)

// TokenBlacklistService interface cho việc blacklist JWT tokens
// Dùng để revoke tokens khi user logout
type TokenBlacklistService interface {
	// Blacklist thêm token vào blacklist với TTL
	// jti: JWT ID (unique identifier của token)
	// ttl: Thời gian giữ trong blacklist (thường = thời gian còn lại của token)
	Blacklist(ctx context.Context, jti string, ttl time.Duration) error

	// IsBlacklisted kiểm tra xem token có trong blacklist không
	IsBlacklisted(ctx context.Context, jti string) (bool, error)

	// Remove xóa token khỏi blacklist (hiếm khi dùng)
	Remove(ctx context.Context, jti string) error

	// ==================== P1: Multi-Device Support ====================

	// TrackUserToken theo dõi token active của user
	// Gọi khi login thành công để track token theo userID
	// userID: ID của user
	// jti: JWT ID của token
	// ttl: Thời gian sống của token (để auto-cleanup)
	TrackUserToken(ctx context.Context, userID string, jti string, ttl time.Duration) error

	// GetUserActiveTokens lấy danh sách tất cả token active của user
	// Trả về slice các jti
	GetUserActiveTokens(ctx context.Context, userID string) ([]string, error)

	// RevokeAllUserTokens thu hồi tất cả token của user
	// Dùng khi: đổi password, bị hack, logout all devices
	RevokeAllUserTokens(ctx context.Context, userID string) error

	// UntrackUserToken xóa token khỏi user_tokens SET và token_user mapping
	// Dùng khi: refresh token (xóa access token cũ đã hết hạn khỏi SET)
	UntrackUserToken(ctx context.Context, userID string, jti string) error

	// GetUserActiveTokenCount lấy số lượng token active của user
	GetUserActiveTokenCount(ctx context.Context, userID string) (int64, error)
}
