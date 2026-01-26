// Package logger cung cấp structured logging sử dụng Uber Zap
package logger

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Log là global logger instance
	Log  *zap.Logger
	once sync.Once
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"

	// Bold colors
	colorBoldRed    = "\033[1;31m"
	colorBoldYellow = "\033[1;33m"

	// Background colors
	bgRed = "\033[41m"
)

// customLevelEncoder tạo encoder với emoji và màu sắc nổi bật
func customLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	var levelStr string
	switch level {
	case zapcore.DebugLevel:
		levelStr = colorCyan + "🔵 DEBUG" + colorReset
	case zapcore.InfoLevel:
		levelStr = colorGreen + "🟢 INFO " + colorReset
	case zapcore.WarnLevel:
		levelStr = colorBoldYellow + "🟡 WARN " + colorReset
	case zapcore.ErrorLevel:
		levelStr = colorBoldRed + "🔴 ERROR" + colorReset
	case zapcore.DPanicLevel:
		levelStr = bgRed + colorWhite + "⚠️  DPANIC" + colorReset
	case zapcore.PanicLevel:
		levelStr = bgRed + colorWhite + "🚨 PANIC" + colorReset
	case zapcore.FatalLevel:
		levelStr = bgRed + colorWhite + "💀 FATAL" + colorReset
	default:
		levelStr = level.CapitalString()
	}
	enc.AppendString(levelStr)
}

// Config chứa cấu hình cho logger
type Config struct {
	Level       string // debug, info, warn, error
	Environment string // development, production
	ServiceName string // Tên service
	Version     string // Version của service

	// File logging config
	EnableFileLog bool   // Bật/tắt ghi log ra file
	LogFilePath   string // Đường dẫn file log (default: logs/app.log)
	MaxSizeMB     int    // Kích thước tối đa mỗi file (MB), default: 100
	MaxBackups    int    // Số file backup giữ lại, default: 5
	MaxAgeDays    int    // Số ngày giữ file cũ, default: 30
	CompressLog   bool   // Nén file cũ (gzip), default: true
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

	var consoleEncoder zapcore.Encoder
	var fileEncoder zapcore.Encoder
	var cores []zapcore.Core

	// === CONSOLE OUTPUT ===
	if cfg.Environment == "production" {
		// Production: JSON format cho console
		consoleEncoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		// Development: Console format với emoji và màu sắc nổi bật
		devEncoderConfig := encoderConfig
		devEncoderConfig.EncodeLevel = customLevelEncoder // Custom encoder với emoji
		consoleEncoder = zapcore.NewConsoleEncoder(devEncoderConfig)
	}
	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level)
	cores = append(cores, consoleCore)

	// === FILE OUTPUT (với rotation theo ngày + size) ===
	if cfg.EnableFileLog {
		// Set defaults nếu chưa config
		if cfg.LogFilePath == "" {
			cfg.LogFilePath = "logs/app.log"
		}
		if cfg.MaxSizeMB == 0 {
			cfg.MaxSizeMB = 100 // 100MB mỗi file
		}
		if cfg.MaxBackups == 0 {
			cfg.MaxBackups = 5 // Giữ 5 file backup
		}
		if cfg.MaxAgeDays == 0 {
			cfg.MaxAgeDays = 30 // Giữ 30 ngày
		}

		// Tạo thư mục logs nếu chưa có
		logDir := filepath.Dir(cfg.LogFilePath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, err
		}

		// Tạo pattern cho file rotation
		// logs/app.log → logs/app.2024-01-27.log
		ext := filepath.Ext(cfg.LogFilePath)
		basePath := strings.TrimSuffix(cfg.LogFilePath, ext)
		pattern := basePath + ".%Y-%m-%d" + ext

		// Cấu hình file-rotatelogs
		// - Rotate mỗi ngày (24h)
		// - Rotate khi file vượt MaxSizeMB
		// - Giữ file trong MaxAgeDays ngày
		// Lưu ý: Không dùng cả MaxAge và RotationCount cùng lúc
		fileWriter, err := rotatelogs.New(
			pattern,
			rotatelogs.WithLinkName(cfg.LogFilePath),                          // Symlink đến file mới nhất
			rotatelogs.WithRotationTime(24*time.Hour),                         // Rotate mỗi 24h
			rotatelogs.WithRotationSize(int64(cfg.MaxSizeMB)*1024*1024),       // Rotate khi vượt size
			rotatelogs.WithMaxAge(time.Duration(cfg.MaxAgeDays)*24*time.Hour), // Xóa file cũ hơn X ngày
		)
		if err != nil {
			return nil, err
		}

		// File luôn dùng JSON format (dễ parse)
		fileEncoder = zapcore.NewJSONEncoder(encoderConfig)
		fileCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(fileWriter), level)
		cores = append(cores, fileCore)
	}

	// Kết hợp tất cả cores
	core := zapcore.NewTee(cores...)

	// Tạo logger với default fields
	logger := zap.New(core,
		zap.AddCaller(),                       // Thêm file:line vào log
		zap.AddStacktrace(zapcore.ErrorLevel), // Stack trace cho Error+
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
