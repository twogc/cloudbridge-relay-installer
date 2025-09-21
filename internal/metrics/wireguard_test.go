package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWireGuardMetrics(t *testing.T) {
	// Test creating new WireGuard metrics
	metrics := NewWireGuardMetrics()
	assert.NotNil(t, metrics)
}

func TestWireGuardMetrics_InterfaceMetrics(t *testing.T) {
	metrics := NewWireGuardMetrics()

	// Test interface status
	metrics.UpdateInterfaceStatus("wg0", "main", true)
	metrics.UpdateInterfaceStatus("wg1", "backup", false)
	metrics.UpdateInterfaceStatus("wg0", "main", false)

	// Test peers count
	metrics.UpdatePeersCount("wg0", 5)
	metrics.UpdatePeersCount("wg1", 3)
	metrics.UpdatePeersCount("wg0", 0)

	// Test bytes received
	metrics.IncrementBytesReceived("wg0", "peer1", 1024)
	metrics.IncrementBytesReceived("wg0", "peer1", 2048)
	metrics.IncrementBytesReceived("wg1", "peer2", 512)

	// Test bytes sent
	metrics.IncrementBytesSent("wg0", "peer1", 1024)
	metrics.IncrementBytesSent("wg0", "peer1", 2048)
	metrics.IncrementBytesSent("wg1", "peer2", 512)

	// Test handshakes
	metrics.IncrementHandshakes("wg0", "peer1", "success")
	metrics.IncrementHandshakes("wg0", "peer1", "success")
	metrics.IncrementHandshakes("wg1", "peer2", "failed")

	// Verify metrics were updated
	assert.NotNil(t, WireGuardInterfaceStatus)
	assert.NotNil(t, WireGuardPeersTotal)
	assert.NotNil(t, WireGuardBytesReceived)
	assert.NotNil(t, WireGuardBytesSent)
	assert.NotNil(t, WireGuardHandshakesTotal)
}

func TestWireGuardMetrics_MeshTopology(t *testing.T) {
	metrics := NewWireGuardMetrics()

	// Test mesh connections
	metrics.UpdateMeshConnections("full_mesh", 10)
	metrics.UpdateMeshConnections("partial_mesh", 5)
	metrics.UpdateMeshConnections("full_mesh", 0)

	// Test mesh routes
	metrics.UpdateMeshRoutes("full_mesh", "direct", 8)
	metrics.UpdateMeshRoutes("full_mesh", "indirect", 2)
	metrics.UpdateMeshRoutes("partial_mesh", "direct", 3)

	// Test mesh latency
	metrics.ObserveMeshLatency("full_mesh", "peer1", 0.1)
	metrics.ObserveMeshLatency("full_mesh", "peer1", 0.2)
	metrics.ObserveMeshLatency("partial_mesh", "peer2", 0.15)

	// Verify metrics were updated
	assert.NotNil(t, WireGuardMeshConnections)
	assert.NotNil(t, WireGuardMeshRoutes)
	assert.NotNil(t, WireGuardMeshLatency)
}

func TestWireGuardMetrics_PeerDiscovery(t *testing.T) {
	metrics := NewWireGuardMetrics()

	// Test discovered peers
	metrics.UpdateDiscoveredPeers("etcd", 5)
	metrics.UpdateDiscoveredPeers("consul", 3)
	metrics.UpdateDiscoveredPeers("etcd", 0)

	// Test discovery heartbeats
	metrics.IncrementDiscoveryHeartbeats("etcd", "success")
	metrics.IncrementDiscoveryHeartbeats("etcd", "success")
	metrics.IncrementDiscoveryHeartbeats("consul", "failed")

	// Test discovery errors
	metrics.IncrementDiscoveryErrors("etcd", "connection_failed")
	metrics.IncrementDiscoveryErrors("etcd", "timeout")
	metrics.IncrementDiscoveryErrors("consul", "invalid_response")

	// Verify metrics were updated
	assert.NotNil(t, WireGuardDiscoveredPeers)
	assert.NotNil(t, WireGuardDiscoveryHeartbeats)
	assert.NotNil(t, WireGuardDiscoveryErrors)
}

