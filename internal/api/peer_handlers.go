package api

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/twogc/cloudbridge-relay-installer/internal/p2p"
	"go.uber.org/zap"
)

// PeerManager управляет пирами
type PeerManager struct {
	peers        map[string]*p2p.Peer
	connections  map[string]*PeerConnection
	logger       *zap.Logger
	validator    TokenValidator
	ipValidator  IPValidatorInterface
	peerStore    PeerStoreInterface
	allocatedIPs map[string]string // peer_id -> allocated_ip
	ipMutex      sync.RWMutex
}

// PeerStoreInterface интерфейс для PeerStore
type PeerStoreInterface interface {
	Store(tenantID, peerID string, peer *p2p.Peer)
	Get(tenantID, peerID string) (*p2p.Peer, bool)
	Update(tenantID, peerID string, peer *p2p.Peer)
	GetAll(tenantID string) []*p2p.Peer
	IsPeerOnline(peerID string) bool
	GetLastSeen(peerID string) time.Time
}

// IPValidatorInterface интерфейс для IPValidator
type IPValidatorInterface interface {
	IsIPAllowed(ip string) bool
	ValidateP2PQUICEndpoint(endpoint string) error
	ValidateAllowedIPs(allowedIPs []string) error
}

// PeerStore управляет хранением пиров
type PeerStore struct {
	peers map[string]map[string]*p2p.Peer // tenant_id -> peer_id -> peer
	mu    sync.RWMutex
}

// NewPeerStore создает новый PeerStore
func NewPeerStore() *PeerStore {
	return &PeerStore{
		peers: make(map[string]map[string]*p2p.Peer),
	}
}

// Store сохраняет пира
func (ps *PeerStore) Store(tenantID, peerID string, peer *p2p.Peer) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.peers[tenantID] == nil {
		ps.peers[tenantID] = make(map[string]*p2p.Peer)
	}
	ps.peers[tenantID][peerID] = peer
}

// Get получает пира
func (ps *PeerStore) Get(tenantID, peerID string) (*p2p.Peer, bool) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if tenantPeers, exists := ps.peers[tenantID]; exists {
		if peer, exists := tenantPeers[peerID]; exists {
			return peer, true
		}
	}
	return nil, false
}

// Update обновляет пира
func (ps *PeerStore) Update(tenantID, peerID string, peer *p2p.Peer) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.peers[tenantID] == nil {
		ps.peers[tenantID] = make(map[string]*p2p.Peer)
	}
	ps.peers[tenantID][peerID] = peer
}

// GetAll получает всех пиров для тенанта
func (ps *PeerStore) GetAll(tenantID string) []*p2p.Peer {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	var peers []*p2p.Peer
	if tenantPeers, exists := ps.peers[tenantID]; exists {
		for _, peer := range tenantPeers {
			peers = append(peers, peer)
		}
	}
	return peers
}

// IsPeerOnline проверяет, онлайн ли пир
func (ps *PeerStore) IsPeerOnline(peerID string) bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	for _, tenantPeers := range ps.peers {
		if peer, exists := tenantPeers[peerID]; exists {
			// Простая проверка: если последний раз видели менее 5 минут назад
			return time.Since(peer.LastSeen) < 5*time.Minute
		}
	}
	return false
}

// GetLastSeen возвращает время последнего контакта с пиром
func (ps *PeerStore) GetLastSeen(peerID string) time.Time {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	for _, tenantPeers := range ps.peers {
		if peer, exists := tenantPeers[peerID]; exists {
			return peer.LastSeen
		}
	}
	return time.Time{}
}

// TokenValidator интерфейс для валидации токенов
type TokenValidator interface {
	ValidateP2PToken(token string) (*P2PClaims, error)
}

// P2PClaims JWT claims для P2P токенов
type P2PClaims struct {
	JTI            string   `json:"jti"`
	Subject        string   `json:"sub"`
	TenantID       string   `json:"tenant_id"`
	ServerID       string   `json:"server_id"`
	ConnectionType string   `json:"connection_type"`
	ProtocolType   string   `json:"protocol_type"`
	Permissions    []string `json:"permissions"`
	IssuedAt       int64    `json:"iat"`
	NotBefore      int64    `json:"nbf"`
	Issuer         string   `json:"iss"`
	Audience       string   `json:"aud"`
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
		peers:        make(map[string]*p2p.Peer),
		connections:  make(map[string]*PeerConnection),
		logger:       logger,
		validator:    validator,
		ipValidator:  ipValidator,
		peerStore:    NewPeerStore(),
		allocatedIPs: make(map[string]string),
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
		peers:       make(map[string]*p2p.Peer),
		connections: make(map[string]*PeerConnection),
		logger:      logger,
		validator:   validator,
		ipValidator: ipValidator,
		peerStore:   NewPeerStore(),
	}
}

