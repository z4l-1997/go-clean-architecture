// Package entity chứa các Domain Entity
package entity

import (
	"errors"
	"time"
)

// TrangThaiOrder định nghĩa các trạng thái của đơn hàng
type TrangThaiOrder string

const (
	OrderMoi        TrangThaiOrder = "moi"         // Đơn mới tạo
	OrderDaXacNhan  TrangThaiOrder = "da_xac_nhan" // Đã xác nhận
	OrderDangNau    TrangThaiOrder = "dang_nau"    // Đang nấu
	OrderDaNau      TrangThaiOrder = "da_nau"      // Đã nấu xong
	OrderDangGiao   TrangThaiOrder = "dang_giao"   // Đang giao
	OrderHoanThanh  TrangThaiOrder = "hoan_thanh"  // Hoàn thành
	OrderDaHuy      TrangThaiOrder = "da_huy"      // Đã hủy
)

// LoaiOrder định nghĩa loại đơn hàng
type LoaiOrder string

const (
	OrderTaiCho   LoaiOrder = "tai_cho"   // Ăn tại chỗ
	OrderMangVe   LoaiOrder = "mang_ve"   // Mang về
	OrderGiaoHang LoaiOrder = "giao_hang" // Giao hàng
)

// OrderItem đại diện cho một món trong đơn hàng
type OrderItem struct {
	MonAnID   string // ID của món ăn
	TenMon    string // Tên món (snapshot tại thời điểm đặt)
	SoLuong   int    // Số lượng
	DonGia    int64  // Đơn giá tại thời điểm đặt (đã tính giảm giá)
	GhiChu    string // Ghi chú (ít cay, không hành,...)
	ThanhTien int64  // Thành tiền = SoLuong * DonGia
}

// Order là Entity đại diện cho đơn hàng
// Lưu trong MongoDB vì:
// - Schema linh hoạt (OrderItem có thể thay đổi)
// - Write-heavy (nhiều order/ngày)
// - Phù hợp embed OrderItem[] trực tiếp
// - Dễ query theo thời gian, trạng thái
type Order struct {
	ID             string         // MongoDB ObjectID hoặc UUID
	KhachHangID    string         // ID khách hàng (optional - khách vãng lai)
	NhanVienID     string         // ID nhân viên phục vụ (optional)
	DauBepID       string         // ID đầu bếp thực hiện (optional)
	SoBan          int            // Số bàn (cho order tại chỗ)
	LoaiOrder      LoaiOrder      // Loại đơn hàng
	TrangThai      TrangThaiOrder // Trạng thái hiện tại
	Items          []OrderItem    // Danh sách món
	TongTien       int64          // Tổng tiền trước giảm giá
	GiamGia        int64          // Số tiền giảm giá
	TienThanhToan  int64          // Tiền thực thanh toán
	GhiChu         string         // Ghi chú chung
	DiaChiGiao     string         // Địa chỉ giao hàng (cho delivery)
	ThoiGianDat    time.Time      // Thời gian đặt
	ThoiGianCapNhat time.Time     // Thời gian cập nhật cuối
	ThoiGianHoanThanh *time.Time  // Thời gian hoàn thành (nullable)
}

// NewOrder tạo một Order mới
func NewOrder(id string, loaiOrder LoaiOrder) (*Order, error) {
	now := time.Now()
	return &Order{
		ID:              id,
		LoaiOrder:       loaiOrder,
		TrangThai:       OrderMoi,
		Items:           make([]OrderItem, 0),
		TongTien:        0,
		GiamGia:         0,
		TienThanhToan:   0,
		ThoiGianDat:     now,
		ThoiGianCapNhat: now,
	}, nil
}

