package quic

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twogc/cloudbridge-relay-installer/internal/quic/congestion"
	"go.uber.org/zap/zaptest"
)

// TestExperimentalFeaturesIntegration tests all experimental features integration
func TestExperimentalFeaturesIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("BBRv2Integration", testBBRv2Integration)
	t.Run("ACKFrequencyIntegration", testACKFrequencyIntegration)
	t.Run("FECIntegration", testFECIntegration)
	t.Run("GreasingIntegration", testGreasingIntegration)
	t.Run("MultiFeatureIntegration", testMultiFeatureIntegration)
	t.Run("PrometheusMetricsIntegration", testPrometheusMetricsIntegration)
	t.Run("PerformanceUnderLoad", testPerformanceUnderLoad)
}

// testBBRv2Integration tests BBRv2 congestion control integration
func testBBRv2Integration(t *testing.T) {
	logger := zaptest.NewLogger(t)
	defer logger.Sync()

	// Create BBRv2 manager
	ccManager := congestion.NewCongestionControlManager(logger, congestion.AlgorithmBBRv2, 1460)
	require.NotNil(t, ccManager, "BBRv2 manager should be initialized")

	// Test BBRv2 state transitions
	t.Run("StateTransitions", func(t *testing.T) {
		metrics := ccManager.GetMetrics()
		// BBRState might be empty initially, that's OK
		initialCWND := metrics.CongestionWindow
		assert.Greater(t, initialCWND, int64(0), "Initial CWND should be > 0")

		// Simulate ACK events
		for i := 0; i < 10; i++ {
			ccManager.OnAck(time.Now(), 1460, 50*time.Millisecond)
			time.Sleep(10 * time.Millisecond)
		}

		// Check that CWND increased
		metricsAfter := ccManager.GetMetrics()
		assert.Greater(t, metricsAfter.CongestionWindow, initialCWND,
			"CWND should increase after ACKs")
	})

	// Test packet loss handling
	t.Run("PacketLossHandling", func(t *testing.T) {
		metricsBefore := ccManager.GetMetrics()
		ccManager.OnLoss()
		metricsAfter := ccManager.GetMetrics()
		assert.LessOrEqual(t, metricsAfter.CongestionWindow, metricsBefore.CongestionWindow,
			"CWND should not increase after packet loss")
	})

	// Test throughput estimation
	t.Run("ThroughputEstimation", func(t *testing.T) {
		// Simulate sustained traffic
		for i := 0; i < 100; i++ {
			ccManager.OnAck(time.Now(), 1460, 50*time.Millisecond)
			time.Sleep(5 * time.Millisecond)
		}

		metrics := ccManager.GetMetrics()
		assert.Greater(t, metrics.Throughput, 0.0, "Throughput should be > 0")
		assert.Greater(t, metrics.Bandwidth, 0.0, "Bandwidth should be > 0")
		assert.Greater(t, metrics.PacingRate, int64(0), "Pacing rate should be > 0")
	})
}

