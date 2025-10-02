package nat

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

func createTestLogger() *logger.Logger {
	testLogger, _ := logger.New(logger.Config{
		Level:  logger.ERROR, // Уменьшаем уровень логирования для тестов
		Output: "stdout",
		JSON:   false,
	})
	return testLogger
}

func TestTURNPoolManager_CreatePool(t *testing.T) {
	testLogger := createTestLogger()
	config := &TURNPoolConfig{
		DefaultPoolName:         "test",
		AllocationTimeout:       30 * time.Second,
		HealthCheckInterval:     30 * time.Second,
		HealthCheckTimeout:      5 * time.Second,
		MaxAllocationsPerClient: 10,
		LoadBalancingStrategy:   "round_robin",
		WeightedThreshold:       0.8,
		MetricsEnabled:          true,
		StatsInterval:           60 * time.Second,
		LatencyHistorySize:      100,
		FailoverEnabled:         true,
		FailoverTimeout:         10 * time.Second,
		MaxRetries:              3,
		BackoffMultiplier:       2.0,
	}

	manager := NewTURNPoolManager(config, testLogger.GetZapLogger())
	defer manager.Stop()

	// Создаем пул
	err := manager.CreatePool("test-pool")
	require.NoError(t, err)

	// Проверяем, что пул создан
	stats := manager.GetPoolStats()
	assert.Contains(t, stats, "test-pool")
}

func TestTURNPoolManager_AddServerToPool(t *testing.T) {
	testLogger := createTestLogger()
	config := &TURNPoolConfig{
		DefaultPoolName:         "test",
		AllocationTimeout:       30 * time.Second,
		HealthCheckInterval:     30 * time.Second,
		HealthCheckTimeout:      5 * time.Second,
		MaxAllocationsPerClient: 10,
		LoadBalancingStrategy:   "round_robin",
		WeightedThreshold:       0.8,
		MetricsEnabled:          true,
		StatsInterval:           60 * time.Second,
		LatencyHistorySize:      100,
		FailoverEnabled:         true,
		FailoverTimeout:         10 * time.Second,
		MaxRetries:              3,
		BackoffMultiplier:       2.0,
	}

	manager := NewTURNPoolManager(config, testLogger.GetZapLogger())
	defer manager.Stop()

	// Создаем пул
	err := manager.CreatePool("test-pool")
	require.NoError(t, err)

	// Создаем сервер
	server := &TURNServerInfo{
		TURNServerPool: &TURNServerPool{
			Address:        "turn1.example.com:3478",
			Username:       "testuser",
			Password:       "testpass",
			Realm:          "example.com",
			Priority:       100,
			Weight:         80,
			MaxConnections: 1000,
			Region:         "us-east",
			Provider:       "aws",
			Capabilities:   []string{"udp", "tcp"},
			Tags:           map[string]string{"env": "test"},
		},
		Region:        "us-east",
		Provider:      "aws",
		Capabilities:  []string{"udp", "tcp"},
		Tags:          map[string]string{"env": "test"},
		Status:        "active",
		StatusChanged: time.Now(),
	}

	// Добавляем сервер в пул
	err = manager.AddServerToPool("test-pool", server)
	require.NoError(t, err)

	// Проверяем, что сервер добавлен
	stats := manager.GetPoolStats()
	assert.Contains(t, stats, "test-pool")
}

