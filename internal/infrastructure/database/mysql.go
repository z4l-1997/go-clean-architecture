package database

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"restaurant_project/internal/infrastructure/config"

	_ "github.com/go-sql-driver/mysql"
)

// Đảm bảo MySQLConnection implement IConnection
var _ IConnection = (*MySQLConnection)(nil)

// MySQLConnection quản lý kết nối MySQL
type MySQLConnection struct {
	db      *sql.DB
	config  *config.MySQLConfig
	options ConnectionOptions
	state   ConnectionState
	mu      sync.RWMutex
}

// NewMySQLConnection tạo instance mới (chưa kết nối)
func NewMySQLConnection(cfg *config.MySQLConfig, opts ...ConnectionOptions) *MySQLConnection {
	options := DefaultConnectionOptions()
	if len(opts) > 0 {
		options = opts[0]
	}

	return &MySQLConnection{
		config:  cfg,
		options: options,
		state:   StateDisconnected,
	}
}

// Connect thiết lập kết nối đến MySQL với retry logic
func (c *MySQLConnection) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state == StateConnected && c.db != nil {
		return nil // Đã kết nối
	}

	c.state = StateConnecting

	var lastErr error
	for attempt := 0; attempt <= c.options.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				c.state = StateError
				return ctx.Err()
			case <-time.After(c.options.RetryInterval * time.Duration(attempt)): // Exponential backoff
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
	return fmt.Errorf("không thể kết nối MySQL sau %d lần thử: %w", c.options.MaxRetries+1, lastErr)
}

// connect thực hiện kết nối thực tế
func (c *MySQLConnection) connect(ctx context.Context) error {
	db, err := sql.Open("mysql", c.config.DSN())
	if err != nil {
		return fmt.Errorf("không thể mở kết nối: %w", err)
	}

	// Cấu hình connection pool
	db.SetMaxOpenConns(c.options.MaxOpenConns)
	db.SetMaxIdleConns(c.options.MaxIdleConns)
	db.SetConnMaxLifetime(c.options.ConnMaxLifetime)
	db.SetConnMaxIdleTime(c.options.ConnMaxIdleTime)

	// Test connection
	pingCtx, cancel := context.WithTimeout(ctx, c.options.PingTimeout)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		db.Close()
		return fmt.Errorf("không thể ping: %w", err)
	}

	c.db = db
	return nil
}

// Disconnect đóng kết nối
func (c *MySQLConnection) Disconnect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.db == nil {
		c.state = StateDisconnected
		return nil
	}

	if err := c.db.Close(); err != nil {
		c.state = StateError
		return fmt.Errorf("lỗi đóng kết nối MySQL: %w", err)
	}

	c.db = nil
	c.state = StateDisconnected
	return nil
}

// Ping kiểm tra kết nối
func (c *MySQLConnection) Ping(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.db == nil {
		return fmt.Errorf("chưa kết nối")
	}

	return c.db.PingContext(ctx)
}

// HealthCheck trả về trạng thái chi tiết
func (c *MySQLConnection) HealthCheck(ctx context.Context) HealthStatus {
	start := time.Now()

	status := HealthStatus{
		CheckedAt: start,
		Details:   make(map[string]any),
	}

	c.mu.RLock()
	status.State = c.state
	c.mu.RUnlock()

	if c.db == nil {
		status.Healthy = false
		status.Message = "Chưa kết nối"
		return status
	}

	// Ping với timeout
	pingCtx, cancel := context.WithTimeout(ctx, c.options.PingTimeout)
	defer cancel()

	if err := c.db.PingContext(pingCtx); err != nil {
		status.Healthy = false
		status.State = StateError
		status.Message = err.Error()
		return status
	}

	status.Latency = time.Since(start)
	status.Healthy = true
	status.State = StateConnected
	status.Message = "OK"

	// Thêm pool stats
	stats := c.db.Stats()
	status.Details["open_connections"] = stats.OpenConnections
	status.Details["in_use"] = stats.InUse
	status.Details["idle"] = stats.Idle
	status.Details["max_open_connections"] = stats.MaxOpenConnections
	status.Details["wait_count"] = stats.WaitCount
	status.Details["wait_duration_ms"] = stats.WaitDuration.Milliseconds()

	return status
}

// Info trả về thông tin connection
func (c *MySQLConnection) Info() ConnectionInfo {
	return ConnectionInfo{
		Name:     "mysql",
		Type:     TypeMySQL,
		Host:     c.config.Host,
		Port:     c.config.Port,
		Database: c.config.Database,
	}
}

// IsConnected kiểm tra trạng thái
func (c *MySQLConnection) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state == StateConnected && c.db != nil
}

// Reconnect thử kết nối lại
func (c *MySQLConnection) Reconnect(ctx context.Context) error {
	if err := c.Disconnect(ctx); err != nil {
		return err
	}
	return c.Connect(ctx)
}

// DB trả về *sql.DB để thực hiện queries
func (c *MySQLConnection) DB() *sql.DB {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.db
}

// BeginTx bắt đầu transaction
func (c *MySQLConnection) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.db == nil {
		return nil, fmt.Errorf("chưa kết nối MySQL")
	}

	return c.db.BeginTx(ctx, opts)
}

// Stats trả về thống kê connection pool
func (c *MySQLConnection) Stats() sql.DBStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.db == nil {
		return sql.DBStats{}
	}

	return c.db.Stats()
}
