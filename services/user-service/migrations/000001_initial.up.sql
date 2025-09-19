-- Create user_service schema
CREATE SCHEMA IF NOT EXISTS user_service;

-- Create users table in user_service schema
CREATE TABLE IF NOT EXISTS user_service.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index on email for faster lookups
CREATE INDEX IF NOT EXISTS idx_users_email ON user_service.users(email);

-- Create index on created_at for ordering
CREATE INDEX IF NOT EXISTS idx_users_created_at ON user_service.users(created_at DESC);