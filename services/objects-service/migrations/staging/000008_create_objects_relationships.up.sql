-- Environment: all
-- Create objects_relationships CTI table for relationship instances
CREATE TABLE objects_service.objects_relationships (
    object_id BIGINT PRIMARY KEY REFERENCES objects_service.objects(id) ON DELETE CASCADE,
    source_object_id BIGINT NOT NULL REFERENCES objects_service.objects(id) ON DELETE CASCADE,
    target_object_id BIGINT NOT NULL REFERENCES objects_service.objects(id) ON DELETE CASCADE,
    relationship_type_id BIGINT NOT NULL REFERENCES objects_service.objects(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'active',
    relationship_metadata JSONB DEFAULT '{}',
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_relationship UNIQUE (source_object_id, target_object_id, relationship_type_id)
);

-- Indexes for common queries
CREATE INDEX idx_relationships_source ON objects_service.objects_relationships(source_object_id);
CREATE INDEX idx_relationships_target ON objects_service.objects_relationships(target_object_id);
CREATE INDEX idx_relationships_type_status ON objects_service.objects_relationships(relationship_type_id, status);
CREATE INDEX idx_relationships_type ON objects_service.objects_relationships(relationship_type_id);
CREATE INDEX idx_relationships_source_target ON objects_service.objects_relationships(source_object_id, target_object_id);

COMMENT ON TABLE objects_service.objects_relationships IS 'CTI concrete table for relationship instances';