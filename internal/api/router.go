package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai"
	"github.com/twogc/cloudbridge-relay-installer/internal/keycloak"
	"go.uber.org/zap"
)

// APIRouter настраивает маршруты API
type APIRouter struct {
	peerManager     *PeerManager
	relayManager    *RelayManager
	healthManager   *HealthManager
	protocolHandler *ProtocolHandler
	hub             *Hub
	logger          *zap.Logger
	aiMiddleware    *AIMiddleware
}

// NewAPIRouter создает новый APIRouter
func NewAPIRouter(logger *zap.Logger, validator TokenValidator) *APIRouter {
	logger.Info("Creating API router components...")

	// Создание общих структур данных
	peers := make(map[string]*Peer)
	connections := make(map[string]*PeerConnection)

	// Создание Hub для WebSocket
	logger.Info("Creating WebSocket Hub...")
	hub := NewHub(logger)

	// Создание менеджеров
	logger.Info("Creating PeerManager...")
	peerManager := NewPeerManager(logger, validator)

	logger.Info("Creating RelayManager...")
	relayManager := NewRelayManager(hub, peers, connections, logger, validator)

	logger.Info("Creating HealthManager...")
	healthManager := NewHealthManager(hub, peers, connections, logger)

	// Создание AI middleware
	logger.Info("Creating AI Middleware...")
	aiClient := ai.NewClient("http://localhost:8083", logger)
	aiConfig := ai.GetDefaultConfig()
	aiMiddleware := NewAIMiddleware(aiClient, aiConfig, logger)

	logger.Info("API router components created successfully")

	return &APIRouter{
		peerManager:   peerManager,
		relayManager:  relayManager,
		healthManager: healthManager,
		hub:           hub,
		logger:        logger,
		aiMiddleware:  aiMiddleware,
	}
}

// NewAPIRouterWithConfig создает новый APIRouter с конфигурацией
func NewAPIRouterWithConfig(logger *zap.Logger, validator TokenValidator, allowedIPs []string) *APIRouter {
	logger.Info("Creating API router components with configuration...")

	// Создание общих структур данных
	peers := make(map[string]*Peer)
	connections := make(map[string]*PeerConnection)

	// Создание Hub для WebSocket
	logger.Info("Creating WebSocket Hub...")
	hub := NewHub(logger)

	// Создание менеджеров с конфигурацией
	logger.Info("Creating PeerManager with configuration...")
	peerManager := NewPeerManagerWithConfig(logger, validator, allowedIPs)

	logger.Info("Creating RelayManager...")
	relayManager := NewRelayManager(hub, peers, connections, logger, validator)

	logger.Info("Creating HealthManager...")
	healthManager := NewHealthManager(hub, peers, connections, logger)

	logger.Info("API router components created successfully with configuration")

	return &APIRouter{
		peerManager:   peerManager,
		relayManager:  relayManager,
		healthManager: healthManager,
		hub:           hub,
		logger:        logger,
	}
}

// NewAPIRouterWithKeycloak создает новый APIRouter с Keycloak валидацией
func NewAPIRouterWithKeycloak(logger *zap.Logger, validator TokenValidator, keycloakValidator *keycloak.KeycloakValidator, allowedIPs []string) *APIRouter {
	logger.Info("Creating API router components with Keycloak validation...")

	// Создание общих структур данных
	peers := make(map[string]*Peer)
	connections := make(map[string]*PeerConnection)

	// Создание Hub для WebSocket
	logger.Info("Creating WebSocket Hub...")
	hub := NewHub(logger)

	// Создание менеджеров с конфигурацией
	logger.Info("Creating PeerManager with configuration...")
	peerManager := NewPeerManagerWithConfig(logger, validator, allowedIPs)

	logger.Info("Creating RelayManager...")
	relayManager := NewRelayManager(hub, peers, connections, logger, validator)

	logger.Info("Creating HealthManager...")
	healthManager := NewHealthManager(hub, peers, connections, logger)

	// Создание ProtocolHandler если есть Keycloak валидатор
	var protocolHandler *ProtocolHandler
	if keycloakValidator != nil {
		logger.Info("Creating ProtocolHandler with Keycloak validation...")
		protocolHandler = NewProtocolHandler(keycloakValidator, logger, peerManager, relayManager)
	}

	logger.Info("API router components created successfully with Keycloak validation")

	return &APIRouter{
		peerManager:     peerManager,
		relayManager:    relayManager,
		healthManager:   healthManager,
		protocolHandler: protocolHandler,
		hub:             hub,
		logger:          logger,
	}
}

