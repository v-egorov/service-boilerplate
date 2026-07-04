-- Environment: development only
-- Migration: Add object-types read permission to mcp-agent-read-only role
-- Description: The MCP agent needs object-types:read:all to call list_object_types through
-- the normal RBAC path (not just the gateway-trust skip path). This complements the existing
-- objects:read:all/own permissions that were added in 000012.

-- Insert object-types scoped read permission if it doesn't exist
INSERT INTO auth_service.permissions (name, resource, action) VALUES
    ('object-types:read:all', 'object-types', 'read')
ON CONFLICT (name) DO NOTHING;

-- Assign to mcp-agent-read-only role
INSERT INTO auth_service.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth_service.roles r
CROSS JOIN auth_service.permissions p
WHERE r.name = 'mcp-agent-read-only'
  AND p.name IN ('object-types:read:all')
ON CONFLICT DO NOTHING;
