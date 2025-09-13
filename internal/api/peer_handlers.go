package api

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// PeerManager управляет пирами
type PeerManager struct {
	peers       map[string]*Peer
	connections map[string]*PeerConnection
	logger      *zap.Logger
	validator   TokenValidator
	ipValidator *IPValidator
}

// TokenValidator интерфейс для валидации токенов
type TokenValidator interface {
	ValidateP2PToken(token string) (*P2PClaims, error)
}

// P2PClaims JWT claims для P2P токенов
type P2PClaims struct {
	JTI               string           `json:"jti"`
	Subject           string           `json:"sub"`
	OrgID             string           `json:"org_id"`
	TenantID          string           `json:"tenant_id"`
	DeviceID          string           `json:"device_id"`
	DeviceFingerprint string           `json:"device_fingerprint"`
	ConnectionType    string           `json:"connection_type"`
	Permissions       []string         `json:"permissions"`
	WireGuardConfig   *WireGuardConfig `json:"wireguard_config"`
	MeshConfig        *MeshConfig      `json:"mesh_config"`
	PeerWhitelist     *PeerWhitelist   `json:"peer_whitelist"`
	NetworkConfig     *NetworkConfig   `json:"network_config"`
	IssuedAt          int64            `json:"iat"`
	NotBefore         int64            `json:"nbf"`
	Issuer            string           `json:"iss"`
	Audience          string           `json:"aud"`
	jwt.RegisteredClaims
}

// NewPeerManager создает новый PeerManager
func NewPeerManager(logger *zap.Logger, validator TokenValidator) *PeerManager {
	ipValidator := NewIPValidator(logger)

	// Добавляем разрешенные IP адреса для тестирования
	// RFC 1918 private networks
	ipValidator.AddAllowedCIDR("10.0.0.0/8")
	ipValidator.AddAllowedCIDR("172.16.0.0/12")
	ipValidator.AddAllowedCIDR("192.168.0.0/16")

	// RFC 5737 documentation networks
	ipValidator.AddAllowedCIDR("203.0.113.0/24")
	ipValidator.AddAllowedCIDR("198.51.100.0/24")
	ipValidator.AddAllowedCIDR("192.0.2.0/24")

	// Localhost
	ipValidator.AddAllowedIP("127.0.0.1")
	ipValidator.AddAllowedIP("::1")

	// Minikube IP
	ipValidator.AddAllowedIP("192.168.49.2")
	ipValidator.AddAllowedIP("95.163.250.190")

	logger.Info("IP validator initialized with allowed networks")

	return &PeerManager{
		peers:       make(map[string]*Peer),
		connections: make(map[string]*PeerConnection),
		logger:      logger,
		validator:   validator,
		ipValidator: ipValidator,
	}
}

// NewPeerManagerWithConfig создает новый PeerManager с конфигурацией
func NewPeerManagerWithConfig(logger *zap.Logger, validator TokenValidator, allowedIPs []string) *PeerManager {
	ipValidator := NewIPValidator(logger)

	// Добавляем разрешенные IP адреса из конфигурации
	for _, ip := range allowedIPs {
		if strings.Contains(ip, "/") {
			// Это CIDR
			if err := ipValidator.AddAllowedCIDR(ip); err != nil {
				logger.Warn("Failed to add CIDR to allowed list",
					zap.String("cidr", ip),
					zap.Error(err))
			} else {
				logger.Info("Added CIDR to allowed list", zap.String("cidr", ip))
			}
		} else {
			// Это отдельный IP
			if err := ipValidator.AddAllowedIP(ip); err != nil {
				logger.Warn("Failed to add IP to allowed list",
					zap.String("ip", ip),
					zap.Error(err))
			} else {
				logger.Info("Added IP to allowed list", zap.String("ip", ip))
			}
		}
	}

	logger.Info("IP validator initialized with configuration",
		zap.Int("allowed_ips_count", len(allowedIPs)))

	return &PeerManager{
		peers:       make(map[string]*Peer),
		connections: make(map[string]*PeerConnection),
		logger:      logger,
		validator:   validator,
		ipValidator: ipValidator,
	}
}

