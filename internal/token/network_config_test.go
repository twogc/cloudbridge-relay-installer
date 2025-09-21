package token

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNetworkConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  NetworkConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config with all fields",
			config: NetworkConfig{
				Subnet:        "10.0.0.0/16",
				Gateway:       "10.0.0.1",
				DNS:           []string{"8.8.8.8", "1.1.1.1"},
				FirewallRules: []string{"allow_ssh", "allow_http"},
				MTU:           1500,
				EnableIPv6:    true,
			},
			wantErr: false,
		},
		{
			name: "valid minimal config",
			config: NetworkConfig{
				Subnet: "192.168.1.0/24",
				MTU:    1500,
			},
			wantErr: false,
		},
		{
			name: "missing subnet",
			config: NetworkConfig{
				Gateway: "10.0.0.1",
				MTU:     1500,
			},
			wantErr: true,
			errMsg:  "subnet is required",
		},
		{
			name: "empty subnet",
			config: NetworkConfig{
				Subnet: "",
				MTU:    1500,
			},
			wantErr: true,
			errMsg:  "subnet is required",
		},
		{
			name: "invalid subnet - too short",
			config: NetworkConfig{
				Subnet: "10.0.0",
				MTU:    1500,
			},
			wantErr: true,
			errMsg:  "subnet appears to be invalid CIDR",
		},
		{
			name: "valid subnet formats",
			config: NetworkConfig{
				Subnet: "0.0.0.0/0",
				MTU:    1500,
			},
			wantErr: false,
		},
		{
			name: "empty DNS entry",
			config: NetworkConfig{
				Subnet: "10.0.0.0/16",
				DNS:    []string{"8.8.8.8", "", "1.1.1.1"},
				MTU:    1500,
			},
			wantErr: true,
			errMsg:  "dns[1] cannot be empty",
		},
		{
			name: "valid DNS entries",
			config: NetworkConfig{
				Subnet: "10.0.0.0/16",
				DNS:    []string{"8.8.8.8", "1.1.1.1", "9.9.9.9"},
				MTU:    1500,
			},
			wantErr: false,
		},
		{
			name: "empty DNS array",
			config: NetworkConfig{
				Subnet: "10.0.0.0/16",
				DNS:    []string{},
				MTU:    1500,
			},
			wantErr: false,
		},
		{
			name: "nil DNS array",
			config: NetworkConfig{
				Subnet: "10.0.0.0/16",
				DNS:    nil,
				MTU:    1500,
			},
			wantErr: false,
		},
		{
			name: "zero MTU",
			config: NetworkConfig{
				Subnet: "10.0.0.0/16",
				MTU:    0,
			},
			wantErr: true,
			errMsg:  "mtu must be positive",
		},
		{
			name: "negative MTU",
			config: NetworkConfig{
				Subnet: "10.0.0.0/16",
				MTU:    -1500,
			},
			wantErr: true,
			errMsg:  "mtu must be positive",
		},
		{
			name: "valid MTU values",
			config: NetworkConfig{
				Subnet: "10.0.0.0/16",
				MTU:    1500,
			},
			wantErr: false,
		},
		{
			name: "large MTU value",
			config: NetworkConfig{
				Subnet: "10.0.0.0/16",
				MTU:    9000,
			},
			wantErr: false,
		},
		{
			name: "small MTU value",
			config: NetworkConfig{
				Subnet: "10.0.0.0/16",
				MTU:    576,
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

func TestNetworkConfig_StructFields(t *testing.T) {
	t.Run("struct field access", func(t *testing.T) {
		config := NetworkConfig{
			Subnet:        "192.168.1.0/24",
			Gateway:       "192.168.1.1",
			DNS:           []string{"8.8.8.8", "1.1.1.1"},
			FirewallRules: []string{"allow_ssh", "allow_http", "allow_https"},
			MTU:           1500,
			EnableIPv6:    true,
		}

		assert.Equal(t, "192.168.1.0/24", config.Subnet)
		assert.Equal(t, "192.168.1.1", config.Gateway)
		assert.Len(t, config.DNS, 2)
		assert.Equal(t, "8.8.8.8", config.DNS[0])
		assert.Equal(t, "1.1.1.1", config.DNS[1])
		assert.Len(t, config.FirewallRules, 3)
		assert.Equal(t, "allow_ssh", config.FirewallRules[0])
		assert.Equal(t, "allow_http", config.FirewallRules[1])
		assert.Equal(t, "allow_https", config.FirewallRules[2])
		assert.Equal(t, 1500, config.MTU)
		assert.True(t, config.EnableIPv6)
	})

	t.Run("empty struct", func(t *testing.T) {
		config := NetworkConfig{}

		assert.Empty(t, config.Subnet)
		assert.Empty(t, config.Gateway)
		assert.Nil(t, config.DNS)
		assert.Nil(t, config.FirewallRules)
		assert.Equal(t, 0, config.MTU)
		assert.False(t, config.EnableIPv6)
	})

	t.Run("IPv6 disabled", func(t *testing.T) {
		config := NetworkConfig{
			Subnet:     "10.0.0.0/16",
			MTU:        1500,
			EnableIPv6: false,
		}

		assert.False(t, config.EnableIPv6)
	})

	t.Run("IPv6 enabled", func(t *testing.T) {
		config := NetworkConfig{
			Subnet:     "10.0.0.0/16",
			MTU:        1500,
			EnableIPv6: true,
		}

		assert.True(t, config.EnableIPv6)
	})
}

func TestNetworkConfig_EdgeCases(t *testing.T) {
	t.Run("minimum valid subnet", func(t *testing.T) {
		config := NetworkConfig{
			Subnet: "0.0.0.0/0",
			MTU:    1,
		}
		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("maximum valid subnet", func(t *testing.T) {
		config := NetworkConfig{
			Subnet: "255.255.255.255/32",
			MTU:    65535,
		}
		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("single DNS entry", func(t *testing.T) {
		config := NetworkConfig{
			Subnet: "10.0.0.0/16",
			DNS:    []string{"8.8.8.8"},
			MTU:    1500,
		}
		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("multiple DNS entries", func(t *testing.T) {
		config := NetworkConfig{
			Subnet: "10.0.0.0/16",
			DNS:    []string{"8.8.8.8", "1.1.1.1", "9.9.9.9", "208.67.222.222"},
			MTU:    1500,
		}
		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("empty firewall rules", func(t *testing.T) {
		config := NetworkConfig{
			Subnet:        "10.0.0.0/16",
			FirewallRules: []string{},
			MTU:           1500,
		}
		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("nil firewall rules", func(t *testing.T) {
		config := NetworkConfig{
			Subnet:        "10.0.0.0/16",
			FirewallRules: nil,
			MTU:           1500,
		}
		err := config.Validate()
		assert.NoError(t, err)
	})
}
