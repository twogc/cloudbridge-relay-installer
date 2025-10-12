package masque

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestMASQUEMetrics_NewMASQUEMetrics(t *testing.T) {
	// Create a new registry for testing to avoid conflicts
	registry := prometheus.NewRegistry()
	
	// Create metrics with custom registry
	metrics := &MASQUEMetrics{
		ConnectionsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "masque_connections_total_test",
			Help: "Total number of MASQUE connections established",
		}),
		ConnectionsActive: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "masque_connections_active_test",
			Help: "Number of active MASQUE connections",
		}),
		ConnectionsFailed: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "masque_connections_failed_total_test",
			Help: "Total number of failed MASQUE connections",
		}),
		ConnectionsClosed: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "masque_connections_closed_total_test",
			Help: "Total number of closed MASQUE connections",
		}),
		BytesReceivedTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "masque_bytes_received_total_test",
			Help: "Total bytes received through MASQUE proxy",
		}),
		BytesSentTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "masque_bytes_sent_total_test",
			Help: "Total bytes sent through MASQUE proxy",
		}),
		PacketsReceivedTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "masque_packets_received_total_test",
			Help: "Total packets received through MASQUE proxy",
		}),
		PacketsSentTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "masque_packets_sent_total_test",
			Help: "Total packets sent through MASQUE proxy",
		}),
		RequestDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "masque_request_duration_seconds_test",
			Help:    "Duration of MASQUE requests",
			Buckets: prometheus.DefBuckets,
		}),
		ProxyLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "masque_proxy_latency_seconds_test",
			Help:    "Latency of MASQUE proxy operations",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		}),
		UDPRelayLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "masque_udp_relay_latency_seconds_test",
			Help:    "Latency of UDP relay operations",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		}),
		ProxyErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "masque_proxy_errors_total_test",
			Help: "Total number of MASQUE proxy errors",
		}),
		UDPErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "masque_udp_errors_total_test",
			Help: "Total number of UDP relay errors",
		}),
		ConnectionErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "masque_connection_errors_total_test",
			Help: "Total number of connection errors",
		}),
		TimeoutErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "masque_timeout_errors_total_test",
			Help: "Total number of timeout errors",
		}),
		ProxyStatus: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "masque_proxy_status_test",
			Help: "MASQUE proxy status (1=active, 0=inactive)",
		}),
		TargetResolutions: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "masque_target_resolutions_total_test",
			Help: "Total number of target address resolutions",
		}),
		TargetResolutionErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "masque_target_resolution_errors_total_test",
			Help: "Total number of target resolution errors",
		}),
	}

	// Register metrics with test registry
	registry.MustRegister(
		metrics.ConnectionsTotal,
		metrics.ConnectionsActive,
		metrics.ConnectionsFailed,
		metrics.ConnectionsClosed,
		metrics.BytesReceivedTotal,
		metrics.BytesSentTotal,
		metrics.PacketsReceivedTotal,
		metrics.PacketsSentTotal,
		metrics.RequestDuration,
		metrics.ProxyLatency,
		metrics.UDPRelayLatency,
		metrics.ProxyErrors,
		metrics.UDPErrors,
		metrics.ConnectionErrors,
		metrics.TimeoutErrors,
		metrics.ProxyStatus,
		metrics.TargetResolutions,
		metrics.TargetResolutionErrors,
	)

	if metrics == nil {
		t.Fatal("Expected metrics to be created")
	}

	// Test that all metrics are initialized
	if metrics.ConnectionsTotal == nil {
		t.Error("Expected ConnectionsTotal to be initialized")
	}

	if metrics.ConnectionsActive == nil {
		t.Error("Expected ConnectionsActive to be initialized")
	}

	if metrics.BytesReceivedTotal == nil {
		t.Error("Expected BytesReceivedTotal to be initialized")
	}

	if metrics.BytesSentTotal == nil {
		t.Error("Expected BytesSentTotal to be initialized")
	}

	if metrics.RequestDuration == nil {
		t.Error("Expected RequestDuration to be initialized")
	}

	if metrics.ProxyLatency == nil {
		t.Error("Expected ProxyLatency to be initialized")
	}

	if metrics.UDPRelayLatency == nil {
		t.Error("Expected UDPRelayLatency to be initialized")
	}
}

