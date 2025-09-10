package config

import (
	"fmt"
	"time"
)

// P2PConfig represents P2P configuration
type P2PConfig struct {
	WireGuard WireGuardConfig `yaml:"wireguard"`
	Mesh      MeshConfig      `yaml:"mesh"`
	Security  SecurityConfig  `yaml:"security"`
}

// WireGuardConfig represents WireGuard configuration
type WireGuardConfig struct {
	Enabled   bool               `yaml:"enabled"`
	Interface WireGuardInterface `yaml:"interface"`
}

// WireGuardInterface represents WireGuard interface configuration
type WireGuardInterface struct {
	Name       string `yaml:"name"`
	ListenPort int    `yaml:"listen_port"`
	Address    string `yaml:"address"`
	MTU        int    `yaml:"mtu"`
}

// MeshConfig represents mesh network configuration
type MeshConfig struct {
	Enabled             bool          `yaml:"enabled"`
	RegistryURL         string        `yaml:"registry_url"`
	HeartbeatInterval   time.Duration `yaml:"heartbeat_interval"`
	MaxPeers            int           `yaml:"max_peers"`
	RoutingStrategy     string        `yaml:"routing_strategy"`
	EnableAutoDiscovery bool          `yaml:"enable_auto_discovery"`
}

// SecurityConfig represents P2P security configuration
type SecurityConfig struct {
	PeerWhitelistEnabled bool          `yaml:"peer_whitelist_enabled"`
	KeyRotationInterval  time.Duration `yaml:"key_rotation_interval"`
	StrictMode           bool          `yaml:"strict_mode"`
}

// Validate validates P2P configuration
func (c *P2PConfig) Validate() error {
	if c.WireGuard.Enabled {
		if c.WireGuard.Interface.ListenPort == 0 {
			return fmt.Errorf("wireguard listen_port is required when enabled")
		}
		if c.WireGuard.Interface.Name == "" {
			return fmt.Errorf("wireguard interface name is required when enabled")
		}
		if c.WireGuard.Interface.Address == "" {
			return fmt.Errorf("wireguard interface address is required when enabled")
		}
	}

	if c.Mesh.Enabled {
		if c.Mesh.RegistryURL == "" {
			return fmt.Errorf("mesh registry_url is required when enabled")
		}
		if c.Mesh.MaxPeers <= 0 {
			return fmt.Errorf("mesh max_peers must be positive")
		}
		if c.Mesh.HeartbeatInterval <= 0 {
			return fmt.Errorf("mesh heartbeat_interval must be positive")
		}
	}

	return nil
}

// GetDefaultP2PConfig returns default P2P configuration
func GetDefaultP2PConfig() *P2PConfig {
	return &P2PConfig{
		WireGuard: WireGuardConfig{
			Enabled: true,
			Interface: WireGuardInterface{
				Name:       "wg0",
				ListenPort: 51820,
				Address:    "10.0.0.1/24",
				MTU:        1420,
			},
		},
		Mesh: MeshConfig{
			Enabled:             true,
			RegistryURL:         "https://registry.cloudbridge.com",
			HeartbeatInterval:   30 * time.Second,
			MaxPeers:            100,
			RoutingStrategy:     "hybrid",
			EnableAutoDiscovery: true,
		},
		Security: SecurityConfig{
			PeerWhitelistEnabled: false,
			KeyRotationInterval:  24 * time.Hour,
			StrictMode:           false,
		},
	}
}


