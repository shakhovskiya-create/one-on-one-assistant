# EKF Hub

Корпоративная платформа для управления командой, коммуникациями и бизнес-процессами.

## Обзор

EKF Hub — это интегрированное корпоративное решение, объединяющее управление встречами, задачами, проектами, мессенджер, почту и календарь в едином интерфейсе. Система интегрируется с корпоративной инфраструктурой (Active Directory, Exchange) и использует AI для автоматизации рутинных задач.

## Ключевые возможности

- **Дашборд** — сводка по встречам, задачам и настроению команды
- **Сотрудники** — карточки сотрудников с досье, историей встреч и аналитикой
- **Проекты** — управление проектами с привязкой встреч и задач
- **Встречи** — планирование, проведение (с AI-скриптом) и анализ 1-на-1
- **Задачи** — канбан-доска с фильтрами и drag-and-drop
- **Мессенджер** — корпоративный чат в стиле Telegram с контактами и каналами
- **Почта** — полноценный почтовый клиент через Exchange EWS
- **Календарь** — просмотр и создание событий через Exchange
- **Аналитика** — графики настроения, статистика договорённостей, красные флаги

## Технологии

| Компонент | Технология |
|-----------|------------|
| Backend | Go 1.23 + Fiber v2 |
| Frontend | SvelteKit 2 + Svelte 5 + TailwindCSS |
| Database | PostgreSQL (Supabase) |
| Calendar & Mail | Exchange Web Services (EWS) |
| AI Transcription | OpenAI Whisper / Yandex SpeechKit |
| AI Analysis | Anthropic Claude |
| Auth | Active Directory (LDAP) |
| Desktop Connector | Tauri 2 (Rust) |

## Структура проекта

```
ekf-hub/
├── backend/                    # Go API сервер
│   ├── cmd/server/            # Точка входа
│   └── internal/
│       ├── config/            # Конфигурация
│       ├── database/          # Supabase клиент
│       ├── handlers/          # HTTP/WebSocket обработчики
│       ├── models/            # Модели данных
│       └── services/          # Бизнес-логика
│           ├── ai/            # OpenAI/Anthropic/Yandex
│           ├── ews/           # Exchange Web Services
│           ├── mail/          # Почтовый сервис
│           └── messenger/     # Мессенджер сервис
│
├── frontend/                   # SvelteKit приложение
│   └── src/
│       ├── lib/
│       │   ├── api/           # TypeScript API клиент
│       │   ├── components/    # UI компоненты
│       │   └── stores/        # Svelte stores (auth, app)
│       └── routes/
│           ├── login/         # Авторизация AD
│           ├── employees/     # Сотрудники
│           ├── projects/      # Проекты с Gantt
│           ├── meetings/      # Встречи + загрузка + скрипт
│           ├── tasks/         # Канбан-доска
│           ├── messenger/     # Telegram-style мессенджер
│           ├── mail/          # Outlook-style почта
│           ├── calendar/      # Календарь EWS
│           └── analytics/     # Аналитика с графиками
│
├── connector-app/             # Tauri Desktop приложение
│   └── src-tauri/             # Rust backend для AD sync
│
└── docker-compose.yml         # Docker deployment
```

## Функциональность

### Авторизация
- Вход через Active Directory (EWS аутентификация)
- Автоматическое определение организационной структуры из AD
- Принудительная смена пароля при первом входе
- JWT токены для API-запросов

### Сотрудники
- Список сотрудников с фото из AD
- Карточка с контактами, должностью, отделом
- Досье: история встреч, договорённости, настроение
- Быстрая навигация к аналитике и календарю
- Иерархия подчинённых

### Встречи
- **Категории**: 1-на-1, проект, команда, планирование, ретро
- **Загрузка записи**: drag-and-drop аудиофайлов с AI-обработкой
- **Скрипт 1-на-1**: таймер секций, чеклист вопросов, заметки
- **Результаты**: транскрипт, резюме, договорённости, красные флаги
- **Оценка настроения**: шкала 1-10 с визуализацией

### Задачи (Канбан)
- **Виды отображения**: доска, список, таблица
- **Колонки**: Backlog → К выполнению → В работе → На проверке → Готово
- **Фильтры**: сотрудник, проект, приоритет, статус, поиск
- **Приоритеты**: цветные флаги (низкий/средний/высокий/критичный)
- Drag-and-drop перемещение

### Мессенджер
- **Telegram-style интерфейс** с тремя вкладками
- **Чаты**: личные диалоги с историей сообщений
- **Контакты**: иерархия по отделам из AD
- **Каналы**: корпоративные каналы для объявлений
- WebSocket для real-time обновлений
- Индикаторы набора текста и статуса прочтения

### Почта
- **Outlook-style интерфейс** с панелями
- Полноценная интеграция с Exchange EWS
- **Папки**: Входящие, Отправленные, Черновики, Удалённые
- Чтение, отправка и удаление писем
- Поддержка HTML-контента и вложений
- Сессионная авторизация для безопасности

### Календарь
- **Три режима**: день, неделя, месяц
- Интеграция с Exchange EWS
- Создание событий с выбором переговорной
- Поддержка онлайн-встреч (Teams, Zoom, Meet)
- Добавление участников из списка сотрудников
- Синхронизация с корпоративным календарём
- Индикатор текущего времени

### Аналитика
- **Дашборд**: ключевые метрики команды
- **Графики настроения**: линейные и столбчатые диаграммы
- **Статистика договорённостей**: donut chart
- **Красные флаги**: выгорание, риск ухода
- **Фильтры**: сотрудник, период (неделя/месяц/квартал/год)
- Список сотрудников, требующих внимания

