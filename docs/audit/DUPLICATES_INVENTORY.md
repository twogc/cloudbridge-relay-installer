# Duplicates Inventory — Go implementations

**Дата:** 2026-07-09  
**Шаг:** cloudbridge-relay-medium / step 1  
**Источник:** [MIMO_AUDIT_REPORT.md §2](MIMO_AUDIT_REPORT.md), grep по символам в `*.go`

Методика: для каждой пары/тройки подсчитаны файлы, ссылающиеся на конструктор/тип (ripgrep).  
«Активен» = есть вызовы из `cmd/`, `internal/` (не только из файла-определения и его тестов).  
«Мёртв» = 0 внешних ссылок, только self-reference в пакете.

---

## Сводная таблица

| path | keep | delete/merge | reason | risk |
|------|------|--------------|--------|------|
| `internal/network/score_tenant_network_allocator.go` | — | **deleted step 1 (next)** | `NewSCORETenantNetworkAllocatorV2` — 0 внешних вызовов | — |
| `internal/network/score_tenant_network_allocator_improved.go` | — | **deleted step 1 (next)** | `NewSCORETenantNetworkAllocatorImproved` — 0 внешних вызовов | — |
| `internal/network/score_allocator_production.go` | **keep** | — | Каноническая версия (production-ready, circuit breaker, `RetryConfigImproved`); 0 внешних вызовов — кандидат на wiring в step 2 | Средний |
| `internal/network/retry_config.go` | — | **deleted step 1 (next)** | `network.RetryConfig` использовался только v2 allocator; `connection/retry.go` — отдельный тип | — |
| `internal/network/retry_config_improved.go` | **keep** | — | Используется production allocator; jitter, Retry-After | Низкий |
| `internal/cadence/service_client.go` | **keep** (ai-worker path) | merge позже | `CreateWorkflowServiceClient*` — 4 файла; вызывается из `client.go` → `cmd/ai-worker` | Средний — два параллельных пути |
| `internal/cadence/service_client_improved.go` | **keep** (relay path) | merge позже | `CreateWorkflowServiceClientImproved` — 2 файла; вызывается из `client_production.go` → `cmd/relay` | Средний |
| `internal/cadence/client.go` | **keep** (ai-worker) | rename/merge step 2 | `cadence.NewClient` — `cmd/ai-worker`, worker, tests | Средний |
| `internal/cadence/client_production.go` | **keep** (relay) | rename/merge step 2 | `cadence.NewProductionClient` — `cmd/relay/main.go` | Средний |
| `internal/nat/turn_server.go` | — | **deleted step 1** | `NewTURNServer` — 0 prod-вызовов; заменён enhanced | — |
| `internal/nat/turn_server_enhanced.go` | **keep** | — | `NewEnhancedTURNServer` — `cmd/relay`, integration tests | Низкий |
| `internal/ai/anomaly/detector.go` | **keep** (Service path) | merge step 2+ | `NewNetworkAnomalyDetector` — `internal/ai/service.go`, `cmd/ai-service`, relay middleware | Средний |
| `internal/ai/anomaly/optimized_detector.go` | defer | merge | `NewOptimizedDetector` — только `optimized_service.go` + tests | Низкий |
| `internal/ai/anomaly/real_anomaly_detector.go` | defer | merge | `NewRealAnomalyDetector` — `real_service.go` + tests | Низкий |
| `internal/ai/anomaly/isolation_forest.go` | **keep** (с detector.go) | — | `SimpleIsolationForest` — используется `detector.go` | Низкий |
| `internal/ai/anomaly/real_isolation_forest.go` | defer | merge с real detector | `RealIsolationForest` — только `real_anomaly_detector.go` | Низкий |
| `internal/ai/predictive/load_predictor.go` | **keep** (Service path) | merge step 2+ | `NewLoadPredictor` — `service.go`, relay middleware | Средний |
| `internal/ai/predictive/optimized_predictor.go` | defer | merge | только `optimized_service.go` | Низкий |
| `internal/ai/predictive/real_load_predictor.go` | defer | merge | `real_service.go` + tests | Низкий |
| `internal/ai/ml/model_implementations.go` | **keep** | — | Методы для типов из `real_models.go` (тот же пакет `ml`) | Низкий |
| `internal/ai/ml/real_models.go` | defer | merge | `NewRealMLModels` — `ml/integration.go` + tests; не в prod `cmd/ai-service` | Низкий |
| `internal/ai/ml/integration.go` | defer | merge или delete | `NewRealAIIntegration` — только `real_models_test.go` | Низкий |
| `internal/ai/pipeline/collectors.go` | **keep** | — | `NewRealPrometheusCollector` — `data_pipeline.go` (канонический pipeline) | Низкий |
| `internal/ai/pipeline/real_data_collector.go` | defer | merge | `NewRealDataCollector` — `real_service.go` + tests | Низкий |
| `internal/ai/routing/optimizer.go` | **keep** (Service path) | merge step 2+ | `NewRouteOptimizer` — `service.go`, relay middleware | Средний |
| `internal/ai/routing/optimized_optimizer.go` | defer | merge | только `optimized_service.go` | Низкий |

