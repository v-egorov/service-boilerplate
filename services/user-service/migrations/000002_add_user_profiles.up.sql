-- Migration: 000002_add_user_profiles
-- Description: Add user profiles table with extended user information
-- Created: Auto-generated

-- Create user profiles table
CREATE TABLE IF NOT EXISTS user_service.user_profiles (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES user_service.users(id) ON DELETE CASCADE,
    bio TEXT,
    avatar_url VARCHAR(500),
    website VARCHAR(255),
    location VARCHAR(100),
    timezone VARCHAR(50) DEFAULT 'UTC',
    theme VARCHAR(20) DEFAULT 'light' CHECK (theme IN ('light', 'dark', 'auto')),
    email_notifications BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(user_id)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_user_profiles_user_id ON user_service.user_profiles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_profiles_theme ON user_service.user_profiles(theme);
CREATE INDEX IF NOT EXISTS idx_user_profiles_created_at ON user_service.user_profiles(created_at);

-- Add table and column comments
COMMENT ON TABLE user_service.user_profiles IS 'Extended user profile information and preferences';
COMMENT ON COLUMN user_service.user_profiles.bio IS 'User biography text (max 1000 characters)';
COMMENT ON COLUMN user_service.user_profiles.theme IS 'UI theme preference';
COMMENT ON COLUMN user_service.user_profiles.email_notifications IS 'Whether user wants email notifications';