// ThemMon thêm món vào đơn hàng
func (o *Order) ThemMon(monAnID, tenMon string, soLuong int, donGia int64, ghiChu string) error {
	if soLuong <= 0 {
		return errors.New("số lượng phải lớn hơn 0")
	}
	if donGia < 0 {
		return errors.New("đơn giá không được âm")
	}

	item := OrderItem{
		MonAnID:   monAnID,
		TenMon:    tenMon,
		SoLuong:   soLuong,
		DonGia:    donGia,
		GhiChu:    ghiChu,
		ThanhTien: int64(soLuong) * donGia,
	}

	o.Items = append(o.Items, item)
	o.tinhTongTien()
	o.ThoiGianCapNhat = time.Now()

	return nil
}

// XoaMon xóa món khỏi đơn hàng theo index
func (o *Order) XoaMon(index int) error {
	if index < 0 || index >= len(o.Items) {
		return errors.New("index không hợp lệ")
	}

	o.Items = append(o.Items[:index], o.Items[index+1:]...)
	o.tinhTongTien()
	o.ThoiGianCapNhat = time.Now()

	return nil
}

// tinhTongTien tính lại tổng tiền từ các items
func (o *Order) tinhTongTien() {
	var tong int64 = 0
	for _, item := range o.Items {
		tong += item.ThanhTien
	}
	o.TongTien = tong
	o.TienThanhToan = tong - o.GiamGia
}

// ApDungGiamGia áp dụng giảm giá
func (o *Order) ApDungGiamGia(soTien int64) error {
	if soTien < 0 {
		return errors.New("số tiền giảm không được âm")
	}
	if soTien > o.TongTien {
		return errors.New("số tiền giảm không được lớn hơn tổng tiền")
	}

	o.GiamGia = soTien
	o.TienThanhToan = o.TongTien - o.GiamGia
	o.ThoiGianCapNhat = time.Now()

	return nil
}

// ChuyenTrangThai chuyển trạng thái đơn hàng
func (o *Order) ChuyenTrangThai(trangThaiMoi TrangThaiOrder) error {
	// Validate state transitions
	validTransitions := map[TrangThaiOrder][]TrangThaiOrder{
		OrderMoi:       {OrderDaXacNhan, OrderDaHuy},
		OrderDaXacNhan: {OrderDangNau, OrderDaHuy},
		OrderDangNau:   {OrderDaNau, OrderDaHuy},
		OrderDaNau:     {OrderDangGiao, OrderHoanThanh}, // Tại chỗ có thể hoàn thành luôn
		OrderDangGiao:  {OrderHoanThanh, OrderDaHuy},
	}

	allowed, exists := validTransitions[o.TrangThai]
	if !exists {
		return errors.New("không thể chuyển trạng thái từ trạng thái hiện tại")
	}

	for _, s := range allowed {
		if s == trangThaiMoi {
			o.TrangThai = trangThaiMoi
			o.ThoiGianCapNhat = time.Now()

			if trangThaiMoi == OrderHoanThanh {
				now := time.Now()
				o.ThoiGianHoanThanh = &now
			}

			return nil
		}
	}

	return errors.New("chuyển trạng thái không hợp lệ")
}

// DaHoanThanh kiểm tra order đã hoàn thành chưa
func (o *Order) DaHoanThanh() bool {
	return o.TrangThai == OrderHoanThanh
}

// DaBiHuy kiểm tra order đã bị hủy chưa
func (o *Order) DaBiHuy() bool {
	return o.TrangThai == OrderDaHuy
}

// CoTheSua kiểm tra order có thể sửa đổi không
func (o *Order) CoTheSua() bool {
	return o.TrangThai == OrderMoi || o.TrangThai == OrderDaXacNhan
}

// GanDauBep gán đầu bếp cho order
func (o *Order) GanDauBep(dauBepID string) {
	o.DauBepID = dauBepID
	o.ThoiGianCapNhat = time.Now()
}

// GanNhanVien gán nhân viên phục vụ
func (o *Order) GanNhanVien(nhanVienID string) {
	o.NhanVienID = nhanVienID
	o.ThoiGianCapNhat = time.Now()
}
