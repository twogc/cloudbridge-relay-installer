package zitadel

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// TenantManager управляет тенантами через Zitadel организации
type TenantManager struct {
	client *Client
	logger *zap.Logger

	// Cache for tenant information
	tenants    map[string]*Tenant
	tenantsMux sync.RWMutex

	// Configuration
	autoCreate    bool
	defaultQuotas *TenantQuotas
	cacheExpiry   time.Duration
}

// Tenant представляет тенант (организацию в Zitadel)
type Tenant struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	DisplayName string                 `json:"display_name"`
	Domain      string                 `json:"domain"`
	State       TenantState            `json:"state"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Quotas      *TenantQuotas          `json:"quotas"`
	Settings    *TenantSettings        `json:"settings"`
	Metadata    map[string]interface{} `json:"metadata"`

	// CloudBridge specific
	RelayEndpoints []string `json:"relay_endpoints"`
	ActivePeers    int      `json:"active_peers"`
	TotalTraffic   int64    `json:"total_traffic"`
}

// TenantState состояние тенанта
type TenantState int

const (
	TenantStateActive TenantState = iota
	TenantStateInactive
	TenantStateSuspended
	TenantStatePendingActivation
)

func (ts TenantState) String() string {
	switch ts {
	case TenantStateActive:
		return "active"
	case TenantStateInactive:
		return "inactive"
	case TenantStateSuspended:
		return "suspended"
	case TenantStatePendingActivation:
		return "pending_activation"
	default:
		return "unknown"
	}
}

// TenantQuotas квоты для тенанта
type TenantQuotas struct {
	MaxPeers       int    `json:"max_peers"`
	MaxConnections int    `json:"max_connections"`
	MaxTunnels     int    `json:"max_tunnels"`
	BandwidthLimit string `json:"bandwidth_limit"`
	StorageLimit   string `json:"storage_limit"`

	// Usage tracking
	CurrentPeers       int   `json:"current_peers"`
	CurrentConnections int   `json:"current_connections"`
	CurrentTunnels     int   `json:"current_tunnels"`
	BandwidthUsed      int64 `json:"bandwidth_used"`
	StorageUsed        int64 `json:"storage_used"`
}

// TenantSettings настройки тенанта
type TenantSettings struct {
	EnableP2P          bool     `json:"enable_p2p"`
	EnableMASQUE       bool     `json:"enable_masque"`
	EnableICE          bool     `json:"enable_ice"`
	AllowedProtocols   []string `json:"allowed_protocols"`
	SecurityLevel      string   `json:"security_level"`
	LoggingLevel       string   `json:"logging_level"`
	SessionTimeout     string   `json:"session_timeout"`
	MaxSessionDuration string   `json:"max_session_duration"`
}

// TenantCreateRequest запрос на создание тенанта
type TenantCreateRequest struct {
	Name        string                 `json:"name"`
	DisplayName string                 `json:"display_name"`
	Domain      string                 `json:"domain"`
	AdminEmail  string                 `json:"admin_email"`
	Quotas      *TenantQuotas          `json:"quotas,omitempty"`
	Settings    *TenantSettings        `json:"settings,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// NewTenantManager создает новый менеджер тенантов
func NewTenantManager(client *Client, logger *zap.Logger) *TenantManager {
	return &TenantManager{
		client:      client,
		logger:      logger,
		tenants:     make(map[string]*Tenant),
		autoCreate:  false,
		cacheExpiry: 10 * time.Minute,
		defaultQuotas: &TenantQuotas{
			MaxPeers:       10,
			MaxConnections: 100,
			MaxTunnels:     20,
			BandwidthLimit: "1GB",
			StorageLimit:   "10GB",
		},
	}
}

