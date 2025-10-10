package p2p

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"github.com/twogc/cloudbridge-relay-installer/internal/config"
	"github.com/twogc/cloudbridge-relay-installer/internal/database"
	"github.com/twogc/cloudbridge-relay-installer/internal/wireguard"
)

// RelayManagerSCORE управляет P2P relay соединениями с интеграцией SCORE
type RelayManagerSCORE struct {
	tenants          map[string]*Tenant
	connections      map[string]*RelayP2PConnection
	quicConnections  map[string]*QUICConnection
	scoreClient      *config.SCOREClient
	repository       *database.Repository
	wireguardManager *wireguard.WireGuardTenantManager
	mutex            sync.RWMutex
	logger           *zap.Logger
	ctx              context.Context
	cancel           context.CancelFunc
	configCache      map[string]*config.TenantNetworkConfig
	cacheMutex       sync.RWMutex
}

// NewRelayManagerSCORE создает новый Relay Manager с интеграцией SCORE
func NewRelayManagerSCORE(logger *zap.Logger, repository *database.Repository, scoreClient *config.SCOREClient) *RelayManagerSCORE {
	ctx, cancel := context.WithCancel(context.Background())

	rm := &RelayManagerSCORE{
		tenants:         make(map[string]*Tenant),
		connections:     make(map[string]*RelayP2PConnection),
		quicConnections: make(map[string]*QUICConnection),
		scoreClient:     scoreClient,
		repository:      repository,
		logger:          logger,
		ctx:             ctx,
		cancel:          cancel,
		configCache:     make(map[string]*config.TenantNetworkConfig),
	}

	// Инициализируем WireGuard менеджер для VIP tenant'ов
	rm.wireguardManager = wireguard.NewWireGuardTenantManager(repository, logger)

	// Загружаем конфигурации из SCORE
	if err := rm.loadConfigurationsFromSCORE(); err != nil {
		logger.Error("Failed to load configurations from SCORE", zap.Error(err))
	}

	// Запускаем watcher для изменений конфигураций
	go rm.watchConfigurationChanges()

	return rm
}

// loadConfigurationsFromSCORE загружает все конфигурации сетей из SCORE
func (rm *RelayManagerSCORE) loadConfigurationsFromSCORE() error {
	configs, err := rm.scoreClient.GetAllTenantNetworkConfigs(rm.ctx)
	if err != nil {
		return fmt.Errorf("failed to load configurations from SCORE: %w", err)
	}

	rm.cacheMutex.Lock()
	defer rm.cacheMutex.Unlock()

	for _, config := range configs {
		if config.IsActive {
			rm.configCache[config.TenantID] = config
			
			// Создаем Tenant объект
			tenant := &Tenant{
				TenantID:       config.TenantID,
				NetworkCIDR:    config.NetworkCIDR,
				Peers:          make(map[string]*Peer),
				Connections:    make(map[string]*RelayP2PConnection),
				MaxPeers:       100, // Default value
				MaxConnections: 1000, // Default value
				CreatedAt:      config.CreatedAt,
				LastActivity:   time.Now(),
			}
			
			rm.tenants[config.TenantID] = tenant
			
			rm.logger.Info("Loaded tenant configuration from SCORE",
				zap.String("tenant_id", config.TenantID),
				zap.String("network_cidr", config.NetworkCIDR),
				zap.Int("version", config.Version))
		}
	}

	rm.logger.Info("Loaded tenant configurations from SCORE",
		zap.Int("count", len(rm.configCache)))

	return nil
}

// watchConfigurationChanges подписывается на изменения конфигураций в SCORE
func (rm *RelayManagerSCORE) watchConfigurationChanges() {
	rm.logger.Info("Starting configuration watcher for SCORE")
	
	callback := func(config *config.TenantNetworkConfig) {
		rm.handleConfigurationChange(config)
	}
	
	if err := rm.scoreClient.WatchTenantNetworkConfigs(rm.ctx, callback); err != nil {
		rm.logger.Error("Configuration watcher failed", zap.Error(err))
	}
}

