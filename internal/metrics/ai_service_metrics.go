package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// AIServiceMetrics представляет метрики для AI Service
type AIServiceMetrics struct {
	// Anomaly Detection
	anomalyDetectionsTotal   *prometheus.CounterVec
	anomalyDetectionsActive  prometheus.Gauge
	anomalyFalsePositives    *prometheus.CounterVec
	anomalyFalseNegatives    *prometheus.CounterVec
	anomalyAccuracy          *prometheus.GaugeVec
	anomalyDetectionDuration *prometheus.HistogramVec

	// Load Prediction
	loadPredictionsTotal     *prometheus.CounterVec
	loadPredictionErrors     *prometheus.CounterVec
	loadPredictionAccuracy   *prometheus.GaugeVec
	loadPredictionDuration   *prometheus.HistogramVec
	loadPredictionConfidence *prometheus.GaugeVec

	// Route Optimization
	routeOptimizationsTotal   *prometheus.CounterVec
	routeOptimizationGains    *prometheus.HistogramVec
	routeOptimizationDuration *prometheus.HistogramVec
	routeOptimizationSuccess  *prometheus.CounterVec

	// Data Pipeline
	dataProcessingTotal      *prometheus.CounterVec
	dataProcessingDuration   *prometheus.HistogramVec
	dataProcessingErrors     *prometheus.CounterVec
	dataProcessingThroughput *prometheus.CounterVec
	dataProcessingLatency    *prometheus.HistogramVec

	// ML Models
	modelInferencesTotal   *prometheus.CounterVec
	modelInferenceDuration *prometheus.HistogramVec
	modelAccuracy          *prometheus.GaugeVec
	modelConfidence        *prometheus.GaugeVec
	modelTrainingDuration  *prometheus.HistogramVec
	modelTrainingErrors    *prometheus.CounterVec

	// Analytics
	analyticsReportsGenerated *prometheus.CounterVec
	analyticsReportDuration   *prometheus.HistogramVec
	analyticsInsightsTotal    *prometheus.CounterVec
	analyticsInsightAccuracy  *prometheus.GaugeVec

	// Performance
	processingTimeTotal    *prometheus.CounterVec
	processingTimeDuration *prometheus.HistogramVec
	processingQueueSize    *prometheus.GaugeVec
	processingQueueLatency *prometheus.HistogramVec

	// Resource Usage
	memoryUsageBytes  *prometheus.GaugeVec
	cpuUsagePercent   *prometheus.GaugeVec
	gpuUsagePercent   *prometheus.GaugeVec
	diskUsageBytes    *prometheus.GaugeVec
	networkUsageBytes *prometheus.CounterVec

	// Cache Performance
	cacheHits      *prometheus.CounterVec
	cacheMisses    *prometheus.CounterVec
	cacheSize      *prometheus.GaugeVec
	cacheEvictions *prometheus.CounterVec

	// Error Handling
	errorsTotal           *prometheus.CounterVec
	errorRecoveryAttempts *prometheus.CounterVec
	errorRecoverySuccess  *prometheus.CounterVec
	circuitBreakerEvents  *prometheus.CounterVec

	// Monitoring
	healthChecksTotal   *prometheus.CounterVec
	healthCheckDuration *prometheus.HistogramVec
	healthStatus        *prometheus.GaugeVec
	uptimeSeconds       prometheus.Gauge

	// Configuration
	configUpdates *prometheus.CounterVec
	configErrors  *prometheus.CounterVec
	featureFlags  *prometheus.GaugeVec
}

