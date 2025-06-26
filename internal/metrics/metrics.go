package metrics

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Metrics represents server metrics
// Все метрики инкапсулированы в этой структуре
// Доступ к метрикам только через экземпляр Metrics
// Регистрация метрик только в NewMetrics()
type Metrics struct {
	mu sync.RWMutex

	// Счетчики
	connectionsTotal      prometheus.Counter
	rejectedConnections   prometheus.Counter
	errorsTotal           *prometheus.CounterVec
	heartbeatsTotal       prometheus.Counter
	heartbeatErrorsTotal  prometheus.Counter

	// Гейджи
	activeConnections prometheus.Gauge
	activeTunnels     prometheus.Gauge
	cacheSize         prometheus.Gauge

	// Гистограммы
	connectionDuration      prometheus.Histogram
	heartbeatLatency        prometheus.Histogram
	tokenValidationDuration *prometheus.HistogramVec
	messageLatency          *prometheus.HistogramVec

	// CounterVec
	tokenValidations *prometheus.CounterVec
	messageErrors    *prometheus.CounterVec

	// Cache
	cacheHits   prometheus.Counter
	cacheMisses prometheus.Counter

	// Локальные счетчики для получения текущих значений
	activeConnectionsCount int64
	activeTunnelsCount     int64

	// Новые метрики для Keycloak
	keycloakTokenValidations *prometheus.CounterVec
	keycloakJWKSCacheHits    prometheus.Counter
	keycloakJWKSCacheMisses  prometheus.Counter
	keycloakJWKSCacheSize    prometheus.Gauge
	keycloakJWKSFetchDuration prometheus.Histogram
	keycloakTokenValidationDuration *prometheus.HistogramVec
	keycloakUserInfoDuration prometheus.Histogram
	keycloakPermissionChecks *prometheus.CounterVec

	// Бизнес-метрики
	businessMetrics struct {
		activeUsers       prometheus.Gauge
		revenuePerHour    prometheus.Counter
		customerSatisfaction prometheus.Gauge
	}

	// Технические метрики
	technicalMetrics struct {
		memoryUsage       prometheus.Gauge
		gcPauseDuration   prometheus.Histogram
		networkLatency    prometheus.Histogram
		diskIO            prometheus.Counter
	}
}

