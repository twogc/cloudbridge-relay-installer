# НЕДЕЛЯ 1: Dashboard BFF Integration - ACTION PLAN
**Statbility Phase - Month 1**
**Дата начала**: 2025-10-28
**Дата окончания**: 2025-11-04

---

## ЦЕЛЬ НЕДЕЛИ
Подключить реальные данные к Dashboard BFF от VictoriaMetrics, Redis, и других сервисов CloudBridge. Реализовать real-time обновления через SSE.

**Текущий статус Dashboard**: Phase 2 (40% готовности) - Core Dashboard In Progress

---

## ЗАДАЧА 1.1: Установка и запуск всех сервисов
**Статус**: PENDING
**Часов**: 4 часа

### Описание
Установить и убедиться, что работают все зависимые сервисы:
- VictoriaMetrics (для метрик)
- Redis (для кэширования)
- Control Plane (Zitadel, для auth)
- AI Service (для insights)
- Relay Service (для трафика)
- DDoS Protection (для защиты)

### Инструкции
```bash
# 1. Перейти в каталог каждого сервиса
cd /Users/maxlanies/Git/2GC/cloudbridge-relay-installer/oss-repositories

# 2. Для каждого сервиса запустить Docker Compose
for repo in cloudbridge-monitoring cloudbridge-control-plane cloudbridge-ai-service \
            cloudbridge-scalable-relay cloudbridge-ddos-protection; do
  cd $repo
  docker-compose up -d
  cd ..
done

# 3. Проверить статус
docker ps | grep cloudbridge
```

### Проверочный список
- [ ] VictoriaMetrics доступен на `http://localhost:8428`
- [ ] Redis доступен на `localhost:6379`
- [ ] Control Plane/Zitadel доступен на определенном порту
- [ ] AI Service работает на HTTP
- [ ] Relay Service получает трафик
- [ ] DDoS Protection собирает метрики

### Результат
Файл: `/Users/maxlanies/Git/2GC/cloudbridge-relay-installer/SERVICES_STATUS.md`

---

## ЗАДАЧА 1.2: Проверка и расширение API структуры Dashboard BFF
**Статус**: PENDING
**Часов**: 6 часов

### Описание
Проверить текущую структуру BFF API и убедиться, что все endpoint'ы работают с реальными данными.

### Текущие endpoint'ы (в main.go)
```
GET /api/v1/dashboard/summary       - Собирает данные в реальном времени
GET /api/v1/dashboard/health        - Здоровье всех сервисов
GET /api/v1/dashboard/metrics       - Метрики VictoriaMetrics
GET /api/v1/dashboard/analytics     - Аналитика трафика (от/до времени)
GET /api/v1/relay/status            - Статус Relay сервиса
GET /api/v1/p2p/status              - Статус P2P сервиса
GET /api/v1/ai/insights             - AI insights
GET /api/v1/edge-pops/status        - Статус Edge PoP
```

### Требуемые работы
1. **Проверить каждый endpoint**:
   ```bash
   # Получить токен (нужен Bearer token от Zitadel)
   curl -X GET http://localhost:8080/api/v1/dashboard/summary \
     -H "Authorization: Bearer <TOKEN>"
   ```

2. **Расширить `GetTrafficAnalytics`** (dashboard.go:222-295)
   - Добавить поддержку фильтрации по PoP location
   - Добавить фильтрацию по protocol (QUIC, MASQUE, WebSocket)
   - Добавить DDoS attack frequency

3. **Добавить новые endpoint'ы** в main.go (118-135):
   ```go
   // Новые endpoint'ы
   r.Get("/api/v1/alerts/list", dashboardHandler.GetAlerts)
   r.Get("/api/v1/alerts/{id}", dashboardHandler.GetAlertDetails)
   r.Post("/api/v1/alerts/{id}/acknowledge", dashboardHandler.AckAlert)

   r.Get("/api/v1/dns/queries", dashboardHandler.GetDNSQueries)
   r.Get("/api/v1/ddos/active-attacks", dashboardHandler.GetActiveAttacks)
   r.Get("/api/v1/perf/top-paths", dashboardHandler.GetTopPaths)
   ```

### Проверочный список
- [ ] Все endpoint'ы возвращают 200 OK с реальными данными
- [ ] Кэширование работает (Redis)
- [ ] Timeout обработан правильно
- [ ] Error handling работает корректно
- [ ] Новые endpoint'ы документированы

### Результат
- Файл: `backend/bff/handlers/dashboard_extended.go` (новые обработчики)
- Обновлено: `backend/bff/main.go` (новые маршруты)

---

