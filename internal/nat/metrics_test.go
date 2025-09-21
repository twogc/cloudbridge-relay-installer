package nat

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

func TestNATMetrics_RecordClientNATBehavior(t *testing.T) {
	testLogger, err := logger.New(logger.Config{
		Level:  logger.DEBUG,
		JSON:   true,
		Output: "stdout",
	})
	require.NoError(t, err)

	nm := NewNATMetrics(testLogger)
	defer nm.Stop()

	clientID := "test-client-1"
	publicIP := "192.168.1.100"

	// Записываем NAT поведение
	nm.RecordClientNATBehavior(clientID, publicIP, NATBehaviorFullCone)

	// Проверяем, что метрики записались
	metrics, exists := nm.GetClientMetrics(clientID)
	require.True(t, exists)
	assert.Equal(t, clientID, metrics.ClientID)
	assert.Equal(t, publicIP, metrics.PublicIP)
	assert.Equal(t, NATBehaviorFullCone, metrics.NATBehavior)
	assert.Equal(t, uint64(1), nm.totalClients)
}

func TestNATMetrics_RecordSTUNRequest(t *testing.T) {
	testLogger, err := logger.New(logger.Config{
		Level:  logger.DEBUG,
		JSON:   true,
		Output: "stdout",
	})
	require.NoError(t, err)

	nm := NewNATMetrics(testLogger)
	defer nm.Stop()

	clientID := "test-client-1"
	serverAddress := "stun.l.google.com:19302"

	// Сначала создаем клиента
	nm.RecordClientNATBehavior(clientID, "192.168.1.100", NATBehaviorUnknown)

	// Записываем STUN запрос
	nm.RecordSTUNRequest(clientID, serverAddress, true, 50*time.Millisecond)

	// Проверяем метрики клиента
	clientMetrics, exists := nm.GetClientMetrics(clientID)
	require.True(t, exists)
	assert.Equal(t, uint64(1), clientMetrics.STUNRequests)
	assert.Equal(t, 50*time.Millisecond, clientMetrics.LastLatency)
	assert.Equal(t, 50*time.Millisecond, clientMetrics.AverageLatency)

	// Проверяем метрики сервера
	serverMetrics, exists := nm.GetSTUNServerMetrics(serverAddress)
	require.True(t, exists)
	assert.Equal(t, uint64(1), serverMetrics.TotalRequests)
	assert.Equal(t, uint64(1), serverMetrics.SuccessfulRequests)
	assert.Equal(t, uint64(0), serverMetrics.FailedRequests)
}

func TestNATMetrics_RecordTURNRequest(t *testing.T) {
	testLogger, err := logger.New(logger.Config{
		Level:  logger.DEBUG,
		JSON:   true,
		Output: "stdout",
	})
	require.NoError(t, err)

	nm := NewNATMetrics(testLogger)
	defer nm.Stop()

	clientID := "test-client-1"
	serverAddress := "turn.example.com:3478"

	// Сначала создаем клиента
	nm.RecordClientNATBehavior(clientID, "192.168.1.100", NATBehaviorUnknown)

	// Записываем TURN запрос
	nm.RecordTURNRequest(clientID, serverAddress, true, 100*time.Millisecond)

	// Проверяем метрики клиента
	clientMetrics, exists := nm.GetClientMetrics(clientID)
	require.True(t, exists)
	assert.Equal(t, uint64(1), clientMetrics.TURNRequests)
	assert.Equal(t, 100*time.Millisecond, clientMetrics.LastLatency)

	// Проверяем метрики сервера
	serverMetrics, exists := nm.GetTURNServerMetrics(serverAddress)
	require.True(t, exists)
	assert.Equal(t, uint64(1), serverMetrics.TotalRequests)
	assert.Equal(t, uint64(1), serverMetrics.SuccessfulRequests)
}

func TestNATMetrics_RecordTURNAllocation(t *testing.T) {
	testLogger, err := logger.New(logger.Config{
		Level:  logger.DEBUG,
		JSON:   true,
		Output: "stdout",
	})
	require.NoError(t, err)

	nm := NewNATMetrics(testLogger)
	defer nm.Stop()

	clientID := "test-client-1"
	serverAddress := "turn.example.com:3478"

	// Сначала создаем клиента
	nm.RecordClientNATBehavior(clientID, "192.168.1.100", NATBehaviorUnknown)

	// Создаем TURN сервер через запрос
	nm.RecordTURNRequest(clientID, serverAddress, true, 50*time.Millisecond)

	// Записываем TURN allocation
	nm.RecordTURNAllocation(clientID, serverAddress, 1024, 2048)

	// Проверяем метрики клиента
	clientMetrics, exists := nm.GetClientMetrics(clientID)
	require.True(t, exists)
	assert.Equal(t, uint64(1), clientMetrics.Allocations)

	// Проверяем метрики сервера
	serverMetrics, exists := nm.GetTURNServerMetrics(serverAddress)
	require.True(t, exists)
	assert.Equal(t, uint64(1), serverMetrics.ActiveAllocations)
	assert.Equal(t, uint64(1), serverMetrics.TotalAllocations)
	assert.Equal(t, uint64(1024), serverMetrics.BandwidthIn)
	assert.Equal(t, uint64(2048), serverMetrics.BandwidthOut)
}

