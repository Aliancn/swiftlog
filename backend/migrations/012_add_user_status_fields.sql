-- Migration: Add user status and management fields
-- Description: Adds fields for user account status management

-- Add is_active column
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE;

-- Add last_login column
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login TIMESTAMPTZ;

-- Comments
COMMENT ON COLUMN users.is_active IS 'Whether the user account is active (can be disabled by admin)';
COMMENT ON COLUMN users.last_login IS 'Timestamp of last successful login';

-- Create index for filtering active users
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);
