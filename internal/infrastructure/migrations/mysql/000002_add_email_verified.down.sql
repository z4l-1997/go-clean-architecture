-- Rollback: Remove is_email_verified column
DROP INDEX idx_users_is_email_verified ON users;
ALTER TABLE users DROP COLUMN is_email_verified;
