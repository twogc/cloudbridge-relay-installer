# 📡 НЕДЕЛЯ 1, ЗАДАЧА 1.2: API СТРУКТУРА И РАСШИРЕНИЕ
**Duration**: 6 часов
**Status**: IN PROGRESS
**Focus**: Dashboard BFF API endpoints

---

## ЦЕЛЬ
Проверить и расширить Dashboard BFF API структуру. Добавить новые endpoints для получения данных от всех CloudBridge сервисов.

---

## ТЕКУЩАЯ СТРУКТУРА API (в main.go)

### Уже реализовано (в bff/main.go):
```go
// Public routes (без auth)
GET /healthz                    // Health check
GET /readyz                     // Readiness check
GET /metrics                    // Prometheus metrics

// Authenticated routes (требуют Bearer token)
GET /api/v1/dashboard/summary       // Сводка всех метрик
GET /api/v1/dashboard/health        // Здоровье всех сервисов
GET /api/v1/dashboard/metrics       // Обзорные метрики
GET /api/v1/dashboard/analytics     // Аналитика трафика (с временным диапазоном)

// Service status endpoints
GET /api/v1/relay/status            // Статус Relay
GET /api/v1/p2p/status              // Статус P2P
GET /api/v1/ai/insights             // AI insights
GET /api/v1/edge-pops/status        // Статус Edge PoP

// Real-time
GET /api/v1/ws                      // WebSocket для real-time
```

---

## НЕДОСТАЮЩИЕ ENDPOINTS (для реализации)

### 1. Alert Management
```go
GET    /api/v1/alerts                    // Список всех алертов
POST   /api/v1/alerts                    // Создать алерт (admin)
GET    /api/v1/alerts/{id}               // Детали алерта
POST   /api/v1/alerts/{id}/acknowledge   // Подтвердить алерт
DELETE /api/v1/alerts/{id}               // Удалить алерт
GET    /api/v1/alerts/search?query=...   // Поиск по алертам
```

### 2. DNS Analytics
```go
GET    /api/v1/dns/queries               // DNS запросы
GET    /api/v1/dns/zones                 // Список зон
GET    /api/v1/dns/performance           // DNS performance metrics
GET    /api/v1/dns/threats               // Обнаруженные угрозы
```

### 3. DDoS Monitoring
```go
GET    /api/v1/ddos/active-attacks       // Активные атаки
GET    /api/v1/ddos/statistics           // Статистика DDoS
GET    /api/v1/ddos/mitigations          // Примененные смягчения
GET    /api/v1/ddos/threats              // Известные угрозы
```

### 4. Performance Analytics
```go
GET    /api/v1/perf/top-paths            // Top paths by traffic
GET    /api/v1/perf/latency-distribution // Распределение латенси
GET    /api/v1/perf/throughput           // Пропускная способность
GET    /api/v1/perf/connections          // Активные соединения
```

### 5. Configuration Management
```go
GET    /api/v1/config/settings           // Текущие настройки
PUT    /api/v1/config/settings           // Обновить настройки
GET    /api/v1/config/policies           // Политики доступа
```

### 6. Admin Operations
```go
GET    /api/v1/admin/services            // Статус всех сервисов
POST   /api/v1/admin/services/{name}/restart   // Перезагрузить сервис
GET    /api/v1/admin/logs                // Системные логи
GET    /api/v1/admin/audit               // Audit trail
```

---

## РЕАЛИЗАЦИЯ ПО ЭТАПАМ

### ЭТАП 1: Alert Management (Hour 1)

#### 1.1 Создать типы для алертов
**File**: `backend/shared/types/alerts.go`

```go
package types

import "time"

// Alert represents an alert in the system
type Alert struct {
    ID        string    `json:"id"`
    Severity  string    `json:"severity"`  // critical, warning, info
    Message   string    `json:"message"`
    Service   string    `json:"service"`   // which service triggered
    Timestamp time.Time `json:"timestamp"`
    Resolved  bool      `json:"resolved"`
    ResolvedAt *time.Time `json:"resolved_at,omitempty"`
}

type AlertList struct {
    Alerts []Alert `json:"alerts"`
    Total  int     `json:"total"`
}

type AlertRequest struct {
    Severity string `json:"severity"`
    Message  string `json:"message"`
    Service  string `json:"service"`
}
```

#### 1.2 Создать handlers для алертов
**File**: `backend/bff/handlers/alerts.go`

