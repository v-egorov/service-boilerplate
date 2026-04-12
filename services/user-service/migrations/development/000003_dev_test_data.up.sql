-- Environment: development
-- Development Migration: Add test data for development
-- This migration only runs in development environment
-- Created: $(date)

-- Insert development test users
INSERT INTO user_service.users (email, first_name, last_name, password_hash) VALUES
    ('dev.admin@example.com', 'Dev', 'Admin', '$2a$10$OUymIhBsngVFUOY7FldRhekCex3hts/jK1m7W6HJYR1vY5ofa2uKy'),
    ('test.user@example.com', 'Test', 'User', '$2a$10$testuser123.hash.placeholder'),
    ('qa.tester@example.com', 'QA', 'Tester', '$2a$10$qatester123.hash.placeholder')
ON CONFLICT (email) DO UPDATE SET
    password_hash = EXCLUDED.password_hash;

-- Add development-specific indexes for testing
CREATE INDEX IF NOT EXISTS idx_users_dev_email ON user_service.users(email) WHERE email LIKE 'dev.%';
CREATE INDEX IF NOT EXISTS idx_users_test_email ON user_service.users(email) WHERE email LIKE 'test.%';

-- Assign roles to development users (cross-service dependency on auth_service)
INSERT INTO auth_service.user_roles (user_id, role_id)
SELECT u.id, r.id
FROM user_service.users u
CROSS JOIN auth_service.roles r
WHERE u.email = 'dev.admin@example.com' AND r.name = 'admin'
ON CONFLICT DO NOTHING;

INSERT INTO auth_service.user_roles (user_id, role_id)
SELECT u.id, r.id
FROM user_service.users u
CROSS JOIN auth_service.roles r
WHERE u.email = 'dev.admin@example.com' AND r.name = 'user'
ON CONFLICT DO NOTHING;

INSERT INTO auth_service.user_roles (user_id, role_id)
SELECT u.id, r.id
FROM user_service.users u
CROSS JOIN auth_service.roles r
WHERE u.email = 'test.user@example.com' AND r.name = 'user'
ON CONFLICT DO NOTHING;

INSERT INTO auth_service.user_roles (user_id, role_id)
SELECT u.id, r.id
FROM user_service.users u
CROSS JOIN auth_service.roles r
WHERE u.email = 'qa.tester@example.com' AND r.name = 'user'
ON CONFLICT DO NOTHING;

-- Create object-type-admin role (if not exists)
INSERT INTO auth_service.roles (name, description) VALUES
    ('object-type-admin', 'Dedicated role for managing object types and objects')
ON CONFLICT (name) DO NOTHING;

-- Assign all permissions to admin role
INSERT INTO auth_service.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth_service.roles r
CROSS JOIN auth_service.permissions p
WHERE r.name = 'admin'
  AND (
    p.name LIKE 'object-types:%'
    OR p.name LIKE 'objects:%'
  )
ON CONFLICT DO NOTHING;

-- Assign all object-types permissions to object-type-admin role
INSERT INTO auth_service.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth_service.roles r
JOIN auth_service.permissions p ON p.name LIKE 'object-types:%'
WHERE r.name = 'object-type-admin'
ON CONFLICT DO NOTHING;

-- Assign objects:read:all and objects:read:own to object-type-admin role
INSERT INTO auth_service.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth_service.roles r
JOIN auth_service.permissions p ON p.name IN ('objects:read:all', 'objects:read:own')
WHERE r.name = 'object-type-admin'
ON CONFLICT DO NOTHING;

-- Assign objects:update:all and objects:delete:all to object-type-admin role
INSERT INTO auth_service.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth_service.roles r
JOIN auth_service.permissions p ON p.name IN ('objects:update:all', 'objects:delete:all')
WHERE r.name = 'object-type-admin'
ON CONFLICT DO NOTHING;