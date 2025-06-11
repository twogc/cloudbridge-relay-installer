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