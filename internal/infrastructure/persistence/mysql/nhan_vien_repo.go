// Package mysql chứa các MySQL repository implementations
package mysql

import (
	"context"
	"database/sql"
	"errors"

	"restaurant_project/internal/domain/entity"
	"restaurant_project/internal/domain/repository"
)

// NhanVienMySQLRepo là implementation của INhanVienRepository sử dụng MySQL
type NhanVienMySQLRepo struct {
	db *sql.DB
}

// NewNhanVienMySQLRepo tạo mới NhanVienMySQLRepo
func NewNhanVienMySQLRepo(db *sql.DB) *NhanVienMySQLRepo {
	return &NhanVienMySQLRepo{db: db}
}

// Verify interface implementation at compile time
var _ repository.INhanVienRepository = (*NhanVienMySQLRepo)(nil)

// FindByID tìm nhân viên theo ID
func (r *NhanVienMySQLRepo) FindByID(ctx context.Context, id string) (*entity.NhanVien, error) {
	query := `SELECT id, user_id, ho_ten, chuc_vu, so_dien_thoai, email,
			  trang_thai, luong_co_ban, ngay_vao_lam, ngay_tao, ngay_cap_nhat
			  FROM nhan_vien WHERE id = ?`

	nv := &entity.NhanVien{}
	var email sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&nv.ID, &nv.UserID, &nv.HoTen, &nv.ChucVu, &nv.SoDienThoai, &email,
		&nv.TrangThai, &nv.LuongCoBan, &nv.NgayVaoLam, &nv.NgayTao, &nv.NgayCapNhat,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if email.Valid {
		nv.Email = email.String
	}

	return nv, nil
}

// FindByUserID tìm nhân viên theo UserID
func (r *NhanVienMySQLRepo) FindByUserID(ctx context.Context, userID string) (*entity.NhanVien, error) {
	query := `SELECT id, user_id, ho_ten, chuc_vu, so_dien_thoai, email,
			  trang_thai, luong_co_ban, ngay_vao_lam, ngay_tao, ngay_cap_nhat
			  FROM nhan_vien WHERE user_id = ?`

	nv := &entity.NhanVien{}
	var email sql.NullString

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&nv.ID, &nv.UserID, &nv.HoTen, &nv.ChucVu, &nv.SoDienThoai, &email,
		&nv.TrangThai, &nv.LuongCoBan, &nv.NgayVaoLam, &nv.NgayTao, &nv.NgayCapNhat,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if email.Valid {
		nv.Email = email.String
	}

	return nv, nil
}

