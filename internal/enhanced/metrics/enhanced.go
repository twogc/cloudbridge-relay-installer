package metrics

import (
	"context"
	_ "net/http/pprof"
	"runtime"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// EnhancedMetrics расширяет существующую систему метрик дополнительными возможностями
type EnhancedMetrics struct {
	mu sync.RWMutex

	// Build Information метрики
	buildInfo *prometheus.GaugeVec

	// Protocol Performance метрики
	protocolLatency  *prometheus.HistogramVec
	protocolErrors   *prometheus.CounterVec
	protocolSwitches *prometheus.CounterVec
	protocolSuccess  *prometheus.CounterVec

	// Health Check метрики
	healthCheckStatus   *prometheus.GaugeVec
	healthCheckDuration *prometheus.HistogramVec

	// Connection Pool метрики
	connectionPoolSize   *prometheus.GaugeVec
	connectionPoolIdle   *prometheus.GaugeVec
	connectionPoolActive *prometheus.GaugeVec

	// QUIC специфичные метрики
	quicStreamsTotal     *prometheus.CounterVec
	quicStreamsActive    *prometheus.GaugeVec
	quicStreamLatency    *prometheus.HistogramVec
	quicConnectionErrors *prometheus.CounterVec

	// Ingress метрики
	ingressRequestsTotal   *prometheus.CounterVec
	ingressRequestDuration *prometheus.HistogramVec
	ingressErrorsTotal     *prometheus.CounterVec
	ingressBandwidth       *prometheus.CounterVec

	// System метрики
	systemMemoryUsage prometheus.Gauge
	systemCPUUsage    prometheus.Gauge
	systemGoroutines  prometheus.Gauge
	systemHeapAlloc   prometheus.Gauge
	systemHeapSys     prometheus.Gauge

	// Diagnostic метрики
	diagnosticRequestsTotal *prometheus.CounterVec
	diagnosticErrorsTotal   *prometheus.CounterVec
}

// NewEnhancedMetrics создает новую расширенную систему метрик
func NewEnhancedMetrics(reg prometheus.Registerer) *EnhancedMetrics {
	em := &EnhancedMetrics{
		// Build Information
		buildInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "build_info",
				Help: "Build and version information",
			},
			[]string{"goversion", "type", "revision", "version"},
		),

		// Protocol Performance
		protocolLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "protocol_latency_seconds",
				Help:    "Protocol latency in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 5.0},
			},
			[]string{"protocol", "operation"},
		),
		protocolErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "protocol_errors_total",
				Help: "Total number of protocol errors",
			},
			[]string{"protocol", "error_type"},
		),
		protocolSwitches: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "protocol_switches_total",
				Help: "Total number of protocol switches",
			},
			[]string{"from_protocol", "to_protocol", "reason"},
		),
		protocolSuccess: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "protocol_success_total",
				Help: "Total number of successful protocol operations",
			},
			[]string{"protocol", "operation"},
		),

		// Health Check
		healthCheckStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "health_check_status",
				Help: "Health check status (1 = healthy, 0 = unhealthy)",
			},
			[]string{"check", "component"},
		),
		healthCheckDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "health_check_duration_seconds",
				Help:    "Health check duration in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0},
			},
			[]string{"check", "component"},
		),

		// Connection Pool
		connectionPoolSize: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "connection_pool_size",
				Help: "Total size of connection pool",
			},
			[]string{"pool", "protocol"},
		),
		connectionPoolIdle: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "connection_pool_idle",
				Help: "Number of idle connections in pool",
			},
			[]string{"pool", "protocol"},
		),
		connectionPoolActive: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "connection_pool_active",
				Help: "Number of active connections in pool",
			},
			[]string{"pool", "protocol"},
		),

		// QUIC метрики
		quicStreamsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "quic_streams_total",
				Help: "Total number of QUIC streams",
			},
			[]string{"direction", "type"},
		),
		quicStreamsActive: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "quic_streams_active",
				Help: "Number of active QUIC streams",
			},
			[]string{"direction", "type"},
		),
		quicStreamLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "quic_stream_latency_seconds",
				Help:    "QUIC stream latency in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0},
			},
			[]string{"direction", "type"},
		),
		quicConnectionErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "quic_connection_errors_total",
				Help: "Total number of QUIC connection errors",
			},
			[]string{"error_type"},
		),

		// Ingress метрики
		ingressRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ingress_requests_total",
				Help: "Total number of ingress requests",
			},
			[]string{"host", "path", "method", "status_code"},
		),
		ingressRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "ingress_request_duration_seconds",
				Help:    "Ingress request duration in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 5.0},
			},
			[]string{"host", "path", "method"},
		),
		ingressErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ingress_errors_total",
				Help: "Total number of ingress errors",
			},
			[]string{"host", "error_type"},
		),
		ingressBandwidth: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ingress_bandwidth_bytes_total",
				Help: "Total bandwidth usage in bytes",
			},
			[]string{"host", "direction"},
		),

		// System метрики
		systemMemoryUsage: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "system_memory_usage_bytes",
				Help: "System memory usage in bytes",
			},
		),
		systemCPUUsage: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "system_cpu_usage_percent",
				Help: "System CPU usage percentage",
			},
		),
		systemGoroutines: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "system_goroutines",
				Help: "Number of active goroutines",
			},
		),
		systemHeapAlloc: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "system_heap_alloc_bytes",
				Help: "Heap memory allocated in bytes",
			},
		),
		systemHeapSys: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "system_heap_sys_bytes",
				Help: "Heap memory obtained from system in bytes",
			},
		),

		// Diagnostic метрики
		diagnosticRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "diagnostic_requests_total",
				Help: "Total number of diagnostic requests",
			},
			[]string{"endpoint", "method"},
		),
		diagnosticErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "diagnostic_errors_total",
				Help: "Total number of diagnostic errors",
			},
			[]string{"endpoint", "error_type"},
		),
	}

	// Регистрируем все метрики
	reg.MustRegister(
		em.buildInfo,
		em.protocolLatency,
		em.protocolErrors,
		em.protocolSwitches,
		em.protocolSuccess,
		em.healthCheckStatus,
		em.healthCheckDuration,
		em.connectionPoolSize,
		em.connectionPoolIdle,
		em.connectionPoolActive,
		em.quicStreamsTotal,
		em.quicStreamsActive,
		em.quicStreamLatency,
		em.quicConnectionErrors,
		em.ingressRequestsTotal,
		em.ingressRequestDuration,
		em.ingressErrorsTotal,
		em.ingressBandwidth,
		em.systemMemoryUsage,
		em.systemCPUUsage,
		em.systemGoroutines,
		em.systemHeapAlloc,
		em.systemHeapSys,
		em.diagnosticRequestsTotal,
		em.diagnosticErrorsTotal,
	)

	return em
}

