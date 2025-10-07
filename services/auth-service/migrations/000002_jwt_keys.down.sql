-- Drop JWT keys table
DROP TABLE IF EXISTS auth_service.jwt_keys;

-- Drop indexes
DROP INDEX IF EXISTS auth_service.idx_jwt_keys_active;
DROP INDEX IF EXISTS auth_service.idx_jwt_keys_key_id;
DROP INDEX IF EXISTS auth_service.idx_jwt_keys_active_created;