-- Development Migration: Add test data for development
-- This migration only runs in development environment
-- Created: $(date)

-- Insert development test users
INSERT INTO user_service.users (email, first_name, last_name) VALUES
    ('dev.admin@example.com', 'Dev', 'Admin'),
    ('test.user@example.com', 'Test', 'User'),
    ('qa.tester@example.com', 'QA', 'Tester')
ON CONFLICT (email) DO NOTHING;

-- Add development-specific indexes for testing
CREATE INDEX IF NOT EXISTS idx_users_dev_email ON user_service.users(email) WHERE email LIKE 'dev.%';
CREATE INDEX IF NOT EXISTS idx_users_test_email ON user_service.users(email) WHERE email LIKE 'test.%';