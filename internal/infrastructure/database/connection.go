// Package database cung cấp interface và implementations cho các kết nối database
package database

import (
	"context"
	"time"
)

// DatabaseType định nghĩa loại database
type DatabaseType string

const (
	TypeMySQL   DatabaseType = "mysql"
	TypeMongoDB DatabaseType = "mongodb"
	TypeRedis   DatabaseType = "redis"
)

// ConnectionState trạng thái kết nối
type ConnectionState string

const (
	StateDisconnected ConnectionState = "disconnected"
	StateConnecting   ConnectionState = "connecting"
	StateConnected    ConnectionState = "connected"
	StateError        ConnectionState = "error"
)

// HealthStatus chứa thông tin health check
type HealthStatus struct {
	Healthy   bool              `json:"healthy"`
	State     ConnectionState   `json:"state"`
	Latency   time.Duration     `json:"latency_ms"`
	Message   string            `json:"message,omitempty"`
	Details   map[string]any    `json:"details,omitempty"`
	CheckedAt time.Time         `json:"checked_at"`
}

// ConnectionInfo chứa thông tin về connection
type ConnectionInfo struct {
	Name     string       `json:"name"`
	Type     DatabaseType `json:"type"`
	Host     string       `json:"host"`
	Port     int          `json:"port,omitempty"`
	Database string       `json:"database"`
}

// ConnectionOptions cấu hình chung cho connection
type ConnectionOptions struct {
	// Retry settings
	MaxRetries    int           // Số lần thử lại tối đa
	RetryInterval time.Duration // Khoảng cách giữa các lần thử

	// Timeout settings
	ConnectTimeout time.Duration // Timeout khi kết nối
	PingTimeout    time.Duration // Timeout khi ping

	// Pool settings
	MaxOpenConns    int           // Số connection tối đa
	MaxIdleConns    int           // Số idle connection tối đa
	ConnMaxLifetime time.Duration // Thời gian sống tối đa của connection
	ConnMaxIdleTime time.Duration // Thời gian idle tối đa
}

// DefaultConnectionOptions trả về options mặc định
func DefaultConnectionOptions() ConnectionOptions {
	return ConnectionOptions{
		MaxRetries:      3,
		RetryInterval:   time.Second * 2,
		ConnectTimeout:  time.Second * 10,
		PingTimeout:     time.Second * 5,
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Minute * 5,
		ConnMaxIdleTime: time.Minute * 5,
	}
}

// IConnection interface chung cho tất cả database connections
type IConnection interface {
	// Connect thiết lập kết nối đến database
	Connect(ctx context.Context) error

	// Disconnect đóng kết nối
	Disconnect(ctx context.Context) error

	// Ping kiểm tra kết nối còn hoạt động
	Ping(ctx context.Context) error

	// HealthCheck trả về trạng thái chi tiết của connection
	HealthCheck(ctx context.Context) HealthStatus

	// Info trả về thông tin connection
	Info() ConnectionInfo

	// IsConnected kiểm tra trạng thái kết nối
	IsConnected() bool

	// Reconnect thử kết nối lại
	Reconnect(ctx context.Context) error
}
