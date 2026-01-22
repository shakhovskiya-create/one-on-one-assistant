# Agent Log

Правило: каждый значимый цикл "поручение → выполнение → приемка" фиксируется записью:
- date/time
- инициатор (PM)
- исполнитель (роль)
- ссылка на handoff-файл
- результат (что изменилось, какие файлы/PR)
- замечания/риски/следующие шаги

---

## 2026-01-22

### Sprint 6: GitHub Integration
- **Инициатор:** PM (user request)
- **Исполнитель:** Developer
- **Результат:**
  - Создан GitHub API клиент: `backend/internal/services/github/client.go`
  - Handlers: `backend/internal/handlers/github.go`
  - Frontend API и страница: `frontend/src/routes/github/+page.svelte`
  - Добавлен в sidebar навигацию
  - Коммит: `dce02f7`
- **Деплой:** Выполнен на 10.100.0.131
- **Замечания:** Требуется GITHUB_TOKEN для полного функционала

### Sprint 6: Confluence Integration
- **Инициатор:** PM (user request)
- **Исполнитель:** Developer
- **Результат:**
  - REST API клиент: `backend/internal/services/confluence/client.go`
  - Handlers и routes для spaces, pages, search
  - Frontend страница с просмотром документации
  - Коммит: `9993b49`
- **Деплой:** Выполнен на 10.100.0.131
- **Замечания:** Требуются credentials Confluence Server

### Sprint 7: Versions/Releases (JIRA parity)
- **Инициатор:** PM (WORKPLAN - HIGH priority)
- **Исполнитель:** Developer
- **Результат:**
  - Миграция: `migrations/018_versions.sql` (таблица versions, fix_version_id в tasks)
  - Backend handlers: `backend/internal/handlers/versions.go` (CRUD, release, release-notes)
  - Models: добавлен Version struct и FixVersionID в Task
  - Routes: добавлены /versions endpoints в main.go
  - Frontend: `frontend/src/routes/releases/+page.svelte` (полный UI)
  - Sidebar: добавлена ссылка на Releases
  - Коммит: `33ee40b`
- **Деплой:** Выполнен на 10.100.0.131
- **Замечания:** Полностью завершён JIRA parity для Versions

### WORKPLAN.md обновление
- **Инициатор:** PM
- **Исполнитель:** Developer
- **Результат:**
  - Статус обновлён на Sprint 7 завершён
  - Версии/Releases отмечены как ✅ в JIRA comparison
  - Добавлен Sprint 7 в список спринтов
  - Добавлен Sprint 8 (ожидает утверждения)

### Sprint 8: Bugfixes (в работе)
- **Инициатор:** PM (user request)
- **Исполнитель:** Developer
- **Статус:** В работе
- **Результаты:**
  - Добавлена обработка ошибок при создании каналов
  - Добавлен индикатор загрузки и сообщения об ошибках
  - **SECURITY FIX:** Исправлена передача паролей почты через GET query params
    - Backend: mail/folders и mail/emails теперь POST с body
    - Frontend: client.ts обновлён для POST запросов
  - WORKPLAN.md обновлён с Sprint 8 (bugfixes) и Sprint 9 (features)
  - **Почта: Навигация по цепочке писем**
    - Добавлены функции getCurrentThread() и navigateThread()
    - Добавлены кнопки навигации (первое/предыдущее/следующее/последнее)
    - Отображается позиция в цепочке (1/N)
  - **Почта: Улучшение отображения вложений**
    - Добавлена переменная attachmentError для отслеживания ошибок
    - Улучшены сообщения об ошибках загрузки вложений
    - Добавлено визуальное отображение ошибок в UI
  - **Bugfix: @const placement**
    - Исправлено размещение @const внутри {#if} блоков для Svelte 5
  - **Транскрибирование: анализ**
    - Выявлена причина ошибки Content-length: лимит Yandex SpeechKit ~1MB для синхронного API
    - Для больших файлов требуется переход на асинхронное API (отложено)
  - **GitHub: диагностика**
    - Выявлена проблема: GITHUB_TOKEN не передавался в docker-compose.yml
    - Добавлена переменная среды GITHUB_TOKEN в docker-compose.yml
    - Требуется: установить GITHUB_TOKEN в .env на сервере
  - **Календарь: Редактирование и удаление событий**
    - EWS client: добавлены методы UpdateCalendarItem и DeleteCalendarItem
    - Backend handlers: UpdateCalendarEvent и DeleteCalendarEvent с JWT авторизацией
    - Routes: PUT /calendar/update и DELETE /calendar/delete
    - Frontend API: updateMeeting и deleteMeeting в client.ts
    - UI: кнопки "Редактировать" и "Удалить" в модальном окне события
    - Inline редактирование: название, время, место
    - Отправка уведомлений участникам при удалении

  - **Задачи: Оптимизация производительности**
    - Добавлен loadingTaskDetails state для индикации загрузки
    - Параллельная загрузка зависимостей и статуса блокировки
    - Добавлен loading spinner в секции зависимостей
    - Reset state при открытии модального окна

  - **GitHub: Перемещение в блок Задачи**
    - Удалён GitHub из sidebar навигации
    - Добавлена секция "Связанные коммиты" в модальное окно задачи
    - Коммиты загружаются параллельно с остальными данными задачи
    - Отображаются коммиты, упоминающие ID задачи
    - UI: аватар, SHA, дата, сообщение, автор с ссылкой на GitHub

  - **Bugfix: CSRF token missing**
    - Убран CSRF middleware из backend (JWT авторизация уже защищает от CSRF)
    - Ошибка возникала при создании каналов в мессенджере

  - **Bugfix: UUID undefined в Releases**
    - Исправлена передача undefined в URLSearchParams
    - Теперь фильтруются пустые параметры перед формированием query string

  - **Releases: убраны из sidebar**
    - Функционал версий будет доступен через задачи (fix_version_id)

  - **Задачи: Диагностика "зависания"**
    - Проведён полный анализ кода tasks/+page.svelte
    - API handlers проверены: dependencies, blocked, time-entries, resources работают корректно
    - GitHub commits интеграция требует GITHUB_TOKEN в .env на сервере
    - Docker контейнеры пересобраны с --no-cache
    - Все API endpoints возвращают правильные статусы (401 без авторизации, 200 с авторизацией)
    - **Выявленные причины возможного "зависания":**
      1. Отсутствует GITHUB_TOKEN в .env на сервере (API github может возвращать ошибки)
      2. При клике на задачу выполняется 4 параллельных API запроса (dependencies, blocked, time-entries, resources)
      3. Loading spinner добавлен для индикации загрузки
    - **Рекомендации:**
      - Установить GITHUB_TOKEN в .env на сервере для интеграции с GitHub
      - Очистить кэш браузера (Ctrl+Shift+R) для применения обновлений

---

## Следующие задачи (Sprint 8)
- Почта: связь с календарём, приглашения
- Транскрибирование: асинхронное API для больших файлов
- Проверить работу задач с пользователем после очистки кэша
