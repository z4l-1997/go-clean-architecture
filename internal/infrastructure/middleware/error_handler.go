package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"restaurant_project/pkg/logger"
)

// AppError là custom error với HTTP status code
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
	Internal   error  `json:"-"` // Error gốc (không expose ra client)
}

func (e *AppError) Error() string {
	return e.Message
}

// NewAppError tạo AppError mới
func NewAppError(code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// WithInternal thêm internal error để logging
func (e *AppError) WithInternal(err error) *AppError {
	e.Internal = err
	return e
}

// Common errors
var (
	ErrNotFound         = NewAppError("NOT_FOUND", "Resource not found", http.StatusNotFound)
	ErrBadRequest       = NewAppError("BAD_REQUEST", "Invalid request", http.StatusBadRequest)
	ErrUnauthorized     = NewAppError("UNAUTHORIZED", "Unauthorized", http.StatusUnauthorized)
	ErrForbidden        = NewAppError("FORBIDDEN", "Forbidden", http.StatusForbidden)
	ErrInternal         = NewAppError("INTERNAL_ERROR", "Internal server error", http.StatusInternalServerError)
	ErrValidation       = NewAppError("VALIDATION_ERROR", "Validation failed", http.StatusBadRequest)
	ErrConflict         = NewAppError("CONFLICT", "Resource conflict", http.StatusConflict)
	ErrServiceUnavail   = NewAppError("SERVICE_UNAVAILABLE", "Service temporarily unavailable", http.StatusServiceUnavailable)
)

// ErrorHandler middleware xử lý lỗi tập trung
// Catch các error từ handlers và format response thống nhất
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Kiểm tra có error không
		if len(c.Errors) == 0 {
			return
		}

		// Lấy error đầu tiên
		err := c.Errors.Last().Err

		// Lấy logger từ context
		reqLogger := logger.GetLogger(c)
		if reqLogger == nil {
			reqLogger = logger.Log
		}

		// Xử lý theo loại error
		var appErr *AppError
		if errors.As(err, &appErr) {
			// AppError - đã có status code
			if appErr.Internal != nil {
				reqLogger.Error("Request error",
					zap.String("code", appErr.Code),
					zap.String("message", appErr.Message),
					zap.Error(appErr.Internal),
				)
			} else {
				reqLogger.Warn("Request error",
					zap.String("code", appErr.Code),
					zap.String("message", appErr.Message),
				)
			}

			c.JSON(appErr.StatusCode, gin.H{
				"error":      appErr.Message,
				"code":       appErr.Code,
				"request_id": logger.GetRequestID(c),
			})
			return
		}

		// Generic error - log và trả về 500
		reqLogger.Error("Unhandled error",
			zap.Error(err),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Internal server error",
			"code":       "INTERNAL_ERROR",
			"request_id": logger.GetRequestID(c),
		})
	}
}

// AbortWithError helper để handlers báo lỗi
func AbortWithError(c *gin.Context, err *AppError) {
	c.Error(err)
	c.Abort()
}

// AbortWithNotFound helper
func AbortWithNotFound(c *gin.Context, message string) {
	err := NewAppError("NOT_FOUND", message, http.StatusNotFound)
	AbortWithError(c, err)
}

// AbortWithBadRequest helper
func AbortWithBadRequest(c *gin.Context, message string) {
	err := NewAppError("BAD_REQUEST", message, http.StatusBadRequest)
	AbortWithError(c, err)
}

// AbortWithValidationError helper cho validation errors
func AbortWithValidationError(c *gin.Context, details interface{}) {
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
		"error":      "Validation failed",
		"code":       "VALIDATION_ERROR",
		"details":    details,
		"request_id": logger.GetRequestID(c),
	})
}

// AbortWithInternalError helper - không expose chi tiết
func AbortWithInternalError(c *gin.Context, internalErr error) {
	err := ErrInternal.WithInternal(internalErr)
	AbortWithError(c, err)
}
