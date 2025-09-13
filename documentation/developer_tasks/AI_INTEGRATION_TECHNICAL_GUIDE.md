# 🔧 AI Integration - Technical Implementation Guide

## 📋 Техническое руководство по реализации

Этот документ содержит детальные технические инструкции для выполнения задач AI Integration Phase 2.

---

## 🏗️ Архитектура интеграции

### **Текущая архитектура**:
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Relay Server  │────│   AI Service    │────│   AI Worker     │
│                 │    │                 │    │                 │
│ • Connection    │    │ • Anomaly       │    │ • Cadence       │
│   Processing    │    │   Detection     │    │   Workflows     │
│ • Traffic       │    │ • Load          │    │ • Activities    │
│   Routing       │    │   Prediction    │    │ • Long-running  │
│ • Load          │    │ • Route         │    │   Tasks         │
│   Balancing     │    │   Optimization  │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │    Cadence      │
                    │                 │
                    │ • Workflow      │
                    │   Orchestration │
                    │ • State         │
                    │   Management    │
                    │ • Fault         │
                    │   Tolerance     │
                    └─────────────────┘
```

### **Целевая архитектура**:
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Relay Server  │────│   AI Middleware │────│   AI Service    │
│                 │    │                 │    │                 │
│ • Connection    │    │ • Real-time     │    │ • Anomaly       │
│   Processing    │    │   Analysis      │    │   Detection     │
│ • AI-Enhanced   │    │ • Caching       │    │ • Load          │
│   Routing       │    │ • Metrics       │    │   Prediction    │
│ • Intelligent   │    │ • Circuit       │    │ • Route         │
│   Load Bal.     │    │   Breaker       │    │   Optimization  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │   Monitoring    │
                    │                 │
                    │ • Prometheus    │
                    │ • Grafana       │
                    │ • Alerts        │
                    │ • Health Checks │
                    └─────────────────┘
```

---

## 🎯 Задача 1: Интеграция AI с Relay Server

### **1.1 Изучение структуры Relay Server**

```bash
# Изучите структуру проекта
tree cmd/relay/
tree internal/api/
```

**Ключевые файлы для изучения**:
- `cmd/relay/main.go` - Точка входа
- `internal/api/` - API обработчики
- `internal/keycloak/` - Keycloak интеграция
- `internal/api/protocol_handlers.go` - Обработчики протоколов

### **1.2 Создание AI Middleware**

Создайте файл `internal/api/ai_middleware.go`:

