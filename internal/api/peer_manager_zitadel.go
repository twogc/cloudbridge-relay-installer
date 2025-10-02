package api

import (
	"context"
	"fmt"
	"net"
	"sync"

	"go.uber.org/zap"
	"github.com/twogc/cloudbridge-relay-installer/internal/zitadel"
	"github.com/twogc/cloudbridge-relay-installer/internal/p2p"
)

// ExtendedPeerStore - расширенный PeerStore для работы с IP адресами
type ExtendedPeerStore struct {
	peers        map[string]map[string]*p2p.Peer // tenantID -> peerID -> peer
	peerIPs      map[string]string               // peerID -> IP
	peerTenants  map[string]string               // peerID -> tenantID
	tenantIPs    map[string]map[string]string    // tenantID -> peerID -> IP
	mutex        sync.RWMutex
}

// NewExtendedPeerStore создает новый расширенный PeerStore
func NewExtendedPeerStore() *ExtendedPeerStore {
	return &ExtendedPeerStore{
		peers:       make(map[string]map[string]*p2p.Peer),
		peerIPs:     make(map[string]string),
		peerTenants: make(map[string]string),
		tenantIPs:   make(map[string]map[string]string),
	}
}

// AddPeer добавляет пира с IP адресом
func (eps *ExtendedPeerStore) AddPeer(peerID, tenantID, ip string) {
	eps.mutex.Lock()
	defer eps.mutex.Unlock()
	
	// Инициализируем мапу для тенанта если нужно
	if eps.peers[tenantID] == nil {
		eps.peers[tenantID] = make(map[string]*p2p.Peer)
	}
	if eps.tenantIPs[tenantID] == nil {
		eps.tenantIPs[tenantID] = make(map[string]string)
	}
	
	// Сохраняем данные
	eps.peerIPs[peerID] = ip
	eps.peerTenants[peerID] = tenantID
	eps.tenantIPs[tenantID][peerID] = ip
}

// GetPeerIP получает IP адрес пира
func (eps *ExtendedPeerStore) GetPeerIP(peerID string) (string, bool) {
	eps.mutex.RLock()
	defer eps.mutex.RUnlock()
	
	ip, exists := eps.peerIPs[peerID]
	return ip, exists
}

// GetPeerTenantID получает tenantID пира
func (eps *ExtendedPeerStore) GetPeerTenantID(peerID string) (string, bool) {
	eps.mutex.RLock()
	defer eps.mutex.RUnlock()
	
	tenantID, exists := eps.peerTenants[peerID]
	return tenantID, exists
}

// GetTenantIPs получает все IP для тенанта
func (eps *ExtendedPeerStore) GetTenantIPs(tenantID string) map[string]string {
	eps.mutex.RLock()
	defer eps.mutex.RUnlock()
	
	if tenantIPs, exists := eps.tenantIPs[tenantID]; exists {
		// Возвращаем копию
		result := make(map[string]string)
		for peerID, ip := range tenantIPs {
			result[peerID] = ip
		}
		return result
	}
	
	return make(map[string]string)
}

// RemovePeer удаляет пира
func (eps *ExtendedPeerStore) RemovePeer(peerID string) {
	eps.mutex.Lock()
	defer eps.mutex.Unlock()
	
	tenantID, exists := eps.peerTenants[peerID]
	if !exists {
		return
	}
	
	// Удаляем из всех мап
	delete(eps.peerIPs, peerID)
	delete(eps.peerTenants, peerID)
	if eps.tenantIPs[tenantID] != nil {
		delete(eps.tenantIPs[tenantID], peerID)
	}
	if eps.peers[tenantID] != nil {
		delete(eps.peers[tenantID], peerID)
	}
}

// GetAllTenants получает все тенанты
func (eps *ExtendedPeerStore) GetAllTenants() map[string]bool {
	eps.mutex.RLock()
	defer eps.mutex.RUnlock()
	
	tenants := make(map[string]bool)
	for tenantID := range eps.peers {
		tenants[tenantID] = true
	}
	
	return tenants
}

