-- Environment: development only
-- Migration: Add MCP agent system user for internal service-to-service communication
-- Description: Creates a system-level user account used by mcp-server to make authenticated
-- requests through objects-service. This allows the auth-service permission check chain to work
-- without needing per-user JWT validation or special-casing in permiddleware.
-- Applies to: Development environment only

-- Create mcp-agent system user with deterministic UUID for reproducibility
INSERT INTO user_service.users (id, email, first_name, last_name, password_hash) VALUES
    ('00000000-0000-4000-8000-000000000001', 'mcp-agent@system.internal', 'MCP', 'Agent', '$2a$10$OUymIhBsngVFUOY7FldRhekCex3hts/jK1m7W6HJYR1vY5ofa2uKy')
ON CONFLICT (id) DO NOTHING;
