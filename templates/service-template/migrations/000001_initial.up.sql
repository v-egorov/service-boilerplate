-- Create SCHEMA_NAME schema
CREATE SCHEMA IF NOT EXISTS SCHEMA_NAME;

-- Create migration executions tracking table for migration orchestrator
CREATE TABLE IF NOT EXISTS SCHEMA_NAME.migration_executions (
    id BIGSERIAL PRIMARY KEY,
    migration_id VARCHAR(255) NOT NULL,
    migration_version VARCHAR(255) NOT NULL,
    environment VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL CHECK (status IN ('pending', 'running', 'completed', 'failed', 'rolled_back')),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    duration_ms BIGINT,
    executed_by VARCHAR(255),
    checksum VARCHAR(255),
    dependencies JSONB,
    metadata JSONB,
    error_message TEXT,
    rollback_version VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(migration_id, environment)
);

-- Create indexes for migration executions table
CREATE INDEX IF NOT EXISTS idx_migration_executions_migration_id ON SCHEMA_NAME.migration_executions(migration_id);
CREATE INDEX IF NOT EXISTS idx_migration_executions_environment ON SCHEMA_NAME.migration_executions(environment);
CREATE INDEX IF NOT EXISTS idx_migration_executions_status ON SCHEMA_NAME.migration_executions(status);
CREATE INDEX IF NOT EXISTS idx_migration_executions_created_at ON SCHEMA_NAME.migration_executions(created_at);

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