// FindAll lấy tất cả nhân viên
func (r *NhanVienMySQLRepo) FindAll(ctx context.Context) ([]*entity.NhanVien, error) {
	query := `SELECT id, user_id, ho_ten, chuc_vu, so_dien_thoai, email,
			  trang_thai, luong_co_ban, ngay_vao_lam, ngay_tao, ngay_cap_nhat
			  FROM nhan_vien ORDER BY ngay_vao_lam DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*entity.NhanVien
	for rows.Next() {
		nv := &entity.NhanVien{}
		var email sql.NullString

		err := rows.Scan(
			&nv.ID, &nv.UserID, &nv.HoTen, &nv.ChucVu, &nv.SoDienThoai, &email,
			&nv.TrangThai, &nv.LuongCoBan, &nv.NgayVaoLam, &nv.NgayTao, &nv.NgayCapNhat,
		)
		if err != nil {
			return nil, err
		}

		if email.Valid {
			nv.Email = email.String
		}

		list = append(list, nv)
	}

	return list, rows.Err()
}

// FindByChucVu lấy nhân viên theo chức vụ
func (r *NhanVienMySQLRepo) FindByChucVu(ctx context.Context, chucVu entity.ChucVu) ([]*entity.NhanVien, error) {
	query := `SELECT id, user_id, ho_ten, chuc_vu, so_dien_thoai, email,
			  trang_thai, luong_co_ban, ngay_vao_lam, ngay_tao, ngay_cap_nhat
			  FROM nhan_vien WHERE chuc_vu = ? ORDER BY ho_ten`

	rows, err := r.db.QueryContext(ctx, query, chucVu)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*entity.NhanVien
	for rows.Next() {
		nv := &entity.NhanVien{}
		var email sql.NullString

		err := rows.Scan(
			&nv.ID, &nv.UserID, &nv.HoTen, &nv.ChucVu, &nv.SoDienThoai, &email,
			&nv.TrangThai, &nv.LuongCoBan, &nv.NgayVaoLam, &nv.NgayTao, &nv.NgayCapNhat,
		)
		if err != nil {
			return nil, err
		}

		if email.Valid {
			nv.Email = email.String
		}

		list = append(list, nv)
	}

	return list, rows.Err()
}

// FindByTrangThai lấy nhân viên theo trạng thái làm việc
func (r *NhanVienMySQLRepo) FindByTrangThai(ctx context.Context, trangThai entity.TrangThaiLamViec) ([]*entity.NhanVien, error) {
	query := `SELECT id, user_id, ho_ten, chuc_vu, so_dien_thoai, email,
			  trang_thai, luong_co_ban, ngay_vao_lam, ngay_tao, ngay_cap_nhat
			  FROM nhan_vien WHERE trang_thai = ? ORDER BY ho_ten`

	rows, err := r.db.QueryContext(ctx, query, trangThai)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*entity.NhanVien
	for rows.Next() {
		nv := &entity.NhanVien{}
		var email sql.NullString

		err := rows.Scan(
			&nv.ID, &nv.UserID, &nv.HoTen, &nv.ChucVu, &nv.SoDienThoai, &email,
			&nv.TrangThai, &nv.LuongCoBan, &nv.NgayVaoLam, &nv.NgayTao, &nv.NgayCapNhat,
		)
		if err != nil {
			return nil, err
		}

		if email.Valid {
			nv.Email = email.String
		}

		list = append(list, nv)
	}

	return list, rows.Err()
}

// FindDauBepRanh tìm đầu bếp đang rảnh
func (r *NhanVienMySQLRepo) FindDauBepRanh(ctx context.Context) ([]*entity.NhanVien, error) {
	query := `SELECT id, user_id, ho_ten, chuc_vu, so_dien_thoai, email,
			  trang_thai, luong_co_ban, ngay_vao_lam, ngay_tao, ngay_cap_nhat
			  FROM nhan_vien WHERE chuc_vu = ? AND trang_thai = ? ORDER BY ho_ten`

	rows, err := r.db.QueryContext(ctx, query, entity.ChucVuBep, entity.TrangThaiRanh)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*entity.NhanVien
	for rows.Next() {
		nv := &entity.NhanVien{}
		var email sql.NullString

		err := rows.Scan(
			&nv.ID, &nv.UserID, &nv.HoTen, &nv.ChucVu, &nv.SoDienThoai, &email,
			&nv.TrangThai, &nv.LuongCoBan, &nv.NgayVaoLam, &nv.NgayTao, &nv.NgayCapNhat,
		)
		if err != nil {
			return nil, err
		}

		if email.Valid {
			nv.Email = email.String
		}

		list = append(list, nv)
	}

	return list, rows.Err()
}

// Save lưu nhân viên mới hoặc cập nhật
func (r *NhanVienMySQLRepo) Save(ctx context.Context, nv *entity.NhanVien) error {
	query := `INSERT INTO nhan_vien (id, user_id, ho_ten, chuc_vu, so_dien_thoai, email,
			  trang_thai, luong_co_ban, ngay_vao_lam, ngay_tao, ngay_cap_nhat)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			  ON DUPLICATE KEY UPDATE
			  user_id = VALUES(user_id),
			  ho_ten = VALUES(ho_ten),
			  chuc_vu = VALUES(chuc_vu),
			  so_dien_thoai = VALUES(so_dien_thoai),
			  email = VALUES(email),
			  trang_thai = VALUES(trang_thai),
			  luong_co_ban = VALUES(luong_co_ban),
			  ngay_cap_nhat = VALUES(ngay_cap_nhat)`

	var email interface{}
	if nv.Email != "" {
		email = nv.Email
	}

	_, err := r.db.ExecContext(ctx, query,
		nv.ID, nv.UserID, nv.HoTen, nv.ChucVu, nv.SoDienThoai, email,
		nv.TrangThai, nv.LuongCoBan, nv.NgayVaoLam, nv.NgayTao, nv.NgayCapNhat,
	)

	return err
}

// Delete xóa nhân viên theo ID
func (r *NhanVienMySQLRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM nhan_vien WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("không tìm thấy nhân viên để xóa")
	}

	return nil
}

// UpdateTrangThai cập nhật trạng thái làm việc
func (r *NhanVienMySQLRepo) UpdateTrangThai(ctx context.Context, id string, trangThai entity.TrangThaiLamViec) error {
	query := `UPDATE nhan_vien SET trang_thai = ?, ngay_cap_nhat = NOW() WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, trangThai, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("không tìm thấy nhân viên để cập nhật")
	}

	return nil
}

// Count đếm tổng số nhân viên
func (r *NhanVienMySQLRepo) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM nhan_vien`
	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}
