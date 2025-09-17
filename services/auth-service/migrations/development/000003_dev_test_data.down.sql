-- Remove development test data for auth-service service

-- Delete test entities
DELETE FROM auth_service.entities WHERE name IN (
    'Test Entity 1',
    'Test Entity 2',
    'Sample Item',
    'Demo Entity'
);