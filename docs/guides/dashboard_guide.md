# CloudBridge Relay Dashboard Guide

## Обзор

CloudBridge Relay v2.0 предоставляет набор специализированных дашбордов для комплексного мониторинга и управления системой туннелирования. Каждый дашборд фокусируется на конкретной области и предоставляет детальную аналитику.

## Доступные дашборды

### 1. CloudBridge Relay Dashboard (Основной)
**Файл:** `deploy/grafana/dashboards/relay_dashboard.json`

Основной дашборд с общими метриками системы:
- Активные соединения и туннели
- Производительность Keycloak аутентификации
- Технические метрики (память, CPU, сеть)
- Бизнес-метрики (пользователи, доход, удовлетворенность)

### 2. Tunnel Management Dashboard
**Файл:** `deploy/grafana/dashboards/tunnel_management_dashboard.json`

Специализированный дашборд для управления туннелями:

#### Основные возможности:
- **Обзор туннелей**: Количество активных туннелей, скорость создания/закрытия
- **Анализ трафика**: Передача данных в обоих направлениях, использование полосы пропускания
- **Здоровье туннелей**: Статус, качество соединения, ошибки
- **Управление ресурсами**: Использование памяти и CPU по туннелям
- **Операции**: Управление туннелями, очистка неактивных

#### Ключевые метрики:
- `relay_tunnel_creations_total` - Общее количество созданных туннелей
- `relay_tunnel_closures_total` - Общее количество закрытых туннелей
- `relay_tunnel_bytes_from_client_total` - Байты от клиента к серверу
- `relay_tunnel_bytes_to_client_total` - Байты от сервера к клиенту
- `relay_tunnel_bandwidth_utilization_percent` - Использование полосы пропускания
- `relay_tunnel_latency_seconds` - Задержка туннелей

### 3. Security Analytics Dashboard
**Файл:** `deploy/grafana/dashboards/security_analytics_dashboard.json`

Дашборд для анализа безопасности и угроз:

#### Основные возможности:
- **Аутентификация**: Успешность аутентификации, неудачные попытки
- **Производительность токенов**: Кэш JWKS, время валидации
- **Анализ разрешений**: Проверки разрешений, отказы
- **Обнаружение угроз**: Подозрительная активность, DDoS атаки
- **ML анализ**: Аномалии, предсказания угроз
- **TLS безопасность**: Статус сертификатов, ошибки handshake

#### Ключевые метрики:
- `relay_keycloak_token_validations_total` - Валидации токенов
- `relay_security_suspicious_activities_total` - Подозрительная активность
- `relay_security_rate_limit_violations_total` - Нарушения rate limiting
- `relay_ml_anomaly_score` - ML оценки аномалий
- `relay_tls_certificate_status` - Статус TLS сертификатов

### 4. Multi-Tenant Analytics Dashboard
**Файл:** `deploy/grafana/dashboards/multi_tenant_analytics_dashboard.json`

Дашборд для анализа мультитенантности:

#### Основные возможности:
- **Обзор тенантов**: Активные тенанты, общий доход, удовлетворенность
- **Использование ресурсов**: Туннели, соединения, полоса пропускания по тенантам
- **Производительность**: Время отклика, ошибки, SLA
- **Безопасность**: Инциденты, нарушения изоляции, оценки рисков
- **Бизнес-аналитика**: Доход, удовлетворенность клиентов, тренды использования

#### Ключевые метрики:
- `relay_tenant_active_count` - Количество активных тенантов
- `relay_tenant_active_tunnels` - Активные туннели по тенантам
- `relay_tenant_bandwidth_bytes_total` - Использование полосы пропускания
- `relay_tenant_sla_compliance_percent` - Соответствие SLA
- `relay_tenant_risk_score` - Оценки рисков по тенантам

## Установка и настройка

### 1. Автоматическая установка

Дашборды автоматически устанавливаются при запуске системы:

```bash
# Запуск всех сервисов
docker-compose up -d

# Дашборды будут доступны в Grafana
# URL: http://localhost:3000
# Логин: admin/admin
```

### 2. Ручная установка

Для ручной установки дашбордов:

```bash
# Копирование дашбордов в Grafana
cp deploy/grafana/dashboards/*.json /var/lib/grafana/dashboards/

# Перезапуск Grafana
docker-compose restart grafana
```

### 3. Настройка Prometheus

