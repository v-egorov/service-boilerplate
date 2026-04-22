-- Environment: all
-- Seed test objects and relationships for development
-- Creates comprehensive test data demonstrating all cardinality patterns

-- Step 1: Ensure Relationship marker exists (idempotent)
INSERT INTO objects_service.object_types (name, description, created_at, updated_at)
VALUES ('Relationship', 'Marker type for relationship instances', NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- Step 2: Create test objects for relationship sources/targets
-- Using subqueries to get object_type_id by name (not hard-coded IDs)

-- 2a. Test Portfolios (Category type - one_to_many parents)
INSERT INTO objects_service.objects (public_id, object_type_id, name, status, created_at, updated_at)
SELECT
    gen_random_uuid(),
    (SELECT id FROM objects_service.object_types WHERE name = 'Category'),
    'Test Portfolio A',
    'active',
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM objects_service.objects
    WHERE name = 'Test Portfolio A'
);

INSERT INTO objects_service.objects (public_id, object_type_id, name, status, created_at, updated_at)
SELECT
    gen_random_uuid(),
    (SELECT id FROM objects_service.object_types WHERE name = 'Category'),
    'Test Portfolio B',
    'active',
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM objects_service.objects
    WHERE name = 'Test Portfolio B'
);

-- 2b. Test Assets (Product type - one_to_many targets)
INSERT INTO objects_service.objects (public_id, object_type_id, name, status, created_at, updated_at)
SELECT
    gen_random_uuid(),
    (SELECT id FROM objects_service.object_types WHERE name = 'Product'),
    'Test Asset X',
    'active',
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM objects_service.objects
    WHERE name = 'Test Asset X'
);

INSERT INTO objects_service.objects (public_id, object_type_id, name, status, created_at, updated_at)
SELECT
    gen_random_uuid(),
    (SELECT id FROM objects_service.object_types WHERE name = 'Product'),
    'Test Asset Y',
    'active',
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM objects_service.objects
    WHERE name = 'Test Asset Y'
);

INSERT INTO objects_service.objects (public_id, object_type_id, name, status, created_at, updated_at)
SELECT
    gen_random_uuid(),
    (SELECT id FROM objects_service.object_types WHERE name = 'Product'),
    'Test Asset Z',
    'active',
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM objects_service.objects
    WHERE name = 'Test Asset Z'
);

-- 2c. Test Articles (Article type - many_to_many)
INSERT INTO objects_service.objects (public_id, object_type_id, name, status, created_at, updated_at)
SELECT
    gen_random_uuid(),
    (SELECT id FROM objects_service.object_types WHERE name = 'Article'),
    'Test Article Alpha',
    'active',
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM objects_service.objects
    WHERE name = 'Test Article Alpha'
);

INSERT INTO objects_service.objects (public_id, object_type_id, name, status, created_at, updated_at)
SELECT
    gen_random_uuid(),
    (SELECT id FROM objects_service.object_types WHERE name = 'Article'),
    'Test Article Beta',
    'active',
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM objects_service.objects
    WHERE name = 'Test Article Beta'
);

INSERT INTO objects_service.objects (public_id, object_type_id, name, status, created_at, updated_at)
SELECT
    gen_random_uuid(),
    (SELECT id FROM objects_service.object_types WHERE name = 'Article'),
    'Test Article Gamma',
    'active',
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM objects_service.objects
    WHERE name = 'Test Article Gamma'
);

-- 2d. Test Categories for parent_of (hierarchy)
INSERT INTO objects_service.objects (public_id, object_type_id, name, status, created_at, updated_at)
SELECT
    gen_random_uuid(),
    (SELECT id FROM objects_service.object_types WHERE name = 'Category'),
    'Test Category Parent',
    'active',
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM objects_service.objects
    WHERE name = 'Test Category Parent'
);

INSERT INTO objects_service.objects (public_id, object_type_id, name, status, created_at, updated_at)
SELECT
    gen_random_uuid(),
    (SELECT id FROM objects_service.object_types WHERE name = 'Category'),
    'Test Category Child',
    'active',
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM objects_service.objects
    WHERE name = 'Test Category Child'
);

-- Step 3: Create relationship marker objects (one per relationship type)
-- These are the CTI records in the objects table

INSERT INTO objects_service.objects (public_id, object_type_id, name, status, created_at, updated_at)
SELECT
    gen_random_uuid(),
    (SELECT id FROM objects_service.object_types WHERE name = 'Relationship'),
    'Relationship: Portfolio A contains Asset X',
    'active',
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM objects_service.objects o
    JOIN objects_service.object_types ot ON o.object_type_id = ot.id
    WHERE ot.name = 'Relationship'
    AND o.name = 'Relationship: Portfolio A contains Asset X'
);