> **Примечание по `internal/ai/data/`:** директории нет; аудит имел в виду `internal/ai/pipeline/collectors.go` + `real_data_collector.go`.

---

## Параллельные AI-сервисы (контекст)

Три оболочки дублируют одни и те же подсистемы:

| Сервис | Файл | Prod entrypoint | Статус |
|--------|------|-----------------|--------|
| `ai.Service` | `service.go` | `cmd/ai-service`, `internal/relay/server_integration.go` | **активен** |
| `ai.RealService` | `real_service.go` | — (только тесты) | мёртв в prod |
| `ai.OptimizedService` | `optimized_service.go` | — (только тесты) | мёртв в prod |

Консолидация компонентов (anomaly/predictive/routing) должна учитывать, какой сервис остаётся canonical.

---

## Cadence: две production-линии

```
cmd/relay/main.go          → NewProductionClient → CreateWorkflowServiceClientImproved
cmd/ai-worker/main.go      → NewClient           → CreateWorkflowServiceClient(WithOptions)
internal/cadence/worker/   → *cadence.Client
```

`integration_test/cadence_integration_test.go` ссылается на несуществующие `CadenceClient` / `NewCadenceClient` (build tag `integration`) — stale, не блокирует `go build ./cmd/relay/`.

---

## SCORE allocators: canonical production, не wired

`internal/p2p/relay_manager_hybrid.go` принимает `network.TenantNetworkAllocator`, но `NewSCORETenantNetworkAllocatorProduction` не вызывается в репозитории. Удалены v2/improved (`score_tenant_network_allocator*.go`) и `retry_config.go`; остаётся `score_allocator_production.go` + `retry_config_improved.go`.

---

## Выполнено

### cloudbridge-relay-medium / step 1

- [x] Инвентаризация grep
- [x] `docs/audit/DUPLICATES_INVENTORY.md`
- [x] Удалён `internal/nat/turn_server.go` (0 внешних вызовов)
- [x] Обновлены ссылки в `openwiki/domain-networking.md`

### cloudbridge-relay-next / step 1

- [x] Удалён `internal/network/score_tenant_network_allocator.go` (V2, 0 внешних вызовов)
- [x] Удалён `internal/network/score_tenant_network_allocator_improved.go` (0 внешних вызовов)
- [x] Удалён `internal/network/retry_config.go` (только v2 allocator)
- [x] Canonical: `score_allocator_production.go` + `retry_config_improved.go`
- [ ] Wiring production allocator в `cmd/relay` — **step 2+**
- [ ] Cadence dual-path merge — **step 2+** (не трогать в step 1)
- [ ] Остальные дубликаты (AI тройки и т.д.) — **step 2+**
