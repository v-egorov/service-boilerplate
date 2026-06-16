-- Environment: development
-- Seed type_key values for existing object_types rows
-- Mapping derived from auth_service.permissions.resource column
-- Uses individual UPDATEs — silently skips rows that don't exist

UPDATE objects_service.object_types SET type_key = 'relationship-types' WHERE name = 'RelationshipType';
UPDATE objects_service.object_types SET type_key = 'relationships' WHERE name = 'Relationship';
UPDATE objects_service.object_types SET type_key = 'category' WHERE name = 'Category';
UPDATE objects_service.object_types SET type_key = 'product' WHERE name = 'Product';
UPDATE objects_service.object_types SET type_key = 'article' WHERE name = 'Article';
UPDATE objects_service.object_types SET type_key = 'location' WHERE name = 'Location';

-- Add unique constraint on type_key (idempotent — only if not already present)
DO $$ BEGIN
    ALTER TABLE objects_service.object_types ADD CONSTRAINT object_types_type_key_unique UNIQUE (type_key);
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
