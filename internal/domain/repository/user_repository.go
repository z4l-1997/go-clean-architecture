// Package repository định nghĩa các Interface cho việc lưu trữ dữ liệu
package repository

import (
	"context"

	"restaurant_project/internal/domain/entity"
)

// IUserRepository là interface định nghĩa các thao tác với dữ liệu User
// Implementation: MySQL (dữ liệu ổn định, cần ACID)
type IUserRepository interface {
	// FindByID tìm user theo ID
	FindByID(ctx context.Context, id string) (*entity.User, error)

	// FindByUsername tìm user theo username
	FindByUsername(ctx context.Context, username string) (*entity.User, error)

	// FindByEmail tìm user theo email
	FindByEmail(ctx context.Context, email string) (*entity.User, error)

	// FindAll lấy tất cả users
	FindAll(ctx context.Context) ([]*entity.User, error)

	// FindByRole lấy users theo role
	FindByRole(ctx context.Context, role entity.UserRole) ([]*entity.User, error)

	// Save lưu user mới hoặc cập nhật
	Save(ctx context.Context, user *entity.User) error

	// Delete xóa user theo ID
	Delete(ctx context.Context, id string) error

	// ExistsByUsername kiểm tra username đã tồn tại chưa
	ExistsByUsername(ctx context.Context, username string) (bool, error)

	// ExistsByEmail kiểm tra email đã tồn tại chưa
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}
