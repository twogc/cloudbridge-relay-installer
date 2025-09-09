package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RelayManager управляет fallback маршрутизацией
type RelayManager struct {
	hub         *Hub
	peers       map[string]*Peer
	connections map[string]*PeerConnection
	logger      *zap.Logger
	validator   TokenValidator
}

// NewRelayManager создает новый RelayManager
func NewRelayManager(hub *Hub, peers map[string]*Peer, connections map[string]*PeerConnection, logger *zap.Logger, validator TokenValidator) *RelayManager {
	return &RelayManager{
		hub:         hub,
		peers:       peers,
		connections: connections,
		logger:      logger,
		validator:   validator,
	}
}

// RouteThroughRelay маршрутизирует трафик через relay
func (rm *RelayManager) RouteThroughRelay(c *gin.Context) {
	var req RelayRouteRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		rm.logger.Error("Failed to bind relay route request", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request format",
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	// Валидация JWT токена
	token := c.GetHeader("Authorization")
	claims, err := rm.validator.ValidateP2PToken(token)
	if err != nil {
		rm.logger.Error("Failed to validate P2P token", zap.Error(err))
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Invalid token",
			Code:    http.StatusUnauthorized,
			Message: err.Error(),
		})
		return
	}

	// Проверка существования пиров
	fromPeer, exists := rm.peers[req.FromPeer]
	if !exists {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "From peer not found",
			Code:    http.StatusNotFound,
			Message: "Source peer does not exist",
		})
		return
	}

	toPeer, exists := rm.peers[req.ToPeer]
	if !exists {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "To peer not found",
			Code:    http.StatusNotFound,
			Message: "Destination peer does not exist",
		})
		return
	}

	// Проверка прав доступа
	if fromPeer.TenantID != claims.TenantID || toPeer.TenantID != claims.TenantID {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "Tenant mismatch",
			Code:    http.StatusForbidden,
			Message: "Peers must belong to the same tenant",
		})
		return
	}

	// Проверка whitelist
	if !rm.validatePeerAccess(req.ToPeer, fromPeer.PeerWhitelist) {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "Peer not in whitelist",
			Code:    http.StatusForbidden,
			Message: "Destination peer is not in source peer's whitelist",
		})
		return
	}

	// Маршрутизация через WebSocket
	startTime := time.Now()
	success := rm.routeToPeer(req.ToPeer, req.Data, req.MessageType)
	latency := time.Since(startTime)

	if !success {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Routing failed",
			Code:    http.StatusInternalServerError,
			Message: "Failed to route message to destination peer",
		})
		return
	}

	// Обновление метрик
	rm.updateRoutingMetrics(req.FromPeer, req.ToPeer, len(req.Data), latency)

	rm.logger.Info("Message routed through relay",
		zap.String("from_peer", req.FromPeer),
		zap.String("to_peer", req.ToPeer),
		zap.Int("data_size", len(req.Data)),
		zap.Duration("latency", latency),
	)

	c.JSON(http.StatusOK, RelayRouteResponse{
		Success:   true,
		RoutedAt:  time.Now(),
		LatencyMS: latency.Milliseconds(),
	})
}

// GetRelayStats возвращает статистику маршрутизации
func (rm *RelayManager) GetRelayStats(c *gin.Context) {
	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing tenant_id",
			Code:    http.StatusBadRequest,
			Message: "tenant_id query parameter is required",
		})
		return
	}

	// Валидация JWT токена
	token := c.GetHeader("Authorization")
	claims, err := rm.validator.ValidateP2PToken(token)
	if err != nil {
		rm.logger.Error("Failed to validate P2P token", zap.Error(err))
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Invalid token",
			Code:    http.StatusUnauthorized,
			Message: err.Error(),
		})
		return
	}

	// Проверка прав доступа
	if claims.TenantID != tenantID {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "Tenant mismatch",
			Code:    http.StatusForbidden,
			Message: "Token tenant does not match request tenant",
		})
		return
	}

	// Подсчет статистики
	stats := rm.calculateStats(tenantID)

	rm.logger.Info("Relay stats requested",
		zap.String("tenant_id", tenantID),
		zap.Int("total_peers", stats.TotalPeers),
		zap.Int("online_peers", stats.OnlinePeers),
	)

	c.JSON(http.StatusOK, RelayStatsResponse{
		Success:  true,
		TenantID: tenantID,
		Stats:    stats,
	})
}

