# ARCHITECTURE.md
# EKF Hub — Architecture Overview (Companion to CONTEXT.md)

## Назначение

Этот файл — карта системы: компоненты, границы ответственности, потоки данных.
Инварианты и запреты — в CONTEXT.md. При конфликте побеждает CONTEXT.md.

---

## High-level схема

Production:
User → nginx (TLS termination, security headers, gzip)
     → Frontend (SvelteKit)
     → Backend (Go/Fiber)
         → Domain/App
         → Infrastructure
             → PostgreSQL
             → Redis (optional cache)

Connector Service (on-prem):
  ↔ Active Directory
  ↔ Exchange Web Services
  → общается с системой только через Backend APIs

---

## Границы ответственности

Frontend:
- визуализация и UX
- нет domain authority

Backend:
- единственный источник истины
- бизнес-логика и валидации
- state transitions

Connector:
- anti-corruption layer
- интеграции и маппинг протоколов
- без доменной истины

DB:
- пассивное хранилище

Redis:
- кэш производных данных
- не источник истины

---

## Backend layering (референс)

```
/cmd
/internal
  /handlers     HTTP handlers (thin)
  /ws           WS handlers (thin)
  /services     orchestration / use cases
  /domain       entities, rules, state machines
  /infra        DB/Redis/external clients/adapters
```

Dependency direction:
- handlers/ws → services → domain
- services → infra
- domain → nothing

Reverse dependencies forbidden.

---

## Networking (production expectation)

- Снаружи доступен только nginx
- Backend/Frontend не торчат наружу напрямую
- PostgreSQL не экспонируется наружу

Фактическая конфигурация фиксируется в IMPROVEMENTS.md.
