-- Development Migration: Remove object-type-admin role and assignments (rollback)
-- Description: Rollback role assignments for dev/staging
-- Applies to: Development and Staging environments

-- Remove object-type-admin role from users
DELETE FROM auth_service.user_roles
WHERE role_id IN (
    SELECT id FROM auth_service.roles WHERE name = 'object-type-admin'
);

-- Remove role-permission assignments for object-type-admin role
DELETE FROM auth_service.role_permissions
WHERE role_id IN (
    SELECT id FROM auth_service.roles WHERE name = 'object-type-admin'
);

-- Remove role-permission assignments for user role (objects-service permissions only)
DELETE FROM auth_service.role_permissions
WHERE role_id IN (
    SELECT id FROM auth_service.roles WHERE name = 'user'
) AND permission_id IN (
    SELECT id FROM auth_service.permissions
    WHERE name LIKE 'object-types:%'
       OR name LIKE 'objects:%'
);

-- Remove role-permission assignments for admin role (objects-service permissions only)
DELETE FROM auth_service.role_permissions
WHERE role_id IN (
    SELECT id FROM auth_service.roles WHERE name = 'admin'
) AND permission_id IN (
    SELECT id FROM auth_service.permissions
    WHERE name LIKE 'object-types:%'
       OR name LIKE 'objects:%'
);

-- Delete object-type-admin role
DELETE FROM auth_service.roles WHERE name = 'object-type-admin';
