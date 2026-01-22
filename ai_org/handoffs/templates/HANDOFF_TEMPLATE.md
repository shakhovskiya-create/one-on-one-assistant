# HANDOFF: <FROM_ROLE> → <TO_ROLE>
<!--
Пример имени файла:
YYYY-MM-DD__PM__ANALYST__auth-flow.md
-->

## 1. Meta
- ID: <auto / YYYY-MM-DD__FROM__TO__slug>
- Date created: YYYY-MM-DD
- From role: <PM / ANALYST / DESIGNER / QA / DEV / DEVOPS / RELEASE>
- To role: <PM / ANALYST / DESIGNER / QA / DEV / DEVOPS / RELEASE>
- Owner (responsible): <role>
- Status: Draft | Active | Blocked | Completed | Accepted
- Related initiative / epic: <link or short name>

---

## 2. Goal (ЗАЧЕМ)
Кратко и однозначно:
- какую проблему решаем
- какой результат считается успехом
- зачем это бизнесу / продукту

> ❗ Один handoff = одна цель.  
> Если целей больше — разбей.

---

## 3. Context (ЧТО НУЖНО ЗНАТЬ)
Ссылки на **источники истины** и обязательный контекст:

- `/CONTEXT.md` — релевантные инварианты
- `/ARCHITECTURE.md` — затрагиваемые части
- `/WORKPLAN.md` — пункт плана (если есть)
- `/DECISIONS.md` — связанные решения
- `/DOMAIN_GLOSSARY.md` — термины
- `/STATE_MACHINES.md` — состояния / бизнес-логика
- Другие файлы/PR/Issues: …

> ❗ Если чего-то **нет в источниках истины** — это считается неизвестным.

---

## 4. Scope (ГРАНИЦЫ РАБОТ)
### In scope
- что именно нужно сделать
- какие артефакты обновить / создать

### Out of scope
- что **делать нельзя**
- какие зоны не трогать

---

## 5. Requirements / Tasks
Чёткий список того, что ожидается от роли-исполнителя.

Пример:
- [ ] Проанализировать текущий флоу X
- [ ] Подготовить PRD / код / UX-flow / тест-кейсы
- [ ] Зафиксировать допущения и вопросы
- [ ] Обновить связанные документы (если требуется)

---

## 6. Acceptance Criteria (КРИТЕРИИ ПРИЁМКИ)
Проверяемые условия, без воды.

Пример:
- [ ] Все требования описаны и не противоречат `CONTEXT.md`
- [ ] Нет незафиксированных допущений
- [ ] Обновлены необходимые md-файлы
- [ ] Нет изменений вне scope

> ❗ Если критерий нельзя проверить — он плохой.

---

## 7. Constraints & Invariants (ОГРАНИЧЕНИЯ)
Обязательные правила, которые **нельзя нарушать**:

- архитектурные инварианты
- security-ограничения
- performance / UX ограничения
- запреты из `AI_SAFE_CHECKLIST.md`

---

## 8. Risks, Assumptions, Open Questions
### Risks
- …

### Assumptions
- …

### Open questions (обязательно задать PM, если есть)
- …

> ❗ Если есть вопросы — **не додумывать**, а зафиксировать здесь.

---

## 9. Expected Outputs (АРТЕФАКТЫ НА ВЫХОДЕ)
Что должно появиться в репозитории после выполнения:

- файлы / директории
- PR / commit
- обновлённые документы
- ссылки

---

## 10. Result (ЗАПОЛНЯЕТ ИСПОЛНИТЕЛЬ)
Что фактически сделано:

- …
- ссылки на изменённые файлы / PR
- отклонения от исходных требований (если были)

Дата выполнения: YYYY-MM-DD

---

## 11. Acceptance (ЗАПОЛНЯЕТ PM)
- Status: Accepted | Rejected | Needs changes
- Accepted by: PM
- Date: YYYY-MM-DD
- Notes:
  - …

> ❗ Только после заполнения этого блока handoff можно переносить в `done/`.

---

## 12. Follow-ups / Next steps
Если из этого handoff вытекают новые задачи:

- [ ] Новый handoff: <описание / ссылка>
- [ ] Обновить WORKPLAN.md
- [ ] Зафиксировать решение в DECISIONS.md

---

## 13. Log reference
Ссылка на запись в:
- `ai_org/logs/AGENT_LOG.md`: <link>

- Linked AGENT_LOG entry: <link>