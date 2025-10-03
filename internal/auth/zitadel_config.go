package auth

import (
	"crypto/rsa"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// ZitadelConfig представляет конфигурацию для интеграции с Zitadel
type ZitadelConfig struct {
	// Основные настройки
	Domain   string `yaml:"domain" json:"domain"`
	Issuer   string `yaml:"issuer" json:"issuer"`
	Audience string `yaml:"audience" json:"audience"`

	// Ключи для валидации JWT
	PublicKey *rsa.PublicKey `yaml:"-" json:"-"`
	KeyID     string         `yaml:"key_id" json:"key_id"`

	// Настройки валидации
	ClockSkew time.Duration `yaml:"clock_skew" json:"clock_skew"`

	// Настройки клиента
	HTTPTimeout time.Duration `yaml:"http_timeout" json:"http_timeout"`

	// Логирование
	Logger *zap.Logger `yaml:"-" json:"-"`
}

// DefaultZitadelConfig возвращает конфигурацию по умолчанию
func DefaultZitadelConfig() *ZitadelConfig {
	return &ZitadelConfig{
		Domain:      "https://edge.2gc.ru",
		Issuer:      "https://edge.2gc.ru/oauth/v2",
		Audience:    "cloudbridge-relay",
		ClockSkew:   5 * time.Minute,
		HTTPTimeout: 30 * time.Second,
	}
}

// Validate проверяет корректность конфигурации
func (c *ZitadelConfig) Validate() error {
	if c.Domain == "" {
		return fmt.Errorf("domain is required")
	}

	if c.Issuer == "" {
		return fmt.Errorf("issuer is required")
	}

	if c.Audience == "" {
		return fmt.Errorf("audience is required")
	}

	if c.PublicKey == nil {
		return fmt.Errorf("public key is required")
	}

	if c.ClockSkew < 0 {
		return fmt.Errorf("clock skew must be non-negative")
	}

	if c.HTTPTimeout <= 0 {
		return fmt.Errorf("http timeout must be positive")
	}

	return nil
}

// GetValidatorConfig возвращает конфигурацию для ZitadelValidator
func (c *ZitadelConfig) GetValidatorConfig() ZitadelValidatorConfig {
	return ZitadelValidatorConfig{
		PublicKey: c.PublicKey,
		Issuer:    c.Issuer,
		Audience:  c.Audience,
		KeyID:     c.KeyID,
		ClockSkew: c.ClockSkew,
		Logger:    c.Logger,
	}
}

// GetClientConfig возвращает конфигурацию для ZitadelClient
func (c *ZitadelConfig) GetClientConfig() ZitadelClientConfig {
	return ZitadelClientConfig{
		Domain:     c.Domain,
		HTTPClient: &http.Client{Timeout: c.HTTPTimeout},
		Logger:     c.Logger,
	}
}

// RelayPermissions определяет разрешения для relay серверов
var RelayPermissions = struct {
	// Основные разрешения
	P2PConnect  string
	MeshJoin    string
	RelayAccess string

	// Сетевые разрешения
	NetworkAllocate   string
	NetworkDeallocate string

	// Разрешения для тенантов
	TenantNetworkRead  string
	TenantNetworkWrite string

	// Административные разрешения
	AdminAccess   string
	ServiceAccess string
}{
	P2PConnect:         "p2p_connect",
	MeshJoin:           "mesh_join",
	RelayAccess:        "relay_access",
	NetworkAllocate:    "network:allocate",
	NetworkDeallocate:  "network:deallocate",
	TenantNetworkRead:  "tenant:network:read",
	TenantNetworkWrite: "tenant:network:write",
	AdminAccess:        "admin:access",
	ServiceAccess:      "service:access",
}

// RelayRoles определяет роли для relay серверов
var RelayRoles = struct {
	// Основные роли
	Service            string
	CloudBridgeService string
	RelayService       string
	NetworkManager     string

	// Административные роли
	Admin       string
	SystemAdmin string
}{
	Service:            "service",
	CloudBridgeService: "cloudbridge-service",
	RelayService:       "relay-service",
	NetworkManager:     "network-manager",
	Admin:              "admin",
	SystemAdmin:        "system-admin",
}

// GetRequiredPermissionsForRelay возвращает необходимые разрешения для relay сервера
func GetRequiredPermissionsForRelay() []string {
	return []string{
		RelayPermissions.P2PConnect,
		RelayPermissions.MeshJoin,
		RelayPermissions.RelayAccess,
		RelayPermissions.NetworkAllocate,
		RelayPermissions.NetworkDeallocate,
		RelayPermissions.TenantNetworkRead,
		RelayPermissions.TenantNetworkWrite,
	}
}

// GetRequiredRolesForRelay возвращает необходимые роли для relay сервера
func GetRequiredRolesForRelay() []string {
	return []string{
		RelayRoles.Service,
		RelayRoles.CloudBridgeService,
		RelayRoles.RelayService,
		RelayRoles.NetworkManager,
	}
}

// GetRequiredPermissionsForAdmin возвращает необходимые разрешения для администратора
func GetRequiredPermissionsForAdmin() []string {
	return []string{
		RelayPermissions.AdminAccess,
		RelayPermissions.ServiceAccess,
		RelayPermissions.NetworkAllocate,
		RelayPermissions.NetworkDeallocate,
		RelayPermissions.TenantNetworkRead,
		RelayPermissions.TenantNetworkWrite,
	}
}

// GetRequiredRolesForAdmin возвращает необходимые роли для администратора
func GetRequiredRolesForAdmin() []string {
	return []string{
		RelayRoles.Admin,
		RelayRoles.SystemAdmin,
	}
}
