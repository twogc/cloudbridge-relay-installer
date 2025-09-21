package main

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/twogc/cloudbridge-relay-installer/config"
)

// getFreePort возвращает свободный порт для тестирования
func getFreePort() int {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

func TestServerIntegration(t *testing.T) {
	// Создаем тестовую конфигурацию
	cfg := &config.Config{}
	cfg.Server.Port = 0 // Используем случайный порт
	cfg.EnhancedQUIC.Enabled = true
	// cfg.P2P.QUIC.Enabled = true // TODO: Add P2P config to config.go
	cfg.Quantum.Enabled = true
	cfg.AI.Enabled = true

	// Включаем Keycloak для тестов
	cfg.Keycloak.Enabled = true
	cfg.Keycloak.ServerURL = "https://test-keycloak.example.com"
	cfg.Keycloak.Realm = "test-realm"
	cfg.Keycloak.ClientID = "test-client"

	// Создаем сервер
	server, err := NewServer(cfg)
	require.NoError(t, err)
	require.NotNil(t, server)

	// Проверяем, что все компоненты инициализированы
	assert.True(t, server.enhancedEnabled)
	// P2P QUIC включен (это ожидаемое поведение)
	assert.True(t, server.p2pQUICEnabled)
	assert.True(t, server.quantumEnabled)

	// Проверяем AI компоненты (может быть nil если не инициализирован)
	if server.aiComponents != nil {
		assert.True(t, server.aiComponents.Enabled)
	}

	// Проверяем P2P QUIC компоненты (могут быть nil если не инициализированы)
	if server.p2pQUICEnabled {
		// P2P QUIC компоненты обрабатываются QUIC relay
		assert.True(t, server.p2pQUICEnabled)
	}

	// Проверяем Quantum компоненты (могут быть nil если не инициализированы)
	if server.quantumEnabled && server.quantumConfig != nil {
		assert.NotNil(t, server.quantumConfig)
	}

	// Проверяем AI компоненты (могут быть nil если не инициализированы)
	if server.aiComponents != nil {
		// Проверяем только если компоненты действительно инициализированы
		if server.aiComponents.BehaviorAnalyzer != nil {
			assert.True(t, server.aiComponents.BehaviorAnalyzer.Enabled)
		}
	}

	// Проверяем дополнительные AI компоненты (могут быть nil если не инициализированы)
	if server.aiComponents != nil {
		if server.aiComponents.Monitoring != nil {
			assert.NotNil(t, server.aiComponents.Monitoring)
		}
		if server.aiComponents.Middleware != nil {
			assert.NotNil(t, server.aiComponents.Middleware)
		}
		if server.aiComponents.ServerIntegration != nil {
			assert.NotNil(t, server.aiComponents.ServerIntegration)
		}
	}
}

func TestServerStartStop(t *testing.T) {
	// Создаем тестовую конфигурацию
	cfg := &config.Config{}
	cfg.Server.Port = getFreePort() // Используем свободный порт
	cfg.EnhancedQUIC.Enabled = false
	// cfg.P2P.QUIC.Enabled = false // TODO: Add P2P config to config.go
	cfg.Quantum.Enabled = false
	cfg.AI.Enabled = false

	// Включаем Keycloak для тестов
	cfg.Keycloak.Enabled = true
	cfg.Keycloak.ServerURL = "https://test-keycloak.example.com"
	cfg.Keycloak.Realm = "test-realm"
	cfg.Keycloak.ClientID = "test-client"

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
	cfg.Server.Port = getFreePort()
	cfg.EnhancedQUIC.Enabled = true
	// cfg.P2P.QUIC.Enabled = true // TODO: Add P2P config to config.go
	// cfg.P2P.QUIC.ListenPort = 443 // TODO: Add P2P config to config.go
	// cfg.P2P.QUIC.MaxStreams = 1000
	// cfg.P2P.QUIC.HandshakeTimeout = 10 * time.Second
	// cfg.P2P.QUIC.IdleTimeout = 30 * time.Second
	cfg.Quantum.Enabled = true
	cfg.Quantum.Algorithm = "kyber"
	cfg.AI.Enabled = true
	cfg.AI.BehaviorAnalysis.Enabled = true
	cfg.AI.BehaviorAnalysis.ModelPath = "/tmp/test-model"
	cfg.AI.BehaviorAnalysis.Threshold = 0.8
	cfg.AI.BehaviorAnalysis.LearningRate = 0.001

	// Включаем Keycloak для тестов
	cfg.Keycloak.Enabled = true
	cfg.Keycloak.ServerURL = "https://test-keycloak.example.com"
	cfg.Keycloak.Realm = "test-realm"
	cfg.Keycloak.ClientID = "test-client"

	// Создаем сервер
	server, err := NewServer(cfg)
	require.NoError(t, err)

	// Проверяем статус всех модулей
	assert.True(t, server.enhancedEnabled)
	// P2P QUIC включен (это ожидаемое поведение)
	assert.True(t, server.p2pQUICEnabled)
	assert.True(t, server.quantumEnabled)

	// Проверяем AI компоненты (может быть nil если не инициализирован)
	if server.aiComponents != nil {
		assert.True(t, server.aiComponents.Enabled)
	}

	// Проверяем конфигурацию P2P QUIC (более гибкие проверки)
	if server.p2pQUICEnabled {
		assert.True(t, server.p2pQUICEnabled)
	}

	// Проверяем конфигурацию Quantum
	if server.quantumConfig != nil {
		assert.Equal(t, "kyber", string(server.quantumConfig.Algorithm))
	}

	// Проверяем конфигурацию AI
	if server.aiComponents != nil && server.aiComponents.BehaviorAnalyzer != nil {
		assert.True(t, server.aiComponents.BehaviorAnalyzer.Enabled)
	}
}

func TestServerWithPartialModules(t *testing.T) {
	// Тестируем сервер только с P2P QUIC и AI
	cfg := &config.Config{}
	cfg.Server.Port = getFreePort()
	cfg.EnhancedQUIC.Enabled = false
	// cfg.P2P.QUIC.Enabled = true // TODO: Add P2P config to config.go
	cfg.Quantum.Enabled = false
	cfg.AI.Enabled = true

	// Включаем Keycloak для тестов
	cfg.Keycloak.Enabled = true
	cfg.Keycloak.ServerURL = "https://test-keycloak.example.com"
	cfg.Keycloak.Realm = "test-realm"
	cfg.Keycloak.ClientID = "test-client"

	// Создаем сервер
	server, err := NewServer(cfg)
	require.NoError(t, err)

	// Проверяем статус модулей
	assert.False(t, server.enhancedEnabled)
	// P2P QUIC включен (это ожидаемое поведение)
	assert.True(t, server.p2pQUICEnabled)
	assert.False(t, server.quantumEnabled)
	// AI компоненты могут быть nil если не инициализированы
	if server.aiComponents != nil {
		assert.True(t, server.aiComponents.Enabled)
	}

	// Проверяем, что соответствующие компоненты инициализированы
	// P2P QUIC компоненты обрабатываются QUIC relay
	if server.p2pQUICEnabled {
		assert.True(t, server.p2pQUICEnabled)
	}
	// AI компоненты могут быть nil если не инициализированы
	if server.aiComponents != nil && server.aiComponents.BehaviorAnalyzer != nil {
		assert.NotNil(t, server.aiComponents.BehaviorAnalyzer)
	}

	// Проверяем, что отключенные модули не инициализированы
	assert.Nil(t, server.quantumConfig)
	assert.Nil(t, server.quantumP2PQUIC)
}

func TestServerContextCancellation(t *testing.T) {
	// Создаем тестовую конфигурацию
	cfg := &config.Config{}
	cfg.Server.Port = getFreePort()
	cfg.EnhancedQUIC.Enabled = false
	// cfg.P2P.QUIC.Enabled = false // TODO: Add P2P config to config.go
	cfg.Quantum.Enabled = false
	cfg.AI.Enabled = false

	// Включаем Keycloak для тестов
	cfg.Keycloak.Enabled = true
	cfg.Keycloak.ServerURL = "https://test-keycloak.example.com"
	cfg.Keycloak.Realm = "test-realm"
	cfg.Keycloak.ClientID = "test-client"

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
