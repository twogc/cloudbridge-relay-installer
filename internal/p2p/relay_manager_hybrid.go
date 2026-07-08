package p2p

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/database"
	"github.com/twogc/cloudbridge-relay-installer/internal/network"
	"github.com/twogc/cloudbridge-relay-installer/internal/wireguard"
	"go.uber.org/zap"
)

// RelayManagerHybrid исправленная версия RelayManager с гибридной архитектурой
type RelayManagerHybrid struct {
	tenants          map[string]*Tenant
	connections      map[string]*RelayP2PConnection
	quicConnections  map[string]*QUICConnection
	networkStore     network.TenantNetworkStore
	networkAllocator network.TenantNetworkAllocator
	peerIPAMRepo     *database.PeerIPAMRepository
	wireguardManager *wireguard.WireGuardTenantManager
	mutex            sync.RWMutex
	logger           *zap.Logger
	ctx              context.Context
	cancel           context.CancelFunc
	configCache      map[string]*network.TenantNetwork
	cacheMutex       sync.RWMutex
	syncStatus       string // "fresh", "stale", "error"
	resourceVersion  network.ResourceVersion
}

// NewRelayManagerHybrid создает новый исправленный Relay Manager
func NewRelayManagerHybrid(
	logger *zap.Logger,
	repository *database.Repository,
	networkStore network.TenantNetworkStore,
	networkAllocator network.TenantNetworkAllocator,
) *RelayManagerHybrid {
	ctx, cancel := context.WithCancel(context.Background())

	var peerIPAMRepo *database.PeerIPAMRepository
	if repository != nil && repository.GetDB() != nil {
		peerIPAMRepo = database.NewPeerIPAMRepository(repository.GetDB(), logger)
	} else if logger != nil {
		logger.Warn("Repository not initialized; hybrid manager peer IPAM will be in-memory unavailable")
	}

	rm := &RelayManagerHybrid{
		tenants:          make(map[string]*Tenant),
		connections:      make(map[string]*RelayP2PConnection),
		quicConnections:  make(map[string]*QUICConnection),
		networkStore:     networkStore,
		networkAllocator: networkAllocator,
		peerIPAMRepo:     peerIPAMRepo,
		logger:           logger,
		ctx:              ctx,
		cancel:           cancel,
		configCache:      make(map[string]*network.TenantNetwork),
		syncStatus:       "fresh",
	}

	// Инициализируем WireGuard менеджер (repository may be nil)
	rm.wireguardManager = wireguard.NewWireGuardTenantManager(repository, logger)

	// Загружаем конфигурации из хранилища
	if err := rm.initializeFromStore(); err != nil {
		logger.Error("Failed to initialize from network store", zap.Error(err))
		rm.syncStatus = "error"
	}

	// Запускаем watcher для изменений конфигураций
	go rm.consumeNetworkEvents()

	return rm
}

// initializeFromStore загружает все конфигурации сетей из хранилища
func (rm *RelayManagerHybrid) initializeFromStore() error {
	networks, resourceVersion, err := rm.networkStore.ListActive(rm.ctx)
	if err != nil {
		return fmt.Errorf("failed to load networks from store: %w", err)
	}

	rm.cacheMutex.Lock()
	defer rm.cacheMutex.Unlock()

	for _, network := range networks {
		if network.IsActive {
			rm.configCache[network.TenantID] = network

			// Создаем Tenant объект
			tenant := &Tenant{
				TenantID:       network.TenantID,
				NetworkCIDR:    network.NetworkCIDR,
				Peers:          make(map[string]*Peer),
				Connections:    make(map[string]*RelayP2PConnection),
				MaxPeers:       100,  // Default value
				MaxConnections: 1000, // Default value
				CreatedAt:      time.Unix(network.AllocatedAtUnix, 0),
				LastActivity:   time.Now(),
			}

			rm.tenants[network.TenantID] = tenant

			rm.logger.Info("Loaded tenant configuration from store",
				zap.String("tenant_id", network.TenantID),
				zap.String("network_cidr", network.NetworkCIDR),
				zap.Int("version", network.Version))
		}
	}

	rm.resourceVersion = resourceVersion

	rm.logger.Info("Loaded tenant configurations from store",
		zap.Int("count", len(rm.configCache)),
		zap.String("resource_version", string(resourceVersion)))

	return nil
}

