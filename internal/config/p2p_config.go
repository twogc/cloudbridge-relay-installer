package config

import (
	"fmt"
	"time"
)

// P2PConfig represents P2P configuration
type P2PConfig struct {
	QUIC     QUICConfig     `yaml:"quic"`
	Mesh     MeshConfig     `yaml:"mesh"`
	Security SecurityConfig `yaml:"security"`
}

// QUICConfig represents QUIC configuration
type QUICConfig struct {
	Enabled             bool          `yaml:"enabled"`
	ListenPort          int           `yaml:"listen_port"`
	MaxStreams          int           `yaml:"max_streams"`
	HandshakeTimeout    time.Duration `yaml:"handshake_timeout"`
	IdleTimeout         time.Duration `yaml:"idle_timeout"`
	ConnectionMigration bool          `yaml:"connection_migration"`
	Multiplexing        bool          `yaml:"multiplexing"`
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
	if c.QUIC.Enabled {
		if c.QUIC.ListenPort == 0 {
			return fmt.Errorf("quic listen_port is required when enabled")
		}
		if c.QUIC.MaxStreams <= 0 {
			return fmt.Errorf("quic max_streams must be positive")
		}
		if c.QUIC.HandshakeTimeout <= 0 {
			return fmt.Errorf("quic handshake_timeout must be positive")
		}
		if c.QUIC.IdleTimeout <= 0 {
			return fmt.Errorf("quic idle_timeout must be positive")
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
		QUIC: QUICConfig{
			Enabled:             true,
			ListenPort:          443,
			MaxStreams:          1000,
			HandshakeTimeout:    10 * time.Second,
			IdleTimeout:         30 * time.Second,
			ConnectionMigration: true,
			Multiplexing:        true,
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
