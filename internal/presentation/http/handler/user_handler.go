// Package handler chứa HTTP Handlers
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"restaurant_project/internal/application/usecase"
	"restaurant_project/internal/domain/entity"
	"restaurant_project/internal/infrastructure/middleware"
	"restaurant_project/internal/presentation/http/dto"
)

// UserHandler xử lý các HTTP request liên quan đến User
type UserHandler struct {
	useCase *usecase.UserUseCase
	jwtAuth *middleware.JWTAuthMiddleware
}

// NewUserHandler tạo mới UserHandler
func NewUserHandler(uc *usecase.UserUseCase, jwtAuth *middleware.JWTAuthMiddleware) *UserHandler {
	return &UserHandler{
		useCase: uc,
		jwtAuth: jwtAuth,
	}
}

// GetMe xử lý GET /api/users/me - Lấy thông tin user hiện tại
// @Summary Lấy thông tin user hiện tại
// @Description Lấy thông tin của user đang đăng nhập
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.APIResponse{data=dto.UserResponse}
// @Failure 401 {object} dto.APIResponse
// @Router /api/users/me [get]
func (h *UserHandler) GetMe(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized,
			dto.NewErrorResponse("Unauthorized", nil))
		return
	}

	user, err := h.useCase.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound,
			dto.NewErrorResponse("User không tồn tại", err))
		return
	}

	c.JSON(http.StatusOK,
		dto.NewSuccessResponse("Lấy thông tin user thành công", dto.ToUserResponse(user)))
}

// ChangePassword xử lý PUT /api/users/me/password - Đổi mật khẩu
// @Summary Đổi mật khẩu
// @Description Đổi mật khẩu của user đang đăng nhập
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.ChangePasswordRequest true "Thông tin đổi mật khẩu"
// @Success 200 {object} dto.APIResponse
// @Failure 400 {object} dto.APIResponse
// @Router /api/users/me/password [put]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized,
			dto.NewErrorResponse("Unauthorized", nil))
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest,
			dto.NewErrorResponse("Dữ liệu không hợp lệ", err))
		return
	}

	err := h.useCase.ChangePassword(c.Request.Context(), userID, req.OldPassword, req.NewPassword)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err == usecase.ErrWrongPassword {
			statusCode = http.StatusUnauthorized
		}
		c.JSON(statusCode,
			dto.NewErrorResponse("Không thể đổi mật khẩu", err))
		return
	}

	c.JSON(http.StatusOK,
		dto.NewSuccessResponse("Đổi mật khẩu thành công", nil))
}

// GetUsers xử lý GET /api/users - Lấy danh sách users
// @Summary Lấy danh sách users
// @Description Lấy danh sách tất cả users có phân trang (Manager+)
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param role query string false "Filter by role"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 20, max: 100)"
// @Success 200 {object} dto.APIResponse{data=dto.PaginatedResponse}
// @Failure 403 {object} dto.APIResponse
// @Router /api/users [get]
func (h *UserHandler) GetUsers(c *gin.Context) {
	var pagination dto.PaginationRequest
	if err := c.ShouldBindQuery(&pagination); err != nil {
		c.JSON(http.StatusBadRequest,
			dto.NewErrorResponse("Tham số phân trang không hợp lệ", err))
		return
	}

	roleFilter := c.Query("role")
	ctx := c.Request.Context()

	var users []*entity.User
	var total int64
	var err error

	if roleFilter != "" {
		users, total, err = h.useCase.GetUsersByRolePaginated(ctx, entity.UserRole(roleFilter), pagination.Offset(), pagination.Limit)
	} else {
		users, total, err = h.useCase.GetAllUsersPaginated(ctx, pagination.Offset(), pagination.Limit)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError,
			dto.NewErrorResponse("Không thể lấy danh sách users", err))
		return
	}

	c.JSON(http.StatusOK,
		dto.NewSuccessResponse("Lấy danh sách users thành công",
			dto.NewPaginatedResponse(dto.ToUserResponseList(users), total, pagination.Page, pagination.Limit)))
}

// CreateUser xử lý POST /api/users - Tạo user mới
// @Summary Tạo user mới
// @Description Tạo user mới (Manager+)
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateUserRequest true "Thông tin user mới"
// @Success 201 {object} dto.APIResponse{data=dto.UserResponse}
// @Failure 400 {object} dto.APIResponse
// @Failure 403 {object} dto.APIResponse
// @Router /api/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	creatorRole, exists := middleware.GetUserRole(c)
	if !exists {
		c.JSON(http.StatusUnauthorized,
			dto.NewErrorResponse("Unauthorized", nil))
		return
	}

	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest,
			dto.NewErrorResponse("Dữ liệu không hợp lệ", err))
		return
	}

	input := usecase.CreateUserInput{
		Username:    req.Username,
		Email:       req.Email,
		Password:    req.Password,
		Role:        entity.UserRole(req.Role),
		CreatorRole: entity.UserRole(creatorRole),
	}

	user, err := h.useCase.CreateUser(c.Request.Context(), input)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err == usecase.ErrPermissionDenied ||
			err == usecase.ErrCannotCreateAdmin ||
			err == usecase.ErrCannotCreateManager {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode,
			dto.NewErrorResponse("Không thể tạo user", err))
		return
	}

	c.JSON(http.StatusCreated,
		dto.NewSuccessResponse("Tạo user thành công", dto.ToUserResponse(user)))
}

