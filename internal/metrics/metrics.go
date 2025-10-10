package metrics

import (
	"fmt"
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
	connectionsTotal     prometheus.Counter
	rejectedConnections  prometheus.Counter
	errorsTotal          *prometheus.CounterVec
	heartbeatsTotal      prometheus.Counter
	heartbeatErrorsTotal prometheus.Counter

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

	// Бизнес-метрики
	businessMetrics struct {
		activeUsers          prometheus.Gauge
		revenuePerHour       prometheus.Counter
		customerSatisfaction prometheus.Gauge
		totalRevenue         prometheus.Gauge
	}

	// Технические метрики
	technicalMetrics struct {
		memoryUsage     prometheus.Gauge
		gcPauseDuration prometheus.Histogram
		networkLatency  prometheus.Histogram
		diskIO          prometheus.Counter
	}

	// Новые метрики для туннелей
	tunnelMetrics struct {
		creationsTotal         prometheus.Counter
		closuresTotal          prometheus.Counter
		durationSeconds        prometheus.Histogram
		bytesFromClientTotal   *prometheus.CounterVec
		bytesToClientTotal     *prometheus.CounterVec
		bandwidthUtilization   *prometheus.GaugeVec
		errorsTotal            *prometheus.CounterVec
		status                 *prometheus.GaugeVec
		activityHeatmap        *prometheus.GaugeVec
		latencySeconds         *prometheus.HistogramVec
		operationsTotal        *prometheus.CounterVec
		cleanupOperationsTotal *prometheus.CounterVec
		memoryUsageBytes       *prometheus.GaugeVec
		cpuUsageSecondsTotal   *prometheus.CounterVec
	}

	// Метрики безопасности
	securityMetrics struct {
		suspiciousActivitiesTotal    prometheus.Counter
		rateLimitViolationsTotal     *prometheus.CounterVec
		ddosAttemptsTotal            *prometheus.CounterVec
		securityEventsTotal          *prometheus.CounterVec
		failedLoginsByIP             *prometheus.GaugeVec
		tlsCertificateStatus         *prometheus.GaugeVec
		tlsHandshakeFailuresTotal    *prometheus.CounterVec
		tlsCertificateExpirationDays *prometheus.GaugeVec
	}

	// ML метрики
	mlMetrics struct {
		anomalyScore           *prometheus.GaugeVec
		threatPredictionsTotal *prometheus.CounterVec
	}

	// Метрики мультитенантности
	tenantMetrics struct {
		activeCount                prometheus.Gauge
		activeTunnels              *prometheus.GaugeVec
		activeConnections          *prometheus.GaugeVec
		bandwidthBytesTotal        *prometheus.CounterVec
		dataTransferGB             *prometheus.GaugeVec
		responseTimeSeconds        *prometheus.HistogramVec
		errorsTotal                *prometheus.CounterVec
		slaCompliancePercent       *prometheus.GaugeVec
		uptimePercent              *prometheus.GaugeVec
		securityIncidentsTotal     *prometheus.CounterVec
		authFailuresTotal          *prometheus.CounterVec
		riskScore                  *prometheus.GaugeVec
		isolationViolationsTotal   *prometheus.CounterVec
		accessPatternsTotal        *prometheus.CounterVec
		revenueTotal               *prometheus.CounterVec
		customerSatisfaction       *prometheus.GaugeVec
		dailyActiveUsers           *prometheus.GaugeVec
		usageHeatmap               *prometheus.GaugeVec
		comparisonMetrics          *prometheus.GaugeVec
		resourceUtilizationPercent prometheus.Gauge
	}

	// P2P метрики
	p2pMetrics *P2PMetrics

	// QUIC метрики
	quicMetrics *QUICMetrics

	// WireGuard метрики
	wireguardMetrics *WireGuardMetrics

	// gRPC метрики
	grpcMetrics *GRPCMetrics

	// AI Service метрики
	aiServiceMetrics *AIServiceMetrics

	// Cadence метрики
	cadenceMetrics *CadenceMetrics
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

		// Бизнес-метрики
		businessMetrics: struct {
			activeUsers          prometheus.Gauge
			revenuePerHour       prometheus.Counter
			customerSatisfaction prometheus.Gauge
			totalRevenue         prometheus.Gauge
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
			totalRevenue: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "relay_business_total_revenue",
				Help: "Total revenue",
			}),
		},

		// Технические метрики
		technicalMetrics: struct {
			memoryUsage     prometheus.Gauge
			gcPauseDuration prometheus.Histogram
			networkLatency  prometheus.Histogram
			diskIO          prometheus.Counter
		}{
			memoryUsage: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "relay_technical_memory_usage_bytes",
				Help: "Memory usage in bytes",
			}),
			gcPauseDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
				Name:    "relay_technical_gc_pause_duration_seconds",
				Help:    "GC pause duration in seconds",
				Buckets: prometheus.DefBuckets,
			}),
			networkLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
				Name:    "relay_technical_network_latency_seconds",
				Help:    "Network latency in seconds",
				Buckets: prometheus.DefBuckets,
			}),
			diskIO: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "relay_technical_disk_io_total",
				Help: "Total disk I/O operations",
			}),
		},

		// Метрики туннелей
		tunnelMetrics: struct {
			creationsTotal         prometheus.Counter
			closuresTotal          prometheus.Counter
			durationSeconds        prometheus.Histogram
			bytesFromClientTotal   *prometheus.CounterVec
			bytesToClientTotal     *prometheus.CounterVec
			bandwidthUtilization   *prometheus.GaugeVec
			errorsTotal            *prometheus.CounterVec
			status                 *prometheus.GaugeVec
			activityHeatmap        *prometheus.GaugeVec
			latencySeconds         *prometheus.HistogramVec
			operationsTotal        *prometheus.CounterVec
			cleanupOperationsTotal *prometheus.CounterVec
			memoryUsageBytes       *prometheus.GaugeVec
			cpuUsageSecondsTotal   *prometheus.CounterVec
		}{
			creationsTotal: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "relay_tunnel_creations_total",
				Help: "Total number of tunnel creations",
			}),
			closuresTotal: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "relay_tunnel_closures_total",
				Help: "Total number of tunnel closures",
			}),
			durationSeconds: prometheus.NewHistogram(prometheus.HistogramOpts{
				Name:    "relay_tunnel_duration_seconds",
				Help:    "Tunnel duration in seconds",
				Buckets: []float64{1, 5, 10, 30, 60, 300, 600, 1800, 3600},
			}),
			bytesFromClientTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "relay_tunnel_bytes_from_client_total",
				Help: "Total bytes transferred from client",
			}, []string{"tunnel_id", "remote_host", "remote_port"}),
			bytesToClientTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "relay_tunnel_bytes_to_client_total",
				Help: "Total bytes transferred to client",
			}, []string{"tunnel_id", "remote_host", "remote_port"}),
			bandwidthUtilization: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "relay_tunnel_bandwidth_utilization_percent",
				Help: "Tunnel bandwidth utilization percentage",
			}, []string{"tunnel_id"}),
			errorsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "relay_tunnel_errors_total",
				Help: "Total tunnel errors",
			}, []string{"tunnel_id", "error_type"}),
			status: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "relay_tunnel_status",
				Help: "Tunnel status (1=active, 0=inactive)",
			}, []string{"tunnel_id", "remote_host", "remote_port"}),
			activityHeatmap: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "relay_tunnel_activity_heatmap",
				Help: "Tunnel activity heatmap",
			}, []string{"tunnel_id", "hour", "day"}),
			latencySeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Name:    "relay_tunnel_latency_seconds",
				Help:    "Tunnel latency in seconds",
				Buckets: prometheus.DefBuckets,
			}, []string{"tunnel_id"}),
			operationsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "relay_tunnel_operations_total",
				Help: "Total tunnel operations",
			}, []string{"operation"}),
			cleanupOperationsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "relay_tunnel_cleanup_operations_total",
				Help: "Total tunnel cleanup operations",
			}, []string{"reason"}),
			memoryUsageBytes: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "relay_tunnel_memory_usage_bytes",
				Help: "Tunnel memory usage in bytes",
			}, []string{"tunnel_id"}),
			cpuUsageSecondsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "relay_tunnel_cpu_usage_seconds_total",
				Help: "Tunnel CPU usage in seconds",
			}, []string{"tunnel_id"}),
		},

		// Метрики безопасности
		securityMetrics: struct {
			suspiciousActivitiesTotal    prometheus.Counter
			rateLimitViolationsTotal     *prometheus.CounterVec
			ddosAttemptsTotal            *prometheus.CounterVec
			securityEventsTotal          *prometheus.CounterVec
			failedLoginsByIP             *prometheus.GaugeVec
			tlsCertificateStatus         *prometheus.GaugeVec
			tlsHandshakeFailuresTotal    *prometheus.CounterVec
			tlsCertificateExpirationDays *prometheus.GaugeVec
		}{
			suspiciousActivitiesTotal: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "relay_security_suspicious_activities_total",
				Help: "Total suspicious activities detected",
			}),
			rateLimitViolationsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "relay_security_rate_limit_violations_total",
				Help: "Total rate limit violations",
			}, []string{"client_ip", "limit_type"}),
			ddosAttemptsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "relay_security_ddos_attempts_total",
				Help: "Total DDoS attack attempts",
			}, []string{"attack_type", "source_ip"}),
			securityEventsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "relay_security_events_total",
				Help: "Total security events",
			}, []string{"event_type", "severity"}),
			failedLoginsByIP: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "relay_security_failed_logins_by_ip",
				Help: "Failed login attempts by IP",
			}, []string{"client_ip"}),
			tlsCertificateStatus: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "relay_tls_certificate_status",
				Help: "TLS certificate status (1=valid, 0=invalid)",
			}, []string{"cert_type"}),
			tlsHandshakeFailuresTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "relay_tls_handshake_failures_total",
				Help: "Total TLS handshake failures",
			}, []string{"failure_reason"}),
			tlsCertificateExpirationDays: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "relay_tls_certificate_expiration_days",
				Help: "Days until certificate expiration",
			}, []string{"cert_type"}),
		},

		// ML метрики
		mlMetrics: struct {
			anomalyScore           *prometheus.GaugeVec
			threatPredictionsTotal *prometheus.CounterVec
		}{
			anomalyScore: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "relay_ml_anomaly_score",
				Help: "ML anomaly score",
			}, []string{"user_id", "feature"}),
			threatPredictionsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "relay_ml_threat_predictions_total",
				Help: "Total ML threat predictions",
			}, []string{"prediction_type"}),
		},

		// Метрики мультитенантности - инициализация
		tenantMetrics: struct {
			activeCount                prometheus.Gauge
			activeTunnels              *prometheus.GaugeVec
			activeConnections          *prometheus.GaugeVec
			bandwidthBytesTotal        *prometheus.CounterVec
			dataTransferGB             *prometheus.GaugeVec
			responseTimeSeconds        *prometheus.HistogramVec
			errorsTotal                *prometheus.CounterVec
			slaCompliancePercent       *prometheus.GaugeVec
			uptimePercent              *prometheus.GaugeVec
			securityIncidentsTotal     *prometheus.CounterVec
			authFailuresTotal          *prometheus.CounterVec
			riskScore                  *prometheus.GaugeVec
			isolationViolationsTotal   *prometheus.CounterVec
			accessPatternsTotal        *prometheus.CounterVec
			revenueTotal               *prometheus.CounterVec
			customerSatisfaction       *prometheus.GaugeVec
			dailyActiveUsers           *prometheus.GaugeVec
			usageHeatmap               *prometheus.GaugeVec
			comparisonMetrics          *prometheus.GaugeVec
			resourceUtilizationPercent prometheus.Gauge
		}{
			activeCount: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "relay_tenant_active_count",
				Help: "Number of active tenants",
			}),
			activeTunnels: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "relay_tenant_active_tunnels",
				Help: "Active tunnels per tenant",
			}, []string{"tenant_id", "tenant_name"}),
			activeConnections: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "relay_tenant_active_connections",
				Help: "Active connections per tenant",
			}, []string{"tenant_id", "tenant_name"}),
			bandwidthBytesTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "relay_tenant_bandwidth_bytes_total",
				Help: "Total bandwidth usage per tenant",
			}, []string{"tenant_id", "tenant_name"}),
			dataTransferGB: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "relay_tenant_data_transfer_gb",
				Help: "Data transfer in GB per tenant",
			}, []string{"tenant_id", "tenant_name"}),
			responseTimeSeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Name:    "relay_tenant_response_time_seconds",
				Help:    "Response time per tenant",
				Buckets: prometheus.DefBuckets,
			}, []string{"tenant_id"}),
			errorsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "relay_tenant_errors_total",
				Help: "Total errors per tenant",
			}, []string{"tenant_id", "error_type"}),
			slaCompliancePercent: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "relay_tenant_sla_compliance_percent",
				Help: "SLA compliance percentage per tenant",
			}, []string{"tenant_id", "sla_type"}),
			uptimePercent: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "relay_tenant_uptime_percent",
				Help: "Uptime percentage per tenant",
			}, []string{"tenant_id", "tenant_name"}),
			securityIncidentsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "relay_tenant_security_incidents_total",
				Help: "Security incidents per tenant",
			}, []string{"tenant_id", "incident_type"}),
			authFailuresTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "relay_tenant_auth_failures_total",
				Help: "Authentication failures per tenant",
			}, []string{"tenant_id", "failure_reason"}),
			riskScore: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "relay_tenant_risk_score",
				Help: "Risk score per tenant",
			}, []string{"tenant_id", "tenant_name"}),
			isolationViolationsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "relay_tenant_isolation_violations_total",
				Help: "Isolation violations per tenant",
			}, []string{"tenant_id", "violation_type"}),
			accessPatternsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "relay_tenant_access_patterns_total",
				Help: "Access patterns per tenant",
			}, []string{"tenant_id", "pattern_type"}),
			revenueTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "relay_tenant_revenue_total",
				Help: "Revenue per tenant",
			}, []string{"tenant_id", "tenant_name"}),
			customerSatisfaction: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "relay_tenant_customer_satisfaction",
				Help: "Customer satisfaction per tenant",
			}, []string{"tenant_id", "tenant_name"}),
			dailyActiveUsers: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "relay_tenant_daily_active_users",
				Help: "Daily active users per tenant",
			}, []string{"tenant_id", "tenant_name"}),
			usageHeatmap: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "relay_tenant_usage_heatmap",
				Help: "Usage heatmap per tenant",
			}, []string{"tenant_id", "hour", "day"}),
			comparisonMetrics: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "relay_tenant_comparison_metric",
				Help: "Comparison metric per tenant",
			}, []string{"tenant_id", "tenant_name", "metric"}),
			resourceUtilizationPercent: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "relay_tenant_resource_utilization_percent",
				Help: "Overall resource utilization percentage",
			}),
		},
		p2pMetrics:       NewP2PMetrics(reg),
		quicMetrics:      NewQUICMetrics(reg),
		wireguardMetrics: NewWireGuardMetrics(reg),
		grpcMetrics:      NewGRPCMetrics(reg),
		aiServiceMetrics: NewAIServiceMetrics(reg),
		cadenceMetrics:   NewCadenceMetrics(nil),
	}

	// Регистрация только основных метрик (подметрики уже зарегистрированы в своих конструкторах)
	// Убираем регистрацию всех метрик, так как они уже зарегистрированы в подметриках
	// reg.MustRegister(...)

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

