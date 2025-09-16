package token

import (
	"testing"
	"time"
)

func TestP2PClaimsWithNewFields(t *testing.T) {
	// Test WireGuardConfig with endpoint
	wgConfig := &WireGuardConfig{
		InterfaceName: "wg0",
		ListenPort:    51820,
		Address:       "10.0.0.100/24",
		MTU:           1420,
		AllowedIPs:    []string{"10.0.0.0/24", "192.168.1.0/24"},
		PrivateKey:    "test-private-key",
		PublicKey:     "test-public-key",
		Endpoint:      "192.168.1.100:51820",
	}

	if err := wgConfig.Validate(); err != nil {
		t.Errorf("WireGuardConfig validation failed: %v", err)
	}

	// Test MeshConfig with network_id and subnet
	meshConfig := &MeshConfig{
		NetworkID:           "mesh-network-001",
		Subnet:              "10.0.0.0/16",
		RegistryURL:         "https://mesh-registry.2gc.ru",
		HeartbeatInterval:   30 * time.Second,
		MaxPeers:            10,
		RoutingStrategy:     "performance_optimal",
		EnableAutoDiscovery: true,
		TrustLevel:          "basic",
	}

	if err := meshConfig.Validate(); err != nil {
		t.Errorf("MeshConfig validation failed: %v", err)
	}

	// Test NetworkConfig
	networkConfig := &NetworkConfig{
		Subnet:        "10.0.0.0/24",
		Gateway:       "10.0.0.1",
		DNS:           []string{"8.8.8.8", "1.1.1.1"},
		MTU:           1420,
		EnableIPv6:    false,
		FirewallRules: []string{"allow_ssh", "allow_http"},
	}

	if err := networkConfig.Validate(); err != nil {
		t.Errorf("NetworkConfig validation failed: %v", err)
	}

	// Test complete P2PClaims
	claims := &P2PClaims{
		ConnectionType:  "wireguard",
		WireGuardConfig: wgConfig,
		MeshConfig:      meshConfig,
		NetworkConfig:   networkConfig,
		PeerWhitelist:   []string{"peer-001", "peer-002"},
		TenantID:        "test-tenant",
		Permissions:     []string{"mesh_join", "wireguard_connect"},
	}

	if err := claims.Validate(); err != nil {
		t.Errorf("P2PClaims validation failed: %v", err)
	}

	// Test getters
	if claims.GetConnectionType() != "wireguard" {
		t.Errorf("Expected connection type 'wireguard', got '%s'", claims.GetConnectionType())
	}

	if claims.GetWireGuardConfig().Endpoint != "192.168.1.100:51820" {
		t.Errorf("Expected endpoint '192.168.1.100:51820', got '%s'", claims.GetWireGuardConfig().Endpoint)
	}

	if claims.GetMeshConfig().NetworkID != "mesh-network-001" {
		t.Errorf("Expected network_id 'mesh-network-001', got '%s'", claims.GetMeshConfig().NetworkID)
	}

	if claims.GetMeshConfig().Subnet != "10.0.0.0/16" {
		t.Errorf("Expected subnet '10.0.0.0/16', got '%s'", claims.GetMeshConfig().Subnet)
	}

	if !claims.HasPermission("mesh_join") {
		t.Error("Expected permission 'mesh_join' to be present")
	}

	if claims.HasPermission("non_existent") {
		t.Error("Expected permission 'non_existent' to be absent")
	}
}

func TestWireGuardConfigEndpointValidation(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		wantErr  bool
	}{
		{"valid endpoint", "192.168.1.100:51820", false},
		{"valid endpoint with domain", "example.com:51820", false},
		{"invalid short endpoint", "1:1", true},
		{"empty endpoint", "", false}, // empty is allowed (omitempty)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &WireGuardConfig{
				InterfaceName: "wg0",
				ListenPort:    51820,
				Address:       "10.0.0.100/24",
				MTU:           1420,
				AllowedIPs:    []string{"10.0.0.0/24"},
				Endpoint:      tt.endpoint,
			}

			err := config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("WireGuardConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMeshConfigNewFieldsValidation(t *testing.T) {
	tests := []struct {
		name      string
		networkID string
		subnet    string
		wantErr   bool
	}{
		{"valid network_id and subnet", "mesh-001", "10.0.0.0/16", false},
		{"short network_id", "ab", "10.0.0.0/16", true},
		{"invalid subnet", "mesh-001", "10.0.0.0", true},
		{"empty fields", "", "", false}, // empty is allowed (omitempty)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &MeshConfig{
				RegistryURL:       "https://registry.example.com",
				HeartbeatInterval: 30 * time.Second,
				MaxPeers:          10,
				NetworkID:         tt.networkID,
				Subnet:            tt.subnet,
			}

			err := config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("MeshConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}


