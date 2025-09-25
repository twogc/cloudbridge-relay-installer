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

func TestTURNPool_AddServer(t *testing.T) {
	// Создаем logger для тестов
	testLogger := createTestLogger()
	pool := NewTURNPool(testLogger)
	defer pool.Stop()

	server := &TURNServerPool{
		Address:        "turn1.example.com:3478",
		Username:       "testuser",
		Password:       "testpass",
		Realm:          "example.com",
		Priority:       100,
		Weight:         80,
		MaxConnections: 1000,
		IsHealthy:      true,
	}

	// Добавляем сервер
	pool.AddServer(server)

	// Проверяем, что сервер добавлен
	servers := pool.GetServerList()
	require.Len(t, servers, 1)
	assert.Equal(t, server.Address, servers[0].Address)
	assert.Equal(t, server.Priority, servers[0].Priority)
}

func TestTURNPool_RemoveServer(t *testing.T) {
	testLogger := createTestLogger()
	pool := NewTURNPool(testLogger)
	defer pool.Stop()

	server := &TURNServerPool{
		Address:        "turn1.example.com:3478",
		Username:       "testuser",
		Password:       "testpass",
		Realm:          "example.com",
		Priority:       100,
		Weight:         80,
		MaxConnections: 1000,
		IsHealthy:      true,
	}

	// Добавляем и удаляем сервер
	pool.AddServer(server)
	pool.RemoveServer(server.Address)

	// Проверяем, что сервер удален
	servers := pool.GetServerList()
	assert.Len(t, servers, 0)
}

func TestTURNPool_GetBestServer(t *testing.T) {
	testLogger := createTestLogger()
	pool := NewTURNPool(testLogger)
	defer pool.Stop()

	// Добавляем серверы с разными приоритетами
	server1 := &TURNServerPool{
		Address:        "turn1.example.com:3478",
		Priority:       100,
		Weight:         80,
		MaxConnections: 1000,
		IsHealthy:      true,
	}
	server2 := &TURNServerPool{
		Address:        "turn2.example.com:3478",
		Priority:       90,
		Weight:         60,
		MaxConnections: 500,
		IsHealthy:      true,
	}
	server3 := &TURNServerPool{
		Address:        "turn3.example.com:3478",
		Priority:       100,
		Weight:         70,
		MaxConnections: 800,
		IsHealthy:      false, // Нездоровый сервер
	}

	pool.AddServer(server1)
	pool.AddServer(server2)
	pool.AddServer(server3)

	// Получаем лучший сервер - должен быть server1 (наивысший приоритет и здоровый)
	bestServer, err := pool.GetBestServer()
	require.NoError(t, err)
	assert.Equal(t, server1.Address, bestServer.Address)
}

func TestTURNPool_GetBestServer_NoHealthyServers(t *testing.T) {
	testLogger := createTestLogger()
	pool := NewTURNPool(testLogger)
	defer pool.Stop()

	// Добавляем только нездоровые серверы
	server1 := &TURNServerPool{
		Address:        "turn1.example.com:3478",
		Priority:       100,
		Weight:         80,
		MaxConnections: 1000,
		IsHealthy:      false,
	}
	server2 := &TURNServerPool{
		Address:        "turn2.example.com:3478",
		Priority:       90,
		Weight:         60,
		MaxConnections: 500,
		IsHealthy:      false,
	}

	pool.AddServer(server1)
	pool.AddServer(server2)

	// Должна быть ошибка - нет здоровых серверов
	_, err := pool.GetBestServer()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no healthy TURN servers available")
}

