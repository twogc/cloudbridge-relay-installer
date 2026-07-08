# Отчёт агента — cursor: inventory Go duplicates

**Дата:** 2026-07-09 01:33  
**Миссия:** cloudbridge-relay-medium  
**Шаг:** 1

## Выполнено

1. Прочитаны `docs/audit/MIMO_AUDIT_REPORT.md` (§2 Средний срок) и `AGENTS.md`.
2. Найдены пары/тройки `*_improved`, `*_enhanced`, `*_production` и AI-дубликаты (`optimized_*`, `real_*`) в:
   - `internal/network/` (score allocator, retry_config)
   - `internal/cadence/` (client, service_client)
   - `internal/nat/` (turn_server)
   - `internal/ai/anomaly/`, `predictive/`, `ml/`, `pipeline/`, `routing/`
3. Для каждой группы выполнен grep по конструкторам/типам — зафиксировано, какой файл активен в prod (`cmd/relay`, `cmd/ai-service`, `cmd/ai-worker`) vs только self/tests.
4. Создан `docs/audit/DUPLICATES_INVENTORY.md` — таблица path | keep | delete/merge | reason | risk.
5. Удалён очевидно мёртвый `internal/nat/turn_server.go` (`NewTURNServer` — 0 prod-вызовов; winner — `turn_server_enhanced.go`).
6. Исправлены ссылки в `openwiki/domain-networking.md` (путь TURN server → enhanced).
7. Остальные дубликаты **не удалялись** — у каждой группы 2+ активных пути или нет единого winner без merge API (правило шага 1).

## Изменённые файлы

| Путь | Действие |
|------|----------|
| `docs/audit/DUPLICATES_INVENTORY.md` | создан |
| `internal/nat/turn_server.go` | удалён |
| `openwiki/domain-networking.md` | обновлены 2 ссылки на TURN |
| `docs/agent-reports/cloudbridge-relay-medium/step01-inventory-go-duplicates.md` | создан (этот отчёт) |

## Проверки (команды и результат)

```bash
export PATH="$HOME/.local/go-install/go/bin:$PATH"
go version
# go version go1.25.12 linux/amd64

go build -o /tmp/cloudbridge-relay ./cmd/relay/
# exit 0 (успех после удаления turn_server.go)

# Подсчёт файлов с символами (выборочно):
rg -l 'NewTURNServer' --glob '*.go' .
# 1 файл до удаления (только turn_server.go); 0 после

rg -l 'NewEnhancedTURNServer' --glob '*.go' .
# cmd/relay/main.go, internal/tests/vpn_integration_test.go, turn_server_enhanced.go

rg -l 'NewProductionClient' --glob '*.go' .
# client_production.go, cmd/relay/main.go

rg -l 'NewClient' --glob '*.go' internal/cadence/
# client.go, client_test.go, cmd/ai-worker/main.go, ...
```

## Блокеры / не сделано

| Item | Причина |
|------|---------|
| Удаление score allocators (v2/improved/production) | Все 3 без внешних вызовов, но нет wiring — winner `production` выбран в аудите, merge отложен на шаг 2 |
| Удаление `retry_config.go` | Привязан к v2 allocator; удалять вместе с allocators |
| Merge cadence `client.go` + `client_production.go` | Обе линии активны: relay vs ai-worker |
| Merge AI anomaly/predictive/routing троек | Три параллельных сервиса (`Service`, `RealService`, `OptimizedService`); prod использует только `Service` + его компоненты |
| `integration_test/cadence_integration_test.go` | Ссылается на несуществующий `NewCadenceClient`; build tag `integration`, не влияет на relay build |
| `go build ./...` | Не запускался (scope: `./cmd/relay/` per task); `rep/` известно ломает full build |

**Следующий шаг (2/5):** консолидация по `DUPLICATES_INVENTORY.md` — начать с SCORE allocators + retry_config, затем cadence clients.
