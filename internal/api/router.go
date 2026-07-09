package api

import (
	"github.com/twogc/cloudbridge-relay-installer/internal/p2p"
	"go.uber.org/zap"
)

// APIRouter основной роутер API
type APIRouter struct {
	logger                *zap.Logger
	peerManager           *PeerManager
	relayManager          *RelayManager
	healthManager         *HealthManager
	p2pManager            *PeerManager
	iceManager            *PeerManager
	hub                   *Hub
	authMiddleware        interface{}
	rateLimiter           interface{}
	aiMiddleware          interface{}
	aiEndpoints           *AIEndpoints
	dataPipelineEndpoints *DataPipelineEndpoints
	monitoringEndpoints   *MonitoringEndpoints
	heartbeatHandler      *HeartbeatHandler
	wireguardEndpoints    *WireGuardSimpleEndpoints
}

// NewAPIRouter создает новый APIRouter
func NewAPIRouter(logger *zap.Logger, validator TokenValidator) *APIRouter {
	// Создаем простые заглушки для тестов
	hub := &Hub{}
	peers := make(map[string]*Peer)
	connections := make(map[string]*PeerConnection)

	// Создаем менеджеры
	peerManager := NewPeerManager(logger, validator)
	relayManager := NewRelayManager(hub, peers, connections, logger, validator)

	// Создаем WireGuard endpoints (пока с заглушками)
	wireguardEndpoints := NewWireGuardSimpleEndpoints(
		peerManager,
		nil, // RelayManager будет nil пока не интегрируем с p2p.RelayManager
		nil, // RelayManagerHybrid будет nil пока не интегрируем
		nil, // WireGuardTenantManager будет nil пока не интегрируем
		nil, // ConfigManager будет nil пока не интегрируем
		logger,
	)

	return &APIRouter{
		logger:                logger,
		peerManager:           peerManager,
		relayManager:          relayManager,
		healthManager:         NewHealthManager(hub, peers, connections, logger),
		p2pManager:            NewPeerManager(logger, validator),
		iceManager:            NewPeerManager(logger, validator),
		hub:                   hub,
		authMiddleware:        nil,                                   // Заглушка
		rateLimiter:           nil,                                   // Заглушка
		aiMiddleware:          nil,                                   // Заглушка
		aiEndpoints:           NewAIEndpoints(nil, logger),           // nil для AIServiceInterface
		dataPipelineEndpoints: NewDataPipelineEndpoints(nil, logger), // nil для DataPipeline
		monitoringEndpoints:   NewMonitoringEndpoints(logger),
		heartbeatHandler:      NewHeartbeatHandler(nil, nil, logger), // nil для HeartbeatManager и RelayManager
		wireguardEndpoints:    wireguardEndpoints,
	}
}

// NewAPIRouterWithConfig создает новый APIRouter с конфигурацией
func NewAPIRouterWithConfig(logger *zap.Logger, validator TokenValidator, allowedIPs []string) *APIRouter {
	return NewAPIRouter(logger, validator)
}

// SetupRoutes настраивает маршруты
func (r *APIRouter) SetupRoutes(engine interface{}) {
	// Заглушка для тестов
}

// SetupSecurityMiddleware настраивает security middleware
func (r *APIRouter) SetupSecurityMiddleware(engine interface{}, env string, config interface{}) {
	// Заглушка для тестов
}

// StartHub запускает Hub
func (r *APIRouter) StartHub() {
	if r.hub != nil {
		go r.hub.Run()
	}
}

// GetHub возвращает Hub
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

// GetWireGuardEndpoints возвращает WireGuardEndpoints
func (r *APIRouter) GetWireGuardEndpoints() *WireGuardSimpleEndpoints {
	return r.wireguardEndpoints
}

// SetRelayManagerHybrid attaches SCORE-backed hybrid manager to WireGuard endpoints.
func (r *APIRouter) SetRelayManagerHybrid(hybrid *p2p.RelayManagerHybrid) {
	if r.wireguardEndpoints != nil {
		r.wireguardEndpoints.SetRelayManagerHybrid(hybrid)
	}
}

// rateLimitMiddleware возвращает rate limiter middleware
func (r *APIRouter) rateLimitMiddleware() interface{} {
	return r.rateLimiter
}
