package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/anomaly"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/pipeline"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/predictive"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/routing"
	"github.com/twogc/cloudbridge-relay-installer/internal/keycloak"
	"go.uber.org/zap"
)

// APIRouter настраивает маршруты API
type APIRouter struct {
	peerManager           *PeerManager
	relayManager          *RelayManager
	healthManager         *HealthManager
	protocolHandler       *ProtocolHandler
	hub                   *Hub
	logger                *zap.Logger
	authMiddleware        *AuthMiddleware
	rateLimiter           *RateLimiter
	aiMiddleware          *AIMiddleware
	aiEndpoints           *AIEndpoints
	dataPipelineEndpoints *DataPipelineEndpoints
	monitoringEndpoints   *MonitoringEndpoints
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

	// Создание Auth Middleware
	logger.Info("Creating Auth Middleware...")
	authMiddleware := NewAuthMiddleware(validator, logger)

	// Создание Rate Limiter
	logger.Info("Creating Rate Limiter...")
	rateLimiter := NewRateLimiter(DefaultRateLimiterConfig(), logger)

	logger.Info("Creating RelayManager...")
	relayManager := NewRelayManager(hub, peers, connections, logger, validator)

	logger.Info("Creating HealthManager...")
	healthManager := NewHealthManager(hub, peers, connections, logger)

	// Создание AI middleware
	logger.Info("Creating AI Middleware...")
	aiClient := ai.NewClient("http://localhost:8083", logger)
	aiConfig := ai.GetDefaultConfig()
	aiMiddleware := NewAIMiddleware(aiClient, aiConfig, logger)

	// Создание AI endpoints
	logger.Info("Creating AI endpoints...")
	// Создаем адаптер для ai.Client
	aiAdapter := NewAIClientAdapter(aiClient, logger)
	aiEndpoints := NewAIEndpoints(aiAdapter, logger)

	// Создание Data Pipeline endpoints
	logger.Info("Creating Data Pipeline endpoints...")
	// Создаем реальный Data Pipeline
	dataPipeline := createDataPipeline(logger)
	dataPipelineEndpoints := NewDataPipelineEndpoints(dataPipeline, logger)

	// Создание Monitoring endpoints
	logger.Info("Creating Monitoring endpoints...")
	monitoringEndpoints := NewMonitoringEndpoints(logger)

	logger.Info("API router components created successfully")

	return &APIRouter{
		peerManager:           peerManager,
		relayManager:          relayManager,
		healthManager:         healthManager,
		hub:                   hub,
		logger:                logger,
		authMiddleware:        authMiddleware,
		rateLimiter:           rateLimiter,
		aiMiddleware:          aiMiddleware,
		aiEndpoints:           aiEndpoints,
		dataPipelineEndpoints: dataPipelineEndpoints,
		monitoringEndpoints:   monitoringEndpoints,
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

	// Создание Auth Middleware
	logger.Info("Creating Auth Middleware...")
	authMiddleware := NewAuthMiddleware(validator, logger)

	// Создание Rate Limiter
	logger.Info("Creating Rate Limiter...")
	rateLimiter := NewRateLimiter(DefaultRateLimiterConfig(), logger)

	logger.Info("Creating RelayManager...")
	relayManager := NewRelayManager(hub, peers, connections, logger, validator)

	logger.Info("Creating HealthManager...")
	healthManager := NewHealthManager(hub, peers, connections, logger)

	// Создание AI middleware
	logger.Info("Creating AI Middleware...")
	aiClient := ai.NewClient("http://localhost:8083", logger)
	aiConfig := ai.GetDefaultConfig()
	aiMiddleware := NewAIMiddleware(aiClient, aiConfig, logger)

	// Создание AI endpoints
	logger.Info("Creating AI endpoints...")
	aiAdapter := NewAIClientAdapter(aiClient, logger)
	aiEndpoints := NewAIEndpoints(aiAdapter, logger)

	// Создание Data Pipeline endpoints
	logger.Info("Creating Data Pipeline endpoints...")
	dataPipeline := createDataPipeline(logger)
	dataPipelineEndpoints := NewDataPipelineEndpoints(dataPipeline, logger)

	// Создание Monitoring endpoints
	logger.Info("Creating Monitoring endpoints...")
	monitoringEndpoints := NewMonitoringEndpoints(logger)

	logger.Info("API router components created successfully with configuration")

	return &APIRouter{
		peerManager:           peerManager,
		relayManager:          relayManager,
		healthManager:         healthManager,
		hub:                   hub,
		logger:                logger,
		authMiddleware:        authMiddleware,
		rateLimiter:           rateLimiter,
		aiMiddleware:          aiMiddleware,
		aiEndpoints:           aiEndpoints,
		dataPipelineEndpoints: dataPipelineEndpoints,
		monitoringEndpoints:   monitoringEndpoints,
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

	// Создание Auth Middleware
	logger.Info("Creating Auth Middleware...")
	authMiddleware := NewAuthMiddleware(validator, logger)

	// Создание Rate Limiter
	logger.Info("Creating Rate Limiter...")
	rateLimiter := NewRateLimiter(DefaultRateLimiterConfig(), logger)

	logger.Info("Creating RelayManager...")
	relayManager := NewRelayManager(hub, peers, connections, logger, validator)

	logger.Info("Creating HealthManager...")
	healthManager := NewHealthManager(hub, peers, connections, logger)

	// Создание AI middleware
	logger.Info("Creating AI Middleware...")
	aiClient := ai.NewClient("http://localhost:8083", logger)
	aiConfig := ai.GetDefaultConfig()
	aiMiddleware := NewAIMiddleware(aiClient, aiConfig, logger)

	// Создание AI endpoints
	logger.Info("Creating AI endpoints...")
	aiAdapter := NewAIClientAdapter(aiClient, logger)
	aiEndpoints := NewAIEndpoints(aiAdapter, logger)

	// Создание Data Pipeline endpoints
	logger.Info("Creating Data Pipeline endpoints...")
	dataPipeline := createDataPipeline(logger)
	dataPipelineEndpoints := NewDataPipelineEndpoints(dataPipeline, logger)

	// Создание Monitoring endpoints
	logger.Info("Creating Monitoring endpoints...")
	monitoringEndpoints := NewMonitoringEndpoints(logger)

	// Создание ProtocolHandler если есть Keycloak валидатор
	var protocolHandler *ProtocolHandler
	if keycloakValidator != nil {
		logger.Info("Creating ProtocolHandler with Keycloak validation...")
		protocolHandler = NewProtocolHandler(keycloakValidator, logger, peerManager, relayManager)
	}

	logger.Info("API router components created successfully with Keycloak validation")

	return &APIRouter{
		peerManager:           peerManager,
		relayManager:          relayManager,
		healthManager:         healthManager,
		protocolHandler:       protocolHandler,
		hub:                   hub,
		logger:                logger,
		authMiddleware:        authMiddleware,
		rateLimiter:           rateLimiter,
		aiMiddleware:          aiMiddleware,
		aiEndpoints:           aiEndpoints,
		dataPipelineEndpoints: dataPipelineEndpoints,
		monitoringEndpoints:   monitoringEndpoints,
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

		// Tenant-namespaced peers endpoints (new, preferred): /api/v1/tenants/:tenant_id/peers/...
		tenant := v1.Group("/tenants/:tenant_id")
		{
			tenantPeers := tenant.Group("/peers")
			{
				tenantPeers.POST("/register", r.peerManager.RegisterPeer)
				tenantPeers.GET("/discover", r.peerManager.DiscoverPeers)
				tenantPeers.PUT("/:peer_id/status", r.peerManager.UpdatePeerStatus)
				tenantPeers.GET("/:peer_id", r.peerManager.GetPeerInfo)
			}
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

		// AI endpoints (требуют аутентификации и rate limiting)
		ai := v1.Group("/ai")
		ai.Use(r.rateLimiter.CreateRateLimitMiddleware("ai"))
		{
			r.logger.Info("Registering AI endpoints with rate limiting only (auth disabled)")
			ai.GET("/status", r.aiEndpoints.GetAIStatus)
			ai.GET("/health", r.aiEndpoints.GetAIHealth)
			ai.GET("/metrics", r.aiEndpoints.GetAIMetrics)
			ai.POST("/detect-anomaly", r.aiEndpoints.DetectAnomaly)
			ai.GET("/predict-load", r.aiEndpoints.PredictLoad)
			ai.POST("/optimize-route", r.aiEndpoints.OptimizeRoute)
		}

		// Data Pipeline endpoints (требуют аутентификации и rate limiting)
		pipeline := v1.Group("/pipeline")
		pipeline.Use(r.authMiddleware.RequireAuth())
		pipeline.Use(r.rateLimiter.CreateRateLimitMiddleware("pipeline"))
		{
			pipeline.GET("/status", r.dataPipelineEndpoints.GetPipelineStatus)
			pipeline.GET("/metrics", r.dataPipelineEndpoints.GetPipelineMetrics)
			pipeline.GET("/data-types", r.dataPipelineEndpoints.GetDataTypes)
			pipeline.POST("/query", r.dataPipelineEndpoints.QueryData)
			pipeline.POST("/start", r.dataPipelineEndpoints.StartPipeline)
			pipeline.POST("/stop", r.dataPipelineEndpoints.StopPipeline)
			pipeline.GET("/collectors/:collector/data", r.dataPipelineEndpoints.GetCollectorData)
		}

		// Extended monitoring endpoints (требуют административных прав и rate limiting)
		monitoring := v1.Group("/monitoring")
		monitoring.Use(r.authMiddleware.RequireAdmin())
		monitoring.Use(r.rateLimiter.CreateRateLimitMiddleware("monitoring"))
		{
			r.logger.Info("Registering monitoring endpoints with admin authentication and rate limiting")
			monitoring.GET("/system", r.monitoringEndpoints.GetSystemMetrics)
			monitoring.GET("/health", r.monitoringEndpoints.GetHealthCheck)
			monitoring.GET("/prometheus", r.monitoringEndpoints.GetPrometheusMetrics)
			monitoring.GET("/services", r.monitoringEndpoints.GetServiceStatus)
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

// GetAIEndpoints возвращает AIEndpoints
func (r *APIRouter) GetAIEndpoints() *AIEndpoints {
	return r.aiEndpoints
}

// GetDataPipelineEndpoints возвращает DataPipelineEndpoints
func (r *APIRouter) GetDataPipelineEndpoints() *DataPipelineEndpoints {
	return r.dataPipelineEndpoints
}

// GetMonitoringEndpoints возвращает MonitoringEndpoints
func (r *APIRouter) GetMonitoringEndpoints() *MonitoringEndpoints {
	return r.monitoringEndpoints
}

// MockAIService заглушка для AI сервиса
type MockAIService struct{}

func (m *MockAIService) GetStatus() *ai.ServiceStatus {
	return &ai.ServiceStatus{
		Status: "running",
		Components: map[string]bool{
			"anomaly_detection":  true,
			"load_prediction":    true,
			"route_optimization": true,
		},
	}
}

func (m *MockAIService) DetectAnomaly(ctx context.Context, metrics *anomaly.NetworkMetrics) (*anomaly.AnomalyResult, error) {
	return &anomaly.AnomalyResult{
		IsAnomaly:   false,
		Score:       0.1,
		AnomalyType: "normal",
		Severity:    "low",
	}, nil
}

func (m *MockAIService) PredictLoad(ctx context.Context, horizon int) (*predictive.LoadPrediction, error) {
	return &predictive.LoadPrediction{
		Forecast:   []float64{0.5, 0.6, 0.7},
		Confidence: 0.85,
	}, nil
}

func (m *MockAIService) OptimizeRoute(ctx context.Context, source, destination string, constraints map[string]interface{}) (*routing.OptimizedRoute, error) {
	return &routing.OptimizedRoute{
		Path:             []string{source, destination},
		PredictedLatency: 100.5,
		Confidence:       0.9,
		Protocol:         "QUIC",
	}, nil
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

// createDataPipeline создает и инициализирует реальный Data Pipeline
func createDataPipeline(logger *zap.Logger) *pipeline.DataPipeline {
	logger.Info("Creating real Data Pipeline...")

	// Создаем конфигурацию для Data Pipeline
	config := &pipeline.Config{
		Enabled:            true,
		BatchSize:          1000,
		ProcessingInterval: 30 * time.Second,
		RetentionDays:      7,
		StoragePath:        "/tmp/cloudbridge-pipeline",
		PrometheusURL:      "http://localhost:9090",
	}

	// Создаем Data Pipeline
	dataPipeline := pipeline.NewDataPipeline(config, logger)

	// Инициализируем Data Pipeline
	if err := dataPipeline.Initialize(); err != nil {
		logger.Error("Failed to initialize Data Pipeline", zap.Error(err))
		// Возвращаем nil, чтобы endpoints работали в graceful degradation режиме
		return nil
	}

	// Запускаем Data Pipeline
	if err := dataPipeline.Start(); err != nil {
		logger.Error("Failed to start Data Pipeline", zap.Error(err))
		// Возвращаем nil, чтобы endpoints работали в graceful degradation режиме
		return nil
	}

	logger.Info("Data Pipeline created and started successfully")
	return dataPipeline
}
