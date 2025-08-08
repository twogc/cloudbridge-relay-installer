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
				Name: "p2p_latency_seconds",
				Help: "P2P connection latency",
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

// Register регистрирует метрики в Prometheus
func (pm *P2PMetrics) Register(registry prometheus.Registerer) error {
	collectors := []prometheus.Collector{
		pm.connections,
		pm.latency,
		pm.throughput,
		pm.peerCount,
	}

	for _, collector := range collectors {
		if err := registry.Register(collector); err != nil {
			return err
		}
	}

	return nil
}
