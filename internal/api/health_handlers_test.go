package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewHealthManager(t *testing.T) {
	hub := NewHub(zap.NewNop())
	peers := make(map[string]*Peer)
	connections := make(map[string]*PeerConnection)
	logger := zap.NewNop()

	healthManager := NewHealthManager(hub, peers, connections, logger)

	assert.NotNil(t, healthManager)
}

func TestHealthManager_HealthCheck(t *testing.T) {
	hub := NewHub(zap.NewNop())
	peers := make(map[string]*Peer)
	connections := make(map[string]*PeerConnection)
	logger := zap.NewNop()

	healthManager := NewHealthManager(hub, peers, connections, logger)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/health", nil)

	healthManager.HealthCheck(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHealthManager_GetSystemMetrics(t *testing.T) {
	hub := NewHub(zap.NewNop())
	peers := make(map[string]*Peer)
	connections := make(map[string]*PeerConnection)
	logger := zap.NewNop()

	healthManager := NewHealthManager(hub, peers, connections, logger)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/metrics", nil)

	healthManager.GetSystemMetrics(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHealthManager_GetPeerMetrics(t *testing.T) {
	hub := NewHub(zap.NewNop())
	peers := make(map[string]*Peer)
	connections := make(map[string]*PeerConnection)
	logger := zap.NewNop()

	healthManager := NewHealthManager(hub, peers, connections, logger)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/health/peers", nil)

	healthManager.GetPeerMetrics(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHealthManager_GetConnectionMetrics(t *testing.T) {
	hub := NewHub(zap.NewNop())
	peers := make(map[string]*Peer)
	connections := make(map[string]*PeerConnection)
	logger := zap.NewNop()

	healthManager := NewHealthManager(hub, peers, connections, logger)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/health/connections", nil)

	healthManager.GetConnectionMetrics(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHealthManager_GetRelayStatus(t *testing.T) {
	hub := NewHub(zap.NewNop())
	peers := make(map[string]*Peer)
	connections := make(map[string]*PeerConnection)
	logger := zap.NewNop()

	healthManager := NewHealthManager(hub, peers, connections, logger)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/relay/status", nil)

	healthManager.GetRelayStatus(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHealthManager_OpenConnection(t *testing.T) {
	hub := NewHub(zap.NewNop())
	peers := make(map[string]*Peer)
	connections := make(map[string]*PeerConnection)
	logger := zap.NewNop()

	healthManager := NewHealthManager(hub, peers, connections, logger)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	requestBody := `{"tenant_id": "tenant1", "peer_id": "peer1", "session_id": "session1", "protocol": "tcp"}`
	c.Request = httptest.NewRequest("POST", "/relay/connections/open", strings.NewReader(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	healthManager.OpenConnection(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHealthManager_CloseConnection(t *testing.T) {
	hub := NewHub(zap.NewNop())
	peers := make(map[string]*Peer)
	connections := make(map[string]*PeerConnection)
	logger := zap.NewNop()

	healthManager := NewHealthManager(hub, peers, connections, logger)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	requestBody := `{"tenant_id": "tenant1", "peer_id": "peer1", "session_id": "session1", "protocol": "tcp"}`
	c.Request = httptest.NewRequest("POST", "/relay/connections/close", strings.NewReader(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	healthManager.CloseConnection(c)

	assert.Equal(t, http.StatusOK, w.Code)
}
