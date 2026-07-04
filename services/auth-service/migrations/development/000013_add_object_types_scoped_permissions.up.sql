-- Environment: development only
-- Migration: Add scoped permission variants for object-types (R2.14)
-- Description: Inserts scoped read/update/delete permissions for object-types that were
-- defined in the architecture spec but never added during June 2026 type_key implementation.
-- Create action stays flat per architecture spec — you always own what you create.

INSERT INTO auth_service.permissions (name, resource, action) VALUES
    ('object-types:read:all',     'object-types', 'read'),
    ('object-types:read:own',     'object-types', 'read'),
    ('object-types:update:all',   'object-types', 'update'),
    ('object-types:update:own',   'object-types', 'update'),
    ('object-types:delete:all',   'object-types', 'delete'),
    ('object-types:delete:own',   'object-types', 'delete')

ON CONFLICT (name) DO NOTHING;
