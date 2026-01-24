package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"restaurant_project/internal/infrastructure/config"

	"github.com/redis/go-redis/v9"
)

// Đảm bảo RedisConnection implement IConnection
var _ IConnection = (*RedisConnection)(nil)

// RedisConnection quản lý kết nối Redis
type RedisConnection struct {
	client  *redis.Client
	config  *config.RedisConfig
	options ConnectionOptions
	state   ConnectionState
	mu      sync.RWMutex
}

// NewRedisConnection tạo instance mới (chưa kết nối)
func NewRedisConnection(cfg *config.RedisConfig, opts ...ConnectionOptions) *RedisConnection {
	options := DefaultConnectionOptions()
	if len(opts) > 0 {
		options = opts[0]
	}

	return &RedisConnection{
		config:  cfg,
		options: options,
		state:   StateDisconnected,
	}
}

// Connect thiết lập kết nối đến Redis với retry logic
func (c *RedisConnection) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state == StateConnected && c.client != nil {
		return nil
	}

	c.state = StateConnecting

	var lastErr error
	for attempt := 0; attempt <= c.options.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				c.state = StateError
				return ctx.Err()
			case <-time.After(c.options.RetryInterval * time.Duration(attempt)):
			}
		}

		if err := c.connect(ctx); err != nil {
			lastErr = err
			continue
		}

		c.state = StateConnected
		return nil
	}

	c.state = StateError
	return fmt.Errorf("không thể kết nối Redis sau %d lần thử: %w", c.options.MaxRetries+1, lastErr)
}

// connect thực hiện kết nối thực tế
func (c *RedisConnection) connect(ctx context.Context) error {
	opts := &redis.Options{
		Addr:            c.config.Addr,
		Password:        c.config.Password,
		DB:              c.config.DB,
		PoolSize:        c.options.MaxOpenConns,
		MinIdleConns:    c.options.MaxIdleConns,
		ConnMaxIdleTime: c.options.ConnMaxIdleTime,
		ConnMaxLifetime: c.options.ConnMaxLifetime,
	}

	client := redis.NewClient(opts)

	// Ping để verify
	pingCtx, cancel := context.WithTimeout(ctx, c.options.PingTimeout)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		client.Close()
		return fmt.Errorf("không thể ping Redis: %w", err)
	}

	c.client = client
	return nil
}

// Disconnect đóng kết nối
func (c *RedisConnection) Disconnect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client == nil {
		c.state = StateDisconnected
		return nil
	}

	if err := c.client.Close(); err != nil {
		c.state = StateError
		return fmt.Errorf("lỗi đóng kết nối Redis: %w", err)
	}

	c.client = nil
	c.state = StateDisconnected
	return nil
}

// Ping kiểm tra kết nối
func (c *RedisConnection) Ping(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.client == nil {
		return fmt.Errorf("chưa kết nối")
	}

	return c.client.Ping(ctx).Err()
}

// HealthCheck trả về trạng thái chi tiết
func (c *RedisConnection) HealthCheck(ctx context.Context) HealthStatus {
	start := time.Now()

	status := HealthStatus{
		CheckedAt: start,
		Details:   make(map[string]any),
	}

	c.mu.RLock()
	status.State = c.state
	c.mu.RUnlock()

	if c.client == nil {
		status.Healthy = false
		status.Message = "Chưa kết nối"
		return status
	}

	// Ping với timeout
	pingCtx, cancel := context.WithTimeout(ctx, c.options.PingTimeout)
	defer cancel()

	if err := c.client.Ping(pingCtx).Err(); err != nil {
		status.Healthy = false
		status.State = StateError
		status.Message = err.Error()
		return status
	}

	status.Latency = time.Since(start)
	status.Healthy = true
	status.State = StateConnected
	status.Message = "OK"

	// Lấy pool stats
	poolStats := c.client.PoolStats()
	status.Details["hits"] = poolStats.Hits
	status.Details["misses"] = poolStats.Misses
	status.Details["timeouts"] = poolStats.Timeouts
	status.Details["total_conns"] = poolStats.TotalConns
	status.Details["idle_conns"] = poolStats.IdleConns
	status.Details["stale_conns"] = poolStats.StaleConns

	return status
}

// Info trả về thông tin connection
func (c *RedisConnection) Info() ConnectionInfo {
	return ConnectionInfo{
		Name:     "redis",
		Type:     TypeRedis,
		Host:     c.config.Addr,
		Database: fmt.Sprintf("db%d", c.config.DB),
	}
}

// IsConnected kiểm tra trạng thái
func (c *RedisConnection) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state == StateConnected && c.client != nil
}

// Reconnect thử kết nối lại
func (c *RedisConnection) Reconnect(ctx context.Context) error {
	if err := c.Disconnect(ctx); err != nil {
		return err
	}
	return c.Connect(ctx)
}

// Client trả về *redis.Client
func (c *RedisConnection) Client() *redis.Client {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.client
}
