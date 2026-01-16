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
│   │   │   ├── components/# UI компоненты
│   │   │   └── stores/    # Svelte stores
│   │   └── routes/        # Страницы
│   └── static/
│
├── connector-app/          # Tauri приложение для AD sync
│
└── _legacy/               # Старый код (Python + Vue)
```

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