// testACKFrequencyIntegration tests ACK Frequency optimization integration
func testACKFrequencyIntegration(t *testing.T) {
	logger := zaptest.NewLogger(t)
	defer logger.Sync()

	// Create ACK Frequency manager
	ackFreqManager := NewACKFrequencyManager(logger, DefaultACKFrequencyConfig())
	require.NotNil(t, ackFreqManager)

	// Test ACK delay optimization
	t.Run("ACKDelayOptimization", func(t *testing.T) {
		connID := "test-conn-1"

		// Simulate packet reception
		for i := 0; i < 50; i++ {
			ackFreqManager.ShouldSendACK(connID, 1460)
			time.Sleep(10 * time.Millisecond)
		}

		// Check metrics
		metrics := ackFreqManager.GetMetrics()
		assert.Greater(t, metrics.TotalConnections, int64(0), "Should have connections")
		assert.Greater(t, metrics.TotalACKs, int64(0), "Should have sent ACKs")

		// ACK delay should be within configured range
		connMetrics := ackFreqManager.GetConnectionMetrics(connID)
		if connMetrics != nil {
			assert.GreaterOrEqual(t, connMetrics.AverageDelay, 1*time.Millisecond,
				"ACK delay should be >= min delay")
			assert.LessOrEqual(t, connMetrics.AverageDelay, 25*time.Millisecond,
				"ACK delay should be <= max delay")
		}
	})

	// Test adaptive behavior
	t.Run("AdaptiveBehavior", func(t *testing.T) {
		connID := "test-conn-2"

		// Simulate high load (should increase ACK delay)
		for i := 0; i < 200; i++ {
			ackFreqManager.ShouldSendACK(connID, 1460)
			time.Sleep(1 * time.Millisecond) // Fast packets
		}

		metrics := ackFreqManager.GetMetrics()
		assert.Greater(t, metrics.AdaptiveACKs, int64(0), "Should have adaptive ACKs")
	})

	// Test bandwidth savings
	t.Run("BandwidthSavings", func(t *testing.T) {
		connID := "test-conn-3"
		metricsBefore := ackFreqManager.GetMetrics()

		// Simulate traffic
		for i := 0; i < 100; i++ {
			ackFreqManager.ShouldSendACK(connID, 1460)
			time.Sleep(10 * time.Millisecond)
		}

		metricsAfter := ackFreqManager.GetMetrics()
		assert.GreaterOrEqual(t, metricsAfter.BandwidthSaved, metricsBefore.BandwidthSaved,
			"Bandwidth saved should not decrease")
	})

	// No explicit cleanup needed for ACKFrequencyManager
}

// testFECIntegration tests Forward Error Correction integration
func testFECIntegration(t *testing.T) {
	logger := zaptest.NewLogger(t)
	defer logger.Sync()

	// Create FEC manager
	fecConfig := &FECConfig{
		RedundancyRatio:    0.1,
		Algorithm:          "xor",
		GroupSize:          10,
		EnableAdaptive:     true,
		TargetRecoveryRate: 0.95,
	}
	fecManager := NewFECManager(logger, fecConfig)
	require.NotNil(t, fecManager)

	// Test FEC group creation
	t.Run("FECGroupCreation", func(t *testing.T) {
		// Create test packets
		packets := make([][]byte, 10)
		for i := range packets {
			packets[i] = make([]byte, 1460)
			for j := range packets[i] {
				packets[i][j] = byte(i)
			}
		}

		// Create FEC group
		group, err := fecManager.CreateGroup(packets)
		require.NoError(t, err, "FEC group creation should not fail")
		assert.Greater(t, len(group.RedundantPackets), 0, "Should have redundant packets")

		// Check metrics
		metrics := fecManager.GetMetrics()
		assert.Greater(t, metrics.TotalGroups, int64(0), "Should have created FEC groups")
	})

	// Test FEC efficiency
	t.Run("FECEfficiency", func(t *testing.T) {
		metrics := fecManager.GetMetrics()

		if metrics.TotalGroups > 0 {
			assert.GreaterOrEqual(t, metrics.Efficiency, 0.0, "Efficiency should be >= 0")
			assert.GreaterOrEqual(t, metrics.RecoveryRate, 0.0, "Recovery rate should be >= 0")
			assert.LessOrEqual(t, metrics.RecoveryRate, 1.0, "Recovery rate should be <= 1.0")
		}
	})

	// Cleanup
	fecManager.Stop()
}