// GetUser xử lý GET /api/users/:id - Lấy user theo ID
// @Summary Lấy user theo ID
// @Description Lấy thông tin user theo ID (Manager+)
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} dto.APIResponse{data=dto.UserResponse}
// @Failure 404 {object} dto.APIResponse
// @Router /api/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest,
			dto.NewErrorResponse("ID không được để trống", nil))
		return
	}

	user, err := h.useCase.GetUserByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound,
			dto.NewErrorResponse("User không tồn tại", err))
		return
	}

	c.JSON(http.StatusOK,
		dto.NewSuccessResponse("Lấy user thành công", dto.ToUserResponse(user)))
}

// UpdateUser xử lý PUT /api/users/:id - Cập nhật user
// @Summary Cập nhật user
// @Description Cập nhật thông tin user (Manager+)
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param request body dto.UpdateUserRequest true "Thông tin cập nhật"
// @Success 200 {object} dto.APIResponse{data=dto.UserResponse}
// @Failure 400 {object} dto.APIResponse
// @Router /api/users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest,
			dto.NewErrorResponse("ID không được để trống", nil))
		return
	}

	callerRole, exists := middleware.GetUserRole(c)
	if !exists {
		c.JSON(http.StatusUnauthorized,
			dto.NewErrorResponse("Unauthorized", nil))
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest,
			dto.NewErrorResponse("Dữ liệu không hợp lệ", err))
		return
	}

	input := usecase.UpdateUserInput{
		ID:         id,
		Email:      req.Email,
		IsActive:   req.IsActive,
		CallerRole: entity.UserRole(callerRole),
	}

	user, err := h.useCase.UpdateUser(c.Request.Context(), input)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err == usecase.ErrPermissionDenied {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode,
			dto.NewErrorResponse("Không thể cập nhật user", err))
		return
	}

	c.JSON(http.StatusOK,
		dto.NewSuccessResponse("Cập nhật user thành công", dto.ToUserResponse(user)))
}

// DeactivateUser xử lý DELETE /api/users/:id - Vô hiệu hóa user
// @Summary Vô hiệu hóa user
// @Description Vô hiệu hóa user (Admin only)
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} dto.APIResponse{data=dto.UserResponse}
// @Failure 400 {object} dto.APIResponse
// @Failure 403 {object} dto.APIResponse
// @Router /api/users/{id} [delete]
func (h *UserHandler) DeactivateUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest,
			dto.NewErrorResponse("ID không được để trống", nil))
		return
	}

	requestorID, _ := middleware.GetUserID(c)

	user, err := h.useCase.DeactivateUser(c.Request.Context(), id, requestorID)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err == usecase.ErrCannotDeactivateSelf {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode,
			dto.NewErrorResponse("Không thể vô hiệu hóa user", err))
		return
	}

	c.JSON(http.StatusOK,
		dto.NewSuccessResponse("Vô hiệu hóa user thành công", dto.ToUserResponse(user)))
}

// ============================================================
// RouteRegistrar Interface Implementation
// ============================================================

// BasePath trả về base path cho User module
func (h *UserHandler) BasePath() string {
	return "/users"
}

// RegisterRoutes đăng ký tất cả routes của User module
// Note: Middleware JWT đã được áp dụng ở cấp group trong app.go
func (h *UserHandler) RegisterRoutes(rg *gin.RouterGroup) {
	// Authenticated routes - tất cả user đã đăng nhập
	rg.GET("/me", h.GetMe)
	rg.PUT("/me/password", h.ChangePassword)

	// Manager+ routes
	rg.GET("", middleware.RequireMinRole(middleware.RoleManager), h.GetUsers)
	rg.POST("", middleware.RequireMinRole(middleware.RoleManager), h.CreateUser)
	rg.GET("/:id", middleware.RequireMinRole(middleware.RoleManager), h.GetUser)
	rg.PUT("/:id", middleware.RequireMinRole(middleware.RoleManager), h.UpdateUser)

	// Admin only
	rg.DELETE("/:id", middleware.RequireRole(middleware.RoleAdmin), h.DeactivateUser)
}