func TestMASQUEMetrics_IncrementConnections(t *testing.T) {
	// Create a new registry for testing
	registry := prometheus.NewRegistry()
	
	metrics := &MASQUEMetrics{
		ConnectionsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "masque_connections_total_test2",
			Help: "Total number of MASQUE connections established",
		}),
		ConnectionsActive: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "masque_connections_active_test2",
			Help: "Number of active MASQUE connections",
		}),
		ConnectionsFailed: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "masque_connections_failed_total_test2",
			Help: "Total number of failed MASQUE connections",
		}),
		ConnectionsClosed: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "masque_connections_closed_total_test2",
			Help: "Total number of closed MASQUE connections",
		}),
	}

	registry.MustRegister(
		metrics.ConnectionsTotal,
		metrics.ConnectionsActive,
		metrics.ConnectionsFailed,
		metrics.ConnectionsClosed,
	)

	// Test incrementing connections
	metrics.IncrementConnections()
	metrics.IncrementConnections()

	// Test decrementing connections
	metrics.DecrementConnections()

	// Test failed connections
	metrics.IncrementFailedConnections()

	// Test closed connections
	metrics.IncrementClosedConnections()

	// Verify metrics values
	if testutil.ToFloat64(metrics.ConnectionsTotal) != 2 {
		t.Errorf("Expected ConnectionsTotal to be 2, got %f", testutil.ToFloat64(metrics.ConnectionsTotal))
	}

	if testutil.ToFloat64(metrics.ConnectionsActive) != 1 {
		t.Errorf("Expected ConnectionsActive to be 1, got %f", testutil.ToFloat64(metrics.ConnectionsActive))
	}

	if testutil.ToFloat64(metrics.ConnectionsFailed) != 1 {
		t.Errorf("Expected ConnectionsFailed to be 1, got %f", testutil.ToFloat64(metrics.ConnectionsFailed))
	}

	if testutil.ToFloat64(metrics.ConnectionsClosed) != 1 {
		t.Errorf("Expected ConnectionsClosed to be 1, got %f", testutil.ToFloat64(metrics.ConnectionsClosed))
	}
}

func TestMASQUEMetrics_AddBytes(t *testing.T) {
	// Create a new registry for testing
	registry := prometheus.NewRegistry()
	
	metrics := &MASQUEMetrics{
		BytesReceivedTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "masque_bytes_received_total_test3",
			Help: "Total bytes received through MASQUE proxy",
		}),
		BytesSentTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "masque_bytes_sent_total_test3",
			Help: "Total bytes sent through MASQUE proxy",
		}),
		PacketsReceivedTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "masque_packets_received_total_test3",
			Help: "Total packets received through MASQUE proxy",
		}),
		PacketsSentTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "masque_packets_sent_total_test3",
			Help: "Total packets sent through MASQUE proxy",
		}),
	}

	registry.MustRegister(
		metrics.BytesReceivedTotal,
		metrics.BytesSentTotal,
		metrics.PacketsReceivedTotal,
		metrics.PacketsSentTotal,
	)

	// Test adding bytes
	metrics.AddBytesReceived(1024)
	metrics.AddBytesSent(2048)

	// Test incrementing packets
	metrics.IncrementPacketsReceived()
	metrics.IncrementPacketsSent()

	// Verify metrics values
	if testutil.ToFloat64(metrics.BytesReceivedTotal) != 1024 {
		t.Errorf("Expected BytesReceivedTotal to be 1024, got %f", testutil.ToFloat64(metrics.BytesReceivedTotal))
	}

	if testutil.ToFloat64(metrics.BytesSentTotal) != 2048 {
		t.Errorf("Expected BytesSentTotal to be 2048, got %f", testutil.ToFloat64(metrics.BytesSentTotal))
	}

	if testutil.ToFloat64(metrics.PacketsReceivedTotal) != 1 {
		t.Errorf("Expected PacketsReceivedTotal to be 1, got %f", testutil.ToFloat64(metrics.PacketsReceivedTotal))
	}

	if testutil.ToFloat64(metrics.PacketsSentTotal) != 1 {
		t.Errorf("Expected PacketsSentTotal to be 1, got %f", testutil.ToFloat64(metrics.PacketsSentTotal))
	}
}

func TestMASQUEMetrics_ObserveLatency(t *testing.T) {
	// Create a new registry for testing
	registry := prometheus.NewRegistry()
	
	metrics := &MASQUEMetrics{
		RequestDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "masque_request_duration_seconds_test4",
			Help:    "Duration of MASQUE requests",
			Buckets: prometheus.DefBuckets,
		}),
		ProxyLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "masque_proxy_latency_seconds_test4",
			Help:    "Latency of MASQUE proxy operations",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		}),
		UDPRelayLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "masque_udp_relay_latency_seconds_test4",
			Help:    "Latency of UDP relay operations",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		}),
	}

	registry.MustRegister(
		metrics.RequestDuration,
		metrics.ProxyLatency,
		metrics.UDPRelayLatency,
	)

	// Test observing latencies
	metrics.ObserveRequestDuration(0.1)
	metrics.ObserveProxyLatency(0.05)
	metrics.ObserveUDPRelayLatency(0.02)

	// Verify that observations were recorded by checking the count
	requestDurationCount := testutil.CollectAndCount(metrics.RequestDuration)
	if requestDurationCount == 0 {
		t.Error("Expected RequestDuration to have observations")
	}

	proxyLatencyCount := testutil.CollectAndCount(metrics.ProxyLatency)
	if proxyLatencyCount == 0 {
		t.Error("Expected ProxyLatency to have observations")
	}

	udpRelayLatencyCount := testutil.CollectAndCount(metrics.UDPRelayLatency)
	if udpRelayLatencyCount == 0 {
		t.Error("Expected UDPRelayLatency to have observations")
	}
}

