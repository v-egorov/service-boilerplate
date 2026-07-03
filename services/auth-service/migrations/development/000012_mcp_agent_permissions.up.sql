-- Environment: development only
-- Migration: Assign MCP agent read-only permissions and link to user account
-- Description: Inserts the required permission entries for objects-service RBAC (if not already present),
-- assigns them to mcp-agent-read-only role, and links the role to the mcp-agent system user.

-- Insert read-only permissions for objects-service if they don't exist
INSERT INTO auth_service.permissions (name, resource, action) VALUES
    ('objects:read:all', 'objects', 'read:all'),
    ('objects:read:own', 'objects', 'read:own')

ON CONFLICT (name) DO NOTHING;

-- Assign read-only permissions to mcp-agent-read-only role
INSERT INTO auth_service.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth_service.roles r
CROSS JOIN auth_service.permissions p
WHERE r.name = 'mcp-agent-read-only'
  AND p.name IN ('objects:read:all', 'objects:read:own')
ON CONFLICT DO NOTHING;

-- Link mcp-agent user to mcp-agent-read-only role
INSERT INTO auth_service.user_roles (user_id, role_id)
SELECT u.id, r.id
FROM user_service.users u
CROSS JOIN auth_service.roles r
WHERE u.email = 'mcp-agent@system.internal' AND r.name = 'mcp-agent-read-only'
ON CONFLICT DO NOTHING;
