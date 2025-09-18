# 🚀 **CloudBridge Relay - План доработки P2P маршрутизации**

## 🎯 **Цель доработки**

Реализовать полноценную P2P маршрутизацию, чтобы удаленные серверы клиентов могли общаться друг с другом через CloudBridge Relay как в локальной сети, с полной поддержкой мультитунельности и изоляции.

## 📊 **Текущее состояние (по результатам тестирования)**

### **✅ ЧТО УЖЕ РАБОТАЕТ:**
- ✅ Клиенты успешно запускаются и регистрируются
- ✅ Сетевые интерфейсы настроены (10.100.77.77, 10.100.77.78)
- ✅ Relay сервер доступен и отвечает
- ✅ Echo функциональность работает
- ✅ JWT аутентификация работает
- ✅ Все порты доступны с внешней сети
- ✅ P2P API endpoints работают

### **❌ ЧТО ТРЕБУЕТ ДОРАБОТКИ:**
- ❌ **P2P маршрутизация не реализована** - Relay не пересылает пакеты между клиентами
- ❌ **Heartbeat/Keep-alive отсутствует** - Клиенты не остаются онлайн
- ❌ **QUIC + ICE соединения не устанавливаются** - Нет P2P mesh сети
- ❌ **Мультитунельность не реализована** - Нет изоляции между tenant

## 🏗️ **Архитектура P2P маршрутизации**

### **Целевая архитектура:**
```
┌─────────────────────────────────────────────────────────────────┐
│                    CloudBridge Relay                            │
├─────────────────────────────────────────────────────────────────┤
│  Tenant A (tenant-216420165)                                   │
│  ┌─────────────────┐    ┌─────────────────┐                    │
│  │   Client A1     │◄──►│   Client A2     │                    │
│  │ 10.100.77.10    │    │ 10.100.77.11    │                    │
│  │ QUIC Conn       │    │ QUIC Conn       │                    │
│  │ ICE Candidates  │    │ ICE Candidates  │                    │
│  └─────────────────┘    └─────────────────┘                    │
│           │                       │                            │
│           └─────────── P2P Relay ──────────────┘                │
├─────────────────────────────────────────────────────────────────┤
│  Tenant B (tenant-123456789)                                   │
│  ┌─────────────────┐    ┌─────────────────┐                    │
│  │   Client B1     │◄──►│   Client B2     │                    │
│  │ 10.100.78.10    │    │ 10.100.78.11    │                    │
│  │ QUIC Conn       │    │ QUIC Conn       │                    │
│  │ ICE Candidates  │    │ ICE Candidates  │                    │
│  └─────────────────┘    └─────────────────┘                    │
│           │                       │                            │
│           └─────────── P2P Relay ──────────────┘                │
└─────────────────────────────────────────────────────────────────┘
```

## 📋 **План реализации по этапам**

### **Этап 1: Базовая P2P маршрутизация (Приоритет: ВЫСОКИЙ)**

#### **1.1 Структуры данных для P2P соединений:**

```go
// internal/p2p/types.go
type P2PConnection struct {
    ConnectionID    string            `json:"connection_id"`
    TenantID        string            `json:"tenant_id"`
    SourcePeer      string            `json:"source_peer"`
    TargetPeer      string            `json:"target_peer"`
    SourceIP        string            `json:"source_ip"`
    TargetIP        string            `json:"target_ip"`
    ConnectionType  string            `json:"connection_type"` // "quic", "quic-ice"
    Status          string            `json:"status"`          // "connecting", "established", "disconnected"
    CreatedAt       time.Time         `json:"created_at"`
    LastActivity    time.Time         `json:"last_activity"`
    BytesTransferred int64            `json:"bytes_transferred"`
    QUICConn        *quic.Conn        `json:"-"`
    ICEConn         *ice.Conn         `json:"-"`
    Mutex           sync.RWMutex      `json:"-"`
}

type Peer struct {
    PeerID          string                    `json:"peer_id"`
    TenantID        string                    `json:"tenant_id"`
    PublicKey       string                    `json:"public_key"`
    AllowedIPs      []string                  `json:"allowed_ips"`
    AllocatedIP     string                    `json:"allocated_ip"`
    Status          string                    `json:"status"`          // "online", "offline"
    LastSeen        time.Time                 `json:"last_seen"`
    HeartbeatTicker *time.Ticker              `json:"-"`
    Connections     map[string]*P2PConnection `json:"-"`
    QUICConn        *quic.Conn                `json:"-"`
    Mutex           sync.RWMutex              `json:"-"`
}

type Tenant struct {
    TenantID        string                    `json:"tenant_id"`
    NetworkCIDR     string                    `json:"network_cidr"`
    Peers           map[string]*Peer          `json:"peers"`
    Connections     map[string]*P2PConnection `json:"connections"`
    MaxPeers        int                       `json:"max_peers"`
    MaxConnections  int                       `json:"max_connections"`
    CreatedAt       time.Time                 `json:"created_at"`
    LastActivity    time.Time                 `json:"last_activity"`
    Mutex           sync.RWMutex              `json:"-"`
}
```

