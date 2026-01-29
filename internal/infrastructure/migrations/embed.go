// Package migrations chứa embedded migration files
package migrations

import "embed"

// MySQLMigrationFiles chứa tất cả SQL migration files cho MySQL
// golang-migrate sử dụng format: {version}_{description}.{up|down}.sql
//
//go:embed mysql/*.sql
var MySQLMigrationFiles embed.FS
