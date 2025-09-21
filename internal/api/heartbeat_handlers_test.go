package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/twogc/cloudbridge-relay-installer/internal/p2p"
	"go.uber.org/zap"
)

// HeartbeatManagerInterface интерфейс для HeartbeatManager
type HeartbeatManagerInterface interface {
	ProcessHeartbeat(peerID, tenantID string) (*p2p.HeartbeatResponse, error)
	GetPeerStatus(peerID string) (*p2p.PeerHeartbeat, error)
	GetTenantStats(tenantID string) map[string]interface{}
}

// RelayManagerInterface интерфейс для RelayManager
type RelayManagerInterface interface {
	GetRelayStats(tenantID string) (*RelayStats, error)
	GetTenantInfo(tenantID string) (*p2p.Tenant, error)
}

// TestHeartbeatHandler тестовая версия хендлера с интерфейсами
type TestHeartbeatHandler struct {
	heartbeatManager HeartbeatManagerInterface
	relayManager     RelayManagerInterface
	logger           *zap.Logger
}

// HandleHeartbeat обрабатывает heartbeat запрос
func (hh *TestHeartbeatHandler) HandleHeartbeat(c *gin.Context) {
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

	c.JSON(http.StatusOK, response)
}

// GetPeerStatus возвращает статус пира
func (hh *TestHeartbeatHandler) GetPeerStatus(c *gin.Context) {
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
func (hh *TestHeartbeatHandler) GetTenantStats(c *gin.Context) {
	tenantID := c.Param("tenant_id")

	if tenantID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing required parameter",
			Code:    http.StatusBadRequest,
			Message: "tenant_id is required",
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
			"tenant_id":       tenantInfo.TenantID,
			"network_cidr":    tenantInfo.NetworkCIDR,
			"max_peers":       tenantInfo.MaxPeers,
			"max_connections": tenantInfo.MaxConnections,
			"created_at":      tenantInfo.CreatedAt.Unix(),
			"last_activity":   tenantInfo.LastActivity.Unix(),
		}
	}

	c.JSON(http.StatusOK, stats)
}

