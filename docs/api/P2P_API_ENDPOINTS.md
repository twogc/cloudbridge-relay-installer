# 🌐 **P2P API Endpoints - Документация**

## 📋 **Обзор**

Документация по P2P API endpoints для CloudBridge Relay. Все endpoints доступны через P2P API Service (порт 8082).

---

## 🔧 **P2P API (порт 8082)**

### **Health и Status**

#### **GET /health**
Проверка состояния P2P API Service.

**Ответ:**
```json
{
  "status": "healthy",
  "timestamp": "2025-09-20T13:21:16.974571811Z",
  "services": {
    "database": "healthy",
    "redis": "healthy",
    "websocket": "healthy"
  },
  "metrics": {
    "active_connections": 0,
    "goroutines": 35,
    "memory_usage_mb": 5,
    "online_peers": 0,
    "total_peers": 0,
    "uptime_seconds": 92.061661609,
    "websocket_connections": 0
  },
  "version": "v1.0.0",
  "uptime": "1m32.061661609s"
}
```

### **Peer Management**

#### **POST /api/v1/tenants/{tenant_id}/peers/register**
Регистрация нового пира в P2P сети.

**Параметры:**
- `tenant_id` (string) - ID тенанта

**Заголовки:**
- `Authorization: Bearer {JWT_TOKEN}` - JWT токен для аутентификации
- `Content-Type: application/json`

**Тело запроса:**
```json
{
  "public_key": "test-peer-1-key",
  "allowed_ips": ["10.100.77.10/32"],
  "peer_info": {
    "name": "Test Peer 1",
    "location": "Test Location 1",
    "capabilities": ["quic", "ice", "masque"],
    "endpoints": {
      "quic": "peer1.example.com:9090",
      "stun": "peer1.example.com:19302"
    }
  }
}
```

**Ответ:**
```json
{
  "peer_id": "peer-001",
  "status": "registered",
  "assigned_ip": "10.100.77.10",
  "tunnel_config": {
    "interface": "wg0",
    "private_key": "generated-private-key",
    "public_key": "test-peer-1-key",
    "allowed_ips": ["10.100.77.0/24"],
    "endpoint": "peer1.example.com:51820",
    "persistent_keepalive": 25
  },
  "created_at": "2025-09-20T13:30:42Z"
}
```

#### **GET /api/v1/tenants/{tenant_id}/peers**
Получение списка зарегистрированных пиров.

**Параметры:**
- `tenant_id` (string) - ID тенанта

**Заголовки:**
- `Authorization: Bearer {JWT_TOKEN}` - JWT токен для аутентификации

**Ответ:**
```json
{
  "peers": [
    {
      "peer_id": "peer-001",
      "public_key": "test-peer-1-key",
      "assigned_ip": "10.100.77.10",
      "status": "online",
      "last_seen": "2025-09-20T13:30:42Z",
      "peer_info": {
        "name": "Test Peer 1",
        "location": "Test Location 1"
      }
    }
  ],
  "total_count": 1
}
```

#### **GET /api/v1/tenants/{tenant_id}/peers/{peer_id}**
Получение информации о конкретном пире.

**Параметры:**
- `tenant_id` (string) - ID тенанта
- `peer_id` (string) - ID пира

**Заголовки:**
- `Authorization: Bearer {JWT_TOKEN}` - JWT токен для аутентификации

**Ответ:**
```json
{
  "peer_id": "peer-001",
  "public_key": "test-peer-1-key",
  "assigned_ip": "10.100.77.10",
  "status": "online",
  "last_seen": "2025-09-20T13:30:42Z",
  "peer_info": {
    "name": "Test Peer 1",
    "location": "Test Location 1",
    "capabilities": ["quic", "ice", "masque"]
  },
  "connection_stats": {
    "bytes_received": 1024,
    "bytes_sent": 2048,
    "last_handshake": "2025-09-20T13:30:42Z",
    "latency": 45
  }
}
```

#### **DELETE /api/v1/tenants/{tenant_id}/peers/{peer_id}**
Удаление пира из P2P сети.

**Параметры:**
- `tenant_id` (string) - ID тенанта
- `peer_id` (string) - ID пира

**Заголовки:**
- `Authorization: Bearer {JWT_TOKEN}` - JWT токен для аутентификации

**Ответ:**
```json
{
  "peer_id": "peer-001",
  "status": "deleted",
  "deleted_at": "2025-09-20T13:30:42Z"
}
```

### **Connection Management**

#### **POST /api/v1/tenants/{tenant_id}/connections/establish**
Установка P2P соединения между пирами.

**Параметры:**
- `tenant_id` (string) - ID тенанта

**Заголовки:**
- `Authorization: Bearer {JWT_TOKEN}` - JWT токен для аутентификации
- `Content-Type: application/json`

**Тело запроса:**
```json
{
  "source_peer_id": "peer-001",
  "target_peer_id": "peer-002",
  "connection_type": "direct",
  "protocol": "quic",
  "configuration": {
    "timeout": 30,
    "retry_attempts": 3,
    "compression": true,
    "encryption": true
  }
}
```