// NewAIServiceMetrics создает новые AI Service метрики
func NewAIServiceMetrics(reg prometheus.Registerer) *AIServiceMetrics {
	m := &AIServiceMetrics{
		// Anomaly Detection
		anomalyDetectionsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_ai_service_anomaly_detections_total",
			Help: "Total number of anomaly detections",
		}, []string{"model", "type", "severity"}),
		anomalyDetectionsActive: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "relay_ai_service_anomaly_detections_active",
			Help: "Number of active anomaly detections",
		}),
		anomalyFalsePositives: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_ai_service_anomaly_false_positives_total",
			Help: "Total number of false positive anomaly detections",
		}, []string{"model", "type"}),
		anomalyFalseNegatives: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_ai_service_anomaly_false_negatives_total",
			Help: "Total number of false negative anomaly detections",
		}, []string{"model", "type"}),
		anomalyAccuracy: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_ai_service_anomaly_accuracy",
			Help: "Anomaly detection accuracy",
		}, []string{"model", "type"}),
		anomalyDetectionDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "relay_ai_service_anomaly_detection_duration_seconds",
			Help:    "Anomaly detection duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.2, 0.5, 1.0, 2.0, 5.0},
		}, []string{"model", "type"}),

		// Load Prediction
		loadPredictionsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_ai_service_load_predictions_total",
			Help: "Total number of load predictions",
		}, []string{"model", "prediction_type"}),
		loadPredictionErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_ai_service_load_prediction_errors_total",
			Help: "Total number of load prediction errors",
		}, []string{"model", "error_type"}),
		loadPredictionAccuracy: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_ai_service_load_prediction_accuracy",
			Help: "Load prediction accuracy",
		}, []string{"model", "prediction_type"}),
		loadPredictionDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "relay_ai_service_load_prediction_duration_seconds",
			Help:    "Load prediction duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.2, 0.5, 1.0, 2.0, 5.0},
		}, []string{"model", "prediction_type"}),
		loadPredictionConfidence: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_ai_service_load_prediction_confidence",
			Help: "Load prediction confidence score",
		}, []string{"model", "prediction_type"}),

		// Route Optimization
		routeOptimizationsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_ai_service_route_optimizations_total",
			Help: "Total number of route optimizations",
		}, []string{"algorithm", "optimization_type"}),
		routeOptimizationGains: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "relay_ai_service_route_optimization_gains",
			Help:    "Route optimization performance gains",
			Buckets: []float64{0.01, 0.05, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0},
		}, []string{"algorithm", "optimization_type"}),
		routeOptimizationDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "relay_ai_service_route_optimization_duration_seconds",
			Help:    "Route optimization duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.2, 0.5, 1.0, 2.0, 5.0},
		}, []string{"algorithm", "optimization_type"}),
		routeOptimizationSuccess: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_ai_service_route_optimization_success_total",
			Help: "Total number of successful route optimizations",
		}, []string{"algorithm", "optimization_type"}),

		// Data Pipeline
		dataProcessingTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_ai_service_data_processing_total",
			Help: "Total number of data processing operations",
		}, []string{"pipeline", "stage", "data_type"}),
		dataProcessingDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "relay_ai_service_data_processing_duration_seconds",
			Help:    "Data processing duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.2, 0.5, 1.0, 2.0, 5.0, 10.0},
		}, []string{"pipeline", "stage", "data_type"}),
		dataProcessingErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_ai_service_data_processing_errors_total",
			Help: "Total number of data processing errors",
		}, []string{"pipeline", "stage", "error_type"}),
		dataProcessingThroughput: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_ai_service_data_processing_throughput_total",
			Help: "Total data processing throughput",
		}, []string{"pipeline", "stage", "data_type", "direction"}),
		dataProcessingLatency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "relay_ai_service_data_processing_latency_seconds",
			Help:    "Data processing latency in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.2, 0.5, 1.0, 2.0, 5.0},
		}, []string{"pipeline", "stage", "data_type"}),

		// ML Models
		modelInferencesTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_ai_service_model_inferences_total",
			Help: "Total number of ML model inferences",
		}, []string{"model", "model_type", "version"}),
		modelInferenceDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "relay_ai_service_model_inference_duration_seconds",
			Help:    "ML model inference duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.2, 0.5, 1.0, 2.0, 5.0},
		}, []string{"model", "model_type", "version"}),
		modelAccuracy: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_ai_service_model_accuracy",
			Help: "ML model accuracy",
		}, []string{"model", "model_type", "version"}),
		modelConfidence: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_ai_service_model_confidence",
			Help: "ML model confidence score",
		}, []string{"model", "model_type", "version"}),
		modelTrainingDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "relay_ai_service_model_training_duration_seconds",
			Help:    "ML model training duration in seconds",
			Buckets: []float64{1, 5, 10, 30, 60, 300, 600, 1800, 3600, 7200},
		}, []string{"model", "model_type", "version"}),
		modelTrainingErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_ai_service_model_training_errors_total",
			Help: "Total number of ML model training errors",
		}, []string{"model", "model_type", "version", "error_type"}),

		// Analytics
		analyticsReportsGenerated: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_ai_service_analytics_reports_generated_total",
			Help: "Total number of analytics reports generated",
		}, []string{"report_type", "format"}),
		analyticsReportDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "relay_ai_service_analytics_report_duration_seconds",
			Help:    "Analytics report generation duration in seconds",
			Buckets: []float64{0.1, 0.5, 1.0, 2.0, 5.0, 10.0, 30.0, 60.0, 300.0},
		}, []string{"report_type", "format"}),
		analyticsInsightsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_ai_service_analytics_insights_total",
			Help: "Total number of analytics insights generated",
		}, []string{"insight_type", "category"}),
		analyticsInsightAccuracy: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_ai_service_analytics_insight_accuracy",
			Help: "Analytics insight accuracy",
		}, []string{"insight_type", "category"}),

		// Performance
		processingTimeTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_ai_service_processing_time_total",
			Help: "Total AI processing time",
		}, []string{"component", "operation"}),
		processingTimeDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "relay_ai_service_processing_time_duration_seconds",
			Help:    "AI processing time duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.2, 0.5, 1.0, 2.0, 5.0, 10.0},
		}, []string{"component", "operation"}),
		processingQueueSize: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_ai_service_processing_queue_size",
			Help: "AI processing queue size",
		}, []string{"queue_type", "priority"}),
		processingQueueLatency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "relay_ai_service_processing_queue_latency_seconds",
			Help:    "AI processing queue latency in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.2, 0.5, 1.0, 2.0, 5.0},
		}, []string{"queue_type", "priority"}),

		// Resource Usage
		memoryUsageBytes: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_ai_service_memory_usage_bytes",
			Help: "AI service memory usage in bytes",
		}, []string{"component", "usage_type"}),
		cpuUsagePercent: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_ai_service_cpu_usage_percent",
			Help: "AI service CPU usage percentage",
		}, []string{"component", "usage_type"}),
		gpuUsagePercent: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_ai_service_gpu_usage_percent",
			Help: "AI service GPU usage percentage",
		}, []string{"component", "usage_type"}),
		diskUsageBytes: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_ai_service_disk_usage_bytes",
			Help: "AI service disk usage in bytes",
		}, []string{"component", "usage_type"}),
		networkUsageBytes: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_ai_service_network_usage_bytes_total",
			Help: "AI service network usage in bytes",
		}, []string{"component", "direction"}),

		// Cache Performance
		cacheHits: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_ai_service_cache_hits_total",
			Help: "Total number of AI cache hits",
		}, []string{"cache_type", "component"}),
		cacheMisses: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_ai_service_cache_misses_total",
			Help: "Total number of AI cache misses",
		}, []string{"cache_type", "component"}),
		cacheSize: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_ai_service_cache_size",
			Help: "AI cache size",
		}, []string{"cache_type", "component"}),
		cacheEvictions: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_ai_service_cache_evictions_total",
			Help: "Total number of AI cache evictions",
		}, []string{"cache_type", "component", "reason"}),

		// Error Handling
		errorsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_ai_service_errors_total",
			Help: "Total number of AI service errors",
		}, []string{"component", "error_type", "severity"}),
		errorRecoveryAttempts: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_ai_service_error_recovery_attempts_total",
			Help: "Total number of AI error recovery attempts",
		}, []string{"component", "error_type"}),
		errorRecoverySuccess: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_ai_service_error_recovery_success_total",
			Help: "Total number of successful AI error recoveries",
		}, []string{"component", "error_type"}),
		circuitBreakerEvents: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_ai_service_circuit_breaker_events_total",
			Help: "Total number of AI circuit breaker events",
		}, []string{"component", "event_type"}),

		// Monitoring
		healthChecksTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_ai_service_health_checks_total",
			Help: "Total number of AI health checks",
		}, []string{"component", "status"}),
		healthCheckDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "relay_ai_service_health_check_duration_seconds",
			Help:    "AI health check duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.2, 0.5, 1.0},
		}, []string{"component"}),
		healthStatus: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_ai_service_health_status",
			Help: "AI service health status (1=healthy, 0=unhealthy)",
		}, []string{"component"}),
		uptimeSeconds: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "relay_ai_service_uptime_seconds",
			Help: "AI service uptime in seconds",
		}),

		// Configuration
		configUpdates: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_ai_service_config_updates_total",
			Help: "Total number of AI configuration updates",
		}, []string{"component", "update_type"}),
		configErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "relay_ai_service_config_errors_total",
			Help: "Total number of AI configuration errors",
		}, []string{"component", "error_type"}),
		featureFlags: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "relay_ai_service_feature_flags",
			Help: "AI service feature flags (1=enabled, 0=disabled)",
		}, []string{"feature", "component"}),
	}

	// Регистрация метрик
	reg.MustRegister(
		m.anomalyDetectionsTotal,
		m.anomalyDetectionsActive,
		m.anomalyFalsePositives,
		m.anomalyFalseNegatives,
		m.anomalyAccuracy,
		m.anomalyDetectionDuration,
		m.loadPredictionsTotal,
		m.loadPredictionErrors,
		m.loadPredictionAccuracy,
		m.loadPredictionDuration,
		m.loadPredictionConfidence,
		m.routeOptimizationsTotal,
		m.routeOptimizationGains,
		m.routeOptimizationDuration,
		m.routeOptimizationSuccess,
		m.dataProcessingTotal,
		m.dataProcessingDuration,
		m.dataProcessingErrors,
		m.dataProcessingThroughput,
		m.dataProcessingLatency,
		m.modelInferencesTotal,
		m.modelInferenceDuration,
		m.modelAccuracy,
		m.modelConfidence,
		m.modelTrainingDuration,
		m.modelTrainingErrors,
		m.analyticsReportsGenerated,
		m.analyticsReportDuration,
		m.analyticsInsightsTotal,
		m.analyticsInsightAccuracy,
		m.processingTimeTotal,
		m.processingTimeDuration,
		m.processingQueueSize,
		m.processingQueueLatency,
		m.memoryUsageBytes,
		m.cpuUsagePercent,
		m.gpuUsagePercent,
		m.diskUsageBytes,
		m.networkUsageBytes,
		m.cacheHits,
		m.cacheMisses,
		m.cacheSize,
		m.cacheEvictions,
		m.errorsTotal,
		m.errorRecoveryAttempts,
		m.errorRecoverySuccess,
		m.circuitBreakerEvents,
		m.healthChecksTotal,
		m.healthCheckDuration,
		m.healthStatus,
		m.uptimeSeconds,
		m.configUpdates,
		m.configErrors,
		m.featureFlags,
	)

	return m
}