```go
package handlers

import (
    "net/http"
    "github.com/go-chi/chi/v5"
    "github.com/twogc/cloudbridge-dashboard/bff/aggregators"
)

type AlertHandler struct {
    aggregator *aggregators.DashboardAggregator
}

func NewAlertHandler(agg *aggregators.DashboardAggregator) *AlertHandler {
    return &AlertHandler{aggregator: agg}
}

// GET /api/v1/alerts
func (ah *AlertHandler) GetAlerts(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // Get alerts from aggregator (will query Alertmanager)
    alerts, err := ah.aggregator.GetAlerts(ctx)
    if err != nil {
        http.Error(w, "Failed to get alerts", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(alerts)
}

// GET /api/v1/alerts/{id}
func (ah *AlertHandler) GetAlertDetails(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    ctx := r.Context()

    alert, err := ah.aggregator.GetAlertByID(ctx, id)
    if err != nil {
        http.Error(w, "Alert not found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(alert)
}

// POST /api/v1/alerts/{id}/acknowledge
func (ah *AlertHandler) AckAlert(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    ctx := r.Context()

    err := ah.aggregator.AcknowledgeAlert(ctx, id)
    if err != nil {
        http.Error(w, "Failed to acknowledge alert", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
}
```

#### 1.3 Добавить методы в Aggregator
**File**: `backend/bff/aggregators/dashboard.go` (добавить к существующему)

```go
// GetAlerts retrieves recent alerts from Prometheus AlertManager
func (d *DashboardAggregator) GetAlerts(ctx context.Context) (*types.AlertList, error) {
    // TODO: Query Prometheus AlertManager for alerts
    // For now, return mock data (same as in getRecentAlerts)
    return &types.AlertList{
        Alerts: []types.Alert{
            {
                ID:        "alert-1",
                Severity:  "warning",
                Message:   "High latency detected on frankfurt-amsterdam link",
                Service:   "relay",
                Timestamp: time.Now().Add(-5 * time.Minute),
                Resolved:  false,
            },
        },
        Total: 1,
    }, nil
}

// GetAlertByID retrieves specific alert
func (d *DashboardAggregator) GetAlertByID(ctx context.Context, id string) (*types.Alert, error) {
    // TODO: Query for specific alert by ID
    return nil, fmt.Errorf("alert not found")
}

// AcknowledgeAlert marks alert as acknowledged
func (d *DashboardAggregator) AcknowledgeAlert(ctx context.Context, id string) error {
    // TODO: Send acknowledgement to AlertManager
    return nil
}
```

### ЭТАП 2: DNS Analytics (Hour 2)

#### 2.1 DNS типы
**File**: `backend/shared/types/dns.go`

```go
package types

type DNSQuery struct {
    ID        string `json:"id"`
    Query     string `json:"query"`
    Type      string `json:"type"`       // A, AAAA, CNAME, MX, etc
    Response  string `json:"response"`
    Latency   int    `json:"latency_ms"`
    Timestamp int64  `json:"timestamp"`
    Status    string `json:"status"`     // success, failed, blocked
}

type DNSMetrics struct {
    QueriesPerSecond float64 `json:"queries_per_second"`
    SuccessRate      float64 `json:"success_rate"`
    AvgLatency       float64 `json:"avg_latency_ms"`
    BlockedQueries   int64   `json:"blocked_queries"`
}

type DNSZone struct {
    Name       string `json:"name"`
    RecordCount int   `json:"record_count"`
    LastUpdate int64  `json:"last_update"`
}
```

#### 2.2 DNS Handler
**File**: `backend/bff/handlers/dns.go`

```go
package handlers

import "net/http"

type DNSHandler struct {
    aggregator *aggregators.DashboardAggregator
}

func NewDNSHandler(agg *aggregators.DashboardAggregator) *DNSHandler {
    return &DNSHandler{aggregator: agg}
}

// GET /api/v1/dns/queries
func (dh *DNSHandler) GetDNSQueries(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // Parse query params for filtering
    query := r.URL.Query().Get("query")
    limit := r.URL.Query().Get("limit")

    queries, err := dh.aggregator.GetDNSQueries(ctx, query, limit)
    if err != nil {
        http.Error(w, "Failed to get DNS queries", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(queries)
}

// Similar methods for GetDNSZones, GetDNSPerformance, GetDNSThreats
```

### ЭТАП 3: DDoS Monitoring (Hour 3)

#### 3.1 DDoS типы
**File**: `backend/shared/types/ddos.go`

