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
		ProtocolType:   "p2p-mesh",
		Scope:          "relay",
		ConnectionType: "quic",
		TenantID:       "test-tenant",
		ServerID:       "test-server",
		Permissions:    []string{"mesh_join", "quic_connect", "p2p_connect"},
	}

	if err := claims.Validate(); err != nil {
		t.Errorf("P2PClaims validation failed: %v", err)
	}

	// Test getters
	if claims.GetConnectionType() != "quic" {
		t.Errorf("Expected connection type 'quic', got '%s'", claims.GetConnectionType())
	}

	if claims.ProtocolType != "p2p-mesh" {
		t.Errorf("Expected protocol type 'p2p-mesh', got '%s'", claims.ProtocolType)
	}

	if claims.TenantID != "test-tenant" {
		t.Errorf("Expected tenant ID 'test-tenant', got '%s'", claims.TenantID)
	}

	if claims.ServerID != "test-server" {
		t.Errorf("Expected server ID 'test-server', got '%s'", claims.ServerID)
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
