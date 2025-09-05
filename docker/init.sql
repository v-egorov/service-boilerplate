-- Database initialization script
-- This script runs when the PostgreSQL container starts

-- Create the service database if it doesn't exist
-- (Note: This is handled by POSTGRES_DB environment variable)

-- You can add additional database setup here
-- For example: creating extensions, setting up roles, etc.

-- Example: Enable UUID extension if needed
-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Example: Create additional roles if needed
-- CREATE ROLE readonly_user WITH LOGIN PASSWORD 'readonly_password';
-- GRANT CONNECT ON DATABASE service_db TO readonly_user;