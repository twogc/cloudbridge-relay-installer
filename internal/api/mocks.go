package api

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/anomaly"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/predictive"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/routing"
	"github.com/twogc/cloudbridge-relay-installer/internal/keycloak"
	"github.com/twogc/cloudbridge-relay-installer/internal/p2p"
)

// MockDataPipeline мок для DataPipeline
type MockDataPipeline struct {
	mock.Mock
}

func (m *MockDataPipeline) GetStatus() (map[string]interface{}, error) {
	args := m.Called()
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockDataPipeline) GetCollectorData() ([]map[string]interface{}, error) {
	args := m.Called()
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *MockDataPipeline) QueryData(query string) ([]map[string]interface{}, error) {
	args := m.Called(query)
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *MockDataPipeline) GetMetrics() (map[string]interface{}, error) {
	args := m.Called()
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockDataPipeline) Start() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDataPipeline) Stop() error {
	args := m.Called()
	return args.Error(0)
}

// MockInputValidator мок для InputValidator
type MockInputValidator struct {
	mock.Mock
}

func (m *MockInputValidator) ValidateQuery(query string) error {
	args := m.Called(query)
	return args.Error(0)
}

// MockAIMiddleware мок для AIMiddleware
type MockAIMiddleware struct {
	mock.Mock
}

func (m *MockAIMiddleware) GetMetrics() *AIMetrics {
	args := m.Called()
	return args.Get(0).(*AIMetrics)
}

func (m *MockAIMiddleware) GetCacheStats() map[string]int {
	args := m.Called()
	return args.Get(0).(map[string]int)
}

func (m *MockAIMiddleware) GetCircuitBreakerStats() map[string]interface{} {
	args := m.Called()
	return args.Get(0).(map[string]interface{})
}

// MockAIClient мок для AI клиента
type MockAIClient struct {
	mock.Mock
}

func (m *MockAIClient) DetectAnomaly(ctx context.Context, metrics *anomaly.NetworkMetrics) (*anomaly.AnomalyResult, error) {
	args := m.Called(ctx, metrics)
	return args.Get(0).(*anomaly.AnomalyResult), args.Error(1)
}

func (m *MockAIClient) PredictLoad(ctx context.Context, horizon int) (*predictive.LoadPrediction, error) {
	args := m.Called(ctx, horizon)
	return args.Get(0).(*predictive.LoadPrediction), args.Error(1)
}

func (m *MockAIClient) OptimizeRoute(ctx context.Context, source, destination string, constraints map[string]interface{}) (*routing.OptimizedRoute, error) {
	args := m.Called(ctx, source, destination, constraints)
	return args.Get(0).(*routing.OptimizedRoute), args.Error(1)
}

func (m *MockAIClient) GetStatus(ctx context.Context) (*ai.ClientServiceStatus, error) {
	args := m.Called(ctx)
	return args.Get(0).(*ai.ClientServiceStatus), args.Error(1)
}

func (m *MockAIClient) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockTokenValidator мок для валидатора токенов
type MockTokenValidator struct {
	mock.Mock
}

func (m *MockTokenValidator) ValidateToken(token string) (*P2PClaims, error) {
	args := m.Called(token)
	return args.Get(0).(*P2PClaims), args.Error(1)
}

func (m *MockTokenValidator) ValidateP2PToken(token string) (*P2PClaims, error) {
	args := m.Called(token)
	return args.Get(0).(*P2PClaims), args.Error(1)
}

// MockHub мок для Hub
type MockHub struct {
	mock.Mock
}

func (m *MockHub) SendRelayMessage(peerID, messageType, data string) bool {
	args := m.Called(peerID, messageType, data)
	return args.Bool(0)
}

func (m *MockHub) GetConnectionCount() int {
	args := m.Called()
	return args.Int(0)
}

// MockHeartbeatManager мок для HeartbeatManager
type MockHeartbeatManager struct {
	mock.Mock
}

func (m *MockHeartbeatManager) GetPeerStatus(peerID string) (*p2p.PeerHeartbeat, error) {
	args := m.Called(peerID)
	return args.Get(0).(*p2p.PeerHeartbeat), args.Error(1)
}

func (m *MockHeartbeatManager) ProcessHeartbeat(peerID, tenantID string) (*p2p.HeartbeatResponse, error) {
	args := m.Called(peerID, tenantID)
	return args.Get(0).(*p2p.HeartbeatResponse), args.Error(1)
}

func (m *MockHeartbeatManager) GetTenantStats(tenantID string) map[string]interface{} {
	args := m.Called(tenantID)
	return args.Get(0).(map[string]interface{})
}

// MockRelayManager мок для RelayManager
type MockRelayManager struct {
	mock.Mock
}

func (m *MockRelayManager) GetRelayStats(tenantID string) (*RelayStats, error) {
	args := m.Called(tenantID)
	return args.Get(0).(*RelayStats), args.Error(1)
}

func (m *MockRelayManager) GetTenantInfo(tenantID string) (*p2p.Tenant, error) {
	args := m.Called(tenantID)
	return args.Get(0).(*p2p.Tenant), args.Error(1)
}

func (m *MockRelayManager) RelayMessage(c interface{}) {
	m.Called(c)
}

func (m *MockRelayManager) RegisterClient(clientID, protocol string) error {
	args := m.Called(clientID, protocol)
	return args.Error(0)
}

func (m *MockRelayManager) GetClientStatus(clientID string) (string, error) {
	args := m.Called(clientID)
	return args.String(0), args.Error(1)
}

// MockPeerManager мок для PeerManager
type MockPeerManager struct {
	mock.Mock
}

func (m *MockPeerManager) RegisterPeerDirect(peer *p2p.Peer) error {
	args := m.Called(peer)
	return args.Error(0)
}

func (m *MockPeerManager) GetPeer(peerID string) (*p2p.Peer, error) {
	args := m.Called(peerID)
	return args.Get(0).(*p2p.Peer), args.Error(1)
}

// MockKeycloakValidator мок для Keycloak валидатора
type MockKeycloakValidator struct {
	mock.Mock
}

func (m *MockKeycloakValidator) ValidateQuickToken(token string) (*keycloak.QuickTokenClaims, error) {
	args := m.Called(token)
	return args.Get(0).(*keycloak.QuickTokenClaims), args.Error(1)
}

func (m *MockKeycloakValidator) ValidateP2PMeshToken(token string) (*keycloak.P2PMeshTokenClaims, error) {
	args := m.Called(token)
	return args.Get(0).(*keycloak.P2PMeshTokenClaims), args.Error(1)
}
