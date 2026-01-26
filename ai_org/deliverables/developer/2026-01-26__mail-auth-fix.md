# Mail Auth Persistence Fix

**Дата:** 2026-01-26
**Статус:** ✅ DONE

## Summary

Исправлена проблема повторной авторизации в почте. Добавлена опция "Запомнить меня" для сохранения credentials в localStorage.

## Changes

### File: `frontend/src/routes/mail/+page.svelte`

1. **State**: Добавлен `rememberMe = $state(true)`
2. **onMount**: Проверяет localStorage, затем sessionStorage
3. **handleLogin**: Сохраняет в localStorage или sessionStorage в зависимости от rememberMe
4. **logout**: Очищает оба хранилища
5. **UI**: Добавлен checkbox "Запомнить меня" в форму входа

### File: `frontend/src/routes/mail/[id]/+page.svelte`

1. **onMount**: Проверяет localStorage первым, затем sessionStorage

## Before/After

### Before
- Credentials в sessionStorage
- Теряются при закрытии вкладки
- Нет опции сохранения

### After
- По умолчанию rememberMe = true
- Credentials в localStorage (persistent)
- Опция выбора для пользователя
- Credentials сохраняются между сессиями браузера

## Testing

- [x] Вход с "Запомнить меня" включённым
- [x] Вход с "Запомнить меня" выключённым
- [x] Logout очищает credentials
- [x] Переход на mail/[id] работает с localStorage
