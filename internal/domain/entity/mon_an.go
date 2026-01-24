// Package entity chứa các Domain Entity - "trái tim" của hệ thống
// Entity là đối tượng có danh tính (identity) và chứa business logic cốt lõi
package entity

import (
	"errors"
	"time"
)

// MonAn là Entity đại diện cho một món ăn trong menu
// Đây giống như "công thức phở" - quy tắc kinh doanh không thay đổi
// dù bạn đổi database (MySQL → MongoDB) hay đổi framework
type MonAn struct {
	ID          string    // Mã định danh duy nhất
	Ten         string    // Tên món ăn (VD: "Phở tái")
	Gia         int64     // Giá gốc (đơn vị: VND)
	MoTa        string    // Mô tả món ăn
	ConHang     bool      // Còn bán không?
	GiamGia     int       // Phần trăm giảm giá (0-100)
	NgayTao     time.Time // Ngày tạo món
	NgayCapNhat time.Time // Ngày cập nhật cuối
}

// NewMonAn tạo một MonAn mới với validation
// Đây là "constructor" - đảm bảo Entity luôn ở trạng thái hợp lệ
func NewMonAn(id, ten string, gia int64, moTa string) (*MonAn, error) {
	// Business rule: Tên không được rỗng
	if ten == "" {
		return nil, errors.New("tên món ăn không được để trống")
	}

	// Business rule: Giá phải > 0
	if gia <= 0 {
		return nil, errors.New("giá món ăn phải lớn hơn 0")
	}

	now := time.Now()
	return &MonAn{
		ID:          id,
		Ten:         ten,
		Gia:         gia,
		MoTa:        moTa,
		ConHang:     true, // Mặc định còn hàng
		GiamGia:     0,    // Mặc định không giảm giá
		NgayTao:     now,
		NgayCapNhat: now,
	}, nil
}

// TinhGia tính giá sau khi áp dụng giảm giá
// Business logic: Giá cuối = Giá gốc - (Giá gốc * GiảmGiá / 100)
// VD: Phở 50,000đ giảm 10% → 45,000đ
func (m *MonAn) TinhGia() int64 {
	if m.GiamGia <= 0 {
		return m.Gia
	}

	// Tính số tiền giảm
	soTienGiam := m.Gia * int64(m.GiamGia) / 100
	return m.Gia - soTienGiam
}

// CoTheBan kiểm tra món có thể bán được không
// Business rule: Phải còn hàng VÀ giá > 0
func (m *MonAn) CoTheBan() bool {
	return m.ConHang && m.Gia > 0
}

// ApDungGiamGia áp dụng giảm giá cho món
// Business rule: Giảm giá phải trong khoảng 0-100%
func (m *MonAn) ApDungGiamGia(phanTram int) error {
	if phanTram < 0 || phanTram > 100 {
		return errors.New("phần trăm giảm giá phải từ 0 đến 100")
	}

	m.GiamGia = phanTram
	m.NgayCapNhat = time.Now()
	return nil
}

// HetHang đánh dấu món hết hàng
func (m *MonAn) HetHang() {
	m.ConHang = false
	m.NgayCapNhat = time.Now()
}

// CoHang đánh dấu món có hàng trở lại
func (m *MonAn) CoHang() {
	m.ConHang = true
	m.NgayCapNhat = time.Now()
}

// CapNhatGia cập nhật giá mới cho món
func (m *MonAn) CapNhatGia(giaMoi int64) error {
	if giaMoi <= 0 {
		return errors.New("giá món ăn phải lớn hơn 0")
	}

	m.Gia = giaMoi
	m.NgayCapNhat = time.Now()
	return nil
}
