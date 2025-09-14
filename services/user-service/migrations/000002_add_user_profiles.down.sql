-- Migration Rollback: 000002_add_user_profiles
-- Description: Remove user profiles table
-- Created: Auto-generated

-- Drop table (CASCADE removes indexes and constraints)
DROP TABLE IF EXISTS user_service.user_profiles CASCADE;