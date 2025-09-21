package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/twogc/cloudbridge-relay-installer/internal/p2p"
	"go.uber.org/zap"
)

// MockP2PConnectionManager мок для P2PConnectionManager
type MockP2PConnectionManager struct {
	mock.Mock
}

// Убеждаемся, что MockP2PConnectionManager реализует интерфейс
var _ p2p.P2PConnectionManagerInterface = (*MockP2PConnectionManager)(nil)

func (m *MockP2PConnectionManager) EstablishConnection(ctx context.Context, sessionID, peer1ID, peer2ID, protocol string) (*p2p.P2PConnection, error) {
	args := m.Called(ctx, sessionID, peer1ID, peer2ID, protocol)
	return args.Get(0).(*p2p.P2PConnection), args.Error(1)
}

func (m *MockP2PConnectionManager) GetConnection(sessionID string) (*p2p.P2PConnection, bool) {
	args := m.Called(sessionID)
	return args.Get(0).(*p2p.P2PConnection), args.Bool(1)
}

func (m *MockP2PConnectionManager) CloseConnection(sessionID string) error {
	args := m.Called(sessionID)
	return args.Error(0)
}

func (m *MockP2PConnectionManager) GetActiveConnections() []*p2p.P2PConnection {
	args := m.Called()
	return args.Get(0).([]*p2p.P2PConnection)
}

func (m *MockP2PConnectionManager) RegisterPeer(peer *p2p.Peer) error {
	args := m.Called(peer)
	return args.Error(0)
}

func (m *MockP2PConnectionManager) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestNewP2PManager(t *testing.T) {
	logger := zap.NewNop()

	manager := NewP2PManager(logger)

	assert.NotNil(t, manager)
	assert.Equal(t, logger, manager.logger)
	assert.NotNil(t, manager.connectionManager)
}

func TestP2PManager_EstablishConnection(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockP2PConnectionManager)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful connection establishment",
			requestBody: map[string]interface{}{
				"session_id":     "session1",
				"target_peer_id": "peer2",
				"protocol":       "quic",
			},
			mockSetup: func(m *MockP2PConnectionManager) {
				m.On("EstablishConnection", mock.Anything, "session1", "peer1", "peer2", "quic").Return(&p2p.P2PConnection{
					SessionID:      "session1",
					Peer1ID:        "peer1",
					Peer2ID:        "peer2",
					ConnectionType: "direct",
					Protocol:       "quic",
					State:          p2p.ConnectionStateConnected,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"success":true,"session_id":"session1","connection_type":"direct"}`,
		},
		{
			name: "connection establishment error",
			requestBody: map[string]interface{}{
				"session_id":     "session1",
				"target_peer_id": "peer2",
				"protocol":       "quic",
			},
			mockSetup: func(m *MockP2PConnectionManager) {
				m.On("EstablishConnection", mock.Anything, "session1", "peer1", "peer2", "quic").Return((*p2p.P2PConnection)(nil), assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Connection failed","code":500,"message":"assert.AnError general error for testing"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			requestBody, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest("POST", "/p2p/connect", bytes.NewBuffer(requestBody))

			manager := &MockP2PConnectionManager{}
			tt.mockSetup(manager)

			handler := &P2PManager{
				connectionManager: manager,
				logger:            zap.NewNop(),
			}

			handler.EstablishConnection(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
			manager.AssertExpectations(t)
		})
	}
}

func TestP2PManager_GetSessionStatus(t *testing.T) {
	tests := []struct {
		name           string
		sessionID      string
		mockSetup      func(*MockP2PConnectionManager)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:      "successful status retrieval",
			sessionID: "session1",
			mockSetup: func(m *MockP2PConnectionManager) {
				m.On("GetConnection", "session1").Return(&p2p.P2PConnection{
					SessionID:      "session1",
					Peer1ID:        "peer1",
					Peer2ID:        "peer2",
					ConnectionType: "direct",
					Protocol:       "quic",
					State:          p2p.ConnectionStateConnected,
				}, true)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"success":true,"status":{"session_id":"session1","peer1_id":"peer1","peer2_id":"peer2","connection_type":"direct","protocol":"quic","state":"connected","established_at":"0001-01-01T00:00:00Z","last_activity":"0001-01-01T00:00:00Z","bytes_transferred":0,"latency_ms":0}}`,
		},
		{
			name:      "connection not found",
			sessionID: "nonexistent",
			mockSetup: func(m *MockP2PConnectionManager) {
				m.On("GetConnection", "nonexistent").Return((*p2p.P2PConnection)(nil), false)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"Session not found","code":404,"message":"Session with specified ID not found"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "session_id", Value: tt.sessionID}}
			c.Request = httptest.NewRequest("GET", "/p2p/status/"+tt.sessionID, nil)

			manager := &MockP2PConnectionManager{}
			tt.mockSetup(manager)

			handler := &P2PManager{
				connectionManager: manager,
				logger:            zap.NewNop(),
			}

			handler.GetSessionStatus(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
			manager.AssertExpectations(t)
		})
	}
}

