package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ZitadelMetrics содержит метрики для Zitadel интеграции
type ZitadelMetrics struct {
	// API вызовы
	APICallsTotal *prometheus.CounterVec
	APICallDuration *prometheus.HistogramVec
	
	// JWT валидация
	JWTValidationTotal *prometheus.CounterVec
	JWTValidationDuration *prometheus.HistogramVec
	
	// Синхронизация тенантов
	TenantSyncTotal *prometheus.CounterVec
	TenantSyncDuration *prometheus.HistogramVec
	TenantSyncErrors *prometheus.CounterVec
	TenantLimitsChanged *prometheus.CounterVec
	
	// Fallback использование
	FallbackUsageTotal *prometheus.CounterVec
}

// NewZitadelMetrics создает новые метрики для Zitadel
func NewZitadelMetrics() *ZitadelMetrics {
	return &ZitadelMetrics{
		// API вызовы
		APICallsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "zitadel_api_calls_total",
				Help: "Total number of Zitadel API calls",
			},
			[]string{"method", "status"},
		),
		APICallDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "zitadel_api_call_duration_seconds",
				Help:    "Duration of Zitadel API calls",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method"},
		),
		
		// JWT валидация
		JWTValidationTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "jwt_validation_total",
				Help: "Total number of JWT validations",
			},
			[]string{"issuer", "status"},
		),
		JWTValidationDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "jwt_validation_duration_seconds",
				Help:    "Duration of JWT validation",
				Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
			},
			[]string{"issuer"},
		),
		
		// Синхронизация тенантов
		TenantSyncTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "tenant_sync_total",
				Help: "Total number of tenant sync operations",
			},
			[]string{"result"},
		),
		TenantSyncDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "tenant_sync_duration_seconds",
				Help:    "Duration of tenant sync operations",
				Buckets: []float64{1, 5, 10, 30, 60, 120, 300, 600},
			},
			[]string{"result"},
		),
		TenantSyncErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "tenant_sync_errors_total",
				Help: "Total number of tenant sync errors",
			},
			[]string{"error_type"},
		),
		TenantLimitsChanged: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "tenant_limits_changed_total",
				Help: "Total number of tenant limits changes",
			},
			[]string{"tenant_id", "limit_type"},
		),
		
		// Fallback использование
		FallbackUsageTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "fallback_usage_total",
				Help: "Total number of fallback usages",
			},
			[]string{"fallback_type", "result"},
		),
	}
}

// RecordAPICall записывает метрику API вызова
func (m *ZitadelMetrics) RecordAPICall(method, status string, duration time.Duration) {
	m.APICallsTotal.WithLabelValues(method, status).Inc()
	m.APICallDuration.WithLabelValues(method).Observe(duration.Seconds())
}

// RecordJWTValidation записывает метрику JWT валидации
func (m *ZitadelMetrics) RecordJWTValidation(issuer, status string, duration time.Duration) {
	m.JWTValidationTotal.WithLabelValues(issuer, status).Inc()
	m.JWTValidationDuration.WithLabelValues(issuer).Observe(duration.Seconds())
}

// RecordTenantSync записывает метрику синхронизации тенантов
func (m *ZitadelMetrics) RecordTenantSync(result string, duration time.Duration) {
	m.TenantSyncTotal.WithLabelValues(result).Inc()
	m.TenantSyncDuration.WithLabelValues(result).Observe(duration.Seconds())
}

// RecordTenantSyncError записывает метрику ошибки синхронизации
func (m *ZitadelMetrics) RecordTenantSyncError(errorType string) {
	m.TenantSyncErrors.WithLabelValues(errorType).Inc()
}

// RecordTenantLimitsChange записывает метрику изменения лимитов
func (m *ZitadelMetrics) RecordTenantLimitsChange(tenantID, limitType string) {
	m.TenantLimitsChanged.WithLabelValues(tenantID, limitType).Inc()
}

// RecordFallbackUsage записывает метрику использования fallback
func (m *ZitadelMetrics) RecordFallbackUsage(fallbackType, result string) {
	m.FallbackUsageTotal.WithLabelValues(fallbackType, result).Inc()
}


