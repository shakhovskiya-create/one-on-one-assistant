# FIGMA PROMPT SPEC

> Этот файл — ЕДИНЫЙ источник задания для Figma Chat.
> Его отправляет в Figma Chat только Product/Owner (Антон) после проверки v0.x → v1.0.

## 0) Meta
- Project: IKF Hub (employee hub)
- Platform: Web (desktop-first)
- Language: RU
- Design style: корпоративный, нейтральный, без “креативной самодеятельности”
- Source of truth (screens): макеты 01–06 (если применимо)
- Global Top Navigation (fixed): Главная / Сотрудники / Задачи / Встречи / Почта / Сообщения / SD / Аналитика
- Context Sidebar: строго контекстный (не дублировать top nav)

## 1) Goal
Опиши одним абзацем: что создаём (какой экран/поток) и зачем (польза/результат).

## 2) Users & Roles
- Employee:
- Manager:
- Admin (если есть):
Опиши ключевые права (read/write) и ограничения.

## 3) Screens / Frames to generate
Перечисли фреймы, которые должны быть созданы (с точными названиями), например:
1) <Screen Name> — Default
2) <Screen Name> — Empty state
3) <Screen Name> — Loading
4) <Screen Name> — Error
5) <Screen Name> — Read-only

## 4) Information Architecture (IA)
- Какие разделы/подразделы видны пользователю
- Что находится в top bar
- Что находится в left sidebar (контекстно)

## 5) Layout rules (hard)
- 12-column grid (desktop)
- Auto-layout enabled for main containers
- Components reuse where possible
- Primary actions: right/top (указать правило)
- No duplicated global items in sidebar
- No new global menu items (не изобретать)

## 6) Components (required)
Перечисли обязательные компоненты:
- Top Bar (global)
- Context Sidebar (per section)
- Header with title + actions
- Table/List with sorting/filtering (если нужно)
- Status badges
- Modal/Drawer (если нужно)
- Toast/Notification (если нужно)

## 7) Data & Content rules
- Никакого lorem ipsum
- Реалистичные подписи/поля на русском
- Даты/время: формат RU
- Примеры сущностей: задачи, заявки SD, встречи, участники и т.д.

## 8) States & Validation
- Empty / Loading / Error / Permission denied
- Form validation rules (если есть формы)
- Edge cases (минимум 3)

## 9) Accessibility (minimum)
- Focus states
- Contrast readable
- Keyboard navigation for forms/tables (базово)

## 10) Output expectation (Figma)
- Frames grouped by screen
- Naming: <Section>/<Screen>/<State>
- Auto-layout: ON
- Responsive constraints: set for desktop containers
- Components: reused (where possible)

## 11) Prompt to paste into Figma Chat (FINAL)
Скопируй отсюда в Figma Chat (без изменений) после утверждения:

PROMPT:
<PASTE YOUR FINAL FIGMA CHAT PROMPT HERE>
