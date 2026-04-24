-- Assign relationships permissions to roles (R2.14)
-- This migration grants relationship instance permissions to admin and object-type-admin roles

-- Assign relationships permissions to admin role
INSERT INTO auth_service.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth_service.roles r
CROSS JOIN auth_service.permissions p
WHERE r.name = 'admin'
  AND p.name LIKE 'relationships:%'
ON CONFLICT DO NOTHING;

-- Assign relationships permissions to object-type-admin role
INSERT INTO auth_service.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth_service.roles r
CROSS JOIN auth_service.permissions p
WHERE r.name = 'object-type-admin'
  AND p.name LIKE 'relationships:%'
ON CONFLICT DO NOTHING;