package token

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWireGuardConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  WireGuardConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config with all fields",
			config: WireGuardConfig{
				InterfaceName: "wg0",
				ListenPort:    51820,
				Address:       "10.0.0.1/24",
				MTU:           1420,
				AllowedIPs:    []string{"10.0.0.0/24", "192.168.1.0/24"},
				PrivateKey:    "private_key_here",
				PublicKey:     "public_key_here",
				Endpoint:      "example.com:51820",
			},
			wantErr: false,
		},
		{
			name: "valid minimal config",
			config: WireGuardConfig{
				InterfaceName: "wg0",
				ListenPort:    51820,
				Address:       "10.0.0.1/24",
				MTU:           1420,
			},
			wantErr: false,
		},
		{
			name: "missing interface_name",
			config: WireGuardConfig{
				ListenPort: 51820,
				Address:    "10.0.0.1/24",
				MTU:        1420,
			},
			wantErr: true,
			errMsg:  "interface_name is required",
		},
		{
			name: "empty interface_name",
			config: WireGuardConfig{
				InterfaceName: "",
				ListenPort:    51820,
				Address:       "10.0.0.1/24",
				MTU:           1420,
			},
			wantErr: true,
			errMsg:  "interface_name is required",
		},
		{
			name: "zero listen_port",
			config: WireGuardConfig{
				InterfaceName: "wg0",
				ListenPort:    0,
				Address:       "10.0.0.1/24",
				MTU:           1420,
			},
			wantErr: true,
			errMsg:  "listen_port must be between 1 and 65535",
		},
		{
			name: "negative listen_port",
			config: WireGuardConfig{
				InterfaceName: "wg0",
				ListenPort:    -1,
				Address:       "10.0.0.1/24",
				MTU:           1420,
			},
			wantErr: true,
			errMsg:  "listen_port must be between 1 and 65535",
		},
		{
			name: "too large listen_port",
			config: WireGuardConfig{
				InterfaceName: "wg0",
				ListenPort:    65536,
				Address:       "10.0.0.1/24",
				MTU:           1420,
			},
			wantErr: true,
			errMsg:  "listen_port must be between 1 and 65535",
		},
		{
			name: "valid listen_port range",
			config: WireGuardConfig{
				InterfaceName: "wg0",
				ListenPort:    1,
				Address:       "10.0.0.1/24",
				MTU:           1420,
			},
			wantErr: false,
		},
		{
			name: "maximum valid listen_port",
			config: WireGuardConfig{
				InterfaceName: "wg0",
				ListenPort:    65535,
				Address:       "10.0.0.1/24",
				MTU:           1420,
			},
			wantErr: false,
		},
		{
			name: "missing address",
			config: WireGuardConfig{
				InterfaceName: "wg0",
				ListenPort:    51820,
				Address:       "",
				MTU:           1420,
			},
			wantErr: true,
			errMsg:  "address is required",
		},
		{
			name: "zero MTU",
			config: WireGuardConfig{
				InterfaceName: "wg0",
				ListenPort:    51820,
				Address:       "10.0.0.1/24",
				MTU:           0,
			},
			wantErr: true,
			errMsg:  "mtu must be positive",
		},
		{
			name: "negative MTU",
			config: WireGuardConfig{
				InterfaceName: "wg0",
				ListenPort:    51820,
				Address:       "10.0.0.1/24",
				MTU:           -1420,
			},
			wantErr: true,
			errMsg:  "mtu must be positive",
		},
		{
			name: "valid MTU values",
			config: WireGuardConfig{
				InterfaceName: "wg0",
				ListenPort:    51820,
				Address:       "10.0.0.1/24",
				MTU:           1420,
			},
			wantErr: false,
		},
		{
			name: "empty allowed_ips entry",
			config: WireGuardConfig{
				InterfaceName: "wg0",
				ListenPort:    51820,
				Address:       "10.0.0.1/24",
				MTU:           1420,
				AllowedIPs:    []string{"10.0.0.0/24", "", "192.168.1.0/24"},
			},
			wantErr: true,
			errMsg:  "allowed_ips[1] cannot be empty",
		},
		{
			name: "invalid allowed_ips entry - too short",
			config: WireGuardConfig{
				InterfaceName: "wg0",
				ListenPort:    51820,
				Address:       "10.0.0.1/24",
				MTU:           1420,
				AllowedIPs:    []string{"10.0.0.0/24", "1.1.1", "192.168.1.0/24"},
			},
			wantErr: true,
			errMsg:  "allowed_ips[1] appears to be invalid CIDR",
		},
		{
			name: "valid allowed_ips",
			config: WireGuardConfig{
				InterfaceName: "wg0",
				ListenPort:    51820,
				Address:       "10.0.0.1/24",
				MTU:           1420,
				AllowedIPs:    []string{"10.0.0.0/24", "192.168.1.0/24", "0.0.0.0/0"},
			},
			wantErr: false,
		},
		{
			name: "empty allowed_ips array",
			config: WireGuardConfig{
				InterfaceName: "wg0",
				ListenPort:    51820,
				Address:       "10.0.0.1/24",
				MTU:           1420,
				AllowedIPs:    []string{},
			},
			wantErr: false,
		},
		{
			name: "nil allowed_ips array",
			config: WireGuardConfig{
				InterfaceName: "wg0",
				ListenPort:    51820,
				Address:       "10.0.0.1/24",
				MTU:           1420,
				AllowedIPs:    nil,
			},
			wantErr: false,
		},
		{
			name: "invalid endpoint - too short",
			config: WireGuardConfig{
				InterfaceName: "wg0",
				ListenPort:    51820,
				Address:       "10.0.0.1/24",
				MTU:           1420,
				Endpoint:      "1.1:1",
			},
			wantErr: true,
			errMsg:  "endpoint appears to be invalid format",
		},
		{
			name: "valid endpoint",
			config: WireGuardConfig{
				InterfaceName: "wg0",
				ListenPort:    51820,
				Address:       "10.0.0.1/24",
				MTU:           1420,
				Endpoint:      "example.com:51820",
			},
			wantErr: false,
		},
		{
			name: "empty endpoint",
			config: WireGuardConfig{
				InterfaceName: "wg0",
				ListenPort:    51820,
				Address:       "10.0.0.1/24",
				MTU:           1420,
				Endpoint:      "",
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

func TestWireGuardConfig_StructFields(t *testing.T) {
	t.Run("struct field access", func(t *testing.T) {
		config := WireGuardConfig{
			InterfaceName: "wg1",
			ListenPort:    51821,
			Address:       "192.168.1.1/24",
			MTU:           1500,
			AllowedIPs:    []string{"192.168.1.0/24", "10.0.0.0/16"},
			PrivateKey:    "test_private_key",
			PublicKey:     "test_public_key",
			Endpoint:      "test.example.com:51821",
		}

		assert.Equal(t, "wg1", config.InterfaceName)
		assert.Equal(t, 51821, config.ListenPort)
		assert.Equal(t, "192.168.1.1/24", config.Address)
		assert.Equal(t, 1500, config.MTU)
		assert.Len(t, config.AllowedIPs, 2)
		assert.Equal(t, "192.168.1.0/24", config.AllowedIPs[0])
		assert.Equal(t, "10.0.0.0/16", config.AllowedIPs[1])
		assert.Equal(t, "test_private_key", config.PrivateKey)
		assert.Equal(t, "test_public_key", config.PublicKey)
		assert.Equal(t, "test.example.com:51821", config.Endpoint)
	})

	t.Run("empty struct", func(t *testing.T) {
		config := WireGuardConfig{}

		assert.Empty(t, config.InterfaceName)
		assert.Equal(t, 0, config.ListenPort)
		assert.Empty(t, config.Address)
		assert.Equal(t, 0, config.MTU)
		assert.Nil(t, config.AllowedIPs)
		assert.Empty(t, config.PrivateKey)
		assert.Empty(t, config.PublicKey)
		assert.Empty(t, config.Endpoint)
	})
}

func TestWireGuardConfig_EdgeCases(t *testing.T) {
	t.Run("minimum valid values", func(t *testing.T) {
		config := WireGuardConfig{
			InterfaceName: "wg0",
			ListenPort:    1,
			Address:       "10.0.0.1/24",
			MTU:           1,
		}
		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("maximum valid values", func(t *testing.T) {
		config := WireGuardConfig{
			InterfaceName: "wg0",
			ListenPort:    65535,
			Address:       "10.0.0.1/24",
			MTU:           65535,
		}
		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("single allowed_ip", func(t *testing.T) {
		config := WireGuardConfig{
			InterfaceName: "wg0",
			ListenPort:    51820,
			Address:       "10.0.0.1/24",
			MTU:           1420,
			AllowedIPs:    []string{"0.0.0.0/0"},
		}
		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("multiple allowed_ips", func(t *testing.T) {
		config := WireGuardConfig{
			InterfaceName: "wg0",
			ListenPort:    51820,
			Address:       "10.0.0.1/24",
			MTU:           1420,
			AllowedIPs:    []string{"10.0.0.0/24", "192.168.1.0/24", "172.16.0.0/16", "0.0.0.0/0"},
		}
		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("endpoint with IP address", func(t *testing.T) {
		config := WireGuardConfig{
			InterfaceName: "wg0",
			ListenPort:    51820,
			Address:       "10.0.0.1/24",
			MTU:           1420,
			Endpoint:      "192.168.1.1:51820",
		}
		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("endpoint with domain name", func(t *testing.T) {
		config := WireGuardConfig{
			InterfaceName: "wg0",
			ListenPort:    51820,
			Address:       "10.0.0.1/24",
			MTU:           1420,
			Endpoint:      "wireguard.example.com:51820",
		}
		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("optional fields only", func(t *testing.T) {
		config := WireGuardConfig{
			InterfaceName: "wg0",
			ListenPort:    51820,
			Address:       "10.0.0.1/24",
			MTU:           1420,
			PrivateKey:    "optional_private_key",
			PublicKey:     "optional_public_key",
			Endpoint:      "optional.example.com:51820",
		}
		err := config.Validate()
		assert.NoError(t, err)
	})
}
