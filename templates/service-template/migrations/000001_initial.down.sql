-- Drop services table and related objects
DROP TRIGGER IF EXISTS update_services_updated_at ON services;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP INDEX IF EXISTS idx_services_created_at;
DROP INDEX IF EXISTS idx_services_name;
DROP TABLE IF EXISTS services;