package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// AI Behavior Analysis Metrics
	AIBBehaviorAnalysisTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_behavior_analysis_total",
			Help: "Total number of behavior analysis operations",
		},
		[]string{"user_id", "action_type", "status"},
	)

	AIBAnomalyScore = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ai_behavior_anomaly_score",
			Help:    "Distribution of behavior anomaly scores",
			Buckets: []float64{0.0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0},
		},
		[]string{"user_id", "action_type"},
	)

	AIBAnomalyDetections = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_behavior_anomaly_detections_total",
			Help: "Total number of anomaly detections",
		},
		[]string{"user_id", "anomaly_type", "severity"},
	)

	AIBUserProfiles = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ai_behavior_user_profiles",
			Help: "Number of user behavior profiles",
		},
		[]string{"status"},
	)

	AIBModelAccuracy = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ai_behavior_model_accuracy",
			Help: "Accuracy of behavior analysis model",
		},
		[]string{"model_version"},
	)

	AIBModelTrainingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ai_behavior_model_training_duration_seconds",
			Help:    "Time taken to train behavior analysis model",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"model_version"},
	)

	AIBModelInferenceDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ai_behavior_model_inference_duration_seconds",
			Help:    "Time taken for behavior analysis inference",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"model_version"},
	)

	// AI Predictive Security Metrics
	AIPredictiveThreats = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_predictive_threats_total",
			Help: "Total number of predicted threats",
		},
		[]string{"threat_type", "confidence_level", "severity"},
	)

	AIPredictiveAccuracy = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ai_predictive_accuracy",
			Help: "Accuracy of threat predictions",
		},
		[]string{"threat_type", "model_version"},
	)

	AIPredictiveFalsePositives = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_predictive_false_positives_total",
			Help: "Total number of false positive predictions",
		},
		[]string{"threat_type", "model_version"},
	)

	AIPredictiveFalseNegatives = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_predictive_false_negatives_total",
			Help: "Total number of false negative predictions",
		},
		[]string{"threat_type", "model_version"},
	)

	AIPredictiveModelTrainingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ai_predictive_model_training_duration_seconds",
			Help:    "Time taken to train predictive security model",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"threat_type", "model_version"},
	)

	AIPredictiveModelInferenceDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ai_predictive_model_inference_duration_seconds",
			Help:    "Time taken for threat prediction inference",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"threat_type", "model_version"},
	)

	// AI Automated Response Metrics
	AIAutomatedResponses = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_automated_responses_total",
			Help: "Total number of automated responses triggered",
		},
		[]string{"response_type", "threat_type", "status"},
	)

	AIAutomatedResponseDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ai_automated_response_duration_seconds",
			Help:    "Time taken to execute automated response",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"response_type"},
	)

	AIAutomatedResponseSuccess = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ai_automated_response_success_rate",
			Help: "Success rate of automated responses",
		},
		[]string{"response_type"},
	)

	AIAutomatedResponseRules = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ai_automated_response_rules",
			Help: "Number of automated response rules",
		},
		[]string{"status"},
	)

	AIAutomatedResponseErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_automated_response_errors_total",
			Help: "Total number of automated response errors",
		},
		[]string{"response_type", "error_type"},
	)

	// AI Monitoring Metrics
	AIMonitoringAlerts = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_monitoring_alerts_total",
			Help: "Total number of AI monitoring alerts",
		},
		[]string{"alert_type", "severity", "status"},
	)

	AIMonitoringMetrics = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ai_monitoring_metrics",
			Help: "AI monitoring metrics values",
		},
		[]string{"metric_name", "component"},
	)

	AIMonitoringDashboardStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ai_monitoring_dashboard_status",
			Help: "AI monitoring dashboard status (1 = available, 0 = unavailable)",
		},
		[]string{"dashboard_url"},
	)

	AIMonitoringDataPoints = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_monitoring_data_points_total",
			Help: "Total number of data points collected for monitoring",
		},
		[]string{"data_type", "source"},
	)

	AIMonitoringDataProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ai_monitoring_data_processing_duration_seconds",
			Help:    "Time taken to process monitoring data",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"data_type"},
	)

	// AI Integration Metrics
	AIIntegrationWireGuardAnalysis = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_integration_wireguard_analysis_total",
			Help: "Total number of WireGuard behavior analyses",
		},
		[]string{"analysis_type", "status"},
	)

	AIIntegrationQUICAnalysis = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_integration_quic_analysis_total",
			Help: "Total number of QUIC behavior analyses",
		},
		[]string{"analysis_type", "status"},
	)

	AIIntegrationQuantumAnalysis = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_integration_quantum_analysis_total",
			Help: "Total number of quantum threat analyses",
		},
		[]string{"analysis_type", "status"},
	)

	AIIntegrationWireGuardThreats = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_integration_wireguard_threats_total",
			Help: "Total number of WireGuard threat responses",
		},
		[]string{"threat_type", "response_type"},
	)

	AIIntegrationQUICThreats = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_integration_quic_threats_total",
			Help: "Total number of QUIC threat responses",
		},
		[]string{"threat_type", "response_type"},
	)

	AIIntegrationQuantumThreats = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_integration_quantum_threats_total",
			Help: "Total number of quantum threat responses",
		},
		[]string{"threat_type", "response_type"},
	)

	// AI System Metrics
	AISystemStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ai_system_status",
			Help: "AI system component status (1 = running, 0 = stopped)",
		},
		[]string{"component"},
	)

	AISystemMemoryUsage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ai_system_memory_usage_bytes",
			Help: "AI system memory usage in bytes",
		},
		[]string{"component"},
	)

	AISystemCPUUsage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ai_system_cpu_usage_percent",
			Help: "AI system CPU usage percentage",
		},
		[]string{"component"},
	)

	AISystemModelLoadTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ai_system_model_load_time_seconds",
			Help:    "Time taken to load AI models",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"model_type", "model_version"},
	)

	AISystemErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_system_errors_total",
			Help: "Total number of AI system errors",
		},
		[]string{"component", "error_type"},
	)
)

