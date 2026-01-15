# EKF Team Hub

Корпоративная платформа управления командой с интеграцией Active Directory и Exchange. Планирование встреч, аналитика и работа с сотрудниками в едином интерфейсе.

## Архитектура

```
┌─────────────────────────────────────────────────────────────────┐
│                         Cloud (Railway)                          │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────────┐  │
│  │  Frontend   │    │   Backend   │    │      Supabase       │  │
│  │  Next.js    │◄──►│   FastAPI   │◄──►│  PostgreSQL + Auth  │  │
│  │  Port 3000  │    │  Port 8000  │    │                     │  │
│  └─────────────┘    └──────┬──────┘    └─────────────────────┘  │
│                            │ WebSocket                           │
└────────────────────────────┼────────────────────────────────────┘
                             │
                    ┌────────▼────────┐
                    │   On-Prem       │
                    │   Connector     │
                    │   Python        │
                    └────────┬────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
       ┌──────▼──────┐ ┌─────▼─────┐ ┌─────▼─────┐
       │    LDAP     │ │  Exchange │ │   File    │
       │  (AD Users) │ │   (EWS)   │ │  Shares   │
       └─────────────┘ └───────────┘ └───────────┘
```

## Возможности

### Базовые функции
- **Интерактивный скрипт встречи** - чеклист с таймером для проведения встречи
- **Автоматическая транскрипция** - загрузите аудио/видео, получите текст
- **AI-анализ** - автоматическое заполнение протокола, выделение договоренностей
- **Красные флаги** - обнаружение признаков выгорания и риска ухода
- **История** - все встречи и договоренности в одном месте

### Интеграция с Active Directory
- **Синхронизация оргструктуры** - автоматический импорт сотрудников с manager_id
- **Фото из AD** - thumbnailPhoto отображается в профиле
- **Департаменты и должности** - из атрибутов AD
- **Пагинированный импорт** - батчи по 100 пользователей

### Интеграция с Exchange/Outlook
- **Синхронизация календаря** - импорт встреч из Exchange
- **Определение участников** - автоматическое сопоставление email → сотрудник
- **Планирование встреч** - интерфейс в стиле нового Outlook
- **Поиск свободных слотов** - автоматический подбор времени

### Авторизация с оргструктурой
- **Вход через AD** - LDAP аутентификация
- **Иерархия доступа** - сотрудник видит только своих подчинённых
- **Рекурсивные подчинённые** - все уровни вложенности

### Брендинг EKF
- **Цвета бренда** - оранжевый `#E63312`, тёмный `#1A1A1A`
- **Шрифт Inter** - современный корпоративный шрифт
- **Единый стиль** - все компоненты в фирменных цветах

### Страницы приложения

| Страница | Путь | Описание |
|----------|------|----------|
| Дашборд | `/` | Обзор предстоящих встреч и задач |
| Логин | `/login` | Авторизация через AD |
| Команда | `/employees` | Сотрудники в 3 видах: дерево / плитки / список |
| Досье | `/employees/[id]` | Карточка сотрудника с историей встреч |
| Календарь | `/calendar` | Планирование встреч в стиле Outlook |
| Встречи | `/meetings` | История всех встреч |
| Аналитика | `/analytics` | Аналитика по категориям встреч |
| Скрипт | `/script` | Интерактивный скрипт проведения встречи |
| Загрузка | `/upload` | Загрузка записей встреч |
| Настройки | `/settings` | Настройки коннектора AD/Exchange |

### Аналитика по категориям встреч
Для каждой категории встреч своя аналитика:

| Категория | Метрики |
|-----------|---------|
| **1-on-1** | Тренды настроения, красные флаги, выполнение договорённостей |
| **Ретро** | Action items, частые темы, динамика улучшений |
| **Планирование** | Velocity спринтов, точность оценок, распределение задач |
| **Статусы** | Частота блокеров, типы проблем, время решения |
| **Демо** | Количество фич, feedback, участие стейкхолдеров |

## Быстрый старт

### 1. Создайте аккаунты и получите ключи