#### **1.2 P2P Manager:**

```go
// internal/p2p/manager.go
type P2PManager struct {
    tenants map[string]*Tenant
    mutex   sync.RWMutex
    logger  *logrus.Logger
}

func NewP2PManager(logger *logrus.Logger) *P2PManager {
    return &P2PManager{
        tenants: make(map[string]*Tenant),
        logger:  logger,
    }
}

func (pm *P2PManager) RegisterPeer(tenantID, peerID, publicKey string, allowedIPs []string) (*Peer, error) {
    pm.mutex.Lock()
    defer pm.mutex.Unlock()
    
    // Создаем tenant если не существует
    if _, exists := pm.tenants[tenantID]; !exists {
        pm.tenants[tenantID] = &Tenant{
            TenantID:       tenantID,
            NetworkCIDR:    pm.getTenantNetwork(tenantID),
            Peers:          make(map[string]*Peer),
            Connections:    make(map[string]*P2PConnection),
            MaxPeers:       100,
            MaxConnections: 1000,
            CreatedAt:      time.Now(),
        }
    }
    
    tenant := pm.tenants[tenantID]
    
    // Проверяем лимиты
    if len(tenant.Peers) >= tenant.MaxPeers {
        return nil, fmt.Errorf("tenant %s has reached max peers limit: %d", tenantID, tenant.MaxPeers)
    }
    
    // Выделяем IP адрес
    allocatedIP := pm.allocateIPForTenant(tenantID)
    
    // Создаем peer
    peer := &Peer{
        PeerID:      peerID,
        TenantID:    tenantID,
        PublicKey:   publicKey,
        AllowedIPs:  allowedIPs,
        AllocatedIP: allocatedIP,
        Status:      "online",
        LastSeen:    time.Now(),
        Connections: make(map[string]*P2PConnection),
    }
    
    // Добавляем в tenant
    tenant.Peers[peerID] = peer
    
    pm.logger.Infof("Peer registered: %s in tenant %s with IP %s", peerID, tenantID, allocatedIP)
    
    return peer, nil
}

func (pm *P2PManager) EstablishP2PConnection(sourcePeerID, targetPeerID string) (*P2PConnection, error) {
    pm.mutex.Lock()
    defer pm.mutex.Unlock()
    
    // Находим source peer
    var sourcePeer *Peer
    var sourceTenant *Tenant
    
    for _, tenant := range pm.tenants {
        if peer, exists := tenant.Peers[sourcePeerID]; exists {
            sourcePeer = peer
            sourceTenant = tenant
            break
        }
    }
    
    if sourcePeer == nil {
        return nil, fmt.Errorf("source peer not found: %s", sourcePeerID)
    }
    
    // Находим target peer в том же tenant
    targetPeer, exists := sourceTenant.Peers[targetPeerID]
    if !exists {
        return nil, fmt.Errorf("target peer not found in same tenant: %s", targetPeerID)
    }
    
    // Проверяем лимиты соединений
    if len(sourceTenant.Connections) >= sourceTenant.MaxConnections {
        return nil, fmt.Errorf("tenant %s has reached max connections limit: %d", sourceTenant.TenantID, sourceTenant.MaxConnections)
    }
    
    // Создаем соединение
    connection := &P2PConnection{
        ConnectionID:   pm.generateConnectionID(),
        TenantID:       sourceTenant.TenantID,
        SourcePeer:     sourcePeerID,
        TargetPeer:     targetPeerID,
        SourceIP:       sourcePeer.AllocatedIP,
        TargetIP:       targetPeer.AllocatedIP,
        ConnectionType: "quic-ice",
        Status:         "connecting",
        CreatedAt:      time.Now(),
        LastActivity:   time.Now(),
    }
    
    // Добавляем в tenant
    sourceTenant.Connections[connection.ConnectionID] = connection
    
    pm.logger.Infof("P2P connection established: %s -> %s in tenant %s", sourcePeerID, targetPeerID, sourceTenant.TenantID)
    
    return connection, nil
}
```

