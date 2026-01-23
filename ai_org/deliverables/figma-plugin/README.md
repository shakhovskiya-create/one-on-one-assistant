# EKF Hub UI Kit - Figma Plugin

Figma плагин для автоматической генерации UI компонентов в стиле EKF Hub.

## Установка

### Способ 1: Локальная установка (для разработки)

1. Открой Figma Desktop App
2. Перейди в **Plugins** > **Development** > **Import plugin from manifest...**
3. Выбери файл `manifest.json` из этой папки
4. Плагин появится в меню **Plugins** > **Development** > **EKF Hub UI Kit**

### Способ 2: Быстрый запуск

1. В Figma нажми `Cmd+/` (Mac) или `Ctrl+/` (Win)
2. Введи "Import plugin from manifest"
3. Выбери `manifest.json`

## Использование

1. Открой плагин: **Plugins** > **Development** > **EKF Hub UI Kit**
2. Появится панель с кнопками
3. Нажми на нужную кнопку чтобы сгенерировать компонент

## Доступные компоненты

### Полные страницы
- **Tasks Page** - Полная страница задач с Kanban доской, фильтрами, sidebar
- **Component Library** - Библиотека всех UI компонентов

### Отдельные компоненты
- **Sidebar** - Боковая навигация (темная)
- **Header** - Шапка страницы
- **Filter Bar** - Панель фильтров с поиском
- **Kanban Column** - Колонка Kanban доски
- **Task Card** - Карточка задачи

## Цвета EKF Hub

| Название | HEX | Использование |
|----------|-----|---------------|
| EKF Red | `#E53935` | Primary, активные элементы |
| EKF Dark | `#1a1a2e` | Sidebar, темный текст |
| Gray 100 | `#f3f4f6` | Фон контента |
| Gray 200 | `#e5e7eb` | Границы, разделители |
| Gray 500 | `#6b7280` | Вторичный текст |
| Gray 900 | `#111827` | Основной текст |

## Файлы

```
figma-plugin/
├── manifest.json   # Конфигурация плагина
├── code.js         # Логика плагина (JavaScript)
├── code.ts         # Исходник (TypeScript)
├── ui.html         # Интерфейс плагина
└── README.md       # Эта документация
```

## Шрифты

Плагин использует шрифт **Inter** (Regular, Medium, SemiBold, Bold).
Убедись что шрифт установлен или доступен в Figma.

## Troubleshooting

**"Font not found"** - Установи шрифт Inter или замени в коде на доступный системный шрифт.

**Компонент не появляется** - Проверь что у тебя есть хотя бы один Frame/Page в документе.