// SetupRoutes настраивает все маршруты API
func (r *APIRouter) SetupRoutes(router *gin.Engine) {
	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(r.corsMiddleware())
	router.Use(r.rateLimitMiddleware())

	// API v1 группа
	v1 := router.Group("/api/v1")
	{
		// Peer management endpoints
		peers := v1.Group("/peers")
		{
			peers.POST("/register", r.peerManager.RegisterPeer)
			peers.GET("/discover", r.peerManager.DiscoverPeers)
			peers.PUT("/:peer_id/status", r.peerManager.UpdatePeerStatus)
			peers.GET("/:peer_id", r.peerManager.GetPeerInfo)
		}

		// Relay endpoints
		relay := v1.Group("/relay")
		{
			relay.POST("/route", r.relayManager.RouteThroughRelay)
			relay.GET("/stats", r.relayManager.GetRelayStats)
		}

		// Protocol endpoints (только если есть ProtocolHandler)
		if r.protocolHandler != nil {
			// Quick protocol endpoints
			quick := v1.Group("/quick")
			{
				quick.POST("/connect", gin.WrapF(r.protocolHandler.RegisterClient))
			}

			// P2P Mesh protocol endpoints
			p2p := v1.Group("/p2p")
			{
				p2p.POST("/mesh/join", gin.WrapF(r.protocolHandler.RegisterClient))
			}
		}

		// Health and monitoring endpoints
		health := v1.Group("/health")
		{
			health.GET("/", r.healthManager.HealthCheck)
			health.GET("/metrics", r.healthManager.GetSystemMetrics)
			health.GET("/peers", r.healthManager.GetPeerMetrics)
			health.GET("/connections", r.healthManager.GetConnectionMetrics)
		}
	}

	// WebSocket endpoint
	router.GET("/ws/peers", r.hub.HandleWebSocket)

	// Root health check
	router.GET("/health", r.healthManager.HealthCheck)

	// Metrics endpoint for Prometheus
	router.GET("/metrics", r.healthManager.GetSystemMetrics)
}

// StartHub запускает WebSocket Hub
func (r *APIRouter) StartHub() {
	go r.hub.Run()
	r.logger.Info("WebSocket Hub started")
}

// corsMiddleware настраивает CORS
func (r *APIRouter) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// rateLimitMiddleware настраивает rate limiting
func (r *APIRouter) rateLimitMiddleware() gin.HandlerFunc {
	// Простая реализация rate limiting
	// В реальной реализации здесь должен быть более сложный алгоритм
	return func(c *gin.Context) {
		// Пропускаем все запросы
		c.Next()
	}
}

// GetHub возвращает WebSocket Hub
func (r *APIRouter) GetHub() *Hub {
	return r.hub
}

// GetPeerManager возвращает PeerManager
func (r *APIRouter) GetPeerManager() *PeerManager {
	return r.peerManager
}

// GetRelayManager возвращает RelayManager
func (r *APIRouter) GetRelayManager() *RelayManager {
	return r.relayManager
}

// GetHealthManager возвращает HealthManager
func (r *APIRouter) GetHealthManager() *HealthManager {
	return r.healthManager
}

// MockTokenValidator для тестирования
type MockTokenValidator struct{}

// ValidateP2PToken валидирует P2P токен (mock реализация)
func (m *MockTokenValidator) ValidateP2PToken(token string) (*P2PClaims, error) {
	// Mock реализация - всегда возвращает успех
	return &P2PClaims{
		JTI:            "mock-jti",
		Subject:        "mock-peer-id",
		OrgID:          "mock-org-id",
		TenantID:       "mock-tenant-id",
		DeviceID:       "mock-device-id",
		ConnectionType: "p2p-mesh",
		Permissions:    []string{"read", "write"},
		WireGuardConfig: &WireGuardConfig{
			PublicKey:  "mock-public-key",
			Endpoint:   "mock.example.com:51820",
			ListenPort: 51820,
			AllowedIPs: []string{"10.0.1.0/24"},
		},
		MeshConfig: &MeshConfig{
			AutoDiscovery: true,
			Routing:       "hybrid",
			Encryption:    "aes256",
		},
		PeerWhitelist: &PeerWhitelist{
			AllowedPeers: []string{"peer1", "peer2"},
			AutoApprove:  false,
			MaxPeers:     10,
		},
		NetworkConfig: &NetworkConfig{
			Subnet: "10.0.1.0/24",
			DNS:    []string{"8.8.8.8", "8.8.4.4"},
			MTU:    1500,
		},
		IssuedAt:  time.Now().Unix(),
		NotBefore: time.Now().Unix(),
		Issuer:    "mock-issuer",
		Audience:  "mock-audience",
	}, nil
}
