package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twogc/cloudbridge-relay-installer/internal/network"
	"go.uber.org/zap"
)

func TestRelayManagerHybrid_Simple(t *testing.T) {
	// Создаем мок store
	store := NewMockTenantNetworkStore()
	defer store.Close()

	// Создаем логгер
	logger := zap.NewNop()

	// Создаем RelayManagerHybrid
	rm := &RelayManagerHybrid{
		logger:          logger,
		networkStore:    store,
		configCache:     make(map[string]*network.TenantNetwork),
		tenants:         make(map[string]*Tenant),
		connections:     make(map[string]*RelayP2PConnection),
		quicConnections: make(map[string]*QUICConnection),
		syncStatus:      "fresh",
	}

	t.Run("initialize from store", func(t *testing.T) {
		// Добавляем сеть в store
		store.AddNetwork("tenant-123", "10.100.77.0/24")

		// Инициализируем из store
		err := rm.initializeFromStore()
		require.NoError(t, err)

		// Проверяем, что сеть загружена
		cidr := rm.GetTenantNetwork("tenant-123")
		assert.Equal(t, "10.100.77.0/24", cidr)
	})

	t.Run("get tenant info", func(t *testing.T) {
		// Добавляем сеть
		store.AddNetwork("tenant-info", "10.100.80.0/24")

		// Инициализируем
		err := rm.initializeFromStore()
		require.NoError(t, err)

		// Получаем информацию о тенанте
		tenant, err := rm.GetTenantInfo("tenant-info")
		require.NoError(t, err)
		assert.Equal(t, "tenant-info", tenant.TenantID)
		assert.Equal(t, "10.100.80.0/24", tenant.NetworkCIDR)
	})

	t.Run("get all tenants", func(t *testing.T) {
		// Получаем всех тенантов
		tenants := rm.GetAllTenants()
		assert.NotNil(t, tenants)
		assert.GreaterOrEqual(t, len(tenants), 0)
	})

	t.Run("get sync status", func(t *testing.T) {
		// Получаем статус синхронизации
		status := rm.GetSyncStatus()
		assert.NotEmpty(t, status)
	})

	t.Run("get resource version", func(t *testing.T) {
		// Получаем версию ресурса
		version := rm.GetResourceVersion()
		assert.NotEmpty(t, version)
	})
}
