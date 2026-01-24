// Package handler chứa HTTP Handlers - "bồi bàn" của hệ thống
// Handler nhận request từ client, gọi UseCase, và trả response
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"restaurant_project/internal/application/usecase"
	"restaurant_project/internal/presentation/http/dto"
)

// MonAnHandler xử lý các HTTP request liên quan đến MonAn
//
// VAI TRÒ CỦA HANDLER:
// - Giống như "bồi bàn" trong nhà hàng
// - Nhận order (HTTP request) từ khách (client)
// - Chuyển order đến bếp (UseCase)
// - Nhận món và mang ra cho khách (HTTP response)
// - KHÔNG xử lý business logic
type MonAnHandler struct {
	useCase *usecase.MonAnUseCase
}

// NewMonAnHandler tạo mới MonAnHandler
func NewMonAnHandler(uc *usecase.MonAnUseCase) *MonAnHandler {
	return &MonAnHandler{
		useCase: uc,
	}
}

// idCounter để tạo ID tự động (trong thực tế sẽ dùng UUID)
var idCounter = 0

// generateID tạo ID đơn giản (trong thực tế dùng UUID)
func generateID() string {
	idCounter++
	return string(rune('0'+idCounter)) + "_mon"
}

// ThemMon xử lý POST /api/mon-an - Thêm món mới
// @Summary Thêm món ăn mới
// @Description Thêm một món ăn mới vào menu
// @Tags MonAn
// @Accept json
// @Produce json
// @Param request body dto.ThemMonRequest true "Thông tin món ăn mới"
// @Success 201 {object} dto.APIResponse{data=dto.MonAnResponse} "Thêm món thành công"
// @Failure 400 {object} dto.APIResponse "Dữ liệu không hợp lệ"
// @Router /api/mon-an [post]
func (h *MonAnHandler) ThemMon(c *gin.Context) {
	// Bước 1: Parse JSON từ request body
	var req dto.ThemMonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest,
			dto.NewErrorResponse("Dữ liệu không hợp lệ", err))
		return
	}

	// Bước 2: Gọi UseCase để thêm món
	input := usecase.ThemMonInput{
		ID:   generateID(),
		Ten:  req.Ten,
		Gia:  req.Gia,
		MoTa: req.MoTa,
	}

	mon, err := h.useCase.ThemMon(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusBadRequest,
			dto.NewErrorResponse("Không thể thêm món", err))
		return
	}

	// Bước 3: Trả response
	c.JSON(http.StatusCreated,
		dto.NewSuccessResponse("Thêm món thành công", dto.ToMonAnResponse(mon)))
}

// XemMenu xử lý GET /api/mon-an - Xem tất cả món
// @Summary Xem menu
// @Description Lấy danh sách tất cả món ăn hoặc chỉ món còn hàng
// @Tags MonAn
// @Accept json
// @Produce json
// @Param con_hang query bool false "Chỉ lấy món còn hàng" default(false)
// @Success 200 {object} dto.APIResponse{data=[]dto.MonAnResponse} "Lấy menu thành công"
// @Failure 500 {object} dto.APIResponse "Lỗi server"
// @Router /api/mon-an [get]
func (h *MonAnHandler) XemMenu(c *gin.Context) {
	// Kiểm tra query param ?con_hang=true
	conHangOnly := c.Query("con_hang") == "true"

	var err error
	if conHangOnly {
		menu, e := h.useCase.XemMenuConHang(c.Request.Context())
		err = e
		if err == nil {
			responses := dto.ToMonAnResponseList(menu)
			c.JSON(http.StatusOK,
				dto.NewSuccessResponse("Lấy menu còn hàng thành công", responses))
			return
		}
	} else {
		menu, e := h.useCase.XemMenu(c.Request.Context())
		err = e
		if err == nil {
			responses := dto.ToMonAnResponseList(menu)
			c.JSON(http.StatusOK,
				dto.NewSuccessResponse("Lấy menu thành công", responses))
			return
		}
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError,
			dto.NewErrorResponse("Không thể lấy menu", err))
		return
	}
}

// TimMon xử lý GET /api/mon-an/:id - Tìm món theo ID
// @Summary Tìm món theo ID
// @Description Lấy thông tin chi tiết một món ăn theo ID
// @Tags MonAn
// @Accept json
// @Produce json
// @Param id path string true "ID món ăn"
// @Success 200 {object} dto.APIResponse{data=dto.MonAnResponse} "Tìm món thành công"
// @Failure 400 {object} dto.APIResponse "ID không hợp lệ"
// @Failure 404 {object} dto.APIResponse "Không tìm thấy món"
// @Router /api/mon-an/{id} [get]
func (h *MonAnHandler) TimMon(c *gin.Context) {
	// Lấy ID từ URL param
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest,
			dto.NewErrorResponse("ID không được để trống", nil))
		return
	}

	mon, err := h.useCase.TimMon(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound,
			dto.NewErrorResponse("Không tìm thấy món", err))
		return
	}

	c.JSON(http.StatusOK,
		dto.NewSuccessResponse("Tìm món thành công", dto.ToMonAnResponse(mon)))
}

