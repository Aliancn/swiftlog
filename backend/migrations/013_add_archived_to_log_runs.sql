-- Add archived flag to log_runs table
-- This allows users to hide completed/failed runs from the status page without deleting them

ALTER TABLE log_runs ADD COLUMN IF NOT EXISTS is_archived BOOLEAN NOT NULL DEFAULT FALSE;

-- Add index for better query performance when filtering archived runs
CREATE INDEX IF NOT EXISTS idx_log_runs_is_archived ON log_runs(is_archived);

-- Add composite index for user-based queries with archived filter
CREATE INDEX IF NOT EXISTS idx_log_runs_user_archived ON log_runs(is_archived, created_at DESC);
