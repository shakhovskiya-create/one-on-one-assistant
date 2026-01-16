-- Add Gantt chart features to tasks table
-- Run this in Supabase SQL Editor

-- Add start_date for Gantt chart visualization
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS start_date DATE;

-- Add progress percentage for task completion tracking
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS progress INTEGER DEFAULT 0;

-- Create task dependencies table for Gantt relationships
CREATE TABLE IF NOT EXISTS task_dependencies (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    task_id UUID REFERENCES tasks(id) ON DELETE CASCADE,
    depends_on_task_id UUID REFERENCES tasks(id) ON DELETE CASCADE,
    dependency_type VARCHAR(20) DEFAULT 'finish_to_start', -- finish_to_start, start_to_start, finish_to_finish, start_to_finish
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(task_id, depends_on_task_id)
);

-- Index for fast dependency lookups
CREATE INDEX IF NOT EXISTS idx_task_dependencies_task ON task_dependencies(task_id);
CREATE INDEX IF NOT EXISTS idx_task_dependencies_depends ON task_dependencies(depends_on_task_id);

-- Update existing tasks to have start_date = created_at or due_date - 7 days
UPDATE tasks
SET start_date = COALESCE(
    created_at::date,
    (COALESCE(due_date::date, CURRENT_DATE) - INTERVAL '7 days')::date
)
WHERE start_date IS NULL;
