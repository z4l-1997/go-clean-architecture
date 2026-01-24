package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"restaurant_project/internal/infrastructure/database"
)

// HealthHandler xử lý các health check endpoints
type HealthHandler struct {
	dbManager *database.DBManager
	startTime time.Time
}

// NewHealthHandler tạo instance mới
func NewHealthHandler(dbManager *database.DBManager) *HealthHandler {
	return &HealthHandler{
		dbManager: dbManager,
		startTime: time.Now(),
	}
}

// HealthResponse là response cho health check
type HealthResponse struct {
	Status    string                            `json:"status"`
	Timestamp time.Time                         `json:"timestamp"`
	Uptime    string                            `json:"uptime"`
	Services  map[string]database.HealthStatus  `json:"services,omitempty"`
}

// LivenessResponse cho Kubernetes liveness probe
type LivenessResponse struct {
	Status string `json:"status"`
}

// ReadinessResponse cho Kubernetes readiness probe
type ReadinessResponse struct {
	Status string `json:"status"`
	Ready  bool   `json:"ready"`
}

// Health xử lý GET /health - Full health check
// @Summary Full health check
// @Description Kiểm tra sức khỏe tổng thể của hệ thống bao gồm tất cả services
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse "Hệ thống healthy"
// @Failure 503 {object} HealthResponse "Hệ thống unhealthy"
// @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	ctx := c.Request.Context()

	// Lấy health status của tất cả services
	services := h.dbManager.HealthCheckAll(ctx)

	// Xác định status tổng thể
	status := "healthy"
	for _, svc := range services {
		if !svc.Healthy {
			status = "unhealthy"
			break
		}
	}

	response := HealthResponse{
		Status:    status,
		Timestamp: time.Now(),
		Uptime:    time.Since(h.startTime).String(),
		Services:  services,
	}

	if status == "unhealthy" {
		c.JSON(http.StatusServiceUnavailable, response)
	} else {
		c.JSON(http.StatusOK, response)
	}
}

// Liveness xử lý GET /health/live - Kubernetes liveness probe
// Chỉ kiểm tra app có đang chạy không (không kiểm tra dependencies)
// @Summary Liveness probe
// @Description Kubernetes liveness probe - kiểm tra app có đang chạy không
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} LivenessResponse "App đang chạy"
// @Router /health/live [get]
func (h *HealthHandler) Liveness(c *gin.Context) {
	response := LivenessResponse{
		Status: "alive",
	}

	c.JSON(http.StatusOK, response)
}

// Readiness xử lý GET /health/ready - Kubernetes readiness probe
// Kiểm tra app có sẵn sàng nhận traffic không
// @Summary Readiness probe
// @Description Kubernetes readiness probe - kiểm tra app có sẵn sàng nhận traffic không
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} ReadinessResponse "App sẵn sàng"
// @Failure 503 {object} ReadinessResponse "App chưa sẵn sàng"
// @Router /health/ready [get]
func (h *HealthHandler) Readiness(c *gin.Context) {
	ctx := c.Request.Context()

	ready := h.dbManager.IsAllHealthy(ctx)

	response := ReadinessResponse{
		Ready: ready,
	}

	if ready {
		response.Status = "ready"
		c.JSON(http.StatusOK, response)
	} else {
		response.Status = "not_ready"
		c.JSON(http.StatusServiceUnavailable, response)
	}
}

// ============================================================
// RouteRegistrar Interface Implementation
// ============================================================

// BasePath trả về base path cho Health module
func (h *HealthHandler) BasePath() string {
	return "/health"
}

// RegisterRoutes đăng ký tất cả routes của Health module
func (h *HealthHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.Health)
	rg.GET("/live", h.Liveness)
	rg.GET("/ready", h.Readiness)
}
