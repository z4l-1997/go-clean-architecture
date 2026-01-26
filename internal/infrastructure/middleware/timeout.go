package middleware

import (
	"net/http"
	"time"

	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"

	"restaurant_project/internal/infrastructure/config"
	"restaurant_project/pkg/logger"
)

// Timeout middleware giới hạn thời gian xử lý request
// Ngăn chặn request treo quá lâu làm cạn kiệt resources
func Timeout(cfg config.TimeoutConfig) gin.HandlerFunc {
	if !cfg.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	// Parse duration
	duration, err := time.ParseDuration(cfg.Duration)
	if err != nil {
		duration = 30 * time.Second // Default 30s
	}

	return timeout.New(
		timeout.WithTimeout(duration),
		timeout.WithHandler(func(c *gin.Context) {
			c.Next()
		}),
		timeout.WithResponse(timeoutResponse),
	)
}

// timeoutResponse trả về response khi request timeout
func timeoutResponse(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusGatewayTimeout, gin.H{
		"error":      "Request timeout",
		"code":       "REQUEST_TIMEOUT",
		"request_id": logger.GetRequestID(c),
	})
}