**Ответ:**
```json
{
  "connection_id": "conn-001",
  "status": "established",
  "source_peer_id": "peer-001",
  "target_peer_id": "peer-002",
  "connection_type": "direct",
  "protocol": "quic",
  "latency": 45,
  "throughput": 2000,
  "established_at": "2025-09-20T13:30:42Z",
  "connection_stats": {
    "bytes_transferred": 0,
    "packets_sent": 0,
    "packets_received": 0,
    "errors": 0
  }
}
```

#### **GET /api/v1/tenants/{tenant_id}/connections**
Получение списка активных соединений.

**Параметры:**
- `tenant_id` (string) - ID тенанта

**Заголовки:**
- `Authorization: Bearer {JWT_TOKEN}` - JWT токен для аутентификации

**Ответ:**
```json
{
  "connections": [
    {
      "connection_id": "conn-001",
      "source_peer_id": "peer-001",
      "target_peer_id": "peer-002",
      "status": "active",
      "protocol": "quic",
      "latency": 45,
      "throughput": 2000,
      "established_at": "2025-09-20T13:30:42Z"
    }
  ],
  "total_count": 1
}
```

#### **GET /api/v1/tenants/{tenant_id}/connections/{connection_id}**
Получение информации о конкретном соединении.

**Параметры:**
- `tenant_id` (string) - ID тенанта
- `connection_id` (string) - ID соединения

**Заголовки:**
- `Authorization: Bearer {JWT_TOKEN}` - JWT токен для аутентификации

**Ответ:**
```json
{
  "connection_id": "conn-001",
  "source_peer_id": "peer-001",
  "target_peer_id": "peer-002",
  "status": "active",
  "protocol": "quic",
  "latency": 45,
  "throughput": 2000,
  "established_at": "2025-09-20T13:30:42Z",
  "connection_stats": {
    "bytes_transferred": 1024,
    "packets_sent": 10,
    "packets_received": 8,
    "errors": 0,
    "last_activity": "2025-09-20T13:30:42Z"
  }
}
```

#### **DELETE /api/v1/tenants/{tenant_id}/connections/{connection_id}**
Закрытие P2P соединения.

**Параметры:**
- `tenant_id` (string) - ID тенанта
- `connection_id` (string) - ID соединения

**Заголовки:**
- `Authorization: Bearer {JWT_TOKEN}` - JWT токен для аутентификации

**Ответ:**
```json
{
  "connection_id": "conn-001",
  "status": "closed",
  "closed_at": "2025-09-20T13:30:42Z"
}
```

### **Tunnel Management**

#### **POST /api/v1/tenants/{tenant_id}/tunnels/create**
Создание P2P туннеля.

**Параметры:**
- `tenant_id` (string) - ID тенанта

**Заголовки:**
- `Authorization: Bearer {JWT_TOKEN}` - JWT токен для аутентификации
- `Content-Type: application/json`

**Тело запроса:**
```json
{
  "tunnel_name": "tunnel-001",
  "source_peer_id": "peer-001",
  "target_peer_id": "peer-002",
  "tunnel_type": "quic",
  "configuration": {
    "encryption": true,
    "compression": true,
    "keepalive_interval": 30,
    "max_bandwidth": 1000000
  }
}
```

**Ответ:**
```json
{
  "tunnel_id": "tunnel-001",
  "status": "created",
  "source_peer_id": "peer-001",
  "target_peer_id": "peer-002",
  "tunnel_type": "quic",
  "created_at": "2025-09-20T13:30:42Z",
  "tunnel_config": {
    "interface": "tun0",
    "local_ip": "10.100.77.1",
    "remote_ip": "10.100.77.2",
    "mtu": 1420
  }
}
```

#### **GET /api/v1/tenants/{tenant_id}/tunnels**
Получение списка туннелей.

**Параметры:**
- `tenant_id` (string) - ID тенанта

**Заголовки:**
- `Authorization: Bearer {JWT_TOKEN}` - JWT токен для аутентификации

**Ответ:**
```json
{
  "tunnels": [
    {
      "tunnel_id": "tunnel-001",
      "source_peer_id": "peer-001",
      "target_peer_id": "peer-002",
      "status": "active",
      "tunnel_type": "quic",
      "created_at": "2025-09-20T13:30:42Z"
    }
  ],
  "total_count": 1
}
```

### **AI Integration**

#### **POST /api/v1/ai/optimize-routing**
AI-оптимизированная маршрутизация (интеграция с AI Service).

**Заголовки:**
- `Authorization: Bearer {JWT_TOKEN}` - JWT токен для аутентификации
- `Content-Type: application/json`

**Тело запроса:**
```json
{
  "tenant_id": "test-tenant",
  "source_peer_id": "peer-001",
  "target_peer_id": "peer-002",
  "data_size": 1024,
  "priority": "medium",
  "network_conditions": {
    "latency": 50,
    "bandwidth": 1000,
    "packet_loss": 0.01
  }
}
```

