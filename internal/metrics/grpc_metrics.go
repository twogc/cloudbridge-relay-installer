package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// GRPCMetrics представляет метрики для gRPC сервера
type GRPCMetrics struct {
	// Соединения
	connectionsTotal   prometheus.Counter
	connectionsActive  prometheus.Gauge
	connectionsFailed  *prometheus.CounterVec
	connectionsDropped *prometheus.CounterVec

	// Запросы
	requestsTotal    *prometheus.CounterVec
	requestsInFlight *prometheus.GaugeVec
	requestsDuration *prometheus.HistogramVec
	requestsSize     *prometheus.HistogramVec
	responsesSize    *prometheus.HistogramVec

	// Методы
	methodCalls      *prometheus.CounterVec
	methodDuration   *prometheus.HistogramVec
	methodErrors     *prometheus.CounterVec
	methodThroughput *prometheus.CounterVec

	// Потоки (Streams)
	streamsTotal   prometheus.Counter
	streamsActive  prometheus.Gauge
	streamsCreated *prometheus.CounterVec
	streamsClosed  *prometheus.CounterVec
	streamsErrors  *prometheus.CounterVec

	// Производительность
	handshakeDuration    prometheus.Histogram
	connectionLatency    *prometheus.HistogramVec
	throughputBytes      *prometheus.CounterVec
	throughputMessages   *prometheus.CounterVec
	bandwidthUtilization *prometheus.GaugeVec

	// Качество соединения
	keepaliveEvents   *prometheus.CounterVec
	keepaliveFailures *prometheus.CounterVec
	connectionUptime  *prometheus.GaugeVec
	lastActivity      *prometheus.GaugeVec

	// Ошибки и сбои
	serverErrors         *prometheus.CounterVec
	clientErrors         *prometheus.CounterVec
	timeoutEvents        *prometheus.CounterVec
	reconnectionAttempts *prometheus.CounterVec
	circuitBreakerEvents *prometheus.CounterVec

	// Безопасность
	authenticationFailures *prometheus.CounterVec
	authorizationFailures  *prometheus.CounterVec
	rateLimitViolations    *prometheus.CounterVec
	rbacViolations         *prometheus.CounterVec

	// Middleware
	middlewareDuration *prometheus.HistogramVec
	middlewareErrors   *prometheus.CounterVec
	interceptorCalls   *prometheus.CounterVec

	// Тенанты
	tenantRequests    *prometheus.CounterVec
	tenantConnections *prometheus.GaugeVec
	tenantBandwidth   *prometheus.CounterVec
	tenantErrors      *prometheus.CounterVec

	// Мониторинг ресурсов
	memoryUsageBytes    *prometheus.GaugeVec
	cpuUsagePercent     *prometheus.GaugeVec
	goroutineCount      *prometheus.GaugeVec
	fileDescriptorCount prometheus.Gauge

	// Конфигурация
	serverStatus  prometheus.Gauge
	serverVersion *prometheus.GaugeVec
	configUpdates *prometheus.CounterVec
	configErrors  *prometheus.CounterVec
}

