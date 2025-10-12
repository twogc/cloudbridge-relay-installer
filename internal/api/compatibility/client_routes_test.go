package compatibility

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func setupTestRouter() (*gin.Engine, *ClientCompatibilityLayer) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger, _ := zap.NewDevelopment()

	ccl := NewClientCompatibilityLayer(logger)

	// Create native handlers struct
	handlers := &NativeHandlers{
		PeerRegister:   mockPeerRegisterHandler,
		PeerDiscovery:  mockPeerDiscoveryHandler,
		PeerHeartbeat:  mockHeartbeatHandler,
		PeerStatus:     mockStatusHandler,
		GetPeer:        mockGetPeerHandler,
		DeletePeer:     mockDeletePeerHandler,
		ICECredentials: mockICECredentialsHandler,
	}

	// Register compatibility routes with handlers
	ccl.RegisterRoutesWithHandlers(router, handlers)

	return router, ccl
}

// Mock handlers for native endpoints
func mockPeerRegisterHandler(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"peer_id":   "test-peer-123",
		"tenant_id": tenantID,
		"message":   "peer registered successfully",
	})
}

func mockPeerDiscoveryHandler(c *gin.Context) {
	peerID := c.Param("peer_id")
	tenantID, _ := c.Get("tenant_id")
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"tenant_id": tenantID,
		"requester": peerID,
		"peers": []map[string]interface{}{
			{
				"peer_id":   "peer-1",
				"is_online": true,
			},
			{
				"peer_id":   "peer-2",
				"is_online": true,
			},
		},
	})
}

func mockHeartbeatHandler(c *gin.Context) {
	peerID := c.Param("peer_id")
	tenantID, _ := c.Get("tenant_id")
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"peer_id":   peerID,
		"tenant_id": tenantID,
		"status":    "active",
	})
}

func mockStatusHandler(c *gin.Context) {
	peerID := c.Param("peer_id")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"peer_id": peerID,
	})
}

func mockGetPeerHandler(c *gin.Context) {
	peerID := c.Param("peer_id")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"peer_id": peerID,
		"status":  "online",
	})
}

func mockDeletePeerHandler(c *gin.Context) {
	peerID := c.Param("peer_id")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"peer_id": peerID,
		"message": "peer deleted",
	})
}

func mockICECredentialsHandler(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"tenant_id": tenantID,
		"ufrag":     "test-ufrag",
		"pwd":       "test-pwd",
	})
}

// Tests

func TestClientCompatibilityLayer_PeerRegister(t *testing.T) {
	router, _ := setupTestRouter()

	// Test legacy client path
	reqBody := map[string]interface{}{
		"public_key": "test-key",
		"allowed_ips": []string{"10.0.0.1/32"},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/v1/tenants/test-tenant/peers/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))
	assert.Equal(t, "test-tenant", response["tenant_id"])
	assert.Equal(t, "test-peer-123", response["peer_id"])
}

func TestClientCompatibilityLayer_PeerDiscovery(t *testing.T) {
	router, _ := setupTestRouter()

	req, _ := http.NewRequest("GET", "/api/v1/tenants/test-tenant/peers/discover?peer_id=test-peer", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))
	assert.Equal(t, "test-tenant", response["tenant_id"])

	peers := response["peers"].([]interface{})
	assert.Equal(t, 2, len(peers))
}

func TestClientCompatibilityLayer_Heartbeat(t *testing.T) {
	router, _ := setupTestRouter()

	reqBody := map[string]interface{}{
		"status": "active",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/v1/tenants/test-tenant/peers/test-peer/heartbeat", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))
	assert.Equal(t, "test-peer", response["peer_id"])
	assert.Equal(t, "test-tenant", response["tenant_id"])
}

func TestClientCompatibilityLayer_Status(t *testing.T) {
	router, _ := setupTestRouter()

	reqBody := map[string]interface{}{
		"status": "online",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/v1/tenants/test-tenant/peers/test-peer/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))
}

func TestClientCompatibilityLayer_GetPeer(t *testing.T) {
	router, _ := setupTestRouter()

	req, _ := http.NewRequest("GET", "/api/v1/tenants/test-tenant/peers/test-peer", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))
	assert.Equal(t, "test-peer", response["peer_id"])
}

