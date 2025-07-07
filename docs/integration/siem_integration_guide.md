# Практическое руководство по интеграции SIEM

## 🎯 Цель

Это руководство предоставляет пошаговые инструкции по интеграции SIEM системы с CloudBridge Relay v2.0 для обеспечения комплексного мониторинга безопасности в multi-tenant окружении.

## 📋 Предварительные требования

### Системные требования
- Docker 20.10+ и Docker Compose 2.0+
- 8+ GB RAM (рекомендуется 16+ GB для production)
- 50+ GB свободного места на диске
- Linux с поддержкой Docker

### Сетевая конфигурация
- Доступ к портам: 5601 (Kibana), 9200 (Elasticsearch), 9092 (Kafka)
- Настроенный firewall для защиты SIEM компонентов
- SSL сертификаты для production окружения

## 🚀 Этап 1: Базовая установка SIEM

### 1.1. Подготовка окружения

```bash
# Переход в директорию SIEM
cd siem

# Создание необходимых директорий
mkdir -p {logs,config,data,backup}

# Копирование конфигурации
cp config/siem_config.yaml config/siem/

# Установка зависимостей Python
pip install -r requirements.txt
```

### 1.2. Настройка конфигурации

```yaml
# config/siem/siem_config.yaml
siem:
  # Основные настройки
  environment: "production"
  log_level: "INFO"
  
  # Настройки сбора данных
  collection:
    relay_logs_interval: 60
    relay_metrics_interval: 30
    keycloak_events_interval: 120
    django_audit_interval: 60
    
  # Настройки обработки
  processing:
    batch_size: 1000
    max_workers: 4
    retry_attempts: 3
    
  # Настройки хранения
  storage:
    elasticsearch:
      hosts: ["http://elasticsearch:9200"]
      index_prefix: "siem"
      retention_days: 90
      
  # Настройки уведомлений
  alerting:
    min_severity: "medium"
    channels:
      telegram:
        enabled: true
        bot_token: "${TELEGRAM_BOT_TOKEN}"
        chat_id: "${TELEGRAM_CHAT_ID}"
      email:
        enabled: false
        smtp_server: "${EMAIL_SMTP_SERVER}"
        username: "${EMAIL_USERNAME}"
        password: "${EMAIL_PASSWORD}"
```

### 1.3. Запуск базовой SIEM

```bash
# Запуск всех сервисов
./start_siem.sh

# Проверка статуса
docker-compose -f docker-compose.siem.yml ps

# Проверка логов
docker-compose -f docker-compose.siem.yml logs -f
```

## 🔧 Этап 2: Multi-tenant конфигурация

### 2.1. Настройка Elasticsearch для multi-tenancy

```yaml
# config/siem/elasticsearch_multi_tenant.yml
elasticsearch:
  indices:
    pattern: "siem-{tenant_id}-{date}"
    settings:
      number_of_shards: 3
      number_of_replicas: 1
      refresh_interval: "1s"
      
  mappings:
    tenant_id:
      type: "keyword"
      required: true
    event_type:
      type: "keyword"
    severity:
      type: "keyword"
    timestamp:
      type: "date"
    source_ip:
      type: "ip"
    user_id:
      type: "keyword"
    action:
      type: "keyword"
```

### 2.2. Конфигурация Kafka для изоляции данных

```yaml
# config/siem/kafka_multi_tenant.yml
kafka:
  topics:
    pattern: "siem-{tenant_id}-{event_type}"
    partitions: 3
    replication_factor: 1
    
  consumer_groups:
    pattern: "siem-processor-{tenant_id}"
    
  security:
    sasl_enabled: false
    ssl_enabled: false
```

## 🛡️ Этап 3: Настройка правил детекции

### 3.1. Tenant-specific правила

```yaml
# config/siem/detection_rules.yaml
detection_rules:
  tenant_specific:
    brute_force:
      enabled: true
      conditions:
        event_type: "authentication_failure"
        time_window: 300  # 5 минут
        threshold: 5
        group_by: ["source_ip", "tenant_id"]
      actions:
        - type: "alert"
          severity: "high"
          message: "Brute force attack detected for tenant {tenant_id}"
        - type: "block_ip"
          duration: 3600  # 1 час
          
    data_exfiltration:
      enabled: true
      conditions:
        event_type: "data_transfer"
        data_volume_gb: 100
        time_window: 3600  # 1 час
        group_by: ["tenant_id", "user_id"]
      actions:
        - type: "alert"
          severity: "critical"
          message: "Potential data exfiltration detected"
        - type: "quarantine_user"
          duration: 7200  # 2 часа
```