// RegisterPeer регистрирует новый пир
func (pm *PeerManager) RegisterPeer(c *gin.Context) {
	var req PeerRegistrationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		pm.logger.Error("Failed to bind registration request", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request format",
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	// Валидация JWT токена
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Missing Authorization header",
			Code:    http.StatusUnauthorized,
			Message: "Authorization header is required",
		})
		return
	}

	// Извлекаем токен из "Bearer <token>"
	token := authHeader
	if strings.HasPrefix(authHeader, "Bearer ") {
		token = strings.TrimPrefix(authHeader, "Bearer ")
	}
	if token == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Missing authorization header",
			Code:    http.StatusUnauthorized,
			Message: "Authorization header is required",
		})
		return
	}

	claims, err := pm.validator.ValidateP2PToken(token)
	if err != nil {
		pm.logger.Error("Failed to validate P2P token", zap.Error(err))
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Invalid token",
			Code:    http.StatusUnauthorized,
			Message: err.Error(),
		})
		return
	}

	// Проверка прав доступа
	if claims.TenantID != req.TenantID {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "Tenant mismatch",
			Code:    http.StatusForbidden,
			Message: "Token tenant does not match request tenant",
		})
		return
	}

	// Проверка существования пира
	if _, exists := pm.peers[req.PeerID]; exists {
		c.JSON(http.StatusConflict, ErrorResponse{
			Error:   "Peer already exists",
			Code:    http.StatusConflict,
			Message: "Peer with this ID is already registered",
		})
		return
	}

	// Валидация WireGuard конфигурации
	if req.WireGuardConfig != nil {
		// Валидация endpoint IP
		if err := pm.ipValidator.ValidateWireGuardEndpoint(req.WireGuardConfig.Endpoint); err != nil {
			pm.logger.Error("Invalid WireGuard endpoint", zap.Error(err))
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Invalid WireGuard endpoint",
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			})
			return
		}

		// Валидация allowed IPs
		if err := pm.ipValidator.ValidateAllowedIPs(req.WireGuardConfig.AllowedIPs); err != nil {
			pm.logger.Error("Invalid WireGuard allowed IPs", zap.Error(err))
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Invalid WireGuard allowed IPs",
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			})
			return
		}
	}

	// Валидация peer whitelist
	if req.PeerWhitelist != nil {
		// Проверяем, что все allowed peers являются валидными
		for _, allowedPeer := range req.PeerWhitelist.AllowedPeers {
			if allowedPeer == "" {
				pm.logger.Error("Invalid peer whitelist: empty peer ID")
				c.JSON(http.StatusBadRequest, ErrorResponse{
					Error:   "Invalid peer whitelist",
					Code:    http.StatusBadRequest,
					Message: "Empty peer ID in whitelist",
				})
				return
			}
		}
	}

	// Создание пира
	peerID := generateID()
	peer := &Peer{
		ID:                peerID,
		PeerID:            req.PeerID,
		TenantID:          req.TenantID,
		OrgID:             req.OrgID,
		DeviceID:          req.DeviceID,
		ConnectionType:    req.ConnectionType,
		WireGuardConfig:   req.WireGuardConfig,
		MeshConfig:        req.MeshConfig,
		PeerWhitelist:     req.PeerWhitelist,
		NetworkConfig:     req.NetworkConfig,
		IsOnline:          true,
		LastSeen:          time.Now(),
		RegisteredAt:      time.Now(),
		ActiveConnections: 0,
		MeshStatus:        "connected",
		LatencyMS:         0,
	}

	// Сохранение пира
	pm.peers[req.PeerID] = peer

	pm.logger.Info("Peer registered successfully",
		zap.String("peer_id", req.PeerID),
		zap.String("tenant_id", req.TenantID),
		zap.String("device_id", req.DeviceID),
	)

	c.JSON(http.StatusOK, PeerRegistrationResponse{
		Success:           true,
		PeerID:            peer.PeerID,
		RegisteredAt:      peer.RegisteredAt,
		DiscoveryEndpoint: "wss://relay.2gc.ru/ws/peers",
	})
}

// DiscoverPeers возвращает список пиров для тенанта
func (pm *PeerManager) DiscoverPeers(c *gin.Context) {
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
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Missing Authorization header",
			Code:    http.StatusUnauthorized,
			Message: "Authorization header is required",
		})
		return
	}

	// Извлекаем токен из "Bearer <token>"
	token := authHeader
	if strings.HasPrefix(authHeader, "Bearer ") {
		token = strings.TrimPrefix(authHeader, "Bearer ")
	}
	claims, err := pm.validator.ValidateP2PToken(token)
	if err != nil {
		pm.logger.Error("Failed to validate P2P token", zap.Error(err))
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

	// Получение пиров из памяти (в реальной реализации - из базы данных)
	var peers []*Peer
	for _, peer := range pm.peers {
		if peer.TenantID == tenantID {
			// Фильтрация по whitelist
			if pm.validatePeerAccess(peer.PeerID, claims.PeerWhitelist) {
				// Обновление статуса онлайн
				peer.IsOnline = pm.isPeerOnline(peer.PeerID)
				peer.LastSeen = pm.getLastSeen(peer.PeerID)
				peers = append(peers, peer)
			}
		}
	}

	pm.logger.Info("Peer discovery completed",
		zap.String("tenant_id", tenantID),
		zap.Int("peers_found", len(peers)),
	)

	c.JSON(http.StatusOK, PeerDiscoveryResponse{
		Success:  true,
		TenantID: tenantID,
		Peers:    peers,
	})
}