// handleConfigurationChange обрабатывает изменения конфигурации
func (rm *RelayManagerSCORE) handleConfigurationChange(config *config.TenantNetworkConfig) {
	rm.cacheMutex.Lock()
	defer rm.cacheMutex.Unlock()

	tenantID := config.TenantID
	
	if config.IsActive {
		// Обновляем или создаем конфигурацию
		rm.configCache[tenantID] = config
		
		// Обновляем Tenant объект
		if tenant, exists := rm.tenants[tenantID]; exists {
			tenant.NetworkCIDR = config.NetworkCIDR
			tenant.LastActivity = time.Now()
		} else {
			// Создаем новый Tenant
			tenant := &Tenant{
				TenantID:       tenantID,
				NetworkCIDR:    config.NetworkCIDR,
				Peers:          make(map[string]*Peer),
				Connections:    make(map[string]*RelayP2PConnection),
				MaxPeers:       100,
				MaxConnections: 1000,
				CreatedAt:      config.CreatedAt,
				LastActivity:   time.Now(),
			}
			rm.tenants[tenantID] = tenant
		}
		
		rm.logger.Info("Updated tenant configuration from SCORE",
			zap.String("tenant_id", tenantID),
			zap.String("network_cidr", config.NetworkCIDR),
			zap.Int("version", config.Version))
	} else {
		// Деактивируем конфигурацию
		delete(rm.configCache, tenantID)
		
		// Закрываем все соединения тенанта
		if tenant, exists := rm.tenants[tenantID]; exists {
			rm.deactivateTenantConnections(tenant)
			delete(rm.tenants, tenantID)
		}
		
		rm.logger.Info("Deactivated tenant configuration from SCORE",
			zap.String("tenant_id", tenantID))
	}
}

// deactivateTenantConnections закрывает все соединения тенанта
func (rm *RelayManagerSCORE) deactivateTenantConnections(tenant *Tenant) {
	tenant.Mutex.Lock()
	defer tenant.Mutex.Unlock()

	// Закрываем все P2P соединения
	for _, conn := range tenant.Connections {
		if conn.QUICConn != nil {
			conn.QUICConn.CloseWithError(0, "tenant deactivated")
		}
	}
	
	// Закрываем все QUIC соединения
	rm.mutex.Lock()
	for peerID, quicConn := range rm.quicConnections {
		if quicConn.TenantID == tenant.TenantID {
			quicConn.Conn.CloseWithError(0, "tenant deactivated")
			delete(rm.quicConnections, peerID)
		}
	}
	rm.mutex.Unlock()
}

