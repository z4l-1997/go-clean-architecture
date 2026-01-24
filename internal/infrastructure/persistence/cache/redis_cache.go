// Package cache chứa các cache repository implementations
package cache

import (
	"context"
	"time"

	"restaurant_project/internal/domain/cache"

	"github.com/redis/go-redis/v9"
)

// Đảm bảo RedisCacheRepository implement ICacheRepository
var _ cache.ICacheRepository = (*RedisCacheRepository)(nil)

// RedisCacheRepository implementation của ICacheRepository sử dụng Redis
type RedisCacheRepository struct {
	client  *redis.Client
	options cache.CacheOptions
}

// NewRedisCacheRepository tạo instance mới
func NewRedisCacheRepository(client *redis.Client, opts ...cache.CacheOptions) *RedisCacheRepository {
	options := cache.DefaultCacheOptions()
	if len(opts) > 0 {
		options = opts[0]
	}

	return &RedisCacheRepository{
		client:  client,
		options: options,
	}
}

// buildKey tạo key với prefix
func (r *RedisCacheRepository) buildKey(key string) string {
	return r.options.KeyPrefix + key
}

// Get lấy giá trị từ cache
func (r *RedisCacheRepository) Get(ctx context.Context, key string) ([]byte, error) {
	fullKey := r.buildKey(key)

	val, err := r.client.Get(ctx, fullKey).Bytes()
	if err == redis.Nil {
		return nil, nil // Key không tồn tại
	}
	if err != nil {
		return nil, err
	}

	return val, nil
}

// Set lưu giá trị vào cache
func (r *RedisCacheRepository) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	fullKey := r.buildKey(key)

	if ttl == 0 {
		ttl = r.options.DefaultTTL
	}

	return r.client.Set(ctx, fullKey, value, ttl).Err()
}

// Delete xóa key khỏi cache
func (r *RedisCacheRepository) Delete(ctx context.Context, key string) error {
	fullKey := r.buildKey(key)
	return r.client.Del(ctx, fullKey).Err()
}

// Exists kiểm tra key có tồn tại không
func (r *RedisCacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := r.buildKey(key)

	count, err := r.client.Exists(ctx, fullKey).Result()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// DeleteByPattern xóa tất cả keys match pattern
func (r *RedisCacheRepository) DeleteByPattern(ctx context.Context, pattern string) error {
	fullPattern := r.buildKey(pattern)

	iter := r.client.Scan(ctx, 0, fullPattern, 100).Iterator()
	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	return r.client.Del(ctx, keys...).Err()
}

// GetOrSet lấy từ cache, nếu không có thì gọi loader và cache kết quả
func (r *RedisCacheRepository) GetOrSet(ctx context.Context, key string, ttl time.Duration, loader func() ([]byte, error)) ([]byte, error) {
	// Thử lấy từ cache trước
	data, err := r.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if data != nil {
		return data, nil // Cache hit
	}

	// Cache miss - gọi loader
	data, err = loader()
	if err != nil {
		return nil, err
	}

	// Cache kết quả
	if err := r.Set(ctx, key, data, ttl); err != nil {
		// Log lỗi nhưng vẫn trả về data
		// Trong production nên log error ở đây
	}

	return data, nil
}
