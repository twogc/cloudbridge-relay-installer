package main

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/twogc/cloudbridge-relay-installer/internal/zitadel"
)

// ZitadelIntegration интеграция с Zitadel
type ZitadelIntegration struct {
	client        *zitadel.Client
	tenantManager *zitadel.TenantManager
	logger        *zap.Logger
	enabled       bool
}

// ZitadelConfig конфигурация Zitadel
type ZitadelConfig struct {
	Enabled        bool               `json:"enabled"`
	Issuer         string             `json:"issuer"`
	Audience       string             `json:"audience"`
	TokenURL       string             `json:"token_url"`
	ClientID       string             `json:"client_id"`
	ClientSecret   string             `json:"client_secret"`
	FallbackSecret string             `json:"fallback_secret"`
	K8sConfig      *zitadel.K8sConfig `json:"k8s_config,omitempty"`

	TokenValidation struct {
		Audience  string `json:"audience"`
		ClockSkew string `json:"clock_skew"`
	} `json:"token_validation"`

	CacheTTL    string `json:"cache_ttl"`
	HTTPTimeout string `json:"http_timeout"`

	TenantManagement struct {
		Enabled       bool `json:"enabled"`
		AutoCreate    bool `json:"auto_create"`
		DefaultQuotas struct {
			MaxPeers       int    `json:"max_peers"`
			MaxConnections int    `json:"max_connections"`
			MaxTunnels     int    `json:"max_tunnels"`
			BandwidthLimit string `json:"bandwidth_limit"`
			StorageLimit   string `json:"storage_limit"`
		} `json:"default_quotas"`
	} `json:"tenant_management"`
}

// NewZitadelIntegration создает новую интеграцию с Zitadel
func NewZitadelIntegration(config *ZitadelConfig, logger *zap.Logger) (*ZitadelIntegration, error) {
	if !config.Enabled {
		return &ZitadelIntegration{
			enabled: false,
			logger:  logger,
		}, nil
	}

	ctx := context.Background()

	// Check if running in k8s environment
	var client *zitadel.Client
	var err error

	if config.K8sConfig != nil {
		// Use k8s-specific client
		logger.Info("Creating Zitadel client for k8s environment")
		client, err = zitadel.NewK8sClient(ctx, config.K8sConfig, logger)
	} else {
		// Use standard client
		logger.Info("Creating standard Zitadel client")
		baseConfig := &zitadel.Config{
			Issuer:       config.Issuer,
			Audience:     config.Audience,
			TokenURL:     config.TokenURL,
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			Scopes:       []string{"openid", "profile", "email"},
			HTTPTimeout:  30 * time.Second,
		}
		client, err = zitadel.NewClient(ctx, baseConfig, logger)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create Zitadel client: %w", err)
	}

	// Get tenant manager from client
	tenantManager := client.GetTenantManager()

	// Configure tenant management
	if config.TenantManagement.Enabled {
		tenantManager.SetAutoCreate(config.TenantManagement.AutoCreate)
		// Проверяем, есть ли настройки квот (проверка на ненулевые значения)
		if config.TenantManagement.DefaultQuotas.MaxPeers > 0 {
			quotas := &zitadel.TenantQuotas{
				MaxPeers:       config.TenantManagement.DefaultQuotas.MaxPeers,
				MaxConnections: config.TenantManagement.DefaultQuotas.MaxConnections,
				MaxTunnels:     config.TenantManagement.DefaultQuotas.MaxTunnels,
				BandwidthLimit: config.TenantManagement.DefaultQuotas.BandwidthLimit,
				StorageLimit:   config.TenantManagement.DefaultQuotas.StorageLimit,
			}
			tenantManager.SetDefaultQuotas(quotas)
		}
	}

	logger.Info("Zitadel integration initialized successfully",
		zap.String("issuer", config.Issuer),
		zap.String("audience", config.Audience),
		zap.String("client_id", config.ClientID),
		zap.Bool("tenant_management", config.TenantManagement.Enabled),
		zap.Bool("k8s_mode", config.K8sConfig != nil))

	return &ZitadelIntegration{
		client:        client,
		tenantManager: tenantManager,
		logger:        logger,
		enabled:       true,
	}, nil
}