func TestTURNPool_AllocateTURN(t *testing.T) {
	testLogger := createTestLogger()
	pool := NewTURNPool(testLogger)
	defer pool.Stop()

	server := &TURNServerPool{
		Address:        "turn1.example.com:3478",
		Username:       "testuser",
		Password:       "testpass",
		Realm:          "example.com",
		Priority:       100,
		Weight:         80,
		MaxConnections: 1000,
		IsHealthy:      true,
	}

	pool.AddServer(server)

	// Создаем allocation
	allocation, err := pool.AllocateTURN("test-client")
	require.NoError(t, err)
	assert.NotEmpty(t, allocation.AllocationID)
	assert.Equal(t, "test-client", allocation.ClientID)
	assert.Equal(t, server.Address, allocation.Server.Address)

	// Проверяем, что счетчик активных соединений увеличился
	assert.Equal(t, 1, server.ActiveConnections)
}

func TestTURNPool_ReleaseTURN(t *testing.T) {
	testLogger := createTestLogger()
	pool := NewTURNPool(testLogger)
	defer pool.Stop()

	server := &TURNServerPool{
		Address:        "turn1.example.com:3478",
		Username:       "testuser",
		Password:       "testpass",
		Realm:          "example.com",
		Priority:       100,
		Weight:         80,
		MaxConnections: 1000,
		IsHealthy:      true,
	}

	pool.AddServer(server)

	// Создаем и освобождаем allocation
	allocation, err := pool.AllocateTURN("test-client")
	require.NoError(t, err)

	pool.ReleaseTURN(allocation)

	// Проверяем, что счетчик активных соединений уменьшился
	assert.Equal(t, 0, server.ActiveConnections)
}

func TestTURNPool_WeightedSelection(t *testing.T) {
	testLogger := createTestLogger()
	pool := NewTURNPool(testLogger)
	defer pool.Stop()

	// Добавляем серверы с одинаковым приоритетом, но разными весами
	server1 := &TURNServerPool{
		Address:        "turn1.example.com:3478",
		Priority:       100,
		Weight:         80, // Больший вес
		MaxConnections: 1000,
		IsHealthy:      true,
	}
	server2 := &TURNServerPool{
		Address:        "turn2.example.com:3478",
		Priority:       100,
		Weight:         20, // Меньший вес
		MaxConnections: 1000,
		IsHealthy:      true,
	}

	pool.AddServer(server1)
	pool.AddServer(server2)

	// Делаем несколько запросов и проверяем распределение
	server1Count := 0
	server2Count := 0

	for i := 0; i < 100; i++ {
		bestServer, err := pool.GetBestServer()
		require.NoError(t, err)

		if bestServer.Address == server1.Address {
			server1Count++
		} else if bestServer.Address == server2.Address {
			server2Count++
		}
	}

	// server1 должен получать больше запросов из-за большего веса
	assert.Greater(t, server1Count, server2Count)
	t.Logf("Server1 count: %d, Server2 count: %d", server1Count, server2Count)
}

func TestTURNPool_MaxConnections(t *testing.T) {
	testLogger := createTestLogger()
	pool := NewTURNPool(testLogger)
	defer pool.Stop()

	server := &TURNServerPool{
		Address:        "turn1.example.com:3478",
		Priority:       100,
		Weight:         80,
		MaxConnections: 2, // Ограничиваем до 2 соединений
		IsHealthy:      true,
	}

	pool.AddServer(server)

	// Создаем максимальное количество allocations
	allocation1, err := pool.AllocateTURN("client1")
	require.NoError(t, err)
	allocation2, err := pool.AllocateTURN("client2")
	require.NoError(t, err)

	// Следующая allocation должна завершиться ошибкой
	_, err = pool.AllocateTURN("client3")
	assert.Error(t, err)

	// Освобождаем одну allocation
	pool.ReleaseTURN(allocation1)

	// Теперь должна быть возможность создать новую allocation
	_, err = pool.AllocateTURN("client3")
	require.NoError(t, err)

	// Очистка
	pool.ReleaseTURN(allocation2)
}