// NewGRPCMetrics создает новые gRPC метрики
func NewGRPCMetrics(reg prometheus.Registerer) *GRPCMetrics {
	m := &GRPCMetrics{
		// Соединения
		connectionsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "relay_grpc_connections_total",
			Help: "Total number of gRPC connections",
		}),
		connectionsActive: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "relay_grpc_connections_active",
			Help: "Number of active gRPC connections",
		}),
		connectionsFailed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_grpc_connections_failed_total",
			Help: "Total number of failed gRPC connections",
		}, []string{"reason", "client_ip"}),
		connectionsDropped: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_grpc_connections_dropped_total",
			Help: "Total number of dropped gRPC connections",
		}, []string{"reason", "client_ip"}),

		// Запросы
		requestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_grpc_requests_total",
			Help: "Total number of gRPC requests",
		}, []string{"method", "service", "status_code"}),
		requestsInFlight: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_grpc_requests_in_flight",
			Help: "Number of gRPC requests in flight",
		}, []string{"method", "service"}),
		requestsDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "relay_grpc_requests_duration_seconds",
			Help:    "gRPC request duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.2, 0.5, 1.0, 2.0, 5.0},
		}, []string{"method", "service"}),
		requestsSize: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "relay_grpc_requests_size_bytes",
			Help:    "gRPC request size in bytes",
			Buckets: []float64{64, 256, 1024, 4096, 16384, 65536, 262144, 1048576},
		}, []string{"method", "service"}),
		responsesSize: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "relay_grpc_responses_size_bytes",
			Help:    "gRPC response size in bytes",
			Buckets: []float64{64, 256, 1024, 4096, 16384, 65536, 262144, 1048576},
		}, []string{"method", "service"}),

		// Методы
		methodCalls: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_grpc_method_calls_total",
			Help: "Total number of gRPC method calls",
		}, []string{"method", "service", "status"}),
		methodDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "relay_grpc_method_duration_seconds",
			Help:    "gRPC method duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.2, 0.5, 1.0, 2.0, 5.0},
		}, []string{"method", "service"}),
		methodErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_grpc_method_errors_total",
			Help: "Total number of gRPC method errors",
		}, []string{"method", "service", "error_type"}),
		methodThroughput: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_grpc_method_throughput_total",
			Help: "Total gRPC method throughput",
		}, []string{"method", "service", "direction"}),

		// Потоки
		streamsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "relay_grpc_streams_total",
			Help: "Total number of gRPC streams",
		}),
		streamsActive: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "relay_grpc_streams_active",
			Help: "Number of active gRPC streams",
		}),
		streamsCreated: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_grpc_streams_created_total",
			Help: "Total number of created gRPC streams",
		}, []string{"method", "service"}),
		streamsClosed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_grpc_streams_closed_total",
			Help: "Total number of closed gRPC streams",
		}, []string{"method", "service", "reason"}),
		streamsErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_grpc_streams_errors_total",
			Help: "Total number of gRPC stream errors",
		}, []string{"method", "service", "error_type"}),

		// Производительность
		handshakeDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "relay_grpc_handshake_duration_seconds",
			Help:    "gRPC handshake duration in seconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.2, 0.5, 1.0, 2.0, 5.0},
		}),
		connectionLatency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "relay_grpc_connection_latency_seconds",
			Help:    "gRPC connection latency in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.2, 0.5, 1.0},
		}, []string{"client_ip"}),
		throughputBytes: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_grpc_throughput_bytes_total",
			Help: "Total gRPC throughput in bytes",
		}, []string{"method", "service", "direction"}),
		throughputMessages: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_grpc_throughput_messages_total",
			Help: "Total gRPC throughput in messages",
		}, []string{"method", "service", "direction"}),
		bandwidthUtilization: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_grpc_bandwidth_utilization_percent",
			Help: "gRPC bandwidth utilization percentage",
		}, []string{"method", "service"}),

		// Качество соединения
		keepaliveEvents: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_grpc_keepalive_events_total",
			Help: "Total number of gRPC keepalive events",
		}, []string{"client_ip", "event_type"}),
		keepaliveFailures: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_grpc_keepalive_failures_total",
			Help: "Total number of gRPC keepalive failures",
		}, []string{"client_ip", "failure_type"}),
		connectionUptime: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_grpc_connection_uptime_seconds",
			Help: "gRPC connection uptime in seconds",
		}, []string{"client_ip"}),
		lastActivity: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_grpc_last_activity_timestamp",
			Help: "Timestamp of last gRPC activity",
		}, []string{"client_ip"}),

		// Ошибки и сбои
		serverErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_grpc_server_errors_total",
			Help: "Total number of gRPC server errors",
		}, []string{"error_type", "method", "service"}),
		clientErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_grpc_client_errors_total",
			Help: "Total number of gRPC client errors",
		}, []string{"error_type", "client_ip"}),
		timeoutEvents: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_grpc_timeout_events_total",
			Help: "Total number of gRPC timeout events",
		}, []string{"timeout_type", "method", "service"}),
		reconnectionAttempts: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_grpc_reconnection_attempts_total",
			Help: "Total number of gRPC reconnection attempts",
		}, []string{"client_ip"}),
		circuitBreakerEvents: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_grpc_circuit_breaker_events_total",
			Help: "Total number of gRPC circuit breaker events",
		}, []string{"method", "service", "event_type"}),

		// Безопасность
		authenticationFailures: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_grpc_authentication_failures_total",
			Help: "Total number of gRPC authentication failures",
		}, []string{"client_ip", "failure_type"}),
		authorizationFailures: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_grpc_authorization_failures_total",
			Help: "Total number of gRPC authorization failures",
		}, []string{"client_ip", "method", "service"}),
		rateLimitViolations: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_grpc_rate_limit_violations_total",
			Help: "Total number of gRPC rate limit violations",
		}, []string{"client_ip", "limit_type"}),
		rbacViolations: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_grpc_rbac_violations_total",
			Help: "Total number of gRPC RBAC violations",
		}, []string{"client_ip", "method", "service", "violation_type"}),

		// Middleware
		middlewareDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "relay_grpc_middleware_duration_seconds",
			Help:    "gRPC middleware duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.2, 0.5, 1.0},
		}, []string{"middleware", "method", "service"}),
		middlewareErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_grpc_middleware_errors_total",
			Help: "Total number of gRPC middleware errors",
		}, []string{"middleware", "error_type"}),
		interceptorCalls: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_grpc_interceptor_calls_total",
			Help: "Total number of gRPC interceptor calls",
		}, []string{"interceptor", "method", "service"}),

		// Тенанты
		tenantRequests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_grpc_tenant_requests_total",
			Help: "Total number of gRPC requests per tenant",
		}, []string{"tenant_id", "method", "service"}),
		tenantConnections: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_grpc_tenant_connections",
			Help: "Number of gRPC connections per tenant",
		}, []string{"tenant_id"}),
		tenantBandwidth: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_grpc_tenant_bandwidth_bytes_total",
			Help: "Total gRPC bandwidth per tenant",
		}, []string{"tenant_id", "direction"}),
		tenantErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_grpc_tenant_errors_total",
			Help: "Total number of gRPC errors per tenant",
		}, []string{"tenant_id", "error_type"}),

		// Мониторинг ресурсов
		memoryUsageBytes: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_grpc_memory_usage_bytes",
			Help: "gRPC memory usage in bytes",
		}, []string{"component"}),
		cpuUsagePercent: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_grpc_cpu_usage_percent",
			Help: "gRPC CPU usage percentage",
		}, []string{"component"}),
		goroutineCount: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_grpc_goroutine_count",
			Help: "gRPC goroutine count",
		}, []string{"component"}),
		fileDescriptorCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "relay_grpc_file_descriptor_count",
			Help: "gRPC file descriptor count",
		}),

		// Конфигурация
		serverStatus: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "relay_grpc_server_status",
			Help: "gRPC server status (1=up, 0=down)",
		}),
		serverVersion: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_grpc_server_version",
			Help: "gRPC server version",
		}, []string{"version"}),
		configUpdates: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_grpc_config_updates_total",
			Help: "Total number of gRPC configuration updates",
		}, []string{"update_type"}),
		configErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_grpc_config_errors_total",
			Help: "Total number of gRPC configuration errors",
		}, []string{"error_type"}),
	}

	// Регистрация метрик
	reg.MustRegister(
		m.connectionsTotal,
		m.connectionsActive,
		m.connectionsFailed,
		m.connectionsDropped,
		m.requestsTotal,
		m.requestsInFlight,
		m.requestsDuration,
		m.requestsSize,
		m.responsesSize,
		m.methodCalls,
		m.methodDuration,
		m.methodErrors,
		m.methodThroughput,
		m.streamsTotal,
		m.streamsActive,
		m.streamsCreated,
		m.streamsClosed,
		m.streamsErrors,
		m.handshakeDuration,
		m.connectionLatency,
		m.throughputBytes,
		m.throughputMessages,
		m.bandwidthUtilization,
		m.keepaliveEvents,
		m.keepaliveFailures,
		m.connectionUptime,
		m.lastActivity,
		m.serverErrors,
		m.clientErrors,
		m.timeoutEvents,
		m.reconnectionAttempts,
		m.circuitBreakerEvents,
		m.authenticationFailures,
		m.authorizationFailures,
		m.rateLimitViolations,
		m.rbacViolations,
		m.middlewareDuration,
		m.middlewareErrors,
		m.interceptorCalls,
		m.tenantRequests,
		m.tenantConnections,
		m.tenantBandwidth,
		m.tenantErrors,
		m.memoryUsageBytes,
		m.cpuUsagePercent,
		m.goroutineCount,
		m.fileDescriptorCount,
		m.serverStatus,
		m.serverVersion,
		m.configUpdates,
		m.configErrors,
	)

	return m
}

