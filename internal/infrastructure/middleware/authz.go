package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"restaurant_project/pkg/logger"
)

// Role constants
const (
	RoleAdmin   = "admin"
	RoleManager = "manager"
	RoleStaff   = "staff"
	RoleUser    = "user"
)

// roleHierarchy định nghĩa thứ tự quyền hạn
// Số càng cao = quyền càng lớn
var roleHierarchy = map[string]int{
	RoleUser:    1,
	RoleStaff:   2,
	RoleManager: 3,
	RoleAdmin:   4,
}

// RequireAuth middleware yêu cầu user phải đăng nhập
// Sử dụng sau JWTAuth middleware
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := GetUserID(c)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":      "Authentication required",
				"code":       "AUTH_REQUIRED",
				"request_id": logger.GetRequestID(c),
			})
			return
		}
		c.Next()
	}
}

// RequireRole middleware yêu cầu user phải có role cụ thể
// Sử dụng: router.Use(RequireRole("admin"))
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := GetUserRole(c)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":      "Authentication required",
				"code":       "AUTH_REQUIRED",
				"request_id": logger.GetRequestID(c),
			})
			return
		}

		// Kiểm tra role
		for _, role := range roles {
			if userRole == role {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error":          "Insufficient permissions",
			"code":           "FORBIDDEN",
			"required_roles": roles,
			"your_role":      userRole,
			"request_id":     logger.GetRequestID(c),
		})
	}
}

// RequireMinRole middleware yêu cầu user phải có role tối thiểu
// Sử dụng: router.Use(RequireMinRole("manager")) - manager và admin đều được phép
func RequireMinRole(minRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := GetUserRole(c)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":      "Authentication required",
				"code":       "AUTH_REQUIRED",
				"request_id": logger.GetRequestID(c),
			})
			return
		}

		userLevel, userOK := roleHierarchy[userRole]
		minLevel, minOK := roleHierarchy[minRole]

		if !userOK || !minOK {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":      "Invalid role configuration",
				"code":       "INVALID_ROLE",
				"request_id": logger.GetRequestID(c),
			})
			return
		}

		if userLevel < minLevel {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":        "Insufficient permissions",
				"code":         "FORBIDDEN",
				"min_required": minRole,
				"your_role":    userRole,
				"request_id":   logger.GetRequestID(c),
			})
			return
		}

		c.Next()
	}
}

// RequireOwnerOrRole middleware cho phép nếu là chủ sở hữu hoặc có role đủ quyền
// ownerIDParam là tên param chứa owner ID (ví dụ: "user_id", "id")
func RequireOwnerOrRole(ownerIDParam string, roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserID(c)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":      "Authentication required",
				"code":       "AUTH_REQUIRED",
				"request_id": logger.GetRequestID(c),
			})
			return
		}

		// Kiểm tra có phải chủ sở hữu không
		ownerID := c.Param(ownerIDParam)
		if ownerID != "" && ownerID == userID {
			c.Next()
			return
		}

		// Kiểm tra role
		userRole, _ := GetUserRole(c)
		for _, role := range roles {
			if userRole == role {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error":      "Insufficient permissions",
			"code":       "FORBIDDEN",
			"request_id": logger.GetRequestID(c),
		})
	}
}

// IsAdmin helper kiểm tra user có phải admin không
func IsAdmin(c *gin.Context) bool {
	role, exists := GetUserRole(c)
	return exists && role == RoleAdmin
}

// IsManager helper kiểm tra user có phải manager trở lên không
func IsManager(c *gin.Context) bool {
	role, exists := GetUserRole(c)
	if !exists {
		return false
	}
	level, ok := roleHierarchy[role]
	return ok && level >= roleHierarchy[RoleManager]
}

// IsStaff helper kiểm tra user có phải staff trở lên không
func IsStaff(c *gin.Context) bool {
	role, exists := GetUserRole(c)
	if !exists {
		return false
	}
	level, ok := roleHierarchy[role]
	return ok && level >= roleHierarchy[RoleStaff]
}
