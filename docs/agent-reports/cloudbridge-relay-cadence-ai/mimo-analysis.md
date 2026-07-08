# MiMo Analysis — Cadence dual-path + AI duplicates

**Дата:** 2026-07-09  
**Модель:** mimo-v2.5-pro (шаг 1/4 миссии `cloudbridge-relay-cadence-ai`)  
**Источники:** `docs/audit/DUPLICATES_INVENTORY.md`, `docs/audit/MIMO_AUDIT_REPORT.md` §2, live grep по `*.go`

> **Примечание (step 2):** MiMo на шаге 1 выдаёт только текст (без записи на диск). Этот файл зафиксирован Cursor на шаге 2 по структуре задания step 1 и верифицирован grep-командами 2026-07-09.

---

## 1. Cadence dual-path

### Текущая схема

```
cmd/relay/main.go
  → cadence.NewProductionClient()
  → CreateWorkflowServiceClientImproved()
  → internal/cadence/client_production.go + service_client_improved.go

cmd/ai-worker/main.go
  → cadence.NewClient()
  → CreateWorkflowServiceClient() / CreateWorkflowServiceClientWithOptions()
  → internal/cadence/client.go + service_client.go

internal/cadence/worker/
  → *cadence.Client (ai-worker path)
```

### Grep (prod, без `_test.go`)

| Символ | Файлы-использования |
|--------|---------------------|
| `NewProductionClient` | `cmd/relay/main.go`, `client_production.go` |
| `NewClient` (cadence) | `cmd/ai-worker/main.go`, `client.go` |
| `CreateWorkflowServiceClientImproved` | `client_production.go`, `service_client_improved.go` |
| `CreateWorkflowServiceClient*` | `client.go`, `service_client.go`, также fallback в `client_production.go` |

### Вердикт

| Файл | Действие | Риск | Обоснование |
|------|----------|------|-------------|
| `client_production.go` | **KEEP** | **medium** | Единственный путь `cmd/relay`; circuit breaker, retry, domain cache |
| `client.go` | **KEEP** | **medium** | Единственный путь `cmd/ai-worker` + worker |
| `service_client_improved.go` | **KEEP** | **medium** | Используется production client |
| `service_client.go` | **KEEP** | **medium** | Используется ai-worker client |
| `integration_test/cadence_integration_test.go` | **DELETE** | **low** | Stale: ссылается на несуществующие `CadenceClient` / `NewCadenceClient`; build tag `integration`; 0 prod-вызовов |

**Можно ли свести к одному API без поломки обоих entrypoints?**  
Нет в step 2. Разные конфиги, разные фабрики gRPC, разное поведение retry/CB. Безопасный merge — только thin wrapper / shared helper (step 3+), с регрессионными тестами для обоих бинарников.

---

## 2. AI тройки (Service / Real / Optimized)

### Prod entrypoints

| Оболочка | Файл | Вызовы из `cmd/` / prod `internal/` |
|----------|------|-------------------------------------|
| `ai.Service` | `service.go` | `cmd/ai-service`, `internal/relay/server_integration.go` — **активен** |
| `ai.RealService` | `real_service.go` | **0** (только `real_service.go` + `*_test.go`) |
| `ai.OptimizedService` | `optimized_service.go` | **0** (только `optimized_service.go` + `*_test.go`) |

### Компоненты real_* / optimized_* (prod grep)

| Конструктор | Prod-вызовы (не тест) |
|-------------|----------------------|
| `NewRealAnomalyDetector` | `real_service.go`, `ml/real_models.go` |
| `NewOptimizedDetector` | `optimized_service.go` |
| `NewRealLoadPredictor` | `real_service.go`, `ml/real_models.go` |
| `NewOptimizedPredictor` | `optimized_service.go` |
| `NewOptimizedOptimizer` | `optimized_service.go` |
| `NewRealDataCollector` | `real_service.go`, `real_anomaly_detector.go`, `real_load_predictor.go` |
| `NewRealMLModels` | `ml/integration.go` |
| `NewRealAIIntegration` | `ml/integration.go` (только) |

### Canonical (prod path `ai.Service`)

| Пакет | KEEP | Использование |
|-------|------|---------------|
| `anomaly/detector.go` | ✅ | `service.go`, relay middleware |
| `anomaly/isolation_forest.go` | ✅ | `detector.go` |
| `predictive/load_predictor.go` | ✅ | `service.go`, relay middleware |
| `routing/optimizer.go` | ✅ | `service.go`, relay middleware |
| `pipeline/collectors.go` | ✅ | `data_pipeline.go` |
| `ml/model_implementations.go` | ✅ | типы в пакете `ml` |