// Методы для работы с gRPC метриками

// Connection Metrics
func (m *GRPCMetrics) IncConnections() {
	m.connectionsTotal.Inc()
	m.connectionsActive.Inc()
}

func (m *GRPCMetrics) DecConnections() {
	m.connectionsActive.Dec()
}

func (m *GRPCMetrics) IncConnectionsFailed(reason, clientIP string) {
	m.connectionsFailed.WithLabelValues(reason, clientIP).Inc()
}

func (m *GRPCMetrics) IncConnectionsDropped(reason, clientIP string) {
	m.connectionsDropped.WithLabelValues(reason, clientIP).Inc()
}

// Request Metrics
func (m *GRPCMetrics) IncRequests(method, service, statusCode string) {
	m.requestsTotal.WithLabelValues(method, service, statusCode).Inc()
}

func (m *GRPCMetrics) IncRequestsInFlight(method, service string) {
	m.requestsInFlight.WithLabelValues(method, service).Inc()
}

func (m *GRPCMetrics) DecRequestsInFlight(method, service string) {
	m.requestsInFlight.WithLabelValues(method, service).Dec()
}

func (m *GRPCMetrics) ObserveRequestDuration(method, service string, duration float64) {
	m.requestsDuration.WithLabelValues(method, service).Observe(duration)
}

