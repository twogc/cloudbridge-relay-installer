package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/twogc/cloudbridge-relay-installer/config"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	"github.com/twogc/cloudbridge-relay-installer/internal/nat"
)

// TestMASQUEConnectionBasic тестирует базовую функциональность MASQUE соединений
func TestMASQUEConnectionBasic(t *testing.T) {
	// Создаем тестовое соединение с datagram channel
	conn := &MASQUEConnection{
		ID:              "test_conn",
		DatagramChannel: make(chan []byte, 10),
		CreatedAt:       time.Now(),
		LastActivity:    time.Now(),
	}

	// Тест отправки и получения datagrams
	testData := []byte("test datagram")
	err := conn.SendDatagram(testData)
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	receivedData, err := conn.ReceiveDatagram(ctx)
	assert.NoError(t, err)
	assert.Equal(t, testData, receivedData)
}

// TestNATMetricsBasic тестирует базовую функциональность NAT метрик
func TestNATMetricsBasic(t *testing.T) {
	// Создаем logger для NAT компонентов
	natLogger, err := logger.New(logger.Config{
		Level:  logger.DEBUG,
		JSON:   true,
		Output: "stdout",
	})
	assert.NoError(t, err)

	stunServers := []string{"stun1.example.com:3478"}

	// Создаем ICE агент
	agent := nat.NewICEAgent(natLogger, stunServers, nil)

	// Записываем несколько попыток соединения через метрики
	agent.Metrics.UpdateConnectivityMetrics(true, 100*time.Millisecond)
	agent.Metrics.UpdateConnectivityMetrics(false, 200*time.Millisecond)
	agent.Metrics.UpdateConnectivityMetrics(true, 150*time.Millisecond)

	// Проверяем метрики
	successRate := agent.Metrics.GetSuccessRate()
	assert.Greater(t, successRate, 0.0)

	// Проверяем общее количество кандидатов
	totalCandidates := agent.Metrics.GetTotalCandidates()
	assert.GreaterOrEqual(t, totalCandidates, 0)
}

// TestFallbackManagerBasic тестирует базовую функциональность FallbackManager
func TestFallbackManagerBasic(t *testing.T) {
	cfg := &config.Config{}

	// Создаем FallbackManager (без логгера для упрощения)
	manager := &FallbackManager{
		config: cfg,
	}

	// Проверяем детекцию блокирующих сетей
	isBlocking := manager.DetectBlockingNetwork("192.168.1.1:8080")
	assert.True(t, isBlocking)

	isBlocking = manager.DetectBlockingNetwork("8.8.8.8:8080")
	assert.False(t, isBlocking)
}

// TestWebSocketFallbackBasic тестирует базовую функциональность WebSocket fallback
func TestWebSocketFallbackBasic(t *testing.T) {
	cfg := &config.Config{}

	// Создаем WebSocket fallback (без логгера для упрощения)
	fallback := &WebSocketFallback{
		config:      cfg,
		connections: make(map[string]*WebSocketConnection),
	}

	// Создаем тестовое соединение
	conn := &WebSocketConnection{
		ID:           "test_conn",
		RemoteAddr:   "127.0.0.1:8080",
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
		IsActive:     true,
	}

	// Добавляем соединение
	fallback.mu.Lock()
	fallback.connections[conn.ID] = conn
	fallback.mu.Unlock()

	// Проверяем статистику
	count := fallback.GetConnectionCount()
	assert.Equal(t, 1, count)

	info := fallback.GetConnectionInfo()
	assert.Equal(t, 1, info["total_connections"])
}

// TestTCPFallbackBasic тестирует базовую функциональность TCP fallback
func TestTCPFallbackBasic(t *testing.T) {
	cfg := &config.Config{}

	// Создаем TCP fallback (без логгера для упрощения)
	fallback := &TCPFallback{
		config:      cfg,
		connections: make(map[string]*TCPConnection),
	}

	// Создаем тестовое соединение
	conn := &TCPConnection{
		ID:           "test_conn",
		RemoteAddr:   "127.0.0.1:8081",
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
		IsActive:     true,
	}

	// Добавляем соединение
	fallback.mu.Lock()
	fallback.connections[conn.ID] = conn
	fallback.mu.Unlock()

	// Проверяем статистику
	count := fallback.GetConnectionCount()
	assert.Equal(t, 1, count)

	info := fallback.GetConnectionInfo()
	assert.Equal(t, 1, info["total_connections"])
}

// TestFallbackStrategyBasic тестирует базовую функциональность стратегий fallback
func TestFallbackStrategyBasic(t *testing.T) {
	cfg := &config.Config{}

	manager := &FallbackManager{
		config: cfg,
	}

	// Тест определения стратегии
	strategy := manager.determineFallbackStrategy("127.0.0.1:8080")
	assert.Equal(t, FallbackStrategyQUICFirst, strategy)
}

// TestPerformanceBasic тестирует базовую производительность
func TestPerformanceBasic(t *testing.T) {
	cfg := &config.Config{}

	manager := &FallbackManager{
		config: cfg,
	}

	// Тест производительности с множественными вызовами
	numCalls := 1000
	start := time.Now()

	for i := 0; i < numCalls; i++ {
		manager.DetectBlockingNetwork("192.168.1.1:8080")
	}

	duration := time.Since(start)
	t.Logf("Processed %d calls in %v (%.2f calls/sec)",
		numCalls, duration, float64(numCalls)/duration.Seconds())

	// Проверяем, что все вызовы завершились быстро
	assert.Less(t, duration, 100*time.Millisecond)
}
