// Package cache định nghĩa interface cho caching layer
package cache

import (
	"context"
	"time"
)

// ICacheRepository interface chung cho tất cả cache implementations
type ICacheRepository interface {
	// Get lấy giá trị từ cache theo key
	// Trả về nil nếu key không tồn tại
	Get(ctx context.Context, key string) ([]byte, error)

	// Set lưu giá trị vào cache với TTL
	// TTL = 0 nghĩa là không hết hạn
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete xóa key khỏi cache
	Delete(ctx context.Context, key string) error

	// Exists kiểm tra key có tồn tại không
	Exists(ctx context.Context, key string) (bool, error)

	// DeleteByPattern xóa tất cả keys match pattern
	// VD: "mon_an:*" sẽ xóa tất cả keys bắt đầu bằng "mon_an:"
	DeleteByPattern(ctx context.Context, pattern string) error

	// GetOrSet lấy từ cache, nếu không có thì gọi loader và cache kết quả
	GetOrSet(ctx context.Context, key string, ttl time.Duration, loader func() ([]byte, error)) ([]byte, error)
}

// CacheOptions tùy chọn cho cache
type CacheOptions struct {
	DefaultTTL time.Duration // TTL mặc định
	KeyPrefix  string        // Prefix cho tất cả keys
}

// DefaultCacheOptions trả về options mặc định
func DefaultCacheOptions() CacheOptions {
	return CacheOptions{
		DefaultTTL: 5 * time.Minute,
		KeyPrefix:  "restaurant:",
	}
}
