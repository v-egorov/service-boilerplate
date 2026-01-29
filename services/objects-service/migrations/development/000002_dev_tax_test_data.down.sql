-- Remove development test data from objects_service taxonomy system
-- This migration removes sample object types and objects

-- Drop all objects (child objects first due to foreign key constraints)
DELETE FROM objects_service.objects WHERE parent_object_id IS NOT NULL;

-- Drop root objects
DELETE FROM objects_service.objects WHERE parent_object_id IS NULL;

-- Drop child object types first
DELETE FROM objects_service.object_types WHERE parent_type_id IS NOT NULL;

-- Drop root object types
DELETE FROM objects_service.object_types WHERE parent_type_id IS NULL;