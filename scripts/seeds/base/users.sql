-- Base user seed data
-- This data is loaded in all environments
-- Contains essential users required for application functionality

INSERT INTO user_service.users (email, first_name, last_name) VALUES
    ('system.admin@example.com', 'System', 'Administrator'),
    ('noreply@example.com', 'No', 'Reply')
ON CONFLICT (email) DO NOTHING;