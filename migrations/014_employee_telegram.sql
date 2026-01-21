-- Add Telegram integration for employees (personal notifications)
-- Run this in PostgreSQL

-- Add Telegram chat ID for personal notifications
ALTER TABLE employees
ADD COLUMN IF NOT EXISTS telegram_chat_id BIGINT;

-- Index for Telegram-linked employees
CREATE INDEX IF NOT EXISTS idx_employees_telegram ON employees(telegram_chat_id) WHERE telegram_chat_id IS NOT NULL;

-- Add telegram_username for linking
ALTER TABLE employees
ADD COLUMN IF NOT EXISTS telegram_username VARCHAR(100);
