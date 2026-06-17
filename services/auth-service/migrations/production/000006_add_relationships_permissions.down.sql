-- Remove relationships permissions (R2.13 down)

DELETE FROM auth_service.permissions WHERE name LIKE 'relationships:%';