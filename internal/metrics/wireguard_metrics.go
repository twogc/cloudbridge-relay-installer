package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// WireGuardMetrics представляет метрики для WireGuard VPN
type WireGuardMetrics struct {
	// Интерфейсы
	interfacesTotal     prometheus.Counter
	interfacesActive    prometheus.Gauge
	interfacesFailed    *prometheus.CounterVec
	interfacesCreated   *prometheus.CounterVec
	interfacesDestroyed *prometheus.CounterVec

	// Пиры (Peers)
	peersTotal     prometheus.Counter
	peersActive    prometheus.Gauge
	peersConnected *prometheus.GaugeVec
	peersAdded     *prometheus.CounterVec
	peersRemoved   *prometheus.CounterVec
	peersFailed    *prometheus.CounterVec

	// Трафик
	trafficBytesReceived   *prometheus.CounterVec
	trafficBytesSent       *prometheus.CounterVec
	trafficPacketsReceived *prometheus.CounterVec
	trafficPacketsSent     *prometheus.CounterVec
	trafficRate            *prometheus.GaugeVec

	// Производительность
	handshakeDuration    *prometheus.HistogramVec
	connectionLatency    *prometheus.HistogramVec
	throughputBytes      *prometheus.CounterVec
	throughputPackets    *prometheus.CounterVec
	bandwidthUtilization *prometheus.GaugeVec

	// Качество соединения
	keepaliveInterval *prometheus.GaugeVec
	lastHandshake     *prometheus.GaugeVec
	handshakeAttempts *prometheus.CounterVec
	handshakeFailures *prometheus.CounterVec
	connectionUptime  *prometheus.GaugeVec

	// Ошибки и сбои
	interfaceErrors      *prometheus.CounterVec
	peerErrors           *prometheus.CounterVec
	handshakeErrors      *prometheus.CounterVec
	timeoutEvents        *prometheus.CounterVec
	reconnectionAttempts *prometheus.CounterVec

	// Безопасность
	keyRotationEvents      *prometheus.CounterVec
	authenticationFailures *prometheus.CounterVec
	encryptionFailures     *prometheus.CounterVec
	unauthorizedAccess     *prometheus.CounterVec

	// Site-to-Site VPN
	siteToSiteConnections *prometheus.GaugeVec
	siteToSiteActive      *prometheus.GaugeVec
	siteToSiteBandwidth   *prometheus.GaugeVec
	siteToSiteLatency     *prometheus.HistogramVec

	// Маршрутизация
	routingRules    *prometheus.GaugeVec
	routingUpdates  *prometheus.CounterVec
	routingFailures *prometheus.CounterVec
	routeChanges    *prometheus.CounterVec

	// Мониторинг ресурсов
	memoryUsageBytes    *prometheus.GaugeVec
	cpuUsagePercent     *prometheus.GaugeVec
	goroutineCount      *prometheus.GaugeVec
	fileDescriptorCount *prometheus.GaugeVec

	// Конфигурация
	configUpdates       *prometheus.CounterVec
	configErrors        *prometheus.CounterVec
	keyGenerationEvents *prometheus.CounterVec
	interfaceStatus     *prometheus.GaugeVec
}

