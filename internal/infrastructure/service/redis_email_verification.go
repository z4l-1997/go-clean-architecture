// Package service chứa các Infrastructure Service implementations
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"restaurant_project/internal/domain/service"
	"restaurant_project/internal/infrastructure/config"
)

// Đảm bảo RedisEmailVerificationService implement EmailVerificationService
var _ service.EmailVerificationService = (*RedisEmailVerificationService)(nil)

const (
	// Key patterns for email verification
	emailVerifyTokenKeyPrefix    = "email_verify:"       // email_verify:{token} -> userID
	emailVerifyUserKeyPrefix     = "email_verify_user:"  // email_verify_user:{userID} -> Set of tokens
	emailVerifyCooldownKeyPrefix = "email_verify_cd:"    // email_verify_cd:{userID} -> "1"
)

// RedisEmailVerificationService implementation của EmailVerificationService sử dụng Redis
type RedisEmailVerificationService struct {
	client      *redis.Client
	tokenTTL    time.Duration
	cooldownTTL time.Duration
	enabled     bool
}

// NewRedisEmailVerificationService tạo mới RedisEmailVerificationService
func NewRedisEmailVerificationService(
	client *redis.Client,
	cfg config.EmailVerificationConfig,
) *RedisEmailVerificationService {
	return &RedisEmailVerificationService{
		client:      client,
		tokenTTL:    cfg.TokenTTL,
		cooldownTTL: cfg.CooldownTTL,
		enabled:     cfg.Enabled,
	}
}

// GenerateToken tạo token mới cho user
func (s *RedisEmailVerificationService) GenerateToken(ctx context.Context, userID string) (string, error) {
	if !s.enabled {
		return "", nil
	}

	// Generate random token (32 bytes = 64 hex characters)
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)

	// Store token -> userID mapping
	tokenKey := emailVerifyTokenKeyPrefix + token
	userKey := emailVerifyUserKeyPrefix + userID

	pipe := s.client.Pipeline()

	// Set token -> userID với TTL
	pipe.Set(ctx, tokenKey, userID, s.tokenTTL)

	// Add token to user's token set (để có thể invalidate all tokens của user)
	pipe.SAdd(ctx, userKey, token)
	pipe.Expire(ctx, userKey, s.tokenTTL)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to store verification token: %w", err)
	}

	return token, nil
}

// ValidateToken kiểm tra token và trả về userID nếu hợp lệ
func (s *RedisEmailVerificationService) ValidateToken(ctx context.Context, token string) (string, error) {
	if !s.enabled {
		return "", fmt.Errorf("email verification is disabled")
	}

	tokenKey := emailVerifyTokenKeyPrefix + token

	userID, err := s.client.Get(ctx, tokenKey).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("invalid or expired verification token")
	}
	if err != nil {
		return "", fmt.Errorf("failed to validate token: %w", err)
	}

	return userID, nil
}

// InvalidateToken xóa một token cụ thể
func (s *RedisEmailVerificationService) InvalidateToken(ctx context.Context, token string) error {
	if !s.enabled {
		return nil
	}

	tokenKey := emailVerifyTokenKeyPrefix + token

	// Get userID first to remove from user's token set
	userID, err := s.client.Get(ctx, tokenKey).Result()
	if err == redis.Nil {
		// Token already expired or doesn't exist
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	userKey := emailVerifyUserKeyPrefix + userID

	pipe := s.client.Pipeline()
	pipe.Del(ctx, tokenKey)
	pipe.SRem(ctx, userKey, token)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to invalidate token: %w", err)
	}

	return nil
}

// InvalidateAllUserTokens xóa tất cả tokens của một user
func (s *RedisEmailVerificationService) InvalidateAllUserTokens(ctx context.Context, userID string) error {
	if !s.enabled {
		return nil
	}

	userKey := emailVerifyUserKeyPrefix + userID

	// Get all tokens of user
	tokens, err := s.client.SMembers(ctx, userKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get user tokens: %w", err)
	}

	if len(tokens) == 0 {
		return nil
	}

	// Delete all token keys
	pipe := s.client.Pipeline()
	for _, token := range tokens {
		pipe.Del(ctx, emailVerifyTokenKeyPrefix+token)
	}
	pipe.Del(ctx, userKey)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to invalidate all user tokens: %w", err)
	}

	return nil
}

// CanResend kiểm tra xem có thể gửi lại email verification không
func (s *RedisEmailVerificationService) CanResend(ctx context.Context, userID string) (bool, int64, error) {
	if !s.enabled {
		return true, 0, nil
	}

	cooldownKey := emailVerifyCooldownKeyPrefix + userID

	ttl, err := s.client.TTL(ctx, cooldownKey).Result()
	if err != nil {
		return false, 0, fmt.Errorf("failed to check resend cooldown: %w", err)
	}

	// Key doesn't exist or has expired
	if ttl < 0 {
		return true, 0, nil
	}

	return false, int64(ttl.Seconds()), nil
}

// SetResendCooldown đặt cooldown sau khi gửi email
func (s *RedisEmailVerificationService) SetResendCooldown(ctx context.Context, userID string) error {
	if !s.enabled {
		return nil
	}

	cooldownKey := emailVerifyCooldownKeyPrefix + userID

	err := s.client.Set(ctx, cooldownKey, "1", s.cooldownTTL).Err()
	if err != nil {
		return fmt.Errorf("failed to set resend cooldown: %w", err)
	}

	return nil
}

// IsEnabled kiểm tra xem service có được bật không
func (s *RedisEmailVerificationService) IsEnabled() bool {
	return s.enabled
}
