package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestNewAIServiceMetrics(t *testing.T) {
	// Создаем новый регистр для тестов
	reg := prometheus.NewRegistry()

	// Создаем AI Service метрики
	aiServiceMetrics := NewAIServiceMetrics(reg)

	// Проверяем, что метрики созданы
	assert.NotNil(t, aiServiceMetrics)
	assert.NotNil(t, aiServiceMetrics.anomalyDetectionsTotal)
	assert.NotNil(t, aiServiceMetrics.anomalyDetectionsActive)
	assert.NotNil(t, aiServiceMetrics.loadPredictionsTotal)
	assert.NotNil(t, aiServiceMetrics.routeOptimizationsTotal)
	assert.NotNil(t, aiServiceMetrics.dataProcessingTotal)
	assert.NotNil(t, aiServiceMetrics.modelInferencesTotal)
	assert.NotNil(t, aiServiceMetrics.analyticsReportsGenerated)
	assert.NotNil(t, aiServiceMetrics.processingTimeTotal)
	assert.NotNil(t, aiServiceMetrics.memoryUsageBytes)
	assert.NotNil(t, aiServiceMetrics.cacheHits)
	assert.NotNil(t, aiServiceMetrics.errorsTotal)
	assert.NotNil(t, aiServiceMetrics.healthChecksTotal)
	assert.NotNil(t, aiServiceMetrics.configUpdates)
}

func TestAIServiceAnomalyDetectionMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	aiServiceMetrics := NewAIServiceMetrics(reg)

	// Тестируем детекцию аномалий
	aiServiceMetrics.IncAnomalyDetections("isolation_forest", "network_anomaly", "high")
	aiServiceMetrics.IncAnomalyDetections("isolation_forest", "traffic_anomaly", "medium")
	aiServiceMetrics.IncAnomalyDetections("lstm", "time_series_anomaly", "low")

	// Тестируем активные детекции
	aiServiceMetrics.SetAnomalyDetectionsActive(5)
	aiServiceMetrics.SetAnomalyDetectionsActive(3)

	// Тестируем ложные срабатывания
	aiServiceMetrics.IncAnomalyFalsePositives("isolation_forest", "network_anomaly")
	aiServiceMetrics.IncAnomalyFalsePositives("lstm", "time_series_anomaly")

	// Тестируем пропущенные аномалии
	aiServiceMetrics.IncAnomalyFalseNegatives("isolation_forest", "network_anomaly")
	aiServiceMetrics.IncAnomalyFalseNegatives("lstm", "time_series_anomaly")

	// Тестируем точность
	aiServiceMetrics.SetAnomalyAccuracy("isolation_forest", "network_anomaly", 0.95)
	aiServiceMetrics.SetAnomalyAccuracy("lstm", "time_series_anomaly", 0.87)

	// Тестируем длительность детекции
	aiServiceMetrics.ObserveAnomalyDetectionDuration("isolation_forest", "network_anomaly", 0.1)
	aiServiceMetrics.ObserveAnomalyDetectionDuration("lstm", "time_series_anomaly", 0.05)

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestAIServiceLoadPredictionMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	aiServiceMetrics := NewAIServiceMetrics(reg)

	// Тестируем предсказания нагрузки
	aiServiceMetrics.IncLoadPredictions("arima", "cpu_usage")
	aiServiceMetrics.IncLoadPredictions("arima", "memory_usage")
	aiServiceMetrics.IncLoadPredictions("lstm", "network_throughput")

	// Тестируем ошибки предсказаний
	aiServiceMetrics.IncLoadPredictionErrors("arima", "model_error")
	aiServiceMetrics.IncLoadPredictionErrors("lstm", "data_error")

	// Тестируем точность предсказаний
	aiServiceMetrics.SetLoadPredictionAccuracy("arima", "cpu_usage", 0.92)
	aiServiceMetrics.SetLoadPredictionAccuracy("lstm", "network_throughput", 0.88)

	// Тестируем длительность предсказаний
	aiServiceMetrics.ObserveLoadPredictionDuration("arima", "cpu_usage", 0.2)
	aiServiceMetrics.ObserveLoadPredictionDuration("lstm", "network_throughput", 0.15)

	// Тестируем уверенность предсказаний
	aiServiceMetrics.SetLoadPredictionConfidence("arima", "cpu_usage", 0.85)
	aiServiceMetrics.SetLoadPredictionConfidence("lstm", "network_throughput", 0.92)

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestAIServiceRouteOptimizationMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	aiServiceMetrics := NewAIServiceMetrics(reg)

	// Тестируем оптимизации маршрутов
	aiServiceMetrics.IncRouteOptimizations("genetic_algorithm", "latency_optimization")
	aiServiceMetrics.IncRouteOptimizations("genetic_algorithm", "bandwidth_optimization")
	aiServiceMetrics.IncRouteOptimizations("reinforcement_learning", "cost_optimization")

	// Тестируем выигрыши от оптимизации
	aiServiceMetrics.ObserveRouteOptimizationGains("genetic_algorithm", "latency_optimization", 0.15)
	aiServiceMetrics.ObserveRouteOptimizationGains("genetic_algorithm", "bandwidth_optimization", 0.25)
	aiServiceMetrics.ObserveRouteOptimizationGains("reinforcement_learning", "cost_optimization", 0.30)

	// Тестируем длительность оптимизации
	aiServiceMetrics.ObserveRouteOptimizationDuration("genetic_algorithm", "latency_optimization", 1.5)
	aiServiceMetrics.ObserveRouteOptimizationDuration("reinforcement_learning", "cost_optimization", 2.0)

	// Тестируем успешные оптимизации
	aiServiceMetrics.IncRouteOptimizationSuccess("genetic_algorithm", "latency_optimization")
	aiServiceMetrics.IncRouteOptimizationSuccess("reinforcement_learning", "cost_optimization")

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestAIServiceDataPipelineMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	aiServiceMetrics := NewAIServiceMetrics(reg)

	// Тестируем обработку данных
	aiServiceMetrics.IncDataProcessing("real_time_pipeline", "collection", "network_metrics")
	aiServiceMetrics.IncDataProcessing("real_time_pipeline", "processing", "network_metrics")
	aiServiceMetrics.IncDataProcessing("batch_pipeline", "aggregation", "historical_data")

	// Тестируем длительность обработки
	aiServiceMetrics.ObserveDataProcessingDuration("real_time_pipeline", "collection", "network_metrics", 0.05)
	aiServiceMetrics.ObserveDataProcessingDuration("batch_pipeline", "aggregation", "historical_data", 2.0)

	// Тестируем ошибки обработки
	aiServiceMetrics.IncDataProcessingErrors("real_time_pipeline", "collection", "connection_error")
	aiServiceMetrics.IncDataProcessingErrors("batch_pipeline", "aggregation", "memory_error")

	// Тестируем пропускную способность
	aiServiceMetrics.IncDataProcessingThroughput("real_time_pipeline", "collection", "network_metrics", "incoming", 1000)
	aiServiceMetrics.IncDataProcessingThroughput("batch_pipeline", "aggregation", "historical_data", "outgoing", 500)

	// Тестируем латентность
	aiServiceMetrics.ObserveDataProcessingLatency("real_time_pipeline", "collection", "network_metrics", 0.01)
	aiServiceMetrics.ObserveDataProcessingLatency("batch_pipeline", "aggregation", "historical_data", 0.5)

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestAIServiceMLModelMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	aiServiceMetrics := NewAIServiceMetrics(reg)

	// Тестируем инференсы моделей
	aiServiceMetrics.IncModelInferences("anomaly_detector", "isolation_forest", "v1.0")
	aiServiceMetrics.IncModelInferences("load_predictor", "lstm", "v2.1")
	aiServiceMetrics.IncModelInferences("route_optimizer", "genetic_algorithm", "v1.5")

	// Тестируем длительность инференса
	aiServiceMetrics.ObserveModelInferenceDuration("anomaly_detector", "isolation_forest", "v1.0", 0.1)
	aiServiceMetrics.ObserveModelInferenceDuration("load_predictor", "lstm", "v2.1", 0.2)

	// Тестируем точность моделей
	aiServiceMetrics.SetModelAccuracy("anomaly_detector", "isolation_forest", "v1.0", 0.95)
	aiServiceMetrics.SetModelAccuracy("load_predictor", "lstm", "v2.1", 0.88)

	// Тестируем уверенность моделей
	aiServiceMetrics.SetModelConfidence("anomaly_detector", "isolation_forest", "v1.0", 0.92)
	aiServiceMetrics.SetModelConfidence("load_predictor", "lstm", "v2.1", 0.85)

	// Тестируем длительность обучения
	aiServiceMetrics.ObserveModelTrainingDuration("anomaly_detector", "isolation_forest", "v1.0", 3600)
	aiServiceMetrics.ObserveModelTrainingDuration("load_predictor", "lstm", "v2.1", 7200)

	// Тестируем ошибки обучения
	aiServiceMetrics.IncModelTrainingErrors("anomaly_detector", "isolation_forest", "v1.0", "convergence_error")
	aiServiceMetrics.IncModelTrainingErrors("load_predictor", "lstm", "v2.1", "memory_error")

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestAIServiceAnalyticsMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	aiServiceMetrics := NewAIServiceMetrics(reg)

	// Тестируем генерацию отчетов
	aiServiceMetrics.IncAnalyticsReportsGenerated("performance_report", "json")
	aiServiceMetrics.IncAnalyticsReportsGenerated("performance_report", "csv")
	aiServiceMetrics.IncAnalyticsReportsGenerated("anomaly_report", "pdf")

	// Тестируем длительность генерации отчетов
	aiServiceMetrics.ObserveAnalyticsReportDuration("performance_report", "json", 5.0)
	aiServiceMetrics.ObserveAnalyticsReportDuration("anomaly_report", "pdf", 10.0)

	// Тестируем инсайты
	aiServiceMetrics.IncAnalyticsInsights("performance_insight", "optimization")
	aiServiceMetrics.IncAnalyticsInsights("anomaly_insight", "detection")

	// Тестируем точность инсайтов
	aiServiceMetrics.SetAnalyticsInsightAccuracy("performance_insight", "optimization", 0.90)
	aiServiceMetrics.SetAnalyticsInsightAccuracy("anomaly_insight", "detection", 0.85)

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestAIServicePerformanceMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	aiServiceMetrics := NewAIServiceMetrics(reg)

	// Тестируем время обработки
	aiServiceMetrics.IncProcessingTime("anomaly_detector", "detection", 0.1)
	aiServiceMetrics.IncProcessingTime("load_predictor", "prediction", 0.2)
	aiServiceMetrics.IncProcessingTime("route_optimizer", "optimization", 1.5)

	// Тестируем длительность обработки
	aiServiceMetrics.ObserveProcessingTimeDuration("anomaly_detector", "detection", 0.1)
	aiServiceMetrics.ObserveProcessingTimeDuration("load_predictor", "prediction", 0.2)

	// Тестируем размер очереди
	aiServiceMetrics.SetProcessingQueueSize("high_priority", "urgent", 10)
	aiServiceMetrics.SetProcessingQueueSize("normal_priority", "standard", 50)

	// Тестируем латентность очереди
	aiServiceMetrics.ObserveProcessingQueueLatency("high_priority", "urgent", 0.01)
	aiServiceMetrics.ObserveProcessingQueueLatency("normal_priority", "standard", 0.05)

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestAIServiceResourceMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	aiServiceMetrics := NewAIServiceMetrics(reg)

	// Тестируем использование памяти
	aiServiceMetrics.SetMemoryUsage("anomaly_detector", "heap", 1024*1024*1024)
	aiServiceMetrics.SetMemoryUsage("load_predictor", "heap", 512*1024*1024)
	aiServiceMetrics.SetMemoryUsage("route_optimizer", "stack", 256*1024*1024)

	// Тестируем использование CPU
	aiServiceMetrics.SetCPUUsage("anomaly_detector", "user", 25.5)
	aiServiceMetrics.SetCPUUsage("load_predictor", "system", 15.2)
	aiServiceMetrics.SetCPUUsage("route_optimizer", "user", 30.8)

	// Тестируем использование GPU
	aiServiceMetrics.SetGPUUsage("model_training", "inference", 45.0)
	aiServiceMetrics.SetGPUUsage("model_training", "training", 80.0)

	// Тестируем использование диска
	aiServiceMetrics.SetDiskUsage("model_storage", "models", 10*1024*1024*1024)
	aiServiceMetrics.SetDiskUsage("data_storage", "datasets", 50*1024*1024*1024)

	// Тестируем использование сети
	aiServiceMetrics.IncNetworkUsage("data_collection", "incoming", 1024*1024)
	aiServiceMetrics.IncNetworkUsage("model_updates", "outgoing", 512*1024)

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestAIServiceCacheMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	aiServiceMetrics := NewAIServiceMetrics(reg)

	// Тестируем попадания в кэш
	aiServiceMetrics.IncCacheHits("model_cache", "anomaly_detector")
	aiServiceMetrics.IncCacheHits("model_cache", "load_predictor")
	aiServiceMetrics.IncCacheHits("data_cache", "network_metrics")

	// Тестируем промахи кэша
	aiServiceMetrics.IncCacheMisses("model_cache", "anomaly_detector")
	aiServiceMetrics.IncCacheMisses("data_cache", "network_metrics")

	// Тестируем размер кэша
	aiServiceMetrics.SetCacheSize("model_cache", "anomaly_detector", 100)
	aiServiceMetrics.SetCacheSize("data_cache", "network_metrics", 1000)

	// Тестируем вытеснения из кэша
	aiServiceMetrics.IncCacheEvictions("model_cache", "anomaly_detector", "lru")
	aiServiceMetrics.IncCacheEvictions("data_cache", "network_metrics", "size_limit")

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestAIServiceErrorMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	aiServiceMetrics := NewAIServiceMetrics(reg)

	// Тестируем ошибки
	aiServiceMetrics.IncErrors("anomaly_detector", "model_error", "high")
	aiServiceMetrics.IncErrors("load_predictor", "data_error", "medium")
	aiServiceMetrics.IncErrors("route_optimizer", "convergence_error", "low")

	// Тестируем попытки восстановления
	aiServiceMetrics.IncErrorRecoveryAttempts("anomaly_detector", "model_error")
	aiServiceMetrics.IncErrorRecoveryAttempts("load_predictor", "data_error")

	// Тестируем успешные восстановления
	aiServiceMetrics.IncErrorRecoverySuccess("anomaly_detector", "model_error")
	aiServiceMetrics.IncErrorRecoverySuccess("load_predictor", "data_error")

	// Тестируем события circuit breaker
	aiServiceMetrics.IncCircuitBreakerEvents("anomaly_detector", "opened")
	aiServiceMetrics.IncCircuitBreakerEvents("load_predictor", "closed")

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestAIServiceMonitoringMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	aiServiceMetrics := NewAIServiceMetrics(reg)

	// Тестируем проверки здоровья
	aiServiceMetrics.IncHealthChecks("anomaly_detector", "healthy")
	aiServiceMetrics.IncHealthChecks("load_predictor", "healthy")
	aiServiceMetrics.IncHealthChecks("route_optimizer", "unhealthy")

	// Тестируем длительность проверок здоровья
	aiServiceMetrics.ObserveHealthCheckDuration("anomaly_detector", 0.1)
	aiServiceMetrics.ObserveHealthCheckDuration("load_predictor", 0.05)

	// Тестируем статус здоровья
	aiServiceMetrics.SetHealthStatus("anomaly_detector", true)
	aiServiceMetrics.SetHealthStatus("load_predictor", true)
	aiServiceMetrics.SetHealthStatus("route_optimizer", false)

	// Тестируем время работы
	aiServiceMetrics.SetUptimeSeconds(3600.0)
	aiServiceMetrics.SetUptimeSeconds(7200.0)

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestAIServiceConfigurationMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	aiServiceMetrics := NewAIServiceMetrics(reg)

	// Тестируем обновления конфигурации
	aiServiceMetrics.IncConfigUpdates("anomaly_detector", "model_config")
	aiServiceMetrics.IncConfigUpdates("load_predictor", "threshold_config")
	aiServiceMetrics.IncConfigUpdates("route_optimizer", "algorithm_config")

	// Тестируем ошибки конфигурации
	aiServiceMetrics.IncConfigErrors("anomaly_detector", "invalid_config")
	aiServiceMetrics.IncConfigErrors("load_predictor", "parse_error")

	// Тестируем флаги функций
	aiServiceMetrics.SetFeatureFlag("anomaly_detection", "anomaly_detector", true)
	aiServiceMetrics.SetFeatureFlag("load_prediction", "load_predictor", true)
	aiServiceMetrics.SetFeatureFlag("route_optimization", "route_optimizer", false)

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestAIServiceMetricsIntegration(t *testing.T) {
	// Создаем основной Metrics с AI Service метриками
	reg := prometheus.NewRegistry()
	mainMetrics := NewMetrics(reg)

	// Проверяем, что AI Service метрики доступны
	aiServiceMetrics := mainMetrics.GetAIServiceMetrics()
	assert.NotNil(t, aiServiceMetrics)

	// Тестируем методы основного Metrics
	mainMetrics.IncAIAnomalyDetections("isolation_forest", "network_anomaly", "high")
	mainMetrics.SetAIAnomalyDetectionsActive(5)
	mainMetrics.IncAIAnomalyFalsePositives("isolation_forest", "network_anomaly")
	mainMetrics.IncAIAnomalyFalseNegatives("isolation_forest", "network_anomaly")
	mainMetrics.SetAIAnomalyAccuracy("isolation_forest", "network_anomaly", 0.95)
	mainMetrics.ObserveAIAnomalyDetectionDuration("isolation_forest", "network_anomaly", 0.1)
	mainMetrics.IncAILoadPredictions("arima", "cpu_usage")
	mainMetrics.IncAILoadPredictionErrors("arima", "model_error")
	mainMetrics.SetAILoadPredictionAccuracy("arima", "cpu_usage", 0.92)
	mainMetrics.ObserveAILoadPredictionDuration("arima", "cpu_usage", 0.2)
	mainMetrics.SetAILoadPredictionConfidence("arima", "cpu_usage", 0.85)
	mainMetrics.IncAIRouteOptimizations("genetic_algorithm", "latency_optimization")
	mainMetrics.ObserveAIRouteOptimizationGains("genetic_algorithm", "latency_optimization", 0.15)
	mainMetrics.ObserveAIRouteOptimizationDuration("genetic_algorithm", "latency_optimization", 1.5)
	mainMetrics.IncAIRouteOptimizationSuccess("genetic_algorithm", "latency_optimization")
	mainMetrics.IncAIDataProcessing("real_time_pipeline", "collection", "network_metrics")
	mainMetrics.ObserveAIDataProcessingDuration("real_time_pipeline", "collection", "network_metrics", 0.05)
	mainMetrics.IncAIDataProcessingErrors("real_time_pipeline", "collection", "connection_error")
	mainMetrics.IncAIDataProcessingThroughput("real_time_pipeline", "collection", "network_metrics", "incoming", 1000)
	mainMetrics.ObserveAIDataProcessingLatency("real_time_pipeline", "collection", "network_metrics", 0.01)
	mainMetrics.IncAIModelInferences("anomaly_detector", "isolation_forest", "v1.0")
	mainMetrics.ObserveAIModelInferenceDuration("anomaly_detector", "isolation_forest", "v1.0", 0.1)
	mainMetrics.SetAIModelAccuracy("anomaly_detector", "isolation_forest", "v1.0", 0.95)
	mainMetrics.SetAIModelConfidence("anomaly_detector", "isolation_forest", "v1.0", 0.92)
	mainMetrics.ObserveAIModelTrainingDuration("anomaly_detector", "isolation_forest", "v1.0", 3600)
	mainMetrics.IncAIModelTrainingErrors("anomaly_detector", "isolation_forest", "v1.0", "convergence_error")
	mainMetrics.IncAIAnalyticsReportsGenerated("performance_report", "json")
	mainMetrics.ObserveAIAnalyticsReportDuration("performance_report", "json", 5.0)
	mainMetrics.IncAIAnalyticsInsights("performance_insight", "optimization")
	mainMetrics.SetAIAnalyticsInsightAccuracy("performance_insight", "optimization", 0.90)
	mainMetrics.IncAIProcessingTime("anomaly_detector", "detection", 0.1)
	mainMetrics.ObserveAIProcessingTimeDuration("anomaly_detector", "detection", 0.1)
	mainMetrics.SetAIProcessingQueueSize("high_priority", "urgent", 10)
	mainMetrics.ObserveAIProcessingQueueLatency("high_priority", "urgent", 0.01)
	mainMetrics.SetAIMemoryUsage("anomaly_detector", "heap", 1024*1024*1024)
	mainMetrics.SetAICPUUsage("anomaly_detector", "user", 25.5)
	mainMetrics.SetAIGPUUsage("model_training", "inference", 45.0)
	mainMetrics.SetAIDiskUsage("model_storage", "models", 10*1024*1024*1024)
	mainMetrics.IncAINetworkUsage("data_collection", "incoming", 1024*1024)
	mainMetrics.IncAICacheHits("model_cache", "anomaly_detector")
	mainMetrics.IncAICacheMisses("model_cache", "anomaly_detector")
	mainMetrics.SetAICacheSize("model_cache", "anomaly_detector", 100)
	mainMetrics.IncAICacheEvictions("model_cache", "anomaly_detector", "lru")
	mainMetrics.IncAIErrors("anomaly_detector", "model_error", "high")
	mainMetrics.IncAIErrorRecoveryAttempts("anomaly_detector", "model_error")
	mainMetrics.IncAIErrorRecoverySuccess("anomaly_detector", "model_error")
	mainMetrics.IncAICircuitBreakerEvents("anomaly_detector", "opened")
	mainMetrics.IncAIHealthChecks("anomaly_detector", "healthy")
	mainMetrics.ObserveAIHealthCheckDuration("anomaly_detector", 0.1)
	mainMetrics.SetAIHealthStatus("anomaly_detector", true)
	mainMetrics.SetAIUptimeSeconds(3600.0)
	mainMetrics.IncAIConfigUpdates("anomaly_detector", "model_config")
	mainMetrics.IncAIConfigErrors("anomaly_detector", "invalid_config")
	mainMetrics.SetAIFeatureFlag("anomaly_detection", "anomaly_detector", true)

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}



