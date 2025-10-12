package nat

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// NATPrometheusMetrics содержит Prometheus метрики для NAT traversal
type NATPrometheusMetrics struct {
	// STUN метрики
	stunRequestsTotal *prometheus.CounterVec
	stunSuccessTotal  *prometheus.CounterVec
	stunFailuresTotal *prometheus.CounterVec
	stunResponseTime  prometheus.Histogram

	// ICE кандидаты
	iceCandidatesGathered *prometheus.CounterVec
	iceGatheringDuration  prometheus.Histogram

	// NAT типы
	natTypeCount *prometheus.CounterVec

	// ICE connectivity checks
	connectivityChecksTotal *prometheus.CounterVec
	connectivitySuccessRate prometheus.Gauge
	connectionTime          prometheus.Histogram

	// TURN метрики
	turnAllocationsTotal        *prometheus.CounterVec
	turnAllocationFailuresTotal prometheus.Counter
	turnActiveallocations       prometheus.Gauge
	turnDataTransferred         *prometheus.CounterVec
	turnResponseTime            prometheus.Histogram

	// Bandwidth метрики
	bandwidthBytes   *prometheus.CounterVec
	packetsTotal     *prometheus.CounterVec
	packetsLostTotal prometheus.Counter
	packetLossRate   prometheus.Gauge

	// Performance метрики
	rtt    prometheus.Gauge
	jitter prometheus.Gauge
}

var (
	prometheusMetricsInstance *NATPrometheusMetrics
	prometheusMetricsOnce     sync.Once
)

// NewNATPrometheusMetrics создает новые Prometheus метрики для NAT (синглтон)
func NewNATPrometheusMetrics() *NATPrometheusMetrics {
	prometheusMetricsOnce.Do(func() {
		prometheusMetricsInstance = &NATPrometheusMetrics{
		// STUN метрики
		stunRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "cloudbridge",
				Subsystem: "nat",
				Name:      "stun_requests_total",
				Help:      "Total number of STUN requests",
			},
			[]string{"server"},
		),
		stunSuccessTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "cloudbridge",
				Subsystem: "nat",
				Name:      "stun_success_total",
				Help:      "Total number of successful STUN requests",
			},
			[]string{"server"},
		),
		stunFailuresTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "cloudbridge",
				Subsystem: "nat",
				Name:      "stun_failures_total",
				Help:      "Total number of failed STUN requests",
			},
			[]string{"server", "error_type"},
		),
		stunResponseTime: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: "cloudbridge",
				Subsystem: "nat",
				Name:      "stun_response_time_seconds",
				Help:      "STUN response time in seconds",
				Buckets:   prometheus.ExponentialBuckets(0.001, 2, 10), // 1ms to ~1s
			},
		),

		// ICE кандидаты
		iceCandidatesGathered: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "cloudbridge",
				Subsystem: "nat",
				Name:      "ice_candidates_gathered_total",
				Help:      "Total number of ICE candidates gathered",
			},
			[]string{"type"}, // host, srflx, relay, prflx
		),
		iceGatheringDuration: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: "cloudbridge",
				Subsystem: "nat",
				Name:      "ice_gathering_duration_seconds",
				Help:      "ICE candidate gathering duration in seconds",
				Buckets:   prometheus.ExponentialBuckets(0.1, 2, 10), // 100ms to ~100s
			},
		),

		// NAT типы
		natTypeCount: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "cloudbridge",
				Subsystem: "nat",
				Name:      "nat_type_count",
				Help:      "Count of detected NAT types",
			},
			[]string{"nat_type"}, // Open, Full Cone, Restricted Cone, Port Restricted, Symmetric, Blocked
		),

		// ICE connectivity checks
		connectivityChecksTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "cloudbridge",
				Subsystem: "nat",
				Name:      "connectivity_checks_total",
				Help:      "Total number of connectivity checks",
			},
			[]string{"result"}, // success, failure
		),
		connectivitySuccessRate: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "cloudbridge",
				Subsystem: "nat",
				Name:      "connectivity_success_rate",
				Help:      "Connectivity check success rate (0-100)",
			},
		),
		connectionTime: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: "cloudbridge",
				Subsystem: "nat",
				Name:      "connection_time_seconds",
				Help:      "Connection establishment time in seconds",
				Buckets:   prometheus.ExponentialBuckets(0.01, 2, 12), // 10ms to ~40s
			},
		),

		// TURN метрики
		turnAllocationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "cloudbridge",
				Subsystem: "nat",
				Name:      "turn_allocations_total",
				Help:      "Total number of TURN allocations",
			},
			[]string{"server", "result"}, // success, failure
		),
		turnAllocationFailuresTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: "cloudbridge",
				Subsystem: "nat",
				Name:      "turn_allocation_failures_total",
				Help:      "Total number of TURN allocation failures",
			},
		),
		turnActiveallocations: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "cloudbridge",
				Subsystem: "nat",
				Name:      "turn_active_allocations",
				Help:      "Number of active TURN allocations",
			},
		),
		turnDataTransferred: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "cloudbridge",
				Subsystem: "nat",
				Name:      "turn_data_transferred_bytes",
				Help:      "Total data transferred through TURN in bytes",
			},
			[]string{"direction"}, // sent, received
		),
		turnResponseTime: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: "cloudbridge",
				Subsystem: "nat",
				Name:      "turn_response_time_seconds",
				Help:      "TURN response time in seconds",
				Buckets:   prometheus.ExponentialBuckets(0.001, 2, 10), // 1ms to ~1s
			},
		),

		// Bandwidth метрики
		bandwidthBytes: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "cloudbridge",
				Subsystem: "nat",
				Name:      "bandwidth_bytes_total",
				Help:      "Total bytes transferred",
			},
			[]string{"direction"}, // sent, received
		),
		packetsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "cloudbridge",
				Subsystem: "nat",
				Name:      "packets_total",
				Help:      "Total packets",
			},
			[]string{"direction"}, // sent, received
		),
		packetsLostTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: "cloudbridge",
				Subsystem: "nat",
				Name:      "packets_lost_total",
				Help:      "Total packets lost",
			},
		),
		packetLossRate: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "cloudbridge",
				Subsystem: "nat",
				Name:      "packet_loss_rate",
				Help:      "Packet loss rate (0-100)",
			},
		),

		// Performance метрики
		rtt: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "cloudbridge",
				Subsystem: "nat",
				Name:      "rtt_milliseconds",
				Help:      "Round-trip time in milliseconds",
			},
		),
		jitter: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "cloudbridge",
				Subsystem: "nat",
				Name:      "jitter_milliseconds",
				Help:      "Jitter in milliseconds",
			},
		),
		}
	})
	return prometheusMetricsInstance
}

