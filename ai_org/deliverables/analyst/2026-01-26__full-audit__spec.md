# Full Audit — Specification (Enterprise)

## Objective
Полный аудит EKF Hub по архитектуре, backend, security, frontend, QA/testability, product/ops readiness.

## Deliverables
- Detailed audit report with evidence in code/configs.
- Risk register: CRITICAL/HIGH/MEDIUM/LOW.
- Verdict on production readiness.

## Non-goals
- Исправления в коде.
- Деплой/миграции.

## Evidence Sources
- Backend: handlers, middleware, config, database layer, services.
- Frontend: auth store, API client, routes (mail/confluence/messenger).
- Infra: docker-compose, nginx configs.
- Docs: README/CONTEXT/ARCHITECTURE/SECURITY_AUDIT.

## Output Format
- Architecture issues
- Backend defects
- Security vulnerabilities (AppSec/InfraSec/Data)
- Frontend defects
- QA/Testability
- Product/Operations
- Final verdict + Top-10 CRITICAL/HIGH

## Open Questions
See questions deliverable.