// allocateIPForTenant выделяет IP адрес для пира в рамках tenant
func (pm *PeerManager) allocateIPForTenant(tenantID, peerID string) (string, error) {
	pm.ipMutex.Lock()
	defer pm.ipMutex.Unlock()

	// Проверяем, есть ли уже выделенный IP для этого пира
	if allocatedIP, exists := pm.allocatedIPs[peerID]; exists {
		return allocatedIP, nil
	}

	// Определяем базовый IP для tenant
	// tenant-216420165 -> 10.100.77.0/24
	baseIP := "10.100.77"

	// Находим следующий доступный IP
	for i := 10; i <= 254; i++ {
		candidateIP := fmt.Sprintf("%s.%d", baseIP, i)

		// Проверяем, не занят ли этот IP
		alreadyAllocated := false
		for _, allocatedIP := range pm.allocatedIPs {
			if allocatedIP == candidateIP {
				alreadyAllocated = true
				break
			}
		}

		if !alreadyAllocated {
			pm.allocatedIPs[peerID] = candidateIP
			pm.logger.Info("Allocated IP for peer",
				zap.String("peer_id", peerID),
				zap.String("tenant_id", tenantID),
				zap.String("allocated_ip", candidateIP))
			return candidateIP, nil
		}
	}

	return "", fmt.Errorf("no available IP addresses for tenant %s", tenantID)
}

// deallocateIP освобождает IP адрес пира
func (pm *PeerManager) deallocateIP(peerID string) {
	pm.ipMutex.Lock()
	defer pm.ipMutex.Unlock()

	if allocatedIP, exists := pm.allocatedIPs[peerID]; exists {
		delete(pm.allocatedIPs, peerID)
		pm.logger.Info("Deallocated IP for peer",
			zap.String("peer_id", peerID),
			zap.String("deallocated_ip", allocatedIP))
	}
}