```go
package types

type DDoSAttack struct {
    ID           string `json:"id"`
    SourceIP     string `json:"source_ip"`
    TargetIP     string `json:"target_ip"`
    Protocol     string `json:"protocol"`
    Severity     string `json:"severity"`
    PacketsPerSec int64  `json:"packets_per_sec"`
    BytesPerSec   int64  `json:"bytes_per_sec"`
    StartTime    int64  `json:"start_time"`
    EndTime      int64  `json:"end_time"`
    Status       string `json:"status"`  // active, mitigated, blocked
}

type DDoSMitigation struct {
    AttackID    string   `json:"attack_id"`
    Method      string   `json:"method"`    // rate-limit, blackhole, scrubbing
    DropRate    float64  `json:"drop_rate"` // % of packets dropped
    AppliedAt   int64    `json:"applied_at"`
    Duration    int64    `json:"duration_seconds"`
}

type DDoSStatistics struct {
    TotalAttacks      int64   `json:"total_attacks"`
    ActiveAttacks     int     `json:"active_attacks"`
    MitigatedPackets  int64   `json:"mitigated_packets"`
    BlockedPackets    int64   `json:"blocked_packets"`
    SuccessRate       float64 `json:"success_rate"`
}
```

#### 3.2 DDoS Handler
**File**: `backend/bff/handlers/ddos.go`

```go
package handlers

type DDoSHandler struct {
    aggregator *aggregators.DashboardAggregator
}

// GET /api/v1/ddos/active-attacks
func (dh *DDoSHandler) GetActiveAttacks(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    attacks, err := dh.aggregator.GetActiveAttacks(ctx)
    if err != nil {
        http.Error(w, "Failed to get attacks", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(attacks)
}

// Similar for GetDDoSStatistics, GetMitigations, GetThreats
```

### ЭТАП 4: Performance Analytics (Hour 4)

#### 4.1 Performance типы
**File**: `backend/shared/types/performance.go`

```go
package types

type PathMetric struct {
    Path       string  `json:"path"`
    Protocol   string  `json:"protocol"`   // QUIC, MASQUE, WebSocket
    Traffic    float64 `json:"traffic_mbps"`
    Latency    float64 `json:"latency_ms"`
    ErrorRate  float64 `json:"error_rate"`
}

type LatencyBucket struct {
    Range  string `json:"range"`         // "0-10ms", "10-50ms", etc
    Count  int64  `json:"count"`
    Percent float64 `json:"percent"`
}

type ConnectionMetrics struct {
    Active    int64 `json:"active"`
    Pending   int64 `json:"pending"`
    Closing   int64 `json:"closing"`
    Closed    int64 `json:"closed"`
    PerSecond float64 `json:"per_second"` // new connections/sec
}
```

#### 4.2 Performance Handler
**File**: `backend/bff/handlers/performance.go`

```go
package handlers

type PerformanceHandler struct {
    aggregator *aggregators.DashboardAggregator
}

// GET /api/v1/perf/top-paths
func (ph *PerformanceHandler) GetTopPaths(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    limit := r.URL.Query().Get("limit")
    sortBy := r.URL.Query().Get("sort_by") // traffic, latency, errors

    paths, err := ph.aggregator.GetTopPaths(ctx, limit, sortBy)
    if err != nil {
        http.Error(w, "Failed to get paths", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(paths)
}

// Similar for GetLatencyDistribution, GetThroughput, GetConnections
```

### ЭТАП 5: Маршруты в main.go (Hour 5)

**File**: `backend/bff/main.go` (обновить authenticated routes section)

```go
// Add after dashboard routes
r.Group(func(r chi.Router) {
    r.Use(authMiddleware.VerifyToken)

    // Dashboard routes (existing)
    r.Get("/api/v1/dashboard/summary", dashboardHandler.GetSummary)
    r.Get("/api/v1/dashboard/health", dashboardHandler.GetHealth)
    r.Get("/api/v1/dashboard/metrics", dashboardHandler.GetOverviewMetrics)
    r.Get("/api/v1/dashboard/analytics", dashboardHandler.GetTrafficAnalytics)

    // NEW: Alert management
    alertHandler := handlers.NewAlertHandler(aggregator)
    r.Get("/api/v1/alerts", alertHandler.GetAlerts)
    r.Get("/api/v1/alerts/{id}", alertHandler.GetAlertDetails)
    r.Post("/api/v1/alerts/{id}/acknowledge", alertHandler.AckAlert)

    // NEW: DNS analytics
    dnsHandler := handlers.NewDNSHandler(aggregator)
    r.Get("/api/v1/dns/queries", dnsHandler.GetDNSQueries)
    r.Get("/api/v1/dns/zones", dnsHandler.GetDNSZones)
    r.Get("/api/v1/dns/performance", dnsHandler.GetDNSPerformance)
    r.Get("/api/v1/dns/threats", dnsHandler.GetDNSThreats)

    // NEW: DDoS monitoring
    ddosHandler := handlers.NewDDoSHandler(aggregator)
    r.Get("/api/v1/ddos/active-attacks", ddosHandler.GetActiveAttacks)
    r.Get("/api/v1/ddos/statistics", ddosHandler.GetDDoSStatistics)
    r.Get("/api/v1/ddos/mitigations", ddosHandler.GetMitigations)
    r.Get("/api/v1/ddos/threats", ddosHandler.GetThreats)

    // NEW: Performance analytics
    perfHandler := handlers.NewPerformanceHandler(aggregator)
    r.Get("/api/v1/perf/top-paths", perfHandler.GetTopPaths)
    r.Get("/api/v1/perf/latency-distribution", perfHandler.GetLatencyDistribution)
    r.Get("/api/v1/perf/throughput", perfHandler.GetThroughput)
    r.Get("/api/v1/perf/connections", perfHandler.GetConnections)
})
```

