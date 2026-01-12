-- Remove development test data for objects-service service

-- Delete test entities
DELETE FROM objects_service.entities WHERE name IN (
    'Test Entity 1',
    'Test Entity 2',
    'Sample Item',
    'Demo Entity'
);