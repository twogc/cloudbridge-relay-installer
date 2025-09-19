package api

import (
	"encoding/json"
	"io"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// HealthManager управляет health check и мониторингом
type HealthManager struct {
	hub         *Hub
	peers       map[string]*Peer
	connections map[string]*PeerConnection
	logger      *zap.Logger
	startTime   time.Time
}

// NewHealthManager создает новый HealthManager
func NewHealthManager(hub *Hub, peers map[string]*Peer, connections map[string]*PeerConnection, logger *zap.Logger) *HealthManager {
	return &HealthManager{
		hub:         hub,
		peers:       peers,
		connections: connections,
		logger:      logger,
		startTime:   time.Now(),
	}
}

// HealthCheckResponse ответ health check
type HealthCheckResponse struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Services  map[string]string      `json:"services"`
	Metrics   map[string]interface{} `json:"metrics"`
	Version   string                 `json:"version"`
	Uptime    string                 `json:"uptime"`
}

// HealthCheck возвращает статус здоровья сервиса
func (hm *HealthManager) HealthCheck(c *gin.Context) {
	// Проверка базы данных (в реальной реализации)
	dbStatus := "healthy"
	// if err := db.Ping(); err != nil {
	//     dbStatus = "unhealthy"
	// }

	// Проверка Redis (в реальной реализации)
	redisStatus := "healthy"
	// if err := redis.Ping(); err != nil {
	//     redisStatus = "unhealthy"
	// }

	// Проверка WebSocket
	wsStatus := "healthy"
	connectionCount := hm.hub.GetConnectionCount()
	if connectionCount > 1000 {
		wsStatus = "warning"
	}

	// Получение метрик
	totalPeers := len(hm.peers)
	onlinePeers := hm.getOnlinePeersCount()
	activeConnections := hm.getActiveConnectionsCount()

	status := "healthy"
	if dbStatus != "healthy" || redisStatus != "healthy" {
		status = "unhealthy"
	}

	uptime := time.Since(hm.startTime)

	response := HealthCheckResponse{
		Status:    status,
		Timestamp: time.Now(),
		Services: map[string]string{
			"database":  dbStatus,
			"redis":     redisStatus,
			"websocket": wsStatus,
		},
		Metrics: map[string]interface{}{
			"total_peers":           totalPeers,
			"online_peers":          onlinePeers,
			"active_connections":    activeConnections,
			"websocket_connections": connectionCount,
			"uptime_seconds":        uptime.Seconds(),
			"memory_usage_mb":       hm.getMemoryUsage(),
			"goroutines":            runtime.NumGoroutine(),
		},
		Version: "v1.0.0",
		Uptime:  uptime.String(),
	}

	hm.logger.Debug("Health check requested",
		zap.String("status", status),
		zap.Int("total_peers", totalPeers),
		zap.Int("online_peers", onlinePeers),
		zap.Int("active_connections", activeConnections),
	)

	c.JSON(http.StatusOK, response)
}

// GetSystemMetrics возвращает системные метрики
func (hm *HealthManager) GetSystemMetrics(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metrics := map[string]interface{}{
		"memory": map[string]interface{}{
			"alloc_mb":        bToMb(m.Alloc),
			"total_alloc_mb":  bToMb(m.TotalAlloc),
			"sys_mb":          bToMb(m.Sys),
			"num_gc":          m.NumGC,
			"gc_cpu_fraction": m.GCCPUFraction,
		},
		"runtime": map[string]interface{}{
			"goroutines": runtime.NumGoroutine(),
			"cpus":       runtime.NumCPU(),
			"go_version": runtime.Version(),
		},
		"peers": map[string]interface{}{
			"total":  len(hm.peers),
			"online": hm.getOnlinePeersCount(),
		},
		"connections": map[string]interface{}{
			"total":  len(hm.connections),
			"active": hm.getActiveConnectionsCount(),
		},
		"websocket": map[string]interface{}{
			"connections": hm.hub.GetConnectionCount(),
		},
		"uptime": time.Since(hm.startTime).Seconds(),
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"metrics":   metrics,
		"timestamp": time.Now(),
	})
}

