-- Development Admin Account Setup
-- This migration sets up the dev admin account with proper RBAC roles
-- Created for development and testing purposes

-- Assign admin role to dev admin user
INSERT INTO auth_service.user_roles (user_id, role_id)
SELECT u.id, r.id
FROM user_service.users u
CROSS JOIN auth_service.roles r
WHERE u.email = 'dev.admin@example.com' AND r.name = 'admin'
ON CONFLICT DO NOTHING;

-- Assign user role to dev admin user (for completeness)
INSERT INTO auth_service.user_roles (user_id, role_id)
SELECT u.id, r.id
FROM user_service.users u
CROSS JOIN auth_service.roles r
WHERE u.email = 'dev.admin@example.com' AND r.name = 'user'
ON CONFLICT DO NOTHING;

-- Assign user role to test.user@example.com
INSERT INTO auth_service.user_roles (user_id, role_id)
SELECT u.id, r.id
FROM user_service.users u
CROSS JOIN auth_service.roles r
WHERE u.email = 'test.user@example.com' AND r.name = 'user'
ON CONFLICT DO NOTHING;

-- Assign user role to qa.tester@example.com
INSERT INTO auth_service.user_roles (user_id, role_id)
SELECT u.id, r.id
FROM user_service.users u
CROSS JOIN auth_service.roles r
WHERE u.email = 'qa.tester@example.com' AND r.name = 'user'
ON CONFLICT DO NOTHING;