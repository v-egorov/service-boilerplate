-- Seed standard relationship types for development
-- This migration creates base objects for relationship types and then the relationship type entries

-- Step 1: Ensure the RelationshipType marker exists in object_types
INSERT INTO objects_service.object_types (name, description, created_at, updated_at)
VALUES ('RelationshipType', 'Marker type for relationship type instances', NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- Step 2: Create base objects for relationship types (if not exist)
-- Uses subquery to get the type_id each time

-- 1. contains / contained_by (one_to_many)
INSERT INTO objects_service.objects (public_id, object_type_id, name, status, created_at, updated_at)
SELECT 
    gen_random_uuid(),
    (SELECT id FROM objects_service.object_types WHERE name = 'RelationshipType'),
    'Relationship: contains',
    'active',
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM objects_service.objects o
    JOIN objects_service.object_types ot ON o.object_type_id = ot.id
    WHERE ot.name = 'RelationshipType'
    AND o.name = 'Relationship: contains'
);

-- 2. belongs_to / owns (many_to_one)
INSERT INTO objects_service.objects (public_id, object_type_id, name, status, created_at, updated_at)
SELECT 
    gen_random_uuid(),
    (SELECT id FROM objects_service.object_types WHERE name = 'RelationshipType'),
    'Relationship: belongs_to',
    'active',
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM objects_service.objects o
    JOIN objects_service.object_types ot ON o.object_type_id = ot.id
    WHERE ot.name = 'RelationshipType'
    AND o.name = 'Relationship: belongs_to'
);

-- 3. references (many_to_many, unidirectional)
INSERT INTO objects_service.objects (public_id, object_type_id, name, status, created_at, updated_at)
SELECT 
    gen_random_uuid(),
    (SELECT id FROM objects_service.object_types WHERE name = 'RelationshipType'),
    'Relationship: references',
    'active',
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM objects_service.objects o
    JOIN objects_service.object_types ot ON o.object_type_id = ot.id
    WHERE ot.name = 'RelationshipType'
    AND o.name = 'Relationship: references'
);

-- 4. parent_of / child_of (one_to_many)
INSERT INTO objects_service.objects (public_id, object_type_id, name, status, created_at, updated_at)
SELECT 
    gen_random_uuid(),
    (SELECT id FROM objects_service.object_types WHERE name = 'RelationshipType'),
    'Relationship: parent_of',
    'active',
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM objects_service.objects o
    JOIN objects_service.object_types ot ON o.object_type_id = ot.id
    WHERE ot.name = 'RelationshipType'
    AND o.name = 'Relationship: parent_of'
);

-- 5. depends_on (many_to_many, unidirectional)
INSERT INTO objects_service.objects (public_id, object_type_id, name, status, created_at, updated_at)
SELECT 
    gen_random_uuid(),
    (SELECT id FROM objects_service.object_types WHERE name = 'RelationshipType'),
    'Relationship: depends_on',
    'active',
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM objects_service.objects o
    JOIN objects_service.object_types ot ON o.object_type_id = ot.id
    WHERE ot.name = 'RelationshipType'
    AND o.name = 'Relationship: depends_on'
);

-- Step 3: Create relationship type entries

-- 1. contains / contained_by (one_to_many, bidirectional)
INSERT INTO objects_service.objects_relationship_types (
    object_id, type_key, relationship_name, reverse_type_key, 
    cardinality, required, min_count, max_count, validation_rules,
    created_at, updated_at
)
SELECT 
    o.id,
    'contains',
    'contains',
    'contained_by',
    'one_to_many',
    false,
    0,
    -1,
    '{}',
    NOW(),
    NOW()
FROM objects_service.objects o
JOIN objects_service.object_types ot ON o.object_type_id = ot.id
WHERE ot.name = 'RelationshipType' AND o.name = 'Relationship: contains'
ON CONFLICT (type_key) DO NOTHING;

-- 2. belongs_to / owns (many_to_one, bidirectional)
INSERT INTO objects_service.objects_relationship_types (
    object_id, type_key, relationship_name, reverse_type_key, 
    cardinality, required, min_count, max_count, validation_rules,
    created_at, updated_at
)
SELECT 
    o.id,
    'belongs_to',
    'belongs to',
    'owns',
    'many_to_one',
    false,
    0,
    -1,
    '{}',
    NOW(),
    NOW()
FROM objects_service.objects o
JOIN objects_service.object_types ot ON o.object_type_id = ot.id
WHERE ot.name = 'RelationshipType' AND o.name = 'Relationship: belongs_to'
ON CONFLICT (type_key) DO NOTHING;

-- 3. references (many_to_many, unidirectional)
INSERT INTO objects_service.objects_relationship_types (
    object_id, type_key, relationship_name, reverse_type_key, 
    cardinality, required, min_count, max_count, validation_rules,
    created_at, updated_at
)
SELECT 
    o.id,
    'references',
    'references',
    NULL,
    'many_to_many',
    false,
    0,
    -1,
    '{}',
    NOW(),
    NOW()
FROM objects_service.objects o
JOIN objects_service.object_types ot ON o.object_type_id = ot.id
WHERE ot.name = 'RelationshipType' AND o.name = 'Relationship: references'
ON CONFLICT (type_key) DO NOTHING;

-- 4. parent_of / child_of (one_to_many, bidirectional)
INSERT INTO objects_service.objects_relationship_types (
    object_id, type_key, relationship_name, reverse_type_key, 
    cardinality, required, min_count, max_count, validation_rules,
    created_at, updated_at
)
SELECT 
    o.id,
    'parent_of',
    'parent of',
    'child_of',
    'one_to_many',
    false,
    0,
    -1,
    '{}',
    NOW(),
    NOW()
FROM objects_service.objects o
JOIN objects_service.object_types ot ON o.object_type_id = ot.id
WHERE ot.name = 'RelationshipType' AND o.name = 'Relationship: parent_of'
ON CONFLICT (type_key) DO NOTHING;

-- 5. depends_on (many_to_many, unidirectional)
INSERT INTO objects_service.objects_relationship_types (
    object_id, type_key, relationship_name, reverse_type_key, 
    cardinality, required, min_count, max_count, validation_rules,
    created_at, updated_at
)
SELECT 
    o.id,
    'depends_on',
    'depends on',
    NULL,
    'many_to_many',
    false,
    0,
    -1,
    '{}',
    NOW(),
    NOW()
FROM objects_service.objects o
JOIN objects_service.object_types ot ON o.object_type_id = ot.id
WHERE ot.name = 'RelationshipType' AND o.name = 'Relationship: depends_on'
ON CONFLICT (type_key) DO NOTHING;
