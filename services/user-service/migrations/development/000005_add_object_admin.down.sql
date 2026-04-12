-- Environment: development/staging
-- Development/Staging Migration: Remove object admin user (rollback)
-- Description: Remove dedicated admin account for objects management

-- Remove role assignment first (FK dependency)
DELETE FROM auth_service.user_roles 
WHERE user_id = (SELECT id FROM user_service.users WHERE email = 'object.admin@example.com')
  AND role_id = (SELECT id FROM auth_service.roles WHERE name = 'object-type-admin');

-- Then remove the user
DELETE FROM user_service.users WHERE email = 'object.admin@example.com';