// NewMetrics creates new metrics
func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		connectionsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "relay_connections_total",
			Help: "Total number of connections",
		}),
		rejectedConnections: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "relay_rejected_connections_total",
			Help: "Total number of rejected connections",
		}),
		errorsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_errors_total",
			Help: "Total number of errors by code",
		}, []string{"code"}),
		heartbeatsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "relay_heartbeats_total",
			Help: "Total number of heartbeats",
		}),
		heartbeatErrorsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "relay_heartbeat_errors_total",
			Help: "Total number of heartbeat errors",
		}),
		activeConnections: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "relay_active_connections",
			Help: "Number of active connections",
		}),
		activeTunnels: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "relay_active_tunnels",
			Help: "Number of active tunnels",
		}),
		cacheSize: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "relay_token_cache_size",
			Help: "The number of tokens in the cache",
		}),
		connectionDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "relay_connection_duration_seconds",
			Help:    "Connection duration in seconds",
			Buckets: []float64{0.1, 0.5, 1.0, 5.0, 10.0, 30.0, 60.0},
		}),
		heartbeatLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "relay_heartbeat_latency_seconds",
			Help:    "Heartbeat latency in seconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1.0},
		}),
		tokenValidationDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "relay_token_validation_duration_seconds",
			Help:    "The duration of token validations in seconds",
			Buckets: prometheus.DefBuckets,
		}, []string{"method"}),
		messageLatency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "relay_message_latency_seconds",
			Help:    "The latency of message processing in seconds",
			Buckets: prometheus.DefBuckets,
		}, []string{"type"}),
		tokenValidations: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_token_validations",
			Help: "The number of token validations",
		}, []string{"method", "result"}),
		messageErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_message_errors",
			Help: "The number of message processing errors",
		}, []string{"type", "error"}),
		cacheHits: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "relay_token_cache_hits",
			Help: "The number of token cache hits",
		}),
		cacheMisses: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "relay_token_cache_misses",
			Help: "The number of token cache misses",
		}),

		// Новые метрики для Keycloak
		keycloakTokenValidations: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_keycloak_token_validations_total",
			Help: "Total number of Keycloak token validations",
		}, []string{"result", "provider"}),
		keycloakJWKSCacheHits: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "relay_keycloak_jwks_cache_hits_total",
			Help: "Total number of JWKS cache hits",
		}),
		keycloakJWKSCacheMisses: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "relay_keycloak_jwks_cache_misses_total",
			Help: "Total number of JWKS cache misses",
		}),
		keycloakJWKSCacheSize: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "relay_keycloak_jwks_cache_size",
			Help: "Number of cached JWKS keys",
		}),
		keycloakJWKSFetchDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "relay_keycloak_jwks_fetch_duration_seconds",
			Help:    "Duration of JWKS fetch operations",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1.0, 2.0, 5.0},
		}),
		keycloakTokenValidationDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "relay_keycloak_token_validation_duration_seconds",
			Help:    "Duration of Keycloak token validation",
			Buckets: prometheus.DefBuckets,
		}, []string{"method"}),
		keycloakUserInfoDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "relay_keycloak_user_info_duration_seconds",
			Help:    "Duration of user info extraction from tokens",
			Buckets: prometheus.DefBuckets,
		}),
		keycloakPermissionChecks: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_keycloak_permission_checks_total",
			Help: "Total number of permission checks",
		}, []string{"resource", "action", "result"}),

		// Бизнес-метрики
		businessMetrics: struct {
			activeUsers       prometheus.Gauge
			revenuePerHour    prometheus.Counter
			customerSatisfaction prometheus.Gauge
		}{
			activeUsers: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "relay_business_active_users",
				Help: "Number of active users",
			}),
			revenuePerHour: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "relay_business_revenue_per_hour_total",
				Help: "Total revenue per hour",
			}),
			customerSatisfaction: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "relay_business_customer_satisfaction",
				Help: "Customer satisfaction score (0-100)",
			}),
		},

		// Технические метрики
		technicalMetrics: struct {
			memoryUsage       prometheus.Gauge
			gcPauseDuration   prometheus.Histogram
			networkLatency    prometheus.Histogram
			diskIO            prometheus.Counter
		}{
			memoryUsage: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "relay_technical_memory_usage_bytes",
				Help: "Memory usage in bytes",
			}),
			gcPauseDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
				Name:    "relay_technical_gc_pause_duration_seconds",
				Help:    "GC pause duration",
				Buckets: []float64{0.001, 0.01, 0.1, 0.5, 1.0, 2.0},
			}),
			networkLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
				Name:    "relay_technical_network_latency_seconds",
				Help:    "Network latency",
				Buckets: []float64{0.001, 0.01, 0.1, 0.5, 1.0, 2.0},
			}),
			diskIO: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "relay_technical_disk_io_total",
				Help: "Total disk I/O operations",
			}),
		},
	}

	// Регистрируем все метрики через переданный реестр
	reg.MustRegister(
		m.connectionsTotal,
		m.rejectedConnections,
		m.errorsTotal,
		m.heartbeatsTotal,
		m.heartbeatErrorsTotal,
		m.activeConnections,
		m.activeTunnels,
		m.cacheSize,
		m.connectionDuration,
		m.heartbeatLatency,
		m.tokenValidationDuration,
		m.messageLatency,
		m.tokenValidations,
		m.messageErrors,
		m.cacheHits,
		m.cacheMisses,
		// Новые метрики Keycloak
		m.keycloakTokenValidations,
		m.keycloakJWKSCacheHits,
		m.keycloakJWKSCacheMisses,
		m.keycloakJWKSCacheSize,
		m.keycloakJWKSFetchDuration,
		m.keycloakTokenValidationDuration,
		m.keycloakUserInfoDuration,
		m.keycloakPermissionChecks,
		// Бизнес-метрики
		m.businessMetrics.activeUsers,
		m.businessMetrics.revenuePerHour,
		m.businessMetrics.customerSatisfaction,
		// Технические метрики
		m.technicalMetrics.memoryUsage,
		m.technicalMetrics.gcPauseDuration,
		m.technicalMetrics.networkLatency,
		m.technicalMetrics.diskIO,
	)

	return m
}

// Методы для инкрементации/наблюдения метрик
func (m *Metrics) IncConnections() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connectionsTotal.Inc()
	m.activeConnections.Inc()
	m.activeConnectionsCount++
}

func (m *Metrics) DecConnections() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.activeConnections.Dec()
	m.activeConnectionsCount--
}

