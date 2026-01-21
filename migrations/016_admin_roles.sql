-- Admin roles and audit logging
-- Run this in PostgreSQL

-- Add role field to employees
ALTER TABLE employees ADD COLUMN IF NOT EXISTS role VARCHAR(20) DEFAULT 'user';
-- Roles: 'user', 'admin', 'super_admin'

-- Create audit_logs table
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID REFERENCES employees(id),
    action VARCHAR(100) NOT NULL,  -- 'user.create', 'user.update', 'workflow.change', etc.
    entity_type VARCHAR(50),       -- 'employee', 'task', 'workflow', etc.
    entity_id UUID,
    old_value JSONB,
    new_value JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- System settings table
CREATE TABLE IF NOT EXISTS system_settings (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    key VARCHAR(100) NOT NULL UNIQUE,
    value JSONB NOT NULL,
    description TEXT,
    updated_by UUID REFERENCES employees(id),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert default system settings
INSERT INTO system_settings (key, value, description) VALUES
('allow_self_registration', 'false', 'Allow users to register without admin approval'),
('default_workflow', '"simple"', 'Default workflow mode for new departments'),
('session_timeout_minutes', '480', 'Session timeout in minutes (8 hours default)'),
('max_file_size_mb', '100', 'Maximum file upload size in MB'),
('maintenance_mode', 'false', 'Enable maintenance mode (only admins can access)')
ON CONFLICT (key) DO NOTHING;

-- Indexes
CREATE INDEX IF NOT EXISTS idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created ON audit_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_employees_role ON employees(role);

-- Set first user as super_admin (optional - run manually if needed)
-- UPDATE employees SET role = 'super_admin' WHERE id = (SELECT id FROM employees ORDER BY created_at LIMIT 1);
