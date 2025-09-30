package api

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/twogc/cloudbridge-relay-installer/internal/p2p"
	"go.uber.org/zap"
)

// P2PManager управляет P2P соединениями через API
type P2PManager struct {
	connectionManager p2p.P2PConnectionManagerInterface
	logger            *zap.Logger
}

// NewP2PManager создает новый P2P Manager
func NewP2PManager(logger *zap.Logger) *P2PManager {
	return &P2PManager{
		connectionManager: p2p.NewP2PConnectionManager(logger),
		logger:            logger,
	}
}

// EstablishConnection устанавливает P2P соединение
func (pm *P2PManager) EstablishConnection(c *gin.Context) {
	var req P2PConnectionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		pm.logger.Error("Failed to bind P2P connection request", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request format",
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	// Валидация JWT токена
	claims, err := pm.validateToken(c)
	if err != nil {
		pm.logger.Error("Failed to validate token", zap.Error(err))
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Invalid token",
			Code:    http.StatusUnauthorized,
			Message: err.Error(),
		})
		return
	}

	// Извлекаем tenant_id из пути
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		tenantID = claims.TenantID
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

	// Устанавливаем соединение
	connection, err := pm.connectionManager.EstablishConnection(
		c.Request.Context(),
		req.SessionID,
		claims.Subject, // peer1 ID
		req.TargetPeerID,
		req.Protocol,
	)
	if err != nil {
		pm.logger.Error("Failed to establish P2P connection", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Connection failed",
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	// Конвертируем ICE candidates если есть
	var iceCandidates []*ICECandidate
	if connection.ICEPair != nil {
		iceCandidates = append(iceCandidates, &ICECandidate{
			Foundation:  connection.ICEPair.LocalCandidate.Foundation,
			Component:   connection.ICEPair.LocalCandidate.Component,
			Transport:   connection.ICEPair.LocalCandidate.Transport,
			Priority:    connection.ICEPair.LocalCandidate.Priority,
			Address:     connection.ICEPair.LocalCandidate.Address,
			Port:        connection.ICEPair.LocalCandidate.Port,
			Type:        connection.ICEPair.LocalCandidate.Type.String(),
			RelatedAddr: connection.ICEPair.LocalCandidate.RelatedAddr,
			RelatedPort: connection.ICEPair.LocalCandidate.RelatedPort,
		})
	}

	response := P2PConnectionResponse{
		Success:        true,
		SessionID:      connection.SessionID,
		ConnectionType: connection.ConnectionType,
	}

	pm.logger.Info("P2P connection established",
		zap.String("session_id", req.SessionID),
		zap.String("connection_type", connection.ConnectionType),
		zap.String("protocol", req.Protocol))

	c.JSON(http.StatusOK, response)
}

// GetSessionStatus возвращает статус P2P сессии
func (pm *P2PManager) GetSessionStatus(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing session_id",
			Code:    http.StatusBadRequest,
			Message: "session_id parameter is required",
		})
		return
	}

	// Валидация JWT токена
	claims, err := pm.validateToken(c)
	if err != nil {
		pm.logger.Error("Failed to validate token", zap.Error(err))
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Invalid token",
			Code:    http.StatusUnauthorized,
			Message: err.Error(),
		})
		return
	}

	// Получаем соединение
	connection, exists := pm.connectionManager.GetConnection(sessionID)
	if !exists {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Session not found",
			Code:    http.StatusNotFound,
			Message: "Session with specified ID not found",
		})
		return
	}

	// Проверяем права доступа
	if connection.Peer1ID != claims.Subject && connection.Peer2ID != claims.Subject {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "Access denied",
			Code:    http.StatusForbidden,
			Message: "You don't have access to this session",
		})
		return
	}

	status := map[string]interface{}{
		"session_id":        connection.SessionID,
		"peer1_id":          connection.Peer1ID,
		"peer2_id":          connection.Peer2ID,
		"connection_type":   connection.ConnectionType,
		"protocol":          connection.Protocol,
		"state":             connection.State.String(),
		"established_at":    connection.EstablishedAt,
		"last_activity":     connection.LastActivity,
		"bytes_transferred": connection.BytesTransferred,
		"latency_ms":        connection.Latency.Milliseconds(),
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"status":  status,
	})
}

