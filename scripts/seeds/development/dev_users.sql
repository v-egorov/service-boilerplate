-- Development seed data
-- This data is only loaded in development environment
-- Contains test users and sample data for development

INSERT INTO user_service.users (email, first_name, last_name) VALUES
    ('alice.developer@example.com', 'Alice', 'Developer'),
    ('bob.manager@example.com', 'Bob', 'Manager'),
    ('charlie.tester@example.com', 'Charlie', 'Tester'),
    ('diana.designer@example.com', 'Diana', 'Designer'),
    ('eve.analyst@example.com', 'Eve', 'Analyst')
ON CONFLICT (email) DO NOTHING;

-- Add some additional development data
-- You can add more tables here as your schema grows