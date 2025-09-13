# 🛠️ Руководство разработчика: AI система CloudBridge Relay

## 📚 Содержание

1. [Быстрый старт](#быстрый-старт)
2. [Архитектура системы](#архитектура-системы)
3. [API Reference](#api-reference)
4. [Cadence Workflows](#cadence-workflows)
5. [Разработка и тестирование](#разработка-и-тестирование)
6. [Развертывание](#развертывание)
7. [Мониторинг](#мониторинг)

## 🚀 Быстрый старт

### Предварительные требования

```bash
# Go 1.21+
go version

# Docker
docker --version

# Kubernetes (Minikube)
minikube version
kubectl version
```

### Запуск локальной разработки

```bash
# 1. Клонирование репозитория
git clone https://github.com/twogc/cloudbridge-relay-installer.git
cd cloudbridge-relay-installer

# 2. Установка зависимостей
go mod download

# 3. Запуск AI Service
go run cmd/ai-service/main.go

# 4. Запуск AI Worker
go run cmd/ai-worker/main.go
```

### Тестирование API

```bash
# Health check
curl http://localhost:8083/health

# Список моделей
curl http://localhost:8083/api/v1/ai/models

# Детекция аномалий
curl -X POST http://localhost:8083/api/v1/ai/anomaly/detect \
  -H "Content-Type: application/json" \
  -d '{"metrics": {"connections_per_second": 150, "bytes_per_second": 2000000}}'
```

## 🏗️ Архитектура системы

### Компоненты

```
┌─────────────────────────────────────────────────────────────┐
│                    AI Service (Port 8083)                 │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │   Anomaly   │  │  Predictive │  │   Routing   │        │
│  │  Detector   │  │  Analytics  │  │ Optimizer   │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
├─────────────────────────────────────────────────────────────┤
│                    Optimized Cache                         │
│                    Circuit Breaker                         │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    AI Worker (Port 8085)                  │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐                 │
│  │ Cadence Client  │  │ AI Activities   │                 │
│  │                 │  │                 │                 │
│  │ • Workflows     │  │ • Data Collection│                │
│  │ • Activities    │  │ • Analysis      │                │
│  │ • Monitoring    │  │ • Response      │                │
│  └─────────────────┘  └─────────────────┘                 │
└─────────────────────────────────────────────────────────────┘
```

### Структура проекта

```
internal/ai/
├── anomaly/                 # Детекция аномалий
│   ├── detector.go         # Основной детектор
│   ├── isolation_forest.go # Isolation Forest алгоритм
│   └── optimized_detector.go # Оптимизированная версия
├── predictive/             # Предсказательная аналитика
│   ├── load_predictor.go  # Основной предсказатель
│   ├── arima.go          # ARIMA модель
│   └── optimized_predictor.go # Оптимизированная версия
├── routing/               # Оптимизация маршрутов
│   ├── optimizer.go      # Основной оптимизатор
│   ├── path_finder.go    # Алгоритмы поиска пути
│   └── optimized_optimizer.go # Оптимизированная версия
├── cache/                 # Кэширование
│   └── optimized_cache.go # LRU кэш с TTL
├── service.go            # Основной AI сервис
├── optimized_service.go  # Оптимизированный сервис
└── config.go            # Конфигурация

cmd/
├── ai-service/           # AI Service приложение
│   ├── main.go
│   └── handlers.go      # HTTP handlers
└── ai-worker/           # AI Worker приложение
    └── main.go

internal/cadence/
├── workflows/
│   └── ai.go           # AI workflows
├── client.go           # Cadence client
└── worker/
    └── worker.go       # Cadence worker
```

## 📖 API Reference

### AI Service API

#### Health Check
```http
GET /health
```

**Response:**
```json
{
  "service": "ai",
  "status": "healthy"
}
```

#### Models List
```http
GET /api/v1/ai/models
```

**Response:**
```json
{
  "models": [
    {
      "id": "anomaly-detector",
      "name": "Network Anomaly Detector",
      "type": "isolation_forest",
      "accuracy": 0.95,
      "status": "active",
      "last_trained": "2025-09-13T15:50:41Z"
    }
  ]
}
```

#### Anomaly Detection
```http
POST /api/v1/ai/anomaly/detect
Content-Type: application/json
```

**Request:**
```json
{
  "metrics": {
    "timestamp": "2025-09-13T15:50:00Z",
    "connections_per_second": 150,
    "bytes_per_second": 2000000,
    "packet_loss_rate": 0.05,
    "average_latency": 45.0,
    "cpu_usage": 65.0,
    "memory_usage": 70.0,
    "active_connections": 200,
    "failed_connections": 8,
    "protocol_distribution": {
      "QUIC": 60,
      "WireGuard": 25,
      "TCP": 15
    }
  }
}
```

**Response:**
```json
{
  "result": {
    "score": 0.3,
    "is_anomaly": false,
    "timestamp": "2025-09-13T15:50:59Z",
    "features": {
      "connections": 100,
      "bytes": 1000000,
      "loss": 0.01,
      "latency": 50,
      "cpu": 45,
      "memory": 60
    },
    "confidence": 0.85,
    "anomaly_type": "normal",
    "severity": "low"
  }
}
```

#### Load Prediction
```http
POST /api/v1/ai/predict/load
Content-Type: application/json
```

**Request:**
```json
{
  "horizon": 24,
  "confidence_level": 0.8
}
```

**Response:**
```json
{
  "prediction": {
    "forecast": [45, 50, 55, 60, 65],
    "confidence": 0.8,
    "horizon": 24,
    "timestamp": "2025-09-13T15:50:59Z",
    "confidence_interval": {
      "lower": [40, 45, 50, 55, 60],
      "upper": [50, 55, 60, 65, 70]
    },
    "model_id": "load-predictor-v1",
    "features": {
      "historical_load": 45,
      "time_of_day": 14,
      "day_of_week": 3
    },
    "processing_time": "45ms",
    "seasonality": true,
    "trend": "increasing"
  }
}
```

#### Route Optimization
```http
POST /api/v1/ai/route/optimize
Content-Type: application/json
```

**Request:**
```json
{
  "source": "node1",
  "destination": "node2",
  "constraints": {
    "max_latency": 100,
    "max_cost": 50,
    "min_reliability": 0.9
  }
}
```

**Response:**
```json
{
  "route": {
    "protocol": "QUIC",
    "path": ["node1", "node1", "node2", "node2"],
    "predicted_latency": 25.5,
    "confidence": 0.9,
    "timestamp": "2025-09-13T15:50:59Z"
  }
}
```

## 🔄 Cadence Workflows

### ThreatAnalysisWorkflow

Анализ угроз в реальном времени с использованием AI моделей.

```go
type ThreatAnalysisWorkflowInput struct {
    AnalysisType  string                 `json:"analysis_type"` // realtime, batch, deep, predictive
    DataSources   []string               `json:"data_sources"`  // logs, network, behavior, external
    TimeWindow    string                 `json:"time_window"`   // 1h, 24h, 7d, 30d
    ThreatModels  []string               `json:"threat_models"` // ddos, intrusion, malware, anomaly
    Configuration map[string]interface{} `json:"configuration"`
}
```

**Использование:**
```go
// Запуск workflow
workflowID, err := cadenceClient.ExecuteWorkflow(
    ctx,
    "ThreatAnalysisWorkflow",
    ThreatAnalysisWorkflowInput{
        AnalysisType: "realtime",
        DataSources:  []string{"network", "logs"},
        TimeWindow:   "1h",
        ThreatModels: []string{"ddos", "anomaly"},
    },
)
```

### BehaviorMonitoringWorkflow

Мониторинг поведения пользователей с поддержкой различных режимов.

```go
type BehaviorMonitoringWorkflowInput struct {
    MonitoringScope  string                 `json:"monitoring_scope"` // user, network, system, global
    UserIDs          []string               `json:"user_ids"`
    MonitoringPeriod string                 `json:"monitoring_period"` // continuous, scheduled, triggered
    AlertThresholds  map[string]float64     `json:"alert_thresholds"`
    Configuration    map[string]interface{} `json:"configuration"`
}
```

### AutomatedResponseWorkflow

Автоматическое реагирование на угрозы с различными политиками.

```go
type AutomatedResponseWorkflowInput struct {
    TriggerEvent    string                 `json:"trigger_event"`
    EventData       map[string]interface{} `json:"event_data"`
    ResponsePolicy  string                 `json:"response_policy"`  // immediate, escalated, manual_approval
    ResponseActions []string               `json:"response_actions"` // block, quarantine, alert, log
    Configuration   map[string]interface{} `json:"configuration"`
}
```

### ModelTrainingWorkflow

Обучение AI моделей с полным циклом валидации.

```go
type ModelTrainingWorkflowInput struct {
    ModelType       string                 `json:"model_type"`    // anomaly_detection, threat_classification, behavior_analysis
    TrainingData    string                 `json:"training_data"` // data source or path
    ModelVersion    string                 `json:"model_version"`
    TrainingMode    string                 `json:"training_mode"` // full, incremental, transfer
    HyperParameters map[string]interface{} `json:"hyperparameters"`
    Configuration   map[string]interface{} `json:"configuration"`
}
```

## 🧪 Разработка и тестирование

### Unit Tests

```bash
# Запуск всех тестов
go test ./internal/ai/... -v

# Тесты с покрытием
go test ./internal/ai/... -cover

# Бенчмарки
go test ./internal/ai/... -bench=.
```

### Integration Tests

```bash
# Тестирование AI workflows
./scripts/setup/test-ai-workflows.sh

# Load testing
go run cmd/load-test/main.go
```

### Code Quality

```bash
# Линтинг
golangci-lint run ./internal/ai/...

# Проверка безопасности
gosec ./internal/ai/...

# Форматирование
go fmt ./internal/ai/...
```

### Пример теста

```go
func TestAnomalyDetector(t *testing.T) {
    detector := anomaly.NewNetworkAnomalyDetector()
    
    metrics := &anomaly.NetworkMetrics{
        ConnectionsPerSecond: 150,
        BytesPerSecond:       2000000,
        PacketLossRate:       0.05,
        AverageLatency:       45.0,
        CPUUsage:             65.0,
        MemoryUsage:          70.0,
    }
    
    result, err := detector.DetectAnomaly(metrics)
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.False(t, result.IsAnomaly)
    assert.Greater(t, result.Confidence, 0.8)
}
```

## 🚀 Развертывание

### Локальное развертывание

```bash
# 1. Запуск Minikube
minikube start

# 2. Развертывание AI компонентов
kubectl apply -f k8s/ai/

# 3. Проверка статуса
kubectl get pods -n cloudbridge | grep ai
```

### Production развертывание

```bash
# 1. Сборка образов
docker build -t cloudbridge/ai-service:latest -f cmd/ai-service/Dockerfile .
docker build -t cloudbridge/ai-worker:latest -f cmd/ai-worker/Dockerfile .

# 2. Push в registry
docker push cloudbridge/ai-service:latest
docker push cloudbridge/ai-worker:latest

# 3. Развертывание в кластере
kubectl apply -f k8s/production/ai/
```

### Конфигурация

```yaml
# config/ai-config.yaml
ai:
  models:
    anomaly_detector:
      algorithm: "isolation_forest"
      threshold: 0.5
      retrain_interval: "24h"
    load_predictor:
      algorithm: "ensemble"
      horizon: 24
      confidence_level: 0.8
    route_optimizer:
      algorithm: "multi_armed_bandit"
      exploration_rate: 0.1
      learning_rate: 0.01
  
  cache:
    size: 10000
    ttl: "1h"
    cleanup_interval: "10m"
  
  circuit_breaker:
    failure_threshold: 5
    recovery_timeout: "30s"
    half_open_max_calls: 3
```

## 📊 Мониторинг

### Prometheus Metrics

```go
// Регистрация метрик
prometheus.MustRegister(aiRequestsTotal)
prometheus.MustRegister(aiRequestDuration)
prometheus.MustRegister(aiModelAccuracy)
```

### Grafana Dashboards

1. **AI Performance Dashboard**
   - Время отклика AI моделей
   - Точность предсказаний
   - Использование ресурсов

2. **Cadence Workflow Dashboard**
   - Статус выполнения workflows
   - Время выполнения activities
   - Ошибки и retry

3. **System Health Dashboard**
   - Состояние сервисов
   - Использование ресурсов
   - Health checks

### Логирование

```go
// Структурированное логирование
logger.Info("AI request processed",
    zap.String("model", "anomaly-detector"),
    zap.Float64("confidence", 0.95),
    zap.Duration("processing_time", 45*time.Millisecond),
    zap.Bool("is_anomaly", false),
)
```

### Алерты

```yaml
# prometheus/alerts.yml
groups:
- name: ai-alerts
  rules:
  - alert: AIHighLatency
    expr: ai_request_duration_seconds > 0.1
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "AI request latency is high"
      
  - alert: AIModelAccuracyLow
    expr: ai_model_accuracy < 0.8
    for: 10m
    labels:
      severity: critical
    annotations:
      summary: "AI model accuracy is below threshold"
```

## 🔧 Troubleshooting

### Частые проблемы

#### 1. AI Service не запускается
```bash
# Проверка логов
kubectl logs -n cloudbridge deployment/ai-service

# Проверка конфигурации
kubectl describe configmap -n cloudbridge ai-config
```

#### 2. Cadence connection failed
```bash
# Проверка Cadence сервиса
kubectl get pods -n cloudbridge | grep cadence

# Проверка network connectivity
kubectl exec -n cloudbridge ai-worker-xxx -- nslookup cadence-service
```

#### 3. High memory usage
```bash
# Проверка использования ресурсов
kubectl top pods -n cloudbridge

# Анализ heap
kubectl exec -n cloudbridge ai-service-xxx -- curl localhost:8084/debug/pprof/heap
```

### Отладка

```bash
# Включение debug логирования
export LOG_LEVEL=debug

# Профилирование CPU
go tool pprof http://localhost:8084/debug/pprof/profile

# Профилирование памяти
go tool pprof http://localhost:8084/debug/pprof/heap
```

## 📚 Дополнительные ресурсы

- [Полный отчет](./AI_INTEGRATION_REPORT.md)
- [API документация](./api/)
- [Примеры использования](./examples/)
- [Cadence документация](https://cadenceworkflow.io/docs/)

---

*Последнее обновление: 13 сентября 2025*

