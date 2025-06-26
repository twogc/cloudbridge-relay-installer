package tunnel

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTunnelCreation(t *testing.T) {
	config := &Config{
		ID:          "test-tunnel",
		Protocol:    "rdp",
		LocalPort:   3389,
		RemoteHost:  "internal-server",
		RemotePort:  3389,
		Options:     map[string]string{"username": "test"},
	}

	tunnel, err := NewTunnel(config)
	assert.NoError(t, err)
	assert.Equal(t, config.ID, tunnel.ID)
	assert.Equal(t, config.Protocol, tunnel.Protocol)
	assert.Equal(t, config.LocalPort, tunnel.LocalPort)
}

func TestTunnelValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				ID:          "test-tunnel",
				Protocol:    "rdp",
				LocalPort:   3389,
				RemoteHost:  "internal-server",
				RemotePort:  3389,
			},
			wantErr: false,
		},
		{
			name: "invalid port",
			config: &Config{
				ID:          "test-tunnel",
				Protocol:    "rdp",
				LocalPort:   80, // privileged port
				RemoteHost:  "internal-server",
				RemotePort:  3389,
			},
			wantErr: true,
		},
		{
			name: "invalid protocol",
			config: &Config{
				ID:          "test-tunnel",
				Protocol:    "invalid",
				LocalPort:   3389,
				RemoteHost:  "internal-server",
				RemotePort:  3389,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewTunnel(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTunnelLifecycle(t *testing.T) {
	config := &Config{
		ID:          "test-tunnel",
		Protocol:    "rdp",
		LocalPort:   3389,
		RemoteHost:  "internal-server",
		RemotePort:  3389,
	}

	tunnel, err := NewTunnel(config)
	assert.NoError(t, err)

	// Test start
	err = tunnel.Start()
	assert.NoError(t, err)
	assert.True(t, tunnel.IsRunning())

	// Test stop
	err = tunnel.Stop()
	assert.NoError(t, err)
	assert.False(t, tunnel.IsRunning())
}

func TestTunnelReconnection(t *testing.T) {
	config := &Config{
		ID:          "test-tunnel",
		Protocol:    "rdp",
		LocalPort:   3389,
		RemoteHost:  "internal-server",
		RemotePort:  3389,
	}

	tunnel, err := NewTunnel(config)
	assert.NoError(t, err)

	// Start tunnel
	err = tunnel.Start()
	assert.NoError(t, err)

	// Simulate connection loss
	tunnel.simulateConnectionLoss()

	// Wait for reconnection
	time.Sleep(100 * time.Millisecond)

	// Check if reconnected
	assert.True(t, tunnel.IsRunning())
} 