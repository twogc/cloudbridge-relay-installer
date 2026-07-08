# Duplicates Inventory — Go implementations

**Дата:** 2026-07-09  
**Шаг:** cloudbridge-relay-score-ports / step 4 (verify + wiring status)  
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
| `internal/network/score_allocator_production.go` | **keep** | — | Каноническая версия (production-ready, circuit breaker, `RetryConfigImproved`); **не wired** — 0 prod-вызовов (verified score-ports step 4) | Средний |
| `internal/network/retry_config.go` | — | **deleted step 1 (next)** | `network.RetryConfig` использовался только v2 allocator; `connection/retry.go` — отдельный тип | — |
| `internal/network/retry_config_improved.go` | **keep** | — | Используется production allocator; jitter, Retry-After | Низкий |
| `internal/cadence/service_client.go` | **keep** (ai-worker path) | merge позже | `CreateWorkflowServiceClient*` — 4 файла; вызывается из `client.go` → `cmd/ai-worker` | Средний — два параллельных пути |
| `internal/cadence/service_client_improved.go` | **keep** (relay path) | merge позже | `CreateWorkflowServiceClientImproved` — 2 файла; вызывается из `client_production.go` → `cmd/relay` | Средний |
| `internal/cadence/client.go` | **keep** (ai-worker) | rename/merge step 2 | `cadence.NewClient` — `cmd/ai-worker`, worker, tests | Средний |
| `internal/cadence/client_production.go` | **keep** (relay) | rename/merge step 2 | `cadence.NewProductionClient` — `cmd/relay/main.go` | Средний |
| `internal/nat/turn_server.go` | — | **deleted step 1** | `NewTURNServer` — 0 prod-вызовов; заменён enhanced | — |
| `internal/nat/turn_server_enhanced.go` | **keep** | — | `NewEnhancedTURNServer` — `cmd/relay`, integration tests | Низкий |
| `internal/ai/anomaly/detector.go` | **keep** (Service path) | — | `NewNetworkAnomalyDetector` — `internal/ai/service.go`, `cmd/ai-service`, relay middleware | — |
| `internal/ai/anomaly/optimized_detector.go` | — | **deleted step 3** | Только `optimized_service.go` + tests; 0 prod | — |
| `internal/ai/anomaly/real_anomaly_detector.go` | — | **deleted step 3** | Только `real_service.go` + tests; 0 prod | — |
| `internal/ai/anomaly/isolation_forest.go` | **keep** (с detector.go) | — | `SimpleIsolationForest` — используется `detector.go` | Низкий |
| `internal/ai/anomaly/real_isolation_forest.go` | — | **deleted step 3** | Только `real_anomaly_detector.go` | — |
| `internal/ai/predictive/load_predictor.go` | **keep** (Service path) | — | `NewLoadPredictor` — `service.go`, relay middleware | — |
| `internal/ai/predictive/optimized_predictor.go` | — | **deleted step 3** | Только `optimized_service.go` | — |
| `internal/ai/predictive/real_load_predictor.go` | — | **deleted step 3** | `real_service.go` + tests; 0 prod | — |
| `internal/ai/ml/model_implementations.go` | **keep** | — | Методы ML-типов; типы в `ml/types.go` | Низкий |
| `internal/ai/ml/types.go` | **keep** (new step 3) | — | Типы/конструкторы, вынесены из `real_models.go` | Низкий |
| `internal/ai/ml/real_models.go` | — | **deleted step 3** | Orchestrator real_*; типы перенесены в `types.go` | — |
| `internal/ai/ml/integration.go` | — | **deleted step 3** | `NewRealAIIntegration` — только tests | — |
| `internal/ai/pipeline/collectors.go` | **keep** | — | `NewRealPrometheusCollector` — `data_pipeline.go` (канонический pipeline) | Низкий |
| `internal/ai/pipeline/real_data_collector.go` | — | **deleted step 3** | Только `real_service.go` + tests | — |
| `internal/ai/routing/optimizer.go` | **keep** (Service path) | — | `NewRouteOptimizer` — `service.go`, relay middleware | — |
| `internal/ai/routing/optimized_optimizer.go` | — | **deleted step 3** | Только `optimized_service.go` | — |
| `internal/ai/cache/optimized_cache.go` | — | **deleted step 3** | Только `optimized_service.go` | — |
| `internal/ai/real_service.go` | — | **deleted step 3** | 0 prod entrypoints | — |
| `internal/ai/optimized_service.go` | — | **deleted step 3** | 0 prod entrypoints | — |

