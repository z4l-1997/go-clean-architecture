// Package router chứa infrastructure để đăng ký routes
// Theo pattern từ Evrone go-clean-template và go-kit
package router

import "github.com/gin-gonic/gin"

// RouteRegistrar - interface chuẩn cho tất cả handlers
// Mỗi handler implement interface này để tự đăng ký routes
type RouteRegistrar interface {
	// BasePath trả về base path cho module (VD: "/mon-an", "/health")
	BasePath() string

	// RegisterRoutes đăng ký tất cả routes của module vào RouterGroup
	RegisterRoutes(rg *gin.RouterGroup)
}

// RouteGroup - config cho một nhóm routes
type RouteGroup struct {
	// Prefix là prefix cho nhóm (VD: "/api", "/admin")
	Prefix string

	// Middlewares áp dụng cho tất cả routes trong nhóm
	Middlewares []gin.HandlerFunc

	// Registrars là danh sách handlers cần đăng ký
	Registrars []RouteRegistrar
}
