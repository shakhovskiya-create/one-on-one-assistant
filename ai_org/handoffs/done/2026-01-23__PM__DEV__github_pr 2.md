# HANDOFF: PM → DEV

## 1. Meta
- ID: 2026-01-23__PM__DEV__github_pr
- Date created: 2026-01-23
- From role: PM
- To role: DEV
- Owner (responsible): DEV
- Status: Completed
- Related initiative / epic: JIRA-like functionality - GitHub integration

---

## 2. Goal (ЗАЧЕМ)111
Добавить отображение Pull Requests в модальном окне задачи, аналогично коммитам.

---

## 3. Context (ЧТО НУЖНО ЗНАТЬ)
- Уже есть интеграция с GitHub для отображения коммитов
- Backend имеет GetPullRequests метод в client.go
- Нужно добавить поиск PR по task ID

---

## 4. Scope (ГРАНИЦЫ РАБОТ)
### In scope
- Backend: метод поиска PR по task ID
- Backend: handler для endpoint
- Frontend: API метод
- Frontend: UI секция в модальном окне задачи

### Out of scope
- Создание PR из интерфейса
- Детальный просмотр PR

---

## 5. Requirements / Tasks
- [x] Добавить SearchPullRequestsForTaskID в client.go
- [x] Добавить GetTaskPullRequests handler
- [x] Добавить route в main.go
- [x] Добавить getTaskPullRequests в frontend API
- [x] Добавить UI секцию для PRs

---

## 6. Acceptance Criteria (КРИТЕРИИ ПРИЁМКИ)
- [x] PRs загружаются параллельно с коммитами
- [x] Отображается номер, статус, название, автор
- [x] Цветовая индикация статуса работает
- [x] Ссылка на PR открывается в новой вкладке

---

## 10. Result (ЗАПОЛНЯЕТ ИСПОЛНИТЕЛЬ)
Реализовано:
- Backend: SearchPullRequestsForTaskID, GetTaskPullRequests
- Route: GET /github/tasks/:id/pulls
- Frontend API: github.getTaskPullRequests()
- UI: секция "Связанные Pull Requests" с цветовыми индикаторами статуса

Дата выполнения: 2026-01-23

---

## 11. Acceptance (ЗАПОЛНЯЕТ PM)
- Status: Accepted
- Accepted by: PM
- Date: 2026-01-23

---

## 13. Log reference
- `ai_org/logs/AGENT_LOG.md`: GitHub: Pull Requests в задачах