// GetStats получает статистику
func (eps *ExtendedPeerStore) GetStats() map[string]interface{} {
	eps.mutex.RLock()
	defer eps.mutex.RUnlock()
	
	stats := map[string]interface{}{
		"total_peers":    len(eps.peerIPs),
		"total_tenants":  len(eps.peers),
		"tenant_counts":  make(map[string]int),
	}
	
	for tenantID, tenantIPs := range eps.tenantIPs {
		stats["tenant_counts"].(map[string]int)[tenantID] = len(tenantIPs)
	}
	
	return stats
}

// TenantNetworkClientInterface - интерфейс для TenantNetworkClient
type TenantNetworkClientInterface interface {
	GetTenantNetworkConfig(ctx context.Context, tenantID string) (*zitadel.TenantNetworkConfig, error)
	CreateTenantNetworkConfig(ctx context.Context, tenantID string, config *zitadel.TenantNetworkConfig) error
	UpdateTenantNetworkConfig(ctx context.Context, tenantID string, config *zitadel.TenantNetworkConfig) error
	AllocateIPForTenant(ctx context.Context, tenantID, peerID, allocatedIP string) error
	DeallocateIPForTenant(ctx context.Context, tenantID, peerID string) error
	GetIPAMPools(ctx context.Context) ([]*zitadel.IPAMPool, error)
	AllocateSubnetFromPool(ctx context.Context, poolID string) (string, error)
	ValidateTenantIsolation(ctx context.Context) (map[string]interface{}, error)
}

// PeerManagerZitadel - PeerManager с интеграцией Zitadel для управления подсетями
type PeerManagerZitadel struct {
	zitadelClient TenantNetworkClientInterface
	ipMutex       sync.RWMutex
	logger        *zap.Logger
	peerStore     *ExtendedPeerStore
}

// NewPeerManagerZitadel создает новый PeerManager с интеграцией Zitadel
func NewPeerManagerZitadel(zitadelClient TenantNetworkClientInterface, logger *zap.Logger) *PeerManagerZitadel {
	return &PeerManagerZitadel{
		zitadelClient: zitadelClient,
		logger:        logger,
		peerStore:     NewExtendedPeerStore(),
	}
}

// AllocateIPForTenantWithZitadel выделяет IP адрес для пира через Zitadel
func (pm *PeerManagerZitadel) AllocateIPForTenantWithZitadel(ctx context.Context, tenantID, peerID string) (string, error) {
	pm.ipMutex.Lock()
	defer pm.ipMutex.Unlock()

	// Проверяем, есть ли уже выделенный IP для этого пира
	if allocatedIP, exists := pm.peerStore.GetPeerIP(peerID); exists {
		return allocatedIP, nil
	}

	// Получаем или создаем сетевую конфигурацию тенанта
	config, err := pm.getOrCreateTenantNetworkConfig(ctx, tenantID)
	if err != nil {
		return "", fmt.Errorf("failed to get or create tenant network config: %w", err)
	}

	// Выделяем IP в рамках подсети тенанта
	allocatedIP, err := pm.allocateIPInTenantSubnet(tenantID, peerID, config.AllocatedSubnet)
	if err != nil {
		return "", fmt.Errorf("failed to allocate IP in tenant subnet: %w", err)
	}

	// Сохраняем IP в Zitadel
	if err := pm.zitadelClient.AllocateIPForTenant(ctx, tenantID, peerID, allocatedIP); err != nil {
		return "", fmt.Errorf("failed to save IP allocation to Zitadel: %w", err)
	}

	// Сохраняем в локальном кэше
	pm.peerStore.AddPeer(peerID, tenantID, allocatedIP)

	pm.logger.Info("Allocated IP for peer with Zitadel integration",
		zap.String("peer_id", peerID),
		zap.String("tenant_id", tenantID),
		zap.String("allocated_ip", allocatedIP),
		zap.String("tenant_subnet", config.AllocatedSubnet))

	return allocatedIP, nil
}

