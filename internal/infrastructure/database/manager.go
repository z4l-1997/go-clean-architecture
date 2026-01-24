package database

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DBManager quản lý tất cả database connections (Registry Pattern)
type DBManager struct {
	connections map[string]IConnection
	mu          sync.RWMutex
}

// NewDBManager tạo DBManager mới
func NewDBManager() *DBManager {
	return &DBManager{
		connections: make(map[string]IConnection),
	}
}

// Register đăng ký một connection mới
func (m *DBManager) Register(name string, conn IConnection) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.connections[name]; exists {
		return fmt.Errorf("connection '%s' đã tồn tại", name)
	}

	m.connections[name] = conn
	return nil
}

// Get lấy connection theo tên
func (m *DBManager) Get(name string) (IConnection, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	conn, exists := m.connections[name]
	if !exists {
		return nil, fmt.Errorf("connection '%s' không tồn tại", name)
	}

	return conn, nil
}

// MustGet lấy connection, panic nếu không tồn tại
func (m *DBManager) MustGet(name string) IConnection {
	conn, err := m.Get(name)
	if err != nil {
		panic(err)
	}
	return conn
}

// ConnectAll kết nối tất cả databases
func (m *DBManager) ConnectAll(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var errs []error

	for name, conn := range m.connections {
		if err := conn.Connect(ctx); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("lỗi kết nối: %v", errs)
	}

	return nil
}

// DisconnectAll ngắt kết nối tất cả databases
func (m *DBManager) DisconnectAll(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var errs []error

	for name, conn := range m.connections {
		if err := conn.Disconnect(ctx); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("lỗi ngắt kết nối: %v", errs)
	}

	return nil
}

// HealthCheckAll kiểm tra health của tất cả connections
func (m *DBManager) HealthCheckAll(ctx context.Context) map[string]HealthStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	results := make(map[string]HealthStatus)

	for name, conn := range m.connections {
		results[name] = conn.HealthCheck(ctx)
	}

	return results
}

// IsAllHealthy kiểm tra tất cả connections có healthy không
func (m *DBManager) IsAllHealthy(ctx context.Context) bool {
	statuses := m.HealthCheckAll(ctx)
	for _, status := range statuses {
		if !status.Healthy {
			return false
		}
	}
	return true
}

// Shutdown graceful shutdown tất cả connections
func (m *DBManager) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error

	// Disconnect theo thứ tự ngược lại (LIFO)
	names := make([]string, 0, len(m.connections))
	for name := range m.connections {
		names = append(names, name)
	}

	// Reverse order
	for i := len(names) - 1; i >= 0; i-- {
		name := names[i]
		conn := m.connections[name]

		// Tạo context với timeout cho mỗi connection
		disconnectCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		if err := conn.Disconnect(disconnectCtx); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", name, err))
		}
		cancel()
	}

	// Clear map
	m.connections = make(map[string]IConnection)

	if len(errs) > 0 {
		return fmt.Errorf("lỗi shutdown: %v", errs)
	}

	return nil
}

// List liệt kê tất cả connections
func (m *DBManager) List() []ConnectionInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	infos := make([]ConnectionInfo, 0, len(m.connections))
	for _, conn := range m.connections {
		infos = append(infos, conn.Info())
	}

	return infos
}

// Count trả về số lượng connections
func (m *DBManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.connections)
}
