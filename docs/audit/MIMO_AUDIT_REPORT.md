# MiMo Audit Report — CloudBridge Relay Installer

**Дата аудита (snapshot):** 2026-07-08  
**Модель:** mimo-v2.5-pro  
**Последнее согласование docs/status:** 2026-07-09 (doc-sync после VOLY-миссий; см. [agent-reports](../agent-reports/))

> **Как читать:**  
> - **Чеклист §«Рекомендуемый план»** и **§«Статус после миссий»** — актуальная правда.  
> - Таблицы findings ниже обновлены: закрытое помечено ✅ / ~~strikethrough~~; open items без галочки.  
> - Исторический snapshot 2026-07-08 сохранён в формулировках «было»; не трактовать закрытые rows как TODO.

**Статус тестов (актуально):** score-ports + k8s-prod-fix (2026-07-09): `go build ./cmd/relay/ ./cmd/ipam-service/ ./cmd/ai-worker/ ./cmd/ai-service/` OK; auxiliary порты IPAM **8080**, AI **8083**, AI Worker **8085** = `openwiki/port-scheme.md` + `cmd/*` + `k8s/` ops + `SERVICE_CATALOG`. `k8s-manifests/05–06` без `nginx:latest`. Cadence-ai: partial test FAIL pre-existing. Go **1.25.12**.

---

## Executive Summary

CloudBridge Relay Installer — Go-монорепозиторий (~175 000 строк Go, 49 internal-пакетов, 13 cmd-бинарников) для P2P/QUIC/WireGuard-ретрансляции.

**На 2026-07-08 (аудит):** мёртвый код, дубли `_improved`/`_enhanced`/`_production`, три auth-пути в конфиге (Zitadel / Keycloak / Django), drift портов в docs vs runtime, nginx-stubs в prod manifests.

**На 2026-07-09 (после VOLY fixes → medium → next → cadence-ai → k8s-consol → score-ports → k8s-prod-fix):**

| Было (критика аудита) | Стало |
|----------------------|--------|
| Keycloak + Django + quantum в коде/скриптах | ✅ Удалены (fixes); IAM canonical = **Zitadel** |
| Порты в 3+ источниках | ✅ SoT = `openwiki/port-scheme.md`; IPAM 8080 / AI 8083 / Worker 8085 |
| nginx stubs в `k8s-manifests/05–06` | ✅ P0 fix (k8s-prod-fix): production images + ports |
| Network score/retry тройки, AI real/optimized | ✅ Canonical production / `ai.Service` |
| Неясный deploy path | ✅ `openwiki/deploy.md` + roles `k8s-manifests/` prod / `k8s/` ops |

**Остаётся:** Cadence dual-path merge; wiring SCORE allocator; full k8s/terraform/monitoring merge; residual port drift (AI metrics 8084, relay prometheus annotation); `docs/legacy/`; часть Keycloak-упоминаний в legacy Go examples.

---

## Результаты тестирования

| Проверка | Статус | Примечание |
|---|---|---|
| `go build ./...` | Частично | Основной код OK; `go build ./...` падает из‑за `rep/` (`.gitignore`, go-rod examples) — Makefile `GO_PKGS` исключает |
| `go test -short ./internal/auth/... ./internal/zitadel/... ./internal/network/... ./config/...` | ✅ OK | auth + zitadel с `-short`; network/config — нет test files |
| `go build ./cmd/relay/ ./cmd/ipam-service/ ./cmd/ai-worker/ ./cmd/ai-service/` | ✅ OK | score-ports / k8s-prod-fix 2026-07-09 |
| `go test -short ./internal/cadence/... ./internal/ai/...` | Частично | activities/workflows/pipeline OK; FAIL client_test, integrations_test, routing real_data (pre-existing) |
| `go test ./...` (полный) | Частично | Без `-short` zitadel может FAIL (live OIDC) |
| `make test-critical` | ✅ | `./internal/zitadel/...` (не keycloak) |
| go.mod Go version | ✅ | **`1.25.12`** |

---

## Лишний / мёртвый код

