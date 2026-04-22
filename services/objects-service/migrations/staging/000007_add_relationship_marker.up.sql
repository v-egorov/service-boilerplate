-- Environment: all
-- Create Relationship marker in object_types (if not exists)
INSERT INTO objects_service.object_types (name, description, created_at, updated_at)
VALUES ('Relationship', 'Marker type for relationship instances', NOW(), NOW())
ON CONFLICT (name) DO NOTHING
RETURNING id;