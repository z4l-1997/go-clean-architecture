// Package config chứa các cấu hình cho ứng dụng
package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config chứa tất cả cấu hình của ứng dụng
type Config struct {
	Server     ServerConfig
	Log        LogConfig
	MySQL      MySQLConfig
	MongoDB    MongoDBConfig
	Redis      RedisConfig
	Migration  MigrationConfig
	Middleware MiddlewareConfig
}

// MigrationConfig cấu hình cho hệ thống migration tự động
type MigrationConfig struct {
	AutoMigrate bool // Tự động chạy migration khi startup
}

// MiddlewareConfig chứa cấu hình cho tất cả middleware
type MiddlewareConfig struct {
	CORS              CORSConfig
	RateLimit         RateLimitConfig
	AuthRateLimit     AuthRateLimitConfig
	JWT               JWTConfig
	Timeout           TimeoutConfig
	Gzip              GzipConfig
	Security          SecurityConfig
	BodyLimit         BodyLimitConfig
	AccountLockout    AccountLockoutConfig
	TokenBlacklist    TokenBlacklistConfig
	EmailVerification EmailVerificationConfig
	Email             EmailConfig
}

// CORSConfig cấu hình Cross-Origin Resource Sharing
type CORSConfig struct {
	Enabled      bool     // Bật/tắt CORS
	AllowOrigins []string // Origins được phép (ví dụ: http://localhost:3000)
	AllowMethods []string // Methods được phép (GET, POST, PUT, DELETE)
	AllowHeaders []string // Headers được phép
	MaxAge       int      // Thời gian cache preflight (giây)
}

// RateLimitConfig cấu hình giới hạn số request
type RateLimitConfig struct {
	Enabled bool    // Bật/tắt rate limiting
	RPS     float64 // Requests per second
	Burst   int     // Burst size
}

// AuthRateLimitConfig cấu hình rate limit riêng cho auth endpoints
// Nghiêm ngặt hơn global rate limit để chống brute force
type AuthRateLimitConfig struct {
	Enabled bool    // Bật/tắt auth rate limiting
	RPS     float64 // Requests per second (mặc định 5)
	Burst   int     // Burst size (mặc định 10)
}

// JWTConfig cấu hình JWT Authentication
type JWTConfig struct {
	Enabled         bool   // Bật/tắt JWT auth
	SecretKey       string // Secret key để sign token
	AccessTokenTTL  string // Thời gian sống access token (ví dụ: 15m)
	RefreshTokenTTL string // Thời gian sống refresh token (ví dụ: 168h)
}

// TimeoutConfig cấu hình request timeout
type TimeoutConfig struct {
	Enabled  bool   // Bật/tắt timeout
	Duration string // Thời gian timeout (ví dụ: 30s)
}

// GzipConfig cấu hình nén response
type GzipConfig struct {
	Enabled bool // Bật/tắt gzip compression
	Level   int  // Compression level (1-9)
}

// SecurityConfig cấu hình security headers
type SecurityConfig struct {
	Enabled    bool // Bật/tắt security headers
	HSTSMaxAge int  // Max-Age cho HSTS header (giây)
}

// BodyLimitConfig cấu hình giới hạn kích thước request body
type BodyLimitConfig struct {
	Enabled bool  // Bật/tắt body size limit
	MaxSize int64 // Kích thước tối đa (bytes)
}

// AccountLockoutConfig cấu hình khóa tài khoản sau nhiều lần login sai
type AccountLockoutConfig struct {
	Enabled         bool          // Bật/tắt account lockout
	MaxAttempts     int           // Số lần thử tối đa trước khi khóa (mặc định 5)
	LockoutDuration time.Duration // Thời gian khóa (mặc định 15 phút)
}

// TokenBlacklistConfig cấu hình token blacklist
type TokenBlacklistConfig struct {
	Enabled bool // Bật/tắt token blacklist (dùng Redis có sẵn)
}

// EmailVerificationConfig cấu hình xác thực email
type EmailVerificationConfig struct {
	Enabled     bool          // Bật/tắt email verification
	TokenTTL    time.Duration // Thời gian sống của verification token (mặc định 24h)
	CooldownTTL time.Duration // Thời gian chờ giữa các lần gửi lại (mặc định 60s)
}

