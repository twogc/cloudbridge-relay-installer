# k8s-manifests — Canonical Production Deploy

Нумерованные манифесты (`00–15`) для упорядоченного развёртывания CloudBridge Relay в Kubernetes.

## Canonical path

| Слой | Путь | Кто применяет |
|------|------|---------------|
| **Production deploy** | `k8s-manifests/00–15` | [`scripts/deploy/deploy.sh`](../scripts/deploy/deploy.sh) |
| Документация | [openwiki/deploy.md](../openwiki/deploy.md) | — |

## Порядок apply

```bash
# Автоматически (рекомендуется)
./scripts/deploy/deploy.sh

# Вручную — core stack (00–08)
kubectl apply -f k8s-manifests/00-namespace.yaml
kubectl apply -f k8s-manifests/01-pvc.yaml
# … 02–08

# Zitadel (09–15) — после core или через zitadel-setup
kubectl apply -f k8s-manifests/09-zitadel.yaml
# … 10–15
```

## Связь с `k8s/`

- **`k8s-manifests/`** — ordered prod deploy (relay, AI, Cadence, ingress, Zitadel wrappers).
- **`k8s/`** — domain/ops manifests (Helm values, schema jobs, BGP, Kyverno, cert-manager) — **не** заменяет нумерованный стек.

Подробнее: [CONSOLIDATION_PLAN §1](../docs/audit/CONSOLIDATION_PLAN.md), [openwiki/deploy.md](../openwiki/deploy.md).

## Образы и порты (`05–06`, fixed 2026-07-09)

| Компонент | Image | Источник | Порты (deployment / service) |
|-----------|-------|----------|------------------------------|
| cloudbridge-relay | `docker-registry:5000/cloudbridge-relay:latest` | in-cluster registry (`deploy.sh` build + push) | 9090 TCP/UDP, 51820 UDP, 8444 TCP, 9091 TCP |
| ai-service | `docker-registry:5000/ai-service:latest` | in-cluster registry (`docker/dockerfiles/Dockerfile.ai-service`) | 8083 HTTP, 8084 metrics (Service) |
| cadence | `ubercadence/server:v0.22.4` | Docker Hub (push в локальный registry **не** нужен) | 7933–7935, 7939 (tcpSocket probe на 7933) |

### Стратегия pull (step 3, 2026-07-09)

1. **`deploy.sh` → `load_images()`** собирает relay/ai локально, делает `kubectl port-forward` на `docker-registry:5000`, затем `docker tag/push localhost:5000/...`.
2. **Поды в кластере** тянут relay/ai как `docker-registry:5000/<name>:latest` (Service `docker-registry` в namespace `cloudbridge`) с `imagePullPolicy: IfNotPresent` и секретом `k8s-registry-creds`.
3. **Cadence** — публичный образ `ubercadence/server:v0.22.4` с Docker Hub; `load_images()` его не собирает и не пушит.

Согласовано с [`openwiki/port-scheme.md`](../openwiki/port-scheme.md) и [`scripts/deploy/deploy.sh`](../scripts/deploy/deploy.sh).

## Связь с `k8s/` (ops)

- **`k8s-manifests/`** — ordered prod deploy (relay, AI, Cadence, ingress, Zitadel wrappers).
- **`k8s/`** — domain/ops manifests (Helm values, schema jobs, BGP, Kyverno, cert-manager) — **не** заменяет нумерованный стек.

Оставшийся drift (relay prometheus annotation **8084**, ops-only sidecars) — см. [openwiki/port-scheme.md §k8s drift](../openwiki/port-scheme.md).
