# Frontend Redesign: Unified Navigation

## Summary
Применены утверждённые HTML прототипы к реальному Svelte фронтенду.

## Changes Made

### New Files
- `frontend/src/lib/components/GlobalNav.svelte` - Глобальная верхняя навигация

### Modified Files
- `frontend/src/routes/+layout.svelte` - Новая структура layout
- `frontend/src/lib/components/Sidebar.svelte` - Контекстная навигация

## Architecture

```
┌─────────────────────────────────────────────────────┐
│ GlobalNav (h-12, bg-ekf-dark, fixed top)            │
│ [EKF Hub] Дашборд | Задачи | Встречи | ...          │
├────────────┬────────────────────────────────────────┤
│ Sidebar    │                                        │
│ (w-60)     │     Main Content                       │
│ contextual │     (slot)                             │
│            │                                        │
└────────────┴────────────────────────────────────────┘
```

## Sidebar Modes

| Path Pattern | Sidebar Title | Items |
|--------------|---------------|-------|
| /tasks, /projects, /sprints, /releases | Управление проектом | Доска, Проекты, Спринты, Релизы |
| /meetings, /calendar | Встречи | Календарь, Расписание |
| default | Инструменты | Сообщения, Почта, Confluence, GitHub, Настройки |

## Design System
- ekf-red: #E53935
- ekf-dark: #1a1a2e
- Font: system (inherited)
- Tailwind CSS classes

## Related
- Prototypes: `ai_org/deliverables/prototypes/`
- Handoff: `ai_org/handoffs/active/2026-01-23__PM__DEV__frontend-redesign.md`