### 3.2. Cross-tenant правила

```yaml
# config/siem/cross_tenant_rules.yaml
cross_tenant_rules:
  shared_ip_attacks:
    enabled: true
    conditions:
      event_type: "authentication_failure"
      correlation_window: 600  # 10 минут
      min_tenants_affected: 2
      group_by: ["source_ip"]
    actions:
      - type: "global_alert"
        severity: "critical"
        message: "Cross-tenant attack detected from IP {source_ip}"
      - type: "global_block_ip"
        duration: 86400  # 24 часа
```

## 🔔 Этап 4: Настройка алертов и уведомлений

### 4.1. Конфигурация каналов уведомлений

```yaml
# config/siem/alerting.yaml
alerting:
  channels:
    telegram:
      enabled: true
      bot_token: "${TELEGRAM_BOT_TOKEN}"
      chat_id: "${TELEGRAM_CHAT_ID}"
      template: |
        🚨 {severity} Alert: {title}
        
        📊 Details:
        • Tenant: {tenant_id}
        • Source: {source_ip}
        • User: {user_id}
        • Time: {timestamp}
        
        📝 Description: {description}
        
        🔗 View in SIEM: {siem_url}
        
    email:
      enabled: true
      smtp_server: "${EMAIL_SMTP_SERVER}"
      smtp_port: 587
      username: "${EMAIL_USERNAME}"
      password: "${EMAIL_PASSWORD}"
      from_address: "siem@yourcompany.com"
      to_addresses: ["security@yourcompany.com", "admin@yourcompany.com"]
      subject_template: "[SIEM Alert] {severity}: {title}"
      
  escalation:
    enabled: true
    levels:
      - level: 1
        delay_minutes: 5
        channels: ["telegram"]
      - level: 2
        delay_minutes: 15
        channels: ["telegram", "email"]
      - level: 3
        delay_minutes: 30
        channels: ["telegram", "email", "slack"]
```

## 📊 Этап 5: Мониторинг и дашборды

### 5.1. Настройка Grafana дашбордов

```json
// grafana/dashboards/siem_overview.json
{
  "dashboard": {
    "title": "SIEM Overview - Multi-tenant",
    "panels": [
      {
        "title": "Events per Tenant",
        "type": "stat",
        "targets": [
          {
            "expr": "sum(rate(siem_events_total[5m])) by (tenant_id)",
            "legendFormat": "{{tenant_id}}"
          }
        ]
      },
      {
        "title": "Alerts by Severity",
        "type": "piechart",
        "targets": [
          {
            "expr": "sum(siem_alerts_total) by (severity)",
            "legendFormat": "{{severity}}"
          }
        ]
      }
    ]
  }
}
```

## 🔧 Этап 6: Интеграция с CloudBridge Relay

### 6.1. Настройка Relay для отправки событий в SIEM

```go
// internal/siem/client.go
package siem

import (
    "encoding/json"
    "net/http"
    "time"
)

type SIEMClient struct {
    baseURL    string
    httpClient *http.Client
}

type SIEMEvent struct {
    TenantID    string                 `json:"tenant_id"`
    EventType   string                 `json:"event_type"`
    Severity    string                 `json:"severity"`
    Timestamp   time.Time              `json:"timestamp"`
    SourceIP    string                 `json:"source_ip"`
    UserID      string                 `json:"user_id"`
    Action      string                 `json:"action"`
    Metadata    map[string]interface{} `json:"metadata"`
}

func NewSIEMClient(baseURL string) *SIEMClient {
    return &SIEMClient{
        baseURL: baseURL,
        httpClient: &http.Client{
            Timeout: 10 * time.Second,
        },
    }
}

func (c *SIEMClient) SendEvent(event SIEMEvent) error {
    eventJSON, err := json.Marshal(event)
    if err != nil {
        return err
    }
    
    resp, err := c.httpClient.Post(
        c.baseURL+"/api/v1/events",
        "application/json",
        bytes.NewBuffer(eventJSON),
    )
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    return nil
}
```

## 🧪 Этап 7: Тестирование и валидация

### 7.1. Тестовые сценарии

