// Package entity chứa các Domain Entity
package entity

import (
	"errors"
	"time"
)

// UserRole định nghĩa vai trò của user trong hệ thống
type UserRole string

const (
	RoleAdmin    UserRole = "admin"     // Quản trị viên
	RoleManager  UserRole = "manager"   // Quản lý
	RoleStaff    UserRole = "staff"     // Nhân viên
	RoleCustomer UserRole = "customer"  // Khách hàng
)

// User là Entity đại diện cho tài khoản người dùng
// Lưu trong MySQL vì:
// - Dữ liệu ổn định, ít thay đổi schema
// - Cần ACID transactions cho authentication
// - Cần foreign key relationships với KhachHang, NhanVien
type User struct {
	ID              string    // UUID hoặc auto-increment ID
	Username        string    // Tên đăng nhập (unique)
	Email           string    // Email (unique)
	PasswordHash    string    // Mật khẩu đã hash (bcrypt)
	Role            UserRole  // Vai trò trong hệ thống
	IsActive        bool      // Tài khoản có hoạt động không
	IsEmailVerified bool      // Email đã được xác thực chưa
	NgayTao         time.Time // Ngày tạo tài khoản
	NgayCapNhat     time.Time // Ngày cập nhật cuối
}

// NewUser tạo một User mới với validation
func NewUser(id, username, email, passwordHash string, role UserRole) (*User, error) {
	if username == "" {
		return nil, errors.New("username không được để trống")
	}
	if email == "" {
		return nil, errors.New("email không được để trống")
	}
	if passwordHash == "" {
		return nil, errors.New("password không được để trống")
	}

	now := time.Now()
	return &User{
		ID:              id,
		Username:        username,
		Email:           email,
		PasswordHash:    passwordHash,
		Role:            role,
		IsActive:        true,
		IsEmailVerified: false, // Email chưa được xác thực khi mới tạo
		NgayTao:         now,
		NgayCapNhat:     now,
	}, nil
}

// IsAdmin kiểm tra user có phải admin không
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// CanManage kiểm tra user có quyền quản lý không (admin hoặc manager)
func (u *User) CanManage() bool {
	return u.Role == RoleAdmin || u.Role == RoleManager
}

// Deactivate vô hiệu hóa tài khoản
func (u *User) Deactivate() {
	u.IsActive = false
	u.NgayCapNhat = time.Now()
}

// Activate kích hoạt tài khoản
func (u *User) Activate() {
	u.IsActive = true
	u.NgayCapNhat = time.Now()
}

// UpdatePassword cập nhật password hash
func (u *User) UpdatePassword(newPasswordHash string) error {
	if newPasswordHash == "" {
		return errors.New("password không được để trống")
	}
	u.PasswordHash = newPasswordHash
	u.NgayCapNhat = time.Now()
	return nil
}

// VerifyEmail đánh dấu email đã được xác thực
func (u *User) VerifyEmail() {
	u.IsEmailVerified = true
	u.NgayCapNhat = time.Now()
}
