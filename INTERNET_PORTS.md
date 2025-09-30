# CloudBridge Relay - Порты для интернет-доступа

## Обзор архитектуры

CloudBridge Relay использует несколько портов для различных функций:
- **HTTP API** - для регистрации клиентов и управления
- **P2P QUIC** - для P2P соединений
- **ICE/STUN/TURN** - для NAT traversal
- **Health Check** - для мониторинга
- **Metrics** - для метрик

## Критические порты для интернета

### 1. HTTP API (Обязательно)
```
Порт: 5552 (TCP)
Назначение: P2P API для регистрации клиентов
Доступ: Публичный
Безопасность: HTTPS рекомендуется
```

### 2. P2P QUIC (Обязательно)
```
Порт: 5553 (UDP)
Назначение: P2P соединения между клиентами
Доступ: Публичный
Протокол: QUIC over UDP
```

### 3. ICE/STUN (Обязательно)
```
Порт: 19302 (UDP)
Назначение: STUN сервер для NAT traversal
Доступ: Публичный
Протокол: STUN over UDP
```

## Дополнительные порты

### 4. Health Check (Рекомендуется)
```
Порт: 5551 (TCP)
Назначение: Health check для load balancer
Доступ: Публичный (только для LB)
Безопасность: Внутренний доступ
```

### 5. Metrics (Опционально)
```
Порт: 9090 (TCP)
Назначение: Prometheus метрики
Доступ: Ограниченный (только для мониторинга)
Безопасность: Аутентификация обязательна
```

## Конфигурация для продакшена

### Минимальная конфигурация (3 порта)
```yaml
# Обязательные порты для работы
ports:
  - port: 5552    # HTTP API
    protocol: TCP
    public: true
  - port: 5553    # P2P QUIC
    protocol: UDP
    public: true
  - port: 19302   # STUN
    protocol: UDP
    public: true
```

### Полная конфигурация (5 портов)
```yaml
# Все порты для полной функциональности
ports:
  - port: 5551    # Health Check
    protocol: TCP
    public: true
  - port: 5552    # HTTP API
    protocol: TCP
    public: true
  - port: 5553    # P2P QUIC
    protocol: UDP
    public: true
  - port: 19302   # STUN
    protocol: UDP
    public: true
  - port: 9090    # Metrics (опционально)
    protocol: TCP
    public: false  # Только для мониторинга
```

## Настройка Load Balancer

### Nginx конфигурация
```nginx
# HTTP API
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
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    # Health check
    location /health {
        proxy_pass http://cloudbridge_api;
    }
}
```

### Kubernetes Service
```yaml
apiVersion: v1
kind: Service
metadata:
  name: cloudbridge-relay-external
  namespace: cloudbridge
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

## Firewall правила

### UFW (Ubuntu)
```bash
# HTTP API
sudo ufw allow 5552/tcp comment "CloudBridge HTTP API"

# P2P QUIC
sudo ufw allow 5553/udp comment "CloudBridge P2P QUIC"

# STUN
sudo ufw allow 19302/udp comment "CloudBridge STUN"

# Health Check
sudo ufw allow 5551/tcp comment "CloudBridge Health Check"

# Metrics (опционально)
sudo ufw allow from 10.0.0.0/8 to any port 9090 comment "CloudBridge Metrics"
```

### iptables
```bash
# HTTP API
iptables -A INPUT -p tcp --dport 5552 -j ACCEPT

# P2P QUIC
iptables -A INPUT -p udp --dport 5553 -j ACCEPT

# STUN
iptables -A INPUT -p udp --dport 19302 -j ACCEPT

# Health Check
iptables -A INPUT -p tcp --dport 5551 -j ACCEPT
```

## Безопасность

### 1. HTTPS для HTTP API
```yaml
# Включить TLS для HTTP API
tls:
  enabled: true
  cert_file: "/app/certs/tls.crt"
  key_file: "/app/certs/tls.key"