// NewWireGuardMetrics создает новые WireGuard метрики
func NewWireGuardMetrics(reg prometheus.Registerer) *WireGuardMetrics {
	m := &WireGuardMetrics{
		// Интерфейсы
		interfacesTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "relay_wireguard_interfaces_total",
			Help: "Total number of WireGuard interfaces",
		}),
		interfacesActive: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "relay_wireguard_interfaces_active",
			Help: "Number of active WireGuard interfaces",
		}),
		interfacesFailed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_wireguard_interfaces_failed_total",
			Help: "Total number of failed WireGuard interfaces",
		}, []string{"interface", "reason"}),
		interfacesCreated: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_wireguard_interfaces_created_total",
			Help: "Total number of created WireGuard interfaces",
		}, []string{"interface", "type"}),
		interfacesDestroyed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_wireguard_interfaces_destroyed_total",
			Help: "Total number of destroyed WireGuard interfaces",
		}, []string{"interface", "reason"}),

		// Пиры
		peersTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "relay_wireguard_peers_total",
			Help: "Total number of WireGuard peers",
		}),
		peersActive: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "relay_wireguard_peers_active",
			Help: "Number of active WireGuard peers",
		}),
		peersConnected: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_wireguard_peers_connected",
			Help: "Number of connected WireGuard peers",
		}, []string{"interface", "peer"}),
		peersAdded: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_wireguard_peers_added_total",
			Help: "Total number of added WireGuard peers",
		}, []string{"interface", "peer"}),
		peersRemoved: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_wireguard_peers_removed_total",
			Help: "Total number of removed WireGuard peers",
		}, []string{"interface", "peer", "reason"}),
		peersFailed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_wireguard_peers_failed_total",
			Help: "Total number of failed WireGuard peers",
		}, []string{"interface", "peer", "reason"}),

		// Трафик
		trafficBytesReceived: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_wireguard_traffic_bytes_received_total",
			Help: "Total bytes received through WireGuard",
		}, []string{"interface", "peer"}),
		trafficBytesSent: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_wireguard_traffic_bytes_sent_total",
			Help: "Total bytes sent through WireGuard",
		}, []string{"interface", "peer"}),
		trafficPacketsReceived: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_wireguard_traffic_packets_received_total",
			Help: "Total packets received through WireGuard",
		}, []string{"interface", "peer"}),
		trafficPacketsSent: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_wireguard_traffic_packets_sent_total",
			Help: "Total packets sent through WireGuard",
		}, []string{"interface", "peer"}),
		trafficRate: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_wireguard_traffic_rate_bps",
			Help: "WireGuard traffic rate in bytes per second",
		}, []string{"interface", "peer", "direction"}),

		// Производительность
		handshakeDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "relay_wireguard_handshake_duration_seconds",
			Help:    "WireGuard handshake duration in seconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.2, 0.5, 1.0, 2.0, 5.0},
		}, []string{"interface", "peer"}),
		connectionLatency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "relay_wireguard_connection_latency_seconds",
			Help:    "WireGuard connection latency in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.2, 0.5, 1.0},
		}, []string{"interface", "peer"}),
		throughputBytes: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_wireguard_throughput_bytes_total",
			Help: "Total WireGuard throughput in bytes",
		}, []string{"interface", "peer", "direction"}),
		throughputPackets: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_wireguard_throughput_packets_total",
			Help: "Total WireGuard throughput in packets",
		}, []string{"interface", "peer", "direction"}),
		bandwidthUtilization: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_wireguard_bandwidth_utilization_percent",
			Help: "WireGuard bandwidth utilization percentage",
		}, []string{"interface", "peer"}),

		// Качество соединения
		keepaliveInterval: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_wireguard_keepalive_interval_seconds",
			Help: "WireGuard keepalive interval in seconds",
		}, []string{"interface", "peer"}),
		lastHandshake: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_wireguard_last_handshake_timestamp",
			Help: "Timestamp of last WireGuard handshake",
		}, []string{"interface", "peer"}),
		handshakeAttempts: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_wireguard_handshake_attempts_total",
			Help: "Total number of WireGuard handshake attempts",
		}, []string{"interface", "peer"}),
		handshakeFailures: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_wireguard_handshake_failures_total",
			Help: "Total number of WireGuard handshake failures",
		}, []string{"interface", "peer", "reason"}),
		connectionUptime: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_wireguard_connection_uptime_seconds",
			Help: "WireGuard connection uptime in seconds",
		}, []string{"interface", "peer"}),

		// Ошибки и сбои
		interfaceErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_wireguard_interface_errors_total",
			Help: "Total number of WireGuard interface errors",
		}, []string{"interface", "error_type"}),
		peerErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_wireguard_peer_errors_total",
			Help: "Total number of WireGuard peer errors",
		}, []string{"interface", "peer", "error_type"}),
		handshakeErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_wireguard_handshake_errors_total",
			Help: "Total number of WireGuard handshake errors",
		}, []string{"interface", "peer", "error_type"}),
		timeoutEvents: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_wireguard_timeout_events_total",
			Help: "Total number of WireGuard timeout events",
		}, []string{"interface", "peer", "timeout_type"}),
		reconnectionAttempts: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_wireguard_reconnection_attempts_total",
			Help: "Total number of WireGuard reconnection attempts",
		}, []string{"interface", "peer"}),

		// Безопасность
		keyRotationEvents: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_wireguard_key_rotation_events_total",
			Help: "Total number of WireGuard key rotation events",
		}, []string{"interface", "key_type"}),
		authenticationFailures: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_wireguard_authentication_failures_total",
			Help: "Total number of WireGuard authentication failures",
		}, []string{"interface", "peer", "failure_type"}),
		encryptionFailures: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_wireguard_encryption_failures_total",
			Help: "Total number of WireGuard encryption failures",
		}, []string{"interface", "peer", "failure_type"}),
		unauthorizedAccess: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_wireguard_unauthorized_access_total",
			Help: "Total number of WireGuard unauthorized access attempts",
		}, []string{"interface", "peer", "access_type"}),

		// Site-to-Site VPN
		siteToSiteConnections: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_wireguard_site_to_site_connections",
			Help: "Number of WireGuard site-to-site connections",
		}, []string{"connection_id", "status"}),
		siteToSiteActive: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_wireguard_site_to_site_active",
			Help: "Number of active WireGuard site-to-site connections",
		}, []string{"connection_id"}),
		siteToSiteBandwidth: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_wireguard_site_to_site_bandwidth_bps",
			Help: "WireGuard site-to-site bandwidth in bytes per second",
		}, []string{"connection_id"}),
		siteToSiteLatency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "relay_wireguard_site_to_site_latency_seconds",
			Help:    "WireGuard site-to-site latency in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.2, 0.5, 1.0},
		}, []string{"connection_id"}),

		// Маршрутизация
		routingRules: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_wireguard_routing_rules",
			Help: "Number of WireGuard routing rules",
		}, []string{"interface", "rule_type"}),
		routingUpdates: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_wireguard_routing_updates_total",
			Help: "Total number of WireGuard routing updates",
		}, []string{"interface", "update_type"}),
		routingFailures: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_wireguard_routing_failures_total",
			Help: "Total number of WireGuard routing failures",
		}, []string{"interface", "failure_type"}),
		routeChanges: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_wireguard_route_changes_total",
			Help: "Total number of WireGuard route changes",
		}, []string{"interface", "change_type"}),

		// Мониторинг ресурсов
		memoryUsageBytes: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_wireguard_memory_usage_bytes",
			Help: "WireGuard memory usage in bytes",
		}, []string{"interface", "component"}),
		cpuUsagePercent: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_wireguard_cpu_usage_percent",
			Help: "WireGuard CPU usage percentage",
		}, []string{"interface", "component"}),
		goroutineCount: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_wireguard_goroutine_count",
			Help: "WireGuard goroutine count",
		}, []string{"interface", "component"}),
		fileDescriptorCount: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_wireguard_file_descriptor_count",
			Help: "WireGuard file descriptor count",
		}, []string{"interface"}),

		// Конфигурация
		configUpdates: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_wireguard_config_updates_total",
			Help: "Total number of WireGuard configuration updates",
		}, []string{"interface", "update_type"}),
		configErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_wireguard_config_errors_total",
			Help: "Total number of WireGuard configuration errors",
		}, []string{"interface", "error_type"}),
		keyGenerationEvents: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_wireguard_key_generation_events_total",
			Help: "Total number of WireGuard key generation events",
		}, []string{"interface", "key_type"}),
		interfaceStatus: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_wireguard_interface_status",
			Help: "WireGuard interface status (1=up, 0=down)",
		}, []string{"interface", "status"}),
	}

	// Регистрация метрик
	reg.MustRegister(
		m.interfacesTotal,
		m.interfacesActive,
		m.interfacesFailed,
		m.interfacesCreated,
		m.interfacesDestroyed,
		m.peersTotal,
		m.peersActive,
		m.peersConnected,
		m.peersAdded,
		m.peersRemoved,
		m.peersFailed,
		m.trafficBytesReceived,
		m.trafficBytesSent,
		m.trafficPacketsReceived,
		m.trafficPacketsSent,
		m.trafficRate,
		m.handshakeDuration,
		m.connectionLatency,
		m.throughputBytes,
		m.throughputPackets,
		m.bandwidthUtilization,
		m.keepaliveInterval,
		m.lastHandshake,
		m.handshakeAttempts,
		m.handshakeFailures,
		m.connectionUptime,
		m.interfaceErrors,
		m.peerErrors,
		m.handshakeErrors,
		m.timeoutEvents,
		m.reconnectionAttempts,
		m.keyRotationEvents,
		m.authenticationFailures,
		m.encryptionFailures,
		m.unauthorizedAccess,
		m.siteToSiteConnections,
		m.siteToSiteActive,
		m.siteToSiteBandwidth,
		m.siteToSiteLatency,
		m.routingRules,
		m.routingUpdates,
		m.routingFailures,
		m.routeChanges,
		m.memoryUsageBytes,
		m.cpuUsagePercent,
		m.goroutineCount,
		m.fileDescriptorCount,
		m.configUpdates,
		m.configErrors,
		m.keyGenerationEvents,
		m.interfaceStatus,
	)

	return m
}

