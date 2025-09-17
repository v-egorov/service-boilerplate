-- Create auth_service schema
CREATE SCHEMA IF NOT EXISTS auth_service;

-- Authentication tokens table
CREATE TABLE IF NOT EXISTS auth_service.auth_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    token_type VARCHAR(50) NOT NULL, -- 'access', 'refresh', 'api_key'
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    revoked_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- User sessions table
CREATE TABLE IF NOT EXISTS auth_service.user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    session_token VARCHAR(255) UNIQUE NOT NULL,
    ip_address INET,
    user_agent TEXT,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Roles table
CREATE TABLE IF NOT EXISTS auth_service.roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Permissions table
CREATE TABLE IF NOT EXISTS auth_service.permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL, -- 'create', 'read', 'update', 'delete'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Role permissions junction table
CREATE TABLE IF NOT EXISTS auth_service.role_permissions (
    role_id UUID REFERENCES auth_service.roles(id) ON DELETE CASCADE,
    permission_id UUID REFERENCES auth_service.permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- User roles junction table
CREATE TABLE IF NOT EXISTS auth_service.user_roles (
    user_id UUID NOT NULL, -- References user_service.users(id)
    role_id UUID REFERENCES auth_service.roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, role_id)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_auth_tokens_user_id ON auth_service.auth_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_auth_tokens_token_hash ON auth_service.auth_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_auth_tokens_expires_at ON auth_service.auth_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON auth_service.user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON auth_service.user_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_roles_name ON auth_service.roles(name);
CREATE INDEX IF NOT EXISTS idx_permissions_name ON auth_service.permissions(name);
CREATE INDEX IF NOT EXISTS idx_permissions_resource_action ON auth_service.permissions(resource, action);

-- Create updated_at trigger for auth_tokens
CREATE OR REPLACE FUNCTION auth_service.update_auth_tokens_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_auth_tokens_updated_at
    BEFORE UPDATE ON auth_service.auth_tokens
    FOR EACH ROW EXECUTE FUNCTION auth_service.update_auth_tokens_updated_at_column();

-- Insert default roles
INSERT INTO auth_service.roles (name, description) VALUES
    ('admin', 'Administrator with full access'),
    ('user', 'Regular user with basic access')
ON CONFLICT (name) DO NOTHING;

-- Insert default permissions
INSERT INTO auth_service.permissions (name, resource, action) VALUES
    ('read:profile', 'profile', 'read'),
    ('update:profile', 'profile', 'update'),
    ('read:users', 'users', 'read'),
    ('create:users', 'users', 'create'),
    ('update:users', 'users', 'update'),
    ('delete:users', 'users', 'delete'),
    ('manage:roles', 'roles', 'manage'),
    ('manage:permissions', 'permissions', 'manage')
ON CONFLICT (name) DO NOTHING;

-- Assign permissions to admin role
INSERT INTO auth_service.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth_service.roles r
CROSS JOIN auth_service.permissions p
WHERE r.name = 'admin'
ON CONFLICT DO NOTHING;

-- Assign basic permissions to user role
INSERT INTO auth_service.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth_service.roles r
JOIN auth_service.permissions p ON p.name IN ('read:profile', 'update:profile')
WHERE r.name = 'user'
ON CONFLICT DO NOTHING;