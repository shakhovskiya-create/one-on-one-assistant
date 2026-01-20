-- Add Agile/Scrum fields to tasks table
-- story_points for task estimation (Fibonacci: 1, 2, 3, 5, 8, 13, 21)
-- sprint for grouping tasks into sprints

ALTER TABLE tasks ADD COLUMN IF NOT EXISTS story_points INTEGER;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS sprint VARCHAR(100);

-- Add index for sprint filtering
CREATE INDEX IF NOT EXISTS idx_tasks_sprint ON tasks(sprint);

-- Add comment
COMMENT ON COLUMN tasks.story_points IS 'Story points estimation (Fibonacci sequence)';
COMMENT ON COLUMN tasks.sprint IS 'Sprint name/identifier';
