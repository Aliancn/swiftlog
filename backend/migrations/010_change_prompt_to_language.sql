-- Migration: Change ai_system_prompt to ai_prompt_language
-- This migration changes the prompt configuration from free-text to language selection

-- Update user_settings table
ALTER TABLE user_settings
    DROP COLUMN IF EXISTS ai_system_prompt,
    ADD COLUMN ai_prompt_language VARCHAR(5) NOT NULL DEFAULT 'en';

-- Update project_settings table
ALTER TABLE project_settings
    DROP COLUMN IF EXISTS ai_system_prompt,
    ADD COLUMN ai_prompt_language VARCHAR(5);

-- Add check constraint to ensure valid language codes
ALTER TABLE user_settings
    ADD CONSTRAINT user_settings_ai_prompt_language_check
    CHECK (ai_prompt_language IN ('en', 'zh'));

ALTER TABLE project_settings
    ADD CONSTRAINT project_settings_ai_prompt_language_check
    CHECK (ai_prompt_language IS NULL OR ai_prompt_language IN ('en', 'zh'));

-- Add comments
COMMENT ON COLUMN user_settings.ai_prompt_language IS 'Language for AI analysis prompts: en (English) or zh (Chinese)';
COMMENT ON COLUMN project_settings.ai_prompt_language IS 'Language for AI analysis prompts: en (English) or zh (Chinese). NULL inherits from user settings.';
