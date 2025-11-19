-- Migrate from global_settings to user-level settings

-- Drop old global_settings table and related triggers
DROP TRIGGER IF EXISTS trigger_update_global_settings_updated_at ON global_settings;
DROP FUNCTION IF EXISTS update_global_settings_updated_at();
DROP TABLE IF EXISTS global_settings CASCADE;

-- Create user_settings table (one row per user)
CREATE TABLE IF NOT EXISTS user_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,

    -- AI Configuration
    ai_enabled BOOLEAN NOT NULL DEFAULT true,
    ai_base_url TEXT NOT NULL DEFAULT 'https://api.openai.com/v1',
    ai_api_key TEXT, -- Encrypted or user-specific key
    ai_model TEXT NOT NULL DEFAULT 'gpt-4o-mini',
    ai_max_tokens INTEGER NOT NULL DEFAULT 500,
    ai_auto_analyze BOOLEAN NOT NULL DEFAULT false,
    ai_max_log_lines INTEGER NOT NULL DEFAULT 1000,
    ai_log_truncate_strategy TEXT NOT NULL DEFAULT 'tail', -- 'head', 'tail', 'smart'
    ai_system_prompt TEXT NOT NULL DEFAULT 'You are a helpful assistant analyzing script execution logs. Identify errors, warnings, and provide actionable recommendations.',

    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create index
CREATE INDEX idx_user_settings_user_id ON user_settings(user_id);

-- Updated at trigger for user_settings
CREATE OR REPLACE FUNCTION update_user_settings_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_user_settings_updated_at
    BEFORE UPDATE ON user_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_user_settings_updated_at();

-- Initialize default settings for all existing users
INSERT INTO user_settings (
    user_id,
    ai_enabled,
    ai_base_url,
    ai_model,
    ai_max_tokens,
    ai_auto_analyze,
    ai_max_log_lines,
    ai_log_truncate_strategy,
    ai_system_prompt
)
SELECT
    id,
    true,
    'https://api.openai.com/v1',
    'gpt-4o-mini',
    500,
    false,
    1000,
    'tail',
    'You are a helpful assistant analyzing script execution logs. Identify errors, warnings, and provide actionable recommendations.'
FROM users
ON CONFLICT (user_id) DO NOTHING;

-- Update project_settings to remove updated_by (no longer needed)
-- Project settings now inherit from user settings via project ownership
ALTER TABLE project_settings DROP COLUMN IF EXISTS updated_by;
