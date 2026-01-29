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

// Đảm bảo RedisTokenBlacklistService implement TokenBlacklistService
var _ service.TokenBlacklistService = (*RedisTokenBlacklistService)(nil)

const (
	// Key pattern cho token blacklist
	tokenBlacklistKeyPrefix = "token_blacklist:"
	// Key pattern cho user tokens tracking
	userTokensKeyPrefix = "user_tokens:"
	// Key pattern cho token -> user mapping (để biết token thuộc user nào)
	tokenUserKeyPrefix = "token_user:"
)

// RedisTokenBlacklistService implementation của TokenBlacklistService sử dụng Redis
type RedisTokenBlacklistService struct {
	client  *redis.Client
	enabled bool
}

// NewRedisTokenBlacklistService tạo mới RedisTokenBlacklistService
func NewRedisTokenBlacklistService(
	client *redis.Client,
	cfg config.TokenBlacklistConfig,
) *RedisTokenBlacklistService {
	return &RedisTokenBlacklistService{
		client:  client,
		enabled: cfg.Enabled,
	}
}

// Blacklist thêm token vào blacklist với TTL
func (s *RedisTokenBlacklistService) Blacklist(ctx context.Context, jti string, ttl time.Duration) error {
	if !s.enabled {
		return nil
	}

	if jti == "" {
		return fmt.Errorf("jti cannot be empty")
	}

	// Đảm bảo TTL ít nhất 1 giây
	if ttl < time.Second {
		ttl = time.Second
	}

	key := tokenBlacklistKeyPrefix + jti
	err := s.client.Set(ctx, key, "1", ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	return nil
}

// IsBlacklisted kiểm tra xem token có trong blacklist không
func (s *RedisTokenBlacklistService) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	if !s.enabled {
		return false, nil
	}

	if jti == "" {
		return false, nil
	}

	key := tokenBlacklistKeyPrefix + jti
	exists, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check token blacklist: %w", err)
	}

	return exists > 0, nil
}

// Remove xóa token khỏi blacklist
func (s *RedisTokenBlacklistService) Remove(ctx context.Context, jti string) error {
	if !s.enabled {
		return nil
	}

	key := tokenBlacklistKeyPrefix + jti
	err := s.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to remove token from blacklist: %w", err)
	}

	return nil
}

// IsEnabled kiểm tra xem service có được bật không
func (s *RedisTokenBlacklistService) IsEnabled() bool {
	return s.enabled
}

// ==================== P1: Multi-Device Support ====================

// TrackUserToken theo dõi token active của user
// Redis structure:
//   - user_tokens:{userID} = SET of {jti1, jti2, ...}
//   - token_user:{jti} = userID (để reverse lookup)
func (s *RedisTokenBlacklistService) TrackUserToken(ctx context.Context, userID string, jti string, ttl time.Duration) error {
	if !s.enabled {
		return nil
	}

	if userID == "" || jti == "" {
		return fmt.Errorf("userID and jti cannot be empty")
	}

	// Đảm bảo TTL ít nhất 1 giây
	if ttl < time.Second {
		ttl = time.Second
	}

	userTokensKey := userTokensKeyPrefix + userID
	tokenUserKey := tokenUserKeyPrefix + jti

	// Kiểm tra TTL hiện tại của user_tokens SET
	// Chỉ ghi đè TTL nếu TTL mới lớn hơn, tránh trường hợp
	// access token (15m) ghi đè lên refresh token (7d)
	currentTTL, err := s.client.TTL(ctx, userTokensKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get current TTL: %w", err)
	}

	pipe := s.client.Pipeline()

	// Thêm jti vào SET của user
	pipe.SAdd(ctx, userTokensKey, jti)

	// Chỉ set TTL khi key chưa tồn tại (currentTTL < 0) hoặc TTL mới lớn hơn
	if currentTTL < 0 || ttl > currentTTL {
		pipe.Expire(ctx, userTokensKey, ttl)
	}

	// Lưu mapping token -> user (để biết token thuộc user nào khi cần)
	pipe.Set(ctx, tokenUserKey, userID, ttl)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to track user token: %w", err)
	}

	return nil
}

// GetUserActiveTokens lấy danh sách tất cả token active của user
func (s *RedisTokenBlacklistService) GetUserActiveTokens(ctx context.Context, userID string) ([]string, error) {
	if !s.enabled {
		return []string{}, nil
	}

	if userID == "" {
		return []string{}, nil
	}

	userTokensKey := userTokensKeyPrefix + userID
	tokens, err := s.client.SMembers(ctx, userTokensKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user active tokens: %w", err)
	}

	// Lọc ra những token đã bị blacklist
	activeTokens := make([]string, 0, len(tokens))
	for _, jti := range tokens {
		blacklisted, err := s.IsBlacklisted(ctx, jti)
		if err != nil {
			continue // Skip on error
		}
		if !blacklisted {
			activeTokens = append(activeTokens, jti)
		}
	}

	return activeTokens, nil
}

// RevokeAllUserTokens thu hồi tất cả token của user
// Flow:
//  1. Lấy tất cả jti từ user_tokens:{userID}
//  2. Blacklist từng jti
//  3. Xóa user_tokens:{userID} set
func (s *RedisTokenBlacklistService) RevokeAllUserTokens(ctx context.Context, userID string) error {
	if !s.enabled {
		return nil
	}

	if userID == "" {
		return fmt.Errorf("userID cannot be empty")
	}

	userTokensKey := userTokensKeyPrefix + userID

	// Lấy tất cả token của user
	tokens, err := s.client.SMembers(ctx, userTokensKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get user tokens: %w", err)
	}

	if len(tokens) == 0 {
		return nil // Không có token nào để revoke
	}

	pipe := s.client.Pipeline()

	// Blacklist tất cả token
	// TTL mặc định 24h (đủ cho access token, refresh token sẽ tự expire)
	defaultBlacklistTTL := 24 * time.Hour
	for _, jti := range tokens {
		blacklistKey := tokenBlacklistKeyPrefix + jti
		pipe.Set(ctx, blacklistKey, "1", defaultBlacklistTTL)

		// Xóa token -> user mapping
		tokenUserKey := tokenUserKeyPrefix + jti
		pipe.Del(ctx, tokenUserKey)
	}

	// Xóa user tokens set
	pipe.Del(ctx, userTokensKey)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to revoke all user tokens: %w", err)
	}

	return nil
}

// GetUserActiveTokenCount lấy số lượng token active của user
func (s *RedisTokenBlacklistService) GetUserActiveTokenCount(ctx context.Context, userID string) (int64, error) {
	if !s.enabled {
		return 0, nil
	}

	if userID == "" {
		return 0, nil
	}

	userTokensKey := userTokensKeyPrefix + userID
	count, err := s.client.SCard(ctx, userTokensKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get user token count: %w", err)
	}

	return count, nil
}

// UntrackUserToken xóa token khỏi tracking khi logout đơn lẻ
// (Helper method, không bắt buộc trong interface)
func (s *RedisTokenBlacklistService) UntrackUserToken(ctx context.Context, userID string, jti string) error {
	if !s.enabled {
		return nil
	}

	if userID == "" || jti == "" {
		return nil
	}

	pipe := s.client.Pipeline()

	userTokensKey := userTokensKeyPrefix + userID
	pipe.SRem(ctx, userTokensKey, jti)

	tokenUserKey := tokenUserKeyPrefix + jti
	pipe.Del(ctx, tokenUserKey)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to untrack user token: %w", err)
	}

	return nil
}
