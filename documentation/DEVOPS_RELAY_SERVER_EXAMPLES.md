# 💻 Примеры кода для DevOps Relay сервера

## 🔧 **1. Структуры данных**

### **Go структуры:**
```go
package main

import (
    "time"
    "encoding/json"
)

// Peer представляет зарегистрированный пир
type Peer struct {
    ID                string                 `json:"id" db:"id"`
    PeerID            string                 `json:"peer_id" db:"peer_id"`
    TenantID          string                 `json:"tenant_id" db:"tenant_id"`
    OrgID             string                 `json:"org_id" db:"org_id"`
    DeviceID          string                 `json:"device_id" db:"device_id"`
    ConnectionType    string                 `json:"connection_type" db:"connection_type"`
    WireGuardConfig   *WireGuardConfig       `json:"wireguard_config" db:"wireguard_config"`
    MeshConfig        *MeshConfig            `json:"mesh_config" db:"mesh_config"`
    PeerWhitelist     *PeerWhitelist         `json:"peer_whitelist" db:"peer_whitelist"`
    NetworkConfig     *NetworkConfig         `json:"network_config" db:"network_config"`
    IsOnline          bool                   `json:"is_online" db:"is_online"`
    LastSeen          time.Time              `json:"last_seen" db:"last_seen"`
    RegisteredAt      time.Time              `json:"registered_at" db:"registered_at"`
    ActiveConnections int                    `json:"active_connections" db:"active_connections"`
    MeshStatus        string                 `json:"mesh_status" db:"mesh_status"`
}

// WireGuardConfig конфигурация WireGuard
type WireGuardConfig struct {
    PublicKey   string   `json:"public_key"`
    Endpoint    string   `json:"endpoint"`
    ListenPort  int      `json:"listen_port"`
    AllowedIPs  []string `json:"allowed_ips"`
}

// MeshConfig конфигурация mesh сети
type MeshConfig struct {
    AutoDiscovery bool   `json:"auto_discovery"`
    Routing       string `json:"routing"`
    Encryption    string `json:"encryption"`
}

// PeerWhitelist список разрешенных пиров
type PeerWhitelist struct {
    AllowedPeers []string `json:"allowed_peers"`
    AutoApprove  bool     `json:"auto_approve"`
    MaxPeers     int      `json:"max_peers"`
}

// NetworkConfig сетевая конфигурация
type NetworkConfig struct {
    Subnet string   `json:"subnet"`
    DNS    []string `json:"dns"`
    MTU    int      `json:"mtu"`
}

// PeerConnection представляет соединение между пирами
type PeerConnection struct {
    ID                string    `json:"id" db:"id"`
    FromPeerID        string    `json:"from_peer_id" db:"from_peer_id"`
    ToPeerID          string    `json:"to_peer_id" db:"to_peer_id"`
    ConnectionType    string    `json:"connection_type" db:"connection_type"`
    IsActive          bool      `json:"is_active" db:"is_active"`
    EstablishedAt     time.Time `json:"established_at" db:"established_at"`
    LastActivity      time.Time `json:"last_activity" db:"last_activity"`
    BytesTransferred  int64     `json:"bytes_transferred" db:"bytes_transferred"`
    LatencyMS         int       `json:"latency_ms" db:"latency_ms"`
}

// P2PClaims JWT claims для P2P токенов
type P2PClaims struct {
    JTI              string           `json:"jti"`
    Subject          string           `json:"sub"`
    OrgID            string           `json:"org_id"`
    TenantID         string           `json:"tenant_id"`
    DeviceID         string           `json:"device_id"`
    DeviceFingerprint string          `json:"device_fingerprint"`
    ConnectionType   string           `json:"connection_type"`
    Permissions      []string         `json:"permissions"`
    WireGuardConfig  *WireGuardConfig `json:"wireguard_config"`
    MeshConfig       *MeshConfig      `json:"mesh_config"`
    PeerWhitelist    *PeerWhitelist   `json:"peer_whitelist"`
    NetworkConfig    *NetworkConfig   `json:"network_config"`
    IssuedAt         int64            `json:"iat"`
    NotBefore        int64            `json:"nbf"`
    Issuer           string           `json:"iss"`
    Audience         string           `json:"aud"`
    jwt.RegisteredClaims
}
```

## 🌐 **2. API Handlers**

