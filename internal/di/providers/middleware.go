// Package providers chứa các Wire providers
package providers

import (
	"github.com/gin-gonic/gin"

	"restaurant_project/internal/domain/service"
	"restaurant_project/internal/infrastructure/config"
	"restaurant_project/internal/infrastructure/middleware"
)

// MiddlewareCollection chứa tất cả middleware đã được khởi tạo
type MiddlewareCollection struct {
	CORS            gin.HandlerFunc
	RateLimit       gin.HandlerFunc
	AuthRateLimit   gin.HandlerFunc
	JWTAuth         *middleware.JWTAuthMiddleware
	Timeout         gin.HandlerFunc
	Gzip            gin.HandlerFunc
	SecurityHeaders gin.HandlerFunc
	BodySizeLimit   gin.HandlerFunc
	ErrorHandler    gin.HandlerFunc
}

// ProvideJWTAuth tạo JWTAuthMiddleware từ config
func ProvideJWTAuth(cfg *config.Config, blacklistService service.TokenBlacklistService) *middleware.JWTAuthMiddleware {
	return middleware.NewJWTAuth(cfg.Middleware.JWT, blacklistService)
}

// ProvideMiddlewareCollection tạo MiddlewareCollection từ config
func ProvideMiddlewareCollection(cfg *config.Config, jwtAuth *middleware.JWTAuthMiddleware) *MiddlewareCollection {
	return &MiddlewareCollection{
		CORS:            middleware.CORS(cfg.Middleware.CORS),
		RateLimit:       middleware.RateLimit(cfg.Middleware.RateLimit),
		AuthRateLimit:   middleware.AuthRateLimit(cfg.Middleware.AuthRateLimit),
		JWTAuth:         jwtAuth,
		Timeout:         middleware.Timeout(cfg.Middleware.Timeout),
		Gzip:            middleware.Gzip(cfg.Middleware.Gzip),
		SecurityHeaders: middleware.SecurityHeaders(cfg.Middleware.Security),
		BodySizeLimit:   middleware.BodySizeLimit(cfg.Middleware.BodyLimit),
		ErrorHandler:    middleware.ErrorHandler(),
	}
}
