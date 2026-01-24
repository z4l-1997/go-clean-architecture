// Package providers chứa các provider functions cho Wire DI
package providers

import (
	"restaurant_project/internal/infrastructure/config"
)

// ProvideConfig tạo và trả về Config từ environment
func ProvideConfig() *config.Config {
	return config.Load()
}

// ProvideMongoDBConfig trích xuất MongoDBConfig từ Config
func ProvideMongoDBConfig(cfg *config.Config) *config.MongoDBConfig {
	return &cfg.MongoDB
}

// ProvideRedisConfig trích xuất RedisConfig từ Config
func ProvideRedisConfig(cfg *config.Config) *config.RedisConfig {
	return &cfg.Redis
}

// ProvideServerConfig trích xuất ServerConfig từ Config
func ProvideServerConfig(cfg *config.Config) *config.ServerConfig {
	return &cfg.Server
}