// Методы для работы с AI Service метриками

// Anomaly Detection Methods
func (m *AIServiceMetrics) IncAnomalyDetections(model, anomalyType, severity string) {
	m.anomalyDetectionsTotal.WithLabelValues(model, anomalyType, severity).Inc()
}

func (m *AIServiceMetrics) SetAnomalyDetectionsActive(count int) {
	m.anomalyDetectionsActive.Set(float64(count))
}

func (m *AIServiceMetrics) IncAnomalyFalsePositives(model, anomalyType string) {
	m.anomalyFalsePositives.WithLabelValues(model, anomalyType).Inc()
}

func (m *AIServiceMetrics) IncAnomalyFalseNegatives(model, anomalyType string) {
	m.anomalyFalseNegatives.WithLabelValues(model, anomalyType).Inc()
}

func (m *AIServiceMetrics) SetAnomalyAccuracy(model, anomalyType string, accuracy float64) {
	m.anomalyAccuracy.WithLabelValues(model, anomalyType).Set(accuracy)
}

func (m *AIServiceMetrics) ObserveAnomalyDetectionDuration(model, anomalyType string, duration float64) {
	m.anomalyDetectionDuration.WithLabelValues(model, anomalyType).Observe(duration)
}

// Load Prediction Methods
func (m *AIServiceMetrics) IncLoadPredictions(model, predictionType string) {
	m.loadPredictionsTotal.WithLabelValues(model, predictionType).Inc()
}

