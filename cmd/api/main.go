// Package main là entry point của ứng dụng
package main

import (
	"fmt"
	"log"

	"restaurant_project/internal/app"
	"restaurant_project/internal/di"
)

func main() {
	fmt.Println("==============================================")
	fmt.Println("   RESTAURANT MENU API - Clean Architecture")
	fmt.Println("   Dependency Injection: Google Wire")
	fmt.Println("==============================================")

	// Initialize application với Wire DI
	diApp, err := di.InitializeApp()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// Create and run application
	runner := app.NewRunner(diApp)

	if err := runner.Run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
