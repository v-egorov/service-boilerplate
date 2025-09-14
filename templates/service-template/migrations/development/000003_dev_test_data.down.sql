-- Remove development test data for SERVICE_NAME service

-- Delete test entities
DELETE FROM entities WHERE name IN (
    'Test Entity 1',
    'Test Entity 2',
    'Sample Item',
    'Demo Entity'
);