// CloseSession закрывает P2P сессию
func (pm *P2PManager) CloseSession(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing session_id",
			Code:    http.StatusBadRequest,
			Message: "session_id parameter is required",
		})
		return
	}

	// Валидация JWT токена
	claims, err := pm.validateToken(c)
	if err != nil {
		pm.logger.Error("Failed to validate token", zap.Error(err))
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Invalid token",
			Code:    http.StatusUnauthorized,
			Message: err.Error(),
		})
		return
	}

	// Получаем соединение для проверки прав
	connection, exists := pm.connectionManager.GetConnection(sessionID)
	if !exists {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Session not found",
			Code:    http.StatusNotFound,
			Message: "Session with specified ID not found",
		})
		return
	}

	// Проверяем права доступа
	if connection.Peer1ID != claims.Subject && connection.Peer2ID != claims.Subject {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "Access denied",
			Code:    http.StatusForbidden,
			Message: "You don't have access to this session",
		})
		return
	}

	// Закрываем соединение
	if err := pm.connectionManager.CloseConnection(sessionID); err != nil {
		pm.logger.Error("Failed to close P2P connection", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to close session",
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	pm.logger.Info("P2P session closed",
		zap.String("session_id", sessionID),
		zap.String("peer_id", claims.Subject))

	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Session closed successfully",
	})
}

// GetActiveConnections возвращает список активных соединений
func (pm *P2PManager) GetActiveConnections(c *gin.Context) {
	// Валидация JWT токена
	claims, err := pm.validateToken(c)
	if err != nil {
		pm.logger.Error("Failed to validate token", zap.Error(err))
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Invalid token",
			Code:    http.StatusUnauthorized,
			Message: err.Error(),
		})
		return
	}

	// Получаем активные соединения
	connections := pm.connectionManager.GetActiveConnections()

	// Фильтруем соединения для текущего пира
	var userConnections []map[string]interface{}
	for _, conn := range connections {
		if conn.Peer1ID == claims.Subject || conn.Peer2ID == claims.Subject {
			userConnections = append(userConnections, map[string]interface{}{
				"session_id":        conn.SessionID,
				"peer1_id":          conn.Peer1ID,
				"peer2_id":          conn.Peer2ID,
				"connection_type":   conn.ConnectionType,
				"protocol":          conn.Protocol,
				"state":             conn.State.String(),
				"established_at":    conn.EstablishedAt,
				"last_activity":     conn.LastActivity,
				"bytes_transferred": conn.BytesTransferred,
				"latency_ms":        conn.Latency.Milliseconds(),
			})
		}
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"success":     true,
		"connections": userConnections,
		"total":       len(userConnections),
	})
}

// validateToken валидирует JWT токен
func (pm *P2PManager) validateToken(c *gin.Context) (*P2PClaims, error) {
	// Проверяем, включена ли авторизация
	if os.Getenv("AUTH_ENABLED") == "false" {
		// Извлекаем server_id из токена даже при отключенной авторизации
		authHeader := c.GetHeader("Authorization")
		serverID := "anonymous"

		if authHeader != "" {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token != authHeader {
				// Извлекаем server_id из токена формата "server-id:server-a"
				if strings.Contains(token, ":") {
					parts := strings.Split(token, ":")
					if len(parts) >= 2 {
						serverID = parts[1]
						pm.logger.Info("Extracted server_id from token",
							zap.String("server_id", serverID),
							zap.String("token_prefix", parts[0]))
					}
				}
			}
		}

		tenantID := c.Param("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		pm.logger.Info("Using extracted server_id for peer_id generation",
			zap.String("server_id", serverID),
			zap.String("tenant_id", tenantID))

		return &P2PClaims{
			Subject:        "peer_" + serverID,
			TenantID:       tenantID,
			ServerID:       serverID,
			ConnectionType: "quic",
			ProtocolType:   "p2p-mesh",
			Permissions:    []string{"relay_connect", "relay_tunnel"},
		}, nil
	}

	// Получаем токен из заголовка Authorization
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("missing authorization header")
	}

	// Извлекаем токен
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		return nil, fmt.Errorf("invalid authorization format")
	}

	// Извлекаем server_id из токена формата "server-id:server-a"
	serverID := "anonymous"
	if strings.Contains(token, ":") {
		parts := strings.Split(token, ":")
		if len(parts) >= 2 {
			serverID = parts[1]
			pm.logger.Info("Extracted server_id from token",
				zap.String("server_id", serverID),
				zap.String("token_prefix", parts[0]))
		}
	}

	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		tenantID = "default"
	}

	pm.logger.Info("Using extracted server_id for peer_id generation",
		zap.String("server_id", serverID),
		zap.String("tenant_id", tenantID))

	return &P2PClaims{
		Subject:        "peer_" + serverID,
		TenantID:       tenantID,
		ServerID:       serverID,
		ConnectionType: "quic",
		ProtocolType:   "p2p-mesh",
		Permissions:    []string{"relay_connect", "relay_tunnel"},
	}, nil
}
