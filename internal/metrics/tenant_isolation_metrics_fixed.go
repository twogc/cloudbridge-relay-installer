package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// TenantIsolationMetrics - исправленные метрики с правильными лейблами
type TenantIsolationMetrics struct {
	// Общие метрики
	TotalTenants   prometheus.Gauge
	TotalSubnets   prometheus.Gauge
	TotalAllocated prometheus.Gauge
	TotalAvailable prometheus.Gauge

	// Метрики изоляции с лейблами
	IsolationViolations *prometheus.CounterVec
	PolicyViolations    *prometheus.CounterVec
	TenantHealth        *prometheus.GaugeVec

	// Метрики по тенантам
	TenantSubnetCount  *prometheus.GaugeVec
	TenantAllocatedIPs *prometheus.GaugeVec
	TenantAvailableIPs *prometheus.GaugeVec

	// Метрики политик
	PolicyCount  *prometheus.GaugeVec
	PolicyStatus *prometheus.GaugeVec

	// Метрики производительности
	PolicyReconcileDuration *prometheus.HistogramVec
	IPAllocationDuration    *prometheus.HistogramVec
}

// NewTenantIsolationMetrics - создание метрик с правильными лейблами
func NewTenantIsolationMetrics() *TenantIsolationMetrics {
	return &TenantIsolationMetrics{
		// Общие метрики
		TotalTenants: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "cloudbridge_tenant_total",
			Help: "Total number of tenants",
		}),
		TotalSubnets: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "cloudbridge_tenant_subnets_total",
			Help: "Total number of tenant subnets",
		}),
		TotalAllocated: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "cloudbridge_tenant_allocated_ips_total",
			Help: "Total number of allocated IPs",
		}),
		TotalAvailable: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "cloudbridge_tenant_available_ips_total",
			Help: "Total number of available IPs",
		}),

		// Метрики изоляции с лейблами
		IsolationViolations: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "cloudbridge_tenant_isolation_violations_total",
			Help: "Total number of tenant isolation violations",
		}, []string{"src_tenant", "dst_tenant", "protocol", "violation_type"}),

		PolicyViolations: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "cloudbridge_tenant_policy_violations_total",
			Help: "Total number of policy violations",
		}, []string{"tenant_id", "policy_type", "violation_reason"}),

		TenantHealth: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "cloudbridge_tenant_health",
			Help: "Health status of tenants",
		}, []string{"tenant_id", "status", "component"}),

		// Метрики по тенантам
		TenantSubnetCount: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "cloudbridge_tenant_subnet_count",
			Help: "Number of subnets per tenant",
		}, []string{"tenant_id"}),

		TenantAllocatedIPs: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "cloudbridge_tenant_allocated_ips",
			Help: "Number of allocated IPs per tenant",
		}, []string{"tenant_id"}),

		TenantAvailableIPs: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "cloudbridge_tenant_available_ips",
			Help: "Number of available IPs per tenant",
		}, []string{"tenant_id"}),

		// Метрики политик
		PolicyCount: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "cloudbridge_tenant_policy_count",
			Help: "Number of policies per tenant",
		}, []string{"tenant_id", "policy_type"}),

		PolicyStatus: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "cloudbridge_tenant_policy_status",
			Help: "Status of policies",
		}, []string{"tenant_id", "policy_name", "status"}),

		// Метрики производительности
		PolicyReconcileDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "cloudbridge_tenant_policy_reconcile_duration_seconds",
			Help:    "Duration of policy reconciliation",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
		}, []string{"tenant_id", "operation"}),

		IPAllocationDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "cloudbridge_tenant_ip_allocation_duration_seconds",
			Help:    "Duration of IP allocation",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
		}, []string{"tenant_id", "operation"}),
	}
}

// UpdateTenantMetrics - обновление метрик тенанта
func (m *TenantIsolationMetrics) UpdateTenantMetrics(tenantID string, subnetCount, allocatedIPs, availableIPs int) {
	m.TenantSubnetCount.WithLabelValues(tenantID).Set(float64(subnetCount))
	m.TenantAllocatedIPs.WithLabelValues(tenantID).Set(float64(allocatedIPs))
	m.TenantAvailableIPs.WithLabelValues(tenantID).Set(float64(availableIPs))
}

