# 🎯 Техническое задание: Интеграция SIEM с CloudBridge Relay v2.0

## 📋 Обзор проекта

### Цель
Интегрировать SIEM систему с существующей архитектурой CloudBridge Relay v2.0, используя уже развернутые компоненты мониторинга (Prometheus, Grafana, AlertManager) и следуя рекомендациям по развертыванию на отдельном сервере.

### Архитектура проекта
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

## 🏗️ Существующая инфраструктура

### Мониторинг (корень проекта)
- **Prometheus**: `prometheus/prometheus.yml` - сбор метрик
- **Grafana**: `grafana/provisioning/` - визуализация
- **AlertManager**: `deploy/alertmanager/alertmanager.yml` - алерты
- **Дашборды**: `deploy/grafana/dashboards/` - готовые дашборды

### Основные сервисы
- **Relay**: TCP proxy с multi-tenancy
- **Keycloak**: аутентификация и авторизация
- **Django API**: управление и биллинг
- **Nginx**: reverse proxy

## 🎯 Задачи разработчика

### 1. Адаптация SIEM к существующей архитектуре

#### 1.1. Интеграция с существующим Prometheus
```yaml
# Добавить в prometheus/prometheus.yml новые scrape jobs:
scrape_configs:
  # Существующие jobs...
  
  # SIEM Backend API
  - job_name: 'siem-backend'
    static_configs:
      - targets: ['siem-server:8000']
    metrics_path: '/metrics'
    scrape_interval: 15s
    
  # SIEM Event Processor
  - job_name: 'siem-processor'
    static_configs:
      - targets: ['siem-server:8001']
    metrics_path: '/metrics'
    scrape_interval: 15s
    
  # SIEM Alert Manager
  - job_name: 'siem-alerts'
    static_configs:
      - targets: ['siem-server:8002']
    metrics_path: '/metrics'
    scrape_interval: 15s
```

#### 1.2. Интеграция с существующим AlertManager
```yaml
# Добавить в deploy/alertmanager/alertmanager.yml новые receivers:
receivers:
  # Существующие receivers...
  
  - name: 'siem-critical'
    email_configs:
    - to: "security@2gc.ru"
      from: "siem@2gc.ru"
      smarthost: "smtp.rusender.ru:465"
      auth_username: "siem@2gc.ru"
      auth_password: "${SIEM_EMAIL_PASSWORD}"
    telegram_configs:
    - bot_token: "${TELEGRAM_BOT_TOKEN}"
      chat_id: "${TELEGRAM_CHAT_ID}"
      
  - name: 'siem-warning'
    email_configs:
    - to: "admin@2gc.ru"
      from: "siem@2gc.ru"
      smarthost: "smtp.rusender.ru:465"
      auth_username: "siem@2gc.ru"
      auth_password: "${SIEM_EMAIL_PASSWORD}"
```

#### 1.3. Создание SIEM дашбордов для Grafana
```json
// Создать в deploy/grafana/dashboards/:
// - siem_events_dashboard.json
// - siem_alerts_dashboard.json  
// - siem_correlation_dashboard.json
// - siem_security_dashboard.json
```

### 2. Разработка SIEM компонентов

#### 2.1. Backend API (FastAPI)
```python
# siem/backend/app/main.py
# Реализовать endpoints:
# - /api/v1/events - сбор событий
# - /api/v1/alerts - управление алертами
# - /api/v1/correlation - правила корреляции
# - /api/v1/analytics - аналитика
# - /api/v1/webhooks - интеграция с внешними системами
```

#### 2.2. Event Collector
```python
# siem/collectors/relay_collector.py
# Сбор данных из:
# - Relay logs: /var/log/relay/relay.log
# - Relay metrics: http://relay:8081/metrics
# - Keycloak events: Keycloak Admin API
# - Django audit logs: Django API
# - System logs: /var/log/syslog, /var/log/auth.log
```

#### 2.3. Event Processor
```python
# siem/processors/event_processor.py
# Обработка событий:
# - Парсинг и нормализация
# - Обогащение данными
# - Применение правил детекции
# - Корреляция событий
# - Генерация алертов
```

#### 2.4. Alert Manager
```python
# siem/alerting/alert_manager.py
# Управление алертами:
# - Группировка алертов
# - Дедупликация
# - Эскалация
# - Отправка уведомлений
```

