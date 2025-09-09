package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewIPValidator(t *testing.T) {
	logger := zap.NewNop()
	validator := NewIPValidator(logger)

	assert.NotNil(t, validator)
	assert.NotNil(t, validator.allowedCIDRs)
	assert.NotNil(t, validator.allowedIPs)
	assert.Equal(t, logger, validator.logger)
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
			name:        "valid IPv6 CIDR",
			cidr:        "2001:db8::/32",
			expectError: false,
		},
		{
			name:        "invalid CIDR",
			cidr:        "invalid-cidr",
			expectError: true,
		},
		{
			name:        "empty CIDR",
			cidr:        "",
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
			ip:          "2001:db8::1",
			expectError: false,
		},
		{
			name:        "invalid IP",
			ip:          "invalid-ip",
			expectError: true,
		},
		{
			name:        "empty IP",
			ip:          "",
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

func TestIPValidator_IsIPAllowed(t *testing.T) {
	logger := zap.NewNop()
	validator := NewIPValidator(logger)

	// Add allowed IPs and CIDRs
	validator.AddAllowedIP("192.168.1.1")
	validator.AddAllowedCIDR("10.0.0.0/8")
	validator.AddAllowedCIDR("2001:db8::/32")

	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{
			name:     "exact IP match",
			ip:       "192.168.1.1",
			expected: true,
		},
		{
			name:     "IP in CIDR range",
			ip:       "10.0.0.1",
			expected: true,
		},
		{
			name:     "IP in IPv6 CIDR range",
			ip:       "2001:db8::1",
			expected: true,
		},
		{
			name:     "IP not allowed",
			ip:       "172.16.0.1",
			expected: false,
		},
		{
			name:     "invalid IP",
			ip:       "invalid-ip",
			expected: false,
		},
		{
			name:     "empty IP",
			ip:       "",
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

func TestIPValidator_ValidateWireGuardEndpoint(t *testing.T) {
	logger := zap.NewNop()
	validator := NewIPValidator(logger)

	// Add allowed IPs
	validator.AddAllowedIP("192.168.1.1")
	validator.AddAllowedIP("10.0.0.1")

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
			name:        "valid endpoint with allowed IP and different port",
			endpoint:    "10.0.0.1:8080",
			expectError: false,
		},
		{
			name:        "invalid endpoint format",
			endpoint:    "192.168.1.1",
			expectError: true,
		},
		{
			name:        "endpoint with not allowed IP",
			endpoint:    "172.16.0.1:51820",
			expectError: true,
		},
		{
			name:        "endpoint with invalid port",
			endpoint:    "192.168.1.1:99999",
			expectError: true,
		},
		{
			name:        "endpoint with zero port",
			endpoint:    "192.168.1.1:0",
			expectError: true,
		},
		{
			name:        "empty endpoint",
			endpoint:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateWireGuardEndpoint(tt.endpoint)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIPValidator_ValidateAllowedIPs(t *testing.T) {
	logger := zap.NewNop()
	validator := NewIPValidator(logger)

	tests := []struct {
		name        string
		allowedIPs  []string
		expectError bool
	}{
		{
			name:        "valid IPs and CIDRs",
			allowedIPs:  []string{"192.168.1.1", "10.0.0.0/8", "2001:db8::1"},
			expectError: false,
		},
		{
			name:        "empty list",
			allowedIPs:  []string{},
			expectError: false,
		},
		{
			name:        "invalid IP",
			allowedIPs:  []string{"invalid-ip"},
			expectError: true,
		},
		{
			name:        "invalid CIDR",
			allowedIPs:  []string{"192.168.1.0/33"},
			expectError: true,
		},
		{
			name:        "empty IP in list",
			allowedIPs:  []string{"192.168.1.1", ""},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateAllowedIPs(tt.allowedIPs)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIPValidator_LoadFromNetworkConfig(t *testing.T) {
	// Simple test without complex setup
	validator := NewIPValidator(zap.NewNop())
	
	// Test nil config
	err := validator.LoadFromNetworkConfig(nil)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(validator.GetAllowedCIDRs()))
	
	// Test valid config
	config := &NetworkConfig{
		Subnet: "192.168.1.0/24",
		DNS:    []string{"8.8.8.8"},
		MTU:    1500,
	}
	err = validator.LoadFromNetworkConfig(config)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(validator.GetAllowedCIDRs()))
}

func TestIPValidator_GetAllowedIPs(t *testing.T) {
	logger := zap.NewNop()
	validator := NewIPValidator(logger)

	// Add some allowed IPs
	validator.AddAllowedIP("192.168.1.1")
	validator.AddAllowedIP("10.0.0.1")

	allowedIPs := validator.GetAllowedIPs()
	assert.Len(t, allowedIPs, 2)
	assert.Contains(t, allowedIPs, "192.168.1.1")
	assert.Contains(t, allowedIPs, "10.0.0.1")
}

func TestIPValidator_GetAllowedCIDRs(t *testing.T) {
	logger := zap.NewNop()
	validator := NewIPValidator(logger)

	// Add some allowed CIDRs
	validator.AddAllowedCIDR("192.168.1.0/24")
	validator.AddAllowedCIDR("10.0.0.0/8")

	allowedCIDRs := validator.GetAllowedCIDRs()
	assert.Len(t, allowedCIDRs, 2)
	assert.Contains(t, allowedCIDRs, "192.168.1.0/24")
	assert.Contains(t, allowedCIDRs, "10.0.0.0/8")
}

func TestIPValidator_GetCount(t *testing.T) {
	logger := zap.NewNop()
	validator := NewIPValidator(logger)

	// Initially no allowed IPs or CIDRs
	assert.Equal(t, 0, validator.GetCount())

	// Add some IPs and CIDRs
	validator.AddAllowedIP("192.168.1.1")
	validator.AddAllowedIP("10.0.0.1")
	validator.AddAllowedCIDR("172.16.0.0/16")

	assert.Equal(t, 3, validator.GetCount())
}

func TestIPValidator_Clear(t *testing.T) {
	logger := zap.NewNop()
	validator := NewIPValidator(logger)

	// Add some IPs and CIDRs
	validator.AddAllowedIP("192.168.1.1")
	validator.AddAllowedCIDR("10.0.0.0/8")

	assert.Equal(t, 2, validator.GetCount())

	// Clear
	validator.Clear()

	assert.Equal(t, 0, validator.GetCount())
	assert.Len(t, validator.GetAllowedIPs(), 0)
	assert.Len(t, validator.GetAllowedCIDRs(), 0)
}

func TestIsValidPort(t *testing.T) {
	tests := []struct {
		name     string
		port     string
		expected bool
	}{
		{
			name:     "valid port",
			port:     "8080",
			expected: true,
		},
		{
			name:     "valid port 1",
			port:     "1",
			expected: true,
		},
		{
			name:     "valid port 65535",
			port:     "65535",
			expected: true,
		},
		{
			name:     "invalid port 0",
			port:     "0",
			expected: false,
		},
		{
			name:     "invalid port 65536",
			port:     "65536",
			expected: false,
		},
		{
			name:     "invalid port string",
			port:     "abc",
			expected: false,
		},
		{
			name:     "empty port",
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
			name:     "valid IP",
			input:    "192.168.1.1",
			expected: true,
		},
		{
			name:     "valid IPv6",
			input:    "2001:db8::1",
			expected: true,
		},
		{
			name:     "valid CIDR",
			input:    "192.168.1.0/24",
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
