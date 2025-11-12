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