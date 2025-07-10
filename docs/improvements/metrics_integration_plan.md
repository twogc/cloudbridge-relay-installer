# 📊 План интеграции метрик cloudflared в CloudBridge Relay

## 🎯 Цель

Интегрировать продвинутую систему метрик из Cloudflare cloudflared в CloudBridge Relay для улучшения мониторинга, диагностики и производительности.

## 📋 Текущее состояние

### Существующие метрики CloudBridge Relay
```go
// internal/metrics/metrics.go
type Metrics struct {
    // Базовые метрики
    ConnectionsTotal    prometheus.Counter
    ConnectionsActive   prometheus.Gauge
    RequestsTotal       prometheus.Counter
    RequestDuration     prometheus.Histogram
    ErrorsTotal         prometheus.Counter
}
```

### Метрики cloudflared для интеграции
```go
// temp/cloudflared/metrics/metrics.go
type Config struct {
    ReadyServer         *ReadyServer
    DiagnosticHandler   *diagnostic.Handler
    QuickTunnelHostname string
    Orchestrator        orchestrator
}
```

## 🔧 Компоненты для интеграции

### 1. **Health Check System**
```go
// Из temp/cloudflared/metrics/readiness.go
type ReadyServer struct {
    ready bool
    mu    sync.RWMutex
}

func (rs *ReadyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    rs.mu.RLock()
    ready := rs.ready
    rs.mu.RUnlock()
    
    if ready {
        w.WriteHeader(http.StatusOK)
        fmt.Fprintf(w, "OK\n")
    } else {
        w.WriteHeader(http.StatusServiceUnavailable)
        fmt.Fprintf(w, "NOT READY\n")
    }
}
```

### 2. **Diagnostic Endpoints**
```go
// Из temp/cloudflared/diagnostic/
type Handler struct {
    // Диагностические endpoints
    // - /debug/pprof/
    // - /debug/pprof/goroutine
    // - /debug/pprof/heap
    // - /debug/pprof/profile
}
```

### 3. **Build Info Metrics**
```go
// Из temp/cloudflared/metrics/metrics.go
func RegisterBuildInfo(buildType, buildTime, version string) {
    buildInfo := prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "build_info",
            Help: "Build and version information",
        },
        []string{"goversion", "type", "revision", "version"},
    )
    prometheus.MustRegister(buildInfo)
    buildInfo.WithLabelValues(runtime.Version(), buildType, buildTime, version).Set(1)
}
```

## 🚀 План реализации

### Этап 1: Расширение существующих метрик (1-2 дня)

#### 1.1 Обновить internal/metrics/metrics.go
```go
// Добавить новые метрики
type EnhancedMetrics struct {
    // Существующие метрики
    ConnectionsTotal    prometheus.Counter
    ConnectionsActive   prometheus.Gauge
    RequestsTotal       prometheus.Counter
    RequestDuration     prometheus.Histogram
    ErrorsTotal         prometheus.Counter
    
    // Новые метрики из cloudflared
    BuildInfo          *prometheus.GaugeVec
    ProtocolLatency    *prometheus.HistogramVec
    ProtocolErrors     *prometheus.CounterVec
    ProtocolSwitches   *prometheus.CounterVec
    ProtocolSuccess    *prometheus.CounterVec
    
    // Health check метрики
    HealthCheckStatus  prometheus.Gauge
    StartupTime        prometheus.Gauge
    
    // Connection pool метрики
    ConnectionPoolSize prometheus.Gauge
    ConnectionPoolIdle prometheus.Gauge
    ConnectionPoolActive prometheus.Gauge
}
```

#### 1.2 Добавить health check систему
```go
// internal/health/health_checker.go
type HealthChecker struct {
    ready bool
    mu    sync.RWMutex
    checks []HealthCheck
}

type HealthCheck interface {
    Check() error
    Name() string
}

func (hc *HealthChecker) AddCheck(check HealthCheck) {
    hc.checks = append(hc.checks, check)
}

func (hc *HealthChecker) IsReady() bool {
    hc.mu.RLock()
    defer hc.mu.RUnlock()
    return hc.ready
}
```

