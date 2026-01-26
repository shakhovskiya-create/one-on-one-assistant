-- Resource Planning Migration
-- Планирование ресурсов (GAP-006)

-- Extend employees table with resource planning fields
ALTER TABLE employees
ADD COLUMN IF NOT EXISTS work_hours_per_week INTEGER DEFAULT 40,
ADD COLUMN IF NOT EXISTS availability_percent INTEGER DEFAULT 100,
ADD COLUMN IF NOT EXISTS hourly_rate DECIMAL(10,2);

-- Resource Allocation table
CREATE TABLE IF NOT EXISTS resource_allocations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    task_id UUID REFERENCES tasks(id) ON DELETE SET NULL,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,

    role VARCHAR(100),  -- Role on the task/project

    allocated_hours_per_week INTEGER NOT NULL DEFAULT 0,
    period_start DATE NOT NULL,
    period_end DATE,

    notes TEXT,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES employees(id)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_resource_allocations_employee ON resource_allocations(employee_id);
CREATE INDEX IF NOT EXISTS idx_resource_allocations_task ON resource_allocations(task_id);
CREATE INDEX IF NOT EXISTS idx_resource_allocations_project ON resource_allocations(project_id);
CREATE INDEX IF NOT EXISTS idx_resource_allocations_period ON resource_allocations(period_start, period_end);

-- Absences table (for future integration with Exchange/EWS)
CREATE TABLE IF NOT EXISTS employee_absences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    absence_type VARCHAR(50) NOT NULL,  -- vacation, sick_leave, holiday, out_of_office
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    description TEXT,
    source VARCHAR(50) DEFAULT 'manual',  -- manual, exchange, hr_system
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_employee_absences_employee ON employee_absences(employee_id);
CREATE INDEX IF NOT EXISTS idx_employee_absences_dates ON employee_absences(start_date, end_date);
