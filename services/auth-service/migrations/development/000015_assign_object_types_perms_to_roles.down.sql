-- Environment: development only
-- Migration: Remove object-types role permission assignments (rollback)

DELETE FROM auth_service.role_permissions
WHERE permission_id IN (
    SELECT id FROM auth_service.permissions
    WHERE name LIKE 'object-types:%'
      AND name NOT LIKE '%create%'  -- keep flat create permissions
);
