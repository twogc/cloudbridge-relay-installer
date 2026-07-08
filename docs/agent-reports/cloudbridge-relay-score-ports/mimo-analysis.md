# MiMo Analysis — IPAM / AI port alignment

**Дата:** 2026-07-09  
**Модель:** mimo-v2.5-pro (миссия `cloudbridge-relay-score-ports`, шаг 2)  
**Источники:** `docs/audit/MIMO_AUDIT_REPORT.md` §«Несогласованность портовой схемы», `openwiki/port-scheme.md`, live grep по `cmd/*`, `k8s/`, `configs/`

---

## 1. Проблема (MiMo Executive Summary)

MiMo зафиксировал **три независимых источника** портов relay и вспомогательных сервисов:

| Источник | Main Relay | AI Service | IPAM |
|---|---:|---:|---:|
| `openwiki/quickstart.md` (до sync) | 9090 | 8082 | — |
| `docs/operations/SERVICE_CATALOG.md` (до step 2) | 8080 | 8085 | 8083 |
| `configs/config.json` | 5550 (+ 9090 http_api) | — | — |
| `cmd/*` runtime | — | **8083** | **8080** |
| `k8s/` ops | — | **8083** | — (CNI → **8080**) |

**Влияние:** деплой и мониторинг по устаревшему каталогу не совпадают с реальными listener'ами; CNI (`configs/cni/cloudbridge-cni.conf`) указывает IPAM на **8080**, а каталог — на **8083**.

**Рекомендация MiMo:** единый источник правды в `openwiki/` + синхронизация остальных документов. Relay-порты уже выровнены (medium/next steps). **Остаток score-ports:** IPAM + AI + AI Worker.

---

## 2. Runtime evidence (grep, 2026-07-09)

### `cmd/*` defaults

| Бинарник | Порт | Источник |
|---|---:|---|
| `cmd/ipam-service` | **8080** | `main.go:54` `Addr: ":8080"` |
| `cmd/ai-service` | **8083** | `main.go:74-76` `AI_PORT` default `"8083"` |
| `cmd/ai-worker` | **8085** | `main.go:46` health `Addr: ":8085"` |

### `k8s/` ops (canonical deploy path)

| Объект | Порт | Файл |
|---|---:|---|
| AI Service HTTP | **8083** | `k8s/deployments/ai-service-deployment.yaml`, `k8s/services/ai-service.yaml` |
| AI Worker metrics | **8085** | `k8s/deployments/ai-worker-deployment.yaml`, `k8s/services/ai-service.yaml` |
| Prometheus scrape AI | **8083** | `k8s/monitoring/prometheus-config.yaml` |
| Cronjob → AI API | **8083** | `k8s/cronjobs/ai-telegram-cronjob.yaml` |

### Интеграции вне k8s YAML

| Объект | Порт | Файл |
|---|---:|---|
| CNI → IPAM | **8080** | `configs/cni/cloudbridge-cni.conf` `serviceUrl: ...:8080` |
| AI dashboard URL | **8083** | `configs/ai-config.yaml` |

### Конфликт при смене IPAM на 8083

`configs/config.json` → `p2p_api_https.port` = **8083** (relay HTTPS API). Перенос IPAM на 8083 создал бы коллизию имён в документации и риск путаницы при локальном compose. **8080** для IPAM согласован с CNI и не пересекается с relay listener'ами.

### Host / smarty Docker (не трогаем)

Prometheus на dev-хосте (smarty Docker) слушает **9090 TCP** — вне scope installer; в `port-scheme.md` оставлен комментарий. Relay QUIC использует **9090 UDP** — конфликта с host Prometheus нет.

---

## 3. Вердикт MiMo → решение step 2

| Сервис | Старый SERVICE_CATALOG | Runtime (`cmd` + `k8s/`) | Действие |
|---|---:|---:|---|
| IPAM | 8083 | **8080** | Документация → **8080**; код не менять |
| AI Service | 8085 | **8083** | Документация → **8083**; код не менять |
| AI Worker | 8086 | **8085** | Документация → **8085**; код не менять |

**Одна правда (step 2):** для auxiliary microservices authoritative — **`cmd/*` defaults + `k8s/` ops manifests**; `openwiki/port-scheme.md` и `SERVICE_CATALOG.md` синхронизированы с runtime.

**Не меняли:** `k8s-manifests/06-services.yaml` (stub **8080/9090** для AI) — prod placeholders; ops-дерево `k8s/` актуальнее.

---

## 4. Оставшийся drift (вне scope step 2)

| Item | Статус |
|---|---|
| `k8s-manifests/06-services.yaml` AI **8080/9090** | Stub; исправление numbered manifests — отдельный k8s-consol шаг |
| `k8s/services/ai-service.yaml` metrics port **8084** | Declared в Service; `/metrics` фактически на **8083** в `cmd/ai-service` |
| `config/config.json` (dev profile) | `server.port` 9090 — dev-only, не installer canonical |

---

*Анализ зафиксирован Cursor на шаге 2 миссии cloudbridge-relay-score-ports.*