// Tunnel Metrics Methods
func (m *Metrics) IncTunnelCreations() {
	m.tunnelMetrics.creationsTotal.Inc()
}

func (m *Metrics) IncTunnelClosures() {
	m.tunnelMetrics.closuresTotal.Inc()
}

func (m *Metrics) ObserveTunnelDuration(duration time.Duration) {
	m.tunnelMetrics.durationSeconds.Observe(duration.Seconds())
}

func (m *Metrics) IncTunnelBytesFromClient(tunnelID, remoteHost string, remotePort int, bytes int64) {
	m.tunnelMetrics.bytesFromClientTotal.WithLabelValues(tunnelID, remoteHost, fmt.Sprintf("%d", remotePort)).Add(float64(bytes))
}

func (m *Metrics) IncTunnelBytesToClient(tunnelID, remoteHost string, remotePort int, bytes int64) {
	m.tunnelMetrics.bytesToClientTotal.WithLabelValues(tunnelID, remoteHost, fmt.Sprintf("%d", remotePort)).Add(float64(bytes))
}

func (m *Metrics) SetTunnelBandwidthUtilization(tunnelID string, utilization float64) {
	m.tunnelMetrics.bandwidthUtilization.WithLabelValues(tunnelID).Set(utilization)
}

func (m *Metrics) IncTunnelErrors(tunnelID, errorType string) {
	m.tunnelMetrics.errorsTotal.WithLabelValues(tunnelID, errorType).Inc()
}