func TestP2PManager_CloseSession(t *testing.T) {
	tests := []struct {
		name           string
		sessionID      string
		mockSetup      func(*MockP2PConnectionManager)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:      "successful connection closure",
			sessionID: "session1",
			mockSetup: func(m *MockP2PConnectionManager) {
				m.On("GetConnection", "session1").Return(&p2p.P2PConnection{
					SessionID: "session1",
					Peer1ID:   "peer1",
					Peer2ID:   "peer2",
				}, true)
				m.On("CloseConnection", "session1").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"success":true,"message":"Session closed successfully"}`,
		},
		{
			name:      "connection closure error",
			sessionID: "session1",
			mockSetup: func(m *MockP2PConnectionManager) {
				m.On("GetConnection", "session1").Return(&p2p.P2PConnection{
					SessionID: "session1",
					Peer1ID:   "peer1",
					Peer2ID:   "peer2",
				}, true)
				m.On("CloseConnection", "session1").Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to close session","code":500,"message":"assert.AnError general error for testing"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "session_id", Value: tt.sessionID}}
			c.Request = httptest.NewRequest("DELETE", "/p2p/close/"+tt.sessionID, nil)

			manager := &MockP2PConnectionManager{}
			tt.mockSetup(manager)

			handler := &P2PManager{
				connectionManager: manager,
				logger:            zap.NewNop(),
			}

			handler.CloseSession(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
			manager.AssertExpectations(t)
		})
	}
}

func TestP2PManager_GetActiveConnections(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*MockP2PConnectionManager)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful connections listing",
			mockSetup: func(m *MockP2PConnectionManager) {
				m.On("GetActiveConnections").Return([]*p2p.P2PConnection{
					{SessionID: "session1", Peer1ID: "peer1", Peer2ID: "peer2", ConnectionType: "direct", Protocol: "quic", State: p2p.ConnectionStateConnected},
					{SessionID: "session2", Peer1ID: "peer3", Peer2ID: "peer4", ConnectionType: "relay", Protocol: "websocket", State: p2p.ConnectionStateConnected},
				})
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"success":true,"connections":[{"session_id":"session1","peer1_id":"peer1","peer2_id":"peer2","connection_type":"direct","protocol":"quic","state":"connected","established_at":"0001-01-01T00:00:00Z","last_activity":"0001-01-01T00:00:00Z","bytes_transferred":0,"latency_ms":0},{"session_id":"session2","peer1_id":"peer3","peer2_id":"peer4","connection_type":"relay","protocol":"websocket","state":"connected","established_at":"0001-01-01T00:00:00Z","last_activity":"0001-01-01T00:00:00Z","bytes_transferred":0,"latency_ms":0}],"total":2}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/p2p/connections", nil)

			manager := &MockP2PConnectionManager{}
			tt.mockSetup(manager)

			handler := &P2PManager{
				connectionManager: manager,
				logger:            zap.NewNop(),
			}

			handler.GetActiveConnections(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
			manager.AssertExpectations(t)
		})
	}
}

func TestP2PManager_InvalidInput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	requestBody := `{"session_id": "session1", "target_peer_id": "peer2", "protocol": "quic", "invalid": "json"`
	c.Request = httptest.NewRequest("POST", "/p2p/connect", bytes.NewBufferString(requestBody))

	manager := &MockP2PConnectionManager{}

	handler := &P2PManager{
		connectionManager: manager,
		logger:            zap.NewNop(),
	}

	handler.EstablishConnection(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"Invalid request format","code":400,"message":"invalid character '}' looking for beginning of object key string"}`, w.Body.String())
}

func TestP2PManager_MissingFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	requestBody := `{"session_id":"","target_peer_id":"","protocol":""}`
	c.Request = httptest.NewRequest("POST", "/p2p/connect", bytes.NewBufferString(requestBody))

	manager := &MockP2PConnectionManager{}

	handler := &P2PManager{
		connectionManager: manager,
		logger:            zap.NewNop(),
	}

	handler.EstablishConnection(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"Invalid request format","code":400,"message":"Key: 'P2PConnectionRequest.SessionID' Error:Field validation for 'SessionID' failed on the 'required' tag\nKey: 'P2PConnectionRequest.TargetPeerID' Error:Field validation for 'TargetPeerID' failed on the 'required' tag\nKey: 'P2PConnectionRequest.Protocol' Error:Field validation for 'Protocol' failed on the 'required' tag"}`, w.Body.String())
}
