# GAP-010: Service Desk MVP

## Summary
Реализован полноценный Service Desk (ITSM) портал для обработки обращений пользователей.
Следует макетам `02-service-desk-portal.html` и `03-service-desk-agent.html`.

## Backend

### New Files
- `backend/internal/handlers/servicedesk.go` - Handlers для Service Desk API
- `backend/migrations/003_service_desk.sql` - Миграция БД

### Modified Files
- `backend/internal/models/models.go` - Добавлены модели Service Desk
- `backend/cmd/server/main.go` - Добавлены routes

### Models
```go
// ServiceTicket - основная сущность тикета
type ServiceTicket struct {
    ID           string
    Number       string  // INC-2026-0001, REQ-2026-0001
    Type         string  // incident, service_request, change, problem
    Title        string
    Description  *string
    Priority     string  // low, medium, high, critical
    Impact       string  // individual, department, organization
    Status       string  // new, in_progress, pending, resolved, closed
    SLADeadline  *time.Time
    ...
}

// ServiceTicketCategory - категории услуг
// ServiceTicketComment - комментарии к тикетам
// ServiceTicketActivity - журнал активности
```

### API Endpoints
| Method | Path | Description |
|--------|------|-------------|
| GET | /service-desk/tickets | Список тикетов (с фильтрами) |
| GET | /service-desk/tickets/my | Мои тикеты |
| GET | /service-desk/tickets/:id | Детали тикета |
| POST | /service-desk/tickets | Создать тикет |
| PUT | /service-desk/tickets/:id | Обновить тикет |
| POST | /service-desk/tickets/:id/comments | Добавить комментарий |
| GET | /service-desk/categories | Список категорий |
| GET | /service-desk/stats | Статистика |

## Frontend

### New Files
- `frontend/src/routes/service-desk/+page.svelte` - User Portal
- `frontend/src/routes/service-desk/create/+page.svelte` - Форма создания
- `frontend/src/routes/service-desk/tickets/[id]/+page.svelte` - Детали тикета

### Modified Files
- `frontend/src/lib/api/client.ts` - API функции и типы

### Pages

**User Portal (`/service-desk`)**
- Hero секция с CTA
- Блок "Мои заявки" (последние 5)
- Каталог услуг (categories)
- Quick actions (создать инцидент/запрос)

**Create Ticket (`/service-desk/create`)**
- Выбор типа: инцидент, запрос на услугу, запрос на изменение
- Тема и описание
- Выбор приоритета с SLA информацией
- Масштаб влияния (для инцидентов)

**Ticket Detail (`/service-desk/tickets/[id]`)**
- Полная информация о тикете
- История активности
- Комментарии (публичные и internal)
- Добавление комментария

## Database Schema

```sql
-- Категории услуг
CREATE TABLE service_ticket_categories (
    id UUID PRIMARY KEY,
    name VARCHAR(100),
    description TEXT,
    icon VARCHAR(50),
    color VARCHAR(20),
    sla_hours INTEGER DEFAULT 24
);

-- Тикеты
CREATE TABLE service_tickets (
    id UUID PRIMARY KEY,
    number VARCHAR(20) UNIQUE,  -- INC-2026-0001
    type VARCHAR(50),           -- incident, service_request, change, problem
    title VARCHAR(500),
    description TEXT,
    category_id UUID REFERENCES service_ticket_categories,
    priority VARCHAR(20),       -- low, medium, high, critical
    impact VARCHAR(50),         -- individual, department, organization
    status VARCHAR(50),         -- new, in_progress, pending, resolved, closed
    requester_id UUID REFERENCES employees,
    assignee_id UUID REFERENCES employees,
    sla_deadline TIMESTAMP WITH TIME ZONE,
    resolution TEXT,
    resolved_at TIMESTAMP WITH TIME ZONE,
    closed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE
);

-- Комментарии
CREATE TABLE service_ticket_comments (
    id UUID PRIMARY KEY,
    ticket_id UUID REFERENCES service_tickets,
    author_id UUID REFERENCES employees,
    content TEXT,
    is_internal BOOLEAN DEFAULT FALSE
);

-- Журнал активности
CREATE TABLE service_ticket_activity (
    id UUID PRIMARY KEY,
    ticket_id UUID REFERENCES service_tickets,
    actor_id UUID REFERENCES employees,
    action VARCHAR(100),
    old_value TEXT,
    new_value TEXT
);
```

## ITIL Features

### Ticket Types
| Type | Prefix | Description |
|------|--------|-------------|
| incident | INC | Что-то не работает |
| service_request | REQ | Запрос на услугу |
| change | CHG | Запрос на изменение |
| problem | PRB | Корневая причина |

### SLA by Priority
| Priority | SLA Hours | Use Case |
|----------|-----------|----------|
| critical | 4h | Массовый сбой |
| high | 8h | Блокирующая проблема |
| medium | 24h | Стандартный запрос |
| low | 72h | Пожелание |

### Default Categories
1. Оборудование (24h SLA)
2. Программное обеспечение (24h SLA)
3. Доступы (8h SLA)
4. Сеть и VPN (8h SLA)
5. Почта и Календарь (24h SLA)
6. HR сервисы (48h SLA)
7. Хозяйственные (48h SLA)
8. Другое (24h SLA)

## Deployment

### Migration
```bash
psql -U postgres -d oneonondb -f backend/migrations/003_service_desk.sql
```

### Rebuild
```bash
docker-compose down && docker-compose build && docker-compose up -d
```

## Design System
- EKF Red (#E53935) - primary actions
- EKF Dark (#1a1a2e) - sidebar
- Status colors: green (new), blue (in_progress), yellow (pending), purple (resolved)
- Svelte 5 runes ($state, $derived)
- Tailwind CSS

## Bugfixes

### Svelte 5 Event Syntax
- **Problem:** Mixed Svelte 4 (`on:submit`) and Svelte 5 (`onclick`) event handlers
- **File:** `frontend/src/routes/service-desk/create/+page.svelte`
- **Fix:** Changed `on:submit|preventDefault={handleSubmit}` to `onsubmit={(e) => { e.preventDefault(); handleSubmit(); }}`

## Related
- Prototypes: `ai_org/deliverables/prototypes/02-service-desk-portal.html`
- Prototypes: `ai_org/deliverables/prototypes/03-service-desk-agent.html`
- Handoff: `ai_org/handoffs/active/2026-01-26__PM__ANALYST__critical-gaps-analysis.md`
- Sprint: Sprint 9 CRITICAL
