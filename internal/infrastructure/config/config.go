// Package config chứa các cấu hình cho ứng dụng
package config

import (
	"os"
	"strconv"
	"time"
)

// Config chứa tất cả cấu hình của ứng dụng
type Config struct {
	Server  ServerConfig
	MySQL   MySQLConfig
	MongoDB MongoDBConfig
	Redis   RedisConfig
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
	}
}

// DSN trả về Data Source Name cho MySQL
// Format: username:password@tcp(host:port)/database?parseTime=true
func (c *MySQLConfig) DSN() string {
	return c.Username + ":" + c.Password +
		"@tcp(" + c.Host + ":" + strconv.Itoa(c.Port) + ")/" +
		c.Database + "?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci"
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
