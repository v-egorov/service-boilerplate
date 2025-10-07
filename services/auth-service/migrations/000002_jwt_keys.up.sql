-- JWT keys table for persistent RSA key storage
CREATE TABLE IF NOT EXISTS auth_service.jwt_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_id VARCHAR(100) UNIQUE NOT NULL, -- Identifier for the key
    private_key_pem TEXT NOT NULL, -- PEM-encoded private key
    public_key_pem TEXT NOT NULL,  -- PEM-encoded public key
    algorithm VARCHAR(50) NOT NULL DEFAULT 'RS256', -- Signing algorithm
    is_active BOOLEAN NOT NULL DEFAULT true, -- Only one active key at a time
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE, -- Optional expiration for key rotation
    metadata JSONB -- Additional metadata (key size, etc.)
);

-- Ensure only one active key
CREATE UNIQUE INDEX IF NOT EXISTS idx_jwt_keys_active
ON auth_service.jwt_keys(is_active)
WHERE is_active = true;

-- Index for key lookup
CREATE INDEX IF NOT EXISTS idx_jwt_keys_key_id ON auth_service.jwt_keys(key_id);
CREATE INDEX IF NOT EXISTS idx_jwt_keys_active_created ON auth_service.jwt_keys(is_active, created_at DESC);