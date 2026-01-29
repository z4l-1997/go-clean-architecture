package middleware

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"restaurant_project/internal/infrastructure/config"
)

// CORS middleware xử lý Cross-Origin Resource Sharing
// Cho phép frontend từ các domain khác truy cập API
func CORS(cfg config.CORSConfig) gin.HandlerFunc {
	if !cfg.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	env := os.Getenv("ENVIRONMENT")
	origins := cfg.AllowOrigins

	// Production safety: warn nếu dùng localhost hoặc wildcard
	if env == "production" {
		for _, origin := range origins {
			if origin == "*" {
				log.Fatalf("[CORS] FATAL: AllowOrigins chứa '*' trong production. Set CORS_ALLOW_ORIGINS với domain cụ thể")
			}
			if strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1") {
				log.Fatalf("[CORS] FATAL: AllowOrigins chứa '%s' trong production. Set CORS_ALLOW_ORIGINS với domain cụ thể (ví dụ: https://myapp.com)", origin)
			}
			if !strings.HasPrefix(origin, "https://") {
				log.Printf("[CORS] WARNING: Origin '%s' không dùng HTTPS trong production", origin)
			}
		}
	}

	corsConfig := cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     cfg.AllowMethods,
		AllowHeaders:     cfg.AllowHeaders,
		ExposeHeaders:    []string{"Content-Length", "Content-Type", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           time.Duration(cfg.MaxAge) * time.Second,
	}

	return cors.New(corsConfig)
}
