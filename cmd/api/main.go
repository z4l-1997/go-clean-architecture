// Package main là entry point của ứng dụng
// @title Restaurant Menu API
// @version 2.0.0
// @description API quản lý menu nhà hàng - Xây dựng theo Clean Architecture
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@restaurant.local

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /
// @schemes http

package main

import (
	"fmt"

	"go.uber.org/zap"

	"restaurant_project/internal/app"
	"restaurant_project/internal/di"
	"restaurant_project/internal/infrastructure/config"
	"restaurant_project/pkg/logger"

	_ "restaurant_project/docs" // Swagger docs
)

func main() {
	// 1. Load config TRƯỚC
	cfg := config.Load()

	// 2. Khởi tạo Structured Logger với File Rotation
	if err := logger.Init(logger.Config{
		Level:       cfg.Log.Level,
		Environment: cfg.Log.Environment,
		ServiceName: "restaurant-api",
		Version:     "2.0.0",
		// File logging config
		EnableFileLog: cfg.Log.EnableFileLog,
		LogFilePath:   cfg.Log.LogFilePath,
		MaxSizeMB:     cfg.Log.MaxSizeMB,
		MaxBackups:    cfg.Log.MaxBackups,
		MaxAgeDays:    cfg.Log.MaxAgeDays,
		CompressLog:   cfg.Log.CompressLog,
	}); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Sync() // Flush logs trước khi exit

	// Banner
	fmt.Println("==============================================")
	fmt.Println("   RESTAURANT MENU API - Clean Architecture")
	fmt.Println("   Dependency Injection: Google Wire")
	fmt.Println("   Logger: Uber Zap (Structured)")
	fmt.Println("==============================================")

	logger.Info("Starting Restaurant API",
		zap.String("environment", cfg.Log.Environment),
		zap.String("log_level", cfg.Log.Level),
	)

	// 3. Initialize application với Wire DI
	diApp, err := di.InitializeApp()
	if err != nil {
		logger.Fatal("Failed to initialize application", zap.Error(err))
	}

	logger.Info("Application initialized successfully")

	// 4. Create and run application
	runner := app.NewRunner(diApp)

	if err := runner.Run(); err != nil {
		logger.Fatal("Application error", zap.Error(err))
	}
}
