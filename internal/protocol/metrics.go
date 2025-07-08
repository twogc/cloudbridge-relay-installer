package protocol

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// ProtocolMetricsCollector collects and manages protocol-specific metrics
type ProtocolMetricsCollector struct {
	// Prometheus metrics
	latencyHistogram    *prometheus.HistogramVec
	throughputGauge     *prometheus.GaugeVec
	errorRateCounter    *prometheus.CounterVec
	successRateGauge    *prometheus.GaugeVec
	switchCounter       *prometheus.CounterVec
	connectionGauge     *prometheus.GaugeVec
	bandwidthGauge      *prometheus.GaugeVec
	packetLossGauge     *prometheus.GaugeVec
	rttGauge            *prometheus.GaugeVec

	// Internal metrics storage
	metrics     map[Protocol]*DetailedMetrics
	mu          sync.RWMutex
	logger      *zap.Logger
	windowSize  time.Duration
}

// DetailedMetrics holds detailed metrics for a protocol
type DetailedMetrics struct {
	Latency     *RollingWindow
	Throughput  *RollingWindow
	ErrorRate   *RollingWindow
	SuccessRate *RollingWindow
	Connections int64
	Bandwidth   float64
	PacketLoss  float64
	RTT         time.Duration
	LastUpdate  time.Time
}

// RollingWindow implements a rolling window for metrics
type RollingWindow struct {
	values    []float64
	times     []time.Time
	window    time.Duration
	maxSize   int
	mu        sync.RWMutex
}

// NewRollingWindow creates a new rolling window
func NewRollingWindow(window time.Duration, maxSize int) *RollingWindow {
	return &RollingWindow{
		values:  make([]float64, 0, maxSize),
		times:   make([]time.Time, 0, maxSize),
		window:  window,
		maxSize: maxSize,
	}
}

// Add adds a value to the rolling window
func (rw *RollingWindow) Add(value float64) {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	now := time.Now()
	rw.values = append(rw.values, value)
	rw.times = append(rw.times, now)

	// Remove old values outside the window
	rw.cleanup(now)
}

// cleanup removes values outside the time window
func (rw *RollingWindow) cleanup(now time.Time) {
	cutoff := now.Add(-rw.window)
	
	// Find the first index that's within the window
	start := 0
	for i, t := range rw.times {
		if t.After(cutoff) {
			start = i
			break
		}
	}

	// Remove old values
	if start > 0 {
		rw.values = rw.values[start:]
		rw.times = rw.times[start:]
	}

	// Limit size
	if len(rw.values) > rw.maxSize {
		rw.values = rw.values[len(rw.values)-rw.maxSize:]
		rw.times = rw.times[len(rw.times)-rw.maxSize:]
	}
}

// Average returns the average of values in the window
func (rw *RollingWindow) Average() float64 {
	rw.mu.RLock()
	defer rw.mu.RUnlock()

	if len(rw.values) == 0 {
		return 0
	}

	sum := 0.0
	for _, v := range rw.values {
		sum += v
	}
	return sum / float64(len(rw.values))
}

// Count returns the number of values in the window
func (rw *RollingWindow) Count() int {
	rw.mu.RLock()
	defer rw.mu.RUnlock()
	return len(rw.values)
}

// NewProtocolMetricsCollector creates a new metrics collector
func NewProtocolMetricsCollector(windowSize time.Duration, logger *zap.Logger) *ProtocolMetricsCollector {
	collector := &ProtocolMetricsCollector{
		latencyHistogram: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "protocol_latency_seconds",
				Help: "Protocol latency in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
			},
			[]string{"protocol"},
		),
		throughputGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "protocol_throughput_bytes_per_second",
				Help: "Protocol throughput in bytes per second",
			},
			[]string{"protocol"},
		),
		errorRateCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "protocol_errors_total",
				Help: "Total number of protocol errors",
			},
			[]string{"protocol", "error_type"},
		),
		successRateGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "protocol_success_rate",
				Help: "Protocol success rate (0.0 to 1.0)",
			},
			[]string{"protocol"},
		),
		switchCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "protocol_switches_total",
				Help: "Total number of protocol switches",
			},
			[]string{"from", "to", "reason"},
		),
		connectionGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "protocol_active_connections",
				Help: "Number of active connections per protocol",
			},
			[]string{"protocol"},
		),
		bandwidthGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "protocol_bandwidth_utilization",
				Help: "Bandwidth utilization percentage",
			},
			[]string{"protocol"},
		),
		packetLossGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "protocol_packet_loss_percentage",
				Help: "Packet loss percentage",
			},
			[]string{"protocol"},
		),
		rttGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "protocol_rtt_seconds",
				Help: "Round-trip time in seconds",
			},
			[]string{"protocol"},
		),
		metrics:    make(map[Protocol]*DetailedMetrics),
		logger:     logger,
		windowSize: windowSize,
	}

	// Register Prometheus metrics
	prometheus.MustRegister(
		collector.latencyHistogram,
		collector.throughputGauge,
		collector.errorRateCounter,
		collector.successRateGauge,
		collector.switchCounter,
		collector.connectionGauge,
		collector.bandwidthGauge,
		collector.packetLossGauge,
		collector.rttGauge,
	)

	return collector
}