// Методы для работы с WireGuard метриками

// Interface Metrics
func (m *WireGuardMetrics) IncInterfaces() {
	m.interfacesTotal.Inc()
	m.interfacesActive.Inc()
}

func (m *WireGuardMetrics) DecInterfaces() {
	m.interfacesActive.Dec()
}

func (m *WireGuardMetrics) IncInterfacesFailed(interfaceName, reason string) {
	m.interfacesFailed.WithLabelValues(interfaceName, reason).Inc()
}

func (m *WireGuardMetrics) IncInterfacesCreated(interfaceName, interfaceType string) {
	m.interfacesCreated.WithLabelValues(interfaceName, interfaceType).Inc()
}

func (m *WireGuardMetrics) IncInterfacesDestroyed(interfaceName, reason string) {
	m.interfacesDestroyed.WithLabelValues(interfaceName, reason).Inc()
}

// Peer Metrics
func (m *WireGuardMetrics) IncPeers() {
	m.peersTotal.Inc()
	m.peersActive.Inc()
}

func (m *WireGuardMetrics) DecPeers() {
	m.peersActive.Dec()
}

func (m *WireGuardMetrics) SetPeersConnected(interfaceName, peer string, connected int) {
	m.peersConnected.WithLabelValues(interfaceName, peer).Set(float64(connected))
}