func (m *GRPCMetrics) ObserveRequestSize(method, service string, size float64) {
	m.requestsSize.WithLabelValues(method, service).Observe(size)
}

func (m *GRPCMetrics) ObserveResponseSize(method, service string, size float64) {
	m.responsesSize.WithLabelValues(method, service).Observe(size)
}

// Method Metrics
func (m *GRPCMetrics) IncMethodCalls(method, service, status string) {
	m.methodCalls.WithLabelValues(method, service, status).Inc()
}

func (m *GRPCMetrics) ObserveMethodDuration(method, service string, duration float64) {
	m.methodDuration.WithLabelValues(method, service).Observe(duration)
}

func (m *GRPCMetrics) IncMethodErrors(method, service, errorType string) {
	m.methodErrors.WithLabelValues(method, service, errorType).Inc()
}

func (m *GRPCMetrics) IncMethodThroughput(method, service, direction string, count int64) {
	m.methodThroughput.WithLabelValues(method, service, direction).Add(float64(count))
}

// Stream Metrics
func (m *GRPCMetrics) IncStreams() {
	m.streamsTotal.Inc()
	m.streamsActive.Inc()
}

func (m *GRPCMetrics) DecStreams() {
	m.streamsActive.Dec()
}

func (m *GRPCMetrics) IncStreamsCreated(method, service string) {
	m.streamsCreated.WithLabelValues(method, service).Inc()
}

func (m *GRPCMetrics) IncStreamsClosed(method, service, reason string) {
	m.streamsClosed.WithLabelValues(method, service, reason).Inc()
}

func (m *GRPCMetrics) IncStreamsErrors(method, service, errorType string) {
	m.streamsErrors.WithLabelValues(method, service, errorType).Inc()
}

// Performance Metrics
func (m *GRPCMetrics) ObserveHandshakeDuration(duration float64) {
	m.handshakeDuration.Observe(duration)
}

func (m *GRPCMetrics) ObserveConnectionLatency(clientIP string, duration float64) {
	m.connectionLatency.WithLabelValues(clientIP).Observe(duration)
}

func (m *GRPCMetrics) IncThroughputBytes(method, service, direction string, bytes int64) {
	m.throughputBytes.WithLabelValues(method, service, direction).Add(float64(bytes))
}

func (m *GRPCMetrics) IncThroughputMessages(method, service, direction string, messages int64) {
	m.throughputMessages.WithLabelValues(method, service, direction).Add(float64(messages))
}

func (m *GRPCMetrics) SetBandwidthUtilization(method, service string, utilization float64) {
	m.bandwidthUtilization.WithLabelValues(method, service).Set(utilization)
}

// Connection Quality Metrics
func (m *GRPCMetrics) IncKeepaliveEvents(clientIP, eventType string) {
	m.keepaliveEvents.WithLabelValues(clientIP, eventType).Inc()
}

func (m *GRPCMetrics) IncKeepaliveFailures(clientIP, failureType string) {
	m.keepaliveFailures.WithLabelValues(clientIP, failureType).Inc()
}

func (m *GRPCMetrics) SetConnectionUptime(clientIP string, uptime float64) {
	m.connectionUptime.WithLabelValues(clientIP).Set(uptime)
}

