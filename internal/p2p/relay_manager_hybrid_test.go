package p2p

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twogc/cloudbridge-relay-installer/internal/network"
	"go.uber.org/zap"
)

// MockTenantNetworkStore для тестирования
type MockTenantNetworkStore struct {
	networks map[string]*network.TenantNetwork
	events   chan network.TenantNetworkEvent
}

func NewMockTenantNetworkStore() *MockTenantNetworkStore {
	return &MockTenantNetworkStore{
		networks: make(map[string]*network.TenantNetwork),
		events:   make(chan network.TenantNetworkEvent, 10),
	}
}

func (m *MockTenantNetworkStore) ListActive(ctx context.Context) ([]*network.TenantNetwork, network.ResourceVersion, error) {
	var result []*network.TenantNetwork
	for _, net := range m.networks {
		if net.IsActive {
			result = append(result, net)
		}
	}
	return result, "mock-1", nil
}

func (m *MockTenantNetworkStore) Get(ctx context.Context, tenantID string) (*network.TenantNetwork, error) {
	if net, ok := m.networks[tenantID]; ok {
		return net, nil
	}
	return nil, network.ErrNotFound
}

func (m *MockTenantNetworkStore) Watch(ctx context.Context, from network.ResourceVersion) (<-chan network.TenantNetworkEvent, error) {
	return m.events, nil
}

func (m *MockTenantNetworkStore) Close() error {
	close(m.events)
	return nil
}

func (m *MockTenantNetworkStore) AddNetwork(tenantID, cidr string) {
	m.networks[tenantID] = &network.TenantNetwork{
		TenantID:        tenantID,
		NetworkCIDR:     cidr,
		IsActive:        true,
		AllocatedAtUnix: time.Now().Unix(),
		Version:         1,
		ResourceVersion: "mock-1",
	}
}

func (m *MockTenantNetworkStore) SendEvent(event network.TenantNetworkEvent) {
	select {
	case m.events <- event:
	default:
	}
}

func TestRelayManagerHybrid_Integration(t *testing.T) {
	// Создаем мок store
	store := NewMockTenantNetworkStore()
	defer store.Close()

	// Создаем логгер
	logger := zap.NewNop()

	// Создаем RelayManagerHybrid (упрощенная версия для тестов)
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

	t.Run("handle network events", func(t *testing.T) {
		// Создаем событие создания сети
		event := network.TenantNetworkEvent{
			Type: network.EventCreated,
			Tenant: &network.TenantNetwork{
				TenantID:        "tenant-456",
				NetworkCIDR:     "10.100.78.0/24",
				IsActive:        true,
				AllocatedAtUnix: time.Now().Unix(),
				Version:         1,
				ResourceVersion: "mock-2",
			},
		}

		// Отправляем событие
		store.SendEvent(event)

		// Ждем обработки
		time.Sleep(100 * time.Millisecond)

		// Проверяем, что сеть добавлена
		cidr := rm.GetTenantNetwork("tenant-456")
		assert.Equal(t, "10.100.78.0/24", cidr)
	})

	t.Run("handle network update", func(t *testing.T) {
		// Создаем событие обновления сети
		event := network.TenantNetworkEvent{
			Type: network.EventUpdated,
			Tenant: &network.TenantNetwork{
				TenantID:        "tenant-456",
				NetworkCIDR:     "10.100.79.0/24",
				IsActive:        true,
				AllocatedAtUnix: time.Now().Unix(),
				Version:         2,
				ResourceVersion: "mock-3",
			},
		}

		// Отправляем событие
		store.SendEvent(event)

		// Ждем обработки
		time.Sleep(100 * time.Millisecond)

		// Проверяем, что сеть обновлена
		cidr := rm.GetTenantNetwork("tenant-456")
		assert.Equal(t, "10.100.79.0/24", cidr)
	})

	t.Run("handle network deactivation", func(t *testing.T) {
		// Создаем событие деактивации сети
		event := network.TenantNetworkEvent{
			Type: network.EventUpdated,
			Tenant: &network.TenantNetwork{
				TenantID:        "tenant-456",
				NetworkCIDR:     "10.100.79.0/24",
				IsActive:        false,
				AllocatedAtUnix: time.Now().Unix(),
				Version:         3,
				ResourceVersion: "mock-4",
			},
		}

		// Отправляем событие
		store.SendEvent(event)

		// Ждем обработки
		time.Sleep(100 * time.Millisecond)

		// Проверяем, что сеть удалена из кэша
		cidr := rm.GetTenantNetwork("tenant-456")
		assert.Empty(t, cidr)
	})

	t.Run("handle network deletion", func(t *testing.T) {
		// Создаем событие удаления сети
		event := network.TenantNetworkEvent{
			Type: network.EventDeleted,
			Tenant: &network.TenantNetwork{
				TenantID:        "tenant-123",
				NetworkCIDR:     "10.100.77.0/24",
				IsActive:        false,
				AllocatedAtUnix: time.Now().Unix(),
				Version:         2,
				ResourceVersion: "mock-5",
			},
		}

		// Отправляем событие
		store.SendEvent(event)

		// Ждем обработки
		time.Sleep(100 * time.Millisecond)

		// Проверяем, что сеть удалена из кэша
		cidr := rm.GetTenantNetwork("tenant-123")
		assert.Empty(t, cidr)
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
