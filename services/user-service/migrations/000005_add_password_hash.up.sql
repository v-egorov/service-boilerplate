-- Migration: 000005_add_password_hash
-- Description: Add password_hash column to users table for secure password storage
-- Created: Auto-generated

-- Add password_hash column to users table
ALTER TABLE user_service.users
ADD COLUMN password_hash VARCHAR(255);

-- Add comment for documentation
COMMENT ON COLUMN user_service.users.password_hash IS 'Bcrypt hashed password for user authentication';

-- Create index on password_hash (though it's not typically queried directly)
-- This is mainly for consistency and potential future use
CREATE INDEX IF NOT EXISTS idx_users_password_hash ON user_service.users(password_hash);