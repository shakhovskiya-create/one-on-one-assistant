-- Versions/Releases table for JIRA-like release management
CREATE TABLE IF NOT EXISTS versions (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    status VARCHAR(20) DEFAULT 'unreleased' CHECK (status IN ('unreleased', 'released', 'archived')),
    start_date DATE,
    release_date DATE,
    released_at TIMESTAMPTZ,
    created_by UUID REFERENCES employees(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Add fix_version to tasks
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS fix_version_id UUID REFERENCES versions(id) ON DELETE SET NULL;

-- Index for fast lookups
CREATE INDEX IF NOT EXISTS idx_versions_project ON versions(project_id);
CREATE INDEX IF NOT EXISTS idx_versions_status ON versions(status);
CREATE INDEX IF NOT EXISTS idx_tasks_fix_version ON tasks(fix_version_id);

-- Comment
COMMENT ON TABLE versions IS 'Release versions for projects (JIRA-like)';
COMMENT ON COLUMN versions.status IS 'unreleased - in progress, released - done, archived - old';
COMMENT ON COLUMN tasks.fix_version_id IS 'Target release version for this task';
