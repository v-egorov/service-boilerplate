-- Environment: development only
-- Migration: Assign object-types scoped permissions to roles (R2.14)
-- Description: Per architecture spec, admin and object-type-admin get :all on all actions;
-- user gets read:own only. Create stays flat per architecture spec.

-- Ensure required roles exist before permission assignments
INSERT INTO auth_service.roles (name, description) VALUES
    ('admin', 'Full system administrator'),
    ('object-type-admin', 'Administrator for object type definitions and schema management')
ON CONFLICT (name) DO NOTHING;

-- Assign :all scoped permissions to admin role (read, update, delete — all :all)
INSERT INTO auth_service.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth_service.roles r
CROSS JOIN auth_service.permissions p
WHERE r.name = 'admin'
  AND p.name IN ('object-types:read:all', 'object-types:update:all', 'object-types:delete:all')
ON CONFLICT DO NOTHING;

-- Assign :all scoped permissions to object-type-admin role (same as admin)
INSERT INTO auth_service.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth_service.roles r
CROSS JOIN auth_service.permissions p
WHERE r.name = 'object-type-admin'
  AND p.name IN ('object-types:read:all', 'object-types:update:all', 'object-types:delete:all')
ON CONFLICT DO NOTHING;

-- Assign :own scoped permissions to user role (read only — no update/delete)
INSERT INTO auth_service.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth_service.roles r
CROSS JOIN auth_service.permissions p
WHERE r.name = 'user'
  AND p.name IN ('object-types:read:own')
ON CONFLICT DO NOTHING;
