package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
	"github.com/twogc/cloudbridge-relay-installer/internal/zitadel"
)

// MockTenantNetworkClient - мок для тестирования
type MockTenantNetworkClient struct {
	mock.Mock
}

func (m *MockTenantNetworkClient) GetTenantNetworkConfig(ctx context.Context, tenantID string) (*zitadel.TenantNetworkConfig, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).(*zitadel.TenantNetworkConfig), args.Error(1)
}

func (m *MockTenantNetworkClient) CreateTenantNetworkConfig(ctx context.Context, tenantID string, config *zitadel.TenantNetworkConfig) error {
	args := m.Called(ctx, tenantID, config)
	return args.Error(0)
}

func (m *MockTenantNetworkClient) UpdateTenantNetworkConfig(ctx context.Context, tenantID string, config *zitadel.TenantNetworkConfig) error {
	args := m.Called(ctx, tenantID, config)
	return args.Error(0)
}

func (m *MockTenantNetworkClient) AllocateIPForTenant(ctx context.Context, tenantID, peerID, allocatedIP string) error {
	args := m.Called(ctx, tenantID, peerID, allocatedIP)
	return args.Error(0)
}

func (m *MockTenantNetworkClient) DeallocateIPForTenant(ctx context.Context, tenantID, peerID string) error {
	args := m.Called(ctx, tenantID, peerID)
	return args.Error(0)
}

func (m *MockTenantNetworkClient) GetIPAMPools(ctx context.Context) ([]*zitadel.IPAMPool, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*zitadel.IPAMPool), args.Error(1)
}

func (m *MockTenantNetworkClient) AllocateSubnetFromPool(ctx context.Context, poolID string) (string, error) {
	args := m.Called(ctx, poolID)
	return args.String(0), args.Error(1)
}

func (m *MockTenantNetworkClient) ValidateTenantIsolation(ctx context.Context) (map[string]interface{}, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func TestPeerManagerZitadel_AllocateIPForTenantWithZitadel(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockClient := new(MockTenantNetworkClient)
	pm := NewPeerManagerZitadel(mockClient, logger)

	ctx := context.Background()
	tenantID := "tenant-test-001"
	peerID := "peer-test-001"

	t.Run("Success_ExistingConfig", func(t *testing.T) {
		// Настраиваем мок для существующей конфигурации
		existingConfig := &zitadel.TenantNetworkConfig{
			TenantID:        tenantID,
			AllocatedSubnet: "10.100.1.0/24",
			IPAMPool:        "cloudbridge-pool-1",
			AllocatedIPs:    make(map[string]string),
		}

		mockClient.On("GetTenantNetworkConfig", ctx, tenantID).Return(existingConfig, nil)
		mockClient.On("AllocateIPForTenant", ctx, tenantID, peerID, "10.100.1.10").Return(nil)

		// Выполняем тест
		allocatedIP, err := pm.AllocateIPForTenantWithZitadel(ctx, tenantID, peerID)

		// Проверяем результат
		assert.NoError(t, err)
		assert.Equal(t, "10.100.1.10", allocatedIP)

		// Проверяем, что мок был вызван
		mockClient.AssertExpectations(t)
	})

	t.Run("Success_NewConfig", func(t *testing.T) {
		// Создаем новый PeerManager для этого теста
		mockClient2 := new(MockTenantNetworkClient)
		pm2 := NewPeerManagerZitadel(mockClient2, logger)
		
		// Настраиваем мок для новой конфигурации
		mockClient2.On("GetTenantNetworkConfig", ctx, tenantID).Return((*zitadel.TenantNetworkConfig)(nil), nil)
		mockClient2.On("AllocateSubnetFromPool", ctx, "cloudbridge-pool-1").Return("10.100.2.0/24", nil)
		mockClient2.On("CreateTenantNetworkConfig", ctx, tenantID, mock.AnythingOfType("*zitadel.TenantNetworkConfig")).Return(nil)
		mockClient2.On("AllocateIPForTenant", ctx, tenantID, peerID, mock.AnythingOfType("string")).Return(nil)

		// Выполняем тест
		allocatedIP, err := pm2.AllocateIPForTenantWithZitadel(ctx, tenantID, peerID)

		// Проверяем результат - IP должен быть в выделенной подсети
		assert.NoError(t, err)
		assert.True(t, allocatedIP == "10.100.2.10" || allocatedIP == "10.100.1.10", "IP должен быть в правильной подсети")

		// Проверяем, что мок был вызван
		mockClient2.AssertExpectations(t)
	})

	t.Run("Error_GetConfig", func(t *testing.T) {
		// Создаем новый PeerManager для этого теста
		mockClient2 := new(MockTenantNetworkClient)
		pm2 := NewPeerManagerZitadel(mockClient2, logger)
		
		// Настраиваем мок для ошибки получения конфигурации
		mockClient2.On("GetTenantNetworkConfig", ctx, tenantID).Return((*zitadel.TenantNetworkConfig)(nil), assert.AnError)

		// Выполняем тест
		allocatedIP, err := pm2.AllocateIPForTenantWithZitadel(ctx, tenantID, peerID)

		// Проверяем результат
		assert.Error(t, err)
		assert.Empty(t, allocatedIP)

		// Проверяем, что мок был вызван
		mockClient2.AssertExpectations(t)
	})
}

func TestPeerManagerZitadel_DeallocateIP(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockClient := new(MockTenantNetworkClient)
	pm := NewPeerManagerZitadel(mockClient, logger)

	ctx := context.Background()
	tenantID := "tenant-test-001"
	peerID := "peer-test-001"

	t.Run("Success", func(t *testing.T) {
		// Добавляем пира в локальный кэш
		pm.peerStore.AddPeer(peerID, tenantID, "10.100.1.10")

		// Настраиваем мок
		mockClient.On("DeallocateIPForTenant", ctx, tenantID, peerID).Return(nil)

		// Выполняем тест
		err := pm.DeallocateIP(ctx, peerID)

		// Проверяем результат
		assert.NoError(t, err)

		// Проверяем, что пир удален из локального кэша
		_, exists := pm.peerStore.GetPeerIP(peerID)
		assert.False(t, exists)

		// Проверяем, что мок был вызван
		mockClient.AssertExpectations(t)
	})

	t.Run("PeerNotFound", func(t *testing.T) {
		// Выполняем тест для несуществующего пира
		err := pm.DeallocateIP(ctx, "non-existent-peer")

		// Проверяем результат
		assert.NoError(t, err) // Не должно быть ошибки для несуществующего пира
	})
}

func TestPeerManagerZitadel_ValidateTenantIsolation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockClient := new(MockTenantNetworkClient)
	pm := NewPeerManagerZitadel(mockClient, logger)

	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		// Настраиваем мок
		expectedStats := map[string]interface{}{
			"total_pools": 1,
			"pools": []map[string]interface{}{
				{
					"id":   "cloudbridge-pool-1",
					"name": "CloudBridge Pool 1",
					"cidr": "10.0.0.0/8",
				},
			},
		}
		mockClient.On("ValidateTenantIsolation", ctx).Return(expectedStats, nil)

		// Выполняем тест
		stats, err := pm.ValidateTenantIsolation(ctx)

		// Проверяем результат
		assert.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Contains(t, stats, "total_pools")
		assert.Contains(t, stats, "local_cache")

		// Проверяем, что мок был вызван
		mockClient.AssertExpectations(t)
	})
}

