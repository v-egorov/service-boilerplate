-- Migration: 000005_add_password_hash (rollback)
-- Description: Remove password_hash column from users table
-- Created: Auto-generated

-- Drop index first
DROP INDEX IF EXISTS user_service.idx_users_password_hash;

-- Remove password_hash column from users table
ALTER TABLE user_service.users
DROP COLUMN IF EXISTS password_hash;