// EmailConfig cấu hình gửi email
type EmailConfig struct {
	Enabled             bool   // Bật/tắt gửi email thật (false = console log)
	VerificationBaseURL string // Base URL cho link xác thực (vd: http://localhost:3000)
}

// LogConfig cấu hình cho structured logger
type LogConfig struct {
	Level       string // debug, info, warn, error
	Environment string // development, production

	// File logging với rotation
	EnableFileLog bool   // Bật/tắt ghi log ra file
	LogFilePath   string // Đường dẫn file log
	MaxSizeMB     int    // Kích thước tối đa mỗi file (MB)
	MaxBackups    int    // Số file backup giữ lại
	MaxAgeDays    int    // Số ngày giữ file cũ
	CompressLog   bool   // Nén file cũ (gzip)
}

// ServerConfig cấu hình HTTP server
type ServerConfig struct {
	Port            string        // Port server lắng nghe
	Host            string        // Host address
	ReadTimeout     time.Duration // Timeout đọc request
	WriteTimeout    time.Duration // Timeout ghi response
	ShutdownTimeout time.Duration // Timeout graceful shutdown
}

// MySQLConfig cấu hình kết nối MySQL
// Dùng cho: User, KhachHang, NhanVien
type MySQLConfig struct {
	Host     string // Địa chỉ host
	Port     int    // Port (mặc định 3306)
	Username string // Username
	Password string // Password
	Database string // Tên database
}

// MongoDBConfig cấu hình kết nối MongoDB
// Dùng cho: MonAn (Menu), Order
type MongoDBConfig struct {
	URI      string // MongoDB connection URI
	Database string // Tên database
}

// RedisConfig cấu hình kết nối Redis
// Dùng cho: Caching, Session
type RedisConfig struct {
	Addr     string // Host:Port (mặc định localhost:6379)
	Password string // Password (để trống nếu không có)
	DB       int    // Database number (mặc định 0)
}

