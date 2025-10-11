package tenant

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	"github.com/twogc/cloudbridge-relay-installer/internal/metrics"
	"github.com/twogc/cloudbridge-relay-installer/internal/quic"
	"github.com/twogc/cloudbridge-relay-installer/internal/tunnel"
	"go.uber.org/zap"
)

// Tenant represents a tenant (company/organization)
type Tenant struct {
	ID             string
	Name           string
	MaxTunnels     int
	MaxConnections int
	RateLimit      float64
	Burst          int
	CreatedAt      time.Time
	LastActivity   time.Time
	mu             sync.RWMutex
}

// TenantLimits represents tenant resource limits
type TenantLimits struct {
	MaxTunnels     int
	MaxConnections int
	RateLimit      float64
	Burst          int
	AllowedIPs     []string
	AllowedPorts   []int
}

// Manager manages all tenants
type Manager struct {
	tenants       map[string]*Tenant
	tenantLimits  map[string]*TenantLimits
	tunnelManager *tunnel.Manager
	logger        *logger.Logger
	metrics       *metrics.Metrics
	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc

	// Enhanced QUIC интеграция
	enhancedTenantManager *quic.TenantManager
	enhancedEnabled       bool
}

// NewManager creates a new tenant manager
func NewManager(logger *logger.Logger, metrics *metrics.Metrics) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	return &Manager{
		tenants:         make(map[string]*Tenant),
		tenantLimits:    make(map[string]*TenantLimits),
		logger:          logger,
		metrics:         metrics,
		ctx:             ctx,
		cancel:          cancel,
		enhancedEnabled: false, // По умолчанию отключено
	}
}

// NewManagerWithEnhanced создает новый tenant manager с Enhanced QUIC поддержкой
func NewManagerWithEnhanced(logger *logger.Logger, metrics *metrics.Metrics) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	m := &Manager{
		tenants:         make(map[string]*Tenant),
		tenantLimits:    make(map[string]*TenantLimits),
		logger:          logger,
		metrics:         metrics,
		ctx:             ctx,
		cancel:          cancel,
		enhancedEnabled: true,
	}

	// Инициализируем Enhanced TenantManager
	m.InitEnhancedTenantManager()

	return m
}

// CreateTenant creates a new tenant
func (m *Manager) CreateTenant(id, name string, limits *TenantLimits) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tenants[id]; exists {
		return fmt.Errorf("tenant %s already exists", id)
	}

	tenant := &Tenant{
		ID:             id,
		Name:           name,
		MaxTunnels:     limits.MaxTunnels,
		MaxConnections: limits.MaxConnections,
		RateLimit:      limits.RateLimit,
		Burst:          limits.Burst,
		CreatedAt:      time.Now(),
		LastActivity:   time.Now(),
	}

	m.tenants[id] = tenant
	m.tenantLimits[id] = limits

	// Добавляем в Enhanced TenantManager, если включен
	if m.enhancedEnabled && m.enhancedTenantManager != nil {
		// Конвертируем лимиты в формат Enhanced QUIC
		bandwidthLimit := int64(limits.MaxConnections * 1024 * 1024) // 1MB на соединение
		m.enhancedTenantManager.AddTenant(id, bandwidthLimit, limits.MaxConnections)
	}

	m.logger.Info("Tenant created",
		zap.String("tenant_id", id),
		zap.String("tenant_name", name),
		zap.Int("max_tunnels", limits.MaxTunnels),
		zap.Int("max_connections", limits.MaxConnections),
		zap.Bool("enhanced_enabled", m.enhancedEnabled),
	)

	return nil
}

// GetTenant returns a tenant by ID
func (m *Manager) GetTenant(id string) (*Tenant, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tenant, exists := m.tenants[id]
	return tenant, exists
}

// GetTenantLimits returns tenant limits
func (m *Manager) GetTenantLimits(id string) (*TenantLimits, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	limits, exists := m.tenantLimits[id]
	return limits, exists
}

