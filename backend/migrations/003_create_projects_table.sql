-- Migration: Create projects table
-- Description: Top-level container for organizing logs, owned by a user

CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT project_name_format CHECK (name ~ '^[a-zA-Z0-9 _-]{1,255}$'),
    CONSTRAINT unique_user_project UNIQUE (user_id, name)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_projects_user_id ON projects(user_id);
CREATE INDEX IF NOT EXISTS idx_projects_user_id_name ON projects(user_id, name);

-- Comments
COMMENT ON TABLE projects IS 'Top-level container for organizing script logs';
COMMENT ON COLUMN projects.id IS 'Unique project identifier';
COMMENT ON COLUMN projects.user_id IS 'Foreign key to users table (owner)';
COMMENT ON COLUMN projects.name IS 'Project name (unique per user, 1-255 chars)';
COMMENT ON COLUMN projects.created_at IS 'Project creation timestamp';
