# Status (оперативный)

**Дата:** 2026-01-26
**Обновлено:** 11:00 UTC

## Текущий фокус
- ✅ Sprint 9 CRITICAL в процессе
- ✅ GAP-001: Global Navigation исправлен
- ✅ GAP-002: GitHub добавлен в Tasks sidebar
- ✅ GAP-009: Tasks sidebar полностью по макету
- ✅ GAP-012: Профиль исправлен (полное ФИО)
- ✅ GAP-010: Service Desk MVP — DEPLOYED
- ✅ GAP-005: Заявка на улучшение — DEPLOYED
- 🔄 GAP-006: Планирование ресурсов — В РЕАЛИЗАЦИИ
- 🔄 Enterprise Full Audit — В ПРОГРЕССЕ (PM→Analyst)

## Завершено сегодня

### GAP-006: Планирование ресурсов 🔄 IN PROGRESS
**Backend:**
- Модели: ResourceAllocation, EmployeeAbsence, ResourceCapacity, ResourceStats
- Расширение Employee: work_hours_per_week, availability_percent
- Handlers: CRUD для allocations, absences, capacity, stats
- Routes: /api/v1/resources/*
- Миграция: `backend/migrations/005_resource_planning.sql`

**Frontend:**
- `/resources` — Capacity overview с таблицей загрузки
- Модальная форма создания аллокации
- API client: resources functions

**Метрики:**
- utilization_percent — загрузка сотрудника
- available_hours — доступные часы
- overloaded — флаг перегрузки (>100%)

**Файлы:**
- `backend/internal/handlers/resources.go`
- `backend/internal/models/models.go` (добавлены модели)
- `backend/cmd/server/main.go` (добавлены routes)
- `frontend/src/lib/api/client.ts` (добавлен API)
- `frontend/src/routes/resources/+page.svelte`

### GAP-005: Заявка на улучшение ✅ DEPLOYED
**Backend:**
- Модели: ImprovementRequest, ImprovementRequestType, ImprovementRequestComment, ImprovementRequestApproval, ImprovementRequestActivity
- Handlers: CRUD, submit, approve, reject, create-project
- Routes: /api/v1/improvements/*
- Миграция: `backend/migrations/004_improvement_requests.sql`

**Frontend:**
- `/improvements` — Список заявок с фильтрами
- `/improvements/create` — Форма создания заявки
- `/improvements/[id]` — Детали заявки с workflow

**Workflow (9 статусов):**
draft → submitted → screening → evaluation → manager_approval → committee_review → budgeting → project_created → in_progress/completed

**Коммит:** c28671d

### GAP-010: Service Desk MVP ✅ DEPLOYED
**Backend:**
- Модели: ServiceTicket, ServiceTicketCategory, ServiceTicketComment, ServiceTicketActivity
- Handlers: CRUD для tickets, comments, categories, stats
- Routes: /api/v1/service-desk/*
- Миграция: `backend/migrations/003_service_desk.sql`

**Frontend:**
- `/service-desk` — User Portal (hero, my tickets, catalog)
- `/service-desk/create` — Create ticket form
- `/service-desk/tickets/[id]` — Ticket detail view

**Коммит:** f8be668, 3dcd4b2

### GAP-001: Global Navigation ✅ DONE
- GlobalNav.svelte: 8 модулей (утверждённый состав)
- Sidebar.svelte: контекстная навигация по разделам
- **Коммит:** f39c006

### GAP-009: Tasks Sidebar ✅ DONE
- Project Selector
- Планирование: Доска, Бэклог, Roadmap
- Спринты: динамическая загрузка
- Релизы: динамическая загрузка
- Тестирование, Документация, Интеграции
- **Коммит:** c6da16c

### GAP-008: Meetings Sidebar ✅ DONE
- Sidebar секции: Календарь, Встречи, Записи
- Main content: требует calendar view (MEDIUM priority)

### GAP-012: Профиль ✅ DONE
- Отображается полное ФИО
- Убрано дублирование в sidebar

## Диагностика GAP-007 (зависимости)
**Статус:** 🔍 DIAGNOSED — требует проверки с тестовыми данными
- База данных: 0 зависимостей
- Код: правильная обработка ошибок
- Рекомендация: создать тестовые данные

## Следующие шаги

### CRITICAL (осталось):
- ✅ Все CRITICAL GAPs закрыты!

### HIGH:
- GAP-003: Связь задача→проект (требует миграцию БД)

### MEDIUM:
- GAP-008: Meetings main content — calendar view (не sidebar)

## Для деплоя GAP-006
```bash
# На сервере 10.100.0.131
cd /opt/one-on-one/app

# 1. Pull последние изменения
git pull origin main

# 2. Применить миграцию
cat backend/migrations/005_resource_planning.sql | docker exec -i oneonone-postgres psql -U postgres -d oneonone

# 3. Пересобрать и запустить
docker-compose down && docker-compose build --no-cache backend frontend && docker-compose up -d
```

## Артефакты
- Handoff: `ai_org/handoffs/active/2026-01-26__PM__ANALYST__critical-gaps-analysis.md`
- Спецификация: `ai_org/deliverables/analyst/2026-01-26__critical-gaps__spec.md`
- Функциональная модель: `ai_org/deliverables/analyst/2026-01-26__critical-gaps__fm.md`
- Deliverable GAP-005: `ai_org/deliverables/developer/2026-01-26__improvement-requests.md`