### 3. Интеграция с существующими сервисами

#### 3.1. Relay Integration
```python
# Сбор данных из Relay:
# - Логи подключений и ошибок
# - Метрики производительности
# - Информация о туннелях
# - Multi-tenant данные
```

#### 3.2. Keycloak Integration
```python
# Сбор событий аутентификации:
# - Успешные/неуспешные логины
# - Изменения пользователей
# - Административные действия
# - MFA события
```

#### 3.3. Django API Integration
```python
# Сбор бизнес-событий:
# - Создание/изменение компаний
# - Управление серверами
# - Биллинг события
# - Аудит действий
```

### 4. Multi-tenant поддержка

#### 4.1. Elasticsearch индексы
```yaml
# Индексы по тенантам:
# - siem-{tenant_id}-events-{date}
# - siem-{tenant_id}-alerts-{date}
# - siem-{tenant_id}-correlations-{date}
```

#### 4.2. Kafka topics
```yaml
# Topics по тенантам:
# - siem.events.{tenant_id}
# - siem.alerts.{tenant_id}
# - siem.correlations.{tenant_id}
```

#### 4.3. Изоляция данных
```python
# Обеспечить изоляцию:
# - Данные тенантов не пересекаются
# - Правила детекции per tenant
# - Алерты per tenant
# - Доступ к данным только для своего тенанта
```

### 5. Правила детекции

#### 5.1. Базовые правила
```yaml
# siem/config/detection_rules.yaml
rules:
  brute_force:
    description: "Множественные неудачные попытки входа"
    conditions:
      event_type: "authentication_failure"
      time_window: 300
      threshold: 5
    actions:
      - alert: "high"
      - block_ip: 3600
        
  suspicious_activity:
    description: "Подозрительная активность"
    conditions:
      event_type: "unusual_access"
      time_window: 600
      threshold: 3
    actions:
      - alert: "medium"
      - log: true
```

#### 5.2. Корреляционные правила
```yaml
# siem/config/correlation_rules.yaml
correlations:
  attack_chain:
    description: "Цепочка атаки"
    events:
      - authentication_failure
      - successful_login
      - data_access
    time_window: 1800
    severity: "high"
```

### 6. Уведомления

#### 6.1. Каналы уведомлений
```python
# Поддержка каналов:
# - Email (SMTP)
# - Telegram
# - Slack
# - Webhooks
# - SMS (опционально)
```

#### 6.2. Шаблоны уведомлений
```yaml
# siem/config/notification_templates.yaml
templates:
  critical_alert:
    subject: "🚨 КРИТИЧЕСКАЯ УГРОЗА: {alert_name}"
    body: |
      Обнаружена критическая угроза безопасности!
      
      Сервис: {service}
      Время: {timestamp}
      Описание: {description}
      
      Немедленно примите меры!
```

### 7. Веб-интерфейс

#### 7.1. React Frontend
```typescript
// siem/frontend/src/
// Компоненты:
// - Dashboard - главная панель
// - Events - просмотр событий
// - Alerts - управление алертами
// - Rules - настройка правил
// - Analytics - аналитика
// - Settings - настройки
```

#### 7.2. Multi-tenant UI
```typescript
// Обеспечить:
// - Переключение между тенантами
// - Изоляцию данных в UI
// - Tenant-specific настройки
// - Роли и права доступа
```

### 8. Производительность и масштабирование

#### 8.1. Оптимизация
```python
# Обеспечить:
# - Обработка > 10,000 событий/сек
# - Задержка < 100ms для алертов
# - Эффективное использование памяти
# - Оптимизация запросов к Elasticsearch
```

#### 8.2. Масштабирование
```yaml
# Поддержка:
# - Горизонтальное масштабирование
# - Auto-scaling на основе нагрузки
# - Load balancing
# - High availability
```

## 🔧 Технические требования

### Стек технологий
- **Backend**: Python 3.10+, FastAPI, Celery, Redis
- **Frontend**: React 18+, TypeScript, Material-UI
- **Database**: PostgreSQL, Elasticsearch
- **Message Queue**: RabbitMQ, Kafka
- **Monitoring**: Prometheus, Grafana (интеграция с существующими)
- **Containerization**: Docker, Docker Compose

