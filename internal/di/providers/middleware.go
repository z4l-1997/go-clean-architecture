// Package providers chứa các Wire providers
package providers

import (
	"github.com/gin-gonic/gin"

	"restaurant_project/internal/infrastructure/config"
	"restaurant_project/internal/infrastructure/middleware"
)

// MiddlewareCollection chứa tất cả middleware đã được khởi tạo
type MiddlewareCollection struct {
	CORS            gin.HandlerFunc
	RateLimit       gin.HandlerFunc
	JWTAuth         *middleware.JWTAuthMiddleware
	Timeout         gin.HandlerFunc
	Gzip            gin.HandlerFunc
	SecurityHeaders gin.HandlerFunc
	BodySizeLimit   gin.HandlerFunc
	ErrorHandler    gin.HandlerFunc
}

// ProvideMiddlewareCollection tạo MiddlewareCollection từ config
func ProvideMiddlewareCollection(cfg *config.Config) *MiddlewareCollection {
	return &MiddlewareCollection{
		CORS:            middleware.CORS(cfg.Middleware.CORS),
		RateLimit:       middleware.RateLimit(cfg.Middleware.RateLimit),
		JWTAuth:         middleware.NewJWTAuth(cfg.Middleware.JWT),
		Timeout:         middleware.Timeout(cfg.Middleware.Timeout),
		Gzip:            middleware.Gzip(cfg.Middleware.Gzip),
		SecurityHeaders: middleware.SecurityHeaders(cfg.Middleware.Security),
		BodySizeLimit:   middleware.BodySizeLimit(cfg.Middleware.BodyLimit),
		ErrorHandler:    middleware.ErrorHandler(),
	}
}

// ProvideMiddlewareConfig trích xuất MiddlewareConfig từ Config
func ProvideMiddlewareConfig(cfg *config.Config) config.MiddlewareConfig {
	return cfg.Middleware
}
