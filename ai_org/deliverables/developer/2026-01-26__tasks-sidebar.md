# Tasks Project Sidebar

**Дата:** 2026-01-26
**Статус:** ✅ DONE

## Summary

Добавлен полноценный Project Management Sidebar на страницу Задачи, соответствующий макету 01-tasks.html.

## Changes

### File: `frontend/src/routes/tasks/+page.svelte`

#### 1. Layout Structure

Новая структура страницы:
```
<div class="flex h-[calc(100vh-4rem)] -m-4">
  <aside class="w-60 bg-ekf-dark">  <!-- Project Sidebar -->
  <main class="flex-1 flex flex-col">  <!-- Main Content -->
```

#### 2. Project Sidebar Sections

- **Project Selector**: Выбор активного проекта (EKF Hub)
- **Планирование**: Доска задач, Бэклог, Roadmap
- **Спринты**: Динамический список спринтов с фильтрацией
- **Релизы**: Версии из API (unreleased/released)
- **Тестирование**: Тест-планы, Тест-кейсы, Прогоны
- **Документация**: Wiki, Требования
- **User Profile**: Фото и данные текущего пользователя

#### 3. Sprint Header

Динамический заголовок при выбранном спринте:
- Название спринта + бейдж "Активный"
- Даты (start_date — end_date)
- Прогресс-бар с процентом выполнения
- Кнопка "Новая задача"

#### 4. Sidebar Navigation Items

```typescript
// Спринты - фильтрация по клику
<button onclick={() => filterSprint = activeSprint?.id || ''}>
  {activeSprint.name}
</button>
```

## Design System

- EKF Dark sidebar: `bg-ekf-dark`
- Active nav item: `bg-ekf-red text-white`
- Inactive nav: `text-gray-300 hover:bg-gray-700`
- Section headers: `text-xs text-gray-500 uppercase tracking-wider`
- Status badges:
  - Active sprint: `bg-green-500/20 text-green-400`
  - Dev release: `bg-yellow-500/20 text-yellow-400`

## Testing

- [x] Sidebar отображается корректно
- [x] Навигация по секциям работает
- [x] Фильтрация по спринту работает
- [x] User profile отображается
- [x] Sprint header показывает прогресс