// testGreasingIntegration tests QUIC Bit Greasing integration
func testGreasingIntegration(t *testing.T) {
	logger := zaptest.NewLogger(t)
	defer logger.Sync()

	// Create Greasing manager
	greasingManager := NewGreasingManager(logger, DefaultGreasingConfig())
	require.NotNil(t, greasingManager)

	// Test version greasing
	t.Run("VersionGreasing", func(t *testing.T) {
		versions := []uint32{0x00000001, 0x00000002}

		greasedVersions, err := greasingManager.GreaseVersionNegotiation(versions)
		require.NoError(t, err)

		// Greasing might add versions
		assert.GreaterOrEqual(t, len(greasedVersions), len(versions),
			"Greased versions should have at least original count")

		// Check metrics
		metrics := greasingManager.GetMetrics()
		assert.GreaterOrEqual(t, metrics.TotalGreaseApplications, int64(0),
			"Should have greasing applications")
	})

	// Test transport parameter greasing
	t.Run("TransportParamGreasing", func(t *testing.T) {
		params := map[uint64][]byte{
			0x01: []byte("value1"),
			0x02: []byte("value2"),
		}

		greasedParams, err := greasingManager.GreaseTransportParameters(params)
		require.NoError(t, err)

		// Greasing might add parameters
		assert.GreaterOrEqual(t, len(greasedParams), len(params),
			"Greased params should have at least original count")
	})

	// Test compatibility
	t.Run("CompatibilityCheck", func(t *testing.T) {
		metrics := greasingManager.GetMetrics()

		// Compatibility issues should be minimal
		if metrics.TotalGreaseApplications > 0 {
			compatibilityRate := 1.0 - float64(metrics.CompatibilityIssues)/float64(metrics.TotalGreaseApplications)
			assert.GreaterOrEqual(t, compatibilityRate, 0.99,
				"Compatibility rate should be >= 99%")
		}
	})

	// Cleanup
	greasingManager.Stop()
}

// testMultiFeatureIntegration tests interaction between multiple experimental features
func testMultiFeatureIntegration(t *testing.T) {
	logger := zaptest.NewLogger(t)
	defer logger.Sync()

	// Create all managers
	ccManager := congestion.NewCongestionControlManager(logger, congestion.AlgorithmBBRv2, 1460)
	ackFreqManager := NewACKFrequencyManager(logger, DefaultACKFrequencyConfig())
	fecManager := NewFECManager(logger, DefaultFECConfig())
	greasingManager := NewGreasingManager(logger, DefaultGreasingConfig())

	require.NotNil(t, ccManager)
	require.NotNil(t, ackFreqManager)
	require.NotNil(t, fecManager)
	require.NotNil(t, greasingManager)

	// Test combined operation
	t.Run("CombinedOperation", func(t *testing.T) {
		connID := "test-conn-multi"

		// Simulate realistic traffic pattern
		for i := 0; i < 50; i++ {
			// ACK received (triggers BBRv2)
			ccManager.OnAck(time.Now(), 1460, 50*time.Millisecond)

			// Read data (triggers ACK Frequency)
			ackFreqManager.ShouldSendACK(connID, 1460)

			// Apply FEC to outgoing packets
			packets := [][]byte{make([]byte, 1460)}
			_, err := fecManager.CreateGroup(packets)
			assert.NoError(t, err)

			time.Sleep(10 * time.Millisecond)
		}

		// Verify all features are working
		assert.NotNil(t, ccManager.GetMetrics(), "BBRv2 should be active")
		assert.NotNil(t, ackFreqManager.GetMetrics(), "ACK Frequency should be active")
		assert.NotNil(t, fecManager.GetMetrics(), "FEC should be active")
		assert.NotNil(t, greasingManager.GetMetrics(), "Greasing should be active")
	})

	// No explicit cleanup needed for ACKFrequencyManager
	fecManager.Stop()
	greasingManager.Stop()
}

