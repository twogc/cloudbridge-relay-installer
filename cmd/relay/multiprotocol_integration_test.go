package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twogc/cloudbridge-relay-installer/config"
	"github.com/twogc/cloudbridge-relay-installer/internal/protocol"
)

func TestMultiprotocolIntegration(t *testing.T) {
	// Создание тестовой конфигурации
	cfg := createTestConfig()
	// Keycloak больше не используется в конфигурации

	server, err := NewServer(cfg)
	require.NoError(t, err)
	require.NotNil(t, server)

	defer server.Stop()

	// Проверка, что мультипротокольные компоненты инициализированы
	assert.NotNil(t, server.protocolSelector, "Protocol selector should be initialized")
	assert.NotNil(t, server.protocolEngine, "Protocol engine should be initialized")
	assert.NotNil(t, server.p2pManager, "P2P manager should be initialized")
	assert.NotNil(t, server.p2cManager, "P2C manager should be initialized")
	assert.NotNil(t, server.connectionHandler, "Connection handler should be initialized")
	assert.NotNil(t, server.autoSwitchManager, "Auto switch manager should be initialized")

	t.Run("Protocol Selection", func(t *testing.T) {
		// Проверка выбора протокола
		selectedProtocol := server.protocolSelector.SelectBest()
		assert.NotEqual(t, protocol.Protocol(999), selectedProtocol, "Should select a valid protocol")

		t.Logf("Selected protocol: %s", selectedProtocol.String())
	})

	t.Run("P2P Availability Check", func(t *testing.T) {
		// Проверка доступности P2P
		available := server.p2pManager.IsAvailable(context.Background(), "127.0.0.1", "test-tenant")
		// P2P может быть недоступен в тестовой среде, это нормально
		t.Logf("P2P available: %v", available)
	})

	t.Run("Connection Handler Metrics", func(t *testing.T) {
		// Проверка метрик обработчика соединений
		metrics := server.connectionHandler.GetMetrics()
		assert.NotNil(t, metrics, "Connection handler metrics should be available")
	})

	t.Run("Auto Switch Manager", func(t *testing.T) {
		// Проверка метрик автоматического переключения
		switchMetrics := server.autoSwitchManager.GetMetrics()
		assert.NotNil(t, switchMetrics, "Auto switch metrics should be available")
	})
}

func TestMultiprotocolConfiguration(t *testing.T) {
	// Тест конфигурации мультипротокольности
	cfg := createTestConfig()

	// Проверка настроек протокола
	assert.True(t, cfg.Protocol.AutoSelect, "Auto select should be enabled")
	assert.Equal(t, "quic", cfg.Protocol.Preferred, "Preferred protocol should be QUIC")
	assert.Equal(t, "http2", cfg.Protocol.Fallback, "Fallback protocol should be HTTP2")
	assert.Equal(t, 0.8, cfg.Protocol.SwitchThreshold, "Switch threshold should be 0.8")
	assert.Equal(t, 30*time.Second, cfg.Protocol.HealthCheckInterval, "Health check interval should be 30s")

	// Проверка настроек Enhanced QUIC
	assert.True(t, cfg.EnhancedQUIC.Enabled, "Enhanced QUIC should be enabled")
	assert.True(t, cfg.EnhancedQUIC.DDoSProtection, "DDoS protection should be enabled")
	assert.True(t, cfg.EnhancedQUIC.GeographicRouting, "Geographic routing should be enabled")
}

func createTestConfig() *config.Config {
	// Создание тестовой конфигурации
	return &config.Config{
		Server: struct {
			Address         string        `json:"address"`
			Port            int           `json:"port"`
			Timeout         time.Duration `json:"timeout"`
			MaxTunnels      int           `json:"max_tunnels"`
			PingInterval    time.Duration `json:"ping_interval"`
			CleanupInterval time.Duration `json:"cleanup_interval"`
		}{
			Port:    8080,
			Address: "localhost",
		},
		Protocol: struct {
			AutoSelect          bool          `json:"auto_select"`
			Preferred           string        `json:"preferred"`
			Fallback            string        `json:"fallback"`
			SwitchThreshold     float64       `json:"switch_threshold"`
			HealthCheckInterval time.Duration `json:"health_check_interval"`
		}{
			AutoSelect:          true,
			Preferred:           "quic",
			Fallback:            "http2",
			SwitchThreshold:     0.8,
			HealthCheckInterval: 30 * time.Second,
		},
		EnhancedQUIC: struct {
			Enabled           bool          `json:"enabled"`
			MaxStreams        int           `json:"max_streams"`
			KeepAlive         bool          `json:"keep_alive"`
			IdleTimeout       time.Duration `json:"idle_timeout"`
			HandshakeTimeout  time.Duration `json:"handshake_timeout"`
			DDoSProtection    bool          `json:"ddos_protection"`
			GeographicRouting bool          `json:"geographic_routing"`
		}{
			Enabled:           true,
			MaxStreams:        100,
			KeepAlive:         true,
			IdleTimeout:       30 * time.Second,
			HandshakeTimeout:  10 * time.Second,
			DDoSProtection:    true,
			GeographicRouting: true,
		},
		// Добавляем минимальные настройки для других компонентов
		ZeroTrust: struct {
			Enabled          bool `json:"enabled"`
			ClientDaemon     bool `json:"client_daemon"`
			ServiceDiscovery bool `json:"service_discovery"`
			AutoReconnect    bool `json:"auto_reconnect"`

			CircuitBreaker struct {
				Enabled          bool
				FailureThreshold int64
				SuccessThreshold int64
				Timeout          time.Duration
				WindowSize       time.Duration
			}

			HealthCheck struct {
				Enabled  bool
				Interval time.Duration
				Timeout  time.Duration
				Checks   []config.HealthCheckConfig
			}

			RateLimit struct {
				Enabled     bool
				RequestsPer int64
				WindowSize  int64
			}

			Authentication struct {
				Enabled     bool
				Type        string
				JWTSecret   string
				KeycloakURL string
			}
		}{
			Enabled: false,
		},
		Metrics: struct {
			Enabled bool   `json:"enabled"`
			Port    int    `json:"port"`
			Path    string `json:"path"`
		}{
			Enabled: false,
		},
		Security: struct {
			TLSEnabled bool     `json:"tls_enabled"`
			CertFile   string   `json:"cert_file"`
			KeyFile    string   `json:"key_file"`
			AllowedIPs []string `json:"allowed_ips"`
			RateLimit  struct {
				Enabled           bool    `json:"enabled"`
				RequestsPerSecond float64 `json:"requests_per_second"`
				Burst             int     `json:"burst"`
			} `json:"rate_limit"`
			MTLS struct {
				Enabled bool   `json:"enabled"`
				CACert  string `json:"ca_file"`
			} `json:"mtls"`
			CipherSuites  []string `json:"cipher_suites"`
			MinTLSVersion string   `json:"min_tls_version"`
			Audit         struct {
				Enabled bool   `json:"enabled"`
				LogFile string `json:"log_file"`
			} `json:"audit"`
		}{
			TLSEnabled: false,
		},
	}
}

// TunnelInfo представляет информацию о туннеле для тестов
type TunnelInfo struct {
	ID       string
	Protocol string
	Endpoint string
}
