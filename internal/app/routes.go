package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// registerRoutes đăng ký tất cả HTTP routes với Gin
func (a *App) registerRoutes(router *gin.Engine) {
	// Health check routes
	router.GET("/health", a.healthHandler.Health)
	router.GET("/health/live", a.healthHandler.Liveness)
	router.GET("/health/ready", a.healthHandler.Readiness)

	// API routes group
	api := router.Group("/api")
	{
		// MonAn routes
		monAn := api.Group("/mon-an")
		{
			monAn.GET("", a.monAnHandler.XemMenu)
			monAn.POST("", a.monAnHandler.ThemMon)
			monAn.GET("/:id", a.monAnHandler.TimMon)
			monAn.DELETE("/:id", a.monAnHandler.XoaMon)
			monAn.PUT("/:id/gia", a.monAnHandler.CapNhatGia)
			monAn.PUT("/:id/giam-gia", a.monAnHandler.ApDungGiamGia)
			monAn.PUT("/:id/het-hang", a.monAnHandler.DanhDauHetHang)
		}
	}

	// Root route
	router.GET("/", a.handleRoot)
}

// handleRoot xử lý root endpoint
func (a *App) handleRoot(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Restaurant Menu API",
		"version": "2.0.0",
		"endpoints": gin.H{
			"GET /health":                "Full health check",
			"GET /health/live":           "Liveness probe",
			"GET /health/ready":          "Readiness probe",
			"GET /api/mon-an":            "List all dishes",
			"GET /api/mon-an?con_hang=true": "List available dishes",
			"GET /api/mon-an/:id":        "Get dish by ID",
			"POST /api/mon-an":           "Create new dish",
			"PUT /api/mon-an/:id/gia":    "Update price",
			"PUT /api/mon-an/:id/giam-gia": "Apply discount",
			"PUT /api/mon-an/:id/het-hang": "Mark as out of stock",
			"DELETE /api/mon-an/:id":     "Delete dish",
		},
	})
}