// RegisterPeer регистрирует новый пир для relay-assisted режима
func (pm *PeerManager) RegisterPeer(c *gin.Context) {
	var req RelayAssistedRegistrationRequest

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

	// Извлекаем tenant_id из пути или токена
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

	// Генерируем peer_id из server_id в токене
	pm.logger.Info("DEBUG: ServerID from token", zap.String("server_id", claims.ServerID))

	// Debug: Check the raw token
	authHeader := c.GetHeader("Authorization")
	pm.logger.Info("DEBUG: Raw Authorization header", zap.String("auth_header", authHeader))

	// Check if token is in special format "server-id:server-a"
	var actualServerID string
	if strings.HasPrefix(claims.ServerID, "server-id:") {
		actualServerID = strings.TrimPrefix(claims.ServerID, "server-id:")
		pm.logger.Info("DEBUG: Extracted server_id from special format", zap.String("extracted_server_id", actualServerID))
	} else {
		actualServerID = claims.ServerID
	}

	peerID := generatePeerID(actualServerID)
	pm.logger.Info("DEBUG: Generated peer_id", zap.String("peer_id", peerID))

	// Проверка существования пира — делаем регистрацию идемпотентной
	if existingPeer, exists := pm.peerStore.Get(tenantID, peerID); exists {
		// Обновляем relay_session, если отсутствует
		if strings.TrimSpace(existingPeer.RelaySessionID) == "" {
			existingPeer.RelaySessionID = generateRelaySessionID()
		}
		if existingPeer.RegisteredAt.IsZero() {
			existingPeer.RegisteredAt = time.Now()
		} else {
			// Обновляем время регистрации, чтобы определить самую свежую запись
			existingPeer.RegisteredAt = time.Now()
		}

		// Выделяем IP адрес, если он еще не выделен
		var allocatedIP string
		if existingPeer.AllocatedIP == "" {
			var err error
			allocatedIP, err = pm.allocateIPForTenant(tenantID, peerID)
			if err != nil {
				pm.logger.Error("Failed to allocate IP for existing peer",
					zap.String("peer_id", peerID),
					zap.String("tenant_id", tenantID),
					zap.Error(err))
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Failed to allocate IP address",
					Code:    http.StatusInternalServerError,
					Message: err.Error(),
				})
				return
			}
			existingPeer.AllocatedIP = allocatedIP
		} else {
			allocatedIP = existingPeer.AllocatedIP
		}

		pm.peerStore.Update(tenantID, peerID, existingPeer)
		pm.logger.Info("Peer re-registered",
			zap.String("peer_id", peerID),
			zap.String("tenant_id", tenantID),
			zap.String("relay_session_id", existingPeer.RelaySessionID),
			zap.String("allocated_ip", allocatedIP))

		c.JSON(http.StatusOK, RelayAssistedRegistrationResponse{
			Success:        true,
			PeerID:         existingPeer.PeerID,
			RelaySessionID: existingPeer.RelaySessionID,
			RegisteredAt:   existingPeer.RegisteredAt,
			AllocatedIP:    allocatedIP,
		})
		return
	}

	// Валидация allowed IPs
	if err := pm.ipValidator.ValidateAllowedIPs(req.AllowedIPs); err != nil {
		pm.logger.Error("Invalid allowed IPs", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid allowed IPs",
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	// Генерируем relay_session_id
	relaySessionID := generateRelaySessionID()

	// Выделяем IP адрес для пира
	allocatedIP, err := pm.allocateIPForTenant(tenantID, peerID)
	if err != nil {
		pm.logger.Error("Failed to allocate IP for peer",
			zap.String("peer_id", peerID),
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to allocate IP address",
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	// Создание пира для P2P mesh режима
	peer := &p2p.Peer{
		ID:             peerID,
		PeerID:         peerID,
		TenantID:       tenantID,
		DeviceID:       claims.ServerID,
		ConnectionType: claims.ConnectionType,
		RelaySessionID: relaySessionID,
		IsOnline:       true,
		LastSeen:       time.Now(),
		RegisteredAt:   time.Now(),
		MeshStatus:     "connected",
		AllocatedIP:    allocatedIP, // Добавляем выделенный IP
	}

	// Сохранение пира в peerStore
	pm.peerStore.Store(tenantID, peerID, peer)

	pm.logger.Info("Peer registered successfully for P2P mesh mode",
		zap.String("peer_id", peerID),
		zap.String("tenant_id", tenantID),
		zap.String("relay_session_id", relaySessionID),
		zap.String("connection_type", claims.ConnectionType),
	)

	c.JSON(http.StatusOK, RelayAssistedRegistrationResponse{
		Success:        true,
		PeerID:         peerID,
		RelaySessionID: relaySessionID,
		RegisteredAt:   peer.RegisteredAt,
		AllocatedIP:    allocatedIP,
	})
}

// DiscoverPeers возвращает список пиров для тенанта в relay-assisted режиме
func (pm *PeerManager) DiscoverPeers(c *gin.Context) {
	// Support both namespaced and query variants
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		tenantID = c.Query("tenant_id")
	}
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing tenant_id",
			Code:    http.StatusBadRequest,
			Message: "tenant_id query parameter is required",
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

	// Проверка прав доступа
	if claims.TenantID != tenantID {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "Tenant mismatch",
			Code:    http.StatusForbidden,
			Message: "Token tenant does not match request tenant",
		})
		return
	}

	// Получение пиров из peerStore
	peers := pm.peerStore.GetAll(tenantID)

	// Обновление статуса и возврат всех пиров для P2P mesh
	var filteredPeers []*p2p.Peer
	for _, peer := range peers {
		// Обновление статуса онлайн
		peer.IsOnline = pm.isPeerOnline(peer.PeerID)
		peer.LastSeen = pm.getLastSeen(peer.PeerID)
		filteredPeers = append(filteredPeers, peer)
	}

	// P2P mesh не требует синтетического relay peer

	pm.logger.Info("Peer discovery completed",
		zap.String("tenant_id", tenantID),
		zap.Int("peers_found", len(filteredPeers)),
	)

	c.JSON(http.StatusOK, RelayAssistedDiscoveryResponse{
		Success:  true,
		TenantID: tenantID,
		Peers:    filteredPeers,
	})
}

