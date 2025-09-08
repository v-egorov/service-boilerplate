-- Migration: 000004_add_user_settings
-- Description: Add new table to user_service schema
-- Created: Вс 07 сен 2025 23:54:28 MSK

-- Create table
CREATE TABLE IF NOT EXISTS user_service.add_user_settings (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Add indexes if needed
-- CREATE INDEX IF NOT EXISTS idx_add_user_settings_created_at ON user_service.add_user_settings(created_at);

-- Add comments
COMMENT ON TABLE user_service.add_user_settings IS 'Table for add_user_settings functionality';
