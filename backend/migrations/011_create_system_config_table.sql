-- Migration: Create system_config table
-- Description: Stores system-wide configuration settings (admin-managed)

CREATE TABLE IF NOT EXISTS system_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key VARCHAR(100) NOT NULL UNIQUE,
    value TEXT NOT NULL,
    description TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by UUID REFERENCES users(id),
    CONSTRAINT key_format CHECK (key ~ '^[a-z_]+$')
);

-- Index for key lookups
CREATE INDEX IF NOT EXISTS idx_system_config_key ON system_config(key);

-- Insert default configuration
INSERT INTO system_config (key, value, description) VALUES
    ('allow_registration', 'true', 'Allow new users to register accounts'),
    ('require_admin_approval', 'false', 'Require admin approval for new registrations'),
    ('default_user_quota_projects', '10', 'Default maximum number of projects per user'),
    ('default_user_quota_storage_mb', '1000', 'Default storage quota in MB per user')
ON CONFLICT (key) DO NOTHING;

-- Comments
COMMENT ON TABLE system_config IS 'System-wide configuration settings managed by administrators';
COMMENT ON COLUMN system_config.key IS 'Unique configuration key (lowercase with underscores)';
COMMENT ON COLUMN system_config.value IS 'Configuration value (stored as text, parse as needed)';
COMMENT ON COLUMN system_config.description IS 'Human-readable description of this setting';
COMMENT ON COLUMN system_config.updated_at IS 'Last update timestamp';
COMMENT ON COLUMN system_config.updated_by IS 'User who last updated this setting';
