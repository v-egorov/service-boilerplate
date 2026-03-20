-- Remove relationship-types permissions
DELETE FROM auth_service.permissions WHERE name LIKE 'relationship-types:%';
