-- Fix RLS policies for messenger tables
-- Run this in Supabase SQL Editor

-- Option 1: Disable RLS (simplest, backend handles auth)
ALTER TABLE conversations DISABLE ROW LEVEL SECURITY;
ALTER TABLE conversation_participants DISABLE ROW LEVEL SECURITY;
ALTER TABLE messages DISABLE ROW LEVEL SECURITY;

-- Or if you prefer to keep RLS enabled, use these permissive policies:
-- DROP POLICY IF EXISTS "Enable all for service_role" ON conversations;
-- DROP POLICY IF EXISTS "Enable all for service_role" ON conversation_participants;
-- DROP POLICY IF EXISTS "Enable all for service_role" ON messages;

-- CREATE POLICY "Allow all" ON conversations FOR ALL USING (true) WITH CHECK (true);
-- CREATE POLICY "Allow all" ON conversation_participants FOR ALL USING (true) WITH CHECK (true);
-- CREATE POLICY "Allow all" ON messages FOR ALL USING (true) WITH CHECK (true);