```python
# tests/test_siem_integration.py
import pytest
import requests
import time
from typing import Dict

class TestSIEMIntegration:
    def setup_method(self):
        self.siem_url = "http://localhost:8081"
        self.elasticsearch_url = "http://localhost:9200"
        
    def test_event_collection(self):
        """Тест сбора событий"""
        # Генерация тестового события
        test_event = {
            "tenant_id": "test_tenant",
            "event_type": "test_event",
            "severity": "medium",
            "timestamp": time.time(),
            "source_ip": "192.168.1.100",
            "user_id": "test_user",
            "action": "test_action"
        }
        
        # Отправка события
        response = requests.post(
            f"{self.siem_url}/api/v1/events",
            json=test_event
        )
        assert response.status_code == 200
        
        # Проверка сохранения в Elasticsearch
        time.sleep(5)  # Ждем обработки
        es_response = requests.get(
            f"{self.elasticsearch_url}/siem-test_tenant-*/_search",
            json={"query": {"term": {"user_id": "test_user"}}}
        )
        assert es_response.status_code == 200
        
        results = es_response.json()
        assert len(results['hits']['hits']) > 0
```

## 🔧 Этап 8: Production развертывание

### 8.1. Настройка production окружения

```yaml
# docker-compose.siem.prod.yml
version: '3.8'

services:
  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.11.0
    environment:
      - discovery.type=multi-node
      - cluster.name=siem-cluster
      - node.name=siem-node-1
      - xpack.security.enabled=true
      - ELASTIC_PASSWORD=${ELASTICSEARCH_PASSWORD}
    volumes:
      - elasticsearch_data:/usr/share/elasticsearch/data
      - elasticsearch_config:/usr/share/elasticsearch/config
    networks:
      - siem-network
    deploy:
      resources:
        limits:
          memory: 4G
        reservations:
          memory: 2G
          
  kafka:
    image: confluentinc/cp-kafka:latest
    environment:
      KAFKA_BROKER_ID: 1
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://kafka:9092
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 3
      KAFKA_DEFAULT_REPLICATION_FACTOR: 3
      KAFKA_MIN_INSYNC_REPLICAS: 2
    deploy:
      replicas: 3
```

## 🔍 Troubleshooting

### Общие проблемы и решения

#### 1. SIEM не собирает события
```bash
# Проверка статуса коллекторов
docker-compose -f docker-compose.siem.yml logs siem-collector

# Проверка подключения к источникам данных
curl -X GET http://relay:8081/metrics
curl -X GET http://keycloak:8080/admin/events

# Проверка Kafka
docker exec -it siem-kafka kafka-topics --list --bootstrap-server localhost:9092
```

#### 2. Высокая задержка обработки
```bash
# Проверка ресурсов Elasticsearch
curl -X GET http://elasticsearch:9200/_cluster/health
curl -X GET http://elasticsearch:9200/_nodes/stats

# Оптимизация индексов
curl -X POST http://elasticsearch:9200/siem-*/_forcemerge
```

## 📈 Оптимизация производительности

### 1. Настройка Elasticsearch
```yaml
# config/elasticsearch.yml
cluster:
  name: siem-cluster
  
node:
  name: siem-node-1
  
path:
  data: /usr/share/elasticsearch/data
  
network:
  host: 0.0.0.0
  
discovery:
  type: single-node
  
xpack:
  security:
    enabled: true
    
indices:
  memory:
    index_buffer_size: 30%
    
thread_pool:
  write:
    queue_size: 1000
    max_queue_size: 2000
```

### 2. Мониторинг производительности
```yaml
# grafana/dashboards/siem_performance.json
{
  "dashboard": {
    "title": "SIEM Performance",
    "panels": [
      {
        "title": "Events Processing Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(siem_events_processed_total[5m])",
            "legendFormat": "Events/sec"
          }
        ]
      },
      {
        "title": "Processing Latency",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(siem_processing_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          }
        ]
      }
    ]
  }
}
```

## 🔗 Заключение

Это руководство предоставляет комплексный подход к интеграции SIEM с CloudBridge Relay v2.0. Ключевые моменты:

1. **Поэтапная реализация** - от базовой установки до production
2. **Multi-tenant изоляция** - полная изоляция данных между тенантами
3. **Масштабируемость** - горизонтальное масштабирование компонентов
4. **Безопасность** - соответствие российским стандартам
5. **Мониторинг** - комплексное наблюдение за производительностью

Для успешной интеграции рекомендуется:
- Тщательное тестирование на каждом этапе
- Мониторинг производительности и ресурсов
- Регулярное обновление правил детекции
- Непрерывное улучшение на основе метрик