// GetPeerMetrics возвращает метрики пиров
func (hm *HealthManager) GetPeerMetrics(c *gin.Context) {
	tenantID := c.Query("tenant_id")

	var peers []*Peer
	if tenantID != "" {
		// Фильтрация по тенанту
		for _, peer := range hm.peers {
			if peer.TenantID == tenantID {
				peers = append(peers, peer)
			}
		}
	} else {
		// Все пиры
		for _, peer := range hm.peers {
			peers = append(peers, peer)
		}
	}

	// Подсчет метрик
	metrics := map[string]interface{}{
		"total_peers":      len(peers),
		"online_peers":     0,
		"offline_peers":    0,
		"avg_latency_ms":   0,
		"connection_types": make(map[string]int),
		"mesh_statuses":    make(map[string]int),
	}

	totalLatency := 0
	latencyCount := 0

	for _, peer := range peers {
		if peer.IsOnline {
			metrics["online_peers"] = metrics["online_peers"].(int) + 1
		} else {
			metrics["offline_peers"] = metrics["offline_peers"].(int) + 1
		}

		// LatencyMS removed from simplified Peer structure

		// Подсчет типов соединений
		connType := peer.ConnectionType
		if connType == "" {
			connType = "unknown"
		}
		metrics["connection_types"].(map[string]int)[connType]++

		// Подсчет статусов mesh
		meshStatus := peer.MeshStatus
		if meshStatus == "" {
			meshStatus = "unknown"
		}
		metrics["mesh_statuses"].(map[string]int)[meshStatus]++
	}

	if latencyCount > 0 {
		metrics["avg_latency_ms"] = totalLatency / latencyCount
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"tenant_id": tenantID,
		"metrics":   metrics,
		"timestamp": time.Now(),
	})
}

// GetConnectionMetrics возвращает метрики соединений
func (hm *HealthManager) GetConnectionMetrics(c *gin.Context) {
	tenantID := c.Query("tenant_id")

	var connections []*PeerConnection
	if tenantID != "" {
		// Фильтрация по тенанту
		for _, conn := range hm.connections {
			fromPeer, exists := hm.peers[conn.FromPeerID]
			if exists && fromPeer.TenantID == tenantID {
				connections = append(connections, conn)
			}
		}
	} else {
		// Все соединения
		for _, conn := range hm.connections {
			connections = append(connections, conn)
		}
	}

	// Подсчет метрик
	metrics := map[string]interface{}{
		"total_connections":    len(connections),
		"active_connections":   0,
		"inactive_connections": 0,
		"total_bytes":          int64(0),
		"avg_latency_ms":       0,
		"connection_types":     make(map[string]int),
	}

	totalLatency := 0
	latencyCount := 0

	for _, conn := range connections {
		if conn.IsActive {
			metrics["active_connections"] = metrics["active_connections"].(int) + 1
		} else {
			metrics["inactive_connections"] = metrics["inactive_connections"].(int) + 1
		}

		metrics["total_bytes"] = metrics["total_bytes"].(int64) + conn.BytesTransferred

		if conn.LatencyMS > 0 {
			totalLatency += conn.LatencyMS
			latencyCount++
		}

		// Подсчет типов соединений
		connType := conn.ConnectionType
		if connType == "" {
			connType = "unknown"
		}
		metrics["connection_types"].(map[string]int)[connType]++
	}

	if latencyCount > 0 {
		metrics["avg_latency_ms"] = totalLatency / latencyCount
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"tenant_id": tenantID,
		"metrics":   metrics,
		"timestamp": time.Now(),
	})
}