func (m *AIServiceMetrics) IncLoadPredictionErrors(model, errorType string) {
	m.loadPredictionErrors.WithLabelValues(model, errorType).Inc()
}

func (m *AIServiceMetrics) SetLoadPredictionAccuracy(model, predictionType string, accuracy float64) {
	m.loadPredictionAccuracy.WithLabelValues(model, predictionType).Set(accuracy)
}

func (m *AIServiceMetrics) ObserveLoadPredictionDuration(model, predictionType string, duration float64) {
	m.loadPredictionDuration.WithLabelValues(model, predictionType).Observe(duration)
}

func (m *AIServiceMetrics) SetLoadPredictionConfidence(model, predictionType string, confidence float64) {
	m.loadPredictionConfidence.WithLabelValues(model, predictionType).Set(confidence)
}

// Route Optimization Methods
func (m *AIServiceMetrics) IncRouteOptimizations(algorithm, optimizationType string) {
	m.routeOptimizationsTotal.WithLabelValues(algorithm, optimizationType).Inc()
}

func (m *AIServiceMetrics) ObserveRouteOptimizationGains(algorithm, optimizationType string, gain float64) {
	m.routeOptimizationGains.WithLabelValues(algorithm, optimizationType).Observe(gain)
}

func (m *AIServiceMetrics) ObserveRouteOptimizationDuration(algorithm, optimizationType string, duration float64) {
	m.routeOptimizationDuration.WithLabelValues(algorithm, optimizationType).Observe(duration)
}

