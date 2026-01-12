# Phase 7: Development Test Data

**Estimated Time**: 1 hour
**Status**: ⬜ Not Started
**Dependencies**: Phase 6 (Main Application)

## Overview

Create development migration with sample taxonomy data for testing and demonstration. This data will be automatically loaded in development environment.

## Tasks

### 7.1 Create Development Test Data Migration

**File**: `migrations/development/000002_dev_tax_test_data.up.sql`

**Steps**:
1. Create sample object types hierarchy
2. Create sample objects of various types
3. Include various relationships and metadata
4. Test different statuses and tags

```sql
-- Development Test Data for Objects Service
-- This migration creates sample taxonomy data for testing

-- Insert Root Object Types
INSERT INTO object_types (name, parent_type_id, concrete_table_name, description, is_sealed, metadata) VALUES
('Category', NULL, NULL, 'Root category for taxonomic classification', false, '{"validation_schema": {"type": "object", "properties": {"icon": {"type": "string"}}}}'),
('Document', NULL, NULL, 'Base type for all documents', false, '{"validation_schema": {"type": "object", "properties": {"mime_type": {"type": "string"}, "size_bytes": {"type": "integer"}}}}'),
('Location', NULL, NULL, 'Geographic locations', true, '{"validation_schema": {"type": "object", "properties": {"latitude": {"type": "number"}, "longitude": {"type": "number"}}}}')
ON CONFLICT (name) DO NOTHING;

-- Get type IDs for reference (will be replaced by actual IDs in migration)
DO $$
DECLARE
    category_id BIGINT;
    document_id BIGINT;
    location_id BIGINT;
BEGIN
    SELECT id INTO category_id FROM object_types WHERE name = 'Category';
    SELECT id INTO document_id FROM object_types WHERE name = 'Document';
    SELECT id INTO location_id FROM object_types WHERE name = 'Location';
    
    -- Insert Sub-Types for Category
    INSERT INTO object_types (name, parent_type_id, description, is_sealed, metadata) VALUES
    ('Product Category', category_id, 'Product categorization', false, '{"icon": "product", "color": "#FF5722"}'),
    ('Service Category', category_id, 'Service categorization', false, '{"icon": "service", "color": "#2196F3"}'),
    ('Resource Category', category_id, 'Resource categorization', false, '{"icon": "resource", "color": "#4CAF50"}')
    ON CONFLICT (name) DO NOTHING;
    
    -- Insert Sub-Types for Document
    INSERT INTO object_types (name, parent_type_id, description, is_sealed, metadata) VALUES
    ('PDF Document', document_id, 'PDF formatted documents', false, '{"mime_type": "application/pdf"}'),
    ('Image Document', document_id, 'Image files', false, '{"mime_types": ["image/jpeg", "image/png", "image/gif"]}'),
    ('Text Document', document_id, 'Plain text documents', false, '{"mime_type": "text/plain"}')
    ON CONFLICT (name) DO NOTHING;
    
    -- Insert Sample Objects for Categories
    INSERT INTO objects (object_type_id, parent_object_id, name, description, metadata, status, tags, created_by, updated_by) VALUES
    ((SELECT id FROM object_types WHERE name = 'Product Category'), NULL, 'Electronics', 'Electronic products and accessories', '{"icon": "electronics", "priority": 1}', 'active', ARRAY['products', 'hardware'], 'system', 'system'),
    ((SELECT id FROM object_types WHERE name = 'Product Category'), NULL, 'Clothing', 'Apparel and fashion items', '{"icon": "clothing", "priority": 2}', 'active', ARRAY['products', 'fashion'], 'system', 'system'),
    ((SELECT id FROM object_types WHERE name = 'Service Category'), NULL, 'Consulting', 'Professional consulting services', '{"icon": "consulting", "rate_range": "high"}', 'active', ARRAY['services', 'business'], 'system', 'system'),
    ((SELECT id FROM object_types WHERE name = 'Service Category'), NULL, 'Support', 'Technical support services', '{"icon": "support", "sla": "24h"}', 'active', ARRAY['services', 'technical'], 'system', 'system')
    ON CONFLICT DO NOTHING;
    
    -- Insert Child Categories
    INSERT INTO objects (object_type_id, parent_object_id, name, description, metadata, status, tags, created_by, updated_by) VALUES
    ((SELECT id FROM object_types WHERE name = 'Product Category'), (SELECT id FROM objects WHERE name = 'Electronics'), 'Smartphones', 'Mobile phone devices', '{"subcategories": ["ios", "android"]}', 'active', ARRAY['mobile', 'phones'], 'system', 'system'),
    ((SELECT id FROM object_types WHERE name = 'Product Category'), (SELECT id FROM objects WHERE name = 'Electronics'), 'Laptops', 'Portable computers', '{"subcategories": ["ultrabook", "gaming"]}', 'active', ARRAY['computers', 'portable'], 'system', 'system'),
    ((SELECT id FROM object_types WHERE name = 'Product Category'), (SELECT id FROM objects WHERE name = 'Clothing'), 'Men''s Wear', 'Clothing for men', '{"seasonal": true}', 'active', ARRAY['clothing', 'men'], 'system', 'system'),
    ((SELECT id FROM object_types WHERE name = 'Product Category'), (SELECT id FROM objects WHERE name = 'Clothing'), 'Women''s Wear', 'Clothing for women', '{"seasonal": true}', 'active', ARRAY['clothing', 'women'], 'system', 'system')
    ON CONFLICT DO NOTHING;
    
    -- Insert Sample Documents
    INSERT INTO objects (object_type_id, parent_object_id, name, description, metadata, status, tags, created_by, updated_by) VALUES
    ((SELECT id FROM object_types WHERE name = 'PDF Document'), NULL, 'Product Specification', 'Technical specifications document', '{"mime_type": "application/pdf", "size_bytes": 245760}', 'active', ARRAY['documentation', 'technical'], 'system', 'system'),
    ((SELECT id FROM object_types WHERE name = 'Image Document'), NULL, 'Product Photo', 'Main product image', '{"mime_type": "image/jpeg", "size_bytes": 1048576, "width": 1920, "height": 1080}', 'active', ARRAY['image', 'product'], 'system', 'system'),
    ((SELECT id FROM object_types WHERE name = 'Text Document'), NULL, 'README', 'Project readme file', '{"mime_type": "text/plain", "size_bytes": 8192}', 'active', ARRAY['documentation', 'readme'], 'system', 'system')
    ON CONFLICT DO NOTHING;
    
    -- Insert Sample Locations (sealed type)
    INSERT INTO objects (object_type_id, parent_object_id, name, description, metadata, status, tags, created_by, updated_by) VALUES
    ((SELECT id FROM object_types WHERE name = 'Location'), NULL, 'San Francisco', 'San Francisco, California', '{"latitude": 37.7749, "longitude": -122.4194, "country": "USA"}', 'active', ARRAY['city', 'usa'], 'system', 'system'),
    ((SELECT id FROM object_types WHERE name = 'Location'), NULL, 'New York', 'New York City, New York', '{"latitude": 40.7128, "longitude": -74.0060, "country": "USA"}', 'active', ARRAY['city', 'usa'], 'system', 'system'),
    ((SELECT id FROM object_types WHERE name = 'Location'), NULL, 'London', 'London, United Kingdom', '{"latitude": 51.5074, "longitude": -0.1278, "country": "UK"}', 'active', ARRAY['city', 'europe'], 'system', 'system')
    ON CONFLICT DO NOTHING;
    
    -- Insert Objects with Different Statuses
    INSERT INTO objects (object_type_id, parent_object_id, name, description, metadata, status, tags, created_by, updated_by, deleted_at) VALUES
    ((SELECT id FROM object_types WHERE name = 'Product Category'), NULL, 'Archived Category', 'Category no longer in use', '{"archived_reason": "product_discontinued"}', 'archived', ARRAY['archived'], 'system', 'system', NULL),
    ((SELECT id FROM object_types WHERE name = 'Product Category'), NULL, 'Deleted Category', 'Deleted category', '{"deleted_reason": "duplicate"}', 'deleted', ARRAY['deleted'], 'system', 'system', NOW()),
    ((SELECT id FROM object_types WHERE name = 'Product Category'), NULL, 'Pending Category', 'Category pending approval', '{"pending_since": "2024-01-01"}', 'pending', ARRAY['pending'], 'system', 'system', NULL)
    ON CONFLICT DO NOTHING;
    
END $$;

-- Create Indexes for Better Query Performance on Test Data
CREATE INDEX IF NOT EXISTS idx_dev_objects_metadata_category ON objects USING GIN ((metadata->>'category') gin_trgm_ops) WHERE metadata ? 'category';
CREATE INDEX IF NOT EXISTS idx_dev_objects_name_trgm ON objects USING GIN (name gin_trgm_ops);

-- Add Comments for Documentation
COMMENT ON TABLE object_types IS 'Object type definitions with hierarchical relationships';
COMMENT ON TABLE objects IS 'Generic objects with flexible attributes and taxonomy support';
COMMENT ON COLUMN objects.public_id IS 'Public UUID for external API references';
COMMENT ON COLUMN objects.metadata IS 'Flexible JSONB attributes for custom data';
COMMENT ON COLUMN objects.tags IS 'Array of string tags for categorization';
```

