-- Environment: development
-- Rollback: Reset test.user password to placeholder

UPDATE user_service.users 
SET password_hash = '$2a$10$testuser123.hash.placeholder',
    updated_at = CURRENT_TIMESTAMP
WHERE email = 'test.user@example.com';
