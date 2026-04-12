-- Environment: development/staging
-- Development/Staging Migration: Add object admin user
-- Description: Create dedicated admin account for objects management
-- Applies to: Development and Staging environments only

INSERT INTO user_service.users (email, first_name, last_name, password_hash) VALUES 
    ('object.admin@example.com', 'Object', 'Admin', '$2a$10$OUymIhBsngVFUOY7FldRhekCex3hts/jK1m7W6HJYR1vY5ofa2uKy')
ON CONFLICT (email) DO NOTHING;

-- Assign object-type-admin role to object.admin user
INSERT INTO auth_service.user_roles (user_id, role_id)
SELECT u.id, r.id
FROM user_service.users u
CROSS JOIN auth_service.roles r
WHERE u.email = 'object.admin@example.com' AND r.name = 'object-type-admin'
ON CONFLICT DO NOTHING;
