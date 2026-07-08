# 2GC: cloudbridge-relay-score-ports

**Date:** 2026-07-09 02:10:51
**Total duration:** 313.7s
**Tasks:** 3 succeeded, 1 failed
**Tokens:** 650 in / 804 out
**Cost:** $0.0015

---

##  [mimo] mimo: SCORE wiring + port drift
*02:05:37 → 02:05:49*
Tokens: 650↑ 804↓ | 11206ms

```
```bash
cat docs/audit/DUPLICATES_INVENTORY.md
```
```

##  [cursor] cursor: write mimo + port alignment
*02:05:50 → 02:08:49*
Tokens: 0↑ 0↓ | 177090ms

```
**Шаг 2/4 выполнен.** Порты IPAM/AI выровнены по MiMo; отчёты записаны на диск.

## Решение «одна правда»

Для auxiliary microservices источником правды выбраны **`cmd/*` defaults + `k8s/` ops manifests** (не старый SERVICE_CATALOG 8083/8085/8086):

| Сервис | Канонический порт | Источник |
|---|---:|---|
| IPAM | **8080** | `cmd/ipam-service`, `configs/cni/cloudbridge-cni.conf` |
| AI Service | **8083** | `AI_PORT`, `k8s/deployments/ai-service-deployment.yaml` |
| AI Worker | **8085** | `cmd/ai-worker`, `k8s/services/ai-service.yaml` |

Код не менялся — он уже совпадал с `k8s/`. Обновлены только документы. Перенос IPAM на 8083 отклонён: конфликт с relay `p2p_api_https.port` и CNI URL.

## Файлы

- `docs/agent-reports/cloudbridge-relay-score-ports/mimo-analysis.md` — MiMo-анализ
- `docs/ag
... [619 more chars]
```

##  [cursor] cursor: SCORE production wiring
*02:08:50 → 02:08:51*
Tokens: 0↑ 0↓ | 502ms

**Error:** Bridge exited before discovery with status 1: cursor-sdk-bridge failed: Error: Missing value for --tool-callback-auth-token
    at takeValue (file:///home/user1/smarty/cloudbridge-relay-installer/voly/.venv/lib/python3.10/site-packages/cursor_sdk/_vendor/bridge/dist/bin/cursor-sdk-bridge.js:155:15)
    at parseArgs (file:///home/user1/smarty/cloudbridge-relay-installer/voly/.venv/lib/python3.10/site-packages/cursor_sdk/_vendor/bridge/dist/bin/cursor-sdk-bridge.js:125:48)
    at main (file:///home/user1/smarty/cloudbridge-relay-installer/voly/.venv/lib/python3.10/site-packages/cursor_sdk/_vendor/bridge/dist/bin/cursor-sdk-bridge.js:24:18)
    at file:///home/user1/smarty/cloudbridge-relay-installer/voly/.venv/lib/python3.10/site-packages/cursor_sdk/_vendor/bridge/dist/bin/cursor-sdk-bridge.js:94:1
    at ModuleJob.run (node:internal/modules/esm/module_job:371:25)
    at async onImport.tracePromise.__proto__ (node:internal/modules/esm/loader:669:26)
    at async asyncRunEntryPointWithESMLoader (node:internal/modules/run_main:101:5)


##  [cursor] cursor: verify + SUMMARY
*02:08:51 → 02:10:51*
Tokens: 0↑ 0↓ | 116512ms

```
**Шаг 4/4 завершён.** Миссия `cloudbridge-relay-score-ports` закрыта.

## Результаты верификации

**Сборка** — успешно:
```bash
PATH="$HOME/.local/go-install/go/bin:$PATH" go build ./cmd/relay/ ./cmd/ipam-service/ ./cmd/ai-worker/ ./cmd/ai-service/
# exit 0, Go 1.25.12
```

**Порты auxiliary** — совпадают с `openwiki/port-scheme.md`:

| Сервис | Порт | cmd | k8s/CNI |
|--------|-----:|-----|---------|
| IPAM | **8080** | `cmd/ipam-service` | CNI `serviceUrl` |
| AI Service | **8083** | `AI_PORT` default | `k8s/deployments/`, `k8s/services/` |
| AI Worker | **8085** | health server | `k8s/services/ai-service.yaml` |

**Wiring SCORE allocator** — не подключён: `NewSCORETenantNetworkAllocatorProduction` имеет 0 prod-вызовов; interface в `relay_manager_hybrid.go` готов → отдельная пачка.

## С
... [531 more chars]
```
