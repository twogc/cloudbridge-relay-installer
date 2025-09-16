# 📢 Уведомление DevOps: Обновление JWT токенов

## Тема: Обновление JWT токенов - добавление полей для P2P Mesh

---

**Уважаемые коллеги DevOps,**

Информируем о важном обновлении в структуре JWT токенов для CloudBridge.

## 🎯 Что изменилось

### ✅ Добавлены новые поля в JWT токены:

#### 1️⃣ **WireGuardConfig** - добавлено поле `endpoint`:
```json
{
  "wireguard_config": {
    "interface_name": "wg0",
    "listen_port": 51820,
    "address": "10.0.0.100/24",
    "mtu": 1420,
    "allowed_ips": ["10.0.0.0/24", "192.168.1.0/24"],
    "private_key": "WG_PRIVATE_KEY_HERE",
    "public_key": "WG_PUBLIC_KEY_HERE",
    "endpoint": "192.168.1.100:51820"  // ← НОВОЕ ПОЛЕ
  }
}
```

#### 2️⃣ **MeshConfig** - добавлены поля `network_id` и `subnet`:
```json
{
  "mesh_config": {
    "network_id": "mesh-network-001",  // ← НОВОЕ ПОЛЕ
    "subnet": "10.0.0.0/16",          // ← НОВОЕ ПОЛЕ
    "registry_url": "https://mesh-registry.2gc.ru",
    "heartbeat_interval": "30s",
    "max_peers": 10,
    "routing_strategy": "performance_optimal",
    "enable_auto_discovery": true,
    "trust_level": "basic"
  }
}
```

#### 3️⃣ **NetworkConfig** - уже содержит все необходимые поля:
```json
{
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

## 🔧 Технические детали

### **Валидация токенов:**
- ✅ Все новые поля проходят валидацию
- ✅ Поддерживается обратная совместимость
- ✅ Пустые поля разрешены (omitempty)

### **Примеры валидации:**
- `endpoint` должен быть в формате "IP:PORT" или "domain:PORT"
- `network_id` минимум 3 символа
- `subnet` должен быть в CIDR нотации (например, "10.0.0.0/16")

## 📋 Полный пример обновленного токена

```json
{
  "header": {
    "alg": "HS256",
    "typ": "JWT",
    "kid": "fallback-key"
  },
  "payload": {
    "protocol_type": "p2p-mesh",
    "scope": "p2p-mesh-claims",
    "tenant_id": "b2dbb292-d833-4de2-921a-2e12aeacf880",
    "server_id": "server-1757766138083",
    "connection_type": "wireguard",
    
    "network_config": {
      "subnet": "10.0.0.0/24",
      "gateway": "10.0.0.1",
      "dns": ["8.8.8.8", "1.1.1.1"],
      "mtu": 1420,
      "firewall_rules": ["allow_ssh", "allow_http"],
      "enable_ipv6": false
    },
    
    "wireguard_config": {
      "interface_name": "wg0",
      "listen_port": 51820,
      "address": "10.0.0.100/24",
      "mtu": 1420,
      "allowed_ips": ["10.0.0.0/24", "192.168.1.0/24"],
      "private_key": "WG_PRIVATE_KEY_HERE",
      "public_key": "WG_PUBLIC_KEY_HERE",
      "endpoint": "192.168.1.100:51820"
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
    
    "peer_whitelist": [
      "peer-001",
      "peer-002",
      "peer-003"
    ],
    
    "permissions": [
      "mesh_join",
      "mesh_manage",
      "wireguard_connect",
      "peer_discovery"
    ],
    
    "exp": 1757766441,
    "iat": 1757766141,
    "iss": "https://auth.2gc.ru/realms/cloudbridge",
    "aud": "account",
    "sub": "server-client-server-1757766138083",
    "jti": "jwt_p2p_mesh_token"
  }
}
```

## 🚀 Что нужно сделать DevOps

### ✅ **Никаких действий не требуется:**
- Токены остаются обратно совместимыми
- Старые токены продолжают работать
- Новые поля опциональны

### 🔍 **Рекомендации:**
1. **Обновить документацию** клиентов с новыми полями
2. **Протестировать** генерацию токенов с новыми полями
3. **Использовать новые поля** для улучшенной конфигурации P2P Mesh

## 📞 Контакты

При возникновении вопросов по новым полям токенов обращайтесь к команде разработки.

**Документация:** `docs/examples/p2p_mesh_token_example.json`  
**Тесты:** `internal/token/p2p_claims_test.go`

---
**Дата обновления:** $(date)  
**Версия:** CloudBridge Relay v1.0.0  
**Статус:** ✅ Готово к использованию