// GetTenantNetwork возвращает выделенную сеть для тенанта из SCORE
func (rm *RelayManagerSCORE) GetTenantNetwork(tenantID string) string {
	rm.cacheMutex.RLock()
	defer rm.cacheMutex.RUnlock()

	// Сначала проверяем кэш
	if config, exists := rm.configCache[tenantID]; exists {
		return config.NetworkCIDR
	}

	// Если нет в кэше, запрашиваем из SCORE
	config, err := rm.scoreClient.GetTenantNetworkConfig(rm.ctx, tenantID)
	if err != nil {
		rm.logger.Error("Failed to get tenant network config from SCORE",
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		return ""
	}

	// Обновляем кэш
	rm.configCache[tenantID] = config
	return config.NetworkCIDR
}

// CreateTenantNetwork создает новую сеть для тенанта через SCORE
func (rm *RelayManagerSCORE) CreateTenantNetwork(tenantID string) (string, error) {
	config, err := rm.scoreClient.GetTenantNetworkConfig(rm.ctx, tenantID)
	if err != nil {
		return "", fmt.Errorf("failed to create tenant network in SCORE: %w", err)
	}

	// Обновляем кэш
	rm.cacheMutex.Lock()
	rm.configCache[tenantID] = config
	rm.cacheMutex.Unlock()

	// Создаем Tenant объект
	tenant := &Tenant{
		TenantID:       tenantID,
		NetworkCIDR:    config.NetworkCIDR,
		Peers:          make(map[string]*Peer),
		Connections:    make(map[string]*RelayP2PConnection),
		MaxPeers:       100,
		MaxConnections: 1000,
		CreatedAt:      config.CreatedAt,
		LastActivity:   time.Now(),
	}

	rm.mutex.Lock()
	rm.tenants[tenantID] = tenant
	rm.mutex.Unlock()

	rm.logger.Info("Created tenant network through SCORE",
		zap.String("tenant_id", tenantID),
		zap.String("network_cidr", config.NetworkCIDR))

	return config.NetworkCIDR, nil
}

// DeactivateTenantNetwork деактивирует сеть тенанта через SCORE
func (rm *RelayManagerSCORE) DeactivateTenantNetwork(tenantID string) error {
	if err := rm.scoreClient.DeactivateTenantNetworkConfig(rm.ctx, tenantID); err != nil {
		return fmt.Errorf("failed to deactivate tenant network in SCORE: %w", err)
	}

	// Удаляем из кэша
	rm.cacheMutex.Lock()
	delete(rm.configCache, tenantID)
	rm.cacheMutex.Unlock()

	// Закрываем соединения и удаляем Tenant
	rm.mutex.Lock()
	if tenant, exists := rm.tenants[tenantID]; exists {
		rm.deactivateTenantConnections(tenant)
		delete(rm.tenants, tenantID)
	}
	rm.mutex.Unlock()

	rm.logger.Info("Deactivated tenant network through SCORE",
		zap.String("tenant_id", tenantID))

	return nil
}

// GetTenantInfo возвращает информацию о тенанте
func (rm *RelayManagerSCORE) GetTenantInfo(tenantID string) (*Tenant, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	tenant, exists := rm.tenants[tenantID]
	if !exists {
		return nil, fmt.Errorf("tenant not found: %s", tenantID)
	}

	return tenant, nil
}

// GetAllTenants возвращает все активные тенанты
func (rm *RelayManagerSCORE) GetAllTenants() map[string]*Tenant {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	// Возвращаем копию для безопасности
	tenants := make(map[string]*Tenant)
	for id, tenant := range rm.tenants {
		tenants[id] = tenant
	}

	return tenants
}

// GetTenantStats возвращает статистику тенанта
func (rm *RelayManagerSCORE) GetTenantStats(tenantID string) map[string]interface{} {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	tenant, exists := rm.tenants[tenantID]
	if !exists {
		return map[string]interface{}{
			"error": "tenant not found",
		}
	}

	tenant.Mutex.RLock()
	peerCount := len(tenant.Peers)
	connectionCount := len(tenant.Connections)
	tenant.Mutex.RUnlock()

	return map[string]interface{}{
		"tenant_id":        tenantID,
		"network_cidr":     tenant.NetworkCIDR,
		"peer_count":       peerCount,
		"connection_count": connectionCount,
		"max_peers":        tenant.MaxPeers,
		"max_connections":  tenant.MaxConnections,
		"created_at":       tenant.CreatedAt,
		"last_activity":    tenant.LastActivity,
	}
}

// Shutdown корректно завершает работу RelayManager
func (rm *RelayManagerSCORE) Shutdown() error {
	rm.logger.Info("Shutting down RelayManager with SCORE integration")

	// Отменяем контекст
	rm.cancel()

	// Закрываем все соединения
	rm.mutex.Lock()
	for _, tenant := range rm.tenants {
		rm.deactivateTenantConnections(tenant)
	}
	rm.mutex.Unlock()

	rm.logger.Info("RelayManager shutdown completed")
	return nil
}

