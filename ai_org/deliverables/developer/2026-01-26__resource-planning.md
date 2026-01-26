# GAP-006: Resource Planning (Планирование ресурсов)

## Summary
Реализовано управление ресурсами и загрузкой сотрудников с расчётом утилизации.

## Backend

### New Files
- `backend/internal/handlers/resources.go` - Handlers для Resource Planning API
- `backend/migrations/005_resource_planning.sql` - Миграция БД

### Modified Files
- `backend/internal/models/models.go` - Добавлены модели и расширен Employee
- `backend/cmd/server/main.go` - Добавлены routes

### Models
```go
// Employee extensions
type Employee struct {
    // ... existing fields
    WorkHoursPerWeek    *int     // Норма часов в неделю (40)
    AvailabilityPercent *int     // Доступность (0-100%)
}

type ResourceAllocation struct {
    ID                    string
    EmployeeID            string
    TaskID                *string
    ProjectID             *string
    Role                  *string
    AllocatedHoursPerWeek int
    PeriodStart           string
    PeriodEnd             *string
    Notes                 *string
    ...
}

type EmployeeAbsence struct {
    ID          string
    EmployeeID  string
    AbsenceType string  // vacation, sick_leave, holiday, out_of_office
    StartDate   string
    EndDate     string
    Source      string  // manual, exchange, hr_system
    ...
}

type ResourceCapacity struct {
    EmployeeID         string
    EmployeeName       string
    Position           string
    WeeklyHours        int      // work_hours_per_week
    AvailabilityPct    int      // availability_percent
    AvailableHours     float64  // weekly_hours * availability_pct / 100
    AllocatedHours     float64  // sum of allocations
    FreeHours          float64  // available - allocated
    UtilizationPercent float64  // allocated / available * 100
    Overloaded         bool     // utilization > 100
}

type ResourceAllocationStats struct {
    TotalEmployees    int
    TotalAllocations  int
    OverloadedCount   int
    UnderutilizedCnt  int
    AvgUtilization    float64
}
```

### API Endpoints
| Method | Path | Description |
|--------|------|-------------|
| GET | /resources/allocations | Список аллокаций |
| GET | /resources/allocations/:id | Детали аллокации |
| POST | /resources/allocations | Создать аллокацию |
| PUT | /resources/allocations/:id | Обновить аллокацию |
| DELETE | /resources/allocations/:id | Удалить аллокацию |
| GET | /resources/capacity | Загрузка сотрудников |
| GET | /resources/stats | Статистика |
| GET | /resources/absences | Список отсутствий |
| POST | /resources/absences | Создать отсутствие |
| DELETE | /resources/absences/:id | Удалить отсутствие |
| PUT | /employees/:id/resource-settings | Обновить настройки ресурсов |

## Frontend

### New Files
- `frontend/src/routes/resources/+page.svelte` - Страница планирования ресурсов

### Modified Files
- `frontend/src/lib/api/client.ts` - API функции и типы

### Pages

**Resources (`/resources`)**
- Статистика (сотрудников, аллокаций, перегружено, недозагружено, ср. загрузка)
- Фильтр по проекту
- Таблица загрузки сотрудников:
  - Имя, должность
  - Норма ч/нед, доступность
  - Доступно, выделено, свободно
  - Прогресс-бар загрузки с цветовой индикацией
- Список текущих аллокаций
- Модальная форма создания аллокации

## Database Schema

```sql
-- Extend employees
ALTER TABLE employees
ADD COLUMN work_hours_per_week INTEGER DEFAULT 40,
ADD COLUMN availability_percent INTEGER DEFAULT 100,
ADD COLUMN hourly_rate DECIMAL(10,2);

-- Resource Allocation
CREATE TABLE resource_allocations (
    id UUID PRIMARY KEY,
    employee_id UUID REFERENCES employees,
    task_id UUID REFERENCES tasks,
    project_id UUID REFERENCES projects,
    role VARCHAR(100),
    allocated_hours_per_week INTEGER,
    period_start DATE NOT NULL,
    period_end DATE,
    notes TEXT,
    created_at TIMESTAMP,
    created_by UUID
);

-- Employee Absences
CREATE TABLE employee_absences (
    id UUID PRIMARY KEY,
    employee_id UUID REFERENCES employees,
    absence_type VARCHAR(50),  -- vacation, sick_leave, holiday, out_of_office
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    description TEXT,
    source VARCHAR(50) DEFAULT 'manual'
);
```

## Metrics

### Capacity Calculation
```
available_hours = work_hours_per_week × (availability_percent / 100)
allocated_hours = SUM(resource_allocations.allocated_hours_per_week)
free_hours = available_hours - allocated_hours
utilization_percent = (allocated_hours / available_hours) × 100
overloaded = utilization_percent > 100
```

### Color Coding
- > 100%: Red (перегружен)
- > 80%: Yellow (высокая загрузка)
- > 50%: Green (нормальная)
- < 50%: Gray (недозагружен)

## Deployment

### Migration
```bash
cat backend/migrations/005_resource_planning.sql | docker exec -i oneonone-postgres psql -U postgres -d oneonone
```

### Rebuild
```bash
docker-compose down && docker-compose build --no-cache backend frontend && docker-compose up -d
```

## Design System
- EKF Red (#E53935) - primary actions
- Progress bar visualization for utilization
- Svelte 5 runes ($state, $derived, $effect)
- Tailwind CSS

## Related
- Functional Model: `ai_org/deliverables/analyst/2026-01-26__critical-gaps__fm.md` (Section 4)
- Sprint: Sprint 9 CRITICAL
