// Package providers chứa các provider functions cho Wire DI
package providers

import (
	"restaurant_project/internal/application/usecase"
	"restaurant_project/internal/domain/repository"
	"restaurant_project/internal/domain/service"
	"restaurant_project/internal/infrastructure/middleware"
)

// ProvideMonAnUseCase tạo MonAn use case
func ProvideMonAnUseCase(repo repository.IMonAnRepository) *usecase.MonAnUseCase {
	return usecase.NewMonAnUseCase(repo)
}

// ProvideUserUseCase tạo User use case
func ProvideUserUseCase(repo repository.IUserRepository, tokenBlacklist service.TokenBlacklistService) *usecase.UserUseCase {
	return usecase.NewUserUseCase(repo, tokenBlacklist)
}

// ProvideAuthUseCase tạo Auth use case
func ProvideAuthUseCase(
	repo repository.IUserRepository,
	jwtAuth *middleware.JWTAuthMiddleware,
	loginAttemptService service.LoginAttemptService,
	emailVerificationService service.EmailVerificationService,
	emailService service.EmailService,
) *usecase.AuthUseCase {
	return usecase.NewAuthUseCase(repo, jwtAuth, loginAttemptService, emailVerificationService, emailService)
}
