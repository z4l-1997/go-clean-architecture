// Package repository định nghĩa các Interface cho việc lưu trữ dữ liệu
package repository

import (
	"context"

	"restaurant_project/internal/domain/entity"
)

// IKhachHangRepository là interface định nghĩa các thao tác với dữ liệu KhachHang
// Implementation: MySQL (cần ACID cho loyalty points, foreign key với User)
type IKhachHangRepository interface {
	// FindByID tìm khách hàng theo ID
	FindByID(ctx context.Context, id string) (*entity.KhachHang, error)

	// FindByUserID tìm khách hàng theo UserID
	FindByUserID(ctx context.Context, userID string) (*entity.KhachHang, error)

	// FindBySoDienThoai tìm khách hàng theo số điện thoại
	FindBySoDienThoai(ctx context.Context, soDienThoai string) (*entity.KhachHang, error)

	// FindAll lấy tất cả khách hàng
	FindAll(ctx context.Context) ([]*entity.KhachHang, error)

	// FindByCapThanhVien lấy khách hàng theo cấp thành viên
	FindByCapThanhVien(ctx context.Context, cap string) ([]*entity.KhachHang, error)

	// Save lưu khách hàng mới hoặc cập nhật
	Save(ctx context.Context, khachHang *entity.KhachHang) error

	// Delete xóa khách hàng theo ID
	Delete(ctx context.Context, id string) error

	// UpdateDiemTichLuy cập nhật điểm tích lũy (atomic operation)
	UpdateDiemTichLuy(ctx context.Context, id string, diemMoi int64) error

	// Count đếm tổng số khách hàng
	Count(ctx context.Context) (int64, error)
}
