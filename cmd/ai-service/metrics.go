package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai"
)

// Metrics структура для хранения Prometheus метрик
type Metrics struct {
	// HTTP метрики
	httpRequestsTotal     *prometheus.CounterVec
	httpRequestDuration   *prometheus.HistogramVec
	httpRequestsInFlight  prometheus.Gauge

	// AI Service метрики
	aiRequestsProcessed   prometheus.Counter
	aiAnomaliesDetected   prometheus.Counter
	aiPredictionsMade     prometheus.Counter
	aiRoutesOptimized     prometheus.Counter
	aiActiveAlerts        prometheus.Gauge
	aiPredictionAccuracy  prometheus.Gauge
	aiServiceUptime       prometheus.Gauge

	// Системные метрики
	cpuUsage              prometheus.Gauge
	memoryUsage           prometheus.Gauge
	diskUsage             prometheus.Gauge

	// Ссылка на AI Service для получения реальных данных
	aiService             *ai.Service
}

// NewMetrics создает новый экземпляр метрик
func NewMetrics(aiService *ai.Service) *Metrics {
	return &Metrics{
		// HTTP метрики
		httpRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status"},
		),
		httpRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		httpRequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_requests_in_flight",
				Help: "Current number of HTTP requests being processed",
			},
		),

		// AI Service метрики
		aiRequestsProcessed: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "ai_requests_processed_total",
				Help: "Total number of AI requests processed",
			},
		),
		aiAnomaliesDetected: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "ai_anomalies_detected_total",
				Help: "Total number of anomalies detected",
			},
		),
		aiPredictionsMade: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "ai_predictions_made_total",
				Help: "Total number of predictions made",
			},
		),
		aiRoutesOptimized: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "ai_routes_optimized_total",
				Help: "Total number of routes optimized",
			},
		),
		aiActiveAlerts: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "ai_active_alerts",
				Help: "Number of active alerts",
			},
		),
		aiPredictionAccuracy: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "ai_prediction_accuracy",
				Help: "Current prediction accuracy (0-1)",
			},
		),
		aiServiceUptime: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "ai_service_uptime_seconds",
				Help: "AI service uptime in seconds",
			},
		),

		// Системные метрики
		cpuUsage: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "ai_cpu_usage_percent",
				Help: "AI service CPU usage percentage",
			},
		),
		memoryUsage: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "ai_memory_usage_bytes",
				Help: "AI service memory usage in bytes",
			},
		),
		diskUsage: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "ai_disk_usage_bytes",
				Help: "AI service disk usage in bytes",
			},
		),

		// Инициализируем ссылку на AI Service
		aiService: aiService,
	}
}

// PrometheusMiddleware создает middleware для сбора HTTP метрик
func (m *Metrics) PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Увеличиваем счетчик активных запросов
		m.httpRequestsInFlight.Inc()
		defer m.httpRequestsInFlight.Dec()

		// Обрабатываем запрос
		c.Next()

		// Записываем метрики
		duration := time.Since(start).Seconds()
		status := http.StatusText(c.Writer.Status())
		
		m.httpRequestsTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			status,
		).Inc()
		
		m.httpRequestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
		).Observe(duration)
	}
}

// UpdateAIMetrics обновляет AI метрики
func (m *Metrics) UpdateAIMetrics() {
	if m.aiService == nil {
		// Fallback к базовым значениям если AI Service недоступен
		m.aiRequestsProcessed.Add(0)
		m.aiAnomaliesDetected.Add(0)
		m.aiPredictionsMade.Add(0)
		m.aiRoutesOptimized.Add(0)
		m.aiActiveAlerts.Set(0)
		m.aiPredictionAccuracy.Set(0.0)
		m.aiServiceUptime.Set(0)
		m.cpuUsage.Set(0)
		m.memoryUsage.Set(0)
		m.diskUsage.Set(0)
		return
	}

	// Получаем реальные данные из AI Service
	status := m.aiService.GetStatus()
	aiMetrics := status.Performance["metrics"].(map[string]interface{})

	// Обновляем метрики на основе реальных данных
	if anomalyMetrics, ok := aiMetrics["anomaly_detection"].(map[string]interface{}); ok {
		if totalDetections, ok := anomalyMetrics["total_detections"].(int64); ok {
			// Для Counter используем Add с разностью
			m.aiAnomaliesDetected.Add(float64(totalDetections))
		}
	}

	if loadPredictionMetrics, ok := aiMetrics["load_prediction"].(map[string]interface{}); ok {
		if totalPredictions, ok := loadPredictionMetrics["total_predictions"].(int64); ok {
			m.aiPredictionsMade.Add(float64(totalPredictions))
		}
		if accuracy, ok := loadPredictionMetrics["accuracy"].(float64); ok {
			m.aiPredictionAccuracy.Set(accuracy)
		}
	}

	if routeOptimizationMetrics, ok := aiMetrics["route_optimization"].(map[string]interface{}); ok {
		if totalOptimizations, ok := routeOptimizationMetrics["total_optimizations"].(int64); ok {
			m.aiRoutesOptimized.Add(float64(totalOptimizations))
		}
	}

	if processingMetrics, ok := aiMetrics["processing"].(map[string]interface{}); ok {
		if totalCount, ok := processingMetrics["total_count"].(int64); ok {
			m.aiRequestsProcessed.Add(float64(totalCount))
		}
	}

	if systemMetrics, ok := aiMetrics["system"].(map[string]interface{}); ok {
		if cpuUsage, ok := systemMetrics["cpu_usage"].(float64); ok {
			m.cpuUsage.Set(cpuUsage)
		}
		if memoryUsage, ok := systemMetrics["memory_usage"].(int64); ok {
			m.memoryUsage.Set(float64(memoryUsage))
		}
		if activeConnections, ok := systemMetrics["active_connections"].(int64); ok {
			// Используем активные соединения как индикатор алертов
			if activeConnections > 100 {
				m.aiActiveAlerts.Set(1) // Высокая нагрузка
			} else if activeConnections > 50 {
				m.aiActiveAlerts.Set(0.5) // Средняя нагрузка
			} else {
				m.aiActiveAlerts.Set(0) // Низкая нагрузка
			}
		}
	}

	// Устанавливаем uptime на основе реального времени работы сервиса
	if status.Uptime > 0 {
		m.aiServiceUptime.Set(status.Uptime.Seconds())
	}

	// Системные метрики диска (пока используем базовые значения)
	m.diskUsage.Set(1024 * 1024 * 1024) // 1 GB диска
}

// GetPrometheusHandler возвращает handler для Prometheus метрик
func (m *Metrics) GetPrometheusHandler() gin.HandlerFunc {
	return gin.WrapH(promhttp.Handler())
}

// StartMetricsUpdater запускает обновление метрик
func (m *Metrics) StartMetricsUpdater() {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			m.UpdateAIMetrics()
		}
	}()
}


