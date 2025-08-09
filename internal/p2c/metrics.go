package p2c

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// P2CMetrics представляет метрики P2C соединений
type P2CMetrics struct {
	connections *prometheus.CounterVec
	latency     *prometheus.HistogramVec
	throughput  *prometheus.CounterVec
	serverLoad  *prometheus.GaugeVec
}

// NewP2CMetrics создает новые P2C метрики
func NewP2CMetrics() *P2CMetrics {
	return &P2CMetrics{
		connections: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "p2c_connections_total",
				Help: "Total number of P2C connections",
			},
			[]string{"server", "status", "tenant_id"},
		),
		latency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "p2c_latency_seconds",
				Help:    "P2C connection latency",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"server", "tenant_id"},
		),
		throughput: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "p2c_throughput_bytes",
				Help: "P2C data throughput",
			},
			[]string{"server", "tenant_id"},
		),
		serverLoad: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "p2c_server_load",
				Help: "P2C server load",
			},
			[]string{"server"},
		),
	}
}

// RecordConnection записывает метрику соединения
func (pc *P2CMetrics) RecordConnection(server, status, tenantID string, duration time.Duration) {
	pc.connections.WithLabelValues(server, status, tenantID).Inc()
	pc.latency.WithLabelValues(server, tenantID).Observe(duration.Seconds())
}

// RecordThroughput записывает метрику пропускной способности
func (pc *P2CMetrics) RecordThroughput(server, tenantID string, bytes int64) {
	pc.throughput.WithLabelValues(server, tenantID).Add(float64(bytes))
}

// SetServerLoad устанавливает нагрузку сервера
func (pc *P2CMetrics) SetServerLoad(server string, load float64) {
	pc.serverLoad.WithLabelValues(server).Set(load)
}

// Register регистрирует метрики в Prometheus
func (pc *P2CMetrics) Register(registry prometheus.Registerer) error {
	collectors := []prometheus.Collector{
		pc.connections,
		pc.latency,
		pc.throughput,
		pc.serverLoad,
	}

	for _, collector := range collectors {
		if err := registry.Register(collector); err != nil {
			return err
		}
	}

	return nil
}
