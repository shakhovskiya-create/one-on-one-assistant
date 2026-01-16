# EKF Team Hub

Система управления встречами 1-на-1 и командной работой.

## Технологии

- **Backend**: Go + Fiber
- **Frontend**: SvelteKit + Svelte 5 + TailwindCSS
- **Database**: Supabase (PostgreSQL)
- **Calendar**: Exchange EWS (on-prem)
- **AI**: OpenAI Whisper + Anthropic Claude

## Структура проекта

```
one-on-one-assistant/
├── backend/                 # Go API сервер
│   ├── cmd/server/         # Точка входа
│   ├── internal/
│   │   ├── config/         # Конфигурация
│   │   ├── database/       # Supabase клиент
│   │   ├── handlers/       # HTTP обработчики
│   │   ├── models/         # Модели данных
│   │   └── services/       # Бизнес-логика
│   └── pkg/
│       ├── ai/             # OpenAI/Anthropic/Yandex
│       ├── ews/            # Exchange Web Services
│       └── telegram/       # Telegram Bot
│
├── frontend/               # SvelteKit приложение
│   ├── src/
│   │   ├── lib/
│   │   │   ├── api/       # API клиент
│   │   │   ├── components/# UI компоненты (Header, Sidebar)
│   │   │   └── stores/    # Svelte stores (auth, app)
│   │   └── routes/
│   │       ├── login/     # Авторизация через AD
│   │       ├── employees/ # Сотрудники и досье
│   │       ├── meetings/  # Встречи с категориями
│   │       ├── tasks/     # Канбан-доска (board/list/table)
│   │       ├── projects/  # Проекты
│   │       ├── calendar/  # Календарь EWS
│   │       ├── analytics/ # Аналитика с графиками
│   │       ├── upload/    # Загрузка аудио встреч
│   │       └── script/    # Скрипт проведения 1-на-1
│   └── static/
│
├── connector-app/          # Tauri приложение для AD sync
│
└── _legacy/               # Старый код (Python + Vue)
```

## Функциональность

### Авторизация
- Вход через Active Directory (EWS)
- Автоматическое определение подчиненных из AD-иерархии
- Хранение сессии в localStorage

### Сотрудники
- Список подчиненных сотрудников
- Карточка сотрудника с контактами
- Досье: история встреч, договоренности, красные флаги
- Быстрый переход к аналитике и календарю

### Встречи
- Категории встреч: 1-на-1, проект, команда, планирование, ретро
- Фильтрация по категориям
- Оценка настроения (mood score)
- Краткое содержание и длительность

### Загрузка аудио
- Drag-and-drop загрузка аудиофайлов
- Выбор категории, сотрудника и проекта
- AI-обработка: транскрипция (Whisper), анализ (Claude)
- Результат: транскрипт, резюме, договоренности, красные флаги

### Скрипт встречи 1-на-1
- Таймер секций и общего времени
- 6 секций: Чекин, Повестка сотрудника, Повестка руководителя, Развитие, Обратная связь, Договоренности
- Чеклист вопросов с заметками
- Визуальный прогресс по секциям

### Задачи (Канбан)
- 3 вида отображения: доска, список, таблица
- 5 колонок: Backlog, К выполнению, В работе, На проверке, Готово
- Drag-and-drop перемещение между колонками
- Фильтры: сотрудник, проект, приоритет, статус, поиск
- Цветные флаги приоритетов
- Создание и редактирование задач в модальных окнах

### Аналитика
- Дашборд с ключевыми метриками
- График динамики настроения (SVG line chart)
- Статистика по договоренностям (donut chart)
- Тренды по категориям встреч (bar chart)
- Список красных флагов
- Фильтрация по сотрудникам

### Календарь
- Интеграция с Exchange EWS
- Просмотр событий сотрудников
- Поиск свободных слотов для встреч
- Синхронизация календаря

### Проекты
- Список проектов с описанием
- Привязка сотрудников к проектам
- Статус проекта

## Запуск

### Backend

```bash
cd backend
cp .env.example .env
# Настройте переменные окружения

go mod tidy
go run ./cmd/server
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```

## Переменные окружения

### Backend (.env)

```env
PORT=8080
SUPABASE_URL=https://xxx.supabase.co
SUPABASE_KEY=eyJ...
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...
EWS_URL=https://post.ekf.su/EWS/Exchange.asmx
EWS_DOMAIN=ekf.local
TELEGRAM_BOT_TOKEN=xxx
```

### Frontend

```env
VITE_API_URL=http://localhost:8080
```

## API Endpoints

### Сотрудники
- `GET /api/v1/employees` - Список сотрудников
- `POST /api/v1/employees` - Создать сотрудника
- `GET /api/v1/employees/:id` - Получить сотрудника
- `PUT /api/v1/employees/:id` - Обновить сотрудника
- `DELETE /api/v1/employees/:id` - Удалить сотрудника
- `GET /api/v1/employees/:id/dossier` - Досье сотрудника

### Проекты
- `GET /api/v1/projects` - Список проектов
- `POST /api/v1/projects` - Создать проект
- `GET /api/v1/projects/:id` - Получить проект
- `PUT /api/v1/projects/:id` - Обновить проект
- `DELETE /api/v1/projects/:id` - Удалить проект

### Встречи
- `GET /api/v1/meetings` - Список встреч
- `GET /api/v1/meetings/:id` - Получить встречу
- `GET /api/v1/meeting-categories` - Категории встреч
- `POST /api/v1/process-meeting` - Обработать аудио встречи

### Задачи
- `GET /api/v1/tasks` - Список задач
- `POST /api/v1/tasks` - Создать задачу
- `GET /api/v1/tasks/:id` - Получить задачу
- `PUT /api/v1/tasks/:id` - Обновить задачу
- `DELETE /api/v1/tasks/:id` - Удалить задачу
- `GET /api/v1/kanban` - Канбан-доска
- `PUT /api/v1/kanban/move` - Переместить задачу

### Календарь (EWS)
- `POST /api/v1/calendar/:id` - Получить события
- `GET /api/v1/calendar/:id/simple` - Простой календарь
- `POST /api/v1/calendar/free-slots` - Найти свободные слоты
- `POST /api/v1/calendar/sync` - Синхронизировать

### Аналитика
- `GET /api/v1/analytics/dashboard` - Дашборд
- `GET /api/v1/analytics/employee/:id` - Аналитика сотрудника
- `GET /api/v1/analytics/employee/:id/by-category` - По категориям

### AD Интеграция
- `GET /api/v1/connector/status` - Статус коннектора
- `POST /api/v1/ad/sync` - Синхронизация AD
- `POST /api/v1/ad/authenticate` - Аутентификация

## Лицензия

Proprietary - EKF
