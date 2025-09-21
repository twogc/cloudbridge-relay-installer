package api

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewIPValidator(t *testing.T) {
	logger := zap.NewNop()

	validator := NewIPValidator(logger)

	assert.NotNil(t, validator)
}

func TestIPValidator_AddAllowedIP(t *testing.T) {
	logger := zap.NewNop()
	validator := NewIPValidator(logger)

	tests := []struct {
		name        string
		ip          string
		expectError bool
	}{
		{
			name:        "valid IPv4",
			ip:          "192.168.1.1",
			expectError: false,
		},
		{
			name:        "valid IPv6",
			ip:          "::1",
			expectError: false,
		},
		{
			name:        "invalid IP",
			ip:          "invalid-ip",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.AddAllowedIP(tt.ip)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIPValidator_AddAllowedCIDR(t *testing.T) {
	logger := zap.NewNop()
	validator := NewIPValidator(logger)

	tests := []struct {
		name        string
		cidr        string
		expectError bool
	}{
		{
			name:        "valid CIDR",
			cidr:        "192.168.1.0/24",
			expectError: false,
		},
		{
			name:        "invalid CIDR",
			cidr:        "invalid-cidr",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.AddAllowedCIDR(tt.cidr)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIPValidator_IsAllowed(t *testing.T) {
	logger := zap.NewNop()
	validator := NewIPValidator(logger)

	// Add allowed IPs
	validator.AddAllowedIP("192.168.1.1")
	validator.AddAllowedCIDR("10.0.0.0/8")

	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{
			name:     "allowed IP",
			ip:       "192.168.1.1",
			expected: true,
		},
		{
			name:     "IP in allowed CIDR",
			ip:       "10.0.0.1",
			expected: true,
		},
		{
			name:     "not allowed IP",
			ip:       "172.16.0.1",
			expected: false,
		},
		{
			name:     "invalid IP",
			ip:       "invalid-ip",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.IsIPAllowed(tt.ip)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIPValidator_ValidateAllowedIPs(t *testing.T) {
	logger := zap.NewNop()
	validator := NewIPValidator(logger)

	// Add allowed IPs
	validator.AddAllowedIP("192.168.1.1")
	validator.AddAllowedCIDR("10.0.0.0/8")

	tests := []struct {
		name        string
		ips         []string
		expectError bool
	}{
		{
			name:        "all allowed IPs",
			ips:         []string{"192.168.1.1", "10.0.0.1"},
			expectError: false,
		},
		{
			name:        "contains invalid IP format",
			ips:         []string{"192.168.1.1", "invalid-ip"},
			expectError: true,
		},
		{
			name:        "empty list",
			ips:         []string{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateAllowedIPs(tt.ips)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIPValidator_ValidateP2PQUICEndpoint(t *testing.T) {
	logger := zap.NewNop()
	validator := NewIPValidator(logger)

	// Add allowed IPs
	validator.AddAllowedIP("192.168.1.1")
	validator.AddAllowedCIDR("10.0.0.0/8")

	tests := []struct {
		name        string
		endpoint    string
		expectError bool
	}{
		{
			name:        "valid endpoint with allowed IP",
			endpoint:    "192.168.1.1:51820",
			expectError: false,
		},
		{
			name:        "valid endpoint with IP in CIDR",
			endpoint:    "10.0.0.1:51820",
			expectError: false,
		},
		{
			name:        "invalid endpoint format - missing port",
			endpoint:    "192.168.1.1",
			expectError: true,
		},
		{
			name:        "invalid endpoint format - too many parts",
			endpoint:    "192.168.1.1:51820:extra",
			expectError: true,
		},
		{
			name:        "invalid port",
			endpoint:    "192.168.1.1:99999",
			expectError: true,
		},
		{
			name:        "not allowed IP",
			endpoint:    "172.16.0.1:51820",
			expectError: true,
		},
		{
			name:        "invalid IP",
			endpoint:    "invalid-ip:51820",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateP2PQUICEndpoint(tt.endpoint)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIPValidator_LoadFromNetworkConfig(t *testing.T) {
	logger := zap.NewNop()
	validator := NewIPValidator(logger)

	tests := []struct {
		name        string
		config      *NetworkConfig
		expectError bool
	}{
		{
			name: "valid network config",
			config: &NetworkConfig{
				Subnet: "192.168.1.0/24",
				DNS:    []string{"8.8.8.8", "8.8.4.4"},
				MTU:    1500,
			},
			expectError: false,
		},
		{
			name:        "nil config",
			config:      nil,
			expectError: false,
		},
		{
			name: "invalid subnet CIDR",
			config: &NetworkConfig{
				Subnet: "invalid-cidr",
				DNS:    []string{"8.8.8.8"},
				MTU:    1500,
			},
			expectError: true,
		},
		{
			name: "empty subnet",
			config: &NetworkConfig{
				Subnet: "",
				DNS:    []string{"8.8.8.8"},
				MTU:    1500,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.LoadFromNetworkConfig(tt.config)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIPValidator_GetAllowedIPs(t *testing.T) {
	logger := zap.NewNop()
	validator := NewIPValidator(logger)

	// Initially empty
	ips := validator.GetAllowedIPs()
	assert.Empty(t, ips)

	// Add some IPs
	validator.AddAllowedIP("192.168.1.1")
	validator.AddAllowedIP("192.168.1.2")

	ips = validator.GetAllowedIPs()
	assert.Len(t, ips, 2)
	assert.Contains(t, ips, "192.168.1.1")
	assert.Contains(t, ips, "192.168.1.2")
}

func TestIPValidator_GetAllowedCIDRs(t *testing.T) {
	logger := zap.NewNop()
	validator := NewIPValidator(logger)

	// Initially empty
	cidrs := validator.GetAllowedCIDRs()
	assert.Empty(t, cidrs)

	// Add some CIDRs
	validator.AddAllowedCIDR("192.168.1.0/24")
	validator.AddAllowedCIDR("10.0.0.0/8")

	cidrs = validator.GetAllowedCIDRs()
	assert.Len(t, cidrs, 2)
	assert.Contains(t, cidrs, "192.168.1.0/24")
	assert.Contains(t, cidrs, "10.0.0.0/8")
}

func TestIPValidator_GetCount(t *testing.T) {
	logger := zap.NewNop()
	validator := NewIPValidator(logger)

	// Initially zero
	count := validator.GetCount()
	assert.Equal(t, 0, count)

	// Add some IPs and CIDRs
	validator.AddAllowedIP("192.168.1.1")
	validator.AddAllowedIP("192.168.1.2")
	validator.AddAllowedCIDR("10.0.0.0/8")

	count = validator.GetCount()
	assert.Equal(t, 3, count)
}

func TestIPValidator_Clear(t *testing.T) {
	logger := zap.NewNop()
	validator := NewIPValidator(logger)

	// Add some data
	validator.AddAllowedIP("192.168.1.1")
	validator.AddAllowedCIDR("10.0.0.0/8")

	// Verify data exists
	assert.Equal(t, 2, validator.GetCount())

	// Clear
	validator.Clear()

	// Verify data is cleared
	assert.Equal(t, 0, validator.GetCount())
	assert.Empty(t, validator.GetAllowedIPs())
	assert.Empty(t, validator.GetAllowedCIDRs())
}

func TestIPValidator_NewIPValidatorMetrics(t *testing.T) {
	metrics := NewIPValidatorMetrics()

	assert.NotNil(t, metrics)
	assert.NotNil(t, metrics.checksTotal)
	assert.NotNil(t, metrics.allowedTotal)
	assert.NotNil(t, metrics.deniedTotal)
	assert.NotNil(t, metrics.violationsTotal)
}

func TestIPValidator_RegisterMetrics(t *testing.T) {
	logger := zap.NewNop()
	validator := NewIPValidator(logger)

	// This should panic when reg is nil
	assert.Panics(t, func() {
		validator.RegisterMetrics(nil)
	})
}

func TestIsValidPort(t *testing.T) {
	tests := []struct {
		name     string
		port     string
		expected bool
	}{
		{
			name:     "valid port - 80",
			port:     "80",
			expected: true,
		},
		{
			name:     "valid port - 443",
			port:     "443",
			expected: true,
		},
		{
			name:     "valid port - 51820",
			port:     "51820",
			expected: true,
		},
		{
			name:     "valid port - 65535",
			port:     "65535",
			expected: true,
		},
		{
			name:     "invalid port - 0",
			port:     "0",
			expected: false,
		},
		{
			name:     "invalid port - 65536",
			port:     "65536",
			expected: false,
		},
		{
			name:     "invalid port - negative",
			port:     "-1",
			expected: false,
		},
		{
			name:     "invalid port - non-numeric",
			port:     "abc",
			expected: false,
		},
		{
			name:     "invalid port - empty",
			port:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidPort(tt.port)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidCIDROrIP(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid CIDR",
			input:    "192.168.1.0/24",
			expected: true,
		},
		{
			name:     "valid IPv4",
			input:    "192.168.1.1",
			expected: true,
		},
		{
			name:     "valid IPv6",
			input:    "::1",
			expected: true,
		},
		{
			name:     "valid IPv6 CIDR",
			input:    "2001:db8::/32",
			expected: true,
		},
		{
			name:     "invalid input",
			input:    "invalid",
			expected: false,
		},
		{
			name:     "empty input",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidCIDROrIP(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIPValidator_ConcurrentAccess(t *testing.T) {
	logger := zap.NewNop()
	validator := NewIPValidator(logger)

	// Test concurrent access
	done := make(chan bool, 10)

	// Start multiple goroutines
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			// Add IPs
			validator.AddAllowedIP(fmt.Sprintf("192.168.1.%d", id))

			// Check if allowed
			validator.IsIPAllowed(fmt.Sprintf("192.168.1.%d", id))

			// Get stats
			validator.GetCount()
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify final state
	assert.Equal(t, 10, validator.GetCount())
}