func (m *Metrics) SetTunnelStatus(tunnelID, remoteHost string, remotePort int, active bool) {
	v := 0.0
	if active {
		v = 1.0
	}
	m.tunnelMetrics.status.WithLabelValues(tunnelID, remoteHost, fmt.Sprintf("%d", remotePort)).Set(v)
}

func (m *Metrics) SetTunnelActivityHeatmap(tunnelID, hour, day string, value float64) {
	m.tunnelMetrics.activityHeatmap.WithLabelValues(tunnelID, hour, day).Set(value)
}

func (m *Metrics) ObserveTunnelLatency(tunnelID string, duration time.Duration) {
	m.tunnelMetrics.latencySeconds.WithLabelValues(tunnelID).Observe(duration.Seconds())
}

func (m *Metrics) IncTunnelOperations(operation string) {
	m.tunnelMetrics.operationsTotal.WithLabelValues(operation).Inc()
}

func (m *Metrics) IncTunnelCleanupOperations(reason string) {
	m.tunnelMetrics.cleanupOperationsTotal.WithLabelValues(reason).Inc()
}

func (m *Metrics) SetTunnelMemoryUsage(tunnelID string, bytes int64) {
	m.tunnelMetrics.memoryUsageBytes.WithLabelValues(tunnelID).Set(float64(bytes))
}

func (m *Metrics) IncTunnelCPUUsage(tunnelID string, seconds float64) {
	m.tunnelMetrics.cpuUsageSecondsTotal.WithLabelValues(tunnelID).Add(seconds)
}

// Security Metrics Methods
func (m *Metrics) IncSuspiciousActivities() {
	m.securityMetrics.suspiciousActivitiesTotal.Inc()
}

func (m *Metrics) IncRateLimitViolations(clientIP, limitType string) {
	m.securityMetrics.rateLimitViolationsTotal.WithLabelValues(clientIP, limitType).Inc()
}

func (m *Metrics) IncDDoSAttempts(attackType, sourceIP string) {
	m.securityMetrics.ddosAttemptsTotal.WithLabelValues(attackType, sourceIP).Inc()
}

func (m *Metrics) IncSecurityEvents(eventType, severity string) {
	m.securityMetrics.securityEventsTotal.WithLabelValues(eventType, severity).Inc()
}

func (m *Metrics) SetFailedLoginsByIP(clientIP string, failedAttempts int) {
	m.securityMetrics.failedLoginsByIP.WithLabelValues(clientIP).Set(float64(failedAttempts))
}

func (m *Metrics) SetTLSCertificateStatus(certType string, valid bool) {
	value := 0.0
	if valid {
		value = 1.0
	}
	m.securityMetrics.tlsCertificateStatus.WithLabelValues(certType).Set(value)
}

func (m *Metrics) IncTLSHandshakeFailures(failureReason string) {
	m.securityMetrics.tlsHandshakeFailuresTotal.WithLabelValues(failureReason).Inc()
}

func (m *Metrics) SetTLSCertificateExpirationDays(certType string, days float64) {
	m.securityMetrics.tlsCertificateExpirationDays.WithLabelValues(certType).Set(days)
}

// ML Metrics Methods
func (m *Metrics) SetMLAnomalyScore(userID, feature string, score float64) {
	m.mlMetrics.anomalyScore.WithLabelValues(userID, feature).Set(score)
}

func (m *Metrics) IncMLThreatPredictions(predictionType string) {
	m.mlMetrics.threatPredictionsTotal.WithLabelValues(predictionType).Inc()
}

// Tenant Metrics Methods
func (m *Metrics) SetActiveTenantCount(count int) {
	m.tenantMetrics.activeCount.Set(float64(count))
}

func (m *Metrics) SetTenantActiveTunnels(tenantID, tenantName string, count int) {
	m.tenantMetrics.activeTunnels.WithLabelValues(tenantID, tenantName).Set(float64(count))
}

func (m *Metrics) SetTenantActiveConnections(tenantID, tenantName string, count int) {
	m.tenantMetrics.activeConnections.WithLabelValues(tenantID, tenantName).Set(float64(count))
}

func (m *Metrics) IncTenantBandwidthBytes(tenantID, tenantName string, bytes int64) {
	m.tenantMetrics.bandwidthBytesTotal.WithLabelValues(tenantID, tenantName).Add(float64(bytes))
}

func (m *Metrics) SetTenantDataTransferGB(tenantID, tenantName string, gb float64) {
	m.tenantMetrics.dataTransferGB.WithLabelValues(tenantID, tenantName).Set(gb)
}

func (m *Metrics) ObserveTenantResponseTime(tenantID string, duration time.Duration) {
	m.tenantMetrics.responseTimeSeconds.WithLabelValues(tenantID).Observe(duration.Seconds())
}

func (m *Metrics) IncTenantErrors(tenantID, errorType string) {
	m.tenantMetrics.errorsTotal.WithLabelValues(tenantID, errorType).Inc()
}

func (m *Metrics) SetTenantSLACompliance(tenantID, slaType string, compliance float64) {
	m.tenantMetrics.slaCompliancePercent.WithLabelValues(tenantID, slaType).Set(compliance)
}

func (m *Metrics) SetTenantUptime(tenantID, tenantName string, uptime float64) {
	m.tenantMetrics.uptimePercent.WithLabelValues(tenantID, tenantName).Set(uptime)
}

func (m *Metrics) IncTenantSecurityIncidents(tenantID, incidentType string) {
	m.tenantMetrics.securityIncidentsTotal.WithLabelValues(tenantID, incidentType).Inc()
}

func (m *Metrics) IncTenantAuthFailures(tenantID, failureReason string) {
	m.tenantMetrics.authFailuresTotal.WithLabelValues(tenantID, failureReason).Inc()
}

func (m *Metrics) SetTenantRiskScore(tenantID, tenantName string, score float64) {
	m.tenantMetrics.riskScore.WithLabelValues(tenantID, tenantName).Set(score)
}

func (m *Metrics) IncTenantIsolationViolations(tenantID, violationType string) {
	m.tenantMetrics.isolationViolationsTotal.WithLabelValues(tenantID, violationType).Inc()
}

func (m *Metrics) IncTenantAccessPatterns(tenantID, patternType string) {
	m.tenantMetrics.accessPatternsTotal.WithLabelValues(tenantID, patternType).Inc()
}

func (m *Metrics) IncTenantRevenue(tenantID, tenantName string, amount float64) {
	m.tenantMetrics.revenueTotal.WithLabelValues(tenantID, tenantName).Add(amount)
}

func (m *Metrics) SetTenantCustomerSatisfaction(tenantID, tenantName string, satisfaction float64) {
	m.tenantMetrics.customerSatisfaction.WithLabelValues(tenantID, tenantName).Set(satisfaction)
}

func (m *Metrics) SetTenantDailyActiveUsers(tenantID, tenantName string, users int) {
	m.tenantMetrics.dailyActiveUsers.WithLabelValues(tenantID, tenantName).Set(float64(users))
}

func (m *Metrics) SetTenantUsageHeatmap(tenantID, hour, day string, value float64) {
	m.tenantMetrics.usageHeatmap.WithLabelValues(tenantID, hour, day).Set(value)
}

func (m *Metrics) SetTenantComparisonMetric(tenantID, tenantName, metric string, value float64) {
	m.tenantMetrics.comparisonMetrics.WithLabelValues(tenantID, tenantName, metric).Set(value)
}

func (m *Metrics) SetTenantResourceUtilization(utilization float64) {
	m.tenantMetrics.resourceUtilizationPercent.Set(utilization)
}

// Business Metrics Methods
func (m *Metrics) SetTotalRevenue(revenue float64) {
	m.businessMetrics.totalRevenue.Set(revenue)
}

// P2P Metrics Methods
func (m *Metrics) GetP2PMetrics() *P2PMetrics {
	return m.p2pMetrics
}

func (m *Metrics) IncP2PConnections() {
	m.p2pMetrics.IncP2PConnections()
}

func (m *Metrics) SetActivePeers(count int) {
	m.p2pMetrics.SetActivePeers(count)
}

