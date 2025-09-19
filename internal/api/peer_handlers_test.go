package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// TestTokenValidator is a test implementation of TokenValidator
type TestTokenValidator struct {
	claims *P2PClaims
	err    error
}

func (t *TestTokenValidator) ValidateP2PToken(token string) (*P2PClaims, error) {
	return t.claims, t.err
}

func TestPeerManager_RegisterPeer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	validator := &TestTokenValidator{}
	pm := NewPeerManager(logger, validator)

	tests := []struct {
		name           string
		request        PeerRegistrationRequest
		mockClaims     *P2PClaims
		mockError      error
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful registration",
			request: PeerRegistrationRequest{
				PeerID:         "test-peer-1",
				TenantID:       "tenant-1",
				DeviceID:       "device-1",
				ConnectionType: "p2p-mesh",
			},
			mockClaims: &P2PClaims{
				TenantID: "tenant-1",
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid token",
			request: PeerRegistrationRequest{
				PeerID:         "test-peer-2",
				TenantID:       "tenant-1",
				DeviceID:       "device-1",
				ConnectionType: "p2p-mesh",
			},
			mockClaims:     nil,
			mockError:      assert.AnError,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid token",
		},
		{
			name: "tenant mismatch",
			request: PeerRegistrationRequest{
				PeerID:         "test-peer-3",
				TenantID:       "tenant-1",
				DeviceID:       "device-1",
				ConnectionType: "p2p-mesh",
			},
			mockClaims: &P2PClaims{
				TenantID: "tenant-2",
			},
			mockError:      nil,
			expectedStatus: http.StatusForbidden,
			expectedError:  "Tenant mismatch",
		},
		{
			name: "missing required fields",
			request: PeerRegistrationRequest{
				PeerID: "",
			},
			mockClaims:     nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup validator
			validator.claims = tt.mockClaims
			validator.err = tt.mockError

			// Create request
			reqBody, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/peers/register", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer test-token")

			// Create response recorder
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Execute
			pm.RegisterPeer(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response.Error, tt.expectedError)
			} else if tt.expectedStatus == http.StatusOK {
				var response PeerRegistrationResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response.Success)
				assert.Equal(t, tt.request.PeerID, response.PeerID)
			}
		})
	}
}

func TestPeerManager_DiscoverPeers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	validator := &TestTokenValidator{}
	pm := NewPeerManager(logger, validator)

	// Add a test peer
	testPeer := &Peer{
		ID:             "test-id",
		PeerID:         "test-peer-1",
		TenantID:       "tenant-1",
		DeviceID:       "device-1",
		ConnectionType: "p2p-mesh",
		IsOnline:       true,
		LastSeen:       time.Now(),
		RegisteredAt:   time.Now(),
		MeshStatus:     "connected",
	}
	pm.peers["test-peer-1"] = testPeer

	tests := []struct {
		name           string
		tenantID       string
		mockClaims     *P2PClaims
		mockError      error
		expectedStatus int
		expectedError  string
	}{
		{
			name:     "successful discovery",
			tenantID: "tenant-1",
			mockClaims: &P2PClaims{
				TenantID: "tenant-1",
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing tenant_id",
			tenantID:       "",
			mockClaims:     nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Missing tenant_id",
		},
		{
			name:     "invalid token",
			tenantID: "tenant-1",
			mockClaims: &P2PClaims{
				TenantID: "tenant-1",
			},
			mockError:      assert.AnError,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid token",
		},
		{
			name:     "tenant mismatch",
			tenantID: "tenant-1",
			mockClaims: &P2PClaims{
				TenantID: "tenant-2",
			},
			mockError:      nil,
			expectedStatus: http.StatusForbidden,
			expectedError:  "Tenant mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup validator
			validator.claims = tt.mockClaims
			validator.err = tt.mockError

			// Create request
			url := "/api/v1/peers/discover"
			if tt.tenantID != "" {
				url += "?tenant_id=" + tt.tenantID
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set("Authorization", "Bearer test-token")

			// Create response recorder
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Execute
			pm.DiscoverPeers(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response.Error, tt.expectedError)
			} else if tt.expectedStatus == http.StatusOK {
				var response PeerDiscoveryResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response.Success)
				assert.Equal(t, tt.tenantID, response.TenantID)
				assert.Len(t, response.Peers, 1)
			}
		})
	}
}

func TestPeerManager_validatePeerAccess(t *testing.T) {
	// Создаем RelayManager для тестирования validatePeerAccess
	rm := &RelayManager{}

	tests := []struct {
		name      string
		peerID    string
		whitelist *PeerWhitelist
		expected  bool
	}{
		{
			name:      "no whitelist - all allowed",
			peerID:    "test-peer-1",
			whitelist: nil,
			expected:  true,
		},
		{
			name:   "auto approve enabled",
			peerID: "test-peer-1",
			whitelist: &PeerWhitelist{
				AutoApprove: true,
			},
			expected: true,
		},
		{
			name:   "peer in whitelist",
			peerID: "test-peer-1",
			whitelist: &PeerWhitelist{
				AllowedPeers: []string{"test-peer-1", "test-peer-2"},
				AutoApprove:  false,
			},
			expected: true,
		},
		{
			name:   "peer not in whitelist",
			peerID: "test-peer-3",
			whitelist: &PeerWhitelist{
				AllowedPeers: []string{"test-peer-1", "test-peer-2"},
				AutoApprove:  false,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rm.validatePeerAccess(tt.peerID, tt.whitelist)
			assert.Equal(t, tt.expected, result)
		})
	}
}