**Ответ:**
```json
{
  "optimization_id": "uuid",
  "status": "completed",
  "optimal_route": {
    "path": ["peer-001", "peer-002"],
    "latency": 45,
    "bandwidth": 1200,
    "reliability": 0.95
  },
  "ai_recommendations": [
    "use_direct_connection",
    "enable_compression"
  ],
  "confidence": 0.9,
  "processing_time_ms": 180
}
```

### **Network Discovery**

#### **GET /api/v1/tenants/{tenant_id}/network/topology**
Получение топологии P2P сети.

**Параметры:**
- `tenant_id` (string) - ID тенанта

**Заголовки:**
- `Authorization: Bearer {JWT_TOKEN}` - JWT токен для аутентификации

**Ответ:**
```json
{
  "topology": {
    "nodes": [
      {
        "peer_id": "peer-001",
        "ip": "10.100.77.10",
        "status": "online",
        "capabilities": ["quic", "ice", "masque"]
      },
      {
        "peer_id": "peer-002",
        "ip": "10.100.77.11",
        "status": "online",
        "capabilities": ["quic", "ice"]
      }
    ],
    "edges": [
      {
        "source": "peer-001",
        "target": "peer-002",
        "connection_type": "direct",
        "latency": 45,
        "bandwidth": 2000
      }
    ]
  },
  "last_updated": "2025-09-20T13:30:42Z"
}
```

#### **GET /api/v1/tenants/{tenant_id}/network/stats**
Получение статистики сети.

**Параметры:**
- `tenant_id` (string) - ID тенанта

**Заголовки:**
- `Authorization: Bearer {JWT_TOKEN}` - JWT токен для аутентификации

**Ответ:**
```json
{
  "network_stats": {
    "total_peers": 2,
    "online_peers": 2,
    "active_connections": 1,
    "total_bandwidth": 2000,
    "average_latency": 45,
    "packet_loss_rate": 0.01
  },
  "time_range": "24h",
  "last_updated": "2025-09-20T13:30:42Z"
}
```

---

## 🔧 **Использование**

### **Примеры запросов:**

#### **Регистрация пира:**
```bash
curl -X POST "http://localhost:8082/api/v1/tenants/tenant-216420165/peers/register" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "public_key": "test-peer-1-key",
    "allowed_ips": ["10.100.77.10/32"],
    "peer_info": {
      "name": "Test Peer 1",
      "location": "Test Location 1"
    }
  }'
```

#### **Установка соединения:**
```bash
curl -X POST "http://localhost:8082/api/v1/tenants/tenant-216420165/connections/establish" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "source_peer_id": "peer-001",
    "target_peer_id": "peer-002",
    "connection_type": "direct",
    "protocol": "quic"
  }'
```

#### **AI-оптимизированная маршрутизация:**
```bash
curl -X POST "http://localhost:8082/api/v1/ai/optimize-routing" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "test-tenant",
    "source_peer_id": "peer-001",
    "target_peer_id": "peer-002",
    "data_size": 1024,
    "priority": "medium",
    "network_conditions": {
      "latency": 50,
      "bandwidth": 1000,
      "packet_loss": 0.01
    }
  }'
```

---

## 📊 **Метрики**

### **P2P API метрики:**
- `p2p_active_connections` - Количество активных соединений
- `p2p_online_peers` - Количество онлайн пиров
- `p2p_total_peers` - Общее количество пиров
- `p2p_websocket_connections` - Количество WebSocket соединений
- `p2p_requests_total` - Общее количество запросов
- `p2p_request_duration_seconds` - Длительность обработки запросов
- `p2p_errors_total` - Общее количество ошибок

---

## 🔍 **Отладка**

### **Проверка статуса:**
```bash
# P2P API Health
curl http://localhost:8082/health

# Статус через NodePort
curl http://$(minikube ip):32500/health
```

### **Логи:**
```bash
# P2P API
kubectl logs deployment/cloudbridge-relay -n cloudbridge
```

### **Port forwarding:**
```bash
# P2P API
kubectl port-forward service/cloudbridge-relay-external 8082:8082 -n cloudbridge
```

---

## 🔐 **Аутентификация**

### **JWT Token:**
Все API endpoints требуют JWT токен в заголовке Authorization:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### **Пример токена:**
```json
{
  "protocol_type": "p2p-mesh",
  "scope": "p2p-mesh-claims",
  "org_id": "tenant-216420165",
  "tenant_id": "tenant-216420165",
  "server_id": "server-1758219195703",
  "connection_type": "quic-ice",
  "max_peers": "10",
  "permissions": ["mesh_join", "mesh_manage"],
  "network_config": {
    "subnet": "10.0.0.0/24",
    "gateway": "10.0.0.1",
    "dns": ["8.8.8.8", "1.1.1.1"],
    "mtu": 1420,
    "firewall_rules": ["allow_ssh", "allow_http"],
    "enable_ipv6": false
  }
}
```

---

**🎯 P2P API Endpoints готовы к использованию для управления P2P mesh сетями!**
