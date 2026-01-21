# CONTEXT.md
# EKF Hub — System Context & Non-Negotiable Invariants
# DO NOT IGNORE. DO NOT VIOLATE. DO NOT GUESS.

## Назначение

Этот файл — "конституция" проекта EKF Hub: архитектурные границы, инварианты и
фактический security posture. Он нужен, чтобы:
- AI и люди не теряли контекст,
- не ломали безопасность и архитектуру,
- не противоречили текущему состоянию production.

Любые изменения делаются только после чтения этого файла.
Если изменение может нарушить инвариант — STOP и запрос уточнений.

---

## Обзор системы

EKF Hub — корпоративная платформа для управления командой, коммуникациями и
процессами. Есть интеграции с Active Directory и Exchange Web Services.

Компоненты:
- Backend: Go 1.23 + Fiber v2 — единственный источник истины (state + бизнес-логика)
- Frontend: SvelteKit 2 + Svelte 5 + Tailwind — UI слой (без domain authority)
- Connector Service — boundary для on-prem интеграций (AD/EWS), anti-corruption layer
- Database: PostgreSQL 16.11 — пассивное хранилище
- nginx — reverse proxy и единая точка входа в production (TLS termination)
- Redis — инфраструктурно добавлен (кэш), но доменной истины не хранит

---

## Инварианты архитектуры (НЕ ОБСУЖДАЕТСЯ)

### Frontend
- Только UI (рендеринг и UX)
- Никакой бизнес-логики
- Никакой доменной валидации (кроме UX-подсказок)
- Нет прямого доступа к БД
- Общение только через HTTP API / WebSockets

### Backend
- Владеет всей бизнес-логикой и инвариантами
- Владеет всеми state transitions
- Валидирует все входные данные
- Единственный источник истины

### Connector Service
- Только интеграции и преобразование протоколов
- Никакой core бизнес-логики
- Никакого владения состоянием домена
- Нет доступа к БД
- Общение с системой только через Backend APIs

### Database
- Пассивное хранилище
- Никакой бизнес-логики в triggers/functions
- Схема отражает домен, а не диктует его
- Все изменения схемы только через миграции

---

## Инварианты безопасности (КРИТИЧНО)

- Никаких секретов/паролей в коде и docker-compose
- Секреты только через env (.env / secret store)
- TLS verify включён в production (никаких skip verify)
- HTTPS обязателен в production (через nginx)
- PostgreSQL не экспонируется наружу (только внутренняя Docker сеть)
- Валидация всех входных данных обязательна (HTTP/WS/connector/uploads)
- Логи не должны содержать токены/пароли/секреты/чувствительные payload'ы

Нарушение любого пункта — критический дефект.

---

## Backend правила (Go)

Слои:
- handlers/ws: thin (parse → call service → return)
- services/app: orchestration use-cases
- domain: сущности, правила, state machines (истина)
- infra: DB/Redis/клиенты внешних систем/адаптеры (только техника)

Запрещено:
- бизнес-логика в handlers/ws
- implicit state transitions
- дублирование правил между слоями
- "временные хаки" в проде

---

## Observability (инварианты)

- Структурированные логи (желательно zap/zerolog)
- Health checks отражают реальные зависимости (DB/Redis если используется)
- В production должны быть мониторинг и алертинг (Prometheus/Grafana) и агрегация логов

---

## Runtime & Production posture (ФАКТЫ)

Docker Compose — эталонный runtime.

Фактические улучшения и текущий статус production фиксируются в IMPROVEMENTS.md.
Если CONTEXT.md конфликтует с IMPROVEMENTS.md по "фактам", нужно обновить CONTEXT.md
через предложенный git diff (после подтверждения).

---

## Протокол изменений

Перед любыми правками:
1) прочитать README.md
2) прочитать этот файл
3) прочитать ARCHITECTURE.md
4) прочитать IMPROVEMENTS.md (если есть)
5) если затрагивается бизнес-логика: DOMAIN_GLOSSARY.md и STATE_MACHINES.md
6) если затрагиваются стандарты: AI_SAFE_CHECKLIST.md

Если информации недостаточно — НЕ ДОДУМЫВАТЬ, а открыть нужные файлы/код и уточнить.
