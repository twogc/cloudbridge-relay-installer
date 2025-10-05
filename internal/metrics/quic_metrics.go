package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// QUICMetrics представляет метрики для QUIC протокола
type QUICMetrics struct {
	// Соединения
	connectionsTotal   prometheus.Counter
	connectionsActive  prometheus.Gauge
	connectionsFailed  *prometheus.CounterVec
	connectionsDropped *prometheus.CounterVec

	// Потоки (Streams)
	streamsTotal   prometheus.Counter
	streamsActive  prometheus.Gauge
	streamsCreated *prometheus.CounterVec
	streamsClosed  *prometheus.CounterVec
	streamsReset   *prometheus.CounterVec

	// Производительность
	handshakeDuration prometheus.Histogram
	handshakeFailures *prometheus.CounterVec
	connectionLatency *prometheus.HistogramVec
	throughputBytes   *prometheus.CounterVec
	throughputPackets *prometheus.CounterVec
	packetLossRate    *prometheus.GaugeVec
	packetLossEvents  *prometheus.CounterVec

	// Качество соединения
	rttMilliseconds  *prometheus.HistogramVec
	rttVariance      *prometheus.GaugeVec
	congestionWindow *prometheus.GaugeVec
	bytesInFlight    *prometheus.GaugeVec
	packetsInFlight  *prometheus.GaugeVec

	// Ошибки и сбои
	packetErrors     *prometheus.CounterVec
	connectionErrors *prometheus.CounterVec
	streamErrors     *prometheus.CounterVec
	timeoutEvents    *prometheus.CounterVec
	retransmissions  *prometheus.CounterVec

	// Безопасность
	encryptionFailures     *prometheus.CounterVec
	authenticationFailures *prometheus.CounterVec
	keyRotationEvents      *prometheus.CounterVec

	// Конфигурация
	versionUsage         *prometheus.GaugeVec
	parameterNegotiation *prometheus.CounterVec
	extensionUsage       *prometheus.CounterVec

	// Мониторинг ресурсов
	memoryUsageBytes *prometheus.GaugeVec
	cpuUsagePercent  *prometheus.GaugeVec
	goroutineCount   *prometheus.GaugeVec
}

