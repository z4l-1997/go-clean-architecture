// Package logger cung cấp structured logging sử dụng Uber Zap
package logger

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Log là global logger instance
	Log  *zap.Logger
	once sync.Once
)

// Config chứa cấu hình cho logger
type Config struct {
	Level       string // debug, info, warn, error
	Environment string // development, production
	ServiceName string // Tên service
	Version     string // Version của service
}

// Init khởi tạo global logger (thread-safe, chỉ chạy 1 lần)
func Init(cfg Config) error {
	var err error
	once.Do(func() {
		Log, err = NewLogger(cfg)
	})
	return err
}

// NewLogger tạo một logger instance mới
func NewLogger(cfg Config) (*zap.Logger, error) {
	// Parse log level từ string
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel // Default là info
	}

	// Cấu hình encoder (format của log)
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var encoder zapcore.Encoder
	var core zapcore.Core

	if cfg.Environment == "production" {
		// Production: JSON format cho log aggregation (ELK, Loki, etc.)
		encoder = zapcore.NewJSONEncoder(encoderConfig)
		core = zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)
	} else {
		// Development: Console format với màu sắc
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
		core = zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)
	}

	// Tạo logger với default fields
	logger := zap.New(core,
		zap.AddCaller(),                          // Thêm file:line vào log
		zap.AddStacktrace(zapcore.ErrorLevel),    // Stack trace cho Error+
	).With(
		zap.String("service", cfg.ServiceName),
		zap.String("version", cfg.Version),
	)

	return logger, nil
}

// Sugar trả về SugaredLogger cho syntax đơn giản hơn
// Ví dụ: Sugar().Infof("User %s logged in", username)
func Sugar() *zap.SugaredLogger {
	if Log == nil {
		// Fallback nếu chưa init
		Log, _ = zap.NewDevelopment()
	}
	return Log.Sugar()
}

// Sync flush tất cả buffered log trước khi application shutdown
// QUAN TRỌNG: Luôn gọi defer logger.Sync() trong main()
func Sync() error {
	if Log != nil {
		return Log.Sync()
	}
	return nil
}

// With tạo một child logger với thêm fields
func With(fields ...zap.Field) *zap.Logger {
	if Log == nil {
		Log, _ = zap.NewDevelopment()
	}
	return Log.With(fields...)
}

// Debug logs a message at DebugLevel
func Debug(msg string, fields ...zap.Field) {
	Log.Debug(msg, fields...)
}

// Info logs a message at InfoLevel
func Info(msg string, fields ...zap.Field) {
	Log.Info(msg, fields...)
}

// Warn logs a message at WarnLevel
func Warn(msg string, fields ...zap.Field) {
	Log.Warn(msg, fields...)
}

// Error logs a message at ErrorLevel
func Error(msg string, fields ...zap.Field) {
	Log.Error(msg, fields...)
}

// Fatal logs a message at FatalLevel then calls os.Exit(1)
func Fatal(msg string, fields ...zap.Field) {
	Log.Fatal(msg, fields...)
}
