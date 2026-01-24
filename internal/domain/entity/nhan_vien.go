// Package entity chứa các Domain Entity
package entity

import (
	"errors"
	"time"
)

// ChucVu định nghĩa chức vụ của nhân viên
type ChucVu string

const (
	ChucVuBep       ChucVu = "bep"       // Đầu bếp
	ChucVuPhucVu    ChucVu = "phuc_vu"   // Phục vụ/Bồi bàn
	ChucVuThuNgan   ChucVu = "thu_ngan"  // Thu ngân
	ChucVuQuanLy    ChucVu = "quan_ly"   // Quản lý
	ChucVuGiaoHang  ChucVu = "giao_hang" // Giao hàng
)

// TrangThaiLamViec định nghĩa trạng thái làm việc
type TrangThaiLamViec string

const (
	TrangThaiRanh   TrangThaiLamViec = "ranh"    // Đang rảnh
	TrangThaiBan    TrangThaiLamViec = "ban"     // Đang bận
	TrangThaiNghi   TrangThaiLamViec = "nghi"    // Nghỉ phép
	TrangThaiOffline TrangThaiLamViec = "offline" // Không online
)

// NhanVien là Entity đại diện cho nhân viên/đầu bếp
// Lưu trong MySQL vì:
// - Dữ liệu ổn định (thông tin cá nhân, lương)
// - Cần foreign key với User
// - Cần ACID cho việc chấm công, tính lương
type NhanVien struct {
	ID          string           // UUID
	UserID      string           // FK -> User.ID
	HoTen       string           // Họ và tên
	ChucVu      ChucVu           // Chức vụ
	SoDienThoai string           // Số điện thoại
	Email       string           // Email công ty
	TrangThai   TrangThaiLamViec // Trạng thái làm việc hiện tại
	LuongCoBan  int64            // Lương cơ bản (VND/tháng)
	NgayVaoLam  time.Time        // Ngày bắt đầu làm việc
	NgayTao     time.Time        // Ngày tạo record
	NgayCapNhat time.Time        // Ngày cập nhật cuối
}

// NewNhanVien tạo một NhanVien mới
func NewNhanVien(id, userID, hoTen string, chucVu ChucVu, soDienThoai string) (*NhanVien, error) {
	if hoTen == "" {
		return nil, errors.New("họ tên không được để trống")
	}
	if soDienThoai == "" {
		return nil, errors.New("số điện thoại không được để trống")
	}

	now := time.Now()
	return &NhanVien{
		ID:          id,
		UserID:      userID,
		HoTen:       hoTen,
		ChucVu:      chucVu,
		SoDienThoai: soDienThoai,
		TrangThai:   TrangThaiOffline,
		LuongCoBan:  0,
		NgayVaoLam:  now,
		NgayTao:     now,
		NgayCapNhat: now,
	}, nil
}

// LaDauBep kiểm tra nhân viên có phải đầu bếp không
func (n *NhanVien) LaDauBep() bool {
	return n.ChucVu == ChucVuBep
}

// LaQuanLy kiểm tra nhân viên có phải quản lý không
func (n *NhanVien) LaQuanLy() bool {
	return n.ChucVu == ChucVuQuanLy
}

// DangRanh kiểm tra nhân viên có đang rảnh không
func (n *NhanVien) DangRanh() bool {
	return n.TrangThai == TrangThaiRanh
}

// CapNhatTrangThai cập nhật trạng thái làm việc
func (n *NhanVien) CapNhatTrangThai(trangThai TrangThaiLamViec) {
	n.TrangThai = trangThai
	n.NgayCapNhat = time.Now()
}

// CapNhatLuong cập nhật lương cơ bản
func (n *NhanVien) CapNhatLuong(luongMoi int64) error {
	if luongMoi < 0 {
		return errors.New("lương không được âm")
	}
	n.LuongCoBan = luongMoi
	n.NgayCapNhat = time.Now()
	return nil
}

// CoThePhanCong kiểm tra có thể phân công việc cho nhân viên không
func (n *NhanVien) CoThePhanCong() bool {
	return n.TrangThai == TrangThaiRanh
}