INSERT INTO objects_service.objects (public_id, object_type_id, name, status, created_at, updated_at)
SELECT
    gen_random_uuid(),
    (SELECT id FROM objects_service.object_types WHERE name = 'Relationship'),
    'Relationship: Portfolio A contains Asset Y',
    'active',
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM objects_service.objects o
    JOIN objects_service.object_types ot ON o.object_type_id = ot.id
    WHERE ot.name = 'Relationship'
    AND o.name = 'Relationship: Portfolio A contains Asset Y'
);

INSERT INTO objects_service.objects (public_id, object_type_id, name, status, created_at, updated_at)
SELECT
    gen_random_uuid(),
    (SELECT id FROM objects_service.object_types WHERE name = 'Relationship'),
    'Relationship: Asset Z belongs to Portfolio A',
    'active',
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM objects_service.objects o
    JOIN objects_service.object_types ot ON o.object_type_id = ot.id
    WHERE ot.name = 'Relationship'
    AND o.name = 'Relationship: Asset Z belongs to Portfolio A'
);

INSERT INTO objects_service.objects (public_id, object_type_id, name, status, created_at, updated_at)
SELECT
    gen_random_uuid(),
    (SELECT id FROM objects_service.object_types WHERE name = 'Relationship'),
    'Relationship: Category Parent parent_of Category Child',
    'active',
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM objects_service.objects o
    JOIN objects_service.object_types ot ON o.object_type_id = ot.id
    WHERE ot.name = 'Relationship'
    AND o.name = 'Relationship: Category Parent parent_of Category Child'
);

INSERT INTO objects_service.objects (public_id, object_type_id, name, status, created_at, updated_at)
SELECT
    gen_random_uuid(),
    (SELECT id FROM objects_service.object_types WHERE name = 'Relationship'),
    'Relationship: Article Alpha references Article Beta',
    'active',
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM objects_service.objects o
    JOIN objects_service.object_types ot ON o.object_type_id = ot.id
    WHERE ot.name = 'Relationship'
    AND o.name = 'Relationship: Article Alpha references Article Beta'
);

INSERT INTO objects_service.objects (public_id, object_type_id, name, status, created_at, updated_at)
SELECT
    gen_random_uuid(),
    (SELECT id FROM objects_service.object_types WHERE name = 'Relationship'),
    'Relationship: Article Alpha references Article Gamma',
    'active',
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM objects_service.objects o
    JOIN objects_service.object_types ot ON o.object_type_id = ot.id
    WHERE ot.name = 'Relationship'
    AND o.name = 'Relationship: Article Alpha references Article Gamma'
);

INSERT INTO objects_service.objects (public_id, object_type_id, name, status, created_at, updated_at)
SELECT
    gen_random_uuid(),
    (SELECT id FROM objects_service.object_types WHERE name = 'Relationship'),
    'Relationship: Asset X depends_on Asset Y',
    'active',
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM objects_service.objects o
    JOIN objects_service.object_types ot ON o.object_type_id = ot.id
    WHERE ot.name = 'Relationship'
    AND o.name = 'Relationship: Asset X depends_on Asset Y'
);

-- Step 4: Create relationship instances in objects_relationships table
-- Using separate INSERT statements for reliable subquery resolution

-- 1. one_to_many: Portfolio A contains Asset X
INSERT INTO objects_service.objects_relationships (object_id, source_object_id, target_object_id, relationship_type_id, status, relationship_metadata, created_at, updated_at)
SELECT r.id, src.id, tgt.id, rt.object_id, 'active', '{}', NOW(), NOW()
FROM objects_service.objects r
CROSS JOIN (SELECT id FROM objects_service.objects WHERE name = 'Test Portfolio A') src(id)
CROSS JOIN (SELECT id FROM objects_service.objects WHERE name = 'Test Asset X') tgt(id)
CROSS JOIN (SELECT object_id FROM objects_service.objects_relationship_types WHERE type_key = 'contains') rt(object_id)
WHERE r.name = 'Relationship: Portfolio A contains Asset X'
AND NOT EXISTS (SELECT 1 FROM objects_service.objects_relationships WHERE object_id = r.id);

-- 2. one_to_many: Portfolio A contains Asset Y
INSERT INTO objects_service.objects_relationships (object_id, source_object_id, target_object_id, relationship_type_id, status, relationship_metadata, created_at, updated_at)
SELECT r.id, src.id, tgt.id, rt.object_id, 'active', '{}', NOW(), NOW()
FROM objects_service.objects r
CROSS JOIN (SELECT id FROM objects_service.objects WHERE name = 'Test Portfolio A') src(id)
CROSS JOIN (SELECT id FROM objects_service.objects WHERE name = 'Test Asset Y') tgt(id)
CROSS JOIN (SELECT object_id FROM objects_service.objects_relationship_types WHERE type_key = 'contains') rt(object_id)
WHERE r.name = 'Relationship: Portfolio A contains Asset Y'
AND NOT EXISTS (SELECT 1 FROM objects_service.objects_relationships WHERE object_id = r.id);