| Объект | Тип | Статус (2026-07-09) | Приоритет (было) |
|---|---|---|---|
| ~~`cmd/quantum-service/main.go`~~ | Заглушка | ✅ Удалён; остался `DISABLED.md` | Средний |
| ~~`internal/quantum/`~~ | Пакет | ✅ Удалён (fixes) | Средний |
| ~~`internal/metrics/quantum.go`~~ | Метрики | ✅ Удалён | Средний |
| ~~`internal/cadence/workflows/quantum.go`~~ | Workflow | ✅ Удалён | Средний |
| ~~`src/` (cli.ts, index.ts)~~ | TS-SDK | ✅ Удалён | Высокий |
| `pkg/enhanced-quic/` (отдельный go.mod) | Изолированный модуль | ⏳ Open: merge или удалить | Средний |
| ~~`internal/auth/keycloak_*`, `interfaces.go`~~ | Keycloak | ✅ Удалены (fixes) | Высокий |
| ~~`configs/config.json` keycloak/django~~ | Конфиг | ✅ Удалены (prod installer config) | Высокий |
| ~~`config/config.go` Django/Quantum load~~ | Конфиг-код | ✅ Удалены | Высокий |
| `config/config.json` (dev) | Dev-конфиг | ⏳ Может ещё содержать legacy секции | Низкий |
| ~~Keycloak scripts (8+ldap)~~ | Скрипты | ✅ Удалены (fixes) | Средний |
| ~~`.github/workflows/*.disabled`~~ | CI | ✅ Удалены | Низкий |
| Top-level keycloak в k8s configmaps | K8s | ✅ Top-level blocks removed; nested `keycloak_*` в zero_trust — residual | Средний |
| Residual Keycloak *упоминания* в Go | Examples/metrics | ⏳ `internal/quic/enhanced*.go`, metrics names — не prod auth path | Низкий |

---

## Дублирование и legacy

### Дублированные реализации (`_improved`, `_enhanced`, `_production`)

| Пакет | v1 / v2 | Canonical | Статус |
|---|---|---|---|
| `internal/network/` | ~~score_* / retry_config.go~~ | `score_allocator_production.go`, `retry_config_improved.go` | ✅; **не wired** в prod |
| `internal/cadence/` | `service_client.go` / `client.go` | dual-path: relay → improved/production, ai-worker → base | ⏳ merge API отложен (intentional) |
| `internal/nat/` | ~~`turn_server.go`~~ | `turn_server_enhanced.go` | ✅ v1 удалён (medium) |
| `internal/health/` | `health.go`, `health_checker.go`, `comprehensive_checker.go` | — | ⏳ объединить |
| `internal/ai/*` | ~~real_*/optimized_*~~ | `ai.Service` + detector/predictor/optimizer/collectors | ✅ cadence-ai |

### Дублирование инфраструктуры

| Дубликат | Где | Статус |
|---|---|---|
| Terraform | `terraform/` vs `terraform-k8s/` vs `infra/` | ⏳ план в `CONSOLIDATION_PLAN.md` |
| K8s | `k8s/` vs `k8s-manifests/` | ✅ roles split (k8s-consol); P0 stubs fixed (k8s-prod-fix); full merge ⏳ |
| Prometheus / Zabbix / Deploy / Dockerfiles | см. plan | ⏳ |
| Systemd | `quic-proxy.service` (**9090**) vs `enhanced-quic-proxy.service` (**9092**) | ❌ **не идентичны** — **не удалять** enhanced (разные порты) |
| Systemd | `relay.service` vs `cloudbridge-relay.service` | ⏳ оставить один |
| Test dirs | `test/` vs `tests/` | ⏳ |
| Nginx | `nginx/` — 10 конфигов | ⏳ |

### Legacy и архивы

| Объект | Статус |
|---|---|
| `docs/legacy/` (~142) | Архив; не SoT |
| `archive/*` | Доп. архив |
| ~~`github_actions/`~~ | ✅ Удалена (medium step 4) |

---

## Архитектурные нестыковки

### 1. Портовая схема — выровнена (остаточный drift)

| Источник | Main Relay | AI Service | IPAM | WireGuard |
|---|---|---|---|---|
| `openwiki/port-scheme.md` (canonical) | 5550 + 9090 http_api | **8083** | **8080** | 51820 |
| `cmd/*` runtime | — | **8083** | **8080** | — |
| `k8s/` ops | 9090 probes | **8083** | CNI → **8080** | 51820 |
| `docs/operations/SERVICE_CATALOG.md` | 5550 + 9090 | **8083** | **8080** | 51820 |
| `configs/config.json` | 5550 (+ 9090 http_api) | — | — | 51820 |
| `k8s-manifests/06-services.yaml` | 9090 | **8083/8084** | — | 51820 |

