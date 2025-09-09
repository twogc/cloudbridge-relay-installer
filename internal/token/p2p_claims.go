package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// P2PClaims represents P2P-specific JWT claims
type P2PClaims struct {
	jwt.RegisteredClaims
	ConnectionType  string           `json:"connection_type"`
	WireGuardConfig *WireGuardConfig `json:"wireguard_config,omitempty"`
	MeshConfig      *MeshConfig      `json:"mesh_config,omitempty"`
	PeerWhitelist   []string         `json:"peer_whitelist,omitempty"`
	NetworkConfig   *NetworkConfig   `json:"network_config,omitempty"`
	TenantID        string           `json:"tenant_id,omitempty"`
	Permissions     []string         `json:"permissions,omitempty"`
	ExpiresAt       *jwt.NumericDate `json:"exp,omitempty"`
}

// Validate validates P2P claims
func (p *P2PClaims) Validate() error {
	if p.ConnectionType == "" {
		return fmt.Errorf("connection_type is required")
	}

	validConnectionTypes := map[string]bool{
		"wireguard": true,
		"quic":      true,
		"tls":       true,
		"hybrid":    true,
	}
	if !validConnectionTypes[p.ConnectionType] {
		return fmt.Errorf("invalid connection_type: %s", p.ConnectionType)
	}

	if p.ConnectionType == "wireguard" || p.ConnectionType == "hybrid" {
		if p.WireGuardConfig == nil {
			return fmt.Errorf("wireguard_config is required for wireguard/hybrid connection type")
		}
		if err := p.WireGuardConfig.Validate(); err != nil {
			return fmt.Errorf("wireguard_config validation failed: %v", err)
		}
	}

	if p.MeshConfig != nil {
		if err := p.MeshConfig.Validate(); err != nil {
			return fmt.Errorf("mesh_config validation failed: %v", err)
		}
	}

	if p.NetworkConfig != nil {
		if err := p.NetworkConfig.Validate(); err != nil {
			return fmt.Errorf("network_config validation failed: %v", err)
		}
	}

	for i, peerID := range p.PeerWhitelist {
		if peerID == "" {
			return fmt.Errorf("peer_whitelist[%d] cannot be empty", i)
		}
	}

	return nil
}

// GetConnectionType returns the connection type
func (p *P2PClaims) GetConnectionType() string {
	return p.ConnectionType
}

// GetWireGuardConfig returns WireGuard configuration
func (p *P2PClaims) GetWireGuardConfig() *WireGuardConfig {
	return p.WireGuardConfig
}

// GetMeshConfig returns mesh configuration
func (p *P2PClaims) GetMeshConfig() *MeshConfig {
	return p.MeshConfig
}

// GetPeerWhitelist returns peer whitelist
func (p *P2PClaims) GetPeerWhitelist() []string {
	return p.PeerWhitelist
}

// GetNetworkConfig returns network configuration
func (p *P2PClaims) GetNetworkConfig() *NetworkConfig {
	return p.NetworkConfig
}

// GetTenantID returns tenant ID
func (p *P2PClaims) GetTenantID() string {
	return p.TenantID
}

// GetPermissions returns permissions
func (p *P2PClaims) GetPermissions() []string {
	return p.Permissions
}

// HasPermission checks if the claim has a specific permission
func (p *P2PClaims) HasPermission(permission string) bool {
	for _, perm := range p.Permissions {
		if perm == permission {
			return true
		}
	}
	return false
}

// IsExpired checks if the token is expired
func (p *P2PClaims) IsExpired() bool {
	if p.ExpiresAt == nil {
		return false
	}
	return time.Now().After(p.ExpiresAt.Time)
}