### ЭТАП 6: Тестирование (Hour 6)

#### 6.1 Тесты для новых endpoints
**File**: `backend/bff/handlers/alerts_test.go`

```go
package handlers

import (
    "testing"
    "net/http/httptest"
)

func TestGetAlerts(t *testing.T) {
    // Mock aggregator
    // Make request to GET /api/v1/alerts
    // Verify response structure
    // Check JSON marshalling
}

func TestAckAlert(t *testing.T) {
    // Test acknowledgement flow
}
```

---

## API DOCUMENTATION

### Request/Response Format

#### Alerts Endpoint
```
GET /api/v1/alerts
Authorization: Bearer <token>

Response:
{
  "alerts": [
    {
      "id": "alert-1",
      "severity": "warning",
      "message": "High latency detected",
      "service": "relay",
      "timestamp": "2025-10-28T10:30:00Z",
      "resolved": false
    }
  ],
  "total": 1
}
```

#### DNS Queries Endpoint
```
GET /api/v1/dns/queries?query=example.com&limit=10
Authorization: Bearer <token>

Response:
{
  "queries": [
    {
      "id": "dns-1",
      "query": "example.com",
      "type": "A",
      "response": "93.184.216.34",
      "latency_ms": 12,
      "timestamp": 1730110200,
      "status": "success"
    }
  ]
}
```

#### DDoS Active Attacks Endpoint
```
GET /api/v1/ddos/active-attacks
Authorization: Bearer <token>

Response:
{
  "attacks": [
    {
      "id": "attack-1",
      "source_ip": "192.0.2.1",
      "target_ip": "203.0.113.45",
      "protocol": "UDP",
      "severity": "critical",
      "packets_per_sec": 500000,
      "bytes_per_sec": 2147483648,
      "start_time": 1730110200,
      "status": "active"
    }
  ]
}
```

---

## DELIVERABLES

### Файлы которые нужно создать:
- [ ] `backend/shared/types/alerts.go` - Alert types
- [ ] `backend/shared/types/dns.go` - DNS types
- [ ] `backend/shared/types/ddos.go` - DDoS types
- [ ] `backend/shared/types/performance.go` - Performance types
- [ ] `backend/bff/handlers/alerts.go` - Alert handlers
- [ ] `backend/bff/handlers/dns.go` - DNS handlers
- [ ] `backend/bff/handlers/ddos.go` - DDoS handlers
- [ ] `backend/bff/handlers/performance.go` - Performance handlers
- [ ] `backend/bff/handlers/alerts_test.go` - Alert tests
- [ ] Обновить `backend/bff/main.go` с новыми маршрутами

### Файлы которые нужно обновить:
- [ ] `backend/bff/aggregators/dashboard.go` - добавить методы для новых endpoints

### Что добавить в Aggregator:
```
GetAlerts(ctx) -> AlertList
GetAlertByID(ctx, id) -> Alert
AcknowledgeAlert(ctx, id) -> error
GetDNSQueries(ctx, ...) -> []DNSQuery
GetDNSZones(ctx) -> []DNSZone
GetDNSPerformance(ctx) -> DNSMetrics
GetDNSThreats(ctx) -> []DNSThreat
GetActiveAttacks(ctx) -> []DDoSAttack
GetDDoSStatistics(ctx) -> DDoSStatistics
GetMitigations(ctx) -> []DDoSMitigation
GetThreats(ctx) -> []Threat
GetTopPaths(ctx, limit, sortBy) -> []PathMetric
GetLatencyDistribution(ctx) -> []LatencyBucket
GetThroughput(ctx) -> ConnectionMetrics
GetConnections(ctx) -> ConnectionMetrics
```

---

## SUCCESS CRITERIA

- [ ] Все новые endpoints добавлены в main.go
- [ ] Все handler'ы реализованы (основной функционал)
- [ ] Все типы определены в shared/types
- [ ] API документация полная
- [ ] Базовые тесты написаны
- [ ] Компиляция без ошибок: `go build ./bff`
- [ ] No lint errors: `golangci-lint run ./backend/...`

---

**Total Time**: 6 hours
**Status**: Ready to implement
**Next Task**: Task 1.3 - VictoriaMetrics Integration