// UpdateFromNATMetrics обновляет Prometheus метрики из NATMetrics
func (npm *NATPrometheusMetrics) UpdateFromNATMetrics(metrics *NATMetrics) {
	metrics.mu.RLock()
	defer metrics.mu.RUnlock()

	// ICE кандидаты
	npm.iceCandidatesGathered.WithLabelValues("host").Add(float64(metrics.HostCandidatesGathered))
	npm.iceCandidatesGathered.WithLabelValues("srflx").Add(float64(metrics.ServerReflexiveCandidatesGathered))
	npm.iceCandidatesGathered.WithLabelValues("relay").Add(float64(metrics.RelayCandidatesGathered))

	// Gathering duration
	if metrics.GatheringDuration > 0 {
		npm.iceGatheringDuration.Observe(metrics.GatheringDuration.Seconds())
	}

	// Connectivity success rate
	npm.connectivitySuccessRate.Set(metrics.GetSuccessRate())

	// TURN allocations
	npm.turnActiveallocations.Set(float64(metrics.ActiveTURNAllocations))

	// Bandwidth
	npm.bandwidthBytes.WithLabelValues("sent").Add(float64(metrics.TotalBytesSent))
	npm.bandwidthBytes.WithLabelValues("received").Add(float64(metrics.TotalBytesReceived))

	npm.packetsTotal.WithLabelValues("sent").Add(float64(metrics.PacketsSent))
	npm.packetsTotal.WithLabelValues("received").Add(float64(metrics.PacketsReceived))

	npm.packetsLostTotal.Add(float64(metrics.PacketsLost))
	npm.packetLossRate.Set(metrics.PacketLossRate)

	// Performance
	npm.rtt.Set(float64(metrics.RTT.Milliseconds()))
	npm.jitter.Set(float64(metrics.Jitter.Milliseconds()))
}

// RecordSTUNRequest записывает STUN запрос
func (npm *NATPrometheusMetrics) RecordSTUNRequest(server string, success bool, responseTime float64) {
	npm.stunRequestsTotal.WithLabelValues(server).Inc()
	if success {
		npm.stunSuccessTotal.WithLabelValues(server).Inc()
		npm.stunResponseTime.Observe(responseTime)
	} else {
		npm.stunFailuresTotal.WithLabelValues(server, "timeout").Inc()
	}
}

// RecordICECandidate записывает собранный ICE кандидат
func (npm *NATPrometheusMetrics) RecordICECandidate(candidateType string) {
	npm.iceCandidatesGathered.WithLabelValues(candidateType).Inc()
}

// RecordNATType записывает определенный тип NAT
func (npm *NATPrometheusMetrics) RecordNATType(natType string) {
	npm.natTypeCount.WithLabelValues(natType).Inc()
}

// RecordConnectivityCheck записывает результат connectivity check
func (npm *NATPrometheusMetrics) RecordConnectivityCheck(success bool, connectionTime float64) {
	if success {
		npm.connectivityChecksTotal.WithLabelValues("success").Inc()
		npm.connectionTime.Observe(connectionTime)
	} else {
		npm.connectivityChecksTotal.WithLabelValues("failure").Inc()
	}
}

// RecordTURNAllocation записывает TURN allocation
func (npm *NATPrometheusMetrics) RecordTURNAllocation(server string, success bool, responseTime float64) {
	if success {
		npm.turnAllocationsTotal.WithLabelValues(server, "success").Inc()
		npm.turnResponseTime.Observe(responseTime)
		npm.turnActiveallocations.Inc()
	} else {
		npm.turnAllocationsTotal.WithLabelValues(server, "failure").Inc()
		npm.turnAllocationFailuresTotal.Inc()
	}
}

// RecordTURNDeallocation записывает освобождение TURN allocation
func (npm *NATPrometheusMetrics) RecordTURNDeallocation() {
	npm.turnActiveallocations.Dec()
}

// RecordTURNData записывает данные, переданные через TURN
func (npm *NATPrometheusMetrics) RecordTURNData(sent, received int64) {
	if sent > 0 {
		npm.turnDataTransferred.WithLabelValues("sent").Add(float64(sent))
	}
	if received > 0 {
		npm.turnDataTransferred.WithLabelValues("received").Add(float64(received))
	}
}
