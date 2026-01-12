-- Drop entities table and related objects
DROP TRIGGER IF EXISTS update_entities_updated_at ON objects_service.entities;
DROP FUNCTION IF EXISTS objects_service.update_entities_updated_at_column();
DROP INDEX IF EXISTS objects_service.idx_entities_created_at;
DROP INDEX IF EXISTS objects_service.idx_entities_name;
DROP TABLE IF EXISTS objects_service.entities;

-- Drop migration executions table and indexes
DROP INDEX IF EXISTS objects_service.idx_migration_executions_created_at;
DROP INDEX IF EXISTS objects_service.idx_migration_executions_status;
DROP INDEX IF EXISTS objects_service.idx_migration_executions_environment;
DROP INDEX IF EXISTS objects_service.idx_migration_executions_migration_id;
DROP TABLE IF EXISTS objects_service.migration_executions;

-- Drop objects_service schema
DROP SCHEMA IF EXISTS objects_service CASCADE;