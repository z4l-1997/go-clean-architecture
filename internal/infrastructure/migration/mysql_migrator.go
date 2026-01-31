package migration

import (
	"context"
	"database/sql"
	"io/fs"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// MySQLMigrator thực hiện migration cho MySQL sử dụng golang-migrate
type MySQLMigrator struct {
	db       *sql.DB
	embedFS  fs.FS
	dbName   string
}

// NewMySQLMigrator tạo MySQLMigrator mới
// db: *sql.DB connection đã kết nối
// embedFS: embed.FS chứa SQL migration files
// dbName: tên database (dùng cho logging)
func NewMySQLMigrator(db *sql.DB, embedFS fs.FS, dbName string) *MySQLMigrator {
	return &MySQLMigrator{
		db:      db,
		embedFS: embedFS,
		dbName:  dbName,
	}
}

// Name trả về tên migrator
func (m *MySQLMigrator) Name() string {
	return "mysql"
}

// Migrate chạy tất cả pending migrations
func (m *MySQLMigrator) Migrate(ctx context.Context) (*MigrationResult, error) {
	mig, err := m.createMigrateInstance()
	if err != nil {
		return nil, err
	}
	// Không gọi mig.Close() vì nó sẽ đóng luôn *sql.DB đang được share
	// với toàn bộ ứng dụng. DB connection được quản lý bởi DBManager.

	err = mig.Up()
	if err != nil && err != migrate.ErrNoChange {
		return nil, err
	}

	version, dirty, _ := mig.Version()

	message := "migrations applied successfully"
	if err == migrate.ErrNoChange {
		message = "no changes needed"
	}

	return &MigrationResult{
		Database:       "mysql",
		CurrentVersion: version,
		Dirty:          dirty,
		Message:        message,
	}, nil
}

// Version trả về version hiện tại
func (m *MySQLMigrator) Version(ctx context.Context) (uint, bool, error) {
	mig, err := m.createMigrateInstance()
	if err != nil {
		return 0, false, err
	}

	return mig.Version()
}

// createMigrateInstance tạo golang-migrate instance
func (m *MySQLMigrator) createMigrateInstance() (*migrate.Migrate, error) {
	// Tạo source driver từ embed.FS
	// Subdirectory "mysql" vì embed pattern là mysql/*.sql
	sourceDriver, err := iofs.New(m.embedFS, "mysql")
	if err != nil {
		return nil, err
	}

	// Tạo database driver
	dbDriver, err := mysql.WithInstance(m.db, &mysql.Config{
		DatabaseName: m.dbName,
	})
	if err != nil {
		return nil, err
	}

	// Tạo migrate instance
	return migrate.NewWithInstance("iofs", sourceDriver, m.dbName, dbDriver)
}
