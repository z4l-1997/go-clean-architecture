package migration

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"

	"restaurant_project/pkg/logger"
)

// MigrationManager quản lý tất cả migrators (Registry Pattern giống DBManager)
type MigrationManager struct {
	migrators map[string]IMigrator
	mu        sync.RWMutex
}

// NewMigrationManager tạo MigrationManager mới
func NewMigrationManager() *MigrationManager {
	return &MigrationManager{
		migrators: make(map[string]IMigrator),
	}
}

// Register đăng ký một migrator mới
func (m *MigrationManager) Register(name string, migrator IMigrator) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.migrators[name]; exists {
		return fmt.Errorf("migrator '%s' đã tồn tại", name)
	}

	m.migrators[name] = migrator
	return nil
}

// MigrateAll chạy migration cho tất cả databases đã đăng ký
func (m *MigrationManager) MigrateAll(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	logger.Info("Migration: bắt đầu migrate tất cả databases",
		zap.Int("total_migrators", len(m.migrators)),
	)

	var errs []error

	for name, migrator := range m.migrators {
		logger.Info("Migration: đang chạy",
			zap.String("database", name),
		)

		result, err := migrator.Migrate(ctx)
		if err != nil {
			logger.Error("Migration: lỗi",
				zap.String("database", name),
				zap.Error(err),
			)
			errs = append(errs, fmt.Errorf("%s: %w", name, err))
			continue
		}

		logger.Info("Migration: hoàn tất",
			zap.String("database", result.Database),
			zap.Uint("version", result.CurrentVersion),
			zap.Bool("dirty", result.Dirty),
			zap.String("message", result.Message),
		)
	}

	if len(errs) > 0 {
		return fmt.Errorf("migration errors: %v", errs)
	}

	logger.Info("Migration: tất cả databases đã được migrate thành công")
	return nil
}
