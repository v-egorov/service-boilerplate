-- Environment: all
-- Down migration for type_key column addition (production)
-- Only drop if column exists (idempotent)

ALTER TABLE objects_service.object_types DROP COLUMN IF EXISTS type_key;
