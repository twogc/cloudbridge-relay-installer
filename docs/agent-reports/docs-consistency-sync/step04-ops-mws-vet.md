# step04 — ops k8s + mws ports + go vet cancel

**Дата:** 2026-07-09  
**Ветка:** `docs/consistency-sync-reports-ports`

## Цель

Закрыть residual ops/mws metrics drift и pre-existing `go vet` context cancel leak в `NewServer`.

## Изменения

### Go
| Файл | Правка |
|------|--------|
| `cmd/relay/main.go` | `cancel()` before Cadence client/start failure returns |

### Ops K8s
| Файл | Правка |
|------|--------|
| `k8s/deployments/cloudbridge-relay-deployment.yaml` | http **9090**, metrics **5551**, health **9091**, prom annotation **5551** |
| `k8s/deployments/cloudbridge-relay-deployment-cert-manager.yaml` | ports/env/probes → port-scheme |
| `k8s/services/cloudbridge-services.yaml` | metrics **5551**, +grpc **8444** |
| `k8s/services/cloudbridge-relay-lb-tcp.yaml` | metrics **5551**, p2p **5552** |
| `k8s/services/cloudbridge-https-service.yaml` | prom **5551** |
| `k8s/network-policies/cloudbridge-grpc-netpol.yaml` | 5552/5551 |
| `k8s/configmaps/cloudbridge-config{,-minimal,-simple}.yaml` | metrics.port **5551** |

### MWS / monitoring
| Файл | Правка |
|------|--------|
| `mws/cloud-init/cloudbridge-relay-minimal.yaml` | metrics 5551, health 9090, ufw |
| `mws/templates/docker-compose.minimal.yml` | 5551 publish, healthcheck 9090 |
| `monitoring/grafana/datasources.yml` | scrape URL :5551 |

## Проверки

```bash
go build -o /dev/null ./cmd/relay/
go vet ./cmd/relay/   # expect no cancel-leak on Cadence paths
```

## Не в scope

- SCORE allocator wiring
- `cloudbridge-relay-p2p-deployment.yaml` lab ports
- Full Cadence dual-path merge
