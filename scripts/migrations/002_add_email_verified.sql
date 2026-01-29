-- Migration: Add is_email_verified column to users table
-- Description: Thêm cột is_email_verified để hỗ trợ xác thực email
-- Date: 2026-01-28

-- Thêm cột is_email_verified
ALTER TABLE users
ADD COLUMN is_email_verified BOOLEAN NOT NULL DEFAULT FALSE
AFTER is_active;

-- Tạo index cho is_email_verified (hữu ích cho việc query users chưa verify)
CREATE INDEX idx_users_is_email_verified ON users(is_email_verified);

-- Mark existing users as verified (backward compatibility)
UPDATE users SET is_email_verified = TRUE;
