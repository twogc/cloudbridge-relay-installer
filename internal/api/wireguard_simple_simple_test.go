package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateWGConfig(t *testing.T) {
	t.Run("generate config", func(t *testing.T) {
		config := generateWGConfig("10.100.77.10", "10.100.77.0/24", "RELAY_PUBLIC_KEY", "relay.example.com:51820")

		assert.Contains(t, config, "10.100.77.10/24")
		assert.Contains(t, config, "10.100.77.0/24")
		assert.Contains(t, config, "RELAY_PUBLIC_KEY")
		assert.Contains(t, config, "relay.example.com:51820")
		assert.Contains(t, config, "PersistentKeepalive = 25")
		assert.Contains(t, config, "PrivateKey = <PUT_YOUR_PRIVATE_KEY_HERE>")
		assert.Contains(t, config, "DNS = 1.1.1.1")
	})

	t.Run("generate config with different parameters", func(t *testing.T) {
		config := generateWGConfig("192.168.1.100", "192.168.1.0/24", "DIFFERENT_KEY", "different.example.com:51821")

		assert.Contains(t, config, "192.168.1.100/24")
		assert.Contains(t, config, "192.168.1.0/24")
		assert.Contains(t, config, "DIFFERENT_KEY")
		assert.Contains(t, config, "different.example.com:51821")
		assert.Contains(t, config, "PersistentKeepalive = 25")
	})
}