func (m *AIServiceMetrics) IncRouteOptimizationSuccess(algorithm, optimizationType string) {
	m.routeOptimizationSuccess.WithLabelValues(algorithm, optimizationType).Inc()
}

// Data Pipeline Methods
func (m *AIServiceMetrics) IncDataProcessing(pipeline, stage, dataType string) {
	m.dataProcessingTotal.WithLabelValues(pipeline, stage, dataType).Inc()
}

func (m *AIServiceMetrics) ObserveDataProcessingDuration(pipeline, stage, dataType string, duration float64) {
	m.dataProcessingDuration.WithLabelValues(pipeline, stage, dataType).Observe(duration)
}

func (m *AIServiceMetrics) IncDataProcessingErrors(pipeline, stage, errorType string) {
	m.dataProcessingErrors.WithLabelValues(pipeline, stage, errorType).Inc()
}

func (m *AIServiceMetrics) IncDataProcessingThroughput(pipeline, stage, dataType, direction string, count int64) {
	m.dataProcessingThroughput.WithLabelValues(pipeline, stage, dataType, direction).Add(float64(count))
}

func (m *AIServiceMetrics) ObserveDataProcessingLatency(pipeline, stage, dataType string, latency float64) {
	m.dataProcessingLatency.WithLabelValues(pipeline, stage, dataType).Observe(latency)
}

// ML Model Methods
func (m *AIServiceMetrics) IncModelInferences(model, modelType, version string) {
	m.modelInferencesTotal.WithLabelValues(model, modelType, version).Inc()
}

func (m *AIServiceMetrics) ObserveModelInferenceDuration(model, modelType, version string, duration float64) {
	m.modelInferenceDuration.WithLabelValues(model, modelType, version).Observe(duration)
}

func (m *AIServiceMetrics) SetModelAccuracy(model, modelType, version string, accuracy float64) {
	m.modelAccuracy.WithLabelValues(model, modelType, version).Set(accuracy)
}

func (m *AIServiceMetrics) SetModelConfidence(model, modelType, version string, confidence float64) {
	m.modelConfidence.WithLabelValues(model, modelType, version).Set(confidence)
}

func (m *AIServiceMetrics) ObserveModelTrainingDuration(model, modelType, version string, duration float64) {
	m.modelTrainingDuration.WithLabelValues(model, modelType, version).Observe(duration)
}

func (m *AIServiceMetrics) IncModelTrainingErrors(model, modelType, version, errorType string) {
	m.modelTrainingErrors.WithLabelValues(model, modelType, version, errorType).Inc()
}

// Analytics Methods
func (m *AIServiceMetrics) IncAnalyticsReportsGenerated(reportType, format string) {
	m.analyticsReportsGenerated.WithLabelValues(reportType, format).Inc()
}

func (m *AIServiceMetrics) ObserveAnalyticsReportDuration(reportType, format string, duration float64) {
	m.analyticsReportDuration.WithLabelValues(reportType, format).Observe(duration)
}

func (m *AIServiceMetrics) IncAnalyticsInsights(insightType, category string) {
	m.analyticsInsightsTotal.WithLabelValues(insightType, category).Inc()
}

func (m *AIServiceMetrics) SetAnalyticsInsightAccuracy(insightType, category string, accuracy float64) {
	m.analyticsInsightAccuracy.WithLabelValues(insightType, category).Set(accuracy)
}

// Performance Methods
func (m *AIServiceMetrics) IncProcessingTime(component, operation string, duration float64) {
	m.processingTimeTotal.WithLabelValues(component, operation).Add(duration)
}

