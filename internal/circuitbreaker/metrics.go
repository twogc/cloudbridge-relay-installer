package circuitbreaker

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// CircuitBreakerMetrics holds Prometheus metrics for circuit breakers
type CircuitBreakerMetrics struct {
	// State metrics
	StateChanges *prometheus.CounterVec
	CurrentState *prometheus.GaugeVec

	// Request metrics
	TotalRequests  *prometheus.CounterVec
	TotalFailures  *prometheus.CounterVec
	TotalSuccesses *prometheus.CounterVec
	TotalTimeouts  *prometheus.CounterVec

	// Performance metrics
	SuccessRate  *prometheus.GaugeVec
	FailureRate  *prometheus.GaugeVec
	ResponseTime *prometheus.HistogramVec

	// Recovery metrics
	RecoveryAttempts *prometheus.CounterVec
	RecoveryTime     *prometheus.HistogramVec
	RecoveryProgress *prometheus.GaugeVec

	// Configuration metrics
	FailureThreshold *prometheus.GaugeVec
	SuccessThreshold *prometheus.GaugeVec
	Timeout          *prometheus.GaugeVec
}

// NewCircuitBreakerMetrics creates new circuit breaker metrics
func NewCircuitBreakerMetrics() *CircuitBreakerMetrics {
	return &CircuitBreakerMetrics{
		StateChanges: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "circuit_breaker_state_changes_total",
				Help: "Total number of circuit breaker state changes",
			},
			[]string{"name", "from_state", "to_state"},
		),

		CurrentState: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "circuit_breaker_current_state",
				Help: "Current state of the circuit breaker (0=Closed, 1=Open, 2=HalfOpen)",
			},
			[]string{"name"},
		),

		TotalRequests: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "circuit_breaker_requests_total",
				Help: "Total number of requests through the circuit breaker",
			},
			[]string{"name"},
		),

		TotalFailures: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "circuit_breaker_failures_total",
				Help: "Total number of failures through the circuit breaker",
			},
			[]string{"name"},
		),

		TotalSuccesses: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "circuit_breaker_successes_total",
				Help: "Total number of successful requests through the circuit breaker",
			},
			[]string{"name"},
		),

		TotalTimeouts: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "circuit_breaker_timeouts_total",
				Help: "Total number of timeouts through the circuit breaker",
			},
			[]string{"name"},
		),

		SuccessRate: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "circuit_breaker_success_rate",
				Help: "Success rate of the circuit breaker (0.0 to 1.0)",
			},
			[]string{"name"},
		),

		FailureRate: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "circuit_breaker_failure_rate",
				Help: "Failure rate of the circuit breaker (0.0 to 1.0)",
			},
			[]string{"name"},
		),

		ResponseTime: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "circuit_breaker_response_time_seconds",
				Help:    "Response time of requests through the circuit breaker",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"name", "state"},
		),

		RecoveryAttempts: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "circuit_breaker_recovery_attempts_total",
				Help: "Total number of recovery attempts",
			},
			[]string{"name"},
		),

		RecoveryTime: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "circuit_breaker_recovery_time_seconds",
				Help:    "Time taken for circuit breaker recovery",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"name"},
		),

		RecoveryProgress: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "circuit_breaker_recovery_progress",
				Help: "Recovery progress of the circuit breaker (0.0 to 1.0)",
			},
			[]string{"name"},
		),

		FailureThreshold: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "circuit_breaker_failure_threshold",
				Help: "Failure threshold configuration of the circuit breaker",
			},
			[]string{"name"},
		),

		SuccessThreshold: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "circuit_breaker_success_threshold",
				Help: "Success threshold configuration of the circuit breaker",
			},
			[]string{"name"},
		),

		Timeout: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "circuit_breaker_timeout_seconds",
				Help: "Timeout configuration of the circuit breaker in seconds",
			},
			[]string{"name"},
		),
	}
}

// MetricsManager manages circuit breaker metrics
type MetricsManager struct {
	metrics *CircuitBreakerMetrics
}

// NewMetricsManager creates a new metrics manager
func NewMetricsManager() *MetricsManager {
	return &MetricsManager{
		metrics: NewCircuitBreakerMetrics(),
	}
}

// RecordStateChange records a state change
func (mm *MetricsManager) RecordStateChange(name, fromState, toState string) {
	mm.metrics.StateChanges.WithLabelValues(name, fromState, toState).Inc()
}

// SetCurrentState sets the current state
func (mm *MetricsManager) SetCurrentState(name string, state State) {
	mm.metrics.CurrentState.WithLabelValues(name).Set(float64(state))
}

