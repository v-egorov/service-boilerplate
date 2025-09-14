-- Development test data for SERVICE_NAME service
-- This file is used to populate the database with test data during development

-- Note: This migration should only be applied in development environments
-- The down migration will remove all test data

-- Insert test entities
INSERT INTO SCHEMA_NAME.entities (name, description, created_at, updated_at) VALUES
('Test Entity 1', 'This is a test entity for development purposes', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('Test Entity 2', 'Another test entity with different description', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('Sample Item', 'A sample entity to demonstrate functionality', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('Demo Entity', 'Entity used for demonstration and testing', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);