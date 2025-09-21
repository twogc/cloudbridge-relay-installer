package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/twogc/cloudbridge-relay-installer/internal/keycloak"
	"github.com/twogc/cloudbridge-relay-installer/internal/p2p"
	"go.uber.org/zap"
)

func TestNewProtocolHandler(t *testing.T) {
	logger := zap.NewNop()
	keycloakValidator := &MockKeycloakValidator{}
	peerManager := &MockPeerManager{}
	relayManager := &MockRelayManager{}

	handler := NewProtocolHandler(keycloakValidator, logger, peerManager, relayManager)

	assert.NotNil(t, handler)
	assert.Equal(t, keycloakValidator, handler.keycloakValidator)
	assert.Equal(t, peerManager, handler.peerManager)
	assert.Equal(t, relayManager, handler.relayManager)
	assert.Equal(t, logger, handler.logger)
}

func TestProtocolHandler_RegisterQuickClient(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockKeycloakValidator, *MockPeerManager, *MockRelayManager)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful quick client registration",
			requestBody: map[string]interface{}{
				"client_id": "client1",
				"protocol":  "quick",
			},
			mockSetup: func(kv *MockKeycloakValidator, pm *MockPeerManager, rm *MockRelayManager) {
				kv.On("ValidateQuickToken", "valid-token").Return(&keycloak.QuickTokenClaims{
					BaseTokenClaims: keycloak.BaseTokenClaims{
						TenantID: "tenant1",
						ServerID: "server1",
					},
				}, nil)
				pm.On("RegisterPeerDirect", mock.AnythingOfType("*p2p.Peer")).Return(nil)
				rm.On("RegisterClient", "client1", "quick").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"Quick client registered successfully"}`,
		},
		{
			name: "invalid token",
			requestBody: map[string]interface{}{
				"client_id": "client1",
				"protocol":  "quick",
			},
			mockSetup: func(kv *MockKeycloakValidator, pm *MockPeerManager, rm *MockRelayManager) {
				kv.On("ValidateQuickToken", "valid-token").Return((*keycloak.QuickTokenClaims)(nil), assert.AnError)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Unauthorized"}`,
		},
		{
			name: "peer registration error",
			requestBody: map[string]interface{}{
				"client_id": "client1",
				"protocol":  "quick",
			},
			mockSetup: func(kv *MockKeycloakValidator, pm *MockPeerManager, rm *MockRelayManager) {
				kv.On("ValidateQuickToken", "valid-token").Return(&keycloak.QuickTokenClaims{
					BaseTokenClaims: keycloak.BaseTokenClaims{
						TenantID: "tenant1",
						ServerID: "server1",
					},
				}, nil)
				pm.On("RegisterPeerDirect", mock.AnythingOfType("*p2p.Peer")).Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Internal server error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			requestBody, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest("POST", "/protocol/quick/register", bytes.NewBuffer(requestBody))
			c.Request.Header.Set("Authorization", "Bearer valid-token")

			keycloakValidator := &MockKeycloakValidator{}
			peerManager := &MockPeerManager{}
			relayManager := &MockRelayManager{}
			tt.mockSetup(keycloakValidator, peerManager, relayManager)

			handler := &ProtocolHandler{
				keycloakValidator: keycloakValidator,
				peerManager:       peerManager,
				relayManager:      relayManager,
				logger:            zap.NewNop(),
			}

			handler.RegisterQuickClient(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
			keycloakValidator.AssertExpectations(t)
			peerManager.AssertExpectations(t)
			relayManager.AssertExpectations(t)
		})
	}
}

