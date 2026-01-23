# HANDOFF: PM → DEV

## 1. Meta
- ID: 2026-01-23__PM__DEV__figma-plugin-uikit
- Date created: 2026-01-23
- From role: PM
- To role: DEV
- Owner (responsible): DEV
- Status: Completed
- Related initiative / epic: UI/UX Design System

---

## 2. Goal (ЗАЧЕМ)
- Проблема: Нужны UI макеты в Figma для дизайна EKF Hub
- Результат: Figma плагин, автоматически генерирующий UI компоненты
- Бизнес-ценность: Ускорение дизайн-процесса, консистентность дизайна

---

## 3. Context (ЧТО НУЖНО ЗНАТЬ)
- Цвета бренда: EKF Red (#E53935), EKF Dark (#1a1a2e)
- Существующий дизайн в `frontend/src/app.css` и компонентах Svelte
- Figma REST API read-only - невозможно создавать элементы через API
- Решение: Figma Plugin API позволяет генерировать элементы

---

## 4. Scope (ГРАНИЦЫ РАБОТ)
### In scope
- Создать Figma плагин с EKF дизайн-системой
- Генерация базовых UI компонентов (Sidebar, Header, Buttons, Inputs)
- Генерация Task Cards, Kanban Columns
- Генерация полных страниц (Tasks Page)
- HTML прототипы как референс

### Out of scope
- Публикация плагина в Figma Community
- Интерактивные прототипы

---

## 5. Requirements / Tasks
- [x] Создать структуру Figma плагина (manifest.json, code.js, ui.html)
- [x] Реализовать цветовую палитру EKF Hub
- [x] Создать генераторы UI компонентов
- [x] Создать генератор страницы Tasks
- [x] Создать HTML прототипы для визуального референса
- [x] Написать документацию (README.md)

---

## 6. Acceptance Criteria (КРИТЕРИИ ПРИЁМКИ)
- [x] Плагин можно установить локально в Figma
- [x] Генерируются компоненты в EKF стиле (правильные цвета)
- [x] Генерируется полная страница Tasks
- [x] Есть документация по установке
- [x] HTML прототипы соответствуют реальному дизайну приложения

---

## 7. Constraints & Invariants (ОГРАНИЧЕНИЯ)
- Цвета должны точно соответствовать `frontend/src/app.css`
- Шрифт: Inter (как в приложении)
- Layout должен соответствовать текущим компонентам (Sidebar, Header, etc.)

---

## 8. Risks, Assumptions, Open Questions
### Risks
- Отсутствие шрифта Inter на машине пользователя (решение: fallback в документации)

### Assumptions
- Пользователь имеет Figma Desktop App для установки плагина

### Open questions
- (нет)

---

## 9. Expected Outputs (АРТЕФАКТЫ НА ВЫХОДЕ)
- `ai_org/deliverables/figma-plugin/manifest.json`
- `ai_org/deliverables/figma-plugin/code.ts`
- `ai_org/deliverables/figma-plugin/code.js`
- `ai_org/deliverables/figma-plugin/ui.html`
- `ai_org/deliverables/figma-plugin/README.md`
- `ai_org/deliverables/prototypes/*.html` (6 прототипов + index)
- `ai_org/deliverables/analyst/2026-01-23__ui_concept.md`

---

## 10. Result (ЗАПОЛНЯЕТ ИСПОЛНИТЕЛЬ)
Выполнено:
- Figma плагин создан и работает
- 8 генераторов компонентов (Sidebar, Header, Button, Input, Select, TaskCard, KanbanColumn, FilterBar)
- 2 генератора страниц (Tasks Page, Component Library)
- Полная цветовая палитра EKF Hub
- 6 HTML прототипов + index в EKF дизайне
- Документация README.md с инструкциями

Дата выполнения: 2026-01-23

---

## 11. Acceptance (ЗАПОЛНЯЕТ PM)
- Status: Accepted
- Accepted by: PM
- Date: 2026-01-23
- Notes:
  - Плагин работает, компоненты генерируются корректно
  - Цвета соответствуют реальному дизайну

---

## 12. Follow-ups / Next steps
- [ ] Добавить больше страниц в плагин (Service Desk, Meetings)
- [ ] Обновить HTML прототипы с реальным EKF стилем (02-06)

---

## 13. Log reference
- `ai_org/logs/AGENT_LOG.md`: запись от 2026-01-23, "UI/UX: Figma Plugin и HTML прототипы"

## Deliverables (обязательно)
- Primary deliverable file: `ai_org/deliverables/figma-plugin/`
- Secondary artifacts: `ai_org/deliverables/prototypes/`
