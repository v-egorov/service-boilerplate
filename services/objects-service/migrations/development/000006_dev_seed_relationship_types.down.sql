-- Remove seeded relationship types
DELETE FROM objects_service.objects_relationship_types;

-- Remove the base objects (but keep the marker type)
DELETE FROM objects_service.objects 
WHERE object_type_id = (SELECT id FROM objects_service.object_types WHERE name = 'RelationshipType');