**AI Worker:** **8085** (`cmd/ai-worker` + catalog + port-scheme).

**Остаточный drift (не P0):**

| Item | Проблема |
|------|----------|
| AI metrics Service **8084** | Runtime `/metrics` на **8083** |
| Relay `prometheus.io/port` **8084** | Runtime metrics **5551** |
| Ops `k8s/deployments/` relay http **8080** | Legacy vs prod **9090** |
| Dev `config/config.json` `server.port` | **9090** (dev-only) |

**SCORE wiring:** `NewSCORETenantNetworkAllocatorProduction` — 0 prod callers (см. `DUPLICATES_INVENTORY.md`).

### 2. Auth — Zitadel only (post-fixes)

| Провайдер | Статус 2026-07-09 |
|---|---|
| **Zitadel** | ✅ Единственный IAM (`internal/auth/`, `internal/zitadel/`) |
| **Keycloak** | ✅ Prod path удалён; residual strings/examples в quic/metrics |
| **Django** | ✅ Структуры/секции installer config удалены |

Canonical docs: `openwiki/architecture.md` (Keycloak/Django = removed).  
`docs/legacy/` может ещё упоминать Keycloak — игнорировать.

### 3. Пути деплоя — canonical задокументирован

| Путь | Статус |
|---|---|
| `scripts/deploy/deploy.sh` → `k8s-manifests/` | ✅ **Canonical prod** (`openwiki/deploy.md`) |
| `k8s/` | ✅ **Canonical ops** (Zitadel Helm, Cadence, BGP…) |
| `terraform-k8s/` | ✅ Cluster provisioning |
| `terraform/`, `infra/` | ⏳ Дубликаты — plan only |
| `ansible/` | Минимальный |

### 4. Дублирование мониторинга

Без изменений с аудита: 2–4 копии Prometheus/Grafana/AlertManager/Zabbix.  
~~`github_actions/grafana/`~~ — снято с удалением `github_actions/`.  
План: `CONSOLIDATION_PLAN.md`.

### 5. Противоречия в документации — статус

| Проблема (аудит 07-08) | Статус 07-09 |
|---|---|
| openwiki Keycloak «Alternative OIDC» | ✅ Zitadel-only (medium) |
| ARCHITECTURE_ANALYSIS Keycloak-диаграммы | ~ partial cleanup; legacy sections may remain |
| SERVICE_CATALOG relay 8080 vs 9090 | ✅ Catalog sync (next + score-ports) |
| openwiki Django as provider | ✅ marked removed |
| TURN typo "roduction Ready" | ✅ fixed SERVICE_CATALOG (doc-sync) |
| **agent-reports SUMMARY vs later missions** | ✅ doc-sync 2026-07-09 (Superseded blocks) |
| security README still Keycloak | ✅ Zitadel (docs-consistency pass 2) |
| network docs metrics=9090 UDP / health=5551 mixup | ✅ INTERNET_PORTS*, PORTS_SUMMARY, SUBDOMAINS → metrics **5551** TCP |
| ARCHITECTURE_ANALYSIS service map (quantum, cadence 8082) | ✅ aligned port-scheme (pass 2) |

### 6. Go version

`go.mod` → **`1.25.12`** ✅. Доступна линейка Go 1.26.x.

---

## Рекомендуемый план действий

### 1. Quick wins (1-2 дня)

- [x] **Удалить Keycloak-код** — fixes step 1, 2026-07-09
- [x] **Удалить Keycloak-скрипты** (8 + ldap) — fixes step 3
- [x] **Удалить Django-конфиг** installer — fixes step 2 (`config/config.json` dev — residual)
- [x] **Удалить quantum-код** — fixes step 2
- [x] **Удалить `src/`** — fixes step 2
- [ ] **НЕ удалять `systemd/enhanced-quic-proxy.service`** — **не** дубликат: UDP **9092** vs **9090** у `quic-proxy.service` (оставляем оба)
- [x] **Makefile `test-critical`** → `./internal/zitadel/...`
- [x] **go.mod `1.25.12`**
- [x] **Удалить `.disabled` workflows**

### 2. Средний срок (1-2 недели)

