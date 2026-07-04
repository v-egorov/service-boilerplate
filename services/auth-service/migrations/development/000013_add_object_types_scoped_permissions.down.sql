-- Environment: development only
-- Migration: Remove scoped object-types permissions (rollback)

DELETE FROM auth_service.permissions WHERE name IN (
    'object-types:read:all',
    'object-types:read:own',
    'object-types:update:all',
    'object-types:update:own',
    'object-types:delete:all',
    'object-types:delete:own'
);