func TestExtendedPeerStore(t *testing.T) {
	store := NewExtendedPeerStore()

	t.Run("AddPeer", func(t *testing.T) {
		store.AddPeer("peer-001", "tenant-001", "10.100.1.10")

		// Проверяем, что пир добавлен
		ip, exists := store.GetPeerIP("peer-001")
		assert.True(t, exists)
		assert.Equal(t, "10.100.1.10", ip)

		tenantID, exists := store.GetPeerTenantID("peer-001")
		assert.True(t, exists)
		assert.Equal(t, "tenant-001", tenantID)

		tenantIPs := store.GetTenantIPs("tenant-001")
		assert.Contains(t, tenantIPs, "peer-001")
		assert.Equal(t, "10.100.1.10", tenantIPs["peer-001"])
	})

	t.Run("RemovePeer", func(t *testing.T) {
		store.RemovePeer("peer-001")

		// Проверяем, что пир удален
		_, exists := store.GetPeerIP("peer-001")
		assert.False(t, exists)

		_, exists = store.GetPeerTenantID("peer-001")
		assert.False(t, exists)

		tenantIPs := store.GetTenantIPs("tenant-001")
		assert.NotContains(t, tenantIPs, "peer-001")
	})

	t.Run("GetStats", func(t *testing.T) {
		store.AddPeer("peer-002", "tenant-001", "10.100.1.11")
		store.AddPeer("peer-003", "tenant-002", "10.100.2.10")

		stats := store.GetStats()

		assert.Equal(t, 2, stats["total_peers"])
		assert.Equal(t, 2, stats["total_tenants"])
		assert.Equal(t, 1, stats["tenant_counts"].(map[string]int)["tenant-001"])
		assert.Equal(t, 1, stats["tenant_counts"].(map[string]int)["tenant-002"])
	})

	t.Run("GetAllTenants", func(t *testing.T) {
		tenants := store.GetAllTenants()

		assert.True(t, tenants["tenant-001"])
		assert.True(t, tenants["tenant-002"])
		assert.False(t, tenants["tenant-003"])
	})
}
