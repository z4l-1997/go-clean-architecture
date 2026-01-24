package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"restaurant_project/internal/infrastructure/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Đảm bảo MongoDBConnection implement IConnection
var _ IConnection = (*MongoDBConnection)(nil)

// Collection names - constants
const (
	CollectionMonAn = "mon_an" // Collection chứa menu
	CollectionOrder = "orders" // Collection chứa đơn hàng
)

// MongoDBConnection quản lý kết nối MongoDB
type MongoDBConnection struct {
	client   *mongo.Client
	database *mongo.Database
	config   *config.MongoDBConfig
	options  ConnectionOptions
	state    ConnectionState
	mu       sync.RWMutex
}

// NewMongoDBConnection tạo instance mới (chưa kết nối)
func NewMongoDBConnection(cfg *config.MongoDBConfig, opts ...ConnectionOptions) *MongoDBConnection {
	options := DefaultConnectionOptions()
	if len(opts) > 0 {
		options = opts[0]
	}

	return &MongoDBConnection{
		config:  cfg,
		options: options,
		state:   StateDisconnected,
	}
}

// Connect thiết lập kết nối đến MongoDB với retry logic
func (c *MongoDBConnection) Connect(ctx context.Context) error {
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
	return fmt.Errorf("không thể kết nối MongoDB sau %d lần thử: %w", c.options.MaxRetries+1, lastErr)
}

// connect thực hiện kết nối thực tế
func (c *MongoDBConnection) connect(ctx context.Context) error {
	connectCtx, cancel := context.WithTimeout(ctx, c.options.ConnectTimeout)
	defer cancel()

	clientOpts := options.Client().
		ApplyURI(c.config.URI).
		SetMaxPoolSize(uint64(c.options.MaxOpenConns)).
		SetMinPoolSize(uint64(c.options.MaxIdleConns)).
		SetMaxConnIdleTime(c.options.ConnMaxIdleTime)

	client, err := mongo.Connect(connectCtx, clientOpts)
	if err != nil {
		return fmt.Errorf("không thể kết nối: %w", err)
	}

	// Ping để verify
	pingCtx, pingCancel := context.WithTimeout(ctx, c.options.PingTimeout)
	defer pingCancel()

	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		client.Disconnect(ctx)
		return fmt.Errorf("không thể ping: %w", err)
	}

	c.client = client
	c.database = client.Database(c.config.Database)
	return nil
}

// Disconnect đóng kết nối
func (c *MongoDBConnection) Disconnect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client == nil {
		c.state = StateDisconnected
		return nil
	}

	if err := c.client.Disconnect(ctx); err != nil {
		c.state = StateError
		return fmt.Errorf("lỗi đóng kết nối MongoDB: %w", err)
	}

	c.client = nil
	c.database = nil
	c.state = StateDisconnected
	return nil
}

// Ping kiểm tra kết nối
func (c *MongoDBConnection) Ping(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.client == nil {
		return fmt.Errorf("chưa kết nối")
	}

	return c.client.Ping(ctx, readpref.Primary())
}

// HealthCheck trả về trạng thái chi tiết
func (c *MongoDBConnection) HealthCheck(ctx context.Context) HealthStatus {
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

	if err := c.client.Ping(pingCtx, readpref.Primary()); err != nil {
		status.Healthy = false
		status.State = StateError
		status.Message = err.Error()
		return status
	}

	status.Latency = time.Since(start)
	status.Healthy = true
	status.State = StateConnected
	status.Message = "OK"

	// Lấy server status nếu có thể
	status.Details["database"] = c.config.Database

	return status
}

// Info trả về thông tin connection
func (c *MongoDBConnection) Info() ConnectionInfo {
	return ConnectionInfo{
		Name:     "mongodb",
		Type:     TypeMongoDB,
		Host:     c.config.URI,
		Database: c.config.Database,
	}
}

// IsConnected kiểm tra trạng thái
func (c *MongoDBConnection) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state == StateConnected && c.client != nil
}

// Reconnect thử kết nối lại
func (c *MongoDBConnection) Reconnect(ctx context.Context) error {
	if err := c.Disconnect(ctx); err != nil {
		return err
	}
	return c.Connect(ctx)
}

// Client trả về *mongo.Client
func (c *MongoDBConnection) Client() *mongo.Client {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.client
}

// Database trả về *mongo.Database
func (c *MongoDBConnection) Database() *mongo.Database {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.database
}

// Collection trả về collection theo tên
func (c *MongoDBConnection) Collection(name string) *mongo.Collection {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.database == nil {
		return nil
	}

	return c.database.Collection(name)
}

// StartSession bắt đầu session cho transaction
func (c *MongoDBConnection) StartSession() (mongo.Session, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.client == nil {
		return nil, fmt.Errorf("chưa kết nối MongoDB")
	}

	return c.client.StartSession()
}
