# CloudBridge Relay - Сводка по портам

> **Canonical source:** [openwiki/port-scheme.md](../../openwiki/port-scheme.md) — единая таблица service → port из `configs/config.json`, SERVICE_CATALOG и k8s-manifests. Этот файл — Minikube NodePort и firewall для продакшена.

## Текущая конфигурация Minikube

### Открытые порты:
```
30080/TCP  - Main Server (5550)
30081/TCP  - Metrics (5551) 
32500/TCP  - P2P API (5552)  ОСНОВНОЙ
30087/TCP  - HTTP API (5553)
30302/UDP  - STUN Server (19302)  ОСНОВНОЙ
30090/UDP  - QUIC Transport (9090)
30443/TCP  - MASQUE Proxy (8443)
30092/UDP  - Enhanced QUIC (9092)
30844/TCP  - gRPC (8444)
```

## Для интернет-доступа нужно открыть:

### 🔥 КРИТИЧЕСКИЕ порты:
```
5552/TCP   - P2P API (регистрация клиентов)
5553/UDP   - P2P QUIC (P2P соединения)
19302/UDP  - STUN (NAT traversal)
```

### РЕКОМЕНДУЕМЫЕ порты:
```
5551/TCP   - Health Check (для load balancer)
9090/UDP   - QUIC Transport (дополнительный)
```

### ВНУТРЕННИЕ порты (не открывать наружу):
```
5550/TCP   - Main Server
8443/TCP   - MASQUE Proxy
9092/UDP   - Enhanced QUIC
8444/TCP   - gRPC
```

## Firewall правила для продакшена:

```bash
# Обязательные порты
sudo ufw allow 5552/tcp comment "CloudBridge P2P API"
sudo ufw allow 5553/udp comment "CloudBridge P2P QUIC"
sudo ufw allow 19302/udp comment "CloudBridge STUN"

# Рекомендуемые порты
sudo ufw allow 5551/tcp comment "CloudBridge Health Check"
sudo ufw allow 9090/udp comment "CloudBridge QUIC Transport"
```

## Nginx конфигурация:

```nginx
# Основной API
upstream cloudbridge_api {
    server 192.168.58.2:32500;
}

server {
    listen 443 ssl;
    server_name edge.2gc.ru;
    
    # P2P API
    location /api/ {
        proxy_pass http://cloudbridge_api;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
    
    # Health check
    location /health {
        proxy_pass http://cloudbridge_api;
    }
}
```

## Kubernetes LoadBalancer:

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
  - name: health
    port: 5551
    targetPort: 5551
    protocol: TCP
  selector:
    app: cloudbridge-relay
```

## Тестирование доступности:

```bash
# Проверить HTTP API
curl -X GET https://edge.2gc.ru/health

# Проверить P2P API
curl -X POST https://edge.2gc.ru/api/v1/tenants/test/peers/register \
  -H "Content-Type: application/json" \
  -d '{"peer_id": "test-peer"}'

# Проверить STUN
nc -u edge.2gc.ru 19302

# Проверить QUIC
nc -u edge.2gc.ru 5553
```

## Безопасность:

1. **HTTPS** для HTTP API (порт 5552)
2. **JWT аутентификация** для всех API запросов
3. **Rate limiting** для предотвращения DDoS
4. **Firewall** для ограничения доступа
5. **Monitoring** для отслеживания атак

## Минимальная конфигурация для продакшена:

```yaml
# Только критически важные порты
ports:
  - 5552/tcp   # P2P API
  - 5553/udp   # P2P QUIC  
  - 19302/udp  # STUN
```

## Полная конфигурация для продакшена:

```yaml
# Все рекомендуемые порты
ports:
  - 5551/tcp   # Health Check
  - 5552/tcp   # P2P API
  - 5553/udp   # P2P QUIC
  - 19302/udp  # STUN
  - 9090/udp   # QUIC Transport
```