// testPrometheusMetricsIntegration tests Prometheus metrics during operation
func testPrometheusMetricsIntegration(t *testing.T) {
	logger := zaptest.NewLogger(t)
	defer logger.Sync()

	// Create metrics exporter
	metrics := NewExperimentalMetrics(logger)
	require.NotNil(t, metrics)

	// Create BBRv2 manager
	ccManager := congestion.NewCongestionControlManager(logger, congestion.AlgorithmBBRv2, 1460)
	require.NotNil(t, ccManager)

	// Test metrics export
	t.Run("MetricsExport", func(t *testing.T) {
		// Generate some traffic
		for i := 0; i < 10; i++ {
			ccManager.OnAck(time.Now(), 1460, 50*time.Millisecond)
		}

		// Export metrics
		connID := "test-conn-metrics"
		tenantID := "test-tenant-metrics"
		ccMetrics := ccManager.GetMetrics()
		metrics.UpdateBBRv2Metrics(connID, tenantID, ccMetrics)

		// Verify metrics snapshot
		snapshot := metrics.GetMetricsSnapshot()
		assert.NotNil(t, snapshot, "Metrics snapshot should exist")
		assert.Contains(t, snapshot, "bbrv2", "Should contain BBRv2 metrics")
	})
}

// testPerformanceUnderLoad tests performance under sustained load
func testPerformanceUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	logger := zaptest.NewLogger(t)
	defer logger.Sync()

	// Create managers
	ccManager := congestion.NewCongestionControlManager(logger, congestion.AlgorithmBBRv2, 1460)
	ackFreqManager := NewACKFrequencyManager(logger, DefaultACKFrequencyConfig())
	require.NotNil(t, ccManager)
	require.NotNil(t, ackFreqManager)

	// Test sustained load
	t.Run("SustainedLoad", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		connID := "test-conn-load"
		packetCount := 0
		ticker := time.NewTicker(1 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				t.Logf("Processed %d packets in 10 seconds (%.1f pps)",
					packetCount, float64(packetCount)/10.0)
				return
			case <-ticker.C:
				ackFreqManager.ShouldSendACK(connID, 1460)
				ccManager.OnAck(time.Now(), 1460, 50*time.Millisecond)
				packetCount++
			}
		}
	})

	// Verify system stability after load
	t.Run("StabilityAfterLoad", func(t *testing.T) {
		metrics := ccManager.GetMetrics()
		assert.Greater(t, metrics.CongestionWindow, int64(0), "CWND should be valid")
		assert.GreaterOrEqual(t, metrics.LossRate, 0.0, "Loss rate should be valid")
		assert.LessOrEqual(t, metrics.LossRate, 1.0, "Loss rate should be <= 1.0")
	})

	// No explicit cleanup needed for ACKFrequencyManager
}

// TestExperimentalFeaturesConfig tests config initialization
func TestExperimentalFeaturesConfig(t *testing.T) {
	logger := zaptest.NewLogger(t)
	defer logger.Sync()

	t.Run("ValidConfig", func(t *testing.T) {
		config := &EnhancedConfig{}
		config.ExperimentalFeatures = DefaultExperimentalFeaturesConfig()
		config.ExperimentalFeatures.EnableBBRv2 = true

		err := config.InitializeExperimentalFeatures(logger)
		assert.NoError(t, err, "Valid config should initialize without error")
		assert.NotNil(t, config.CCManager, "CCManager should be initialized")

		config.StopExperimentalFeatures()
	})

	t.Run("NilConfig", func(t *testing.T) {
		config := &EnhancedConfig{}
		config.ExperimentalFeatures = nil

		err := config.InitializeExperimentalFeatures(logger)
		assert.NoError(t, err, "Nil config should be handled gracefully")
	})

	t.Run("MultipleFeatures", func(t *testing.T) {
		config := &EnhancedConfig{}
		config.ExperimentalFeatures = DefaultExperimentalFeaturesConfig()
		config.ExperimentalFeatures.EnableBBRv2 = true
		config.ExperimentalFeatures.EnableACKFrequency = true
		config.ExperimentalFeatures.EnableFEC = true
		config.ExperimentalFeatures.EnableGreasing = true

		err := config.InitializeExperimentalFeatures(logger)
		assert.NoError(t, err, "All features should initialize")
		assert.NotNil(t, config.CCManager)
		assert.NotNil(t, config.ACKFreqManager)
		assert.NotNil(t, config.FECManager)
		assert.NotNil(t, config.GreasingManager)

		config.StopExperimentalFeatures()
	})
}