- [x] **Инвентаризация дубликатов** → `DUPLICATES_INVENTORY.md`; TURN v1 удалён
- [x] **network score/retry** → production + improved
- [x] **AI real/optimized** → `ai.Service` (cadence-ai)
- [ ] **Cadence dual-path merge API** — intentional dual entrypoint; medium risk
- [~] **K8s first step** — roles + README + archive fixed (k8s-consol); full merge ⏳
- [x] **P0 nginx-stubs `05–06` + deploy.sh load_images** (k8s-prod-fix)
- [ ] **Terraform / monitoring / test dirs / Dockerfiles** — plan only
- [x] **Портовая схема** → `openwiki/port-scheme.md` + catalog + auxiliary
- [x] **Архитектурные docs IAM** → Zitadel-only (medium)
- [x] **Удалить `github_actions/`**
- [x] **`CONSOLIDATION_PLAN.md`**
- [ ] **Очистить / сузить `docs/legacy/`**
- [ ] **Residual AI metrics 8084 / prometheus annotations**
- [ ] **Wire SCORE allocator** в `cmd/relay` / hybrid manager

### 3. Стратегические (1+ месяц)

- [ ] Разбить `cmd/relay/main.go`, `internal/metrics/metrics.go`, `internal/ai/ai.go`
- [x] CI `go vet` / `golangci-lint` — `.github/workflows/lint.yml` (vet на relay legacy test still fails)
- [ ] Интеграционные тесты Zitadel (mock)
- [x] Canonical deploy path — `openwiki/deploy.md`
- [ ] Пересмотр `internal/p2p/`

---

## Статус после миссий (индекс)

| Миссия | Итог | Отчёт |
|--------|------|-------|
| cloudbridge-relay-fixes | Keycloak/Django/quantum/src cleanup | [SUMMARY](../agent-reports/cloudbridge-relay-fixes/SUMMARY.md) |
| cloudbridge-relay-medium | port-scheme, IAM docs, inventory, github_actions, plan | [SUMMARY](../agent-reports/cloudbridge-relay-medium/SUMMARY.md) |
| cloudbridge-relay-next | score/retry delete, catalog, lint.yml, deploy.md | [SUMMARY](../agent-reports/cloudbridge-relay-next/SUMMARY.md) |
| cloudbridge-relay-cadence-ai | AI safe merge; Cadence dual-path documented | [SUMMARY](../agent-reports/cloudbridge-relay-cadence-ai/SUMMARY.md) |
| cloudbridge-relay-k8s-consol | K8s roles; no full merge | [SUMMARY](../agent-reports/cloudbridge-relay-k8s-consol/SUMMARY.md) |
| cloudbridge-relay-score-ports | Auxiliary ports = cmd/* + k8s ops | [SUMMARY](../agent-reports/cloudbridge-relay-score-ports/SUMMARY.md) |
| cloudbridge-relay-k8s-prod-fix | P0 real images/ports in 05–06 + deploy.sh | [SUMMARY](../agent-reports/cloudbridge-relay-k8s-prod-fix/SUMMARY.md) |

**Порядок SoT:** `openwiki/port-scheme.md` / `openwiki/deploy.md` > актуальные SUMMARY (с учётом Superseded) > этот файл > `docs/legacy/`.

---

## Приложение: snapshot 2026-07-08 (исторический)

### Проверка Go (среда аудита)

```bash
$ go version
/bin/bash: go: command not found
```

*(Позже: Go 1.25.12 в `~/.local/go-install/`.)*

### Статистика кодовой базы (на момент аудита)

```
# Go-строк: ~174 902; internal pkgs: 49; cmd: 13; TODO/FIXME files: 39
```

### Структура cmd/ (актуализировано)

| Бинарник | Статус |
|---|---|
| `relay` | Основной |
| `api`, `ai-service`, `ai-analytics`, `ai-worker`, `cadence-service`, `ipam-service`, `cni-plugin` | Active |
| `quic-service` | Экспериментальный |
| `tenant-policy-controller`, `kubectl-zitadel`, `zerotrust-demo` | Utilities / demo |
| `quantum-service` | **DISABLED.md only** (main.go удалён) |

### Auth: баланс ссылок (snapshot 07-08 → post-fixes)

```
# 2026-07-08 audit: Keycloak ~312, Zitadel ~315, Keycloak scripts: 8
# 2026-07-09: prod Keycloak packages/scripts removed; residual string refs in examples/metrics
```

---

*Живой статус — чеклист и §«Статус после миссий». Findings tables синхронизированы 2026-07-09 (doc-consistency pass).*
