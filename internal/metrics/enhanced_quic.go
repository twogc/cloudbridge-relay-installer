package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Enhanced QUIC Connection Metrics
	enhancedQuicConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "enhanced_quic_connections_total",
			Help: "Total number of Enhanced QUIC connections",
		},
		[]string{"tenant_id", "status", "region"},
	)

	enhancedQuicConnectionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "enhanced_quic_connection_duration_seconds",
			Help:    "Duration of Enhanced QUIC connections",
			Buckets: []float64{0.1, 0.5, 1, 5, 10, 30, 60, 300, 600},
		},
		[]string{"tenant_id", "region"},
	)

	// Data Transfer Metrics
	enhancedQuicBytesSent = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "enhanced_quic_bytes_sent_total",
			Help: "Total bytes sent via Enhanced QUIC",
		},
		[]string{"tenant_id", "region"},
	)

	enhancedQuicBytesReceived = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "enhanced_quic_bytes_received_total",
			Help: "Total bytes received via Enhanced QUIC",
		},
		[]string{"tenant_id", "region"},
	)

	enhancedQuicThroughput = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "enhanced_quic_throughput_bytes_per_second",
			Help: "Current throughput in bytes per second",
		},
		[]string{"tenant_id", "direction", "region"},
	)

	// Stream Metrics
	enhancedQuicStreams = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "enhanced_quic_streams_active",
			Help: "Number of active Enhanced QUIC streams",
		},
		[]string{"tenant_id", "stream_type", "region"},
	)

	enhancedQuicStreamsCreated = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "enhanced_quic_streams_created_total",
			Help: "Total number of Enhanced QUIC streams created",
		},
		[]string{"tenant_id", "stream_type", "region"},
	)

	// DDoS Protection Metrics
	enhancedQuicDDoSBlocked = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "enhanced_quic_ddos_blocked_total",
			Help: "Total DDoS attacks blocked by Enhanced QUIC",
		},
		[]string{"client_ip", "reason", "region"},
	)

	enhancedQuicDDoSDetected = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "enhanced_quic_ddos_detected_total",
			Help: "Total DDoS attacks detected by Enhanced QUIC",
		},
		[]string{"attack_type", "severity", "region"},
	)

	enhancedQuicRateLimitHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "enhanced_quic_rate_limit_hits_total",
			Help: "Total rate limit hits in Enhanced QUIC",
		},
		[]string{"tenant_id", "limit_type", "region"},
	)

	// Geographic Routing Metrics
	enhancedQuicGeoRoutes = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "enhanced_quic_geo_routes_total",
			Help: "Total geographic routes selected by Enhanced QUIC",
		},
		[]string{"source_region", "target_region", "route_type"},
	)

	enhancedQuicGeoLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "enhanced_quic_geo_routing_latency_seconds",
			Help:    "Latency of geographic routing decisions",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1},
		},
		[]string{"source_region", "target_region"},
	)

	// Tenant Management Metrics
	enhancedQuicTenantBandwidthUsage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "enhanced_quic_tenant_bandwidth_usage_bytes",
			Help: "Current bandwidth usage by tenant",
		},
		[]string{"tenant_id", "direction", "region"},
	)

	enhancedQuicTenantConnectionsActive = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "enhanced_quic_tenant_connections_active",
			Help: "Number of active connections per tenant",
		},
		[]string{"tenant_id", "region"},
	)

	enhancedQuicTenantLimitExceeded = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "enhanced_quic_tenant_limit_exceeded_total",
			Help: "Total tenant limit violations",
		},
		[]string{"tenant_id", "limit_type", "region"},
	)

	// Authentication Metrics
	enhancedQuicAuthenticationAttempts = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "enhanced_quic_authentication_attempts_total",
			Help: "Total authentication attempts in Enhanced QUIC",
		},
		[]string{"tenant_id", "auth_method", "status", "region"},
	)

	enhancedQuicKeycloakRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "enhanced_quic_keycloak_requests_total",
			Help: "Total requests to Keycloak from Enhanced QUIC",
		},
		[]string{"operation", "status", "region"},
	)

	enhancedQuicKeycloakLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "enhanced_quic_keycloak_latency_seconds",
			Help:    "Latency of Keycloak operations",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5},
		},
		[]string{"operation", "region"},
	)

	// Load Balancing Metrics
	enhancedQuicLoadBalancerRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "enhanced_quic_load_balancer_requests_total",
			Help: "Total requests processed by Enhanced QUIC load balancer",
		},
		[]string{"server", "algorithm", "status", "region"},
	)

	enhancedQuicServerHealth = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "enhanced_quic_server_health",
			Help: "Health status of Enhanced QUIC servers (1=healthy, 0=unhealthy)",
		},
		[]string{"server", "region"},
	)

	// Performance Metrics
	enhancedQuicLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "enhanced_quic_latency_seconds",
			Help:    "Enhanced QUIC operation latency",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1},
		},
		[]string{"operation", "tenant_id", "region"},
	)

	enhancedQuicCpuUsage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "enhanced_quic_cpu_usage_percent",
			Help: "CPU usage percentage for Enhanced QUIC operations",
		},
		[]string{"component", "region"},
	)

	enhancedQuicMemoryUsage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "enhanced_quic_memory_usage_bytes",
			Help: "Memory usage in bytes for Enhanced QUIC operations",
		},
		[]string{"component", "region"},
	)

	// Error Metrics
	enhancedQuicErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "enhanced_quic_errors_total",
			Help: "Total errors in Enhanced QUIC operations",
		},
		[]string{"error_type", "component", "tenant_id", "region"},
	)

	enhancedQuicTimeout = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "enhanced_quic_timeouts_total",
			Help: "Total timeouts in Enhanced QUIC operations",
		},
		[]string{"operation", "tenant_id", "region"},
	)
)

