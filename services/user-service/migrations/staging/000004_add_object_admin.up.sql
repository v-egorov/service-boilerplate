-- Environment: development/staging
-- Development/Staging Migration: Add object admin user
-- Description: Create dedicated admin account for objects management
-- Applies to: Development and Staging environments only

INSERT INTO user_service.users (email, first_name, last_name, password_hash) VALUES 
    ('object.admin@example.com', 'Object', 'Admin', '$2a$10$OUymIhBsngVFUOY7FldRhekCex3hts/jK1m7W6HJYR1vY5ofa2uKy')
ON CONFLICT (email) DO NOTHING;