// consumeNetworkEvents подписывается на изменения конфигураций с самовосстановлением
func (rm *RelayManagerHybrid) consumeNetworkEvents() {
	rm.logger.Info("Starting network events consumer")

	for {
		eventChan, err := rm.networkStore.Watch(rm.ctx, rm.resourceVersion)
		if err != nil {
			rm.syncStatus = "error"
			rm.logger.Error("watch failed, retrying...", zap.Error(err))
			select {
			case <-time.After(2 * time.Second):
				continue
			case <-rm.ctx.Done():
				return
			}
		}

		rm.syncStatus = "fresh"
		rm.logger.Info("Successfully subscribed to network events",
			zap.String("resource_version", string(rm.resourceVersion)))

		for {
			select {
			case <-rm.ctx.Done():
				rm.logger.Info("Network events consumer stopped")
				return
			case ev, ok := <-eventChan:
				if !ok {
					rm.logger.Warn("watch channel closed; resubscribing")
					rm.syncStatus = "stale"
					break
				}
				rm.handleNetworkEvent(ev)
			}
		}
	}
}

// handleNetworkEvent обрабатывает события изменения конфигураций
func (rm *RelayManagerHybrid) handleNetworkEvent(event network.TenantNetworkEvent) {
	rm.cacheMutex.Lock()
	defer rm.cacheMutex.Unlock()

	// Проверяем на nil Tenant
	if event.Tenant == nil {
		rm.logger.Warn("Tenant event without body",
			zap.String("type", string(event.Type)),
			zap.String("resource_version", string(event.ResourceVersion)))
		return
	}

	tenantID := event.Tenant.TenantID

	// Обновляем resource version
	rm.resourceVersion = event.ResourceVersion

	switch event.Type {
	case network.EventCreated, network.EventUpdated:
		// Проверяем активность тенанта
		if !event.Tenant.IsActive {
			// Трактуем как удаление/деактивацию
			delete(rm.configCache, tenantID)
			rm.mutex.Lock()
			if tenant, exists := rm.tenants[tenantID]; exists {
				rm.deactivateTenantConnections(tenant)
				delete(rm.tenants, tenantID)
			}
			rm.mutex.Unlock()

			rm.logger.Info("Tenant deactivated via update",
				zap.String("tenant_id", tenantID),
				zap.Int("version", event.Tenant.Version),
				zap.String("reason", event.Reason))
			return
		}

		// Обновляем или создаем конфигурацию
		rm.configCache[tenantID] = event.Tenant

		// Обновляем Tenant объект
		rm.mutex.Lock()
		if tenant, exists := rm.tenants[tenantID]; exists {
			tenant.NetworkCIDR = event.Tenant.NetworkCIDR
			tenant.LastActivity = time.Now()
		} else {
			// Создаем новый Tenant
			tenant := &Tenant{
				TenantID:       tenantID,
				NetworkCIDR:    event.Tenant.NetworkCIDR,
				Peers:          make(map[string]*Peer),
				Connections:    make(map[string]*RelayP2PConnection),
				MaxPeers:       100,
				MaxConnections: 1000,
				CreatedAt:      time.Unix(event.Tenant.AllocatedAtUnix, 0),
				LastActivity:   time.Now(),
			}
			rm.tenants[tenantID] = tenant
		}
		rm.mutex.Unlock()

		rm.logger.Info("Tenant config upserted from store",
			zap.String("tenant_id", tenantID),
			zap.String("network_cidr", event.Tenant.NetworkCIDR),
			zap.Int("version", event.Tenant.Version),
			zap.String("reason", event.Reason))

	case network.EventDeleted:
		// Удаляем конфигурацию
		delete(rm.configCache, tenantID)

		// Закрываем все соединения тенанта
		rm.mutex.Lock()
		if tenant, exists := rm.tenants[tenantID]; exists {
			rm.deactivateTenantConnections(tenant)
			delete(rm.tenants, tenantID)
		}
		rm.mutex.Unlock()

		rm.logger.Info("Tenant config deleted from store",
			zap.String("tenant_id", tenantID),
			zap.String("reason", event.Reason))
	}
}

