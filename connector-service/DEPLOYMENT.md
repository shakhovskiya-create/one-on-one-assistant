# Развертывание Коннектора

## Кратко

Коннектор - это Go-сервис, который работает внутри корпоративной сети (с доступом через OpenVPN) и подключается к облачному бэкенду Railway через WebSocket.

## Зачем нужен?

Exchange Server (`post.ekf.su`) недоступен из интернета. Коннектор решает это:

```
Internet → Railway Backend (облако)
               ↕ WebSocket (wss://)
          AD Connector (внутри сети)
               ↕ EWS/NTLM
          Exchange Server (post.ekf.su)
```

## Быстрый старт

### 1. На сервере внутри сети

```bash
# Клонировать репозиторий
git clone https://github.com/shakhovskiya-create/one-on-one-assistant.git
cd one-on-one-assistant/connector-service

# Создать .env
cp .env.example .env
nano .env
```

### 2. Настроить .env

```env
# URL бэкенда (Railway)
BACKEND_WS_URL=wss://one-on-one-back-production.up.railway.app/ws/connector

# API ключ (получить у администратора или создать в Railway)
CONNECTOR_API_KEY=ваш-секретный-ключ

# Exchange
EWS_URL=https://post.ekf.su/EWS/Exchange.asmx
EWS_DOMAIN=ekfgroup
EWS_USERNAME=ekfgroup\service-account
EWS_PASSWORD=пароль-сервисного-аккаунта
EWS_SKIP_TLS_VERIFY=true
```

### 3. Установить

```bash
sudo ./install.sh
```

### 4. Проверить

```bash
sudo systemctl status ekf-connector
sudo journalctl -u ekf-connector -f
```

## Настройка API ключа на Railway

API ключ нужен для безопасного подключения коннектора к бэкенду.

### Вариант 1: Использовать существующий

Если в `.env` на Railway уже есть `CONNECTOR_API_KEY`, используйте его.

### Вариант 2: Создать новый

1. Сгенерировать ключ:
```bash
openssl rand -hex 32
```

2. Добавить в Railway:
   - Открыть проект в Railway
   - Выбрать сервис `one-on-one-back`
   - Variables → Add Variable
   - `CONNECTOR_API_KEY` = `ваш-сгенерированный-ключ`
   - Deploy

3. Использовать этот же ключ в `.env` коннектора

## Проверка работы

### 1. Проверить подключение к бэкенду

```bash
# Логи коннектора
sudo journalctl -u ekf-connector -n 100

# Должно быть:
# "Connected to backend: wss://..."
# "Heartbeat sent"
```

### 2. Проверить доступность через бэкенд

Из браузера или Postman:
```
GET https://one-on-one-back-production.up.railway.app/api/v1/connector/status
```

Ответ должен быть:
```json
{
  "connected": true,
  "ad_status": "connected",
  ...
}
```

### 3. Проверить календарь

Авторизоваться в приложении и попробовать синхронизировать календарь.

## Troubleshooting

### Коннектор не подключается к бэкенду

```bash
# Проверить доступность
curl -I https://one-on-one-back-production.up.railway.app

# Проверить логи
sudo journalctl -u ekf-connector -xe
```

### Не удается получить календарь

```bash
# Проверить доступность Exchange
curl -I https://post.ekf.su/EWS/Exchange.asmx

# Убедиться, что OpenVPN работает
ip addr show tun0
```

### Сервис падает при запуске

```bash
# Проверить конфиг
/opt/ekf-connector/connector -config /opt/ekf-connector/config.yaml

# Проверить права
ls -la /opt/ekf-connector
```

## Обновление

```bash
cd one-on-one-assistant
git pull
cd connector-service
sudo systemctl stop ekf-connector
go build -o /opt/ekf-connector/connector ./cmd/connector
sudo systemctl start ekf-connector
sudo systemctl status ekf-connector
```

## Команды управления

```bash
# Запустить
sudo systemctl start ekf-connector

# Остановить
sudo systemctl stop ekf-connector

# Перезапустить
sudo systemctl restart ekf-connector

# Статус
sudo systemctl status ekf-connector

# Логи реал-тайм
sudo journalctl -u ekf-connector -f

# Последние 100 строк
sudo journalctl -u ekf-connector -n 100

# Включить автозапуск
sudo systemctl enable ekf-connector

# Выключить автозапуск
sudo systemctl disable ekf-connector
```
