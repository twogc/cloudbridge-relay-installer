package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// P2PMetrics represents P2P-specific metrics
type P2PMetrics struct {
	connectionsTotal    prometheus.Counter
	activePeers         prometheus.Gauge
	trafficBytes        *prometheus.CounterVec
	latencySeconds      *prometheus.HistogramVec
	handshakeDuration   prometheus.Histogram
	meshRoutes          prometheus.Gauge
	peerDiscoveryEvents *prometheus.CounterVec
	fallbackEvents      *prometheus.CounterVec
	whitelistViolations *prometheus.CounterVec
}

// NewP2PMetrics creates new P2P metrics
func NewP2PMetrics(reg prometheus.Registerer) *P2PMetrics {
	m := &P2PMetrics{
		connectionsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "cloudbridge_p2p_connections_total",
			Help: "Total number of P2P connections",
		}),
		activePeers: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "cloudbridge_p2p_active_peers",
			Help: "Number of active P2P peers",
		}),
		trafficBytes: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "cloudbridge_p2p_traffic_bytes_total",
			Help: "Total P2P traffic in bytes",
		}, []string{"direction", "peer_id"}),
		latencySeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "cloudbridge_p2p_latency_seconds",
			Help:    "P2P connection latency in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0},
		}, []string{"peer_id"}),
		handshakeDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "cloudbridge_p2p_handshake_duration_seconds",
			Help:    "P2P handshake duration in seconds",
			Buckets: []float64{0.1, 0.5, 1.0, 2.0, 5.0, 10.0},
		}),
		meshRoutes: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "cloudbridge_p2p_mesh_routes",
			Help: "Number of active mesh routes",
		}),
		peerDiscoveryEvents: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "cloudbridge_p2p_discovery_events_total",
			Help: "Total peer discovery events",
		}, []string{"event_type"}),
		fallbackEvents: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "cloudbridge_p2p_fallback_events_total",
			Help: "Total P2P fallback events",
		}, []string{"from_protocol", "to_protocol", "reason"}),
		whitelistViolations: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "cloudbridge_p2p_whitelist_violations_total",
			Help: "Total peer whitelist violations",
		}, []string{"violation_type"}),
	}

	reg.MustRegister(
		m.connectionsTotal,
		m.activePeers,
		m.trafficBytes,
		m.latencySeconds,
		m.handshakeDuration,
		m.meshRoutes,
		m.peerDiscoveryEvents,
		m.fallbackEvents,
		m.whitelistViolations,
	)

	return m
}

// IncP2PConnections increments P2P connections counter
func (m *P2PMetrics) IncP2PConnections() {
	m.connectionsTotal.Inc()
}

// SetActivePeers sets the number of active peers
func (m *P2PMetrics) SetActivePeers(count int) {
	m.activePeers.Set(float64(count))
}

// IncP2PTraffic increments P2P traffic counter
func (m *P2PMetrics) IncP2PTraffic(direction, peerID string, bytes int64) {
	m.trafficBytes.WithLabelValues(direction, peerID).Add(float64(bytes))
}

// ObserveP2PLatency observes P2P latency
func (m *P2PMetrics) ObserveP2PLatency(peerID string, duration float64) {
	m.latencySeconds.WithLabelValues(peerID).Observe(duration)
}

// ObserveHandshakeDuration observes handshake duration
func (m *P2PMetrics) ObserveHandshakeDuration(duration float64) {
	m.handshakeDuration.Observe(duration)
}

// SetMeshRoutes sets the number of mesh routes
func (m *P2PMetrics) SetMeshRoutes(count int) {
	m.meshRoutes.Set(float64(count))
}

// IncPeerDiscoveryEvent increments peer discovery event counter
func (m *P2PMetrics) IncPeerDiscoveryEvent(eventType string) {
	m.peerDiscoveryEvents.WithLabelValues(eventType).Inc()
}

// IncFallbackEvent increments fallback event counter
func (m *P2PMetrics) IncFallbackEvent(fromProtocol, toProtocol, reason string) {
	m.fallbackEvents.WithLabelValues(fromProtocol, toProtocol, reason).Inc()
}

// IncWhitelistViolation increments whitelist violation counter
func (m *P2PMetrics) IncWhitelistViolation(violationType string) {
	m.whitelistViolations.WithLabelValues(violationType).Inc()
}
