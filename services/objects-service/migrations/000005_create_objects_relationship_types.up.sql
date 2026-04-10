-- Environment: all
CREATE TABLE objects_service.objects_relationship_types (
    object_id BIGINT PRIMARY KEY REFERENCES objects_service.objects(id) ON DELETE CASCADE,
    type_key VARCHAR(100) NOT NULL UNIQUE,
    relationship_name VARCHAR(255),
    reverse_type_key VARCHAR(100) NULL,
    cardinality VARCHAR(20) NOT NULL DEFAULT 'many_to_many',
    required BOOLEAN DEFAULT FALSE,
    min_count INTEGER DEFAULT 0,
    max_count INTEGER DEFAULT -1,
    validation_rules JSONB DEFAULT '{}',
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_relationship_types_reverse_key 
    ON objects_service.objects_relationship_types(reverse_type_key);

CREATE INDEX idx_relationship_types_cardinality 
    ON objects_service.objects_relationship_types(cardinality);

COMMENT ON TABLE objects_service.objects_relationship_types IS 
    'CTI concrete table for relationship type instances';
