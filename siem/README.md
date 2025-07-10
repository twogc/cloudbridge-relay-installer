# 🔒 CloudBridge SIEM System

Полнофункциональная система Security Information and Event Management (SIEM) для CloudBridge Relay v2.0, интегрированная с существующей архитектурой мониторинга.

## 🎯 Обзор

CloudBridge SIEM — это современная система мониторинга безопасности, которая:

- **Собирает события** из CloudBridge Relay, Keycloak и Django API
- **Анализирует угрозы** с помощью правил детекции и поведенческого анализа
- **Отправляет уведомления** через email, Telegram, Slack и другие каналы
- **Визуализирует данные** через интеграцию с существующими Grafana дашбордами
- **Соответствует требованиям** 152-ФЗ и 187-ФЗ
- **Поддерживает multi-tenancy** с полной изоляцией данных

## 🏗️ Архитектура

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
                                 │
                    ┌─────────────────┐
                    │   Monitoring    │
                    │                 │
                    │ - Prometheus    │
                    │ - Grafana       │
                    │ - AlertManager  │
                    └─────────────────┘
```

## 🚀 Быстрый запуск

### ⚠️ ВАЖНО: Развертывание на отдельном сервере

**SIEM систему необходимо развертывать на отдельном сервере** для обеспечения оптимальной производительности, безопасности и масштабируемости.

### Требования к серверу

#### Минимальные требования
- **CPU**: 4 vCPU
- **RAM**: 16 GB DDR4
- **Storage**: 500 GB SSD
- **Network**: 1 Gbps
- **OS**: Ubuntu 20.04+ или 22.04+

#### Рекомендуемые требования
- **CPU**: 8 vCPU
- **RAM**: 32 GB DDR4
- **Storage**: 2 TB NVMe SSD
- **Network**: 10 Gbps

#### Production требования
- **CPU**: 16 vCPU
- **RAM**: 64 GB DDR4 ECC
- **Storage**: 4 TB NVMe SSD (RAID 1)
- **Network**: 25 Gbps

### Автоматическое развертывание

```bash
# Скачивание скрипта развертывания
wget https://raw.githubusercontent.com/maxlanies/cloudbridge-relay-installer/main/scripts/deploy_siem_server.sh

# Установка прав на выполнение
chmod +x deploy_siem_server.sh

# Запуск полного развертывания
sudo ./deploy_siem_server.sh --full

# Или минимальное развертывание
sudo ./deploy_siem_server.sh --minimal
```

### Ручное развертывание

```bash
# Подготовка системы
sudo apt update && sudo apt upgrade -y
sudo apt install -y docker.io docker-compose nginx git curl wget

# Установка SIEM
sudo mkdir -p /opt/siem
sudo chown $USER:$USER /opt/siem
cd /opt/siem
git clone https://github.com/maxlanies/cloudbridge-relay-installer.git .

# Запуск SIEM
cd siem
docker-compose -f docker-compose.siem.yml up -d
```

## 📊 Доступ к сервисам

После запуска будут доступны:

| Сервис | URL | Логин/Пароль | Описание |
|--------|-----|--------------|----------|
| **SIEM Web UI** | http://localhost:8081 | - | Главный веб-интерфейс |
| **Kibana** | http://localhost:5601 | - | Визуализация и поиск событий |
| **Elasticsearch** | http://localhost:9200 | - | API для поиска данных |
| **Kafka** | localhost:9092 | - | Потоковая обработка |
| **Redis** | localhost:6379 | - | Кэширование |
| **RabbitMQ** | http://localhost:15672 | guest/guest | Очереди сообщений |

## 🔧 Интеграция с CloudBridge Relay

### Настройка сбора данных

```bash
# Создание SSH ключей для безопасного доступа
ssh-keygen -t rsa -b 4096 -f /opt/siem/.ssh/id_rsa -N ""

# Копирование ключа на сервер Relay
ssh-copy-id -i /opt/siem/.ssh/id_rsa.pub ubuntu@your_relay_server_ip
```

### Конфигурация удаленного сбора

```yaml
# /opt/siem/siem/config/remote_collectors.yaml
remote_collectors:
  relay_server:
    host: "your_relay_server_ip"
    port: 22
    username: "ubuntu"
    key_file: "/opt/siem/.ssh/id_rsa"
    sources:
      - path: "/var/log/relay/relay.log"
        type: "relay_logs"
        interval: 60
      - path: "http://localhost:8081/metrics"
        type: "relay_metrics"
        interval: 30
      - path: "http://localhost:8080/admin/realms/cloudbridge/events"
        type: "keycloak_events"
        interval: 120
      - path: "http://localhost:8000/api/v1/audit-logs"
        type: "django_audit"
        interval: 60
```

## 📈 Правила детекции

### Встроенные правила

1. **Brute Force Detection**
   - Детекция множественных неудачных попыток входа
   - Настраиваемый порог и временное окно

2. **Suspicious Activity**
   - Обнаружение подозрительной активности
   - Анализ паттернов поведения

3. **Data Exfiltration**
   - Детекция несанкционированного доступа к данным
   - Мониторинг больших объемов передачи данных

4. **Privilege Escalation**
   - Обнаружение попыток повышения привилегий
   - Мониторинг административных действий

### Настройка правил

```yaml
# siem/config/detection_rules.yaml
rules:
  brute_force:
    enabled: true
    description: "Множественные неудачные попытки входа"
    conditions:
      event_type: "authentication_failure"
      time_window: 300
      threshold: 5
    actions:
      - type: "alert"
        severity: "high"
      - type: "block_ip"
        duration: 3600
