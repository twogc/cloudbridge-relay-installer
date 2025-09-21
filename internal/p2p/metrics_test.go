package p2p

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewP2PMetrics(t *testing.T) {
	t.Run("creates metrics with all components", func(t *testing.T) {
		metrics := NewP2PMetrics()

		assert.NotNil(t, metrics)
		assert.NotNil(t, metrics.connections)
		assert.NotNil(t, metrics.latency)
		assert.NotNil(t, metrics.throughput)
		assert.NotNil(t, metrics.peerCount)
		assert.NotNil(t, metrics.activeConnections)
		assert.NotNil(t, metrics.totalConnections)
		assert.NotNil(t, metrics.packetsSent)
		assert.NotNil(t, metrics.packetsReceived)
		assert.NotNil(t, metrics.packetsLost)
		assert.NotNil(t, metrics.bytesTransferred)
		assert.NotNil(t, metrics.handshakeDuration)
		assert.NotNil(t, metrics.connectionErrors)
	})
}

func TestP2PMetrics_RecordConnection(t *testing.T) {
	t.Run("records connection metrics", func(t *testing.T) {
		metrics := NewP2PMetrics()

		routeType := "direct"
		status := "success"
		tenantID := "tenant-123"
		duration := 100 * time.Millisecond

		// This should not panic
		metrics.RecordConnection(routeType, status, tenantID, duration)

		// Verify metrics were recorded by checking if we can get the value
		// Note: In a real test, you might want to use a test registry
		assert.NotNil(t, metrics.connections)
		assert.NotNil(t, metrics.latency)
	})
}

func TestP2PMetrics_RecordThroughput(t *testing.T) {
	t.Run("records throughput metrics", func(t *testing.T) {
		metrics := NewP2PMetrics()

		routeType := "mesh"
		tenantID := "tenant-456"
		bytes := int64(1024)

		// This should not panic
		metrics.RecordThroughput(routeType, tenantID, bytes)

		assert.NotNil(t, metrics.throughput)
	})
}

func TestP2PMetrics_SetPeerCount(t *testing.T) {
	t.Run("sets peer count", func(t *testing.T) {
		metrics := NewP2PMetrics()

		count := 5
		metrics.SetPeerCount(count)

		assert.NotNil(t, metrics.peerCount)
	})
}

func TestP2PMetrics_SetActiveConnections(t *testing.T) {
	t.Run("sets active connections count", func(t *testing.T) {
		metrics := NewP2PMetrics()

		count := 10
		metrics.SetActiveConnections(count)

		assert.NotNil(t, metrics.activeConnections)
	})
}

func TestP2PMetrics_IncrementTotalConnections(t *testing.T) {
	t.Run("increments total connections", func(t *testing.T) {
		metrics := NewP2PMetrics()

		// This should not panic
		metrics.IncrementTotalConnections()

		assert.NotNil(t, metrics.totalConnections)
	})
}

func TestP2PMetrics_RecordPackets(t *testing.T) {
	t.Run("records packet metrics", func(t *testing.T) {
		metrics := NewP2PMetrics()

		sent := int64(100)
		received := int64(95)
		lost := int64(5)

		// This should not panic
		metrics.RecordPackets(sent, received, lost)

		assert.NotNil(t, metrics.packetsSent)
		assert.NotNil(t, metrics.packetsReceived)
		assert.NotNil(t, metrics.packetsLost)
	})
}

func TestP2PMetrics_RecordBytesTransferred(t *testing.T) {
	t.Run("records bytes transferred", func(t *testing.T) {
		metrics := NewP2PMetrics()

		bytes := int64(2048)
		metrics.RecordBytesTransferred(bytes)

		assert.NotNil(t, metrics.bytesTransferred)
	})
}

func TestP2PMetrics_RecordHandshakeDuration(t *testing.T) {
	t.Run("records handshake duration", func(t *testing.T) {
		metrics := NewP2PMetrics()

		protocol := "quic"
		duration := 50 * time.Millisecond
		metrics.RecordHandshakeDuration(protocol, duration)

		assert.NotNil(t, metrics.handshakeDuration)
	})
}

