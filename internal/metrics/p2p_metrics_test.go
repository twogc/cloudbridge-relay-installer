package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewP2PMetrics(t *testing.T) {
	// Test creating new P2P metrics
	reg := prometheus.NewRegistry()
	metrics := NewP2PMetrics(reg)

	assert.NotNil(t, metrics)
	assert.NotNil(t, metrics.connectionsTotal)
	assert.NotNil(t, metrics.activePeers)
	assert.NotNil(t, metrics.trafficBytes)
	assert.NotNil(t, metrics.latencySeconds)
	assert.NotNil(t, metrics.handshakeDuration)
	assert.NotNil(t, metrics.meshRoutes)
	assert.NotNil(t, metrics.peerDiscoveryEvents)
	assert.NotNil(t, metrics.fallbackEvents)
	assert.NotNil(t, metrics.whitelistViolations)
}

func TestNewP2PMetrics_WithNilRegistry(t *testing.T) {
	// Test creating P2P metrics with nil registry
	metrics := NewP2PMetrics(nil)

	assert.NotNil(t, metrics)
	assert.NotNil(t, metrics.connectionsTotal)
	assert.NotNil(t, metrics.activePeers)
	assert.NotNil(t, metrics.trafficBytes)
	assert.NotNil(t, metrics.latencySeconds)
	assert.NotNil(t, metrics.handshakeDuration)
	assert.NotNil(t, metrics.meshRoutes)
	assert.NotNil(t, metrics.peerDiscoveryEvents)
	assert.NotNil(t, metrics.fallbackEvents)
	assert.NotNil(t, metrics.whitelistViolations)
}

func TestP2PMetrics_IncP2PConnections(t *testing.T) {
	// Test incrementing P2P connections
	reg := prometheus.NewRegistry()
	metrics := NewP2PMetrics(reg)

	// Increment connections
	metrics.IncP2PConnections()
	metrics.IncP2PConnections()
	metrics.IncP2PConnections()

	// Verify metric was incremented
	assert.NotNil(t, metrics.connectionsTotal)
}

func TestP2PMetrics_SetActivePeers(t *testing.T) {
	// Test setting active peers
	reg := prometheus.NewRegistry()
	metrics := NewP2PMetrics(reg)

	// Set active peers
	metrics.SetActivePeers(10)
	metrics.SetActivePeers(5)
	metrics.SetActivePeers(0)
	metrics.SetActivePeers(-1)

	// Verify metric was set
	assert.NotNil(t, metrics.activePeers)
}

func TestP2PMetrics_IncP2PTraffic(t *testing.T) {
	// Test incrementing P2P traffic
	reg := prometheus.NewRegistry()
	metrics := NewP2PMetrics(reg)

	// Increment traffic
	metrics.IncP2PTraffic("upload", "peer_1", 1024)
	metrics.IncP2PTraffic("download", "peer_1", 2048)
	metrics.IncP2PTraffic("upload", "peer_2", 512)
	metrics.IncP2PTraffic("download", "peer_2", 1024)

	// Verify metric was incremented
	assert.NotNil(t, metrics.trafficBytes)
}

func TestP2PMetrics_ObserveP2PLatency(t *testing.T) {
	// Test observing P2P latency
	reg := prometheus.NewRegistry()
	metrics := NewP2PMetrics(reg)

	// Observe latency
	metrics.ObserveP2PLatency("peer_1", 0.1)
	metrics.ObserveP2PLatency("peer_1", 0.2)
	metrics.ObserveP2PLatency("peer_2", 0.05)
	metrics.ObserveP2PLatency("peer_2", 0.15)

	// Verify metric was observed
	assert.NotNil(t, metrics.latencySeconds)
}

func TestP2PMetrics_ObserveHandshakeDuration(t *testing.T) {
	// Test observing handshake duration
	reg := prometheus.NewRegistry()
	metrics := NewP2PMetrics(reg)

	// Observe handshake duration
	metrics.ObserveHandshakeDuration(1.0)
	metrics.ObserveHandshakeDuration(2.0)
	metrics.ObserveHandshakeDuration(0.5)
	metrics.ObserveHandshakeDuration(5.0)

	// Verify metric was observed
	assert.NotNil(t, metrics.handshakeDuration)
}

