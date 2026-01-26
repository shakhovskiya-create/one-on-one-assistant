# DECISIONS.md
# EKF Hub — Журнал решений (память об изменениях)

Назначение: фиксировать важные решения, которые AI и люди должны помнить.
Это НЕ changelog коммитов. Это журнал "почему мы сделали именно так".

Правило: любое решение, влияющее на безопасность, архитектуру, домен, контракты API,
инфраструктуру, данные — фиксируем здесь.

Формат записи:

## YYYY-MM-DD — <короткое название решения>
Контекст:
- почему возникло

Решение:
- что решили

Последствия:
- что улучшилось
- какие риски/компромиссы

Связанные изменения:
- файлы/коммиты/релиз (если есть)

---

## 2026-01-26 — CRITICAL: Системное расхождение макетов и реализации

Контекст:
- Владелец продукта провёл ревью всех разделов EKF Hub
- Выявлены КРИТИЧЕСКИЕ расхождения между согласованными макетами и фактической реализацией
- Проблема носит СИСТЕМНЫЙ характер — затронуты ВСЕ разделы

Зафиксированные проблемы (12 GAP):

**CRITICAL (блокируют работу):**
1. GAP-005: Сущность "Заявка на улучшение" — НЕ РЕАЛИЗОВАНА (был согласован полный workflow)
2. GAP-006: Планирование ресурсов — НЕ РЕАЛИЗОВАНО (загрузка исполнителей)
3. GAP-007: Зависимости задач — БАГ, UI зависает
4. GAP-010: Service Desk — НЕ РЕАЛИЗОВАН (был в макетах)
5. GAP-001: Dashboard sidebar — элементы в неправильном месте (должны быть в top-bar)

**HIGH:**
6. GAP-002: GitHub отсутствует в Tasks sidebar (есть в Dashboard где не нужен)
7. GAP-003: Нет обязательной связи задача→проект
8. GAP-008: Meetings НЕ соответствует макету
9. GAP-009: Tasks sidebar не содержит: тестирование, документация, roadmap

**MEDIUM/LOW:**
10. GAP-004: Нет документов в проектах
11. GAP-011: Нет SSO для почты
12. GAP-012: Профиль — дублирование, неполное ФИО

Причина проблемы:
- Отсутствует формальный процесс приёмки макетов
- Отсутствует контроль соответствия реализации ТЗ
- Нет единого источника истины для требований

Решение:
1. Зафиксировать все расхождения в `ai_org/deliverables/analyst/2026-01-26__critical-gaps__spec.md`
2. Создать вопросы к требованиям в `__questions.md`
3. Создать целевую функциональную модель в `__fm.md`
4. Обновить WORKPLAN.md с недостающими сущностями
5. Ввести обязательную верификацию UI соответствия макетам

Последствия:
- Приостановить текущую разработку до утверждения плана исправлений
- Приоритизировать CRITICAL дефекты
- Требуется ответ владельца продукта на вопросы в `__questions.md`

Связанные артефакты:
- `ai_org/handoffs/active/2026-01-26__PM__ANALYST__critical-gaps-analysis.md`
- `ai_org/deliverables/analyst/2026-01-26__critical-gaps__questions.md`
- `ai_org/deliverables/analyst/2026-01-26__critical-gaps__fm.md`
- `ai_org/deliverables/analyst/2026-01-26__critical-gaps__spec.md`

---

## 2026-01-21 — Security Audit Critical Fixes
Контекст:
- Principal Engineer аудит выявил 10 CRITICAL и 10 HIGH уязвимостей
- Основные проблемы: хранение паролей, слабый RNG, отсутствие rate limiting, debug info в ответах

Решение:
1. **Удалено хранение пароля в sessionStorage** (frontend/src/lib/stores/auth.ts)
   - Убран вызов sessionStorage.setItem для EWS credentials

2. **Удалены query params для credentials** (backend/internal/handlers/connector.go)
   - Credentials принимаются ТОЛЬКО из body, не из URL

3. **AD_SKIP_VERIFY теперь false по умолчанию** (backend/internal/config/config.go)
   - Защита от MITM при TLS соединении к AD

4. **JWT_SECRET требует явной установки** (backend/internal/config/config.go)
   - Паника при старте если не установлен (no default)

5. **Rate limiting добавлен** (backend/internal/middleware/ratelimit.go)
   - Auth endpoint: 5 req/min per IP
   - General API: 100 req/min per IP

6. **CSRF protection добавлена** (backend/internal/middleware/csrf.go)
   - Токен валидация для state-changing requests
   - Endpoint GET /api/v1/csrf-token для получения токена

7. **crypto/rand вместо math/rand** (backend/internal/handlers/connector.go)
   - Криптографически стойкий генератор для токенов

8. **PostgreSQL sslmode=prefer** (docker-compose.yml, docker-compose-ssl.yml)
   - TLS при доступности

9. **Удалена debug info из ответов** (backend/internal/handlers/connector.go)
   - Убраны поля domain/url из auth ответов

Последствия:
- CRITICAL уязвимости закрыты
- Требуется пересборка backend
- Frontend должен отправлять X-CSRF-Token header на state-changing requests
- WebSocket auth через URL params - известное ограничение браузеров (документировано)

Связанные изменения:
- SECURITY_AUDIT_2026-01-21.md - полный отчёт аудита
- backend/internal/middleware/ratelimit.go - NEW
- backend/internal/middleware/csrf.go - NEW
- backend/cmd/server/main.go - интеграция middleware

---

## 2026-01-21 — Production hardening baseline
Контекст:
- аудит выявил критичные риски: порты наружу, secrets, TLS verify, отсутствие nginx entrypoint.

Решение:
- включили nginx как единую точку входа,
- закрыли прямые порты frontend/backend/postgres,
- вынесли secrets в env,
- подготовили SSL конфиг,
- включили health checks и бэкапы,
- добавили Redis как infra-компонент (кэш).

Последствия:
- рост безопасности и управляемости,
- часть задач требует пересборки backend/frontend (валидация, structured logging, redis cache).

Связанные изменения:
- см. IMPROVEMENTS.md
