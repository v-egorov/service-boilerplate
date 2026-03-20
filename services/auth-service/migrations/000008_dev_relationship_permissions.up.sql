-- Migration: Add relationship-types permissions
-- Description: Insert permissions for relationship-types management
-- Applies to: All environments

INSERT INTO auth_service.permissions (name, resource, action) VALUES
    ('relationship-types:create', 'relationship-types', 'create'),
    ('relationship-types:read', 'relationship-types', 'read'),
    ('relationship-types:update', 'relationship-types', 'update'),
    ('relationship-types:delete', 'relationship-types', 'delete')

ON CONFLICT (name) DO NOTHING;
