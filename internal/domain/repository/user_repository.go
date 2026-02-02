// Package repository định nghĩa các Interface cho việc lưu trữ dữ liệu
package repository

import (
	"context"
	"errors"

	"restaurant_project/internal/domain/entity"
)

// Repository-level errors
var (
	// ErrDuplicateEntry là lỗi khi INSERT vi phạm UNIQUE constraint
	ErrDuplicateEntry = errors.New("duplicate entry")
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

	// FindAllPaginated lấy users có phân trang
	FindAllPaginated(ctx context.Context, offset, limit int) ([]*entity.User, int64, error)

	// FindByRole lấy users theo role
	FindByRole(ctx context.Context, role entity.UserRole) ([]*entity.User, error)

	// FindByRolePaginated lấy users theo role có phân trang
	FindByRolePaginated(ctx context.Context, role entity.UserRole, offset, limit int) ([]*entity.User, int64, error)

	// Create tạo user mới (trả lỗi nếu trùng unique constraint)
	Create(ctx context.Context, user *entity.User) error

	// Save cập nhật user đã tồn tại
	Save(ctx context.Context, user *entity.User) error

	// Delete xóa user theo ID
	Delete(ctx context.Context, id string) error

	// ExistsByUsername kiểm tra username đã tồn tại chưa
	ExistsByUsername(ctx context.Context, username string) (bool, error)

	// ExistsByEmail kiểm tra email đã tồn tại chưa
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}
