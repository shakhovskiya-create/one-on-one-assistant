# AUTOPILOT Protocol (PM)

При любом входящем сообщении пользователя:

1) Определи тип запроса:
   - Bug/Issue
   - Feature/Enhancement
   - Refactor/Tech debt
   - Research/Analysis request
   - Release/Deploy request

2) Выбери slug (2-4 слова, snake_case).

3) Создай/обнови артефакты:
   - handoff: ai_org/handoffs/active/YYYY-MM-DD__PM__<ROLE>__<slug>.md
   - STATUS: ai_org/state/STATUS.md
   - AGENT_LOG: ai_org/logs/AGENT_LOG.md
   - deliverables:
     * Analyst: ai_org/deliverables/analyst/YYYY-MM-DD__<slug>__questions.md
              ai_org/deliverables/analyst/YYYY-MM-DD__<slug>__fm.md
              ai_org/deliverables/analyst/YYYY-MM-DD__<slug>__spec.md
     * QA:     ai_org/deliverables/qa/YYYY-MM-DD__<slug>__testplan.md
              ai_org/deliverables/qa/YYYY-MM-DD__<slug>__results.md
              ai_org/deliverables/qa/YYYY-MM-DD__<slug>__bugs.md
     * DevOps: ai_org/deliverables/devops/YYYY-MM-DD__<slug>__plan.md
              ai_org/deliverables/devops/YYYY-MM-DD__<slug>__runlog.md
              ai_org/deliverables/devops/YYYY-MM-DD__<slug>__rollback.md

4) Заполнение по этапам:
   - Сразу заполни questions.md (Analyst) и задай вопросы пользователю.
   - Пока нет ответов — FM и Spec заполняй черновиком + пометь OPEN QUESTIONS.
   - После ответов обнови FM/Spec и только затем инициируй Dev/QA/DevOps.

5) Не "закрывай" handoff без:
   - Result (исполнитель)
   - Acceptance (PM)
   - записи в AGENT_LOG
   - перемещения handoff в done/
