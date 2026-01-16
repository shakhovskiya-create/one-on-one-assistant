# Camunda BPMN для EKF Team Hub

## Деплой на Railway

### 1. Создать новый сервис в Railway

```bash
# В Railway Dashboard:
# 1. New Project → Deploy from GitHub repo
# 2. Выбрать этот репозиторий
# 3. Указать Root Directory: camunda
# 4. Deploy
```

### 2. Настроить переменные окружения

В Railway Dashboard для сервиса Camunda:
- `PORT` = 8080 (автоматически)

### 3. Настроить backend

В Railway Dashboard для backend сервиса добавить:
- `CAMUNDA_URL` = https://your-camunda-service.railway.app
- `CAMUNDA_USER` = demo (или ваш пользователь)
- `CAMUNDA_PASSWORD` = demo (или ваш пароль)

## Доступ к Camunda

После деплоя:
- Web UI: https://your-camunda-service.railway.app/camunda
- REST API: https://your-camunda-service.railway.app/engine-rest

Логин по умолчанию: demo / demo

## Процессы

- `task-approval.bpmn` - Согласование задачи
