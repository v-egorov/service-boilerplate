-- Create user_service schema
CREATE SCHEMA IF NOT EXISTS user_service;

-- Create migration executions tracking table for migration orchestrator
CREATE TABLE IF NOT EXISTS user_service.migration_executions (
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
CREATE INDEX IF NOT EXISTS idx_migration_executions_migration_id ON user_service.migration_executions(migration_id);
CREATE INDEX IF NOT EXISTS idx_migration_executions_environment ON user_service.migration_executions(environment);
CREATE INDEX IF NOT EXISTS idx_migration_executions_status ON user_service.migration_executions(status);
CREATE INDEX IF NOT EXISTS idx_migration_executions_created_at ON user_service.migration_executions(created_at);

-- Create users table in user_service schema
CREATE TABLE IF NOT EXISTS user_service.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index on email for faster lookups
CREATE INDEX IF NOT EXISTS idx_users_email ON user_service.users(email);

-- Create index on created_at for ordering
CREATE INDEX IF NOT EXISTS idx_users_created_at ON user_service.users(created_at DESC);

-- Add comment for documentation
COMMENT ON COLUMN user_service.users.password_hash IS 'Bcrypt hashed password for user authentication';

-- Create index on password_hash (though it's not typically queried directly)
-- This is mainly for consistency and potential future use
CREATE INDEX IF NOT EXISTS idx_users_password_hash ON user_service.users(password_hash);