### Требования к производительности
- **Events/sec**: > 10,000
- **Alert latency**: < 100ms
- **Storage**: 90 дней retention
- **Uptime**: 99.9%

### Безопасность
- **Authentication**: Keycloak OIDC
- **Authorization**: RBAC per tenant
- **Encryption**: TLS 1.3, data at rest
- **Audit**: Все действия логируются
- **Compliance**: 152-ФЗ, 187-ФЗ

## 📋 План разработки

### Этап 1: Базовая интеграция (2 недели)
1. Настройка интеграции с существующим Prometheus
2. Создание базовых SIEM дашбордов
3. Разработка Event Collector
4. Базовая обработка событий

### Этап 2: Основная функциональность (3 недели)
1. Разработка Backend API
2. Реализация правил детекции
3. Система алертов и уведомлений
4. Multi-tenant поддержка

### Этап 3: Веб-интерфейс (2 недели)
1. Разработка React frontend
2. Multi-tenant UI
3. Дашборды и аналитика
4. Настройки и конфигурация

### Этап 4: Оптимизация и тестирование (1 неделя)
1. Performance optimization
2. Security testing
3. Load testing
4. Documentation

## 🎯 Критерии приемки

### Функциональные требования
- [ ] Сбор событий из всех источников (Relay, Keycloak, Django)
- [ ] Обработка > 10,000 событий/сек
- [ ] Multi-tenant изоляция данных
- [ ] Правила детекции угроз
- [ ] Система алертов и уведомлений
- [ ] Веб-интерфейс с дашбордами
- [ ] Интеграция с существующим мониторингом

### Технические требования
- [ ] Интеграция с Prometheus/Grafana/AlertManager
- [ ] Поддержка Docker deployment
- [ ] SSL/TLS шифрование
- [ ] Backup и восстановление
- [ ] Логирование и аудит
- [ ] Соответствие 152-ФЗ, 187-ФЗ

### Производительность
- [ ] Latency < 100ms для алертов
- [ ] Uptime > 99.9%
- [ ] Memory usage < 8GB
- [ ] Storage optimization

## 📚 Документация

### Обязательная документация
- [ ] API документация (OpenAPI/Swagger)
- [ ] Руководство по развертыванию
- [ ] Руководство пользователя
- [ ] Troubleshooting guide
- [ ] Security hardening guide

### Код документация
- [ ] Docstrings для всех функций
- [ ] README для каждого компонента
- [ ] Комментарии к конфигурации
- [ ] Примеры использования

## 🔗 Интеграция с существующими системами

### Prometheus Integration
```yaml
# Добавить в prometheus/prometheus.yml:
scrape_configs:
  - job_name: 'siem-backend'
    static_configs:
      - targets: ['siem-server:8000']
    metrics_path: '/metrics'
    scrape_interval: 15s
```

### Grafana Integration
```yaml
# Создать дашборды в deploy/grafana/dashboards/:
# - siem_events_dashboard.json
# - siem_alerts_dashboard.json
# - siem_correlation_dashboard.json
```

### AlertManager Integration
```yaml
# Добавить в deploy/alertmanager/alertmanager.yml:
receivers:
  - name: 'siem-critical'
    email_configs:
    - to: "security@2gc.ru"
    telegram_configs:
    - bot_token: "${TELEGRAM_BOT_TOKEN}"
      chat_id: "${TELEGRAM_CHAT_ID}"
```

## 💡 Рекомендации

### Архитектурные решения
1. **Использовать существующую инфраструктуру** мониторинга
2. **Развертывать SIEM на отдельном сервере** для производительности
3. **Обеспечить multi-tenant изоляцию** данных
4. **Интегрировать с Keycloak** для аутентификации
5. **Использовать event-driven архитектуру** для масштабируемости

### Технические решения
1. **FastAPI** для высокопроизводительного API
2. **Elasticsearch** для хранения и поиска событий
3. **Kafka** для потоковой обработки
4. **Redis** для кэширования и сессий
5. **Celery** для асинхронных задач

### Безопасность
1. **TLS 1.3** для всех соединений
2. **RBAC** для контроля доступа
3. **Audit logging** для всех действий
4. **Data encryption** at rest
5. **Regular security updates**

---

**Срок выполнения**: 8 недель  
**Приоритет**: Высокий  
**Статус**: В разработке 