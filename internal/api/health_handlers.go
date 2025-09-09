package api

import (
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

		if peer.LatencyMS > 0 {
			totalLatency += peer.LatencyMS
			latencyCount++
		}

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
