# CloudBridge Relay

CloudBridge Relay - это высокопроизводительный сервер для туннелирования TCP-соединений с поддержкой аутентификации, балансировки нагрузки и расширенного мониторинга.

## 🚀 Основные возможности

- ✅ **Туннелирование TCP-соединений** с поддержкой TLS/mTLS
- ✅ **JWT аутентификация** через Django API
- ✅ **Интеграция с Keycloak** (OAuth 2.0 / OpenID Connect)
- ✅ **Балансировка нагрузки** (round-robin)
- ✅ **Ограничение скорости запросов** (rate limiting)
- ✅ **Расширенный мониторинг** через Prometheus + Grafana
- ✅ **Алерты и уведомления** через AlertManager
- ✅ **Структурированное логирование** (JSON) с ротацией
- ✅ **Автоматическое переподключение**
- ✅ **Graceful shutdown**
- ✅ **Health checks**
- ✅ **Полное тестовое покрытие** (unit, integration, E2E)

## 🔒 Безопасность

### TLS/mTLS поддержка
- TLS 1.2+ по умолчанию включен
- Поддержка mTLS (mutual TLS)
- Настраиваемые cipher suites
- Автоматическое обновление сертификатов

### Аутентификация
- JWT токены с валидацией через Django API
- Интеграция с Keycloak для централизованной аутентификации
- Кэширование токенов с TTL
- Безопасное хранение секретов через переменные окружения

### Rate Limiting
- Ограничение по IP-адресам
- Настраиваемые лимиты запросов
- Защита от DDoS атак

## 📊 Мониторинг

### Prometheus метрики
- Активные соединения и туннели
- Производительность и задержки
- Ошибки и отказы
- Использование ресурсов
- Keycloak метрики (валидация токенов, JWKS кэш)
- Бизнес-метрики (активные пользователи, доход)

### Grafana дашборды
- Готовые дашборды для мониторинга
- Автоматическая настройка через provisioning
- Визуализация метрик в реальном времени

### Алерты
- Настраиваемые алерты Prometheus
- Уведомления через Slack, Email
- Escalation policies
- Автоматическое разрешение

## 📝 Логирование

### Структурированное логирование
- **Формат:** JSON (по умолчанию), Console
- **Уровни:** DEBUG, INFO, WARN, ERROR
- **Ротация:** Автоматическая с настройкой размера и количества файлов
- **Поля:** timestamp, level, message, user_id, request_id, connection_id, duration, error

### Специализированные методы
- Логирование соединений и аутентификации
- Логирование валидации токенов и проверки разрешений
- Логирование производительности и безопасности
- Логирование бизнес-событий

## 🧪 Тестирование

### Типы тестов
- **Unit тесты** - покрытие всех компонентов
- **Integration тесты** - тестирование взаимодействия
- **E2E тесты** - полный цикл работы
- **Performance тесты** - нагрузочное тестирование

### Покрытие
- Более 80% покрытия кода
- Автоматизированные тесты в CI/CD
- Тестирование безопасности

## 🛠️ Установка

### Требования

- Go 1.18+
- Docker 20.10+ (опционально)
- Systemd (для установки как сервис)

### Быстрый старт

```bash
# Клонирование репозитория
git clone https://github.com/maxlanies/cloudbridge-relay-installer.git
cd cloudbridge-relay-installer

# Сборка
make build

# Запуск в development режиме
RELAY_JWT_SECRET=test-secret ./relay --config config/config.json --log-level DEBUG
```

### Docker

```bash
# Сборка образа
docker build -t cloudbridge-relay .

# Запуск контейнера
docker run -d \
  --name relay \
  -p 8080:8080 \
  -p 8081:8081 \
  -e RELAY_JWT_SECRET=your-secret \
  cloudbridge-relay ./relay --config config/config.json

# Просмотр логов
docker logs relay
```

### Системный сервис

```bash
# Установка как системный сервис
make install

# Проверка статуса
sudo systemctl status relay
```

## ⚙️ Конфигурация

### Основные параметры

```json
{
  "server": {
    "port": 8080,
    "max_tunnels": 1000,
    "timeout": "30s",
    "ping_interval": "30s",
    "cleanup_interval": "1m"
  },
  "django": {
    "api_endpoint": "http://your-django-server:8000/api",
    "jwt_secret": "REPLACE_WITH_SECURE_SECRET_FROM_ENVIRONMENT",
    "rest_timeout": "5s",
    "rest_cache_ttl": "5m"
  },
  "keycloak": {
    "enabled": false,
    "server_url": "https://keycloak.example.com",
    "realm": "cloudbridge",
    "client_id": "relay-client",
    "token_validation": {
      "issuer": "https://keycloak.example.com/realms/cloudbridge",
      "audience": "relay-client",
      "clock_skew": "30s"
    }
  },
  "security": {
    "tls_enabled": true,
    "cert_file": "/etc/relay/certs/fullchain.pem",
    "key_file": "/etc/relay/certs/privkey.pem",
    "rate_limit": {
      "enabled": true,
      "requests_per_second": 100,
      "burst": 200
    }
  },
  "metrics": {
    "enabled": true,
    "port": ":8081"
  },
  "logging": {
    "level": "INFO",
    "format": "json",
    "output": "/var/log/relay/relay.log"
  }
}
```