func (m *WireGuardMetrics) IncPeersAdded(interfaceName, peer string) {
	m.peersAdded.WithLabelValues(interfaceName, peer).Inc()
}

func (m *WireGuardMetrics) IncPeersRemoved(interfaceName, peer, reason string) {
	m.peersRemoved.WithLabelValues(interfaceName, peer, reason).Inc()
}

func (m *WireGuardMetrics) IncPeersFailed(interfaceName, peer, reason string) {
	m.peersFailed.WithLabelValues(interfaceName, peer, reason).Inc()
}

// Traffic Metrics
func (m *WireGuardMetrics) IncTrafficBytesReceived(interfaceName, peer string, bytes int64) {
	m.trafficBytesReceived.WithLabelValues(interfaceName, peer).Add(float64(bytes))
}

func (m *WireGuardMetrics) IncTrafficBytesSent(interfaceName, peer string, bytes int64) {
	m.trafficBytesSent.WithLabelValues(interfaceName, peer).Add(float64(bytes))
}

func (m *WireGuardMetrics) IncTrafficPacketsReceived(interfaceName, peer string, packets int64) {
	m.trafficPacketsReceived.WithLabelValues(interfaceName, peer).Add(float64(packets))
}

func (m *WireGuardMetrics) IncTrafficPacketsSent(interfaceName, peer string, packets int64) {
	m.trafficPacketsSent.WithLabelValues(interfaceName, peer).Add(float64(packets))
}

