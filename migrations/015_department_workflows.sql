-- Workflow режимы по департаментам
-- Run this in PostgreSQL

-- Таблица workflow режимов
CREATE TABLE IF NOT EXISTS workflow_modes (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,  -- 'simple', 'development'
    description TEXT,
    statuses JSONB NOT NULL,  -- [{id, label, color, wipLimit}]
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Таблица связи департамент -> workflow
CREATE TABLE IF NOT EXISTS department_workflows (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    department VARCHAR(200) NOT NULL UNIQUE,  -- Название департамента из AD
    workflow_mode_id UUID REFERENCES workflow_modes(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Вставка базовых workflow режимов
INSERT INTO workflow_modes (name, description, statuses, is_default) VALUES
('simple', 'Простой Kanban для бизнеса',
 '[{"id":"backlog","label":"Backlog","color":"bg-gray-100","wipLimit":0},{"id":"todo","label":"К выполнению","color":"bg-blue-50","wipLimit":10},{"id":"in_progress","label":"В работе","color":"bg-yellow-50","wipLimit":5},{"id":"done","label":"Готово","color":"bg-green-50","wipLimit":0}]',
 true),
('development', 'Полный процесс разработки',
 '[{"id":"backlog","label":"Backlog","color":"bg-gray-100","wipLimit":0},{"id":"analysis","label":"Анализ","color":"bg-purple-50","wipLimit":5},{"id":"todo","label":"К разработке","color":"bg-blue-50","wipLimit":10},{"id":"in_progress","label":"В разработке","color":"bg-yellow-50","wipLimit":5},{"id":"review","label":"Code Review","color":"bg-orange-50","wipLimit":3},{"id":"testing","label":"Тестирование","color":"bg-pink-50","wipLimit":5},{"id":"done","label":"Готово","color":"bg-green-50","wipLimit":0}]',
 false)
ON CONFLICT (name) DO NOTHING;

-- Настройка для Цифрового развития
INSERT INTO department_workflows (department, workflow_mode_id)
SELECT 'Цифровое развитие', id FROM workflow_modes WHERE name = 'development'
ON CONFLICT (department) DO NOTHING;

-- Индексы
CREATE INDEX IF NOT EXISTS idx_department_workflows_dept ON department_workflows(department);
