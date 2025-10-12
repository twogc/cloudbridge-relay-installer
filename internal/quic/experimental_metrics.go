package quic

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/twogc/cloudbridge-relay-installer/internal/quic/congestion"
	"go.uber.org/zap"
)

// ExperimentalMetrics contains Prometheus metrics for all experimental QUIC features
// Includes: BBRv2, ACK Frequency, FEC, and QUIC Bit Greasing
type ExperimentalMetrics struct {
	logger *zap.Logger

	// BBRv2 Congestion Control Metrics
	bbrv2State             *prometheus.GaugeVec
	bbrv2CongestionWindow  *prometheus.GaugeVec
	bbrv2PacingRate        *prometheus.GaugeVec
	bbrv2Bandwidth         *prometheus.GaugeVec
	bbrv2MinRTT            *prometheus.GaugeVec
	bbrv2Throughput        *prometheus.GaugeVec
	bbrv2LossRate          *prometheus.GaugeVec
	bbrv2StateTransitions  *prometheus.CounterVec
	bbrv2PacketsInFlight   *prometheus.GaugeVec

	// ACK Frequency Optimization Metrics
	ackFreqTotalConnections  prometheus.Gauge
	ackFreqActiveConnections prometheus.Gauge
	ackFreqTotalACKs         prometheus.Counter
	ackFreqDelayedACKs       prometheus.Counter
	ackFreqAdaptiveACKs      prometheus.Counter
	ackFreqAverageDelay      prometheus.Histogram
	ackFreqBandwidthSaved    prometheus.Counter
	ackFreqPacketsReceived   *prometheus.CounterVec
	ackFreqBytesReceived     *prometheus.CounterVec

	// Forward Error Correction (FEC) Metrics
	fecTotalGroups       prometheus.Counter
	fecActiveGroups      prometheus.Gauge
	fecRedundancyBytes   prometheus.Counter
	fecRecoveryEvents    prometheus.Counter
	fecFailedRecoveries  prometheus.Counter
	fecRecoveryRate      prometheus.Gauge
	fecEfficiency        prometheus.Gauge
	fecLatency           prometheus.Histogram
	fecBytesSaved        prometheus.Counter
	fecDataPackets       *prometheus.CounterVec
	fecRedundantPackets  *prometheus.CounterVec

	// QUIC Bit Greasing (RFC 9287) Metrics
	greasingVersionCount        prometheus.Counter
	greasingTransportParamCount prometheus.Counter
	greasingFrameTypeCount      prometheus.Counter
	greasingTotalApplications   prometheus.Counter
	greasingCompatibilityIssues prometheus.Counter
	greasingActiveConnections   prometheus.Gauge

	// Cross-cutting Performance Metrics
	experimentalConnectionsTotal       *prometheus.CounterVec
	experimentalConnectionsActive      *prometheus.GaugeVec
	experimentalLatencyImprovement     *prometheus.GaugeVec
	experimentalThroughputImprovement  *prometheus.GaugeVec
	experimentalPacketLossReduction    *prometheus.GaugeVec
}

