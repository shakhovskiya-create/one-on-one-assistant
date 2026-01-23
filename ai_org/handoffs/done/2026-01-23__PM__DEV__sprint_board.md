# HANDOFF: PM → DEV

## 1. Meta
- ID: 2026-01-23__PM__DEV__sprint_board
- Date created: 2026-01-23
- From role: PM
- To role: DEV
- Owner (responsible): DEV
- Status: Completed
- Related initiative / epic: JIRA-like functionality

---

## 2. Goal (ЗАЧЕМ)
Создать страницу управления спринтами с полным CRUD функционалом, аналогичную JIRA Sprint Board.

---

## 3. Context (ЧТО НУЖНО ЗНАТЬ)
- `/ai_org/state/CONTEXT.md` — контекст проекта
- `/ai_org/state/ARCHITECTURE.md` — SvelteKit + Go Fiber архитектура
- `/ai_org/state/WORKPLAN.md` — Sprint 8 JIRA parity
- Backend sprints handlers уже созданы: `backend/internal/handlers/sprints.go`
- API endpoints уже есть: `/api/sprints/*`

---

## 4. Scope (ГРАНИЦЫ РАБОТ)
### In scope
- Создать страницу `/sprints` для управления спринтами
- Добавить навигацию в sidebar
- Добавить иконку sprint

### Out of scope
- Burndown chart
- Sprint planning board (drag-and-drop)

---

## 5. Requirements / Tasks
- [x] Создать `frontend/src/routes/sprints/+page.svelte`
- [x] Добавить создание, редактирование, удаление спринтов
- [x] Добавить действия start/complete спринта
- [x] Отображать прогресс, velocity, tasks count
- [x] Добавить ссылки Sprints и Releases в Sidebar
- [x] Добавить иконку sprint в Sidebar

---

## 6. Acceptance Criteria (КРИТЕРИИ ПРИЁМКИ)
- [x] Страница /sprints доступна
- [x] Можно создать новый спринт
- [x] Можно редактировать и удалять спринт
- [x] Можно запустить и завершить спринт
- [x] Отображается прогресс и velocity
- [x] Навигация работает из sidebar

---

## 7. Constraints & Invariants (ОГРАНИЧЕНИЯ)
- Svelte 5 runes ($state, $derived)
- Tailwind CSS для стилей
- API через client.ts

---

## 8. Risks, Assumptions, Open Questions
### Risks
- Нет

### Assumptions
- Backend API работает корректно

### Open questions
- Нет

---

## 9. Expected Outputs (АРТЕФАКТЫ НА ВЫХОДЕ)
- `frontend/src/routes/sprints/+page.svelte` - страница спринтов
- `frontend/src/lib/components/Sidebar.svelte` - обновлён с новыми ссылками и иконкой

---

## 10. Result (ЗАПОЛНЯЕТ ИСПОЛНИТЕЛЬ)
Создана полнофункциональная страница управления спринтами:
- CRUD операции для спринтов
- Start/Complete actions с velocity calculation
- Progress bar и статистика
- Навигация добавлена в sidebar с иконками

Дата выполнения: 2026-01-23

---

## 11. Acceptance (ЗАПОЛНЯЕТ PM)
- Status: Accepted
- Accepted by: PM
- Date: 2026-01-23
- Notes:
  - Функционал соответствует требованиям

---

## 12. Follow-ups / Next steps
- [ ] Burndown chart (отложено)
- [ ] Drag-and-drop sprint planning (отложено)

---

## 13. Log reference
- `ai_org/logs/AGENT_LOG.md`: Sprint Board section