func (m *Metrics) IncRejectedConnections() {
	m.rejectedConnections.Inc()
}

func (m *Metrics) IncErrors(code string) {
	m.errorsTotal.WithLabelValues(code).Inc()
}

func (m *Metrics) IncHeartbeats() {
	m.heartbeatsTotal.Inc()
}

func (m *Metrics) IncHeartbeatErrors() {
	m.heartbeatErrorsTotal.Inc()
}

func (m *Metrics) SetActiveTunnels(count int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.activeTunnels.Set(float64(count))
	m.activeTunnelsCount = int64(count)
}

func (m *Metrics) SetCacheSize(count int) {
	m.cacheSize.Set(float64(count))
}

func (m *Metrics) IncCacheHits() {
	m.cacheHits.Inc()
}

func (m *Metrics) IncCacheMisses() {
	m.cacheMisses.Inc()
}

func (m *Metrics) ObserveConnectionDuration(duration time.Duration) {
	m.connectionDuration.Observe(duration.Seconds())
}

func (m *Metrics) ObserveHeartbeatLatency(duration time.Duration) {
	m.heartbeatLatency.Observe(duration.Seconds())
}

func (m *Metrics) ObserveTokenValidationDuration(method string, duration time.Duration) {
	m.tokenValidationDuration.WithLabelValues(method).Observe(duration.Seconds())
}

func (m *Metrics) IncTokenValidations(method, result string) {
	m.tokenValidations.WithLabelValues(method, result).Inc()
}

func (m *Metrics) ObserveMessageLatency(msgType string, duration time.Duration) {
	m.messageLatency.WithLabelValues(msgType).Observe(duration.Seconds())
}

func (m *Metrics) IncMessageErrors(msgType, err string) {
	m.messageErrors.WithLabelValues(msgType, err).Inc()
}

// Новые методы для Keycloak метрик
func (m *Metrics) IncKeycloakTokenValidations(result, provider string) {
	m.keycloakTokenValidations.WithLabelValues(result, provider).Inc()
}

func (m *Metrics) IncKeycloakJWKSCacheHits() {
	m.keycloakJWKSCacheHits.Inc()
}

func (m *Metrics) IncKeycloakJWKSCacheMisses() {
	m.keycloakJWKSCacheMisses.Inc()
}

func (m *Metrics) SetKeycloakJWKSCacheSize(count int) {
	m.keycloakJWKSCacheSize.Set(float64(count))
}

func (m *Metrics) ObserveKeycloakJWKSFetchDuration(duration time.Duration) {
	m.keycloakJWKSFetchDuration.Observe(duration.Seconds())
}

func (m *Metrics) ObserveKeycloakTokenValidationDuration(method string, duration time.Duration) {
	m.keycloakTokenValidationDuration.WithLabelValues(method).Observe(duration.Seconds())
}

func (m *Metrics) ObserveKeycloakUserInfoDuration(duration time.Duration) {
	m.keycloakUserInfoDuration.Observe(duration.Seconds())
}

func (m *Metrics) IncKeycloakPermissionChecks(resource, action, result string) {
	m.keycloakPermissionChecks.WithLabelValues(resource, action, result).Inc()
}

// Методы для бизнес-метрик
func (m *Metrics) SetActiveUsers(count int) {
	m.businessMetrics.activeUsers.Set(float64(count))
}

func (m *Metrics) IncRevenuePerHour(amount float64) {
	m.businessMetrics.revenuePerHour.Add(amount)
}

func (m *Metrics) SetCustomerSatisfaction(score float64) {
	m.businessMetrics.customerSatisfaction.Set(score)
}

// Методы для технических метрик
func (m *Metrics) SetMemoryUsage(bytes int64) {
	m.technicalMetrics.memoryUsage.Set(float64(bytes))
}

func (m *Metrics) ObserveGCPauseDuration(duration time.Duration) {
	m.technicalMetrics.gcPauseDuration.Observe(duration.Seconds())
}

func (m *Metrics) ObserveNetworkLatency(duration time.Duration) {
	m.technicalMetrics.networkLatency.Observe(duration.Seconds())
}

func (m *Metrics) IncDiskIO() {
	m.technicalMetrics.diskIO.Inc()
}

// Геттеры
func (m *Metrics) GetActiveConnections() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.activeConnectionsCount
}

func (m *Metrics) GetActiveTunnels() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.activeTunnelsCount
} 