```

### 2. Аутентификация
```yaml
# Включить JWT аутентификацию
auth:
  enabled: true
  jwt_secret: "your-secret-key"
  audience: "relay-production"
```

### 3. Rate Limiting
```yaml
# Ограничить количество запросов
rate_limit:
  enabled: true
  requests_per_minute: 100
  burst_size: 20
```

## Мониторинг портов

### Проверка открытых портов
```bash
# Проверить, что порты слушают
netstat -tulpn | grep -E "(5551|5552|5553|19302|9090)"

# Проверить доступность извне
telnet your-server.com 5552
nc -u your-server.com 5553
nc -u your-server.com 19302
```

### Тестирование соединения
```bash
# Тест HTTP API
curl -X GET https://edge.2gc.ru/health

# Тест P2P регистрации
curl -X POST https://edge.2gc.ru/api/v1/tenants/test/peers/register \
  -H "Content-Type: application/json" \
  -d '{"peer_id": "test-peer", "endpoint": "1.2.3.4:5553"}'
```

## Рекомендации по развертыванию

### 1. Минимальная конфигурация
- **Порты**: 5552 (TCP), 5553 (UDP), 19302 (UDP)
- **Безопасность**: HTTPS + JWT
- **Мониторинг**: Внутренний health check

### 2. Продакшен конфигурация
- **Порты**: 5551 (TCP), 5552 (TCP), 5553 (UDP), 19302 (UDP)
- **Безопасность**: HTTPS + JWT + Rate Limiting
- **Мониторинг**: Prometheus + Grafana
- **Load Balancer**: Nginx/HAProxy

### 3. Высокая доступность
- **Множественные инстансы**: 3+ реплики
- **Load Balancer**: Round-robin
- **Health Checks**: Автоматическое восстановление
- **Мониторинг**: Полный стек наблюдения

## Устранение неполадок

### Порт не доступен
```bash
# Проверить, что сервис запущен
kubectl get pods -n cloudbridge
kubectl logs -n cloudbridge deployment/cloudbridge-relay

# Проверить порты
kubectl get services -n cloudbridge
```

### Firewall блокирует
```bash
# Проверить firewall
sudo ufw status
sudo iptables -L

# Открыть порты
sudo ufw allow 5552/tcp
sudo ufw allow 5553/udp
sudo ufw allow 19302/udp
```

### NAT проблемы
```bash
# Проверить NAT
nmap -sU -p 5553 your-server.com
nmap -sU -p 19302 your-server.com

# Тест STUN
stun-client your-server.com:19302
```

## Примеры конфигурации

### Docker Compose
```yaml
version: '3.8'
services:
  cloudbridge-relay:
    image: cloudbridge-relay:latest
    ports:
      - "5551:5551"    # Health Check
      - "5552:5552"    # HTTP API
      - "5553:5553/udp" # P2P QUIC
      - "19302:19302/udp" # STUN
    environment:
      - AUTH_ENABLED=true
      - RELAY_JWT_SECRET=your-secret
```

### Kubernetes
```yaml
apiVersion: v1
kind: Service
metadata:
  name: cloudbridge-relay-external
spec:
  type: LoadBalancer
  ports:
  - port: 5552
    targetPort: 5552
    protocol: TCP
  - port: 5553
    targetPort: 5553
    protocol: UDP
  - port: 19302
    targetPort: 19302
    protocol: UDP
```

## Заключение

Для работы CloudBridge Relay через интернет необходимо открыть:

**Обязательные порты:**
- `5552/TCP` - HTTP API
- `5553/UDP` - P2P QUIC
- `19302/UDP` - STUN

**Рекомендуемые порты:**
- `5551/TCP` - Health Check
- `9090/TCP` - Metrics (опционально)

**Безопасность:**
- HTTPS для HTTP API
- JWT аутентификация
- Rate limiting
- Firewall правила