func TestTURNPoolManager_AllocateTURN(t *testing.T) {
	testLogger := createTestLogger()
	config := &TURNPoolConfig{
		DefaultPoolName:         "test",
		AllocationTimeout:       30 * time.Second,
		HealthCheckInterval:     30 * time.Second,
		HealthCheckTimeout:      5 * time.Second,
		MaxAllocationsPerClient: 10,
		LoadBalancingStrategy:   "round_robin",
		WeightedThreshold:       0.8,
		MetricsEnabled:          true,
		StatsInterval:           60 * time.Second,
		LatencyHistorySize:      100,
		FailoverEnabled:         true,
		FailoverTimeout:         10 * time.Second,
		MaxRetries:              3,
		BackoffMultiplier:       2.0,
	}

	manager := NewTURNPoolManager(config, testLogger.GetZapLogger())
	defer manager.Stop()

	// Создаем пул
	err := manager.CreatePool("test-pool")
	require.NoError(t, err)

	// Создаем сервер
	server := &TURNServerInfo{
		TURNServerPool: &TURNServerPool{
			Address:        "turn1.example.com:3478",
			Username:       "testuser",
			Password:       "testpass",
			Realm:          "example.com",
			Priority:       100,
			Weight:         80,
			MaxConnections: 1000,
			Region:         "us-east",
			Provider:       "aws",
			Capabilities:   []string{"udp", "tcp"},
			Tags:           map[string]string{"env": "test"},
		},
		Region:        "us-east",
		Provider:      "aws",
		Capabilities:  []string{"udp", "tcp"},
		Tags:          map[string]string{"env": "test"},
		Status:        "active",
		StatusChanged: time.Now(),
	}

	// Добавляем сервер в пул
	err = manager.AddServerToPool("test-pool", server)
	require.NoError(t, err)

	// Создаем allocation
	preferences := &AllocationPreferences{
		PreferredRegion:   "us-east",
		PreferredProvider: "aws",
		MaxLatency:        100 * time.Millisecond,
		PoolName:          "test-pool",
	}

	allocation, err := manager.AllocateTURN("test-client", preferences)
	require.NoError(t, err)
	assert.NotNil(t, allocation)
	assert.Equal(t, "test-client", allocation.ClientID)
	assert.Equal(t, "active", allocation.Status)
}

func TestTURNPoolManager_ReleaseTURN(t *testing.T) {
	testLogger := createTestLogger()
	config := &TURNPoolConfig{
		DefaultPoolName:         "test",
		AllocationTimeout:       30 * time.Second,
		HealthCheckInterval:     30 * time.Second,
		HealthCheckTimeout:      5 * time.Second,
		MaxAllocationsPerClient: 10,
		LoadBalancingStrategy:   "round_robin",
		WeightedThreshold:       0.8,
		MetricsEnabled:          true,
		StatsInterval:           60 * time.Second,
		LatencyHistorySize:      100,
		FailoverEnabled:         true,
		FailoverTimeout:         10 * time.Second,
		MaxRetries:              3,
		BackoffMultiplier:       2.0,
	}

	manager := NewTURNPoolManager(config, testLogger.GetZapLogger())
	defer manager.Stop()

	// Создаем пул
	err := manager.CreatePool("test-pool")
	require.NoError(t, err)

	// Создаем сервер
	server := &TURNServerInfo{
		TURNServerPool: &TURNServerPool{
			Address:        "turn1.example.com:3478",
			Username:       "testuser",
			Password:       "testpass",
			Realm:          "example.com",
			Priority:       100,
			Weight:         80,
			MaxConnections: 1000,
			Region:         "us-east",
			Provider:       "aws",
			Capabilities:   []string{"udp", "tcp"},
			Tags:           map[string]string{"env": "test"},
		},
		Region:        "us-east",
		Provider:      "aws",
		Capabilities:  []string{"udp", "tcp"},
		Tags:          map[string]string{"env": "test"},
		Status:        "active",
		StatusChanged: time.Now(),
	}

	// Добавляем сервер в пул
	err = manager.AddServerToPool("test-pool", server)
	require.NoError(t, err)

	// Создаем allocation
	preferences := &AllocationPreferences{
		PreferredRegion:   "us-east",
		PreferredProvider: "aws",
		MaxLatency:        100 * time.Millisecond,
		PoolName:          "test-pool",
	}

	allocation, err := manager.AllocateTURN("test-client", preferences)
	require.NoError(t, err)
	assert.NotNil(t, allocation)

	// Освобождаем allocation
	err = manager.ReleaseTURN(allocation.ID)
	require.NoError(t, err)

	// Проверяем, что allocation освобожден
	_, err = manager.GetAllocationInfo(allocation.ID)
	assert.Error(t, err)
}

