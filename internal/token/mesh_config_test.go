package token

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMeshConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  MeshConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config with all fields",
			config: MeshConfig{
				NetworkID:           "mesh-network-001",
				Subnet:              "10.0.0.0/16",
				RegistryURL:         "https://registry.example.com",
				HeartbeatInterval:   30 * time.Second,
				MaxPeers:            10,
				RoutingStrategy:     "performance_optimal",
				EnableAutoDiscovery: true,
				TrustLevel:          "basic",
			},
			wantErr: false,
		},
		{
			name: "valid minimal config",
			config: MeshConfig{
				RegistryURL:       "https://registry.example.com",
				HeartbeatInterval: 30 * time.Second,
				MaxPeers:          5,
			},
			wantErr: false,
		},
		{
			name: "missing registry_url",
			config: MeshConfig{
				HeartbeatInterval: 30 * time.Second,
				MaxPeers:          5,
			},
			wantErr: true,
			errMsg:  "registry_url is required",
		},
		{
			name: "empty registry_url",
			config: MeshConfig{
				RegistryURL:       "",
				HeartbeatInterval: 30 * time.Second,
				MaxPeers:          5,
			},
			wantErr: true,
			errMsg:  "registry_url is required",
		},
		{
			name: "zero heartbeat_interval",
			config: MeshConfig{
				RegistryURL:       "https://registry.example.com",
				HeartbeatInterval: 0,
				MaxPeers:          5,
			},
			wantErr: true,
			errMsg:  "heartbeat_interval must be positive",
		},
		{
			name: "negative heartbeat_interval",
			config: MeshConfig{
				RegistryURL:       "https://registry.example.com",
				HeartbeatInterval: -30 * time.Second,
				MaxPeers:          5,
			},
			wantErr: true,
			errMsg:  "heartbeat_interval must be positive",
		},
		{
			name: "zero max_peers",
			config: MeshConfig{
				RegistryURL:       "https://registry.example.com",
				HeartbeatInterval: 30 * time.Second,
				MaxPeers:          0,
			},
			wantErr: true,
			errMsg:  "max_peers must be positive",
		},
		{
			name: "negative max_peers",
			config: MeshConfig{
				RegistryURL:       "https://registry.example.com",
				HeartbeatInterval: 30 * time.Second,
				MaxPeers:          -5,
			},
			wantErr: true,
			errMsg:  "max_peers must be positive",
		},
		{
			name: "short network_id",
			config: MeshConfig{
				NetworkID:         "ab",
				RegistryURL:       "https://registry.example.com",
				HeartbeatInterval: 30 * time.Second,
				MaxPeers:          5,
			},
			wantErr: true,
			errMsg:  "network_id must be at least 3 characters long",
		},
		{
			name: "valid network_id",
			config: MeshConfig{
				NetworkID:         "abc",
				RegistryURL:       "https://registry.example.com",
				HeartbeatInterval: 30 * time.Second,
				MaxPeers:          5,
			},
			wantErr: false,
		},
		{
			name: "invalid subnet - too short",
			config: MeshConfig{
				Subnet:            "10.0.0",
				RegistryURL:       "https://registry.example.com",
				HeartbeatInterval: 30 * time.Second,
				MaxPeers:          5,
			},
			wantErr: true,
			errMsg:  "subnet appears to be invalid CIDR",
		},
		{
			name: "invalid subnet - no slash",
			config: MeshConfig{
				Subnet:            "10.0.0.0",
				RegistryURL:       "https://registry.example.com",
				HeartbeatInterval: 30 * time.Second,
				MaxPeers:          5,
			},
			wantErr: true,
			errMsg:  "subnet must be in CIDR notation",
		},
		{
			name: "valid subnet",
			config: MeshConfig{
				Subnet:            "10.0.0.0/16",
				RegistryURL:       "https://registry.example.com",
				HeartbeatInterval: 30 * time.Second,
				MaxPeers:          5,
			},
			wantErr: false,
		},
		{
			name: "invalid routing_strategy",
			config: MeshConfig{
				RegistryURL:       "https://registry.example.com",
				HeartbeatInterval: 30 * time.Second,
				MaxPeers:          5,
				RoutingStrategy:   "invalid_strategy",
			},
			wantErr: true,
			errMsg:  "invalid routing_strategy",
		},
		{
			name: "valid routing_strategies",
			config: MeshConfig{
				RegistryURL:       "https://registry.example.com",
				HeartbeatInterval: 30 * time.Second,
				MaxPeers:          5,
				RoutingStrategy:   "wireguard_first",
			},
			wantErr: false,
		},
		{
			name: "invalid trust_level",
			config: MeshConfig{
				RegistryURL:       "https://registry.example.com",
				HeartbeatInterval: 30 * time.Second,
				MaxPeers:          5,
				TrustLevel:        "invalid_trust",
			},
			wantErr: true,
			errMsg:  "invalid trust_level",
		},
		{
			name: "valid trust_levels",
			config: MeshConfig{
				RegistryURL:       "https://registry.example.com",
				HeartbeatInterval: 30 * time.Second,
				MaxPeers:          5,
				TrustLevel:        "full",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMeshConfig_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected MeshConfig
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid JSON with string heartbeat_interval",
			jsonData: `{
				"network_id": "mesh-network-001",
				"subnet": "10.0.0.0/16",
				"registry_url": "https://registry.example.com",
				"heartbeat_interval": "30s",
				"max_peers": 10,
				"routing_strategy": "performance_optimal",
				"enable_auto_discovery": true,
				"trust_level": "basic"
			}`,
			expected: MeshConfig{
				NetworkID:           "mesh-network-001",
				Subnet:              "10.0.0.0/16",
				RegistryURL:         "https://registry.example.com",
				HeartbeatInterval:   30 * time.Second,
				MaxPeers:            10,
				RoutingStrategy:     "performance_optimal",
				EnableAutoDiscovery: true,
				TrustLevel:          "basic",
			},
			wantErr: false,
		},
		{
			name: "valid JSON with numeric heartbeat_interval (float64)",
			jsonData: `{
				"registry_url": "https://registry.example.com",
				"heartbeat_interval": 30.0,
				"max_peers": 5
			}`,
			expected: MeshConfig{
				RegistryURL:       "https://registry.example.com",
				HeartbeatInterval: 30 * time.Second,
				MaxPeers:          5,
			},
			wantErr: false,
		},
		{
			name: "valid JSON with integer heartbeat_interval",
			jsonData: `{
				"registry_url": "https://registry.example.com",
				"heartbeat_interval": 60,
				"max_peers": 5
			}`,
			expected: MeshConfig{
				RegistryURL:       "https://registry.example.com",
				HeartbeatInterval: 60 * time.Second,
				MaxPeers:          5,
			},
			wantErr: false,
		},
		{
			name: "valid JSON with int64 heartbeat_interval",
			jsonData: `{
				"registry_url": "https://registry.example.com",
				"heartbeat_interval": 120,
				"max_peers": 5
			}`,
			expected: MeshConfig{
				RegistryURL:       "https://registry.example.com",
				HeartbeatInterval: 120 * time.Second,
				MaxPeers:          5,
			},
			wantErr: false,
		},
		{
			name: "invalid JSON format",
			jsonData: `{
				"registry_url": "https://registry.example.com",
				"heartbeat_interval": "invalid_duration",
				"max_peers": 5
			}`,
			wantErr: true,
			errMsg:  "invalid heartbeat_interval format",
		},
		{
			name: "unsupported heartbeat_interval type",
			jsonData: `{
				"registry_url": "https://registry.example.com",
				"heartbeat_interval": true,
				"max_peers": 5
			}`,
			wantErr: true,
			errMsg:  "unsupported heartbeat_interval type",
		},
		{
			name: "malformed JSON",
			jsonData: `{
				"registry_url": "https://registry.example.com",
				"heartbeat_interval": 30,
				"max_peers": 5
			`, // missing closing brace
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config MeshConfig
			err := json.Unmarshal([]byte(tt.jsonData), &config)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, config)
			}
		})
	}
}