### Проекты
- Создание и редактирование проектов
- Привязка встреч к проектам
- Счётчик встреч по проекту
- Ссылка на Gantt-диаграмму

## Запуск

### Docker Compose (рекомендуется)

```bash
# Клонировать репозиторий
git clone <repo-url>
cd ekf-hub

# Настроить переменные окружения
cp .env.example .env
# Отредактировать .env

# Запустить
docker-compose up -d
```

Приложение будет доступно на `http://localhost` (nginx reverse proxy).

### Локальная разработка

#### Backend

```bash
cd backend
cp .env.example .env
# Настроить переменные окружения

go mod tidy
go run ./cmd/server
```

#### Frontend

```bash
cd frontend
npm install
npm run dev
```

## Переменные окружения

### Backend (.env)

```env
# Сервер
PORT=8080
CORS_ORIGINS=http://localhost:5173,http://localhost:3000

# База данных
SUPABASE_URL=https://xxx.supabase.co
SUPABASE_KEY=eyJ...

# AI сервисы
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...
YANDEX_API_KEY=AQV...
YANDEX_FOLDER_ID=b1g...

# Exchange EWS
EWS_URL=https://mail.company.com/EWS/Exchange.asmx
EWS_DOMAIN=company.local

# Telegram (опционально)
TELEGRAM_BOT_TOKEN=xxx
```

### Frontend

```env
VITE_API_URL=http://localhost:8080
```

## API Endpoints

### Аутентификация
| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/api/v1/auth/login` | Вход через AD |
| POST | `/api/v1/auth/change-password` | Смена пароля |
| GET | `/api/v1/auth/me` | Текущий пользователь |

### Сотрудники
| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/api/v1/employees` | Список сотрудников |
| GET | `/api/v1/employees/:id` | Карточка сотрудника |
| GET | `/api/v1/employees/:id/dossier` | Досье сотрудника |

### Встречи
| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/api/v1/meetings` | Список встреч |
| POST | `/api/v1/meetings` | Создать встречу |
| GET | `/api/v1/meetings/:id` | Получить встречу |
| GET | `/api/v1/meeting-categories` | Категории |
| POST | `/api/v1/process-meeting` | AI-обработка аудио |

### Задачи
| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/api/v1/tasks` | Список задач |
| POST | `/api/v1/tasks` | Создать задачу |
| PUT | `/api/v1/tasks/:id` | Обновить задачу |
| DELETE | `/api/v1/tasks/:id` | Удалить задачу |
| GET | `/api/v1/kanban` | Канбан-доска |
| PUT | `/api/v1/kanban/move` | Переместить задачу |

### Мессенджер
| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/api/v1/messenger/chats` | Список чатов |
| GET | `/api/v1/messenger/messages/:chatId` | Сообщения чата |
| POST | `/api/v1/messenger/messages` | Отправить сообщение |
| GET | `/api/v1/messenger/channels` | Список каналов |
| WS | `/ws/messenger` | WebSocket подключение |

### Почта (EWS)
| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/api/v1/mail/folders` | Папки почты |
| POST | `/api/v1/mail/emails` | Список писем |
| POST | `/api/v1/mail/send` | Отправить письмо |
| POST | `/api/v1/mail/read` | Отметить прочитанным |
| DELETE | `/api/v1/mail/emails/:id` | Удалить письмо |

### Календарь (EWS)
| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/api/v1/calendar/:id` | События сотрудника |
| GET | `/api/v1/calendar/:id/simple` | Простой календарь |
| POST | `/api/v1/calendar/free-slots` | Свободные слоты |
| POST | `/api/v1/calendar/sync` | Синхронизация |

### Аналитика
| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/api/v1/analytics/dashboard` | Дашборд |
| GET | `/api/v1/analytics/employee/:id` | Аналитика сотрудника |

### Проекты
| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/api/v1/projects` | Список проектов |
| POST | `/api/v1/projects` | Создать проект |
| GET | `/api/v1/projects/:id` | Получить проект |
| DELETE | `/api/v1/projects/:id` | Удалить проект |

## Бизнес-логика

### Workflow встречи 1-на-1

1. **Планирование**: Руководитель выбирает сотрудника и время через календарь
2. **Проведение**: Используется скрипт с таймером и чеклистом вопросов
3. **Запись**: Аудио загружается для AI-обработки
4. **Анализ**: Whisper транскрибирует, Claude анализирует и выделяет:
   - Краткое резюме встречи
   - Договорённости с дедлайнами
   - Красные флаги (выгорание, риск ухода)
   - Оценка настроения (1-10)
5. **Мониторинг**: Договорённости отслеживаются, аналитика показывает тренды

### Интеграция с AD

- Connector-app (Tauri) работает на on-prem сервере
- Периодическая синхронизация списка сотрудников
- Получение фото, должностей, отделов, иерархии
- EWS используется для авторизации и доступа к календарю/почте

### Модель данных

- **Employee**: сотрудник (id, name, email, position, department, manager_id)
- **Meeting**: встреча (id, employee_id, category, date, mood_score, summary)
- **Agreement**: договорённость (id, meeting_id, text, deadline, status)
- **Task**: задача (id, title, status, priority, assignee_id, project_id)
- **Project**: проект (id, name, description)
- **Message**: сообщение мессенджера (id, chat_id, sender_id, content, timestamp)
- **Channel**: канал (id, name, description, members)

## Deployment

### Production (Docker)

```yaml
# docker-compose.yml уже настроен
# nginx reverse proxy на портах 80/443
# Volumes для данных PostgreSQL
```

### Railway/Vercel

- Backend: Railway с Go buildpack
- Frontend: Vercel с SvelteKit adapter-auto
- Database: Supabase managed PostgreSQL

## Лицензия

Proprietary - EKF Group
