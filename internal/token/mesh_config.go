package token

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// MeshConfig represents mesh network configuration in JWT
type MeshConfig struct {
	NetworkID           string        `json:"network_id,omitempty"`
	Subnet              string        `json:"subnet,omitempty"`
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

	// Validate network_id if provided
	if m.NetworkID != "" && len(m.NetworkID) < 3 {
		return fmt.Errorf("network_id must be at least 3 characters long, got: %s", m.NetworkID)
	}

	// Validate subnet if provided
	if m.Subnet != "" {
		if len(m.Subnet) < 7 { // minimum: "0.0.0.0/0"
			return fmt.Errorf("subnet appears to be invalid CIDR: %s", m.Subnet)
		}
		// Check if it contains a slash (CIDR notation)
		if !strings.Contains(m.Subnet, "/") {
			return fmt.Errorf("subnet must be in CIDR notation (e.g., 10.0.0.0/16), got: %s", m.Subnet)
		}
	}

	validStrategies := map[string]bool{
		"quic_first":          true,
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

// UnmarshalJSON implements custom JSON unmarshaling for MeshConfig
// to support both string and numeric formats for heartbeat_interval
func (m *MeshConfig) UnmarshalJSON(data []byte) error {
	// Create a temporary struct with heartbeat_interval as interface{}
	type TempMeshConfig struct {
		NetworkID           string      `json:"network_id,omitempty"`
		Subnet              string      `json:"subnet,omitempty"`
		RegistryURL         string      `json:"registry_url"`
		HeartbeatInterval   interface{} `json:"heartbeat_interval"`
		MaxPeers            int         `json:"max_peers"`
		RoutingStrategy     string      `json:"routing_strategy"`
		EnableAutoDiscovery bool        `json:"enable_auto_discovery"`
		TrustLevel          string      `json:"trust_level"`
	}

	var temp TempMeshConfig
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Copy simple fields
	m.NetworkID = temp.NetworkID
	m.Subnet = temp.Subnet
	m.RegistryURL = temp.RegistryURL
	m.MaxPeers = temp.MaxPeers
	m.RoutingStrategy = temp.RoutingStrategy
	m.EnableAutoDiscovery = temp.EnableAutoDiscovery
	m.TrustLevel = temp.TrustLevel

	// Handle heartbeat_interval conversion
	switch v := temp.HeartbeatInterval.(type) {
	case string:
		// Parse string format like "30s", "1m", "2h"
		duration, err := time.ParseDuration(v)
		if err != nil {
			return fmt.Errorf("invalid heartbeat_interval format: %s", v)
		}
		m.HeartbeatInterval = duration
	case float64:
		// Parse numeric format (seconds)
		m.HeartbeatInterval = time.Duration(v) * time.Second
	case int:
		// Parse integer format (seconds)
		m.HeartbeatInterval = time.Duration(v) * time.Second
	case int64:
		// Parse int64 format (seconds)
		m.HeartbeatInterval = time.Duration(v) * time.Second
	default:
		return fmt.Errorf("unsupported heartbeat_interval type: %T", v)
	}

	return nil
}
