// Package repository định nghĩa các Interface cho việc lưu trữ dữ liệu
package repository

import (
	"context"

	"restaurant_project/internal/domain/entity"
)

// INhanVienRepository là interface định nghĩa các thao tác với dữ liệu NhanVien
// Implementation: MySQL (dữ liệu ổn định, foreign key với User)
type INhanVienRepository interface {
	// FindByID tìm nhân viên theo ID
	FindByID(ctx context.Context, id string) (*entity.NhanVien, error)

	// FindByUserID tìm nhân viên theo UserID
	FindByUserID(ctx context.Context, userID string) (*entity.NhanVien, error)

	// FindAll lấy tất cả nhân viên
	FindAll(ctx context.Context) ([]*entity.NhanVien, error)

	// FindByChucVu lấy nhân viên theo chức vụ
	FindByChucVu(ctx context.Context, chucVu entity.ChucVu) ([]*entity.NhanVien, error)

	// FindByTrangThai lấy nhân viên theo trạng thái làm việc
	FindByTrangThai(ctx context.Context, trangThai entity.TrangThaiLamViec) ([]*entity.NhanVien, error)

	// FindDauBepRanh tìm đầu bếp đang rảnh (để phân công order)
	FindDauBepRanh(ctx context.Context) ([]*entity.NhanVien, error)

	// Save lưu nhân viên mới hoặc cập nhật
	Save(ctx context.Context, nhanVien *entity.NhanVien) error

	// Delete xóa nhân viên theo ID
	Delete(ctx context.Context, id string) error

	// UpdateTrangThai cập nhật trạng thái làm việc
	UpdateTrangThai(ctx context.Context, id string, trangThai entity.TrangThaiLamViec) error

	// Count đếm tổng số nhân viên
	Count(ctx context.Context) (int64, error)
}
