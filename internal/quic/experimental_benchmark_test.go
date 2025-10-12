package quic

import (
	"testing"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/quic/congestion"
	"go.uber.org/zap/zaptest"
)

// BenchmarkBBRv2 benchmarks BBRv2 congestion control performance
func BenchmarkBBRv2(b *testing.B) {
	logger := zaptest.NewLogger(b)
	defer logger.Sync()

	scenarios := []struct {
		name string
		rtt  time.Duration
		mtu  int
	}{
		{"RTT_10ms_MTU_1460", 10 * time.Millisecond, 1460},
		{"RTT_50ms_MTU_1460", 50 * time.Millisecond, 1460},
		{"RTT_100ms_MTU_1460", 100 * time.Millisecond, 1460},
		{"RTT_200ms_MTU_1460", 200 * time.Millisecond, 1460},
		{"RTT_500ms_MTU_1460", 500 * time.Millisecond, 1460},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			ccManager := congestion.NewCongestionControlManager(logger, congestion.AlgorithmBBRv2, scenario.mtu)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ccManager.OnAck(time.Now(), scenario.mtu, scenario.rtt)
			}
			b.StopTimer()

			// Report metrics
			metrics := ccManager.GetMetrics()
			b.ReportMetric(float64(metrics.CongestionWindow), "cwnd_bytes")
			b.ReportMetric(float64(metrics.PacingRate), "pacing_rate_bps")
			b.ReportMetric(metrics.Throughput, "throughput_mbps")
		})
	}
}

// BenchmarkBBRv2vsC UBIC compares BBRv2 vs CUBIC performance
func BenchmarkBBRv2vsCUBIC(b *testing.B) {
	logger := zaptest.NewLogger(b)
	defer logger.Sync()

	scenarios := []struct {
		name      string
		algorithm congestion.CongestionControlAlgorithm
		rtt       time.Duration
	}{
		{"BBRv2_50ms", congestion.AlgorithmBBRv2, 50 * time.Millisecond},
		{"CUBIC_50ms", congestion.AlgorithmCubic, 50 * time.Millisecond},
		{"BBRv2_200ms", congestion.AlgorithmBBRv2, 200 * time.Millisecond},
		{"CUBIC_200ms", congestion.AlgorithmCubic, 200 * time.Millisecond},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			ccManager := congestion.NewCongestionControlManager(logger, scenario.algorithm, 1460)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ccManager.OnAck(time.Now(), 1460, scenario.rtt)
			}
			b.StopTimer()

			metrics := ccManager.GetMetrics()
			b.ReportMetric(float64(metrics.CongestionWindow), "cwnd_bytes")
			b.ReportMetric(metrics.Throughput, "throughput_mbps")
		})
	}
}

// BenchmarkACKFrequency benchmarks ACK Frequency optimization
func BenchmarkACKFrequency(b *testing.B) {
	logger := zaptest.NewLogger(b)
	defer logger.Sync()

	scenarios := []struct {
		name       string
		packetSize int
		adaptive   bool
	}{
		{"PacketSize_1460_Adaptive", 1460, true},
		{"PacketSize_1460_Static", 1460, false},
		{"PacketSize_500_Adaptive", 500, true},
		{"PacketSize_9000_Adaptive", 9000, true},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			config := DefaultACKFrequencyConfig()
			config.EnableAdaptive = scenario.adaptive
			ackFreqManager := NewACKFrequencyManager(logger, config)

			connID := "bench-conn"
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ackFreqManager.ShouldSendACK(connID, scenario.packetSize)
			}
			b.StopTimer()

			metrics := ackFreqManager.GetMetrics()
			if metrics.TotalACKs > 0 {
				overhead := float64(metrics.TotalACKs) / float64(b.N) * 100
				b.ReportMetric(overhead, "ack_overhead_%")
			}
			b.ReportMetric(float64(metrics.BandwidthSaved), "bandwidth_saved_bytes")
		})
	}
}

// BenchmarkFEC benchmarks Forward Error Correction performance
func BenchmarkFEC(b *testing.B) {
	logger := zaptest.NewLogger(b)
	defer logger.Sync()

	scenarios := []struct {
		name       string
		groupSize  int
		redundancy float64
		algorithm  string
	}{
		{"XOR_GroupSize_10_Redundancy_10%", 10, 0.1, "xor"},
		{"XOR_GroupSize_10_Redundancy_20%", 10, 0.2, "xor"},
		{"XOR_GroupSize_20_Redundancy_10%", 20, 0.1, "xor"},
		{"ReedSolomon_GroupSize_10_Redundancy_10%", 10, 0.1, "reed_solomon"},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			config := &FECConfig{
				RedundancyRatio:    scenario.redundancy,
				Algorithm:          scenario.algorithm,
				GroupSize:          scenario.groupSize,
				EnableAdaptive:     false,
				TargetRecoveryRate: 0.95,
			}
			fecManager := NewFECManager(logger, config)

			// Create test packets
			packets := make([][]byte, scenario.groupSize)
			for i := range packets {
				packets[i] = make([]byte, 1460)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := fecManager.CreateGroup(packets)
				if err != nil {
					b.Fatal(err)
				}
			}
			b.StopTimer()

			metrics := fecManager.GetMetrics()
			b.ReportMetric(metrics.RecoveryRate*100, "recovery_rate_%")
			b.ReportMetric(metrics.Efficiency, "efficiency")
		})
	}
}