**Важно:** Все поля типа `time.Duration` должны быть указаны в виде строк с суффиксами времени (например, "30s", "5m", "1h").

### Переменные окружения

```bash
# Обязательные
RELAY_JWT_SECRET=your-secure-jwt-secret

# Опциональные
RELAY_DJANGO_API_ENDPOINT=http://your-django-server:8000/api
RELAY_LOG_LEVEL=info
RELAY_TLS_ENABLED=true

# Keycloak (если включен)
KEYCLOAK_SERVER_URL=https://keycloak.example.com
KEYCLOAK_REALM=cloudbridge
KEYCLOAK_CLIENT_ID=relay-client
KEYCLOAK_CLIENT_SECRET=your-client-secret
```

## 📖 Документация

### Основная документация
- [API Reference](docs/api_reference.md) - Полная документация API
- [Deployment Guide](docs/deployment_guide.md) - Руководство по развертыванию
- [Client Integration](docs/client_integration.md) - Интеграция клиентов
- [Debugging](docs/debugging.md) - Отладка и диагностика
- [Metrics](docs/metrics.md) - Архитектура и использование метрик
- [Keycloak Integration](docs/keycloak_integration.md) - Интеграция с Keycloak

### Примеры использования
- [Python Client](examples/python_client.py)
- [JavaScript Client](examples/javascript_client.js)
- [Go Client](examples/go_client.go)

## 🔧 Разработка

### Команды Makefile

```bash
# Сборка
make build

# Тестирование
make test-all

# Линтинг
make lint

# Форматирование
make fmt

# Docker
make docker-build
make docker-run

# Мониторинг
make setup-monitoring

# Безопасность
make security-audit
```

### Структура проекта

```
cloudbridge-relay-installer/
├── cmd/relay/           # Основной сервер
├── internal/            # Внутренние компоненты
│   ├── auth/           # Аутентификация
│   ├── keycloak/       # Интеграция с Keycloak
│   ├── metrics/        # Prometheus метрики
│   ├── logger/         # Структурированное логирование
│   ├── token/          # Валидация токенов
│   ├── ratelimit/      # Ограничение скорости
│   └── ...
├── config/             # Конфигурация
├── deploy/             # Конфигурации развертывания
├── docs/               # Документация
├── scripts/            # Скрипты управления
└── tests/              # Тесты
```

## 🚨 Устранение неполадок

### Частые проблемы

#### 1. Ошибка парсинга конфигурации
```
Failed to load configuration: failed to parse config file: json: cannot unmarshal string into Go struct field .server.timeout of type time.Duration
```
**Решение:** Убедитесь, что все поля `time.Duration` указаны в виде строк с суффиксами времени.

#### 2. Порт уже занят
```
Failed to start server: failed to create listener: listen tcp :8080: bind: address already in use
```
**Решение:** Остановите процесс, использующий порт 8080, или измените порт в конфигурации.

#### 3. Ошибка Docker контейнера
```
docker: Error response from daemon: failed to create task for container: failed to create shim task: OCI runtime create failed: runc create failed: unable to start container process: error during container init: exec: "--config": executable file not found in $PATH
```
**Решение:** Укажите правильную команду запуска: `cloudbridge-relay ./relay --config config/config.json`

### Диагностика

```bash
# Проверка статуса
sudo systemctl status relay

# Просмотр логов
tail -f /var/log/relay/relay.log

# Проверка метрик
curl http://localhost:8081/metrics

# Проверка конфигурации
./relay --config config/config.json --help
```

## 📈 Мониторинг

### Основные метрики

```promql
# Активные соединения
relay_active_connections

# Процент успешных валидаций токенов
rate(relay_keycloak_token_validations_total{result="success"}[5m]) / 
rate(relay_keycloak_token_validations_total[5m]) * 100

# Время соединения (95-й перцентиль)
histogram_quantile(0.95, rate(relay_connection_duration_seconds_bucket[5m]))
```

### Алерты

- Высокий процент неудачных валидаций токенов
- Медленные операции JWKS
- Низкая производительность кэша
- Высокое использование памяти

## 🤝 Поддержка

### Получение помощи

1. Проверьте [документацию](docs/)
2. Изучите [часто задаваемые вопросы](docs/debugging.md)
3. Создайте issue в репозитории

### Сбор информации для поддержки

```bash
# Логи
sudo journalctl -u relay > relay-logs.txt

# Метрики
curl http://localhost:8081/metrics > metrics.txt

# Конфигурация
cat /etc/relay/config.json > config.txt

# Системная информация
uname -a > system-info.txt
```

## 📄 Лицензия

Этот проект лицензирован под MIT License - см. файл [LICENSE](LICENSE) для деталей.

## 🏗️ Архитектура

CloudBridge Relay построен с использованием современных практик разработки:

- **Модульная архитектура** - четкое разделение ответственности
- **Dependency Injection** - легкое тестирование и конфигурация
- **Graceful Shutdown** - корректное завершение работы
- **Health Checks** - мониторинг состояния сервиса
- **Structured Logging** - удобный анализ логов
- **Metrics First** - детальный мониторинг производительности