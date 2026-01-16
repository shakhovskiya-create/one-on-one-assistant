-- Схема базы данных для Meeting Assistant v3.0
-- Выполнить в Supabase SQL Editor

-- ============ CORE TABLES ============

-- Сотрудники
CREATE TABLE employees (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    position VARCHAR(255) NOT NULL,
    meeting_frequency VARCHAR(50) DEFAULT 'weekly',
    meeting_day VARCHAR(20),
    development_priorities TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Проекты
CREATE TABLE projects (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) DEFAULT 'active' CHECK (status IN ('active', 'completed', 'on_hold', 'cancelled')),
    start_date DATE,
    end_date DATE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Категории встреч
CREATE TABLE meeting_categories (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    prompt_template TEXT,
    analysis_fields JSONB,
    is_system BOOLEAN DEFAULT false
);

-- Предзаполненные категории
INSERT INTO meeting_categories (code, name, description, is_system) VALUES
    ('one_on_one', '1-на-1', 'Индивидуальная встреча руководителя с сотрудником', true),
    ('team_meeting', 'Совещание', 'Командное совещание по проекту или задачам', true),
    ('planning', 'Планирование', 'Планирование спринта/проекта', true),
    ('retro', 'Ретроспектива', 'Анализ прошедшего периода, уроки', true),
    ('kickoff', 'Kickoff', 'Старт проекта или инициативы', true),
    ('interview', 'Интервью', 'Собеседование с кандидатом', true),
    ('status', 'Статус-митинг', 'Регулярный статус по проекту', true),
    ('demo', 'Демо', 'Демонстрация результатов', true);

-- Встречи (расширенная)
CREATE TABLE meetings (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    title VARCHAR(500),
    employee_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    category_id UUID REFERENCES meeting_categories(id) ON DELETE SET NULL,
    date DATE NOT NULL,
    duration_minutes INTEGER,
    duration_seconds INTEGER,
    mood_score INTEGER CHECK (mood_score >= 1 AND mood_score <= 10),
    -- Транскрипты
    transcript TEXT,
    transcript_whisper TEXT,
    transcript_yandex TEXT,
    transcript_merged TEXT,
    -- Анализ
    summary TEXT,
    analysis JSONB,
    audio_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Участники встреч (для командных встреч)
CREATE TABLE meeting_participants (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    meeting_id UUID REFERENCES meetings(id) ON DELETE CASCADE,
    employee_id UUID REFERENCES employees(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(meeting_id, employee_id)
);

-- Договорённости из встреч
CREATE TABLE agreements (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    meeting_id UUID REFERENCES meetings(id) ON DELETE CASCADE,
    task TEXT NOT NULL,
    responsible VARCHAR(255) NOT NULL,
    deadline DATE,
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'overdue', 'cancelled')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============ TASK MANAGEMENT ============

-- Теги
CREATE TABLE tags (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    color VARCHAR(20) DEFAULT 'gray',
    is_system BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Предзаполненные теги
INSERT INTO tags (name, color, is_system) VALUES
    ('срочно', 'red', true),
    ('важно', 'orange', true),
    ('блокер', 'red', true),
    ('баг', 'red', true),
    ('улучшение', 'blue', true),
    ('документация', 'purple', true);

-- Задачи
CREATE TABLE tasks (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    status VARCHAR(50) DEFAULT 'backlog' CHECK (status IN ('backlog', 'todo', 'in_progress', 'review', 'done')),
    priority INTEGER DEFAULT 3 CHECK (priority >= 1 AND priority <= 5),
    flag_color VARCHAR(20),
    -- Связи
    assignee_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    co_assignee_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    creator_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    meeting_id UUID REFERENCES meetings(id) ON DELETE SET NULL,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    parent_id UUID REFERENCES tasks(id) ON DELETE SET NULL,
    -- Эпики
    is_epic BOOLEAN DEFAULT false,
    -- Даты
    due_date DATE,
    original_due_date DATE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Связь задач с тегами
CREATE TABLE task_tags (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    task_id UUID REFERENCES tasks(id) ON DELETE CASCADE,
    tag_id UUID REFERENCES tags(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(task_id, tag_id)
);

-- Связи между задачами (блокирует, связана с)
CREATE TABLE task_links (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    source_task_id UUID REFERENCES tasks(id) ON DELETE CASCADE,
    target_task_id UUID REFERENCES tasks(id) ON DELETE CASCADE,
    link_type VARCHAR(50) NOT NULL CHECK (link_type IN ('blocks', 'blocked_by', 'relates_to', 'duplicates')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(source_task_id, target_task_id, link_type)
);

-- Комментарии к задачам
CREATE TABLE task_comments (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    task_id UUID REFERENCES tasks(id) ON DELETE CASCADE,
    author_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- История изменений задач
CREATE TABLE task_history (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    task_id UUID REFERENCES tasks(id) ON DELETE CASCADE,
    field_name VARCHAR(100) NOT NULL,
    old_value TEXT,
    new_value TEXT,
    changed_by UUID REFERENCES employees(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============ TELEGRAM INTEGRATION ============

CREATE TABLE telegram_users (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    employee_id UUID REFERENCES employees(id) ON DELETE CASCADE UNIQUE,
    telegram_username VARCHAR(100),
    telegram_chat_id BIGINT,
    notifications_enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============ INDEXES ============

-- Meetings
CREATE INDEX idx_meetings_employee_id ON meetings(employee_id);
CREATE INDEX idx_meetings_project_id ON meetings(project_id);
CREATE INDEX idx_meetings_category_id ON meetings(category_id);
CREATE INDEX idx_meetings_date ON meetings(date DESC);

-- Agreements
CREATE INDEX idx_agreements_meeting_id ON agreements(meeting_id);
CREATE INDEX idx_agreements_status ON agreements(status);
CREATE INDEX idx_agreements_deadline ON agreements(deadline);

-- Tasks
CREATE INDEX idx_tasks_assignee_id ON tasks(assignee_id);
CREATE INDEX idx_tasks_project_id ON tasks(project_id);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_parent_id ON tasks(parent_id);
CREATE INDEX idx_tasks_due_date ON tasks(due_date);

-- Meeting participants
CREATE INDEX idx_meeting_participants_meeting ON meeting_participants(meeting_id);
CREATE INDEX idx_meeting_participants_employee ON meeting_participants(employee_id);

-- Task history
CREATE INDEX idx_task_history_task_id ON task_history(task_id);

-- ============ ROW LEVEL SECURITY ============

ALTER TABLE employees ENABLE ROW LEVEL SECURITY;
ALTER TABLE projects ENABLE ROW LEVEL SECURITY;
ALTER TABLE meetings ENABLE ROW LEVEL SECURITY;
ALTER TABLE meeting_participants ENABLE ROW LEVEL SECURITY;
ALTER TABLE agreements ENABLE ROW LEVEL SECURITY;
ALTER TABLE tasks ENABLE ROW LEVEL SECURITY;
ALTER TABLE tags ENABLE ROW LEVEL SECURITY;
ALTER TABLE task_tags ENABLE ROW LEVEL SECURITY;
ALTER TABLE task_comments ENABLE ROW LEVEL SECURITY;
ALTER TABLE task_history ENABLE ROW LEVEL SECURITY;
ALTER TABLE telegram_users ENABLE ROW LEVEL SECURITY;

-- Политики для authenticated users (упрощённые - полный доступ)
CREATE POLICY "Enable all for authenticated" ON employees FOR ALL USING (auth.role() = 'authenticated');
CREATE POLICY "Enable all for authenticated" ON projects FOR ALL USING (auth.role() = 'authenticated');
CREATE POLICY "Enable all for authenticated" ON meetings FOR ALL USING (auth.role() = 'authenticated');
CREATE POLICY "Enable all for authenticated" ON meeting_participants FOR ALL USING (auth.role() = 'authenticated');
CREATE POLICY "Enable all for authenticated" ON agreements FOR ALL USING (auth.role() = 'authenticated');
CREATE POLICY "Enable all for authenticated" ON tasks FOR ALL USING (auth.role() = 'authenticated');
CREATE POLICY "Enable all for authenticated" ON tags FOR ALL USING (auth.role() = 'authenticated');
CREATE POLICY "Enable all for authenticated" ON task_tags FOR ALL USING (auth.role() = 'authenticated');
CREATE POLICY "Enable all for authenticated" ON task_comments FOR ALL USING (auth.role() = 'authenticated');
CREATE POLICY "Enable all for authenticated" ON task_history FOR ALL USING (auth.role() = 'authenticated');
CREATE POLICY "Enable all for authenticated" ON telegram_users FOR ALL USING (auth.role() = 'authenticated');

-- Политики для service_role (полный доступ без ограничений)
CREATE POLICY "Enable all for service_role" ON employees FOR ALL USING (auth.role() = 'service_role');
CREATE POLICY "Enable all for service_role" ON projects FOR ALL USING (auth.role() = 'service_role');
CREATE POLICY "Enable all for service_role" ON meetings FOR ALL USING (auth.role() = 'service_role');
CREATE POLICY "Enable all for service_role" ON meeting_participants FOR ALL USING (auth.role() = 'service_role');
CREATE POLICY "Enable all for service_role" ON agreements FOR ALL USING (auth.role() = 'service_role');
CREATE POLICY "Enable all for service_role" ON tasks FOR ALL USING (auth.role() = 'service_role');
CREATE POLICY "Enable all for service_role" ON tags FOR ALL USING (auth.role() = 'service_role');
CREATE POLICY "Enable all for service_role" ON task_tags FOR ALL USING (auth.role() = 'service_role');
CREATE POLICY "Enable all for service_role" ON task_comments FOR ALL USING (auth.role() = 'service_role');
CREATE POLICY "Enable all for service_role" ON task_history FOR ALL USING (auth.role() = 'service_role');
CREATE POLICY "Enable all for service_role" ON telegram_users FOR ALL USING (auth.role() = 'service_role');

-- ============ FUNCTIONS ============

-- Автоматическая пометка просроченных договоренностей
CREATE OR REPLACE FUNCTION update_overdue_agreements()
RETURNS void AS $$
BEGIN
    UPDATE agreements
    SET status = 'overdue'
    WHERE status = 'pending'
    AND deadline < CURRENT_DATE;
END;
$$ LANGUAGE plpgsql;

-- Автоматическое обновление updated_at для tasks
CREATE OR REPLACE FUNCTION update_task_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_task_timestamp
    BEFORE UPDATE ON tasks
    FOR EACH ROW
    EXECUTE FUNCTION update_task_timestamp();

-- ============ NOTES ============
--
-- Для запуска автоматической пометки просроченных договоренностей:
-- SELECT cron.schedule('update-overdue', '0 9 * * *', 'SELECT update_overdue_agreements()');
--
-- Для использования Yandex SpeechKit добавьте переменные:
-- YANDEX_API_KEY, YANDEX_FOLDER_ID