// BenchmarkGreasing benchmarks QUIC Bit Greasing performance
func BenchmarkGreasing(b *testing.B) {
	logger := zaptest.NewLogger(b)
	defer logger.Sync()

	scenarios := []struct {
		name        string
		probability float64
	}{
		{"Probability_10%", 0.1},
		{"Probability_50%", 0.5},
		{"Probability_100%", 1.0},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			config := DefaultGreasingConfig()
			config.GreaseProbability = scenario.probability
			greasingManager := NewGreasingManager(logger, config)

			versions := []uint32{0x00000001, 0x00000002}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := greasingManager.GreaseVersionNegotiation(versions)
				if err != nil {
					b.Fatal(err)
				}
			}
			b.StopTimer()

			metrics := greasingManager.GetMetrics()
			if metrics.TotalGreaseApplications > 0 {
				successRate := (1.0 - float64(metrics.CompatibilityIssues)/float64(metrics.TotalGreaseApplications)) * 100
				b.ReportMetric(successRate, "success_rate_%")
			}
		})
	}
}

// BenchmarkMultiFeature benchmarks combined feature performance
func BenchmarkMultiFeature(b *testing.B) {
	logger := zaptest.NewLogger(b)
	defer logger.Sync()

	scenarios := []struct {
		name         string
		enableBBRv2  bool
		enableACKFreq bool
		enableFEC    bool
		rtt          time.Duration
	}{
		{"BBRv2_Only_50ms", true, false, false, 50 * time.Millisecond},
		{"BBRv2_ACKFreq_50ms", true, true, false, 50 * time.Millisecond},
		{"BBRv2_FEC_50ms", true, false, true, 50 * time.Millisecond},
		{"All_Features_50ms", true, true, true, 50 * time.Millisecond},
		{"All_Features_200ms", true, true, true, 200 * time.Millisecond},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			var ccManager *congestion.CongestionControlManager
			var ackFreqManager *ACKFrequencyManager
			var fecManager *FECManager

			if scenario.enableBBRv2 {
				ccManager = congestion.NewCongestionControlManager(logger, congestion.AlgorithmBBRv2, 1460)
			}
			if scenario.enableACKFreq {
				ackFreqManager = NewACKFrequencyManager(logger, DefaultACKFrequencyConfig())
			}
			if scenario.enableFEC {
				fecManager = NewFECManager(logger, DefaultFECConfig())
			}

			connID := "bench-conn"
			packets := [][]byte{make([]byte, 1460)}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if ccManager != nil {
					ccManager.OnAck(time.Now(), 1460, scenario.rtt)
				}
				if ackFreqManager != nil {
					ackFreqManager.ShouldSendACK(connID, 1460)
				}
				if fecManager != nil {
					_, _ = fecManager.CreateGroup(packets)
				}
			}
			b.StopTimer()

			// Report combined metrics
			if ccManager != nil {
				metrics := ccManager.GetMetrics()
				b.ReportMetric(metrics.Throughput, "throughput_mbps")
			}
		})
	}
}

// BenchmarkPacketLoss benchmarks performance under packet loss
func BenchmarkPacketLoss(b *testing.B) {
	logger := zaptest.NewLogger(b)
	defer logger.Sync()

	scenarios := []struct {
		name     string
		lossRate float64 // 0.0 - 1.0
	}{
		{"Loss_0%", 0.0},
		{"Loss_0.5%", 0.005},
		{"Loss_1%", 0.01},
		{"Loss_5%", 0.05},
		{"Loss_10%", 0.10},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			ccManager := congestion.NewCongestionControlManager(logger, congestion.AlgorithmBBRv2, 1460)
			fecManager := NewFECManager(logger, DefaultFECConfig())

			lossCounter := 0
			lossThreshold := int(1.0 / scenario.lossRate)
			if scenario.lossRate == 0 {
				lossThreshold = b.N + 1 // No loss
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Simulate packet loss
				if i%lossThreshold == 0 && i > 0 {
					ccManager.OnLoss()
					lossCounter++
				} else {
					ccManager.OnAck(time.Now(), 1460, 50*time.Millisecond)
				}

				// FEC can recover some losses
				packets := [][]byte{make([]byte, 1460)}
				_, _ = fecManager.CreateGroup(packets)
			}
			b.StopTimer()

			metrics := ccManager.GetMetrics()
			fecMetrics := fecManager.GetMetrics()

			b.ReportMetric(metrics.Throughput, "throughput_mbps")
			b.ReportMetric(metrics.LossRate*100, "observed_loss_rate_%")
			b.ReportMetric(fecMetrics.RecoveryRate*100, "fec_recovery_rate_%")
		})
	}
}

