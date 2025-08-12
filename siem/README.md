# 🎯 CloudBridge SIEM v2.0

Система управления информацией и событиями безопасности (SIEM) для CloudBridge Relay v2.0

## 📋 Обзор

CloudBridge SIEM - это комплексная система мониторинга безопасности, которая интегрируется с существующей архитектурой CloudBridge Relay v2.0. Система обеспечивает сбор, анализ и корреляцию событий безопасности в реальном времени.

### 🏗️ Архитектура

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Relay Server  │    │  Keycloak Server│    │  Django Server  │
│                 │    │                 │    │                 │
│ - CloudBridge   │    │ - Authentication│    │ - Management    │
│ - TCP Proxy     │    │ - User Mgmt     │    │ - API           │
│ - Multi-tenant  │    │ - SSO           │    │ - Dashboard     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │   SIEM Server   │
                    │                 │
                    │ - Elasticsearch │
                    │ - Kibana        │
                    │ - Kafka         │
                    │ - Collectors    │
                    │ - Processors    │
                    │ - Alert Manager │
                    │ - Web UI        │
                    └─────────────────┘
```

## 🚀 Быстрый старт

### Предварительные требования

- Docker и Docker Compose
- Минимум 8GB RAM
- 100GB свободного места на диске
- Linux/Unix система

### 1. Клонирование репозитория

```bash
git clone <repository-url>
cd cloudbridge-relay-installer/siem
```

### 2. Настройка переменных окружения

Создайте файл `.env` в корне SIEM директории:

```bash
# SIEM Configuration
SIEM_EMAIL_PASSWORD=your-email-password
TELEGRAM_BOT_TOKEN=your-telegram-bot-token
TELEGRAM_CHAT_ID=your-telegram-chat-id
JWT_SECRET_KEY=your-jwt-secret-key

# External Services
RELAY_METRICS_URL=http://relay:8081/metrics
KEYCLOAK_ADMIN_URL=http://keycloak:8080/auth/admin
DJANGO_API_URL=http://django:8000/api

# Monitoring
PROMETHEUS_URL=http://localhost:9090
GRAFANA_URL=http://localhost:3000
```

### 3. Запуск системы

```bash
# Запуск всех сервисов
docker-compose up -d

# Проверка статуса
docker-compose ps

# Просмотр логов
docker-compose logs -f siem-api
```

### 4. Доступ к интерфейсам

- **SIEM Web UI**: http://localhost:3000
- **Grafana**: http://localhost:3001 (admin/admin)
- **Kibana**: http://localhost:5601
- **Prometheus**: http://localhost:9090
- **AlertManager**: http://localhost:9093

## 🔧 Конфигурация

### Основные настройки

Основная конфигурация находится в файле `config/siem_config.yaml`:

```yaml
# Application Settings
app:
  name: "CloudBridge SIEM"
  version: "2.0.0"
  debug: false
  log_level: "INFO"
  timezone: "Europe/Moscow"

# Server Configuration
server:
  host: "0.0.0.0"
  port: 8000
  workers: 4
```

### Правила детекции

Правила детекции настраиваются в файле `config/detection_rules.yaml`:

```yaml
rules:
  brute_force:
    name: "Brute Force Attack Detection"
    description: "Множественные неудачные попытки входа"
    enabled: true
    severity: "high"
    conditions:
      event_type: "authentication_failure"
      time_window: 300
      threshold: 5
```

### Уведомления

Настройка уведомлений в `config/siem_config.yaml`:

```yaml
notifications:
  email:
    enabled: true
    smtp_server: "smtp.rusender.ru"
    smtp_port: 465
    username: "siem@2gc.ru"
    password: "${SIEM_EMAIL_PASSWORD}"
  
  telegram:
    enabled: true
    bot_token: "${TELEGRAM_BOT_TOKEN}"
    chat_id: "${TELEGRAM_CHAT_ID}"
```

## 📊 Мониторинг и метрики

### Prometheus метрики

SIEM система экспортирует следующие метрики:

- `siem_events_total` - общее количество событий
- `siem_events_by_type` - события по типам
- `siem_events_by_source` - события по источникам
- `siem_events_by_severity` - события по важности
- `siem_alerts_active` - активные алерты
- `siem_alerts_critical` - критические алерты
- `siem_correlations_active` - активные корреляции

### Grafana дашборды

Включены готовые дашборды:

- **SIEM Events Dashboard** - обзор событий безопасности
- **SIEM Alerts Dashboard** - управление алертами
- **SIEM Correlation Dashboard** - корреляция событий

## 🔐 Безопасность

### Multi-tenant изоляция

Система поддерживает изоляцию данных по тенантам:

```python
# Автоматическое определение тенанта
tenant_id = await get_current_tenant(request)

# Изоляция данных
events = await event_service.get_events(tenant_id=tenant_id)
```

### Аутентификация

Интеграция с Keycloak для аутентификации:

```yaml
keycloak:
  admin_url: "http://keycloak:8080/auth/admin"
  realm: "cloudbridge"
  client_id: "siem-client"
  client_secret: "siem-secret"
