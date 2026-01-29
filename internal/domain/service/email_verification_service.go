// Package service chứa các Domain Service interfaces
package service

import (
	"context"
)

// EmailVerificationService interface cho việc quản lý email verification tokens
// Sử dụng Redis để lưu trữ tokens với TTL
type EmailVerificationService interface {
	// GenerateToken tạo token mới cho user
	// Token được lưu trong Redis với TTL (mặc định 24h)
	GenerateToken(ctx context.Context, userID string) (string, error)

	// ValidateToken kiểm tra token và trả về userID nếu hợp lệ
	// Trả về error nếu token không tồn tại hoặc đã hết hạn
	ValidateToken(ctx context.Context, token string) (string, error)

	// InvalidateToken xóa một token cụ thể
	InvalidateToken(ctx context.Context, token string) error

	// InvalidateAllUserTokens xóa tất cả tokens của một user
	InvalidateAllUserTokens(ctx context.Context, userID string) error

	// CanResend kiểm tra xem có thể gửi lại email verification không
	// Trả về true nếu có thể gửi, remainingSeconds là số giây còn lại phải chờ
	CanResend(ctx context.Context, userID string) (canResend bool, remainingSeconds int64, err error)

	// SetResendCooldown đặt cooldown sau khi gửi email
	SetResendCooldown(ctx context.Context, userID string) error
}
