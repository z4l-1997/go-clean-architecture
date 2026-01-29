// Package app chứa Application Runner với Wire DI
package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"restaurant_project/internal/di"
	"restaurant_project/pkg/logger"
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
	logger.Info("Setting up Gin HTTP server...")

	// Thiết lập Gin mode dựa vào environment
	if r.app.Config.Log.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Tạo Gin router
	r.router = gin.New()

	// ============================================================
	// MIDDLEWARE CHAIN - Thứ tự rất quan trọng!
	// ============================================================

	// 1. Recovery - catch panics trước tiên
	r.router.Use(logger.GinRecovery())

	// 2. Logger - log tất cả requests
	r.router.Use(logger.GinLogger())

	// 3. Security Headers - thêm security headers sớm
	r.router.Use(r.app.Middlewares.SecurityHeaders)

	// 4. CORS - xử lý preflight requests
	r.router.Use(r.app.Middlewares.CORS)

	// 5. Rate Limit - throttle requests sớm để bảo vệ resources
	r.router.Use(r.app.Middlewares.RateLimit)

	// 6. Body Size Limit - kiểm tra kích thước request
	r.router.Use(r.app.Middlewares.BodySizeLimit)

	// 7. Timeout - giới hạn thời gian xử lý request
	r.router.Use(r.app.Middlewares.Timeout)

	// 8. Gzip - nén response (cuối cùng trong chain)
	r.router.Use(r.app.Middlewares.Gzip)

	// 9. Error Handler - catch errors từ handlers
	r.router.Use(r.app.Middlewares.ErrorHandler)

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

	logger.Info("Gin HTTP server configured",
		zap.String("addr", r.server.Addr),
		zap.Duration("read_timeout", r.app.Config.Server.ReadTimeout),
		zap.Duration("write_timeout", r.app.Config.Server.WriteTimeout),
	)
}

// registerRoutes đăng ký tất cả routes
func (r *Runner) registerRoutes(router *gin.Engine) {
	// Root endpoint
	router.GET("/", r.handleRoot)

	// Health endpoints (không có prefix /api)
	healthGroup := router.Group(r.app.HealthHandler.BasePath())
	r.app.HealthHandler.RegisterRoutes(healthGroup)

	// Swagger endpoints - đăng ký trực tiếp trên router
	r.app.SwaggerHandler.RegisterRoutesOnEngine(router)

	// API routes
	api := router.Group("/api")
	{
		// Auth routes (PUBLIC - không cần JWT)
		// Áp dụng Auth Rate Limit nghiêm ngặt hơn để chống brute force
		authGroup := api.Group(r.app.AuthHandler.BasePath())
		authGroup.Use(r.app.Middlewares.AuthRateLimit)
		r.app.AuthHandler.RegisterRoutes(authGroup)

		// Auth protected routes (cần JWT) - cho logout
		authProtectedGroup := api.Group(r.app.AuthHandler.BasePath())
		authProtectedGroup.Use(r.app.Middlewares.AuthRateLimit)
		authProtectedGroup.Use(r.app.Middlewares.JWTAuth.Middleware())
		r.app.AuthHandler.RegisterProtectedRoutes(authProtectedGroup)

		// MonAn routes
		monAnGroup := api.Group(r.app.MonAnHandler.BasePath())
		r.app.MonAnHandler.RegisterRoutes(monAnGroup)

		// User routes (PROTECTED - cần JWT)
		userGroup := api.Group(r.app.UserHandler.BasePath())
		userGroup.Use(r.app.Middlewares.JWTAuth.Middleware())
		r.app.UserHandler.RegisterRoutes(userGroup)
	}

	logger.Debug("Routes registered successfully")
}

// handleRoot xử lý root endpoint
func (r *Runner) handleRoot(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Restaurant Menu API",
		"version": "2.0.0",
		"di":      "Google Wire",
		"logger":  "Uber Zap",
		"endpoints": gin.H{
			"GET /swagger/index.html":       "Swagger UI",
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
			"POST /api/auth/register":       "Register new customer",
			"POST /api/auth/login":          "Login",
			"POST /api/auth/refresh":        "Refresh access token",
			"POST /api/auth/logout":         "Logout (revoke token) [Auth]",
			"GET /api/users/me":             "Get current user [Auth]",
			"PUT /api/users/me/password":    "Change password [Auth]",
			"GET /api/users":                "List all users [Manager+]",
			"POST /api/users":               "Create user [Manager+]",
			"GET /api/users/:id":            "Get user by ID [Manager+]",
			"PUT /api/users/:id":            "Update user [Manager+]",
			"DELETE /api/users/:id":         "Deactivate user [Admin]",
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
		// Banner for console
		fmt.Println("\n==============================================")
		fmt.Printf("Server running at http://%s:%s\n", r.app.Config.Server.Host, r.app.Config.Server.Port)
		fmt.Println("Powered by Gin Framework + Google Wire DI")
		fmt.Println("Logger: Uber Zap (Structured)")
		fmt.Println("==============================================")
		fmt.Println("\nEndpoints:")
		fmt.Printf("  curl http://localhost:%s/health\n", r.app.Config.Server.Port)
		fmt.Printf("  curl http://localhost:%s/api/mon-an\n", r.app.Config.Server.Port)
		fmt.Println("\nPress Ctrl+C to stop...")

		logger.Info("HTTP server started",
			zap.String("addr", r.server.Addr),
		)

		if err := r.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Chờ signal hoặc error
	select {
	case err := <-serverErr:
		logger.Error("Server error", zap.Error(err))
		return fmt.Errorf("server error: %w", err)
	case sig := <-quit:
		logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
		return r.shutdown()
	}
}

// shutdown thực hiện graceful shutdown
func (r *Runner) shutdown() error {
	logger.Info("Starting graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), r.app.Config.Server.ShutdownTimeout)
	defer cancel()

	// Shutdown HTTP server
	logger.Info("Stopping HTTP server...")
	if err := r.server.Shutdown(ctx); err != nil {
		logger.Error("HTTP server shutdown error", zap.Error(err))
	} else {
		logger.Info("HTTP server stopped")
	}

	// Shutdown databases
	logger.Info("Closing database connections...")
	if err := r.app.DBManager.Shutdown(ctx); err != nil {
		logger.Error("Database shutdown error", zap.Error(err))
	} else {
		logger.Info("Database connections closed")
	}

	logger.Info("Graceful shutdown completed")
	return nil
}
