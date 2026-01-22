# AI Org (мульти-агентная команда)

Это “операционная система” AI-команды для проекта.

Принципы:
- Один главный агент: Product Manager (PM).
- Все остальные роли работают через артефакты (md-файлы).
- Источник истины по проекту — корневые файлы репозитория (CONTEXT.md / ARCHITECTURE.md / WORKPLAN.md / DECISIONS.md / DOMAIN_GLOSSARY.md / STATE_MACHINES.md и т.п.).
- Любая работа должна оставлять след в: ai_org/logs/ + обновлениях state/ и корневых артефактах.

Точка входа: ai_org/state/INDEX.md