// getOrCreateTenantNetworkConfig получает или создает сетевую конфигурацию тенанта
func (pm *PeerManagerZitadel) getOrCreateTenantNetworkConfig(ctx context.Context, tenantID string) (*zitadel.TenantNetworkConfig, error) {
	// Пытаемся получить существующую конфигурацию
	config, err := pm.zitadelClient.GetTenantNetworkConfig(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant network config: %w", err)
	}

	// Если конфигурация существует, возвращаем её
	if config != nil {
		return config, nil
	}

	// Создаем новую конфигурацию
	pm.logger.Info("Creating new network config for tenant", zap.String("tenant_id", tenantID))

	// Выделяем подсеть из IPAM пула
	subnet, err := pm.allocateTenantSubnet(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate tenant subnet: %w", err)
	}

	// Создаем новую конфигурацию
	newConfig := &zitadel.TenantNetworkConfig{
		AllocatedSubnet: subnet,
		IPAMPool:        "cloudbridge-pool-1", // По умолчанию
		NetworkPolicies: []string{"allow-intra-tenant", "deny-cross-tenant"},
		AllocatedIPs:    make(map[string]string),
	}

	// Сохраняем в Zitadel
	if err := pm.zitadelClient.CreateTenantNetworkConfig(ctx, tenantID, newConfig); err != nil {
		return nil, fmt.Errorf("failed to create tenant network config: %w", err)
	}

	pm.logger.Info("Created new network config for tenant",
		zap.String("tenant_id", tenantID),
		zap.String("subnet", subnet))

	return newConfig, nil
}

// allocateTenantSubnet выделяет подсеть для тенанта
func (pm *PeerManagerZitadel) allocateTenantSubnet(ctx context.Context, tenantID string) (string, error) {
	// Пытаемся выделить подсеть из IPAM пула
	subnet, err := pm.zitadelClient.AllocateSubnetFromPool(ctx, "cloudbridge-pool-1")
	if err != nil {
		// Если IPAM пул недоступен, используем детерминированную генерацию
		pm.logger.Warn("Failed to allocate subnet from IPAM pool, using deterministic generation",
			zap.String("tenant_id", tenantID),
			zap.Error(err))

		return pm.generateDeterministicSubnet(tenantID), nil
	}

	return subnet, nil
}

// generateDeterministicSubnet генерирует детерминированную подсеть для тенанта
func (pm *PeerManagerZitadel) generateDeterministicSubnet(tenantID string) string {
	// Используем детерминированную функцию на основе хеша tenantID
	hash := 0
	for _, r := range tenantID {
		hash = (hash*31 + int(r)) % 65536 // 2^16 = 65536 возможных подсетей /24
	}

	// Распределяем подсети в диапазоне 10.0.0.0/8
	// Например, 10.X.Y.0/24
	x := hash / 256
	y := hash % 256

	subnet := fmt.Sprintf("10.%d.%d.0/24", x, y)
	
	pm.logger.Info("Generated deterministic subnet for tenant",
		zap.String("tenant_id", tenantID),
		zap.String("subnet", subnet),
		zap.Int("subnet_id", hash))

	return subnet
}

// allocateIPInTenantSubnet выделяет IP в рамках подсети тенанта
func (pm *PeerManagerZitadel) allocateIPInTenantSubnet(tenantID, peerID, subnet string) (string, error) {
	// Парсим подсеть (например, "10.100.1.0/24")
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return "", fmt.Errorf("invalid subnet format: %s", subnet)
	}

	// Получаем базовый IP адрес подсети
	baseIP := ipNet.IP.To4()
	if baseIP == nil {
		return "", fmt.Errorf("invalid IPv4 subnet: %s", subnet)
	}

	// Получаем существующие IP для тенанта
	existingIPs := pm.peerStore.GetTenantIPs(tenantID)

	// Находим следующий доступный IP в подсети тенанта
	// Начинаем с .10, чтобы избежать конфликтов с шлюзами и т.д.
	for i := 10; i <= 254; i++ {
		candidateIP := net.IPv4(baseIP[0], baseIP[1], baseIP[2], byte(i)).String()

		// Проверяем, не занят ли этот IP
		alreadyAllocated := false
		for _, allocated := range existingIPs {
			if allocated == candidateIP {
				alreadyAllocated = true
				break
			}
		}

		if !alreadyAllocated {
			return candidateIP, nil
		}
	}

	return "", fmt.Errorf("no available IP addresses in subnet %s for tenant %s", subnet, tenantID)
}

