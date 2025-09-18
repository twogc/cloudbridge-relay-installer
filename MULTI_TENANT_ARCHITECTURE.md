# 🏢 **CloudBridge Relay - Архитектура мультитунельности**

## 🎯 **Цель мультитунельности**

Обеспечить полную изоляцию между разными пользователями/организациями, позволяя каждому создавать свою собственную P2P mesh сеть из удаленных серверов.

## 🏗️ **Архитектура изоляции**

### **Уровень 1: JWT Token изоляция**

```json
{
  "tenant_id": "tenant-216420165",  // Уникальный идентификатор организации
  "org_id": "tenant-216420165",     // Дублирование для совместимости
  "server_id": "server-1758219195703", // Уникальный сервер в рамках tenant
  "permissions": ["mesh_join", "mesh_manage"], // Права доступа
  "peer_whitelist": ["peer-001", "peer-002", "peer-003"] // Разрешенные пиры
}
```

### **Уровень 2: Сетевая изоляция**

```
┌─────────────────────────────────────────────────────────────────┐
│                    CloudBridge Relay                            │
├─────────────────────────────────────────────────────────────────┤
│  Tenant A (tenant-216420165)                                   │
│  Network: 10.100.77.0/24                                      │
│  ┌─────────────────┐    ┌─────────────────┐                    │
│  │   Server A1     │◄──►│   Server A2     │                    │
│  │ 10.100.77.10    │    │ 10.100.77.11    │                    │
│  │ peer-001        │    │ peer-002        │                    │
│  └─────────────────┘    └─────────────────┘                    │
├─────────────────────────────────────────────────────────────────┤
│  Tenant B (tenant-123456789)                                   │
│  Network: 10.100.78.0/24                                      │
│  ┌─────────────────┐    ┌─────────────────┐                    │
│  │   Server B1     │◄──►│   Server B2     │                    │
│  │ 10.100.78.10    │    │ 10.100.78.11    │                    │
│  │ peer-101        │    │ peer-102        │                    │
│  └─────────────────┘    └─────────────────┘                    │
├─────────────────────────────────────────────────────────────────┤
│  Tenant C (tenant-987654321)                                   │
│  Network: 10.100.79.0/24                                      │
│  ┌─────────────────┐    ┌─────────────────┐                    │
│  │   Server C1     │◄──►│   Server C2     │                    │
│  │ 10.100.79.10    │    │ 10.100.79.11    │                    │
│  │ peer-201        │    │ peer-202        │                    │
│  └─────────────────┘    └─────────────────┘                    │
└─────────────────────────────────────────────────────────────────┘
```

### **Уровень 3: P2P соединения изоляция**

```go
type TenantIsolation struct {
    TenantID        string                    `json:"tenant_id"`
    NetworkCIDR     string                    `json:"network_cidr"`     // 10.100.77.0/24
    Peers           map[string]*Peer          `json:"peers"`            // peer-001, peer-002
    Connections     map[string]*P2PConnection `json:"connections"`      // conn-001, conn-002
    AllowedIPs      []string                  `json:"allowed_ips"`      // Разрешенные IP
    MaxPeers        int                       `json:"max_peers"`        // Максимум пиров
    MaxConnections  int                       `json:"max_connections"`  // Максимум соединений
}

type P2PConnection struct {
    ConnectionID    string    `json:"connection_id"`
    TenantID        string    `json:"tenant_id"`        // Обязательная проверка
    SourcePeer      string    `json:"source_peer"`      // peer-001
    TargetPeer      string    `json:"target_peer"`      // peer-002
    SourceIP        string    `json:"source_ip"`        // 10.100.77.10
    TargetIP        string    `json:"target_ip"`        // 10.100.77.11
    ConnectionType  string    `json:"connection_type"`  // "quic", "quic-ice"
    Status          string    `json:"status"`          // "connecting", "established"
    CreatedAt       time.Time `json:"created_at"`
    LastActivity    time.Time `json:"last_activity"`
    BytesTransferred int64    `json:"bytes_transferred"`
}
```

## 🔒 **Механизмы изоляции**

### **1. JWT Token валидация:**

```go
func (s *Server) validateTenantAccess(token *jwt.Token, requiredTenant string) error {
    claims := token.Claims.(*P2PClaims)
    
    // Проверяем tenant_id
    if claims.TenantID != requiredTenant {
        return fmt.Errorf("tenant mismatch: expected %s, got %s", requiredTenant, claims.TenantID)
    }
    
    // Проверяем права доступа
    if !contains(claims.Permissions, "mesh_join") {
        return fmt.Errorf("insufficient permissions: mesh_join required")
    }
    
    return nil
}
```

### **2. Сетевая изоляция:**

