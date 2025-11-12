-- Key rotation enhancements for auth-service
-- Migration: 000003_key_rotation.up.sql

-- Add rotation metadata to jwt_keys table
ALTER TABLE auth_service.jwt_keys
ADD COLUMN IF NOT EXISTS rotation_reason TEXT,
ADD COLUMN IF NOT EXISTS rotated_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS expires_at TIMESTAMP WITH TIME ZONE;

-- Create key rotation configuration table
CREATE TABLE IF NOT EXISTS auth_service.key_rotation_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rotation_type VARCHAR(50) NOT NULL DEFAULT 'time', -- 'time', 'usage', 'manual'
    interval_days INTEGER DEFAULT 30,
    max_tokens INTEGER DEFAULT 100000,
    overlap_minutes INTEGER DEFAULT 60,
    enabled BOOLEAN DEFAULT true,
    last_rotation_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Insert default rotation configuration
INSERT INTO auth_service.key_rotation_config (rotation_type, interval_days, max_tokens, overlap_minutes, enabled)
VALUES ('time', 30, 100000, 60, true)
ON CONFLICT DO NOTHING;

-- Create index for efficient rotation checks
CREATE INDEX IF NOT EXISTS idx_jwt_keys_rotated_at ON auth_service.jwt_keys(rotated_at);
CREATE INDEX IF NOT EXISTS idx_key_rotation_config_enabled ON auth_service.key_rotation_config(enabled);

-- Update existing keys with rotation metadata (set rotated_at to created_at for existing keys)
UPDATE auth_service.jwt_keys
SET rotated_at = created_at
WHERE rotated_at IS NULL;