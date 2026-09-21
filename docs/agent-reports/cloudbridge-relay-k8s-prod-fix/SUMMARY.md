# SUMMARY — cloudbridge-relay-k8s-prod-fix

**Дата:** 2026-07-09
**Миссия:** Устранить P0 production drift в `k8s-manifests/05–06` и `scripts/deploy/deploy.sh` (nginx-stubs → реальные образы/порты).

---

## Проблема

Canonical prod deploy (`scripts/deploy/deploy.sh` → `k8s-manifests/00–08`) применял **nginx:latest stubs** в `05-deployments.yaml` и неверные порты AI Service (**8080/9090**) в `06-services.yaml`. `deploy.sh` → `load_images()` собирал alpine-echo заглушки вместо реальных бинарников relay/ai.

MiMo audit: [MIMO_AUDIT_REPORT.md §1.3](../../audit/MIMO_AUDIT_REPORT.md), drift table в [port-scheme.md §k8s drift](../../../openwiki/port-scheme.md).

---

## Что сделано (steps 2–4)

| Шаг | Содержание | Ключевые файлы |
|-----|------------|----------------|
| **2** | Замена stubs в `05–06`: production images/ports/probes | `k8s-manifests/05-deployments.yaml`, `06-services.yaml`, `README.md` |
| **3** | `load_images()` — real ai build, cadence с Docker Hub, registry push flow | `scripts/deploy/deploy.sh`, `05-deployments.yaml` (image paths) |
| **4** | Документация, проверки, SUMMARY | `openwiki/port-scheme.md`, `deploy.md`, `CONSOLIDATION_PLAN.md`, `MIMO_AUDIT_REPORT.md` |

### Итоговые образы и порты (`05–06`)

| Компонент | Image | Порты |
|-----------|-------|-------|
| cloudbridge-relay | `docker-registry:5000/cloudbridge-relay:latest` | 9090 TCP/UDP, 51820 UDP, 8444 TCP, 9091 TCP |
| ai-service | `docker-registry:5000/ai-service:latest` | 8083 HTTP; Service metrics 8084 |
| cadence | `ubercadence/server:v0.22.4` | 7933–7935, 7939 (tcpSocket probe 7933) |

---

## Проверки (step 4)

| Команда | Ожидание | Результат |
|---------|----------|-----------|
| `rg 'nginx:latest' k8s-manifests/` | 0 | 0 |
| `rg 'containerPort: 80$' k8s-manifests/05-deployments.yaml` | 0 | 0 |
| `rg 'port: 8080' k8s-manifests/06-services.yaml` (ai-service) | 0 | 0 |
| `rg 'alpine:3.18' scripts/deploy/deploy.sh` | 0 | 0 |
| YAML parse `05–06` | OK | python3 yaml.safe_load_all |
| `kubectl` dry-run | optional | ⏭ kubectl not in PATH |

---

## Остаточный drift (на момент k8s-prod-fix)

| Элемент | Описание |
|---------|----------|
| AI metrics Service **8084** | Runtime `/metrics` на **8083** |
| Relay prometheus.io/port **8084** | vs runtime metrics **5551** |
| Ops `k8s/deployments/` relay http **8080** | Legacy vs prod **9090** |
| Cadence runtime env | Минимальный env в `05`; full в `k8s/cadence/` |
| Cadence worker port **7939** | Есть в `06`, нет в части ops services |

### Superseded (docs-consistency residual-port-fix, 2026-07-09)

| Item | Статус |
|------|--------|
| AI Service metrics **8084** in `k8s-manifests/06` + `k8s/services/ai-service.yaml` | removed; scrape **8083** |
| Relay prometheus annotation **8084** | **5551** + containerPort/Service metrics |
| Ops relay http **8080** / mws 8084 | **ещё open** (ops/OVA path) |

---

## Отчёты по шагам

- [mimo-analysis.md](mimo-analysis.md) — выжимка MiMo по drift
- [step02-write-mimo-fix-05-06-manifests.md](step02-write-mimo-fix-05-06-manifests.md)
- [step03-fix-deploy-sh-load-images.md](step03-fix-deploy-sh-load-images.md)
- [step04-docs-verify-summary.md](step04-docs-verify-summary.md)

---

## Открытые пункты

- End-to-end `deploy.sh` push в in-cluster registry не проверен (нет кластера / `kubectl`).
- Cadence pod может не стартовать без env из `k8s/cadence/` (Cassandra/Ringpop).
- Полный merge `k8s/` ↔ `k8s-manifests/` — отложен ([CONSOLIDATION_PLAN §1](../../audit/CONSOLIDATION_PLAN.md)).

---

## Commit/push

Не выполнялся (по инструкции пользователя).
