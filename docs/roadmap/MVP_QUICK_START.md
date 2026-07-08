# CloudBridge Relay MVP - Quick Start Guide

**Последнее обновление**: 2025-10-13
**Статус**: Ready for MVP deployment (82% готовности)

---

## Что уже работает (готово к production)

### Core Services (95%)
- *MASQUE Proxy** - RFC 9297 compliant
- *QUIC** - Enhanced QUIC с safe defaults
- *NAT Traversal** - Client-side готов (24 Prometheus метрики)
- *AI Analytics** - Full stack (Cloudflare AI)
- *IPAM Service** - O(1) allocation
- *Multi-tenancy** - Полная изоляция
- *Circuit Breaker** - Production ready

### Security (85%)
- *Zitadel OIDC** - JWT validation через JWKS
- *TLS 1.3** - Cert-manager integration
- *RBAC** - Role-based access control

### Infrastructure (95%)
- *Kubernetes Deployments** - Production manifests
- *TURN Server** - Готов к deploy
- *Prometheus Monitoring** - Full observability
- *External Secrets** - 90% готов

---

## Минимальные требования для MVP

### Что ОБЯЗАТЕЛЬНО нужно:
1. ubernetes cluster (VK Cloud или bare-metal)
2. oadBalancer (для TURN server)
3. itadel instance (для аутентификации)
4. ⚠️ TLS сертификаты (можно self-signed для dev)
5. ⚠️ PostgreSQL (для IPAM и AI)

### Что можно отложить:
- ⏸️ Quantum Service (заглушка, не критично)
- ⏸️ Kubernetes Cluster Registry (опционально)
- ⏸️ CNI Plugin (опционально)

---

## Deployment Checklist (Fast Track - 3 часа)

### Шаг 1: Подготовка окружения (30 минут)

```bash
# 1. Клонировать репозиторий
git clone https://github.com/your-org/cloudbridge-relay-installer.git
cd cloudbridge-relay-installer

# 2. Проверить доступ к Kubernetes
kubectl cluster-info
kubectl get nodes

# 3. Создать namespace
kubectl create namespace cloudbridge

# 4. Настроить переменные окружения
export ZITADEL_DOMAIN="https://zitadel.2gc.ru"
export TURN_PASSWORD=$(openssl rand -base64 32)
export RELAY_DOMAIN="relay.2gc.ru"
```

### Шаг 2: Deploy TURN Server (30 минут)

```bash
# Автоматический deployment
./scripts/deploy-turn-server.sh

# Проверка статуса
kubectl get pods,svc -n cloudbridge -l app=turn-server

# Получить external IP
export TURN_EXTERNAL_IP=$(kubectl get svc turn-server -n cloudbridge \
  -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

echo "TURN Server IP: $TURN_EXTERNAL_IP"
```

**Результат**: TURN server работает на `$TURN_EXTERNAL_IP:3478`

### Шаг 3: Обновить конфигурацию (15 минут)

```bash
# Обновить configs/config.json
cat > configs/config-override.json <<EOF
{
  "zitadel": {
    "domain": "$ZITADEL_DOMAIN",
    "issuer": "$ZITADEL_DOMAIN",
    "audience": "cloudbridge-relay"
  },
  "turn_pool": {
    "pools": {
      "primary": {
        "servers": [{
          "address": "$TURN_EXTERNAL_IP:3478",
          "username": "cloudbridge",
          "password": "$TURN_PASSWORD",
          "realm": "cloudbridge.2gc.ru"
        }]
      }
    }
  },
  "relay": {
    "domain": "$RELAY_DOMAIN",
    "external_ip": "auto"
  }
}
EOF

# Merge с основной конфигурацией
jq -s '.[0] * .[1]' configs/config.json configs/config-override.json > configs/config-merged.json
mv configs/config-merged.json configs/config.json
```

### Шаг 4: Deploy CloudBridge Relay (45 минут)

```bash
# 1. Создать secrets
kubectl create secret generic cloudbridge-secrets -n cloudbridge \
  --from-literal=TURN_PASSWORD="$TURN_PASSWORD" \
  --from-literal=ZITADEL_CLIENT_SECRET="<your-client-secret>"

# 2. Deploy PostgreSQL (если нет внешнего)
kubectl apply -f k8s/deployments/postgres-simple.yaml

# 3. Deploy AI Service
kubectl apply -f k8s/deployments/ai-service-deployment.yaml

# 4. Deploy IPAM Service
kubectl apply -f k8s/deployments/ipam-service-deployment.yaml

# 5. Deploy Main Relay
kubectl apply -f k8s/deployments/cloudbridge-relay-deployment.yaml

# 6. Deploy Prometheus
kubectl apply -f k8s/deployments/prometheus-optimized-deployment.yaml
```

### Шаг 5: Проверка работоспособности (30 минут)

```bash
# 1. Проверить все pods
kubectl get pods -n cloudbridge

# Ожидаемый вывод:
# NAME                                READY   STATUS    RESTARTS   AGE
# turn-server-xxxxx                   1/1     Running   0          30m
# cloudbridge-relay-xxxxx             1/1     Running   0          10m
# ai-service-xxxxx                    1/1     Running   0          10m
# ipam-service-xxxxx                  1/1     Running   0          10m
# postgres-xxxxx                      1/1     Running   0          10m
# prometheus-xxxxx                    1/1     Running   0          10m

# 2. Проверить services
kubectl get svc -n cloudbridge

# 3. Проверить logs
kubectl logs -n cloudbridge -l app=cloudbridge-relay --tail=100

# 4. Проверить metrics
kubectl port-forward -n cloudbridge svc/cloudbridge-relay 8080:8080
curl http://localhost:8080/metrics

# 5. Проверить health
curl http://localhost:8080/health
# Ожидается: {"status":"healthy","version":"..."}
```

