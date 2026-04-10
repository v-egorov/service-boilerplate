-- Environment: development
-- Development Admin Account Setup - Down Migration
-- This removes the dev admin account RBAC assignments
-- Note: object.admin@example.com is NOT removed here - handled by 000005_dev_object_permissions_seed.down.sql

-- Remove all role assignments for dev admin user
DELETE FROM auth_service.user_roles
WHERE user_id IN (
    SELECT id FROM user_service.users
    WHERE email IN ('dev.admin@example.com', 'test.user@example.com', 'qa.tester@example.com')
);