// AIMetrics предоставляет методы для обновления метрик AI Analytics
type AIMetrics struct{}

// NewAIMetrics создает новый экземпляр AIMetrics
func NewAIMetrics() *AIMetrics {
	return &AIMetrics{}
}

// IncrementBehaviorAnalysis увеличивает счетчик анализа поведения
func (am *AIMetrics) IncrementBehaviorAnalysis(userID, actionType, status string) {
	AIBBehaviorAnalysisTotal.WithLabelValues(userID, actionType, status).Inc()
}

// ObserveAnomalyScore наблюдает балл аномалии
func (am *AIMetrics) ObserveAnomalyScore(userID, actionType string, score float64) {
	AIBAnomalyScore.WithLabelValues(userID, actionType).Observe(score)
}

// IncrementAnomalyDetections увеличивает счетчик обнаружения аномалий
func (am *AIMetrics) IncrementAnomalyDetections(userID, anomalyType, severity string) {
	AIBAnomalyDetections.WithLabelValues(userID, anomalyType, severity).Inc()
}

// UpdateUserProfiles обновляет количество профилей пользователей
func (am *AIMetrics) UpdateUserProfiles(status string, count int) {
	AIBUserProfiles.WithLabelValues(status).Set(float64(count))
}

// UpdateModelAccuracy обновляет точность модели
func (am *AIMetrics) UpdateModelAccuracy(modelVersion string, accuracy float64) {
	AIBModelAccuracy.WithLabelValues(modelVersion).Set(accuracy)
}

// ObserveModelTrainingDuration наблюдает время обучения модели
func (am *AIMetrics) ObserveModelTrainingDuration(modelVersion string, duration float64) {
	AIBModelTrainingDuration.WithLabelValues(modelVersion).Observe(duration)
}

