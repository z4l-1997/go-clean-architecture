package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"restaurant_project/internal/infrastructure/config"
	"restaurant_project/pkg/logger"
)

// BodySizeLimit middleware giới hạn kích thước request body
// Ngăn chặn các request quá lớn gây quá tải server
func BodySizeLimit(cfg config.BodyLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.Enabled {
			c.Next()
			return
		}

		// Kiểm tra Content-Length header
		if c.Request.ContentLength > cfg.MaxSize {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
				"error":      "Request body too large",
				"code":       "BODY_TOO_LARGE",
				"max_size":   cfg.MaxSize,
				"request_id": logger.GetRequestID(c),
			})
			return
		}

		// Wrap body với MaxBytesReader để giới hạn đọc
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, cfg.MaxSize)

		c.Next()
	}
}
