package providers

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"

	"restaurant_project/internal/infrastructure/config"
	"restaurant_project/internal/infrastructure/migration"
	"restaurant_project/internal/infrastructure/migrations"
	"restaurant_project/pkg/logger"
)

// ProvideMySQLMigrator tạo MySQLMigrator với embedded SQL files
func ProvideMySQLMigrator(db *sql.DB, cfg *config.MySQLConfig) *migration.MySQLMigrator {
	return migration.NewMySQLMigrator(db, migrations.MySQLMigrationFiles, cfg.Database)
}

// ProvideMigrationManager tạo MigrationManager, đăng ký migrators và chạy migration
func ProvideMigrationManager(
	cfg *config.Config,
	mysqlMigrator *migration.MySQLMigrator,
) (*migration.MigrationManager, error) {
	manager := migration.NewMigrationManager()

	// Đăng ký migrators
	manager.Register("mysql", mysqlMigrator)

	// Chạy migration nếu auto-migrate được bật
	if cfg.Migration.AutoMigrate {
		logger.Info("Migration: auto-migrate enabled, đang chạy migrations...",
			zap.Bool("auto_migrate", cfg.Migration.AutoMigrate),
		)

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		if err := manager.MigrateAll(ctx); err != nil {
			return nil, fmt.Errorf("auto-migration failed: %w", err)
		}
	} else {
		logger.Info("Migration: auto-migrate disabled, bỏ qua migrations")
	}

	return manager, nil
}
