package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// AI Connection Metrics
	AIConnectionsProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_connections_processed_total",
			Help: "Total number of connections processed by AI",
		},
		[]string{"status", "anomaly_detected", "service"},
	)
	
	AIProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ai_processing_duration_seconds",
			Help:    "Time spent processing connections through AI",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "operation"},
	)
	
	// Anomaly Detection Metrics
	AnomaliesDetected = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_anomalies_detected_total",
			Help: "Total number of anomalies detected",
		},
		[]string{"severity", "type", "source"},
	)
	
	AnomalyDetectionAccuracy = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ai_anomaly_detection_accuracy",
			Help: "Accuracy of anomaly detection model",
		},
		[]string{"model_version"},
	)
	
	// Load Prediction Metrics
	LoadPredictionsMade = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_load_predictions_total",
			Help: "Total number of load predictions made",
		},
		[]string{"horizon", "accuracy_range"},
	)
	
	LoadPredictionAccuracy = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ai_load_prediction_accuracy",
			Help: "Accuracy of load prediction model",
		},
		[]string{"model_version", "horizon"},
	)
	
	// Route Optimization Metrics
	RoutesOptimized = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_routes_optimized_total",
			Help: "Total number of routes optimized",
		},
		[]string{"protocol", "optimization_type"},
	)
	
	RouteOptimizationLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ai_route_optimization_latency_seconds",
			Help:    "Time spent optimizing routes",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0},
		},
		[]string{"protocol"},
	)
	
	// Cache Metrics
	AICacheHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_cache_hits_total",
			Help: "Total number of AI cache hits",
		},
		[]string{"cache_type"},
	)
	
	AICacheMisses = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_cache_misses_total",
			Help: "Total number of AI cache misses",
		},
		[]string{"cache_type"},
	)
	
	// Circuit Breaker Metrics
	AICircuitBreakerState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ai_circuit_breaker_state",
			Help: "Current state of AI circuit breaker (0=closed, 1=open, 2=half-open)",
		},
		[]string{"service"},
	)
	
	AICircuitBreakerFailures = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_circuit_breaker_failures_total",
			Help: "Total number of AI circuit breaker failures",
		},
		[]string{"service"},
	)
	
	// System Metrics
	AIServiceHealth = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ai_service_health",
			Help: "Health status of AI services (1=healthy, 0=unhealthy)",
		},
		[]string{"service", "component"},
	)
	
	AIModelVersion = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ai_model_version",
			Help: "Version of AI models in use",
		},
		[]string{"model_type", "version"},
	)
	
	// Performance Metrics
	AIMemoryUsage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ai_memory_usage_bytes",
			Help: "Memory usage of AI components",
		},
		[]string{"component"},
	)
	
	AICPUUsage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ai_cpu_usage_percent",
			Help: "CPU usage of AI components",
		},
		[]string{"component"},
	)
	
	// Business Metrics
	AICostSavings = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_cost_savings_total",
			Help: "Total cost savings from AI optimizations",
		},
		[]string{"optimization_type"},
	)
	
	AIAutomationRate = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ai_automation_rate_percent",
			Help: "Percentage of operations automated by AI",
		},
		[]string{"operation_type"},
	)
)
