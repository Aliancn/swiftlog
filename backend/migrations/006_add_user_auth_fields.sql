-- Migration: Add authentication fields to users table
-- Description: Adds password_hash and is_admin columns for user authentication

-- Add password_hash column
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255);

-- Add is_admin column
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_admin BOOLEAN NOT NULL DEFAULT FALSE;

-- Comments
COMMENT ON COLUMN users.password_hash IS 'Bcrypt hashed password';
COMMENT ON COLUMN users.is_admin IS 'Whether the user has admin privileges';
