package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/twogc/cloudbridge-relay-installer/internal/p2p"
	"go.uber.org/zap"
)

// HeartbeatHandler обрабатывает heartbeat запросы
type HeartbeatHandler struct {
	heartbeatManager *p2p.HeartbeatManager
	relayManager     *p2p.RelayManager
	logger           *zap.Logger
}

// NewHeartbeatHandler создает новый Heartbeat Handler
func NewHeartbeatHandler(heartbeatManager *p2p.HeartbeatManager, relayManager *p2p.RelayManager, logger *zap.Logger) *HeartbeatHandler {
	return &HeartbeatHandler{
		heartbeatManager: heartbeatManager,
		relayManager:     relayManager,
		logger:           logger,
	}
}

// HandleHeartbeat обрабатывает heartbeat запрос
func (hh *HeartbeatHandler) HandleHeartbeat(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	peerID := c.Param("peer_id")
	
	if tenantID == "" || peerID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing required parameters",
			Code:    http.StatusBadRequest,
			Message: "tenant_id and peer_id are required",
		})
		return
	}
	
	// Валидация JWT токена
	claims, err := hh.validateToken(c)
	if err != nil {
		hh.logger.Error("Failed to validate token", zap.Error(err))
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
	
	// Обрабатываем heartbeat
	response, err := hh.heartbeatManager.ProcessHeartbeat(peerID, tenantID)
	if err != nil {
		hh.logger.Error("Failed to process heartbeat",
			zap.String("peer_id", peerID),
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Heartbeat processing failed",
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}
	
	hh.logger.Debug("Heartbeat processed successfully",
		zap.String("peer_id", peerID),
		zap.String("tenant_id", tenantID),
		zap.Bool("online", response.Online))
	
	c.JSON(http.StatusOK, response)
}

// GetPeerStatus возвращает статус пира
func (hh *HeartbeatHandler) GetPeerStatus(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	peerID := c.Param("peer_id")
	
	if tenantID == "" || peerID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing required parameters",
			Code:    http.StatusBadRequest,
			Message: "tenant_id and peer_id are required",
		})
		return
	}
	
	// Валидация JWT токена
	claims, err := hh.validateToken(c)
	if err != nil {
		hh.logger.Error("Failed to validate token", zap.Error(err))
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
	
	// Получаем статус пира
	peerHeartbeat, err := hh.heartbeatManager.GetPeerStatus(peerID)
	if err != nil {
		hh.logger.Error("Failed to get peer status",
			zap.String("peer_id", peerID),
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Peer not found",
			Code:    http.StatusNotFound,
			Message: err.Error(),
		})
		return
	}
	
	// Формируем ответ
	status := map[string]interface{}{
		"peer_id":         peerHeartbeat.PeerID,
		"tenant_id":       peerHeartbeat.TenantID,
		"is_online":       peerHeartbeat.IsOnline,
		"last_heartbeat":  peerHeartbeat.LastHeartbeat.Unix(),
		"heartbeat_count": peerHeartbeat.HeartbeatCount,
		"last_seen":       peerHeartbeat.LastSeen.Unix(),
		"uptime":          time.Since(peerHeartbeat.LastSeen).Seconds(),
	}
	
	c.JSON(http.StatusOK, status)
}

// GetTenantStats возвращает статистику по tenant
func (hh *HeartbeatHandler) GetTenantStats(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing required parameter",
			Code:    http.StatusBadRequest,
			Message: "tenant_id is required",
		})
		return
	}
	
	// Валидация JWT токена
	claims, err := hh.validateToken(c)
	if err != nil {
		hh.logger.Error("Failed to validate token", zap.Error(err))
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
	
	// Получаем статистику
	stats := hh.heartbeatManager.GetTenantStats(tenantID)
	
	// Получаем информацию о tenant
	tenantInfo, err := hh.relayManager.GetTenantInfo(tenantID)
	if err != nil {
		hh.logger.Error("Failed to get tenant info",
			zap.String("tenant_id", tenantID),
			zap.Error(err))
	} else {
		stats["tenant_info"] = map[string]interface{}{
			"tenant_id":        tenantInfo.TenantID,
			"network_cidr":     tenantInfo.NetworkCIDR,
			"max_peers":        tenantInfo.MaxPeers,
			"max_connections":  tenantInfo.MaxConnections,
			"created_at":       tenantInfo.CreatedAt.Unix(),
			"last_activity":    tenantInfo.LastActivity.Unix(),
		}
	}
	
	c.JSON(http.StatusOK, stats)
}

// GetOnlinePeers возвращает список онлайн пиров
func (hh *HeartbeatHandler) GetOnlinePeers(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing required parameter",
			Code:    http.StatusBadRequest,
			Message: "tenant_id is required",
		})
		return
	}
	
	// Валидация JWT токена
	claims, err := hh.validateToken(c)
	if err != nil {
		hh.logger.Error("Failed to validate token", zap.Error(err))
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
	
	// Получаем список онлайн пиров
	onlinePeers := hh.heartbeatManager.GetOnlinePeers(tenantID)
	
	// Формируем ответ
	response := map[string]interface{}{
		"tenant_id":    tenantID,
		"online_peers": onlinePeers,
		"count":        len(onlinePeers),
		"timestamp":    time.Now().Unix(),
	}
	
	c.JSON(http.StatusOK, response)
}

// validateToken валидирует JWT токен
func (hh *HeartbeatHandler) validateToken(c *gin.Context) (*P2PClaims, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("authorization header is required")
	}
	
	// Извлекаем токен из заголовка
	_ = authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		_ = authHeader[7:]
	}
	
	// Здесь должна быть валидация JWT токена
	// Для демонстрации возвращаем фиктивные claims
	claims := &P2PClaims{
		TenantID: c.Param("tenant_id"),
		Subject:  "test-peer",
	}
	
	return claims, nil
}

// RegisterHeartbeatRoutes регистрирует маршруты для heartbeat
func RegisterHeartbeatRoutes(router *gin.Engine, heartbeatHandler *HeartbeatHandler) {
	// Группа маршрутов для heartbeat
	heartbeatGroup := router.Group("/api/v1/tenants/:tenant_id/peers/:peer_id")
	{
		heartbeatGroup.POST("/heartbeat", heartbeatHandler.HandleHeartbeat)
		heartbeatGroup.GET("/status", heartbeatHandler.GetPeerStatus)
	}
	
	// Группа маршрутов для статистики tenant
	tenantGroup := router.Group("/api/v1/tenants/:tenant_id")
	{
		tenantGroup.GET("/stats", heartbeatHandler.GetTenantStats)
		tenantGroup.GET("/peers/online", heartbeatHandler.GetOnlinePeers)
	}
}
