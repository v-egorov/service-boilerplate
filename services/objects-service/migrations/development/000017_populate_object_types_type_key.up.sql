-- Populate NULL type_key values in object_types table
-- 8 rows have name but no type_key — these cause pgx ScanArgError when scanning into Go string fields.
-- Using lowercase, hyphenated values consistent with existing type_keys (e.g., "blog-post", "news-article").

UPDATE objects_service.object_types SET type_key = 'electronics' WHERE id = 5;
UPDATE objects_service.object_types SET type_key = 'clothing'    WHERE id = 6;
UPDATE objects_service.object_types SET type_key = 'books'       WHERE id = 7;
UPDATE objects_service.object_types SET type_key = 'news-article' WHERE id = 8;
UPDATE objects_service.object_types SET type_key = 'blog-post'   WHERE id = 9;
UPDATE objects_service.object_types SET type_key = 'tutorial'    WHERE id = 10;
UPDATE objects_service.object_types SET type_key = 'country'     WHERE id = 11;
UPDATE objects_service.object_types SET type_key = 'city'        WHERE id = 12;
