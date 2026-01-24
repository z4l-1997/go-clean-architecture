// Package providers chứa các provider functions cho Wire DI
package providers

import (
	"restaurant_project/internal/application/usecase"
	"restaurant_project/internal/domain/repository"
)

// ProvideMonAnUseCase tạo MonAn use case
func ProvideMonAnUseCase(repo repository.IMonAnRepository) *usecase.MonAnUseCase {
	return usecase.NewMonAnUseCase(repo)
}
