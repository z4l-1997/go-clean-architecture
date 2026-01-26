package middleware

import (
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"

	"restaurant_project/internal/infrastructure/config"
)

// Gzip middleware nén response để giảm bandwidth
// Hỗ trợ các level từ 1 (nhanh, ít nén) đến 9 (chậm, nén tốt)
func Gzip(cfg config.GzipConfig) gin.HandlerFunc {
	if !cfg.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	// Đảm bảo level trong khoảng hợp lệ
	level := cfg.Level
	if level < gzip.BestSpeed {
		level = gzip.BestSpeed // 1
	}
	if level > gzip.BestCompression {
		level = gzip.BestCompression // 9
	}

	return gzip.Gzip(level)
}
