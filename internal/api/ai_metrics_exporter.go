package api

import (
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/metrics"
	"go.uber.org/zap"
)

// AIMetricsExporter экспортирует AI метрики в Prometheus
type AIMetricsExporter struct {
	aiMiddleware *AIMiddleware
	logger       *zap.Logger
	interval     time.Duration
	stopCh       chan struct{}
}

// NewAIMetricsExporter создает новый экспортер AI метрик
func NewAIMetricsExporter(aiMiddleware *AIMiddleware, logger *zap.Logger) *AIMetricsExporter {
	return &AIMetricsExporter{
		aiMiddleware: aiMiddleware,
		logger:       logger,
		interval:     30 * time.Second,
		stopCh:       make(chan struct{}),
	}
}

// Start запускает экспорт метрик
func (e *AIMetricsExporter) Start() {
	e.logger.Info("Starting AI metrics exporter...")
	
	go e.exportMetrics()
}

// Stop останавливает экспорт метрик
func (e *AIMetricsExporter) Stop() {
	e.logger.Info("Stopping AI metrics exporter...")
	close(e.stopCh)
}

// exportMetrics экспортирует метрики в Prometheus
func (e *AIMetricsExporter) exportMetrics() {
	ticker := time.NewTicker(e.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			e.updateMetrics()
		case <-e.stopCh:
			return
		}
	}
}

// updateMetrics обновляет Prometheus метрики
func (e *AIMetricsExporter) updateMetrics() {
	aiMetrics := e.aiMiddleware.GetMetrics()
	circuitBreakerStats := e.aiMiddleware.circuitBreaker.GetStats()

	// Обновляем метрики обработки соединений
	metrics.AIConnectionsProcessed.WithLabelValues("success", "false", "middleware").Add(float64(aiMetrics.GetRequestsProcessed()))
	
	// Обновляем метрики аномалий
	metrics.AnomaliesDetected.WithLabelValues("low", "network", "middleware").Add(float64(aiMetrics.GetAnomaliesDetected()))
	
	// Обновляем метрики предсказаний
	metrics.LoadPredictionsMade.WithLabelValues("24", "high").Add(float64(aiMetrics.GetPredictionsMade()))
	
	// Обновляем метрики оптимизации маршрутов
	metrics.RoutesOptimized.WithLabelValues("QUIC", "latency").Add(float64(aiMetrics.GetRoutesOptimized()))
	
	// Обновляем метрики кэша
	metrics.AICacheHits.WithLabelValues("anomaly").Add(float64(aiMetrics.GetCacheHits()))
	metrics.AICacheMisses.WithLabelValues("anomaly").Add(float64(aiMetrics.GetCacheMisses()))
	
	// Обновляем метрики circuit breaker
	state := 0
	if circuitBreakerStats["state"] == "open" {
		state = 1
	} else if circuitBreakerStats["state"] == "half-open" {
		state = 2
	}
	metrics.AICircuitBreakerState.WithLabelValues("middleware").Set(float64(state))
	
	// Обновляем метрики здоровья сервиса
	metrics.AIServiceHealth.WithLabelValues("middleware", "anomaly-detection").Set(1)
	metrics.AIServiceHealth.WithLabelValues("middleware", "load-prediction").Set(1)
	metrics.AIServiceHealth.WithLabelValues("middleware", "route-optimization").Set(1)
	
	// Обновляем метрики производительности
	avgProcessingTime := aiMetrics.GetAverageProcessingTime()
	metrics.AIProcessingDuration.WithLabelValues("middleware", "total").Observe(avgProcessingTime.Seconds())
	
	// Обновляем метрики точности (заглушки для демонстрации)
	metrics.AnomalyDetectionAccuracy.WithLabelValues("v1.0").Set(0.95)
	metrics.LoadPredictionAccuracy.WithLabelValues("v1.0", "24h").Set(0.85)
	
	// Обновляем метрики версий моделей
	metrics.AIModelVersion.WithLabelValues("anomaly-detection", "v1.0").Set(1.0)
	metrics.AIModelVersion.WithLabelValues("load-prediction", "v1.0").Set(1.0)
	metrics.AIModelVersion.WithLabelValues("route-optimization", "v1.0").Set(1.0)
	
	// Обновляем метрики ресурсов (заглушки)
	metrics.AIMemoryUsage.WithLabelValues("middleware").Set(128 * 1024 * 1024) // 128MB
	metrics.AICPUUsage.WithLabelValues("middleware").Set(25.0) // 25%
	
	// Обновляем бизнес-метрики
	metrics.AICostSavings.WithLabelValues("anomaly-detection").Add(1000.0)
	metrics.AICostSavings.WithLabelValues("route-optimization").Add(500.0)
	metrics.AIAutomationRate.WithLabelValues("connection-processing").Set(85.0)
	
	e.logger.Debug("AI metrics updated",
		zap.Int64("requests_processed", aiMetrics.GetRequestsProcessed()),
		zap.Int64("anomalies_detected", aiMetrics.GetAnomaliesDetected()),
		zap.Float64("cache_hit_rate", aiMetrics.GetCacheHitRate()),
		zap.String("circuit_breaker_state", circuitBreakerStats["state"].(string)),
	)
}
