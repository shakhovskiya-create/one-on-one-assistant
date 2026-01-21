-- Resource planning fields for tasks and employees
-- Run this in PostgreSQL

-- Add resource planning fields to tasks
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS estimated_hours DECIMAL(10, 2);
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS actual_hours DECIMAL(10, 2);
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS estimated_cost DECIMAL(12, 2);
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS actual_cost DECIMAL(12, 2);

-- Add hourly rate to employees
ALTER TABLE employees ADD COLUMN IF NOT EXISTS hourly_rate DECIMAL(10, 2);

-- Add time tracking entries table
CREATE TABLE IF NOT EXISTS time_entries (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    task_id UUID REFERENCES tasks(id) ON DELETE CASCADE,
    employee_id UUID REFERENCES employees(id) ON DELETE CASCADE,
    hours DECIMAL(10, 2) NOT NULL,
    description TEXT,
    date DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for time entries
CREATE INDEX IF NOT EXISTS idx_time_entries_task ON time_entries(task_id);
CREATE INDEX IF NOT EXISTS idx_time_entries_employee ON time_entries(employee_id);
CREATE INDEX IF NOT EXISTS idx_time_entries_date ON time_entries(date);

-- View for task cost calculation
CREATE OR REPLACE VIEW task_costs AS
SELECT
    t.id AS task_id,
    t.title,
    t.estimated_hours,
    t.actual_hours,
    t.estimated_cost,
    t.actual_cost,
    COALESCE(SUM(te.hours), 0) AS logged_hours,
    e.hourly_rate,
    COALESCE(SUM(te.hours), 0) * COALESCE(e.hourly_rate, 0) AS calculated_cost
FROM tasks t
LEFT JOIN time_entries te ON t.id = te.task_id
LEFT JOIN employees e ON t.assignee_id = e.id
GROUP BY t.id, t.title, t.estimated_hours, t.actual_hours, t.estimated_cost, t.actual_cost, e.hourly_rate;

-- Function to update actual_hours on task from time entries
CREATE OR REPLACE FUNCTION update_task_actual_hours()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        UPDATE tasks
        SET actual_hours = (
            SELECT COALESCE(SUM(hours), 0)
            FROM time_entries
            WHERE task_id = OLD.task_id
        )
        WHERE id = OLD.task_id;
        RETURN OLD;
    ELSE
        UPDATE tasks
        SET actual_hours = (
            SELECT COALESCE(SUM(hours), 0)
            FROM time_entries
            WHERE task_id = NEW.task_id
        )
        WHERE id = NEW.task_id;
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Trigger to auto-update actual_hours
DROP TRIGGER IF EXISTS trigger_update_task_actual_hours ON time_entries;
CREATE TRIGGER trigger_update_task_actual_hours
AFTER INSERT OR UPDATE OR DELETE ON time_entries
FOR EACH ROW
EXECUTE FUNCTION update_task_actual_hours();