// NewExperimentalMetrics creates a new Prometheus metrics exporter for experimental features
func NewExperimentalMetrics(logger *zap.Logger) *ExperimentalMetrics {
	return &ExperimentalMetrics{
		logger: logger,

		// BBRv2 Metrics
		bbrv2State: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "cloudbridge_quic_bbrv2_state",
			Help: "BBRv2 congestion control state (0=Startup, 1=Drain, 2=ProbeBW, 3=ProbeRTT)",
		}, []string{"connection_id", "tenant_id"}),

		bbrv2CongestionWindow: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "cloudbridge_quic_bbrv2_congestion_window_bytes",
			Help: "BBRv2 congestion window size in bytes",
		}, []string{"connection_id", "tenant_id"}),

		bbrv2PacingRate: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "cloudbridge_quic_bbrv2_pacing_rate_bps",
			Help: "BBRv2 pacing rate in bits per second",
		}, []string{"connection_id", "tenant_id"}),

		bbrv2Bandwidth: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "cloudbridge_quic_bbrv2_bandwidth_bps",
			Help: "BBRv2 estimated bandwidth in bits per second",
		}, []string{"connection_id", "tenant_id"}),

		bbrv2MinRTT: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "cloudbridge_quic_bbrv2_min_rtt_seconds",
			Help: "BBRv2 minimum RTT in seconds",
		}, []string{"connection_id", "tenant_id"}),

		bbrv2Throughput: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "cloudbridge_quic_bbrv2_throughput_mbps",
			Help: "BBRv2 throughput in Mbps",
		}, []string{"connection_id", "tenant_id"}),

		bbrv2LossRate: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "cloudbridge_quic_bbrv2_loss_rate",
			Help: "BBRv2 packet loss rate (0.0 to 1.0)",
		}, []string{"connection_id", "tenant_id"}),

		bbrv2StateTransitions: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "cloudbridge_quic_bbrv2_state_transitions_total",
			Help: "Total number of BBRv2 state transitions",
		}, []string{"from_state", "to_state", "tenant_id"}),

		bbrv2PacketsInFlight: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "cloudbridge_quic_bbrv2_packets_in_flight",
			Help: "Number of packets in flight (sent but not acknowledged)",
		}, []string{"connection_id", "tenant_id"}),

		// ACK Frequency Metrics
		ackFreqTotalConnections: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "cloudbridge_quic_ack_freq_total_connections",
			Help: "Total number of connections with ACK frequency optimization",
		}),

		ackFreqActiveConnections: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "cloudbridge_quic_ack_freq_active_connections",
			Help: "Number of active connections with ACK frequency optimization",
		}),

		ackFreqTotalACKs: promauto.NewCounter(prometheus.CounterOpts{
			Name: "cloudbridge_quic_ack_freq_total_acks",
			Help: "Total number of ACKs sent",
		}),

		ackFreqDelayedACKs: promauto.NewCounter(prometheus.CounterOpts{
			Name: "cloudbridge_quic_ack_freq_delayed_acks_total",
			Help: "Total number of delayed ACKs",
		}),

		ackFreqAdaptiveACKs: promauto.NewCounter(prometheus.CounterOpts{
			Name: "cloudbridge_quic_ack_freq_adaptive_acks_total",
			Help: "Total number of adaptive ACKs",
		}),

		ackFreqAverageDelay: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "cloudbridge_quic_ack_freq_average_delay_seconds",
			Help:    "Average ACK delay in seconds",
			Buckets: []float64{0.001, 0.002, 0.005, 0.01, 0.025, 0.05, 0.1},
		}),

		ackFreqBandwidthSaved: promauto.NewCounter(prometheus.CounterOpts{
			Name: "cloudbridge_quic_ack_freq_bandwidth_saved_bytes",
			Help: "Total bandwidth saved by ACK frequency optimization (bytes)",
		}),

		ackFreqPacketsReceived: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "cloudbridge_quic_ack_freq_packets_received_total",
			Help: "Total packets received per connection",
		}, []string{"connection_id", "tenant_id"}),

		ackFreqBytesReceived: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "cloudbridge_quic_ack_freq_bytes_received_total",
			Help: "Total bytes received per connection",
		}, []string{"connection_id", "tenant_id"}),

		// FEC Metrics
		fecTotalGroups: promauto.NewCounter(prometheus.CounterOpts{
			Name: "cloudbridge_quic_fec_total_groups",
			Help: "Total number of FEC groups created",
		}),

		fecActiveGroups: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "cloudbridge_quic_fec_active_groups",
			Help: "Number of active FEC groups",
		}),

		fecRedundancyBytes: promauto.NewCounter(prometheus.CounterOpts{
			Name: "cloudbridge_quic_fec_redundancy_bytes_total",
			Help: "Total redundancy bytes added by FEC",
		}),

		fecRecoveryEvents: promauto.NewCounter(prometheus.CounterOpts{
			Name: "cloudbridge_quic_fec_recovery_events_total",
			Help: "Total number of successful packet recoveries",
		}),

		fecFailedRecoveries: promauto.NewCounter(prometheus.CounterOpts{
			Name: "cloudbridge_quic_fec_failed_recoveries_total",
			Help: "Total number of failed recovery attempts",
		}),

		fecRecoveryRate: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "cloudbridge_quic_fec_recovery_rate",
			Help: "FEC recovery rate (0.0 to 1.0)",
		}),

		fecEfficiency: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "cloudbridge_quic_fec_efficiency",
			Help: "FEC efficiency (recovered_bytes / redundancy_bytes)",
		}),

		fecLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "cloudbridge_quic_fec_latency_seconds",
			Help:    "FEC encoding/decoding latency in seconds",
			Buckets: []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05},
		}),

		fecBytesSaved: promauto.NewCounter(prometheus.CounterOpts{
			Name: "cloudbridge_quic_fec_bytes_saved_total",
			Help: "Total bytes saved by avoiding retransmission",
		}),

		fecDataPackets: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "cloudbridge_quic_fec_data_packets_total",
			Help: "Total data packets in FEC groups",
		}, []string{"algorithm", "tenant_id"}),

		fecRedundantPackets: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "cloudbridge_quic_fec_redundant_packets_total",
			Help: "Total redundant packets created by FEC",
		}, []string{"algorithm", "tenant_id"}),

		// Greasing Metrics
		greasingVersionCount: promauto.NewCounter(prometheus.CounterOpts{
			Name: "cloudbridge_quic_greasing_version_total",
			Help: "Total number of version greasing applications",
		}),

		greasingTransportParamCount: promauto.NewCounter(prometheus.CounterOpts{
			Name: "cloudbridge_quic_greasing_transport_param_total",
			Help: "Total number of transport parameter greasing applications",
		}),

		greasingFrameTypeCount: promauto.NewCounter(prometheus.CounterOpts{
			Name: "cloudbridge_quic_greasing_frame_type_total",
			Help: "Total number of frame type greasing applications",
		}),

		greasingTotalApplications: promauto.NewCounter(prometheus.CounterOpts{
			Name: "cloudbridge_quic_greasing_applications_total",
			Help: "Total number of greasing applications",
		}),

		greasingCompatibilityIssues: promauto.NewCounter(prometheus.CounterOpts{
			Name: "cloudbridge_quic_greasing_compatibility_issues_total",
			Help: "Total number of compatibility issues detected",
		}),

		greasingActiveConnections: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "cloudbridge_quic_greasing_active_connections",
			Help: "Number of active connections with greasing enabled",
		}),

		// Cross-cutting Metrics
		experimentalConnectionsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "cloudbridge_quic_experimental_connections_total",
			Help: "Total connections using experimental features",
		}, []string{"feature", "tenant_id"}),

		experimentalConnectionsActive: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "cloudbridge_quic_experimental_connections_active",
			Help: "Active connections using experimental features",
		}, []string{"feature", "tenant_id"}),

		experimentalLatencyImprovement: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "cloudbridge_quic_experimental_latency_improvement_percent",
			Help: "Latency improvement compared to baseline (%)",
		}, []string{"feature", "tenant_id"}),

		experimentalThroughputImprovement: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "cloudbridge_quic_experimental_throughput_improvement_percent",
			Help: "Throughput improvement compared to baseline (%)",
		}, []string{"feature", "tenant_id"}),

		experimentalPacketLossReduction: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "cloudbridge_quic_experimental_packet_loss_reduction_percent",
			Help: "Packet loss reduction compared to baseline (%)",
		}, []string{"feature", "tenant_id"}),
	}
}