func (m *GRPCMetrics) SetLastActivity(clientIP string, timestamp float64) {
	m.lastActivity.WithLabelValues(clientIP).Set(timestamp)
}

// Error Metrics
func (m *GRPCMetrics) IncServerErrors(errorType, method, service string) {
	m.serverErrors.WithLabelValues(errorType, method, service).Inc()
}

func (m *GRPCMetrics) IncClientErrors(errorType, clientIP string) {
	m.clientErrors.WithLabelValues(errorType, clientIP).Inc()
}

func (m *GRPCMetrics) IncTimeoutEvents(timeoutType, method, service string) {
	m.timeoutEvents.WithLabelValues(timeoutType, method, service).Inc()
}

func (m *GRPCMetrics) IncReconnectionAttempts(clientIP string) {
	m.reconnectionAttempts.WithLabelValues(clientIP).Inc()
}

func (m *GRPCMetrics) IncCircuitBreakerEvents(method, service, eventType string) {
	m.circuitBreakerEvents.WithLabelValues(method, service, eventType).Inc()
}

// Security Metrics
func (m *GRPCMetrics) IncAuthenticationFailures(clientIP, failureType string) {
	m.authenticationFailures.WithLabelValues(clientIP, failureType).Inc()
}

func (m *GRPCMetrics) IncAuthorizationFailures(clientIP, method, service string) {
	m.authorizationFailures.WithLabelValues(clientIP, method, service).Inc()
}

func (m *GRPCMetrics) IncRateLimitViolations(clientIP, limitType string) {
	m.rateLimitViolations.WithLabelValues(clientIP, limitType).Inc()
}

func (m *GRPCMetrics) IncRBACViolations(clientIP, method, service, violationType string) {
	m.rbacViolations.WithLabelValues(clientIP, method, service, violationType).Inc()
}

// Middleware Metrics
func (m *GRPCMetrics) ObserveMiddlewareDuration(middleware, method, service string, duration float64) {
	m.middlewareDuration.WithLabelValues(middleware, method, service).Observe(duration)
}

func (m *GRPCMetrics) IncMiddlewareErrors(middleware, errorType string) {
	m.middlewareErrors.WithLabelValues(middleware, errorType).Inc()
}

func (m *GRPCMetrics) IncInterceptorCalls(interceptor, method, service string) {
	m.interceptorCalls.WithLabelValues(interceptor, method, service).Inc()
}

// Tenant Metrics
func (m *GRPCMetrics) IncTenantRequests(tenantID, method, service string) {
	m.tenantRequests.WithLabelValues(tenantID, method, service).Inc()
}

func (m *GRPCMetrics) SetTenantConnections(tenantID string, connections int) {
	m.tenantConnections.WithLabelValues(tenantID).Set(float64(connections))
}

func (m *GRPCMetrics) IncTenantBandwidth(tenantID, direction string, bytes int64) {
	m.tenantBandwidth.WithLabelValues(tenantID, direction).Add(float64(bytes))
}

func (m *GRPCMetrics) IncTenantErrors(tenantID, errorType string) {
	m.tenantErrors.WithLabelValues(tenantID, errorType).Inc()
}

// Resource Monitoring Metrics
func (m *GRPCMetrics) SetMemoryUsage(component string, bytes int64) {
	m.memoryUsageBytes.WithLabelValues(component).Set(float64(bytes))
}

func (m *GRPCMetrics) SetCPUUsage(component string, percent float64) {
	m.cpuUsagePercent.WithLabelValues(component).Set(percent)
}

func (m *GRPCMetrics) SetGoroutineCount(component string, count int) {
	m.goroutineCount.WithLabelValues(component).Set(float64(count))
}

func (m *GRPCMetrics) SetFileDescriptorCount(count int) {
	m.fileDescriptorCount.Set(float64(count))
}

// Configuration Metrics
func (m *GRPCMetrics) SetServerStatus(up bool) {
	value := 0.0
	if up {
		value = 1.0
	}
	m.serverStatus.Set(value)
}

func (m *GRPCMetrics) SetServerVersion(version string) {
	m.serverVersion.WithLabelValues(version).Set(1.0)
}

func (m *GRPCMetrics) IncConfigUpdates(updateType string) {
	m.configUpdates.WithLabelValues(updateType).Inc()
}

func (m *GRPCMetrics) IncConfigErrors(errorType string) {
	m.configErrors.WithLabelValues(errorType).Inc()
}