// ObserveModelInferenceDuration наблюдает время инференса модели
func (am *AIMetrics) ObserveModelInferenceDuration(modelVersion string, duration float64) {
	AIBModelInferenceDuration.WithLabelValues(modelVersion).Observe(duration)
}

// IncrementPredictiveThreats увеличивает счетчик предсказанных угроз
func (am *AIMetrics) IncrementPredictiveThreats(threatType, confidenceLevel, severity string) {
	AIPredictiveThreats.WithLabelValues(threatType, confidenceLevel, severity).Inc()
}

// UpdatePredictiveAccuracy обновляет точность предсказаний
func (am *AIMetrics) UpdatePredictiveAccuracy(threatType, modelVersion string, accuracy float64) {
	AIPredictiveAccuracy.WithLabelValues(threatType, modelVersion).Set(accuracy)
}

// IncrementFalsePositives увеличивает счетчик ложных срабатываний
func (am *AIMetrics) IncrementFalsePositives(threatType, modelVersion string) {
	AIPredictiveFalsePositives.WithLabelValues(threatType, modelVersion).Inc()
}

// IncrementFalseNegatives увеличивает счетчик пропущенных угроз
func (am *AIMetrics) IncrementFalseNegatives(threatType, modelVersion string) {
	AIPredictiveFalseNegatives.WithLabelValues(threatType, modelVersion).Inc()
}

// ObservePredictiveModelTrainingDuration наблюдает время обучения модели предсказания
func (am *AIMetrics) ObservePredictiveModelTrainingDuration(threatType, modelVersion string, duration float64) {
	AIPredictiveModelTrainingDuration.WithLabelValues(threatType, modelVersion).Observe(duration)
}

// ObservePredictiveModelInferenceDuration наблюдает время инференса модели предсказания
func (am *AIMetrics) ObservePredictiveModelInferenceDuration(threatType, modelVersion string, duration float64) {
	AIPredictiveModelInferenceDuration.WithLabelValues(threatType, modelVersion).Observe(duration)
}

// IncrementAutomatedResponses увеличивает счетчик автоматических ответов
func (am *AIMetrics) IncrementAutomatedResponses(responseType, threatType, status string) {
	AIAutomatedResponses.WithLabelValues(responseType, threatType, status).Inc()
}

// ObserveAutomatedResponseDuration наблюдает время выполнения автоматического ответа
func (am *AIMetrics) ObserveAutomatedResponseDuration(responseType string, duration float64) {
	AIAutomatedResponseDuration.WithLabelValues(responseType).Observe(duration)
}

// UpdateAutomatedResponseSuccess обновляет успешность автоматических ответов
func (am *AIMetrics) UpdateAutomatedResponseSuccess(responseType string, successRate float64) {
	AIAutomatedResponseSuccess.WithLabelValues(responseType).Set(successRate)
}

// UpdateAutomatedResponseRules обновляет количество правил автоматических ответов
func (am *AIMetrics) UpdateAutomatedResponseRules(status string, count int) {
	AIAutomatedResponseRules.WithLabelValues(status).Set(float64(count))
}

// IncrementAutomatedResponseErrors увеличивает счетчик ошибок автоматических ответов
func (am *AIMetrics) IncrementAutomatedResponseErrors(responseType, errorType string) {
	AIAutomatedResponseErrors.WithLabelValues(responseType, errorType).Inc()
}

// IncrementMonitoringAlerts увеличивает счетчик мониторинговых алертов
func (am *AIMetrics) IncrementMonitoringAlerts(alertType, severity, status string) {
	AIMonitoringAlerts.WithLabelValues(alertType, severity, status).Inc()
}

// UpdateMonitoringMetrics обновляет метрики мониторинга
func (am *AIMetrics) UpdateMonitoringMetrics(metricName, component string, value float64) {
	AIMonitoringMetrics.WithLabelValues(metricName, component).Set(value)
}