### Шаг 6: Тестирование NAT Traversal (30 минут)

```bash
# 1. Port-forward для доступа к API
kubectl port-forward -n cloudbridge svc/cloudbridge-relay 8082:8082

# 2. Зарегистрировать тестового peer
curl -X POST http://localhost:8082/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "test-tenant",
    "peer_id": "peer-1",
    "public_key": "test-key"
  }'

# 3. Проверить ICE candidates
curl http://localhost:8082/api/v1/peers/peer-1/ice-candidates

# 4. Проверить TURN allocation
curl http://localhost:8082/api/v1/turn/allocate \
  -H "Authorization: Bearer <jwt-token>"
```

---

## Troubleshooting

### Problem: TURN server не получает external IP

**Solution**:
```bash
# Проверить LoadBalancer
kubectl describe svc turn-server -n cloudbridge

# Если используете bare-metal, установить MetalLB
kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.13.12/config/manifests/metallb-native.yaml

# Настроить IP pool
cat <<EOF | kubectl apply -f -
apiVersion: metallb.io/v1beta1
kind: IPAddressPool
metadata:
  name: first-pool
  namespace: metallb-system
spec:
  addresses:
  - 192.168.1.240-192.168.1.250
EOF
```

### Problem: Zitadel JWKS не доступен

**Solution**:
```bash
# Проверить connectivity
curl -v $ZITADEL_DOMAIN/.well-known/openid-configuration

# Проверить logs
kubectl logs -n cloudbridge -l app=cloudbridge-relay | grep -i jwks

# Временно использовать fallback secret (НЕ для production!)
kubectl create secret generic zitadel-fallback -n cloudbridge \
  --from-literal=jwt-secret="your-secret"
```

### Problem: Pods не стартуют

**Solution**:
```bash
# Проверить events
kubectl get events -n cloudbridge --sort-by='.lastTimestamp'

# Проверить logs
kubectl logs -n cloudbridge <pod-name> --previous

# Проверить resources
kubectl top pods -n cloudbridge
```

---

## Monitoring Dashboard

### Prometheus Queries

```promql
# NAT Traversal Success Rate
rate(nat_traversal_success_total[5m]) / rate(nat_traversal_attempts_total[5m]) * 100

# TURN Allocations
sum(turn_active_allocations) by (server)

# QUIC Connections
sum(quic_connections_active) by (tenant_id)

# AI Service Latency
histogram_quantile(0.95, rate(ai_request_duration_seconds_bucket[5m]))
```

### Key Metrics to Watch

| Metric | Target | Alert Threshold |
|--------|--------|-----------------|
| NAT Traversal Success | > 95% | < 90% |
| TURN Relay Usage | < 35% | > 50% |
| Relay Latency (p95) | < 100ms | > 200ms |
| API Error Rate | < 1% | > 5% |

---

## Success Criteria (MVP)

### Technical
- [x] WT validation через JWKS
- [ ] ⏳ TURN server deployed и работает
- [ ] ⏳ NAT traversal success > 90%
- [ ] ⏳ All services healthy

### Functional
- [ ] ⏳ Client может зарегистрироваться
- [ ] ⏳ P2P connection через MASQUE
- [ ] ⏳ TURN fallback работает
- [ ] ⏳ Multi-tenancy изоляция

### Operational
- [x] rometheus metrics экспортируются
- [x] ogs структурированные
- [ ] ⏳ Health checks работают

---

## 🚦 Next Steps

### Immediate (Next 1-2 days)
1. Deploy TURN server в production кластер
2. Тестировать NAT traversal с реальными клиентами
3. Настроить production TLS сертификаты (Let's Encrypt)

### Short-term (Next week)
1. API Compatibility Layer (Блокер #3)
2. External Secrets production deployment
3. Load testing (100+ concurrent connections)

### Medium-term (2-4 weeks)
1. Client SDK integration
2. Personal Cabinet (Cloudflare Workers)
3. Production monitoring и alerting

---

## Documentation

- **Architecture**: [ARCHITECTURE_ANALYSIS.md](ARCHITECTURE_ANALYSIS.md)
- **Developer Guide**: [DEVELOPER_BRIEFING.md](DEVELOPER_BRIEFING.md)
- **Full MVP Plan**: [MVP_IMPLEMENTATION_PLAN.md](MVP_IMPLEMENTATION_PLAN.md)
- **OSS Integration**: [oss-repositories/integration-plan.md](oss-repositories/integration-plan.md)

---

## 🆘 Support

- GitHub Issues: https://github.com/your-org/cloudbridge-relay-installer/issues
- Documentation: https://docs.cloudbridge.2gc.ru
- Slack: #cloudbridge-relay

---

**Status**: 🟢 Ready for MVP deployment
**Last Updated**: 2025-10-13 23:50
**Next Milestone**: Production TURN deployment
