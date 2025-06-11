package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// ActiveTunnels tracks the number of active tunnels
	ActiveTunnels = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "relay_active_tunnels",
		Help: "The number of active tunnels",
	})

	// TotalTunnels tracks the total number of tunnels created
	TotalTunnels = promauto.NewCounter(prometheus.CounterOpts{
		Name: "relay_total_tunnels",
		Help: "The total number of tunnels created",
	})

	// FailedTunnels tracks the number of failed tunnel creations
	FailedTunnels = promauto.NewCounter(prometheus.CounterOpts{
		Name: "relay_failed_tunnels",
		Help: "The number of failed tunnel creations",
	})

	// TokenValidations tracks token validation metrics
	TokenValidations = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "relay_token_validations",
		Help: "The number of token validations",
	}, []string{"method", "result"})

	// TokenValidationDuration tracks token validation duration
	TokenValidationDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "relay_token_validation_duration_seconds",
		Help:    "The duration of token validations in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method"})

	// MessageLatency tracks message processing latency
	MessageLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "relay_message_latency_seconds",
		Help:    "The latency of message processing in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"type"})

	// MessageErrors tracks message processing errors
	MessageErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "relay_message_errors",
		Help: "The number of message processing errors",
	}, []string{"type", "error"})

	// CacheSize tracks the size of the token cache
	CacheSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "relay_token_cache_size",
		Help: "The number of tokens in the cache",
	})

	// CacheHits tracks token cache hits
	CacheHits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "relay_token_cache_hits",
		Help: "The number of token cache hits",
	})

	// CacheMisses tracks token cache misses
	CacheMisses = promauto.NewCounter(prometheus.CounterOpts{
		Name: "relay_token_cache_misses",
		Help: "The number of token cache misses",
	})
)

// Metrics represents server metrics
type Metrics struct {
	activeConnections    prometheus.Gauge
	rejectedConnections  prometheus.Counter
	connectionDuration   prometheus.Histogram
	bytesTransferred     prometheus.Counter
	errors              prometheus.Counter
	latency             prometheus.Histogram
}

// NewMetrics creates new metrics
func NewMetrics() *Metrics {
	return &Metrics{
		activeConnections: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "relay_active_connections",
			Help: "Number of active connections",
		}),
		rejectedConnections: promauto.NewCounter(prometheus.CounterOpts{
			Name: "relay_rejected_connections",
			Help: "Number of rejected connections",
		}),
		connectionDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "relay_connection_duration_seconds",
			Help:    "Connection duration in seconds",
			Buckets: prometheus.DefBuckets,
		}),
		bytesTransferred: promauto.NewCounter(prometheus.CounterOpts{
			Name: "relay_bytes_transferred",
			Help: "Total bytes transferred",
		}),
		errors: promauto.NewCounter(prometheus.CounterOpts{
			Name: "relay_errors",
			Help: "Number of errors",
		}),
		latency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "relay_latency_seconds",
			Help:    "Request latency in seconds",
			Buckets: prometheus.DefBuckets,
		}),
	}
}

// IncActiveConnections increments active connections counter
func (m *Metrics) IncActiveConnections() {
	m.activeConnections.Inc()
}

// DecActiveConnections decrements active connections counter
func (m *Metrics) DecActiveConnections() {
	m.activeConnections.Dec()
}

// IncRejectedConnections increments rejected connections counter
func (m *Metrics) IncRejectedConnections() {
	m.rejectedConnections.Inc()
}

// ObserveConnectionDuration records connection duration
func (m *Metrics) ObserveConnectionDuration(duration float64) {
	m.connectionDuration.Observe(duration)
}

// AddBytesTransferred adds bytes to transferred counter
func (m *Metrics) AddBytesTransferred(bytes float64) {
	m.bytesTransferred.Add(bytes)
}

// IncErrors increments errors counter
func (m *Metrics) IncErrors() {
	m.errors.Inc()
}

// ObserveLatency records request latency
func (m *Metrics) ObserveLatency(latency float64) {
	m.latency.Observe(latency)
} 