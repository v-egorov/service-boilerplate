-- Create objects_service schema
CREATE SCHEMA IF NOT EXISTS objects_service;

-- Enable pgcrypto extension for UUID generation
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Object Types Table
CREATE TABLE objects_service.object_types (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    parent_type_id BIGINT REFERENCES objects_service.object_types(id) ON DELETE RESTRICT,
    concrete_table_name VARCHAR(255) UNIQUE,
    description TEXT,
    is_sealed BOOLEAN DEFAULT FALSE,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT no_self_parent CHECK (parent_type_id IS NULL OR parent_type_id != id)
);

-- Objects Table
CREATE TABLE objects_service.objects (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    public_id UUID NOT NULL DEFAULT gen_random_uuid(),
    object_type_id BIGINT NOT NULL REFERENCES objects_service.object_types(id) ON DELETE RESTRICT,
    parent_object_id BIGINT REFERENCES objects_service.objects(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    version BIGINT DEFAULT 0,
    created_by VARCHAR(255) DEFAULT 'system',
    updated_by VARCHAR(255) DEFAULT 'system',
    metadata JSONB DEFAULT '{}'::jsonb,
    tags TEXT[] DEFAULT '{}',
    status VARCHAR(50) DEFAULT 'active' CHECK (
        status IN ('active', 'inactive', 'archived', 'deleted', 'pending')
    ),
    CONSTRAINT no_self_parent CHECK (parent_object_id IS NULL OR parent_object_id != id),
    CONSTRAINT objects_public_id_uniq UNIQUE (public_id)
);

-- Indexes for object_types
CREATE INDEX idx_object_types_parent ON objects_service.object_types(parent_type_id);
CREATE INDEX idx_object_types_name ON objects_service.object_types(name);
CREATE INDEX idx_object_types_sealed ON objects_service.object_types(is_sealed) WHERE is_sealed = TRUE;

-- Indexes for objects
CREATE INDEX idx_objects_type_id ON objects_service.objects(object_type_id);
CREATE INDEX idx_objects_type_status ON objects_service.objects(object_type_id, status);
CREATE INDEX idx_objects_status ON objects_service.objects(status);
CREATE INDEX idx_objects_parent_id ON objects_service.objects(parent_object_id);
CREATE INDEX idx_objects_public_id ON objects_service.objects(public_id);
CREATE INDEX idx_objects_metadata_gin ON objects_service.objects USING GIN (metadata);
CREATE INDEX idx_objects_tags_gin ON objects_service.objects USING GIN (tags);
CREATE INDEX idx_objects_created_at ON objects_service.objects(created_at DESC);
CREATE INDEX idx_objects_updated_at ON objects_service.objects(updated_at DESC);
CREATE INDEX idx_objects_deleted_at ON objects_service.objects(deleted_at) WHERE deleted_at IS NOT NULL;

-- Trigger function for updated_at columns
CREATE OR REPLACE FUNCTION objects_service.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers for updated_at
CREATE TRIGGER update_object_types_updated_at
    BEFORE UPDATE ON objects_service.object_types
    FOR EACH ROW
    EXECUTE FUNCTION objects_service.update_updated_at_column();

CREATE TRIGGER update_objects_updated_at
    BEFORE UPDATE ON objects_service.objects
    FOR EACH ROW
    EXECUTE FUNCTION objects_service.update_updated_at_column();