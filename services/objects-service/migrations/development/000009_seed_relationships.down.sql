-- Down migration: Remove seed relationships and test objects
-- Delete relationship instances first (foreign key constraint)
DELETE FROM objects_service.objects_relationships;

-- Delete relationship marker objects
DELETE FROM objects_service.objects
WHERE object_type_id = (SELECT id FROM objects_service.object_types WHERE name = 'Relationship');

-- Delete test objects
DELETE FROM objects_service.objects WHERE name IN (
    'Test Portfolio A', 'Test Portfolio B',
    'Test Asset X', 'Test Asset Y', 'Test Asset Z',
    'Test Article Alpha', 'Test Article Beta', 'Test Article Gamma',
    'Test Category Parent', 'Test Category Child'
);