func TestTURNPoolManager_GetAllocationInfo(t *testing.T) {
	testLogger := createTestLogger()
	config := &TURNPoolConfig{
		DefaultPoolName:         "test",
		AllocationTimeout:       30 * time.Second,
		HealthCheckInterval:     30 * time.Second,
		HealthCheckTimeout:      5 * time.Second,
		MaxAllocationsPerClient: 10,
		LoadBalancingStrategy:   "round_robin",
		WeightedThreshold:       0.8,
		MetricsEnabled:          true,
		StatsInterval:           60 * time.Second,
		LatencyHistorySize:      100,
		FailoverEnabled:         true,
		FailoverTimeout:         10 * time.Second,
		MaxRetries:              3,
		BackoffMultiplier:       2.0,
	}

	manager := NewTURNPoolManager(config, testLogger.GetZapLogger())
	defer manager.Stop()

	// Создаем пул
	err := manager.CreatePool("test-pool")
	require.NoError(t, err)

	// Создаем сервер
	server := &TURNServerInfo{
		TURNServerPool: &TURNServerPool{
			Address:        "turn1.example.com:3478",
			Username:       "testuser",
			Password:       "testpass",
			Realm:          "example.com",
			Priority:       100,
			Weight:         80,
			MaxConnections: 1000,
			Region:         "us-east",
			Provider:       "aws",
			Capabilities:   []string{"udp", "tcp"},
			Tags:           map[string]string{"env": "test"},
		},
		Region:        "us-east",
		Provider:      "aws",
		Capabilities:  []string{"udp", "tcp"},
		Tags:          map[string]string{"env": "test"},
		Status:        "active",
		StatusChanged: time.Now(),
	}

	// Добавляем сервер в пул
	err = manager.AddServerToPool("test-pool", server)
	require.NoError(t, err)

	// Создаем allocation
	preferences := &AllocationPreferences{
		PreferredRegion:   "us-east",
		PreferredProvider: "aws",
		MaxLatency:        100 * time.Millisecond,
		PoolName:          "test-pool",
	}

	allocation, err := manager.AllocateTURN("test-client", preferences)
	require.NoError(t, err)
	assert.NotNil(t, allocation)

	// Получаем информацию об allocation
	savedAllocation, err := manager.GetAllocationInfo(allocation.ID)
	require.NoError(t, err)
	assert.Equal(t, allocation.ID, savedAllocation.ID)
	assert.Equal(t, allocation.ClientID, savedAllocation.ClientID)
	assert.Equal(t, allocation.Status, savedAllocation.Status)
}

func TestTURNPoolManager_ListAllocations(t *testing.T) {
	testLogger := createTestLogger()
	config := &TURNPoolConfig{
		DefaultPoolName:         "test",
		AllocationTimeout:       30 * time.Second,
		HealthCheckInterval:     30 * time.Second,
		HealthCheckTimeout:      5 * time.Second,
		MaxAllocationsPerClient: 10,
		LoadBalancingStrategy:   "round_robin",
		WeightedThreshold:       0.8,
		MetricsEnabled:          true,
		StatsInterval:           60 * time.Second,
		LatencyHistorySize:      100,
		FailoverEnabled:         true,
		FailoverTimeout:         10 * time.Second,
		MaxRetries:              3,
		BackoffMultiplier:       2.0,
	}

	manager := NewTURNPoolManager(config, testLogger.GetZapLogger())
	defer manager.Stop()

	// Создаем пул
	err := manager.CreatePool("test-pool")
	require.NoError(t, err)

	// Создаем сервер
	server := &TURNServerInfo{
		TURNServerPool: &TURNServerPool{
			Address:        "turn1.example.com:3478",
			Username:       "testuser",
			Password:       "testpass",
			Realm:          "example.com",
			Priority:       100,
			Weight:         80,
			MaxConnections: 1000,
			Region:         "us-east",
			Provider:       "aws",
			Capabilities:   []string{"udp", "tcp"},
			Tags:           map[string]string{"env": "test"},
		},
		Region:        "us-east",
		Provider:      "aws",
		Capabilities:  []string{"udp", "tcp"},
		Tags:          map[string]string{"env": "test"},
		Status:        "active",
		StatusChanged: time.Now(),
	}

	// Добавляем сервер в пул
	err = manager.AddServerToPool("test-pool", server)
	require.NoError(t, err)

	// Создаем несколько allocations
	preferences := &AllocationPreferences{
		PreferredRegion:   "us-east",
		PreferredProvider: "aws",
		MaxLatency:        100 * time.Millisecond,
		PoolName:          "test-pool",
	}

	allocation1, err := manager.AllocateTURN("test-client-1", preferences)
	require.NoError(t, err)
	assert.NotNil(t, allocation1)

	allocation2, err := manager.AllocateTURN("test-client-2", preferences)
	require.NoError(t, err)
	assert.NotNil(t, allocation2)

	// Получаем список allocations
	allocations := manager.ListAllocations()
	assert.Len(t, allocations, 2)

	// Проверяем, что оба allocation в списке
	allocationIDs := make(map[string]bool)
	for _, allocation := range allocations {
		allocationIDs[allocation.ID] = true
	}
	assert.True(t, allocationIDs[allocation1.ID])
	assert.True(t, allocationIDs[allocation2.ID])
}

