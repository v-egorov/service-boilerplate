-- Development Admin Account Setup - Down Migration
-- This removes the dev admin account RBAC assignments

-- Remove all role assignments for dev admin user
DELETE FROM auth_service.user_roles
WHERE user_id IN (
    SELECT id FROM user_service.users
    WHERE email IN ('dev.admin@example.com', 'test.user@example.com', 'qa.tester@example.com')
);