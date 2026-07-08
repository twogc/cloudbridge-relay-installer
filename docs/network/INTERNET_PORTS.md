# CloudBridge Relay — порты для интернет-доступа

> **Canonical:** [openwiki/port-scheme.md](../../openwiki/port-scheme.md).  
> Этот файл — firewall / LB / internet exposure; не переопределяет SoT.

## Обзор

| Роль | Порт | Протокол | Примечание |
|------|-----:|----------|------------|
| P2P REST API | **5552** | TCP | Публичный API клиентов |
| P2P QUIC | **5553** | UDP | P2P data plane |
| STUN | **19302** | UDP | NAT discovery |
| QUIC transport | **9090** | UDP | Основной QUIC listener |
| HTTP API | **9090** | TCP | Health/REST via ingress (не metrics) |
| HTTPS API | **5553** | TCP | TLS API (ingress) |
| Prometheus metrics | **5551** | TCP | **Внутренний** scrape target |
| Health / TLS tunnel | **9091** | TCP | Internal |
| MASQUE | **8443** | TCP/UDP | Optional |
| TURN | **3478** | UDP/TCP | Built-in or coturn |
| WireGuard | **51820** | UDP | Mesh (when enabled) |
| gRPC | **8444** | TCP | Control plane (internal) |

**Не путать:**

- **5551 TCP** = relay Prometheus metrics (`metrics.port`)
- **9090 UDP** = QUIC transport
- **9090 TCP** = HTTP API (`http_api.port`), **не** metrics
- Host Prometheus UI часто тоже **9090 TCP** — внешний к installer (см. port-scheme host note)

## Критические (обязательно снаружи)

### 1. P2P REST API — TCP 5552

```
Порт: 5552 (TCP)
Назначение: регистрация peers / P2P API
Доступ: публичный
Безопасность: JWT (Zitadel)
```

### 2. P2P QUIC — UDP 5553

```
Порт: 5553 (UDP)
Назначение: P2P data plane
Доступ: публичный
Протокол: QUIC over UDP
```

### 3. STUN — UDP 19302

```
Порт: 19302 (UDP)
Назначение: NAT traversal
Доступ: публичный
```

## Рекомендуемые

### 4. QUIC transport — UDP 9090

```
Порт: 9090 (UDP)
Назначение: основной QUIC transport relay
```

### 5. MASQUE — TCP 8443 (optional)

```
Порт: 8443 (TCP/UDP)
Назначение: HTTP/3 MASQUE
```

### 6. TURN — UDP/TCP 3478 (optional)

```
Порт: 3478
Назначение: relay для restrictive NAT
```

### 7. WireGuard — UDP 51820 (optional)

```
Порт: 51820 (UDP)
Назначение: WireGuard mesh
```

## Внутренние (не открывать в интернет)

| Порт | Протокол | Назначение |
|-----:|----------|------------|
| 5550 | TCP | Main TCP server |
| **5551** | TCP | **Prometheus metrics** |
| 9091 | TCP | Health / TLS tunnel |
| 8444 | TCP | gRPC |
| 9090 | TCP | HTTP API — только через ingress / private LB |

## Конфигурация firewall

### Минимальная (P2P)

```bash
sudo ufw allow 5552/tcp comment "CloudBridge P2P API"
sudo ufw allow 5553/udp comment "CloudBridge P2P QUIC"
sudo ufw allow 19302/udp comment "CloudBridge STUN"
```

### Полная публичная поверхность

```yaml
ports:
  - port: 5552    # P2P REST API
    protocol: TCP
    public: true
  - port: 5553    # P2P QUIC
    protocol: UDP
    public: true
  - port: 19302   # STUN
    protocol: UDP
    public: true
  - port: 9090    # QUIC transport
    protocol: UDP
    public: true
  - port: 8443    # MASQUE (optional)
    protocol: TCP
    public: true
  - port: 3478    # TURN (optional)
    protocol: UDP
    public: true
  - port: 51820   # WireGuard (optional)
    protocol: UDP
    public: true
  # Internal only:
  # 5551/TCP metrics — scrape from Prometheus network
```

### UFW (рекомендуемые + metrics private)

```bash
# Public
sudo ufw allow 5552/tcp comment "CloudBridge P2P API"
sudo ufw allow 5553/udp comment "CloudBridge P2P QUIC"
sudo ufw allow 19302/udp comment "CloudBridge STUN"
sudo ufw allow 9090/udp comment "CloudBridge QUIC Transport"
sudo ufw allow 8443/tcp comment "CloudBridge MASQUE"
sudo ufw allow 51820/udp comment "CloudBridge WireGuard"

# Metrics: only from monitoring CIDR (example)
# sudo ufw allow from 10.0.0.0/8 to any port 5551 proto tcp comment "CloudBridge metrics"
```

## Load balancer / Nginx

Проксируйте **P2P API (5552)** и при необходимости **HTTP API (9090)**; не публикуйте **5551** наружу без auth/network policy.

```nginx
upstream cloudbridge_p2p_api {
    server RELAY_IP:5552;
}

server {
    listen 443 ssl;
    server_name edge.example.com;

    location /api/ {
        proxy_pass http://cloudbridge_p2p_api;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## Проверка

```bash
sudo ufw status
ss -tulpn | grep -E '(5551|5552|5553|8443|9090|19302|51820|3478)'
curl -sS "http://127.0.0.1:5551/metrics" | head
```

## См. также

- [PORTS_SUMMARY.md](PORTS_SUMMARY.md) — Minikube NodePort + firewall checklist
- [INTERNET_PORTS_FULL.md](INTERNET_PORTS_FULL.md) — expanded tables
- [docs/operations/SERVICE_CATALOG.md](../operations/SERVICE_CATALOG.md)
