# 🔑 Руководство по миграции JWT токенов для QUIC + ICE/STUN/TURN

## 📊 **Анализ текущего токена:**

### **Текущая структура (WireGuard):**
```json
{
  "protocol_type": "p2p-mesh",
  "scope": "p2p-mesh-claims",
  "org_id": "tenant-216420165",
  "tenant_id": "tenant-216420165",
  "server_id": "server-1758105689169",
  "connection_type": "wireguard",           // ❌ УДАЛИТЬ
  "max_peers": "10",
  "permissions": ["mesh_join", "mesh_manage"],
  "network_config": {
    "subnet": "10.0.0.0/24",
    "gateway": "10.0.0.1",
    "dns": ["8.8.8.8", "1.1.1.1"],
    "mtu": 1420,
    "firewall_rules": ["allow_ssh", "allow_http"],
    "enable_ipv6": false
  },
  "wireguard_config": {                     // ❌ УДАЛИТЬ
    "interface_name": "wg0",
    "listen_port": 51820,
    "address": "10.0.0.100/24",
    "mtu": 1420,
    "allowed_ips": ["10.0.0.0/24", "192.168.1.0/24"]
  },
  "mesh_config": {
    "network_id": "mesh-network-001",
    "subnet": "10.0.0.0/16",
    "registry_url": "https://mesh-registry.2gc.ru",
    "heartbeat_interval": "30s",
    "max_peers": 10,
    "routing_strategy": "performance_optimal",
    "enable_auto_discovery": true,
    "trust_level": "basic"
  },
  "peer_whitelist": ["peer-001", "peer-002", "peer-003"]
}
```

---

## 🚀 **Новая структура токена (QUIC + ICE/STUN/TURN):**

### **Обновленная структура:**
```json
{
  "protocol_type": "p2p-mesh",
  "scope": "p2p-mesh-claims",
  "org_id": "tenant-216420165",
  "tenant_id": "tenant-216420165",
  "server_id": "server-1758105689169",
  "connection_type": "quic-ice",            // ✅ ИЗМЕНИТЬ
  "max_peers": "10",
  "permissions": ["mesh_join", "mesh_manage", "ice_signaling", "quic_connect"],
  "network_config": {
    "subnet": "10.0.0.0/24",
    "gateway": "10.0.0.1",
    "dns": ["8.8.8.8", "1.1.1.1"],
    "mtu": 1420,
    "firewall_rules": ["allow_ssh", "allow_http"],
    "enable_ipv6": false
  },
  "quic_config": {                          // ✅ ДОБАВИТЬ
    "handshake_timeout": "10s",
    "idle_timeout": "30s",
    "max_streams": 100,
    "connection_migration": true,
    "multiplexing": true
  },
  "ice_config": {                           // ✅ ДОБАВИТЬ
    "stun_servers": ["edge.2gc.ru:19302"],
    "turn_servers": [],
    "ice_timeout": "30s",
    "connection_timeout": "10s",
    "gathering_timeout": "5s",
    "candidate_types": ["host", "srflx", "relay"]
  },
  "masque_config": {                        // ✅ ДОБАВИТЬ
    "enabled": true,
    "proxy_url": "https://edge.2gc.ru:8443",
    "alpn_protocols": ["h3", "h3-29"],
    "fallback_enabled": true
  },
  "mesh_config": {
    "network_id": "mesh-network-001",
    "subnet": "10.0.0.0/16",
    "registry_url": "https://edge.2gc.ru:8080",  // ✅ ИЗМЕНИТЬ
    "heartbeat_interval": "30s",
    "max_peers": 10,
    "routing_strategy": "performance_optimal",
    "enable_auto_discovery": true,
    "trust_level": "basic"
  },
  "peer_whitelist": ["peer-001", "peer-002", "peer-003"],
  "relay_endpoints": {                      // ✅ ДОБАВИТЬ
    "http_api": "edge.2gc.ru:8080",
    "p2p_api": "edge.2gc.ru:8082",
    "quic_transport": "edge.2gc.ru:9090",
    "stun_server": "edge.2gc.ru:19302",
    "masque_proxy": "edge.2gc.ru:8443",
    "enhanced_quic": "edge.2gc.ru:9092"
  }
}
```

---

## 🔧 **Изменения в личном кабинете:**

### **1. Удалить поля:**
```json
// ❌ УДАЛИТЬ эти поля из формы генерации токена:
"connection_type": "wireguard"
"wireguard_config": {
  "interface_name": "wg0",
  "listen_port": 51820,
  "address": "10.0.0.100/24",
  "mtu": 1420,
  "allowed_ips": ["10.0.0.0/24", "192.168.1.0/24"]
}
```

### **2. Изменить поля:**
```json
// ✅ ИЗМЕНИТЬ:
"connection_type": "quic-ice"  // было "wireguard"

"permissions": [
  "mesh_join", 
  "mesh_manage",
  "ice_signaling",    // ✅ ДОБАВИТЬ
  "quic_connect"      // ✅ ДОБАВИТЬ
]

"mesh_config": {
  "registry_url": "https://edge.2gc.ru:8080"  // было "https://mesh-registry.2gc.ru"
}
```

### **3. Добавить новые поля:**

#### **QUIC Configuration:**
```json
"quic_config": {
  "handshake_timeout": "10s",
  "idle_timeout": "30s", 
  "max_streams": 100,
  "connection_migration": true,
  "multiplexing": true
}
```

#### **ICE Configuration:**
```json
"ice_config": {
  "stun_servers": ["edge.2gc.ru:19302"],
  "turn_servers": [],
  "ice_timeout": "30s",
  "connection_timeout": "10s",
  "gathering_timeout": "5s",
  "candidate_types": ["host", "srflx", "relay"]
}
```

