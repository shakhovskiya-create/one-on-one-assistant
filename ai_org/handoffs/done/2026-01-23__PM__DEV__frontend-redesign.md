# HANDOFF: PM → DEV

## 1. Meta
- ID: 2026-01-23__PM__DEV__frontend-redesign
- Date created: 2026-01-23
- From role: PM
- To role: DEV
- Owner (responsible): DEV
- Status: Active
- Related initiative / epic: EKF Hub UI Redesign

---

## 2. Goal (ЗАЧЕМ)
Применить утверждённые HTML прототипы к реальному Svelte фронтенду:
- Единая структура: Global Top Nav + Dark Sidebar + Content
- Соответствие дизайн-системе EKF Hub

---

## 3. Context (ЧТО НУЖНО ЗНАТЬ)
- Референс: `ai_org/deliverables/prototypes/01-tasks.html`
- Референс: `ai_org/deliverables/prototypes/04-meetings.html`
- Дизайн-система: ekf-red #E53935, ekf-dark #1a1a2e
- Текущий фронтенд: Svelte 5, SvelteKit

---

## 4. Scope (ГРАНИЦЫ РАБОТ)
### In scope
- Создать GlobalNav.svelte - верхняя навигация модулей
- Обновить +layout.svelte - новая структура
- Обновить Sidebar.svelte - контекстная навигация
- Убрать Header.svelte (интегрировать в GlobalNav)

### Out of scope
- Отдельные страницы (tasks, meetings) - следующий этап

---

## 5. Requirements / Tasks
- [ ] Создать GlobalNav.svelte с модулями
- [ ] Обновить +layout.svelte под новую структуру
- [ ] Sidebar - убрать модули (оставить контекстную навигацию)
- [ ] Деплой на сервер

---

## 10. Result (ЗАПОЛНЯЕТ ИСПОЛНИТЕЛЬ)
(в процессе)

---

## Deliverables (обязательно)
- Primary deliverable: `frontend/src/lib/components/GlobalNav.svelte`
- Secondary: `frontend/src/routes/+layout.svelte`
