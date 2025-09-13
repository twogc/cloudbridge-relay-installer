package token

import (
	"fmt"
	"time"
)

// MeshConfig represents mesh network configuration in JWT
type MeshConfig struct {
	RegistryURL         string        `json:"registry_url"`
	HeartbeatInterval   time.Duration `json:"heartbeat_interval"`
	MaxPeers            int           `json:"max_peers"`
	RoutingStrategy     string        `json:"routing_strategy"`
	EnableAutoDiscovery bool          `json:"enable_auto_discovery"`
	TrustLevel          string        `json:"trust_level"`
}

// Validate validates mesh configuration
func (m *MeshConfig) Validate() error {
	if m.RegistryURL == "" {
		return fmt.Errorf("registry_url is required")
	}

	if m.HeartbeatInterval <= 0 {
		return fmt.Errorf("heartbeat_interval must be positive, got: %v", m.HeartbeatInterval)
	}

	if m.MaxPeers <= 0 {
		return fmt.Errorf("max_peers must be positive, got: %d", m.MaxPeers)
	}

	validStrategies := map[string]bool{
		"wireguard_first":     true,
		"performance_optimal": true,
		"geographic_optimal":  true,
		"hybrid":              true,
	}
	if m.RoutingStrategy != "" && !validStrategies[m.RoutingStrategy] {
		return fmt.Errorf("invalid routing_strategy: %s", m.RoutingStrategy)
	}

	validTrustLevels := map[string]bool{
		"none":  true,
		"basic": true,
		"full":  true,
	}
	if m.TrustLevel != "" && !validTrustLevels[m.TrustLevel] {
		return fmt.Errorf("invalid trust_level: %s", m.TrustLevel)
	}

	return nil
}








