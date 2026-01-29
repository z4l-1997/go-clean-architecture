// Package migration cung cấp hệ thống migration tự động cho databases
package migration

import "context"

// IMigrator interface cho database migrator
// Mỗi database (MySQL, MongoDB) implement interface này
type IMigrator interface {
	// Migrate chạy tất cả pending migrations
	Migrate(ctx context.Context) (*MigrationResult, error)

	// Version trả về version hiện tại của database schema
	Version(ctx context.Context) (uint, bool, error)

	// Name trả về tên của migrator (ví dụ: "mysql", "mongodb")
	Name() string
}

// MigrationResult chứa kết quả sau khi chạy migration
type MigrationResult struct {
	Database       string // Tên database (mysql, mongodb)
	CurrentVersion uint   // Version hiện tại sau migration
	Dirty          bool   // Schema có bị dirty không
	Message        string // Thông báo kết quả
}