func (m *Metrics) IncP2PTraffic(direction, peerID string, bytes int64) {
	m.p2pMetrics.IncP2PTraffic(direction, peerID, bytes)
}

func (m *Metrics) ObserveP2PLatency(peerID string, duration float64) {
	m.p2pMetrics.ObserveP2PLatency(peerID, duration)
}

func (m *Metrics) ObserveHandshakeDuration(duration float64) {
	m.p2pMetrics.ObserveHandshakeDuration(duration)
}

func (m *Metrics) SetMeshRoutes(count int) {
	m.p2pMetrics.SetMeshRoutes(count)
}

func (m *Metrics) IncPeerDiscoveryEvent(eventType string) {
	m.p2pMetrics.IncPeerDiscoveryEvent(eventType)
}

func (m *Metrics) IncFallbackEvent(fromProtocol, toProtocol, reason string) {
	m.p2pMetrics.IncFallbackEvent(fromProtocol, toProtocol, reason)
}

func (m *Metrics) IncWhitelistViolation(violationType string) {
	m.p2pMetrics.IncWhitelistViolation(violationType)
}

// QUIC Metrics Methods
func (m *Metrics) GetQUICMetrics() *QUICMetrics {
	return m.quicMetrics
}

// Connection Methods
func (m *Metrics) IncQUICConnections() {
	m.quicMetrics.IncConnections()
}

func (m *Metrics) DecQUICConnections() {
	m.quicMetrics.DecConnections()
}

func (m *Metrics) IncQUICConnectionsFailed(reason, version string) {
	m.quicMetrics.IncConnectionsFailed(reason, version)
}

func (m *Metrics) IncQUICConnectionsDropped(reason string) {
	m.quicMetrics.IncConnectionsDropped(reason)
}

// Stream Methods
func (m *Metrics) IncQUICStreams() {
	m.quicMetrics.IncStreams()
}

func (m *Metrics) DecQUICStreams() {
	m.quicMetrics.DecStreams()
}

func (m *Metrics) IncQUICStreamsCreated(streamType, direction string) {
	m.quicMetrics.IncStreamsCreated(streamType, direction)
}

func (m *Metrics) IncQUICStreamsClosed(streamType, reason string) {
	m.quicMetrics.IncStreamsClosed(streamType, reason)
}

func (m *Metrics) IncQUICStreamsReset(streamType, reason string) {
	m.quicMetrics.IncStreamsReset(streamType, reason)
}

// Performance Methods
func (m *Metrics) ObserveQUICHandshakeDuration(duration float64) {
	m.quicMetrics.ObserveHandshakeDuration(duration)
}

func (m *Metrics) IncQUICHandshakeFailures(reason, version string) {
	m.quicMetrics.IncHandshakeFailures(reason, version)
}

func (m *Metrics) ObserveQUICConnectionLatency(connectionID string, duration float64) {
	m.quicMetrics.ObserveConnectionLatency(connectionID, duration)
}

func (m *Metrics) IncQUICThroughputBytes(connectionID, direction string, bytes int64) {
	m.quicMetrics.IncThroughputBytes(connectionID, direction, bytes)
}

func (m *Metrics) IncQUICThroughputPackets(connectionID, direction string, packets int64) {
	m.quicMetrics.IncThroughputPackets(connectionID, direction, packets)
}

func (m *Metrics) SetQUICPacketLossRate(connectionID string, rate float64) {
	m.quicMetrics.SetPacketLossRate(connectionID, rate)
}

func (m *Metrics) IncQUICPacketLossEvents(connectionID, reason string) {
	m.quicMetrics.IncPacketLossEvents(connectionID, reason)
}

// Connection Quality Methods
func (m *Metrics) ObserveQUICRTT(connectionID string, rtt float64) {
	m.quicMetrics.ObserveRTT(connectionID, rtt)
}

func (m *Metrics) SetQUICRTTVariance(connectionID string, variance float64) {
	m.quicMetrics.SetRTTVariance(connectionID, variance)
}

func (m *Metrics) SetQUICCongestionWindow(connectionID string, windowSize int64) {
	m.quicMetrics.SetCongestionWindow(connectionID, windowSize)
}

func (m *Metrics) SetQUICBytesInFlight(connectionID string, bytes int64) {
	m.quicMetrics.SetBytesInFlight(connectionID, bytes)
}

func (m *Metrics) SetQUICPacketsInFlight(connectionID string, packets int64) {
	m.quicMetrics.SetPacketsInFlight(connectionID, packets)
}

// Error Methods
func (m *Metrics) IncQUICPacketErrors(errorType, connectionID string) {
	m.quicMetrics.IncPacketErrors(errorType, connectionID)
}

func (m *Metrics) IncQUICConnectionErrors(errorType, connectionID string) {
	m.quicMetrics.IncConnectionErrors(errorType, connectionID)
}

func (m *Metrics) IncQUICStreamErrors(errorType, streamID string) {
	m.quicMetrics.IncStreamErrors(errorType, streamID)
}

func (m *Metrics) IncQUICTimeoutEvents(timeoutType, connectionID string) {
	m.quicMetrics.IncTimeoutEvents(timeoutType, connectionID)
}

func (m *Metrics) IncQUICRetransmissions(connectionID, reason string) {
	m.quicMetrics.IncRetransmissions(connectionID, reason)
}

// Security Methods
func (m *Metrics) IncQUICEncryptionFailures(failureType, connectionID string) {
	m.quicMetrics.IncEncryptionFailures(failureType, connectionID)
}

func (m *Metrics) IncQUICAuthenticationFailures(failureType, connectionID string) {
	m.quicMetrics.IncAuthenticationFailures(failureType, connectionID)
}

func (m *Metrics) IncQUICKeyRotationEvents(connectionID, keyType string) {
	m.quicMetrics.IncKeyRotationEvents(connectionID, keyType)
}

// Configuration Methods
func (m *Metrics) SetQUICVersionUsage(version string, count int) {
	m.quicMetrics.SetVersionUsage(version, count)
}

func (m *Metrics) IncQUICParameterNegotiation(parameter, result string) {
	m.quicMetrics.IncParameterNegotiation(parameter, result)
}

func (m *Metrics) IncQUICExtensionUsage(extension, connectionID string) {
	m.quicMetrics.IncExtensionUsage(extension, connectionID)
}

// Resource Monitoring Methods
func (m *Metrics) SetQUICMemoryUsage(connectionID, component string, bytes int64) {
	m.quicMetrics.SetMemoryUsage(connectionID, component, bytes)
}

func (m *Metrics) SetQUICCPUUsage(connectionID, component string, percent float64) {
	m.quicMetrics.SetCPUUsage(connectionID, component, percent)
}

func (m *Metrics) SetQUICGoroutineCount(connectionID, component string, count int) {
	m.quicMetrics.SetGoroutineCount(connectionID, component, count)
}

// WireGuard Metrics Methods
func (m *Metrics) GetWireGuardMetrics() *WireGuardMetrics {
	return m.wireguardMetrics
}

// Interface Methods
func (m *Metrics) IncWireGuardInterfaces() {
	m.wireguardMetrics.IncInterfaces()
}

func (m *Metrics) DecWireGuardInterfaces() {
	m.wireguardMetrics.DecInterfaces()
}

func (m *Metrics) IncWireGuardInterfacesFailed(interfaceName, reason string) {
	m.wireguardMetrics.IncInterfacesFailed(interfaceName, reason)
}

func (m *Metrics) IncWireGuardInterfacesCreated(interfaceName, interfaceType string) {
	m.wireguardMetrics.IncInterfacesCreated(interfaceName, interfaceType)
}

func (m *Metrics) IncWireGuardInterfacesDestroyed(interfaceName, reason string) {
	m.wireguardMetrics.IncInterfacesDestroyed(interfaceName, reason)
}

// Peer Methods
func (m *Metrics) IncWireGuardPeers() {
	m.wireguardMetrics.IncPeers()
}

func (m *Metrics) DecWireGuardPeers() {
	m.wireguardMetrics.DecPeers()
}

func (m *Metrics) SetWireGuardPeersConnected(interfaceName, peer string, connected int) {
	m.wireguardMetrics.SetPeersConnected(interfaceName, peer, connected)
}

func (m *Metrics) IncWireGuardPeersAdded(interfaceName, peer string) {
	m.wireguardMetrics.IncPeersAdded(interfaceName, peer)
}