func TestMeshConfig_UnmarshalJSON_EdgeCases(t *testing.T) {
	t.Run("empty JSON object", func(t *testing.T) {
		var config MeshConfig
		err := json.Unmarshal([]byte("{}"), &config)
		// Empty JSON object should fail because heartbeat_interval is nil
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported heartbeat_interval type")
	})

	t.Run("heartbeat_interval with different time units", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected time.Duration
		}{
			{"1s", 1 * time.Second},
			{"1m", 1 * time.Minute},
			{"1h", 1 * time.Hour},
			{"2m30s", 2*time.Minute + 30*time.Second},
			{"1h30m45s", 1*time.Hour + 30*time.Minute + 45*time.Second},
		}

		for _, tc := range testCases {
			t.Run(tc.input, func(t *testing.T) {
				jsonData := `{
					"registry_url": "https://registry.example.com",
					"heartbeat_interval": "` + tc.input + `",
					"max_peers": 5
				}`

				var config MeshConfig
				err := json.Unmarshal([]byte(jsonData), &config)
				require.NoError(t, err)
				assert.Equal(t, tc.expected, config.HeartbeatInterval)
			})
		}
	})
}

func TestMeshConfig_StructFields(t *testing.T) {
	t.Run("struct field access", func(t *testing.T) {
		config := MeshConfig{
			NetworkID:           "test-network",
			Subnet:              "192.168.1.0/24",
			RegistryURL:         "https://test.registry.com",
			HeartbeatInterval:   45 * time.Second,
			MaxPeers:            15,
			RoutingStrategy:     "geographic_optimal",
			EnableAutoDiscovery: true,
			TrustLevel:          "full",
		}

		assert.Equal(t, "test-network", config.NetworkID)
		assert.Equal(t, "192.168.1.0/24", config.Subnet)
		assert.Equal(t, "https://test.registry.com", config.RegistryURL)
		assert.Equal(t, 45*time.Second, config.HeartbeatInterval)
		assert.Equal(t, 15, config.MaxPeers)
		assert.Equal(t, "geographic_optimal", config.RoutingStrategy)
		assert.True(t, config.EnableAutoDiscovery)
		assert.Equal(t, "full", config.TrustLevel)
	})

	t.Run("empty struct", func(t *testing.T) {
		config := MeshConfig{}

		assert.Empty(t, config.NetworkID)
		assert.Empty(t, config.Subnet)
		assert.Empty(t, config.RegistryURL)
		assert.Equal(t, time.Duration(0), config.HeartbeatInterval)
		assert.Equal(t, 0, config.MaxPeers)
		assert.Empty(t, config.RoutingStrategy)
		assert.False(t, config.EnableAutoDiscovery)
		assert.Empty(t, config.TrustLevel)
	})
}
