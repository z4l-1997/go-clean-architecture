// Package service chứa các Infrastructure Service implementations
package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"restaurant_project/internal/domain/service"
	"restaurant_project/internal/infrastructure/config"
	"restaurant_project/pkg/logger"
)

// Đảm bảo ConsoleEmailService implement EmailService
var _ service.EmailService = (*ConsoleEmailService)(nil)

// ConsoleEmailService implementation của EmailService cho development
// Thay vì gửi email thật, service này log verification link ra console
type ConsoleEmailService struct {
	baseURL string
	enabled bool
}

// NewConsoleEmailService tạo mới ConsoleEmailService
func NewConsoleEmailService(cfg config.EmailConfig) *ConsoleEmailService {
	return &ConsoleEmailService{
		baseURL: cfg.VerificationBaseURL,
		enabled: !cfg.Enabled, // Console service được dùng khi Email.Enabled = false
	}
}

// SendVerificationEmail log verification link ra console
func (s *ConsoleEmailService) SendVerificationEmail(ctx context.Context, toEmail, token string) error {
	if !s.enabled {
		return nil
	}

	// Build verification URL
	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", s.baseURL, token)

	// Log to console using structured logger
	logger.Info("[EMAIL] Verification email would be sent",
		zap.String("to", toEmail),
		zap.String("verification_url", verificationURL),
	)

	// Also print a prominent message to console for easy copy-paste
	fmt.Println("")
	fmt.Println("╔════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    EMAIL VERIFICATION (DEV MODE)                   ║")
	fmt.Println("╠════════════════════════════════════════════════════════════════════╣")
	fmt.Printf("║ To: %s\n", toEmail)
	fmt.Println("║────────────────────────────────────────────────────────────────────║")
	fmt.Println("║ Verification Link:                                                 ║")
	fmt.Printf("║ %s\n", verificationURL)
	fmt.Println("║────────────────────────────────────────────────────────────────────║")
	fmt.Printf("║ Token: %s\n", token)
	fmt.Println("╚════════════════════════════════════════════════════════════════════╝")
	fmt.Println("")

	return nil
}

// IsEnabled kiểm tra xem service có được bật không
func (s *ConsoleEmailService) IsEnabled() bool {
	return s.enabled
}