// UpdatePeerStatus обновляет статус пира (heartbeat)
func (pm *PeerManager) UpdatePeerStatus(c *gin.Context) {
	peerID := c.Param("peer_id")
	tenantID := c.Param("tenant_id")

	// ДОБАВИТЬ: Логирование для диагностики
	pm.logger.Info("Heartbeat request",
		zap.String("peer_id", peerID),
		zap.String("tenant_id", tenantID))

	var req RelayAssistedStatusUpdateRequest
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

	// Проверка прав доступа
	if claims.TenantID != tenantID {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "Tenant mismatch",
			Code:    http.StatusForbidden,
			Message: "Token tenant does not match request tenant",
		})
		return
	}

	// ДОБАВИТЬ: Проверка существования пира
	peer, exists := pm.peerStore.Get(tenantID, peerID)
	if !exists {
		// ДОБАВИТЬ: Автоматическая регистрация при heartbeat
		pm.logger.Info("Peer not found during heartbeat, attempting auto-registration",
			zap.String("peer_id", peerID),
			zap.String("tenant_id", tenantID))

		// Создаем пира автоматически на основе heartbeat запроса
		relaySessionID := generateRelaySessionID()
		peer = &p2p.Peer{
			ID:             peerID,
			PeerID:         peerID,
			TenantID:       tenantID,
			DeviceID:       claims.ServerID,
			ConnectionType: claims.ConnectionType,
			RelaySessionID: relaySessionID,
			IsOnline:       true,
			LastSeen:       time.Now(),
			RegisteredAt:   time.Now(),
			MeshStatus:     req.Status,
		}

		// Сохраняем пира
		pm.peerStore.Store(tenantID, peerID, peer)

		pm.logger.Info("Peer auto-registered during heartbeat",
			zap.String("peer_id", peerID),
			zap.String("tenant_id", tenantID),
			zap.String("relay_session_id", relaySessionID))
	}

	// Обновление статуса
	peer.IsOnline = req.Status == "active"
	peer.LastSeen = time.Now()
	peer.MeshStatus = req.Status

	// Сохраняем обновленный пир
	pm.peerStore.Update(tenantID, peerID, peer)

	pm.logger.Info("Peer status updated (heartbeat)",
		zap.String("peer_id", peerID),
		zap.String("tenant_id", tenantID),
		zap.String("status", req.Status),
		zap.String("relay_session_id", req.RelaySessionID),
	)

	c.JSON(http.StatusOK, PeerStatusResponse{
		Success:   true,
		UpdatedAt: time.Now(),
	})
}

// GetPeerInfo возвращает информацию о пире
func (pm *PeerManager) GetPeerInfo(c *gin.Context) {
	peerID := c.Param("peer_id")
	tenantID := c.Param("tenant_id")

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

	// Проверка прав доступа
	if claims.TenantID != tenantID {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "Tenant mismatch",
			Code:    http.StatusForbidden,
			Message: "Token tenant does not match request tenant",
		})
		return
	}

	// Проверка существования пира
	peer, exists := pm.peerStore.Get(tenantID, peerID)
	if !exists {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Peer not found",
			Code:    http.StatusNotFound,
			Message: "Peer with this ID does not exist",
		})
		return
	}

	c.JSON(http.StatusOK, PeerInfoResponse{
		Success: true,
		Peer:    peer,
	})
}

// isPeerOnline проверяет, онлайн ли пир
func (pm *PeerManager) isPeerOnline(peerID string) bool {
	return pm.peerStore.IsPeerOnline(peerID)
}

// getLastSeen возвращает время последнего контакта с пиром
func (pm *PeerManager) getLastSeen(peerID string) time.Time {
	return pm.peerStore.GetLastSeen(peerID)
}

// GetPeer получает пира по ID (для интерфейса PeerManagerInterface)
func (pm *PeerManager) GetPeer(peerID string) (*p2p.Peer, error) {
	// Ищем пира во всех тенантах
	// Используем известные tenant ID или получаем их из других источников
	// Для простоты используем фиксированный список или получаем из контекста
	tenantIDs := []string{"default", "tenant1", "tenant2"} // Можно расширить

	for _, tenantID := range tenantIDs {
		if peer, exists := pm.peerStore.Get(tenantID, peerID); exists {
			return peer, nil
		}
	}
	return nil, fmt.Errorf("peer not found: %s", peerID)
}

// RegisterPeerDirect регистрирует пира напрямую (для интерфейса PeerManagerInterface)
func (pm *PeerManager) RegisterPeerDirect(peer *p2p.Peer) error {
	// Простая реализация - сохраняем пира в store
	tenantID := peer.TenantID
	if tenantID == "" {
		tenantID = "default"
	}

	pm.peerStore.Store(tenantID, peer.ID, peer)
	pm.logger.Info("Peer registered",
		zap.String("peer_id", peer.ID),
		zap.String("tenant_id", tenantID))
	return nil
}

