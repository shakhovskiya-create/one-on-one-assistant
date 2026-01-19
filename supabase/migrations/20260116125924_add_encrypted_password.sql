-- Add encrypted_password column to employees table for storing user EWS passwords
ALTER TABLE employees ADD COLUMN IF NOT EXISTS encrypted_password TEXT;

-- Create index for faster lookups (optional)
CREATE INDEX IF NOT EXISTS idx_employees_encrypted_password ON employees(id) WHERE encrypted_password IS NOT NULL;