// UpdateMonitoringDashboardStatus обновляет статус дашборда мониторинга
func (am *AIMetrics) UpdateMonitoringDashboardStatus(dashboardURL string, available bool) {
	status := 0.0
	if available {
		status = 1.0
	}
	AIMonitoringDashboardStatus.WithLabelValues(dashboardURL).Set(status)
}

// IncrementMonitoringDataPoints увеличивает счетчик точек данных мониторинга
func (am *AIMetrics) IncrementMonitoringDataPoints(dataType, source string) {
	AIMonitoringDataPoints.WithLabelValues(dataType, source).Inc()
}

// ObserveMonitoringDataProcessingDuration наблюдает время обработки данных мониторинга
func (am *AIMetrics) ObserveMonitoringDataProcessingDuration(dataType string, duration float64) {
	AIMonitoringDataProcessingDuration.WithLabelValues(dataType).Observe(duration)
}

// IncrementIntegrationWireGuardAnalysis увеличивает счетчик анализа WireGuard
func (am *AIMetrics) IncrementIntegrationWireGuardAnalysis(analysisType, status string) {
	AIIntegrationWireGuardAnalysis.WithLabelValues(analysisType, status).Inc()
}

// IncrementIntegrationQUICAnalysis увеличивает счетчик анализа QUIC
func (am *AIMetrics) IncrementIntegrationQUICAnalysis(analysisType, status string) {
	AIIntegrationQUICAnalysis.WithLabelValues(analysisType, status).Inc()
}

// IncrementIntegrationQuantumAnalysis увеличивает счетчик анализа квантовых угроз
func (am *AIMetrics) IncrementIntegrationQuantumAnalysis(analysisType, status string) {
	AIIntegrationQuantumAnalysis.WithLabelValues(analysisType, status).Inc()
}

// IncrementIntegrationWireGuardThreats увеличивает счетчик ответов на угрозы WireGuard
func (am *AIMetrics) IncrementIntegrationWireGuardThreats(threatType, responseType string) {
	AIIntegrationWireGuardThreats.WithLabelValues(threatType, responseType).Inc()
}

// IncrementIntegrationQUICThreats увеличивает счетчик ответов на угрозы QUIC
func (am *AIMetrics) IncrementIntegrationQUICThreats(threatType, responseType string) {
	AIIntegrationQUICThreats.WithLabelValues(threatType, responseType).Inc()
}

// IncrementIntegrationQuantumThreats увеличивает счетчик ответов на квантовые угрозы
func (am *AIMetrics) IncrementIntegrationQuantumThreats(threatType, responseType string) {
	AIIntegrationQuantumThreats.WithLabelValues(threatType, responseType).Inc()
}

// UpdateSystemStatus обновляет статус компонента системы
func (am *AIMetrics) UpdateSystemStatus(component string, running bool) {
	status := 0.0
	if running {
		status = 1.0
	}
	AISystemStatus.WithLabelValues(component).Set(status)
}

// UpdateSystemMemoryUsage обновляет использование памяти
func (am *AIMetrics) UpdateSystemMemoryUsage(component string, bytes int64) {
	AISystemMemoryUsage.WithLabelValues(component).Set(float64(bytes))
}

// UpdateSystemCPUUsage обновляет использование CPU
func (am *AIMetrics) UpdateSystemCPUUsage(component string, percentage float64) {
	AISystemCPUUsage.WithLabelValues(component).Set(percentage)
}

// ObserveSystemModelLoadTime наблюдает время загрузки модели
func (am *AIMetrics) ObserveSystemModelLoadTime(modelType, modelVersion string, duration float64) {
	AISystemModelLoadTime.WithLabelValues(modelType, modelVersion).Observe(duration)
}

// IncrementSystemErrors увеличивает счетчик ошибок системы
func (am *AIMetrics) IncrementSystemErrors(component, errorType string) {
	AISystemErrors.WithLabelValues(component, errorType).Inc()
}
