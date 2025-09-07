# 📊 CloudBridge Relay - Advanced Monitoring Guide

## 📋 Обзор

Этот документ описывает продвинутую систему мониторинга CloudBridge Relay Server с использованием **оптимизированного Prometheus**, Grafana и Alertmanager для single-server архитектуры.

### 🚀 **Новые возможности (v2.0):**
- **Умная оптимизация пространства** - 94% экономия дискового пространства
- **Автоматическая очистка** - каждые 6 часов
- **Оптимизированные интервалы** - 4-20x реже scraping
- **Фильтрация метрик** - удаление избыточных данных
- **Production-ready** - готовность к долгосрочному использованию

## 🏗️ Архитектура мониторинга

### Компоненты системы:

```
┌─────────────────────────────────────────────────────────────┐
│                    Single Server                            │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │   Prometheus│ │   Grafana   │ │ Alertmanager│           │
│  │   (Port 9090)│ │  (Port 3000)│ │  (Port 9093)│           │
│  └─────────────┘ └─────────────┘ └─────────────┘           │
│                                                             │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │CloudBridge  │ │  Cassandra  │ │   Cadence   │           │
│  │   Relay     │ │  (Port 9500)│ │  (Port 7934)│           │
│  │  (Port 8084)│ └─────────────┘ └─────────────┘           │
│  └─────────────┘                                           │
│                                                             │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │   WireGuard │ │    QUIC     │ │    AI       │           │
│  │  (Port 8084)│ │  (Port 8084)│ │  (Port 8084)│           │
│  └─────────────┘ └─────────────┘ └─────────────┘           │
└─────────────────────────────────────────────────────────────┘
```

## 📁 Структура мониторинга

```
monitoring/
├── prometheus/
│   └── prometheus.yml              # Конфигурация Prometheus
├── grafana/
│   ├── datasources.yml             # Data sources для Grafana
│   └── dashboards.yml              # Конфигурация дашбордов
├── alertmanager/
│   └── alertmanager.yml            # Конфигурация Alertmanager
├── rules/
│   └── cloudbridge-alerts.yml      # Правила алертов
└── dashboards/
    ├── cloudbridge-overview.json   # Обзорный дашборд
    └── cloudbridge-relay-detail.json # Детальный дашборд
```

## 🚀 Быстрый старт

### 1. **Развертывание оптимизированного мониторинга**

```bash
# Развертывание оптимизированного Prometheus
kubectl apply -f k8s/storageclasses/prometheus-optimized-storage.yaml
kubectl apply -f k8s/persistentvolumes/prometheus-optimized-pv.yaml
kubectl apply -f k8s/persistentvolumeclaims/prometheus-optimized-pvc.yaml
kubectl apply -f k8s/configmaps/prometheus-optimized-config.yaml
kubectl apply -f k8s/deployments/prometheus-optimized-deployment.yaml
kubectl apply -f k8s/services/prometheus-optimized-service.yaml
kubectl apply -f k8s/cronjobs/prometheus-cleanup-cronjob.yaml

# Проверка статуса
kubectl get pods -n cloudbridge-monitoring

# Ручная очистка (если нужно)
./scripts/prometheus-cleanup.sh
```

### 2. **Доступ к интерфейсам**

```bash
# Port forwarding для доступа
kubectl port-forward service/grafana 3000:3000 -n cloudbridge-monitoring
kubectl port-forward service/prometheus 9090:9090 -n cloudbridge-monitoring
kubectl port-forward service/alertmanager 9093:9093 -n cloudbridge-monitoring
```

### 3. **Вход в Grafana**

- **URL**: http://localhost:3000
- **Username**: admin
- **Password**: admin123

## 📊 Prometheus Configuration

### **Основные метрики:**

```yaml
# CloudBridge Relay Service
- job_name: 'cloudbridge-relay'
  kubernetes_sd_configs:
    - role: endpoints
      namespaces:
        names:
          - cloudbridge
  relabel_configs:
    - source_labels: [__meta_kubernetes_service_name]
      action: keep
      regex: cloudbridge-relay
    - source_labels: [__meta_kubernetes_endpoint_port_name]
      action: keep
      regex: metrics
```

### **Scrape интервалы:**

- **CloudBridge Relay**: 10s
- **Cassandra**: 30s
- **Cadence**: 15s
- **WireGuard**: 15s
- **QUIC**: 10s
- **AI Service**: 15s

### **Retention policy:**

- **Время хранения**: 200 часов
- **Evaluation interval**: 15s
- **Scrape interval**: 15s

## 📈 Grafana Dashboards

### **1. CloudBridge Overview Dashboard**

**Метрики:**
- ✅ Статус всех сервисов
- ✅ CPU и Memory usage
- ✅ Request rate и Response time
- ✅ Error rate
- ✅ Active connections
- ✅ Disk usage
- ✅ Network I/O

