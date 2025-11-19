-- Settings tables for global and project-level configuration

-- Global settings table (single row)
CREATE TABLE IF NOT EXISTS global_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- AI Configuration
    ai_enabled BOOLEAN NOT NULL DEFAULT true,
    ai_base_url TEXT NOT NULL DEFAULT 'https://api.openai.com/v1',
    ai_api_key TEXT, -- Encrypted or hashed
    ai_model TEXT NOT NULL DEFAULT 'gpt-4o-mini',
    ai_max_tokens INTEGER NOT NULL DEFAULT 500,
    ai_auto_analyze BOOLEAN NOT NULL DEFAULT false,
    ai_max_log_lines INTEGER NOT NULL DEFAULT 1000,
    ai_log_truncate_strategy TEXT NOT NULL DEFAULT 'tail', -- 'head', 'tail', 'smart'
    ai_system_prompt TEXT NOT NULL DEFAULT 'You are a helpful assistant analyzing script execution logs. Identify errors, warnings, and provide actionable recommendations.',

    -- Metadata
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Ensure only one row exists
    CONSTRAINT single_row_constraint CHECK (id = '00000000-0000-0000-0000-000000000001'::uuid)
);

-- Project-specific settings table
CREATE TABLE IF NOT EXISTS project_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL UNIQUE REFERENCES projects(id) ON DELETE CASCADE,

    -- AI Configuration (nullable = inherit from global)
    ai_enabled BOOLEAN,
    ai_base_url TEXT,
    ai_api_key TEXT,
    ai_model TEXT,
    ai_max_tokens INTEGER,
    ai_auto_analyze BOOLEAN,
    ai_max_log_lines INTEGER,
    ai_log_truncate_strategy TEXT,
    ai_system_prompt TEXT,

    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL
);

-- Insert default global settings
INSERT INTO global_settings (id, ai_enabled, ai_base_url, ai_model, ai_max_tokens, ai_auto_analyze, ai_max_log_lines, ai_log_truncate_strategy, ai_system_prompt)
VALUES (
    '00000000-0000-0000-0000-000000000001'::uuid,
    true,
    'https://api.openai.com/v1',
    'gpt-4o-mini',
    500,
    false,
    1000,
    'tail',
    'You are a helpful assistant analyzing script execution logs. Identify errors, warnings, and provide actionable recommendations.'
) ON CONFLICT (id) DO NOTHING;

-- Indexes
CREATE INDEX idx_project_settings_project_id ON project_settings(project_id);

-- Updated at trigger for global_settings
CREATE OR REPLACE FUNCTION update_global_settings_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_global_settings_updated_at
    BEFORE UPDATE ON global_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_global_settings_updated_at();

-- Updated at trigger for project_settings
CREATE OR REPLACE FUNCTION update_project_settings_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_project_settings_updated_at
    BEFORE UPDATE ON project_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_project_settings_updated_at();