// RecordRequest records a request
func (mm *MetricsManager) RecordRequest(name string) {
	mm.metrics.TotalRequests.WithLabelValues(name).Inc()
}

// RecordFailure records a failure
func (mm *MetricsManager) RecordFailure(name string) {
	mm.metrics.TotalFailures.WithLabelValues(name).Inc()
}

// RecordSuccess records a success
func (mm *MetricsManager) RecordSuccess(name string) {
	mm.metrics.TotalSuccesses.WithLabelValues(name).Inc()
}

// RecordTimeout records a timeout
func (mm *MetricsManager) RecordTimeout(name string) {
	mm.metrics.TotalTimeouts.WithLabelValues(name).Inc()
}

// SetSuccessRate sets the success rate
func (mm *MetricsManager) SetSuccessRate(name string, rate float64) {
	mm.metrics.SuccessRate.WithLabelValues(name).Set(rate)
}

// SetFailureRate sets the failure rate
func (mm *MetricsManager) SetFailureRate(name string, rate float64) {
	mm.metrics.FailureRate.WithLabelValues(name).Set(rate)
}

// RecordResponseTime records response time
func (mm *MetricsManager) RecordResponseTime(name, state string, duration float64) {
	mm.metrics.ResponseTime.WithLabelValues(name, state).Observe(duration)
}

// RecordRecoveryAttempt records a recovery attempt
func (mm *MetricsManager) RecordRecoveryAttempt(name string) {
	mm.metrics.RecoveryAttempts.WithLabelValues(name).Inc()
}

// RecordRecoveryTime records recovery time
func (mm *MetricsManager) RecordRecoveryTime(name string, duration float64) {
	mm.metrics.RecoveryTime.WithLabelValues(name).Observe(duration)
}

// SetRecoveryProgress sets recovery progress
func (mm *MetricsManager) SetRecoveryProgress(name string, progress float64) {
	mm.metrics.RecoveryProgress.WithLabelValues(name).Set(progress)
}

// SetFailureThreshold sets failure threshold
func (mm *MetricsManager) SetFailureThreshold(name string, threshold float64) {
	mm.metrics.FailureThreshold.WithLabelValues(name).Set(threshold)
}

// SetSuccessThreshold sets success threshold
func (mm *MetricsManager) SetSuccessThreshold(name string, threshold float64) {
	mm.metrics.SuccessThreshold.WithLabelValues(name).Set(threshold)
}

// SetTimeout sets timeout
func (mm *MetricsManager) SetTimeout(name string, timeout float64) {
	mm.metrics.Timeout.WithLabelValues(name).Set(timeout)
}

// UpdateAllMetrics updates all metrics for a circuit breaker
func (mm *MetricsManager) UpdateAllMetrics(cb *CircuitBreaker) {
	metrics := cb.GetMetrics()
	name := cb.GetName()

	// Update state
	state := cb.GetState()
	mm.SetCurrentState(name, state)

	// Update request counts
	if totalRequests, ok := metrics["total_requests"].(int64); ok {
		for i := int64(0); i < totalRequests; i++ {
			mm.RecordRequest(name)
		}
	}

	if totalFailures, ok := metrics["total_failures"].(int64); ok {
		for i := int64(0); i < totalFailures; i++ {
			mm.RecordFailure(name)
		}
	}

	if totalSuccesses, ok := metrics["total_successes"].(int64); ok {
		for i := int64(0); i < totalSuccesses; i++ {
			mm.RecordSuccess(name)
		}
	}

	if totalTimeouts, ok := metrics["total_timeouts"].(int64); ok {
		for i := int64(0); i < totalTimeouts; i++ {
			mm.RecordTimeout(name)
		}
	}

	// Update rates
	if successRate, ok := metrics["success_rate"].(float64); ok {
		mm.SetSuccessRate(name, successRate)
		mm.SetFailureRate(name, 1.0-successRate)
	}

	// Update recovery metrics
	if recoveryProgress, ok := metrics["recovery_progress"].(float64); ok {
		mm.SetRecoveryProgress(name, recoveryProgress)
	}

	// Update configuration
	config := cb.GetStats()
	if failureThreshold, ok := config["failure_threshold"].(int64); ok {
		mm.SetFailureThreshold(name, float64(failureThreshold))
	}

	if successThreshold, ok := config["success_threshold"].(int64); ok {
		mm.SetSuccessThreshold(name, float64(successThreshold))
	}

	if timeout, ok := config["timeout"].(time.Duration); ok {
		mm.SetTimeout(name, timeout.Seconds())
	}
}
