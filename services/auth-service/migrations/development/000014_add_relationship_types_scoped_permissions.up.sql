-- Environment: development only
-- Migration: Add scoped permission variants for relationship-types (R2.14)
-- Description: Inserts scoped read/update/delete permissions for relationship-types that were
-- defined in the architecture spec but never added during June 2026 type_key implementation.
-- Create action stays flat per architecture spec — you always own what you create.

INSERT INTO auth_service.permissions (name, resource, action) VALUES
    ('relationship-types:read:all',   'relationship-types', 'read'),
    ('relationship-types:read:own',   'relationship-types', 'read'),
    ('relationship-types:update:all',  'relationship-types', 'update'),
    ('relationship-types:update:own',  'relationship-types', 'update'),
    ('relationship-types:delete:all',  'relationship-types', 'delete'),
    ('relationship-types:delete:own',  'relationship-types', 'delete')

ON CONFLICT (name) DO NOTHING;
