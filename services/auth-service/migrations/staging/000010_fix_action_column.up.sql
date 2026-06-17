-- Environment: all
-- Migration: Fix action column scope suffixes on objects permissions
-- Description: Strip scope suffixes (:all, :own) from action column for objects-scoped permissions
-- Applies to: All environments (development, staging, production)

UPDATE auth_service.permissions SET action = 'read' WHERE resource = 'objects' AND action IN ('read:all', 'read:own');
UPDATE auth_service.permissions SET action = 'update' WHERE resource = 'objects' AND action IN ('update:all', 'update:own');
UPDATE auth_service.permissions SET action = 'delete' WHERE resource = 'objects' AND action IN ('delete:all', 'delete:own');
