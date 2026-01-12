-- Create objects_service schema
CREATE SCHEMA IF NOT EXISTS objects_service;

-- Create entities table in objects_service schema
CREATE TABLE IF NOT EXISTS objects_service.entities (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_entities_name ON objects_service.entities(name);
CREATE INDEX IF NOT EXISTS idx_entities_created_at ON objects_service.entities(created_at);

-- Create updated_at trigger
CREATE OR REPLACE FUNCTION objects_service.update_entities_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_entities_updated_at
    BEFORE UPDATE ON objects_service.entities
    FOR EACH ROW EXECUTE FUNCTION objects_service.update_entities_updated_at_column();