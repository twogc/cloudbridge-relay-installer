package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// KubeProxyMetrics holds all kube proxy related metrics
type KubeProxyMetrics struct {
	RequestsTotal       *prometheus.CounterVec
	RequestDuration     *prometheus.HistogramVec
	WSHandshakesTotal   *prometheus.CounterVec
	WSUpgradesTotal     *prometheus.CounterVec
	WSActiveStreams     prometheus.Gauge
	WSErrorsTotal       *prometheus.CounterVec
	UpgradeFallbacks    *prometheus.CounterVec
	UpstreamErrorsTotal *prometheus.CounterVec
}

// NewKubeProxyMetrics creates and registers all kube proxy metrics
func NewKubeProxyMetrics() *KubeProxyMetrics {
	return &KubeProxyMetrics{
		RequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kube_proxy_requests_total",
				Help: "Total number of kube proxy requests",
			},
			[]string{"cluster_id", "verb", "resource", "code"},
		),
		RequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "kube_proxy_request_duration_seconds",
				Help:    "Duration of kube proxy requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"cluster_id", "verb", "resource"},
		),
		WSHandshakesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kube_ws_handshakes_total",
				Help: "Total number of WebSocket handshakes",
			},
			[]string{"cluster_id", "code", "subprotocol"},
		),
		WSUpgradesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kube_ws_upgrades_total",
				Help: "Total number of WebSocket protocol upgrades",
			},
			[]string{"cluster_id", "subprotocol"},
		),
		WSActiveStreams: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "kube_ws_active_streams",
				Help: "Number of active WebSocket streams",
			},
		),
		WSErrorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kube_ws_errors_total",
				Help: "Total number of WebSocket errors",
			},
			[]string{"cluster_id", "reason"},
		),
		UpgradeFallbacks: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kube_upgrade_fallbacks_total",
				Help: "Total number of protocol upgrade fallbacks",
			},
			[]string{"cluster_id", "from", "to"},
		),
		UpstreamErrorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kube_upstream_errors_total",
				Help: "Total number of upstream errors",
			},
			[]string{"cluster_id", "code"},
		),
	}
}

// RecordRequest records a completed request with timing
func (m *KubeProxyMetrics) RecordRequest(clusterID, verb, resource, code string, duration time.Duration) {
	m.RequestsTotal.WithLabelValues(clusterID, verb, resource, code).Inc()
	m.RequestDuration.WithLabelValues(clusterID, verb, resource).Observe(duration.Seconds())
}

// RecordWSHandshake records a WebSocket handshake attempt
func (m *KubeProxyMetrics) RecordWSHandshake(clusterID, code, subprotocol string) {
	m.WSHandshakesTotal.WithLabelValues(clusterID, code, subprotocol).Inc()
}

// RecordWSUpgrade records a successful WebSocket upgrade
func (m *KubeProxyMetrics) RecordWSUpgrade(clusterID, subprotocol string) {
	m.WSUpgradesTotal.WithLabelValues(clusterID, subprotocol).Inc()
}

// IncrementWSStreams increments active WebSocket streams counter
func (m *KubeProxyMetrics) IncrementWSStreams() {
	m.WSActiveStreams.Inc()
}

// DecrementWSStreams decrements active WebSocket streams counter
func (m *KubeProxyMetrics) DecrementWSStreams() {
	m.WSActiveStreams.Dec()
}

// RecordWSError records a WebSocket error
func (m *KubeProxyMetrics) RecordWSError(clusterID, reason string) {
	m.WSErrorsTotal.WithLabelValues(clusterID, reason).Inc()
}

// RecordUpgradeFallback records a protocol upgrade fallback
func (m *KubeProxyMetrics) RecordUpgradeFallback(clusterID, from, to string) {
	m.UpgradeFallbacks.WithLabelValues(clusterID, from, to).Inc()
}

// RecordUpstreamError records an upstream error
func (m *KubeProxyMetrics) RecordUpstreamError(clusterID, code string) {
	m.UpstreamErrorsTotal.WithLabelValues(clusterID, code).Inc()
}