func (m *WireGuardMetrics) SetTrafficRate(interfaceName, peer, direction string, rate float64) {
	m.trafficRate.WithLabelValues(interfaceName, peer, direction).Set(rate)
}

// Performance Metrics
func (m *WireGuardMetrics) ObserveHandshakeDuration(interfaceName, peer string, duration float64) {
	m.handshakeDuration.WithLabelValues(interfaceName, peer).Observe(duration)
}

func (m *WireGuardMetrics) ObserveConnectionLatency(interfaceName, peer string, duration float64) {
	m.connectionLatency.WithLabelValues(interfaceName, peer).Observe(duration)
}

func (m *WireGuardMetrics) IncThroughputBytes(interfaceName, peer, direction string, bytes int64) {
	m.throughputBytes.WithLabelValues(interfaceName, peer, direction).Add(float64(bytes))
}

func (m *WireGuardMetrics) IncThroughputPackets(interfaceName, peer, direction string, packets int64) {
	m.throughputPackets.WithLabelValues(interfaceName, peer, direction).Add(float64(packets))
}

func (m *WireGuardMetrics) SetBandwidthUtilization(interfaceName, peer string, utilization float64) {
	m.bandwidthUtilization.WithLabelValues(interfaceName, peer).Set(utilization)
}

// Connection Quality Metrics
func (m *WireGuardMetrics) SetKeepaliveInterval(interfaceName, peer string, interval float64) {
	m.keepaliveInterval.WithLabelValues(interfaceName, peer).Set(interval)
}

func (m *WireGuardMetrics) SetLastHandshake(interfaceName, peer string, timestamp float64) {
	m.lastHandshake.WithLabelValues(interfaceName, peer).Set(timestamp)
}

func (m *WireGuardMetrics) IncHandshakeAttempts(interfaceName, peer string) {
	m.handshakeAttempts.WithLabelValues(interfaceName, peer).Inc()
}

func (m *WireGuardMetrics) IncHandshakeFailures(interfaceName, peer, reason string) {
	m.handshakeFailures.WithLabelValues(interfaceName, peer, reason).Inc()
}

func (m *WireGuardMetrics) SetConnectionUptime(interfaceName, peer string, uptime float64) {
	m.connectionUptime.WithLabelValues(interfaceName, peer).Set(uptime)
}

// Error Metrics
func (m *WireGuardMetrics) IncInterfaceErrors(interfaceName, errorType string) {
	m.interfaceErrors.WithLabelValues(interfaceName, errorType).Inc()
}

func (m *WireGuardMetrics) IncPeerErrors(interfaceName, peer, errorType string) {
	m.peerErrors.WithLabelValues(interfaceName, peer, errorType).Inc()
}

func (m *WireGuardMetrics) IncHandshakeErrors(interfaceName, peer, errorType string) {
	m.handshakeErrors.WithLabelValues(interfaceName, peer, errorType).Inc()
}

func (m *WireGuardMetrics) IncTimeoutEvents(interfaceName, peer, timeoutType string) {
	m.timeoutEvents.WithLabelValues(interfaceName, peer, timeoutType).Inc()
}

func (m *WireGuardMetrics) IncReconnectionAttempts(interfaceName, peer string) {
	m.reconnectionAttempts.WithLabelValues(interfaceName, peer).Inc()
}

// Security Metrics
func (m *WireGuardMetrics) IncKeyRotationEvents(interfaceName, keyType string) {
	m.keyRotationEvents.WithLabelValues(interfaceName, keyType).Inc()
}

