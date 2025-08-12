package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// WireGuard Interface Metrics
	WireGuardPeersTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "wireguard_peers_total",
			Help: "Total number of WireGuard peers",
		},
		[]string{"interface"},
	)

	WireGuardInterfaceStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "wireguard_interface_status",
			Help: "WireGuard interface status (1 = running, 0 = stopped)",
		},
		[]string{"interface", "name"},
	)

	WireGuardBytesReceived = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "wireguard_bytes_received_total",
			Help: "Total bytes received through WireGuard interface",
		},
		[]string{"interface", "peer"},
	)

	WireGuardBytesSent = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "wireguard_bytes_sent_total",
			Help: "Total bytes sent through WireGuard interface",
		},
		[]string{"interface", "peer"},
	)

	WireGuardHandshakesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "wireguard_handshakes_total",
			Help: "Total number of WireGuard handshakes",
		},
		[]string{"interface", "peer", "status"},
	)

	// Mesh Topology Metrics
	WireGuardMeshConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "wireguard_mesh_connections",
			Help: "Number of active mesh connections",
		},
		[]string{"topology"},
	)

	WireGuardMeshRoutes = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "wireguard_mesh_routes",
			Help: "Number of routes in mesh topology",
		},
		[]string{"topology", "type"},
	)

	WireGuardMeshLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "wireguard_mesh_latency_seconds",
			Help:    "Mesh connection latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"topology", "peer"},
	)

	// Peer Discovery Metrics
	WireGuardDiscoveredPeers = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "wireguard_discovered_peers",
			Help: "Number of discovered peers",
		},
		[]string{"registry"},
	)

	WireGuardDiscoveryHeartbeats = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "wireguard_discovery_heartbeats_total",
			Help: "Total number of discovery heartbeats",
		},
		[]string{"registry", "status"},
	)

	WireGuardDiscoveryErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "wireguard_discovery_errors_total",
			Help: "Total number of discovery errors",
		},
		[]string{"registry", "error_type"},
	)

	// QUIC over WireGuard Metrics
	WireGuardQUICConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "wireguard_quic_connections",
			Help: "Number of active QUIC connections over WireGuard",
		},
		[]string{"interface"},
	)

	WireGuardQUICStreams = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "wireguard_quic_streams",
			Help: "Number of active QUIC streams over WireGuard",
		},
		[]string{"interface", "direction"},
	)

	WireGuardQUICErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "wireguard_quic_errors_total",
			Help: "Total number of QUIC over WireGuard errors",
		},
		[]string{"interface", "error_type"},
	)

	// Hybrid Router Metrics
	WireGuardRoutingDecisions = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "wireguard_routing_decisions_total",
			Help: "Total number of routing decisions",
		},
		[]string{"strategy", "route_type"},
	)

	WireGuardRoutingLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "wireguard_routing_latency_seconds",
			Help:    "Routing decision latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"strategy"},
	)

	WireGuardRoutingErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "wireguard_routing_errors_total",
			Help: "Total number of routing errors",
		},
		[]string{"strategy", "error_type"},
	)

	// Site-to-Site Metrics
	WireGuardSiteToSiteConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "wireguard_site_to_site_connections",
			Help: "Number of active site-to-site connections",
		},
		[]string{"network_id", "tenant_id"},
	)

	WireGuardSiteToSiteRoutes = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "wireguard_site_to_site_routes",
			Help: "Number of site-to-site routes",
		},
		[]string{"from_network", "to_network"},
	)

	WireGuardVPNReplacementStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "wireguard_vpn_replacement_status",
			Help: "VPN replacement status (1 = active, 0 = inactive)",
		},
		[]string{"network_id", "tenant_id"},
	)
)

// WireGuardMetrics предоставляет методы для обновления метрик WireGuard
type WireGuardMetrics struct{}

// NewWireGuardMetrics создает новый экземпляр WireGuardMetrics
func NewWireGuardMetrics() *WireGuardMetrics {
	return &WireGuardMetrics{}
}

// UpdateInterfaceStatus обновляет статус интерфейса
func (wm *WireGuardMetrics) UpdateInterfaceStatus(interfaceName, name string, running bool) {
	status := 0.0
	if running {
		status = 1.0
	}
	WireGuardInterfaceStatus.WithLabelValues(interfaceName, name).Set(status)
}

// UpdatePeersCount обновляет количество пиров
func (wm *WireGuardMetrics) UpdatePeersCount(interfaceName string, count int) {
	WireGuardPeersTotal.WithLabelValues(interfaceName).Set(float64(count))
}

// IncrementBytesReceived увеличивает счетчик полученных байт
func (wm *WireGuardMetrics) IncrementBytesReceived(interfaceName, peer string, bytes int64) {
	WireGuardBytesReceived.WithLabelValues(interfaceName, peer).Add(float64(bytes))
}

