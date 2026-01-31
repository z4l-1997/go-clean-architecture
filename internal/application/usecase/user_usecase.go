// Package usecase chứa Application Use Cases
package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"restaurant_project/internal/domain/entity"
	"restaurant_project/internal/domain/repository"
	"restaurant_project/internal/domain/service"
	"restaurant_project/pkg/logger"
	"restaurant_project/pkg/password"
)

// User use case errors
var (
	ErrUserNotFound       = errors.New("không tìm thấy user")
	ErrUsernameExists     = errors.New("username đã tồn tại")
	ErrEmailExists        = errors.New("email đã tồn tại")
	ErrInvalidRole        = errors.New("role không hợp lệ")
	ErrPermissionDenied   = errors.New("không có quyền thực hiện thao tác này")
	ErrCannotCreateAdmin  = errors.New("chỉ Admin mới có quyền tạo Admin")
	ErrCannotCreateManager = errors.New("chỉ Admin mới có quyền tạo Manager")
	ErrWrongPassword      = errors.New("mật khẩu cũ không đúng")
	ErrCannotDeactivateSelf = errors.New("không thể tự vô hiệu hóa tài khoản")
)

// CreateUserInput là input để tạo user mới
type CreateUserInput struct {
	Username    string
	Email       string
	Password    string
	Role        entity.UserRole
	CreatorRole entity.UserRole // Role của người tạo
}

// UpdateUserInput là input để cập nhật user
type UpdateUserInput struct {
	ID       string
	Email    *string
	IsActive *bool
}

// UserUseCase xử lý business logic liên quan đến User
type UserUseCase struct {
	repo           repository.IUserRepository
	tokenBlacklist service.TokenBlacklistService
}

// NewUserUseCase tạo mới UserUseCase
func NewUserUseCase(repo repository.IUserRepository, tokenBlacklist service.TokenBlacklistService) *UserUseCase {
	return &UserUseCase{
		repo:           repo,
		tokenBlacklist: tokenBlacklist,
	}
}

// CreateUser tạo user mới với kiểm tra quyền
func (uc *UserUseCase) CreateUser(ctx context.Context, input CreateUserInput) (*entity.User, error) {
	// Validate role permission
	if err := uc.validateCreatePermission(input.CreatorRole, input.Role); err != nil {
		return nil, err
	}

	// Check username exists
	exists, err := uc.repo.ExistsByUsername(ctx, input.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUsernameExists
	}

	// Check email exists
	exists, err = uc.repo.ExistsByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailExists
	}

	// Hash password
	passwordHash, err := password.Hash(input.Password)
	if err != nil {
		return nil, err
	}

	// Create user entity
	user, err := entity.NewUser(
		uuid.New().String(),
		input.Username,
		input.Email,
		passwordHash,
		input.Role,
	)
	if err != nil {
		return nil, err
	}

	// Save to repository
	if err := uc.repo.Save(ctx, user); err != nil {
		logger.CtxError(ctx, "failed to save new user",
			zap.String("target_username", input.Username),
			zap.Error(err),
		)
		return nil, err
	}

	logger.CtxInfo(ctx, "user created",
		zap.String("target_user_id", user.ID),
		zap.String("target_username", user.Username),
		zap.String("target_role", string(user.Role)),
		zap.String("creator_role", string(input.CreatorRole)),
	)

	return user, nil
}

// validateCreatePermission kiểm tra quyền tạo user theo role
// Permission Matrix:
// | Creator     | Admin | Manager | Staff | Customer |
// |-------------|-------|---------|-------|----------|
// | Admin       | YES   | YES     | YES   | YES      |
// | Manager     | NO    | NO      | YES   | YES      |
// | Staff       | NO    | NO      | NO    | NO       |
// | Customer    | NO    | NO      | NO    | NO       |
func (uc *UserUseCase) validateCreatePermission(creatorRole, targetRole entity.UserRole) error {
	switch targetRole {
	case entity.RoleAdmin:
		if creatorRole != entity.RoleAdmin {
			return ErrCannotCreateAdmin
		}
	case entity.RoleManager:
		if creatorRole != entity.RoleAdmin {
			return ErrCannotCreateManager
		}
	case entity.RoleStaff:
		if creatorRole != entity.RoleAdmin && creatorRole != entity.RoleManager {
			return ErrPermissionDenied
		}
	case entity.RoleCustomer:
		if creatorRole == entity.RoleCustomer || creatorRole == entity.RoleStaff {
			return ErrPermissionDenied
		}
	default:
		return ErrInvalidRole
	}
	return nil
}