func TestClientCompatibilityLayer_DeletePeer(t *testing.T) {
	router, _ := setupTestRouter()

	req, _ := http.NewRequest("DELETE", "/api/v1/tenants/test-tenant/peers/test-peer", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))
	assert.Equal(t, "peer deleted", response["message"])
}

func TestClientCompatibilityLayer_ICECredentials(t *testing.T) {
	router, _ := setupTestRouter()

	reqBody := map[string]interface{}{
		"ufrag": "client-ufrag",
		"pwd":   "client-pwd",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/v1/tenants/test-tenant/ice/credentials", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))
	assert.Equal(t, "test-tenant", response["tenant_id"])
}

func TestExtractPeerID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ccl := NewClientCompatibilityLayer(logger)

	tests := []struct {
		name     string
		setup    func(*gin.Context)
		expected string
	}{
		{
			name: "from context",
			setup: func(c *gin.Context) {
				c.Set("peer_id", "peer-from-context")
			},
			expected: "peer-from-context",
		},
		{
			name: "from claims",
			setup: func(c *gin.Context) {
				c.Set("claims", map[string]interface{}{
					"peer_id": "peer-from-claims",
				})
			},
			expected: "peer-from-claims",
		},
		{
			name: "from claims sub",
			setup: func(c *gin.Context) {
				c.Set("claims", map[string]interface{}{
					"sub": "peer-from-sub",
				})
			},
			expected: "peer-from-sub",
		},
		{
			name: "from query parameter",
			setup: func(c *gin.Context) {
				c.Request, _ = http.NewRequest("GET", "/?peer_id=peer-from-query", nil)
			},
			expected: "peer-from-query",
		},
		{
			name: "from header",
			setup: func(c *gin.Context) {
				c.Request, _ = http.NewRequest("GET", "/", nil)
				c.Request.Header.Set("X-Peer-ID", "peer-from-header")
			},
			expected: "peer-from-header",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest("GET", "/", nil)

			tt.setup(c)

			result := ccl.extractPeerID(c)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTenantIDMiddleware(t *testing.T) {
	router := gin.New()
	router.Use(TenantIDMiddleware())
	router.GET("/test/:tenant_id/path", func(c *gin.Context) {
		tenantID, exists := c.Get("tenant_id")
		assert.True(t, exists)
		c.JSON(http.StatusOK, gin.H{"tenant_id": tenantID})
	})

	req, _ := http.NewRequest("GET", "/test/my-tenant/path", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "my-tenant", response["tenant_id"])
}

func TestHealthCheck(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ccl := NewClientCompatibilityLayer(logger)

	health := ccl.HealthCheck()

	assert.Equal(t, "client_compatibility_layer", health["name"])
	assert.Equal(t, "healthy", health["status"])
	assert.Equal(t, "1.0.0", health["version"])

	routes := health["routes_registered"].([]string)
	assert.Greater(t, len(routes), 0)
}

func TestCompatibilityMetrics(t *testing.T) {
	metrics := &CompatibilityMetrics{
		PathTranslations: make(map[string]int64),
	}

	router := gin.New()
	router.Use(MetricsMiddleware(metrics))
	router.GET("/api/v1/tenants/:tenant_id/peers/register", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// Make request
	req, _ := http.NewRequest("GET", "/api/v1/tenants/test/peers/register", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check metrics
	assert.Equal(t, int64(1), metrics.RequestsTotal)
	assert.Equal(t, int64(1), metrics.SuccessCount)
	assert.Equal(t, int64(0), metrics.ErrorCount)
	assert.Equal(t, int64(1), metrics.PathTranslations["/api/v1/tenants/test/peers/register"])
}

// Benchmark tests

func BenchmarkPeerRegister(b *testing.B) {
	router, _ := setupTestRouter()
	reqBody := map[string]interface{}{
		"public_key":  "test-key",
		"allowed_ips": []string{"10.0.0.1/32"},
	}
	body, _ := json.Marshal(reqBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/api/v1/tenants/test-tenant/peers/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkPeerDiscovery(b *testing.B) {
	router, _ := setupTestRouter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/v1/tenants/test-tenant/peers/discover?peer_id=test-peer", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