```go
package api

import (
    "context"
    "fmt"
    "time"
    
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
    "github.com/twogc/cloudbridge-relay-installer/internal/ai"
)

// AIMiddleware обеспечивает интеграцию AI с Relay Server
type AIMiddleware struct {
    aiClient   *ai.Client
    config     *ai.Config
    logger     *zap.Logger
    cache      *AICache
    metrics    *AIMetrics
}

// AICache кэширует результаты AI анализа
type AICache struct {
    anomalyCache    map[string]*ai.AnomalyResult
    predictionCache map[string]*ai.LoadPrediction
    routeCache      map[string]*ai.OptimizedRoute
    ttl             time.Duration
}

// AIMetrics собирает метрики AI обработки
type AIMetrics struct {
    requestsProcessed    int64
    anomaliesDetected    int64
    predictionsMade      int64
    routesOptimized      int64
    processingTime       time.Duration
    cacheHits           int64
    cacheMisses         int64
}

// NewAIMiddleware создает новый AI middleware
func NewAIMiddleware(aiClient *ai.Client, config *ai.Config, logger *zap.Logger) *AIMiddleware {
    return &AIMiddleware{
        aiClient: aiClient,
        config:   config,
        logger:   logger,
        cache: &AICache{
            anomalyCache:    make(map[string]*ai.AnomalyResult),
            predictionCache: make(map[string]*ai.LoadPrediction),
            routeCache:      make(map[string]*ai.OptimizedRoute),
            ttl:             5 * time.Minute,
        },
        metrics: &AIMetrics{},
    }
}

// ProcessConnection обрабатывает соединение через AI
func (m *AIMiddleware) ProcessConnection(ctx context.Context, conn *Connection) error {
    start := time.Now()
    defer func() {
        m.metrics.processingTime += time.Since(start)
        m.metrics.requestsProcessed++
    }()

    // 1. Анализ аномалий
    anomalyResult, err := m.analyzeAnomalies(ctx, conn)
    if err != nil {
        m.logger.Error("Failed to analyze anomalies", zap.Error(err))
        return err
    }

    // 2. Предсказание нагрузки
    loadPrediction, err := m.predictLoad(ctx, conn)
    if err != nil {
        m.logger.Error("Failed to predict load", zap.Error(err))
        return err
    }

    // 3. Оптимизация маршрута
    optimizedRoute, err := m.optimizeRoute(ctx, conn)
    if err != nil {
        m.logger.Error("Failed to optimize route", zap.Error(err))
        return err
    }

    // 4. Применение результатов
    return m.applyAIResults(ctx, conn, anomalyResult, loadPrediction, optimizedRoute)
}

// analyzeAnomalies анализирует аномалии в соединении
func (m *AIMiddleware) analyzeAnomalies(ctx context.Context, conn *Connection) (*ai.AnomalyResult, error) {
    // Проверяем кэш
    cacheKey := fmt.Sprintf("anomaly:%s:%d", conn.RemoteAddr, time.Now().Unix()/300)
    if result, exists := m.cache.anomalyCache[cacheKey]; exists {
        m.metrics.cacheHits++
        return result, nil
    }
    m.metrics.cacheMisses++

    // Собираем метрики соединения
    metrics := map[string]interface{}{
        "packets_per_second":  conn.PacketsPerSecond,
        "bytes_per_second":    conn.BytesPerSecond,
        "connection_count":    conn.ConnectionCount,
        "latency":            conn.Latency,
        "bandwidth":          conn.Bandwidth,
    }

    // Вызываем AI Service
    result, err := m.aiClient.DetectAnomaly(ctx, metrics)
    if err != nil {
        return nil, err
    }

    // Кэшируем результат
    m.cache.anomalyCache[cacheKey] = result
    
    // Обновляем метрики
    if result.IsAnomaly {
        m.metrics.anomaliesDetected++
    }

    return result, nil
}

// predictLoad предсказывает нагрузку
func (m *AIMiddleware) predictLoad(ctx context.Context, conn *Connection) (*ai.LoadPrediction, error) {
    // Проверяем кэш
    cacheKey := fmt.Sprintf("load:%d", time.Now().Unix()/300)
    if result, exists := m.cache.predictionCache[cacheKey]; exists {
        m.metrics.cacheHits++
        return result, nil
    }
    m.metrics.cacheMisses++

    // Вызываем AI Service
    result, err := m.aiClient.PredictLoad(ctx, 24) // 24 часа
    if err != nil {
        return nil, err
    }

    // Кэшируем результат
    m.cache.predictionCache[cacheKey] = result
    m.metrics.predictionsMade++

    return result, nil
}

// optimizeRoute оптимизирует маршрут
func (m *AIMiddleware) optimizeRoute(ctx context.Context, conn *Connection) (*ai.OptimizedRoute, error) {
    // Проверяем кэш
    cacheKey := fmt.Sprintf("route:%s:%s", conn.Source, conn.Destination)
    if result, exists := m.cache.routeCache[cacheKey]; exists {
        m.metrics.cacheHits++
        return result, nil
    }
    m.metrics.cacheMisses++

    // Вызываем AI Service
    constraints := map[string]interface{}{
        "max_latency":   100, // ms
        "min_bandwidth": 1000, // Mbps
    }
    
    result, err := m.aiClient.OptimizeRoute(ctx, conn.Source, conn.Destination, constraints)
    if err != nil {
        return nil, err
    }

    // Кэшируем результат
    m.cache.routeCache[cacheKey] = result
    m.metrics.routesOptimized++

    return result, nil
}

// applyAIResults применяет результаты AI анализа
func (m *AIMiddleware) applyAIResults(ctx context.Context, conn *Connection, 
    anomalyResult *ai.AnomalyResult, loadPrediction *ai.LoadPrediction, 
    optimizedRoute *ai.OptimizedRoute) error {

    // 1. Обработка аномалий
    if anomalyResult.IsAnomaly {
        m.logger.Warn("Anomaly detected in connection",
            zap.String("remote_addr", conn.RemoteAddr),
            zap.Float64("score", anomalyResult.Score),
            zap.String("type", anomalyResult.AnomalyType),
        )
        
        // Применяем политику реагирования
        switch anomalyResult.Severity {
        case "high", "critical":
            // Блокируем соединение
            return m.blockConnection(ctx, conn, "anomaly_detected")
        case "medium":
            // Ограничиваем пропускную способность
            return m.throttleConnection(ctx, conn, 0.5)
        case "low":
            // Логируем и продолжаем
            m.logger.Info("Low severity anomaly, continuing")
        }
    }

    // 2. Применение предсказания нагрузки
    if loadPrediction != nil {
        // Адаптируем ресурсы на основе предсказания
        m.adjustResources(ctx, loadPrediction)
    }

    // 3. Применение оптимизированного маршрута
    if optimizedRoute != nil {
        // Обновляем маршрут соединения
        conn.Route = optimizedRoute.Path
        conn.Protocol = optimizedRoute.Protocol
    }

    return nil
}

// blockConnection блокирует соединение
func (m *AIMiddleware) blockConnection(ctx context.Context, conn *Connection, reason string) error {
    m.logger.Warn("Blocking connection",
        zap.String("remote_addr", conn.RemoteAddr),
        zap.String("reason", reason),
    )
    
    // Закрываем соединение
    return conn.Close()
}

// throttleConnection ограничивает пропускную способность
func (m *AIMiddleware) throttleConnection(ctx context.Context, conn *Connection, factor float64) error {
    m.logger.Info("Throttling connection",
        zap.String("remote_addr", conn.RemoteAddr),
        zap.Float64("factor", factor),
    )
    
    // Применяем ограничение
    conn.ThrottleFactor = factor
    return nil
}

// adjustResources адаптирует ресурсы на основе предсказания
func (m *AIMiddleware) adjustResources(ctx context.Context, prediction *ai.LoadPrediction) {
    // Анализируем предсказание и адаптируем ресурсы
    if len(prediction.Forecast) > 0 {
        avgLoad := 0.0
        for _, load := range prediction.Forecast {
            avgLoad += load
        }
        avgLoad /= float64(len(prediction.Forecast))
        
        if avgLoad > 80.0 {
            // Высокая нагрузка - увеличиваем ресурсы
            m.logger.Info("High load predicted, scaling up resources",
                zap.Float64("predicted_load", avgLoad),
            )
        } else if avgLoad < 20.0 {
            // Низкая нагрузка - уменьшаем ресурсы
            m.logger.Info("Low load predicted, scaling down resources",
                zap.Float64("predicted_load", avgLoad),
            )
        }
    }
}

// GetMetrics возвращает метрики AI middleware
func (m *AIMiddleware) GetMetrics() *AIMetrics {
    return m.metrics
}

// CleanupCache очищает устаревшие записи кэша
func (m *AIMiddleware) CleanupCache() {
    now := time.Now()
    
    // Очищаем кэш аномалий
    for key := range m.cache.anomalyCache {
        // Простая логика очистки по времени
        if now.Unix()%300 == 0 { // Каждые 5 минут
            delete(m.cache.anomalyCache, key)
        }
    }
    
    // Аналогично для других кэшей
}
```