// routeToPeer отправляет данные пиру через WebSocket
func (rm *RelayManager) routeToPeer(peerID string, data string, messageType string) bool {
	return rm.hub.SendRelayMessage(peerID, messageType, data)
}

// validatePeerAccess проверяет доступ пира
func (rm *RelayManager) validatePeerAccess(peerID string, whitelist *PeerWhitelist) bool {
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

// updateRoutingMetrics обновляет метрики маршрутизации
func (rm *RelayManager) updateRoutingMetrics(fromPeer, toPeer string, dataSize int, latency time.Duration) {
	// В реальной реализации здесь должны обновляться Prometheus метрики
	rm.logger.Debug("Routing metrics updated",
		zap.String("from_peer", fromPeer),
		zap.String("to_peer", toPeer),
		zap.Int("data_size", dataSize),
		zap.Duration("latency", latency),
	)
}

// calculateStats вычисляет статистику для тенанта
func (rm *RelayManager) calculateStats(tenantID string) *RelayStats {
	stats := &RelayStats{
		TotalPeers:        0,
		OnlinePeers:       0,
		DirectConnections: 0,
		RelayConnections:  0,
		TotalTrafficBytes: 0,
		AverageLatencyMS:  0,
	}

	// Подсчет пиров
	for _, peer := range rm.peers {
		if peer.TenantID == tenantID {
			stats.TotalPeers++
			if peer.IsOnline {
				stats.OnlinePeers++
			}
		}
	}

	// Подсчет соединений
	for _, conn := range rm.connections {
		fromPeer, exists := rm.peers[conn.FromPeerID]
		if !exists {
			continue
		}

		if fromPeer.TenantID == tenantID {
			if conn.ConnectionType == "direct" {
				stats.DirectConnections++
			} else if conn.ConnectionType == "relay" {
				stats.RelayConnections++
			}
			stats.TotalTrafficBytes += conn.BytesTransferred
		}
	}

	// Подсчет средней задержки
	totalLatency := 0
	latencyCount := 0
	for _, peer := range rm.peers {
		if peer.TenantID == tenantID && peer.LatencyMS > 0 {
			totalLatency += peer.LatencyMS
			latencyCount++
		}
	}

	if latencyCount > 0 {
		stats.AverageLatencyMS = totalLatency / latencyCount
	}

	return stats
}

// CreateConnection создает новое соединение между пирами
func (rm *RelayManager) CreateConnection(fromPeerID, toPeerID, connectionType string) *PeerConnection {
	connID := generateID()
	conn := &PeerConnection{
		ID:               connID,
		FromPeerID:       fromPeerID,
		ToPeerID:         toPeerID,
		ConnectionType:   connectionType,
		IsActive:         true,
		EstablishedAt:    time.Now(),
		LastActivity:     time.Now(),
		BytesTransferred: 0,
		LatencyMS:        0,
	}

	rm.connections[connID] = conn

	rm.logger.Info("Connection created",
		zap.String("connection_id", connID),
		zap.String("from_peer", fromPeerID),
		zap.String("to_peer", toPeerID),
		zap.String("connection_type", connectionType),
	)

	return conn
}

// UpdateConnection обновляет соединение
func (rm *RelayManager) UpdateConnection(connID string, bytesTransferred int64, latencyMS int) {
	conn, exists := rm.connections[connID]
	if !exists {
		return
	}

	conn.BytesTransferred += bytesTransferred
	conn.LatencyMS = latencyMS
	conn.LastActivity = time.Now()

	rm.logger.Debug("Connection updated",
		zap.String("connection_id", connID),
		zap.Int64("bytes_transferred", conn.BytesTransferred),
		zap.Int("latency_ms", latencyMS),
	)
}

// CloseConnection закрывает соединение
func (rm *RelayManager) CloseConnection(connID string) {
	conn, exists := rm.connections[connID]
	if !exists {
		return
	}

	conn.IsActive = false
	conn.LastActivity = time.Now()

	rm.logger.Info("Connection closed",
		zap.String("connection_id", connID),
		zap.String("from_peer", conn.FromPeerID),
		zap.String("to_peer", conn.ToPeerID),
	)
}
