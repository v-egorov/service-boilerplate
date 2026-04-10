-- Environment: all
-- Migration: Add objects-service permissions
-- Description: Insert fine-grained permissions for objects-service RBAC
-- Applies to: All environments (development, staging, production)

-- Insert permissions for object types management
INSERT INTO auth_service.permissions (name, resource, action) VALUES
    ('object-types:create', 'object-types', 'create'),
    ('object-types:read', 'object-types', 'read'),
    ('object-types:update', 'object-types', 'update'),
    ('object-types:delete', 'object-types', 'delete')

ON CONFLICT (name) DO NOTHING;

-- Insert permissions for objects management
INSERT INTO auth_service.permissions (name, resource, action) VALUES
    ('objects:create', 'objects', 'create'),
    ('objects:read:all', 'objects', 'read:all'),
    ('objects:read:own', 'objects', 'read:own'),
    ('objects:update:all', 'objects', 'update:all'),
    ('objects:update:own', 'objects', 'update:own'),
    ('objects:delete:all', 'objects', 'delete:all'),
    ('objects:delete:own', 'objects', 'delete:own')

ON CONFLICT (name) DO NOTHING;