// CapNhatGia xử lý PUT /api/mon-an/:id/gia - Cập nhật giá
// @Summary Cập nhật giá món ăn
// @Description Cập nhật giá mới cho một món ăn
// @Tags MonAn
// @Accept json
// @Produce json
// @Param id path string true "ID món ăn"
// @Param request body dto.CapNhatGiaRequest true "Giá mới"
// @Success 200 {object} dto.APIResponse{data=dto.MonAnResponse} "Cập nhật giá thành công"
// @Failure 400 {object} dto.APIResponse "Dữ liệu không hợp lệ"
// @Router /api/mon-an/{id}/gia [put]
func (h *MonAnHandler) CapNhatGia(c *gin.Context) {
	// Lấy ID từ URL param
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest,
			dto.NewErrorResponse("ID không được để trống", nil))
		return
	}

	// Parse request
	var req dto.CapNhatGiaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest,
			dto.NewErrorResponse("Dữ liệu không hợp lệ", err))
		return
	}

	// Gọi UseCase
	input := usecase.CapNhatGiaInput{
		ID:     id,
		GiaMoi: req.Gia,
	}

	mon, err := h.useCase.CapNhatGia(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusBadRequest,
			dto.NewErrorResponse("Không thể cập nhật giá", err))
		return
	}

	c.JSON(http.StatusOK,
		dto.NewSuccessResponse("Cập nhật giá thành công", dto.ToMonAnResponse(mon)))
}

// ApDungGiamGia xử lý PUT /api/mon-an/:id/giam-gia - Áp dụng giảm giá
// @Summary Áp dụng giảm giá
// @Description Áp dụng phần trăm giảm giá cho một món ăn
// @Tags MonAn
// @Accept json
// @Produce json
// @Param id path string true "ID món ăn"
// @Param request body dto.ApDungGiamGiaRequest true "Phần trăm giảm giá"
// @Success 200 {object} dto.APIResponse{data=dto.MonAnResponse} "Áp dụng giảm giá thành công"
// @Failure 400 {object} dto.APIResponse "Dữ liệu không hợp lệ"
// @Router /api/mon-an/{id}/giam-gia [put]
func (h *MonAnHandler) ApDungGiamGia(c *gin.Context) {
	// Lấy ID từ URL param
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest,
			dto.NewErrorResponse("ID không được để trống", nil))
		return
	}

	// Parse request
	var req dto.ApDungGiamGiaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest,
			dto.NewErrorResponse("Dữ liệu không hợp lệ", err))
		return
	}

	// Gọi UseCase
	input := usecase.ApDungGiamGiaInput{
		ID:       id,
		PhanTram: req.PhanTram,
	}

	mon, err := h.useCase.ApDungGiamGia(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusBadRequest,
			dto.NewErrorResponse("Không thể áp dụng giảm giá", err))
		return
	}

	c.JSON(http.StatusOK,
		dto.NewSuccessResponse("Áp dụng giảm giá thành công", dto.ToMonAnResponse(mon)))
}

// XoaMon xử lý DELETE /api/mon-an/:id - Xóa món
// @Summary Xóa món ăn
// @Description Xóa một món ăn khỏi menu
// @Tags MonAn
// @Accept json
// @Produce json
// @Param id path string true "ID món ăn"
// @Success 200 {object} dto.APIResponse "Xóa món thành công"
// @Failure 400 {object} dto.APIResponse "Không thể xóa món"
// @Router /api/mon-an/{id} [delete]
func (h *MonAnHandler) XoaMon(c *gin.Context) {
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest,
			dto.NewErrorResponse("ID không được để trống", nil))
		return
	}

	err := h.useCase.XoaMon(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest,
			dto.NewErrorResponse("Không thể xóa món", err))
		return
	}

	c.JSON(http.StatusOK,
		dto.NewSuccessResponse("Xóa món thành công", nil))
}

// DanhDauHetHang xử lý PUT /api/mon-an/:id/het-hang - Đánh dấu hết hàng
// @Summary Đánh dấu hết hàng
// @Description Đánh dấu một món ăn là hết hàng
// @Tags MonAn
// @Accept json
// @Produce json
// @Param id path string true "ID món ăn"
// @Success 200 {object} dto.APIResponse{data=dto.MonAnResponse} "Đánh dấu hết hàng thành công"
// @Failure 400 {object} dto.APIResponse "Không thể đánh dấu hết hàng"
// @Router /api/mon-an/{id}/het-hang [put]
func (h *MonAnHandler) DanhDauHetHang(c *gin.Context) {
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest,
			dto.NewErrorResponse("ID không được để trống", nil))
		return
	}

	mon, err := h.useCase.DanhDauHetHang(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest,
			dto.NewErrorResponse("Không thể đánh dấu hết hàng", err))
		return
	}

	c.JSON(http.StatusOK,
		dto.NewSuccessResponse("Đánh dấu hết hàng thành công", dto.ToMonAnResponse(mon)))
}

// ============================================================
// RouteRegistrar Interface Implementation
// ============================================================

// BasePath trả về base path cho MonAn module
func (h *MonAnHandler) BasePath() string {
	return "/mon-an"
}

// RegisterRoutes đăng ký tất cả routes của MonAn module
func (h *MonAnHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.XemMenu)
	rg.POST("", h.ThemMon)
	rg.GET("/:id", h.TimMon)
	rg.DELETE("/:id", h.XoaMon)
	rg.PUT("/:id/gia", h.CapNhatGia)
	rg.PUT("/:id/giam-gia", h.ApDungGiamGia)
	rg.PUT("/:id/het-hang", h.DanhDauHetHang)
}