// UpdateGeneralMetrics - обновление общих метрик
func (m *TenantIsolationMetrics) UpdateGeneralMetrics(tenantCount, subnetCount, allocatedCount, availableCount int) {
	m.TotalTenants.Set(float64(tenantCount))
	m.TotalSubnets.Set(float64(subnetCount))
	m.TotalAllocated.Set(float64(allocatedCount))
	m.TotalAvailable.Set(float64(availableCount))
}

// RecordIsolationViolation - регистрация нарушения изоляции
func (m *TenantIsolationMetrics) RecordIsolationViolation(srcTenant, dstTenant, protocol, violationType string) {
	m.IsolationViolations.WithLabelValues(srcTenant, dstTenant, protocol, violationType).Inc()

	// Обновляем статус здоровья тенантов
	m.TenantHealth.WithLabelValues(srcTenant, "isolation_violation", "network").Set(0)
	m.TenantHealth.WithLabelValues(dstTenant, "isolation_violation", "network").Set(0)
}

// RecordPolicyViolation - регистрация нарушения политики
func (m *TenantIsolationMetrics) RecordPolicyViolation(tenantID, policyType, violationReason string) {
	m.PolicyViolations.WithLabelValues(tenantID, policyType, violationReason).Inc()
	m.TenantHealth.WithLabelValues(tenantID, "policy_violation", "security").Set(0)
}

// RecordTenantHealth - регистрация здоровья тенанта
func (m *TenantIsolationMetrics) RecordTenantHealth(tenantID, status, component string) {
	healthValue := 0.0
	if status == "healthy" {
		healthValue = 1.0
	}
	m.TenantHealth.WithLabelValues(tenantID, status, component).Set(healthValue)
}

// RecordPolicyCount - регистрация количества политик
func (m *TenantIsolationMetrics) RecordPolicyCount(tenantID, policyType string, count int) {
	m.PolicyCount.WithLabelValues(tenantID, policyType).Set(float64(count))
}

// RecordPolicyStatus - регистрация статуса политики
func (m *TenantIsolationMetrics) RecordPolicyStatus(tenantID, policyName, status string) {
	statusValue := 0.0
	if status == "active" {
		statusValue = 1.0
	}
	m.PolicyStatus.WithLabelValues(tenantID, policyName, status).Set(statusValue)
}

// RecordReconcileDuration - регистрация времени reconciliation
func (m *TenantIsolationMetrics) RecordReconcileDuration(tenantID, operation string, duration float64) {
	m.PolicyReconcileDuration.WithLabelValues(tenantID, operation).Observe(duration)
}

// RecordIPAllocationDuration - регистрация времени выделения IP
func (m *TenantIsolationMetrics) RecordIPAllocationDuration(tenantID, operation string, duration float64) {
	m.IPAllocationDuration.WithLabelValues(tenantID, operation).Observe(duration)
}

// ResetTenantMetrics - сброс метрик тенанта
func (m *TenantIsolationMetrics) ResetTenantMetrics(tenantID string) {
	m.TenantSubnetCount.DeleteLabelValues(tenantID)
	m.TenantAllocatedIPs.DeleteLabelValues(tenantID)
	m.TenantAvailableIPs.DeleteLabelValues(tenantID)
	m.PolicyCount.DeleteLabelValues(tenantID, "allow-intra")
	m.PolicyCount.DeleteLabelValues(tenantID, "allow-dns")
	m.PolicyCount.DeleteLabelValues(tenantID, "default-deny")
	m.PolicyStatus.DeleteLabelValues(tenantID, "tenant-"+tenantID+"-allow-intra", "active")
	m.PolicyStatus.DeleteLabelValues(tenantID, "tenant-"+tenantID+"-allow-dns", "active")
	m.PolicyStatus.DeleteLabelValues(tenantID, "tenant-"+tenantID+"-default-deny", "active")
	m.TenantHealth.DeleteLabelValues(tenantID, "healthy", "network")
	m.TenantHealth.DeleteLabelValues(tenantID, "healthy", "security")
}