```go
func (s *Server) allocateIPForTenant(tenantID string) (string, error) {
    // Каждый tenant получает свой подсеть
    tenantNetworks := map[string]string{
        "tenant-216420165": "10.100.77.0/24",
        "tenant-123456789": "10.100.78.0/24", 
        "tenant-987654321": "10.100.79.0/24",
    }
    
    network, exists := tenantNetworks[tenantID]
    if !exists {
        return "", fmt.Errorf("unknown tenant: %s", tenantID)
    }
    
    // Выделяем свободный IP в подсети tenant
    ip := s.allocateIPInSubnet(network)
    return ip, nil
}
```

### **3. P2P маршрутизация с изоляцией:**

```go
func (s *Server) routeP2PPacket(sourcePeer, targetPeer string, data []byte) error {
    // 1. Получаем информацию о пирах
    sourcePeerInfo := s.getPeerInfo(sourcePeer)
    targetPeerInfo := s.getPeerInfo(targetPeer)
    
    // 2. Проверяем изоляцию
    if sourcePeerInfo.TenantID != targetPeerInfo.TenantID {
        return fmt.Errorf("cross-tenant communication not allowed: %s -> %s", 
            sourcePeerInfo.TenantID, targetPeerInfo.TenantID)
    }
    
    // 3. Проверяем whitelist
    if !s.isPeerAllowed(sourcePeerInfo.TenantID, targetPeer) {
        return fmt.Errorf("peer %s not in whitelist for tenant %s", 
            targetPeer, sourcePeerInfo.TenantID)
    }
    
    // 4. Пересылаем пакет
    return s.forwardPacket(sourcePeerInfo, targetPeerInfo, data)
}
```

## 📊 **Примеры использования**

### **Сценарий 1: Компания A (tenant-216420165)**

```bash
# Сервер 1 (Office A)
curl -X POST "http://edge.2gc.ru:8082/api/v1/tenants/tenant-216420165/peers/register" \
  -H "Authorization: Bearer $JWT_TOKEN_A" \
  -H "Content-Type: application/json" \
  -d '{
    "public_key": "server-a1-public-key",
    "allowed_ips": ["10.100.77.10/32"],
    "peer_info": {
      "name": "Office A Server",
      "location": "Moscow Office"
    }
  }'

# Сервер 2 (Office B) 
curl -X POST "http://edge.2gc.ru:8082/api/v1/tenants/tenant-216420165/peers/register" \
  -H "Authorization: Bearer $JWT_TOKEN_A" \
  -H "Content-Type: application/json" \
  -d '{
    "public_key": "server-a2-public-key", 
    "allowed_ips": ["10.100.77.11/32"],
    "peer_info": {
      "name": "Office B Server",
      "location": "SPB Office"
    }
  }'

# Результат: Серверы A1 и A2 могут общаться друг с другом
# IP адреса: 10.100.77.10 ↔ 10.100.77.11
```

### **Сценарий 2: Компания B (tenant-123456789)**

```bash
# Сервер 1 (Company B Office)
curl -X POST "http://edge.2gc.ru:8082/api/v1/tenants/tenant-123456789/peers/register" \
  -H "Authorization: Bearer $JWT_TOKEN_B" \
  -H "Content-Type: application/json" \
  -d '{
    "public_key": "server-b1-public-key",
    "allowed_ips": ["10.100.78.10/32"],
    "peer_info": {
      "name": "Company B Server",
      "location": "Company B Office"
    }
  }'

# Результат: Сервер B1 НЕ может общаться с серверами A1/A2
# Изоляция: 10.100.77.0/24 (Company A) ≠ 10.100.78.0/24 (Company B)
```

## 🔧 **Реализация в коде**

### **1. Tenant Manager:**

```go
type TenantManager struct {
    tenants map[string]*Tenant
    mutex   sync.RWMutex
}

type Tenant struct {
    TenantID        string
    NetworkCIDR     string
    Peers           map[string]*Peer
    Connections     map[string]*P2PConnection
    MaxPeers        int
    MaxConnections  int
    CreatedAt       time.Time
    LastActivity    time.Time
}

func (tm *TenantManager) GetTenant(tenantID string) (*Tenant, error) {
    tm.mutex.RLock()
    defer tm.mutex.RUnlock()
    
    tenant, exists := tm.tenants[tenantID]
    if !exists {
        return nil, fmt.Errorf("tenant not found: %s", tenantID)
    }
    
    return tenant, nil
}

func (tm *TenantManager) AddPeer(tenantID string, peer *Peer) error {
    tm.mutex.Lock()
    defer tm.mutex.Unlock()
    
    tenant, exists := tm.tenants[tenantID]
    if !exists {
        return fmt.Errorf("tenant not found: %s", tenantID)
    }
    
    if len(tenant.Peers) >= tenant.MaxPeers {
        return fmt.Errorf("tenant %s has reached max peers limit: %d", tenantID, tenant.MaxPeers)
    }
    
    tenant.Peers[peer.PeerID] = peer
    return nil
}
```

### **2. P2P Connection с изоляцией:**