// RecordEnhancedQuicConnection записывает метрику соединения
func RecordEnhancedQuicConnection(tenantID, status, region string, delta float64) {
	enhancedQuicConnections.WithLabelValues(tenantID, status, region).Add(delta)
}

// RecordEnhancedQuicConnectionDuration записывает длительность соединения
func RecordEnhancedQuicConnectionDuration(tenantID, region string, duration float64) {
	enhancedQuicConnectionDuration.WithLabelValues(tenantID, region).Observe(duration)
}

// RecordEnhancedQuicBytes записывает метрики переданных данных
func RecordEnhancedQuicBytes(tenantID, region string, sent, received int64) {
	enhancedQuicBytesSent.WithLabelValues(tenantID, region).Add(float64(sent))
	enhancedQuicBytesReceived.WithLabelValues(tenantID, region).Add(float64(received))
}

// UpdateEnhancedQuicThroughput обновляет метрику пропускной способности
func UpdateEnhancedQuicThroughput(tenantID, direction, region string, bytesPerSecond float64) {
	enhancedQuicThroughput.WithLabelValues(tenantID, direction, region).Set(bytesPerSecond)
}

// RecordEnhancedQuicStream записывает метрики потоков
func RecordEnhancedQuicStream(tenantID, streamType, region string, delta float64) {
	enhancedQuicStreams.WithLabelValues(tenantID, streamType, region).Add(delta)
	if delta > 0 {
		enhancedQuicStreamsCreated.WithLabelValues(tenantID, streamType, region).Inc()
	}
}

// RecordEnhancedQuicDDoS записывает метрику DDoS блокировки
func RecordEnhancedQuicDDoS(clientIP, reason, region string) {
	enhancedQuicDDoSBlocked.WithLabelValues(clientIP, reason, region).Inc()
}

// RecordEnhancedQuicDDoSDetection записывает обнаружение DDoS атаки
func RecordEnhancedQuicDDoSDetection(attackType, severity, region string) {
	enhancedQuicDDoSDetected.WithLabelValues(attackType, severity, region).Inc()
}

// RecordEnhancedQuicRateLimit записывает превышение лимита скорости
func RecordEnhancedQuicRateLimit(tenantID, limitType, region string) {
	enhancedQuicRateLimitHits.WithLabelValues(tenantID, limitType, region).Inc()
}

// RecordEnhancedQuicGeoRoute записывает метрику географической маршрутизации
func RecordEnhancedQuicGeoRoute(sourceRegion, targetRegion, routeType string) {
	enhancedQuicGeoRoutes.WithLabelValues(sourceRegion, targetRegion, routeType).Inc()
}

// RecordEnhancedQuicGeoLatency записывает задержку географической маршрутизации
func RecordEnhancedQuicGeoLatency(sourceRegion, targetRegion string, latency float64) {
	enhancedQuicGeoLatency.WithLabelValues(sourceRegion, targetRegion).Observe(latency)
}

// UpdateEnhancedQuicTenantBandwidth обновляет использование пропускной способности тенанта
func UpdateEnhancedQuicTenantBandwidth(tenantID, direction, region string, bytes float64) {
	enhancedQuicTenantBandwidthUsage.WithLabelValues(tenantID, direction, region).Set(bytes)
}