// NewQUICMetrics создает новые QUIC метрики
func NewQUICMetrics(reg prometheus.Registerer) *QUICMetrics {
	m := &QUICMetrics{
		// Соединения
		connectionsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "relay_quic_connections_total",
			Help: "Total number of QUIC connections",
		}),
		connectionsActive: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "relay_quic_connections_active",
			Help: "Number of active QUIC connections",
		}),
		connectionsFailed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_quic_connections_failed_total",
			Help: "Total number of failed QUIC connections",
		}, []string{"reason", "version"}),
		connectionsDropped: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_quic_connections_dropped_total",
			Help: "Total number of dropped QUIC connections",
		}, []string{"reason"}),

		// Потоки
		streamsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "relay_quic_streams_total",
			Help: "Total number of QUIC streams",
		}),
		streamsActive: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "relay_quic_streams_active",
			Help: "Number of active QUIC streams",
		}),
		streamsCreated: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_quic_streams_created_total",
			Help: "Total number of created QUIC streams",
		}, []string{"stream_type", "direction"}),
		streamsClosed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_quic_streams_closed_total",
			Help: "Total number of closed QUIC streams",
		}, []string{"stream_type", "reason"}),
		streamsReset: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_quic_streams_reset_total",
			Help: "Total number of reset QUIC streams",
		}, []string{"stream_type", "reason"}),

		// Производительность
		handshakeDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "relay_quic_handshake_duration_seconds",
			Help:    "QUIC handshake duration in seconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.2, 0.5, 1.0, 2.0, 5.0},
		}),
		handshakeFailures: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_quic_handshake_failures_total",
			Help: "Total number of QUIC handshake failures",
		}, []string{"reason", "version"}),
		connectionLatency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "relay_quic_connection_latency_seconds",
			Help:    "QUIC connection latency in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.2, 0.5, 1.0},
		}, []string{"connection_id"}),
		throughputBytes: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_quic_throughput_bytes_total",
			Help: "Total QUIC throughput in bytes",
		}, []string{"connection_id", "direction"}),
		throughputPackets: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_quic_throughput_packets_total",
			Help: "Total QUIC throughput in packets",
		}, []string{"connection_id", "direction"}),
		packetLossRate: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_quic_packet_loss_rate",
			Help: "QUIC packet loss rate (0-1)",
		}, []string{"connection_id"}),
		packetLossEvents: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_quic_packet_loss_events_total",
			Help: "Total number of QUIC packet loss events",
		}, []string{"connection_id", "reason"}),

		// Качество соединения
		rttMilliseconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "relay_quic_rtt_milliseconds",
			Help:    "QUIC round-trip time in milliseconds",
			Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000},
		}, []string{"connection_id"}),
		rttVariance: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_quic_rtt_variance_milliseconds",
			Help: "QUIC RTT variance in milliseconds",
		}, []string{"connection_id"}),
		congestionWindow: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_quic_congestion_window_bytes",
			Help: "QUIC congestion window size in bytes",
		}, []string{"connection_id"}),
		bytesInFlight: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_quic_bytes_in_flight",
			Help: "QUIC bytes in flight",
		}, []string{"connection_id"}),
		packetsInFlight: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_quic_packets_in_flight",
			Help: "QUIC packets in flight",
		}, []string{"connection_id"}),

		// Ошибки и сбои
		packetErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_quic_packet_errors_total",
			Help: "Total number of QUIC packet errors",
		}, []string{"error_type", "connection_id"}),
		connectionErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_quic_connection_errors_total",
			Help: "Total number of QUIC connection errors",
		}, []string{"error_type", "connection_id"}),
		streamErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_quic_stream_errors_total",
			Help: "Total number of QUIC stream errors",
		}, []string{"error_type", "stream_id"}),
		timeoutEvents: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_quic_timeout_events_total",
			Help: "Total number of QUIC timeout events",
		}, []string{"timeout_type", "connection_id"}),
		retransmissions: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_quic_retransmissions_total",
			Help: "Total number of QUIC retransmissions",
		}, []string{"connection_id", "reason"}),

		// Безопасность
		encryptionFailures: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_quic_encryption_failures_total",
			Help: "Total number of QUIC encryption failures",
		}, []string{"failure_type", "connection_id"}),
		authenticationFailures: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_quic_authentication_failures_total",
			Help: "Total number of QUIC authentication failures",
		}, []string{"failure_type", "connection_id"}),
		keyRotationEvents: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_quic_key_rotation_events_total",
			Help: "Total number of QUIC key rotation events",
		}, []string{"connection_id", "key_type"}),

		// Конфигурация
		versionUsage: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_quic_version_usage",
			Help: "QUIC version usage count",
		}, []string{"version"}),
		parameterNegotiation: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_quic_parameter_negotiation_total",
			Help: "Total number of QUIC parameter negotiations",
		}, []string{"parameter", "result"}),
		extensionUsage: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_quic_extension_usage_total",
			Help: "Total number of QUIC extension usage",
		}, []string{"extension", "connection_id"}),

		// Мониторинг ресурсов
		memoryUsageBytes: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_quic_memory_usage_bytes",
			Help: "QUIC memory usage in bytes",
		}, []string{"connection_id", "component"}),
		cpuUsagePercent: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_quic_cpu_usage_percent",
			Help: "QUIC CPU usage percentage",
		}, []string{"connection_id", "component"}),
		goroutineCount: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_quic_goroutine_count",
			Help: "QUIC goroutine count",
		}, []string{"connection_id", "component"}),
	}

	// Регистрация метрик
	reg.MustRegister(
		m.connectionsTotal,
		m.connectionsActive,
		m.connectionsFailed,
		m.connectionsDropped,
		m.streamsTotal,
		m.streamsActive,
		m.streamsCreated,
		m.streamsClosed,
		m.streamsReset,
		m.handshakeDuration,
		m.handshakeFailures,
		m.connectionLatency,
		m.throughputBytes,
		m.throughputPackets,
		m.packetLossRate,
		m.packetLossEvents,
		m.rttMilliseconds,
		m.rttVariance,
		m.congestionWindow,
		m.bytesInFlight,
		m.packetsInFlight,
		m.packetErrors,
		m.connectionErrors,
		m.streamErrors,
		m.timeoutEvents,
		m.retransmissions,
		m.encryptionFailures,
		m.authenticationFailures,
		m.keyRotationEvents,
		m.versionUsage,
		m.parameterNegotiation,
		m.extensionUsage,
		m.memoryUsageBytes,
		m.cpuUsagePercent,
		m.goroutineCount,
	)

	return m
}

// Методы для работы с QUIC метриками

// Connection Metrics
func (m *QUICMetrics) IncConnections() {
	m.connectionsTotal.Inc()
	m.connectionsActive.Inc()
}

func (m *QUICMetrics) DecConnections() {
	m.connectionsActive.Dec()
}

func (m *QUICMetrics) IncConnectionsFailed(reason, version string) {
	m.connectionsFailed.WithLabelValues(reason, version).Inc()
}

func (m *QUICMetrics) IncConnectionsDropped(reason string) {
	m.connectionsDropped.WithLabelValues(reason).Inc()
}

// Stream Metrics
func (m *QUICMetrics) IncStreams() {
	m.streamsTotal.Inc()
	m.streamsActive.Inc()
}

func (m *QUICMetrics) DecStreams() {
	m.streamsActive.Dec()
}

func (m *QUICMetrics) IncStreamsCreated(streamType, direction string) {
	m.streamsCreated.WithLabelValues(streamType, direction).Inc()
}

