// Package providers chứa các Wire providers
package providers

import (
	"restaurant_project/internal/domain/service"
	"restaurant_project/internal/infrastructure/config"
	infraservice "restaurant_project/internal/infrastructure/service"

	"github.com/redis/go-redis/v9"
)

// ProvideLoginAttemptService tạo LoginAttemptService từ Redis client
func ProvideLoginAttemptService(
	client *redis.Client,
	cfg *config.Config,
) service.LoginAttemptService {
	return infraservice.NewRedisLoginAttemptService(client, cfg.Middleware.AccountLockout)
}

// ProvideTokenBlacklistService tạo TokenBlacklistService từ Redis client
func ProvideTokenBlacklistService(
	client *redis.Client,
	cfg *config.Config,
) service.TokenBlacklistService {
	return infraservice.NewRedisTokenBlacklistService(client, cfg.Middleware.TokenBlacklist)
}

// ProvideEmailVerificationService tạo EmailVerificationService từ Redis client
func ProvideEmailVerificationService(
	client *redis.Client,
	cfg *config.Config,
) service.EmailVerificationService {
	return infraservice.NewRedisEmailVerificationService(client, cfg.Middleware.EmailVerification)
}

// ProvideEmailService tạo EmailService (Console mode cho development)
func ProvideEmailService(
	cfg *config.Config,
) service.EmailService {
	return infraservice.NewConsoleEmailService(cfg.Middleware.Email)
}
