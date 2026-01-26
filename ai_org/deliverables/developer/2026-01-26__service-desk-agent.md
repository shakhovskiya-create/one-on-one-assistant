# Service Desk Agent Console

**Дата:** 2026-01-26
**Статус:** ✅ DONE

## Summary

Реализована консоль агента Service Desk по макету 03-service-desk-agent.html.

## New Files

### `frontend/src/routes/service-desk/agent/+page.svelte`

Agent Console с:
- Sidebar навигация (Queue, Incidents, Requests, My Tickets)
- Stats Bar (SLA, открыто, просрочено, решено)
- Таблица тикетов с фильтрами
- Выбор тикета с подсветкой

## Modified Files

### `frontend/src/routes/service-desk/+page.svelte`

- Добавлен переключатель "Консоль агента"
- Добавлен timeout 10 сек для предотвращения зависания
- Улучшена обработка ошибок авторизации

## Features

### Agent Console Sidebar
```
┌─────────────────────────┐
│ Service Desk            │
│ Agent Console           │
├─────────────────────────┤
│ Queue              [24] │  ← Все открытые
│ Incidents          [18] │  ← Только инциденты
│ Service Requests    [6] │  ← Запросы
│ ───────────────────     │
│ My Tickets              │  ← Назначенные
├─────────────────────────┤
│ ← Портал пользователя   │
└─────────────────────────┘
```

### Stats Bar
- SLA % (зелёный)
- Открыто (синий)
- Просрочено (жёлтый)
- Решено сегодня (фиолетовый)

### Tickets Table
| Номер | Тип | Приоритет | Тема | Заявитель | Статус | SLA | Создано |

### Portal/Agent Toggle
- На Portal: кнопка "Консоль агента" (верхний правый угол)
- На Agent: ссылка "Портал пользователя" (нижняя часть sidebar)

## Design System

- EKF Dark sidebar: `bg-ekf-dark`
- Active nav: `bg-ekf-red`
- Status colors: blue/yellow/purple/green/gray
- Priority colors: red/orange/yellow/green

## Bugfixes

- `{@const}` обёрнут в `{#if true}` - Svelte 5 требует `@const` внутри блоков
