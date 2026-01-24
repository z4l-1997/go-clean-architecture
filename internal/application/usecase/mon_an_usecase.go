// Package usecase chứa Application Use Cases - "quản lý bếp" của hệ thống
// UseCase điều phối luồng nghiệp vụ: nhận input → gọi domain → gọi repository → trả output
package usecase

import (
	"context"
	"errors"
	"fmt"

	"restaurant_project/internal/domain/entity"
	"restaurant_project/internal/domain/repository"
)

// MonAnUseCase xử lý các use case liên quan đến MonAn
//
// VAI TRÒ CỦA USECASE:
// - Giống như "quản lý bếp" trong nhà hàng
// - Nhận order (input) từ bồi bàn (Handler)
// - Điều phối đầu bếp (Domain Entity) và kho (Repository)
// - Không chứa business logic cốt lõi (đã có trong Entity)
// - Chỉ điều phối workflow
type MonAnUseCase struct {
	repo repository.IMonAnRepository
}

// NewMonAnUseCase tạo mới MonAnUseCase với dependency injection
// DEPENDENCY INJECTION:
// - UseCase KHÔNG tự tạo repository
// - Repository được "inject" (tiêm) từ bên ngoài
// - Điều này giúp dễ dàng thay đổi implementation (VD: từ Memory → MySQL)
func NewMonAnUseCase(repo repository.IMonAnRepository) *MonAnUseCase {
	return &MonAnUseCase{
		repo: repo,
	}
}

// ThemMonInput là dữ liệu đầu vào để thêm món mới
type ThemMonInput struct {
	ID    string
	Ten   string
	Gia   int64
	MoTa  string
}

// ThemMon thêm một món mới vào menu
// Workflow:
// 1. Validate input (trong Entity)
// 2. Tạo Entity mới
// 3. Lưu vào repository
func (uc *MonAnUseCase) ThemMon(ctx context.Context, input ThemMonInput) (*entity.MonAn, error) {
	// Bước 1: Tạo Entity (validation xảy ra ở đây)
	mon, err := entity.NewMonAn(input.ID, input.Ten, input.Gia, input.MoTa)
	if err != nil {
		return nil, fmt.Errorf("không thể tạo món ăn: %w", err)
	}

	// Bước 2: Lưu vào repository
	if err := uc.repo.Save(ctx, mon); err != nil {
		return nil, fmt.Errorf("không thể lưu món ăn: %w", err)
	}

	return mon, nil
}

// XemMenu lấy danh sách tất cả món ăn
func (uc *MonAnUseCase) XemMenu(ctx context.Context) ([]*entity.MonAn, error) {
	menu, err := uc.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("không thể lấy menu: %w", err)
	}

	return menu, nil
}

// XemMenuConHang lấy danh sách món còn bán
func (uc *MonAnUseCase) XemMenuConHang(ctx context.Context) ([]*entity.MonAn, error) {
	menu, err := uc.repo.FindByConHang(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("không thể lấy menu còn hàng: %w", err)
	}

	return menu, nil
}

// TimMon tìm món theo ID
func (uc *MonAnUseCase) TimMon(ctx context.Context, id string) (*entity.MonAn, error) {
	if id == "" {
		return nil, errors.New("ID không được để trống")
	}

	mon, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("không thể tìm món: %w", err)
	}

	if mon == nil {
		return nil, fmt.Errorf("không tìm thấy món với ID: %s", id)
	}

	return mon, nil
}

// CapNhatGiaInput là dữ liệu đầu vào để cập nhật giá
type CapNhatGiaInput struct {
	ID      string
	GiaMoi  int64
}

// CapNhatGia cập nhật giá cho món ăn
func (uc *MonAnUseCase) CapNhatGia(ctx context.Context, input CapNhatGiaInput) (*entity.MonAn, error) {
	// Bước 1: Tìm món
	mon, err := uc.TimMon(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	// Bước 2: Cập nhật giá (business logic trong Entity)
	if err := mon.CapNhatGia(input.GiaMoi); err != nil {
		return nil, fmt.Errorf("không thể cập nhật giá: %w", err)
	}

	// Bước 3: Lưu lại
	if err := uc.repo.Save(ctx, mon); err != nil {
		return nil, fmt.Errorf("không thể lưu món ăn: %w", err)
	}

	return mon, nil
}

// ApDungGiamGiaInput là dữ liệu đầu vào để áp dụng giảm giá
type ApDungGiamGiaInput struct {
	ID        string
	PhanTram  int
}

// ApDungGiamGia áp dụng giảm giá cho món ăn
func (uc *MonAnUseCase) ApDungGiamGia(ctx context.Context, input ApDungGiamGiaInput) (*entity.MonAn, error) {
	// Bước 1: Tìm món
	mon, err := uc.TimMon(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	// Bước 2: Áp dụng giảm giá (business logic trong Entity)
	if err := mon.ApDungGiamGia(input.PhanTram); err != nil {
		return nil, fmt.Errorf("không thể áp dụng giảm giá: %w", err)
	}

	// Bước 3: Lưu lại
	if err := uc.repo.Save(ctx, mon); err != nil {
		return nil, fmt.Errorf("không thể lưu món ăn: %w", err)
	}

	return mon, nil
}

// XoaMon xóa món khỏi menu
func (uc *MonAnUseCase) XoaMon(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("ID không được để trống")
	}

	if err := uc.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("không thể xóa món: %w", err)
	}

	return nil
}

// DanhDauHetHang đánh dấu món hết hàng
func (uc *MonAnUseCase) DanhDauHetHang(ctx context.Context, id string) (*entity.MonAn, error) {
	// Bước 1: Tìm món
	mon, err := uc.TimMon(ctx, id)
	if err != nil {
		return nil, err
	}

	// Bước 2: Đánh dấu hết hàng
	mon.HetHang()

	// Bước 3: Lưu lại
	if err := uc.repo.Save(ctx, mon); err != nil {
		return nil, fmt.Errorf("không thể lưu món ăn: %w", err)
	}

	return mon, nil
}