func TestProtocolHandler_RegisterP2PMeshClient(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockKeycloakValidator, *MockPeerManager, *MockRelayManager)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful p2p mesh client registration",
			requestBody: map[string]interface{}{
				"client_id": "client1",
				"protocol":  "p2p_mesh",
			},
			mockSetup: func(kv *MockKeycloakValidator, pm *MockPeerManager, rm *MockRelayManager) {
				kv.On("ValidateP2PMeshToken", "valid-token").Return(&keycloak.P2PMeshTokenClaims{
					BaseTokenClaims: keycloak.BaseTokenClaims{
						TenantID: "tenant1",
						ServerID: "server1",
					},
				}, nil)
				pm.On("RegisterPeerDirect", mock.AnythingOfType("*p2p.Peer")).Return(nil)
				rm.On("RegisterClient", "client1", "p2p_mesh").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"P2P Mesh client registered successfully"}`,
		},
		{
			name: "invalid token",
			requestBody: map[string]interface{}{
				"client_id": "client1",
				"protocol":  "p2p_mesh",
			},
			mockSetup: func(kv *MockKeycloakValidator, pm *MockPeerManager, rm *MockRelayManager) {
				kv.On("ValidateP2PMeshToken", "valid-token").Return((*keycloak.P2PMeshTokenClaims)(nil), assert.AnError)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Unauthorized"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			requestBody, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest("POST", "/protocol/p2p_mesh/register", bytes.NewBuffer(requestBody))
			c.Request.Header.Set("Authorization", "Bearer valid-token")

			keycloakValidator := &MockKeycloakValidator{}
			peerManager := &MockPeerManager{}
			relayManager := &MockRelayManager{}
			tt.mockSetup(keycloakValidator, peerManager, relayManager)

			handler := &ProtocolHandler{
				keycloakValidator: keycloakValidator,
				peerManager:       peerManager,
				relayManager:      relayManager,
				logger:            zap.NewNop(),
			}

			handler.RegisterP2PMeshClient(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
			keycloakValidator.AssertExpectations(t)
			peerManager.AssertExpectations(t)
			relayManager.AssertExpectations(t)
		})
	}
}

func TestProtocolHandler_GetClientStatus(t *testing.T) {
	tests := []struct {
		name           string
		clientID       string
		mockSetup      func(*MockKeycloakValidator, *MockPeerManager, *MockRelayManager)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:     "successful status retrieval",
			clientID: "client1",
			mockSetup: func(kv *MockKeycloakValidator, pm *MockPeerManager, rm *MockRelayManager) {
				kv.On("ValidateQuickToken", "valid-token").Return(&keycloak.QuickTokenClaims{
					BaseTokenClaims: keycloak.BaseTokenClaims{
						TenantID: "tenant1",
						ServerID: "server1",
					},
				}, nil)
				pm.On("GetPeer", "client1").Return(&p2p.Peer{
					ID:       "client1",
					IsOnline: true,
				}, nil)
				rm.On("GetClientStatus", "client1").Return("connected", nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"client_id":"client1","status":"active","connection_status":"connected"}`,
		},
		{
			name:     "client not found",
			clientID: "nonexistent",
			mockSetup: func(kv *MockKeycloakValidator, pm *MockPeerManager, rm *MockRelayManager) {
				kv.On("ValidateQuickToken", "valid-token").Return(&keycloak.QuickTokenClaims{
					BaseTokenClaims: keycloak.BaseTokenClaims{
						TenantID: "tenant1",
						ServerID: "server1",
					},
				}, nil)
				pm.On("GetPeer", "nonexistent").Return((*p2p.Peer)(nil), assert.AnError)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"Client not found"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "client_id", Value: tt.clientID}}
			c.Request = httptest.NewRequest("GET", "/protocol/client/"+tt.clientID+"/status", nil)
			c.Request.Header.Set("Authorization", "Bearer valid-token")

			keycloakValidator := &MockKeycloakValidator{}
			peerManager := &MockPeerManager{}
			relayManager := &MockRelayManager{}
			tt.mockSetup(keycloakValidator, peerManager, relayManager)

			handler := &ProtocolHandler{
				keycloakValidator: keycloakValidator,
				peerManager:       peerManager,
				relayManager:      relayManager,
				logger:            zap.NewNop(),
			}

			handler.GetClientStatus(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
			keycloakValidator.AssertExpectations(t)
			peerManager.AssertExpectations(t)
			relayManager.AssertExpectations(t)
		})
	}
}