### **1.3 Интеграция в Relay Server**

Обновите `cmd/relay/main.go`:

```go
// Добавьте в структуру Server
type Server struct {
    // ... существующие поля
    aiMiddleware *AIMiddleware
}

// В функции NewServer добавьте инициализацию AI
func NewServer(config *Config) (*Server, error) {
    // ... существующий код
    
    // Инициализация AI клиента
    aiClient, err := ai.NewClient(config.AI)
    if err != nil {
        return nil, fmt.Errorf("failed to create AI client: %w", err)
    }
    
    // Создание AI middleware
    aiMiddleware := NewAIMiddleware(aiClient, config.AI, logger)
    
    server := &Server{
        // ... существующие поля
        aiMiddleware: aiMiddleware,
    }
    
    return server, nil
}

// Обновите обработку соединений
func (s *Server) handleConnection(conn net.Conn) {
    // ... существующий код
    
    // Обработка через AI middleware
    if err := s.aiMiddleware.ProcessConnection(ctx, connection); err != nil {
        s.logger.Error("AI processing failed", zap.Error(err))
        // Решаем, продолжать ли обработку
    }
    
    // ... остальная обработка
}
```

---

## 📊 Задача 2: Мониторинг и дашборд

### **2.1 Prometheus метрики**

Создайте файл `internal/ai/metrics_extended.go`:

```go
package ai

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // AI Connection Metrics
    AIConnectionsProcessed = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "ai_connections_processed_total",
            Help: "Total number of connections processed by AI",
        },
        []string{"status", "anomaly_detected", "service"},
    )
    
    AIProcessingDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "ai_processing_duration_seconds",
            Help:    "Time spent processing connections through AI",
            Buckets: prometheus.DefBuckets,
        },
        []string{"service", "operation"},
    )
    
    // Anomaly Detection Metrics
    AnomaliesDetected = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "ai_anomalies_detected_total",
            Help: "Total number of anomalies detected",
        },
        []string{"severity", "type", "source"},
    )
    
    AnomalyDetectionAccuracy = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "ai_anomaly_detection_accuracy",
            Help: "Accuracy of anomaly detection model",
        },
        []string{"model_version"},
    )
    
    // Load Prediction Metrics
    LoadPredictionsMade = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "ai_load_predictions_total",
            Help: "Total number of load predictions made",
        },
        []string{"horizon", "accuracy_range"},
    )
    
    LoadPredictionAccuracy = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "ai_load_prediction_accuracy",
            Help: "Accuracy of load prediction model",
        },
        []string{"model_version", "horizon"},
    )
    
    // Route Optimization Metrics
    RoutesOptimized = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "ai_routes_optimized_total",
            Help: "Total number of routes optimized",
        },
        []string{"protocol", "optimization_type"},
    )
    
    RouteOptimizationLatency = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "ai_route_optimization_latency_seconds",
            Help:    "Time spent optimizing routes",
            Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0},
        },
        []string{"protocol"},
    )
    
    // Cache Metrics
    AICacheHits = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "ai_cache_hits_total",
            Help: "Total number of AI cache hits",
        },
        []string{"cache_type"},
    )
    
    AICacheMisses = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "ai_cache_misses_total",
            Help: "Total number of AI cache misses",
        },
        []string{"cache_type"},
    )
    
    // System Metrics
    AIServiceHealth = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "ai_service_health",
            Help: "Health status of AI services (1=healthy, 0=unhealthy)",
        },
        []string{"service", "component"},
    )
    
    AIModelVersion = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "ai_model_version",
            Help: "Version of AI models in use",
        },
        []string{"model_type", "version"},
    )
)
```

### **2.2 Grafana дашборд**

Создайте файл `monitoring/grafana/dashboards/ai-dashboard.json`:

```json
{
  "dashboard": {
    "id": null,
    "title": "CloudBridge AI System Dashboard",
    "tags": ["cloudbridge", "ai", "monitoring"],
    "timezone": "browser",
    "panels": [
      {
        "id": 1,
        "title": "AI Connections Processed",
        "type": "stat",
        "targets": [
          {
            "expr": "rate(ai_connections_processed_total[5m])",
            "legendFormat": "Connections/sec"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "color": {
              "mode": "thresholds"
            },
            "thresholds": {
              "steps": [
                {"color": "green", "value": null},
                {"color": "yellow", "value": 100},
                {"color": "red", "value": 500}
              ]
            }
          }
        }
      },
      {
        "id": 2,
        "title": "Anomaly Detection Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(ai_anomalies_detected_total[5m])",
            "legendFormat": "{{severity}} - {{type}}"
          }
        ]
      },
      {
        "id": 3,
        "title": "AI Processing Latency",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(ai_processing_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          },
          {
            "expr": "histogram_quantile(0.50, rate(ai_processing_duration_seconds_bucket[5m]))",
            "legendFormat": "50th percentile"
          }
        ]
      },
      {
        "id": 4,
        "title": "Cache Hit Rate",
        "type": "stat",
        "targets": [
          {
            "expr": "rate(ai_cache_hits_total[5m]) / (rate(ai_cache_hits_total[5m]) + rate(ai_cache_misses_total[5m])) * 100",
            "legendFormat": "Hit Rate %"
          }
        ]
      },
      {
        "id": 5,
        "title": "AI Service Health",
        "type": "stat",
        "targets": [
          {
            "expr": "ai_service_health",
            "legendFormat": "{{service}} - {{component}}"
          }
        ]
      }
    ],
    "time": {
      "from": "now-1h",
      "to": "now"
    },
    "refresh": "30s"
  }
}
```

### **2.3 Алерты**

Создайте файл `monitoring/prometheus/alerts/ai-alerts.yml`:

```yaml
groups:
- name: ai-alerts
  rules:
  - alert: HighAnomalyRate
    expr: rate(ai_anomalies_detected_total[5m]) > 0.1
    for: 2m
    labels:
      severity: warning
    annotations:
      summary: "High anomaly detection rate"
      description: "Anomaly detection rate is {{ $value }} anomalies per second"

  - alert: AIServiceDown
    expr: ai_service_health == 0
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "AI Service is down"
      description: "AI Service {{ $labels.service }} component {{ $labels.component }} is unhealthy"

  - alert: HighAIProcessingLatency
    expr: histogram_quantile(0.95, rate(ai_processing_duration_seconds_bucket[5m])) > 0.1
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High AI processing latency"
      description: "95th percentile AI processing latency is {{ $value }}s"

  - alert: LowCacheHitRate
    expr: rate(ai_cache_hits_total[5m]) / (rate(ai_cache_hits_total[5m]) + rate(ai_cache_misses_total[5m])) < 0.8
    for: 10m
    labels:
      severity: warning
    annotations:
      summary: "Low AI cache hit rate"
      description: "AI cache hit rate is {{ $value }}%"
```

---

## ⚡ Задача 3: Оптимизация производительности

### **3.1 Кэширование**

Создайте файл `internal/ai/cache/redis_cache.go`:

```go
package cache

import (
    "context"
    "encoding/json"
    "time"
    
    "github.com/go-redis/redis/v8"
    "go.uber.org/zap"
)

// RedisCache реализует кэширование с использованием Redis
type RedisCache struct {
    client *redis.Client
    logger *zap.Logger
    ttl    time.Duration
}

// NewRedisCache создает новый Redis кэш
func NewRedisCache(addr string, password string, db int, logger *zap.Logger) *RedisCache {
    rdb := redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: password,
        DB:       db,
    })
    
    return &RedisCache{
        client: rdb,
        logger: logger,
        ttl:    5 * time.Minute,
    }
}

// Set сохраняет значение в кэш
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    
    return c.client.Set(ctx, key, data, c.ttl).Err()
}

// Get получает значение из кэша
func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
    val, err := c.client.Get(ctx, key).Result()
    if err != nil {
        return err
    }
    
    return json.Unmarshal([]byte(val), dest)
}

// Delete удаляет значение из кэша
func (c *RedisCache) Delete(ctx context.Context, key string) error {
    return c.client.Del(ctx, key).Err()
}

// Exists проверяет существование ключа
func (c *RedisCache) Exists(ctx context.Context, key string) bool {
    result, err := c.client.Exists(ctx, key).Result()
    return err == nil && result > 0
}
```

### **3.2 Профилирование**

Добавьте в `cmd/ai-service/main.go`:

```go
import (
    _ "net/http/pprof"
    "net/http"
)

func main() {
    // ... существующий код
    
    // Запуск pprof сервера
    go func() {
        log.Println("Starting pprof server on :6060")
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
    
    // ... остальной код
}
```

---

## 🔒 Задача 4: Подготовка к продакшену

### **4.1 Структурированное логирование**

Создайте файл `internal/ai/logging/structured_logger.go`:

```go
package logging

import (
    "context"
    "time"
    
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

// StructuredLogger обеспечивает структурированное логирование
type StructuredLogger struct {
    logger *zap.Logger
}

// NewStructuredLogger создает новый структурированный логгер
func NewStructuredLogger(level string) (*StructuredLogger, error) {
    config := zap.NewProductionConfig()
    
    switch level {
    case "debug":
        config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
    case "info":
        config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
    case "warn":
        config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
    case "error":
        config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
    }
    
    logger, err := config.Build()
    if err != nil {
        return nil, err
    }
    
    return &StructuredLogger{logger: logger}, nil
}

// LogAIProcessing логирует обработку AI
func (l *StructuredLogger) LogAIProcessing(ctx context.Context, 
    requestID string, operation string, duration time.Duration, 
    result interface{}, err error) {
    
    fields := []zap.Field{
        zap.String("request_id", requestID),
        zap.String("operation", operation),
        zap.Duration("duration", duration),
        zap.Time("timestamp", time.Now()),
    }
    
    if result != nil {
        fields = append(fields, zap.Any("result", result))
    }
    
    if err != nil {
        fields = append(fields, zap.Error(err))
        l.logger.Error("AI processing failed", fields...)
    } else {
        l.logger.Info("AI processing completed", fields...)
    }
}

// LogAnomalyDetected логирует обнаруженную аномалию
func (l *StructuredLogger) LogAnomalyDetected(ctx context.Context,
    requestID string, score float64, anomalyType string, 
    severity string, details map[string]interface{}) {
    
    fields := []zap.Field{
        zap.String("request_id", requestID),
        zap.Float64("anomaly_score", score),
        zap.String("anomaly_type", anomalyType),
        zap.String("severity", severity),
        zap.Time("timestamp", time.Now()),
    }
    
    for key, value := range details {
        fields = append(fields, zap.Any(key, value))
    }
    
    l.logger.Warn("Anomaly detected", fields...)
}
```

