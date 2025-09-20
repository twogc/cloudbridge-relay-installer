package p2p

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// P2PMetrics представляет метрики P2P соединений
type P2PMetrics struct {
	connections *prometheus.CounterVec
	latency     *prometheus.HistogramVec
	throughput  *prometheus.CounterVec
	peerCount   prometheus.Gauge

	// Дополнительные метрики для реальных данных
	activeConnections prometheus.Gauge
	totalConnections  prometheus.Counter
	packetsSent       prometheus.Counter
	packetsReceived   prometheus.Counter
	packetsLost       prometheus.Counter
	bytesTransferred  prometheus.Counter
	handshakeDuration prometheus.HistogramVec
	connectionErrors  *prometheus.CounterVec
}

// NewP2PMetrics создает новые P2P метрики
func NewP2PMetrics() *P2PMetrics {
	return &P2PMetrics{
		connections: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "p2p_connections_total",
				Help: "Total number of P2P connections",
			},
			[]string{"route_type", "status", "tenant_id"},
		),
		latency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "p2p_latency_seconds",
				Help:    "P2P connection latency",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"route_type", "tenant_id"},
		),
		throughput: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "p2p_throughput_bytes",
				Help: "P2P data throughput",
			},
			[]string{"route_type", "tenant_id"},
		),
		peerCount: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "p2p_active_peers",
				Help: "Number of active P2P peers",
			},
		),

		// Дополнительные метрики
		activeConnections: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "p2p_active_connections",
				Help: "Number of active P2P connections",
			},
		),
		totalConnections: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "p2p_total_connections",
				Help: "Total number of P2P connections established",
			},
		),
		packetsSent: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "p2p_packets_sent_total",
				Help: "Total number of packets sent",
			},
		),
		packetsReceived: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "p2p_packets_received_total",
				Help: "Total number of packets received",
			},
		),
		packetsLost: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "p2p_packets_lost_total",
				Help: "Total number of packets lost",
			},
		),
		bytesTransferred: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "p2p_bytes_transferred_total",
				Help: "Total bytes transferred through P2P connections",
			},
		),
		handshakeDuration: *prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "p2p_handshake_duration_seconds",
				Help:    "P2P handshake duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"protocol"},
		),
		connectionErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "p2p_connection_errors_total",
				Help: "Total number of P2P connection errors",
			},
			[]string{"error_type", "protocol"},
		),
	}
}

// RecordConnection записывает метрику соединения
func (pm *P2PMetrics) RecordConnection(routeType, status, tenantID string, duration time.Duration) {
	pm.connections.WithLabelValues(routeType, status, tenantID).Inc()
	pm.latency.WithLabelValues(routeType, tenantID).Observe(duration.Seconds())
}

// RecordThroughput записывает метрику пропускной способности
func (pm *P2PMetrics) RecordThroughput(routeType, tenantID string, bytes int64) {
	pm.throughput.WithLabelValues(routeType, tenantID).Add(float64(bytes))
}

// SetPeerCount устанавливает количество активных пиров
func (pm *P2PMetrics) SetPeerCount(count int) {
	pm.peerCount.Set(float64(count))
}

// SetActiveConnections устанавливает количество активных соединений
func (pm *P2PMetrics) SetActiveConnections(count int) {
	pm.activeConnections.Set(float64(count))
}

// IncrementTotalConnections увеличивает счетчик общих соединений
func (pm *P2PMetrics) IncrementTotalConnections() {
	pm.totalConnections.Inc()
}

// RecordPackets записывает метрики пакетов
func (pm *P2PMetrics) RecordPackets(sent, received, lost int64) {
	pm.packetsSent.Add(float64(sent))
	pm.packetsReceived.Add(float64(received))
	pm.packetsLost.Add(float64(lost))
}

// RecordBytesTransferred записывает переданные байты
func (pm *P2PMetrics) RecordBytesTransferred(bytes int64) {
	pm.bytesTransferred.Add(float64(bytes))
}

// RecordHandshakeDuration записывает время handshake
func (pm *P2PMetrics) RecordHandshakeDuration(protocol string, duration time.Duration) {
	pm.handshakeDuration.WithLabelValues(protocol).Observe(duration.Seconds())
}

// RecordConnectionError записывает ошибку соединения
func (pm *P2PMetrics) RecordConnectionError(errorType, protocol string) {
	pm.connectionErrors.WithLabelValues(errorType, protocol).Inc()
}

// Register регистрирует метрики в Prometheus
func (pm *P2PMetrics) Register(registry prometheus.Registerer) error {
	collectors := []prometheus.Collector{
		pm.connections,
		pm.latency,
		pm.throughput,
		pm.peerCount,
		pm.activeConnections,
		pm.totalConnections,
		pm.packetsSent,
		pm.packetsReceived,
		pm.packetsLost,
		pm.bytesTransferred,
		&pm.handshakeDuration,
		pm.connectionErrors,
	}

	for _, collector := range collectors {
		if err := registry.Register(collector); err != nil {
			return err
		}
	}

	return nil
}