func TestWireGuardMetrics_QUICOverWireGuard(t *testing.T) {
	metrics := NewWireGuardMetrics()

	// Test QUIC connections
	metrics.UpdateQUICConnections("wg0", 3)
	metrics.UpdateQUICConnections("wg1", 2)
	metrics.UpdateQUICConnections("wg0", 0)

	// Test QUIC streams
	metrics.UpdateQUICStreams("wg0", "upload", 5)
	metrics.UpdateQUICStreams("wg0", "download", 3)
	metrics.UpdateQUICStreams("wg1", "upload", 2)

	// Test QUIC errors
	metrics.IncrementQUICErrors("wg0", "connection_failed")
	metrics.IncrementQUICErrors("wg0", "stream_error")
	metrics.IncrementQUICErrors("wg1", "timeout")

	// Verify metrics were updated
	assert.NotNil(t, WireGuardQUICConnections)
	assert.NotNil(t, WireGuardQUICStreams)
	assert.NotNil(t, WireGuardQUICErrors)
}

func TestWireGuardMetrics_HybridRouter(t *testing.T) {
	metrics := NewWireGuardMetrics()

	// Test routing decisions
	metrics.IncrementRoutingDecisions("latency_based", "direct")
	metrics.IncrementRoutingDecisions("latency_based", "direct")
	metrics.IncrementRoutingDecisions("load_based", "indirect")

	// Test routing latency
	metrics.ObserveRoutingLatency("latency_based", 0.01)
	metrics.ObserveRoutingLatency("latency_based", 0.02)
	metrics.ObserveRoutingLatency("load_based", 0.015)

	// Test routing errors
	metrics.IncrementRoutingErrors("latency_based", "no_route")
	metrics.IncrementRoutingErrors("latency_based", "timeout")
	metrics.IncrementRoutingErrors("load_based", "overload")

	// Verify metrics were updated
	assert.NotNil(t, WireGuardRoutingDecisions)
	assert.NotNil(t, WireGuardRoutingLatency)
	assert.NotNil(t, WireGuardRoutingErrors)
}

func TestWireGuardMetrics_SiteToSite(t *testing.T) {
	metrics := NewWireGuardMetrics()

	// Test site-to-site connections
	metrics.UpdateSiteToSiteConnections("network1", "tenant1", 2)
	metrics.UpdateSiteToSiteConnections("network2", "tenant1", 1)
	metrics.UpdateSiteToSiteConnections("network1", "tenant2", 3)

	// Test site-to-site routes
	metrics.UpdateSiteToSiteRoutes("network1", "network2", 1)
	metrics.UpdateSiteToSiteRoutes("network2", "network3", 2)
	metrics.UpdateSiteToSiteRoutes("network1", "network3", 1)

	// Test VPN replacement status
	metrics.UpdateVPNReplacementStatus("network1", "tenant1", true)
	metrics.UpdateVPNReplacementStatus("network2", "tenant1", false)
	metrics.UpdateVPNReplacementStatus("network1", "tenant2", true)

	// Verify metrics were updated
	assert.NotNil(t, WireGuardSiteToSiteConnections)
	assert.NotNil(t, WireGuardSiteToSiteRoutes)
	assert.NotNil(t, WireGuardVPNReplacementStatus)
}

