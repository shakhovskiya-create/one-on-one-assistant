-- Add Telegram bot integration for channels
-- Run this in Supabase SQL Editor

-- Add Telegram bot fields to conversations
ALTER TABLE conversations
ADD COLUMN IF NOT EXISTS telegram_bot_token TEXT;

ALTER TABLE conversations
ADD COLUMN IF NOT EXISTS telegram_chat_id BIGINT;

ALTER TABLE conversations
ADD COLUMN IF NOT EXISTS telegram_enabled BOOLEAN DEFAULT FALSE;

-- Index for Telegram-enabled channels
CREATE INDEX IF NOT EXISTS idx_conversations_telegram ON conversations(telegram_enabled) WHERE telegram_enabled = TRUE;