### Real / Optimized цепочки (мёртвы в prod, связаны между собой)

| Файл | Действие step 2 | Риск step 3 | Зависимости |
|------|-----------------|-------------|-------------|
| `real_service.go` | DEFER | medium | orchestrator для real_* |
| `optimized_service.go` | DEFER | medium | orchestrator для optimized_* |
| `anomaly/real_anomaly_detector.go` | DEFER | low–med | `real_service.go` |
| `anomaly/real_isolation_forest.go` | DEFER | low | `real_anomaly_detector.go` |
| `anomaly/optimized_detector.go` | DEFER | low | `optimized_service.go` |
| `predictive/real_load_predictor.go` | DEFER | low | `real_service.go` |
| `predictive/optimized_predictor.go` | DEFER | low | `optimized_service.go` |
| `routing/optimized_optimizer.go` | DEFER | low | `optimized_service.go` |
| `pipeline/real_data_collector.go` | DEFER | low | `real_service.go` |
| `ml/real_models.go` | DEFER | low | `integration.go`, tests |
| `ml/integration.go` | DEFER | low | только tests + `real_models.go` |

**Почему не DELETE в step 2:** у каждого файла есть внутренние вызовы в пакете `internal/ai` (не 0 внешних ссылок). Удаление по одному ломает сборку. Нужен пакетный cleanup в step 3 (вместе с тестами `real_service_*`, `optimized_service_*`, `ml/real_models_test.go`).

---

## 3. План действий (порядок)

### Step 2 (текущий) — только low-risk DELETE

| # | Действие | Файлы | Риск |
|---|----------|-------|------|
| 1 | DELETE stale integration test | `internal/cadence/integration_test/cadence_integration_test.go` | **low** |
| 2 | Не трогать cadence dual-path | все `client*.go`, `service_client*.go` | medium |
| 3 | Не трогать AI real/optimized цепочки | см. таблицу выше | medium (при пакетном удалении) |

### Step 3 — один безопасный merge (выбрать трек с lowest risk)

**Рекомендация MiMo: трек B (AI)** — lower risk, чем cadence merge:

1. Удалить `real_service.go` + `optimized_service.go` и все exclusive `real_*` / `optimized_*` файлы одним коммитом.
2. Удалить/сузить тесты: `real_service_integration_test.go`, `optimized_service_test.go`, `error_handling_test.go` (real), `performance_test.go` (real), `ml/real_models_test.go`, `real_load_predictor_improved_test.go`.
3. Оставить `ai.Service` + `detector.go` / `load_predictor.go` / `optimizer.go` / `collectors.go` как canonical.

**Альтернатива трек A (Cadence):** shared helper для `CreateWorkflowServiceClient*` без смены публичного API `NewClient` / `NewProductionClient` — риск **medium**, нужны тесты обоих entrypoints.

### Step 4 — верификация

- `go build ./cmd/relay/ ./cmd/ai-worker/ ./cmd/ai-service/`
- `go test -short ./internal/cadence/... ./internal/ai/...`
- Обновить `MIMO_AUDIT_REPORT.md` §2, `DUPLICATES_INVENTORY.md`

---

## 4. Что НЕ трогать в этой пачке

- `voly/`, `rep/`
- `internal/network/` (уже очищено в `cloudbridge-relay-next` step 1)
- `internal/nat/turn_server.go` (уже удалён)
- K8s / Terraform / monitoring consolidation
- Wiring SCORE allocator (`score_allocator_production.go`) — пачка `cloudbridge-relay-score-ports`
- Портовая схема — уже синхронизирована в `cloudbridge-relay-next` step 2
- Keycloak / quantum / `src/` — уже удалены в `cloudbridge-relay-fixes`

---

## 5. Next actions (cadence dual-path, step 3+)

1. Выделить общий `createWorkflowServiceClient(hostPort, opts)` в `internal/cadence/service_client_common.go`.
2. `service_client.go` и `service_client_improved.go` — thin wrappers без изменения сигнатур.
3. Документировать в `DUPLICATES_INVENTORY.md` две линии как «intentional dual entrypoint» до полного merge.
4. Добавить `go test -short` для `internal/cadence/client.go` и `client_production.go` (smoke).

---

*Анализ подготовлен для Cursor step 2/4. Удаления step 2 ограничены файлами с risk=low и 0 prod-вызовов.*