## ЗАДАЧА 1.3: Внедрение реальных данных из VictoriaMetrics
**Статус**: PENDING
**Часов**: 8 часов

### Описание
Убедиться, что все метрики правильно получаются из VictoriaMetrics и кэшируются в Redis.

### Текущее состояние VictoriaMetrics клиента
Файл: `backend/services/metrics/victoriametrics.go`

**Уже реализовано:**
- `Query()` - мгновенные запросы
- `RangeQuery()` - запросы по временному диапазону
- `GetMetricSummary()` - сводка метрик

**Требуемые работы:**
1. **Добавить новые метрики** в `GetMetricSummary()`:
   ```go
   // Добавить к существующим:
   summary.DDoSAttacksPerHour = // Запрос из DDoS Protection
   summary.DNSQueriesPerSecond = // Из DNS Network
   summary.P2PMeshHealth = // Из P2P
   summary.EdgePoPLatency = // Latency к каждому PoP
   ```

2. **Добавить методы для детальной аналитики**:
   ```go
   func (v *VictoriaMetricsClient) GetDDosMetrics(ctx context.Context) (*DDoSMetrics, error)
   func (v *VictoriaMetricsClient) GetDNSMetrics(ctx context.Context) (*DNSMetrics, error)
   func (v *VictoriaMetricsClient) GetP2PMetrics(ctx context.Context) (*P2PMetrics, error)
   func (v *VictoriaMetricsClient) GetEdgePoPMetrics(ctx context.Context, location string) (*EdgePoPMetrics, error)
   ```

3. **Реализовать кэширование в Redis**:
   - Dashboard summary: 30s TTL
   - Traffic analytics: 2m TTL
   - Health checks: 10s TTL
   - Alerts: 5m TTL

### PromQL запросы (примеры)
```promql
# Трафик через Relay
sum(rate(relay_quic_throughput_bytes_total[5m])) / 1e9 * 8

# Активные QUIC соединения
sum(relay_quic_connections_active)

# DDoS attacks
sum(increase(ddos_attacks_total[1h]))

# DNS queries
sum(rate(coredns_dns_request_total[5m]))

# P2P connections
sum(p2p_connections_active)

# Edge PoP latency
histogram_quantile(0.95, rate(edge_pop_latency_milliseconds_bucket[5m]))
```

### Проверочный список
- [ ] Все метрики успешно получаются из VictoriaMetrics
- [ ] Redis кэш работает
- [ ] TTL правильный для каждого типа метрик
- [ ] Fallback на хардкод работает при недоступности VM
- [ ] Логирование показывает реальные данные

### Результат
- Обновлено: `backend/services/metrics/victoriametrics.go`
- Новый файл: `backend/services/metrics/detailed_metrics.go` (расширенные метрики)
- Обновлено: `backend/bff/aggregators/dashboard.go` (использование новых метрик)

---

## ЗАДАЧА 1.4: Реализация Server-Sent Events (SSE) для real-time обновлений
**Статус**: PENDING
**Часов**: 10 часов

### Описание
Реализовать real-time обновления через Server-Sent Events вместо polling.

### Текущее состояние
Файл: `backend/bff/handlers/websocket.go` - содержит WebSocket handler, но нужен SSE

### Требуемые работы

#### 1. Создать SSE sender/broadcaster
```go
// backend/bff/sse/broadcaster.go
type EventBroadcaster struct {
    clients    map[*Client]bool
    broadcast  chan *Event
    register   chan *Client
    unregister chan *Client
}

type Event struct {
    Type string          // "metrics", "health", "alert", "analytics"
    Data interface{}    // JSON payload
    Timestamp time.Time
}
```

#### 2. Реализовать SSE endpoint'ы в main.go
```go
// Прямая трансляция событий в реальном времени
r.Get("/api/v1/events/stream", sseHandler.StreamEvents)  // Все события
r.Get("/api/v1/events/metrics", sseHandler.StreamMetrics)
r.Get("/api/v1/events/alerts", sseHandler.StreamAlerts)
r.Get("/api/v1/events/health", sseHandler.StreamHealth)
```

#### 3. SSE формат
```
event: metrics
data: {"timestamp":"2025-10-28T12:00:00Z","connections":1000,"throughput":5.2}

event: alert
data: {"id":"alert-1","severity":"warning","message":"High latency"}

event: health
data: {"relay":"healthy","p2p":"degraded","ai":"healthy"}
```

#### 4. Frontend интеграция (Next.js)
```typescript
// frontend/lib/hooks/useSSE.ts
export function useSSEMetrics() {
    const [metrics, setMetrics] = useStateRefreshInterval()

    useEffect(() => {
        const eventSource = new EventSource('/api/v1/events/metrics')
        eventSource.addEventListener('metrics', (event) => {
            setMetrics(JSON.parse(event.data))
        })
        return () => eventSource.close()
    }, [])

    return metrics
}
```

