-- Remove relationships permissions from roles (R2.14 down)

DELETE FROM auth_service.role_permissions
WHERE permission_id IN (
    SELECT id FROM auth_service.permissions WHERE name LIKE 'relationships:%'
);