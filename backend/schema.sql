-- Схема базы данных для 1-on-1 Assistant
-- Выполнить в Supabase SQL Editor

-- Таблица сотрудников
CREATE TABLE employees (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    position VARCHAR(255) NOT NULL,
    meeting_frequency VARCHAR(50) DEFAULT 'weekly',
    meeting_day VARCHAR(20),
    development_priorities TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Таблица встреч
CREATE TABLE meetings (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    employee_id UUID REFERENCES employees(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    duration_minutes INTEGER,
    mood_score INTEGER CHECK (mood_score >= 1 AND mood_score <= 10),
    transcript TEXT,
    summary TEXT,
    analysis JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Таблица договоренностей
CREATE TABLE agreements (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    meeting_id UUID REFERENCES meetings(id) ON DELETE CASCADE,
    task TEXT NOT NULL,
    responsible VARCHAR(255) NOT NULL,
    deadline DATE,
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'overdue', 'cancelled')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Индексы
CREATE INDEX idx_meetings_employee_id ON meetings(employee_id);
CREATE INDEX idx_meetings_date ON meetings(date DESC);
CREATE INDEX idx_agreements_meeting_id ON agreements(meeting_id);
CREATE INDEX idx_agreements_status ON agreements(status);
CREATE INDEX idx_agreements_deadline ON agreements(deadline);

-- RLS (Row Level Security) - опционально
ALTER TABLE employees ENABLE ROW LEVEL SECURITY;
ALTER TABLE meetings ENABLE ROW LEVEL SECURITY;
ALTER TABLE agreements ENABLE ROW LEVEL SECURITY;

-- Политики для авторизованных пользователей
CREATE POLICY "Enable all for authenticated users" ON employees
    FOR ALL USING (auth.role() = 'authenticated');

CREATE POLICY "Enable all for authenticated users" ON meetings
    FOR ALL USING (auth.role() = 'authenticated');

CREATE POLICY "Enable all for authenticated users" ON agreements
    FOR ALL USING (auth.role() = 'authenticated');

-- Функция для автоматической пометки просроченных договоренностей
CREATE OR REPLACE FUNCTION update_overdue_agreements()
RETURNS void AS $$
BEGIN
    UPDATE agreements
    SET status = 'overdue'
    WHERE status = 'pending'
    AND deadline < CURRENT_DATE;
END;
$$ LANGUAGE plpgsql;

-- Триггер для обновления (можно вызывать по расписанию через pg_cron)
-- SELECT cron.schedule('update-overdue', '0 9 * * *', 'SELECT update_overdue_agreements()');
