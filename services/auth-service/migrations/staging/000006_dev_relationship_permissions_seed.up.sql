-- Environment: development
-- Development Migration: Seed relationship-types and object permissions
-- Description: Assign relationship-types permissions to admin and object-type-admin roles, and basic object permissions to user role
-- Applies to: Development environment

-- Assign all relationship-types permissions to admin role
INSERT INTO auth_service.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth_service.roles r
CROSS JOIN auth_service.permissions p
WHERE r.name = 'admin'
  AND p.name LIKE 'relationship-types:%'
ON CONFLICT DO NOTHING;

-- Assign all relationship-types permissions to object-type-admin role
INSERT INTO auth_service.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth_service.roles r
CROSS JOIN auth_service.permissions p
WHERE r.name = 'object-type-admin'
  AND p.name LIKE 'relationship-types:%'
ON CONFLICT DO NOTHING;

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
