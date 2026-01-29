// Package service chứa các Domain Service interfaces
package service

import (
	"context"
)

// EmailService interface cho việc gửi email
// Có 2 implementation:
// - ConsoleEmailService: Log ra console (development)
// - SMTPEmailService: Gửi email thật (production) - future
type EmailService interface {
	// SendVerificationEmail gửi email xác thực đến user
	// toEmail: địa chỉ email nhận
	// token: verification token
	// Trong development mode, chỉ log link ra console
	SendVerificationEmail(ctx context.Context, toEmail, token string) error
}