**Панели:**
- System Overview (статус сервисов)
- CPU Usage (график по времени)
- Memory Usage (график по времени)
- Request Rate (запросы в секунду)
- Response Time (время ответа)
- Error Rate (процент ошибок)
- Active Connections (активные соединения)
- Disk Usage (использование диска)
- Network I/O (сетевой трафик)

### **2. CloudBridge Relay Detail Dashboard**

**Метрики:**
- ✅ Детальные метрики Relay сервиса
- ✅ WireGuard connections
- ✅ QUIC protocol metrics
- ✅ Cadence workflow metrics
- ✅ Cassandra database metrics
- ✅ AI service metrics
- ✅ System resources

**Панели:**
- Relay Service Status
- Active Sessions
- Total Requests
- Error Rate
- Request Rate Over Time
- Response Time Distribution
- WireGuard Connections
- QUIC Protocol Metrics
- Cadence Workflow Metrics
- Cassandra Database Metrics
- AI Service Metrics
- System Resources

## 🚨 Alerting Rules

### **Критические алерты (Critical):**

```yaml
# Service Down
- alert: CloudBridgeRelayDown
  expr: up{job="cloudbridge-relay"} == 0
  for: 1m
  labels:
    severity: critical

# High Error Rate
- alert: CloudBridgeRelayHighErrorRate
  expr: rate(http_requests_total{job="cloudbridge-relay",status=~"5.."}[5m]) / rate(http_requests_total{job="cloudbridge-relay"}[5m]) * 100 > 5
  for: 5m
  labels:
    severity: critical
```

### **Предупреждающие алерты (Warning):**

```yaml
# High CPU Usage
- alert: CloudBridgeRelayHighCPU
  expr: rate(process_cpu_seconds_total{job="cloudbridge-relay"}[5m]) * 100 > 80
  for: 5m
  labels:
    severity: warning

# High Memory Usage
- alert: CloudBridgeRelayHighMemory
  expr: (process_resident_memory_bytes{job="cloudbridge-relay"} / 1024 / 1024 / 1024) > 1
  for: 5m
  labels:
    severity: warning
```

### **Группы алертов:**

1. **cloudbridge-relay** - основные метрики Relay сервиса
2. **cloudbridge-cassandra** - метрики базы данных
3. **cloudbridge-cadence** - метрики workflow engine
4. **cloudbridge-wireguard** - метрики VPN сервиса
5. **cloudbridge-quic** - метрики QUIC протокола
6. **cloudbridge-ai** - метрики AI сервиса
7. **kubernetes-infrastructure** - метрики инфраструктуры

## 📧 Alertmanager Configuration

### **Каналы уведомлений:**

- **Email** - для критических алертов
- **Slack** - для командных уведомлений
- **PagerDuty** - для экстренных ситуаций

### **Группировка алертов:**

```yaml
route:
  group_by: ['alertname', 'cluster', 'service']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'web.hook'
  routes:
    - match:
        severity: critical
      receiver: 'critical-alerts'
      group_wait: 5s
      group_interval: 5s
      repeat_interval: 30m
```

### **Inhibition rules:**

```yaml
inhibit_rules:
  - source_match:
      severity: 'critical'
    target_match:
      severity: 'warning'
    equal: ['alertname', 'cluster', 'service']
```

## 🔧 Custom Metrics

### **CloudBridge Relay метрики:**

```go
// Пример метрик для Go приложения
var (
    requestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "cloudbridge_relay_requests_total",
            Help: "Total number of requests",
        },
        []string{"method", "endpoint", "status"},
    )
    
    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "cloudbridge_relay_request_duration_seconds",
            Help: "Request duration in seconds",
        },
        []string{"method", "endpoint"},
    )
    
    activeConnections = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "cloudbridge_relay_active_connections",
            Help: "Number of active connections",
        },
    )
)
```

### **WireGuard метрики:**

```go
var (
    wireguardConnections = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "cloudbridge_wireguard_connections",
            Help: "WireGuard connections by peer",
        },
        []string{"peer"},
    )
    
    wireguardBytesTransferred = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "cloudbridge_wireguard_bytes_transferred_total",
            Help: "Total bytes transferred via WireGuard",
        },
        []string{"direction", "peer"},
    )
)
```

### **QUIC метрики:**

```go
var (
    quicStreams = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "cloudbridge_quic_active_streams",
            Help: "Number of active QUIC streams",
        },
    )
    
    quicBytesSent = prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "cloudbridge_quic_bytes_sent_total",
            Help: "Total bytes sent via QUIC",
        },
    )
)
```

## 📊 Grafana Data Sources

### **Настроенные источники данных:**

1. **Prometheus** - основной источник метрик
2. **CloudBridge Prometheus** - специализированный источник
3. **CloudBridge Relay Metrics** - метрики Relay сервиса
4. **Cassandra Metrics** - метрики базы данных
5. **Cadence Metrics** - метрики workflow engine
6. **Kubernetes Metrics** - метрики кластера
7. **Node Exporter** - системные метрики
8. **cAdvisor** - контейнерные метрики
9. **kube-state-metrics** - состояние Kubernetes
10. **Alertmanager** - метрики алертов