func TestTURNPool_GetServerStats(t *testing.T) {
	testLogger := createTestLogger()
	pool := NewTURNPool(testLogger)
	defer pool.Stop()

	server1 := &TURNServerPool{
		Address:        "turn1.example.com:3478",
		Priority:       100,
		Weight:         80,
		MaxConnections: 1000,
		IsHealthy:      true,
	}
	server2 := &TURNServerPool{
		Address:        "turn2.example.com:3478",
		Priority:       90,
		Weight:         60,
		MaxConnections: 500,
		IsHealthy:      false,
	}

	pool.AddServer(server1)
	pool.AddServer(server2)

	stats := pool.GetServerStats()

	assert.Equal(t, 2, stats["total_servers"])
	assert.Equal(t, 1, stats["healthy_servers"])
	assert.Equal(t, 1, stats["unhealthy_servers"])
	assert.Equal(t, 0, stats["total_connections"])
	assert.Equal(t, 1500, stats["max_connections"])
}

func TestTURNPoolManager_CreatePool(t *testing.T) {
	config := &TURNPoolConfig{
		DefaultPoolName:         "test",
		AllocationTimeout:       time.Minute,
		HealthCheckInterval:     30 * time.Second,
		HealthCheckTimeout:      5 * time.Second,
		MaxAllocationsPerClient: 5,
		LoadBalancingStrategy:   "weighted",
		MetricsEnabled:          false, // Отключаем метрики для тестов
	}

	manager := NewTURNPoolManager(config, createTestLogger().GetZapLogger())
	defer manager.Stop()

	// Создаем пул
	err := manager.CreatePool("test-pool")
	require.NoError(t, err)

	// Попытка создать пул с тем же именем должна завершиться ошибкой
	err = manager.CreatePool("test-pool")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestTURNPoolManager_AddServerToPool(t *testing.T) {
	config := &TURNPoolConfig{
		DefaultPoolName:         "test",
		AllocationTimeout:       time.Minute,
		HealthCheckInterval:     30 * time.Second,
		HealthCheckTimeout:      5 * time.Second,
		MaxAllocationsPerClient: 5,
		LoadBalancingStrategy:   "weighted",
		MetricsEnabled:          false,
	}

	manager := NewTURNPoolManager(config, createTestLogger().GetZapLogger())
	defer manager.Stop()

	// Создаем пул
	err := manager.CreatePool("test-pool")
	require.NoError(t, err)

	// Добавляем сервер в пул
	server := &TURNServerInfo{
		TURNServerPool: &TURNServerPool{
			Address:        "turn1.example.com:3478",
			Username:       "testuser",
			Password:       "testpass",
			Realm:          "example.com",
			Priority:       100,
			Weight:         80,
			MaxConnections: 1000,
		},
		Region:   "us-east1",
		Provider: "aws",
	}

	err = manager.AddServerToPool("test-pool", server)
	require.NoError(t, err)

	// Попытка добавить сервер в несуществующий пул
	err = manager.AddServerToPool("nonexistent-pool", server)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestTURNPoolManager_AllocateTURN(t *testing.T) {
	config := &TURNPoolConfig{
		DefaultPoolName:         "default",
		AllocationTimeout:       time.Minute,
		HealthCheckInterval:     30 * time.Second,
		HealthCheckTimeout:      5 * time.Second,
		MaxAllocationsPerClient: 5,
		LoadBalancingStrategy:   "weighted",
		MetricsEnabled:          false,
	}

	manager := NewTURNPoolManager(config, createTestLogger().GetZapLogger())
	defer manager.Stop()

	// Создаем пул по умолчанию
	err := manager.CreatePool("default")
	require.NoError(t, err)

	// Добавляем сервер
	server := &TURNServerInfo{
		TURNServerPool: &TURNServerPool{
			Address:        "turn1.example.com:3478",
			Username:       "testuser",
			Password:       "testpass",
			Realm:          "example.com",
			Priority:       100,
			Weight:         80,
			MaxConnections: 1000,
		},
		Region:   "us-east1",
		Provider: "aws",
	}

	err = manager.AddServerToPool("default", server)
	require.NoError(t, err)

	// Создаем allocation
	allocation, err := manager.AllocateTURN("test-client", nil)
	require.NoError(t, err)
	assert.NotEmpty(t, allocation.AllocationID)
	assert.Equal(t, "test-client", allocation.ClientID)

	// Проверяем, что allocation сохранена в manager
	savedAllocation, err := manager.GetAllocationInfo(allocation.AllocationID)
	require.NoError(t, err)
	assert.Equal(t, allocation.AllocationID, savedAllocation.AllocationID)
}

func TestTURNPoolManager_AllocationLimit(t *testing.T) {
	config := &TURNPoolConfig{
		DefaultPoolName:         "default",
		AllocationTimeout:       time.Minute,
		HealthCheckInterval:     30 * time.Second,
		HealthCheckTimeout:      5 * time.Second,
		MaxAllocationsPerClient: 2, // Ограничиваем до 2 allocations на клиента
		LoadBalancingStrategy:   "weighted",
		MetricsEnabled:          false,
	}

	manager := NewTURNPoolManager(config, createTestLogger().GetZapLogger())
	defer manager.Stop()

	// Создаем пул и сервер
	err := manager.CreatePool("default")
	require.NoError(t, err)

	server := &TURNServerInfo{
		TURNServerPool: &TURNServerPool{
			Address:        "turn1.example.com:3478",
			Username:       "testuser",
			Password:       "testpass",
			Realm:          "example.com",
			Priority:       100,
			Weight:         80,
			MaxConnections: 1000,
		},
		Region:   "us-east1",
		Provider: "aws",
	}

	err = manager.AddServerToPool("default", server)
	require.NoError(t, err)

	// Создаем максимальное количество allocations
	allocation1, err := manager.AllocateTURN("test-client", nil)
	require.NoError(t, err)
	allocation2, err := manager.AllocateTURN("test-client", nil)
	require.NoError(t, err)

	// Следующая allocation должна завершиться ошибкой
	_, err = manager.AllocateTURN("test-client", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "allocation limit exceeded")

	// Освобождаем allocation
	err = manager.ReleaseTURN(allocation1.AllocationID)
	require.NoError(t, err)

	// Теперь должна быть возможность создать новую allocation
	_, err = manager.AllocateTURN("test-client", nil)
	require.NoError(t, err)

	// Очистка
	err = manager.ReleaseTURN(allocation2.AllocationID)
	require.NoError(t, err)
}

func TestTURNPoolManager_GetPoolStats(t *testing.T) {
	config := &TURNPoolConfig{
		DefaultPoolName:         "default",
		AllocationTimeout:       time.Minute,
		HealthCheckInterval:     30 * time.Second,
		HealthCheckTimeout:      5 * time.Second,
		MaxAllocationsPerClient: 5,
		LoadBalancingStrategy:   "weighted",
		MetricsEnabled:          false,
	}

	manager := NewTURNPoolManager(config, createTestLogger().GetZapLogger())
	defer manager.Stop()

	// Создаем пул
	err := manager.CreatePool("test-pool")
	require.NoError(t, err)

	// Добавляем сервер
	server := &TURNServerInfo{
		TURNServerPool: &TURNServerPool{
			Address:        "turn1.example.com:3478",
			Username:       "testuser",
			Password:       "testpass",
			Realm:          "example.com",
			Priority:       100,
			Weight:         80,
			MaxConnections: 1000,
		},
		Region:   "us-east1",
		Provider: "aws",
	}

	err = manager.AddServerToPool("test-pool", server)
	require.NoError(t, err)

	// Получаем статистику
	stats := manager.GetPoolStats()

	assert.Equal(t, 1, stats["total_pools"])
	assert.Equal(t, 0, stats["total_allocations"])
	assert.Equal(t, 1, stats["total_servers"])
	assert.Equal(t, 1, stats["healthy_servers"])

	// Проверяем статистику конкретного пула
	poolStats := stats["pools"].(map[string]interface{})["test-pool"].(map[string]interface{})
	assert.Equal(t, 1, poolStats["total_servers"])
	assert.Equal(t, 1, poolStats["healthy_servers"])
}
