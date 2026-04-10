-- Environment: development
-- Add read permission for relationship-types to user role
-- This allows regular users to view/list relationship types but not modify them

INSERT INTO auth_service.role_permissions (role_id, permission_id)
SELECT 
    (SELECT id FROM auth_service.roles WHERE name = 'user'),
    (SELECT id FROM auth_service.permissions WHERE name = 'relationship-types:read')
WHERE NOT EXISTS (
    SELECT 1 FROM auth_service.role_permissions rp
    JOIN auth_service.roles r ON rp.role_id = r.id
    WHERE r.name = 'user'
    AND rp.permission_id = (SELECT id FROM auth_service.permissions WHERE name = 'relationship-types:read')
);