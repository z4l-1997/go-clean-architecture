// Package entity chứa các Domain Entity
package entity

import (
	"errors"
	"time"
)

// KhachHang là Entity đại diện cho khách hàng
// Lưu trong MySQL vì:
// - Dữ liệu ổn định, cần tích điểm thành viên
// - Cần foreign key với User
// - Cần ACID cho transactions liên quan đến điểm thưởng
type KhachHang struct {
	ID           string    // UUID
	UserID       string    // FK -> User.ID (nullable - khách vãng lai)
	HoTen        string    // Họ và tên
	SoDienThoai  string    // Số điện thoại (unique)
	Email        string    // Email (optional)
	DiaChi       string    // Địa chỉ giao hàng
	DiemTichLuy  int64     // Điểm tích lũy (loyalty points)
	CapThanhVien string    // Cấp: "bronze", "silver", "gold", "platinum"
	NgayTao      time.Time // Ngày đăng ký
	NgayCapNhat  time.Time // Ngày cập nhật cuối
}

// NewKhachHang tạo một KhachHang mới
func NewKhachHang(id, hoTen, soDienThoai string) (*KhachHang, error) {
	if hoTen == "" {
		return nil, errors.New("họ tên không được để trống")
	}
	if soDienThoai == "" {
		return nil, errors.New("số điện thoại không được để trống")
	}

	now := time.Now()
	return &KhachHang{
		ID:           id,
		HoTen:        hoTen,
		SoDienThoai:  soDienThoai,
		DiemTichLuy:  0,
		CapThanhVien: "bronze",
		NgayTao:      now,
		NgayCapNhat:  now,
	}, nil
}

// ThemDiem thêm điểm tích lũy và tự động nâng cấp
func (k *KhachHang) ThemDiem(diem int64) {
	k.DiemTichLuy += diem
	k.capNhatCapThanhVien()
	k.NgayCapNhat = time.Now()
}

// TruDiem trừ điểm khi đổi quà
func (k *KhachHang) TruDiem(diem int64) error {
	if diem > k.DiemTichLuy {
		return errors.New("không đủ điểm tích lũy")
	}
	k.DiemTichLuy -= diem
	k.NgayCapNhat = time.Now()
	return nil
}

// capNhatCapThanhVien cập nhật cấp dựa trên điểm tích lũy
func (k *KhachHang) capNhatCapThanhVien() {
	switch {
	case k.DiemTichLuy >= 10000:
		k.CapThanhVien = "platinum"
	case k.DiemTichLuy >= 5000:
		k.CapThanhVien = "gold"
	case k.DiemTichLuy >= 2000:
		k.CapThanhVien = "silver"
	default:
		k.CapThanhVien = "bronze"
	}
}

// TinhGiamGiaThanhVien tính % giảm giá dựa trên cấp thành viên
func (k *KhachHang) TinhGiamGiaThanhVien() int {
	switch k.CapThanhVien {
	case "platinum":
		return 15
	case "gold":
		return 10
	case "silver":
		return 5
	default:
		return 0
	}
}

// LinkToUser liên kết khách hàng với tài khoản user
func (k *KhachHang) LinkToUser(userID string) {
	k.UserID = userID
	k.NgayCapNhat = time.Now()
}
