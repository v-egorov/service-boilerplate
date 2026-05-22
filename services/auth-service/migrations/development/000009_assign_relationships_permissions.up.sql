-- Assign relationships scoped permissions to roles (R2.14)
-- Scoped variants: admin gets :all, user gets :own, relationship-admin gets same as admin,
-- relationship-viewer gets only read:all. object-type-admin removed from relationships entirely.

-- Ensure new roles exist before permission assignments (roles created in 000010, but included here for idempotent rollback/reapply)
INSERT INTO auth_service.roles (name, description) VALUES
    ('relationship-admin', 'Dedicated role for full relationship instance management'),
    ('relationship-viewer', 'Read-only access to relationship instances for audit and discovery')
ON CONFLICT (name) DO NOTHING;

-- Assign relationships scoped permissions to admin role (all CRUD :all)
INSERT INTO auth_service.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth_service.roles r
CROSS JOIN auth_service.permissions p
WHERE r.name = 'admin'
  AND p.name IN ('relationships:create:all', 'relationships:read:all', 'relationships:update:all', 'relationships:delete:all')
ON CONFLICT DO NOTHING;

-- Assign relationships scoped permissions to user role (all CRUD :own)
INSERT INTO auth_service.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth_service.roles r
CROSS JOIN auth_service.permissions p
WHERE r.name = 'user'
  AND p.name IN ('relationships:create:own', 'relationships:read:own', 'relationships:update:own', 'relationships:delete:own')
ON CONFLICT DO NOTHING;

-- Assign relationships scoped permissions to relationship-admin role (all CRUD :all, same as admin)
INSERT INTO auth_service.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth_service.roles r
CROSS JOIN auth_service.permissions p
WHERE r.name = 'relationship-admin'
  AND p.name IN ('relationships:create:all', 'relationships:read:all', 'relationships:update:all', 'relationships:delete:all')
ON CONFLICT DO NOTHING;

-- Assign relationships scoped permissions to relationship-viewer role (read-only :all)
INSERT INTO auth_service.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth_service.roles r
CROSS JOIN auth_service.permissions p
WHERE r.name = 'relationship-viewer'
  AND p.name = 'relationships:read:all'
ON CONFLICT DO NOTHING;

-- Remove object-type-admin from relationships (explicit cleanup — no relationship permissions for this role)
DELETE FROM auth_service.role_permissions rp
USING auth_service.roles r
WHERE r.id = rp.role_id
  AND r.name = 'object-type-admin'
  AND EXISTS (SELECT 1 FROM auth_service.permissions p WHERE p.id = rp.permission_id AND p.resource = 'relationships');
