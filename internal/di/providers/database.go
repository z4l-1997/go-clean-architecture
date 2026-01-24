// Package providers chứa các provider functions cho Wire DI
package providers

import (
	"context"
	"fmt"
	"time"

	"restaurant_project/internal/infrastructure/config"
	"restaurant_project/internal/infrastructure/database"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

// ProvideMongoDBConnection tạo MongoDB connection
func ProvideMongoDBConnection(cfg *config.MongoDBConfig) *database.MongoDBConnection {
	return database.NewMongoDBConnection(cfg)
}

// ProvideRedisConnection tạo Redis connection
func ProvideRedisConnection(cfg *config.RedisConfig) *database.RedisConnection {
	return database.NewRedisConnection(cfg)
}

// ProvideDBManager tạo và cấu hình DBManager với các connections
func ProvideDBManager(
	mongoConn *database.MongoDBConnection,
	redisConn *database.RedisConnection,
) (*database.DBManager, error) {
	fmt.Println("[1] Initializing Database Manager...")

	manager := database.NewDBManager()

	// Đăng ký connections
	manager.Register("mongodb", mongoConn)
	manager.Register("redis", redisConn)

	fmt.Printf("    Registered %d database connections\n", manager.Count())

	// Kết nối với timeout
	fmt.Println("[2] Connecting to databases...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := manager.ConnectAll(ctx); err != nil {
		return nil, fmt.Errorf("không thể kết nối databases: %w", err)
	}

	fmt.Println("    All databases connected")
	return manager, nil
}

// ProvideMongoDB trích xuất *mongo.Database từ MongoDBConnection
func ProvideMongoDB(conn *database.MongoDBConnection) *mongo.Database {
	return conn.Database()
}

// ProvideRedisClient trích xuất *redis.Client từ RedisConnection
func ProvideRedisClient(conn *database.RedisConnection) *redis.Client {
	return conn.Client()
}
