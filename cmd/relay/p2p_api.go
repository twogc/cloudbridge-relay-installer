package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/twogc/cloudbridge-relay-installer/internal/api"
	"github.com/twogc/cloudbridge-relay-installer/internal/token"
)

// P2PAPIServer сервер P2P API
type P2PAPIServer struct {
	router    *gin.Engine
	apiRouter *api.APIRouter
	logger    *zap.Logger
	validator *TokenValidatorAdapter
	server    *http.Server
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
	apiClaims := &api.P2PClaims{
		JTI:               claims.ID,
		Subject:           claims.Subject,
		OrgID:             "", // Нет в token.P2PClaims
		TenantID:          claims.TenantID,
		DeviceID:          "", // Нет в token.P2PClaims
		DeviceFingerprint: "", // Нет в token.P2PClaims
		ConnectionType:    claims.ConnectionType,
		Permissions:       claims.Permissions,
		WireGuardConfig:   convertWireGuardConfig(claims.WireGuardConfig),
		MeshConfig:        convertMeshConfig(claims.MeshConfig),
		PeerWhitelist:     convertPeerWhitelist(claims.PeerWhitelist),
		NetworkConfig:     convertNetworkConfig(claims.NetworkConfig),
		IssuedAt:          claims.IssuedAt.Unix(),
		NotBefore:         claims.NotBefore.Unix(),
		Issuer:            claims.Issuer,
		Audience:          claims.Audience[0], // Берем первый элемент
	}

	return apiClaims, nil
}

// convertWireGuardConfig конвертирует WireGuardConfig
func convertWireGuardConfig(cfg *token.WireGuardConfig) *api.WireGuardConfig {
	if cfg == nil {
		return nil
	}
	return &api.WireGuardConfig{
		PublicKey:  cfg.PublicKey,
		Endpoint:   "", // Нет в token.WireGuardConfig
		ListenPort: 0,  // Нет в token.WireGuardConfig
		AllowedIPs: cfg.AllowedIPs,
	}
}

// convertMeshConfig конвертирует MeshConfig
func convertMeshConfig(cfg *token.MeshConfig) *api.MeshConfig {
	if cfg == nil {
		return nil
	}
	return &api.MeshConfig{
		AutoDiscovery: cfg.EnableAutoDiscovery,
		Routing:       cfg.RoutingStrategy,
		Encryption:    cfg.TrustLevel,
	}
}

// convertPeerWhitelist конвертирует PeerWhitelist
func convertPeerWhitelist(whitelist []string) *api.PeerWhitelist {
	if whitelist == nil {
		return nil
	}
	return &api.PeerWhitelist{
		AllowedPeers: whitelist,
		AutoApprove:  false, // Нет в token.P2PClaims
		MaxPeers:     0,     // Нет в token.P2PClaims
	}
}

// convertNetworkConfig конвертирует NetworkConfig
func convertNetworkConfig(cfg *token.NetworkConfig) *api.NetworkConfig {
	if cfg == nil {
		return nil
	}
	return &api.NetworkConfig{
		Subnet: cfg.Subnet,
		DNS:    cfg.DNS,
		MTU:    cfg.MTU,
	}
}

// NewP2PAPIServer создает новый P2P API сервер
func NewP2PAPIServer(logger *zap.Logger, jwtSecret []byte, port int) *P2PAPIServer {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Panic in NewP2PAPIServer", zap.Any("panic", r))
		}
	}()

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

	// Настройка маршрутов
	apiRouter.SetupRoutes(router)

	// Создание HTTP сервера
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return &P2PAPIServer{
		router:    router,
		apiRouter: apiRouter,
		logger:    logger,
		validator: validator,
		server:    server,
	}
}

// Start запускает P2P API сервер
func (s *P2PAPIServer) Start() error {
	// Запуск WebSocket Hub
	s.apiRouter.StartHub()

	s.logger.Info("Starting P2P API server",
		zap.String("addr", s.server.Addr),
	)

	// Запуск HTTP сервера
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

// Stop останавливает P2P API сервер
func (s *P2PAPIServer) Stop() error {
	s.logger.Info("Stopping P2P API server")

	// Остановка HTTP сервера
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return s.server.Shutdown(ctx)
}

// GetHub возвращает WebSocket Hub
func (s *P2PAPIServer) GetHub() *api.Hub {
	return s.apiRouter.GetHub()
}

// GetPeerManager возвращает PeerManager
func (s *P2PAPIServer) GetPeerManager() *api.PeerManager {
	return s.apiRouter.GetPeerManager()
}

// GetRelayManager возвращает RelayManager
func (s *P2PAPIServer) GetRelayManager() *api.RelayManager {
	return s.apiRouter.GetRelayManager()
}

// GetHealthManager возвращает HealthManager
func (s *P2PAPIServer) GetHealthManager() *api.HealthManager {
	return s.apiRouter.GetHealthManager()
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