### Проверочный список
- [ ] SSE endpoint'ы работают без ошибок
- [ ] События отправляются в реальном времени
- [ ] Клиент может подписаться/отписаться
- [ ] Нет утечек памяти при длительных соединениях
- [ ] Frontend правильно потребляет SSE

### Результат
- Новый файл: `backend/bff/sse/broadcaster.go`
- Новый файл: `backend/bff/sse/events.go`
- Новый файл: `backend/bff/handlers/sse.go`
- Обновлено: `backend/bff/main.go` (SSE маршруты)
- Новый файл: `frontend/lib/hooks/useSSE.ts`

---

## ЗАДАЧА 1.5: Integration тесты для Dashboard BFF
**Статус**: PENDING
**Часов**: 6 часов

### Описание
Написать integration тесты для всех BFF endpoint'ов и проверить корректность работы.

### Требуемые тесты

#### 1. API endpoint тесты
```go
// backend/bff/handlers/dashboard_test.go
func TestGetSummary(t *testing.T)
func TestGetHealth(t *testing.T)
func TestGetMetrics(t *testing.T)
func TestGetTrafficAnalytics(t *testing.T)
```

#### 2. Aggregator тесты
```go
// backend/bff/aggregators/dashboard_test.go
func TestGetSummaryWithCaching(t *testing.T)
func TestGetSummaryWithoutCache(t *testing.T)
func TestServiceHealthCheck(t *testing.T)
```

#### 3. VictoriaMetrics тесты
```go
// backend/services/metrics/victoriametrics_test.go
func TestQuery(t *testing.T)
func TestRangeQuery(t *testing.T)
func TestMetricSummary(t *testing.T)
```

#### 4. End-to-End тесты
```bash
# Запустить dashboard-bff
go run ./bff

# Проверить endpoint'ы
curl http://localhost:8080/api/v1/dashboard/summary \
  -H "Authorization: Bearer <TOKEN>"

# Проверить метрики
curl http://localhost:8080/metrics | grep dashboard
```

### Проверочный список
- [ ] Все unit тесты проходят (>80% coverage)
- [ ] Integration тесты работают
- [ ] E2E тесты с реальными данными успешны
- [ ] Mock сервисы работают правильно
- [ ] Документация по тестам написана

### Результат
- Новый файл: `backend/bff/handlers/dashboard_test.go`
- Новый файл: `backend/bff/aggregators/dashboard_test.go`
- Новый файл: `backend/services/metrics/victoriametrics_test.go`
- Новый файл: `backend/test/integration_test.go`

---

## ПРОГРЕСС

| Задача | Статус | % | Часов |
|--------|--------|----|----|
| 1.1 Установка сервисов | ⏳ PENDING | 0% | 4h |
| 1.2 API структура | ⏳ PENDING | 0% | 6h |
| 1.3 VictoriaMetrics | ⏳ PENDING | 0% | 8h |
| 1.4 SSE real-time | ⏳ PENDING | 0% | 10h |
| 1.5 Integration тесты | ⏳ PENDING | 0% | 6h |
| **ИТОГО** | | **0%** | **34h** |

---

## ЗАВИСИМОСТИ

```
1.1 Установка сервисов
  ↓
1.2 API структура (зависит от доступности сервисов)
  ↓
1.3 VictoriaMetrics (зависит от доступности VM)
  ↓
1.4 SSE real-time (может быть параллельно с 1.3)
  ↓
1.5 Integration тесты (должны быть в конце)
```

---

## ССЫЛКИ

- Backend main: [backend/bff/main.go](../oss-repositories/cloudbridge-dashboard/backend/bff/main.go)
- Handlers: [backend/bff/handlers/dashboard.go](../oss-repositories/cloudbridge-dashboard/backend/bff/handlers/dashboard.go)
- VictoriaMetrics: [backend/services/metrics/victoriametrics.go](../oss-repositories/cloudbridge-dashboard/backend/services/metrics/victoriametrics.go)
- Aggregator: [backend/bff/aggregators/dashboard.go](../oss-repositories/cloudbridge-dashboard/backend/bff/aggregators/dashboard.go)
- README: [cloudbridge-dashboard/README.md](../oss-repositories/cloudbridge-dashboard/README.md)

---

## 📞 КОНТАКТЫ И ВОПРОСЫ

При возникновении вопросов обратитесь к:
- Документации в `docs/` каталоге
- API документации в `backend/README.md`
- Issue tracker в GitHub
