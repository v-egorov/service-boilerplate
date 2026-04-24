-- Add relationships permissions for relationship instances (R2.13)
-- This migration adds permissions for CRUD operations on relationship instances

INSERT INTO auth_service.permissions (name, resource, action) VALUES
    ('relationships:create', 'relationships', 'create'),
    ('relationships:read', 'relationships', 'read'),
    ('relationships:update', 'relationships', 'update'),
    ('relationships:delete', 'relationships', 'delete')
ON CONFLICT (name) DO NOTHING;