// UpdateBBRv2Metrics updates BBRv2 congestion control metrics
func (em *ExperimentalMetrics) UpdateBBRv2Metrics(connectionID, tenantID string, metrics *congestion.CCMetrics) {
	if metrics == nil {
		return
	}

	// Map BBRState string to numeric value for Prometheus
	stateValue := 0.0
	switch metrics.BBRState {
	case "Startup":
		stateValue = 0.0
	case "Drain":
		stateValue = 1.0
	case "ProbeBW":
		stateValue = 2.0
	case "ProbeRTT":
		stateValue = 3.0
	}

	em.bbrv2State.WithLabelValues(connectionID, tenantID).Set(stateValue)
	em.bbrv2CongestionWindow.WithLabelValues(connectionID, tenantID).Set(float64(metrics.CongestionWindow))
	em.bbrv2PacingRate.WithLabelValues(connectionID, tenantID).Set(float64(metrics.PacingRate))
	em.bbrv2Bandwidth.WithLabelValues(connectionID, tenantID).Set(metrics.Bandwidth)
	em.bbrv2MinRTT.WithLabelValues(connectionID, tenantID).Set(metrics.MinRTT.Seconds())
	em.bbrv2Throughput.WithLabelValues(connectionID, tenantID).Set(metrics.Throughput)
	em.bbrv2LossRate.WithLabelValues(connectionID, tenantID).Set(metrics.LossRate)
	em.bbrv2PacketsInFlight.WithLabelValues(connectionID, tenantID).Set(float64(metrics.PacketsInFlight))
}