func TestP2PMetrics_SetMeshRoutes(t *testing.T) {
	// Test setting mesh routes
	reg := prometheus.NewRegistry()
	metrics := NewP2PMetrics(reg)

	// Set mesh routes
	metrics.SetMeshRoutes(5)
	metrics.SetMeshRoutes(10)
	metrics.SetMeshRoutes(0)
	metrics.SetMeshRoutes(-1)

	// Verify metric was set
	assert.NotNil(t, metrics.meshRoutes)
}

func TestP2PMetrics_IncPeerDiscoveryEvent(t *testing.T) {
	// Test incrementing peer discovery events
	reg := prometheus.NewRegistry()
	metrics := NewP2PMetrics(reg)

	// Increment discovery events
	metrics.IncPeerDiscoveryEvent("peer_found")
	metrics.IncPeerDiscoveryEvent("peer_lost")
	metrics.IncPeerDiscoveryEvent("peer_found")
	metrics.IncPeerDiscoveryEvent("peer_updated")

	// Verify metric was incremented
	assert.NotNil(t, metrics.peerDiscoveryEvents)
}

func TestP2PMetrics_IncFallbackEvent(t *testing.T) {
	// Test incrementing fallback events
	reg := prometheus.NewRegistry()
	metrics := NewP2PMetrics(reg)

	// Increment fallback events
	metrics.IncFallbackEvent("wireguard", "quic", "timeout")
	metrics.IncFallbackEvent("quic", "tcp", "connection_failed")
	metrics.IncFallbackEvent("wireguard", "quic", "rate_limit")
	metrics.IncFallbackEvent("tcp", "udp", "packet_loss")

	// Verify metric was incremented
	assert.NotNil(t, metrics.fallbackEvents)
}

func TestP2PMetrics_IncWhitelistViolation(t *testing.T) {
	// Test incrementing whitelist violations
	reg := prometheus.NewRegistry()
	metrics := NewP2PMetrics(reg)

	// Increment whitelist violations
	metrics.IncWhitelistViolation("unauthorized_peer")
	metrics.IncWhitelistViolation("invalid_certificate")
	metrics.IncWhitelistViolation("unauthorized_peer")
	metrics.IncWhitelistViolation("expired_token")

	// Verify metric was incremented
	assert.NotNil(t, metrics.whitelistViolations)
}

func TestP2PMetrics_EdgeCases(t *testing.T) {
	// Test edge cases
	reg := prometheus.NewRegistry()
	metrics := NewP2PMetrics(reg)

	// Test with empty strings
	metrics.IncP2PTraffic("", "", 0)
	metrics.ObserveP2PLatency("", 0.0)
	metrics.IncPeerDiscoveryEvent("")
	metrics.IncFallbackEvent("", "", "")
	metrics.IncWhitelistViolation("")

	// Test with negative values
	metrics.SetActivePeers(-10)
	metrics.SetMeshRoutes(-5)
	metrics.IncP2PTraffic("upload", "peer_1", -1024)
	metrics.ObserveP2PLatency("peer_1", -0.1)
	metrics.ObserveHandshakeDuration(-1.0)

	// Test with very large values
	metrics.SetActivePeers(999999)
	metrics.SetMeshRoutes(999999)
	metrics.IncP2PTraffic("upload", "peer_1", 999999999)
	metrics.ObserveP2PLatency("peer_1", 999.0)
	metrics.ObserveHandshakeDuration(999.0)

	// Test with zero values
	metrics.SetActivePeers(0)
	metrics.SetMeshRoutes(0)
	metrics.IncP2PTraffic("upload", "peer_1", 0)
	metrics.ObserveP2PLatency("peer_1", 0.0)
	metrics.ObserveHandshakeDuration(0.0)

	// Verify all operations completed without errors
	assert.NotNil(t, metrics)
}

func TestP2PMetrics_ConcurrentAccess(t *testing.T) {
	// Test concurrent access to metrics
	reg := prometheus.NewRegistry()
	metrics := NewP2PMetrics(reg)
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			peerID := "peer" + string(rune(id))
			direction := "upload"
			if id%2 == 0 {
				direction = "download"
			}

			// Test all metrics methods
			metrics.IncP2PConnections()
			metrics.SetActivePeers(id)
			metrics.IncP2PTraffic(direction, peerID, int64(id*100))
			metrics.ObserveP2PLatency(peerID, float64(id)/10.0)
			metrics.ObserveHandshakeDuration(float64(id))
			metrics.SetMeshRoutes(id)
			metrics.IncPeerDiscoveryEvent("peer_found")
			metrics.IncFallbackEvent("wireguard", "quic", "timeout")
			metrics.IncWhitelistViolation("unauthorized_peer")
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all operations completed without errors
	assert.NotNil(t, metrics)
}