---

### 7.2 Create Rollback Script

**File**: `migrations/development/000002_dev_tax_test_data.down.sql`

**Steps**:
1. Drop indexes created for test data
2. Remove all test data
3. Clean up in reverse order

```sql
-- Rollback Development Test Data

-- Drop indexes created for test data
DROP INDEX IF EXISTS idx_dev_objects_metadata_category;
DROP INDEX IF EXISTS idx_dev_objects_name_trgm;

-- Remove objects in correct order (children first, then parents)
-- Note: We delete all objects since this is test data
-- For production, you would be more selective

-- Remove objects
DELETE FROM objects WHERE created_by = 'system';

-- Remove object types
DELETE FROM object_types WHERE created_by = 'system';

-- Remove comments
COMMENT ON TABLE object_types IS NULL;
COMMENT ON TABLE objects IS NULL;
COMMENT ON COLUMN objects.public_id IS NULL;
COMMENT ON COLUMN objects.metadata IS NULL;
COMMENT ON COLUMN objects.tags IS NULL;
```

---

### 7.3 Update Environments File

**File**: `migrations/environments.json`

**Steps**:
1. Add development environment reference
2. Ensure test data migration is included in dev environment

```json
{
  "development": [
    "000001_initial",
    "000002_dev_tax_test_data"
  ],
  "production": [
    "000001_initial"
  ],
  "staging": [
    "000001_initial"
  ]
}
```