func (m *Metrics) IncWireGuardPeersRemoved(interfaceName, peer, reason string) {
	m.wireguardMetrics.IncPeersRemoved(interfaceName, peer, reason)
}

func (m *Metrics) IncWireGuardPeersFailed(interfaceName, peer, reason string) {
	m.wireguardMetrics.IncPeersFailed(interfaceName, peer, reason)
}

// Traffic Methods
func (m *Metrics) IncWireGuardTrafficBytesReceived(interfaceName, peer string, bytes int64) {
	m.wireguardMetrics.IncTrafficBytesReceived(interfaceName, peer, bytes)
}

func (m *Metrics) IncWireGuardTrafficBytesSent(interfaceName, peer string, bytes int64) {
	m.wireguardMetrics.IncTrafficBytesSent(interfaceName, peer, bytes)
}

func (m *Metrics) IncWireGuardTrafficPacketsReceived(interfaceName, peer string, packets int64) {
	m.wireguardMetrics.IncTrafficPacketsReceived(interfaceName, peer, packets)
}

func (m *Metrics) IncWireGuardTrafficPacketsSent(interfaceName, peer string, packets int64) {
	m.wireguardMetrics.IncTrafficPacketsSent(interfaceName, peer, packets)
}

func (m *Metrics) SetWireGuardTrafficRate(interfaceName, peer, direction string, rate float64) {
	m.wireguardMetrics.SetTrafficRate(interfaceName, peer, direction, rate)
}

// Performance Methods
func (m *Metrics) ObserveWireGuardHandshakeDuration(interfaceName, peer string, duration float64) {
	m.wireguardMetrics.ObserveHandshakeDuration(interfaceName, peer, duration)
}

func (m *Metrics) ObserveWireGuardConnectionLatency(interfaceName, peer string, duration float64) {
	m.wireguardMetrics.ObserveConnectionLatency(interfaceName, peer, duration)
}

func (m *Metrics) IncWireGuardThroughputBytes(interfaceName, peer, direction string, bytes int64) {
	m.wireguardMetrics.IncThroughputBytes(interfaceName, peer, direction, bytes)
}

func (m *Metrics) IncWireGuardThroughputPackets(interfaceName, peer, direction string, packets int64) {
	m.wireguardMetrics.IncThroughputPackets(interfaceName, peer, direction, packets)
}

func (m *Metrics) SetWireGuardBandwidthUtilization(interfaceName, peer string, utilization float64) {
	m.wireguardMetrics.SetBandwidthUtilization(interfaceName, peer, utilization)
}

// Connection Quality Methods
func (m *Metrics) SetWireGuardKeepaliveInterval(interfaceName, peer string, interval float64) {
	m.wireguardMetrics.SetKeepaliveInterval(interfaceName, peer, interval)
}

func (m *Metrics) SetWireGuardLastHandshake(interfaceName, peer string, timestamp float64) {
	m.wireguardMetrics.SetLastHandshake(interfaceName, peer, timestamp)
}

func (m *Metrics) IncWireGuardHandshakeAttempts(interfaceName, peer string) {
	m.wireguardMetrics.IncHandshakeAttempts(interfaceName, peer)
}

func (m *Metrics) IncWireGuardHandshakeFailures(interfaceName, peer, reason string) {
	m.wireguardMetrics.IncHandshakeFailures(interfaceName, peer, reason)
}

func (m *Metrics) SetWireGuardConnectionUptime(interfaceName, peer string, uptime float64) {
	m.wireguardMetrics.SetConnectionUptime(interfaceName, peer, uptime)
}

// Error Methods
func (m *Metrics) IncWireGuardInterfaceErrors(interfaceName, errorType string) {
	m.wireguardMetrics.IncInterfaceErrors(interfaceName, errorType)
}

func (m *Metrics) IncWireGuardPeerErrors(interfaceName, peer, errorType string) {
	m.wireguardMetrics.IncPeerErrors(interfaceName, peer, errorType)
}

func (m *Metrics) IncWireGuardHandshakeErrors(interfaceName, peer, errorType string) {
	m.wireguardMetrics.IncHandshakeErrors(interfaceName, peer, errorType)
}

func (m *Metrics) IncWireGuardTimeoutEvents(interfaceName, peer, timeoutType string) {
	m.wireguardMetrics.IncTimeoutEvents(interfaceName, peer, timeoutType)
}

func (m *Metrics) IncWireGuardReconnectionAttempts(interfaceName, peer string) {
	m.wireguardMetrics.IncReconnectionAttempts(interfaceName, peer)
}

// Security Methods
func (m *Metrics) IncWireGuardKeyRotationEvents(interfaceName, keyType string) {
	m.wireguardMetrics.IncKeyRotationEvents(interfaceName, keyType)
}

func (m *Metrics) IncWireGuardAuthenticationFailures(interfaceName, peer, failureType string) {
	m.wireguardMetrics.IncAuthenticationFailures(interfaceName, peer, failureType)
}

func (m *Metrics) IncWireGuardEncryptionFailures(interfaceName, peer, failureType string) {
	m.wireguardMetrics.IncEncryptionFailures(interfaceName, peer, failureType)
}

func (m *Metrics) IncWireGuardUnauthorizedAccess(interfaceName, peer, accessType string) {
	m.wireguardMetrics.IncUnauthorizedAccess(interfaceName, peer, accessType)
}

// Site-to-Site VPN Methods
func (m *Metrics) SetWireGuardSiteToSiteConnections(connectionID, status string, count int) {
	m.wireguardMetrics.SetSiteToSiteConnections(connectionID, status, count)
}

func (m *Metrics) SetWireGuardSiteToSiteActive(connectionID string, active int) {
	m.wireguardMetrics.SetSiteToSiteActive(connectionID, active)
}

func (m *Metrics) SetWireGuardSiteToSiteBandwidth(connectionID string, bandwidth float64) {
	m.wireguardMetrics.SetSiteToSiteBandwidth(connectionID, bandwidth)
}

func (m *Metrics) ObserveWireGuardSiteToSiteLatency(connectionID string, latency float64) {
	m.wireguardMetrics.ObserveSiteToSiteLatency(connectionID, latency)
}

// Routing Methods
func (m *Metrics) SetWireGuardRoutingRules(interfaceName, ruleType string, count int) {
	m.wireguardMetrics.SetRoutingRules(interfaceName, ruleType, count)
}

func (m *Metrics) IncWireGuardRoutingUpdates(interfaceName, updateType string) {
	m.wireguardMetrics.IncRoutingUpdates(interfaceName, updateType)
}

func (m *Metrics) IncWireGuardRoutingFailures(interfaceName, failureType string) {
	m.wireguardMetrics.IncRoutingFailures(interfaceName, failureType)
}

func (m *Metrics) IncWireGuardRouteChanges(interfaceName, changeType string) {
	m.wireguardMetrics.IncRouteChanges(interfaceName, changeType)
}

// Resource Monitoring Methods
func (m *Metrics) SetWireGuardMemoryUsage(interfaceName, component string, bytes int64) {
	m.wireguardMetrics.SetMemoryUsage(interfaceName, component, bytes)
}

func (m *Metrics) SetWireGuardCPUUsage(interfaceName, component string, percent float64) {
	m.wireguardMetrics.SetCPUUsage(interfaceName, component, percent)
}

func (m *Metrics) SetWireGuardGoroutineCount(interfaceName, component string, count int) {
	m.wireguardMetrics.SetGoroutineCount(interfaceName, component, count)
}

func (m *Metrics) SetWireGuardFileDescriptorCount(interfaceName string, count int) {
	m.wireguardMetrics.SetFileDescriptorCount(interfaceName, count)
}

// Configuration Methods
func (m *Metrics) IncWireGuardConfigUpdates(interfaceName, updateType string) {
	m.wireguardMetrics.IncConfigUpdates(interfaceName, updateType)
}

func (m *Metrics) IncWireGuardConfigErrors(interfaceName, errorType string) {
	m.wireguardMetrics.IncConfigErrors(interfaceName, errorType)
}

func (m *Metrics) IncWireGuardKeyGenerationEvents(interfaceName, keyType string) {
	m.wireguardMetrics.IncKeyGenerationEvents(interfaceName, keyType)
}

func (m *Metrics) SetWireGuardInterfaceStatus(interfaceName, status string, up bool) {
	m.wireguardMetrics.SetInterfaceStatus(interfaceName, status, up)
}

