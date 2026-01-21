-- Media messages support (voice, video)
-- Run this in Supabase SQL Editor

-- Update message_type constraint to allow voice and video
ALTER TABLE messages DROP CONSTRAINT IF EXISTS messages_message_type_check;
ALTER TABLE messages ADD CONSTRAINT messages_message_type_check
    CHECK (message_type IN ('text', 'file', 'system', 'voice', 'video', 'sticker', 'gif'));

-- Add media-specific fields to messages
ALTER TABLE messages ADD COLUMN IF NOT EXISTS file_id UUID REFERENCES files(id) ON DELETE SET NULL;
ALTER TABLE messages ADD COLUMN IF NOT EXISTS duration_seconds INTEGER;
ALTER TABLE messages ADD COLUMN IF NOT EXISTS thumbnail_url TEXT;
ALTER TABLE messages ADD COLUMN IF NOT EXISTS transcription TEXT;
ALTER TABLE messages ADD COLUMN IF NOT EXISTS waveform JSONB; -- For voice message waveform visualization

-- Add reactions support
CREATE TABLE IF NOT EXISTS message_reactions (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    message_id UUID REFERENCES messages(id) ON DELETE CASCADE,
    employee_id UUID REFERENCES employees(id) ON DELETE CASCADE,
    emoji VARCHAR(10) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(message_id, employee_id, emoji)
);

-- Index for reactions
CREATE INDEX IF NOT EXISTS idx_message_reactions_message ON message_reactions(message_id);

-- Enable RLS for reactions
ALTER TABLE message_reactions ENABLE ROW LEVEL SECURITY;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE tablename = 'message_reactions' AND policyname = 'Enable all for service_role') THEN
        CREATE POLICY "Enable all for service_role" ON message_reactions FOR ALL USING (auth.role() = 'service_role');
    END IF;
END $$;

-- Sticker packs support (for future Sprint 2)
CREATE TABLE IF NOT EXISTS sticker_packs (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    thumbnail_url TEXT,
    is_official BOOLEAN DEFAULT FALSE,
    created_by UUID REFERENCES employees(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS stickers (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    pack_id UUID REFERENCES sticker_packs(id) ON DELETE CASCADE,
    emoji VARCHAR(10), -- Associated emoji
    file_id UUID REFERENCES files(id),
    url TEXT NOT NULL,
    width INTEGER,
    height INTEGER,
    is_animated BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Enable RLS for stickers
ALTER TABLE sticker_packs ENABLE ROW LEVEL SECURITY;
ALTER TABLE stickers ENABLE ROW LEVEL SECURITY;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE tablename = 'sticker_packs' AND policyname = 'Enable all for service_role') THEN
        CREATE POLICY "Enable all for service_role" ON sticker_packs FOR ALL USING (auth.role() = 'service_role');
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE tablename = 'stickers' AND policyname = 'Enable all for service_role') THEN
        CREATE POLICY "Enable all for service_role" ON stickers FOR ALL USING (auth.role() = 'service_role');
    END IF;
END $$;

-- Index for file lookups
CREATE INDEX IF NOT EXISTS idx_messages_file ON messages(file_id) WHERE file_id IS NOT NULL;
