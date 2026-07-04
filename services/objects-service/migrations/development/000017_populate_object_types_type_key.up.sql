-- Environment: development
-- Populate NULL type_key values for child object types using name-based lookups.
-- Naming convention: <parent_type_key>-<child_name_slugified>
-- After population, enforce NOT NULL to prevent future NULLs.

UPDATE objects_service.object_types SET type_key = 'product-electronics'  WHERE name = 'Electronics';
UPDATE objects_service.object_types SET type_key = 'product-clothing'     WHERE name = 'Clothing';
UPDATE objects_service.object_types SET type_key = 'product-books'        WHERE name = 'Books';
UPDATE objects_service.object_types SET type_key = 'article-news-article' WHERE name = 'News Article';
UPDATE objects_service.object_types SET type_key = 'article-blog-post'    WHERE name = 'Blog Post';
UPDATE objects_service.object_types SET type_key = 'article-tutorial'     WHERE name = 'Tutorial';
UPDATE objects_service.object_types SET type_key = 'location-country'     WHERE name = 'Country';
UPDATE objects_service.object_types SET type_key = 'location-city'        WHERE name = 'City';

-- Enforce NOT NULL — will fail if any row still has a NULL (transaction-safe)
ALTER TABLE objects_service.object_types ALTER COLUMN type_key SET NOT NULL;
