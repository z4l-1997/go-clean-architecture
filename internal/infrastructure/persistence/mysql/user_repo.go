// Package mysql chứa các MySQL repository implementations
package mysql

import (
	"context"
	"database/sql"
	"errors"

	mysqldriver "github.com/go-sql-driver/mysql"

	"restaurant_project/internal/domain/entity"
	"restaurant_project/internal/domain/repository"
)

// UserMySQLRepo là implementation của IUserRepository sử dụng MySQL
type UserMySQLRepo struct {
	db *sql.DB
}

// NewUserMySQLRepo tạo mới UserMySQLRepo
func NewUserMySQLRepo(db *sql.DB) *UserMySQLRepo {
	return &UserMySQLRepo{db: db}
}

// Verify interface implementation at compile time
var _ repository.IUserRepository = (*UserMySQLRepo)(nil)

// FindByID tìm user theo ID
func (r *UserMySQLRepo) FindByID(ctx context.Context, id string) (*entity.User, error) {
	query := `SELECT id, username, email, password_hash, role, is_active, is_email_verified, ngay_tao, ngay_cap_nhat
			  FROM users WHERE id = ?`

	user := &entity.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.Role, &user.IsActive, &user.IsEmailVerified, &user.NgayTao, &user.NgayCapNhat,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// FindByUsername tìm user theo username
func (r *UserMySQLRepo) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	query := `SELECT id, username, email, password_hash, role, is_active, is_email_verified, ngay_tao, ngay_cap_nhat
			  FROM users WHERE username = ?`

	user := &entity.User{}
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.Role, &user.IsActive, &user.IsEmailVerified, &user.NgayTao, &user.NgayCapNhat,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// FindByEmail tìm user theo email
func (r *UserMySQLRepo) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `SELECT id, username, email, password_hash, role, is_active, is_email_verified, ngay_tao, ngay_cap_nhat
			  FROM users WHERE email = ?`

	user := &entity.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.Role, &user.IsActive, &user.IsEmailVerified, &user.NgayTao, &user.NgayCapNhat,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// FindAll lấy tất cả users
func (r *UserMySQLRepo) FindAll(ctx context.Context) ([]*entity.User, error) {
	query := `SELECT id, username, email, password_hash, role, is_active, is_email_verified, ngay_tao, ngay_cap_nhat
			  FROM users ORDER BY ngay_tao DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*entity.User
	for rows.Next() {
		user := &entity.User{}
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.PasswordHash,
			&user.Role, &user.IsActive, &user.IsEmailVerified, &user.NgayTao, &user.NgayCapNhat,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

// FindAllPaginated lấy users có phân trang
func (r *UserMySQLRepo) FindAllPaginated(ctx context.Context, offset, limit int) ([]*entity.User, int64, error) {
	// Count total
	var total int64
	countQuery := `SELECT COUNT(*) FROM users`
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Fetch paginated
	query := `SELECT id, username, email, password_hash, role, is_active, is_email_verified, ngay_tao, ngay_cap_nhat
			  FROM users ORDER BY ngay_tao DESC LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*entity.User
	for rows.Next() {
		user := &entity.User{}
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.PasswordHash,
			&user.Role, &user.IsActive, &user.IsEmailVerified, &user.NgayTao, &user.NgayCapNhat,
		)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}

	return users, total, rows.Err()
}

// FindByRolePaginated lấy users theo role có phân trang
func (r *UserMySQLRepo) FindByRolePaginated(ctx context.Context, role entity.UserRole, offset, limit int) ([]*entity.User, int64, error) {
	// Count total
	var total int64
	countQuery := `SELECT COUNT(*) FROM users WHERE role = ?`
	if err := r.db.QueryRowContext(ctx, countQuery, role).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Fetch paginated
	query := `SELECT id, username, email, password_hash, role, is_active, is_email_verified, ngay_tao, ngay_cap_nhat
			  FROM users WHERE role = ? ORDER BY ngay_tao DESC LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, query, role, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*entity.User
	for rows.Next() {
		user := &entity.User{}
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.PasswordHash,
			&user.Role, &user.IsActive, &user.IsEmailVerified, &user.NgayTao, &user.NgayCapNhat,
		)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}

	return users, total, rows.Err()
}

// FindByRole lấy users theo role
func (r *UserMySQLRepo) FindByRole(ctx context.Context, role entity.UserRole) ([]*entity.User, error) {
	query := `SELECT id, username, email, password_hash, role, is_active, is_email_verified, ngay_tao, ngay_cap_nhat
			  FROM users WHERE role = ? ORDER BY ngay_tao DESC`

	rows, err := r.db.QueryContext(ctx, query, role)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*entity.User
	for rows.Next() {
		user := &entity.User{}
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.PasswordHash,
			&user.Role, &user.IsActive, &user.IsEmailVerified, &user.NgayTao, &user.NgayCapNhat,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

// Create tạo user mới, trả lỗi nếu trùng unique constraint
func (r *UserMySQLRepo) Create(ctx context.Context, user *entity.User) error {
	query := `INSERT INTO users (id, username, email, password_hash, role, is_active, is_email_verified, ngay_tao, ngay_cap_nhat)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Username, user.Email, user.PasswordHash,
		user.Role, user.IsActive, user.IsEmailVerified, user.NgayTao, user.NgayCapNhat,
	)
	if err != nil {
		var mysqlErr *mysqldriver.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return repository.ErrDuplicateEntry
		}
		return err
	}
	return nil
}

// Save cập nhật user đã tồn tại
func (r *UserMySQLRepo) Save(ctx context.Context, user *entity.User) error {
	query := `INSERT INTO users (id, username, email, password_hash, role, is_active, is_email_verified, ngay_tao, ngay_cap_nhat)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			  ON DUPLICATE KEY UPDATE
			  username = VALUES(username),
			  email = VALUES(email),
			  password_hash = VALUES(password_hash),
			  role = VALUES(role),
			  is_active = VALUES(is_active),
			  is_email_verified = VALUES(is_email_verified),
			  ngay_cap_nhat = VALUES(ngay_cap_nhat)`

	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Username, user.Email, user.PasswordHash,
		user.Role, user.IsActive, user.IsEmailVerified, user.NgayTao, user.NgayCapNhat,
	)

	return err
}

// Delete xóa user theo ID
func (r *UserMySQLRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("không tìm thấy user để xóa")
	}

	return nil
}

// ExistsByUsername kiểm tra username đã tồn tại chưa
func (r *UserMySQLRepo) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, username).Scan(&exists)
	return exists, err
}

// ExistsByEmail kiểm tra email đã tồn tại chưa
func (r *UserMySQLRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	return exists, err
}
