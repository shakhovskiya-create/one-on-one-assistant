-- Add channel support to messenger
-- Run this in Supabase SQL Editor

-- Add channel type and description to conversations
ALTER TABLE conversations
DROP CONSTRAINT IF EXISTS conversations_type_check;

ALTER TABLE conversations
ADD CONSTRAINT conversations_type_check
CHECK (type IN ('direct', 'group', 'channel'));

-- Add description column for channels
ALTER TABLE conversations
ADD COLUMN IF NOT EXISTS description TEXT;

-- Add created_by for channels (owner)
ALTER TABLE conversations
ADD COLUMN IF NOT EXISTS created_by UUID REFERENCES employees(id);

-- Add is_admin flag to participants for channels
ALTER TABLE conversation_participants
ADD COLUMN IF NOT EXISTS is_admin BOOLEAN DEFAULT FALSE;

-- Create index for channel lookups
CREATE INDEX IF NOT EXISTS idx_conversations_type ON conversations(type);
