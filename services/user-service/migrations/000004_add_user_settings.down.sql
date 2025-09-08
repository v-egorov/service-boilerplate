-- Migration Rollback: 000004_add_user_settings
-- Description: Remove add_user_settings table from user_service schema
-- Created: Вс 07 сен 2025 23:54:28 MSK

-- Drop table (with CASCADE to remove dependencies)
DROP TABLE IF EXISTS user_service.add_user_settings CASCADE;
