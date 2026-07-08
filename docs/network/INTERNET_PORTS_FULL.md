# CloudBridge Relay - Порты для интернет доступа

## 🔥 КРИТИЧЕСКИЕ порты (ОБЯЗАТЕЛЬНО открыть):

### 1. **P2P API (TCP 5552)**
- **Назначение**: Регистрация клиентов, P2P API
- **Протокол**: TCP
- **Безопасность**: Высокая (требует JWT токен)
- **Пример**: `ufw allow 5552/tcp`

### 2. **P2P QUIC (UDP 5553)**
- **Назначение**: P2P соединения через QUIC
- **Протокол**: UDP
- **Безопасность**: Средняя (защищено QUIC)
- **Пример**: `ufw allow 5553/udp`

### 3. **STUN (UDP 19302)**
- **Назначение**: NAT traversal для P2P
- **Протокол**: UDP
- **Безопасность**: Средняя (публичный STUN)
- **Пример**: `ufw allow 19302/udp`

## РЕКОМЕНДУЕМЫЕ порты (для полной функциональности):

### 4. **Health Check (TCP 5551)**
- **Назначение**: Health check для load balancer
- **Протокол**: TCP
- **Безопасность**: Низкая (публичный health check)
- **Пример**: `ufw allow 5551/tcp`

### 5. **QUIC Transport (UDP 9090)**
- **Назначение**: Дополнительный QUIC транспорт
- **Протокол**: UDP
- **Безопасность**: Средняя (защищено QUIC)
- **Пример**: `ufw allow 9090/udp`

### 6. **MASQUE HTTP/3 (TCP 8443)**
- **Назначение**: HTTP/3 MASQUE прокси
- **Протокол**: TCP
- **Безопасность**: Высокая (TLS 1.3)
- **Пример**: `ufw allow 8443/tcp`

##  ДОПОЛНИТЕЛЬНЫЕ порты (опционально):

### 7. **Admin API (TCP 5551)**
- **Назначение**: Административная панель
- **Протокол**: TCP
- **Безопасность**: Высокая (требует аутентификации)
- **Пример**: `ufw allow 5551/tcp`

### 8. **Metrics (UDP 9090)**
- **Назначение**: Prometheus метрики
- **Протокол**: UDP
- **Безопасность**: Низкая (публичные метрики)
- **Пример**: `ufw allow 9090/udp`

### 9. **WireGuard (UDP 51820)**
- **Назначение**: WireGuard VPN mesh (`wireguard.interface.listen_port`)
- **Протокол**: UDP
- **Безопасность**: Высокая (WireGuard crypto)
- **Пример**: `ufw allow 51820/udp`

## Команды для открытия портов:

### UFW (Ubuntu Firewall):
```bash
# Критические порты
sudo ufw allow 5552/tcp comment "P2P API"
sudo ufw allow 5553/udp comment "P2P QUIC"
sudo ufw allow 19302/udp comment "STUN"

# Рекомендуемые порты
sudo ufw allow 5551/tcp comment "Health Check"
sudo ufw allow 9090/udp comment "QUIC Transport"
sudo ufw allow 8443/tcp comment "MASQUE HTTP/3"

# Дополнительные порты
sudo ufw allow 51820/udp comment "WireGuard VPN"
```

### iptables (альтернатива):
```bash
# Критические порты
sudo iptables -A INPUT -p tcp --dport 5552 -j ACCEPT
sudo iptables -A INPUT -p udp --dport 5553 -j ACCEPT
sudo iptables -A INPUT -p udp --dport 19302 -j ACCEPT

# Рекомендуемые порты
sudo iptables -A INPUT -p tcp --dport 5551 -j ACCEPT
sudo iptables -A INPUT -p udp --dport 9090 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 8443 -j ACCEPT
```

## Безопасность портов:

### Высокая безопасность (требуют аутентификации):
- **5552/TCP** - P2P API (JWT токен)
- **8443/TCP** - MASQUE HTTP/3 (TLS 1.3)

### Средняя безопасность (защищено протоколом):
- **5553/UDP** - P2P QUIC (QUIC шифрование)
- **19302/UDP** - STUN (публичный, но ограниченный)
- **9090/UDP** - QUIC Transport (QUIC шифрование)
- **51820/UDP** - WebRTC (WebRTC шифрование)

### Низкая безопасность (публичные):
- **5551/TCP** - Health Check (публичный health check)
- **9090/UDP** - Metrics (публичные метрики)

## Минимальная конфигурация:

### Для базовой работы P2P:
```bash
sudo ufw allow 5552/tcp  # P2P API
sudo ufw allow 5553/udp  # P2P QUIC
sudo ufw allow 19302/udp # STUN
```

### Для полной функциональности:
```bash
sudo ufw allow 5552/tcp  # P2P API
sudo ufw allow 5553/udp  # P2P QUIC
sudo ufw allow 19302/udp # STUN
sudo ufw allow 5551/tcp  # Health Check
sudo ufw allow 9090/udp  # QUIC Transport
sudo ufw allow 8443/tcp  # MASQUE HTTP/3
```

## Поддомены и порты:

| Поддомен | Порт | Назначение | Безопасность |
|----------|------|------------|--------------|
| **relay.2gc.ru** | 5552/TCP | P2P API | Высокая (JWT) |
| **masque.2gc.ru** | 8443/TCP | HTTP/3 MASQUE | Высокая (TLS 1.3) |
| **p2p.2gc.ru** | 5552/TCP | P2P Discovery | Высокая (JWT) |
| **admin.2gc.ru** | 5551/TCP | Admin Panel | Высокая (Auth) |
| **monitor.2gc.ru** | 9090/UDP | Metrics | Низкая (Public) |
| **api.2gc.ru** | 5552/TCP | General API | Высокая (JWT) |

## Важные замечания:

1. **Порт 5552** - самый критический, используется для всех API
2. **Порт 5553** - необходим для P2P соединений
3. **Порт 19302** - критичен для NAT traversal
4. **Порт 8443** - для HTTP/3 MASQUE прокси
5. **Порт 5551** - для health check и админки
6. **Порт 9090** - для метрик и дополнительного QUIC

## Проверка открытых портов:

```bash
# Проверить статус UFW
sudo ufw status

# Проверить открытые порты
sudo netstat -tlnp | grep -E "(5551|5552|5553|8443|9090|19302|51820)"

# Проверить подключения
sudo ss -tlnp | grep -E "(5551|5552|5553|8443|9090|19302|51820)"
```

## Приоритет открытия портов:

1. **КРИТИЧЕСКИЕ** (обязательно):
   - 5552/TCP (P2P API)
   - 5553/UDP (P2P QUIC)
   - 19302/UDP (STUN)

2. **РЕКОМЕНДУЕМЫЕ** (для полной функциональности):
   - 5551/TCP (Health Check)
   - 9090/UDP (QUIC Transport)
   - 8443/TCP (MASQUE HTTP/3)

3. **ДОПОЛНИТЕЛЬНЫЕ** (опционально):
   - 51820/UDP (WebRTC P2P)
   - 9090/UDP (Metrics)