// DeallocateIP освобождает IP адрес пира
func (pm *PeerManagerZitadel) DeallocateIP(ctx context.Context, peerID string) error {
	pm.ipMutex.Lock()
	defer pm.ipMutex.Unlock()

	// Получаем информацию о пире
	tenantID, exists := pm.peerStore.GetPeerTenantID(peerID)
	if !exists {
		pm.logger.Warn("Peer not found for deallocation", zap.String("peer_id", peerID))
		return nil
	}

	// Удаляем из Zitadel
	if err := pm.zitadelClient.DeallocateIPForTenant(ctx, tenantID, peerID); err != nil {
		pm.logger.Error("Failed to deallocate IP from Zitadel",
			zap.String("peer_id", peerID),
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		// Продолжаем выполнение, так как локальное удаление все равно нужно
	}

	// Удаляем из локального кэша
	pm.peerStore.RemovePeer(peerID)

	pm.logger.Info("Deallocated IP for peer",
		zap.String("peer_id", peerID),
		zap.String("tenant_id", tenantID))

	return nil
}

// ValidateTenantIsolation проверяет изоляцию тенантов через Zitadel
func (pm *PeerManagerZitadel) ValidateTenantIsolation(ctx context.Context) (map[string]interface{}, error) {
	// Получаем статистику из Zitadel
	stats, err := pm.zitadelClient.ValidateTenantIsolation(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to validate tenant isolation in Zitadel: %w", err)
	}

	// Добавляем локальную статистику
	localStats := pm.peerStore.GetStats()
	stats["local_cache"] = localStats

	pm.logger.Info("Validated tenant isolation with Zitadel integration",
		zap.Int("total_tenants", len(pm.peerStore.GetAllTenants())))

	return stats, nil
}

// GetTenantNetworkConfig получает сетевую конфигурацию тенанта
func (pm *PeerManagerZitadel) GetTenantNetworkConfig(ctx context.Context, tenantID string) (*zitadel.TenantNetworkConfig, error) {
	return pm.zitadelClient.GetTenantNetworkConfig(ctx, tenantID)
}

// GetAllTenantSubnets получает все подсети тенантов
func (pm *PeerManagerZitadel) GetAllTenantSubnets(ctx context.Context) (map[string]string, error) {
	// Получаем все тенанты из локального кэша
	tenants := pm.peerStore.GetAllTenants()
	subnets := make(map[string]string)

	// Запрашиваем конфигурацию для каждого тенанта
	for tenantID := range tenants {
		config, err := pm.zitadelClient.GetTenantNetworkConfig(ctx, tenantID)
		if err != nil {
			pm.logger.Warn("Failed to get network config for tenant",
				zap.String("tenant_id", tenantID),
				zap.Error(err))
			continue
		}
		if config != nil {
			subnets[tenantID] = config.AllocatedSubnet
		}
	}

	return subnets, nil
}

// SyncWithZitadel синхронизирует локальный кэш с Zitadel
func (pm *PeerManagerZitadel) SyncWithZitadel(ctx context.Context) error {
	pm.logger.Info("Starting sync with Zitadel")

	// Получаем все тенанты из локального кэша
	tenants := pm.peerStore.GetAllTenants()

	for tenantID := range tenants {
		// Получаем конфигурацию из Zitadel
		config, err := pm.zitadelClient.GetTenantNetworkConfig(ctx, tenantID)
		if err != nil {
			pm.logger.Warn("Failed to sync tenant config",
				zap.String("tenant_id", tenantID),
				zap.Error(err))
			continue
		}

		if config != nil {
			// Синхронизируем выделенные IP
			for peerID, ip := range config.AllocatedIPs {
				pm.peerStore.AddPeer(peerID, tenantID, ip)
			}
		}
	}

	pm.logger.Info("Completed sync with Zitadel",
		zap.Int("tenants_synced", len(tenants)))

	return nil
}
