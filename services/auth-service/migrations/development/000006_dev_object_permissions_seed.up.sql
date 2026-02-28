-- Development Migration: Seed object-type-admin role and assign permissions
-- Description: Create object-type-admin role and assign permissions for dev/staging
-- Applies to: Development and Staging environments

-- Create object-type-admin role
INSERT INTO auth_service.roles (name, description) VALUES
    ('object-type-admin', 'Dedicated role for managing object types and objects')
ON CONFLICT (name) DO NOTHING;

-- Assign all permissions to admin role
INSERT INTO auth_service.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth_service.roles r
CROSS JOIN auth_service.permissions p
WHERE r.name = 'admin'
  AND (
    p.name LIKE 'object-types:%'
    OR p.name LIKE 'objects:%'
  )
ON CONFLICT DO NOTHING;

-- Assign all object-types permissions to object-type-admin role
INSERT INTO auth_service.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth_service.roles r
JOIN auth_service.permissions p ON p.name LIKE 'object-types:%'
WHERE r.name = 'object-type-admin'
ON CONFLICT DO NOTHING;

-- Assign objects:read:all and objects:read:own to object-type-admin role
INSERT INTO auth_service.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth_service.roles r
JOIN auth_service.permissions p ON p.name IN ('objects:read:all', 'objects:read:own')
WHERE r.name = 'object-type-admin'
ON CONFLICT DO NOTHING;

-- Assign objects:update:all and objects:delete:all to object-type-admin role
INSERT INTO auth_service.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth_service.roles r
JOIN auth_service.permissions p ON p.name IN ('objects:update:all', 'objects:delete:all')
WHERE r.name = 'object-type-admin'
ON CONFLICT DO NOTHING;

-- Assign basic permissions to user role
INSERT INTO auth_service.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth_service.roles r
JOIN auth_service.permissions p ON p.name IN (
    'object-types:read',
    'objects:create',
    'objects:read:own',
    'objects:update:own',
    'objects:delete:own'
)
WHERE r.name = 'user'
ON CONFLICT DO NOTHING;

-- Assign object-type-admin role to dev admin user
INSERT INTO auth_service.user_roles (user_id, role_id)
SELECT u.id, r.id
FROM user_service.users u
JOIN auth_service.roles r ON r.name = 'object-type-admin'
WHERE u.email = 'dev.admin@example.com'
ON CONFLICT DO NOTHING;

-- Assign object-type-admin role to object admin user
INSERT INTO auth_service.user_roles (user_id, role_id)
SELECT u.id, r.id
FROM user_service.users u
JOIN auth_service.roles r ON r.name = 'object-type-admin'
WHERE u.email = 'object.admin@example.com'
ON CONFLICT DO NOTHING;