// deactivateTenantConnections закрывает все соединения тенанта
func (rm *RelayManagerHybrid) deactivateTenantConnections(tenant *Tenant) {
	// Сначала освобождаем IP адреса пиров
	if rm.peerIPAMRepo != nil {
		if count, err := rm.peerIPAMRepo.DeallocateIPsByTenant(rm.ctx, tenant.TenantID); err != nil {
			rm.logger.Error("Failed to deallocate peer IPs for tenant",
				zap.String("tenant_id", tenant.TenantID),
				zap.Error(err))
		} else {
			rm.logger.Info("Deallocated peer IPs for tenant",
				zap.String("tenant_id", tenant.TenantID),
				zap.Int64("count", count))
		}
	}

	// Закрываем и удаляем P2P соединения
	tenant.Mutex.Lock()
	for _, conn := range tenant.Connections {
		if conn.QUICConn != nil {
			conn.QUICConn.CloseWithError(0, "tenant deactivated")
		}
	}
	tenant.Connections = make(map[string]*RelayP2PConnection)
	tenant.Mutex.Unlock()

	// Закрываем и удаляем глобальные соединения
	for id, conn := range rm.connections {
		if conn.TenantID == tenant.TenantID {
			if conn.QUICConn != nil {
				conn.QUICConn.CloseWithError(0, "tenant deactivated")
			}
			delete(rm.connections, id)
		}
	}

	// Закрываем и удаляем QUIC соединения
	for peerID, quicConn := range rm.quicConnections {
		if quicConn.TenantID == tenant.TenantID {
			quicConn.Conn.CloseWithError(0, "tenant deactivated")
			delete(rm.quicConnections, peerID)
		}
	}
}

// AllocatePeerIP выдаёт IP из подсети тенанта
func (rm *RelayManagerHybrid) AllocatePeerIP(tenantID, peerID string) (string, error) {
	cidr := rm.GetTenantNetwork(tenantID)
	if cidr == "" {
		return "", fmt.Errorf("tenant %s has no network assigned", tenantID)
	}
	if rm.peerIPAMRepo == nil {
		return "", fmt.Errorf("peer IPAM repository not available")
	}
	ip, err := rm.peerIPAMRepo.AllocateIPForPeer(tenantID, peerID, cidr)
	if err != nil {
		return "", err
	}
	rm.logger.Info("Peer IP allocated",
		zap.String("tenant_id", tenantID),
		zap.String("peer_id", peerID),
		zap.String("ip", ip),
	)
	return ip, nil
}

// ReleasePeerIP освобождает IP peer'а
func (rm *RelayManagerHybrid) ReleasePeerIP(peerID string) error {
	if rm.peerIPAMRepo == nil {
		return fmt.Errorf("peer IPAM repository not available")
	}
	return rm.peerIPAMRepo.DeallocateIPForPeer(peerID)
}