> **Примечание по `internal/ai/data/`:** директории нет; аудит имел в виду `internal/ai/pipeline/collectors.go` + `real_data_collector.go`.

---

## Параллельные AI-сервисы (контекст)

Три оболочки дублируют одни и те же подсистемы:

| Сервис | Файл | Prod entrypoint | Статус |
|--------|------|-----------------|--------|
| `ai.Service` | `service.go` | `cmd/ai-service`, `internal/relay/server_integration.go` | **активен (canonical)** |
| ~~`ai.RealService`~~ | ~~`real_service.go`~~ | — | **удалён step 3** |
| ~~`ai.OptimizedService`~~ | ~~`optimized_service.go`~~ | — | **удалён step 3** |

Консолидация завершена: canonical — `ai.Service` + `detector.go` / `load_predictor.go` / `optimizer.go` / `collectors.go`.

---

## Cadence: две production-линии

```
cmd/relay/main.go          → NewProductionClient → CreateWorkflowServiceClientImproved
cmd/ai-worker/main.go      → NewClient           → CreateWorkflowServiceClient(WithOptions)
internal/cadence/worker/   → *cadence.Client
```

`integration_test/cadence_integration_test.go` — **удалён step 2** (stale, ссылался на несуществующие `CadenceClient` / `NewCadenceClient`).

---

## SCORE allocators: canonical production, не wired

`internal/p2p/relay_manager_hybrid.go` принимает `network.TenantNetworkAllocator`, но `NewSCORETenantNetworkAllocatorProduction` не вызывается в репозитории (grep step 4: только `score_allocator_production.go`). Удалены v2/improved (`score_tenant_network_allocator*.go`) и `retry_config.go`; остаётся `score_allocator_production.go` + `retry_config_improved.go`.

**Wiring status (2026-07-09):** ✅ optional path — `p2p.NewRelayManagerHybridFromSCORE` + `SCORE_BASE_URL`/`SCORE_API_KEY` env; wired from `cmd/relay` `P2PIntegration` → P2P API WireGuard endpoints. Disabled when URL empty (default).

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
### cloudbridge-relay-cadence-ai / step 3 — AI safe merge (трек B)

- [x] Удалены `real_service.go`, `optimized_service.go` и exclusive `real_*` / `optimized_*` компоненты
- [x] Удалены связанные тесты (`real_service_*`, `optimized_service_*`, `error_handling_test.go`, `performance_test.go`, `ml/real_models_test.go`, `real_load_predictor_improved_test.go`)
- [x] Типы ML вынесены в `internal/ai/ml/types.go` (для `model_implementations.go`)
- [x] `ai.Service.OptimizeRoute` возвращает `*routing.OptimizedRoute` (убрана зависимость от удалённого `routing.Route`)
- [x] Canonical: `ai.Service` + `detector.go` / `load_predictor.go` / `optimizer.go` / `collectors.go`
- [ ] Cadence dual-path merge (трек A) — **step 4+** (medium risk, intentional dual entrypoint)
- [ ] Wiring production SCORE allocator — **отдельная пачка** (verified step 4: 0 prod callers; interface готов в `relay_manager_hybrid.go`)

### cloudbridge-relay-score-ports / step 4 — verify + wiring status

- [x] `go build ./cmd/relay/ ./cmd/ipam-service/ ./cmd/ai-worker/ ./cmd/ai-service/` — OK
- [x] Порты IPAM **8080**, AI **8083**, AI Worker **8085** — совпадают `cmd/*`, `openwiki/port-scheme.md`, `k8s/` ops, CNI
- [x] Wiring status SCORE allocator зафиксирован: **не wired** (только definition в `score_allocator_production.go`)
