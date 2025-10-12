package main

import (
	"fmt"
	"os"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/auth"
	"github.com/twogc/cloudbridge-relay-installer/internal/balancer"
	"github.com/twogc/cloudbridge-relay-installer/internal/metrics"
	"github.com/twogc/cloudbridge-relay-installer/internal/quic"
	"github.com/twogc/cloudbridge-relay-installer/internal/ratelimit"
	"github.com/twogc/cloudbridge-relay-installer/internal/tenant"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// EnhancedQUICManager управляет Enhanced QUIC функциональностью
type EnhancedQUICManager struct {
	config         *quic.EnhancedConfig
	tenantManager  *quic.TenantManager
	ddosProtection *quic.DDoSProtection
	geoRouter      *quic.GeographicRouter
	keycloakAuth   *quic.KeycloakAuth
	logger         *zap.Logger

	// Интеграции с существующими компонентами
	relayTenantManager *tenant.Manager
	relayBalancer      *balancer.Balancer
	relayRateLimit     *ratelimit.IPLimiter
	relayKeycloakAuth  *auth.EnhancedKeycloakAuth

	// Метрики
	metrics *metrics.Metrics

	enabled bool
}

// NewEnhancedQUICManager создает новый Enhanced QUIC Manager
func NewEnhancedQUICManager(logger *zap.Logger, metrics *metrics.Metrics) *EnhancedQUICManager {
	config := &quic.EnhancedConfig{
		MaxStreams:                     100,
		KeepAlive:                      true,
		IdleTimeout:                    30 * time.Second,
		HandshakeTimeout:               10 * time.Second,
		MaxIncomingStreams:             100,
		MaxIncomingUniStreams:          100,
		InitialStreamReceiveWindow:     512 * 1024,      // 512KB
		MaxStreamReceiveWindow:         2 * 1024 * 1024, // 2MB
		InitialConnectionReceiveWindow: 1024 * 1024,     // 1MB
		MaxConnectionReceiveWindow:     5 * 1024 * 1024, // 5MB
		CongestionControlAlgorithm:     "cubic",

		// Enhanced features
		EnableDDoSProtection:      true,
		EnableGeographicRouting:   true,
		EnableKeycloakIntegration: false, // По умолчанию отключено
	}

	return &EnhancedQUICManager{
		config:         config,
		tenantManager:  quic.NewTenantManager(),
		ddosProtection: quic.NewDDoSProtection(),
		geoRouter:      quic.NewGeographicRouter(),
		logger:         logger,
		metrics:        metrics,
		enabled:        false, // По умолчанию отключено
	}
}

// Initialize инициализирует Enhanced QUIC
func (eqm *EnhancedQUICManager) Initialize() error {
	eqm.logger.Info("Initializing Enhanced QUIC Manager")

	// Настраиваем тенантов
	eqm.config.TenantManager = eqm.tenantManager

	// Загружаем tenant'ы из конфигурационного файла
	if err := eqm.loadTenantsFromConfig(); err != nil {
		eqm.logger.Warn("Failed to load tenants from config, using defaults", zap.Error(err))
		// Fallback к тестовым tenant'ам
		eqm.tenantManager.AddTenant("premium", 1024*1024*1024, 100) // 1GB, 100 connections
		eqm.tenantManager.AddTenant("standard", 100*1024*1024, 10)  // 100MB, 10 connections
		eqm.tenantManager.AddTenant("basic", 10*1024*1024, 5)       // 10MB, 5 connections
	}

	// Настраиваем Keycloak, если включен
	if eqm.config.EnableKeycloakIntegration && eqm.config.KeycloakConfig != nil {
		eqm.keycloakAuth = quic.NewKeycloakAuth(eqm.config.KeycloakConfig)
		eqm.logger.Info("Keycloak integration enabled")
	}

	eqm.enabled = true
	eqm.logger.Info("Enhanced QUIC Manager initialized successfully",
		zap.Bool("ddos_protection", eqm.config.EnableDDoSProtection),
		zap.Bool("geographic_routing", eqm.config.EnableGeographicRouting),
		zap.Bool("keycloak_integration", eqm.config.EnableKeycloakIntegration),
	)

	return nil
}

// SetupIntegrations настраивает интеграции с существующими компонентами
func (eqm *EnhancedQUICManager) SetupIntegrations(
	tenantMgr *tenant.Manager,
	balancer *balancer.Balancer,
	rateLimit *ratelimit.IPLimiter,
	keycloakAuth *auth.EnhancedKeycloakAuth,
) error {
	eqm.logger.Info("Setting up Enhanced QUIC integrations")

	// Интеграция с Tenant Manager
	if tenantMgr != nil {
		eqm.relayTenantManager = tenantMgr
		tenantMgr.EnableEnhanced()
		eqm.logger.Info("Tenant Manager integration enabled")
	}

	// Интеграция с Load Balancer
	if balancer != nil {
		eqm.relayBalancer = balancer
		balancer.EnableEnhanced()
		eqm.logger.Info("Load Balancer integration enabled")
	}

	// Интеграция с Rate Limiting
	if rateLimit != nil {
		eqm.relayRateLimit = rateLimit
		rateLimit.EnableEnhanced()
		eqm.logger.Info("Rate Limiting integration enabled")
	}

	// Интеграция с Keycloak Auth
	if keycloakAuth != nil {
		eqm.relayKeycloakAuth = keycloakAuth
		// Включаем Enhanced режим только если не был включен ранее
		if !keycloakAuth.IsEnhancedEnabled() {
			if err := keycloakAuth.EnableEnhanced(nil); err != nil {
				eqm.logger.Warn("Failed to enable Enhanced Keycloak Auth", zap.Error(err))
			} else {
				eqm.logger.Info("Keycloak Auth integration enabled")
			}
		}
	}

	return nil
}

// CreateConnection создает Enhanced соединение
func (eqm *EnhancedQUICManager) CreateConnection(network, address, tenantID string) (*quic.EnhancedConnection, error) {
	if !eqm.enabled {
		return nil, fmt.Errorf("enhanced QUIC manager is not enabled")
	}

	eqm.logger.Debug("Creating Enhanced QUIC connection",
		zap.String("network", network),
		zap.String("address", address),
		zap.String("tenant_id", tenantID),
	)

	// Записываем метрики создания соединения
	if eqm.metrics != nil {
		eqm.metrics.IncQUICConnections()
	}

	// Создаем Enhanced политики для QUIC соединения
	enhancedConn, err := quic.NewEnhancedPolicies(tenantID, eqm.config)
	if err != nil {
		// Записываем метрики ошибки соединения
		if eqm.metrics != nil {
			eqm.metrics.IncQUICConnectionsFailed("dial_failed", "unknown")
		}
		return nil, err
	}

	return enhancedConn, nil
}

// ValidateConnection проверяет соединение через все интегрированные компоненты
func (eqm *EnhancedQUICManager) ValidateConnection(clientIP, tenantID, token string) error {
	if !eqm.enabled {
		return fmt.Errorf("enhanced QUIC manager is not enabled")
	}

	// Проверяем DDoS Protection
	if eqm.config.EnableDDoSProtection && eqm.ddosProtection != nil {
		if err := eqm.ddosProtection.CheckRateLimit(clientIP); err != nil {
			eqm.logger.Warn("DDoS Protection blocked connection",
				zap.String("client_ip", clientIP),
				zap.Error(err),
			)
			// Записываем метрики DDoS блокировки
			if eqm.metrics != nil {
				eqm.metrics.IncQUICConnectionsFailed("ddos_protection", "unknown")
			}
			return err
		}
	}

	// Проверяем лимиты тенанта через интегрированный Tenant Manager
	if eqm.relayTenantManager != nil {
		if err := eqm.relayTenantManager.CheckEnhancedTenantLimits(tenantID); err != nil {
			eqm.logger.Warn("Tenant limits exceeded",
				zap.String("tenant_id", tenantID),
				zap.Error(err),
			)
			// Записываем метрики превышения лимитов тенанта
			if eqm.metrics != nil {
				eqm.metrics.IncQUICConnectionsFailed("tenant_limits", "unknown")
			}
			return err
		}
	}

	// Проверяем через интегрированный Rate Limiter
	if eqm.relayRateLimit != nil {
		if !eqm.relayRateLimit.Allow(clientIP) {
			eqm.logger.Warn("Rate limit exceeded",
				zap.String("client_ip", clientIP),
			)
			// Записываем метрики превышения rate limit
			if eqm.metrics != nil {
				eqm.metrics.IncQUICConnectionsFailed("rate_limit", "unknown")
			}
			return fmt.Errorf("rate limit exceeded for IP %s", clientIP)
		}
	}

	// Проверяем аутентификацию через интегрированный Keycloak
	if eqm.config.EnableKeycloakIntegration && eqm.relayKeycloakAuth != nil && token != "" {
		if isValid, err := eqm.relayKeycloakAuth.ValidateTokenEnhanced(token); err != nil || !isValid {
			eqm.logger.Warn("Keycloak authentication failed",
				zap.String("tenant_id", tenantID),
				zap.Error(err),
				zap.Bool("is_valid", isValid),
			)
			// Записываем метрики ошибок аутентификации
			if eqm.metrics != nil {
				eqm.metrics.IncQUICAuthenticationFailures("token_validation", "unknown")
			}
			return fmt.Errorf("authentication failed: %v", err)
		}
	}

	return nil
}

// GetServerByGeo возвращает сервер на основе географического маршрута
func (eqm *EnhancedQUICManager) GetServerByGeo(clientIP string) (string, error) {
	if !eqm.enabled {
		return "", fmt.Errorf("enhanced QUIC manager is not enabled")
	}

	if eqm.config.EnableGeographicRouting && eqm.geoRouter != nil {
		route := eqm.geoRouter.RouteByLocation(clientIP)
		eqm.logger.Debug("Geographic routing",
			zap.String("client_ip", clientIP),
			zap.String("route", route),
		)
		return route, nil
	}

	// Fallback через интегрированный балансировщик
	if eqm.relayBalancer != nil {
		if server := eqm.relayBalancer.GetServerByGeo(clientIP); server != nil {
			return server.Address, nil
		}
	}

	return "", fmt.Errorf("no server available for IP %s", clientIP)
}

// GetStats возвращает статистику Enhanced QUIC
func (eqm *EnhancedQUICManager) GetStats() map[string]interface{} {
	stats := map[string]interface{}{
		"enabled":                      eqm.enabled,
		"ddos_protection_enabled":      eqm.config.EnableDDoSProtection,
		"geographic_routing_enabled":   eqm.config.EnableGeographicRouting,
		"keycloak_integration_enabled": eqm.config.EnableKeycloakIntegration,
	}

	// Добавляем статистику тенантов
	if eqm.tenantManager != nil {
		tenantStats := make(map[string]interface{})
		// Здесь можно добавить детальную статистику по тенантам
		stats["tenants"] = tenantStats
	}

	return stats
}

// Enable включает Enhanced QUIC Manager
func (eqm *EnhancedQUICManager) Enable() error {
	if eqm.enabled {
		return nil // Уже включен
	}

	eqm.logger.Info("Enabling Enhanced QUIC Manager")
	return eqm.Initialize()
}

// Disable отключает Enhanced QUIC Manager
func (eqm *EnhancedQUICManager) Disable() {
	if !eqm.enabled {
		return // Уже отключен
	}

	eqm.logger.Info("Disabling Enhanced QUIC Manager")

	// Отключаем интеграции
	if eqm.relayTenantManager != nil {
		eqm.relayTenantManager.DisableEnhanced()
	}
	if eqm.relayBalancer != nil {
		eqm.relayBalancer.DisableEnhanced()
	}
	if eqm.relayRateLimit != nil {
		eqm.relayRateLimit.DisableEnhanced()
	}
	if eqm.relayKeycloakAuth != nil {
		eqm.relayKeycloakAuth.DisableEnhanced()
	}

	eqm.enabled = false
	eqm.logger.Info("Enhanced QUIC Manager disabled")
}

// IsEnabled проверяет, включен ли Enhanced QUIC Manager
func (eqm *EnhancedQUICManager) IsEnabled() bool {
	return eqm.enabled
}

// GetConfig возвращает конфигурацию Enhanced QUIC
func (eqm *EnhancedQUICManager) GetConfig() *quic.EnhancedConfig {
	return eqm.config
}

// UpdateConfig обновляет конфигурацию Enhanced QUIC
func (eqm *EnhancedQUICManager) UpdateConfig(config *quic.EnhancedConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	eqm.config = config
	eqm.logger.Info("Enhanced QUIC configuration updated")

	// Переинициализируем, если включен
	if eqm.enabled {
		return eqm.Initialize()
	}

	return nil
}

// loadTenantsFromConfig загружает tenant'ы из конфигурационного файла enhanced_quic.yaml
func (eqm *EnhancedQUICManager) loadTenantsFromConfig() error {
	eqm.logger.Info("Loading tenants from configuration file")

	// Путь к конфигурационному файлу
	configPath := "configs/enhanced_quic.yaml"

	// Читаем файл конфигурации
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	// Парсим YAML
	var config struct {
		Tenants map[string]struct {
			BandwidthLimit  int64    `yaml:"bandwidth_limit"`
			ConnectionLimit int      `yaml:"connection_limit"`
			AllowedIPs      []string `yaml:"allowed_ips"`
			AllowedPorts    []int    `yaml:"allowed_ports"`
			RateLimit       int      `yaml:"rate_limit"`
			Burst           int      `yaml:"burst"`
		} `yaml:"tenants"`
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse YAML config: %w", err)
	}

	// Добавляем tenant'ы в менеджер и синхронизируем с relayTenantManager
	for tenantID, tenantConfig := range config.Tenants {
		eqm.logger.Info("Adding tenant from config",
			zap.String("tenant_id", tenantID),
			zap.Int64("bandwidth_limit", tenantConfig.BandwidthLimit),
			zap.Int("connection_limit", tenantConfig.ConnectionLimit),
		)

		// Внутренний Enhanced TenantManager (для QUIC)
		eqm.tenantManager.AddTenant(tenantID, tenantConfig.BandwidthLimit, tenantConfig.ConnectionLimit)

		// Также добавляем в общий relay TenantManager, чтобы health/валидации видели тенантов
		if eqm.relayTenantManager != nil {
			limits := &tenant.TenantLimits{
				MaxTunnels:     tenantConfig.ConnectionLimit,
				MaxConnections: tenantConfig.ConnectionLimit,
				RateLimit:      float64(tenantConfig.RateLimit),
				Burst:          tenantConfig.Burst,
				AllowedIPs:     tenantConfig.AllowedIPs,
				AllowedPorts:   tenantConfig.AllowedPorts,
			}
			_ = eqm.relayTenantManager.CreateTenant(tenantID, tenantID, limits)
		}
	}

	eqm.logger.Info("Tenants loaded from config successfully",
		zap.Int("total_tenants", len(config.Tenants)),
	)

	return nil
}

// RecordConnectionMetrics записывает метрики производительности соединения
func (eqm *EnhancedQUICManager) RecordConnectionMetrics(connectionID string, metrics map[string]interface{}) {
	if eqm.metrics == nil {
		return
	}

	// Записываем метрики производительности
	if rtt, ok := metrics["rtt"].(float64); ok {
		eqm.metrics.ObserveQUICRTT(connectionID, rtt)
	}

	if latency, ok := metrics["latency"].(float64); ok {
		eqm.metrics.ObserveQUICConnectionLatency(connectionID, latency)
	}

	if throughput, ok := metrics["throughput_bytes"].(int64); ok {
		direction := "unknown"
		if dir, exists := metrics["direction"].(string); exists {
			direction = dir
		}
		eqm.metrics.IncQUICThroughputBytes(connectionID, direction, throughput)
	}

	if packets, ok := metrics["throughput_packets"].(int64); ok {
		direction := "unknown"
		if dir, exists := metrics["direction"].(string); exists {
			direction = dir
		}
		eqm.metrics.IncQUICThroughputPackets(connectionID, direction, packets)
	}

	if packetLoss, ok := metrics["packet_loss_rate"].(float64); ok {
		eqm.metrics.SetQUICPacketLossRate(connectionID, packetLoss)
	}

	if congestionWindow, ok := metrics["congestion_window"].(int64); ok {
		eqm.metrics.SetQUICCongestionWindow(connectionID, congestionWindow)
	}

	if bytesInFlight, ok := metrics["bytes_in_flight"].(int64); ok {
		eqm.metrics.SetQUICBytesInFlight(connectionID, bytesInFlight)
	}

	if packetsInFlight, ok := metrics["packets_in_flight"].(int64); ok {
		eqm.metrics.SetQUICPacketsInFlight(connectionID, packetsInFlight)
	}

	// Записываем метрики ресурсов
	if memoryUsage, ok := metrics["memory_usage"].(int64); ok {
		component := "connection"
		if comp, exists := metrics["component"].(string); exists {
			component = comp
		}
		eqm.metrics.SetQUICMemoryUsage(connectionID, component, memoryUsage)
	}

	if cpuUsage, ok := metrics["cpu_usage"].(float64); ok {
		component := "connection"
		if comp, exists := metrics["component"].(string); exists {
			component = comp
		}
		eqm.metrics.SetQUICCPUUsage(connectionID, component, cpuUsage)
	}

	if goroutineCount, ok := metrics["goroutine_count"].(int); ok {
		component := "connection"
		if comp, exists := metrics["component"].(string); exists {
			component = comp
		}
		eqm.metrics.SetQUICGoroutineCount(connectionID, component, goroutineCount)
	}
}

// RecordStreamMetrics записывает метрики потоков
func (eqm *EnhancedQUICManager) RecordStreamMetrics(streamType, direction, reason string) {
	if eqm.metrics == nil {
		return
	}

	eqm.metrics.IncQUICStreamsCreated(streamType, direction)
}

// RecordErrorMetrics записывает метрики ошибок
func (eqm *EnhancedQUICManager) RecordErrorMetrics(errorType, connectionID, streamID string) {
	if eqm.metrics == nil {
		return
	}

	switch errorType {
	case "packet_error":
		eqm.metrics.IncQUICPacketErrors("packet_corruption", connectionID)
	case "connection_error":
		eqm.metrics.IncQUICConnectionErrors("connection_failed", connectionID)
	case "stream_error":
		eqm.metrics.IncQUICStreamErrors("stream_reset", streamID)
	case "timeout":
		eqm.metrics.IncQUICTimeoutEvents("connection_timeout", connectionID)
	case "retransmission":
		eqm.metrics.IncQUICRetransmissions(connectionID, "packet_loss")
	case "encryption_failure":
		eqm.metrics.IncQUICEncryptionFailures("key_rotation", connectionID)
	}
}