// CreateTenant создает новый тенант (организацию в Zitadel)
func (tm *TenantManager) CreateTenant(ctx context.Context, req *TenantCreateRequest) (*Tenant, error) {
	tm.logger.Info("Creating new tenant",
		zap.String("name", req.Name),
		zap.String("domain", req.Domain))

	// В полной реализации здесь был бы вызов Zitadel Management API
	// для создания новой организации
	tenant := &Tenant{
		ID:          fmt.Sprintf("tenant_%d", time.Now().Unix()),
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Domain:      req.Domain,
		State:       TenantStatePendingActivation,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Quotas:      req.Quotas,
		Settings:    req.Settings,
		Metadata:    req.Metadata,
	}

	// Apply default quotas if not specified
	if tenant.Quotas == nil {
		tenant.Quotas = tm.defaultQuotas
	}

	// Apply default settings if not specified
	if tenant.Settings == nil {
		tenant.Settings = &TenantSettings{
			EnableP2P:          true,
			EnableMASQUE:       true,
			EnableICE:          true,
			AllowedProtocols:   []string{"quic", "tcp", "udp"},
			SecurityLevel:      "standard",
			LoggingLevel:       "info",
			SessionTimeout:     "1h",
			MaxSessionDuration: "24h",
		}
	}

	// Cache tenant
	tm.tenantsMux.Lock()
	tm.tenants[tenant.ID] = tenant
	tm.tenantsMux.Unlock()

	tm.logger.Info("Tenant created successfully",
		zap.String("tenant_id", tenant.ID),
		zap.String("name", tenant.Name))

	return tenant, nil
}

// GetTenant получает тенант по ID
func (tm *TenantManager) GetTenant(ctx context.Context, tenantID string) (*Tenant, error) {
	// Check cache first
	tm.tenantsMux.RLock()
	if tenant, exists := tm.tenants[tenantID]; exists {
		tm.tenantsMux.RUnlock()
		return tenant, nil
	}
	tm.tenantsMux.RUnlock()

	// В полной реализации здесь был бы запрос к Zitadel API
	// для получения информации об организации
	tm.logger.Debug("Tenant not found in cache, would fetch from Zitadel API",
		zap.String("tenant_id", tenantID))

	return nil, fmt.Errorf("tenant not found: %s", tenantID)
}

// GetTenantByDomain получает тенант по домену
func (tm *TenantManager) GetTenantByDomain(ctx context.Context, domain string) (*Tenant, error) {
	tm.tenantsMux.RLock()
	defer tm.tenantsMux.RUnlock()

	for _, tenant := range tm.tenants {
		if tenant.Domain == domain {
			return tenant, nil
		}
	}

	return nil, fmt.Errorf("tenant not found for domain: %s", domain)
}

// UpdateTenant обновляет тенант
func (tm *TenantManager) UpdateTenant(ctx context.Context, tenantID string, updates map[string]interface{}) (*Tenant, error) {
	tm.tenantsMux.Lock()
	defer tm.tenantsMux.Unlock()

	tenant, exists := tm.tenants[tenantID]
	if !exists {
		return nil, fmt.Errorf("tenant not found: %s", tenantID)
	}

	// Apply updates
	for key, value := range updates {
		switch key {
		case "display_name":
			if displayName, ok := value.(string); ok {
				tenant.DisplayName = displayName
			}
		case "state":
			if stateStr, ok := value.(string); ok {
				switch stateStr {
				case "active":
					tenant.State = TenantStateActive
				case "inactive":
					tenant.State = TenantStateInactive
				case "suspended":
					tenant.State = TenantStateSuspended
				}
			}
		}
	}

	tenant.UpdatedAt = time.Now()

	tm.logger.Info("Tenant updated",
		zap.String("tenant_id", tenantID),
		zap.Any("updates", updates))

	return tenant, nil
}

// DeleteTenant удаляет тенант
func (tm *TenantManager) DeleteTenant(ctx context.Context, tenantID string) error {
	tm.tenantsMux.Lock()
	defer tm.tenantsMux.Unlock()

	if _, exists := tm.tenants[tenantID]; !exists {
		return fmt.Errorf("tenant not found: %s", tenantID)
	}

	delete(tm.tenants, tenantID)

	tm.logger.Info("Tenant deleted", zap.String("tenant_id", tenantID))
	return nil
}

// ListTenants возвращает список всех тенантов
func (tm *TenantManager) ListTenants(ctx context.Context) ([]*Tenant, error) {
	tm.tenantsMux.RLock()
	defer tm.tenantsMux.RUnlock()

	tenants := make([]*Tenant, 0, len(tm.tenants))
	for _, tenant := range tm.tenants {
		tenants = append(tenants, tenant)
	}

	return tenants, nil
}

