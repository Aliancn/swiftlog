-- Migration: Create users table
-- Description: Stores authenticated users of the platform

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT username_format CHECK (username ~ '^[a-zA-Z0-9_-]{3,50}$')
);

-- Index for username lookups
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

-- Comments
COMMENT ON TABLE users IS 'Authenticated users of the SwiftLog platform';
COMMENT ON COLUMN users.id IS 'Unique user identifier (UUID v4)';
COMMENT ON COLUMN users.username IS 'Unique username (3-50 chars, alphanumeric + _ -)';
COMMENT ON COLUMN users.created_at IS 'Account creation timestamp';
