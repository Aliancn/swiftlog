-- Migration: Create log_runs table
-- Description: Represents a single execution of a logged script

CREATE TABLE IF NOT EXISTS log_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES log_groups(id) ON DELETE CASCADE,
    start_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    end_time TIMESTAMPTZ,
    status VARCHAR(50) NOT NULL DEFAULT 'running',
    exit_code INTEGER,
    ai_report TEXT,
    ai_status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT status_valid CHECK (status IN ('running', 'completed', 'failed', 'aborted')),
    CONSTRAINT exit_code_range CHECK (exit_code IS NULL OR (exit_code >= -128 AND exit_code <= 255)),
    CONSTRAINT ai_status_valid CHECK (ai_status IN ('pending', 'processing', 'completed', 'failed')),
    CONSTRAINT end_time_after_start CHECK (end_time IS NULL OR end_time >= start_time)
);

-- Foreign key indexes
CREATE INDEX IF NOT EXISTS idx_log_runs_group_id ON log_runs(group_id);

-- Query optimization indexes
CREATE INDEX IF NOT EXISTS idx_log_runs_group_id_start_time ON log_runs(group_id, start_time DESC);
CREATE INDEX IF NOT EXISTS idx_log_runs_start_time ON log_runs(start_time DESC);

-- Partial indexes (space-efficient)
CREATE INDEX IF NOT EXISTS idx_log_runs_status_failed ON log_runs(status) WHERE status IN ('failed', 'aborted');
CREATE INDEX IF NOT EXISTS idx_log_runs_ai_status_pending ON log_runs(ai_status) WHERE ai_status = 'pending';

-- Comments
COMMENT ON TABLE log_runs IS 'Represents a single execution of a logged script';
COMMENT ON COLUMN log_runs.id IS 'Unique run identifier (used as Loki label for log correlation)';
COMMENT ON COLUMN log_runs.group_id IS 'Foreign key to log_groups table';
COMMENT ON COLUMN log_runs.start_time IS 'Script execution start timestamp';
COMMENT ON COLUMN log_runs.end_time IS 'Script execution end timestamp (null if still running)';
COMMENT ON COLUMN log_runs.status IS 'Execution status: running, completed, failed, aborted';
COMMENT ON COLUMN log_runs.exit_code IS 'Script exit code (-128 to 255, null if not finished)';
COMMENT ON COLUMN log_runs.ai_report IS 'AI-generated analysis report (null if not generated)';
COMMENT ON COLUMN log_runs.ai_status IS 'AI report generation status: pending, processing, completed, failed';

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_log_runs_updated_at ON log_runs;
CREATE TRIGGER update_log_runs_updated_at
    BEFORE UPDATE ON log_runs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
