-- Migration: Remove objects-service permissions (rollback)
-- Description: Remove fine-grained permissions for objects-service RBAC
-- Applies to: All environments (development, staging, production)

-- Remove all role-permission assignments for objects-service permissions
DELETE FROM auth_service.role_permissions
WHERE permission_id IN (
    SELECT id FROM auth_service.permissions
    WHERE name LIKE 'object-types:%'
       OR name LIKE 'objects:%'
);

-- Remove permissions
DELETE FROM auth_service.permissions
WHERE name LIKE 'object-types:%'
   OR name LIKE 'objects:%';
