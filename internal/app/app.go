// Package app chứa Application struct và dependency wiring
package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"restaurant_project/internal/application/usecase"
	"restaurant_project/internal/domain/cache"
	"restaurant_project/internal/infrastructure/config"
	"restaurant_project/internal/infrastructure/database"
	persistenceCache "restaurant_project/internal/infrastructure/persistence/cache"
	"restaurant_project/internal/infrastructure/persistence/mongodb"
	"restaurant_project/internal/presentation/http/handler"
)

// App là application struct chứa tất cả dependencies
type App struct {
	cfg       *config.Config
	dbManager *database.DBManager
	server    *http.Server
	router    *gin.Engine

	// Connections
	mongoConn *database.MongoDBConnection
	redisConn *database.RedisConnection

	// Handlers
	monAnHandler  *handler.MonAnHandler
	healthHandler *handler.HealthHandler
}

// New tạo App instance mới
func New(cfg *config.Config) *App {
	return &App{
		cfg: cfg,
	}
}

// Run khởi chạy application
func (a *App) Run() error {
	ctx := context.Background()

	// Setup các components
	if err := a.setupDatabase(ctx); err != nil {
		return fmt.Errorf("setup database: %w", err)
	}

	a.setupRepositories()
	a.setupServer()

	// Chạy server với graceful shutdown
	return a.runWithGracefulShutdown()
}

// setupDatabase khởi tạo và kết nối databases
func (a *App) setupDatabase(ctx context.Context) error {
	fmt.Println("[1] Initializing Database Manager...")

	a.dbManager = database.NewDBManager()

	// Tạo connections
	a.mongoConn = database.NewMongoDBConnection(&a.cfg.MongoDB)
	a.redisConn = database.NewRedisConnection(&a.cfg.Redis)

	// Đăng ký connections
	a.dbManager.Register("mongodb", a.mongoConn)
	a.dbManager.Register("redis", a.redisConn)

	fmt.Printf("    Registered %d database connections\n", a.dbManager.Count())

	// Kết nối với timeout
	fmt.Println("[2] Connecting to databases...")
	connectCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := a.dbManager.ConnectAll(connectCtx); err != nil {
		return err
	}

	fmt.Println("    All databases connected")
	return nil
}

// setupRepositories khởi tạo repositories và use cases
func (a *App) setupRepositories() {
	fmt.Println("[3] Initializing repositories and use cases...")

	// MongoDB repository cho MonAn
	monAnRepo := mongodb.NewMonAnMongoRepo(a.mongoConn.Database())

	// Redis cache
	cacheOpts := cache.DefaultCacheOptions()
	redisCache := persistenceCache.NewRedisCacheRepository(a.redisConn.Client(), cacheOpts)

	// Cached repository
	cachedMonAnRepo := persistenceCache.NewCachedMonAnRepository(monAnRepo, redisCache, 5*time.Minute)

	// Use cases
	monAnUseCase := usecase.NewMonAnUseCase(cachedMonAnRepo)

	// Handlers
	a.monAnHandler = handler.NewMonAnHandler(monAnUseCase)
	a.healthHandler = handler.NewHealthHandler(a.dbManager)

	fmt.Println("    Repositories and handlers ready")
}

// setupServer tạo HTTP server với Gin
func (a *App) setupServer() {
	fmt.Println("[4] Setting up Gin HTTP server...")

	// Thiết lập Gin mode dựa vào environment
	if a.cfg.Server.Host == "0.0.0.0" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Tạo Gin router
	a.router = gin.New()

	// Thêm middleware mặc định
	a.router.Use(gin.Recovery())
	a.router.Use(gin.Logger())

	// Đăng ký routes
	a.registerRoutes(a.router)

	// Tạo HTTP server
	a.server = &http.Server{
		Addr:         a.cfg.Server.Host + ":" + a.cfg.Server.Port,
		Handler:      a.router,
		ReadTimeout:  a.cfg.Server.ReadTimeout,
		WriteTimeout: a.cfg.Server.WriteTimeout,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Println("    Gin HTTP server configured")
}

// runWithGracefulShutdown chạy server và xử lý shutdown
func (a *App) runWithGracefulShutdown() error {
	// Channel để nhận OS signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Channel để nhận server errors
	serverErr := make(chan error, 1)

	// Chạy server trong goroutine
	go func() {
		fmt.Println("\n==============================================")
		fmt.Printf("Server running at http://%s:%s\n", a.cfg.Server.Host, a.cfg.Server.Port)
		fmt.Println("Powered by Gin Framework")
		fmt.Println("==============================================")
		fmt.Println("\nEndpoints:")
		fmt.Printf("  curl http://localhost:%s/health\n", a.cfg.Server.Port)
		fmt.Printf("  curl http://localhost:%s/api/mon-an\n", a.cfg.Server.Port)
		fmt.Println("\nPress Ctrl+C to stop...")

		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Chờ signal hoặc error
	select {
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	case sig := <-quit:
		fmt.Printf("\n\nReceived signal: %v\n", sig)
		return a.shutdown()
	}
}

// shutdown thực hiện graceful shutdown
func (a *App) shutdown() error {
	fmt.Println("Starting graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), a.cfg.Server.ShutdownTimeout)
	defer cancel()

	// Shutdown HTTP server
	fmt.Println("  -> Stopping HTTP server...")
	if err := a.server.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	} else {
		fmt.Println("  -> HTTP server stopped")
	}

	// Shutdown databases
	fmt.Println("  -> Closing database connections...")
	if err := a.dbManager.Shutdown(ctx); err != nil {
		log.Printf("Database shutdown error: %v", err)
	} else {
		fmt.Println("  -> Database connections closed")
	}

	fmt.Println("\nGraceful shutdown completed")
	return nil
}
