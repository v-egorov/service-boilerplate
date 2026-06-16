-- Environment: all
-- Add type_key column to object_types table
-- type_key is the human-readable namespace used in permission strings (e.g., "objects", "relationships")
-- Nullable initially — subsequent migrations populate and add unique constraint

ALTER TABLE objects_service.object_types ADD COLUMN IF NOT EXISTS type_key VARCHAR(100);
