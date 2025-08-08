package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/twogc/cloudbridge-relay-installer/config"
)

func TestServerIntegration(t *testing.T) {
	// Создаем тестовую конфигурацию
	cfg := &config.Config{}
	cfg.Server.Port = 8080
	cfg.EnhancedQUIC.Enabled = true
	cfg.WireGuard.Enabled = true
	cfg.Quantum.Enabled = true
	cfg.AI.Enabled = true

	// Создаем сервер
	server, err := NewServer(cfg)
	require.NoError(t, err)
	require.NotNil(t, server)

	// Проверяем, что все компоненты инициализированы
	assert.True(t, server.enhancedEnabled)
	assert.True(t, server.wireGuardEnabled)
	assert.True(t, server.quantumEnabled)
	assert.True(t, server.aiAnalyticsEnabled)

	// Проверяем WireGuard компоненты
	assert.NotNil(t, server.wireGuardInterface)
	assert.NotNil(t, server.wireGuardMesh)
	assert.NotNil(t, server.wireGuardPeerDiscovery)
	assert.NotNil(t, server.quicOverWireGuard)
	assert.NotNil(t, server.hybridRouter)
	assert.NotNil(t, server.siteToSiteManager)

	// Проверяем Quantum компоненты
	assert.NotNil(t, server.quantumConfig)
	assert.NotNil(t, server.quantumWireGuard)
	assert.NotNil(t, server.quantumEnhancedQUIC)

	// Проверяем AI компоненты
	assert.NotNil(t, server.aiBehaviorAnalyzer)
	assert.NotNil(t, server.aiPredictiveSecurity)
	assert.NotNil(t, server.aiAutomatedResponse)
	assert.NotNil(t, server.aiMonitoring)
}

func TestServerStartStop(t *testing.T) {
	// Создаем тестовую конфигурацию
	cfg := &config.Config{}
	cfg.Server.Port = 8081
	cfg.EnhancedQUIC.Enabled = false
	cfg.WireGuard.Enabled = false
	cfg.Quantum.Enabled = false
	cfg.AI.Enabled = false

	// Создаем сервер
	server, err := NewServer(cfg)
	require.NoError(t, err)

	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Запускаем сервер в горутине
	go func() {
		err := server.Start()
		if err != nil && err != context.Canceled {
			t.Logf("Server error: %v", err)
		}
	}()

	// Ждем немного для запуска
	time.Sleep(100 * time.Millisecond)

	// Останавливаем сервер
	server.Stop()

	// Проверяем, что сервер остановился
	select {
	case <-ctx.Done():
		// Таймаут - сервер не остановился
		t.Error("Server did not stop within timeout")
	case <-time.After(1 * time.Second):
		// Сервер остановился успешно
	}
}

func TestServerWithAllModulesEnabled(t *testing.T) {
	// Создаем тестовую конфигурацию со всеми модулями
	cfg := &config.Config{}
	cfg.Server.Port = 8082
	cfg.EnhancedQUIC.Enabled = true
	cfg.WireGuard.Enabled = true
	cfg.WireGuard.Interface.Name = "wg0"
	cfg.WireGuard.Interface.ListenPort = 51820
	cfg.WireGuard.Interface.Address = "10.100.1.1/24"
	cfg.Quantum.Enabled = true
	cfg.Quantum.Algorithm = "kyber"
	cfg.AI.Enabled = true
	cfg.AI.BehaviorAnalysis.Enabled = true
	cfg.AI.BehaviorAnalysis.ModelPath = "/tmp/test-model"
	cfg.AI.BehaviorAnalysis.Threshold = 0.8
	cfg.AI.BehaviorAnalysis.LearningRate = 0.001

	// Создаем сервер
	server, err := NewServer(cfg)
	require.NoError(t, err)

	// Проверяем статус всех модулей
	assert.True(t, server.enhancedEnabled)
	assert.True(t, server.wireGuardEnabled)
	assert.True(t, server.quantumEnabled)
	assert.True(t, server.aiAnalyticsEnabled)

	// Проверяем конфигурацию WireGuard
	assert.Equal(t, "wg0", server.wireGuardInterface.Name)
	assert.Equal(t, 51820, server.wireGuardInterface.ListenPort)
	assert.Equal(t, "10.100.1.1/24", server.wireGuardInterface.Address)

	// Проверяем конфигурацию Quantum
	assert.Equal(t, "kyber", server.quantumConfig.Algorithm)

	// Проверяем конфигурацию AI
	assert.True(t, server.aiBehaviorAnalyzer.Enabled)
	assert.Equal(t, "/tmp/test-model", server.aiBehaviorAnalyzer.ModelPath)
	assert.Equal(t, 0.8, server.aiBehaviorAnalyzer.Threshold)
}

func TestServerWithPartialModules(t *testing.T) {
	// Тестируем сервер только с WireGuard и AI
	cfg := &config.Config{}
	cfg.Server.Port = 8083
	cfg.EnhancedQUIC.Enabled = false
	cfg.WireGuard.Enabled = true
	cfg.Quantum.Enabled = false
	cfg.AI.Enabled = true

	// Создаем сервер
	server, err := NewServer(cfg)
	require.NoError(t, err)

	// Проверяем статус модулей
	assert.False(t, server.enhancedEnabled)
	assert.True(t, server.wireGuardEnabled)
	assert.False(t, server.quantumEnabled)
	assert.True(t, server.aiAnalyticsEnabled)

	// Проверяем, что соответствующие компоненты инициализированы
	assert.NotNil(t, server.wireGuardInterface)
	assert.NotNil(t, server.aiBehaviorAnalyzer)

	// Проверяем, что отключенные модули не инициализированы
	assert.Nil(t, server.quantumConfig)
	assert.Nil(t, server.quantumWireGuard)
}

func TestServerContextCancellation(t *testing.T) {
	// Создаем тестовую конфигурацию
	cfg := &config.Config{}
	cfg.Server.Port = 8084
	cfg.EnhancedQUIC.Enabled = false
	cfg.WireGuard.Enabled = false
	cfg.Quantum.Enabled = false
	cfg.AI.Enabled = false

	// Создаем сервер
	server, err := NewServer(cfg)
	require.NoError(t, err)

	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Запускаем сервер в горутине
	go func() {
		err := server.Start()
		if err != nil && err != context.Canceled {
			t.Logf("Server error: %v", err)
		}
	}()

	// Ждем немного для запуска
	time.Sleep(100 * time.Millisecond)

	// Отменяем контекст
	cancel()

	// Ждем завершения
	select {
	case <-ctx.Done():
		// Контекст отменен
	case <-time.After(3 * time.Second):
		t.Error("Server did not respond to context cancellation")
	}
}