// BenchmarkLoadLevels benchmarks performance at different load levels
func BenchmarkLoadLevels(b *testing.B) {
	logger := zaptest.NewLogger(b)
	defer logger.Sync()

	scenarios := []struct {
		name string
		pps  int // packets per second
	}{
		{"Load_100pps", 100},
		{"Load_500pps", 500},
		{"Load_1000pps", 1000},
		{"Load_2000pps", 2000},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			ccManager := congestion.NewCongestionControlManager(logger, congestion.AlgorithmBBRv2, 1460)
			ackFreqManager := NewACKFrequencyManager(logger, DefaultACKFrequencyConfig())

			connID := "bench-conn"
			interval := time.Second / time.Duration(scenario.pps)

			b.ResetTimer()
			startTime := time.Now()
			for i := 0; i < b.N; i++ {
				ccManager.OnAck(time.Now(), 1460, 50*time.Millisecond)
				ackFreqManager.ShouldSendACK(connID, 1460)

				// Simulate pacing
				time.Sleep(interval)
			}
			duration := time.Since(startTime)
			b.StopTimer()

			actualPPS := float64(b.N) / duration.Seconds()
			metrics := ccManager.GetMetrics()

			b.ReportMetric(actualPPS, "actual_pps")
			b.ReportMetric(metrics.Throughput, "throughput_mbps")
		})
	}
}

// BenchmarkPrometheusExport benchmarks Prometheus metrics export overhead
func BenchmarkPrometheusExport(b *testing.B) {
	logger := zaptest.NewLogger(b)
	defer logger.Sync()

	metricsExporter := NewExperimentalMetrics(logger)
	ccManager := congestion.NewCongestionControlManager(logger, congestion.AlgorithmBBRv2, 1460)

	connID := "bench-conn"
	tenantID := "bench-tenant"

	// Warm up
	ccManager.OnAck(time.Now(), 1460, 50*time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ccManager.OnAck(time.Now(), 1460, 50*time.Millisecond)
		metrics := ccManager.GetMetrics()
		metricsExporter.UpdateBBRv2Metrics(connID, tenantID, metrics)
	}
	b.StopTimer()

	b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "metrics_ops/sec")
}

// BenchmarkMemoryAllocation benchmarks memory allocation patterns
func BenchmarkMemoryAllocation(b *testing.B) {
	logger := zaptest.NewLogger(b)
	defer logger.Sync()

	scenarios := []struct {
		name string
		size int
	}{
		{"FEC_Group_10_Packets", 10},
		{"FEC_Group_50_Packets", 50},
		{"FEC_Group_100_Packets", 100},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			fecManager := NewFECManager(logger, DefaultFECConfig())

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				packets := make([][]byte, scenario.size)
				for j := range packets {
					packets[j] = make([]byte, 1460)
				}
				_, _ = fecManager.CreateGroup(packets)
			}
			b.StopTimer()
		})
	}
}

// Benchmark helper functions

// simulateRTT simulates network RTT delay
func simulateRTT(rtt time.Duration) {
	time.Sleep(rtt / 2) // Half RTT for one-way delay
}

// simulatePacketLoss simulates packet loss based on probability
func simulatePacketLoss(probability float64, counter int) bool {
	if probability == 0 {
		return false
	}
	return counter%int(1.0/probability) == 0
}

// calculateThroughput calculates throughput from bytes and duration
func calculateThroughput(bytes int64, duration time.Duration) float64 {
	if duration == 0 {
		return 0
	}
	bitsPerSecond := float64(bytes*8) / duration.Seconds()
	return bitsPerSecond / 1_000_000 // Convert to Mbps
}

// Example benchmark usage documentation
func ExampleBenchmarkBBRv2() {
	// To run benchmarks:
	//
	// # Run all benchmarks
	// go test -bench=. -benchmem ./internal/quic/
	//
	// # Run specific benchmark
	// go test -bench=BenchmarkBBRv2 -benchmem ./internal/quic/
	//
	// # Run with custom iterations
	// go test -bench=BenchmarkBBRv2 -benchtime=10s ./internal/quic/
	//
	// # Save benchmark results
	// go test -bench=. -benchmem ./internal/quic/ > benchmark_results.txt
	//
	// # Compare benchmarks
	// go test -bench=BenchmarkBBRv2vsCUBIC -benchmem ./internal/quic/ | tee new.txt
	// benchstat old.txt new.txt
}