### **Peer Registration:**
```go
package main

import (
    "net/http"
    "encoding/json"
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

// RegisterPeer регистрирует новый пир
func RegisterPeer(c *gin.Context) {
    var req struct {
        PeerID          string           `json:"peer_id" binding:"required"`
        TenantID        string           `json:"tenant_id" binding:"required"`
        OrgID           string           `json:"org_id" binding:"required"`
        DeviceID        string           `json:"device_id" binding:"required"`
        ConnectionType  string           `json:"connection_type" binding:"required"`
        WireGuardConfig *WireGuardConfig `json:"wireguard_config" binding:"required"`
        MeshConfig      *MeshConfig      `json:"mesh_config"`
        PeerWhitelist   *PeerWhitelist   `json:"peer_whitelist"`
        NetworkConfig   *NetworkConfig   `json:"network_config"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // Валидация JWT токена
    token := c.GetHeader("Authorization")
    if token == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
        return
    }
    
    claims, err := validateP2PToken(token)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
        return
    }
    
    // Проверка прав доступа
    if claims.TenantID != req.TenantID {
        c.JSON(http.StatusForbidden, gin.H{"error": "tenant mismatch"})
        return
    }
    
    // Создание пира
    peer := &Peer{
        PeerID:          req.PeerID,
        TenantID:        req.TenantID,
        OrgID:           req.OrgID,
        DeviceID:        req.DeviceID,
        ConnectionType:  req.ConnectionType,
        WireGuardConfig: req.WireGuardConfig,
        MeshConfig:      req.MeshConfig,
        PeerWhitelist:   req.PeerWhitelist,
        NetworkConfig:   req.NetworkConfig,
        IsOnline:        true,
        LastSeen:        time.Now(),
        RegisteredAt:    time.Now(),
        MeshStatus:      "connected",
    }
    
    // Сохранение в базу данных
    if err := db.CreatePeer(peer); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create peer"})
        return
    }
    
    // Уведомление других пиров через WebSocket
    notifyPeers(peer.TenantID, "peer_joined", peer)
    
    c.JSON(http.StatusOK, gin.H{
        "success":           true,
        "peer_id":          peer.PeerID,
        "registered_at":    peer.RegisteredAt,
        "discovery_endpoint": "wss://relay.2gc.ru/ws/peers",
    })
}
```

### **Peer Discovery:**
```go
// DiscoverPeers возвращает список пиров для тенанта
func DiscoverPeers(c *gin.Context) {
    tenantID := c.Query("tenant_id")
    if tenantID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "missing tenant_id"})
        return
    }
    
    // Валидация JWT токена
    token := c.GetHeader("Authorization")
    claims, err := validateP2PToken(token)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
        return
    }
    
    // Проверка прав доступа
    if claims.TenantID != tenantID {
        c.JSON(http.StatusForbidden, gin.H{"error": "tenant mismatch"})
        return
    }
    
    // Получение пиров из базы данных
    peers, err := db.GetPeersByTenant(tenantID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get peers"})
        return
    }
    
    // Фильтрация по whitelist
    filteredPeers := filterPeersByWhitelist(peers, claims.PeerWhitelist)
    
    // Обновление статуса онлайн
    for _, peer := range filteredPeers {
        peer.IsOnline = isPeerOnline(peer.PeerID)
        peer.LastSeen = getLastSeen(peer.PeerID)
    }
    
    c.JSON(http.StatusOK, gin.H{
        "success":   true,
        "tenant_id": tenantID,
        "peers":     filteredPeers,
    })
}
```

### **Peer Status Update:**
```go
// UpdatePeerStatus обновляет статус пира
func UpdatePeerStatus(c *gin.Context) {
    peerID := c.Param("peer_id")
    
    var req struct {
        IsOnline         bool `json:"is_online"`
        LatencyMS        int  `json:"latency_ms"`
        ActiveConnections int `json:"active_connections"`
        MeshStatus       string `json:"mesh_status"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // Валидация JWT токена
    token := c.GetHeader("Authorization")
    claims, err := validateP2PToken(token)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
        return
    }
    
    // Обновление статуса в базе данных
    if err := db.UpdatePeerStatus(peerID, req.IsOnline, req.LatencyMS, req.ActiveConnections, req.MeshStatus); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update status"})
        return
    }
    
    // Уведомление других пиров
    notifyPeers(claims.TenantID, "peer_status_update", gin.H{
        "peer_id":          peerID,
        "is_online":        req.IsOnline,
        "latency_ms":       req.LatencyMS,
        "mesh_status":      req.MeshStatus,
        "timestamp":        time.Now(),
    })
    
    c.JSON(http.StatusOK, gin.H{
        "success":    true,
        "updated_at": time.Now(),
    })
}
```

## 🔌 **3. WebSocket Server**

### **WebSocket Handler:**
```go
package main

import (
    "net/http"
    "github.com/gorilla/websocket"
    "github.com/gin-gonic/gin"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // В продакшене нужно добавить проверку origin
    },
}

// WebSocketConnection представляет WebSocket соединение
type WebSocketConnection struct {
    Conn     *websocket.Conn
    PeerID   string
    TenantID string
    Send     chan []byte
    Hub      *Hub
}

// Hub управляет WebSocket соединениями
type Hub struct {
    Connections map[string]*WebSocketConnection
    Register    chan *WebSocketConnection
    Unregister  chan *WebSocketConnection
    Broadcast   chan []byte
}

// NewHub создает новый Hub
func NewHub() *Hub {
    return &Hub{
        Connections: make(map[string]*WebSocketConnection),
        Register:    make(chan *WebSocketConnection),
        Unregister:  make(chan *WebSocketConnection),
        Broadcast:   make(chan []byte),
    }
}

// Run запускает Hub
func (h *Hub) Run() {
    for {
        select {
        case conn := <-h.Register:
            h.Connections[conn.PeerID] = conn
            log.Printf("Peer %s connected to WebSocket", conn.PeerID)
            
        case conn := <-h.Unregister:
            if _, ok := h.Connections[conn.PeerID]; ok {
                delete(h.Connections, conn.PeerID)
                close(conn.Send)
                log.Printf("Peer %s disconnected from WebSocket", conn.PeerID)
            }
            
        case message := <-h.Broadcast:
            for _, conn := range h.Connections {
                select {
                case conn.Send <- message:
                default:
                    close(conn.Send)
                    delete(h.Connections, conn.PeerID)
                }
            }
        }
    }
}

// HandleWebSocket обрабатывает WebSocket соединения
func HandleWebSocket(c *gin.Context) {
    // Валидация JWT токена
    token := c.Query("token")
    if token == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
        return
    }
    
    claims, err := validateP2PToken("Bearer " + token)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
        return
    }
    
    // Обновление до WebSocket
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        log.Printf("WebSocket upgrade error: %v", err)
        return
    }
    
    // Создание WebSocket соединения
    wsConn := &WebSocketConnection{
        Conn:     conn,
        PeerID:   claims.Subject,
        TenantID: claims.TenantID,
        Send:     make(chan []byte, 256),
        Hub:      hub,
    }
    
    // Регистрация в Hub
    hub.Register <- wsConn
    
    // Запуск горутин для чтения и записи
    go wsConn.writePump()
    go wsConn.readPump()
}

// writePump отправляет сообщения клиенту
func (c *WebSocketConnection) writePump() {
    ticker := time.NewTicker(54 * time.Second)
    defer func() {
        ticker.Stop()
        c.Conn.Close()
    }()
    
    for {
        select {
        case message, ok := <-c.Send:
            c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
            if !ok {
                c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }
            
            w, err := c.Conn.NextWriter(websocket.TextMessage)
            if err != nil {
                return
            }
            w.Write(message)
            
            // Отправка дополнительных сообщений
            n := len(c.Send)
            for i := 0; i < n; i++ {
                w.Write([]byte{'\n'})
                w.Write(<-c.Send)
            }
            
            if err := w.Close(); err != nil {
                return
            }
            
        case <-ticker.C:
            c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
            if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}

// readPump читает сообщения от клиента
func (c *WebSocketConnection) readPump() {
    defer func() {
        c.Hub.Unregister <- c
        c.Conn.Close()
    }()
    
    c.Conn.SetReadLimit(512)
    c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
    c.Conn.SetPongHandler(func(string) error {
        c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
        return nil
    })
    
    for {
        _, _, err := c.Conn.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                log.Printf("WebSocket error: %v", err)
            }
            break
        }
    }
}
```

## 🔄 **4. Fallback Routing**

### **Relay Routing:**
```go
// RouteThroughRelay маршрутизирует трафик через relay
func RouteThroughRelay(c *gin.Context) {
    var req struct {
        FromPeer    string `json:"from_peer" binding:"required"`
        ToPeer      string `json:"to_peer" binding:"required"`
        Data        string `json:"data" binding:"required"`
        MessageType string `json:"message_type" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // Валидация JWT токена
    token := c.GetHeader("Authorization")
    claims, err := validateP2PToken(token)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
        return
    }
    
    // Проверка существования пиров
    fromPeer, err := db.GetPeer(req.FromPeer)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "from peer not found"})
        return
    }
    
    toPeer, err := db.GetPeer(req.ToPeer)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "to peer not found"})
        return
    }
    
    // Проверка прав доступа
    if fromPeer.TenantID != claims.TenantID || toPeer.TenantID != claims.TenantID {
        c.JSON(http.StatusForbidden, gin.H{"error": "tenant mismatch"})
        return
    }
    
    // Проверка whitelist
    if !validatePeerAccess(req.ToPeer, fromPeer.PeerWhitelist) {
        c.JSON(http.StatusForbidden, gin.H{"error": "peer not in whitelist"})
        return
    }
    
    // Маршрутизация через WebSocket
    startTime := time.Now()
    success := routeToPeer(req.ToPeer, req.Data, req.MessageType)
    latency := time.Since(startTime)
    
    if !success {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "routing failed"})
        return
    }
    
    // Обновление метрик
    updateRoutingMetrics(req.FromPeer, req.ToPeer, len(req.Data), latency)
    
    c.JSON(http.StatusOK, gin.H{
        "success":    true,
        "routed_at":  time.Now(),
        "latency_ms": latency.Milliseconds(),
    })
}

