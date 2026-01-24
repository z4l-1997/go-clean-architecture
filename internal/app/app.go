// Package app chứa Application Runner với Wire DI
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

	"restaurant_project/internal/di"
)

// Runner quản lý việc chạy application
type Runner struct {
	app    *di.App
	server *http.Server
	router *gin.Engine
}

// NewRunner tạo Runner mới từ di.App
func NewRunner(app *di.App) *Runner {
	return &Runner{
		app: app,
	}
}

// Run khởi chạy application
func (r *Runner) Run() error {
	r.setupServer()
	return r.runWithGracefulShutdown()
}

// setupServer tạo HTTP server với Gin
func (r *Runner) setupServer() {
	fmt.Println("[4] Setting up Gin HTTP server...")

	// Thiết lập Gin mode dựa vào environment
	if r.app.Config.Server.Host == "0.0.0.0" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Tạo Gin router
	r.router = gin.New()

	// Thêm middleware mặc định
	r.router.Use(gin.Recovery())
	r.router.Use(gin.Logger())

	// Đăng ký routes
	r.registerRoutes(r.router)

	// Tạo HTTP server
	r.server = &http.Server{
		Addr:         r.app.Config.Server.Host + ":" + r.app.Config.Server.Port,
		Handler:      r.router,
		ReadTimeout:  r.app.Config.Server.ReadTimeout,
		WriteTimeout: r.app.Config.Server.WriteTimeout,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Println("    Gin HTTP server configured")
}

// registerRoutes đăng ký tất cả routes
func (r *Runner) registerRoutes(router *gin.Engine) {
	// Root endpoint
	router.GET("/", r.handleRoot)

	// Health endpoints (không có prefix /api)
	healthGroup := router.Group(r.app.HealthHandler.BasePath())
	r.app.HealthHandler.RegisterRoutes(healthGroup)

	// API routes
	api := router.Group("/api")
	{
		// MonAn routes
		monAnGroup := api.Group(r.app.MonAnHandler.BasePath())
		r.app.MonAnHandler.RegisterRoutes(monAnGroup)
	}
}

// handleRoot xử lý root endpoint
func (r *Runner) handleRoot(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Restaurant Menu API",
		"version": "2.0.0",
		"di":      "Google Wire",
		"endpoints": gin.H{
			"GET /health":                   "Full health check",
			"GET /health/live":              "Liveness probe",
			"GET /health/ready":             "Readiness probe",
			"GET /api/mon-an":               "List all dishes",
			"GET /api/mon-an?con_hang=true": "List available dishes",
			"GET /api/mon-an/:id":           "Get dish by ID",
			"POST /api/mon-an":              "Create new dish",
			"PUT /api/mon-an/:id/gia":       "Update price",
			"PUT /api/mon-an/:id/giam-gia":  "Apply discount",
			"PUT /api/mon-an/:id/het-hang":  "Mark as out of stock",
			"DELETE /api/mon-an/:id":        "Delete dish",
		},
	})
}

// runWithGracefulShutdown chạy server và xử lý shutdown
func (r *Runner) runWithGracefulShutdown() error {
	// Channel để nhận OS signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Channel để nhận server errors
	serverErr := make(chan error, 1)

	// Chạy server trong goroutine
	go func() {
		fmt.Println("\n==============================================")
		fmt.Printf("Server running at http://%s:%s\n", r.app.Config.Server.Host, r.app.Config.Server.Port)
		fmt.Println("Powered by Gin Framework + Google Wire DI")
		fmt.Println("==============================================")
		fmt.Println("\nEndpoints:")
		fmt.Printf("  curl http://localhost:%s/health\n", r.app.Config.Server.Port)
		fmt.Printf("  curl http://localhost:%s/api/mon-an\n", r.app.Config.Server.Port)
		fmt.Println("\nPress Ctrl+C to stop...")

		if err := r.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Chờ signal hoặc error
	select {
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	case sig := <-quit:
		fmt.Printf("\n\nReceived signal: %v\n", sig)
		return r.shutdown()
	}
}

// shutdown thực hiện graceful shutdown
func (r *Runner) shutdown() error {
	fmt.Println("Starting graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), r.app.Config.Server.ShutdownTimeout)
	defer cancel()

	// Shutdown HTTP server
	fmt.Println("  -> Stopping HTTP server...")
	if err := r.server.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	} else {
		fmt.Println("  -> HTTP server stopped")
	}

	// Shutdown databases
	fmt.Println("  -> Closing database connections...")
	if err := r.app.DBManager.Shutdown(ctx); err != nil {
		log.Printf("Database shutdown error: %v", err)
	} else {
		fmt.Println("  -> Database connections closed")
	}

	fmt.Println("\nGraceful shutdown completed")
	return nil
}
