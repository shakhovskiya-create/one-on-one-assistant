-- Force password change on first login
ALTER TABLE employees ADD COLUMN IF NOT EXISTS force_password_change BOOLEAN DEFAULT TRUE;

-- Set all users to require password change
UPDATE employees SET force_password_change = TRUE WHERE email IS NOT NULL;
