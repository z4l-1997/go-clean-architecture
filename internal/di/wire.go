//go:build wireinject
// +build wireinject

// Package di chứa dependency injection configuration với Google Wire
package di

import (
	"github.com/google/wire"

	"restaurant_project/internal/di/providers"
	"restaurant_project/internal/infrastructure/config"
	"restaurant_project/internal/infrastructure/database"
	"restaurant_project/internal/infrastructure/migration"
	"restaurant_project/internal/presentation/http/handler"
)

// ============================================================
// Service Provider Set
// ============================================================

// ServiceSet chứa các providers cho Domain Service layer
var ServiceSet = wire.NewSet(
	providers.ProvideLoginAttemptService,
	providers.ProvideTokenBlacklistService,
	providers.ProvideEmailVerificationService,
	providers.ProvideEmailService,
)

// ============================================================
// Middleware Provider Set
// ============================================================

// MiddlewareSet chứa các providers cho Middleware layer
var MiddlewareSet = wire.NewSet(
	providers.ProvideJWTAuth,
	providers.ProvideMiddlewareCollection,
)

// MigrationSet chứa các providers cho Migration layer
var MigrationSet = wire.NewSet(
	providers.ProvideMySQLMigrator,
	providers.ProvideMigrationManager,
)

// ============================================================
// Provider Sets - Nhóm các providers theo layer
// ============================================================

// ConfigSet chứa các providers cho Config layer
var ConfigSet = wire.NewSet(
	providers.ProvideConfig,
	providers.ProvideMongoDBConfig,
	providers.ProvideRedisConfig,
	providers.ProvideServerConfig,
	providers.ProvideMySQLConfig,
)

// DatabaseSet chứa các providers cho Database layer
var DatabaseSet = wire.NewSet(
	providers.ProvideMongoDBConnection,
	providers.ProvideRedisConnection,
	providers.ProvideMySQLConnection,
	providers.ProvideDBManager,
	providers.ProvideMongoDB,
	providers.ProvideRedisClient,
	providers.ProvideMySQLDB,
)

// RepositorySet chứa các providers cho Repository layer
var RepositorySet = wire.NewSet(
	providers.ProvideMonAnMongoRepo,
	providers.ProvideRedisCacheRepository,
	providers.ProvideCachedMonAnRepository,
	providers.ProvideMonAnRepository,
	providers.ProvideUserMySQLRepo,
	providers.ProvideUserRepository,
)

// UseCaseSet chứa các providers cho UseCase layer
var UseCaseSet = wire.NewSet(
	providers.ProvideMonAnUseCase,
	providers.ProvideUserUseCase,
	providers.ProvideAuthUseCase,
)

// HandlerSet chứa các providers cho Handler layer
var HandlerSet = wire.NewSet(
	providers.ProvideMonAnHandler,
	providers.ProvideHealthHandler,
	providers.ProvideSwaggerHandler,
	providers.ProvideUserHandler,
	providers.ProvideAuthHandler,
)

// ============================================================
// App struct - Container chứa tất cả dependencies
// ============================================================

// App chứa tất cả dependencies đã được inject
type App struct {
	Config           *config.Config
	DBManager        *database.DBManager
	MigrationManager *migration.MigrationManager
	MonAnHandler     *handler.MonAnHandler
	HealthHandler    *handler.HealthHandler
	SwaggerHandler   *handler.SwaggerHandler
	UserHandler      *handler.UserHandler
	AuthHandler      *handler.AuthHandler
	Middlewares      *providers.MiddlewareCollection

	// Internal connections (để cleanup)
	MongoConn *database.MongoDBConnection
	RedisConn *database.RedisConnection
	MySQLConn *database.MySQLConnection
}

// ============================================================
// Injector function - Wire sẽ generate implementation
// ============================================================

// InitializeApp là injector function - Wire sẽ tự động generate code
// để tạo App với tất cả dependencies
func InitializeApp() (*App, error) {
	wire.Build(
		ConfigSet,
		DatabaseSet,
		MigrationSet,
		ServiceSet,
		RepositorySet,
		UseCaseSet,
		HandlerSet,
		MiddlewareSet,
		wire.Struct(new(App), "*"),
	)
	return nil, nil
}
