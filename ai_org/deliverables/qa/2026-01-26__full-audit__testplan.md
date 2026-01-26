# Full Audit — QA Test Plan

## Scope
- Static review of backend/frontend/infra configs.
- No runtime execution in this audit.

## Risks
- Отсутствие тестов → регрессии не обнаруживаются.
- Нет observability → инциденты трудно расследовать.

## Planned Checks (not executed)
- go test ./...
- go vet ./...
- go test -race ./...
- npm run lint
- npm run build
- docker compose build
- curl /health

## Notes
В рамках аудита тесты не выполнялись (нет цели изменения кода, отсутствуют окружения/зависимости).
