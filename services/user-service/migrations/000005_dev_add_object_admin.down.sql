-- Environment: development
-- Development Migration: Remove object admin user (rollback)
-- Description: Remove dedicated admin account for objects management

DELETE FROM user_service.users WHERE email = 'object.admin@example.com';