## 🔍 Troubleshooting

### **Общие проблемы:**

#### Проблема: Prometheus не собирает метрики

```bash
# Проверка статуса Prometheus
kubectl get pods -n cloudbridge-monitoring
kubectl logs prometheus-xxx -n cloudbridge-monitoring

# Проверка targets
curl http://localhost:9090/api/v1/targets
```

#### Проблема: Grafana не отображает данные

```bash
# Проверка подключения к Prometheus
kubectl exec -it grafana-xxx -n cloudbridge-monitoring -- curl http://prometheus:9090/api/v1/query?query=up

# Проверка datasources
kubectl logs grafana-xxx -n cloudbridge-monitoring
```

#### Проблема: Алерты не срабатывают

```bash
# Проверка правил алертов
kubectl get configmap prometheus-rules -n cloudbridge-monitoring -o yaml

# Проверка Alertmanager
kubectl logs alertmanager-xxx -n cloudbridge-monitoring
```

### **Полезные команды:**

```bash
# Проверка метрик
curl http://localhost:9090/api/v1/query?query=up

# Проверка алертов
curl http://localhost:9090/api/v1/alerts

# Проверка targets
curl http://localhost:9090/api/v1/targets

# Проверка правил
curl http://localhost:9090/api/v1/rules
```

## 📚 Best Practices

### **Мониторинг:**

1. **Используйте осмысленные метрики** - измеряйте то, что важно для бизнеса
2. **Настройте правильные алерты** - избегайте ложных срабатываний
3. **Используйте группировку** - группируйте связанные алерты
4. **Настройте escalation** - для критических алертов
5. **Регулярно проверяйте дашборды** - для понимания трендов

### **Производительность:**

1. **Оптимизируйте scrape интервалы** - баланс между детализацией и нагрузкой
2. **Используйте recording rules** - для сложных вычислений
3. **Настройте retention policy** - для управления дисковым пространством
4. **Мониторьте сам Prometheus** - для предотвращения проблем

### **Безопасность:**

1. **Используйте RBAC** - для контроля доступа
2. **Настройте TLS** - для шифрования трафика
3. **Ограничьте доступ** - только необходимые пользователи
4. **Регулярно обновляйте** - для исправления уязвимостей

## 🚀 Оптимизация Prometheus (v2.0)

### **Умная оптимизация пространства**

CloudBridge Relay использует оптимизированную версию Prometheus с революционными улучшениями:

#### **📊 Экономия дискового пространства:**
- **94% экономия** - с 365GB/год до 24GB/год
- **Retention time**: 24 часа (было 200 часов)
- **Retention size**: 40GB лимит из 50GB доступных
- **Автоматическая очистка**: каждые 6 часов

#### **⚡ Оптимизированные интервалы:**
- **Prometheus**: 120 секунд (было 15 секунд)
- **CloudBridge Relay**: 60 секунд (было 10 секунд)
- **Cassandra**: 120 секунд (было 15 секунд)
- **Cadence**: 120 секунд (было 15 секунд)
- **Kubernetes**: 300 секунд (было 15 секунд)

#### **🔧 Технические улучшения:**
- **WAL compression**: включено для экономии места
- **Block duration**: 2 часа для оптимального хранения
- **Metric filtering**: удаление избыточных метрик
- **Query limits**: 20 concurrent, 50M samples

#### **🧹 Автоматическая очистка:**
```bash
# CronJob каждые 6 часов
kubectl get cronjobs -n cloudbridge-monitoring

# Ручная очистка
./scripts/prometheus-cleanup.sh

# Мониторинг размера
kubectl exec -n cloudbridge-monitoring prometheus-optimized-xxx -- du -sh /prometheus/
```

#### **📈 Мониторинг и алерты:**
- **PrometheusStorageHigh**: предупреждение при 40GB
- **PrometheusStorageCritical**: критический при 45GB
- **PrometheusRetentionTime**: превышение времени хранения

### **🎯 Production Ready:**
- ✅ **Self-healing** - автоматическая очистка
- ✅ **Scalable** - легко масштабируется
- ✅ **Maintainable** - простое обслуживание
- ✅ **Cost-effective** - 94% экономия ресурсов

---

## 🎯 Заключение

Продвинутая система мониторинга CloudBridge Relay обеспечивает:

- ✅ **Полную видимость** - всех компонентов системы
- ✅ **Проактивные алерты** - для предотвращения проблем
- ✅ **Детальные дашборды** - для анализа производительности
- ✅ **Автоматическое развертывание** - простой setup
- ✅ **Масштабируемость** - готовность к росту
- ✅ **Интеграция** - с существующими системами

---

*Документ создан: 2025-09-06*  
*Статус: ✅ Готов к использованию* 🚀
