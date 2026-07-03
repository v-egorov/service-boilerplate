-- Environment: development only
-- Migration: Remove MCP agent read-only role (rollback)

DELETE FROM auth_service.user_roles
WHERE role_id IN (SELECT id FROM auth_service.roles WHERE name = 'mcp-agent-read-only');

DELETE FROM auth_service.roles WHERE name = 'mcp-agent-read-only';
