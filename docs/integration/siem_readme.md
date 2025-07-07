# CloudBridge SIEM System

Полнофункциональная система Security Information and Event Management (SIEM) для CloudBridge Relay, обеспечивающая сбор, анализ и мониторинг событий безопасности в реальном времени.

## 🎯 Обзор

CloudBridge SIEM — это современная система мониторинга безопасности, которая:

- **Собирает события** из CloudBridge Relay, Keycloak и системных логов
- **Анализирует угрозы** с помощью правил детекции и поведенческого анализа
- **Отправляет уведомления** через email, Telegram, Slack и другие каналы
- **Визуализирует данные** через Kibana и Grafana
- **Соответствует требованиям** 152-ФЗ и 187-ФЗ

## 🏗️ Архитектура

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Data Sources  │    │   Data Pipeline │    │   Analytics     │
│                 │    │                 │    │                 │
│ • CloudBridge   │───▶│ • Kafka         │───▶│ • Elasticsearch │
│ • Keycloak      │    │ • Collectors    │    │ • ML Models     │
│ • System Logs   │    │ • Processors    │    │ • Rules Engine  │
│ • Metrics       │    │ • Enrichment    │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │   Visualization │
                       │                 │
                       │ • Kibana        │
                       │ • Grafana       │
                       │ • Web UI        │
                       └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │   Alerting      │
                       │                 │
                       │ • Email         │
                       │ • Telegram      │
                       │ • Slack         │
                       │ • Webhooks      │
                       └─────────────────┘
```

## 🚀 Быстрый запуск

### Требования

- Docker 20.10+
- Docker Compose 2.0+
- 4+ GB RAM
- 20+ GB свободного места

### Автоматическая установка

```bash
# Клонирование репозитория
git clone https://github.com/maxlanies/cloudbridge-relay-installer.git
cd cloudbridge-relay-installer/siem

# Запуск SIEM системы
chmod +x start_siem.sh
./start_siem.sh
```

### Ручная установка

```bash
# Создание директорий
mkdir -p logs/siem config/siem data/{elasticsearch,redis,prometheus,grafana} backup/siem

# Копирование конфигурации
cp siem_config.yaml config/siem/

# Запуск с Docker Compose
docker-compose -f docker-compose.siem.yml up -d
```

## 📊 Доступ к сервисам

После запуска будут доступны:

| Сервис | URL | Логин/Пароль | Описание |
|--------|-----|--------------|----------|
| **Kibana** | http://localhost:5601 | - | Визуализация и поиск событий |
| **Grafana** | http://localhost:3001 | admin/admin | Дашборды и метрики |
| **Elasticsearch** | http://localhost:9200 | - | API для поиска данных |
| **Prometheus** | http://localhost:9091 | - | Метрики системы |
| **SIEM Web UI** | http://localhost:8081 | - | Главный веб-интерфейс |
| **Kafka** | localhost:9092 | - | Потоковая обработка |
| **Redis** | localhost:6379 | - | Кэширование |

## 🔧 Конфигурация

### Основные настройки

Редактируйте файл `config/siem/siem_config.yaml`:

```yaml
# Настройки сбора данных
collection:
  relay_logs_interval: 60
  relay_metrics_interval: 30
  keycloak_events_interval: 120

# Правила детекции
processing:
  detection_rules:
    brute_force_detection:
      enabled: true
      threshold: 5
      time_window: 300
      severity: high

# Уведомления
alerting:
  min_severity: medium
  channels:
    email:
      enabled: true
      smtp_server: smtp.gmail.com
      # ... другие настройки
```

### Переменные окружения

Создайте файл `.env.siem`:

```bash
# Elasticsearch
ELASTICSEARCH_PASSWORD=changeme
ELASTICSEARCH_USERNAME=elastic

# Уведомления
EMAIL_SMTP_SERVER=smtp.gmail.com
EMAIL_USERNAME=your-email@gmail.com
EMAIL_PASSWORD=your-app-password

TELEGRAM_BOT_TOKEN=your-bot-token
TELEGRAM_CHAT_ID=your-chat-id

SLACK_WEBHOOK_URL=https://hooks.slack.com/services/your/webhook/url
```

## 📈 Правила детекции

### Встроенные правила

1. **Brute Force Detection**
   - Детекция множественных неудачных попыток входа
   - Настраиваемый порог и временное окно

2. **Suspicious IP Detection**
   - Блокировка подозрительных IP-адресов
   - Интеграция с threat intelligence

3. **Unusual User Activity**
   - Анализ поведения пользователей
   - Детекция аномалий

4. **Relay Anomaly Detection**
   - Мониторинг аномалий в работе Relay
   - Детекция ошибок и сбоев

5. **Data Exfiltration Detection**
   - Обнаружение эксфильтрации данных
   - Мониторинг больших передач данных

### Создание собственных правил

```yaml
custom_rule:
  enabled: true
  conditions:
    event_type: "relay_operation"
    source: "relay_logs"
    message_contains: "error"
    threshold: 10
    time_window: 600
  severity: high
  tags: [custom, error_detection]
```

## 🔔 Система уведомлений

### Поддерживаемые каналы

- **Email** - SMTP отправка
- **Telegram** - Bot API
- **Slack** - Webhook интеграция
- **Webhook** - HTTP POST запросы
- **PagerDuty** - Incident management
- **Microsoft Teams** - Webhook интеграция

### Настройка уведомлений

```yaml
alerting:
  channels:
    telegram:
      enabled: true
      bot_token: "your-bot-token"
      chat_id: "your-chat-id"
    
    email:
      enabled: true
      smtp_server: "smtp.gmail.com"
      username: "your-email@gmail.com"
      password: "your-app-password"
      to_emails: ["admin@company.com"]
