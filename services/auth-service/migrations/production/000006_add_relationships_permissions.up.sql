-- Add relationships scoped permissions for relationship instances (R2.14)
-- Scoped variants required for all CRUD actions on relationships

INSERT INTO auth_service.permissions (name, resource, action) VALUES
    ('relationships:create:own',     'relationships', 'create'),
    ('relationships:create:all',     'relationships', 'create'),
    ('relationships:read:own',       'relationships', 'read'),
    ('relationships:read:all',       'relationships', 'read'),
    ('relationships:update:own',     'relationships', 'update'),
    ('relationships:update:all',     'relationships', 'update'),
    ('relationships:delete:own',     'relationships', 'delete'),
    ('relationships:delete:all',     'relationships', 'delete')
ON CONFLICT (name) DO NOTHING;