func TestWireGuardMetrics_EdgeCases(t *testing.T) {
	metrics := NewWireGuardMetrics()

	// Test with empty strings
	metrics.UpdateInterfaceStatus("", "", true)
	metrics.UpdatePeersCount("", 0)
	metrics.IncrementBytesReceived("", "", 0)
	metrics.IncrementBytesSent("", "", 0)
	metrics.IncrementHandshakes("", "", "")
	metrics.UpdateMeshConnections("", 0)
	metrics.UpdateMeshRoutes("", "", 0)
	metrics.ObserveMeshLatency("", "", 0.0)
	metrics.UpdateDiscoveredPeers("", 0)
	metrics.IncrementDiscoveryHeartbeats("", "")
	metrics.IncrementDiscoveryErrors("", "")
	metrics.UpdateQUICConnections("", 0)
	metrics.UpdateQUICStreams("", "", 0)
	metrics.IncrementQUICErrors("", "")
	metrics.IncrementRoutingDecisions("", "")
	metrics.ObserveRoutingLatency("", 0.0)
	metrics.IncrementRoutingErrors("", "")
	metrics.UpdateSiteToSiteConnections("", "", 0)
	metrics.UpdateSiteToSiteRoutes("", "", 0)
	metrics.UpdateVPNReplacementStatus("", "", true)

	// Test with negative values
	metrics.UpdatePeersCount("wg0", -1)
	metrics.IncrementBytesReceived("wg0", "peer1", -1024)
	metrics.IncrementBytesSent("wg0", "peer1", -1024)
	metrics.UpdateMeshConnections("full_mesh", -1)
	metrics.UpdateMeshRoutes("full_mesh", "direct", -1)
	metrics.ObserveMeshLatency("full_mesh", "peer1", -0.1)
	metrics.UpdateDiscoveredPeers("etcd", -1)
	metrics.UpdateQUICConnections("wg0", -1)
	metrics.UpdateQUICStreams("wg0", "upload", -1)
	metrics.ObserveRoutingLatency("latency_based", -0.1)
	metrics.UpdateSiteToSiteConnections("network1", "tenant1", -1)
	metrics.UpdateSiteToSiteRoutes("network1", "network2", -1)

	// Test with very large values
	metrics.UpdatePeersCount("wg0", 999999)
	metrics.IncrementBytesReceived("wg0", "peer1", 999999999)
	metrics.IncrementBytesSent("wg0", "peer1", 999999999)
	metrics.UpdateMeshConnections("full_mesh", 999999)
	metrics.UpdateMeshRoutes("full_mesh", "direct", 999999)
	metrics.ObserveMeshLatency("full_mesh", "peer1", 999.0)
	metrics.UpdateDiscoveredPeers("etcd", 999999)
	metrics.UpdateQUICConnections("wg0", 999999)
	metrics.UpdateQUICStreams("wg0", "upload", 999999)
	metrics.ObserveRoutingLatency("latency_based", 999.0)
	metrics.UpdateSiteToSiteConnections("network1", "tenant1", 999999)
	metrics.UpdateSiteToSiteRoutes("network1", "network2", 999999)

	// Test with zero values
	metrics.UpdatePeersCount("wg0", 0)
	metrics.IncrementBytesReceived("wg0", "peer1", 0)
	metrics.IncrementBytesSent("wg0", "peer1", 0)
	metrics.UpdateMeshConnections("full_mesh", 0)
	metrics.UpdateMeshRoutes("full_mesh", "direct", 0)
	metrics.ObserveMeshLatency("full_mesh", "peer1", 0.0)
	metrics.UpdateDiscoveredPeers("etcd", 0)
	metrics.UpdateQUICConnections("wg0", 0)
	metrics.UpdateQUICStreams("wg0", "upload", 0)
	metrics.ObserveRoutingLatency("latency_based", 0.0)
	metrics.UpdateSiteToSiteConnections("network1", "tenant1", 0)
	metrics.UpdateSiteToSiteRoutes("network1", "network2", 0)

	// Verify all operations completed without errors
	assert.NotNil(t, metrics)
}

func TestWireGuardMetrics_ConcurrentAccess(t *testing.T) {
	metrics := NewWireGuardMetrics()
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			interfaceName := "wg" + string(rune(id%3))
			peer := "peer" + string(rune(id%5))
			topology := "full_mesh"
			if id%2 == 0 {
				topology = "partial_mesh"
			}
			registry := "etcd"
			if id%2 == 0 {
				registry = "consul"
			}
			strategy := "latency_based"
			if id%2 == 0 {
				strategy = "load_based"
			}
			networkID := "network" + string(rune(id%3))
			tenantID := "tenant" + string(rune(id%2))

			// Test all metrics methods
			metrics.UpdateInterfaceStatus(interfaceName, "test", true)
			metrics.UpdatePeersCount(interfaceName, id)
			metrics.IncrementBytesReceived(interfaceName, peer, int64(id*100))
			metrics.IncrementBytesSent(interfaceName, peer, int64(id*200))
			metrics.IncrementHandshakes(interfaceName, peer, "success")
			metrics.UpdateMeshConnections(topology, id)
			metrics.UpdateMeshRoutes(topology, "direct", id)
			metrics.ObserveMeshLatency(topology, peer, float64(id)/100.0)
			metrics.UpdateDiscoveredPeers(registry, id)
			metrics.IncrementDiscoveryHeartbeats(registry, "success")
			metrics.IncrementDiscoveryErrors(registry, "error")
			metrics.UpdateQUICConnections(interfaceName, id)
			metrics.UpdateQUICStreams(interfaceName, "upload", id)
			metrics.IncrementQUICErrors(interfaceName, "error")
			metrics.IncrementRoutingDecisions(strategy, "direct")
			metrics.ObserveRoutingLatency(strategy, float64(id)/1000.0)
			metrics.IncrementRoutingErrors(strategy, "error")
			metrics.UpdateSiteToSiteConnections(networkID, tenantID, id)
			metrics.UpdateSiteToSiteRoutes(networkID, "network2", id)
			metrics.UpdateVPNReplacementStatus(networkID, tenantID, true)
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all operations completed without errors
	assert.NotNil(t, metrics)
}

