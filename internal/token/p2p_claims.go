package token

import (
	"encoding/json"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// P2PClaims represents P2P-specific JWT claims
type P2PClaims struct {
	jwt.RegisteredClaims
	ProtocolType   string   `json:"protocol_type"`
	Scope          string   `json:"scope"`
	TenantID       string   `json:"tenant_id"`
	ServerID       string   `json:"server_id"`
	ConnectionType string   `json:"connection_type"`
	Permissions    []string `json:"permissions"`
}

// Validate validates P2P claims
func (p *P2PClaims) Validate() error {
	if p.ProtocolType != "p2p-mesh" {
		return fmt.Errorf("unsupported protocol: %s", p.ProtocolType)
	}

	if p.ConnectionType != "quic" {
		return fmt.Errorf("unsupported connection type: %s", p.ConnectionType)
	}

	if p.TenantID == "" {
		return fmt.Errorf("missing tenant_id")
	}

	if p.ServerID == "" {
		return fmt.Errorf("missing server_id")
	}

	// Проверить разрешения
	if !contains(p.Permissions, "p2p_connect") {
		return fmt.Errorf("missing p2p_connect permission")
	}

	return nil
}

// contains проверяет, содержит ли слайс строку
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetConnectionType returns the connection type
func (p *P2PClaims) GetConnectionType() string {
	return p.ConnectionType
}

// GetTenantID returns tenant ID
func (p *P2PClaims) GetTenantID() string {
	return p.TenantID
}

// GetServerID returns server ID
func (p *P2PClaims) GetServerID() string {
	return p.ServerID
}

// GetPermissions returns permissions
func (p *P2PClaims) GetPermissions() []string {
	return p.Permissions
}

// UnmarshalJSON implements custom JSON unmarshaling for P2PClaims
func (p *P2PClaims) UnmarshalJSON(data []byte) error {
	// Use standard JSON unmarshaling for simplified structure
	type TempP2PClaims struct {
		jwt.RegisteredClaims
		ProtocolType   string   `json:"protocol_type"`
		Scope          string   `json:"scope"`
		TenantID       string   `json:"tenant_id"`
		ServerID       string   `json:"server_id"`
		ConnectionType string   `json:"connection_type"`
		Permissions    []string `json:"permissions"`
	}

	var temp TempP2PClaims
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Copy fields
	p.RegisteredClaims = temp.RegisteredClaims
	p.ProtocolType = temp.ProtocolType
	p.Scope = temp.Scope
	p.TenantID = temp.TenantID
	p.ServerID = temp.ServerID
	p.ConnectionType = temp.ConnectionType
	p.Permissions = temp.Permissions

	return nil
}

// HasPermission checks if the claim has a specific permission
func (p *P2PClaims) HasPermission(permission string) bool {
	return contains(p.Permissions, permission)
}
