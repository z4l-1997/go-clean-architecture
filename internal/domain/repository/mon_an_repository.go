// Package repository định nghĩa các Interface cho việc lưu trữ dữ liệu
// Interface là "ổ cắm" - định nghĩa WHAT (làm gì) chứ không phải HOW (làm như thế nào)
package repository

import (
	"context"

	"restaurant_project/internal/domain/entity"
)

// IMonAnRepository là interface định nghĩa các thao tác với dữ liệu MonAn
//
// TẠI SAO DÙNG INTERFACE?
// - Giống như ổ cắm điện: bạn biết ổ cắm có 2 lỗ, nhưng không cần biết dây điện
//   đi đâu (từ nhà máy thủy điện hay nhiệt điện)
// - UseCase chỉ cần biết "tôi có thể lưu/lấy món ăn", không cần biết
//   lưu vào MySQL, MongoDB, hay file text
//
// DEPENDENCY INVERSION PRINCIPLE:
// - Domain layer định nghĩa interface
// - Infrastructure layer implement interface
// - UseCase phụ thuộc vào interface, KHÔNG phụ thuộc vào implementation cụ thể
type IMonAnRepository interface {
	// FindByID tìm món ăn theo ID
	// Trả về nil nếu không tìm thấy
	FindByID(ctx context.Context, id string) (*entity.MonAn, error)

	// FindAll lấy tất cả món ăn
	FindAll(ctx context.Context) ([]*entity.MonAn, error)

	// FindByConHang lấy các món theo trạng thái còn hàng
	// conHang = true → lấy các món còn bán
	// conHang = false → lấy các món hết hàng
	FindByConHang(ctx context.Context, conHang bool) ([]*entity.MonAn, error)

	// Save lưu món ăn mới hoặc cập nhật món đã có
	// Nếu ID đã tồn tại → update
	// Nếu ID chưa tồn tại → insert
	Save(ctx context.Context, mon *entity.MonAn) error

	// Delete xóa món ăn theo ID
	// Trả về error nếu không tìm thấy món
	Delete(ctx context.Context, id string) error

	// Count đếm tổng số món ăn
	Count(ctx context.Context) (int64, error)
}
