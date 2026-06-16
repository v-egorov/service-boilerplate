-- Environment: staging
-- Down migration for type_key seeding (staging)
-- Remove unique constraint and set type_keys back to NULL

DO $$ BEGIN
    ALTER TABLE objects_service.object_types DROP CONSTRAINT IF EXISTS object_types_type_key_unique;
EXCEPTION WHEN undefined_object THEN NULL;
END $$;

UPDATE objects_service.object_types SET type_key = NULL WHERE name IN ('RelationshipType', 'Relationship', 'Category', 'Product', 'Article', 'Location');
