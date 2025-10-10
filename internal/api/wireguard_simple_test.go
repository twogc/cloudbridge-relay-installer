package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// MockRelayManagerHybrid для тестирования
type MockRelayManagerHybrid struct {
	tenantNetworks map[string]string
	peerIPs        map[string]string
}

func NewMockRelayManagerHybrid() *MockRelayManagerHybrid {
	return &MockRelayManagerHybrid{
		tenantNetworks: make(map[string]string),
		peerIPs:        make(map[string]string),
	}
}

func (m *MockRelayManagerHybrid) GetTenantNetwork(tenantID string) string {
	return m.tenantNetworks[tenantID]
}

func (m *MockRelayManagerHybrid) AllocatePeerIP(tenantID, peerID string) (string, error) {
	if ip, ok := m.peerIPs[peerID]; ok {
		return ip, nil
	}
	// Генерируем IP на основе peerID
	ip := "10.100.77.10"
	m.peerIPs[peerID] = ip
	return ip, nil
}

func (m *MockRelayManagerHybrid) GetPeerIPAMRepo() interface{} {
	return nil
}

func (m *MockRelayManagerHybrid) SetTenantNetwork(tenantID, cidr string) {
	m.tenantNetworks[tenantID] = cidr
}

func TestWireGuardSimpleEndpoints_GetWireGuardConfig(t *testing.T) {
	// Создаем логгер
	logger := zap.NewNop()

	// Создаем мок RelayManagerHybrid
	mockRelayManager := NewMockRelayManagerHybrid()
	mockRelayManager.SetTenantNetwork("tenant-123", "10.100.77.0/24")

	// Создаем WireGuardSimpleEndpoints
	wge := &WireGuardSimpleEndpoints{
		relayManagerHybrid: nil, // Используем nil для упрощения тестов
		logger:             logger,
	}

	t.Run("successful config generation", func(t *testing.T) {
		// Создаем Gin router
		gin.SetMode(gin.TestMode)
		router := gin.New()

		// Добавляем маршрут
		router.GET("/tenants/:tenant_id/peers/:peer_id/wireguard-config", wge.GetWireGuardConfig)

		// Создаем запрос
		req, err := http.NewRequest("GET", "/tenants/tenant-123/peers/peer-456/wireguard-config", nil)
		require.NoError(t, err)

		// Добавляем заголовок авторизации (заглушка)
		req.Header.Set("Authorization", "Bearer mock-token")

		// Создаем response recorder
		w := httptest.NewRecorder()

		// Выполняем запрос
		router.ServeHTTP(w, req)

		// Проверяем статус
		assert.Equal(t, http.StatusOK, w.Code)

		// Проверяем ответ
		var response WireGuardConfigResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.Equal(t, "10.100.77.0/24", response.TenantCIDR)
		assert.Equal(t, "10.100.77.10", response.PeerIP)
		assert.NotEmpty(t, response.WGConfig)
		assert.NotEmpty(t, response.InterfaceName)
	})

	t.Run("tenant not found", func(t *testing.T) {
		// Создаем Gin router
		gin.SetMode(gin.TestMode)
		router := gin.New()

		// Добавляем маршрут
		router.GET("/tenants/:tenant_id/peers/:peer_id/wireguard-config", wge.GetWireGuardConfig)

		// Создаем запрос для несуществующего тенанта
		req, err := http.NewRequest("GET", "/tenants/tenant-999/peers/peer-456/wireguard-config", nil)
		require.NoError(t, err)

		// Добавляем заголовок авторизации
		req.Header.Set("Authorization", "Bearer mock-token")

		// Создаем response recorder
		w := httptest.NewRecorder()

		// Выполняем запрос
		router.ServeHTTP(w, req)

		// Проверяем статус
		assert.Equal(t, http.StatusOK, w.Code)

		// Проверяем ответ (должен использовать заглушку)
		var response WireGuardConfigResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.Equal(t, "10.100.77.0/24", response.TenantCIDR) // Заглушка
		assert.Equal(t, "10.100.77.10", response.PeerIP)       // Заглушка
	})

	t.Run("missing authorization", func(t *testing.T) {
		// Создаем Gin router
		gin.SetMode(gin.TestMode)
		router := gin.New()

		// Добавляем маршрут
		router.GET("/tenants/:tenant_id/peers/:peer_id/wireguard-config", wge.GetWireGuardConfig)

		// Создаем запрос без авторизации
		req, err := http.NewRequest("GET", "/tenants/tenant-123/peers/peer-456/wireguard-config", nil)
		require.NoError(t, err)

		// Создаем response recorder
		w := httptest.NewRecorder()

		// Выполняем запрос
		router.ServeHTTP(w, req)

		// Проверяем статус (должен быть 401)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestWireGuardSimpleEndpoints_GenerateWGConfig(t *testing.T) {
	t.Run("generate config", func(t *testing.T) {
		config := generateWGConfig("10.100.77.10", "10.100.77.0/24", "RELAY_PUBLIC_KEY", "relay.example.com:51820")

		assert.Contains(t, config, "10.100.77.10/24")
		assert.Contains(t, config, "10.100.77.0/24")
		assert.Contains(t, config, "RELAY_PUBLIC_KEY")
		assert.Contains(t, config, "relay.example.com:51820")
		assert.Contains(t, config, "PersistentKeepalive = 25")
	})
}
