// Package repository định nghĩa các Interface cho việc lưu trữ dữ liệu
package repository

import (
	"context"
	"time"

	"restaurant_project/internal/domain/entity"
)

// IOrderRepository là interface định nghĩa các thao tác với dữ liệu Order
// Implementation: MongoDB (schema linh hoạt, write-heavy, embedded documents)
type IOrderRepository interface {
	// FindByID tìm order theo ID
	FindByID(ctx context.Context, id string) (*entity.Order, error)

	// FindAll lấy tất cả orders
	FindAll(ctx context.Context) ([]*entity.Order, error)

	// FindByKhachHangID lấy orders của một khách hàng
	FindByKhachHangID(ctx context.Context, khachHangID string) ([]*entity.Order, error)

	// FindByTrangThai lấy orders theo trạng thái
	FindByTrangThai(ctx context.Context, trangThai entity.TrangThaiOrder) ([]*entity.Order, error)

	// FindByDauBepID lấy orders được gán cho một đầu bếp
	FindByDauBepID(ctx context.Context, dauBepID string) ([]*entity.Order, error)

	// FindByThoiGian lấy orders trong khoảng thời gian
	FindByThoiGian(ctx context.Context, from, to time.Time) ([]*entity.Order, error)

	// FindPending lấy các orders đang chờ xử lý (mới, đã xác nhận, đang nấu)
	FindPending(ctx context.Context) ([]*entity.Order, error)

	// Save lưu order mới hoặc cập nhật
	Save(ctx context.Context, order *entity.Order) error

	// Delete xóa order theo ID
	Delete(ctx context.Context, id string) error

	// UpdateTrangThai cập nhật trạng thái order
	UpdateTrangThai(ctx context.Context, id string, trangThai entity.TrangThaiOrder) error

	// Count đếm tổng số orders
	Count(ctx context.Context) (int64, error)

	// CountByTrangThai đếm orders theo trạng thái
	CountByTrangThai(ctx context.Context, trangThai entity.TrangThaiOrder) (int64, error)

	// TinhDoanhThu tính doanh thu trong khoảng thời gian
	TinhDoanhThu(ctx context.Context, from, to time.Time) (int64, error)
}
