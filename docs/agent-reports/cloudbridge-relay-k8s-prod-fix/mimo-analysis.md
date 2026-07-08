# MiMo Analysis — k8s-manifests prod fix

**Дата:** 2026-07-09  
**Источник:** `docs/audit/MIMO_AUDIT_REPORT.md` (mimo-v2.5-pro, 2026-07-08) + `openwiki/port-scheme.md`  
**Контекст:** cloudbridge-relay-k8s-prod-fix, step 2

---

## Executive Summary (MiMo)

CloudBridge Relay Installer содержит **два независимых дерева K8s-манифестов**:

| Дерево | Роль | Статус до fix |
|--------|------|---------------|
| `k8s-manifests/00–15` | Canonical **production deploy** (`scripts/deploy/deploy.sh`) | `05-deployments.yaml` — **nginx:latest stubs**, порты 80 |
| `k8s/deployments/`, `k8s/services/` | Ops/component manifests | Актуальные образы и порты |

MiMo зафиксировал **критический drift** в numbered prod manifests: образы-заглушки и неверные порты AI Service (**8080/9090** vs canonical **8083/8084**).

---

## Портовая схема (canonical)

Источник истины: `openwiki/port-scheme.md`

| Сервис | Canonical port(s) | Источник |
|--------|-------------------|----------|
| Relay HTTP API | **9090** TCP | `configs/config.json` `http_api.port` |
| Relay QUIC | **9090** UDP | `cmd/relay/main.go` |
| Relay WireGuard | **51820** UDP | config |
| Relay gRPC | **8444** TCP | `grpc.port` |
| Relay health/TLS tunnel | **9091** TCP | `cmd/relay/https.go` |
| AI Service HTTP + `/metrics` | **8083** TCP | `cmd/ai-service`, `AI_PORT` |
| AI Service metrics (Service decl.) | **8084** TCP | `k8s/services/ai-service.yaml` |
| Cadence frontend | **7933** TCP | SERVICE_CATALOG, k8s-manifests |
| Cadence matching/membership | **7934** TCP | k8s |
| Cadence history/frontend svc | **7935** TCP | k8s-manifests |
| Cadence worker | **7939** TCP | k8s-manifests |

---

## Drift: `k8s-manifests/05–06` (до step 2)

### CloudBridge Relay

| Поле | `k8s-manifests/05` (stub) | `k8s/deployments` | Canonical |
|------|----------------------------|-------------------|-----------|
| Image | `nginx:latest` | `cloudbridge-relay:mvp-*` | `cloudbridge-relay:latest` (deploy.sh) |
| containerPort | **80** | 8080 (legacy), 9090 UDP, … | **9090** TCP, 9090 UDP, 51820, 8444, 9091 |
| Health probe | 9090 `/health` ✓ | 9090 `/health` ✓ | 9090 |

**Вывод MiMo:** Service в `06-services.yaml` уже был корректен (9090, 51820, 8444); deployment — placeholder.

### AI Service

| Поле | `k8s-manifests/05–06` (stub) | `k8s/` ops | Canonical |
|------|------------------------------|------------|-----------|
| Image | `nginx:latest` | `ai-service:latest` | `ai-service:latest` |
| containerPort | **80** | **8083** | **8083** |
| Health probe | **8080** | **8083** | **8083** |
| Service ports | **8080** / metrics **9090** | **8083** / **8084** | **8083** / **8084** |
| prometheus.io/port | **9090** | **8083** | **8083** (runtime `/metrics` on 8083) |

**Вывод MiMo:** Полный stub drift; score-ports step 2–4 выровнял docs + `k8s/`, но numbered YAML не обновлялись.

### Cadence

| Поле | `k8s-manifests/05` (stub) | `k8s/cadence/cadence.yaml` |
|------|---------------------------|----------------------------|
| Image | `nginx:latest` | `ubercadence/server:v0.22.4` |
| Ports | 80 | 7933, 7934, 7935, 7939 |
| Probe | httpGet `/health` :7933 | **tcpSocket** :7933 |

**Вывод MiMo:** Cadence не экспонирует HTTP `/health` на frontend; tcpSocket на 7933 — правильный probe (как в `k8s/cadence/cadence.yaml`).

---

## Рекомендации MiMo (применены в step 2)

1. **`05-deployments.yaml`:** заменить nginx stubs на production images (`cloudbridge-relay:latest`, `ai-service:latest`, `ubercadence/server:v0.22.4`); объявить canonical containerPorts; исправить probes.
2. **`06-services.yaml`:** AI Service **8083/8084** (убрать 8080/9090 drift).
3. **`k8s-manifests/README.md`:** убрать упоминание «stub nginx», зафиксировать fixed date.
4. **Не трогать** дерево `k8s/` целиком — minimal diff, ops manifests остаются reference.

---

## Остаточный drift (вне scope step 2)

| Item | Issue |
|------|-------|
| Relay prometheus annotation | `8084` в deployment vs runtime metrics **5551** |
| AI metrics containerPort | Runtime serves `/metrics` on **8083**; Service declares **8084** (k8s/ pattern) |
| Cadence env/DB | `05-deployments.yaml` — минимальный env; полная конфигурация в `k8s/cadence/` |
| Full k8s merge | Отложено (CONSOLIDATION_PLAN §1) |

---

## Команды верификации MiMo (из аудита)

```bash
go build ./cmd/relay/ ./cmd/ipam-service/ ./cmd/ai-worker/ ./cmd/ai-service/  # OK (score-ports)
# Auxiliary: IPAM 8080, AI 8083, AI Worker 8085 — совпадают с port-scheme + cmd/*
```

---

*Извлечено из MiMo Audit Report для миссии k8s-prod-fix. Полный отчёт: `docs/audit/MIMO_AUDIT_REPORT.md`.*
