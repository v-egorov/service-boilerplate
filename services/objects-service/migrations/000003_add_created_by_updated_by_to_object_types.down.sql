-- Environment: all
-- Remove created_by and updated_by columns from object_types
ALTER TABLE objects_service.object_types 
DROP COLUMN IF EXISTS created_by,
DROP COLUMN IF EXISTS updated_by;

-- Drop indexes
DROP INDEX IF EXISTS idx_object_types_created_by;
DROP INDEX IF EXISTS idx_object_types_updated_by;