```

## 🔔 Уведомления

### Поддерживаемые каналы

- **Email** (SMTP)
- **Telegram**
- **Slack**
- **Webhooks**
- **SMS** (опционально)

### Настройка уведомлений

```yaml
# siem/config/notifications.yaml
notifications:
  email:
    enabled: true
    smtp_server: "smtp.gmail.com"
    smtp_port: 587
    username: "your-email@gmail.com"
    password: "your-app-password"
    from: "siem@yourcompany.com"
    to: ["security@yourcompany.com", "admin@yourcompany.com"]
    
  telegram:
    enabled: true
    bot_token: "your-bot-token"
    chat_id: "your-chat-id"
    
  slack:
    enabled: true
    webhook_url: "https://hooks.slack.com/services/your/webhook/url"
```

## 🎨 Веб-интерфейс

### Основные функции

- **Dashboard** - главная панель с обзором событий
- **Events** - просмотр и поиск событий безопасности
- **Alerts** - управление алертами и уведомлениями
- **Rules** - настройка правил детекции
- **Analytics** - аналитика и отчеты
- **Settings** - настройки системы

### Multi-tenant поддержка

- Переключение между тенантами
- Изоляция данных в UI
- Tenant-specific настройки
- Роли и права доступа

## 🔗 Интеграция с существующим мониторингом

### Prometheus Integration

SIEM автоматически интегрируется с существующим Prometheus:

```yaml
# Автоматически добавляется в prometheus/prometheus.yml:
scrape_configs:
  - job_name: 'siem-backend'
    static_configs:
      - targets: ['siem-server:8000']
    metrics_path: '/metrics'
    scrape_interval: 15s
```

### Grafana Integration

Создаются новые дашборды в существующей Grafana:

- **SIEM Events Dashboard** - мониторинг событий безопасности
- **SIEM Alerts Dashboard** - управление алертами
- **SIEM Correlation Dashboard** - анализ корреляций
- **SIEM Security Dashboard** - общая безопасность

### AlertManager Integration

Интеграция с существующим AlertManager для централизованного управления алертами.

## 🔒 Безопасность

### Аутентификация и авторизация

- **Keycloak OIDC** интеграция
- **RBAC** (Role-Based Access Control)
- **Multi-tenant изоляция**
- **Audit logging** всех действий

### Шифрование

- **TLS 1.3** для всех соединений
- **Data encryption** at rest
- **Secure communication** между компонентами

### Соответствие требованиям

- **152-ФЗ** - защита персональных данных
- **187-ФЗ** - безопасность критической информационной инфраструктуры
- **ISO 27001** - информационная безопасность

## 📊 Производительность

### Метрики производительности

- **Events/sec**: > 10,000
- **Alert latency**: < 100ms
- **Storage retention**: 90 дней
- **Uptime**: 99.9%

### Масштабирование

- **Горизонтальное масштабирование**
- **Auto-scaling** на основе нагрузки
- **Load balancing**
- **High availability**

## 🛠️ Управление

### Системные команды

```bash
# Управление сервисом
sudo systemctl start siem
sudo systemctl stop siem
sudo systemctl restart siem
sudo systemctl status siem

# Просмотр логов
sudo journalctl -u siem.service -f
tail -f /opt/siem/logs/*.log

# Мониторинг ресурсов
/opt/siem/scripts/monitor_siem.sh

# Резервное копирование
/opt/siem/scripts/backup_siem.sh
```

### Docker команды

```bash
# Просмотр контейнеров
docker ps

# Просмотр логов
docker-compose -f docker-compose.siem.yml logs -f

# Перезапуск сервисов
docker-compose -f docker-compose.siem.yml restart

# Обновление образов
docker-compose -f docker-compose.siem.yml pull
docker-compose -f docker-compose.siem.yml up -d
```

## 📚 Документация

### Основная документация

- **[Техническое задание](TECHNICAL_SPECIFICATION.md)** - подробное ТЗ для разработчика
- **[Руководство по развертыванию](../docs/deployment/SIEM_DEPLOYMENT_COMPLETE_GUIDE.md)** - полное руководство
- **[Руководство по интеграции](../docs/integration/siem_integration_guide.md)** - интеграция с CloudBridge Relay
- **[Анализ SIEM системы](../docs/integration/siem_integration_analysis.md)** - архитектурный анализ

### API документация

- **OpenAPI/Swagger**: http://localhost:8081/docs
- **API Reference**: http://localhost:8081/redoc

## 🐛 Troubleshooting

### Частые проблемы

1. **SIEM не собирает события**
   - Проверьте подключение к источникам данных
   - Убедитесь в правильности конфигурации коллекторов

2. **Высокое потребление ресурсов**
   - Проверьте настройки Elasticsearch
   - Оптимизируйте правила детекции

3. **Алерты не отправляются**
   - Проверьте настройки уведомлений
   - Убедитесь в доступности внешних сервисов

### Логи и диагностика

```bash
# Просмотр логов SIEM
tail -f /opt/siem/logs/siem.log

# Проверка состояния Elasticsearch
curl -X GET "localhost:9200/_cluster/health?pretty"

# Проверка Kafka
docker exec -it siem_kafka_1 kafka-topics --list --bootstrap-server localhost:9092

# Проверка Redis
docker exec -it siem_redis_1 redis-cli ping
```

## 🤝 Поддержка

### Получение помощи

- **Документация**: [docs/](../docs/)
- **Issues**: GitHub Issues
- **Discussions**: GitHub Discussions
- **Email**: support@2gc.ru

### Сообщество

- **Telegram**: @cloudbridge_support
- **Discord**: CloudBridge Community
- **Blog**: https://blog.2gc.ru

---

**Версия**: 2.0.0  
**Последнее обновление**: 2024  
**Лицензия**: MIT  
**Поддержка**: CloudBridge Team 