// RecordBBRv2StateTransition records a BBRv2 state transition
func (em *ExperimentalMetrics) RecordBBRv2StateTransition(fromState, toState, tenantID string) {
	em.bbrv2StateTransitions.WithLabelValues(fromState, toState, tenantID).Inc()
}

// UpdateACKFrequencyMetrics updates ACK frequency optimization metrics
func (em *ExperimentalMetrics) UpdateACKFrequencyMetrics(metrics *ACKFrequencyMetrics) {
	if metrics == nil {
		return
	}

	em.ackFreqTotalConnections.Set(float64(metrics.TotalConnections))
	em.ackFreqActiveConnections.Set(float64(metrics.ActiveConnections))
	em.ackFreqBandwidthSaved.Add(float64(metrics.BandwidthSaved))
}

// RecordACK records an ACK event
func (em *ExperimentalMetrics) RecordACK(delayed, adaptive bool, delay float64) {
	em.ackFreqTotalACKs.Inc()
	if delayed {
		em.ackFreqDelayedACKs.Inc()
	}
	if adaptive {
		em.ackFreqAdaptiveACKs.Inc()
	}
	em.ackFreqAverageDelay.Observe(delay)
}

// RecordACKFrequencyConnection records per-connection ACK metrics
func (em *ExperimentalMetrics) RecordACKFrequencyConnection(connectionID, tenantID string, packetsReceived, bytesReceived int64) {
	em.ackFreqPacketsReceived.WithLabelValues(connectionID, tenantID).Add(float64(packetsReceived))
	em.ackFreqBytesReceived.WithLabelValues(connectionID, tenantID).Add(float64(bytesReceived))
}

// UpdateFECMetrics updates FEC metrics
func (em *ExperimentalMetrics) UpdateFECMetrics(metrics *FECMetrics) {
	if metrics == nil {
		return
	}

	em.fecActiveGroups.Set(float64(metrics.ActiveGroups))
	em.fecRecoveryRate.Set(metrics.RecoveryRate)
	em.fecEfficiency.Set(metrics.Efficiency)
}

// RecordFECGroup records FEC group creation
func (em *ExperimentalMetrics) RecordFECGroup(algorithm, tenantID string, dataPackets, redundantPackets, redundancyBytes int) {
	em.fecTotalGroups.Inc()
	em.fecDataPackets.WithLabelValues(algorithm, tenantID).Add(float64(dataPackets))
	em.fecRedundantPackets.WithLabelValues(algorithm, tenantID).Add(float64(redundantPackets))
	em.fecRedundancyBytes.Add(float64(redundancyBytes))
}