// Start запускает интеграцию с Zitadel
func (zi *ZitadelIntegration) Start(ctx context.Context) error {
	if !zi.enabled {
		zi.logger.Info("Zitadel integration disabled")
		return nil
	}

	zi.logger.Info("Starting Zitadel integration")

	// Проверка подключения к Zitadel
	if err := zi.healthCheck(ctx); err != nil {
		return fmt.Errorf("Zitadel health check failed: %w", err)
	}

	zi.logger.Info("Zitadel integration started successfully")
	return nil
}

// Stop останавливает интеграцию с Zitadel
func (zi *ZitadelIntegration) Stop(ctx context.Context) error {
	if !zi.enabled {
		return nil
	}

	zi.logger.Info("Stopping Zitadel integration")

	// Закрытие соединений
	if zi.client != nil {
		if err := zi.client.Close(); err != nil {
			zi.logger.Warn("Failed to close Zitadel client", zap.Error(err))
		}
	}

	zi.logger.Info("Zitadel integration stopped")
	return nil
}

// GetClient возвращает Zitadel клиент
func (zi *ZitadelIntegration) GetClient() *zitadel.Client {
	return zi.client
}

// GetTenantManager возвращает менеджер тенантов
func (zi *ZitadelIntegration) GetTenantManager() *zitadel.TenantManager {
	return zi.tenantManager
}

// IsEnabled возвращает статус интеграции
func (zi *ZitadelIntegration) IsEnabled() bool {
	return zi.enabled
}

// healthCheck проверяет доступность Zitadel
func (zi *ZitadelIntegration) healthCheck(ctx context.Context) error {
	if zi.client == nil {
		return fmt.Errorf("client not initialized")
	}

	// TODO: Implement health check when zitadel-go client is available
	// For now, we assume the client is healthy if it was created successfully
	zi.logger.Debug("Zitadel health check skipped - client created successfully")
	return nil
}

// CreateDefaultTenant создает тенант по умолчанию для тестирования
func (zi *ZitadelIntegration) CreateDefaultTenant(ctx context.Context) (string, error) {
	if !zi.enabled || zi.tenantManager == nil {
		return "", fmt.Errorf("Zitadel integration not enabled or tenant manager not available")
	}

	tenantReq := &zitadel.TenantCreateRequest{
		Name:        "CloudBridge Default",
		DisplayName: "CloudBridge Default Tenant",
		Domain:      "cloudbridge.local",
		AdminEmail:  "admin@cloudbridge.local",
		Quotas: &zitadel.TenantQuotas{
			MaxPeers:       10,
			MaxConnections: 100,
			MaxTunnels:     20,
			BandwidthLimit: "1GB",
			StorageLimit:   "10GB",
		},
	}

	tenant, err := zi.tenantManager.CreateTenant(ctx, tenantReq)
	if err != nil {
		return "", fmt.Errorf("failed to create default tenant: %w", err)
	}

	zi.logger.Info("Created default tenant", zap.String("tenant_id", tenant.ID))
	return tenant.ID, nil
}

// MigrationHelper помощник для миграции с Keycloak на Zitadel
type MigrationHelper struct {
	zitadelIntegration *ZitadelIntegration
	logger             *zap.Logger
}

// NewMigrationHelper создает помощника для миграции
func NewMigrationHelper(zi *ZitadelIntegration, logger *zap.Logger) *MigrationHelper {
	return &MigrationHelper{
		zitadelIntegration: zi,
		logger:             logger,
	}
}

// ValidateTokenWithFallback валидирует токен через Zitadel
func (mh *MigrationHelper) ValidateTokenWithFallback(ctx context.Context, token string) (*zitadel.TokenClaims, error) {
	if !mh.zitadelIntegration.enabled {
		return nil, fmt.Errorf("Zitadel integration not enabled")
	}

	// Валидация через Zitadel
	claims, err := mh.zitadelIntegration.client.ValidateToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	mh.logger.Debug("Token validated successfully via Zitadel")
	return claims, nil
}
