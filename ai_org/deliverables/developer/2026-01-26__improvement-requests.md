# GAP-005: Improvement Requests (Заявки на улучшение)

## Summary
Реализован полный workflow управления заявками на улучшение с 9-этапным процессом согласования.

## Backend

### New Files
- `backend/internal/handlers/improvements.go` - Handlers для Improvement Requests API
- `backend/migrations/004_improvement_requests.sql` - Миграция БД

### Modified Files
- `backend/internal/models/models.go` - Добавлены модели Improvement Requests
- `backend/cmd/server/main.go` - Добавлены routes

### Models
```go
type ImprovementRequest struct {
    ID                string
    Number            string     // IR-2026-0001
    Title             string
    Description       *string
    BusinessValue     *string    // Бизнес-ценность
    ExpectedEffect    *string    // Ожидаемый эффект (KPI / финансовый)
    InitiatorID       string
    DepartmentID      *string
    SponsorID         *string
    EstimatedBudget   *float64   // Предварительный бюджет
    ApprovedBudget    *float64   // Утверждённый бюджет
    EstimatedStart    *string
    EstimatedEnd      *string
    Status            string     // 9-stage lifecycle
    CommitteeDate     *string
    CommitteeDecision *string
    ProjectID         *string    // Link to created project
    RejectionReason   *string
    TypeID            *string
    Priority          string
    ...
}

type ImprovementRequestType struct { ... }
type ImprovementRequestComment struct { ... }
type ImprovementRequestApproval struct { ... }
type ImprovementRequestActivity struct { ... }
```

### API Endpoints
| Method | Path | Description |
|--------|------|-------------|
| GET | /improvements | Список заявок (с фильтрами) |
| GET | /improvements/my | Мои заявки |
| GET | /improvements/types | Типы заявок |
| GET | /improvements/stats | Статистика |
| GET | /improvements/:id | Детали заявки |
| POST | /improvements | Создать заявку |
| PUT | /improvements/:id | Обновить заявку |
| POST | /improvements/:id/submit | Подать на рассмотрение |
| POST | /improvements/:id/approve | Одобрить |
| POST | /improvements/:id/reject | Отклонить |
| POST | /improvements/:id/create-project | Создать проект |
| POST | /improvements/:id/comments | Добавить комментарий |

## Frontend

### New Files
- `frontend/src/routes/improvements/+page.svelte` - Список заявок
- `frontend/src/routes/improvements/create/+page.svelte` - Форма создания
- `frontend/src/routes/improvements/[id]/+page.svelte` - Детали заявки

### Modified Files
- `frontend/src/lib/api/client.ts` - API функции и типы

### Pages

**List (`/improvements`)**
- Статистика (всего, черновики, на рассмотрении, одобрено, отклонено)
- Фильтры: только мои, по статусу, по типу
- Карточки заявок с основной информацией

**Create (`/improvements/create`)**
- Название и описание
- Бизнес-обоснование (бизнес-ценность, ожидаемый эффект)
- Тип улучшения, приоритет
- Предварительный бюджет
- Сроки реализации

**Detail (`/improvements/[id]`)**
- Визуализация workflow (8 шагов прогресс-бар)
- Действия: подать, одобрить, отклонить, создать проект
- Комментарии (публичные и внутренние)
- История активности

## Workflow

### 9-этапный процесс согласования
```
draft → submitted → screening → evaluation → manager_approval → committee_review → budgeting → project_created → in_progress/completed
  ↓         ↓           ↓            ↓              ↓                 ↓               ↓
       ←←←←←←←←←←←←←←←←← rejected ←←←←←←←←←←←←←←←←←←←←←←←←←←←←←←←←←←←
```

### Status Transitions
| # | Status | Description | Actor |
|---|--------|-------------|-------|
| 1 | draft | Черновик | Initiator |
| 2 | submitted | Подана | Initiator → click "Submit" |
| 3 | screening | Первичный скрининг | Process Owner / Business Curator |
| 4 | evaluation | Оценка | Analyst + IT + Finance |
| 5 | manager_approval | Согласование руководителя | Line Manager |
| 6 | committee_review | Рассмотрение комитетом | Committee |
| 7 | budgeting | Утверждение бюджета | Finance |
| 8 | project_created | Проект создан | PMO |
| 9 | in_progress / completed | Управление проектом | Project Manager |

## Database Schema

```sql
-- Types
CREATE TABLE improvement_request_types (
    id UUID PRIMARY KEY,
    name VARCHAR(100),
    description TEXT,
    icon VARCHAR(50),
    color VARCHAR(20)
);

-- Main requests
CREATE TABLE improvement_requests (
    id UUID PRIMARY KEY,
    number VARCHAR(20) UNIQUE,  -- IR-2026-0001
    title VARCHAR(500),
    description TEXT,
    business_value TEXT,
    expected_effect TEXT,
    initiator_id UUID REFERENCES employees,
    department_id UUID REFERENCES departments,
    sponsor_id UUID REFERENCES employees,
    estimated_budget DECIMAL(15,2),
    approved_budget DECIMAL(15,2),
    estimated_start DATE,
    estimated_end DATE,
    status VARCHAR(50) DEFAULT 'draft',
    committee_date DATE,
    committee_decision TEXT,
    project_id UUID REFERENCES projects,
    rejection_reason TEXT,
    rejected_by UUID REFERENCES employees,
    type_id UUID REFERENCES improvement_request_types,
    priority VARCHAR(20) DEFAULT 'medium',
    ...
);

-- Comments, Approvals, Activity tables
```

## Default Types
1. Процесс — Улучшение бизнес-процессов
2. Продукт — Новые функции или улучшения продукта
3. Инфраструктура — Улучшения ИТ-инфраструктуры
4. Автоматизация — Автоматизация ручных задач
5. Оптимизация — Оптимизация затрат и ресурсов
6. Другое — Прочие улучшения

## Project Integration

После статуса `project_created` (или `budgeting` при одобрении бюджета):
1. Создаётся проект в таблице `projects`
2. Название проекта = название заявки
3. Описание проекта = описание заявки
4. Сроки = сроки из заявки
5. `project_id` записывается в заявку
6. Статус заявки переходит в `in_progress`

## Deployment

### Migration
```bash
cat backend/migrations/004_improvement_requests.sql | docker exec -i oneonone-postgres psql -U postgres -d oneonone
```

### Rebuild
```bash
docker-compose down && docker-compose build --no-cache backend frontend && docker-compose up -d
```

## Design System
- EKF Red (#E53935) - primary actions
- Green (#10B981) - approved, success
- Red (#EF4444) - rejected, errors
- Progress bar visualization for workflow steps
- Svelte 5 runes ($state, $derived)
- Tailwind CSS

## Related
- Functional Model: `ai_org/deliverables/analyst/2026-01-26__critical-gaps__fm.md` (Section 2)
- Handoff: `ai_org/handoffs/active/2026-01-26__PM__ANALYST__critical-gaps-analysis.md`
- Sprint: Sprint 9 CRITICAL
