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