// GetTenantNetwork возвращает выделенную сеть для тенанта
func (rm *RelayManagerHybrid) GetTenantNetwork(tenantID string) string {
	rm.cacheMutex.RLock()
	if tn, ok := rm.configCache[tenantID]; ok {
		rm.cacheMutex.RUnlock()
		return tn.NetworkCIDR
	}
	rm.cacheMutex.RUnlock()

	// Если нет в кэше, запрашиваем из хранилища
	network, err := rm.networkStore.Get(rm.ctx, tenantID)
	if err != nil {
		rm.logger.Error("Failed to get tenant network from store",
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		rm.syncStatus = "stale"
		// Безопасный дефолт: возвращаем пустую строку
		return ""
	}

	// Обновляем кэш
	rm.cacheMutex.Lock()
	rm.configCache[tenantID] = network
	rm.cacheMutex.Unlock()

	rm.syncStatus = "fresh"
	return network.NetworkCIDR
}

// CreateTenantNetwork создает новую сеть для тенанта через аллокатор
func (rm *RelayManagerHybrid) CreateTenantNetwork(tenantID string) (string, error) {
	if rm.networkAllocator == nil {
		return "", fmt.Errorf("network allocator not available")
	}

	network, err := rm.networkAllocator.ReserveSubnet(rm.ctx, tenantID, rm.resourceVersion)
	if err != nil {
		return "", fmt.Errorf("failed to reserve subnet: %w", err)
	}

	// Обновляем кэш
	rm.cacheMutex.Lock()
	rm.configCache[tenantID] = network
	rm.cacheMutex.Unlock()

	// Создаем Tenant объект
	tenant := &Tenant{
		TenantID:       tenantID,
		NetworkCIDR:    network.NetworkCIDR,
		Peers:          make(map[string]*Peer),
		Connections:    make(map[string]*RelayP2PConnection),
		MaxPeers:       100,
		MaxConnections: 1000,
		CreatedAt:      time.Unix(network.AllocatedAtUnix, 0),
		LastActivity:   time.Now(),
	}

	rm.mutex.Lock()
	rm.tenants[tenantID] = tenant
	rm.mutex.Unlock()

	rm.logger.Info("Created tenant network through allocator",
		zap.String("tenant_id", tenantID),
		zap.String("network_cidr", network.NetworkCIDR))

	return network.NetworkCIDR, nil
}

// DeactivateTenantNetwork деактивирует сеть тенанта через аллокатор
func (rm *RelayManagerHybrid) DeactivateTenantNetwork(tenantID string) error {
	if rm.networkAllocator == nil {
		return fmt.Errorf("network allocator not available")
	}

	newVersion, err := rm.networkAllocator.ReleaseSubnet(rm.ctx, tenantID, rm.resourceVersion)
	if err != nil {
		return fmt.Errorf("failed to release subnet: %w", err)
	}

	// Обновляем resource version
	rm.resourceVersion = newVersion

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

	// Освобождаем все IP пиров тенанта
	if rm.peerIPAMRepo != nil {
		if count, err := rm.peerIPAMRepo.DeallocateIPsByTenant(rm.ctx, tenantID); err != nil {
			rm.logger.Warn("Failed to batch deallocate tenant IPs",
				zap.String("tenant_id", tenantID), zap.Error(err))
		} else {
			rm.logger.Info("Batch deallocated tenant IPs",
				zap.String("tenant_id", tenantID), zap.Int64("count", count))
		}
	}

	rm.logger.Info("Deactivated tenant network through allocator",
		zap.String("tenant_id", tenantID),
		zap.String("new_version", string(newVersion)))

	return nil
}

// GetTenantInfo возвращает информацию о тенанте
func (rm *RelayManagerHybrid) GetTenantInfo(tenantID string) (*Tenant, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	tenant, exists := rm.tenants[tenantID]
	if !exists {
		return nil, fmt.Errorf("tenant not found: %s", tenantID)
	}

	return tenant, nil
}

// GetAllTenants возвращает все активные тенанты (read-only)
func (rm *RelayManagerHybrid) GetAllTenants() map[string]*Tenant {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	// Возвращаем копию для безопасности
	tenants := make(map[string]*Tenant)
	for id, tenant := range rm.tenants {
		tenants[id] = tenant
	}

	return tenants
}

// GetSyncStatus возвращает статус синхронизации
func (rm *RelayManagerHybrid) GetSyncStatus() string {
	rm.cacheMutex.RLock()
	defer rm.cacheMutex.RUnlock()
	return rm.syncStatus
}

// GetResourceVersion возвращает текущую resource version
func (rm *RelayManagerHybrid) GetResourceVersion() network.ResourceVersion {
	rm.cacheMutex.RLock()
	defer rm.cacheMutex.RUnlock()
	return rm.resourceVersion
}

// GetPeerIPAMRepo возвращает PeerIPAMRepository для доступа к IPAM
func (rm *RelayManagerHybrid) GetPeerIPAMRepo() *database.PeerIPAMRepository {
	return rm.peerIPAMRepo
}

// GetTenantStats возвращает статистику тенанта
func (rm *RelayManagerHybrid) GetTenantStats(tenantID string) map[string]interface{} {
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
		"sync_status":      rm.syncStatus,
		"resource_version": string(rm.resourceVersion),
	}
}

// Shutdown корректно завершает работу RelayManager
func (rm *RelayManagerHybrid) Shutdown() error {
	rm.logger.Info("Shutting down RelayManager with hybrid architecture")

	// Отменяем контекст
	rm.cancel()

	// Закрываем все соединения
	rm.mutex.Lock()
	for _, tenant := range rm.tenants {
		rm.deactivateTenantConnections(tenant)
	}
	rm.mutex.Unlock()

	// Закрываем хранилище
	if rm.networkStore != nil {
		rm.networkStore.Close()
	}

	rm.logger.Info("RelayManager shutdown completed")
	return nil
}
