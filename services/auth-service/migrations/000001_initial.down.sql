-- Drop triggers and functions
DROP TRIGGER IF EXISTS update_auth_tokens_updated_at ON auth_service.auth_tokens;
DROP FUNCTION IF EXISTS auth_service.update_auth_tokens_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS auth_service.idx_auth_tokens_user_id;
DROP INDEX IF EXISTS auth_service.idx_auth_tokens_token_hash;
DROP INDEX IF EXISTS auth_service.idx_auth_tokens_expires_at;
DROP INDEX IF EXISTS auth_service.idx_user_sessions_user_id;
DROP INDEX IF EXISTS auth_service.idx_user_sessions_expires_at;
DROP INDEX IF EXISTS auth_service.idx_roles_name;
DROP INDEX IF EXISTS auth_service.idx_permissions_name;
DROP INDEX IF EXISTS auth_service.idx_permissions_resource_action;

-- Drop tables (in reverse dependency order)
DROP TABLE IF EXISTS auth_service.user_roles;
DROP TABLE IF EXISTS auth_service.role_permissions;
DROP TABLE IF EXISTS auth_service.permissions;
DROP TABLE IF EXISTS auth_service.roles;
DROP TABLE IF EXISTS auth_service.user_sessions;
DROP TABLE IF EXISTS auth_service.auth_tokens;

-- Drop auth_service schema
DROP SCHEMA IF EXISTS auth_service CASCADE;