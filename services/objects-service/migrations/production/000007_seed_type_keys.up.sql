-- Environment: production
-- Seed type_key values for existing object_types rows (production)
-- Production only has RelationshipType — no test data or relationship instances yet

UPDATE objects_service.object_types SET type_key = 'relationship-types' WHERE name = 'RelationshipType';

-- Add unique constraint on type_key (idempotent — only if not already present)
DO $$ BEGIN
    ALTER TABLE objects_service.object_types ADD CONSTRAINT object_types_type_key_unique UNIQUE (type_key);
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
