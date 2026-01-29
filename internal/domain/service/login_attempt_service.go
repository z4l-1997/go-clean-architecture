// Package service chứa các Domain Service interfaces
package service

import (
	"context"
)

// LoginAttemptService interface cho việc tracking login attempts
// Dùng để implement Account Lockout - khóa tài khoản sau nhiều lần login sai
type LoginAttemptService interface {
	// IncrementAttempts tăng số lần login thất bại cho username
	// Trả về số lần attempt hiện tại
	IncrementAttempts(ctx context.Context, username string) (int, error)

	// GetAttempts lấy số lần login thất bại hiện tại
	GetAttempts(ctx context.Context, username string) (int, error)

	// ResetAttempts xóa counter khi login thành công
	ResetAttempts(ctx context.Context, username string) error

	// IsLocked kiểm tra xem tài khoản có bị khóa không
	IsLocked(ctx context.Context, username string) (bool, error)

	// Lock khóa tài khoản (gọi khi vượt quá max attempts)
	Lock(ctx context.Context, username string) error

	// GetRemainingLockTime lấy thời gian còn lại bị khóa (giây)
	// Trả về 0 nếu không bị khóa
	GetRemainingLockTime(ctx context.Context, username string) (int64, error)
}
