-- Environment: development only
-- Migration: Remove MCP agent system user (rollback)

DELETE FROM user_service.users WHERE id = '00000000-0000-4000-8000-000000000001';