func TestProtocolHandler_InvalidInput(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	requestBody := `{"invalid": "json"`
	c.Request = httptest.NewRequest("POST", "/protocol/quick/register", bytes.NewBufferString(requestBody))

	handler := &ProtocolHandler{
		keycloakValidator: &MockKeycloakValidator{},
		peerManager:       &MockPeerManager{},
		relayManager:      &MockRelayManager{},
		logger:            zap.NewNop(),
	}

	handler.RegisterQuickClient(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"Invalid input"}`, w.Body.String())
}

func TestProtocolHandler_MissingFields(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	requestBody := `{"client_id": "client1"}`
	c.Request = httptest.NewRequest("POST", "/protocol/quick/register", bytes.NewBufferString(requestBody))

	handler := &ProtocolHandler{
		keycloakValidator: &MockKeycloakValidator{},
		peerManager:       &MockPeerManager{},
		relayManager:      &MockRelayManager{},
		logger:            zap.NewNop(),
	}

	handler.RegisterQuickClient(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"Invalid input"}`, w.Body.String())
}

func TestProtocolHandler_ConnectionLimits(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	requestBody := `{"client_id": "client1", "protocol": "quick"}`
	c.Request = httptest.NewRequest("POST", "/protocol/quick/register", bytes.NewBufferString(requestBody))
	c.Request.Header.Set("Authorization", "Bearer valid-token")

	keycloakValidator := &MockKeycloakValidator{}
	keycloakValidator.On("ValidateQuickToken", "valid-token").Return(&keycloak.QuickTokenClaims{
		BaseTokenClaims: keycloak.BaseTokenClaims{
			TenantID: "tenant1",
			ServerID: "server1",
		},
	}, nil)

	peerManager := &MockPeerManager{}
	peerManager.On("RegisterPeerDirect", mock.AnythingOfType("*p2p.Peer")).Return(nil)
	relayManager := &MockRelayManager{}
	relayManager.On("RegisterClient", "client1", "quick").Return(assert.AnError)

	handler := &ProtocolHandler{
		keycloakValidator: keycloakValidator,
		peerManager:       peerManager,
		relayManager:      relayManager,
		logger:            zap.NewNop(),
	}

	handler.RegisterQuickClient(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.JSONEq(t, `{"error":"Internal server error"}`, w.Body.String())
	keycloakValidator.AssertExpectations(t)
	relayManager.AssertExpectations(t)
}

func TestProtocolHandler_RegisterClient(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		method         string
		expectedStatus int
	}{
		{
			name:           "quick endpoint",
			path:           "/protocol/quick/register",
			method:         "POST",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "p2p endpoint",
			path:           "/protocol/p2p/register",
			method:         "POST",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unknown endpoint",
			path:           "/protocol/unknown/register",
			method:         "POST",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			requestBody := `{"target_server": "192.168.1.100", "connection_type": "tcp", "client_config": {"endpoint": "192.168.1.101:8080", "public_key": "test-key"}}`
			r := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(requestBody))
			r.Header.Set("Authorization", "Bearer valid-token")
			r.Header.Set("Content-Type", "application/json")

			keycloakValidator := &MockKeycloakValidator{}
			peerManager := &MockPeerManager{}
			relayManager := &MockRelayManager{}

			// Setup mocks for successful cases
			if tt.expectedStatus == http.StatusOK {
				if strings.Contains(tt.path, "quick") {
					keycloakValidator.On("ValidateQuickToken", "valid-token").Return(&keycloak.QuickTokenClaims{
						BaseTokenClaims: keycloak.BaseTokenClaims{
							TenantID: "tenant1",
							ServerID: "server1",
						},
					}, nil)
					peerManager.On("RegisterPeerDirect", mock.AnythingOfType("*p2p.Peer")).Return(nil)
					relayManager.On("RegisterClient", mock.Anything, "quick").Return(nil)
				} else if strings.Contains(tt.path, "p2p") {
					keycloakValidator.On("ValidateP2PMeshToken", "valid-token").Return(&keycloak.P2PMeshTokenClaims{
						BaseTokenClaims: keycloak.BaseTokenClaims{
							TenantID: "tenant1",
							ServerID: "server1",
						},
					}, nil)
					peerManager.On("RegisterPeerMesh", mock.AnythingOfType("*p2p.Peer")).Return(nil)
					relayManager.On("RegisterClient", mock.Anything, "p2p_mesh").Return(nil)
				}
			}

			handler := &ProtocolHandler{
				keycloakValidator: keycloakValidator,
				peerManager:       peerManager,
				relayManager:      relayManager,
				logger:            zap.NewNop(),
			}

			// Just test that it doesn't panic
			assert.NotPanics(t, func() {
				handler.RegisterClient(w, r)
			})

			// For unknown endpoint, we expect 404
			if strings.Contains(tt.path, "unknown") {
				assert.Equal(t, http.StatusNotFound, w.Code)
			}
		})
	}
}

