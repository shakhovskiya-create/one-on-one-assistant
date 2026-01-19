# EKF One-on-One AD Connector

Коннектор для интеграции с Active Directory и Exchange Web Services. Работает как сервис внутри корпоративной сети и подключается к облачному бэкенду через WebSocket.

## Возможности

- Подключение к Railway backend через защищенный WebSocket
- Получение календаря из Exchange (EWS)
- Синхронизация календаря
- Поиск свободных слотов
- Автоматическое переподключение при разрыве связи
- Работает как systemd service с автозапуском

## Требования

- Go 1.21 или выше
- Доступ к корпоративной сети (OpenVPN)
- Доступ к Exchange Server (post.ekf.su)
- Linux сервер с systemd

## Установка

### 1. Скачать и собрать

```bash
git clone https://github.com/your-repo/one-on-one-assistant.git
cd one-on-one-assistant/connector-service
```

### 2. Настроить .env

```bash
cp .env.example .env
nano .env
```

Заполните переменные:
```env
BACKEND_WS_URL=wss://one-on-one-back-production.up.railway.app/ws/connector
CONNECTOR_API_KEY=ваш-api-ключ

EWS_URL=https://post.ekf.su/EWS/Exchange.asmx
EWS_DOMAIN=ekfgroup
EWS_USERNAME=domain\service-account
EWS_PASSWORD=пароль-сервисного-аккаунта
EWS_SKIP_TLS_VERIFY=true
```

### 3. Установить как сервис

```bash
sudo ./install.sh
```

### 4. Запустить

```bash
sudo systemctl start ekf-connector
sudo systemctl status ekf-connector
```

### 5. Просмотр логов

```bash
# Реал-тайм логи
sudo journalctl -u ekf-connector -f

# Последние 100 строк
sudo journalctl -u ekf-connector -n 100
```

## Управление сервисом

```bash
# Запустить
sudo systemctl start ekf-connector

# Остановить
sudo systemctl stop ekf-connector

# Перезапустить
sudo systemctl restart ekf-connector

# Статус
sudo systemctl status ekf-connector

# Включить автозапуск
sudo systemctl enable ekf-connector

# Выключить автозапуск
sudo systemctl disable ekf-connector
```

## Архитектура

```
Railway Backend (облако)
    ↕ WebSocket (wss://)
AD Connector (внутри сети)
    ↕ EWS (https://)
Exchange Server (post.ekf.su)
```

## Поддерживаемые команды

- `ping` - Проверка связи
- `get_calendar` - Получить календарь пользователя
- `sync_calendar` - Синхронизировать календарь
- `find_free_slots` - Найти свободные слоты

## Безопасность

- API ключ для аутентификации
- TLS для WebSocket соединения
- Сервис работает от непривилегированного пользователя
- Минимальные права доступа (systemd hardening)

## Troubleshooting

### Коннектор не подключается к бэкенду

1. Проверьте URL и API ключ в .env
2. Проверьте доступность бэкенда: `curl -I https://one-on-one-back-production.up.railway.app`
3. Проверьте логи: `sudo journalctl -u ekf-connector -f`

### Не удается подключиться к Exchange

1. Проверьте доступность: `curl -I https://post.ekf.su/EWS/Exchange.asmx`
2. Проверьте credentials в .env
3. Убедитесь, что OpenVPN работает

### Сервис падает при запуске

1. Проверьте логи: `sudo journalctl -u ekf-connector -xe`
2. Проверьте права на файлы: `ls -la /opt/ekf-connector`
3. Проверьте конфиг: `/opt/ekf-connector/connector -config /opt/ekf-connector/config.yaml`
