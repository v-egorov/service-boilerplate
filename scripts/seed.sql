INSERT INTO user_service.users (email, first_name, last_name) VALUES
    ('john.doe@example.com', 'John', 'Doe'),
    ('jane.smith@example.com', 'Jane', 'Smith'),
    ('bob.wilson@example.com', 'Bob', 'Wilson'),
    ('alice.johnson@example.com', 'Alice', 'Johnson'),
    ('charlie.brown@example.com', 'Charlie', 'Brown')
ON CONFLICT (email) DO NOTHING;