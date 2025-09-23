package zitadel

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// TenantManager менеджер тенантов
type TenantManager struct {
	validator *ZitadelValidator
	logger    *zap.Logger

	// Кэш тенантов
	tenantCache map[string]*TenantInfo
	cacheMutex  sync.RWMutex
	cacheExpiry time.Duration
}

// NewTenantManager создает новый менеджер тенантов
func NewTenantManager(validator *ZitadelValidator, logger *zap.Logger) *TenantManager {
	return &TenantManager{
		validator:   validator,
		logger:      logger,
		tenantCache: make(map[string]*TenantInfo),
		cacheExpiry: 5 * time.Minute,
	}
}

// GetTenant получает информацию о тенанте
func (tm *TenantManager) GetTenant(ctx context.Context, tenantID string) (*TenantInfo, error) {
	// Проверка кэша
	tm.cacheMutex.RLock()
	if tenant, exists := tm.tenantCache[tenantID]; exists {
		tm.cacheMutex.RUnlock()
		return tenant, nil
	}
	tm.cacheMutex.RUnlock()

	// Создание базовой информации о тенанте
	tenant := &TenantInfo{
		ID:        tenantID,
		Name:      fmt.Sprintf("Tenant %s", tenantID),
		Domain:    fmt.Sprintf("%s.local", tenantID),
		State:     "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),

		// Значения по умолчанию
		MaxPeers:       10,
		MaxConnections: 100,
		MaxTunnels:     20,
		BandwidthLimit: "1GB",
		StorageLimit:   "10GB",
	}

	// Сохранение в кэш
	tm.cacheMutex.Lock()
	tm.tenantCache[tenantID] = tenant
	tm.cacheMutex.Unlock()

	tm.logger.Debug("Created tenant info",
		zap.String("tenant_id", tenantID),
		zap.String("name", tenant.Name),
	)

	return tenant, nil
}

// CreateTenant создает новый тенант
func (tm *TenantManager) CreateTenant(ctx context.Context, name, domain string, config map[string]interface{}) (*TenantInfo, error) {
	tenantID := fmt.Sprintf("tenant-%d", time.Now().Unix())

	// Создание информации о тенанте
	tenant := &TenantInfo{
		ID:        tenantID,
		Name:      name,
		Domain:    domain,
		State:     "active",
		Metadata:  config,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),

		// Значения по умолчанию или из конфигурации
		MaxPeers:       getIntFromConfig(config, "max_peers", 10),
		MaxConnections: getIntFromConfig(config, "max_connections", 100),
		MaxTunnels:     getIntFromConfig(config, "max_tunnels", 20),
		BandwidthLimit: getStringFromConfig(config, "bandwidth_limit", "1GB"),
		StorageLimit:   getStringFromConfig(config, "storage_limit", "10GB"),
	}

	// Сохранение в кэш
	tm.cacheMutex.Lock()
	tm.tenantCache[tenantID] = tenant
	tm.cacheMutex.Unlock()

	tm.logger.Info("Created new tenant",
		zap.String("tenant_id", tenantID),
		zap.String("name", name),
		zap.String("domain", domain),
	)

	return tenant, nil
}

// UpdateTenant обновляет информацию о тенанте
func (tm *TenantManager) UpdateTenant(ctx context.Context, tenantID string, updates map[string]interface{}) (*TenantInfo, error) {
	// Получение текущей информации
	tenant, err := tm.GetTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// Обновление полей
	if name, ok := updates["name"].(string); ok {
		tenant.Name = name
	}

	if domain, ok := updates["domain"].(string); ok {
		tenant.Domain = domain
	}

	if maxPeers, ok := updates["max_peers"].(int); ok {
		tenant.MaxPeers = maxPeers
	}

	if maxConnections, ok := updates["max_connections"].(int); ok {
		tenant.MaxConnections = maxConnections
	}

	if maxTunnels, ok := updates["max_tunnels"].(int); ok {
		tenant.MaxTunnels = maxTunnels
	}

	if bandwidthLimit, ok := updates["bandwidth_limit"].(string); ok {
		tenant.BandwidthLimit = bandwidthLimit
	}

	if storageLimit, ok := updates["storage_limit"].(string); ok {
		tenant.StorageLimit = storageLimit
	}

	tenant.UpdatedAt = time.Now()

	// Обновление кэша
	tm.cacheMutex.Lock()
	tm.tenantCache[tenantID] = tenant
	tm.cacheMutex.Unlock()

	tm.logger.Info("Updated tenant",
		zap.String("tenant_id", tenantID),
		zap.Any("updates", updates),
	)

	return tenant, nil
}

// DeleteTenant удаляет тенант
func (tm *TenantManager) DeleteTenant(ctx context.Context, tenantID string) error {
	// Удаление из кэша
	tm.cacheMutex.Lock()
	delete(tm.tenantCache, tenantID)
	tm.cacheMutex.Unlock()

	tm.logger.Info("Deleted tenant", zap.String("tenant_id", tenantID))
	return nil
}

// ListTenants возвращает список всех тенантов
func (tm *TenantManager) ListTenants(ctx context.Context) ([]*TenantInfo, error) {
	tm.cacheMutex.RLock()
	defer tm.cacheMutex.RUnlock()

	var tenants []*TenantInfo
	for _, tenant := range tm.tenantCache {
		tenants = append(tenants, tenant)
	}

	return tenants, nil
}

// ValidateTenantAccess проверяет доступ к тенанту
func (tm *TenantManager) ValidateTenantAccess(ctx context.Context, tenantID, userID string) (bool, error) {
	// Упрощенная проверка - всегда разрешаем доступ
	// TODO: Реализовать через Zitadel API
	return true, nil
}

// GetTenantQuotas возвращает квоты тенанта
func (tm *TenantManager) GetTenantQuotas(ctx context.Context, tenantID string) (map[string]interface{}, error) {
	tenant, err := tm.GetTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"max_peers":       tenant.MaxPeers,
		"max_connections": tenant.MaxConnections,
		"max_tunnels":     tenant.MaxTunnels,
		"bandwidth_limit": tenant.BandwidthLimit,
		"storage_limit":   tenant.StorageLimit,
	}, nil
}

// UpdateTenantQuotas обновляет квоты тенанта
func (tm *TenantManager) UpdateTenantQuotas(ctx context.Context, tenantID string, quotas map[string]interface{}) error {
	_, err := tm.UpdateTenant(ctx, tenantID, quotas)
	return err
}

// ClearCache очищает кэш тенантов
func (tm *TenantManager) ClearCache() {
	tm.cacheMutex.Lock()
	defer tm.cacheMutex.Unlock()

	tm.tenantCache = make(map[string]*TenantInfo)
	tm.logger.Info("Tenant cache cleared")
}

// GetCacheStats возвращает статистику кэша
func (tm *TenantManager) GetCacheStats() map[string]interface{} {
	tm.cacheMutex.RLock()
	defer tm.cacheMutex.RUnlock()

	return map[string]interface{}{
		"cached_tenants": len(tm.tenantCache),
		"cache_expiry":   tm.cacheExpiry.String(),
	}
}
