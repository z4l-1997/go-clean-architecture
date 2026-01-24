// Package mysql chứa các MySQL repository implementations
package mysql

import (
	"context"
	"database/sql"
	"errors"

	"restaurant_project/internal/domain/entity"
	"restaurant_project/internal/domain/repository"
)

// KhachHangMySQLRepo là implementation của IKhachHangRepository sử dụng MySQL
type KhachHangMySQLRepo struct {
	db *sql.DB
}

// NewKhachHangMySQLRepo tạo mới KhachHangMySQLRepo
func NewKhachHangMySQLRepo(db *sql.DB) *KhachHangMySQLRepo {
	return &KhachHangMySQLRepo{db: db}
}

// Verify interface implementation at compile time
var _ repository.IKhachHangRepository = (*KhachHangMySQLRepo)(nil)

// FindByID tìm khách hàng theo ID
func (r *KhachHangMySQLRepo) FindByID(ctx context.Context, id string) (*entity.KhachHang, error) {
	query := `SELECT id, user_id, ho_ten, so_dien_thoai, email, dia_chi,
			  diem_tich_luy, cap_thanh_vien, ngay_tao, ngay_cap_nhat
			  FROM khach_hang WHERE id = ?`

	kh := &entity.KhachHang{}
	var userID sql.NullString
	var email, diaChi sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&kh.ID, &userID, &kh.HoTen, &kh.SoDienThoai, &email, &diaChi,
		&kh.DiemTichLuy, &kh.CapThanhVien, &kh.NgayTao, &kh.NgayCapNhat,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if userID.Valid {
		kh.UserID = userID.String
	}
	if email.Valid {
		kh.Email = email.String
	}
	if diaChi.Valid {
		kh.DiaChi = diaChi.String
	}

	return kh, nil
}

// FindByUserID tìm khách hàng theo UserID
func (r *KhachHangMySQLRepo) FindByUserID(ctx context.Context, userID string) (*entity.KhachHang, error) {
	query := `SELECT id, user_id, ho_ten, so_dien_thoai, email, dia_chi,
			  diem_tich_luy, cap_thanh_vien, ngay_tao, ngay_cap_nhat
			  FROM khach_hang WHERE user_id = ?`

	kh := &entity.KhachHang{}
	var uid sql.NullString
	var email, diaChi sql.NullString

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&kh.ID, &uid, &kh.HoTen, &kh.SoDienThoai, &email, &diaChi,
		&kh.DiemTichLuy, &kh.CapThanhVien, &kh.NgayTao, &kh.NgayCapNhat,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if uid.Valid {
		kh.UserID = uid.String
	}
	if email.Valid {
		kh.Email = email.String
	}
	if diaChi.Valid {
		kh.DiaChi = diaChi.String
	}

	return kh, nil
}

