-- Environment: development
-- Development Migration: Seed relationship-types permissions
-- Description: Assign relationship-types permissions to admin and object-type-admin roles
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
