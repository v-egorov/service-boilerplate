-- Down migration: Remove Relationship marker
DELETE FROM objects_service.object_types WHERE name = 'Relationship';