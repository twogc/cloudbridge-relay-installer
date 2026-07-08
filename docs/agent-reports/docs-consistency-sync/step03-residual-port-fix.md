# step03 — residual port fix (manifests + README)

**Дата:** 2026-07-09  
**Ветка:** `docs/consistency-sync-reports-ports`

## Цель

Закрыть P0-adjacent residual drift: AI metrics **8084** и relay prometheus **8084** → canonical **8083** / **5551**.

## Изменения

| Файл | Правка |
|------|--------|
| `k8s-manifests/05-deployments.yaml` | `prometheus.io/port: "5551"`; `containerPort` metrics **5551** |
| `k8s-manifests/06-services.yaml` | relay Service +metrics **5551**; AI без 8084 |
| `k8s/services/ai-service.yaml` | убран metrics 8084 (ClusterIP + NodePort) |
| `README.md` | таблица портов → port-scheme |
| `openwiki/port-scheme.md` | K8s exposure + drift tables |
| `k8s-manifests/README.md` | ports table |
| `AGENTS.md`, MIMO, k8s-prod-fix SUMMARY | status updates |

## Проверки

```bash
go build -o /dev/null ./cmd/relay/ ./cmd/ai-service/   # OK
go vet ./cmd/relay/   # EXIT 1: pre-existing context cancel leak main.go:181 (not KeycloakURL)
```

## Не в scope

- Wire SCORE allocator
- Fix main.go context cancel (vet)
- mws OVA 8084
- ops relay deployment containerPort 8080
