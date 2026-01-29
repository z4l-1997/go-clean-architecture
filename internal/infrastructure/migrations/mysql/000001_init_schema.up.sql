-- ===========================================
-- Restaurant Project - MySQL Schema
-- ===========================================
-- Polyglot Persistence: MySQL lưu dữ liệu ổn định
-- - users: Tài khoản đăng nhập
-- - khach_hang: Thông tin khách hàng
-- - nhan_vien: Thông tin nhân viên/đầu bếp

-- ===========================================
-- BẢNG USERS - Tài khoản người dùng
-- ===========================================
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,                -- UUID
    username VARCHAR(50) NOT NULL UNIQUE,      -- Tên đăng nhập
    email VARCHAR(100) NOT NULL UNIQUE,        -- Email
    password_hash VARCHAR(255) NOT NULL,       -- Mật khẩu đã hash (bcrypt)
    role ENUM('admin', 'manager', 'staff', 'customer') NOT NULL DEFAULT 'customer',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,   -- Tài khoản có hoạt động không
    ngay_tao DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ngay_cap_nhat DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_username (username),
    INDEX idx_email (email),
    INDEX idx_role (role),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ===========================================
-- BẢNG KHACH_HANG - Khách hàng
-- ===========================================
CREATE TABLE IF NOT EXISTS khach_hang (
    id VARCHAR(36) PRIMARY KEY,                -- UUID
    user_id VARCHAR(36),                       -- FK -> users (nullable - khách vãng lai)
    ho_ten VARCHAR(100) NOT NULL,              -- Họ và tên
    so_dien_thoai VARCHAR(15) NOT NULL UNIQUE, -- Số điện thoại
    email VARCHAR(100),                        -- Email (optional)
    dia_chi TEXT,                              -- Địa chỉ giao hàng
    diem_tich_luy BIGINT NOT NULL DEFAULT 0,   -- Điểm tích lũy (loyalty points)
    cap_thanh_vien ENUM('bronze', 'silver', 'gold', 'platinum') NOT NULL DEFAULT 'bronze',
    ngay_tao DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ngay_cap_nhat DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    INDEX idx_user_id (user_id),
    INDEX idx_so_dien_thoai (so_dien_thoai),
    INDEX idx_cap_thanh_vien (cap_thanh_vien),
    INDEX idx_diem_tich_luy (diem_tich_luy)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ===========================================
-- BẢNG NHAN_VIEN - Nhân viên/Đầu bếp
-- ===========================================
CREATE TABLE IF NOT EXISTS nhan_vien (
    id VARCHAR(36) PRIMARY KEY,                -- UUID
    user_id VARCHAR(36) NOT NULL,              -- FK -> users
    ho_ten VARCHAR(100) NOT NULL,              -- Họ và tên
    chuc_vu ENUM('bep', 'phuc_vu', 'thu_ngan', 'quan_ly', 'giao_hang') NOT NULL,
    so_dien_thoai VARCHAR(15) NOT NULL,        -- Số điện thoại
    email VARCHAR(100),                        -- Email công ty
    trang_thai ENUM('ranh', 'ban', 'nghi', 'offline') NOT NULL DEFAULT 'offline',
    luong_co_ban BIGINT NOT NULL DEFAULT 0,    -- Lương cơ bản (VND/tháng)
    ngay_vao_lam DATE NOT NULL,                -- Ngày bắt đầu làm việc
    ngay_tao DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ngay_cap_nhat DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_chuc_vu (chuc_vu),
    INDEX idx_trang_thai (trang_thai),
    INDEX idx_chuc_vu_trang_thai (chuc_vu, trang_thai)  -- Composite index cho FindDauBepRanh
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ===========================================
-- SAMPLE DATA (Optional - cho development)
-- ===========================================

-- Admin user (password: admin123 - đã hash bằng bcrypt)
INSERT INTO users (id, username, email, password_hash, role, is_active) VALUES
('550e8400-e29b-41d4-a716-446655440001', 'admin', 'admin@restaurant.vn',
 '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'admin', TRUE)
ON DUPLICATE KEY UPDATE username = username;

-- Manager user (password: manager123)
INSERT INTO users (id, username, email, password_hash, role, is_active) VALUES
('550e8400-e29b-41d4-a716-446655440002', 'manager', 'manager@restaurant.vn',
 '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'manager', TRUE)
ON DUPLICATE KEY UPDATE username = username;

-- Staff user - Đầu bếp (password: chef123)
INSERT INTO users (id, username, email, password_hash, role, is_active) VALUES
('550e8400-e29b-41d4-a716-446655440003', 'chef01', 'chef01@restaurant.vn',
 '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'staff', TRUE)
ON DUPLICATE KEY UPDATE username = username;

-- Nhân viên đầu bếp
INSERT INTO nhan_vien (id, user_id, ho_ten, chuc_vu, so_dien_thoai, email, trang_thai, luong_co_ban, ngay_vao_lam) VALUES
('660e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440003',
 'Nguyễn Văn Bếp', 'bep', '0901234567', 'chef01@restaurant.vn', 'ranh', 15000000, '2024-01-15')
ON DUPLICATE KEY UPDATE ho_ten = ho_ten;

-- Khách hàng mẫu
INSERT INTO khach_hang (id, ho_ten, so_dien_thoai, email, diem_tich_luy, cap_thanh_vien) VALUES
('770e8400-e29b-41d4-a716-446655440001', 'Trần Thị Lan', '0987654321', 'lan.tran@email.com', 5500, 'gold'),
('770e8400-e29b-41d4-a716-446655440002', 'Lê Văn Hùng', '0912345678', 'hung.le@email.com', 1200, 'bronze')
ON DUPLICATE KEY UPDATE ho_ten = ho_ten;