// RecordLatency records latency for a protocol
func (pmc *ProtocolMetricsCollector) RecordLatency(protocol Protocol, latency time.Duration) {
	pmc.mu.Lock()
	defer pmc.mu.Unlock()

	// Update Prometheus metric
	pmc.latencyHistogram.WithLabelValues(protocol.String()).Observe(latency.Seconds())

	// Update internal metrics
	metrics := pmc.getOrCreateMetrics(protocol)
	metrics.Latency.Add(float64(latency.Milliseconds()))
	metrics.LastUpdate = time.Now()

	pmc.logger.Debug("Recorded protocol latency",
		zap.String("protocol", protocol.String()),
		zap.Duration("latency", latency),
	)
}

// RecordThroughput records throughput for a protocol
func (pmc *ProtocolMetricsCollector) RecordThroughput(protocol Protocol, bytesPerSecond float64) {
	pmc.mu.Lock()
	defer pmc.mu.Unlock()

	// Update Prometheus metric
	pmc.throughputGauge.WithLabelValues(protocol.String()).Set(bytesPerSecond)

	// Update internal metrics
	metrics := pmc.getOrCreateMetrics(protocol)
	metrics.Throughput.Add(bytesPerSecond)
	metrics.Bandwidth = bytesPerSecond
	metrics.LastUpdate = time.Now()
}

// RecordError records an error for a protocol
func (pmc *ProtocolMetricsCollector) RecordError(protocol Protocol, errorType string) {
	pmc.mu.Lock()
	defer pmc.mu.Unlock()

	// Update Prometheus metric
	pmc.errorRateCounter.WithLabelValues(protocol.String(), errorType).Inc()

	// Update internal metrics
	metrics := pmc.getOrCreateMetrics(protocol)
	metrics.ErrorRate.Add(1.0)
	metrics.LastUpdate = time.Now()

	pmc.logger.Warn("Recorded protocol error",
		zap.String("protocol", protocol.String()),
		zap.String("error_type", errorType),
	)
}

// RecordSuccess records a successful operation for a protocol
func (pmc *ProtocolMetricsCollector) RecordSuccess(protocol Protocol) {
	pmc.mu.Lock()
	defer pmc.mu.Unlock()

	// Update internal metrics
	metrics := pmc.getOrCreateMetrics(protocol)
	metrics.SuccessRate.Add(1.0)
	metrics.LastUpdate = time.Now()

	// Update Prometheus success rate
	successRate := metrics.SuccessRate.Average()
	pmc.successRateGauge.WithLabelValues(protocol.String()).Set(successRate)
}

// RecordSwitch records a protocol switch
func (pmc *ProtocolMetricsCollector) RecordSwitch(from, to Protocol, reason string) {
	pmc.switchCounter.WithLabelValues(from.String(), to.String(), reason).Inc()

	pmc.logger.Info("Recorded protocol switch",
		zap.String("from", from.String()),
		zap.String("to", to.String()),
		zap.String("reason", reason),
	)
}

// UpdateConnections updates the number of active connections
func (pmc *ProtocolMetricsCollector) UpdateConnections(protocol Protocol, count int64) {
	pmc.mu.Lock()
	defer pmc.mu.Unlock()

	// Update Prometheus metric
	pmc.connectionGauge.WithLabelValues(protocol.String()).Set(float64(count))

	// Update internal metrics
	metrics := pmc.getOrCreateMetrics(protocol)
	metrics.Connections = count
	metrics.LastUpdate = time.Now()
}

