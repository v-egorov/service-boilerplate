-- Drop entities table and related objects
DROP TRIGGER IF EXISTS update_entities_updated_at ON SCHEMA_NAME.entities;
DROP FUNCTION IF EXISTS SCHEMA_NAME.update_entities_updated_at_column();
DROP INDEX IF EXISTS SCHEMA_NAME.idx_entities_created_at;
DROP INDEX IF EXISTS SCHEMA_NAME.idx_entities_name;
DROP TABLE IF EXISTS SCHEMA_NAME.entities;

-- Drop SCHEMA_NAME schema
DROP SCHEMA IF EXISTS SCHEMA_NAME CASCADE;