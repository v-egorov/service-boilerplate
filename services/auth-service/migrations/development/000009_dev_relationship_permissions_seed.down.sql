-- Environment: development
-- Remove relationship-types permissions from roles
DELETE FROM auth_service.role_permissions
WHERE permission_id IN (
    SELECT id FROM auth_service.permissions
    WHERE name LIKE 'relationship-types:%'
);