// FindBySoDienThoai tìm khách hàng theo số điện thoại
func (r *KhachHangMySQLRepo) FindBySoDienThoai(ctx context.Context, soDienThoai string) (*entity.KhachHang, error) {
	query := `SELECT id, user_id, ho_ten, so_dien_thoai, email, dia_chi,
			  diem_tich_luy, cap_thanh_vien, ngay_tao, ngay_cap_nhat
			  FROM khach_hang WHERE so_dien_thoai = ?`

	kh := &entity.KhachHang{}
	var userID sql.NullString
	var email, diaChi sql.NullString

	err := r.db.QueryRowContext(ctx, query, soDienThoai).Scan(
		&kh.ID, &userID, &kh.HoTen, &kh.SoDienThoai, &email, &diaChi,
		&kh.DiemTichLuy, &kh.CapThanhVien, &kh.NgayTao, &kh.NgayCapNhat,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if userID.Valid {
		kh.UserID = userID.String
	}
	if email.Valid {
		kh.Email = email.String
	}
	if diaChi.Valid {
		kh.DiaChi = diaChi.String
	}

	return kh, nil
}

// FindAll lấy tất cả khách hàng
func (r *KhachHangMySQLRepo) FindAll(ctx context.Context) ([]*entity.KhachHang, error) {
	query := `SELECT id, user_id, ho_ten, so_dien_thoai, email, dia_chi,
			  diem_tich_luy, cap_thanh_vien, ngay_tao, ngay_cap_nhat
			  FROM khach_hang ORDER BY ngay_tao DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*entity.KhachHang
	for rows.Next() {
		kh := &entity.KhachHang{}
		var userID sql.NullString
		var email, diaChi sql.NullString

		err := rows.Scan(
			&kh.ID, &userID, &kh.HoTen, &kh.SoDienThoai, &email, &diaChi,
			&kh.DiemTichLuy, &kh.CapThanhVien, &kh.NgayTao, &kh.NgayCapNhat,
		)
		if err != nil {
			return nil, err
		}

		if userID.Valid {
			kh.UserID = userID.String
		}
		if email.Valid {
			kh.Email = email.String
		}
		if diaChi.Valid {
			kh.DiaChi = diaChi.String
		}

		list = append(list, kh)
	}

	return list, rows.Err()
}

// FindByCapThanhVien lấy khách hàng theo cấp thành viên
func (r *KhachHangMySQLRepo) FindByCapThanhVien(ctx context.Context, cap string) ([]*entity.KhachHang, error) {
	query := `SELECT id, user_id, ho_ten, so_dien_thoai, email, dia_chi,
			  diem_tich_luy, cap_thanh_vien, ngay_tao, ngay_cap_nhat
			  FROM khach_hang WHERE cap_thanh_vien = ? ORDER BY diem_tich_luy DESC`

	rows, err := r.db.QueryContext(ctx, query, cap)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*entity.KhachHang
	for rows.Next() {
		kh := &entity.KhachHang{}
		var userID sql.NullString
		var email, diaChi sql.NullString

		err := rows.Scan(
			&kh.ID, &userID, &kh.HoTen, &kh.SoDienThoai, &email, &diaChi,
			&kh.DiemTichLuy, &kh.CapThanhVien, &kh.NgayTao, &kh.NgayCapNhat,
		)
		if err != nil {
			return nil, err
		}

		if userID.Valid {
			kh.UserID = userID.String
		}
		if email.Valid {
			kh.Email = email.String
		}
		if diaChi.Valid {
			kh.DiaChi = diaChi.String
		}

		list = append(list, kh)
	}

	return list, rows.Err()
}

// Save lưu khách hàng mới hoặc cập nhật
func (r *KhachHangMySQLRepo) Save(ctx context.Context, kh *entity.KhachHang) error {
	query := `INSERT INTO khach_hang (id, user_id, ho_ten, so_dien_thoai, email, dia_chi,
			  diem_tich_luy, cap_thanh_vien, ngay_tao, ngay_cap_nhat)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			  ON DUPLICATE KEY UPDATE
			  user_id = VALUES(user_id),
			  ho_ten = VALUES(ho_ten),
			  so_dien_thoai = VALUES(so_dien_thoai),
			  email = VALUES(email),
			  dia_chi = VALUES(dia_chi),
			  diem_tich_luy = VALUES(diem_tich_luy),
			  cap_thanh_vien = VALUES(cap_thanh_vien),
			  ngay_cap_nhat = VALUES(ngay_cap_nhat)`

	var userID, email, diaChi interface{}
	if kh.UserID != "" {
		userID = kh.UserID
	}
	if kh.Email != "" {
		email = kh.Email
	}
	if kh.DiaChi != "" {
		diaChi = kh.DiaChi
	}

	_, err := r.db.ExecContext(ctx, query,
		kh.ID, userID, kh.HoTen, kh.SoDienThoai, email, diaChi,
		kh.DiemTichLuy, kh.CapThanhVien, kh.NgayTao, kh.NgayCapNhat,
	)

	return err
}

// Delete xóa khách hàng theo ID
func (r *KhachHangMySQLRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM khach_hang WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("không tìm thấy khách hàng để xóa")
	}

	return nil
}

// UpdateDiemTichLuy cập nhật điểm tích lũy (atomic operation)
func (r *KhachHangMySQLRepo) UpdateDiemTichLuy(ctx context.Context, id string, diemMoi int64) error {
	query := `UPDATE khach_hang SET diem_tich_luy = ?, ngay_cap_nhat = NOW() WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, diemMoi, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("không tìm thấy khách hàng để cập nhật")
	}

	return nil
}

// Count đếm tổng số khách hàng
func (r *KhachHangMySQLRepo) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM khach_hang`
	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}
