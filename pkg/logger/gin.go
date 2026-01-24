package logger

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	// RequestIDHeader là header name cho request ID
	RequestIDHeader = "X-Request-ID"
	// LoggerKey là key để lưu logger trong gin.Context
	LoggerKey = "logger"
	// RequestIDKey là key để lưu request ID trong gin.Context
	RequestIDKey = "request_id"
)

// GinLogger tạo logging middleware cho Gin
// Middleware này:
// - Tạo/lấy Request ID
// - Inject logger vào context
// - Log request/response với timing
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Lấy hoặc tạo Request ID
		requestID := c.GetHeader(RequestIDHeader)
		if requestID == "" {
			requestID = generateRequestID()
		}

		// Tạo logger với request context
		reqLogger := Log.With(
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("client_ip", c.ClientIP()),
		)

		// Inject vào gin context
		c.Set(LoggerKey, reqLogger)
		c.Set(RequestIDKey, requestID)
		c.Header(RequestIDHeader, requestID)

		// Process request
		c.Next()

		// Tính thời gian xử lý
		latency := time.Since(start)
		statusCode := c.Writer.Status()

		// Build log fields
		fields := []zap.Field{
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.Int("body_size", c.Writer.Size()),
		}

		if query != "" {
			fields = append(fields, zap.String("query", query))
		}

		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()))
		}

		// Chọn log level dựa trên status code
		msg := "HTTP Request"
		switch {
		case statusCode >= 500:
			reqLogger.Error(msg, fields...)
		case statusCode >= 400:
			reqLogger.Warn(msg, fields...)
		default:
			reqLogger.Info(msg, fields...)
		}
	}
}

// GinRecovery recovery middleware với structured logging
// Thay thế gin.Recovery() với logging chi tiết hơn
func GinRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Lấy stack trace
				stack := string(debug.Stack())

				// Lấy logger từ context
				reqLogger := GetLogger(c)

				// Log panic với đầy đủ thông tin
				reqLogger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("stack", stack),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
				)

				// Trả về error response
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error":      "Internal Server Error",
					"request_id": c.GetString(RequestIDKey),
				})
			}
		}()
		c.Next()
	}
}

// GetLogger lấy logger từ gin.Context
// Nếu không tìm thấy, trả về global logger
func GetLogger(c *gin.Context) *zap.Logger {
	if c == nil {
		return Log
	}
	if logger, exists := c.Get(LoggerKey); exists {
		if l, ok := logger.(*zap.Logger); ok {
			return l
		}
	}
	return Log
}

// GetRequestID lấy request ID từ gin.Context
func GetRequestID(c *gin.Context) string {
	if c == nil {
		return ""
	}
	return c.GetString(RequestIDKey)
}

// generateRequestID tạo unique request ID
func generateRequestID() string {
	id, err := uuid.NewV7()
	if err != nil {
		// Fallback nếu UUID generation fail
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return id.String()
}

// LoggerWithRequestID tạo child logger với request ID
// Hữu ích khi cần log trong goroutine riêng
func LoggerWithRequestID(requestID string) *zap.Logger {
	return Log.With(zap.String("request_id", requestID))
}
