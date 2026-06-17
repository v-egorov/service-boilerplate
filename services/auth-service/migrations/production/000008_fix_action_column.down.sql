-- Environment: all
-- Migration: Restore action column scope suffixes on objects permissions (rollback of 000012)
-- Description: Re-add scope suffixes to action column for objects-scoped permissions
-- Applies to: All environments (development, staging, production)

UPDATE auth_service.permissions SET action = 'read:all' WHERE resource = 'objects' AND name = 'objects:read:all';
UPDATE auth_service.permissions SET action = 'read:own' WHERE resource = 'objects' AND name = 'objects:read:own';
UPDATE auth_service.permissions SET action = 'update:all' WHERE resource = 'objects' AND name = 'objects:update:all';
UPDATE auth_service.permissions SET action = 'update:own' WHERE resource = 'objects' AND name = 'objects:update:own';
UPDATE auth_service.permissions SET action = 'delete:all' WHERE resource = 'objects' AND name = 'objects:delete:all';
UPDATE auth_service.permissions SET action = 'delete:own' WHERE resource = 'objects' AND name = 'objects:delete:own';
