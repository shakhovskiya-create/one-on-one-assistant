# Deliverable: Sprint Board Page

## Meta
- Date: 2026-01-23
- Role: DEV
- Handoff: `ai_org/handoffs/done/2026-01-23__PM__DEV__sprint_board.md`

## Описание
Создана страница управления спринтами `/sprints` с полным CRUD функционалом.

## Файлы

### Созданные
- `frontend/src/routes/sprints/+page.svelte` - страница управления спринтами

### Изменённые
- `frontend/src/lib/components/Sidebar.svelte` - добавлены ссылки Sprints и Releases, иконка sprint

## Функционал

### Sprint Board Page
- Список всех спринтов с карточками
- Создание нового спринта (название, цель, даты, проект)
- Редактирование спринта (inline form)
- Удаление спринта с подтверждением
- Запуск спринта (переход в статус active)
- Завершение спринта (подсчёт velocity)
- Отображение прогресса (progress bar)
- Статистика: tasks count, done count, total points, completed points, velocity

### Sidebar Navigation
- Ссылка "Спринты" с иконкой молнии (sprint)
- Ссылка "Релизы" с иконкой тега (release)

## API Integration
- `sprints.list()` - получение списка спринтов
- `sprints.create()` - создание спринта
- `sprints.update()` - обновление спринта
- `sprints.delete()` - удаление спринта
- `sprints.start()` - запуск спринта
- `sprints.complete()` - завершение спринта

## Технологии
- Svelte 5 runes ($state, $derived)
- Tailwind CSS
- TypeScript
