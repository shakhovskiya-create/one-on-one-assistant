-- Sprints table for Scrum/Agile sprint management
-- Replaces VARCHAR sprint field with proper sprint entity

CREATE TABLE IF NOT EXISTS sprints (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    goal TEXT,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status VARCHAR(20) DEFAULT 'planning' CHECK (status IN ('planning', 'active', 'completed')),
    velocity INTEGER DEFAULT 0,  -- Calculated from completed story points
    created_by UUID REFERENCES employees(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT valid_dates CHECK (end_date >= start_date)
);

-- Add sprint_id to tasks (replacing VARCHAR sprint field)
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS sprint_id UUID REFERENCES sprints(id) ON DELETE SET NULL;

-- Index for fast lookups
CREATE INDEX IF NOT EXISTS idx_sprints_project ON sprints(project_id);
CREATE INDEX IF NOT EXISTS idx_sprints_status ON sprints(status);
CREATE INDEX IF NOT EXISTS idx_sprints_dates ON sprints(start_date, end_date);
CREATE INDEX IF NOT EXISTS idx_tasks_sprint_id ON tasks(sprint_id);

-- Comments
COMMENT ON TABLE sprints IS 'Scrum sprints for project task grouping';
COMMENT ON COLUMN sprints.status IS 'planning - preparing sprint, active - current sprint, completed - finished';
COMMENT ON COLUMN sprints.velocity IS 'Team velocity (completed story points)';
COMMENT ON COLUMN sprints.goal IS 'Sprint goal/objective';
COMMENT ON COLUMN tasks.sprint_id IS 'Sprint this task belongs to';
