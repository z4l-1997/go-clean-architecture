// Package service chứa các Infrastructure Service implementations
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"restaurant_project/internal/domain/service"
	"restaurant_project/internal/infrastructure/config"
)

// Đảm bảo RedisLoginAttemptService implement LoginAttemptService
var _ service.LoginAttemptService = (*RedisLoginAttemptService)(nil)

const (
	// Key patterns
	loginAttemptsKeyPrefix = "login_attempts:"
	accountLockedKeyPrefix = "account_locked:"
)

// RedisLoginAttemptService implementation của LoginAttemptService sử dụng Redis
type RedisLoginAttemptService struct {
	client          *redis.Client
	maxAttempts     int
	lockoutDuration time.Duration
	enabled         bool
}

// NewRedisLoginAttemptService tạo mới RedisLoginAttemptService
func NewRedisLoginAttemptService(
	client *redis.Client,
	cfg config.AccountLockoutConfig,
) *RedisLoginAttemptService {
	return &RedisLoginAttemptService{
		client:          client,
		maxAttempts:     cfg.MaxAttempts,
		lockoutDuration: cfg.LockoutDuration,
		enabled:         cfg.Enabled,
	}
}

// IncrementAttempts tăng số lần login thất bại
func (s *RedisLoginAttemptService) IncrementAttempts(ctx context.Context, username string) (int, error) {
	if !s.enabled {
		return 0, nil
	}

	key := loginAttemptsKeyPrefix + username

	// Increment và set TTL = lockoutDuration
	pipe := s.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, s.lockoutDuration)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to increment login attempts: %w", err)
	}

	attempts := int(incr.Val())

	// Nếu vượt quá max attempts, khóa tài khoản
	if attempts >= s.maxAttempts {
		if err := s.Lock(ctx, username); err != nil {
			return attempts, err
		}
	}

	return attempts, nil
}

// GetAttempts lấy số lần login thất bại hiện tại
func (s *RedisLoginAttemptService) GetAttempts(ctx context.Context, username string) (int, error) {
	if !s.enabled {
		return 0, nil
	}

	key := loginAttemptsKeyPrefix + username
	val, err := s.client.Get(ctx, key).Int()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get login attempts: %w", err)
	}

	return val, nil
}

// ResetAttempts xóa counter khi login thành công
func (s *RedisLoginAttemptService) ResetAttempts(ctx context.Context, username string) error {
	if !s.enabled {
		return nil
	}

	// Xóa cả attempts counter và locked key
	pipe := s.client.Pipeline()
	pipe.Del(ctx, loginAttemptsKeyPrefix+username)
	pipe.Del(ctx, accountLockedKeyPrefix+username)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to reset login attempts: %w", err)
	}

	return nil
}

// IsLocked kiểm tra xem tài khoản có bị khóa không
func (s *RedisLoginAttemptService) IsLocked(ctx context.Context, username string) (bool, error) {
	if !s.enabled {
		return false, nil
	}

	key := accountLockedKeyPrefix + username
	exists, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check account lock status: %w", err)
	}

	return exists > 0, nil
}

// Lock khóa tài khoản
func (s *RedisLoginAttemptService) Lock(ctx context.Context, username string) error {
	if !s.enabled {
		return nil
	}

	key := accountLockedKeyPrefix + username
	err := s.client.Set(ctx, key, "1", s.lockoutDuration).Err()
	if err != nil {
		return fmt.Errorf("failed to lock account: %w", err)
	}

	return nil
}

// GetRemainingLockTime lấy thời gian còn lại bị khóa (giây)
func (s *RedisLoginAttemptService) GetRemainingLockTime(ctx context.Context, username string) (int64, error) {
	if !s.enabled {
		return 0, nil
	}

	key := accountLockedKeyPrefix + username
	ttl, err := s.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get remaining lock time: %w", err)
	}

	if ttl < 0 {
		return 0, nil
	}

	return int64(ttl.Seconds()), nil
}

// MaxAttempts trả về số lần thử tối đa
func (s *RedisLoginAttemptService) MaxAttempts() int {
	return s.maxAttempts
}

// LockoutDuration trả về thời gian khóa
func (s *RedisLoginAttemptService) LockoutDuration() time.Duration {
	return s.lockoutDuration
}

// IsEnabled kiểm tra xem service có được bật không
func (s *RedisLoginAttemptService) IsEnabled() bool {
	return s.enabled
}