func (m *WireGuardMetrics) IncAuthenticationFailures(interfaceName, peer, failureType string) {
	m.authenticationFailures.WithLabelValues(interfaceName, peer, failureType).Inc()
}

func (m *WireGuardMetrics) IncEncryptionFailures(interfaceName, peer, failureType string) {
	m.encryptionFailures.WithLabelValues(interfaceName, peer, failureType).Inc()
}

func (m *WireGuardMetrics) IncUnauthorizedAccess(interfaceName, peer, accessType string) {
	m.unauthorizedAccess.WithLabelValues(interfaceName, peer, accessType).Inc()
}

// Site-to-Site VPN Metrics
func (m *WireGuardMetrics) SetSiteToSiteConnections(connectionID, status string, count int) {
	m.siteToSiteConnections.WithLabelValues(connectionID, status).Set(float64(count))
}

func (m *WireGuardMetrics) SetSiteToSiteActive(connectionID string, active int) {
	m.siteToSiteActive.WithLabelValues(connectionID).Set(float64(active))
}

func (m *WireGuardMetrics) SetSiteToSiteBandwidth(connectionID string, bandwidth float64) {
	m.siteToSiteBandwidth.WithLabelValues(connectionID).Set(bandwidth)
}

func (m *WireGuardMetrics) ObserveSiteToSiteLatency(connectionID string, latency float64) {
	m.siteToSiteLatency.WithLabelValues(connectionID).Observe(latency)
}

// Routing Metrics
func (m *WireGuardMetrics) SetRoutingRules(interfaceName, ruleType string, count int) {
	m.routingRules.WithLabelValues(interfaceName, ruleType).Set(float64(count))
}

func (m *WireGuardMetrics) IncRoutingUpdates(interfaceName, updateType string) {
	m.routingUpdates.WithLabelValues(interfaceName, updateType).Inc()
}

func (m *WireGuardMetrics) IncRoutingFailures(interfaceName, failureType string) {
	m.routingFailures.WithLabelValues(interfaceName, failureType).Inc()
}

func (m *WireGuardMetrics) IncRouteChanges(interfaceName, changeType string) {
	m.routeChanges.WithLabelValues(interfaceName, changeType).Inc()
}

// Resource Monitoring Metrics
func (m *WireGuardMetrics) SetMemoryUsage(interfaceName, component string, bytes int64) {
	m.memoryUsageBytes.WithLabelValues(interfaceName, component).Set(float64(bytes))
}

func (m *WireGuardMetrics) SetCPUUsage(interfaceName, component string, percent float64) {
	m.cpuUsagePercent.WithLabelValues(interfaceName, component).Set(percent)
}

func (m *WireGuardMetrics) SetGoroutineCount(interfaceName, component string, count int) {
	m.goroutineCount.WithLabelValues(interfaceName, component).Set(float64(count))
}

func (m *WireGuardMetrics) SetFileDescriptorCount(interfaceName string, count int) {
	m.fileDescriptorCount.WithLabelValues(interfaceName).Set(float64(count))
}

// Configuration Metrics
func (m *WireGuardMetrics) IncConfigUpdates(interfaceName, updateType string) {
	m.configUpdates.WithLabelValues(interfaceName, updateType).Inc()
}

func (m *WireGuardMetrics) IncConfigErrors(interfaceName, errorType string) {
	m.configErrors.WithLabelValues(interfaceName, errorType).Inc()
}

func (m *WireGuardMetrics) IncKeyGenerationEvents(interfaceName, keyType string) {
	m.keyGenerationEvents.WithLabelValues(interfaceName, keyType).Inc()
}

func (m *WireGuardMetrics) SetInterfaceStatus(interfaceName, status string, up bool) {
	value := 0.0
	if up {
		value = 1.0
	}
	m.interfaceStatus.WithLabelValues(interfaceName, status).Set(value)
}



