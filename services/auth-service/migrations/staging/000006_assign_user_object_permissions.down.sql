-- Environment: staging
-- Staging Migration: Remove object permissions from user role (rollback)

-- Remove object permissions from user role
DELETE FROM auth_service.role_permissions
WHERE role_id = (SELECT id FROM auth_service.roles WHERE name = 'user')
  AND permission_id IN (
      SELECT id FROM auth_service.permissions WHERE name IN (
          'object-types:read',
          'objects:read:own',
          'objects:create',
          'objects:update:own',
          'objects:delete:own'
      )
  );