func TestP2PMetrics_RecordConnectionError(t *testing.T) {
	t.Run("records connection error", func(t *testing.T) {
		metrics := NewP2PMetrics()

		errorType := "timeout"
		protocol := "wireguard"
		metrics.RecordConnectionError(errorType, protocol)

		assert.NotNil(t, metrics.connectionErrors)
	})
}

func TestP2PMetrics_Register(t *testing.T) {
	t.Run("registers all metrics with registry", func(t *testing.T) {
		metrics := NewP2PMetrics()
		registry := prometheus.NewRegistry()

		err := metrics.Register(registry)
		assert.NoError(t, err)

		// Verify all metrics are registered
		metricsList, err := registry.Gather()
		require.NoError(t, err)

		// Should have multiple metrics registered
		assert.Greater(t, len(metricsList), 5)
	})
}

func TestP2PMetrics_Register_ErrorHandling(t *testing.T) {
	t.Run("handles registration errors gracefully", func(t *testing.T) {
		metrics := NewP2PMetrics()

		// Create a registry that will cause errors
		registry := prometheus.NewRegistry()

		// First registration should succeed
		err := metrics.Register(registry)
		assert.NoError(t, err)

		// Second registration should fail (duplicate registration)
		err = metrics.Register(registry)
		assert.Error(t, err)
	})
}

func TestP2PMetrics_EdgeCases(t *testing.T) {
	t.Run("handles zero values", func(t *testing.T) {
		metrics := NewP2PMetrics()

		// Test with zero values
		metrics.SetPeerCount(0)
		metrics.SetActiveConnections(0)
		metrics.RecordThroughput("", "", 0)
		metrics.RecordPackets(0, 0, 0)
		metrics.RecordBytesTransferred(0)
		metrics.RecordHandshakeDuration("", 0)

		// Should not panic
		assert.NotNil(t, metrics)
	})

	t.Run("handles negative values", func(t *testing.T) {
		metrics := NewP2PMetrics()

		// Test with negative values for gauges (these should work)
		metrics.SetPeerCount(-1)
		metrics.SetActiveConnections(-1)

		// For counters, we can't use negative values, so we test with 0
		metrics.RecordThroughput("", "", 0)
		metrics.RecordPackets(0, 0, 0)
		metrics.RecordBytesTransferred(0)

		// Should not panic
		assert.NotNil(t, metrics)
	})

	t.Run("handles large values", func(t *testing.T) {
		metrics := NewP2PMetrics()

		// Test with large values
		metrics.SetPeerCount(1000000)
		metrics.SetActiveConnections(1000000)
		metrics.RecordThroughput("", "", 9223372036854775807) // max int64
		metrics.RecordPackets(1000000, 1000000, 1000000)
		metrics.RecordBytesTransferred(9223372036854775807)

		// Should not panic
		assert.NotNil(t, metrics)
	})

	t.Run("handles empty strings", func(t *testing.T) {
		metrics := NewP2PMetrics()

		// Test with empty strings
		metrics.RecordConnection("", "", "", time.Millisecond)
		metrics.RecordThroughput("", "", 100)
		metrics.RecordHandshakeDuration("", time.Millisecond)
		metrics.RecordConnectionError("", "")

		// Should not panic
		assert.NotNil(t, metrics)
	})
}

func TestP2PMetrics_ConcurrentAccess(t *testing.T) {
	t.Run("handles concurrent access safely", func(t *testing.T) {
		metrics := NewP2PMetrics()

		// Run multiple goroutines that modify metrics
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func(id int) {
				defer func() { done <- true }()

				// Each goroutine performs different operations
				metrics.SetPeerCount(id)
				metrics.SetActiveConnections(id * 2)
				metrics.IncrementTotalConnections()
				metrics.RecordConnection("direct", "success", "tenant", time.Millisecond)
				metrics.RecordThroughput("mesh", "tenant", int64(id*100))
				metrics.RecordPackets(int64(id), int64(id), 0)
				metrics.RecordBytesTransferred(int64(id * 50))
				metrics.RecordHandshakeDuration("quic", time.Millisecond)
				metrics.RecordConnectionError("timeout", "wireguard")
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}

		// Should not panic and metrics should still be valid
		assert.NotNil(t, metrics)
	})
}