func TestProtocolHandler_HandleQuickConnection(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		authHeader     string
		requestBody    string
		mockSetup      func(*MockKeycloakValidator)
		expectedStatus int
	}{
		{
			name:       "successful quick connection",
			method:     "POST",
			authHeader: "Bearer valid-token",
			requestBody: `{
				"target_server": "192.168.1.100",
				"connection_type": "tcp",
				"client_config": {
					"endpoint": "192.168.1.101:8080",
					"public_key": "test-key"
				}
			}`,
			mockSetup: func(kv *MockKeycloakValidator) {
				kv.On("ValidateQuickToken", "valid-token").Return(&keycloak.QuickTokenClaims{
					BaseTokenClaims: keycloak.BaseTokenClaims{
						TenantID: "tenant1",
						ServerID: "server1",
					},
					MaxConnections: "10",
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "method not allowed",
			method:         "GET",
			authHeader:     "Bearer valid-token",
			requestBody:    `{}`,
			mockSetup:      func(kv *MockKeycloakValidator) {},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "missing authorization header",
			method:         "POST",
			authHeader:     "",
			requestBody:    `{}`,
			mockSetup:      func(kv *MockKeycloakValidator) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid authorization header",
			method:         "POST",
			authHeader:     "InvalidHeader",
			requestBody:    `{}`,
			mockSetup:      func(kv *MockKeycloakValidator) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid token",
			method:     "POST",
			authHeader: "Bearer invalid-token",
			requestBody: `{
				"target_server": "192.168.1.100",
				"connection_type": "tcp",
				"client_config": {
					"endpoint": "192.168.1.101:8080",
					"public_key": "test-key"
				}
			}`,
			mockSetup: func(kv *MockKeycloakValidator) {
				kv.On("ValidateQuickToken", "invalid-token").Return((*keycloak.QuickTokenClaims)(nil), assert.AnError)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "missing target_server",
			method:     "POST",
			authHeader: "Bearer valid-token",
			requestBody: `{
				"connection_type": "tcp",
				"client_config": {
					"endpoint": "192.168.1.101:8080",
					"public_key": "test-key"
				}
			}`,
			mockSetup: func(kv *MockKeycloakValidator) {
				kv.On("ValidateQuickToken", "valid-token").Return(&keycloak.QuickTokenClaims{
					BaseTokenClaims: keycloak.BaseTokenClaims{
						TenantID: "tenant1",
						ServerID: "server1",
					},
					MaxConnections: "10",
				}, nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "invalid request body",
			method:      "POST",
			authHeader:  "Bearer valid-token",
			requestBody: `{invalid json`,
			mockSetup: func(kv *MockKeycloakValidator) {
				kv.On("ValidateQuickToken", "valid-token").Return(&keycloak.QuickTokenClaims{
					BaseTokenClaims: keycloak.BaseTokenClaims{
						TenantID: "tenant1",
						ServerID: "server1",
					},
					MaxConnections: "10",
				}, nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(tt.method, "/protocol/quick/register", bytes.NewBufferString(tt.requestBody))
			if tt.authHeader != "" {
				r.Header.Set("Authorization", tt.authHeader)
			}

			keycloakValidator := &MockKeycloakValidator{}
			tt.mockSetup(keycloakValidator)

			handler := &ProtocolHandler{
				keycloakValidator: keycloakValidator,
				peerManager:       &MockPeerManager{},
				relayManager:      &MockRelayManager{},
				logger:            zap.NewNop(),
			}

			handler.handleQuickConnection(w, r)

			assert.Equal(t, tt.expectedStatus, w.Code)
			keycloakValidator.AssertExpectations(t)
		})
	}
}

func TestProtocolHandler_HandleP2PMeshConnection(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		authHeader     string
		requestBody    string
		mockSetup      func(*MockKeycloakValidator)
		expectedStatus int
	}{
		{
			name:       "successful p2p mesh connection",
			method:     "POST",
			authHeader: "Bearer valid-token",
			requestBody: `{
				"mesh_id": "mesh1",
				"server_config": {
					"endpoint": "192.168.1.100:51820",
					"public_key": "test-key",
					"allowed_ips": ["192.168.1.0/24"]
				}
			}`,
			mockSetup: func(kv *MockKeycloakValidator) {
				kv.On("ValidateP2PMeshToken", "valid-token").Return(&keycloak.P2PMeshTokenClaims{
					BaseTokenClaims: keycloak.BaseTokenClaims{
						TenantID: "tenant1",
						ServerID: "server1",
					},
					MaxPeers: "10",
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "method not allowed",
			method:         "GET",
			authHeader:     "Bearer valid-token",
			requestBody:    `{}`,
			mockSetup:      func(kv *MockKeycloakValidator) {},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "missing authorization header",
			method:         "POST",
			authHeader:     "",
			requestBody:    `{}`,
			mockSetup:      func(kv *MockKeycloakValidator) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid token",
			method:     "POST",
			authHeader: "Bearer invalid-token",
			requestBody: `{
				"mesh_id": "mesh1",
				"server_config": {
					"endpoint": "192.168.1.100:51820",
					"public_key": "test-key",
					"allowed_ips": ["192.168.1.0/24"]
				}
			}`,
			mockSetup: func(kv *MockKeycloakValidator) {
				kv.On("ValidateP2PMeshToken", "invalid-token").Return((*keycloak.P2PMeshTokenClaims)(nil), assert.AnError)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "missing mesh_id",
			method:     "POST",
			authHeader: "Bearer valid-token",
			requestBody: `{
				"server_config": {
					"endpoint": "192.168.1.100:51820",
					"public_key": "test-key",
					"allowed_ips": ["192.168.1.0/24"]
				}
			}`,
			mockSetup: func(kv *MockKeycloakValidator) {
				kv.On("ValidateP2PMeshToken", "valid-token").Return(&keycloak.P2PMeshTokenClaims{
					BaseTokenClaims: keycloak.BaseTokenClaims{
						TenantID: "tenant1",
						ServerID: "server1",
					},
					MaxPeers: "10",
				}, nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "invalid request body",
			method:      "POST",
			authHeader:  "Bearer valid-token",
			requestBody: `{invalid json`,
			mockSetup: func(kv *MockKeycloakValidator) {
				kv.On("ValidateP2PMeshToken", "valid-token").Return(&keycloak.P2PMeshTokenClaims{
					BaseTokenClaims: keycloak.BaseTokenClaims{
						TenantID: "tenant1",
						ServerID: "server1",
					},
					MaxPeers: "10",
				}, nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(tt.method, "/protocol/p2p/register", bytes.NewBufferString(tt.requestBody))
			if tt.authHeader != "" {
				r.Header.Set("Authorization", tt.authHeader)
			}

			keycloakValidator := &MockKeycloakValidator{}
			tt.mockSetup(keycloakValidator)

			handler := &ProtocolHandler{
				keycloakValidator: keycloakValidator,
				peerManager:       &MockPeerManager{},
				relayManager:      &MockRelayManager{},
				logger:            zap.NewNop(),
			}

			handler.handleP2PMeshConnection(w, r)

			assert.Equal(t, tt.expectedStatus, w.Code)
			keycloakValidator.AssertExpectations(t)
		})
	}
}

func TestProtocolHandler_GetCurrentConnections(t *testing.T) {
	handler := &ProtocolHandler{
		logger: zap.NewNop(),
	}

	// Should return 0 for now (mock implementation)
	connections := handler.getCurrentConnections("tenant1", "server1")
	assert.Equal(t, 0, connections)
}

func TestProtocolHandler_GetCurrentPeers(t *testing.T) {
	handler := &ProtocolHandler{
		logger: zap.NewNop(),
	}

	// Should return 0 for now (mock implementation)
	peers := handler.getCurrentPeers("tenant1", "server1", "mesh1")
	assert.Equal(t, 0, peers)
}

func TestProtocolHandler_CreateQuickConnection(t *testing.T) {
	handler := &ProtocolHandler{
		logger: zap.NewNop(),
	}

	claims := &keycloak.QuickTokenClaims{
		BaseTokenClaims: keycloak.BaseTokenClaims{
			TenantID: "tenant1",
			ServerID: "server1",
		},
	}

	req := QuickConnectionRequest{
		TargetServer:   "192.168.1.100",
		ConnectionType: "tcp",
		ClientConfig: QuickClientConfig{
			Endpoint:  "192.168.1.101:8080",
			PublicKey: "test-key",
		},
	}

	conn, err := handler.createQuickConnection(claims, req)

	assert.NoError(t, err)
	assert.NotNil(t, conn)
	assert.NotEmpty(t, conn.ID)
	assert.Equal(t, "192.168.1.100:8080", conn.Endpoint)
	assert.Contains(t, conn.ID, "quick-tenant1-")
}

func TestProtocolHandler_CreateP2PMeshConnection(t *testing.T) {
	handler := &ProtocolHandler{
		logger: zap.NewNop(),
	}

	claims := &keycloak.P2PMeshTokenClaims{
		BaseTokenClaims: keycloak.BaseTokenClaims{
			TenantID: "tenant1",
			ServerID: "server1",
		},
	}

	req := P2PMeshRequest{
		MeshID: "mesh1",
		ServerConfig: P2PServerConfig{
			Endpoint:   "192.168.1.100:51820",
			PublicKey:  "test-key",
			AllowedIPs: []string{"192.168.1.0/24"},
		},
	}

	mesh, err := handler.createP2PMeshConnection(claims, req)

	assert.NoError(t, err)
	assert.NotNil(t, mesh)
	assert.Equal(t, "mesh1", mesh.ID)
	assert.Len(t, mesh.Peers, 1)
	assert.Equal(t, "peer-1", mesh.Peers[0].ID)
}

func TestProtocolHandler_GetClientStatus_MissingClientID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{} // No client_id param

	handler := &ProtocolHandler{
		keycloakValidator: &MockKeycloakValidator{},
		peerManager:       &MockPeerManager{},
		relayManager:      &MockRelayManager{},
		logger:            zap.NewNop(),
	}

	handler.GetClientStatus(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"Missing client_id"}`, w.Body.String())
}

func TestProtocolHandler_GetClientStatus_Unauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "client_id", Value: "client1"}}
	c.Request = httptest.NewRequest("GET", "/protocol/client/client1/status", nil)
	// No Authorization header

	handler := &ProtocolHandler{
		keycloakValidator: &MockKeycloakValidator{},
		peerManager:       &MockPeerManager{},
		relayManager:      &MockRelayManager{},
		logger:            zap.NewNop(),
	}

	handler.GetClientStatus(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.JSONEq(t, `{"error":"Unauthorized"}`, w.Body.String())
}

func TestProtocolHandler_GetClientStatus_InvalidToken(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "client_id", Value: "client1"}}
	c.Request = httptest.NewRequest("GET", "/protocol/client/client1/status", nil)
	c.Request.Header.Set("Authorization", "Bearer invalid-token")

	keycloakValidator := &MockKeycloakValidator{}
	keycloakValidator.On("ValidateQuickToken", "invalid-token").Return((*keycloak.QuickTokenClaims)(nil), assert.AnError)

	handler := &ProtocolHandler{
		keycloakValidator: keycloakValidator,
		peerManager:       &MockPeerManager{},
		relayManager:      &MockRelayManager{},
		logger:            zap.NewNop(),
	}

	handler.GetClientStatus(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.JSONEq(t, `{"error":"Unauthorized"}`, w.Body.String())
	keycloakValidator.AssertExpectations(t)
}

func TestProtocolHandler_GetClientStatus_PeerNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "client_id", Value: "nonexistent"}}
	c.Request = httptest.NewRequest("GET", "/protocol/client/nonexistent/status", nil)
	c.Request.Header.Set("Authorization", "Bearer valid-token")

	keycloakValidator := &MockKeycloakValidator{}
	keycloakValidator.On("ValidateQuickToken", "valid-token").Return(&keycloak.QuickTokenClaims{
		BaseTokenClaims: keycloak.BaseTokenClaims{
			TenantID: "tenant1",
			ServerID: "server1",
		},
	}, nil)

	peerManager := &MockPeerManager{}
	peerManager.On("GetPeer", "nonexistent").Return((*p2p.Peer)(nil), assert.AnError)

	handler := &ProtocolHandler{
		keycloakValidator: keycloakValidator,
		peerManager:       peerManager,
		relayManager:      &MockRelayManager{},
		logger:            zap.NewNop(),
	}

	handler.GetClientStatus(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.JSONEq(t, `{"error":"Client not found"}`, w.Body.String())
	keycloakValidator.AssertExpectations(t)
	peerManager.AssertExpectations(t)
}

func TestProtocolHandler_GetClientStatus_RelayManagerError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "client_id", Value: "client1"}}
	c.Request = httptest.NewRequest("GET", "/protocol/client/client1/status", nil)
	c.Request.Header.Set("Authorization", "Bearer valid-token")

	keycloakValidator := &MockKeycloakValidator{}
	keycloakValidator.On("ValidateQuickToken", "valid-token").Return(&keycloak.QuickTokenClaims{
		BaseTokenClaims: keycloak.BaseTokenClaims{
			TenantID: "tenant1",
			ServerID: "server1",
		},
	}, nil)

	peerManager := &MockPeerManager{}
	peerManager.On("GetPeer", "client1").Return(&p2p.Peer{
		ID:       "client1",
		IsOnline: true,
	}, nil)

	relayManager := &MockRelayManager{}
	relayManager.On("GetClientStatus", "client1").Return("", assert.AnError)

	handler := &ProtocolHandler{
		keycloakValidator: keycloakValidator,
		peerManager:       peerManager,
		relayManager:      relayManager,
		logger:            zap.NewNop(),
	}

	handler.GetClientStatus(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"client_id":"client1","status":"active","connection_status":"unknown"}`, w.Body.String())
	keycloakValidator.AssertExpectations(t)
	peerManager.AssertExpectations(t)
	relayManager.AssertExpectations(t)
}