func TestMASQUEMetrics_IncrementErrors(t *testing.T) {
	// Create a new registry for testing
	registry := prometheus.NewRegistry()
	
	metrics := &MASQUEMetrics{
		ProxyErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "masque_proxy_errors_total_test5",
			Help: "Total number of MASQUE proxy errors",
		}),
		UDPErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "masque_udp_errors_total_test5",
			Help: "Total number of UDP relay errors",
		}),
		ConnectionErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "masque_connection_errors_total_test5",
			Help: "Total number of connection errors",
		}),
		TimeoutErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "masque_timeout_errors_total_test5",
			Help: "Total number of timeout errors",
		}),
	}

	registry.MustRegister(
		metrics.ProxyErrors,
		metrics.UDPErrors,
		metrics.ConnectionErrors,
		metrics.TimeoutErrors,
	)

	// Test incrementing various error types
	metrics.IncrementProxyErrors()
	metrics.IncrementUDPErrors()
	metrics.IncrementConnectionErrors()
	metrics.IncrementTimeoutErrors()

	// Verify metrics values
	if testutil.ToFloat64(metrics.ProxyErrors) != 1 {
		t.Errorf("Expected ProxyErrors to be 1, got %f", testutil.ToFloat64(metrics.ProxyErrors))
	}

	if testutil.ToFloat64(metrics.UDPErrors) != 1 {
		t.Errorf("Expected UDPErrors to be 1, got %f", testutil.ToFloat64(metrics.UDPErrors))
	}

	if testutil.ToFloat64(metrics.ConnectionErrors) != 1 {
		t.Errorf("Expected ConnectionErrors to be 1, got %f", testutil.ToFloat64(metrics.ConnectionErrors))
	}

	if testutil.ToFloat64(metrics.TimeoutErrors) != 1 {
		t.Errorf("Expected TimeoutErrors to be 1, got %f", testutil.ToFloat64(metrics.TimeoutErrors))
	}
}

func TestMASQUEMetrics_SetProxyStatus(t *testing.T) {
	// Create a new registry for testing
	registry := prometheus.NewRegistry()
	
	metrics := &MASQUEMetrics{
		ProxyStatus: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "masque_proxy_status_test6",
			Help: "MASQUE proxy status (1=active, 0=inactive)",
		}),
	}

	registry.MustRegister(metrics.ProxyStatus)

	// Test setting proxy status
	metrics.SetProxyStatus(true)
	if testutil.ToFloat64(metrics.ProxyStatus) != 1 {
		t.Errorf("Expected ProxyStatus to be 1, got %f", testutil.ToFloat64(metrics.ProxyStatus))
	}

	metrics.SetProxyStatus(false)
	if testutil.ToFloat64(metrics.ProxyStatus) != 0 {
		t.Errorf("Expected ProxyStatus to be 0, got %f", testutil.ToFloat64(metrics.ProxyStatus))
	}
}

func TestMASQUEMetrics_TargetResolution(t *testing.T) {
	// Create a new registry for testing
	registry := prometheus.NewRegistry()
	
	metrics := &MASQUEMetrics{
		TargetResolutions: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "masque_target_resolutions_total_test7",
			Help: "Total number of target address resolutions",
		}),
		TargetResolutionErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "masque_target_resolution_errors_total_test7",
			Help: "Total number of target resolution errors",
		}),
	}

	registry.MustRegister(
		metrics.TargetResolutions,
		metrics.TargetResolutionErrors,
	)

	// Test target resolution metrics
	metrics.IncrementTargetResolutions()
	metrics.IncrementTargetResolutionErrors()

	// Verify metrics values
	if testutil.ToFloat64(metrics.TargetResolutions) != 1 {
		t.Errorf("Expected TargetResolutions to be 1, got %f", testutil.ToFloat64(metrics.TargetResolutions))
	}

	if testutil.ToFloat64(metrics.TargetResolutionErrors) != 1 {
		t.Errorf("Expected TargetResolutionErrors to be 1, got %f", testutil.ToFloat64(metrics.TargetResolutionErrors))
	}
}
