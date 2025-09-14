-- Create SCHEMA_NAME schema
CREATE SCHEMA IF NOT EXISTS SCHEMA_NAME;

-- Create entities table in SCHEMA_NAME schema
CREATE TABLE IF NOT EXISTS SCHEMA_NAME.entities (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_entities_name ON SCHEMA_NAME.entities(name);
CREATE INDEX IF NOT EXISTS idx_entities_created_at ON SCHEMA_NAME.entities(created_at);

-- Create updated_at trigger
CREATE OR REPLACE FUNCTION SCHEMA_NAME.update_entities_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_entities_updated_at
    BEFORE UPDATE ON SCHEMA_NAME.entities
    FOR EACH ROW EXECUTE FUNCTION SCHEMA_NAME.update_entities_updated_at_column();