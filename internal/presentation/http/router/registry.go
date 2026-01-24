package router

import "github.com/gin-gonic/gin"

// Registry quản lý việc đăng ký routes theo groups
// Pattern: Explicit wiring giống Uber Go Style Guide
type Registry struct {
	groups []RouteGroup
}

// NewRegistry tạo một Registry mới
func NewRegistry() *Registry {
	return &Registry{
		groups: make([]RouteGroup, 0),
	}
}

// AddGroup thêm một RouteGroup vào registry
// Hỗ trợ method chaining: registry.AddGroup(g1).AddGroup(g2)
func (r *Registry) AddGroup(g RouteGroup) *Registry {
	r.groups = append(r.groups, g)
	return r
}

// Register đăng ký tất cả groups vào Gin Engine
func (r *Registry) Register(engine *gin.Engine) {
	for _, g := range r.groups {
		// Tạo group với prefix và middlewares
		rg := engine.Group(g.Prefix, g.Middlewares...)

		// Đăng ký từng module handler
		for _, reg := range g.Registrars {
			// Tạo sub-group cho mỗi module
			moduleGroup := rg.Group(reg.BasePath())
			reg.RegisterRoutes(moduleGroup)
		}
	}
}
