-- Remove read permission for relationship-types from user role

DELETE FROM auth_service.role_permissions
WHERE role_id = (SELECT id FROM auth_service.roles WHERE name = 'user')
AND permission_id = (SELECT id FROM auth_service.permissions WHERE name = 'relationship-types:read');