// Package middleware chứa các HTTP middleware cho ứng dụng
package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"restaurant_project/internal/infrastructure/config"
)

// SecurityHeaders middleware thêm các security headers vào response
// Bảo vệ khỏi các tấn công phổ biến như XSS, Clickjacking, MIME-sniffing
func SecurityHeaders(cfg config.SecurityConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.Enabled {
			c.Next()
			return
		}

		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// Enable XSS filter (legacy browsers)
		c.Header("X-XSS-Protection", "1; mode=block")

		// Referrer policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy (basic)
		c.Header("Content-Security-Policy", "default-src 'self'")

		// Permissions Policy (formerly Feature-Policy)
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// HSTS - chỉ áp dụng cho HTTPS trong production
		if cfg.HSTSMaxAge > 0 {
			hstsValue := fmt.Sprintf("max-age=%d; includeSubDomains", cfg.HSTSMaxAge)
			c.Header("Strict-Transport-Security", hstsValue)
		}

		c.Next()
	}
}
