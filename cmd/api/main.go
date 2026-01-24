// Package main là entry point của ứng dụng
package main

import (
	"fmt"
	"log"

	"restaurant_project/internal/app"
	"restaurant_project/internal/infrastructure/config"
)

func main() {
	fmt.Println("==============================================")
	fmt.Println("   RESTAURANT MENU API - Clean Architecture")
	fmt.Println("==============================================\n")

	// Load configuration
	cfg := config.Load()

	// Create and run application
	application := app.New(cfg)

	if err := application.Run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
