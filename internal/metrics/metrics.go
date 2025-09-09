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

	// Новые метрики для Keycloak
	keycloakTokenValidations        *prometheus.CounterVec
	keycloakJWKSCacheHits           prometheus.Counter
	keycloakJWKSCacheMisses         prometheus.Counter
	keycloakJWKSCacheSize           prometheus.Gauge
	keycloakJWKSFetchDuration       prometheus.Histogram
	keycloakTokenValidationDuration *prometheus.HistogramVec
	keycloakUserInfoDuration        prometheus.Histogram
	keycloakPermissionChecks        *prometheus.CounterVec

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
			}, []string{"tunnel_id", "remote_host", "remote_port", "status", "duration_seconds", "bytes_transferred", "last_activity"}),
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
			}, []string{"client_ip", "failed_attempts", "last_attempt", "geolocation", "threat_level"}),
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
				Name: "relay_tenant_comparison_metrics",
				Help: "Comparison metrics per tenant",
			}, []string{"tenant_id", "tenant_name", "active_tunnels", "active_connections", "bandwidth_mbps", "data_transfer_gb", "response_time_ms", "error_rate_percent", "sla_compliance_percent", "risk_score", "revenue_per_hour", "customer_satisfaction"}),
			resourceUtilizationPercent: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "relay_tenant_resource_utilization_percent",
				Help: "Overall resource utilization percentage",
			}),
		},
		p2pMetrics: NewP2PMetrics(reg),
	}

	// Регистрация всех метрик
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
		m.keycloakTokenValidations,
		m.keycloakJWKSCacheHits,
		m.keycloakJWKSCacheMisses,
		m.keycloakJWKSCacheSize,
		m.keycloakJWKSFetchDuration,
		m.keycloakTokenValidationDuration,
		m.keycloakUserInfoDuration,
		m.keycloakPermissionChecks,
		m.businessMetrics.activeUsers,
		m.businessMetrics.revenuePerHour,
		m.businessMetrics.customerSatisfaction,
		m.businessMetrics.totalRevenue,
		m.technicalMetrics.memoryUsage,
		m.technicalMetrics.gcPauseDuration,
		m.technicalMetrics.networkLatency,
		m.technicalMetrics.diskIO,
		m.tunnelMetrics.creationsTotal,
		m.tunnelMetrics.closuresTotal,
		m.tunnelMetrics.durationSeconds,
		m.tunnelMetrics.bytesFromClientTotal,
		m.tunnelMetrics.bytesToClientTotal,
		m.tunnelMetrics.bandwidthUtilization,
		m.tunnelMetrics.errorsTotal,
		m.tunnelMetrics.status,
		m.tunnelMetrics.activityHeatmap,
		m.tunnelMetrics.latencySeconds,
		m.tunnelMetrics.operationsTotal,
		m.tunnelMetrics.cleanupOperationsTotal,
		m.tunnelMetrics.memoryUsageBytes,
		m.tunnelMetrics.cpuUsageSecondsTotal,
		m.securityMetrics.suspiciousActivitiesTotal,
		m.securityMetrics.rateLimitViolationsTotal,
		m.securityMetrics.ddosAttemptsTotal,
		m.securityMetrics.securityEventsTotal,
		m.securityMetrics.failedLoginsByIP,
		m.securityMetrics.tlsCertificateStatus,
		m.securityMetrics.tlsHandshakeFailuresTotal,
		m.securityMetrics.tlsCertificateExpirationDays,
		m.mlMetrics.anomalyScore,
		m.mlMetrics.threatPredictionsTotal,
		m.tenantMetrics.activeCount,
		m.tenantMetrics.activeTunnels,
		m.tenantMetrics.activeConnections,
		m.tenantMetrics.bandwidthBytesTotal,
		m.tenantMetrics.dataTransferGB,
		m.tenantMetrics.responseTimeSeconds,
		m.tenantMetrics.errorsTotal,
		m.tenantMetrics.slaCompliancePercent,
		m.tenantMetrics.uptimePercent,
		m.tenantMetrics.securityIncidentsTotal,
		m.tenantMetrics.authFailuresTotal,
		m.tenantMetrics.riskScore,
		m.tenantMetrics.isolationViolationsTotal,
		m.tenantMetrics.accessPatternsTotal,
		m.tenantMetrics.revenueTotal,
		m.tenantMetrics.customerSatisfaction,
		m.tenantMetrics.dailyActiveUsers,
		m.tenantMetrics.usageHeatmap,
		m.tenantMetrics.comparisonMetrics,
		m.tenantMetrics.resourceUtilizationPercent,
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

func (m *Metrics) SetTunnelStatus(tunnelID, remoteHost string, remotePort int, status string, durationSeconds float64, bytesTransferred int64, lastActivity string) {
	m.tunnelMetrics.status.WithLabelValues(tunnelID, remoteHost, fmt.Sprintf("%d", remotePort), status, fmt.Sprintf("%.2f", durationSeconds), fmt.Sprintf("%d", bytesTransferred), lastActivity).Set(1)
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

func (m *Metrics) SetFailedLoginsByIP(clientIP, failedAttempts, lastAttempt, geolocation, threatLevel string) {
	m.securityMetrics.failedLoginsByIP.WithLabelValues(clientIP, failedAttempts, lastAttempt, geolocation, threatLevel).Set(1)
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

func (m *Metrics) SetTenantComparisonMetrics(tenantID, tenantName string, activeTunnels, activeConnections int, bandwidthMbps, dataTransferGB, responseTimeMs, errorRatePercent, slaCompliancePercent, riskScore, revenuePerHour, customerSatisfaction float64) {
	m.tenantMetrics.comparisonMetrics.WithLabelValues(
		tenantID, tenantName,
		fmt.Sprintf("%d", activeTunnels),
		fmt.Sprintf("%d", activeConnections),
		fmt.Sprintf("%.2f", bandwidthMbps),
		fmt.Sprintf("%.2f", dataTransferGB),
		fmt.Sprintf("%.2f", responseTimeMs),
		fmt.Sprintf("%.2f", errorRatePercent),
		fmt.Sprintf("%.2f", slaCompliancePercent),
		fmt.Sprintf("%.2f", riskScore),
		fmt.Sprintf("%.2f", revenuePerHour),
		fmt.Sprintf("%.2f", customerSatisfaction),
	).Set(1)
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