### Этап 2: Диагностические endpoints (2-3 дня)

#### 2.1 Создать diagnostic handler
```go
// internal/diagnostic/handler.go
type DiagnosticHandler struct {
    logger *zap.Logger
}

func (dh *DiagnosticHandler) InstallEndpoints(mux *http.ServeMux) {
    // Pprof endpoints
    mux.HandleFunc("/debug/pprof/", pprof.Index)
    mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
    mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
    mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
    mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
    
    // Custom diagnostic endpoints
    mux.HandleFunc("/debug/status", dh.handleStatus)
    mux.HandleFunc("/debug/config", dh.handleConfig)
    mux.HandleFunc("/debug/connections", dh.handleConnections)
}
```

#### 2.2 Добавить custom diagnostic endpoints
```go
func (dh *DiagnosticHandler) handleStatus(w http.ResponseWriter, r *http.Request) {
    status := map[string]interface{}{
        "version":     Version,
        "build_time":  BuildTime,
        "uptime":      time.Since(startTime).String(),
        "goroutines":  runtime.NumGoroutine(),
        "memory":      getMemoryStats(),
        "connections": getConnectionStats(),
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(status)
}
```

### Этап 3: Обновление HTTP сервера (1-2 дня)

#### 3.1 Расширить metrics server
```go
// internal/metrics/http.go
func NewMetricsServer(config *Config) *MetricsServer {
    mux := http.NewServeMux()
    
    // Prometheus metrics
    mux.Handle("/metrics", promhttp.Handler())
    
    // Health check
    mux.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "OK\n")
    })
    
    // Ready check
    mux.Handle("/ready", config.HealthChecker)
    
    // Diagnostic endpoints
    config.DiagnosticHandler.InstallEndpoints(mux)
    
    return &MetricsServer{
        mux: mux,
        config: config,
    }
}
```

### Этап 4: Конфигурация и документация (1 день)

#### 4.1 Обновить конфигурацию
```yaml
# config.yaml
relay:
  metrics:
    enabled: true
    port: 8081
    detailed: true
    health_check: true
    diagnostics: true
    build_info: true
    
  health:
    enabled: true
    checks:
      - database
      - redis
      - keycloak
      - tcp_proxy
    
  diagnostics:
    enabled: true
    pprof: true
    custom_endpoints: true
```

#### 4.2 Обновить Prometheus конфигурацию
```yaml
# prometheus/prometheus.yml
scrape_configs:
  - job_name: 'cloudbridge-relay'
    static_configs:
      - targets: ['localhost:8081']
    metrics_path: '/metrics'
    scrape_interval: 15s
    
  - job_name: 'cloudbridge-relay-health'
    static_configs:
      - targets: ['localhost:8081']
    metrics_path: '/healthcheck'
    scrape_interval: 30s
```

## 📊 Новые метрики

### 1. **Build Information**
```
# HELP build_info Build and version information
# TYPE build_info gauge
build_info{goversion="go1.21.0",type="release",revision="abc123",version="2.0.0"} 1
```

### 2. **Protocol Performance**
```
# HELP protocol_latency_seconds Protocol latency in seconds
# TYPE protocol_latency_seconds histogram
protocol_latency_seconds_bucket{protocol="quic",le="0.1"} 100
protocol_latency_seconds_bucket{protocol="http2",le="0.1"} 80

# HELP protocol_errors_total Total number of protocol errors
# TYPE protocol_errors_total counter
protocol_errors_total{protocol="quic"} 5
protocol_errors_total{protocol="http2"} 2
```