// GetSystemStats возвращает системную статистику
func (hh *TestHeartbeatHandler) GetSystemStats(c *gin.Context) {
	// Получаем статистику
	stats := hh.heartbeatManager.GetTenantStats("system")

	// Получаем информацию о relay
	relayStats, err := hh.relayManager.GetRelayStats("system")
	if err != nil {
		hh.logger.Error("Failed to get relay stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	stats["relay_stats"] = relayStats
	c.JSON(http.StatusOK, stats)
}

func TestNewHeartbeatHandler(t *testing.T) {
	logger := zap.NewNop()
	heartbeatManager := &MockHeartbeatManager{}
	relayManager := &MockRelayManager{}

	handler := &TestHeartbeatHandler{
		heartbeatManager: heartbeatManager,
		relayManager:     relayManager,
		logger:           logger,
	}

	assert.NotNil(t, handler)
	assert.Equal(t, heartbeatManager, handler.heartbeatManager)
	assert.Equal(t, relayManager, handler.relayManager)
	assert.Equal(t, logger, handler.logger)
}

func TestHeartbeatHandler_ProcessHeartbeat(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockHeartbeatManager)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful heartbeat processing",
			requestBody: map[string]interface{}{
				"peer_id":   "peer1",
				"timestamp": time.Now().Unix(),
			},
			mockSetup: func(m *MockHeartbeatManager) {
				m.On("ProcessHeartbeat", "peer1", "tenant1").Return(&p2p.HeartbeatResponse{
					Status:    "ok",
					Timestamp: 0,
					Online:    true,
					Message:   "Heartbeat processed successfully",
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"ok","timestamp":0,"online":true,"message":"Heartbeat processed successfully"}`,
		},
		{
			name: "heartbeat processing error",
			requestBody: map[string]interface{}{
				"peer_id":   "peer1",
				"timestamp": time.Now().Unix(),
			},
			mockSetup: func(m *MockHeartbeatManager) {
				m.On("ProcessHeartbeat", "peer1", "tenant1").Return((*p2p.HeartbeatResponse)(nil), assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Heartbeat processing failed","code":500,"message":"assert.AnError general error for testing"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			requestBody, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest("POST", "/api/v1/tenants/tenant1/peers/peer1/heartbeat", strings.NewReader(string(requestBody)))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{
				{Key: "tenant_id", Value: "tenant1"},
				{Key: "peer_id", Value: "peer1"},
			}

			heartbeatManager := &MockHeartbeatManager{}
			tt.mockSetup(heartbeatManager)

			handler := &TestHeartbeatHandler{
				heartbeatManager: heartbeatManager,
				relayManager:     &MockRelayManager{},
				logger:           zap.NewNop(),
			}

			handler.HandleHeartbeat(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
			heartbeatManager.AssertExpectations(t)
		})
	}
}

func TestHeartbeatHandler_GetPeerStatus(t *testing.T) {
	tests := []struct {
		name           string
		peerID         string
		mockSetup      func(*MockHeartbeatManager)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "successful status retrieval",
			peerID: "peer1",
			mockSetup: func(m *MockHeartbeatManager) {
				m.On("GetPeerStatus", "peer1").Return(&p2p.PeerHeartbeat{
					PeerID:         "peer1",
					TenantID:       "tenant1",
					IsOnline:       true,
					LastSeen:       time.Unix(0, 0),
					LastHeartbeat:  time.Unix(0, 0),
					HeartbeatCount: 5,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"peer_id":"peer1","tenant_id":"tenant1","is_online":true,"last_heartbeat":0,"heartbeat_count":5,"last_seen":0,"uptime":0}`,
		},
		{
			name:   "peer not found",
			peerID: "nonexistent",
			mockSetup: func(m *MockHeartbeatManager) {
				m.On("GetPeerStatus", "nonexistent").Return((*p2p.PeerHeartbeat)(nil), assert.AnError)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"Peer not found","code":404,"message":"assert.AnError general error for testing"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{
				{Key: "tenant_id", Value: "tenant1"},
				{Key: "peer_id", Value: tt.peerID},
			}
			c.Request = httptest.NewRequest("GET", "/heartbeat/status/tenant1/"+tt.peerID, nil)

			heartbeatManager := &MockHeartbeatManager{}
			tt.mockSetup(heartbeatManager)

			handler := &TestHeartbeatHandler{
				heartbeatManager: heartbeatManager,
				relayManager:     &MockRelayManager{},
				logger:           zap.NewNop(),
			}

			handler.GetPeerStatus(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
			heartbeatManager.AssertExpectations(t)
		})
	}
}

func TestHeartbeatHandler_GetTenantStats(t *testing.T) {
	tests := []struct {
		name           string
		tenantID       string
		mockSetup      func(*MockHeartbeatManager, *MockRelayManager)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:     "successful stats retrieval",
			tenantID: "tenant1",
			mockSetup: func(hm *MockHeartbeatManager, rm *MockRelayManager) {
				hm.On("GetTenantStats", "tenant1").Return(map[string]interface{}{
					"total_peers":   10,
					"online_peers":  5,
					"offline_peers": 5,
				})
				rm.On("GetTenantInfo", "tenant1").Return(&p2p.Tenant{
					TenantID:       "tenant1",
					NetworkCIDR:    "10.0.0.0/24",
					MaxPeers:       100,
					MaxConnections: 50,
					CreatedAt:      time.Unix(0, 0),
					LastActivity:   time.Unix(0, 0),
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"total_peers":10,"online_peers":5,"offline_peers":5,"tenant_info":{"tenant_id":"tenant1","network_cidr":"10.0.0.0/24","max_peers":100,"max_connections":50,"created_at":0,"last_activity":0}}`,
		},
		{
			name:     "tenant not found",
			tenantID: "nonexistent",
			mockSetup: func(hm *MockHeartbeatManager, rm *MockRelayManager) {
				hm.On("GetTenantStats", "nonexistent").Return(map[string]interface{}{
					"total_peers":   0,
					"online_peers":  0,
					"offline_peers": 0,
				})
				rm.On("GetTenantInfo", "nonexistent").Return((*p2p.Tenant)(nil), assert.AnError)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"total_peers":0,"online_peers":0,"offline_peers":0}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "tenant_id", Value: tt.tenantID}}
			c.Request = httptest.NewRequest("GET", "/heartbeat/tenant/"+tt.tenantID+"/stats", nil)

			heartbeatManager := &MockHeartbeatManager{}
			relayManager := &MockRelayManager{}
			tt.mockSetup(heartbeatManager, relayManager)

			handler := &TestHeartbeatHandler{
				heartbeatManager: heartbeatManager,
				relayManager:     relayManager,
				logger:           zap.NewNop(),
			}

			handler.GetTenantStats(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
			heartbeatManager.AssertExpectations(t)
			relayManager.AssertExpectations(t)
		})
	}
}

func TestHeartbeatHandler_GetSystemStats(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*MockHeartbeatManager, *MockRelayManager)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful system stats retrieval",
			mockSetup: func(hm *MockHeartbeatManager, rm *MockRelayManager) {
				hm.On("GetTenantStats", "system").Return(map[string]interface{}{
					"total_peers":   200,
					"online_peers":  100,
					"offline_peers": 100,
				})
				rm.On("GetRelayStats", "system").Return(&RelayStats{
					TotalPeers:        200,
					OnlinePeers:       100,
					DirectConnections: 50,
					RelayConnections:  50,
					TotalTrafficBytes: 1000000,
					AverageLatencyMS:  20,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"total_peers":200,"online_peers":100,"offline_peers":100,"relay_stats":{"total_peers":200,"online_peers":100,"direct_connections":50,"relay_connections":50,"total_traffic_bytes":1000000,"average_latency_ms":20}}`,
		},
		{
			name: "system stats error",
			mockSetup: func(hm *MockHeartbeatManager, rm *MockRelayManager) {
				hm.On("GetTenantStats", "system").Return(map[string]interface{}{
					"total_peers":   0,
					"online_peers":  0,
					"offline_peers": 0,
				})
				rm.On("GetRelayStats", "system").Return((*RelayStats)(nil), assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Internal server error","code":500,"message":"assert.AnError general error for testing"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/heartbeat/system/stats", nil)

			heartbeatManager := &MockHeartbeatManager{}
			relayManager := &MockRelayManager{}
			tt.mockSetup(heartbeatManager, relayManager)

			handler := &TestHeartbeatHandler{
				heartbeatManager: heartbeatManager,
				relayManager:     relayManager,
				logger:           zap.NewNop(),
			}

			handler.GetSystemStats(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
			heartbeatManager.AssertExpectations(t)
			relayManager.AssertExpectations(t)
		})
	}
}

func TestHeartbeatHandler_InvalidInput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest("POST", "/api/v1/tenants/tenant1/peers/peer1/heartbeat", nil)
	c.Params = gin.Params{
		{Key: "tenant_id", Value: "tenant1"},
		{Key: "peer_id", Value: "peer1"},
	}

	mockHeartbeatManager := &MockHeartbeatManager{}
	mockHeartbeatManager.On("ProcessHeartbeat", "peer1", "tenant1").Return((*p2p.HeartbeatResponse)(nil), assert.AnError)

	handler := &TestHeartbeatHandler{
		heartbeatManager: mockHeartbeatManager,
		relayManager:     &MockRelayManager{},
		logger:           zap.NewNop(),
	}

	handler.HandleHeartbeat(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockHeartbeatManager.AssertExpectations(t)
}

func TestHeartbeatHandler_MissingFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	requestBody := `{"peer_id": "peer1"}`
	c.Request = httptest.NewRequest("POST", "/api/v1/tenants/tenant1/peers/peer1/heartbeat", strings.NewReader(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{
		{Key: "tenant_id", Value: "tenant1"},
		{Key: "peer_id", Value: "peer1"},
	}

	mockHeartbeatManager := &MockHeartbeatManager{}
	mockHeartbeatManager.On("ProcessHeartbeat", "peer1", "tenant1").Return(&p2p.HeartbeatResponse{
		Status:    "ok",
		Timestamp: 0,
		Online:    true,
		Message:   "Heartbeat processed successfully",
	}, nil)

	handler := &TestHeartbeatHandler{
		heartbeatManager: mockHeartbeatManager,
		relayManager:     &MockRelayManager{},
		logger:           zap.NewNop(),
	}

	handler.HandleHeartbeat(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockHeartbeatManager.AssertExpectations(t)
}