Убедитесь, что Prometheus настроен для сбора метрик:

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'relay'
    static_configs:
      - targets: ['relay:8081']
    metrics_path: '/metrics'
```

## Использование дашбордов

### Фильтрация и переменные

Все дашборды поддерживают переменные для фильтрации:

- **Tunnel Management**: `tunnel_id`, `remote_host`
- **Security Analytics**: `user_id`, `resource`, `severity`
- **Multi-Tenant Analytics**: `tenant_id`, `tenant_name`, `sla_type`

### Временные диапазоны

Рекомендуемые временные диапазоны:
- **Краткосрочный анализ**: 1-6 часов
- **Среднесрочный анализ**: 1-7 дней
- **Долгосрочный анализ**: 1-30 дней

### Алерты

Настройте алерты для критических метрик:

```yaml
# alertmanager.yml
groups:
  - name: relay_alerts
    rules:
      - alert: HighErrorRate
        expr: rate(relay_errors_total[5m]) > 0.1
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High error rate detected"
          
      - alert: TunnelCreationFailure
        expr: rate(relay_tunnel_creations_total[5m]) == 0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "No tunnel creations detected"
```

## Расширенные возможности

### 1. Экспорт данных

Экспорт метрик для внешнего анализа:

```bash
# Экспорт метрик Prometheus
curl http://localhost:9090/api/v1/export -d 'match[]={job="relay"}' > metrics.json

# Экспорт дашбордов Grafana
curl -H "Authorization: Bearer $GRAFANA_TOKEN" \
  http://localhost:3000/api/dashboards/db/relay-dashboard > dashboard.json
```

### 2. Интеграция с внешними системами

#### SIEM интеграция:
```python
import requests
import json

# Отправка метрик в SIEM
def send_to_siem(metrics):
    siem_url = "https://siem.company.com/api/events"
    headers = {"Authorization": "Bearer $SIEM_TOKEN"}
    
    for metric in metrics:
        requests.post(siem_url, json=metric, headers=headers)
```

#### Webhook уведомления:
```yaml
# grafana.ini
[unified_alerting]
enabled = true

[alerting]
enabled = true
```

### 3. Кастомизация дашбордов

#### Добавление новых панелей:

```json
{
  "title": "Custom Metric",
  "type": "graph",
  "targets": [
    {
      "expr": "custom_metric_name",
      "legendFormat": "{{label}}"
    }
  ],
  "gridPos": {"h": 8, "w": 12, "x": 0, "y": 0}
}
```

#### Создание новых переменных:

```json
{
  "name": "custom_variable",
  "type": "query",
  "query": "label_values(metric_name, label_name)",
  "refresh": 2,
  "includeAll": true,
  "multi": true
}
```

## Мониторинг производительности

### Ключевые KPI

1. **Доступность системы**: > 99.9%
2. **Время отклика**: < 100ms (95th percentile)
3. **Успешность аутентификации**: > 99%
4. **Использование ресурсов**: < 80%

### Метрики для отслеживания

```promql
# Доступность
up{job="relay"}

# Время отклика
histogram_quantile(0.95, rate(relay_connection_duration_seconds_bucket[5m]))

# Успешность аутентификации
rate(relay_keycloak_token_validations_total{result="success"}[5m]) / 
rate(relay_keycloak_token_validations_total[5m]) * 100

# Использование ресурсов
relay_technical_memory_usage_bytes / 1024 / 1024
```

## Устранение неполадок

### Частые проблемы

1. **Дашборды не загружаются**
   - Проверьте доступность Grafana
   - Убедитесь в правильности JSON файлов
   - Проверьте права доступа к файлам

2. **Метрики не отображаются**
   - Проверьте подключение Prometheus к relay
   - Убедитесь в доступности `/metrics` endpoint
   - Проверьте конфигурацию scrape_configs

3. **Медленная загрузка**
   - Оптимизируйте запросы PromQL
   - Увеличьте ресурсы для Grafana
   - Настройте кэширование

### Диагностика

```bash
# Проверка метрик relay
curl http://localhost:8081/metrics | grep relay_

# Проверка статуса Prometheus
curl http://localhost:9090/api/v1/targets

# Проверка статуса Grafana
curl http://localhost:3000/api/health
```

## Заключение

Дашборды CloudBridge Relay предоставляют комплексное решение для мониторинга и управления системой туннелирования. Регулярное использование этих инструментов поможет обеспечить стабильную работу системы и быстрое выявление проблем. 