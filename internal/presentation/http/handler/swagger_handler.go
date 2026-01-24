// Package handler chứa HTTP Handlers
package handler

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SwaggerHandler xử lý Swagger UI endpoints
type SwaggerHandler struct{}

// NewSwaggerHandler tạo SwaggerHandler mới
func NewSwaggerHandler() *SwaggerHandler {
	return &SwaggerHandler{}
}

// BasePath trả về base path cho Swagger module
func (h *SwaggerHandler) BasePath() string {
	return "/swagger"
}

// RegisterRoutes đăng ký Swagger routes (cho RouterGroup)
func (h *SwaggerHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

// RegisterRoutesOnEngine đăng ký Swagger routes trực tiếp trên Engine
// Cần thiết vì gin-swagger yêu cầu path pattern "/swagger/*any"
func (h *SwaggerHandler) RegisterRoutesOnEngine(router *gin.Engine) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
