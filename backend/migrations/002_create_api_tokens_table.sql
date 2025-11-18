-- Migration: Create api_tokens table
-- Description: Stores API tokens for CLI and client authentication

CREATE TABLE IF NOT EXISTS api_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT token_hash_format CHECK (token_hash ~ '^[a-f0-9]{64}$')
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_api_tokens_user_id ON api_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_api_tokens_token_hash ON api_tokens(token_hash);

-- Comments
COMMENT ON TABLE api_tokens IS 'API tokens for authenticating CLI and other clients';
COMMENT ON COLUMN api_tokens.id IS 'Unique token identifier';
COMMENT ON COLUMN api_tokens.user_id IS 'Foreign key to users table';
COMMENT ON COLUMN api_tokens.token_hash IS 'SHA-256 hash of raw token (64 hex chars)';
COMMENT ON COLUMN api_tokens.name IS 'User-provided name for the token (e.g., "My Laptop")';
COMMENT ON COLUMN api_tokens.created_at IS 'Token creation timestamp';
