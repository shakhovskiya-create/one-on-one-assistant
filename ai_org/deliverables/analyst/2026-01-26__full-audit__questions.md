# Full Audit — Questions (Analyst)

## Open Questions
1) Где хранятся production secrets (Vault/SSM/KMS)? В репозитории не описано. (MISSING)
2) Есть ли выделенный audit logging / SIEM / log pipeline? В repo нет конфигурации. (MISSING)
3) Есть ли утвержденные RBAC политики (роль/права по сущностям)? В коде нет явной модели. (MISSING)
4) Есть ли требования по хранению и ротации EWS/AD credentials? (MISSING)
5) Есть ли требования по retention/PII masking? (MISSING)

## Assumptions
- Репозиторий отражает production runtime конфигурацию.
- Внешние сервисы (AD/Exchange/Confluence/GitHub) доступны и настроены вне кода.

## Risks
- Без доступа к production runtime нельзя подтвердить реальный security posture (headers, TLS, rate limit, WAF).
- Нельзя проверить фактические RPO/RTO и восстановление бэкапов.