```

### Эскалация уведомлений

```yaml
escalation:
  enabled: true
  rules:
    - condition: "severity == 'critical'"
      actions:
        - "immediate_notification"
        - "create_incident"
        - "call_oncall"
      timeout: 300
```

## 📊 Визуализация данных

### Kibana дашборды

- **Security Overview** - Общий обзор безопасности
- **Authentication Events** - События аутентификации
- **Network Activity** - Сетевая активность
- **Threat Intelligence** - Информация об угрозах

### Grafana дашборды

- **SIEM Metrics** - Метрики производительности
- **Alert Statistics** - Статистика уведомлений
- **System Health** - Здоровье системы
- **Real-time Monitoring** - Мониторинг в реальном времени

## 🔒 Безопасность

### Соответствие требованиям

- **152-ФЗ** - Защита персональных данных
- **187-ФЗ** - Безопасность критической инфраструктуры
- **ГОСТ Р 34.10-2012** - Алгоритмы подписи
- **ГОСТ Р 34.11-2012** - Алгоритмы хеширования

### Меры безопасности

- Шифрование данных в состоянии покоя и передаче
- Аутентификация через Keycloak
- Аудит всех операций
- Ротация ключей и сертификатов
- Изоляция сетей

## 🛠️ Управление системой

### Команды управления

```bash
# Запуск
./start_siem.sh

# Остановка
./start_siem.sh stop

# Перезапуск
./start_siem.sh restart

# Статус
./start_siem.sh status

# Просмотр логов
./start_siem.sh logs

# Редактирование конфигурации
./start_siem.sh config

# Создание резервной копии
./start_siem.sh backup
```

### Docker Compose команды

```bash
# Запуск всех сервисов
docker-compose -f docker-compose.siem.yml up -d

# Остановка
docker-compose -f docker-compose.siem.yml down

# Просмотр логов
docker-compose -f docker-compose.siem.yml logs -f

# Перезапуск конкретного сервиса
docker-compose -f docker-compose.siem.yml restart siem-collector
```

## 📝 Логирование

### Структура логов

```
logs/siem/
├── collector.log      # Логи сборщика данных
├── processor.log      # Логи обработчика событий
├── alert-manager.log  # Логи системы уведомлений
└── webui.log         # Логи веб-интерфейса
```

### Уровни логирования

- **DEBUG** - Детальная отладочная информация
- **INFO** - Общая информация о работе
- **WARNING** - Предупреждения
- **ERROR** - Ошибки
- **CRITICAL** - Критические ошибки

## 🔧 Интеграции

### Внешние системы

- **Jira** - Управление инцидентами
- **ServiceNow** - IT Service Management
- **Ansible** - Автоматизация
- **Terraform** - Infrastructure as Code

### Threat Intelligence

- **AlienVault OTX** - Открытая threat intelligence
- **VirusTotal** - Анализ файлов и URL
- **AbuseIPDB** - База данных злонамеренных IP

## 📊 Мониторинг и метрики

### Prometheus метрики

- `siem_events_collected_total` - Общее количество собранных событий
- `siem_alerts_generated_total` - Количество сгенерированных уведомлений
- `siem_processing_duration_seconds` - Время обработки событий
- `siem_storage_usage_bytes` - Использование хранилища

### Health Checks

```bash
# Проверка Elasticsearch
curl http://localhost:9200/_cluster/health

# Проверка Kibana
curl http://localhost:5601/api/status

# Проверка Kafka
docker exec siem-kafka kafka-topics --bootstrap-server localhost:9092 --list
```

## 🚨 Устранение неполадок

### Частые проблемы

1. **Elasticsearch не запускается**
   ```bash
   # Проверка логов
   docker logs siem-elasticsearch
   
   # Проверка ресурсов
   docker stats siem-elasticsearch
   ```

2. **Kafka не подключается**
   ```bash
   # Проверка Zookeeper
   docker logs siem-zookeeper
   
   # Проверка Kafka
   docker logs siem-kafka
   ```

3. **Уведомления не отправляются**
   ```bash
   # Проверка конфигурации
   docker logs siem-alert-manager
   
   # Тестирование каналов
   ./test_notifications.sh
   ```

### Диагностика

```bash
# Проверка всех сервисов
./start_siem.sh status

# Просмотр логов всех сервисов
docker-compose -f docker-compose.siem.yml logs

# Проверка сетевых соединений
docker network inspect siem_siem-network
```

## 📚 Дополнительные ресурсы

### Документация

- [Elasticsearch Guide](https://www.elastic.co/guide/index.html)
- [Kibana User Guide](https://www.elastic.co/guide/en/kibana/current/index.html)
- [Grafana Documentation](https://grafana.com/docs/)
- [Kafka Documentation](https://kafka.apache.org/documentation/)

### Сообщество

- [Elastic Community](https://discuss.elastic.co/)
- [Grafana Community](https://community.grafana.com/)
- [Apache Kafka Community](https://kafka.apache.org/community)

## 🤝 Поддержка

### Получение помощи

1. Проверьте [FAQ](docs/faq.md)
2. Изучите [Troubleshooting Guide](docs/troubleshooting.md)
3. Создайте [Issue](https://github.com/maxlanies/cloudbridge-relay-installer/issues)

### Вклад в проект

1. Fork репозитория
2. Создайте feature branch
3. Внесите изменения
4. Создайте Pull Request

## 📄 Лицензия

Этот проект лицензирован под MIT License - см. файл [LICENSE](../LICENSE) для деталей.

---

**CloudBridge SIEM System** - Современное решение для мониторинга безопасности с открытым исходным кодом. 