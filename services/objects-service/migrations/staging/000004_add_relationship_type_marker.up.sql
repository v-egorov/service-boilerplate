-- Environment: all
-- Create RelationshipType marker in object_types (if not exists)
INSERT INTO objects_service.object_types (name, description, created_at, updated_at)
VALUES ('RelationshipType', 'Marker type for relationship type instances', NOW(), NOW())
ON CONFLICT (name) DO NOTHING
RETURNING id;