// GetUserByID lấy user theo ID
func (uc *UserUseCase) GetUserByID(ctx context.Context, id string) (*entity.User, error) {
	user, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// GetAllUsers lấy tất cả users
func (uc *UserUseCase) GetAllUsers(ctx context.Context) ([]*entity.User, error) {
	return uc.repo.FindAll(ctx)
}

// GetAllUsersPaginated lấy tất cả users có phân trang
func (uc *UserUseCase) GetAllUsersPaginated(ctx context.Context, offset, limit int) ([]*entity.User, int64, error) {
	return uc.repo.FindAllPaginated(ctx, offset, limit)
}

// GetUsersByRole lấy users theo role
func (uc *UserUseCase) GetUsersByRole(ctx context.Context, role entity.UserRole) ([]*entity.User, error) {
	return uc.repo.FindByRole(ctx, role)
}

// GetUsersByRolePaginated lấy users theo role có phân trang
func (uc *UserUseCase) GetUsersByRolePaginated(ctx context.Context, role entity.UserRole, offset, limit int) ([]*entity.User, int64, error) {
	return uc.repo.FindByRolePaginated(ctx, role, offset, limit)
}

// UpdateUser cập nhật thông tin user
func (uc *UserUseCase) UpdateUser(ctx context.Context, input UpdateUserInput) (*entity.User, error) {
	user, err := uc.repo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Update email if provided
	if input.Email != nil {
		// Check if email already used by another user
		existingUser, err := uc.repo.FindByEmail(ctx, *input.Email)
		if err != nil {
			return nil, err
		}
		if existingUser != nil && existingUser.ID != user.ID {
			return nil, ErrEmailExists
		}
		user.Email = *input.Email
	}

	// Update is_active if provided
	if input.IsActive != nil {
		user.IsActive = *input.IsActive
	}

	user.NgayCapNhat = time.Now()

	if err := uc.repo.Save(ctx, user); err != nil {
		logger.CtxError(ctx, "failed to update user",
			zap.String("target_user_id", input.ID),
			zap.Error(err),
		)
		return nil, err
	}

	logger.CtxInfo(ctx, "user updated",
		zap.String("target_user_id", input.ID),
		zap.Bool("email_changed", input.Email != nil),
		zap.Bool("active_changed", input.IsActive != nil),
	)

	return user, nil
}

// DeactivateUser vô hiệu hóa user (soft delete)
func (uc *UserUseCase) DeactivateUser(ctx context.Context, id string, requestorID string) (*entity.User, error) {
	// Không cho phép tự vô hiệu hóa chính mình
	if id == requestorID {
		return nil, ErrCannotDeactivateSelf
	}

	user, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	user.Deactivate()

	if err := uc.repo.Save(ctx, user); err != nil {
		logger.CtxError(ctx, "failed to deactivate user",
			zap.String("target_user_id", id),
			zap.Error(err),
		)
		return nil, err
	}

	logger.CtxWarn(ctx, "user deactivated",
		zap.String("target_user_id", id),
		zap.String("target_username", user.Username),
		zap.String("requestor_id", requestorID),
	)

	return user, nil
}

// ChangePassword đổi mật khẩu user
func (uc *UserUseCase) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	user, err := uc.repo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Verify old password
	if !password.Verify(oldPassword, user.PasswordHash) {
		logger.CtxWarn(ctx, "password change failed: wrong old password",
			zap.String("user_id", userID),
		)
		return ErrWrongPassword
	}

	// Hash new password
	newHash, err := password.Hash(newPassword)
	if err != nil {
		return err
	}

	// Update password
	if err := user.UpdatePassword(newHash); err != nil {
		return err
	}

	if err := uc.repo.Save(ctx, user); err != nil {
		logger.CtxError(ctx, "failed to save password change",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return err
	}

	// Revoke tất cả tokens sau khi đổi password thành công (fail-open)
	if err := uc.tokenBlacklist.RevokeAllUserTokens(ctx, userID); err != nil {
		logger.CtxWarn(ctx, "failed to revoke tokens after password change",
			zap.String("user_id", userID),
			zap.Error(err),
		)
	}

	logger.CtxInfo(ctx, "password changed",
		zap.String("user_id", userID),
	)

	return nil
}