func (m *AIServiceMetrics) ObserveProcessingTimeDuration(component, operation string, duration float64) {
	m.processingTimeDuration.WithLabelValues(component, operation).Observe(duration)
}

func (m *AIServiceMetrics) SetProcessingQueueSize(queueType, priority string, size int) {
	m.processingQueueSize.WithLabelValues(queueType, priority).Set(float64(size))
}

func (m *AIServiceMetrics) ObserveProcessingQueueLatency(queueType, priority string, latency float64) {
	m.processingQueueLatency.WithLabelValues(queueType, priority).Observe(latency)
}

// Resource Usage Methods
func (m *AIServiceMetrics) SetMemoryUsage(component, usageType string, bytes int64) {
	m.memoryUsageBytes.WithLabelValues(component, usageType).Set(float64(bytes))
}

func (m *AIServiceMetrics) SetCPUUsage(component, usageType string, percent float64) {
	m.cpuUsagePercent.WithLabelValues(component, usageType).Set(percent)
}

func (m *AIServiceMetrics) SetGPUUsage(component, usageType string, percent float64) {
	m.gpuUsagePercent.WithLabelValues(component, usageType).Set(percent)
}

func (m *AIServiceMetrics) SetDiskUsage(component, usageType string, bytes int64) {
	m.diskUsageBytes.WithLabelValues(component, usageType).Set(float64(bytes))
}

func (m *AIServiceMetrics) IncNetworkUsage(component, direction string, bytes int64) {
	m.networkUsageBytes.WithLabelValues(component, direction).Add(float64(bytes))
}

// Cache Performance Methods
func (m *AIServiceMetrics) IncCacheHits(cacheType, component string) {
	m.cacheHits.WithLabelValues(cacheType, component).Inc()
}

func (m *AIServiceMetrics) IncCacheMisses(cacheType, component string) {
	m.cacheMisses.WithLabelValues(cacheType, component).Inc()
}

func (m *AIServiceMetrics) SetCacheSize(cacheType, component string, size int) {
	m.cacheSize.WithLabelValues(cacheType, component).Set(float64(size))
}

func (m *AIServiceMetrics) IncCacheEvictions(cacheType, component, reason string) {
	m.cacheEvictions.WithLabelValues(cacheType, component, reason).Inc()
}

// Error Handling Methods
func (m *AIServiceMetrics) IncErrors(component, errorType, severity string) {
	m.errorsTotal.WithLabelValues(component, errorType, severity).Inc()
}

func (m *AIServiceMetrics) IncErrorRecoveryAttempts(component, errorType string) {
	m.errorRecoveryAttempts.WithLabelValues(component, errorType).Inc()
}

func (m *AIServiceMetrics) IncErrorRecoverySuccess(component, errorType string) {
	m.errorRecoverySuccess.WithLabelValues(component, errorType).Inc()
}

func (m *AIServiceMetrics) IncCircuitBreakerEvents(component, eventType string) {
	m.circuitBreakerEvents.WithLabelValues(component, eventType).Inc()
}

// Monitoring Methods
func (m *AIServiceMetrics) IncHealthChecks(component, status string) {
	m.healthChecksTotal.WithLabelValues(component, status).Inc()
}

func (m *AIServiceMetrics) ObserveHealthCheckDuration(component string, duration float64) {
	m.healthCheckDuration.WithLabelValues(component).Observe(duration)
}

func (m *AIServiceMetrics) SetHealthStatus(component string, healthy bool) {
	value := 0.0
	if healthy {
		value = 1.0
	}
	m.healthStatus.WithLabelValues(component).Set(value)
}

func (m *AIServiceMetrics) SetUptimeSeconds(uptime float64) {
	m.uptimeSeconds.Set(uptime)
}

// Configuration Methods
func (m *AIServiceMetrics) IncConfigUpdates(component, updateType string) {
	m.configUpdates.WithLabelValues(component, updateType).Inc()
}

func (m *AIServiceMetrics) IncConfigErrors(component, errorType string) {
	m.configErrors.WithLabelValues(component, errorType).Inc()
}

func (m *AIServiceMetrics) SetFeatureFlag(feature, component string, enabled bool) {
	value := 0.0
	if enabled {
		value = 1.0
	}
	m.featureFlags.WithLabelValues(feature, component).Set(value)
}



