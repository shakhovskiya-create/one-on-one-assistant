-- Local authentication support
-- Allows users to login without connector when AD is unavailable

-- Add password_hash column for local authentication
ALTER TABLE employees ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255);

-- Add is_local_user flag to identify users created locally (not synced from AD)
ALTER TABLE employees ADD COLUMN IF NOT EXISTS is_local_user BOOLEAN DEFAULT FALSE;

-- Index for faster login lookups
CREATE INDEX IF NOT EXISTS idx_employees_email_lower ON employees(lower(email));
CREATE INDEX IF NOT EXISTS idx_employees_ad_login_lower ON employees(lower(ad_login));

-- Create a function to set password for local users
-- Usage: SELECT set_user_password('user@example.com', 'password_hash_here');
CREATE OR REPLACE FUNCTION set_user_password(user_email TEXT, pwd_hash TEXT)
RETURNS BOOLEAN AS $$
BEGIN
    UPDATE employees
    SET password_hash = pwd_hash,
        is_local_user = TRUE
    WHERE lower(email) = lower(user_email);
    RETURN FOUND;
END;
$$ LANGUAGE plpgsql;
