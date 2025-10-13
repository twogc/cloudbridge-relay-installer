package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/twogc/cloudbridge-relay-installer/internal/api"
	"github.com/twogc/cloudbridge-relay-installer/internal/api/compatibility"
	"github.com/twogc/cloudbridge-relay-installer/internal/p2p"
	"github.com/twogc/cloudbridge-relay-installer/internal/token"
)

// P2PAPIServer сервер P2P API
type P2PAPIServer struct {
	router           *gin.Engine
	apiRouter        *api.APIRouter
	logger           *zap.Logger
	validator        *TokenValidatorAdapter
	tokenValidator   *token.Validator
	peerManager      *api.PeerManager
	server           *http.Server
	relayManager     *p2p.RelayManager
	heartbeatManager *p2p.HeartbeatManager
}

// TokenValidatorAdapter адаптер для token.Validator
type TokenValidatorAdapter struct {
	validator *token.Validator
}

// ValidateP2PToken реализует интерфейс api.TokenValidator
func (t *TokenValidatorAdapter) ValidateP2PToken(tokenString string) (*api.P2PClaims, error) {
	claims, err := t.validator.ValidateP2PToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Конвертируем token.P2PClaims в api.P2PClaims
	// Безопасно обрабатываем необязательные поля aud/iat/nbf, которые могут отсутствовать
	var audience string
	if len(claims.Audience) > 0 {
		audience = claims.Audience[0]
	}

	var issuedAt int64
	if claims.IssuedAt != nil {
		issuedAt = claims.IssuedAt.Unix()
	}

	var notBefore int64
	if claims.NotBefore != nil {
		notBefore = claims.NotBefore.Unix()
	}
	apiClaims := &api.P2PClaims{
		JTI:            claims.ID,
		Subject:        claims.Subject,
		TenantID:       claims.TenantID,
		ServerID:       claims.ServerID,
		ConnectionType: claims.ConnectionType,
		ProtocolType:   claims.ProtocolType,
		Permissions:    claims.Permissions,
		IssuedAt:       issuedAt,
		NotBefore:      notBefore,
		Issuer:         claims.Issuer,
		Audience:       audience,
	}

	return apiClaims, nil
}

// NewP2PAPIServerWithConfig создает новый P2P API сервер с конфигурацией
func NewP2PAPIServerWithConfig(logger *zap.Logger, jwtSecret []byte, port int, allowedIPs []string) *P2PAPIServer {
	return NewP2PAPIServer(logger, jwtSecret, port)
}

