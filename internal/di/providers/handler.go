// Package providers chứa các provider functions cho Wire DI
package providers

import (
	"fmt"

	"restaurant_project/internal/application/usecase"
	"restaurant_project/internal/infrastructure/database"
	"restaurant_project/internal/presentation/http/handler"
)

// ProvideMonAnHandler tạo MonAn HTTP handler
func ProvideMonAnHandler(uc *usecase.MonAnUseCase) *handler.MonAnHandler {
	return handler.NewMonAnHandler(uc)
}

// ProvideHealthHandler tạo Health HTTP handler
func ProvideHealthHandler(dbManager *database.DBManager) *handler.HealthHandler {
	return handler.NewHealthHandler(dbManager)
}

// ProvideHandlers trả về thông báo khi handlers đã sẵn sàng
func ProvideHandlers(
	monAnHandler *handler.MonAnHandler,
	healthHandler *handler.HealthHandler,
) {
	fmt.Println("[3] Initializing repositories and use cases...")
	fmt.Println("    Repositories and handlers ready")
}