### **4.2 Health Checks**

Создайте файл `internal/ai/health/health_checker.go`:

```go
package health

import (
    "context"
    "time"
    
    "go.uber.org/zap"
)

// HealthChecker проверяет здоровье AI компонентов
type HealthChecker struct {
    logger *zap.Logger
    checks map[string]HealthCheck
}

// HealthCheck интерфейс для проверки здоровья
type HealthCheck interface {
    Check(ctx context.Context) error
    Name() string
}

// HealthStatus статус здоровья
type HealthStatus struct {
    Status    string            `json:"status"`
    Timestamp time.Time         `json:"timestamp"`
    Checks    map[string]string `json:"checks"`
}

// NewHealthChecker создает новый health checker
func NewHealthChecker(logger *zap.Logger) *HealthChecker {
    return &HealthChecker{
        logger: logger,
        checks: make(map[string]HealthCheck),
    }
}

// RegisterCheck регистрирует проверку здоровья
func (h *HealthChecker) RegisterCheck(check HealthCheck) {
    h.checks[check.Name()] = check
}

// CheckHealth проверяет здоровье всех компонентов
func (h *HealthChecker) CheckHealth(ctx context.Context) *HealthStatus {
    status := &HealthStatus{
        Status:    "healthy",
        Timestamp: time.Now(),
        Checks:    make(map[string]string),
    }
    
    for name, check := range h.checks {
        if err := check.Check(ctx); err != nil {
            status.Checks[name] = "unhealthy: " + err.Error()
            status.Status = "unhealthy"
        } else {
            status.Checks[name] = "healthy"
        }
    }
    
    return status
}
```

---

## 🧪 Тестирование

### **Unit тесты**

Создайте файл `internal/ai/middleware_test.go`:

```go
package ai

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// MockAIClient мок для AI клиента
type MockAIClient struct {
    mock.Mock
}

func (m *MockAIClient) DetectAnomaly(ctx context.Context, metrics map[string]interface{}) (*AnomalyResult, error) {
    args := m.Called(ctx, metrics)
    return args.Get(0).(*AnomalyResult), args.Error(1)
}

func (m *MockAIClient) PredictLoad(ctx context.Context, horizon int) (*LoadPrediction, error) {
    args := m.Called(ctx, horizon)
    return args.Get(0).(*LoadPrediction), args.Error(1)
}

func (m *MockAIClient) OptimizeRoute(ctx context.Context, source, destination string, constraints map[string]interface{}) (*OptimizedRoute, error) {
    args := m.Called(ctx, source, destination, constraints)
    return args.Get(0).(*OptimizedRoute), args.Error(1)
}

func TestAIMiddleware_ProcessConnection(t *testing.T) {
    // Создаем мок клиент
    mockClient := new(MockAIClient)
    
    // Настраиваем ожидания
    mockClient.On("DetectAnomaly", mock.Anything, mock.Anything).Return(
        &AnomalyResult{
            Score:      0.3,
            IsAnomaly:  false,
            Confidence: 0.85,
        }, nil)
    
    mockClient.On("PredictLoad", mock.Anything, 24).Return(
        &LoadPrediction{
            Forecast:   []float64{45, 50, 55},
            Confidence: 0.8,
            Horizon:    24,
        }, nil)
    
    mockClient.On("OptimizeRoute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
        &OptimizedRoute{
            Protocol: "QUIC",
            Path:     []string{"source", "dest"},
        }, nil)
    
    // Создаем middleware
    middleware := &AIMiddleware{
        aiClient: mockClient,
        cache:    &AICache{},
        metrics:  &AIMetrics{},
    }
    
    // Создаем тестовое соединение
    conn := &Connection{
        RemoteAddr: "192.168.1.1:8080",
        Source:     "192.168.1.1",
        Destination: "192.168.1.100",
    }
    
    // Тестируем обработку
    err := middleware.ProcessConnection(context.Background(), conn)
    
    // Проверяем результаты
    assert.NoError(t, err)
    mockClient.AssertExpectations(t)
}
```

---

## 📝 Заключение

Этот технический гид содержит все необходимые детали для реализации AI Integration Phase 2. Следуйте инструкциям по порядку, тестируйте каждый компонент и документируйте результаты.

**Удачи в разработке! 🚀**
