# HANDOFF: PM → Developer

## 1. Meta
- ID: 2026-01-26__PM__DEV__security-sprint
- Date created: 2026-01-26
- From role: PM
- To role: Developer
- Owner (responsible): Developer
- Status: Active
- Related initiative / epic: Security Audit Implementation

---

## 2. Goal (ЗАЧЕМ)
Реализовать ВСЕ критические рекомендации из Enterprise Audit в Sprint 10.
Успех: устранены все CRITICAL уязвимости, система готова к production.

---

## 3. Context (ЧТО НУЖНО ЗНАТЬ)
- `/ai_org/deliverables/analyst/2026-01-26__full-audit__spec.md` — полный аудит
- `/CONTEXT.md` — инварианты архитектуры
- `/ARCHITECTURE.md` — компоненты системы

---

## 4. Scope (ГРАНИЦЫ РАБОТ)

### Sprint 10: SECURITY FIXES (72 часа)

#### 10.1 CRITICAL: Убрать reversible password storage (AES)
**Файлы:**
- `backend/internal/utils/crypto.go`
- `connector/calendar handlers`
**Проблема:** JWT_SECRET compromise → дешифровка EWS/AD паролей
**Fix:** SSO/OAuth/Kerberos или vault + short-lived tokens

#### 10.2 CRITICAL: Закрыть IDOR на employees
**Файлы:** `backend/internal/handlers/employees.go`
**Проблема:** Нет проверки прав доступа по user_id/role
**Fix:** RBAC/ABAC проверки в handlers или middleware

#### 10.3 CRITICAL: Закрыть IDOR на календаре
**Файлы:** `backend/internal/handlers/calendar.go`
**Проблема:** Доступ к календарю по произвольному employee_id
**Fix:** Проверка ownership/manager scopes

#### 10.4 CRITICAL: Устранить XSS через {@html}
**Файлы:**
- `frontend/src/routes/confluence/+page.svelte`
- `frontend/src/routes/mail/+page.svelte`
**Проблема:** JS execution → token theft
**Fix:** DOMPurify sanitization + CSP hardening

#### 10.5 CRITICAL: Убрать JWT из localStorage
**Файлы:** `frontend/src/lib/stores/auth.ts`
**Проблема:** XSS → полный захват сессии
**Fix:** HttpOnly cookies + CSRF

#### 10.6 CRITICAL: Исправить WS token в URL
**Файлы:** `frontend/src/lib/api/client.ts`
**Проблема:** Утечки через logs/referrers
**Fix:** Headers/subprotocol auth

#### 10.7 CRITICAL: Включить TLS verify
**Файлы:** `docker-compose.yml`, `backend/.env.example`
**Проблема:** TLS verify disabled
**Fix:** enforce TLS verify; sslmode=require/verify-full

#### 10.8 CRITICAL: Включить sslmode для PostgreSQL
**Файлы:** DB connection strings
**Fix:** sslmode=require/verify-full

#### 10.9 CRITICAL: Убрать Connector API key из URL
**Файлы:** connector service
**Fix:** Headers auth

#### 10.10 CRITICAL: File upload validation
**Файлы:** `backend/internal/handlers/files.go`
**Проблема:** Нет MIME/size/AV контроля
**Fix:** Streaming, лимиты, whitelist, AV-сканирование

### Sprint 10 HIGH Priority (1 месяц)

#### 10.11 HIGH: CSP hardening
**Файлы:** `nginx-ssl.conf`
**Fix:** Убрать unsafe-inline/unsafe-eval, использовать nonce/hash

#### 10.12 HIGH: Mail API credentials binding
**Файлы:** `backend/internal/handlers/mail.go`
**Fix:** Привязка credentials к user_id

#### 10.13 HIGH: State machine validation
**Файлы:** `backend/internal/handlers/tasks.go`
**Fix:** Enforce transitions из STATE_MACHINES.md

#### 10.14 HIGH: Transaction boundaries
**Файлы:** `backend/internal/handlers/tasks.go`
**Fix:** Транзакции для multi-step операций

#### 10.15 HIGH: Health endpoint info disclosure
**Файлы:** `backend/cmd/server/main.go`
**Fix:** Убрать детали EWS URL

---

## 5. Acceptance Criteria
- [ ] Все CRITICAL уязвимости закрыты
- [ ] Тесты security проходят
- [ ] Нет regression в функционале
- [ ] Документация обновлена

---

## 6. Expected Outputs
- Исправленный код
- Обновлённый WORKPLAN.md
- Security test results