#### **1.3 QUIC Relay Handler:**

```go
// cmd/relay/quic_relay.go
func (s *Server) handleQUICRelay(conn *quic.Conn, peerID string, data []byte) error {
    // 1. Определяем отправителя
    sourcePeer := s.p2pManager.GetPeer(peerID)
    if sourcePeer == nil {
        return fmt.Errorf("source peer not found: %s", peerID)
    }
    
    // 2. Парсим пакет для определения получателя
    targetPeerID := s.parseQUICPacketTarget(data)
    if targetPeerID == "" {
        return fmt.Errorf("cannot determine target peer from packet")
    }
    
    // 3. Получаем информацию о получателе
    targetPeer := s.p2pManager.GetPeer(targetPeerID)
    if targetPeer == nil {
        return fmt.Errorf("target peer not found: %s", targetPeerID)
    }
    
    // 4. Проверяем изоляцию tenant
    if sourcePeer.TenantID != targetPeer.TenantID {
        return fmt.Errorf("cross-tenant relay not allowed: %s -> %s", 
            sourcePeer.TenantID, targetPeer.TenantID)
    }
    
    // 5. Находим или создаем соединение
    connection := s.p2pManager.GetConnection(sourcePeer.PeerID, targetPeerID)
    if connection == nil {
        connection, err := s.p2pManager.EstablishP2PConnection(sourcePeer.PeerID, targetPeerID)
        if err != nil {
            return fmt.Errorf("failed to establish P2P connection: %v", err)
        }
    }
    
    // 6. Пересылаем пакет
    if targetPeer.QUICConn != nil {
        _, err := targetPeer.QUICConn.Write(data)
        if err != nil {
            return fmt.Errorf("failed to forward packet: %v", err)
        }
        
        // Обновляем статистику
        connection.BytesTransferred += int64(len(data))
        connection.LastActivity = time.Now()
        
        s.logger.Debugf("Packet relayed: %s -> %s (%d bytes)", sourcePeer.PeerID, targetPeerID, len(data))
    }
    
    return nil
}

func (s *Server) parseQUICPacketTarget(data []byte) string {
    // Простая реализация - в реальности нужно парсить QUIC заголовки
    // Для демонстрации используем первые 32 байта как peer ID
    if len(data) >= 32 {
        return string(data[:32])
    }
    return ""
}
```

### **Этап 2: Heartbeat система (Приоритет: ВЫСОКИЙ)**

#### **2.1 Heartbeat endpoint:**

```go
// internal/api/peer_handlers.go
func (h *PeerHandler) HandleHeartbeat(c *gin.Context) {
    tenantID := c.Param("tenant_id")
    peerID := c.Param("peer_id")
    
    // Обновляем LastSeen и Status
    peer := h.p2pManager.GetPeer(tenantID, peerID)
    if peer == nil {
        c.JSON(404, gin.H{"error": "peer not found"})
        return
    }
    
    peer.Mutex.Lock()
    peer.LastSeen = time.Now()
    peer.Status = "online"
    peer.Mutex.Unlock()
    
    h.logger.Debugf("Heartbeat received from peer %s in tenant %s", peerID, tenantID)
    
    c.JSON(200, gin.H{
        "status": "ok",
        "timestamp": time.Now().Unix(),
        "peer_status": "online",
    })
}
```

#### **2.2 Heartbeat клиент:**