// RegisterBuildInfo регистрирует информацию о сборке
func (em *EnhancedMetrics) RegisterBuildInfo(buildType, buildTime, version string) {
	em.buildInfo.WithLabelValues(runtime.Version(), buildType, buildTime, version).Set(1)
}

// ObserveProtocolLatency наблюдает задержку протокола
func (em *EnhancedMetrics) ObserveProtocolLatency(protocol, operation string, duration time.Duration) {
	em.protocolLatency.WithLabelValues(protocol, operation).Observe(duration.Seconds())
}

// IncProtocolErrors увеличивает счетчик ошибок протокола
func (em *EnhancedMetrics) IncProtocolErrors(protocol, errorType string) {
	em.protocolErrors.WithLabelValues(protocol, errorType).Inc()
}

// IncProtocolSwitches увеличивает счетчик переключений протокола
func (em *EnhancedMetrics) IncProtocolSwitches(fromProtocol, toProtocol, reason string) {
	em.protocolSwitches.WithLabelValues(fromProtocol, toProtocol, reason).Inc()
}

// IncProtocolSuccess увеличивает счетчик успешных операций протокола
func (em *EnhancedMetrics) IncProtocolSuccess(protocol, operation string) {
	em.protocolSuccess.WithLabelValues(protocol, operation).Inc()
}

// SetHealthCheckStatus устанавливает статус health check
func (em *EnhancedMetrics) SetHealthCheckStatus(check, component string, healthy bool) {
	status := 0.0
	if healthy {
		status = 1.0
	}
	em.healthCheckStatus.WithLabelValues(check, component).Set(status)
}

