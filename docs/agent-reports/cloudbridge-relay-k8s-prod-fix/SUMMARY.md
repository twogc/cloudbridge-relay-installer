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
| `rg 'nginx:latest' k8s-manifests/` | 0 | ✅ 0 |
| `rg 'containerPort: 80$' k8s-manifests/05-deployments.yaml` | 0 | ✅ 0 |
| `rg 'port: 8080' k8s-manifests/06-services.yaml` (ai-service) | 0 | ✅ 0 |
| `rg 'alpine:3.18' scripts/deploy/deploy.sh` | 0 | ✅ 0 |
| YAML parse `05–06` | OK | ✅ python3 yaml.safe_load_all |
| `kubectl` dry-run | optional | ⏭ kubectl not in PATH |

---

## Остаточный drift (не P0)

| Элемент | Описание |
|---------|----------|
| AI metrics Service **8084** | Runtime `/metrics` на **8083**; `containerPort: 8084` не объявлен в deployment |
| Relay prometheus.io/port **8084** | Annotation drift vs runtime metrics **5551** |
| Ops `k8s/deployments/` relay http **8080** | Legacy `containerPort` в ops-дереве; `k8s-manifests/05` использует **9090** |
| Cadence runtime env | Минимальный env в `05`; полная конфигурация в `k8s/cadence/` |
| Cadence worker port **7939** | Есть в `k8s-manifests/06`, нет в `k8s/services/cloudbridge-services.yaml` |

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
