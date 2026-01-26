# Handoff: Mail Auth Persistence Fix

**Дата:** 2026-01-26
**От:** PM
**Кому:** Developer
**Приоритет:** HIGH

## Проблема

Почта требует повторную авторизацию при каждом посещении страницы или после закрытия вкладки.

## Причина

Credentials хранились в `sessionStorage`, который очищается при закрытии вкладки браузера.

## Решение

1. Добавлена опция "Запомнить меня" в форму входа
2. При включённой опции credentials сохраняются в `localStorage` (persistent)
3. При выключённой опции credentials сохраняются в `sessionStorage` (temporary)
4. onMount проверяет сначала localStorage, потом sessionStorage
5. Logout очищает оба хранилища

## Файлы

- `frontend/src/routes/mail/+page.svelte` (MODIFIED)
- `frontend/src/routes/mail/[id]/+page.svelte` (MODIFIED)

## Изменения

### State
```typescript
let rememberMe = $state(true); // Remember credentials by default
```

### onMount
```typescript
// First try localStorage (persistent)
const localCreds = localStorage.getItem('ews_credentials');
if (localCreds) {
  credentials = JSON.parse(localCreds);
  // ...
}
// Then try sessionStorage (temporary)
const sessionCreds = sessionStorage.getItem('ews_credentials');
```

### handleLogin
```typescript
if (rememberMe) {
  localStorage.setItem('ews_credentials', JSON.stringify(credentials));
  sessionStorage.removeItem('ews_credentials');
} else {
  sessionStorage.setItem('ews_credentials', JSON.stringify(credentials));
  localStorage.removeItem('ews_credentials');
}
```

### logout
```typescript
localStorage.removeItem('ews_credentials');
sessionStorage.removeItem('ews_credentials');
```

## Статус

✅ DONE
