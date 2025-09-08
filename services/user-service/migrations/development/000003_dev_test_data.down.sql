-- Development Migration Rollback: Remove test data
-- This rollback only runs in development environment
-- Created: $(date)

-- Remove development test users
DELETE FROM user_service.users
WHERE email IN (
    'dev.admin@example.com',
    'test.user@example.com',
    'qa.tester@example.com'
);

-- Remove development-specific indexes
DROP INDEX IF EXISTS user_service.idx_users_dev_email;
DROP INDEX IF EXISTS user_service.idx_users_test_email;