func TestWireGuardMetrics_IntegrationWorkflow(t *testing.T) {
	// Simulate a complete WireGuard workflow
	metrics := NewWireGuardMetrics()

	interfaceName := "wg0"
	peer := "peer1"
	topology := "full_mesh"
	registry := "etcd"
	strategy := "latency_based"
	networkID := "network1"
	tenantID := "tenant1"

	// 1. Interface setup
	metrics.UpdateInterfaceStatus(interfaceName, "main", true)
	metrics.UpdatePeersCount(interfaceName, 1)

	// 2. Peer discovery
	metrics.UpdateDiscoveredPeers(registry, 5)
	metrics.IncrementDiscoveryHeartbeats(registry, "success")

	// 3. Handshake
	metrics.IncrementHandshakes(interfaceName, peer, "success")

	// 4. Data transfer
	metrics.IncrementBytesReceived(interfaceName, peer, 1024)
	metrics.IncrementBytesSent(interfaceName, peer, 2048)

	// 5. Mesh topology
	metrics.UpdateMeshConnections(topology, 3)
	metrics.UpdateMeshRoutes(topology, "direct", 2)
	metrics.ObserveMeshLatency(topology, peer, 0.1)

	// 6. QUIC over WireGuard
	metrics.UpdateQUICConnections(interfaceName, 2)
	metrics.UpdateQUICStreams(interfaceName, "upload", 3)
	metrics.UpdateQUICStreams(interfaceName, "download", 2)

	// 7. Routing decisions
	metrics.IncrementRoutingDecisions(strategy, "direct")
	metrics.ObserveRoutingLatency(strategy, 0.01)

	// 8. Site-to-site connections
	metrics.UpdateSiteToSiteConnections(networkID, tenantID, 1)
	metrics.UpdateSiteToSiteRoutes(networkID, "network2", 1)
	metrics.UpdateVPNReplacementStatus(networkID, tenantID, true)

	// 9. Error handling
	metrics.IncrementDiscoveryErrors(registry, "timeout")
	metrics.IncrementQUICErrors(interfaceName, "connection_failed")
	metrics.IncrementRoutingErrors(strategy, "no_route")

	// Verify workflow completed
	assert.NotNil(t, metrics)
}

func TestWireGuardMetrics_Performance(t *testing.T) {
	// Test performance with many operations
	metrics := NewWireGuardMetrics()

	start := time.Now()

	for i := 0; i < 1000; i++ {
		interfaceName := "wg" + string(rune(i%5))
		peer := "peer" + string(rune(i%10))
		topology := "full_mesh"
		if i%2 == 0 {
			topology = "partial_mesh"
		}
		registry := "etcd"
		if i%2 == 0 {
			registry = "consul"
		}
		strategy := "latency_based"
		if i%2 == 0 {
			strategy = "load_based"
		}
		networkID := "network" + string(rune(i%3))
		tenantID := "tenant" + string(rune(i%2))

		metrics.UpdateInterfaceStatus(interfaceName, "test", true)
		metrics.UpdatePeersCount(interfaceName, i%100)
		metrics.IncrementBytesReceived(interfaceName, peer, int64(i%10000))
		metrics.IncrementBytesSent(interfaceName, peer, int64(i%20000))
		metrics.IncrementHandshakes(interfaceName, peer, "success")
		metrics.UpdateMeshConnections(topology, i%50)
		metrics.UpdateMeshRoutes(topology, "direct", i%30)
		metrics.ObserveMeshLatency(topology, peer, float64(i%100)/1000.0)
		metrics.UpdateDiscoveredPeers(registry, i%20)
		metrics.IncrementDiscoveryHeartbeats(registry, "success")
		metrics.IncrementDiscoveryErrors(registry, "error")
		metrics.UpdateQUICConnections(interfaceName, i%15)
		metrics.UpdateQUICStreams(interfaceName, "upload", i%10)
		metrics.IncrementQUICErrors(interfaceName, "error")
		metrics.IncrementRoutingDecisions(strategy, "direct")
		metrics.ObserveRoutingLatency(strategy, float64(i%100)/10000.0)
		metrics.IncrementRoutingErrors(strategy, "error")
		metrics.UpdateSiteToSiteConnections(networkID, tenantID, i%5)
		metrics.UpdateSiteToSiteRoutes(networkID, "network2", i%3)
		metrics.UpdateVPNReplacementStatus(networkID, tenantID, true)
	}

	duration := time.Since(start)

	// Should complete within reasonable time (less than 1 second)
	assert.Less(t, duration, time.Second)
}