```

### Шифрование

- TLS 1.3 для всех соединений
- Шифрование данных в состоянии покоя
- Безопасное хранение секретов

## 🔄 Интеграция с существующими системами

### Relay Integration

Сбор данных из CloudBridge Relay:

```yaml
relay:
  metrics_url: "http://relay:8081/metrics"
  log_file: "/var/log/relay/relay.log"
  api_url: "http://relay:8081/api"
```

### Keycloak Integration

Сбор событий аутентификации:

```yaml
keycloak:
  admin_url: "http://keycloak:8080/auth/admin"
  realm: "cloudbridge"
```

### Django API Integration

Сбор бизнес-событий:

```yaml
django:
  api_url: "http://django:8000/api"
  api_key: "siem-api-key"
```

## 📈 Производительность

### Требования к производительности

- **Events/sec**: > 10,000
- **Alert latency**: < 100ms
- **Storage**: 90 дней retention
- **Uptime**: 99.9%

### Оптимизация

```yaml
performance:
  max_events_per_second: 10000
  event_processing_timeout: 100
  batch_size: 1000
  worker_threads: 8
  max_memory_usage: "8GB"
```

## 🛠️ Разработка

### Структура проекта

```
siem/
├── backend/                 # FastAPI backend
│   ├── app/
│   │   ├── api/            # API endpoints
│   │   ├── models/         # Data models
│   │   ├── schemas/        # Pydantic schemas
│   │   ├── services/       # Business logic
│   │   └── main.py         # FastAPI app
│   └── requirements.txt
├── frontend/               # React frontend
│   ├── src/
│   │   ├── components/     # React components
│   │   ├── pages/          # Page components
│   │   └── App.tsx
│   └── package.json
├── collectors/             # Event collectors
├── processors/             # Event processors
├── alerting/               # Alert management
├── config/                 # Configuration files
└── docker-compose.yml      # Docker orchestration
```

### Локальная разработка

```bash
# Запуск в режиме разработки
docker-compose -f docker-compose.dev.yml up -d

# Просмотр логов
docker-compose logs -f

# Тестирование API
curl http://localhost:8000/api/v1/events
```

## 🧪 Тестирование

### Unit тесты

```bash
# Запуск тестов backend
cd backend
python -m pytest tests/

# Запуск тестов frontend
cd frontend
npm test
```

### Интеграционные тесты

```bash
# Запуск интеграционных тестов
docker-compose -f docker-compose.test.yml up --abort-on-container-exit
```

### Load тестирование

```bash
# Тестирование производительности
python scripts/load_test.py --events 10000 --duration 60
```

## 📚 API Документация

### Swagger UI

Доступ к API документации: http://localhost:8000/docs

### Основные endpoints

```bash
# События
GET    /api/v1/events          # Получение событий
POST   /api/v1/events          # Создание события
GET    /api/v1/events/{id}     # Получение события по ID

# Алерты
GET    /api/v1/alerts          # Получение алертов
POST   /api/v1/alerts          # Создание алерта
PUT    /api/v1/alerts/{id}     # Обновление алерта

# Корреляции
GET    /api/v1/correlations    # Получение корреляций
POST   /api/v1/correlations    # Создание корреляции

# Аналитика
GET    /api/v1/analytics       # Аналитические данные
```

## 🔧 Обслуживание

### Backup

```bash
# Создание backup
./scripts/backup.sh

# Восстановление из backup
./scripts/restore.sh backup-2024-01-01.tar.gz
```

### Мониторинг

```bash
# Проверка здоровья системы
curl http://localhost:8000/health

# Просмотр метрик
curl http://localhost:8000/metrics
```

### Логи

```bash
# Просмотр логов SIEM
docker-compose logs -f siem-api

# Просмотр логов Elasticsearch
docker-compose logs -f elasticsearch

# Просмотр логов Kafka
docker-compose logs -f kafka
```

## 🚨 Troubleshooting

### Частые проблемы

#### 1. Elasticsearch не запускается

```bash
# Проверка ресурсов
docker stats

# Увеличение памяти для Elasticsearch
export ES_JAVA_OPTS="-Xms2g -Xmx2g"
```

#### 2. Kafka соединение отказано

```bash
# Проверка статуса Kafka
docker-compose ps kafka

# Перезапуск Kafka
docker-compose restart kafka
```

#### 3. Проблемы с аутентификацией

```bash
# Проверка Keycloak
curl http://keycloak:8080/auth/realms/cloudbridge

# Проверка токенов
./scripts/debug_token.sh
```

### Логи и отладка

```bash
# Включение debug режима
export DEBUG=true
docker-compose up -d

# Просмотр детальных логов
docker-compose logs -f --tail=100
```

## 📞 Поддержка

### Контакты

- **Email**: security@2gc.ru
- **Telegram**: @cloudbridge_support
- **Документация**: https://docs.2gc.ru/siem

### Сообщение об ошибках

При обнаружении проблем создайте issue с:

1. Описанием проблемы
2. Логами системы
3. Версией системы
4. Шагами для воспроизведения

## 📄 Лицензия

Copyright © 2024 2GC. Все права защищены.

---

**Версия**: 2.0.0  
**Дата**: 2024-01-01  
**Статус**: Production Ready 