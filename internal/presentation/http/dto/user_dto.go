// Package dto chứa Data Transfer Objects
package dto

import (
	"restaurant_project/internal/domain/entity"
)

// ============================================
// USER REQUEST DTOs
// ============================================

// CreateUserRequest là dữ liệu để tạo user mới
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50" example:"john_doe"`
	Email    string `json:"email" binding:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required,min=6,max=100" example:"password123"`
	Role     string `json:"role" binding:"required,oneof=admin manager staff customer" example:"staff"`
}

// UpdateUserRequest là dữ liệu để cập nhật user
type UpdateUserRequest struct {
	Email    *string `json:"email,omitempty" binding:"omitempty,email" example:"newemail@example.com"`
	IsActive *bool   `json:"is_active,omitempty" example:"true"`
}

// ChangePasswordRequest là dữ liệu để đổi mật khẩu
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required" example:"oldpass123"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=100" example:"newpass123"`
}

// ============================================
// USER RESPONSE DTOs
// ============================================

// UserResponse là dữ liệu trả về cho user
type UserResponse struct {
	ID              string `json:"id" example:"uuid-123"`
	Username        string `json:"username" example:"john_doe"`
	Email           string `json:"email" example:"john@example.com"`
	Role            string `json:"role" example:"staff"`
	IsActive        bool   `json:"is_active" example:"true"`
	IsEmailVerified bool   `json:"is_email_verified" example:"true"`
	CreatedAt       string `json:"created_at" example:"24/01/2026 10:00"`
	UpdatedAt       string `json:"updated_at" example:"24/01/2026 10:30"`
}

// ToUserResponse chuyển đổi Entity sang Response DTO
func ToUserResponse(user *entity.User) UserResponse {
	return UserResponse{
		ID:              user.ID,
		Username:        user.Username,
		Email:           user.Email,
		Role:            string(user.Role),
		IsActive:        user.IsActive,
		IsEmailVerified: user.IsEmailVerified,
		CreatedAt:       user.NgayTao.Format("02/01/2006 15:04"),
		UpdatedAt:       user.NgayCapNhat.Format("02/01/2006 15:04"),
	}
}

// ToUserResponseList chuyển đổi danh sách Entity sang Response DTO
func ToUserResponseList(users []*entity.User) []UserResponse {
	result := make([]UserResponse, len(users))
	for i, user := range users {
		result[i] = ToUserResponse(user)
	}
	return result
}
