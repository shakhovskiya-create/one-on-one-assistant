# Handoff: Tasks Project Sidebar

**Дата:** 2026-01-26
**От:** PM
**Кому:** Developer
**Приоритет:** HIGH

## Задача

Добавить Project Management Sidebar на страницу Задачи (/tasks) по макету 01-tasks.html.

## Выполнено

### 1. Project Sidebar (левый, 240px, ekf-dark)

- Project Selector (EKF Hub)
- Планирование: Доска задач (active), Бэклог, Roadmap
- Спринты: активный, планируемые, архив
- Релизы: unreleased (Dev), released
- Тестирование: Тест-планы, Тест-кейсы, Прогоны
- Документация: Wiki, Требования
- Настройки проекта
- User profile в нижней части

### 2. Sprint Header

- Динамический заголовок с названием активного спринта
- Прогресс-бар с процентом выполнения
- Даты спринта

### 3. Улучшенные фильтры

- Поиск задач
- Фильтр по исполнителям
- Фильтр по статусам
- Фильтр по проектам
- View Toggle (Список/Kanban)

## Файлы

- `frontend/src/routes/tasks/+page.svelte` (MODIFIED)

## Acceptance Criteria

- [x] Sidebar отображается слева
- [x] Навигация по секциям (Планирование, Спринты, Релизы, Тестирование, Документация)
- [x] Sprint header с прогрессом
- [x] Фильтрация по спринту из sidebar
- [x] User profile в sidebar

## Статус

✅ DONE