// GetRelayStatus proxies relay status JSON from the relay's local HTTP endpoint (port 9091).
// It exposes a stable JSON for dashboards on port 30082.
func (hm *HealthManager) GetRelayStatus(c *gin.Context) {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("http://127.0.0.1:9091/relay/status")
	if err == nil && resp != nil && resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		// Merge in-memory API counters (logical connections) into relay JSON
		type relayJSON struct {
			Connections struct {
				Active int64 `json:"active"`
				Total  int64 `json:"total"`
			} `json:"connections"`
			Runtime map[string]interface{} `json:"runtime"`
			Uptime  interface{}            `json:"uptime"`
		}
		var rj relayJSON
		if err := json.Unmarshal(body, &rj); err == nil {
			active := int64(hm.getActiveConnectionsCount())
			total := int64(len(hm.connections))
			if active > rj.Connections.Active {
				rj.Connections.Active = active
			}
			if total > rj.Connections.Total {
				rj.Connections.Total = total
			}
			out, _ := json.Marshal(rj)
			c.Data(http.StatusOK, "application/json", out)
			return
		}
		c.Data(http.StatusOK, "application/json", body)
		return
	}

	// Fallback: synthesize minimal JSON from in-memory maps if relay endpoint is unavailable
	uptime := 0.0
	// We don't have the relay start time here; return 0 as a safe default
	c.JSON(http.StatusOK, gin.H{
		"connections": gin.H{
			"active": hm.getActiveConnectionsCount(),
			"total":  len(hm.connections),
		},
		"peers": gin.H{
			"online": hm.getOnlinePeersCount(),
			"total":  len(hm.peers),
		},
		"runtime": gin.H{
			"goroutines": runtime.NumGoroutine(),
		},
		"uptime":   uptime,
		"fallback": true,
	})
}

// ConnectionEventRequest represents open/close logical connection event
type ConnectionEventRequest struct {
	TenantID  string                 `json:"tenant_id" binding:"required"`
	PeerID    string                 `json:"peer_id" binding:"required"`
	SessionID string                 `json:"session_id" binding:"required"`
	Protocol  string                 `json:"protocol"`
	Meta      map[string]interface{} `json:"meta"`
}

// OpenConnection marks a logical connection as active (idempotent)
func (hm *HealthManager) OpenConnection(c *gin.Context) {
	var req ConnectionEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		hm.logger.Error("Invalid open connection request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Use SessionID as connection ID key
	if req.SessionID == "" || req.PeerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id and peer_id required"})
		return
	}

	// Create or update
	if _, exists := hm.connections[req.SessionID]; !exists {
		hm.connections[req.SessionID] = &PeerConnection{
			ID:               req.SessionID,
			FromPeerID:       req.PeerID,
			ToPeerID:         "",
			ConnectionType:   req.Protocol,
			IsActive:         true,
			EstablishedAt:    time.Now(),
			LastActivity:     time.Now(),
			BytesTransferred: 0,
			LatencyMS:        0,
		}
	} else {
		hm.connections[req.SessionID].IsActive = true
		hm.connections[req.SessionID].LastActivity = time.Now()
	}

	hm.logger.Info("Logical connection opened",
		zap.String("session_id", req.SessionID),
		zap.String("peer_id", req.PeerID),
		zap.String("tenant_id", req.TenantID),
	)

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// CloseConnection marks a logical connection as inactive (idempotent)
func (hm *HealthManager) CloseConnection(c *gin.Context) {
	var req ConnectionEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		hm.logger.Error("Invalid close connection request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.SessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id required"})
		return
	}

	if conn, ok := hm.connections[req.SessionID]; ok {
		conn.IsActive = false
		conn.LastActivity = time.Now()
		hm.logger.Info("Logical connection closed", zap.String("session_id", req.SessionID))
	} else {
		hm.logger.Info("Close called for unknown session", zap.String("session_id", req.SessionID))
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// getOnlinePeersCount возвращает количество онлайн пиров
func (hm *HealthManager) getOnlinePeersCount() int {
	count := 0
	for _, peer := range hm.peers {
		if peer.IsOnline {
			count++
		}
	}
	return count
}

// getActiveConnectionsCount возвращает количество активных соединений
func (hm *HealthManager) getActiveConnectionsCount() int {
	count := 0
	for _, conn := range hm.connections {
		if conn.IsActive {
			count++
		}
	}
	return count
}

// getMemoryUsage возвращает использование памяти в MB
func (hm *HealthManager) getMemoryUsage() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return bToMb(m.Alloc)
}

// bToMb конвертирует байты в мегабайты
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