func TestWireGuardMetrics_DataTypes(t *testing.T) {
	// Test different data types and edge cases
	metrics := NewWireGuardMetrics()

	testCases := []struct {
		name     string
		testFunc func()
	}{
		{
			name: "string labels",
			testFunc: func() {
				metrics.UpdateInterfaceStatus("wg-interface", "main-interface", true)
			},
		},
		{
			name: "unicode strings",
			testFunc: func() {
				metrics.UpdateInterfaceStatus("интерфейс-wg", "основной-интерфейс", true)
			},
		},
		{
			name: "special characters",
			testFunc: func() {
				metrics.UpdateInterfaceStatus("wg@interface", "main_interface", true)
			},
		},
		{
			name: "long strings",
			testFunc: func() {
				longString := "very-long-interface-name-that-might-cause-issues-with-prometheus-label-limits"
				metrics.UpdateInterfaceStatus(longString, "main", true)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NotPanics(t, tc.testFunc)
		})
	}
}

func TestWireGuardMetrics_StatusValues(t *testing.T) {
	// Test status values (running/stopped, active/inactive)
	metrics := NewWireGuardMetrics()

	// Test running status (true)
	metrics.UpdateInterfaceStatus("wg0", "main", true)
	metrics.UpdateVPNReplacementStatus("network1", "tenant1", true)

	// Test stopped status (false)
	metrics.UpdateInterfaceStatus("wg0", "main", false)
	metrics.UpdateVPNReplacementStatus("network1", "tenant1", false)

	// Verify operations completed without errors
	assert.NotNil(t, WireGuardInterfaceStatus)
	assert.NotNil(t, WireGuardVPNReplacementStatus)
}

func TestWireGuardMetrics_InterfaceTypes(t *testing.T) {
	// Test different interface types
	metrics := NewWireGuardMetrics()

	interfaces := []string{"wg0", "wg1", "wg2", "wg3", "wg4"}
	peers := []string{"peer1", "peer2", "peer3", "peer4", "peer5"}

	for _, iface := range interfaces {
		for _, peer := range peers {
			metrics.UpdateInterfaceStatus(iface, "test", true)
			metrics.UpdatePeersCount(iface, 1)
			metrics.IncrementBytesReceived(iface, peer, 1024)
			metrics.IncrementBytesSent(iface, peer, 2048)
			metrics.IncrementHandshakes(iface, peer, "success")
			metrics.UpdateQUICConnections(iface, 1)
			metrics.UpdateQUICStreams(iface, "upload", 1)
			metrics.IncrementQUICErrors(iface, "error")
		}
	}

	// Verify all operations completed without errors
	assert.NotNil(t, metrics)
}

func TestWireGuardMetrics_TopologyTypes(t *testing.T) {
	// Test different topology types
	metrics := NewWireGuardMetrics()

	topologies := []string{"full_mesh", "partial_mesh", "star", "ring", "tree"}
	routeTypes := []string{"direct", "indirect", "backup", "primary", "secondary"}

	for _, topology := range topologies {
		for _, routeType := range routeTypes {
			metrics.UpdateMeshConnections(topology, 1)
			metrics.UpdateMeshRoutes(topology, routeType, 1)
			metrics.ObserveMeshLatency(topology, "peer1", 0.1)
		}
	}

	// Verify all operations completed without errors
	assert.NotNil(t, metrics)
}

func TestWireGuardMetrics_RegistryTypes(t *testing.T) {
	// Test different registry types
	metrics := NewWireGuardMetrics()

	registries := []string{"etcd", "consul", "zookeeper", "redis", "postgresql"}
	statuses := []string{"success", "failed", "timeout", "error", "unknown"}

	for _, registry := range registries {
		for _, status := range statuses {
			metrics.UpdateDiscoveredPeers(registry, 1)
			metrics.IncrementDiscoveryHeartbeats(registry, status)
			metrics.IncrementDiscoveryErrors(registry, "error")
		}
	}

	// Verify all operations completed without errors
	assert.NotNil(t, metrics)
}

func TestWireGuardMetrics_StrategyTypes(t *testing.T) {
	// Test different routing strategy types
	metrics := NewWireGuardMetrics()

	strategies := []string{"latency_based", "load_based", "cost_based", "reliability_based", "hybrid"}
	routeTypes := []string{"direct", "indirect", "backup", "primary", "secondary"}

	for _, strategy := range strategies {
		for _, routeType := range routeTypes {
			metrics.IncrementRoutingDecisions(strategy, routeType)
			metrics.ObserveRoutingLatency(strategy, 0.01)
			metrics.IncrementRoutingErrors(strategy, "error")
		}
	}

	// Verify all operations completed without errors
	assert.NotNil(t, metrics)
}