// gRPC Metrics Methods
func (m *Metrics) GetGRPCMetrics() *GRPCMetrics {
	return m.grpcMetrics
}

// Connection Methods
func (m *Metrics) IncGRPCConnections() {
	m.grpcMetrics.IncConnections()
}

func (m *Metrics) DecGRPCConnections() {
	m.grpcMetrics.DecConnections()
}

func (m *Metrics) IncGRPCConnectionsFailed(reason, clientIP string) {
	m.grpcMetrics.IncConnectionsFailed(reason, clientIP)
}

func (m *Metrics) IncGRPCConnectionsDropped(reason, clientIP string) {
	m.grpcMetrics.IncConnectionsDropped(reason, clientIP)
}

// Request Methods
func (m *Metrics) IncGRPCRequests(method, service, statusCode string) {
	m.grpcMetrics.IncRequests(method, service, statusCode)
}

func (m *Metrics) IncGRPCRequestsInFlight(method, service string) {
	m.grpcMetrics.IncRequestsInFlight(method, service)
}

func (m *Metrics) DecGRPCRequestsInFlight(method, service string) {
	m.grpcMetrics.DecRequestsInFlight(method, service)
}

func (m *Metrics) ObserveGRPCRequestDuration(method, service string, duration float64) {
	m.grpcMetrics.ObserveRequestDuration(method, service, duration)
}

func (m *Metrics) ObserveGRPCRequestSize(method, service string, size float64) {
	m.grpcMetrics.ObserveRequestSize(method, service, size)
}

func (m *Metrics) ObserveGRPCResponseSize(method, service string, size float64) {
	m.grpcMetrics.ObserveResponseSize(method, service, size)
}

// Method Methods
func (m *Metrics) IncGRPCMethodCalls(method, service, status string) {
	m.grpcMetrics.IncMethodCalls(method, service, status)
}

func (m *Metrics) ObserveGRPCMethodDuration(method, service string, duration float64) {
	m.grpcMetrics.ObserveMethodDuration(method, service, duration)
}

func (m *Metrics) IncGRPCMethodErrors(method, service, errorType string) {
	m.grpcMetrics.IncMethodErrors(method, service, errorType)
}

func (m *Metrics) IncGRPCMethodThroughput(method, service, direction string, count int64) {
	m.grpcMetrics.IncMethodThroughput(method, service, direction, count)
}

// Stream Methods
func (m *Metrics) IncGRPCStreams() {
	m.grpcMetrics.IncStreams()
}

func (m *Metrics) DecGRPCStreams() {
	m.grpcMetrics.DecStreams()
}

func (m *Metrics) IncGRPCStreamsCreated(method, service string) {
	m.grpcMetrics.IncStreamsCreated(method, service)
}

func (m *Metrics) IncGRPCStreamsClosed(method, service, reason string) {
	m.grpcMetrics.IncStreamsClosed(method, service, reason)
}

func (m *Metrics) IncGRPCStreamsErrors(method, service, errorType string) {
	m.grpcMetrics.IncStreamsErrors(method, service, errorType)
}

// Performance Methods
func (m *Metrics) ObserveGRPCHandshakeDuration(duration float64) {
	m.grpcMetrics.ObserveHandshakeDuration(duration)
}

func (m *Metrics) ObserveGRPCConnectionLatency(clientIP string, duration float64) {
	m.grpcMetrics.ObserveConnectionLatency(clientIP, duration)
}

func (m *Metrics) IncGRPCThroughputBytes(method, service, direction string, bytes int64) {
	m.grpcMetrics.IncThroughputBytes(method, service, direction, bytes)
}

func (m *Metrics) IncGRPCThroughputMessages(method, service, direction string, messages int64) {
	m.grpcMetrics.IncThroughputMessages(method, service, direction, messages)
}

func (m *Metrics) SetGRPCBandwidthUtilization(method, service string, utilization float64) {
	m.grpcMetrics.SetBandwidthUtilization(method, service, utilization)
}

// Connection Quality Methods
func (m *Metrics) IncGRPCKeepaliveEvents(clientIP, eventType string) {
	m.grpcMetrics.IncKeepaliveEvents(clientIP, eventType)
}

func (m *Metrics) IncGRPCKeepaliveFailures(clientIP, failureType string) {
	m.grpcMetrics.IncKeepaliveFailures(clientIP, failureType)
}

func (m *Metrics) SetGRPCConnectionUptime(clientIP string, uptime float64) {
	m.grpcMetrics.SetConnectionUptime(clientIP, uptime)
}

func (m *Metrics) SetGRPCLastActivity(clientIP string, timestamp float64) {
	m.grpcMetrics.SetLastActivity(clientIP, timestamp)
}

// Error Methods
func (m *Metrics) IncGRPCServerErrors(errorType, method, service string) {
	m.grpcMetrics.IncServerErrors(errorType, method, service)
}

func (m *Metrics) IncGRPCClientErrors(errorType, clientIP string) {
	m.grpcMetrics.IncClientErrors(errorType, clientIP)
}

func (m *Metrics) IncGRPCTimeoutEvents(timeoutType, method, service string) {
	m.grpcMetrics.IncTimeoutEvents(timeoutType, method, service)
}

func (m *Metrics) IncGRPCReconnectionAttempts(clientIP string) {
	m.grpcMetrics.IncReconnectionAttempts(clientIP)
}

func (m *Metrics) IncGRPCCircuitBreakerEvents(method, service, eventType string) {
	m.grpcMetrics.IncCircuitBreakerEvents(method, service, eventType)
}

// Security Methods
func (m *Metrics) IncGRPCAuthenticationFailures(clientIP, failureType string) {
	m.grpcMetrics.IncAuthenticationFailures(clientIP, failureType)
}

func (m *Metrics) IncGRPCAuthorizationFailures(clientIP, method, service string) {
	m.grpcMetrics.IncAuthorizationFailures(clientIP, method, service)
}

func (m *Metrics) IncGRPCRateLimitViolations(clientIP, limitType string) {
	m.grpcMetrics.IncRateLimitViolations(clientIP, limitType)
}

func (m *Metrics) IncGRPCRBACViolations(clientIP, method, service, violationType string) {
	m.grpcMetrics.IncRBACViolations(clientIP, method, service, violationType)
}

// Middleware Methods
func (m *Metrics) ObserveGRPCMiddlewareDuration(middleware, method, service string, duration float64) {
	m.grpcMetrics.ObserveMiddlewareDuration(middleware, method, service, duration)
}

func (m *Metrics) IncGRPCMiddlewareErrors(middleware, errorType string) {
	m.grpcMetrics.IncMiddlewareErrors(middleware, errorType)
}

func (m *Metrics) IncGRPCInterceptorCalls(interceptor, method, service string) {
	m.grpcMetrics.IncInterceptorCalls(interceptor, method, service)
}

// Tenant Methods
func (m *Metrics) IncGRPCTenantRequests(tenantID, method, service string) {
	m.grpcMetrics.IncTenantRequests(tenantID, method, service)
}

func (m *Metrics) SetGRPCTenantConnections(tenantID string, connections int) {
	m.grpcMetrics.SetTenantConnections(tenantID, connections)
}

func (m *Metrics) IncGRPCTenantBandwidth(tenantID, direction string, bytes int64) {
	m.grpcMetrics.IncTenantBandwidth(tenantID, direction, bytes)
}

func (m *Metrics) IncGRPCTenantErrors(tenantID, errorType string) {
	m.grpcMetrics.IncTenantErrors(tenantID, errorType)
}

// Resource Monitoring Methods
func (m *Metrics) SetGRPCMemoryUsage(component string, bytes int64) {
	m.grpcMetrics.SetMemoryUsage(component, bytes)
}

func (m *Metrics) SetGRPCCPUUsage(component string, percent float64) {
	m.grpcMetrics.SetCPUUsage(component, percent)
}

func (m *Metrics) SetGRPCGoroutineCount(component string, count int) {
	m.grpcMetrics.SetGoroutineCount(component, count)
}

func (m *Metrics) SetGRPCFileDescriptorCount(count int) {
	m.grpcMetrics.SetFileDescriptorCount(count)
}

// Configuration Methods
func (m *Metrics) SetGRPCServerStatus(up bool) {
	m.grpcMetrics.SetServerStatus(up)
}

func (m *Metrics) SetGRPCServerVersion(version string) {
	m.grpcMetrics.SetServerVersion(version)
}