// NewP2PAPIServer создает новый P2P API сервер
func NewP2PAPIServer(logger *zap.Logger, jwtSecret []byte, port int) *P2PAPIServer {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Panic in NewP2PAPIServer", zap.Any("panic", r))
		}
	}()

	// JWT secret sanity check
	if len(jwtSecret) == 0 {
		logger.Fatal("JWT secret is empty")
	}

	logger.Info("Creating token validator...")
	// Создание валидатора токенов
	tokenValidator := token.NewValidator(token.Config{
		JWTSecret: string(jwtSecret),
	})
	logger.Info("Token validator created successfully")

	logger.Info("Creating token validator adapter...")

	// Создание адаптера
	validator := &TokenValidatorAdapter{validator: tokenValidator}

	// Создание API router
	logger.Info("Creating API router...")
	apiRouter := api.NewAPIRouter(logger, validator)
	if apiRouter == nil {
		logger.Error("Failed to create API router")
		return nil
	}
	logger.Info("API router created successfully")

	// Настройка Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Add Client Compatibility Layer middleware
	logger.Info("Initializing Client Compatibility Layer")
	router.Use(compatibility.TenantIDMiddleware())
	router.Use(compatibility.LogCompatibilityMiddleware(logger))

	// Initialize compatibility metrics
	compatMetrics := &compatibility.CompatibilityMetrics{
		PathTranslations: make(map[string]int64),
	}
	router.Use(compatibility.MetricsMiddleware(compatMetrics))

	// Настройка маршрутов для P2P API
	// Health endpoints
	ccl := compatibility.NewClientCompatibilityLayer(logger)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":             "healthy",
			"timestamp":          time.Now().Format(time.RFC3339),
			"version":            "1.0.0",
			"compatibility_layer": ccl.HealthCheck(),
		})
	})

	// Smart ready endpoint with dependency checks
	router.GET("/ready", func(c *gin.Context) {
		ready := apiRouter != nil &&
			apiRouter.GetPeerManager() != nil &&
			apiRouter.GetRelayManager() != nil

		status := "degraded"
		if ready {
			status = "ready"
		}

		c.JSON(200, gin.H{
			"status": status,
			"services": gin.H{
				"http":     true,
				"p2p_quic": apiRouter.GetRelayManager() != nil,
				"quic":     false,
			},
		})
	})

	// P2P API endpoints
	v1 := router.Group("/api/v1")
	{
		// Tenant endpoints
		tenants := v1.Group("/tenants/:tenant_id")
		{
			// Peer registration with nil check
			tenants.POST("/peers/register", func(c *gin.Context) {
				if apiRouter.GetPeerManager() == nil {
					c.JSON(http.StatusServiceUnavailable, gin.H{"error": "peer manager unavailable"})
					return
				}
				apiRouter.GetPeerManager().RegisterPeer(c)
			})

			tenants.GET("/peers", func(c *gin.Context) {
				if apiRouter.GetPeerManager() == nil {
					c.JSON(http.StatusServiceUnavailable, gin.H{"error": "peer manager unavailable"})
					return
				}
				apiRouter.GetPeerManager().DiscoverPeers(c)
			})

			tenants.GET("/peers/:peer_id", func(c *gin.Context) {
				if apiRouter.GetPeerManager() == nil {
					c.JSON(http.StatusServiceUnavailable, gin.H{"error": "peer manager unavailable"})
					return
				}
				apiRouter.GetPeerManager().GetPeerInfo(c)
			})

			tenants.PUT("/peers/:peer_id/status", func(c *gin.Context) {
				if apiRouter.GetPeerManager() == nil {
					c.JSON(http.StatusServiceUnavailable, gin.H{"error": "peer manager unavailable"})
					return
				}
				apiRouter.GetPeerManager().UpdatePeerStatus(c)
			})

			// WireGuard configuration endpoint
			tenants.GET("/peers/:peer_id/wireguard-config", func(c *gin.Context) {
				if apiRouter.GetWireGuardEndpoints() == nil {
					c.JSON(http.StatusServiceUnavailable, gin.H{"error": "wireguard endpoints unavailable"})
					return
				}
				apiRouter.GetWireGuardEndpoints().GetWireGuardConfig(c)
			})
		}

		// P2P endpoints
		p2pGrp := v1.Group("/p2p")
		{
			p2pGrp.GET("/discover", func(c *gin.Context) {
				if apiRouter.GetPeerManager() == nil {
					c.JSON(http.StatusServiceUnavailable, gin.H{"error": "peer manager unavailable"})
					return
				}
				apiRouter.GetPeerManager().DiscoverPeers(c)
			})
		}

		// Relay endpoints
		relay := v1.Group("/relay")
		{
			relay.POST("/route", func(c *gin.Context) {
				if apiRouter.GetRelayManager() == nil {
					c.JSON(http.StatusServiceUnavailable, gin.H{"error": "relay manager unavailable"})
					return
				}
				apiRouter.GetRelayManager().RouteThroughRelay(c)
			})

			relay.GET("/stats", func(c *gin.Context) {
				if apiRouter.GetRelayManager() == nil {
					c.JSON(http.StatusServiceUnavailable, gin.H{"error": "relay manager unavailable"})
					return
				}
				apiRouter.GetRelayManager().GetRelayStats(c)
			})
		}
	}

	// Создание HTTP сервера с улучшенными таймаутами
	server := &http.Server{
		Addr:              fmt.Sprintf("0.0.0.0:%d", port),
		Handler:           router,
		ReadTimeout:       30 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1MB
	}

	// Peer Manager уже создан в apiRouter

	return &P2PAPIServer{
		router:         router,
		apiRouter:      apiRouter,
		logger:         logger,
		validator:      validator,
		tokenValidator: tokenValidator,
		peerManager:    apiRouter.GetPeerManager(),
		server:         server,
	}
}

// Start запускает P2P API сервер
func (s *P2PAPIServer) Start() error {
	s.logger.Info("Starting P2P API server", zap.String("addr", s.server.Addr))

	// Запуск HTTP сервера в горутине для неблокирующего старта
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal("HTTP server error", zap.Error(err))
		}
	}()

	return nil
}

// Stop останавливает P2P API сервер
func (s *P2PAPIServer) Stop() error {
	s.logger.Info("Stopping P2P API server")

	// Остановка HTTP сервера
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("HTTP server shutdown error", zap.Error(err))
		return err
	}

	s.logger.Info("P2P API server stopped")
	return nil
}

// GetHub возвращает WebSocket Hub
func (s *P2PAPIServer) GetHub() *api.Hub {
	return s.apiRouter.GetHub()
}

// GetPeerManager возвращает PeerManager
func (s *P2PAPIServer) GetPeerManager() *api.PeerManager {
	return s.peerManager
}

// GetRelayManager возвращает RelayManager
func (s *P2PAPIServer) GetRelayManager() *api.RelayManager {
	return s.apiRouter.GetRelayManager()
}

// GetHealthManager возвращает HealthManager
func (s *P2PAPIServer) GetHealthManager() *api.HealthManager {
	return s.apiRouter.GetHealthManager()
}

// SetManagers устанавливает P2P менеджеры
func (s *P2PAPIServer) SetManagers(relayManager *p2p.RelayManager, heartbeatManager *p2p.HeartbeatManager) {
	s.relayManager = relayManager
	s.heartbeatManager = heartbeatManager
	s.logger.Info("P2P managers set for API server")
}

// GetTokenValidator возвращает token validator
func (s *P2PAPIServer) GetTokenValidator() *token.Validator {
	return s.tokenValidator
}

// P2PAPIServerConfig конфигурация P2P API сервера
type P2PAPIServerConfig struct {
	Port      int    `json:"port"`
	Host      string `json:"host"`
	JWTSecret string `json:"jwt_secret"`
	Enabled   bool   `json:"enabled"`
}

// DefaultP2PAPIServerConfig возвращает конфигурацию по умолчанию
func DefaultP2PAPIServerConfig() *P2PAPIServerConfig {
	return &P2PAPIServerConfig{
		Port:      8080,
		Host:      "0.0.0.0",
		JWTSecret: "your-jwt-secret-key",
		Enabled:   true,
	}
}