// ValidateTenantAccess проверяет доступ тенанта к ресурсу
func (tm *TenantManager) ValidateTenantAccess(ctx context.Context, tenantID, resource string) error {
	tenant, err := tm.GetTenant(ctx, tenantID)
	if err != nil {
		return err
	}

	if tenant.State != TenantStateActive {
		return fmt.Errorf("tenant is not active: %s", tenant.State)
	}

	// Check quotas
	if err := tm.checkQuotas(tenant, resource); err != nil {
		return fmt.Errorf("quota exceeded: %w", err)
	}

	return nil
}

// checkQuotas проверяет квоты тенанта
func (tm *TenantManager) checkQuotas(tenant *Tenant, resource string) error {
	if tenant.Quotas == nil {
		return nil
	}

	switch resource {
	case "peer":
		if tenant.Quotas.CurrentPeers >= tenant.Quotas.MaxPeers {
			return fmt.Errorf("max peers limit reached: %d", tenant.Quotas.MaxPeers)
		}
	case "connection":
		if tenant.Quotas.CurrentConnections >= tenant.Quotas.MaxConnections {
			return fmt.Errorf("max connections limit reached: %d", tenant.Quotas.MaxConnections)
		}
	case "tunnel":
		if tenant.Quotas.CurrentTunnels >= tenant.Quotas.MaxTunnels {
			return fmt.Errorf("max tunnels limit reached: %d", tenant.Quotas.MaxTunnels)
		}
	}

	return nil
}

// UpdateTenantUsage обновляет использование ресурсов тенанта
func (tm *TenantManager) UpdateTenantUsage(ctx context.Context, tenantID string, usage *TenantUsage) error {
	tm.tenantsMux.Lock()
	defer tm.tenantsMux.Unlock()

	tenant, exists := tm.tenants[tenantID]
	if !exists {
		return fmt.Errorf("tenant not found: %s", tenantID)
	}

	if tenant.Quotas == nil {
		tenant.Quotas = &TenantQuotas{}
	}

	// Update usage
	tenant.Quotas.CurrentPeers = usage.Peers
	tenant.Quotas.CurrentConnections = usage.Connections
	tenant.Quotas.CurrentTunnels = usage.Tunnels
	tenant.Quotas.BandwidthUsed += usage.BandwidthDelta
	tenant.Quotas.StorageUsed = usage.StorageUsed

	tenant.UpdatedAt = time.Now()

	return nil
}

// TenantUsage представляет использование ресурсов тенанта
type TenantUsage struct {
	Peers          int   `json:"peers"`
	Connections    int   `json:"connections"`
	Tunnels        int   `json:"tunnels"`
	BandwidthDelta int64 `json:"bandwidth_delta"`
	StorageUsed    int64 `json:"storage_used"`
}

// GetTenantUsage возвращает статистику использования тенанта
func (tm *TenantManager) GetTenantUsage(ctx context.Context, tenantID string) (*TenantUsage, error) {
	tenant, err := tm.GetTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	if tenant.Quotas == nil {
		return &TenantUsage{}, nil
	}

	return &TenantUsage{
		Peers:       tenant.Quotas.CurrentPeers,
		Connections: tenant.Quotas.CurrentConnections,
		Tunnels:     tenant.Quotas.CurrentTunnels,
		StorageUsed: tenant.Quotas.StorageUsed,
	}, nil
}

// SetAutoCreate включает/отключает автоматическое создание тенантов
func (tm *TenantManager) SetAutoCreate(enabled bool) {
	tm.autoCreate = enabled
	tm.logger.Info("Auto-create tenants setting changed", zap.Bool("enabled", enabled))
}

// SetDefaultQuotas устанавливает квоты по умолчанию
func (tm *TenantManager) SetDefaultQuotas(quotas *TenantQuotas) {
	tm.defaultQuotas = quotas
	tm.logger.Info("Default quotas updated", zap.Any("quotas", quotas))
}

// GetStats возвращает статистику менеджера тенантов
func (tm *TenantManager) GetStats() map[string]interface{} {
	tm.tenantsMux.RLock()
	defer tm.tenantsMux.RUnlock()

	stats := map[string]interface{}{
		"total_tenants": len(tm.tenants),
		"auto_create":   tm.autoCreate,
		"cache_expiry":  tm.cacheExpiry,
	}

	// Count by state
	stateCount := make(map[string]int)
	for _, tenant := range tm.tenants {
		stateCount[tenant.State.String()]++
	}
	stats["tenants_by_state"] = stateCount

	return stats
}
