package middleware

import (
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

	corsConfig := cors.Config{
		AllowOrigins:     cfg.AllowOrigins,
		AllowMethods:     cfg.AllowMethods,
		AllowHeaders:     cfg.AllowHeaders,
		ExposeHeaders:    []string{"Content-Length", "Content-Type", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           time.Duration(cfg.MaxAge) * time.Second,
	}

	return cors.New(corsConfig)
}
