-- Environment: all
-- Drop triggers
DROP TRIGGER IF EXISTS update_object_types_updated_at ON objects_service.object_types;
DROP TRIGGER IF EXISTS update_objects_updated_at ON objects_service.objects;

-- Drop trigger function
DROP FUNCTION IF EXISTS objects_service.update_updated_at_column();

-- Drop indexes for objects
DROP INDEX IF EXISTS idx_objects_type_id;
DROP INDEX IF EXISTS idx_objects_type_status;
DROP INDEX IF EXISTS idx_objects_status;
DROP INDEX IF EXISTS idx_objects_parent_id;
DROP INDEX IF EXISTS idx_objects_public_id;
DROP INDEX IF EXISTS idx_objects_metadata_gin;
DROP INDEX IF EXISTS idx_objects_tags_gin;
DROP INDEX IF EXISTS idx_objects_created_at;
DROP INDEX IF EXISTS idx_objects_updated_at;
DROP INDEX IF EXISTS idx_objects_deleted_at;

-- Drop indexes for object_types
DROP INDEX IF EXISTS idx_object_types_parent;
DROP INDEX IF EXISTS idx_object_types_name;
DROP INDEX IF EXISTS idx_object_types_sealed;

-- Drop tables
DROP TABLE IF EXISTS objects_service.objects;
DROP TABLE IF EXISTS objects_service.object_types;

DELETE FROM objects_service.migration_executions;

-- Keep objects_service schema for migration tracking
-- Note: Schema cleanup should be handled separately if needed
-- Drop migration executions table and indexes (if they exist)
-- DROP INDEX IF EXISTS objects_service.idx_migration_executions_created_at;
-- DROP INDEX IF EXISTS objects_service.idx_migration_executions_status;
-- DROP INDEX IF EXISTS objects_service.idx_migration_executions_environment;
-- DROP INDEX IF EXISTS objects_service.idx_migration_executions_migration_id;
-- DROP TABLE IF EXISTS objects_service.migration_executions;

-- Keep objects_service schema for migration tracking
-- Note: Schema cleanup should be handled separately if needed
-- DROP SCHEMA IF EXISTS objects_service CASCADE;