// RegisterPeerDirectWithRequest регистрирует пира напрямую по запросу
func (pm *PeerManager) RegisterPeerDirectWithRequest(req DirectPeerRegistrationRequest) (*DirectPeerRegistrationResponse, error) {
	// Валидация запроса
	if req.PeerID == "" || req.TenantID == "" || req.DeviceID == "" {
		return nil, fmt.Errorf("invalid request: peer_id, tenant_id and device_id are required")
	}

	// Выделяем IP адрес
	allocatedIP, err := pm.allocateIPForTenant(req.TenantID, req.PeerID)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate IP: %w", err)
	}

	// Создаем пира
	peer := &p2p.Peer{
		ID:           req.PeerID,
		PeerID:       req.PeerID,
		TenantID:     req.TenantID,
		DeviceID:     req.DeviceID,
		IsOnline:     true,
		LastSeen:     time.Now(),
		RegisteredAt: time.Now(),
		AllocatedIP:  allocatedIP,
	}

	// Сохраняем пира
	pm.peerStore.Store(req.TenantID, req.PeerID, peer)

	pm.logger.Info("Peer registered directly",
		zap.String("peer_id", req.PeerID),
		zap.String("tenant_id", req.TenantID),
		zap.String("allocated_ip", allocatedIP))

	return &DirectPeerRegistrationResponse{
		Success:      true,
		PeerID:       req.PeerID,
		RegisteredAt: peer.RegisteredAt,
		AllocatedIP:  allocatedIP,
	}, nil
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

// generatePeerID генерирует peer ID из server_id
func generatePeerID(serverID string) string {
	if serverID == "" {
		return "peer_" + generateID()
	}
	return "peer_" + serverID
}

// generateRelaySessionID генерирует relay session ID
func generateRelaySessionID() string {
	return "rs_" + generateID()
}

// truncateForLog возвращает первые 8 символов строки для безопасного логирования
func truncateForLog(s string) string {
	if len(s) <= 8 {
		return s
	}
	return s[:8] + "..."
}

// validateToken валидирует JWT токен
func (pm *PeerManager) validateToken(c *gin.Context) (*P2PClaims, error) {
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

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("missing authorization header")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		return nil, fmt.Errorf("invalid authorization format")
	}

	// Используем переданный валидатор (для тестов) или локальную JWT проверку
	if pm.validator != nil {
		claims, err := pm.validator.ValidateP2PToken(token)
		if err != nil {
			return nil, fmt.Errorf("invalid token: %v", err)
		}

		pm.logger.Info("Token validated with external validator",
			zap.String("server_id", claims.ServerID),
			zap.String("tenant_id", claims.TenantID),
			zap.String("subject", claims.Subject))

		return claims, nil
	}

	// Fallback к локальной HS256/JWT проверке с fallback секретом
	claims, err := pm.validateP2PToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	pm.logger.Info("Token validated with local JWT parser",
		zap.String("server_id", claims.ServerID),
		zap.String("tenant_id", claims.TenantID),
		zap.String("subject", claims.Subject))

	return claims, nil
}

// validateP2PToken валидирует P2P JWT токен
func (pm *PeerManager) validateP2PToken(tokenString string) (*P2PClaims, error) {
	// Удаление "Bearer " префикса
	if len(tokenString) > 7 && strings.HasPrefix(tokenString, "Bearer ") {
		tokenString = tokenString[7:]
	}

	// JWT HMAC secret for P2P tokens
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "test-secret"
	}

	// Парсинг токена с HS256 секретом (raw, без base64 декодирования)
	token, err := jwt.ParseWithClaims(tokenString, &P2PClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*P2PClaims)
	if !ok || !token.Valid {
		return nil, err
	}

	// Проверка обязательных полей для P2P mesh
	if claims.ProtocolType != "p2p-mesh" {
		return nil, fmt.Errorf("unsupported protocol: %s", claims.ProtocolType)
	}

	if claims.ConnectionType != "quic" {
		return nil, fmt.Errorf("unsupported connection type: %s", claims.ConnectionType)
	}

	if strings.TrimSpace(claims.TenantID) == "" {
		return nil, fmt.Errorf("missing tenant_id")
	}

	if strings.TrimSpace(claims.ServerID) == "" {
		return nil, fmt.Errorf("missing server_id")
	}

	return claims, nil
}