// UpdatePacketLoss updates packet loss for a protocol
func (pmc *ProtocolMetricsCollector) UpdatePacketLoss(protocol Protocol, lossPercentage float64) {
	pmc.mu.Lock()
	defer pmc.mu.Unlock()

	// Update Prometheus metric
	pmc.packetLossGauge.WithLabelValues(protocol.String()).Set(lossPercentage)

	// Update internal metrics
	metrics := pmc.getOrCreateMetrics(protocol)
	metrics.PacketLoss = lossPercentage
	metrics.LastUpdate = time.Now()
}

// UpdateRTT updates round-trip time for a protocol
func (pmc *ProtocolMetricsCollector) UpdateRTT(protocol Protocol, rtt time.Duration) {
	pmc.mu.Lock()
	defer pmc.mu.Unlock()

	// Update Prometheus metric
	pmc.rttGauge.WithLabelValues(protocol.String()).Set(rtt.Seconds())

	// Update internal metrics
	metrics := pmc.getOrCreateMetrics(protocol)
	metrics.RTT = rtt
	metrics.LastUpdate = time.Now()
}

// getOrCreateMetrics gets or creates metrics for a protocol
func (pmc *ProtocolMetricsCollector) getOrCreateMetrics(protocol Protocol) *DetailedMetrics {
	metrics, exists := pmc.metrics[protocol]
	if !exists {
		metrics = &DetailedMetrics{
			Latency:     NewRollingWindow(pmc.windowSize, 1000),
			Throughput:  NewRollingWindow(pmc.windowSize, 1000),
			ErrorRate:   NewRollingWindow(pmc.windowSize, 1000),
			SuccessRate: NewRollingWindow(pmc.windowSize, 1000),
			LastUpdate:  time.Now(),
		}
		pmc.metrics[protocol] = metrics
	}
	return metrics
}

// GetMetrics returns detailed metrics for a protocol
func (pmc *ProtocolMetricsCollector) GetMetrics(protocol Protocol) *DetailedMetrics {
	pmc.mu.RLock()
	defer pmc.mu.RUnlock()

	if metrics, exists := pmc.metrics[protocol]; exists {
		return metrics
	}
	return &DetailedMetrics{}
}

// GetAllMetrics returns metrics for all protocols
func (pmc *ProtocolMetricsCollector) GetAllMetrics() map[Protocol]*DetailedMetrics {
	pmc.mu.RLock()
	defer pmc.mu.RUnlock()

	result := make(map[Protocol]*DetailedMetrics)
	for protocol, metrics := range pmc.metrics {
		result[protocol] = metrics
	}
	return result
}

// GetProtocolPerformance returns performance comparison between protocols
func (pmc *ProtocolMetricsCollector) GetProtocolPerformance() map[string]interface{} {
	pmc.mu.RLock()
	defer pmc.mu.RUnlock()

	performance := make(map[string]interface{})
	
	for protocol, metrics := range pmc.metrics {
		performance[protocol.String()] = map[string]interface{}{
			"latency_ms":      metrics.Latency.Average(),
			"throughput_bps":  metrics.Throughput.Average(),
			"error_rate":      metrics.ErrorRate.Average(),
			"success_rate":    metrics.SuccessRate.Average(),
			"connections":     metrics.Connections,
			"bandwidth":       metrics.Bandwidth,
			"packet_loss":     metrics.PacketLoss,
			"rtt_ms":          metrics.RTT.Milliseconds(),
			"last_update":     metrics.LastUpdate,
			"sample_count": map[string]int{
				"latency":     metrics.Latency.Count(),
				"throughput":  metrics.Throughput.Count(),
				"error_rate":  metrics.ErrorRate.Count(),
				"success_rate": metrics.SuccessRate.Count(),
			},
		}
	}

	return performance
}

// Cleanup removes old metrics
func (pmc *ProtocolMetricsCollector) Cleanup() {
	pmc.mu.Lock()
	defer pmc.mu.Unlock()

	cutoff := time.Now().Add(-pmc.windowSize)
	for protocol, metrics := range pmc.metrics {
		if metrics.LastUpdate.Before(cutoff) {
			delete(pmc.metrics, protocol)
			pmc.logger.Debug("Cleaned up old metrics",
				zap.String("protocol", protocol.String()),
			)
		}
	}
} 