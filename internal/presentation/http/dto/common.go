// Package dto chứa Data Transfer Objects
package dto

// ============================================
// PAGINATION DTOs
// ============================================

// PaginationRequest là query params cho pagination
type PaginationRequest struct {
	Page  int `form:"page,default=1" binding:"min=1"`
	Limit int `form:"limit,default=20" binding:"min=1,max=100"`
}

// Offset tính offset từ page và limit
func (p PaginationRequest) Offset() int {
	return (p.Page - 1) * p.Limit
}

// PaginatedResponse là response chứa dữ liệu phân trang
type PaginatedResponse struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int         `json:"total_pages"`
}

// NewPaginatedResponse tạo PaginatedResponse
func NewPaginatedResponse(items interface{}, total int64, page, limit int) PaginatedResponse {
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}
	return PaginatedResponse{
		Items:      items,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}
}