#### Supabase (бесплатно)
1. Зарегистрируйтесь на [supabase.com](https://supabase.com)
2. Создайте новый проект
3. Перейдите в Settings → API и скопируйте:
   - Project URL → `SUPABASE_URL`
   - anon public key → `SUPABASE_KEY`
4. Перейдите в SQL Editor и выполните содержимое файла `backend/schema.sql`

#### OpenAI (для транскрипции)
1. Зарегистрируйтесь на [platform.openai.com](https://platform.openai.com)
2. Создайте API ключ → `OPENAI_API_KEY`
3. Пополните баланс (минимум $5)

#### Anthropic (для анализа)
1. Зарегистрируйтесь на [console.anthropic.com](https://console.anthropic.com)
2. Создайте API ключ → `ANTHROPIC_API_KEY`
3. Пополните баланс (минимум $5)

### 2. Настройте проект

```bash
# Клонируйте проект
git clone <repo-url>
cd one-on-one-assistant

# Создайте файл с переменными окружения
cp .env.example .env

# Отредактируйте .env и добавьте ваши ключи
nano .env
```

### 3. Запуск через Docker (рекомендуется)

```bash
docker-compose up -d
```

Приложение будет доступно:
- Frontend: http://localhost:3000
- Backend API: http://localhost:8000

### 4. Запуск без Docker

#### Backend
```bash
cd backend
python -m venv venv
source venv/bin/activate  # Windows: venv\Scripts\activate
pip install -r requirements.txt
uvicorn main:app --reload
```

#### Frontend
```bash
cd frontend
npm install
npm run dev
```

### 5. Настройка On-Prem Connector

Для интеграции с AD и Exchange нужен коннектор внутри корпоративной сети:

```bash
cd connector
python -m venv venv
source venv/bin/activate
pip install -r requirements.txt

# Настройте переменные
export BACKEND_WS_URL=wss://your-backend.railway.app/ws/connector
export CONNECTOR_TOKEN=your-secret-token
export AD_SERVER=ldap://your-ad-server
export AD_USER=CN=Service,OU=Users,DC=company,DC=local
export AD_PASSWORD=password
export EXCHANGE_SERVER=https://mail.company.local/EWS/Exchange.asmx
export EXCHANGE_EMAIL=service@company.local
export EXCHANGE_PASSWORD=password

# Запустите
python connector.py
```

## Структура проекта

```
one-on-one-assistant/
├── backend/
│   ├── main.py              # FastAPI приложение
│   ├── schema.sql           # SQL схема для Supabase
│   ├── requirements.txt     # Python зависимости
│   └── Dockerfile
├── frontend/
│   ├── app/
│   │   ├── page.tsx         # Дашборд
│   │   ├── layout.tsx       # Основной layout с sidebar
│   │   ├── providers.tsx    # AuthProvider + AuthGuard
│   │   ├── sidebar.tsx      # Навигация с инфо пользователя
│   │   ├── login/           # Страница входа
│   │   ├── calendar/        # Календарь в стиле Outlook
│   │   ├── employees/       # Команда (3 вида)
│   │   │   ├── page.tsx     # Список/дерево/плитки
│   │   │   └── [id]/        # Досье сотрудника
│   │   ├── meetings/        # История встреч
│   │   ├── analytics/       # Аналитика по категориям
│   │   ├── script/          # Скрипт встречи
│   │   ├── upload/          # Загрузка записей
│   │   └── settings/        # Настройки коннектора
│   ├── lib/
│   │   └── auth.tsx         # AuthContext с иерархией доступа
│   ├── package.json
│   └── Dockerfile
├── connector/
│   ├── connector.py         # On-prem коннектор
│   ├── ad_sync.py           # Синхронизация с AD
│   ├── ews_calendar.py      # Синхронизация с Exchange
│   └── requirements.txt
├── docker-compose.yml
├── .env.example
└── README.md
```

## API Endpoints

### Сотрудники
| Метод | Endpoint | Описание |
|-------|----------|----------|
| GET | `/employees` | Список всех сотрудников |
| POST | `/employees` | Добавить сотрудника |
| GET | `/employees/{id}` | Получить сотрудника |
| GET | `/employees/{id}/dossier` | Досье с историей встреч |
| GET | `/employees/my-team` | Мои подчинённые (рекурсивно) |

### Встречи
| Метод | Endpoint | Описание |
|-------|----------|----------|
| GET | `/meetings` | Список встреч |
| POST | `/meetings` | Создать встречу |
| GET | `/meetings/{id}` | Детали встречи |
| POST | `/meetings/{id}/resolve-participants` | Сопоставить участников |

### Календарь и Exchange
| Метод | Endpoint | Описание |
|-------|----------|----------|
| POST | `/calendar/sync` | Синхронизировать календарь из Exchange |
| GET | `/calendar/free-slots` | Найти свободные слоты |

### AD интеграция
| Метод | Endpoint | Описание |
|-------|----------|----------|
| POST | `/ad/sync` | Синхронизировать сотрудников из AD |
| GET | `/ad/subordinates/{id}` | Подчинённые из БД |
| POST | `/auth/login` | Авторизация через AD |

### Аналитика
| Метод | Endpoint | Описание |
|-------|----------|----------|
| GET | `/analytics/dashboard` | Общая аналитика |
| GET | `/analytics/employee/{id}` | Аналитика по сотруднику |
| GET | `/analytics/employee/{id}/by-category` | По категориям встреч |
| GET | `/meeting-categories` | Список категорий встреч |

### Транскрипция и анализ
| Метод | Endpoint | Описание |
|-------|----------|----------|
| POST | `/transcribe` | Транскрибировать аудио |
| POST | `/analyze` | Анализировать транскрипт |
| POST | `/process-meeting` | Полный пайплайн обработки |
| GET | `/script` | Получить скрипт встречи |

### WebSocket
| Endpoint | Описание |
|----------|----------|
| `/ws/connector` | Подключение on-prem коннектора |

## Переменные окружения

```bash
# Supabase
SUPABASE_URL=https://xxx.supabase.co
SUPABASE_KEY=eyJ...

# AI APIs
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...

# Connector auth
CONNECTOR_TOKEN=your-secret-token

# On-prem connector (в connector/.env)
BACKEND_WS_URL=wss://backend.railway.app/ws/connector
AD_SERVER=ldap://dc.company.local
AD_USER=CN=Service,OU=Users,DC=company,DC=local
AD_PASSWORD=...
AD_BASE_DN=DC=company,DC=local
EXCHANGE_SERVER=https://mail.company.local/EWS/Exchange.asmx
EXCHANGE_EMAIL=service@company.local
EXCHANGE_PASSWORD=...
```

## Стоимость

Примерная стоимость на 20 встреч в месяц (по 60 минут):

| Сервис | Стоимость |
|--------|-----------|
| Whisper API | ~$7 |
| Claude API | ~$3 |
| Supabase | $0 (free tier) |
| Railway | $5 (hobby plan) |
| **Итого** | **~$15/месяц** |

## Roadmap

### Выполнено
- [x] Базовый функционал (транскрипция, анализ, красные флаги)
- [x] Интеграция с AD (синхронизация сотрудников с оргструктурой)
- [x] On-prem коннектор для AD/Exchange (WebSocket → Cloud)
- [x] Интеграция с Exchange/Outlook (календарь, участники)
- [x] Авторизация с оргструктурой (AD login, иерархия доступа)
- [x] Досье сотрудника (карточка, история, аналитика)
- [x] Аналитика по категориям встреч
- [x] 3 вида отображения команды (дерево/плитки/список)
- [x] Календарь в стиле Outlook

### В работе
- [ ] Email-уведомления о просроченных договоренностях
- [ ] Экспорт отчетов в PDF/Excel

### Позже
- [ ] Сравнительный анализ по команде
- [ ] Интеграция с Jira/Asana для трекинга задач
- [ ] Десктоп-версия (Electron)
- [ ] Мобильное приложение

## Поддержка

При возникновении проблем:
1. Проверьте правильность API ключей в .env
2. Убедитесь, что выполнили SQL схему в Supabase
3. Проверьте баланс на OpenAI и Anthropic
4. Для AD/Exchange - проверьте логи коннектора

## Лицензия

MIT
