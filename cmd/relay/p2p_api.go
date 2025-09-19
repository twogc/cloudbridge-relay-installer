package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/twogc/cloudbridge-relay-installer/internal/api"
	"github.com/twogc/cloudbridge-relay-installer/internal/keycloak"
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

// NewP2PAPIServer создает новый P2P API сервер
func NewP2PAPIServer(logger *zap.Logger, jwtSecret []byte, port int) *P2PAPIServer {
	return NewP2PAPIServerWithConfig(logger, jwtSecret, port, nil)
}

// NewP2PAPIServerWithConfig создает новый P2P API сервер с конфигурацией
func NewP2PAPIServerWithConfig(logger *zap.Logger, jwtSecret []byte, port int, allowedIPs []string) *P2PAPIServer {
	return NewP2PAPIServerWithKeycloak(logger, jwtSecret, port, allowedIPs, nil)
}

// NewP2PAPIServerWithKeycloak создает новый P2P API сервер с Keycloak валидацией
func NewP2PAPIServerWithKeycloak(logger *zap.Logger, jwtSecret []byte, port int, allowedIPs []string, keycloakValidator *keycloak.KeycloakValidator) *P2PAPIServer {
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
	var apiRouter *api.APIRouter
	if keycloakValidator != nil {
		logger.Info("Creating API router with Keycloak validation")
		apiRouter = api.NewAPIRouterWithKeycloak(logger, validator, keycloakValidator, allowedIPs)
	} else if len(allowedIPs) > 0 {
		logger.Info("Creating API router with IP configuration", zap.Int("allowed_ips_count", len(allowedIPs)))
		apiRouter = api.NewAPIRouterWithConfig(logger, validator, allowedIPs)
	} else {
		logger.Info("Creating API router with default configuration")
		apiRouter = api.NewAPIRouter(logger, validator)
	}
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
		Addr:         fmt.Sprintf("0.0.0.0:%d", port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Создание Peer Manager
	peerManager := api.NewPeerManager(logger, validator)

	return &P2PAPIServer{
		router:         router,
		apiRouter:      apiRouter,
		logger:         logger,
		validator:      validator,
		tokenValidator: tokenValidator,
		peerManager:    peerManager,
		server:         server,
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
