-- Migration: AD and Exchange Integration
-- Run this in Supabase SQL Editor to add AD/Exchange support

-- ============ EMPLOYEES: AD FIELDS ============

-- Email (for Exchange integration)
ALTER TABLE employees ADD COLUMN IF NOT EXISTS email VARCHAR(255) UNIQUE;

-- AD fields
ALTER TABLE employees ADD COLUMN IF NOT EXISTS ad_dn TEXT;  -- Distinguished Name
ALTER TABLE employees ADD COLUMN IF NOT EXISTS ad_login VARCHAR(100);  -- sAMAccountName
ALTER TABLE employees ADD COLUMN IF NOT EXISTS department VARCHAR(255);
ALTER TABLE employees ADD COLUMN IF NOT EXISTS manager_id UUID REFERENCES employees(id);
ALTER TABLE employees ADD COLUMN IF NOT EXISTS manager_dn TEXT;  -- For AD sync
ALTER TABLE employees ADD COLUMN IF NOT EXISTS photo_base64 TEXT;  -- AD thumbnail photo
ALTER TABLE employees ADD COLUMN IF NOT EXISTS phone VARCHAR(50);
ALTER TABLE employees ADD COLUMN IF NOT EXISTS mobile VARCHAR(50);
ALTER TABLE employees ADD COLUMN IF NOT EXISTS last_ad_sync TIMESTAMP WITH TIME ZONE;

-- Index for AD lookup
CREATE INDEX IF NOT EXISTS idx_employees_email ON employees(email);
CREATE INDEX IF NOT EXISTS idx_employees_ad_dn ON employees(ad_dn);
CREATE INDEX IF NOT EXISTS idx_employees_ad_login ON employees(ad_login);
CREATE INDEX IF NOT EXISTS idx_employees_manager_id ON employees(manager_id);

-- ============ MEETINGS: EXCHANGE FIELDS ============

ALTER TABLE meetings ADD COLUMN IF NOT EXISTS exchange_id VARCHAR(500);  -- EWS Item ID
ALTER TABLE meetings ADD COLUMN IF NOT EXISTS exchange_data JSONB;  -- Full Exchange event data
ALTER TABLE meetings ADD COLUMN IF NOT EXISTS location VARCHAR(500);
ALTER TABLE meetings ADD COLUMN IF NOT EXISTS start_time TIMESTAMP WITH TIME ZONE;
ALTER TABLE meetings ADD COLUMN IF NOT EXISTS end_time TIMESTAMP WITH TIME ZONE;
ALTER TABLE meetings ADD COLUMN IF NOT EXISTS organizer_id UUID REFERENCES employees(id);
ALTER TABLE meetings ADD COLUMN IF NOT EXISTS is_recurring BOOLEAN DEFAULT false;
ALTER TABLE meetings ADD COLUMN IF NOT EXISTS recurrence_pattern JSONB;
ALTER TABLE meetings ADD COLUMN IF NOT EXISTS last_exchange_sync TIMESTAMP WITH TIME ZONE;

-- Index for Exchange lookup
CREATE INDEX IF NOT EXISTS idx_meetings_exchange_id ON meetings(exchange_id);
CREATE INDEX IF NOT EXISTS idx_meetings_organizer_id ON meetings(organizer_id);
CREATE INDEX IF NOT EXISTS idx_meetings_start_time ON meetings(start_time);

-- ============ USER SESSIONS (for AD auth) ============

CREATE TABLE IF NOT EXISTS user_sessions (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    employee_id UUID REFERENCES employees(id) ON DELETE CASCADE,
    token VARCHAR(500) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    ip_address VARCHAR(50),
    user_agent TEXT
);

CREATE INDEX IF NOT EXISTS idx_user_sessions_token ON user_sessions(token);
CREATE INDEX IF NOT EXISTS idx_user_sessions_employee ON user_sessions(employee_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires ON user_sessions(expires_at);

-- RLS for sessions
ALTER TABLE user_sessions ENABLE ROW LEVEL SECURITY;
CREATE POLICY "Enable all for service_role" ON user_sessions FOR ALL USING (auth.role() = 'service_role');

-- ============ ORG STRUCTURE VIEW ============

CREATE OR REPLACE VIEW org_tree AS
WITH RECURSIVE subordinates AS (
    -- Base: top managers (no manager)
    SELECT
        id,
        name,
        position,
        department,
        email,
        manager_id,
        1 as level,
        ARRAY[id] as path
    FROM employees
    WHERE manager_id IS NULL

    UNION ALL

    -- Recursive: employees with managers
    SELECT
        e.id,
        e.name,
        e.position,
        e.department,
        e.email,
        e.manager_id,
        s.level + 1,
        s.path || e.id
    FROM employees e
    INNER JOIN subordinates s ON e.manager_id = s.id
)
SELECT * FROM subordinates;

-- ============ HELPER FUNCTION: Get all subordinates ============

CREATE OR REPLACE FUNCTION get_all_subordinates(manager_uuid UUID)
RETURNS TABLE (
    id UUID,
    name VARCHAR,
    position VARCHAR,
    department VARCHAR,
    level INTEGER
) AS $$
WITH RECURSIVE subs AS (
    SELECT
        e.id,
        e.name,
        e.position,
        e.department,
        1 as level
    FROM employees e
    WHERE e.manager_id = manager_uuid

    UNION ALL

    SELECT
        e.id,
        e.name,
        e.position,
        e.department,
        s.level + 1
    FROM employees e
    INNER JOIN subs s ON e.manager_id = s.id
)
SELECT * FROM subs;
$$ LANGUAGE SQL;

-- ============ NOTES ============
--
-- After running this migration:
-- 1. Set CONNECTOR_API_KEY in Railway environment
-- 2. Run the on-prem connector on your Mac
-- 3. Call POST /ad/sync to import users from AD
-- 4. Users will be linked by email to Exchange calendar
--