// RecordFECRecovery records a FEC recovery event
func (em *ExperimentalMetrics) RecordFECRecovery(success bool, bytesSaved int64, latency float64) {
	if success {
		em.fecRecoveryEvents.Inc()
		em.fecBytesSaved.Add(float64(bytesSaved))
	} else {
		em.fecFailedRecoveries.Inc()
	}
	em.fecLatency.Observe(latency)
}

// UpdateGreasingMetrics updates greasing metrics
func (em *ExperimentalMetrics) UpdateGreasingMetrics(metrics *GreasingMetrics) {
	if metrics == nil {
		return
	}

	// Update total applications
	em.greasingTotalApplications.Add(float64(metrics.TotalGreaseApplications))
}

// RecordGreasingApplication records a greasing application
func (em *ExperimentalMetrics) RecordGreasingApplication(greasingType string) {
	switch greasingType {
	case "version":
		em.greasingVersionCount.Inc()
	case "transport_param":
		em.greasingTransportParamCount.Inc()
	case "frame_type":
		em.greasingFrameTypeCount.Inc()
	}
	em.greasingTotalApplications.Inc()
}

// RecordGreasingCompatibilityIssue records a compatibility issue
func (em *ExperimentalMetrics) RecordGreasingCompatibilityIssue() {
	em.greasingCompatibilityIssues.Inc()
}

// SetGreasingActiveConnections sets the number of active connections with greasing
func (em *ExperimentalMetrics) SetGreasingActiveConnections(count int) {
	em.greasingActiveConnections.Set(float64(count))
}

// RecordExperimentalConnection records a connection using experimental features
func (em *ExperimentalMetrics) RecordExperimentalConnection(feature, tenantID string, active bool) {
	em.experimentalConnectionsTotal.WithLabelValues(feature, tenantID).Inc()
	if active {
		em.experimentalConnectionsActive.WithLabelValues(feature, tenantID).Inc()
	} else {
		em.experimentalConnectionsActive.WithLabelValues(feature, tenantID).Dec()
	}
}

// SetLatencyImprovement sets the latency improvement for a feature
func (em *ExperimentalMetrics) SetLatencyImprovement(feature, tenantID string, improvement float64) {
	em.experimentalLatencyImprovement.WithLabelValues(feature, tenantID).Set(improvement)
}

// SetThroughputImprovement sets the throughput improvement for a feature
func (em *ExperimentalMetrics) SetThroughputImprovement(feature, tenantID string, improvement float64) {
	em.experimentalThroughputImprovement.WithLabelValues(feature, tenantID).Set(improvement)
}

// SetPacketLossReduction sets the packet loss reduction for a feature
func (em *ExperimentalMetrics) SetPacketLossReduction(feature, tenantID string, reduction float64) {
	em.experimentalPacketLossReduction.WithLabelValues(feature, tenantID).Set(reduction)
}

// GetMetricsSnapshot returns a snapshot of all experimental metrics
func (em *ExperimentalMetrics) GetMetricsSnapshot() map[string]interface{} {
	return map[string]interface{}{
		"bbrv2": map[string]string{
			"description": "BBRv2 congestion control metrics",
			"metrics": "cloudbridge_quic_bbrv2_*",
		},
		"ack_frequency": map[string]string{
			"description": "ACK Frequency optimization metrics",
			"metrics": "cloudbridge_quic_ack_freq_*",
		},
		"fec": map[string]string{
			"description": "Forward Error Correction metrics",
			"metrics": "cloudbridge_quic_fec_*",
		},
		"greasing": map[string]string{
			"description": "QUIC Bit Greasing (RFC 9287) metrics",
			"metrics": "cloudbridge_quic_greasing_*",
		},
		"cross_cutting": map[string]string{
			"description": "Cross-cutting experimental feature metrics",
			"metrics": "cloudbridge_quic_experimental_*",
		},
	}
}