### 3. **Connection Pool**
```
# HELP connection_pool_size Current connection pool size
# TYPE connection_pool_size gauge
connection_pool_size{pool="tcp"} 50
connection_pool_size{pool="http"} 30

# HELP connection_pool_idle Idle connections in pool
# TYPE connection_pool_idle gauge
connection_pool_idle{pool="tcp"} 20
connection_pool_idle{pool="http"} 10
```

### 4. **Health Status**
```
# HELP health_check_status Health check status (1=healthy, 0=unhealthy)
# TYPE health_check_status gauge
health_check_status{check="database"} 1
health_check_status{check="redis"} 1
health_check_status{check="keycloak"} 0
```

## 🎯 Дашборды Grafana

### 1. **System Overview Dashboard**
- Build information
- Uptime and health status
- Memory and goroutine usage
- Connection statistics

### 2. **Protocol Performance Dashboard**
- Protocol latency comparison
- Error rates by protocol
- Protocol switching frequency
- Throughput metrics

### 3. **Connection Pool Dashboard**
- Pool size and utilization
- Idle vs active connections
- Connection creation/destruction rates
- Pool efficiency metrics

### 4. **Diagnostic Dashboard**
- Pprof data visualization
- Custom diagnostic endpoints
- System status overview
- Performance bottlenecks

## 🔧 Тестирование

### 1. **Unit Tests**
```go
func TestHealthChecker(t *testing.T) {
    hc := NewHealthChecker()
    
    // Add mock checks
    hc.AddCheck(&MockHealthCheck{name: "test", healthy: true})
    
    // Test ready status
    if !hc.IsReady() {
        t.Error("Health checker should be ready")
    }
}
```

### 2. **Integration Tests**
```go
func TestMetricsEndpoints(t *testing.T) {
    server := NewMetricsServer(config)
    
    // Test /metrics endpoint
    resp, err := http.Get("http://localhost:8081/metrics")
    assert.NoError(t, err)
    assert.Equal(t, 200, resp.StatusCode)
    
    // Test /healthcheck endpoint
    resp, err = http.Get("http://localhost:8081/healthcheck")
    assert.NoError(t, err)
    assert.Equal(t, 200, resp.StatusCode)
}
```

### 3. **Load Testing**
```bash
# Тестирование производительности метрик
wrk -t12 -c400 -d30s http://localhost:8081/metrics

# Тестирование health check
ab -n 1000 -c 10 http://localhost:8081/healthcheck
```

## 📈 Ожидаемые результаты

### Количественные улучшения
- **Детализация метрик**: +200% новых метрик
- **Диагностика**: +300% улучшение отладки
- **Мониторинг**: +150% покрытие системы
- **Производительность**: -10% overhead

### Качественные улучшения
- Упрощение диагностики проблем
- Лучший мониторинг состояния системы
- Расширенные возможности отладки
- Улучшенная observability

## 🚀 Следующие шаги

### Немедленные действия
1. **Создать ветку для разработки**
   ```bash
   git checkout -b feature/cloudflared-metrics-integration
   ```

2. **Начать с health check системы**
   - Реализовать HealthChecker
   - Добавить базовые checks
   - Интегрировать в HTTP server

3. **Добавить новые метрики**
   - Расширить существующие метрики
   - Добавить build info
   - Реализовать protocol metrics

### Долгосрочные планы
1. **Полная интеграция диагностики**
2. **Создание дашбордов Grafana**
3. **Оптимизация производительности**
4. **Расширение health checks**

## 🔗 Ссылки

- [Cloudflare cloudflared metrics](https://github.com/cloudflare/cloudflared/tree/master/metrics)
- [Prometheus client_golang](https://github.com/prometheus/client_golang)
- [Grafana dashboards](https://grafana.com/docs/grafana/latest/dashboards/)
- [Go pprof](https://golang.org/pkg/net/http/pprof/)

---

**Статус**: План готов ✅  
**Следующий шаг**: Начало реализации health check системы  
**Время реализации**: 5-7 дней  
**Приоритет**: Высокий 