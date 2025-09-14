-- Migration: 000004_add_user_settings
-- Description: Add new table to user_service schema
-- Created: Вс 07 сен 2025 23:54:28 MSK

-- Create user settings table
CREATE TABLE IF NOT EXISTS user_service.user_settings (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES user_service.users(id) ON DELETE CASCADE,
    setting_key VARCHAR(100) NOT NULL,
    setting_value TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(user_id, setting_key)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_user_settings_user_id ON user_service.user_settings(user_id);
CREATE INDEX IF NOT EXISTS idx_user_settings_key ON user_service.user_settings(setting_key);
CREATE INDEX IF NOT EXISTS idx_user_settings_created_at ON user_service.user_settings(created_at);

-- Add table and column comments
COMMENT ON TABLE user_service.user_settings IS 'User-specific application settings and preferences';
COMMENT ON COLUMN user_service.user_settings.setting_key IS 'Setting name/key identifier';
COMMENT ON COLUMN user_service.user_settings.setting_value IS 'Setting value (can be JSON or text)';