// Load đọc cấu hình từ environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:            getEnv("SERVER_PORT", "8080"),
			Host:            getEnv("SERVER_HOST", "0.0.0.0"),
			ReadTimeout:     getEnvAsDuration("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:    getEnvAsDuration("SERVER_WRITE_TIMEOUT", 15*time.Second),
			ShutdownTimeout: getEnvAsDuration("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
		},
		Log: LogConfig{
			Level:         getEnv("LOG_LEVEL", "info"),
			Environment:   getEnv("ENVIRONMENT", "development"),
			EnableFileLog: getEnvAsBool("LOG_FILE_ENABLED", true),
			LogFilePath:   getEnv("LOG_FILE_PATH", "logs/app.log"),
			MaxSizeMB:     getEnvAsInt("LOG_FILE_MAX_SIZE_MB", 100),
			MaxBackups:    getEnvAsInt("LOG_FILE_MAX_BACKUPS", 5),
			MaxAgeDays:    getEnvAsInt("LOG_FILE_MAX_AGE_DAYS", 30),
			CompressLog:   getEnvAsBool("LOG_FILE_COMPRESS", true),
		},
		MySQL: MySQLConfig{
			Host:     getEnv("MYSQL_HOST", "localhost"),
			Port:     getEnvAsInt("MYSQL_PORT", 3306),
			Username: getEnv("MYSQL_USER", "root"),
			Password: getEnv("MYSQL_PASSWORD", ""),
			Database: getEnv("MYSQL_DATABASE", "restaurant_db"),
		},
		MongoDB: MongoDBConfig{
			URI:      getEnv("MONGODB_URI", "mongodb://localhost:27017"),
			Database: getEnv("MONGODB_DATABASE", "restaurant_db"),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Migration: MigrationConfig{
			AutoMigrate: getEnvAsBool("MIGRATION_AUTO_MIGRATE", true),
		},
		Middleware: MiddlewareConfig{
			CORS: CORSConfig{
				Enabled:      getEnvAsBool("CORS_ENABLED", true),
				AllowOrigins: getEnvAsStringSlice("CORS_ALLOW_ORIGINS", []string{"http://localhost:3000", "http://localhost:5173"}),
				AllowMethods: getEnvAsStringSlice("CORS_ALLOW_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
				AllowHeaders: getEnvAsStringSlice("CORS_ALLOW_HEADERS", []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"}),
				MaxAge:       getEnvAsInt("CORS_MAX_AGE", 86400),
			},
			RateLimit: RateLimitConfig{
				Enabled: getEnvAsBool("RATE_LIMIT_ENABLED", true),
				RPS:     getEnvAsFloat("RATE_LIMIT_RPS", 100),
				Burst:   getEnvAsInt("RATE_LIMIT_BURST", 200),
			},
			JWT: JWTConfig{
				Enabled:         getEnvAsBool("JWT_ENABLED", true),
				SecretKey:       getEnv("JWT_SECRET_KEY", "change-this-in-production"),
				AccessTokenTTL:  getEnv("JWT_ACCESS_TOKEN_TTL", "15m"),
				RefreshTokenTTL: getEnv("JWT_REFRESH_TOKEN_TTL", "168h"),
			},
			Timeout: TimeoutConfig{
				Enabled:  getEnvAsBool("TIMEOUT_ENABLED", true),
				Duration: getEnv("TIMEOUT_DURATION", "30s"),
			},
			Gzip: GzipConfig{
				Enabled: getEnvAsBool("GZIP_ENABLED", true),
				Level:   getEnvAsInt("GZIP_LEVEL", 5),
			},
			Security: SecurityConfig{
				Enabled:    getEnvAsBool("SECURITY_HEADERS_ENABLED", true),
				HSTSMaxAge: getEnvAsInt("SECURITY_HSTS_MAX_AGE", 31536000),
			},
			BodyLimit: BodyLimitConfig{
				Enabled: getEnvAsBool("BODY_LIMIT_ENABLED", true),
				MaxSize: getEnvAsInt64("BODY_LIMIT_MAX_SIZE", 1048576), // 1MB
			},
			AuthRateLimit: AuthRateLimitConfig{
				Enabled: getEnvAsBool("AUTH_RATE_LIMIT_ENABLED", true),
				RPS:     getEnvAsFloat("AUTH_RATE_LIMIT_RPS", 5),
				Burst:   getEnvAsInt("AUTH_RATE_LIMIT_BURST", 10),
			},
			AccountLockout: AccountLockoutConfig{
				Enabled:         getEnvAsBool("ACCOUNT_LOCKOUT_ENABLED", true),
				MaxAttempts:     getEnvAsInt("ACCOUNT_LOCKOUT_MAX_ATTEMPTS", 5),
				LockoutDuration: getEnvAsDuration("ACCOUNT_LOCKOUT_DURATION", 15*time.Minute),
			},
			TokenBlacklist: TokenBlacklistConfig{
				Enabled: getEnvAsBool("TOKEN_BLACKLIST_ENABLED", true),
			},
			EmailVerification: EmailVerificationConfig{
				Enabled:     getEnvAsBool("EMAIL_VERIFICATION_ENABLED", true),
				TokenTTL:    getEnvAsDuration("EMAIL_VERIFICATION_TOKEN_TTL", 24*time.Hour),
				CooldownTTL: getEnvAsDuration("EMAIL_VERIFICATION_COOLDOWN_TTL", 60*time.Second),
			},
			Email: EmailConfig{
				Enabled:             getEnvAsBool("EMAIL_ENABLED", false),
				VerificationBaseURL: getEnv("EMAIL_VERIFICATION_BASE_URL", "http://localhost:3000"),
			},
		},
	}
}

// DSN trả về Data Source Name cho MySQL
// Format: username:password@tcp(host:port)/database?parseTime=true
func (c *MySQLConfig) DSN() string {
	return c.Username + ":" + c.Password +
		"@tcp(" + c.Host + ":" + strconv.Itoa(c.Port) + ")/" +
		c.Database + "?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci&multiStatements=true"
}

// getEnv lấy giá trị từ env, nếu không có thì trả về defaultValue
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt lấy giá trị int từ env
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// getEnvAsDuration lấy giá trị duration từ env
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// getEnvAsBool lấy giá trị bool từ env
func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// getEnvAsFloat lấy giá trị float64 từ env
func getEnvAsFloat(key string, defaultValue float64) float64 {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return value
	}
	return defaultValue
}

// getEnvAsInt64 lấy giá trị int64 từ env
func getEnvAsInt64(key string, defaultValue int64) int64 {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	if value, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
		return value
	}
	return defaultValue
}

// getEnvAsStringSlice lấy giá trị []string từ env (comma-separated)
func getEnvAsStringSlice(key string, defaultValue []string) []string {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	// Split by comma and trim spaces
	parts := strings.Split(valueStr, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return defaultValue
	}
	return result
}