// UpdateTenantActivity updates tenant last activity
func (m *Manager) UpdateTenantActivity(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if tenant, exists := m.tenants[id]; exists {
		tenant.LastActivity = time.Now()
	}
}

// UpdateTenantLimits updates tenant limits
func (m *Manager) UpdateTenantLimits(id string, limits *TenantLimits) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tenant, exists := m.tenants[id]
	if !exists {
		return fmt.Errorf("tenant %s not found", id)
	}

	// Update tenant limits
	tenant.MaxTunnels = limits.MaxTunnels
	tenant.MaxConnections = limits.MaxConnections
	tenant.RateLimit = limits.RateLimit
	tenant.Burst = limits.Burst

	// Update limits map
	m.tenantLimits[id] = limits

	// Update Enhanced TenantManager if enabled
	if m.enhancedEnabled && m.enhancedTenantManager != nil {
		// TODO: Implement UpdateTenantLimits in quic.TenantManager
		// For now, we just log that enhanced mode is enabled
		m.logger.Debug("Enhanced mode enabled, limits updated in standard manager only",
			zap.String("tenant_id", id))
	}

	m.logger.Info("Tenant limits updated",
		zap.String("tenant_id", id),
		zap.Int("max_tunnels", limits.MaxTunnels),
		zap.Int("max_connections", limits.MaxConnections),
	)

	return nil
}

// CheckTenantLimits checks if tenant can create more tunnels/connections
func (m *Manager) CheckTenantLimits(tenantID string, currentTunnels, currentConnections int) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tenant, exists := m.tenants[tenantID]
	if !exists {
		return fmt.Errorf("tenant %s not found", tenantID)
	}

	if currentTunnels >= tenant.MaxTunnels {
		return fmt.Errorf("tenant %s tunnel limit exceeded (%d/%d)", tenantID, currentTunnels, tenant.MaxTunnels)
	}

	if currentConnections >= tenant.MaxConnections {
		return fmt.Errorf("tenant %s connection limit exceeded (%d/%d)", tenantID, currentConnections, tenant.MaxConnections)
	}

	return nil
}

// IsPortAllowed checks if a port is allowed for the tenant
func (m *Manager) IsPortAllowed(tenantID string, port int) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	limits, exists := m.tenantLimits[tenantID]
	if !exists {
		return false
	}

	// If no ports specified, allow all
	if len(limits.AllowedPorts) == 0 {
		return true
	}

	for _, allowedPort := range limits.AllowedPorts {
		if allowedPort == port {
			return true
		}
	}

	return false
}

// IsIPAllowed checks if an IP is allowed for the tenant
func (m *Manager) IsIPAllowed(tenantID string, ip string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	limits, exists := m.tenantLimits[tenantID]
	if !exists {
		return false
	}

	// If no IPs specified, allow all
	if len(limits.AllowedIPs) == 0 {
		return true
	}

	// Parse the IP to check
	checkIP := net.ParseIP(ip)
	if checkIP == nil {
		return false
	}

	for _, allowedIP := range limits.AllowedIPs {
		// Check if it's a CIDR range
		if strings.Contains(allowedIP, "/") {
			_, ipNet, err := net.ParseCIDR(allowedIP)
			if err != nil {
				continue
			}
			if ipNet.Contains(checkIP) {
				return true
			}
		} else {
			// Check exact IP match
			if allowedIP == ip {
				return true
			}
		}
	}

	return false
}

