# Full Audit — Specification (Enterprise)

**Last updated:** 2026-01-26

## Objective
Полный аудит EKF Hub по архитектуре, backend, security, frontend, QA/testability, product/ops readiness.

## Scope
- Backend (handlers/middleware/config/database/services)
- Frontend (auth store, API client, key routes: mail/confluence/messenger)
- Infra (docker-compose, nginx)
- Docs (README/CONTEXT/ARCHITECTURE/SECURITY_AUDIT)

## Mental Map (System Model)
### Entry points
- nginx (80/443), WS: /ws/messenger, /ws/connector
- Backend API: /api/v1/*

### Trust boundaries
- Public: User → nginx
- Internal: nginx → frontend/backend → PostgreSQL/MinIO/Redis
- On‑prem: connector ↔ AD/EWS (WS + API key)

### State ownership
Backend владеет доменной логикой и state transitions. DB — пассивное хранилище.

---

# PART 1 — ARCHITECTURE

## [ARCHITECTURE ISSUE] Business logic в handlers (нарушение layering)
**Где в коде:** backend/internal/handlers/*.go (tasks/employees/calendar)
**Почему проблема:** нарушение separation of concerns, сложность тестирования, дублирование правил.
**Риск:** неконсистентные переходы, рост техдолга.
**Рекомендация:** вынести правила и state transitions в domain/services.

## [ARCHITECTURE ISSUE] State machine задач не соблюдается
**Где в коде:** backend/internal/handlers/tasks.go (MoveTaskKanban/UpdateTask)
**Почему проблема:** статусы "backlog/todo/in_progress" не совпадают с STATE_MACHINES.md (Backlog/ToDo/InProgress).
**Риск:** неконсистентные статусы, сломанная аналитика и workflow.
**Рекомендация:** унифицировать enum статусов + валидация переходов.

## [ARCHITECTURE ISSUE] Нет транзакций для multi‑step операций
**Где в коде:** backend/internal/handlers/tasks.go (CreateTask + tags + history)
**Почему проблема:** операции частично записываются.
**Риск:** битые данные, сложные инциденты.
**Рекомендация:** транзакции/идемпотентность.

---

# PART 2 — BACKEND (GO)

## [BACKEND DEFECT] IDOR на employees (no authz)
**Файл/функция:** backend/internal/handlers/employees.go (List/Get/Update/Delete)
**Описание:** нет проверки прав доступа по user_id/role.
**Риск:** чтение/изменение профилей других пользователей.
**Severity:** CRITICAL
**Как исправить:** RBAC/ABAC проверки в handlers или middleware.

## [BACKEND DEFECT] IDOR на календаре
**Файл/функция:** backend/internal/handlers/calendar.go (GetCalendar)
**Описание:** доступ к календарю по произвольному employee_id без проверки текущего пользователя.
**Риск:** утечка календаря/PII.
**Severity:** CRITICAL
**Как исправить:** проверка ownership/manager scopes.

## [BACKEND DEFECT] Не валидируются переходы статусов задач
**Файл/функция:** backend/internal/handlers/tasks.go (UpdateTask/MoveTaskKanban)
**Описание:** статус меняется на любой строковый value.
**Риск:** нарушены бизнес‑инварианты.
**Severity:** HIGH
**Как исправить:** enforce state machine из STATE_MACHINES.md.

## [BACKEND DEFECT] Mail API принимает произвольные credentials
**Файл/функция:** backend/internal/handlers/mail.go
**Описание:** логин/пароль из body не связан с текущим пользователем.
**Риск:** доступ к чужой почте при наличии credentials.
**Severity:** HIGH
**Как исправить:** привязка к user_id и безопасный credential store.

## [BACKEND DEFECT] File upload без валидации
**Файл/функция:** backend/internal/handlers/files.go (UploadFile)
**Описание:** нет MIME/size/AV‑контроля, файл читается в память целиком.
**Риск:** DoS и загрузка вредоносных файлов.
**Severity:** HIGH
**Как исправить:** streaming, лимиты, whitelist, AV‑сканирование.

---

# PART 3 — SECURITY AUDIT

## AppSec
1) **Reversible password storage (AES)**
   - backend/internal/utils/crypto.go + connector/calendar handlers
   - JWT_SECRET compromise ⇒ дешифровка EWS/AD паролей
   - Severity: CRITICAL
   - Fix: убрать хранение паролей, SSO/OAuth/Kerberos или vault + short‑lived tokens

2) **IDOR (employees/tasks/calendar)**
   - handlers без authz контроля
   - Severity: CRITICAL
   - Fix: RBAC/ABAC проверка ownership

3) **XSS через {@html}**
   - frontend routes: confluence + mail
   - Severity: CRITICAL
   - Fix: HTML sanitization + CSP без unsafe‑inline/unsafe‑eval

4) **JWT/token leakage**
   - localStorage + WS token в URL
   - Severity: CRITICAL/HIGH
   - Fix: HttpOnly cookies, WS auth headers/subprotocol

5) **File uploads без проверки**
   - Severity: HIGH
   - Fix: MIME/size/AV

## InfraSec
1) **TLS verify disabled / sslmode=disable**
   - docker-compose.yml, backend/.env.example, connector-service README
   - Severity: HIGH
   - Fix: enforce TLS verify; sslmode=require/verify-full

2) **HTTP-only nginx config**
   - nginx.conf без TLS/headers
   - Severity: HIGH
   - Fix: production на nginx-ssl.conf

3) **CSP с unsafe-inline/unsafe-eval**
   - nginx-ssl.conf
   - Severity: HIGH
   - Fix: nonce/hash CSP

## Data
1) **Health endpoint раскрывает EWS URL**
   - backend/cmd/server/main.go (/health)
   - Severity: MEDIUM
   - Fix: убрать детали/сделать админским

---

# PART 4 — FRONTEND AUDIT

## [FRONTEND DEFECT] JWT в localStorage
**Файл:** frontend/src/lib/stores/auth.ts
**Риск:** XSS ⇒ полный захват сессии
**Severity:** CRITICAL
**Fix:** HttpOnly cookies + CSRF

## [FRONTEND DEFECT] WS token в URL
**Файл:** frontend/src/lib/api/client.ts
**Риск:** утечки через logs/referrers
**Severity:** HIGH
**Fix:** headers/subprotocol auth

## [FRONTEND DEFECT] XSS через {@html}
**Файл:** frontend/src/routes/confluence/+page.svelte, frontend/src/routes/mail/+page.svelte
**Риск:** JS execution → token theft
**Severity:** CRITICAL
**Fix:** sanitization (DOMPurify), CSP hardening

---

# PART 5 — QA / TESTABILITY

## Факты
- В repo отсутствуют unit/integration tests (_test.go отсутствуют).
- TESTING.md — placeholders, требует заполнения реальных команд.

## Критические тест‑кейсы
- AuthZ: доступ к чужим сущностям (negative tests).
- XSS: payloads в mail/confluence.
- File upload: size/MIME/AV.
- WS auth: token leakage regression.

---

# PART 6 — PRODUCT & OPS

## Что сломается первым
- Почта/календарь при проблемах с decrypt паролей и EWS.

## Где пользователи страдают
- XSS + localStorage ⇒ компрометации аккаунтов.

## Где саппорт утонет
- Нет формализованных RBAC правил → хаос в доступах.

## Где масштабирование сложно
- In‑memory rate limiter и WS hub без shared state.

---

# FINAL VERDICT
**❌ NOT READY FOR PRODUCTION**

## Top‑10 CRITICAL
1) Reversible EWS/AD password storage (AES)
2) IDOR on employees/tasks/calendar
3) XSS in Confluence {@html}
4) XSS in Mail {@html}
5) JWT in localStorage
6) WS token in URL
7) TLS verify disabled defaults
8) sslmode=disable
9) Connector API key in URL
10) File upload validation отсутствует

## Top‑10 HIGH
1) CSP unsafe-inline/unsafe-eval
2) HTTP‑only nginx config
3) Mail API accepts arbitrary credentials
4) State machine transitions not validated
5) In‑memory rate limit non‑scalable
6) Health endpoint info disclosure
7) MinIO default creds in compose defaults
8) No centralized input validation
9) No transactional boundaries for multi‑step ops
10) WS auth model weak

## Срочно (72 часа)
- убрать reversible password storage
- закрыть IDOR
- устранить XSS + localStorage токены
- включить TLS verify/sslmode

## До 1 месяца
- переработать WS auth
- CSP hardening
- transactions + validation

## Техдолг (разрешено после фиксов)
- улучшение /health
- Redis‑rate‑limit

---

## Open Questions
See questions deliverable.
