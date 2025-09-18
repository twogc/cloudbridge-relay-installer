# 🔑 Быстрые изменения JWT токена для QUIC + ICE/STUN/TURN

## ❌ **Удалить из токена:**
```json
"connection_type": "wireguard"
"wireguard_config": {
  "interface_name": "wg0",
  "listen_port": 51820,
  "address": "10.0.0.100/24",
  "mtu": 1420,
  "allowed_ips": ["10.0.0.0/24", "192.168.1.0/24"]
}
```

## ✅ **Изменить в токене:**
```json
"connection_type": "quic-ice"  // было "wireguard"

"permissions": [
  "mesh_join", 
  "mesh_manage",
  "ice_signaling",    // ДОБАВИТЬ
  "quic_connect"      // ДОБАВИТЬ
]

"mesh_config": {
  "registry_url": "https://edge.2gc.ru:8080"  // было "https://mesh-registry.2gc.ru"
}
```

## ✅ **Добавить в токен:**

### **QUIC Configuration:**
```json
"quic_config": {
  "handshake_timeout": "10s",
  "idle_timeout": "30s",
  "max_streams": 100,
  "connection_migration": true,
  "multiplexing": true
}
```

### **ICE Configuration:**
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

### **MASQUE Configuration:**
```json
"masque_config": {
  "enabled": true,
  "proxy_url": "https://edge.2gc.ru:8443",
  "alpn_protocols": ["h3", "h3-29"],
  "fallback_enabled": true
}
```

### **Relay Endpoints:**
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

## 🎯 **Итог:**
**Заменить WireGuard конфигурацию на QUIC + ICE/STUN/TURN конфигурацию с новыми endpoints для edge.2gc.ru**
