-- Migration: Update ai_status constraint to include 'none'
-- Description: Adds 'none' to the ai_status check constraint and updates default value

-- Drop the existing constraint
ALTER TABLE log_runs DROP CONSTRAINT IF EXISTS ai_status_valid;

-- Add the new constraint with 'none' included
ALTER TABLE log_runs ADD CONSTRAINT ai_status_valid
    CHECK (ai_status IN ('none', 'pending', 'processing', 'completed', 'failed'));

-- Update the default value for ai_status to 'none' (AI disabled by default)
ALTER TABLE log_runs ALTER COLUMN ai_status SET DEFAULT 'none';

-- Update the column comment
COMMENT ON COLUMN log_runs.ai_status IS 'AI report generation status: none (disabled), pending, processing, completed, failed';
