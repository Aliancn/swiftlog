-- Migration: Create log_groups table
-- Description: Organizational unit within a project for grouping related script runs

CREATE TABLE IF NOT EXISTS log_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT group_name_format CHECK (name ~ '^[a-zA-Z0-9 _-]{1,255}$'),
    CONSTRAINT unique_project_group UNIQUE (project_id, name)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_log_groups_project_id ON log_groups(project_id);
CREATE INDEX IF NOT EXISTS idx_log_groups_project_id_name ON log_groups(project_id, name);

-- Comments
COMMENT ON TABLE log_groups IS 'Organizational unit within a project for grouping script runs';
COMMENT ON COLUMN log_groups.id IS 'Unique group identifier';
COMMENT ON COLUMN log_groups.project_id IS 'Foreign key to projects table';
COMMENT ON COLUMN log_groups.name IS 'Group name (unique per project, 1-255 chars)';
COMMENT ON COLUMN log_groups.created_at IS 'Group creation timestamp';