func TestNATMetrics_RecordConnectivityCheck(t *testing.T) {
	testLogger, err := logger.New(logger.Config{
		Level:  logger.DEBUG,
		JSON:   true,
		Output: "stdout",
	})
	require.NoError(t, err)

	nm := NewNATMetrics(testLogger)
	defer nm.Stop()

	clientID := "test-client-1"

	// Сначала создаем клиента
	nm.RecordClientNATBehavior(clientID, "192.168.1.100", NATBehaviorUnknown)

	// Записываем connectivity check
	nm.RecordConnectivityCheck(clientID, true)
	nm.RecordConnectivityCheck(clientID, false)

	// Проверяем метрики клиента
	clientMetrics, exists := nm.GetClientMetrics(clientID)
	require.True(t, exists)
	assert.Equal(t, uint64(2), clientMetrics.ConnectivityChecks)
	assert.Equal(t, uint64(1), clientMetrics.SuccessfulChecks)
}

func TestNATMetrics_GetAllMetrics(t *testing.T) {
	testLogger, err := logger.New(logger.Config{
		Level:  logger.DEBUG,
		JSON:   true,
		Output: "stdout",
	})
	require.NoError(t, err)

	nm := NewNATMetrics(testLogger)
	defer nm.Stop()

	// Добавляем тестовые данные
	nm.RecordClientNATBehavior("client1", "192.168.1.100", NATBehaviorFullCone)
	nm.RecordClientNATBehavior("client2", "192.168.1.101", NATBehaviorSymmetric)
	nm.RecordSTUNRequest("client1", "stun1.example.com:19302", true, 50*time.Millisecond)
	nm.RecordTURNRequest("client1", "turn1.example.com:3478", true, 100*time.Millisecond)
	nm.RecordTURNAllocation("client1", "turn1.example.com:3478", 1024, 2048)

	// Получаем все метрики
	allMetrics := nm.GetAllMetrics()

	assert.Equal(t, uint64(2), allMetrics["total_clients"])
	assert.Equal(t, uint64(1), allMetrics["total_stun_requests"])
	assert.Equal(t, uint64(1), allMetrics["total_turn_requests"])
	assert.Equal(t, uint64(1), allMetrics["total_allocations"])
	assert.Equal(t, 2, allMetrics["active_clients"])
	assert.Equal(t, 1, allMetrics["active_stun_servers"])
	assert.Equal(t, 1, allMetrics["active_turn_servers"])
}

func TestNATMetrics_GetNATBehaviorStats(t *testing.T) {
	testLogger, err := logger.New(logger.Config{
		Level:  logger.DEBUG,
		JSON:   true,
		Output: "stdout",
	})
	require.NoError(t, err)

	nm := NewNATMetrics(testLogger)
	defer nm.Stop()

	// Добавляем клиентов с разными типами NAT
	nm.RecordClientNATBehavior("client1", "192.168.1.100", NATBehaviorFullCone)
	nm.RecordClientNATBehavior("client2", "192.168.1.101", NATBehaviorFullCone)
	nm.RecordClientNATBehavior("client3", "192.168.1.102", NATBehaviorSymmetric)
	nm.RecordClientNATBehavior("client4", "192.168.1.103", NATBehaviorUnknown)

	// Получаем статистику по типам NAT
	stats := nm.GetNATBehaviorStats()

	assert.Equal(t, uint64(2), stats[NATBehaviorFullCone])
	assert.Equal(t, uint64(1), stats[NATBehaviorSymmetric])
	assert.Equal(t, uint64(1), stats[NATBehaviorUnknown])
	assert.Equal(t, uint64(0), stats[NATBehaviorRestrictedCone])
}

func TestNATMetrics_CalculateAverageLatency(t *testing.T) {
	testLogger, err := logger.New(logger.Config{
		Level:  logger.DEBUG,
		JSON:   true,
		Output: "stdout",
	})
	require.NoError(t, err)

	nm := NewNATMetrics(testLogger)
	defer nm.Stop()

	// Тестируем вычисление средней задержки
	avg1 := nm.calculateAverageLatency(0, 100*time.Millisecond, 1)
	assert.Equal(t, 100*time.Millisecond, avg1)

	avg2 := nm.calculateAverageLatency(100*time.Millisecond, 200*time.Millisecond, 2)
	assert.Equal(t, 150*time.Millisecond, avg2)

	avg3 := nm.calculateAverageLatency(150*time.Millisecond, 300*time.Millisecond, 3)
	assert.Equal(t, 200*time.Millisecond, avg3)
}

func TestNATMetrics_CleanupExpiredMetrics(t *testing.T) {
	testLogger, err := logger.New(logger.Config{
		Level:  logger.DEBUG,
		JSON:   true,
		Output: "stdout",
	})
	require.NoError(t, err)

	nm := NewNATMetrics(testLogger)
	defer nm.Stop()

	// Добавляем клиента
	nm.RecordClientNATBehavior("client1", "192.168.1.100", NATBehaviorFullCone)

	// Проверяем, что клиент добавлен
	_, exists := nm.GetClientMetrics("client1")
	assert.True(t, exists)

	// Ждем немного и принудительно очищаем
	time.Sleep(100 * time.Millisecond)
	nm.cleanupExpiredClientMetrics()

	// Клиент должен остаться (не истек)
	_, exists = nm.GetClientMetrics("client1")
	assert.True(t, exists)
}