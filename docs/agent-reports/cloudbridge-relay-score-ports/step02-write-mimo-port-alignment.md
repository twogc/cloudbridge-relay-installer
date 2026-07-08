# Отчёт агента — cursor: write mimo + port alignment

**Дата:** 2026-07-09 02:05  
**Миссия:** cloudbridge-relay-score-ports  
**Шаг:** 2

## Выполнено

1. Прочитаны `openwiki/quickstart.md`, `docs/audit/MIMO_AUDIT_REPORT.md`, `openwiki/port-scheme.md`, `docs/operations/SERVICE_CATALOG.md`.
2. Сохранён MiMo-анализ по IPAM/AI в `docs/agent-reports/cloudbridge-relay-score-ports/mimo-analysis.md` (runtime grep, конфликт IPAM 8083 vs `p2p_api_https`, host Prometheus 9090 — только комментарий).
3. **Решение «одна правда»:** для auxiliary microservices authoritative — **`cmd/*` defaults + `k8s/` ops manifests** (не legacy SERVICE_CATALOG 8083/8085/8086):
   - IPAM → **8080** (`cmd/ipam-service`, `configs/cni/cloudbridge-cni.conf`)
   - AI Service → **8083** (`AI_PORT`, `k8s/deployments/ai-service-deployment.yaml`)
   - AI Worker → **8085** (`cmd/ai-worker`, `k8s/services/ai-service.yaml`)
4. Синхронизированы документы (код **не менялся** — уже совпадал с `k8s/`):
   - `openwiki/port-scheme.md` — таблица Auxiliary, K8s exposure, AI drift section, host-note про smarty Prometheus 9090
   - `docs/operations/SERVICE_CATALOG.md` — порты IPAM/AI/AI Worker, metrics, network topology; удалены блоки Drift
5. Smarty Docker / host Prometheus **9090** не трогали.

## Изменённые файлы

| Путь | Действие |
|------|----------|
| `docs/agent-reports/cloudbridge-relay-score-ports/mimo-analysis.md` | создан — MiMo IPAM/AI port analysis |
| `openwiki/port-scheme.md` | IPAM 8080, AI 8083, AI Worker 8085; source-of-truth note; stub drift k8s-manifests |
| `docs/operations/SERVICE_CATALOG.md` | каталожные порты IPAM/AI/Worker + metrics/topology |
| `docs/agent-reports/cloudbridge-relay-score-ports/step02-write-mimo-port-alignment.md` | создан (этот отчёт) |

## Проверки (команды и результат)

```bash
# Runtime defaults
rg -n 'Addr: ":808|AI_PORT|port = "808' cmd/ipam-service/main.go cmd/ai-service/main.go cmd/ai-worker/main.go
```

```
cmd/ipam-service/main.go:54:    Addr:    ":8080",
cmd/ai-service/main.go:74:    port := os.Getenv("AI_PORT")
cmd/ai-service/main.go:76:        port = "8083"
cmd/ai-worker/main.go:46:        Addr: ":8085",
```

```bash
# Docs aligned
rg -n '\| IPAM \||\| AI Service \||\| AI Worker \|' openwiki/port-scheme.md
rg -n '^\*\*Port\*\*:' docs/operations/SERVICE_CATALOG.md | rg -n '8080|8083|8085'
```

```
| IPAM | 8080 | TCP | ...
| AI Service | 8083 | TCP | ...
| AI Worker | 8085 | TCP | ...
# SERVICE_CATALOG: IPAM 8080, AI 8083, AI Worker 8085
```

```bash
# CNI + k8s consistency
rg '8080' configs/cni/cloudbridge-cni.conf
rg '8083' k8s/deployments/ai-service-deployment.yaml | head -3
go build -o /dev/null ./cmd/ipam-service/ ./cmd/ai-service/ ./cmd/ai-worker/
```

```
serviceUrl: ...:8080
containerPort: 8083 / AI_PORT=8083
go build: OK (exit 0)
```

## Блокеры / не сделано

| Item | Причина |
|------|---------|
| `k8s-manifests/06-services.yaml` AI **8080/9090** | Prod stub placeholders; исправление — отдельный k8s-consol scope |
| `k8s/services/ai-service.yaml` metrics **8084** vs runtime `/metrics` on **8083** | Service declaration drift; требует решения по split-port или упрощению Service |
| Изменение `cmd/*` на старые catalog-порты (8083/8085/8086) | Отклонено: ломает CNI (8080), k8s ops, конфликт с relay `p2p_api_https` 8083 |
| Commit / push | Не запрашивались |