---

## Checklist

- [ ] Create `migrations/development/000002_dev_tax_test_data.up.sql`
- [ ] Create `migrations/development/000002_dev_tax_test_data.down.sql`
- [ ] Update `migrations/environments.json`
- [ ] Test migration up in development
- [ ] Verify test data in database
- [ ] Test migration down
- [ ] Verify data cleanup
- [ ] Update progress.md

## Testing

```bash
# Run development migrations
cd services/objects-service
go run cmd/migrate/main.go up --env=development

# Verify test data
psql postgresql://postgres:password@localhost:5432/objects_service -c "SELECT COUNT(*) FROM object_types;"
psql postgresql://postgres:password@localhost:5432/objects_service -c "SELECT COUNT(*) FROM objects;"
psql postgresql://postgres:password@localhost:5432/objects_service -c "SELECT name, status FROM objects ORDER BY created_at;"

# Test rollback
go run cmd/migrate/main.go down --env=development

# Verify cleanup
psql postgresql://postgres:password@localhost:5432/objects_service -c "SELECT COUNT(*) FROM object_types;"
psql postgresql://postgres:password@localhost:5432/objects_service -c "SELECT COUNT(*) FROM objects;"

# Re-run migrations
go run cmd/migrate/main.go up --env=development
```

## Manual Verification Queries

```sql
-- Check object types hierarchy
SELECT 
    ot1.id,
    ot1.name,
    ot2.name as parent_name,
    ot1.is_sealed
FROM object_types ot1
LEFT JOIN object_types ot2 ON ot1.parent_type_id = ot2.id
ORDER BY ot1.name;

-- Check objects with types
SELECT 
    o.id,
    o.public_id,
    o.name,
    ot.name as type_name,
    o.status,
    o.tags
FROM objects o
JOIN object_types ot ON o.object_type_id = ot.id
ORDER BY o.name;

-- Check hierarchical objects
SELECT 
    o1.name as parent_name,
    o2.name as child_name,
    o2.status
FROM objects o1
JOIN objects o2 ON o2.parent_object_id = o1.id
ORDER BY o1.name, o2.name;

-- Check metadata queries
SELECT 
    name,
    metadata->>'icon' as icon,
    metadata->>'priority' as priority
FROM objects
WHERE metadata ? 'icon';

-- Check tag queries
SELECT name, tags
FROM objects
WHERE tags @> ARRAY['electronics']::TEXT[];
```

## Common Issues

**Issue**: Migration fails due to foreign key constraints
**Solution**: Ensure migrations are run in correct order (parent types before child types)

**Issue**: Test data conflicts with existing data
**Solution**: Use ON CONFLICT DO NOTHING to handle duplicate inserts

**Issue**: JSONB metadata format errors
**Solution**: Ensure all JSON is valid and properly escaped in SQL

## Next Phase

Proceed to [Phase 8: Tests](phase-08-tests.md) once all tasks in this phase are complete.
