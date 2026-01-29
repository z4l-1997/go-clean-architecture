// Package providers chứa các provider functions cho Wire DI
package providers

import (
	"database/sql"
	"time"

	"restaurant_project/internal/domain/cache"
	"restaurant_project/internal/domain/repository"
	persistenceCache "restaurant_project/internal/infrastructure/persistence/cache"
	"restaurant_project/internal/infrastructure/persistence/mongodb"
	"restaurant_project/internal/infrastructure/persistence/mysql"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

// ProvideMonAnMongoRepo tạo MonAn MongoDB repository
func ProvideMonAnMongoRepo(db *mongo.Database) *mongodb.MonAnMongoRepo {
	return mongodb.NewMonAnMongoRepo(db)
}

// ProvideRedisCacheRepository tạo Redis cache repository
func ProvideRedisCacheRepository(client *redis.Client) *persistenceCache.RedisCacheRepository {
	cacheOpts := cache.DefaultCacheOptions()
	return persistenceCache.NewRedisCacheRepository(client, cacheOpts)
}

// ProvideCachedMonAnRepository tạo Cached MonAn repository với decorator pattern
func ProvideCachedMonAnRepository(
	repo *mongodb.MonAnMongoRepo,
	cacheRepo *persistenceCache.RedisCacheRepository,
) *persistenceCache.CachedMonAnRepository {
	return persistenceCache.NewCachedMonAnRepository(repo, cacheRepo, 5*time.Minute)
}

// ProvideMonAnRepository binds CachedMonAnRepository to IMonAnRepository interface
func ProvideMonAnRepository(cached *persistenceCache.CachedMonAnRepository) repository.IMonAnRepository {
	return cached
}

// ProvideUserMySQLRepo tạo User MySQL repository
func ProvideUserMySQLRepo(db *sql.DB) *mysql.UserMySQLRepo {
	return mysql.NewUserMySQLRepo(db)
}

// ProvideUserRepository binds UserMySQLRepo to IUserRepository interface
func ProvideUserRepository(repo *mysql.UserMySQLRepo) repository.IUserRepository {
	return repo
}