// GetTenantStats returns statistics for all tenants
func (m *Manager) GetTenantStats() map[string]map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]map[string]interface{})

	for tenantID, tenant := range m.tenants {
		tenantStats := map[string]interface{}{
			"name":            tenant.Name,
			"max_tunnels":     tenant.MaxTunnels,
			"max_connections": tenant.MaxConnections,
			"rate_limit":      tenant.RateLimit,
			"burst":           tenant.Burst,
			"created_at":      tenant.CreatedAt,
			"last_activity":   tenant.LastActivity,
		}

		// Add tunnel statistics if tunnel manager is available
		if m.tunnelManager != nil {
			// This would need to be implemented to track per-tenant tunnels
			// For now, we'll add placeholder stats
			tenantStats["active_tunnels"] = 0
			tenantStats["total_tunnels"] = 0
		}

		stats[tenantID] = tenantStats
	}

	return stats
}

// CleanupInactiveTenants removes inactive tenants
func (m *Manager) CleanupInactiveTenants(timeout time.Duration) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	removed := 0

	for tenantID, tenant := range m.tenants {
		if now.Sub(tenant.LastActivity) > timeout {
			m.logger.Info("Removing inactive tenant",
				zap.String("tenant_id", tenantID),
				zap.String("tenant_name", tenant.Name),
				zap.Duration("inactive_duration", now.Sub(tenant.LastActivity)),
			)

			delete(m.tenants, tenantID)
			delete(m.tenantLimits, tenantID)
			removed++
		}
	}

	if removed > 0 {
		m.logger.Info("Tenant cleanup completed",
			zap.Int("removed_tenants", removed),
			zap.Int("remaining_tenants", len(m.tenants)),
		)
	}

	return removed
}

// StartCleanupLoop starts the cleanup loop
func (m *Manager) StartCleanupLoop(interval, timeout time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-m.ctx.Done():
				return
			case <-ticker.C:
				m.CleanupInactiveTenants(timeout)
			}
		}
	}()
}

// Shutdown gracefully shuts down the manager
func (m *Manager) Shutdown() {
	m.logger.Info("Shutting down tenant manager")
	m.cancel()
}

// GetTenantCount returns the total number of tenants
func (m *Manager) GetTenantCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.tenants)
}

// SetTunnelManager sets the tunnel manager for tenant-specific tunnel tracking
func (m *Manager) SetTunnelManager(tm *tunnel.Manager) {
	m.tunnelManager = tm
}

// InitEnhancedTenantManager инициализирует Enhanced TenantManager
func (m *Manager) InitEnhancedTenantManager() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.enhancedTenantManager = quic.NewTenantManager()
	m.enhancedEnabled = true

	// Синхронизируем существующих тенантов с Enhanced
	for id, limits := range m.tenantLimits {
		bandwidthLimit := int64(limits.MaxConnections * 1024 * 1024) // 1MB на соединение
		m.enhancedTenantManager.AddTenant(id, bandwidthLimit, limits.MaxConnections)
	}

	m.logger.Info("Enhanced TenantManager initialized",
		zap.Int("synced_tenants", len(m.tenants)),
	)
}

// GetEnhancedTenantManager возвращает Enhanced TenantManager
func (m *Manager) GetEnhancedTenantManager() *quic.TenantManager {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.enhancedTenantManager
}

// IsEnhancedEnabled проверяет, включен ли Enhanced режим
func (m *Manager) IsEnhancedEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.enhancedEnabled
}

// EnableEnhanced включает Enhanced режим
func (m *Manager) EnableEnhanced() {
	if !m.enhancedEnabled {
		m.InitEnhancedTenantManager()
	}
}

// DisableEnhanced отключает Enhanced режим
func (m *Manager) DisableEnhanced() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.enhancedEnabled = false
	m.enhancedTenantManager = nil

	m.logger.Info("Enhanced mode disabled")
}

// CheckEnhancedTenantLimits проверяет лимиты через Enhanced TenantManager
func (m *Manager) CheckEnhancedTenantLimits(tenantID string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.enhancedEnabled || m.enhancedTenantManager == nil {
		// Fallback к обычной проверке
		return m.CheckTenantLimits(tenantID, 0, 0)
	}

	return m.enhancedTenantManager.CheckTenantLimits(tenantID)
}
