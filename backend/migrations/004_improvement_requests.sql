-- Improvement Requests Migration
-- Заявки на улучшение (GAP-005)

-- Improvement Request Types (for categorization)
CREATE TABLE IF NOT EXISTS improvement_request_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    icon VARCHAR(50),
    color VARCHAR(20),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Improvement Requests
CREATE TABLE IF NOT EXISTS improvement_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    number VARCHAR(20) NOT NULL UNIQUE,  -- IR-2026-0001
    title VARCHAR(500) NOT NULL,
    description TEXT,
    business_value TEXT,  -- Бизнес-ценность
    expected_effect TEXT,  -- Ожидаемый эффект (KPI / финансовый)

    -- Organizational
    initiator_id UUID NOT NULL REFERENCES employees(id),
    department_id UUID REFERENCES departments(id),
    sponsor_id UUID REFERENCES employees(id),

    -- Budget
    estimated_budget DECIMAL(15,2),  -- Предварительный бюджет
    approved_budget DECIMAL(15,2),   -- Утверждённый бюджет

    -- Timeline
    estimated_start DATE,
    estimated_end DATE,

    -- Status (9-stage lifecycle)
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    -- draft, submitted, screening, evaluation, manager_approval, committee_review, budgeting, project_created, in_progress, completed, rejected

    -- Committee
    committee_date DATE,
    committee_decision TEXT,

    -- Link to project (after approval)
    project_id UUID REFERENCES projects(id),

    -- Rejection
    rejection_reason TEXT,
    rejected_by UUID REFERENCES employees(id),
    rejected_at TIMESTAMP WITH TIME ZONE,

    -- Type/Category
    type_id UUID REFERENCES improvement_request_types(id),
    priority VARCHAR(20) DEFAULT 'medium',  -- low, medium, high, critical

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    submitted_at TIMESTAMP WITH TIME ZONE,
    approved_at TIMESTAMP WITH TIME ZONE
);

-- Improvement Request Comments
CREATE TABLE IF NOT EXISTS improvement_request_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL REFERENCES improvement_requests(id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES employees(id),
    content TEXT NOT NULL,
    is_internal BOOLEAN DEFAULT FALSE,  -- Internal notes for reviewers
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Improvement Request Approvals (tracking workflow)
CREATE TABLE IF NOT EXISTS improvement_request_approvals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL REFERENCES improvement_requests(id) ON DELETE CASCADE,
    approver_id UUID NOT NULL REFERENCES employees(id),
    stage VARCHAR(50) NOT NULL,  -- screening, evaluation, manager_approval, committee_review, budgeting
    decision VARCHAR(20) NOT NULL,  -- approved, rejected, pending
    comment TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Improvement Request Activity Log
CREATE TABLE IF NOT EXISTS improvement_request_activity (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL REFERENCES improvement_requests(id) ON DELETE CASCADE,
    actor_id UUID REFERENCES employees(id),
    action VARCHAR(100) NOT NULL,  -- created, submitted, status_changed, comment_added, approved, rejected, etc.
    old_value TEXT,
    new_value TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Improvement Request Attachments
CREATE TABLE IF NOT EXISTS improvement_request_attachments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL REFERENCES improvement_requests(id) ON DELETE CASCADE,
    uploaded_by UUID NOT NULL REFERENCES employees(id),
    file_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_size INTEGER,
    mime_type VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_improvement_requests_initiator ON improvement_requests(initiator_id);
CREATE INDEX IF NOT EXISTS idx_improvement_requests_sponsor ON improvement_requests(sponsor_id);
CREATE INDEX IF NOT EXISTS idx_improvement_requests_status ON improvement_requests(status);
CREATE INDEX IF NOT EXISTS idx_improvement_requests_department ON improvement_requests(department_id);
CREATE INDEX IF NOT EXISTS idx_improvement_requests_project ON improvement_requests(project_id);
CREATE INDEX IF NOT EXISTS idx_improvement_requests_created_at ON improvement_requests(created_at);
CREATE INDEX IF NOT EXISTS idx_improvement_request_comments_request ON improvement_request_comments(request_id);
CREATE INDEX IF NOT EXISTS idx_improvement_request_approvals_request ON improvement_request_approvals(request_id);
CREATE INDEX IF NOT EXISTS idx_improvement_request_activity_request ON improvement_request_activity(request_id);
CREATE INDEX IF NOT EXISTS idx_improvement_request_attachments_request ON improvement_request_attachments(request_id);

-- Insert default types
INSERT INTO improvement_request_types (name, description, icon, color)
VALUES
    ('Процесс', 'Улучшение бизнес-процессов', 'workflow', 'blue'),
    ('Продукт', 'Новые функции или улучшения продукта', 'box', 'green'),
    ('Инфраструктура', 'Улучшения ИТ-инфраструктуры', 'server', 'purple'),
    ('Автоматизация', 'Автоматизация ручных задач', 'zap', 'orange'),
    ('Оптимизация', 'Оптимизация затрат и ресурсов', 'trending-down', 'teal'),
    ('Другое', 'Прочие улучшения', 'lightbulb', 'gray')
ON CONFLICT DO NOTHING;
