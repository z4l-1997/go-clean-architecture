package logger

import (
	"context"

	"go.uber.org/zap"
)

// ctxKey là key để lưu logger trong context
type ctxKey struct{}

// WithContext thêm logger vào context
// Sử dụng khi cần pass logger qua các layer (handler -> usecase -> repository)
func WithContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, logger)
}

// FromContext lấy logger từ context
// Nếu không tìm thấy, trả về global logger
func FromContext(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return Log
	}
	if logger, ok := ctx.Value(ctxKey{}).(*zap.Logger); ok {
		return logger
	}
	return Log // Fallback to global logger
}

// WithFields tạo context mới với logger có thêm fields
// Hữu ích khi cần thêm thông tin như user_id, order_id vào tất cả log trong một request
//
// Ví dụ:
//
//	ctx = logger.WithFields(ctx,
//	    zap.String("user_id", userID),
//	    zap.String("order_id", orderID),
//	)
func WithFields(ctx context.Context, fields ...zap.Field) context.Context {
	logger := FromContext(ctx).With(fields...)
	return WithContext(ctx, logger)
}

// CtxDebug logs debug message với logger từ context
func CtxDebug(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).Debug(msg, fields...)
}

// CtxInfo logs info message với logger từ context
func CtxInfo(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).Info(msg, fields...)
}

// CtxWarn logs warn message với logger từ context
func CtxWarn(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).Warn(msg, fields...)
}

// CtxError logs error message với logger từ context
func CtxError(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).Error(msg, fields...)
}