// IncrementBytesSent увеличивает счетчик отправленных байт
func (wm *WireGuardMetrics) IncrementBytesSent(interfaceName, peer string, bytes int64) {
	WireGuardBytesSent.WithLabelValues(interfaceName, peer).Add(float64(bytes))
}

// IncrementHandshakes увеличивает счетчик рукопожатий
func (wm *WireGuardMetrics) IncrementHandshakes(interfaceName, peer, status string) {
	WireGuardHandshakesTotal.WithLabelValues(interfaceName, peer, status).Inc()
}

// UpdateMeshConnections обновляет количество mesh соединений
func (wm *WireGuardMetrics) UpdateMeshConnections(topology string, count int) {
	WireGuardMeshConnections.WithLabelValues(topology).Set(float64(count))
}

// UpdateMeshRoutes обновляет количество маршрутов в mesh
func (wm *WireGuardMetrics) UpdateMeshRoutes(topology, routeType string, count int) {
	WireGuardMeshRoutes.WithLabelValues(topology, routeType).Set(float64(count))
}

// ObserveMeshLatency наблюдает задержку mesh соединений
func (wm *WireGuardMetrics) ObserveMeshLatency(topology, peer string, latency float64) {
	WireGuardMeshLatency.WithLabelValues(topology, peer).Observe(latency)
}

// UpdateDiscoveredPeers обновляет количество обнаруженных пиров
func (wm *WireGuardMetrics) UpdateDiscoveredPeers(registry string, count int) {
	WireGuardDiscoveredPeers.WithLabelValues(registry).Set(float64(count))
}

// IncrementDiscoveryHeartbeats увеличивает счетчик heartbeat'ов
func (wm *WireGuardMetrics) IncrementDiscoveryHeartbeats(registry, status string) {
	WireGuardDiscoveryHeartbeats.WithLabelValues(registry, status).Inc()
}

// IncrementDiscoveryErrors увеличивает счетчик ошибок обнаружения
func (wm *WireGuardMetrics) IncrementDiscoveryErrors(registry, errorType string) {
	WireGuardDiscoveryErrors.WithLabelValues(registry, errorType).Inc()
}

// UpdateQUICConnections обновляет количество QUIC соединений
func (wm *WireGuardMetrics) UpdateQUICConnections(interfaceName string, count int) {
	WireGuardQUICConnections.WithLabelValues(interfaceName).Set(float64(count))
}

// UpdateQUICStreams обновляет количество QUIC потоков
func (wm *WireGuardMetrics) UpdateQUICStreams(interfaceName, direction string, count int) {
	WireGuardQUICStreams.WithLabelValues(interfaceName, direction).Set(float64(count))
}

// IncrementQUICErrors увеличивает счетчик QUIC ошибок
func (wm *WireGuardMetrics) IncrementQUICErrors(interfaceName, errorType string) {
	WireGuardQUICErrors.WithLabelValues(interfaceName, errorType).Inc()
}

// IncrementRoutingDecisions увеличивает счетчик решений маршрутизации
func (wm *WireGuardMetrics) IncrementRoutingDecisions(strategy, routeType string) {
	WireGuardRoutingDecisions.WithLabelValues(strategy, routeType).Inc()
}

// ObserveRoutingLatency наблюдает задержку маршрутизации
func (wm *WireGuardMetrics) ObserveRoutingLatency(strategy string, latency float64) {
	WireGuardRoutingLatency.WithLabelValues(strategy).Observe(latency)
}

// IncrementRoutingErrors увеличивает счетчик ошибок маршрутизации
func (wm *WireGuardMetrics) IncrementRoutingErrors(strategy, errorType string) {
	WireGuardRoutingErrors.WithLabelValues(strategy, errorType).Inc()
}

// UpdateSiteToSiteConnections обновляет количество site-to-site соединений
func (wm *WireGuardMetrics) UpdateSiteToSiteConnections(networkID, tenantID string, count int) {
	WireGuardSiteToSiteConnections.WithLabelValues(networkID, tenantID).Set(float64(count))
}

// UpdateSiteToSiteRoutes обновляет количество site-to-site маршрутов
func (wm *WireGuardMetrics) UpdateSiteToSiteRoutes(fromNetwork, toNetwork string, count int) {
	WireGuardSiteToSiteRoutes.WithLabelValues(fromNetwork, toNetwork).Set(float64(count))
}

// UpdateVPNReplacementStatus обновляет статус замены VPN
func (wm *WireGuardMetrics) UpdateVPNReplacementStatus(networkID, tenantID string, active bool) {
	status := 0.0
	if active {
		status = 1.0
	}
	WireGuardVPNReplacementStatus.WithLabelValues(networkID, tenantID).Set(status)
}