```go
// internal/client/heartbeat.go
type HeartbeatClient struct {
    peerID      string
    tenantID    string
    apiURL      string
    jwtToken    string
    ticker      *time.Ticker
    logger      *logrus.Logger
}

func NewHeartbeatClient(peerID, tenantID, apiURL, jwtToken string, logger *logrus.Logger) *HeartbeatClient {
    return &HeartbeatClient{
        peerID:   peerID,
        tenantID: tenantID,
        apiURL:   apiURL,
        jwtToken: jwtToken,
        logger:   logger,
    }
}

func (hc *HeartbeatClient) Start() {
    hc.ticker = time.NewTicker(30 * time.Second)
    go func() {
        for range hc.ticker.C {
            hc.sendHeartbeat()
        }
    }()
}

func (hc *HeartbeatClient) sendHeartbeat() {
    url := fmt.Sprintf("%s/api/v1/tenants/%s/peers/%s/heartbeat", hc.apiURL, hc.tenantID, hc.peerID)
    
    req, err := http.NewRequest("POST", url, nil)
    if err != nil {
        hc.logger.Errorf("Failed to create heartbeat request: %v", err)
        return
    }
    
    req.Header.Set("Authorization", "Bearer "+hc.jwtToken)
    req.Header.Set("Content-Type", "application/json")
    
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        hc.logger.Errorf("Failed to send heartbeat: %v", err)
        return
    }
    defer resp.Body.Close()
    
    if resp.StatusCode == 200 {
        hc.logger.Debugf("Heartbeat sent successfully")
    } else {
        hc.logger.Errorf("Heartbeat failed with status: %d", resp.StatusCode)
    }
}

func (hc *HeartbeatClient) Stop() {
    if hc.ticker != nil {
        hc.ticker.Stop()
    }
}
```

#### **2.3 Cleanup офлайн пиров:**

```go
// internal/p2p/cleanup.go
func (pm *P2PManager) StartCleanupRoutine() {
    ticker := time.NewTicker(60 * time.Second)
    go func() {
        for range ticker.C {
            pm.cleanupOfflinePeers()
        }
    }()
}

func (pm *P2PManager) cleanupOfflinePeers() {
    pm.mutex.Lock()
    defer pm.mutex.Unlock()
    
    cutoff := time.Now().Add(-5 * time.Minute) // 5 минут без активности
    
    for _, tenant := range pm.tenants {
        for peerID, peer := range tenant.Peers {
            peer.Mutex.RLock()
            lastSeen := peer.LastSeen
            peer.Mutex.RUnlock()
            
            if lastSeen.Before(cutoff) {
                // Помечаем как офлайн
                peer.Mutex.Lock()
                peer.Status = "offline"
                peer.Mutex.Unlock()
                
                // Закрываем соединения
                for _, conn := range peer.Connections {
                    conn.Mutex.Lock()
                    conn.Status = "disconnected"
                    conn.Mutex.Unlock()
                }
                
                pm.logger.Infof("Peer marked as offline: %s in tenant %s", peerID, tenant.TenantID)
            }
        }
    }
}
```

### **Этап 3: ICE/STUN интеграция (Приоритет: СРЕДНИЙ)**

#### **3.1 ICE кандидаты обмен:**

```go
// internal/ice/candidates.go
type ICECandidate struct {
    Candidate     string `json:"candidate"`
    SDPMLineIndex int    `json:"sdpMLineIndex"`
    SDPMid        string `json:"sdpMid"`
}

type ICESession struct {
    SessionID     string         `json:"session_id"`
    TenantID      string         `json:"tenant_id"`
    SourcePeer    string         `json:"source_peer"`
    TargetPeer    string         `json:"target_peer"`
    Candidates    []ICECandidate `json:"candidates"`
    Status        string         `json:"status"`
    CreatedAt     time.Time      `json:"created_at"`
    Mutex         sync.RWMutex   `json:"-"`
}

func (s *Server) exchangeICECandidates(sourcePeer, targetPeer string, candidates []ICECandidate) error {
    // 1. Создаем ICE сессию
    session := &ICESession{
        SessionID:  s.generateSessionID(),
        TenantID:   s.getTenantByPeerID(sourcePeer),
        SourcePeer: sourcePeer,
        TargetPeer: targetPeer,
        Candidates: candidates,
        Status:     "exchanging",
        CreatedAt:  time.Now(),
    }
    
    // 2. Сохраняем сессию
    s.iceManager.AddSession(session)
    
    // 3. Отправляем кандидаты целевому пиру
    targetPeerInfo := s.p2pManager.GetPeer(targetPeer)
    if targetPeerInfo != nil && targetPeerInfo.QUICConn != nil {
        // Отправляем через QUIC соединение
        data, _ := json.Marshal(map[string]interface{}{
            "type": "ice_candidates",
            "session_id": session.SessionID,
            "candidates": candidates,
        })
        
        _, err := targetPeerInfo.QUICConn.Write(data)
        if err != nil {
            return fmt.Errorf("failed to send ICE candidates: %v", err)
        }
    }
    
    return nil
}
```