func (m *Metrics) IncGRPCConfigUpdates(updateType string) {
	m.grpcMetrics.IncConfigUpdates(updateType)
}

func (m *Metrics) IncGRPCConfigErrors(errorType string) {
	m.grpcMetrics.IncConfigErrors(errorType)
}

// AI Service Metrics Methods
func (m *Metrics) GetAIServiceMetrics() *AIServiceMetrics {
	return m.aiServiceMetrics
}

// Anomaly Detection Methods
func (m *Metrics) IncAIAnomalyDetections(model, anomalyType, severity string) {
	m.aiServiceMetrics.IncAnomalyDetections(model, anomalyType, severity)
}

func (m *Metrics) SetAIAnomalyDetectionsActive(count int) {
	m.aiServiceMetrics.SetAnomalyDetectionsActive(count)
}

func (m *Metrics) IncAIAnomalyFalsePositives(model, anomalyType string) {
	m.aiServiceMetrics.IncAnomalyFalsePositives(model, anomalyType)
}

func (m *Metrics) IncAIAnomalyFalseNegatives(model, anomalyType string) {
	m.aiServiceMetrics.IncAnomalyFalseNegatives(model, anomalyType)
}

func (m *Metrics) SetAIAnomalyAccuracy(model, anomalyType string, accuracy float64) {
	m.aiServiceMetrics.SetAnomalyAccuracy(model, anomalyType, accuracy)
}

func (m *Metrics) ObserveAIAnomalyDetectionDuration(model, anomalyType string, duration float64) {
	m.aiServiceMetrics.ObserveAnomalyDetectionDuration(model, anomalyType, duration)
}

// Load Prediction Methods
func (m *Metrics) IncAILoadPredictions(model, predictionType string) {
	m.aiServiceMetrics.IncLoadPredictions(model, predictionType)
}

func (m *Metrics) IncAILoadPredictionErrors(model, errorType string) {
	m.aiServiceMetrics.IncLoadPredictionErrors(model, errorType)
}

func (m *Metrics) SetAILoadPredictionAccuracy(model, predictionType string, accuracy float64) {
	m.aiServiceMetrics.SetLoadPredictionAccuracy(model, predictionType, accuracy)
}

func (m *Metrics) ObserveAILoadPredictionDuration(model, predictionType string, duration float64) {
	m.aiServiceMetrics.ObserveLoadPredictionDuration(model, predictionType, duration)
}

func (m *Metrics) SetAILoadPredictionConfidence(model, predictionType string, confidence float64) {
	m.aiServiceMetrics.SetLoadPredictionConfidence(model, predictionType, confidence)
}

// Route Optimization Methods
func (m *Metrics) IncAIRouteOptimizations(algorithm, optimizationType string) {
	m.aiServiceMetrics.IncRouteOptimizations(algorithm, optimizationType)
}

func (m *Metrics) ObserveAIRouteOptimizationGains(algorithm, optimizationType string, gain float64) {
	m.aiServiceMetrics.ObserveRouteOptimizationGains(algorithm, optimizationType, gain)
}

func (m *Metrics) ObserveAIRouteOptimizationDuration(algorithm, optimizationType string, duration float64) {
	m.aiServiceMetrics.ObserveRouteOptimizationDuration(algorithm, optimizationType, duration)
}

func (m *Metrics) IncAIRouteOptimizationSuccess(algorithm, optimizationType string) {
	m.aiServiceMetrics.IncRouteOptimizationSuccess(algorithm, optimizationType)
}

// Data Pipeline Methods
func (m *Metrics) IncAIDataProcessing(pipeline, stage, dataType string) {
	m.aiServiceMetrics.IncDataProcessing(pipeline, stage, dataType)
}

func (m *Metrics) ObserveAIDataProcessingDuration(pipeline, stage, dataType string, duration float64) {
	m.aiServiceMetrics.ObserveDataProcessingDuration(pipeline, stage, dataType, duration)
}

func (m *Metrics) IncAIDataProcessingErrors(pipeline, stage, errorType string) {
	m.aiServiceMetrics.IncDataProcessingErrors(pipeline, stage, errorType)
}

func (m *Metrics) IncAIDataProcessingThroughput(pipeline, stage, dataType, direction string, count int64) {
	m.aiServiceMetrics.IncDataProcessingThroughput(pipeline, stage, dataType, direction, count)
}

func (m *Metrics) ObserveAIDataProcessingLatency(pipeline, stage, dataType string, latency float64) {
	m.aiServiceMetrics.ObserveDataProcessingLatency(pipeline, stage, dataType, latency)
}

// ML Model Methods
func (m *Metrics) IncAIModelInferences(model, modelType, version string) {
	m.aiServiceMetrics.IncModelInferences(model, modelType, version)
}

func (m *Metrics) ObserveAIModelInferenceDuration(model, modelType, version string, duration float64) {
	m.aiServiceMetrics.ObserveModelInferenceDuration(model, modelType, version, duration)
}

func (m *Metrics) SetAIModelAccuracy(model, modelType, version string, accuracy float64) {
	m.aiServiceMetrics.SetModelAccuracy(model, modelType, version, accuracy)
}

func (m *Metrics) SetAIModelConfidence(model, modelType, version string, confidence float64) {
	m.aiServiceMetrics.SetModelConfidence(model, modelType, version, confidence)
}

func (m *Metrics) ObserveAIModelTrainingDuration(model, modelType, version string, duration float64) {
	m.aiServiceMetrics.ObserveModelTrainingDuration(model, modelType, version, duration)
}

func (m *Metrics) IncAIModelTrainingErrors(model, modelType, version, errorType string) {
	m.aiServiceMetrics.IncModelTrainingErrors(model, modelType, version, errorType)
}

// Analytics Methods
func (m *Metrics) IncAIAnalyticsReportsGenerated(reportType, format string) {
	m.aiServiceMetrics.IncAnalyticsReportsGenerated(reportType, format)
}

func (m *Metrics) ObserveAIAnalyticsReportDuration(reportType, format string, duration float64) {
	m.aiServiceMetrics.ObserveAnalyticsReportDuration(reportType, format, duration)
}

func (m *Metrics) IncAIAnalyticsInsights(insightType, category string) {
	m.aiServiceMetrics.IncAnalyticsInsights(insightType, category)
}

func (m *Metrics) SetAIAnalyticsInsightAccuracy(insightType, category string, accuracy float64) {
	m.aiServiceMetrics.SetAnalyticsInsightAccuracy(insightType, category, accuracy)
}

// Performance Methods
func (m *Metrics) IncAIProcessingTime(component, operation string, duration float64) {
	m.aiServiceMetrics.IncProcessingTime(component, operation, duration)
}

func (m *Metrics) ObserveAIProcessingTimeDuration(component, operation string, duration float64) {
	m.aiServiceMetrics.ObserveProcessingTimeDuration(component, operation, duration)
}

func (m *Metrics) SetAIProcessingQueueSize(queueType, priority string, size int) {
	m.aiServiceMetrics.SetProcessingQueueSize(queueType, priority, size)
}

func (m *Metrics) ObserveAIProcessingQueueLatency(queueType, priority string, latency float64) {
	m.aiServiceMetrics.ObserveProcessingQueueLatency(queueType, priority, latency)
}

// Resource Usage Methods
func (m *Metrics) SetAIMemoryUsage(component, usageType string, bytes int64) {
	m.aiServiceMetrics.SetMemoryUsage(component, usageType, bytes)
}

func (m *Metrics) SetAICPUUsage(component, usageType string, percent float64) {
	m.aiServiceMetrics.SetCPUUsage(component, usageType, percent)
}

func (m *Metrics) SetAIGPUUsage(component, usageType string, percent float64) {
	m.aiServiceMetrics.SetGPUUsage(component, usageType, percent)
}

func (m *Metrics) SetAIDiskUsage(component, usageType string, bytes int64) {
	m.aiServiceMetrics.SetDiskUsage(component, usageType, bytes)
}

func (m *Metrics) IncAINetworkUsage(component, direction string, bytes int64) {
	m.aiServiceMetrics.IncNetworkUsage(component, direction, bytes)
}

// Cache Performance Methods
func (m *Metrics) IncAICacheHits(cacheType, component string) {
	m.aiServiceMetrics.IncCacheHits(cacheType, component)
}

