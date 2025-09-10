-- Drop entities table and related objects
DROP TRIGGER IF EXISTS update_entities_updated_at ON entities;
DROP FUNCTION IF EXISTS update_entities_updated_at_column();
DROP INDEX IF EXISTS idx_entities_created_at;
DROP INDEX IF EXISTS idx_entities_name;
DROP TABLE IF EXISTS entities;