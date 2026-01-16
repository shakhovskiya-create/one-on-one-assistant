-- File storage metadata tracking
CREATE TABLE IF NOT EXISTS files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    storage_path VARCHAR(500) NOT NULL,
    bucket VARCHAR(100) NOT NULL DEFAULT 'attachments',
    content_type VARCHAR(100),
    size_bytes BIGINT,
    uploaded_by UUID REFERENCES employees(id),
    entity_type VARCHAR(50), -- 'meeting', 'task', 'message', 'employee'
    entity_id UUID,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for faster lookups
CREATE INDEX IF NOT EXISTS idx_files_entity ON files(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_files_uploaded_by ON files(uploaded_by);
CREATE INDEX IF NOT EXISTS idx_files_bucket ON files(bucket);

-- Add attachments support to messages
ALTER TABLE messages ADD COLUMN IF NOT EXISTS attachments JSONB DEFAULT '[]'::jsonb;

-- Add attachments support to tasks
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS attachments JSONB DEFAULT '[]'::jsonb;
