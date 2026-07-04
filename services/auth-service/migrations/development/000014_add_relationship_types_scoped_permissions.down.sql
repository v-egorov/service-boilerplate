-- Environment: development only
-- Migration: Remove scoped relationship-types permissions (rollback)

DELETE FROM auth_service.permissions WHERE name IN (
    'relationship-types:read:all',
    'relationship-types:read:own',
    'relationship-types:update:all',
    'relationship-types:update:own',
    'relationship-types:delete:all',
    'relationship-types:delete:own'
);