// ObserveHealthCheckDuration наблюдает длительность health check
func (em *EnhancedMetrics) ObserveHealthCheckDuration(check, component string, duration time.Duration) {
	em.healthCheckDuration.WithLabelValues(check, component).Observe(duration.Seconds())
}

// SetConnectionPoolSize устанавливает размер пула соединений
func (em *EnhancedMetrics) SetConnectionPoolSize(pool, protocol string, size int) {
	em.connectionPoolSize.WithLabelValues(pool, protocol).Set(float64(size))
}

// SetConnectionPoolIdle устанавливает количество неактивных соединений
func (em *EnhancedMetrics) SetConnectionPoolIdle(pool, protocol string, idle int) {
	em.connectionPoolIdle.WithLabelValues(pool, protocol).Set(float64(idle))
}

// SetConnectionPoolActive устанавливает количество активных соединений
func (em *EnhancedMetrics) SetConnectionPoolActive(pool, protocol string, active int) {
	em.connectionPoolActive.WithLabelValues(pool, protocol).Set(float64(active))
}

// QUIC методы
func (em *EnhancedMetrics) IncQUICStreams(direction, streamType string) {
	em.quicStreamsTotal.WithLabelValues(direction, streamType).Inc()
}

func (em *EnhancedMetrics) SetQUICStreamsActive(direction, streamType string, count int) {
	em.quicStreamsActive.WithLabelValues(direction, streamType).Set(float64(count))
}

func (em *EnhancedMetrics) ObserveQUICStreamLatency(direction, streamType string, duration time.Duration) {
	em.quicStreamLatency.WithLabelValues(direction, streamType).Observe(duration.Seconds())
}

func (em *EnhancedMetrics) IncQUICConnectionErrors(errorType string) {
	em.quicConnectionErrors.WithLabelValues(errorType).Inc()
}

// Ingress методы
func (em *EnhancedMetrics) IncIngressRequests(host, path, method, statusCode string) {
	em.ingressRequestsTotal.WithLabelValues(host, path, method, statusCode).Inc()
}

func (em *EnhancedMetrics) ObserveIngressRequestDuration(host, path, method string, duration time.Duration) {
	em.ingressRequestDuration.WithLabelValues(host, path, method).Observe(duration.Seconds())
}

func (em *EnhancedMetrics) IncIngressErrors(host, errorType string) {
	em.ingressErrorsTotal.WithLabelValues(host, errorType).Inc()
}

func (em *EnhancedMetrics) IncIngressBandwidth(host, direction string, bytes int64) {
	em.ingressBandwidth.WithLabelValues(host, direction).Add(float64(bytes))
}

// System методы
func (em *EnhancedMetrics) SetSystemMemoryUsage(bytes int64) {
	em.systemMemoryUsage.Set(float64(bytes))
}

func (em *EnhancedMetrics) SetSystemCPUUsage(percent float64) {
	em.systemCPUUsage.Set(percent)
}

func (em *EnhancedMetrics) SetSystemGoroutines(count int) {
	em.systemGoroutines.Set(float64(count))
}

func (em *EnhancedMetrics) SetSystemHeapAlloc(bytes uint64) {
	em.systemHeapAlloc.Set(float64(bytes))
}

func (em *EnhancedMetrics) SetSystemHeapSys(bytes uint64) {
	em.systemHeapSys.Set(float64(bytes))
}

// Diagnostic методы
func (em *EnhancedMetrics) IncDiagnosticRequests(endpoint, method string) {
	em.diagnosticRequestsTotal.WithLabelValues(endpoint, method).Inc()
}

func (em *EnhancedMetrics) IncDiagnosticErrors(endpoint, errorType string) {
	em.diagnosticErrorsTotal.WithLabelValues(endpoint, errorType).Inc()
}

// UpdateSystemMetrics обновляет системные метрики
func (em *EnhancedMetrics) UpdateSystemMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	em.SetSystemMemoryUsage(int64(m.Alloc))
	em.SetSystemHeapAlloc(m.HeapAlloc)
	em.SetSystemHeapSys(m.HeapSys)
	em.SetSystemGoroutines(runtime.NumGoroutine())
}

// StartSystemMetricsCollector запускает сборщик системных метрик
func (em *EnhancedMetrics) StartSystemMetricsCollector(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			em.UpdateSystemMetrics()
		}
	}
}