### **Этап 4: Мультитунельность (Приоритет: СРЕДНИЙ)**

#### **4.1 Tenant Manager:**

```go
// internal/tenant/manager.go
type TenantManager struct {
    tenants map[string]*Tenant
    mutex   sync.RWMutex
    logger  *logrus.Logger
}

func (tm *TenantManager) getTenantNetwork(tenantID string) string {
    // Каждый tenant получает свою подсеть
    tenantNetworks := map[string]string{
        "tenant-216420165": "10.100.77.0/24",
        "tenant-123456789": "10.100.78.0/24",
        "tenant-987654321": "10.100.79.0/24",
    }
    
    network, exists := tenantNetworks[tenantID]
    if !exists {
        // Генерируем новую подсеть для нового tenant
        network = tm.generateNewTenantNetwork(tenantID)
    }
    
    return network
}

func (tm *TenantManager) allocateIPForTenant(tenantID string) (string, error) {
    network := tm.getTenantNetwork(tenantID)
    
    // Парсим CIDR
    _, ipNet, err := net.ParseCIDR(network)
    if err != nil {
        return "", fmt.Errorf("invalid network CIDR: %v", err)
    }
    
    // Находим свободный IP в подсети
    for i := 10; i < 254; i++ { // Пропускаем .0, .1, .255
        ip := net.IPv4(ipNet.IP[0], ipNet.IP[1], ipNet.IP[2], byte(i))
        
        // Проверяем, не занят ли IP
        if !tm.isIPAllocated(tenantID, ip.String()) {
            return ip.String(), nil
        }
    }
    
    return "", fmt.Errorf("no free IP addresses in tenant network: %s", network)
}
```

## 🧪 **Тестирование P2P маршрутизации**

### **1. Тест базовой P2P связности:**

```bash
# Тест 1: Регистрация двух пиров в одном tenant
curl -X POST "http://localhost:8082/api/v1/tenants/tenant-216420165/peers/register" \
  -H "Authorization: Bearer $(cat test-token.txt)" \
  -H "Content-Type: application/json" \
  -d '{
    "public_key": "peer-1-key",
    "allowed_ips": ["10.100.77.10/32"],
    "peer_info": {"name": "Peer 1"}
  }'

curl -X POST "http://localhost:8082/api/v1/tenants/tenant-216420165/peers/register" \
  -H "Authorization: Bearer $(cat test-token.txt)" \
  -H "Content-Type: application/json" \
  -d '{
    "public_key": "peer-2-key",
    "allowed_ips": ["10.100.77.11/32"],
    "peer_info": {"name": "Peer 2"}
  }'

# Тест 2: Установка P2P соединения
curl -X POST "http://localhost:8082/api/v1/tenants/tenant-216420165/connections" \
  -H "Authorization: Bearer $(cat test-token.txt)" \
  -H "Content-Type: application/json" \
  -d '{
    "source_peer": "peer-1",
    "target_peer": "peer-2",
    "connection_type": "quic-ice"
  }'

# Тест 3: Отправка данных через QUIC relay
echo "Hello from Peer 1" | nc -u localhost 31284
```

### **2. Тест мультитунельности:**

```bash
# Тест 1: Регистрация пира в другом tenant
curl -X POST "http://localhost:8082/api/v1/tenants/tenant-123456789/peers/register" \
  -H "Authorization: Bearer $(cat test-token-b.txt)" \
  -H "Content-Type: application/json" \
  -d '{
    "public_key": "peer-b1-key",
    "allowed_ips": ["10.100.78.10/32"],
    "peer_info": {"name": "Peer B1"}
  }'

# Тест 2: Попытка cross-tenant соединения (должна быть заблокирована)
curl -X POST "http://localhost:8082/api/v1/tenants/tenant-216420165/connections" \
  -H "Authorization: Bearer $(cat test-token.txt)" \
  -H "Content-Type: application/json" \
  -d '{
    "source_peer": "peer-1",
    "target_peer": "peer-b1",
    "connection_type": "quic-ice"
  }'
# Ожидаемый результат: 403 Forbidden - cross-tenant connection not allowed
```

