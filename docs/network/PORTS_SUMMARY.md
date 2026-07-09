# CloudBridge Relay — сводка по портам

> **Canonical source:** [openwiki/port-scheme.md](../../openwiki/port-scheme.md)  
> Этот файл — Minikube NodePort examples и firewall checklist. При расхождении верить port-scheme.

## Relay listeners (canonical)

| Listener | Port | Protocol |
|----------|-----:|----------|
| Main TCP | 5550 | TCP |
| Prometheus metrics | **5551** | TCP |
| P2P REST API | 5552 | TCP |
| HTTP API | 9090 | TCP |
| HTTPS API | 5553 | TCP |
| Health / TLS tunnel | 9091 | TCP |
| QUIC transport | 9090 | UDP |
| Enhanced QUIC | 9092 | UDP |
| P2P QUIC | 5553 | UDP |
| STUN | 19302 | UDP |
| TURN | 3478 | UDP/TCP |
| MASQUE | 8443 | TCP/UDP |
| gRPC | 8444 | TCP |
| WireGuard | 51820 | UDP |

**Auxiliary:** IPAM **8080**, AI Service **8083**, AI Worker **8085** (см. port-scheme).

## Пример: Minikube NodePort (исторический lab)

Значения NodePort зависят от Service; ниже — типичный lab mapping, **не** prod contract:

```
30080/TCP  - Main Server (5550)
30081/TCP  - Metrics (5551)
32500/TCP  - P2P API (5552)   # primary public API
30087/TCP  - HTTPS API (5553) # was mislabeled as "HTTP API" in older notes
30090/UDP  - QUIC Transport (9090)
30443/TCP  - MASQUE Proxy (8443)
30092/UDP  - Enhanced QUIC (9092)
30844/TCP  - gRPC (8444)
```

HTTP API (**9090** TCP) и metrics (**5551**) — разные listeners.

## Интернет: открыть

### Критические

```
5552/TCP   - P2P API
5553/UDP   - P2P QUIC
19302/UDP  - STUN
```

### Рекомендуемые

```
9090/UDP   - QUIC transport
8443/TCP   - MASQUE (optional)
3478/UDP   - TURN (optional)
51820/UDP  - WireGuard (optional)
```

### Внутренние (не наружу)

```
5550/TCP   - Main server
5551/TCP   - Prometheus metrics
9090/TCP   - HTTP API (ingress only)
9091/TCP   - Health tunnel
8444/TCP   - gRPC
9092/UDP   - Enhanced QUIC (optional)
```

## Firewall (prod)

```bash
sudo ufw allow 5552/tcp comment "CloudBridge P2P API"
sudo ufw allow 5553/udp comment "CloudBridge P2P QUIC"
sudo ufw allow 19302/udp comment "CloudBridge STUN"
sudo ufw allow 9090/udp comment "CloudBridge QUIC Transport"
# Metrics — only monitoring CIDR:
# sudo ufw allow from 10.0.0.0/8 to any port 5551 proto tcp comment "CloudBridge metrics"
```

## Nginx (edge → P2P API)

```nginx
upstream cloudbridge_api {
    server RELAY_IP:5552;
}

server {
    listen 443 ssl;
    server_name edge.example.com;

    location /api/ {
        proxy_pass http://cloudbridge_api;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location /health {
        proxy_pass http://cloudbridge_api;
    }
}
```

## Kubernetes LoadBalancer (public surface)

```yaml
apiVersion: v1
kind: Service
metadata:
  name: cloudbridge-relay-external
spec:
  type: LoadBalancer
  ports:
  - name: p2p-api
    port: 5552
    targetPort: 5552
    protocol: TCP
  - name: p2p-quic
    port: 5553
    targetPort: 5553
    protocol: UDP
  - name: stun
    port: 19302
    targetPort: 19302
    protocol: UDP
  - name: quic
    port: 9090
    targetPort: 9090
    protocol: UDP
  selector:
    app: cloudbridge-relay
```

Prod numbered stack: `k8s-manifests/06-services.yaml` (9090 TCP/UDP, 51820, 8444) — см. [openwiki/deploy.md](../../openwiki/deploy.md).

## Тестирование

```bash
curl -sS "http://127.0.0.1:5551/metrics" | head
curl -sS "http://127.0.0.1:5552/health" || true
# External
curl -X GET "https://edge.example.com/health"
```

## См. также

- [INTERNET_PORTS.md](INTERNET_PORTS.md)
- [INTERNET_PORTS_FULL.md](INTERNET_PORTS_FULL.md)
- [SERVICE_CATALOG.md](../operations/SERVICE_CATALOG.md)
