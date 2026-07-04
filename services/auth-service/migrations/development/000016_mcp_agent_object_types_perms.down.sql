-- Environment: development only
-- Migration: Remove object-types permission from mcp-agent-read-only role (rollback)

DELETE FROM auth_service.role_permissions
WHERE role_id IN (SELECT id FROM auth_service.roles WHERE name = 'mcp-agent-read-only')
  AND permission_id IN (SELECT id FROM auth_service.permissions WHERE name = 'object-types:read:all');
