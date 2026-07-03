-- Environment: development only
-- Migration: Add MCP agent read-only role and permissions for objects-service
-- Description: Creates a dedicated role for the mcp-server system account that grants
-- read-only access to all object types and objects. This is assigned to the mcp-agent
-- user (created in user-service migration 000007) so that permiddleware permission checks
-- succeed without requiring per-user JWT validation.

-- Create MCP agent read-only role (if not exists from previous runs)
INSERT INTO auth_service.roles (name, description) VALUES
    ('mcp-agent-read-only', 'Read-only access for mcp-server internal service account')
ON CONFLICT (name) DO NOTHING;
