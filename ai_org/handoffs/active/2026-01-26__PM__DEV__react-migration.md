# HANDOFF: PM → Developer

## 1. Meta
- ID: 2026-01-26__PM__DEV__react-migration
- Date created: 2026-01-26
- From role: PM
- To role: Developer
- Owner (responsible): Developer
- Status: Active
- Related initiative / epic: Frontend Stack Replacement

---

## 2. Goal (ЗАЧЕМ)
Полная замена SvelteKit 2 + Svelte 5 на React.
Успех: полностью переписанный frontend на React с сохранением всего функционала.

---

## 3. Context (ЧТО НУЖНО ЗНАТЬ)
- Текущий стек: SvelteKit 2 + Svelte 5 + Tailwind CSS
- Целевой стек: React 18+ + Vite + Tailwind CSS
- Дизайн: Figma reference (EKF Hub design system)
- Backend API: REST + WebSocket (без изменений)

---

## 4. Scope (ГРАНИЦЫ РАБОТ)

### Sprint 11: REACT MIGRATION

#### Фаза 1: Инфраструктура ✅ ЗАВЕРШЕНА
- [x] Создать новый React проект (Vite + React 18)
- [x] Настроить Tailwind CSS с EKF дизайн-системой
- [x] Настроить React Router
- [x] Настроить state management (Zustand)
- [x] Настроить API client (fetch wrapper)
- [x] Настроить WebSocket client

#### Фаза 2: Компоненты дизайн-системы ✅ ЗАВЕРШЕНА
- [x] Layout components (Sidebar, Header, MainContent)
- [x] Navigation components
- [x] Form components (Input, Select, Button, Textarea)
- [x] Card components (Card, CardHeader, CardTitle, CardContent, CardFooter)
- [x] Modal/Dialog components (Modal, ConfirmDialog)
- [x] Table components (Table, TableHeader, TableBody, TableRow, etc.)
- [x] Badge components (Badge, StatusBadge, PriorityBadge)
- [x] Avatar components (Avatar, AvatarGroup)
- [ ] Toast/Notification components

#### Фаза 3: Страницы (по приоритету)
1. [x] Auth (Login) ✅
2. [x] Dashboard ✅
3. [x] Tasks (Kanban + List views) ✅
4. [ ] Employees
5. [ ] Messenger
6. [ ] Mail
7. [ ] Calendar
8. [ ] Meetings
9. [ ] Service Desk (Portal + Agent)
10. [ ] Confluence
11. [ ] Analytics
12. [ ] Admin
13. [ ] Releases
14. [ ] Settings

#### Фаза 4: Интеграции
- [ ] WebSocket real-time updates
- [ ] File upload/download
- [ ] Rich text editor (TipTap for React)
- [ ] Calendar drag-and-drop
- [ ] Kanban drag-and-drop

#### Фаза 5: Testing & Deploy
- [ ] Unit tests (Vitest)
- [ ] E2E tests (Playwright)
- [ ] Docker build
- [ ] nginx configuration update

---

## 5. Design Reference
- Figma: https://www.figma.com/make/PqSbgFYW2fsgAe6GiCeiP0/Complete-Design-Mockup
- EKF Colors: ekf-red (#E53935), ekf-dark (#1a1a2e)

---

## 6. Acceptance Criteria
- [ ] Все страницы переписаны на React
- [ ] Функционал 1:1 с Svelte версией
- [ ] Дизайн соответствует Figma макетам
- [ ] Build проходит без ошибок
- [ ] Tests проходят
- [ ] Deployed на production

---

## 7. Expected Outputs
- `/frontend-react/` — новый React проект
- Обновлённый docker-compose.yml
- Обновлённая документация

