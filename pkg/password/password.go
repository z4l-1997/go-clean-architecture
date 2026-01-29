// Package password cung cấp các hàm tiện ích cho việc hash và verify password
package password

import (
	"golang.org/x/crypto/bcrypt"
)

// DefaultCost là cost mặc định cho bcrypt (12 được khuyến nghị cho production)
const DefaultCost = 12

// Hash tạo bcrypt hash từ password plaintext
func Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Verify kiểm tra password có khớp với hash không
func Verify(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