```go
func (s *Server) EstablishP2PConnection(sourcePeerID, targetPeerID string) (*P2PConnection, error) {
    // 1. Получаем информацию о пирах
    sourcePeer := s.getPeer(sourcePeerID)
    targetPeer := s.getPeer(targetPeerID)
    
    if sourcePeer == nil || targetPeer == nil {
        return nil, fmt.Errorf("peer not found")
    }
    
    // 2. Проверяем изоляцию
    if sourcePeer.TenantID != targetPeer.TenantID {
        return nil, fmt.Errorf("cross-tenant connection not allowed: %s -> %s", 
            sourcePeer.TenantID, targetPeer.TenantID)
    }
    
    // 3. Создаем соединение
    connection := &P2PConnection{
        ConnectionID:   generateConnectionID(),
        TenantID:       sourcePeer.TenantID,
        SourcePeer:     sourcePeerID,
        TargetPeer:     targetPeerID,
        SourceIP:       sourcePeer.AllocatedIP,
        TargetIP:       targetPeer.AllocatedIP,
        ConnectionType: "quic-ice",
        Status:         "connecting",
        CreatedAt:      time.Now(),
    }
    
    // 4. Добавляем в tenant
    tenant := s.tenantManager.GetTenant(sourcePeer.TenantID)
    tenant.Connections[connection.ConnectionID] = connection
    
    return connection, nil
}
```

### **3. QUIC Relay с изоляцией:**

```go
func (s *Server) handleQUICRelay(conn *quic.Conn, peerID string, data []byte) error {
    // 1. Определяем отправителя
    sourcePeer := s.getPeer(peerID)
    if sourcePeer == nil {
        return fmt.Errorf("source peer not found: %s", peerID)
    }
    
    // 2. Парсим пакет для определения получателя
    targetPeerID := s.parseQUICPacketTarget(data)
    if targetPeerID == "" {
        return fmt.Errorf("cannot determine target peer from packet")
    }
    
    // 3. Получаем информацию о получателе
    targetPeer := s.getPeer(targetPeerID)
    if targetPeer == nil {
        return fmt.Errorf("target peer not found: %s", targetPeerID)
    }
    
    // 4. Проверяем изоляцию
    if sourcePeer.TenantID != targetPeer.TenantID {
        return fmt.Errorf("cross-tenant relay not allowed: %s -> %s", 
            sourcePeer.TenantID, targetPeer.TenantID)
    }
    
    // 5. Пересылаем пакет
    targetConn := s.getConnection(targetPeerID)
    if targetConn == nil {
        return fmt.Errorf("no active connection for target peer: %s", targetPeerID)
    }
    
    _, err := targetConn.Write(data)
    return err
}
```

## 📈 **Метрики и мониторинг**

### **Метрики по tenant:**

```go
type TenantMetrics struct {
    TenantID           string    `json:"tenant_id"`
    ActivePeers        int       `json:"active_peers"`
    ActiveConnections  int       `json:"active_connections"`
    BytesTransferred   int64     `json:"bytes_transferred"`
    PacketsTransferred int64     `json:"packets_transferred"`
    LastActivity       time.Time `json:"last_activity"`
    Uptime             time.Duration `json:"uptime"`
}

func (s *Server) GetTenantMetrics(tenantID string) (*TenantMetrics, error) {
    tenant := s.tenantManager.GetTenant(tenantID)
    if tenant == nil {
        return nil, fmt.Errorf("tenant not found: %s", tenantID)
    }
    
    metrics := &TenantMetrics{
        TenantID:           tenantID,
        ActivePeers:        len(tenant.Peers),
        ActiveConnections:  len(tenant.Connections),
        LastActivity:       tenant.LastActivity,
        Uptime:             time.Since(tenant.CreatedAt),
    }
    
    // Подсчитываем трафик
    for _, conn := range tenant.Connections {
        metrics.BytesTransferred += conn.BytesTransferred
    }
    
    return metrics, nil
}
```

## 🚀 **Преимущества мультитунельности**

### **1. Безопасность:**
- ✅ Полная изоляция между организациями
- ✅ JWT токены содержат tenant_id
- ✅ Сетевые пространства разделены
- ✅ Невозможность cross-tenant коммуникации

### **2. Масштабируемость:**
- ✅ Каждый tenant может иметь до 100+ серверов
- ✅ Отдельные лимиты для каждого tenant
- ✅ Независимое управление ресурсами

### **3. Управление:**
- ✅ Централизованное управление через JWT
- ✅ Whitelist пиров на уровне tenant
- ✅ Мониторинг по tenant
- ✅ Graceful shutdown по tenant

### **4. Гибкость:**
- ✅ Разные сетевые конфигурации для tenant
- ✅ Настраиваемые лимиты
- ✅ Разные протоколы (QUIC, ICE/STUN/TURN)
- ✅ Поддержка MASQUE для корпоративных сетей

---

**🎯 Результат: Каждая организация получает свою собственную изолированную P2P mesh сеть, где удаленные серверы могут общаться как в локальной сети, но полностью изолированы от других организаций!**