### **3. Тест heartbeat системы:**

```bash
# Тест 1: Отправка heartbeat
curl -X POST "http://localhost:8082/api/v1/tenants/tenant-216420165/peers/peer-1/heartbeat" \
  -H "Authorization: Bearer $(cat test-token.txt)"

# Тест 2: Проверка статуса пира
curl -X GET "http://localhost:8082/api/v1/tenants/tenant-216420165/peers/peer-1" \
  -H "Authorization: Bearer $(cat test-token.txt)"

# Ожидаемый результат: "status": "online", "last_seen": "2025-09-18T20:30:00Z"
```

## 📊 **Метрики успеха**

### **Критерии готовности:**

1. **P2P соединения:**
   - ✅ Клиенты могут общаться друг с другом через relay
   - ✅ QUIC пакеты пересылаются между пирами
   - ✅ Latency < 100ms для локальных соединений
   - ✅ Throughput > 1Mbps для P2P трафика

2. **Мультитунельность:**
   - ✅ Пиры из разных tenant не могут общаться
   - ✅ Изоляция на уровне JWT токенов
   - ✅ Отдельные сетевые пространства (10.100.77.0/24, 10.100.78.0/24)
   - ✅ Невозможность cross-tenant коммуникации

3. **Стабильность:**
   - ✅ Heartbeat система работает (30s интервал)
   - ✅ Автоматический cleanup офлайн пиров (5min timeout)
   - ✅ Graceful shutdown соединений
   - ✅ Восстановление соединений после сбоев

4. **Масштабируемость:**
   - ✅ Поддержка 100+ одновременных соединений
   - ✅ Memory usage < 500MB
   - ✅ CPU usage < 50%
   - ✅ Поддержка 10+ tenant одновременно

## 🚀 **Следующие шаги**

### **Для разработчика:**

1. **Немедленно (сегодня):**
   - Изучить текущий код в `cmd/relay/`
   - Понять структуру JWT токенов
   - Проанализировать существующие API endpoints
   - Создать базовые структуры данных для P2P

2. **Этап 1 (1-2 дня):**
   - Реализовать P2P Manager
   - Добавить базовую P2P маршрутизацию
   - Реализовать QUIC Relay Handler
   - Протестировать базовую P2P связность

3. **Этап 2 (1-2 дня):**
   - Реализовать heartbeat систему
   - Добавить cleanup офлайн пиров
   - Протестировать стабильность соединений

4. **Этап 3 (2-3 дня):**
   - Реализовать мультитунельность
   - Добавить ICE/STUN интеграцию
   - Протестировать изоляцию между tenant

5. **Этап 4 (1 день):**
   - Оптимизация производительности
   - Добавление метрик и мониторинга
   - Финальное тестирование

### **Для DevOps:**

1. **Подготовить тестовую среду:**
   - Настроить Minikube с новым кодом
   - Подготовить тестовые JWT токены для разных tenant
   - Настроить мониторинг P2P соединений

2. **Тестирование:**
   - Тестировать P2P соединения между разными серверами
   - Проверить изоляцию между tenant
   - Нагрузочное тестирование (100+ соединений)
   - Тестирование восстановления после сбоев

## 📞 **Поддержка и ресурсы**

- **Техническая документация:** `DEVELOPER_TECHNICAL_SPECIFICATION.md`
- **Архитектура мультитунельности:** `MULTI_TENANT_ARCHITECTURE.md`
- **Инструкция по сборке:** `MINIKUBE_BUILD_GUIDE.md`
- **Текущий статус:** Все порты работают, нужна доработка P2P маршрутизации

---

**🎯 Цель: Создать полноценную P2P mesh сеть, где удаленные серверы клиентов могут общаться как в локальной сети, с полной изоляцией между разными пользователями и поддержкой всех современных протоколов!**
