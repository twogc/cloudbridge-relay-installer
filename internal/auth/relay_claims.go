package auth

import (
	"github.com/golang-jwt/jwt/v5"
)

// RelayClaims представляет claims для relay токенов от Zitadel
type RelayClaims struct {
	jwt.RegisteredClaims

	// Zitadel specific claims
	OrgID         string   `json:"org_id"`
	TenantID      string   `json:"tenant_id"`
	ServiceName   string   `json:"service_name"`
	Permissions   []string `json:"permissions"`
	Roles         []string `json:"roles"`
	RelayServerID string   `json:"relay_server_id"`
	ProtocolType  string   `json:"protocol_type"`
	TokenType     string   `json:"token_type"`

	// Дополнительные поля для CloudBridge
	RelaySecret   string         `json:"relay_secret,omitempty"`
	NetworkConfig *NetworkConfig `json:"network_config,omitempty"`
}

// NetworkConfig представляет сетевую конфигурацию для relay сервера
type NetworkConfig struct {
	AllowedSubnets []string `json:"allowed_subnets"`
	MaxConnections int      `json:"max_connections"`
	BandwidthLimit int64    `json:"bandwidth_limit"`
}

// HasPermission проверяет наличие разрешения
func (c *RelayClaims) HasPermission(permission string) bool {
	if c == nil {
		return false
	}
	for _, p := range c.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// HasRole проверяет наличие роли
func (c *RelayClaims) HasRole(role string) bool {
	if c == nil {
		return false
	}
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// IsRelayToken проверяет, является ли токен relay токеном
func (c *RelayClaims) IsRelayToken() bool {
	if c == nil {
		return false
	}
	return c.TokenType == "relay_permanent" || c.TokenType == "relay_session"
}

// IsPermanentToken проверяет, является ли токен постоянным (без времени жизни)
func (c *RelayClaims) IsPermanentToken() bool {
	if c == nil {
		return false
	}
	return c.TokenType == "relay_permanent"
}

// GetExpirationTime возвращает время истечения токена (если есть)
func (c *RelayClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	if c == nil || c.ExpiresAt == nil {
		return nil, nil
	}
	return c.ExpiresAt, nil
}

// ValidatePermissions проверяет наличие всех необходимых разрешений
func (c *RelayClaims) ValidatePermissions(requiredPermissions []string) bool {
	if c == nil {
		return false
	}

	for _, required := range requiredPermissions {
		if !c.HasPermission(required) {
			return false
		}
	}
	return true
}

// ValidateRoles проверяет наличие хотя бы одной из требуемых ролей
func (c *RelayClaims) ValidateRoles(requiredRoles []string) bool {
	if c == nil {
		return false
	}

	for _, required := range requiredRoles {
		if c.HasRole(required) {
			return true
		}
	}
	return false
}

// GetRelayServerID возвращает ID relay сервера
func (c *RelayClaims) GetRelayServerID() string {
	if c == nil {
		return ""
	}
	if c.RelayServerID != "" {
		return c.RelayServerID
	}
	return c.Subject
}

// GetTenantID возвращает ID тенанта
func (c *RelayClaims) GetTenantID() string {
	if c == nil {
		return ""
	}
	return c.TenantID
}

// GetOrgID возвращает ID организации
func (c *RelayClaims) GetOrgID() string {
	if c == nil {
		return ""
	}
	return c.OrgID
}
