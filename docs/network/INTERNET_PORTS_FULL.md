# CloudBridge Relay — порты для интернет-доступа (полный)

> **Canonical:** [openwiki/port-scheme.md](../../openwiki/port-scheme.md).  
> Краткая версия: [INTERNET_PORTS.md](INTERNET_PORTS.md).

## Критические (обязательно открыть)

### 1. P2P API — TCP 5552

- **Назначение:** регистрация клиентов, P2P REST API
- **Безопасность:** высокая (JWT / Zitadel)
- **Пример:** `ufw allow 5552/tcp`

### 2. P2P QUIC — UDP 5553

- **Назначение:** P2P data plane
- **Безопасность:** средняя (QUIC crypto)
- **Пример:** `ufw allow 5553/udp`

### 3. STUN — UDP 19302

- **Назначение:** NAT traversal
- **Пример:** `ufw allow 19302/udp`

## Рекомендуемые

### 4. QUIC transport — UDP 9090

- **Назначение:** основной QUIC listener relay
- **Пример:** `ufw allow 9090/udp`
- **Не путать с** metrics (5551 TCP) и HTTP API (9090 TCP)

### 5. MASQUE HTTP/3 — TCP 8443

- **Назначение:** MASQUE proxy
- **Пример:** `ufw allow 8443/tcp`

### 6. TURN — UDP/TCP 3478

- **Назначение:** TURN (built-in или coturn pool)
- **Пример:** `ufw allow 3478/udp`

### 7. WireGuard — UDP 51820

- **Назначение:** WireGuard mesh (`wireguard.interface.listen_port`)
- **Пример:** `ufw allow 51820/udp`

## Внутренние / ограниченные

### 8. Prometheus metrics — TCP 5551

- **Назначение:** `/metrics` на relay (`metrics.port`)
- **Протокол:** **TCP** (не UDP)
- **Доступ:** только monitoring network / scrape
- **Пример:** `ufw allow from 10.0.0.0/8 to any port 5551 proto tcp`

### 9. HTTP API — TCP 9090

- **Назначение:** `http_api.port` (REST/health via ingress)
- **Не** публиковать «как metrics»

### 10. Health / TLS tunnel — TCP 9091

- **Назначение:** internal health / TLS tunnel
- **Доступ:** cluster-internal

### 11. gRPC — TCP 8444

- **Назначение:** control plane
- **Доступ:** internal

## Команды firewall

### UFW

```bash
# Критические
sudo ufw allow 5552/tcp comment "P2P API"
sudo ufw allow 5553/udp comment "P2P QUIC"
sudo ufw allow 19302/udp comment "STUN"

# Рекомендуемые
sudo ufw allow 9090/udp comment "QUIC Transport"
sudo ufw allow 8443/tcp comment "MASQUE HTTP/3"
sudo ufw allow 3478/udp comment "TURN"
sudo ufw allow 51820/udp comment "WireGuard VPN"

# Metrics — только из monitoring CIDR
# sudo ufw allow from 10.0.0.0/8 to any port 5551 proto tcp comment "Relay metrics"
```

### iptables

```bash
sudo iptables -A INPUT -p tcp --dport 5552 -j ACCEPT
sudo iptables -A INPUT -p udp --dport 5553 -j ACCEPT
sudo iptables -A INPUT -p udp --dport 19302 -j ACCEPT
sudo iptables -A INPUT -p udp --dport 9090 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 8443 -j ACCEPT
```

## Уровни безопасности

| Порт | Протокол | Уровень | Заметка |
|-----:|----------|---------|---------|
| 5552 | TCP | Высокая | JWT |
| 8443 | TCP | Высокая | TLS / MASQUE |
| 5553 | UDP | Средняя | QUIC |
| 9090 | UDP | Средняя | QUIC transport |
| 19302 | UDP | Средняя | STUN |
| 51820 | UDP | Высокая | WireGuard crypto |
| 3478 | UDP/TCP | Средняя | TURN |
| **5551** | **TCP** | Низкая (internal) | **Metrics only** |
| 9091 | TCP | Internal | Health tunnel |

## Минимальные профили

### Базовый P2P

```bash
sudo ufw allow 5552/tcp
sudo ufw allow 5553/udp
sudo ufw allow 19302/udp
```

### Полная публичная поверхность

```bash
sudo ufw allow 5552/tcp
sudo ufw allow 5553/udp
sudo ufw allow 19302/udp
sudo ufw allow 9090/udp   # QUIC transport (not metrics)
sudo ufw allow 8443/tcp
sudo ufw allow 51820/udp
```

## Поддомены (типовая схема)

| Поддомен | Backend port | Назначение |
|----------|-------------:|------------|
| relay / p2p / api | **5552** TCP | P2P REST API |
| masque | **8443** TCP | MASQUE |
| admin | **9090** TCP | HTTP API (ingress) |
| monitor (relay scrape) | **5551** TCP | Relay `/metrics` |
| monitor (Prometheus UI) | **9090** TCP | Prometheus server (host/stack), не listener metrics relay |

## Важные замечания

1. **5552** — критический API entry
2. **5553/UDP** — P2P QUIC
3. **19302** — STUN
4. **9090/UDP** — QUIC transport; **9090/TCP** — HTTP API или host Prometheus
5. **5551/TCP** — **единственный** canonical metrics port relay
6. **Не** открывать metrics (5551) в интернет без network policy

## Проверка

```bash
sudo ufw status
sudo ss -tulpn | grep -E '(5551|5552|5553|8443|9090|19302|51820|3478|9091)'
curl -sS http://127.0.0.1:5551/metrics | head
```

## Приоритет

1. **Критические:** 5552/TCP, 5553/UDP, 19302/UDP  
2. **Рекомендуемые:** 9090/UDP, 8443/TCP, 3478, 51820/UDP  
3. **Internal:** 5551/TCP metrics, 9091 health, 8444 gRPC  

См. [port-scheme.md](../../openwiki/port-scheme.md) для auxiliary (IPAM **8080**, AI **8083**, Worker **8085**).
