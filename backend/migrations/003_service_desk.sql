-- Service Desk Migration
-- Run this on PostgreSQL to create the Service Desk tables

-- Service Ticket Categories
CREATE TABLE IF NOT EXISTS service_ticket_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    icon VARCHAR(50),
    color VARCHAR(20),
    sla_hours INTEGER DEFAULT 24,
    parent_id UUID REFERENCES service_ticket_categories(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Service Tickets
CREATE TABLE IF NOT EXISTS service_tickets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    number VARCHAR(20) NOT NULL UNIQUE,
    type VARCHAR(50) NOT NULL DEFAULT 'service_request', -- incident, service_request, change, problem
    title VARCHAR(500) NOT NULL,
    description TEXT,
    category_id UUID REFERENCES service_ticket_categories(id),
    priority VARCHAR(20) NOT NULL DEFAULT 'medium', -- low, medium, high, critical
    impact VARCHAR(50) DEFAULT 'individual', -- individual, department, organization
    status VARCHAR(50) NOT NULL DEFAULT 'new', -- new, in_progress, pending, resolved, closed
    requester_id UUID NOT NULL REFERENCES employees(id),
    assignee_id UUID REFERENCES employees(id),
    sla_deadline TIMESTAMP WITH TIME ZONE,
    resolution TEXT,
    resolved_at TIMESTAMP WITH TIME ZONE,
    closed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Service Ticket Comments
CREATE TABLE IF NOT EXISTS service_ticket_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id UUID NOT NULL REFERENCES service_tickets(id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES employees(id),
    content TEXT NOT NULL,
    is_internal BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Service Ticket Activity Log
CREATE TABLE IF NOT EXISTS service_ticket_activity (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id UUID NOT NULL REFERENCES service_tickets(id) ON DELETE CASCADE,
    actor_id UUID REFERENCES employees(id),
    action VARCHAR(100) NOT NULL, -- created, status_changed, assigned, comment_added, resolved, etc.
    old_value TEXT,
    new_value TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_service_tickets_requester ON service_tickets(requester_id);
CREATE INDEX IF NOT EXISTS idx_service_tickets_assignee ON service_tickets(assignee_id);
CREATE INDEX IF NOT EXISTS idx_service_tickets_status ON service_tickets(status);
CREATE INDEX IF NOT EXISTS idx_service_tickets_type ON service_tickets(type);
CREATE INDEX IF NOT EXISTS idx_service_tickets_priority ON service_tickets(priority);
CREATE INDEX IF NOT EXISTS idx_service_tickets_created_at ON service_tickets(created_at);
CREATE INDEX IF NOT EXISTS idx_service_ticket_comments_ticket ON service_ticket_comments(ticket_id);
CREATE INDEX IF NOT EXISTS idx_service_ticket_activity_ticket ON service_ticket_activity(ticket_id);

-- Insert default categories
INSERT INTO service_ticket_categories (name, description, icon, color, sla_hours)
VALUES
    ('Оборудование', 'Ноутбуки, мониторы, периферия', 'laptop', 'red', 24),
    ('Программное обеспечение', 'Лицензии, установка, обновления', 'code', 'purple', 24),
    ('Доступы', 'Системы, VPN, папки, учётки', 'key', 'green', 8),
    ('Сеть и VPN', 'Подключение, WiFi, удалённый доступ', 'globe', 'orange', 8),
    ('Почта и Календарь', 'Outlook, общие ящики, календари', 'mail', 'blue', 24),
    ('HR сервисы', 'Приём, увольнение, справки', 'users', 'pink', 48),
    ('Хозяйственные', 'Офис, переговорные, ключи', 'building', 'teal', 48),
    ('Другое', 'Общие вопросы и консультации', 'help', 'gray', 24)
ON CONFLICT DO NOTHING;
