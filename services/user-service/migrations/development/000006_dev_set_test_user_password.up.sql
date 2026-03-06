-- Development Migration: Set known password for test.user
-- Description: Set test.user password to known value for RBAC testing
-- Uses same hash as dev.admin (password: devadmin123)

UPDATE user_service.users 
SET password_hash = '$2a$10$OUymIhBsngVFUOY7FldRhekCex3hts/jK1m7W6HJYR1vY5ofa2uKy',
    updated_at = CURRENT_TIMESTAMP
WHERE email = 'test.user@example.com';
