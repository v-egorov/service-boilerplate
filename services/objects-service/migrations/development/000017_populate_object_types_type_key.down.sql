-- Environment: development
-- Rollback migration 000017: revert type_keys to NULL and drop NOT NULL constraint.

ALTER TABLE objects_service.object_types ALTER COLUMN type_key DROP NOT NULL;

UPDATE objects_service.object_types SET type_key = NULL WHERE name IN ('Electronics', 'Clothing', 'Books', 'News Article', 'Blog Post', 'Tutorial', 'Country', 'City');
