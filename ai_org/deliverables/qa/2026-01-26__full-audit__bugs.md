# Full Audit — QA Bugs (Summary)

## Critical/High (see audit report for details)
- Reversible storage of user passwords (AES) for EWS access.
- JWT tokens and WS tokens exposed via localStorage/URL query params.
- Отсутствует централизованный RBAC/ABAC на уровне handlers (IDOR).
- Input validation непоследовательна; нет обязательной схемы валидации.
- TLS/db sslmode=disable в docker-compose.

## Notes
Полный перечень — в детальном аудит-отчете (в сообщении PM/Analyst).
