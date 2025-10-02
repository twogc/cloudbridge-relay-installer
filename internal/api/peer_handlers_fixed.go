package api

import (
	"crypto/sha256"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

// PeerManagerFixed исправленная версия с правильной изоляцией IP адресов
type PeerManagerFixed struct {
	allocatedIPs  map[string]string            // peerID -> allocatedIP
	tenantSubnets map[string]string            // tenantID -> subnetCIDR
	tenantIPs     map[string]map[string]string // tenantID -> peerID -> IP
	ipMutex       sync.RWMutex
	tenantMutex   sync.RWMutex
	logger        *zap.Logger
}

// NewPeerManagerFixed создает новый PeerManager с правильной изоляцией
func NewPeerManagerFixed(logger *zap.Logger) *PeerManagerFixed {
	return &PeerManagerFixed{
		allocatedIPs:  make(map[string]string),
		tenantSubnets: make(map[string]string),
		tenantIPs:     make(map[string]map[string]string),
		logger:        logger,
	}
}

// allocateIPForTenantFixed исправленная версия с изоляцией тенантов
func (pm *PeerManagerFixed) allocateIPForTenantFixed(tenantID, peerID string) (string, error) {
	pm.ipMutex.Lock()
	defer pm.ipMutex.Unlock()

	// Проверяем, есть ли уже выделенный IP для этого пира
	if allocatedIP, exists := pm.allocatedIPs[peerID]; exists {
		return allocatedIP, nil
	}

	// Получаем или создаем подсеть для тенанта
	tenantSubnet, err := pm.getOrCreateTenantSubnet(tenantID)
	if err != nil {
		return "", fmt.Errorf("failed to get tenant subnet: %w", err)
	}

	// Выделяем IP в рамках подсети тенанта
	allocatedIP, err := pm.allocateIPInTenantSubnet(tenantID, peerID, tenantSubnet)
	if err != nil {
		return "", fmt.Errorf("failed to allocate IP in tenant subnet: %w", err)
	}

	pm.allocatedIPs[peerID] = allocatedIP

	// Инициализируем мапу для тенанта если нужно
	if pm.tenantIPs[tenantID] == nil {
		pm.tenantIPs[tenantID] = make(map[string]string)
	}
	pm.tenantIPs[tenantID][peerID] = allocatedIP

	pm.logger.Info("Allocated IP for peer with tenant isolation",
		zap.String("peer_id", peerID),
		zap.String("tenant_id", tenantID),
		zap.String("allocated_ip", allocatedIP),
		zap.String("tenant_subnet", tenantSubnet))

	return allocatedIP, nil
}

// getOrCreateTenantSubnet получает или создает подсеть для тенанта
func (pm *PeerManagerFixed) getOrCreateTenantSubnet(tenantID string) (string, error) {
	pm.tenantMutex.Lock()
	defer pm.tenantMutex.Unlock()

	// Проверяем, есть ли уже подсеть для этого тенанта
	if subnet, exists := pm.tenantSubnets[tenantID]; exists {
		return subnet, nil
	}

	// Создаем новую подсеть для тенанта
	subnet, err := pm.createTenantSubnet(tenantID)
	if err != nil {
		return "", fmt.Errorf("failed to create tenant subnet: %w", err)
	}

	pm.tenantSubnets[tenantID] = subnet
	pm.logger.Info("Created new subnet for tenant",
		zap.String("tenant_id", tenantID),
		zap.String("subnet", subnet))

	return subnet, nil
}

// createTenantSubnet создает уникальную подсеть для тенанта
func (pm *PeerManagerFixed) createTenantSubnet(tenantID string) (string, error) {
	// Используем детерминированную функцию для выделения подсетей
	// Это гарантирует, что один и тот же тенант всегда получит одну и ту же подсеть

	hash := sha256.Sum256([]byte(tenantID))
	subnetID := int(hash[0])<<8 + int(hash[1]) // 0-65535

	// Генерируем подсеть в диапазоне 10.0.0.0/8
	// Используем формат 10.x.y.0/24 где x,y определяются из hash
	x := (subnetID / 256) % 256
	y := subnetID % 256

	subnet := fmt.Sprintf("10.%d.%d.0/24", x, y)

	pm.logger.Info("Generated subnet for tenant",
		zap.String("tenant_id", tenantID),
		zap.String("subnet", subnet),
		zap.Int("subnet_id", subnetID))

	return subnet, nil
}

// allocateIPInTenantSubnet выделяет IP в рамках подсети тенанта
func (pm *PeerManagerFixed) allocateIPInTenantSubnet(tenantID, peerID, subnet string) (string, error) {
	// Парсим подсеть (например, "10.100.1.0/24")
	var x, y, z int
	if _, err := fmt.Sscanf(subnet, "10.%d.%d.%d/24", &x, &y, &z); err != nil {
		return "", fmt.Errorf("invalid subnet format: %s", subnet)
	}
	baseIP := fmt.Sprintf("10.%d.%d", x, y)

	// Находим следующий доступный IP в подсети тенанта
	for i := 10; i <= 254; i++ {
		candidateIP := fmt.Sprintf("%s.%d", baseIP, i)

		// Проверяем, не занят ли этот IP в рамках тенанта
		if pm.isIPAvailableInTenant(tenantID, candidateIP) {
			return candidateIP, nil
		}
	}

	return "", fmt.Errorf("no available IP addresses in tenant subnet %s", subnet)
}

// isIPAvailableInTenant проверяет, доступен ли IP в рамках тенанта
func (pm *PeerManagerFixed) isIPAvailableInTenant(tenantID, candidateIP string) bool {
	// Проверяем, не занят ли IP в рамках этого тенанта
	if tenantIPs, exists := pm.tenantIPs[tenantID]; exists {
		for _, allocatedIP := range tenantIPs {
			if allocatedIP == candidateIP {
				return false
			}
		}
	}

	// Проверяем, не занят ли IP в других тенантах
	for otherTenantID, otherTenantIPs := range pm.tenantIPs {
		if otherTenantID == tenantID {
			continue // Пропускаем текущий тенант
		}

		for _, allocatedIP := range otherTenantIPs {
			if allocatedIP == candidateIP {
				pm.logger.Warn("IP collision detected between tenants",
					zap.String("ip", candidateIP),
					zap.String("current_tenant", tenantID),
					zap.String("collision_tenant", otherTenantID))
				return false
			}
		}
	}

	return true
}

// GetTenantSubnet возвращает подсеть тенанта
func (pm *PeerManagerFixed) GetTenantSubnet(tenantID string) (string, error) {
	pm.tenantMutex.RLock()
	defer pm.tenantMutex.RUnlock()

	if subnet, exists := pm.tenantSubnets[tenantID]; exists {
		return subnet, nil
	}

	return "", fmt.Errorf("tenant %s not found", tenantID)
}

// GetTenantIPs возвращает все IP адреса тенанта
func (pm *PeerManagerFixed) GetTenantIPs(tenantID string) (map[string]string, error) {
	pm.ipMutex.RLock()
	defer pm.ipMutex.RUnlock()

	if tenantIPs, exists := pm.tenantIPs[tenantID]; exists {
		// Создаем копию для безопасности
		result := make(map[string]string)
		for peerID, ip := range tenantIPs {
			result[peerID] = ip
		}
		return result, nil
	}

	return nil, fmt.Errorf("tenant %s not found", tenantID)
}

// ValidateTenantIsolation проверяет изоляцию тенантов
func (pm *PeerManagerFixed) ValidateTenantIsolation() error {
	pm.ipMutex.RLock()
	defer pm.ipMutex.RUnlock()

	// Собираем все IP адреса по тенантам
	tenantIPMap := make(map[string][]string)
	for tenantID, tenantIPs := range pm.tenantIPs {
		for _, ip := range tenantIPs {
			tenantIPMap[tenantID] = append(tenantIPMap[tenantID], ip)
		}
	}

	// Проверяем, что IP адреса разных тенантов не пересекаются
	for tenantA, ipsA := range tenantIPMap {
		for tenantB, ipsB := range tenantIPMap {
			if tenantA == tenantB {
				continue
			}

			for _, ipA := range ipsA {
				for _, ipB := range ipsB {
					if ipA == ipB {
						return fmt.Errorf("IP collision detected: %s is used by both tenant %s and tenant %s", ipA, tenantA, tenantB)
					}
				}
			}
		}
	}

	pm.logger.Info("Tenant isolation validation passed",
		zap.Int("tenants", len(tenantIPMap)))

	return nil
}

// GetIsolationStats возвращает статистику изоляции
func (pm *PeerManagerFixed) GetIsolationStats() map[string]interface{} {
	pm.ipMutex.RLock()
	defer pm.ipMutex.RUnlock()

	stats := map[string]interface{}{
		"total_tenants":      len(pm.tenantSubnets),
		"total_allocated":    len(pm.allocatedIPs),
		"tenant_subnets":     make(map[string]string),
		"tenant_peer_counts": make(map[string]int),
	}

	// Добавляем подсети тенантов
	for tenantID, subnet := range pm.tenantSubnets {
		stats["tenant_subnets"].(map[string]string)[tenantID] = subnet
	}

	// Добавляем количество пиров по тенантам
	for tenantID, tenantIPs := range pm.tenantIPs {
		stats["tenant_peer_counts"].(map[string]int)[tenantID] = len(tenantIPs)
	}

	return stats
}