// UpdatePeerStatus обновляет статус пира
func (pm *PeerManager) UpdatePeerStatus(c *gin.Context) {
	peerID := c.Param("peer_id")

	var req PeerStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pm.logger.Error("Failed to bind status update request", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request format",
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	// Валидация JWT токена
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Missing Authorization header",
			Code:    http.StatusUnauthorized,
			Message: "Authorization header is required",
		})
		return
	}

	// Извлекаем токен из "Bearer <token>"
	token := authHeader
	if strings.HasPrefix(authHeader, "Bearer ") {
		token = strings.TrimPrefix(authHeader, "Bearer ")
	}
	claims, err := pm.validator.ValidateP2PToken(token)
	if err != nil {
		pm.logger.Error("Failed to validate P2P token", zap.Error(err))
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Invalid token",
			Code:    http.StatusUnauthorized,
			Message: err.Error(),
		})
		return
	}

	// Проверка существования пира
	peer, exists := pm.peers[peerID]
	if !exists {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Peer not found",
			Code:    http.StatusNotFound,
			Message: "Peer with this ID does not exist",
		})
		return
	}

	// Проверка прав доступа
	if peer.TenantID != claims.TenantID {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "Tenant mismatch",
			Code:    http.StatusForbidden,
			Message: "Token tenant does not match peer tenant",
		})
		return
	}

	// Обновление статуса
	peer.IsOnline = req.IsOnline
	peer.LatencyMS = req.LatencyMS
	peer.ActiveConnections = req.ActiveConnections
	peer.MeshStatus = req.MeshStatus
	peer.LastSeen = time.Now()

	pm.logger.Info("Peer status updated",
		zap.String("peer_id", peerID),
		zap.Bool("is_online", req.IsOnline),
		zap.Int("latency_ms", req.LatencyMS),
		zap.String("mesh_status", req.MeshStatus),
	)

	c.JSON(http.StatusOK, PeerStatusResponse{
		Success:   true,
		UpdatedAt: time.Now(),
	})
}

// GetPeerInfo возвращает информацию о пире
func (pm *PeerManager) GetPeerInfo(c *gin.Context) {
	peerID := c.Param("peer_id")

	// Валидация JWT токена
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Missing Authorization header",
			Code:    http.StatusUnauthorized,
			Message: "Authorization header is required",
		})
		return
	}

	// Извлекаем токен из "Bearer <token>"
	token := authHeader
	if strings.HasPrefix(authHeader, "Bearer ") {
		token = strings.TrimPrefix(authHeader, "Bearer ")
	}
	claims, err := pm.validator.ValidateP2PToken(token)
	if err != nil {
		pm.logger.Error("Failed to validate P2P token", zap.Error(err))
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Invalid token",
			Code:    http.StatusUnauthorized,
			Message: err.Error(),
		})
		return
	}

	// Проверка существования пира
	peer, exists := pm.peers[peerID]
	if !exists {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Peer not found",
			Code:    http.StatusNotFound,
			Message: "Peer with this ID does not exist",
		})
		return
	}

	// Проверка прав доступа
	if peer.TenantID != claims.TenantID {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "Tenant mismatch",
			Code:    http.StatusForbidden,
			Message: "Token tenant does not match peer tenant",
		})
		return
	}

	c.JSON(http.StatusOK, PeerInfoResponse{
		Success: true,
		Peer:    peer,
	})
}

// validatePeerAccess проверяет доступ пира
func (pm *PeerManager) validatePeerAccess(peerID string, whitelist *PeerWhitelist) bool {
	if whitelist == nil {
		return true // Нет whitelist - все разрешено
	}

	if whitelist.AutoApprove {
		return true
	}

	// Проверяем, что peerID есть в списке разрешенных
	for _, allowedPeer := range whitelist.AllowedPeers {
		if allowedPeer == peerID {
			return true
		}
	}

	return false
}

// isPeerOnline проверяет, онлайн ли пир
func (pm *PeerManager) isPeerOnline(peerID string) bool {
	peer, exists := pm.peers[peerID]
	if !exists {
		return false
	}

	// Простая проверка: если последний раз видели менее 5 минут назад
	return time.Since(peer.LastSeen) < 5*time.Minute
}

// getLastSeen возвращает время последнего контакта с пиром
func (pm *PeerManager) getLastSeen(peerID string) time.Time {
	peer, exists := pm.peers[peerID]
	if !exists {
		return time.Time{}
	}

	return peer.LastSeen
}

// generateID генерирует уникальный ID
func generateID() string {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		// Fallback to timestamp-based ID if crypto/rand fails
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// validateP2PToken валидирует P2P JWT токен
func (pm *PeerManager) validateP2PToken(tokenString string) (*P2PClaims, error) {
	// Удаление "Bearer " префикса
	if len(tokenString) > 7 && strings.HasPrefix(tokenString, "Bearer ") {
		tokenString = tokenString[7:]
	}

	// Парсинг токена
	token, err := jwt.ParseWithClaims(tokenString, &P2PClaims{}, func(token *jwt.Token) (interface{}, error) {
		// В реальной реализации здесь должен быть секретный ключ
		return []byte("your-jwt-secret-key"), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*P2PClaims)
	if !ok || !token.Valid {
		return nil, err
	}

	// Проверка обязательных полей
	if claims.ConnectionType != "p2p-mesh" {
		return nil, err
	}

	if claims.DeviceID == "" {
		return nil, err
	}

	if claims.WireGuardConfig == nil {
		return nil, err
	}

	return claims, nil
}