// routeToPeer отправляет данные пиру через WebSocket
func routeToPeer(peerID string, data string, messageType string) bool {
    conn, exists := hub.Connections[peerID]
    if !exists {
        return false
    }
    
    message := gin.H{
        "type":         "relay_message",
        "message_type": messageType,
        "data":         data,
        "timestamp":    time.Now(),
    }
    
    messageBytes, err := json.Marshal(message)
    if err != nil {
        return false
    }
    
    select {
    case conn.Send <- messageBytes:
        return true
    default:
        return false
    }
}
```

## 🔍 **5. Валидация токенов**

### **JWT Validation:**
```go
import (
    "github.com/golang-jwt/jwt/v5"
    "time"
)

// validateP2PToken валидирует P2P JWT токен
func validateP2PToken(tokenString string) (*P2PClaims, error) {
    // Удаление "Bearer " префикса
    if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
        tokenString = tokenString[7:]
    }
    
    // Парсинг токена
    token, err := jwt.ParseWithClaims(tokenString, &P2PClaims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(os.Getenv("JWT_SECRET")), nil
    })
    
    if err != nil {
        return nil, err
    }
    
    claims, ok := token.Claims.(*P2PClaims)
    if !ok || !token.Valid {
        return nil, errors.New("invalid token")
    }
    
    // Проверка обязательных полей
    if claims.ConnectionType != "p2p-mesh" {
        return nil, errors.New("invalid connection type")
    }
    
    if claims.DeviceID == "" {
        return nil, errors.New("missing device_id")
    }
    
    if claims.WireGuardConfig == nil {
        return nil, errors.New("missing wireguard_config")
    }
    
    if claims.WireGuardConfig.PublicKey == "" {
        return nil, errors.New("missing wireguard public key")
    }
    
    // Проверка отзыва токена
    if isTokenRevoked(claims.JTI) {
        return nil, errors.New("token revoked")
    }
    
    return claims, nil
}

