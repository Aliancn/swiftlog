-- Add ai_max_concurrent field to user_settings table
ALTER TABLE user_settings
ADD COLUMN ai_max_concurrent INTEGER NOT NULL DEFAULT 3;

COMMENT ON COLUMN user_settings.ai_max_concurrent IS 'Maximum number of concurrent AI analysis tasks (1-10)';

-- Add ai_max_concurrent field to project_settings table
ALTER TABLE project_settings
ADD COLUMN ai_max_concurrent INTEGER;

COMMENT ON COLUMN project_settings.ai_max_concurrent IS 'Project-specific maximum concurrent AI tasks (overrides user setting if set)';
