-- Revert type_key values to NULL for the 8 rows populated in migration 000017
UPDATE objects_service.object_types SET type_key = NULL WHERE id IN (5, 6, 7, 8, 9, 10, 11, 12);
