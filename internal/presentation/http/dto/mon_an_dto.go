// Package dto chứa Data Transfer Objects - "phiếu ghi order" của hệ thống
// DTO chuyển đổi dữ liệu giữa client và hệ thống
package dto

import (
	"restaurant_project/internal/domain/entity"
)

// ============================================
// REQUEST DTOs - Dữ liệu từ client gửi lên
// ============================================

// ThemMonRequest là dữ liệu client gửi khi thêm món mới
// JSON tags định nghĩa tên field khi parse JSON
type ThemMonRequest struct {
	Ten  string `json:"ten" example:"Phở bò tái"`       // Tên món ăn
	Gia  int64  `json:"gia" example:"50000"`            // Giá món (VND)
	MoTa string `json:"mo_ta" example:"Phở truyền thống với thịt bò tái"` // Mô tả món ăn
}

// CapNhatGiaRequest là dữ liệu client gửi khi cập nhật giá
type CapNhatGiaRequest struct {
	Gia int64 `json:"gia" example:"55000"` // Giá mới (VND)
}

// ApDungGiamGiaRequest là dữ liệu client gửi khi áp dụng giảm giá
type ApDungGiamGiaRequest struct {
	PhanTram int `json:"phan_tram" example:"10"` // Phần trăm giảm giá (0-100)
}

// ============================================
// RESPONSE DTOs - Dữ liệu trả về cho client
// ============================================

// MonAnResponse là dữ liệu trả về khi client yêu cầu thông tin món ăn
// TẠI SAO CẦN DTO RIÊNG THAY VÌ DÙNG ENTITY?
// 1. Entity có thể chứa thông tin nhạy cảm không nên expose
// 2. Client có thể cần format khác (VD: giá đã format, ngày theo định dạng khác)
// 3. Tách biệt API contract với internal data model
type MonAnResponse struct {
	ID          string `json:"id" example:"1_mon"`                    // Mã món
	Ten         string `json:"ten" example:"Phở bò tái"`              // Tên món
	Gia         int64  `json:"gia" example:"50000"`                   // Giá gốc
	GiaSauGiam  int64  `json:"gia_sau_giam" example:"45000"`          // Giá sau giảm giá
	MoTa        string `json:"mo_ta" example:"Phở truyền thống"`      // Mô tả
	ConHang     bool   `json:"con_hang" example:"true"`               // Còn bán không
	GiamGia     int    `json:"giam_gia" example:"10"`                 // % giảm giá
	CoTheBan    bool   `json:"co_the_ban" example:"true"`             // Có thể bán không (business logic)
	NgayTao     string `json:"ngay_tao" example:"24/01/2026 10:00"`   // Ngày tạo (format đẹp)
	NgayCapNhat string `json:"ngay_cap_nhat" example:"24/01/2026 10:30"` // Ngày cập nhật
}

// ToMonAnResponse chuyển đổi Entity sang Response DTO
// Đây là pattern Mapper - chuyển đổi giữa các layer
func ToMonAnResponse(mon *entity.MonAn) MonAnResponse {
	return MonAnResponse{
		ID:          mon.ID,
		Ten:         mon.Ten,
		Gia:         mon.Gia,
		GiaSauGiam:  mon.TinhGia(),   // Gọi business logic của Entity
		MoTa:        mon.MoTa,
		ConHang:     mon.ConHang,
		GiamGia:     mon.GiamGia,
		CoTheBan:    mon.CoTheBan(),  // Gọi business logic của Entity
		NgayTao:     mon.NgayTao.Format("02/01/2006 15:04"),
		NgayCapNhat: mon.NgayCapNhat.Format("02/01/2006 15:04"),
	}
}

// ToMonAnResponseList chuyển đổi danh sách Entity sang Response DTO
func ToMonAnResponseList(monList []*entity.MonAn) []MonAnResponse {
	result := make([]MonAnResponse, len(monList))
	for i, mon := range monList {
		result[i] = ToMonAnResponse(mon)
	}
	return result
}

// ============================================
// COMMON RESPONSE - Response dùng chung
// ============================================

// APIResponse là cấu trúc response chuẩn cho tất cả API
type APIResponse struct {
	Success bool        `json:"success" example:"true"`           // Request thành công hay thất bại
	Message string      `json:"message" example:"Thao tác thành công"` // Thông báo
	Data    interface{} `json:"data,omitempty"`                   // Dữ liệu trả về (nếu có)
	Error   string      `json:"error,omitempty" example:""`       // Lỗi (nếu có)
}

// NewSuccessResponse tạo response thành công
func NewSuccessResponse(message string, data interface{}) APIResponse {
	return APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// NewErrorResponse tạo response lỗi
func NewErrorResponse(message string, err error) APIResponse {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	return APIResponse{
		Success: false,
		Message: message,
		Error:   errMsg,
	}
}
