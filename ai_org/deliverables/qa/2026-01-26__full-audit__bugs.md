# Full Audit — QA Bugs (Summary)

**Last updated:** 2026-01-26

## CRITICAL
1) Reversible password storage (AES-GCM) for EWS/AD → полный компромисс паролей при утечке JWT_SECRET.
2) IDOR на employees/tasks/calendar → чтение/изменение чужих сущностей.
3) XSS через {@html} в Confluence/Mail → кража токенов.
4) JWT в localStorage + WS token в URL → утечки через XSS/логи.

## HIGH
1) TLS verify disabled/sslmode=disable по умолчанию.
2) File upload без валидации MIME/size.
3) CSP unsafe-inline/unsafe-eval.
4) Нет валидации state transitions задач.

## Notes
Полный перечень и контекст — в аналитическом отчете (spec).
