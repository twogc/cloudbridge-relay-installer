package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
}

// NewMetrics создает новый экземпляр метрик
func NewMetrics() *Metrics {
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
	// Обновляем метрики AI сервиса
	m.aiRequestsProcessed.Add(1000) // Симуляция обработанных запросов
	m.aiAnomaliesDetected.Add(50)   // Симуляция обнаруженных аномалий
	m.aiPredictionsMade.Add(200)    // Симуляция сделанных предсказаний
	m.aiRoutesOptimized.Add(150)    // Симуляция оптимизированных маршрутов
	
	// Устанавливаем текущие значения
	m.aiActiveAlerts.Set(2)           // 2 активных алерта
	m.aiPredictionAccuracy.Set(0.85)  // 85% точность предсказаний
	m.aiServiceUptime.Set(time.Since(time.Now().Add(-24 * time.Hour)).Seconds()) // 24 часа uptime
	
	// Системные метрики
	m.cpuUsage.Set(51.3)              // 51.3% CPU
	m.memoryUsage.Set(256 * 1024 * 1024) // 256 MB памяти
	m.diskUsage.Set(1024 * 1024 * 1024)  // 1 GB диска
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