-- 3. many_to_one: Asset Z belongs to Portfolio A
INSERT INTO objects_service.objects_relationships (object_id, source_object_id, target_object_id, relationship_type_id, status, relationship_metadata, created_at, updated_at)
SELECT r.id, src.id, tgt.id, rt.object_id, 'active', '{}', NOW(), NOW()
FROM objects_service.objects r
CROSS JOIN (SELECT id FROM objects_service.objects WHERE name = 'Test Asset Z') src(id)
CROSS JOIN (SELECT id FROM objects_service.objects WHERE name = 'Test Portfolio A') tgt(id)
CROSS JOIN (SELECT object_id FROM objects_service.objects_relationship_types WHERE type_key = 'belongs_to') rt(object_id)
WHERE r.name = 'Relationship: Asset Z belongs to Portfolio A'
AND NOT EXISTS (SELECT 1 FROM objects_service.objects_relationships WHERE object_id = r.id);

-- 4. one_to_many hierarchy: Category Parent parent_of Category Child
INSERT INTO objects_service.objects_relationships (object_id, source_object_id, target_object_id, relationship_type_id, status, relationship_metadata, created_at, updated_at)
SELECT r.id, src.id, tgt.id, rt.object_id, 'active', '{}', NOW(), NOW()
FROM objects_service.objects r
CROSS JOIN (SELECT id FROM objects_service.objects WHERE name = 'Test Category Parent') src(id)
CROSS JOIN (SELECT id FROM objects_service.objects WHERE name = 'Test Category Child') tgt(id)
CROSS JOIN (SELECT object_id FROM objects_service.objects_relationship_types WHERE type_key = 'parent_of') rt(object_id)
WHERE r.name = 'Relationship: Category Parent parent_of Category Child'
AND NOT EXISTS (SELECT 1 FROM objects_service.objects_relationships WHERE object_id = r.id);

-- 5. many_to_many: Article Alpha references Article Beta
INSERT INTO objects_service.objects_relationships (object_id, source_object_id, target_object_id, relationship_type_id, status, relationship_metadata, created_at, updated_at)
SELECT r.id, src.id, tgt.id, rt.object_id, 'active', '{}', NOW(), NOW()
FROM objects_service.objects r
CROSS JOIN (SELECT id FROM objects_service.objects WHERE name = 'Test Article Alpha') src(id)
CROSS JOIN (SELECT id FROM objects_service.objects WHERE name = 'Test Article Beta') tgt(id)
CROSS JOIN (SELECT object_id FROM objects_service.objects_relationship_types WHERE type_key = 'references') rt(object_id)
WHERE r.name = 'Relationship: Article Alpha references Article Beta'
AND NOT EXISTS (SELECT 1 FROM objects_service.objects_relationships WHERE object_id = r.id);

-- 6. many_to_many: Article Alpha references Article Gamma
INSERT INTO objects_service.objects_relationships (object_id, source_object_id, target_object_id, relationship_type_id, status, relationship_metadata, created_at, updated_at)
SELECT r.id, src.id, tgt.id, rt.object_id, 'active', '{}', NOW(), NOW()
FROM objects_service.objects r
CROSS JOIN (SELECT id FROM objects_service.objects WHERE name = 'Test Article Alpha') src(id)
CROSS JOIN (SELECT id FROM objects_service.objects WHERE name = 'Test Article Gamma') tgt(id)
CROSS JOIN (SELECT object_id FROM objects_service.objects_relationship_types WHERE type_key = 'references') rt(object_id)
WHERE r.name = 'Relationship: Article Alpha references Article Gamma'
AND NOT EXISTS (SELECT 1 FROM objects_service.objects_relationships WHERE object_id = r.id);

-- 7. many_to_many: Asset X depends_on Asset Y
INSERT INTO objects_service.objects_relationships (object_id, source_object_id, target_object_id, relationship_type_id, status, relationship_metadata, created_at, updated_at)
SELECT r.id, src.id, tgt.id, rt.object_id, 'active', '{}', NOW(), NOW()
FROM objects_service.objects r
CROSS JOIN (SELECT id FROM objects_service.objects WHERE name = 'Test Asset X') src(id)
CROSS JOIN (SELECT id FROM objects_service.objects WHERE name = 'Test Asset Y') tgt(id)
CROSS JOIN (SELECT object_id FROM objects_service.objects_relationship_types WHERE type_key = 'depends_on') rt(object_id)
WHERE r.name = 'Relationship: Asset X depends_on Asset Y'
AND NOT EXISTS (SELECT 1 FROM objects_service.objects_relationships WHERE object_id = r.id);