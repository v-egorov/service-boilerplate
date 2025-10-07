-- Rollback key rotation enhancements
-- Migration: 000003_key_rotation.down.sql

-- Remove rotation metadata from jwt_keys table
ALTER TABLE auth_service.jwt_keys
DROP COLUMN IF EXISTS rotation_reason,
DROP COLUMN IF EXISTS rotated_at,
DROP COLUMN IF EXISTS expires_at;

-- Drop key rotation configuration table
DROP TABLE IF EXISTS auth_service.key_rotation_config;

-- Drop indexes
DROP INDEX IF EXISTS auth_service.idx_jwt_keys_rotated_at;
DROP INDEX IF EXISTS auth_service.idx_key_rotation_config_enabled;