#### **MASQUE Configuration:**
```json
"masque_config": {
  "enabled": true,
  "proxy_url": "https://edge.2gc.ru:8443",
  "alpn_protocols": ["h3", "h3-29"],
  "fallback_enabled": true
}
```

#### **Relay Endpoints:**
```json
"relay_endpoints": {
  "http_api": "edge.2gc.ru:8080",
  "p2p_api": "edge.2gc.ru:8082", 
  "quic_transport": "edge.2gc.ru:9090",
  "stun_server": "edge.2gc.ru:19302",
  "masque_proxy": "edge.2gc.ru:8443",
  "enhanced_quic": "edge.2gc.ru:9092"
}
```

---

## 🎯 **Пример нового токена:**

### **Полный пример JWT payload:**
```json
{
  "protocol_type": "p2p-mesh",
  "scope": "p2p-mesh-claims",
  "org_id": "tenant-216420165",
  "tenant_id": "tenant-216420165", 
  "server_id": "server-1758105689169",
  "connection_type": "quic-ice",
  "max_peers": "10",
  "permissions": ["mesh_join", "mesh_manage", "ice_signaling", "quic_connect"],
  "network_config": {
    "subnet": "10.0.0.0/24",
    "gateway": "10.0.0.1",
    "dns": ["8.8.8.8", "1.1.1.1"],
    "mtu": 1420,
    "firewall_rules": ["allow_ssh", "allow_http"],
    "enable_ipv6": false
  },
  "quic_config": {
    "handshake_timeout": "10s",
    "idle_timeout": "30s",
    "max_streams": 100,
    "connection_migration": true,
    "multiplexing": true
  },
  "ice_config": {
    "stun_servers": ["edge.2gc.ru:19302"],
    "turn_servers": [],
    "ice_timeout": "30s",
    "connection_timeout": "10s",
    "gathering_timeout": "5s",
    "candidate_types": ["host", "srflx", "relay"]
  },
  "masque_config": {
    "enabled": true,
    "proxy_url": "https://edge.2gc.ru:8443",
    "alpn_protocols": ["h3", "h3-29"],
    "fallback_enabled": true
  },
  "mesh_config": {
    "network_id": "mesh-network-001",
    "subnet": "10.0.0.0/16",
    "registry_url": "https://edge.2gc.ru:8080",
    "heartbeat_interval": "30s",
    "max_peers": 10,
    "routing_strategy": "performance_optimal",
    "enable_auto_discovery": true,
    "trust_level": "basic"
  },
  "peer_whitelist": ["peer-001", "peer-002", "peer-003"],
  "relay_endpoints": {
    "http_api": "edge.2gc.ru:8080",
    "p2p_api": "edge.2gc.ru:8082",
    "quic_transport": "edge.2gc.ru:9090",
    "stun_server": "edge.2gc.ru:19302",
    "masque_proxy": "edge.2gc.ru:8443",
    "enhanced_quic": "edge.2gc.ru:9092"
  },
  "iat": 1758196574,
  "iss": "https://auth.2gc.ru/realms/cloudbridge",
  "aud": "account",
  "sub": "server-client-server-1758105689169",
  "jti": "jwt_1758196574771_z97klmj0t"
}
```

---

## 🔧 **Изменения в UI личного кабинета:**

### **1. Удалить секции:**
- ❌ WireGuard Configuration
- ❌ Interface Settings
- ❌ Port Configuration

### **2. Добавить новые секции:**

#### **QUIC Settings:**
```
☐ Enable QUIC Transport
☐ Enable Connection Migration
☐ Enable Multiplexing
Handshake Timeout: [10s]
Idle Timeout: [30s]
Max Streams: [100]
```

#### **ICE Settings:**
```
☐ Enable ICE Signaling
☐ Enable STUN Discovery
☐ Enable TURN Relay
ICE Timeout: [30s]
Connection Timeout: [10s]
Gathering Timeout: [5s]
```

#### **MASQUE Settings:**
```
☐ Enable MASQUE Proxy
☐ Enable Fallback
Proxy URL: [https://edge.2gc.ru:8443]
ALPN Protocols: [h3, h3-29]
```

#### **Relay Endpoints:**
```
HTTP API: [edge.2gc.ru:8080]
P2P API: [edge.2gc.ru:8082]
QUIC Transport: [edge.2gc.ru:9090]
STUN Server: [edge.2gc.ru:19302]
MASQUE Proxy: [edge.2gc.ru:8443]
Enhanced QUIC: [edge.2gc.ru:9092]
```

---

## ✅ **Критерии готовности:**

### **Backend изменения:**
- [ ] Обновить JWT валидацию для новых полей
- [ ] Добавить поддержку quic_config, ice_config, masque_config
- [ ] Обновить relay_endpoints
- [ ] Изменить connection_type на "quic-ice"

### **Frontend изменения:**
- [ ] Удалить WireGuard секции из UI
- [ ] Добавить QUIC, ICE, MASQUE секции
- [ ] Обновить форму генерации токена
- [ ] Добавить валидацию новых полей

### **Тестирование:**
- [ ] Генерация нового токена
- [ ] Валидация токена на relay сервере
- [ ] Подключение клиента с новым токеном
- [ ] Проверка всех endpoints

---

## 🎯 **Итог:**

**Новый JWT токен будет содержать:**
- ✅ QUIC конфигурацию вместо WireGuard
- ✅ ICE/STUN/TURN настройки
- ✅ MASQUE поддержку
- ✅ Обновленные relay endpoints
- ✅ Новые permissions для P2P

**Готово для миграции на QUIC + ICE/STUN/TURN архитектуру!** 🚀
