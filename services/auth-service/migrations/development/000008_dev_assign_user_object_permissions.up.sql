-- Environment: development
-- Development Migration: Assign object permissions to user role
-- Description: Give regular users basic permissions to access object-types and objects

-- Assign object-types:read to user role
INSERT INTO auth_service.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth_service.roles r
JOIN auth_service.permissions p ON p.name = 'object-types:read'
WHERE r.name = 'user'
ON CONFLICT DO NOTHING;

-- Assign basic object permissions to user role (own objects only)
INSERT INTO auth_service.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth_service.roles r
JOIN auth_service.permissions p ON p.name IN (
    'objects:read:own',
    'objects:create',
    'objects:update:own',
    'objects:delete:own'
)
WHERE r.name = 'user'
ON CONFLICT DO NOTHING;