func (m *Metrics) IncAICacheMisses(cacheType, component string) {
	m.aiServiceMetrics.IncCacheMisses(cacheType, component)
}

func (m *Metrics) SetAICacheSize(cacheType, component string, size int) {
	m.aiServiceMetrics.SetCacheSize(cacheType, component, size)
}

func (m *Metrics) IncAICacheEvictions(cacheType, component, reason string) {
	m.aiServiceMetrics.IncCacheEvictions(cacheType, component, reason)
}

// Error Handling Methods
func (m *Metrics) IncAIErrors(component, errorType, severity string) {
	m.aiServiceMetrics.IncErrors(component, errorType, severity)
}

func (m *Metrics) IncAIErrorRecoveryAttempts(component, errorType string) {
	m.aiServiceMetrics.IncErrorRecoveryAttempts(component, errorType)
}

func (m *Metrics) IncAIErrorRecoverySuccess(component, errorType string) {
	m.aiServiceMetrics.IncErrorRecoverySuccess(component, errorType)
}

func (m *Metrics) IncAICircuitBreakerEvents(component, eventType string) {
	m.aiServiceMetrics.IncCircuitBreakerEvents(component, eventType)
}

// Monitoring Methods
func (m *Metrics) IncAIHealthChecks(component, status string) {
	m.aiServiceMetrics.IncHealthChecks(component, status)
}

func (m *Metrics) ObserveAIHealthCheckDuration(component string, duration float64) {
	m.aiServiceMetrics.ObserveHealthCheckDuration(component, duration)
}

func (m *Metrics) SetAIHealthStatus(component string, healthy bool) {
	m.aiServiceMetrics.SetHealthStatus(component, healthy)
}

func (m *Metrics) SetAIUptimeSeconds(uptime float64) {
	m.aiServiceMetrics.SetUptimeSeconds(uptime)
}

// Configuration Methods
func (m *Metrics) IncAIConfigUpdates(component, updateType string) {
	m.aiServiceMetrics.IncConfigUpdates(component, updateType)
}

func (m *Metrics) IncAIConfigErrors(component, errorType string) {
	m.aiServiceMetrics.IncConfigErrors(component, errorType)
}

func (m *Metrics) SetAIFeatureFlag(feature, component string, enabled bool) {
	m.aiServiceMetrics.SetFeatureFlag(feature, component, enabled)
}

// Cadence Metrics Methods
func (m *Metrics) GetCadenceMetrics() *CadenceMetrics {
	return m.cadenceMetrics
}

// Workflow Methods
func (m *Metrics) IncCadenceWorkflows(workflowType, domain string) {
	m.cadenceMetrics.RecordWorkflowStarted(workflowType, domain, "")
}

func (m *Metrics) SetCadenceWorkflowsActive(workflowType, domain string, count int) {
	m.cadenceMetrics.ActiveWorkflows.WithLabelValues(workflowType, domain).Set(float64(count))
}

func (m *Metrics) IncCadenceWorkflowsCompleted(workflowType, domain string) {
	m.cadenceMetrics.RecordWorkflowCompleted(workflowType, domain, "", 0)
}

func (m *Metrics) IncCadenceWorkflowsFailed(workflowType, domain, errorType string) {
	m.cadenceMetrics.RecordWorkflowFailed(workflowType, domain, "", errorType, 0)
}

func (m *Metrics) IncCadenceWorkflowsCancelled(workflowType, domain string) {
	m.cadenceMetrics.RecordWorkflowFailed(workflowType, domain, "", "cancelled", 0)
}

func (m *Metrics) IncCadenceWorkflowsTerminated(workflowType, domain string) {
	m.cadenceMetrics.RecordWorkflowFailed(workflowType, domain, "", "terminated", 0)
}

func (m *Metrics) ObserveCadenceWorkflowDuration(workflowType, domain string, duration float64) {
	m.cadenceMetrics.WorkflowDuration.WithLabelValues(workflowType, domain).Observe(duration)
}

func (m *Metrics) ObserveCadenceWorkflowExecutionTime(workflowType, domain string, duration float64) {
	m.cadenceMetrics.WorkflowDuration.WithLabelValues(workflowType, domain).Observe(duration)
}

// Activity Methods
func (m *Metrics) IncCadenceActivities(activityType, domain string) {
	m.cadenceMetrics.RecordActivityStarted(activityType, domain)
}

func (m *Metrics) SetCadenceActivitiesActive(activityType, domain string, count int) {
	// Activities are tracked via counters, not gauges
}

func (m *Metrics) IncCadenceActivitiesCompleted(activityType, domain string) {
	m.cadenceMetrics.RecordActivityCompleted(activityType, domain, 0)
}

func (m *Metrics) IncCadenceActivitiesFailed(activityType, domain, errorType string) {
	m.cadenceMetrics.RecordActivityFailed(activityType, domain, errorType, 0)
}

func (m *Metrics) IncCadenceActivitiesRetried(activityType, domain string) {
	m.cadenceMetrics.RecordActivityRetry(activityType, domain, 1)
}

func (m *Metrics) IncCadenceActivitiesTimeout(activityType, domain string) {
	m.cadenceMetrics.RecordActivityFailed(activityType, domain, "timeout", 0)
}

func (m *Metrics) ObserveCadenceActivityDuration(activityType, domain string, duration float64) {
	m.cadenceMetrics.ActivityDuration.WithLabelValues(activityType, domain).Observe(duration)
}

func (m *Metrics) ObserveCadenceActivityExecutionTime(activityType, domain string, duration float64) {
	m.cadenceMetrics.ActivityDuration.WithLabelValues(activityType, domain).Observe(duration)
}

// Additional Cadence Methods using existing API
func (m *Metrics) RecordCadenceQUICWorkflow(subtype, status, peerID string) {
	m.cadenceMetrics.RecordQUICWorkflow(subtype, status, peerID)
}

func (m *Metrics) RecordCadenceQuantumWorkflow(subtype, status, algorithm string) {
	m.cadenceMetrics.RecordQuantumWorkflow(subtype, status, algorithm)
}

func (m *Metrics) RecordCadenceAIWorkflow(subtype, status, modelType string) {
	m.cadenceMetrics.RecordAIWorkflow(subtype, status, modelType)
}

func (m *Metrics) RecordCadenceScalingWorkflow(subtype, status, scaleDirection string) {
	m.cadenceMetrics.RecordScalingWorkflow(subtype, status, scaleDirection)
}

func (m *Metrics) RecordCadenceSecurityWorkflow(subtype, status, threatLevel string) {
	m.cadenceMetrics.RecordSecurityWorkflow(subtype, status, threatLevel)
}

func (m *Metrics) UpdateCadenceWorkerConnections(workerIdentity, domain string, count float64) {
	m.cadenceMetrics.UpdateWorkerConnections(workerIdentity, domain, count)
}

func (m *Metrics) UpdateCadenceWorkerPollers(pollerType, workerIdentity, domain string, count float64) {
	m.cadenceMetrics.UpdateWorkerPollers(pollerType, workerIdentity, domain, count)
}

func (m *Metrics) RecordCadenceWorkerLatency(operation, workerIdentity, domain string, duration time.Duration) {
	m.cadenceMetrics.RecordWorkerLatency(operation, workerIdentity, domain, duration)
}

func (m *Metrics) UpdateCadenceWorkerQueueDepth(queueType, workerIdentity, domain string, depth float64) {
	m.cadenceMetrics.UpdateWorkerQueueDepth(queueType, workerIdentity, domain, depth)
}

func (m *Metrics) RecordCadenceClientRequest(operation, status, domain string, duration time.Duration) {
	m.cadenceMetrics.RecordClientRequest(operation, status, domain, duration)
}

func (m *Metrics) RecordCadenceClientError(operation, errorType, domain string) {
	m.cadenceMetrics.RecordClientError(operation, errorType, domain)
}

func (m *Metrics) RecordCadenceHealthCheck(checkType, status, component string) {
	m.cadenceMetrics.RecordHealthCheck(checkType, status, component)
}

func (m *Metrics) UpdateCadenceSystemUptime(component, instance string, uptime time.Duration) {
	m.cadenceMetrics.UpdateSystemUptime(component, instance, uptime)
}

func (m *Metrics) UpdateCadenceResourceUtilization(resourceType, component string, utilization float64) {
	m.cadenceMetrics.UpdateResourceUtilization(resourceType, component, utilization)
}