func (m *QUICMetrics) IncStreamsClosed(streamType, reason string) {
	m.streamsClosed.WithLabelValues(streamType, reason).Inc()
}

func (m *QUICMetrics) IncStreamsReset(streamType, reason string) {
	m.streamsReset.WithLabelValues(streamType, reason).Inc()
}

// Performance Metrics
func (m *QUICMetrics) ObserveHandshakeDuration(duration float64) {
	m.handshakeDuration.Observe(duration)
}

func (m *QUICMetrics) IncHandshakeFailures(reason, version string) {
	m.handshakeFailures.WithLabelValues(reason, version).Inc()
}

func (m *QUICMetrics) ObserveConnectionLatency(connectionID string, duration float64) {
	m.connectionLatency.WithLabelValues(connectionID).Observe(duration)
}

func (m *QUICMetrics) IncThroughputBytes(connectionID, direction string, bytes int64) {
	m.throughputBytes.WithLabelValues(connectionID, direction).Add(float64(bytes))
}

func (m *QUICMetrics) IncThroughputPackets(connectionID, direction string, packets int64) {
	m.throughputPackets.WithLabelValues(connectionID, direction).Add(float64(packets))
}

func (m *QUICMetrics) SetPacketLossRate(connectionID string, rate float64) {
	m.packetLossRate.WithLabelValues(connectionID).Set(rate)
}

func (m *QUICMetrics) IncPacketLossEvents(connectionID, reason string) {
	m.packetLossEvents.WithLabelValues(connectionID, reason).Inc()
}

// Connection Quality Metrics
func (m *QUICMetrics) ObserveRTT(connectionID string, rtt float64) {
	m.rttMilliseconds.WithLabelValues(connectionID).Observe(rtt)
}

func (m *QUICMetrics) SetRTTVariance(connectionID string, variance float64) {
	m.rttVariance.WithLabelValues(connectionID).Set(variance)
}

func (m *QUICMetrics) SetCongestionWindow(connectionID string, windowSize int64) {
	m.congestionWindow.WithLabelValues(connectionID).Set(float64(windowSize))
}

func (m *QUICMetrics) SetBytesInFlight(connectionID string, bytes int64) {
	m.bytesInFlight.WithLabelValues(connectionID).Set(float64(bytes))
}

func (m *QUICMetrics) SetPacketsInFlight(connectionID string, packets int64) {
	m.packetsInFlight.WithLabelValues(connectionID).Set(float64(packets))
}

// Error Metrics
func (m *QUICMetrics) IncPacketErrors(errorType, connectionID string) {
	m.packetErrors.WithLabelValues(errorType, connectionID).Inc()
}

func (m *QUICMetrics) IncConnectionErrors(errorType, connectionID string) {
	m.connectionErrors.WithLabelValues(errorType, connectionID).Inc()
}

func (m *QUICMetrics) IncStreamErrors(errorType, streamID string) {
	m.streamErrors.WithLabelValues(errorType, streamID).Inc()
}

func (m *QUICMetrics) IncTimeoutEvents(timeoutType, connectionID string) {
	m.timeoutEvents.WithLabelValues(timeoutType, connectionID).Inc()
}

func (m *QUICMetrics) IncRetransmissions(connectionID, reason string) {
	m.retransmissions.WithLabelValues(connectionID, reason).Inc()
}

// Security Metrics
func (m *QUICMetrics) IncEncryptionFailures(failureType, connectionID string) {
	m.encryptionFailures.WithLabelValues(failureType, connectionID).Inc()
}

func (m *QUICMetrics) IncAuthenticationFailures(failureType, connectionID string) {
	m.authenticationFailures.WithLabelValues(failureType, connectionID).Inc()
}

func (m *QUICMetrics) IncKeyRotationEvents(connectionID, keyType string) {
	m.keyRotationEvents.WithLabelValues(connectionID, keyType).Inc()
}

// Configuration Metrics
func (m *QUICMetrics) SetVersionUsage(version string, count int) {
	m.versionUsage.WithLabelValues(version).Set(float64(count))
}

func (m *QUICMetrics) IncParameterNegotiation(parameter, result string) {
	m.parameterNegotiation.WithLabelValues(parameter, result).Inc()
}

func (m *QUICMetrics) IncExtensionUsage(extension, connectionID string) {
	m.extensionUsage.WithLabelValues(extension, connectionID).Inc()
}

// Resource Monitoring Metrics
func (m *QUICMetrics) SetMemoryUsage(connectionID, component string, bytes int64) {
	m.memoryUsageBytes.WithLabelValues(connectionID, component).Set(float64(bytes))
}

func (m *QUICMetrics) SetCPUUsage(connectionID, component string, percent float64) {
	m.cpuUsagePercent.WithLabelValues(connectionID, component).Set(percent)
}

func (m *QUICMetrics) SetGoroutineCount(connectionID, component string, count int) {
	m.goroutineCount.WithLabelValues(connectionID, component).Set(float64(count))
}
