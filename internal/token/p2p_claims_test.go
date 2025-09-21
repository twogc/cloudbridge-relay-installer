package token

import (
	"testing"
	"time"
)

func TestP2PClaimsWithNewFields(t *testing.T) {
	// Test QuicConfig
	quicConfig := &QuicConfig{
		MaxStreams:          100,
		HandshakeTimeout:    "10s",
		IdleTimeout:         "60s",
		ConnectionMigration: true,
		Multiplexing:        true,
	}

	if err := quicConfig.Validate(); err != nil {
		t.Errorf("QuicConfig validation failed: %v", err)
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

func TestQuicConfigTimeoutValidation(t *testing.T) {
	tests := []struct {
		name             string
		handshakeTimeout string
		idleTimeout      string
		wantErr          bool
	}{
		{"valid timeouts", "10s", "60s", false},
		{"valid timeouts with minutes", "1m", "5m", false},
		{"invalid handshake timeout", "invalid", "60s", true},
		{"invalid idle timeout", "10s", "invalid", true},
		{"empty timeouts", "", "", false}, // empty is allowed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &QuicConfig{
				MaxStreams:          100,
				HandshakeTimeout:    tt.handshakeTimeout,
				IdleTimeout:         tt.idleTimeout,
				ConnectionMigration: true,
				Multiplexing:        true,
			}

			err := config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("QuicConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
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