// isTokenRevoked проверяет, отозван ли токен
func isTokenRevoked(jti string) bool {
    // Проверка в базе данных или Redis
    revoked, err := db.IsTokenRevoked(jti)
    if err != nil {
        log.Printf("Error checking token revocation: %v", err)
        return false
    }
    return revoked
}

// validatePeerAccess проверяет доступ пира
func validatePeerAccess(peerID string, whitelist *PeerWhitelist) bool {
    if whitelist == nil {
        return true // Нет whitelist - все разрешено
    }
    
    if whitelist.AutoApprove {
        return true
    }
    
    for _, allowedPeer := range whitelist.AllowedPeers {
        if allowedPeer == peerID {
            return true
        }
    }
    
    return false
}
```

## 📊 **6. Мониторинг и метрики**

### **Prometheus метрики:**
```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    peersTotal = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "cloudbridge_peers_total",
            Help: "Total number of registered peers",
        },
        []string{"tenant_id", "connection_type"},
    )
    
    peersOnline = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "cloudbridge_peers_online",
            Help: "Number of online peers",
        },
        []string{"tenant_id"},
    )
    
    peerConnections = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "cloudbridge_peer_connections",
            Help: "Number of active peer connections",
        },
        []string{"tenant_id", "connection_type"},
    )
    
    peerLatency = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "cloudbridge_peer_latency_seconds",
            Help: "Peer connection latency",
            Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
        },
        []string{"tenant_id", "peer_id"},
    )
    
    relayTraffic = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "cloudbridge_relay_traffic_bytes_total",
            Help: "Total traffic routed through relay",
        },
        []string{"tenant_id", "direction"},
    )
)