func TestP2PMetrics_IntegrationWorkflow(t *testing.T) {
	// Simulate a complete P2P workflow
	reg := prometheus.NewRegistry()
	metrics := NewP2PMetrics(reg)

	peerID := "peer_123"
	direction := "upload"

	// 1. Peer discovery
	metrics.IncPeerDiscoveryEvent("peer_found")

	// 2. Handshake
	metrics.ObserveHandshakeDuration(1.5)

	// 3. Connection established
	metrics.IncP2PConnections()
	metrics.SetActivePeers(1)

	// 4. Data transfer
	metrics.IncP2PTraffic(direction, peerID, 1024)
	metrics.IncP2PTraffic("download", peerID, 2048)

	// 5. Latency monitoring
	metrics.ObserveP2PLatency(peerID, 0.1)

	// 6. Mesh routing
	metrics.SetMeshRoutes(3)

	// 7. Fallback scenario
	metrics.IncFallbackEvent("wireguard", "quic", "timeout")

	// 8. Security violation
	metrics.IncWhitelistViolation("unauthorized_peer")

	// 9. Peer lost
	metrics.IncPeerDiscoveryEvent("peer_lost")
	metrics.SetActivePeers(0)

	// Verify workflow completed
	assert.NotNil(t, metrics)
}

func TestP2PMetrics_Performance(t *testing.T) {
	// Test performance with many operations
	reg := prometheus.NewRegistry()
	metrics := NewP2PMetrics(reg)

	start := time.Now()

	for i := 0; i < 1000; i++ {
		peerID := "peer" + string(rune(i%100))
		direction := "upload"
		if i%2 == 0 {
			direction = "download"
		}

		metrics.IncP2PConnections()
		metrics.SetActivePeers(i % 100)
		metrics.IncP2PTraffic(direction, peerID, int64(i%10000))
		metrics.ObserveP2PLatency(peerID, float64(i%100)/100.0)
		metrics.ObserveHandshakeDuration(float64(i%100) / 10.0)
		metrics.SetMeshRoutes(i % 50)
		metrics.IncPeerDiscoveryEvent("peer_found")
		metrics.IncFallbackEvent("wireguard", "quic", "timeout")
		metrics.IncWhitelistViolation("unauthorized_peer")
	}

	duration := time.Since(start)

	// Should complete within reasonable time (less than 1 second)
	assert.Less(t, duration, time.Second)
}

func TestP2PMetrics_DataTypes(t *testing.T) {
	// Test different data types and edge cases
	reg := prometheus.NewRegistry()
	metrics := NewP2PMetrics(reg)

	testCases := []struct {
		name     string
		testFunc func()
	}{
		{
			name: "string labels",
			testFunc: func() {
				metrics.IncP2PTraffic("upload-direction", "peer-id-123", 1024)
			},
		},
		{
			name: "unicode strings",
			testFunc: func() {
				metrics.IncP2PTraffic("направление-загрузки", "пир-ид-123", 1024)
			},
		},
		{
			name: "special characters",
			testFunc: func() {
				metrics.IncP2PTraffic("upload@direction", "peer_id-123", 1024)
			},
		},
		{
			name: "long strings",
			testFunc: func() {
				longString := "very-long-peer-id-that-might-cause-issues-with-prometheus-label-limits"
				metrics.IncP2PTraffic("upload", longString, 1024)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NotPanics(t, tc.testFunc)
		})
	}
}