// UpdateEnhancedQuicTenantConnections обновляет количество активных соединений тенанта
func UpdateEnhancedQuicTenantConnections(tenantID, region string, connections float64) {
	enhancedQuicTenantConnectionsActive.WithLabelValues(tenantID, region).Set(connections)
}

// RecordEnhancedQuicTenantLimitExceeded записывает превышение лимита тенанта
func RecordEnhancedQuicTenantLimitExceeded(tenantID, limitType, region string) {
	enhancedQuicTenantLimitExceeded.WithLabelValues(tenantID, limitType, region).Inc()
}

// RecordEnhancedQuicAuthentication записывает попытку аутентификации
func RecordEnhancedQuicAuthentication(tenantID, authMethod, status, region string) {
	enhancedQuicAuthenticationAttempts.WithLabelValues(tenantID, authMethod, status, region).Inc()
}

// RecordEnhancedQuicKeycloakRequest записывает запрос к Keycloak
func RecordEnhancedQuicKeycloakRequest(operation, status, region string) {
	enhancedQuicKeycloakRequests.WithLabelValues(operation, status, region).Inc()
}

// RecordEnhancedQuicKeycloakLatency записывает задержку операций Keycloak
func RecordEnhancedQuicKeycloakLatency(operation, region string, latency float64) {
	enhancedQuicKeycloakLatency.WithLabelValues(operation, region).Observe(latency)
}

// RecordEnhancedQuicLoadBalancerRequest записывает запрос к балансировщику нагрузки
func RecordEnhancedQuicLoadBalancerRequest(server, algorithm, status, region string) {
	enhancedQuicLoadBalancerRequests.WithLabelValues(server, algorithm, status, region).Inc()
}

// UpdateEnhancedQuicServerHealth обновляет статус здоровья сервера
func UpdateEnhancedQuicServerHealth(server, region string, healthy bool) {
	value := 0.0
	if healthy {
		value = 1.0
	}
	enhancedQuicServerHealth.WithLabelValues(server, region).Set(value)
}

// RecordEnhancedQuicLatency записывает задержку операции
func RecordEnhancedQuicLatency(operation, tenantID, region string, latency float64) {
	enhancedQuicLatency.WithLabelValues(operation, tenantID, region).Observe(latency)
}

// UpdateEnhancedQuicCpuUsage обновляет использование CPU
func UpdateEnhancedQuicCpuUsage(component, region string, cpuPercent float64) {
	enhancedQuicCpuUsage.WithLabelValues(component, region).Set(cpuPercent)
}

// UpdateEnhancedQuicMemoryUsage обновляет использование памяти
func UpdateEnhancedQuicMemoryUsage(component, region string, memoryBytes float64) {
	enhancedQuicMemoryUsage.WithLabelValues(component, region).Set(memoryBytes)
}

// RecordEnhancedQuicError записывает ошибку
func RecordEnhancedQuicError(errorType, component, tenantID, region string) {
	enhancedQuicErrors.WithLabelValues(errorType, component, tenantID, region).Inc()
}

// RecordEnhancedQuicTimeout записывает тайм-аут
func RecordEnhancedQuicTimeout(operation, tenantID, region string) {
	enhancedQuicTimeout.WithLabelValues(operation, tenantID, region).Inc()
}

// GetEnhancedQuicMetrics возвращает список всех Enhanced QUIC метрик
func GetEnhancedQuicMetrics() []prometheus.Collector {
	return []prometheus.Collector{
		enhancedQuicConnections,
		enhancedQuicConnectionDuration,
		enhancedQuicBytesSent,
		enhancedQuicBytesReceived,
		enhancedQuicThroughput,
		enhancedQuicStreams,
		enhancedQuicStreamsCreated,
		enhancedQuicDDoSBlocked,
		enhancedQuicDDoSDetected,
		enhancedQuicRateLimitHits,
		enhancedQuicGeoRoutes,
		enhancedQuicGeoLatency,
		enhancedQuicTenantBandwidthUsage,
		enhancedQuicTenantConnectionsActive,
		enhancedQuicTenantLimitExceeded,
		enhancedQuicAuthenticationAttempts,
		enhancedQuicKeycloakRequests,
		enhancedQuicKeycloakLatency,
		enhancedQuicLoadBalancerRequests,
		enhancedQuicServerHealth,
		enhancedQuicLatency,
		enhancedQuicCpuUsage,
		enhancedQuicMemoryUsage,
		enhancedQuicErrors,
		enhancedQuicTimeout,
	}
}
