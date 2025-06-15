# CloudBridge Relay

CloudBridge Relay - это высокопроизводительный сервер для туннелирования TCP-соединений с поддержкой аутентификации и балансировки нагрузки.

## Основные возможности

- Туннелирование TCP-соединений
- Аутентификация через Django API
- Поддержка TLS/mTLS
- Балансировка нагрузки
- Мониторинг и метрики
- Структурированное логирование
- Автоматическое переподключение
- Ограничение скорости запросов
- Graceful shutdown

## Установка

### Требования

- Go 1.21 или выше
- Docker (опционально)
- Systemd (для установки как сервис)

### Установка из исходного кода

```bash
git clone https://github.com/maxlanies/cloudbridge-relay-installer.git
cd cloudbridge-relay-installer
go build -o relay cmd/relay/main.go
```

### Установка через Docker

```bash
docker pull maxlanies/cloudbridge-relay
```

## Конфигурация

### Основные параметры

```json
{
  "server": {
    "port": 8080,
    "max_tunnels": 1000,
    "timeout": "30s"
  },
  "django": {
    "api_endpoint": "http://localhost:8000/api",
    "jwt_secret": "your-secret-key",
    "rest_timeout": "5s",
    "rest_cache_ttl": "1m"
  },
  "security": {
    "rate_limit": {
      "requests_per_second": 100,
      "burst": 200
    }
  }
}
```

### Параметры командной строки

- `--config` - путь к файлу конфигурации (по умолчанию: config/config.json)
- `--workers` - количество рабочих горутин (по умолчанию: количество CPU * 2)
- `--timeout` - таймаут соединения (по умолчанию: 30s)
- `--rate-limit` - ограничение скорости запросов (по умолчанию: 100)
- `--rate-burst` - размер пакета для ограничения скорости (по умолчанию: 200)
- `--max-tunnels` - максимальное количество туннелей (по умолчанию: 1000)
- `--log-level` - уровень логирования (DEBUG, INFO, WARNING, ERROR)
- `--log-file` - путь к файлу логов (по умолчанию: stdout)

## Использование

### Запуск сервера

```bash
./relay --config config/config.json
```

### Запуск в Docker

```bash
docker run -d \
  -p 8080:8080 \
  -v /path/to/config:/app/config \
  maxlanies/cloudbridge-relay
```

### Установка как системный сервис

```bash
sudo ./install.sh
```

## Протокол

### Формат сообщений

Все сообщения передаются в формате JSON с 4-байтовым префиксом длины (BigEndian):

```go
type Message struct {
    Type    MessageType `json:"type"`
    Token   string     `json:"token"`
    Config  *ConfigMessage  `json:"config,omitempty"`
    Pong    *PongMessage   `json:"pong,omitempty"`
    Error   string     `json:"error,omitempty"`
}
```

### Типы сообщений

- `Register` - регистрация туннеля
- `Connect` - подключение к туннелю
- `Ping` - проверка активности
- `Pong` - ответ на ping
- `Config` - конфигурация туннеля
- `Error` - сообщение об ошибке

## Мониторинг

### Метрики Prometheus

- `relay_active_connections` - количество активных соединений
- `relay_rejected_connections` - количество отклоненных соединений
- `relay_connection_duration_seconds` - длительность соединений
- `relay_bytes_transferred` - количество переданных байт
- `relay_errors_total` - количество ошибок
- `relay_latency_seconds` - задержка обработки запросов

### Логирование

Сервер использует структурированное логирование в формате JSON:

```json
{
  "timestamp": "2024-03-14T12:00:00Z",
  "level": "INFO",
  "message": "Server started",
  "fields": {
    "port": 8080,
    "tls": true,
    "workers": 8
  }
}
```

## Отладка

Подробная информация по отладке доступна в [документации по отладке](docs/debugging.md).

## Безопасность

- Поддержка TLS/mTLS
- Валидация токенов через Django API
- Ограничение скорости запросов
- Проверка IP-адресов
- Graceful shutdown

## Масштабирование

- Балансировка нагрузки
- Мониторинг производительности
- Настраиваемые ограничения соединений
- Кэширование токенов

## Развертывание

- Поддержка Docker
- Интеграция с Systemd
- Скрипты установки
- Конфигурация через JSON

## Мониторинг

- Метрики Prometheus
- Проверки здоровья
- Логирование
- Отслеживание ошибок

## Лицензия

MIT

## Авторы

- Max Lanies

## Вклад в проект

1. Форкните репозиторий
2. Создайте ветку для ваших изменений
3. Внесите изменения
4. Отправьте pull request

## Поддержка

Если у вас возникли проблемы или есть вопросы, создайте issue в репозитории проекта.

## JWT секрет (RELAY_JWT_SECRET)

Для корректной работы relay-сервера, Django API и тест-клиента необходимо использовать один и тот же секрет для подписи и проверки JWT токенов.

**Текущий секрет:**
```
Yk7+4of0fCmcHY+p0QCZHonEMO0+WUvbKsrjfvJL3zk=
```

### Где используется секрет

- **Django:**
  - В настройках (`settings.py`) как `RELAY_JWT_SECRET`.
  - Используется для проверки токенов во view `/api/relay/token/validate/`.

- **Relay-сервер:**
  - В файле `config/config.json` поле `jwt_secret`.
  - Используется для валидации токенов, получаемых от клиентов.

- **Тест-клиент:**
  - В файле `test_client.py` переменная `secret`.
  - Используется для генерации токенов, отправляемых на relay-сервер.

### Важно
- Если вы меняете секрет, обязательно обновите его во всех трёх местах.
- Несовпадение секрета приведёт к ошибкам валидации токена ("invalid token").