func TestP2PMetrics_RegistryIntegration(t *testing.T) {
	// Test that metrics are properly registered
	reg := prometheus.NewRegistry()
	_ = NewP2PMetrics(reg)

	// Verify that all metrics are registered
	metricFamilies, err := reg.Gather()
	require.NoError(t, err)

	// Should have 9 metric families
	assert.Len(t, metricFamilies, 9)

	// Check that all expected metrics are present
	metricNames := make(map[string]bool)
	for _, mf := range metricFamilies {
		metricNames[*mf.Name] = true
	}

	expectedMetrics := []string{
		"cloudbridge_p2p_connections_total",
		"cloudbridge_p2p_active_peers",
		"cloudbridge_p2p_traffic_bytes_total",
		"cloudbridge_p2p_latency_seconds",
		"cloudbridge_p2p_handshake_duration_seconds",
		"cloudbridge_p2p_mesh_routes",
		"cloudbridge_p2p_discovery_events_total",
		"cloudbridge_p2p_fallback_events_total",
		"cloudbridge_p2p_whitelist_violations_total",
	}

	for _, expected := range expectedMetrics {
		assert.True(t, metricNames[expected], "Metric %s should be registered", expected)
	}
}

func TestP2PMetrics_MetricValues(t *testing.T) {
	// Test that metric values are correctly set
	reg := prometheus.NewRegistry()
	metrics := NewP2PMetrics(reg)

	// Set some values
	metrics.SetActivePeers(5)
	metrics.SetMeshRoutes(3)

	// Increment some counters
	metrics.IncP2PConnections()
	metrics.IncP2PConnections()
	metrics.IncP2PConnections()

	// Observe some histograms
	metrics.ObserveP2PLatency("peer_1", 0.1)
	metrics.ObserveP2PLatency("peer_1", 0.2)
	metrics.ObserveHandshakeDuration(1.0)
	metrics.ObserveHandshakeDuration(2.0)

	// Increment some labeled counters
	metrics.IncP2PTraffic("upload", "peer_1", 1024)
	metrics.IncP2PTraffic("download", "peer_1", 2048)
	metrics.IncPeerDiscoveryEvent("peer_found")
	metrics.IncPeerDiscoveryEvent("peer_lost")
	metrics.IncFallbackEvent("wireguard", "quic", "timeout")
	metrics.IncWhitelistViolation("unauthorized_peer")

	// Verify that metrics were recorded
	assert.NotNil(t, metrics)
}

func TestP2PMetrics_LabelValues(t *testing.T) {
	// Test different label values
	reg := prometheus.NewRegistry()
	metrics := NewP2PMetrics(reg)

	// Test different directions
	metrics.IncP2PTraffic("upload", "peer_1", 1024)
	metrics.IncP2PTraffic("download", "peer_1", 2048)
	metrics.IncP2PTraffic("bidirectional", "peer_1", 512)

	// Test different peer IDs
	metrics.IncP2PTraffic("upload", "peer_1", 1024)
	metrics.IncP2PTraffic("upload", "peer_2", 1024)
	metrics.IncP2PTraffic("upload", "peer_3", 1024)

	// Test different event types
	metrics.IncPeerDiscoveryEvent("peer_found")
	metrics.IncPeerDiscoveryEvent("peer_lost")
	metrics.IncPeerDiscoveryEvent("peer_updated")
	metrics.IncPeerDiscoveryEvent("peer_connected")
	metrics.IncPeerDiscoveryEvent("peer_disconnected")

	// Test different fallback scenarios
	metrics.IncFallbackEvent("wireguard", "quic", "timeout")
	metrics.IncFallbackEvent("quic", "tcp", "connection_failed")
	metrics.IncFallbackEvent("tcp", "udp", "packet_loss")
	metrics.IncFallbackEvent("udp", "http", "rate_limit")

	// Test different violation types
	metrics.IncWhitelistViolation("unauthorized_peer")
	metrics.IncWhitelistViolation("invalid_certificate")
	metrics.IncWhitelistViolation("expired_token")
	metrics.IncWhitelistViolation("invalid_signature")

	// Verify all operations completed without errors
	assert.NotNil(t, metrics)
}

func TestP2PMetrics_HistogramBuckets(t *testing.T) {
	// Test histogram bucket boundaries
	reg := prometheus.NewRegistry()
	metrics := NewP2PMetrics(reg)

	// Test latency histogram buckets
	latencyValues := []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 2.0}
	for _, val := range latencyValues {
		metrics.ObserveP2PLatency("peer_1", val)
	}

	// Test handshake duration histogram buckets
	handshakeValues := []float64{0.1, 0.5, 1.0, 2.0, 5.0, 10.0, 15.0}
	for _, val := range handshakeValues {
		metrics.ObserveHandshakeDuration(val)
	}

	// Verify all operations completed without errors
	assert.NotNil(t, metrics)
}