// updateRoutingMetrics обновляет метрики маршрутизации
func updateRoutingMetrics(fromPeer, toPeer string, dataSize int, latency time.Duration) {
    relayTraffic.WithLabelValues("inbound").Add(float64(dataSize))
    relayTraffic.WithLabelValues("outbound").Add(float64(dataSize))
    peerLatency.WithLabelValues("", fromPeer).Observe(latency.Seconds())
}
```

### **Health Check:**
```go
// HealthCheck возвращает статус здоровья сервиса
func HealthCheck(c *gin.Context) {
    // Проверка базы данных
    dbStatus := "healthy"
    if err := db.Ping(); err != nil {
        dbStatus = "unhealthy"
    }
    
    // Проверка Redis
    redisStatus := "healthy"
    if err := redis.Ping(); err != nil {
        redisStatus = "unhealthy"
    }
    
    // Проверка WebSocket
    wsStatus := "healthy"
    if len(hub.Connections) > 1000 {
        wsStatus = "warning"
    }
    
    // Получение метрик
    totalPeers, _ := db.GetTotalPeers()
    onlinePeers, _ := db.GetOnlinePeers()
    activeConnections, _ := db.GetActiveConnections()
    
    status := "healthy"
    if dbStatus != "healthy" || redisStatus != "healthy" {
        status = "unhealthy"
    }
    
    c.JSON(http.StatusOK, gin.H{
        "status":    status,
        "timestamp": time.Now(),
        "services": gin.H{
            "database":  dbStatus,
            "redis":     redisStatus,
            "websocket": wsStatus,
        },
        "metrics": gin.H{
            "total_peers":        totalPeers,
            "online_peers":       onlinePeers,
            "active_connections": activeConnections,
            "uptime_seconds":     time.Since(startTime).Seconds(),
        },
    })
}
```

## 🧪 **7. Тестирование**

### **Unit тесты:**
```go
package main

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestValidateP2PToken(t *testing.T) {
    // Создание тестового токена
    token := createTestP2PToken()
    
    // Валидация токена
    claims, err := validateP2PToken(token)
    
    assert.NoError(t, err)
    assert.Equal(t, "p2p-mesh", claims.ConnectionType)
    assert.Equal(t, "test-device-123", claims.DeviceID)
    assert.NotNil(t, claims.WireGuardConfig)
}

func TestValidatePeerAccess(t *testing.T) {
    whitelist := &PeerWhitelist{
        AllowedPeers: []string{"peer1", "peer2"},
        AutoApprove:  false,
        MaxPeers:     10,
    }
    
    // Тест разрешенного пира
    assert.True(t, validatePeerAccess("peer1", whitelist))
    
    // Тест запрещенного пира
    assert.False(t, validatePeerAccess("peer3", whitelist))
    
    // Тест auto approve
    whitelist.AutoApprove = true
    assert.True(t, validatePeerAccess("peer3", whitelist))
}

func TestPeerRegistration(t *testing.T) {
    // Создание тестового сервера
    server := setupTestServer()
    
    // Тестовые данные
    peerData := gin.H{
        "peer_id":          "test-peer-123",
        "tenant_id":        "test-tenant-456",
        "org_id":           "test-org-789",
        "device_id":        "test-device-123",
        "connection_type":  "p2p-mesh",
        "wireguard_config": gin.H{
            "public_key":  "test-public-key",
            "endpoint":    "test.example.com:51820",
            "listen_port": 51820,
            "allowed_ips": []string{"10.0.1.1/32"},
        },
    }
    
    // Отправка запроса
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("POST", "/api/v1/peers/register", jsonBody(peerData))
    req.Header.Set("Authorization", "Bearer "+createTestJWT())
    
    server.ServeHTTP(w, req)
    
    // Проверка ответа
    assert.Equal(t, http.StatusOK, w.Code)
    
    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.True(t, response["success"].(bool))
    assert.Equal(t, "test-peer-123", response["peer_id"])
}
```

---

**Эти примеры кода помогут DevOps команде быстро реализовать P2P функциональность на Relay сервере!** 🚀
