-- Environment: development only
-- Migration: Remove MCP agent permissions and user-role assignment (rollback)

DELETE FROM auth_service.user_roles
WHERE role_id IN (SELECT id FROM auth_service.roles WHERE name = 'mcp-agent-read-only');

DELETE FROM auth_service.role_permissions
WHERE permission_id IN (
    SELECT id FROM auth_service.permissions
    WHERE name IN ('objects:read:all', 'objects:read:own')
);
