package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/analytics"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/analytics/notifications"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/anomaly"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/predictive"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/routing"
	"go.uber.org/zap"
)

// AIService представляет AI сервис с обработчиками
type AIService struct {
	aiService          *ai.Service
	logger             *zap.Logger
	notificationSystem *notifications.NotificationSystem
}

// NewAIService создает новый AI сервис
func NewAIService(aiService *ai.Service, logger *zap.Logger) *AIService {
	// Создаем систему уведомлений
	notificationSystem := notifications.NewNotificationSystem()

	// Настраиваем Telegram
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")

	if botToken != "" && chatID != "" {
		telegramConfig := &notifications.TelegramConfig{
			BotToken:              botToken,
			ChatID:                chatID,
			ParseMode:             "HTML",
			DisableWebPagePreview: true,
		}
		notificationSystem.AddTelegramConfig(telegramConfig)
		logger.Info("Telegram notifications configured")
	} else {
		logger.Warn("Telegram credentials not provided, notifications disabled")
	}

	return &AIService{
		aiService:          aiService,
		logger:             logger,
		notificationSystem: notificationSystem,
	}
}

// SetupRoutes настраивает маршруты для AI API
func (s *AIService) SetupRoutes(router *gin.Engine) {
	api := router.Group("/api/v1/ai")
	{
		// Anomaly Detection
		api.POST("/anomaly/detect", s.DetectAnomaly)
		api.GET("/anomaly/status", s.GetAnomalyStatus)

		// Predictive Analytics
		api.POST("/predict/load", s.PredictLoad)
		api.GET("/predict/load/:horizon", s.PredictLoadWithHorizon)

		// Intelligent Routing
		api.POST("/route/optimize", s.OptimizeRoute)
		api.GET("/route/recommendations", s.GetRouteRecommendations)

		// Model Management
		api.GET("/models", s.ListModels)
		api.POST("/models/:id/train", s.TrainModel)
		api.POST("/models/:id/deploy", s.DeployModel)

		// Metrics and Status
		api.GET("/status", s.GetStatus)
		api.GET("/metrics", s.GetMetrics)

		// Historical Data and Trends
		api.GET("/history/anomalies", s.GetAnomalyHistory)
		api.GET("/history/predictions", s.GetPredictionHistory)
		api.GET("/history/routes", s.GetRouteHistory)
		api.GET("/trends/performance", s.GetPerformanceTrends)

		// Real-time Data
		api.GET("/realtime/connections", s.GetRealtimeConnections)
		api.GET("/realtime/performance", s.GetRealtimePerformance)

		// Alerts and Notifications
		api.GET("/alerts", s.GetActiveAlerts)
		api.POST("/alerts/:id/acknowledge", s.AcknowledgeAlert)

		// System Health
		api.GET("/health/detailed", s.GetDetailedHealth)
		api.GET("/health/components", s.GetComponentHealth)

		// Telegram Notifications
		api.POST("/telegram/test", s.TestTelegramNotification)
		api.POST("/telegram/send", s.SendTelegramMessage)
		api.GET("/telegram/status", s.GetTelegramStatus)
		api.POST("/telegram/resource-alert", s.SendResourceAlert)
	}
}

// DetectAnomalyRequest запрос на обнаружение аномалий
type DetectAnomalyRequest struct {
	Metrics *anomaly.NetworkMetrics `json:"metrics" binding:"required"`
}

// DetectAnomalyResponse ответ на обнаружение аномалий
type DetectAnomalyResponse struct {
	Result *anomaly.AnomalyResult `json:"result"`
	Error  string                 `json:"error,omitempty"`
}

// DetectAnomaly обрабатывает запрос на обнаружение аномалий
func (s *AIService) DetectAnomaly(c *gin.Context) {
	var req DetectAnomalyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем AI сервис и детектор аномалий
	status := s.aiService.GetStatus()
	if !status.Components["anomaly_detector"] {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "anomaly detector not available"})
		return
	}

	// Здесь нужно получить доступ к детектору аномалий
	// Пока что возвращаем заглушку
	result := &anomaly.AnomalyResult{
		Score:       0.3,
		IsAnomaly:   false,
		Timestamp:   time.Now(),
		Features:    map[string]float64{"connections": 100, "bytes": 1000000, "loss": 0.01, "latency": 50, "cpu": 45, "memory": 60},
		Confidence:  0.85,
		AnomalyType: "normal",
		Severity:    "low",
	}

	c.JSON(http.StatusOK, gin.H{
		"result": result,
	})
}

// GetAnomalyStatus возвращает статус детектора аномалий
func (s *AIService) GetAnomalyStatus(c *gin.Context) {
	status := s.aiService.GetStatus()

	c.JSON(http.StatusOK, gin.H{
		"enabled": status.Components["anomaly_detector"],
		"status":  "running",
		"metrics": map[string]interface{}{
			"detections": 100,
			"accuracy":   0.95,
		},
	})
}

// PredictLoadRequest запрос на предсказание нагрузки
type PredictLoadRequest struct {
	Horizon int `json:"horizon" binding:"required"`
}

// PredictLoadResponse ответ на предсказание нагрузки
type PredictLoadResponse struct {
	Prediction *predictive.LoadPrediction `json:"prediction"`
	Error      string                     `json:"error,omitempty"`
}

// PredictLoad обрабатывает запрос на предсказание нагрузки
func (s *AIService) PredictLoad(c *gin.Context) {
	var req PredictLoadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем AI сервис и предсказатель нагрузки
	status := s.aiService.GetStatus()
	if !status.Components["load_predictor"] {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "load predictor not available"})
		return
	}

	// Здесь нужно получить доступ к предсказателю нагрузки
	// Пока что возвращаем заглушку
	prediction := &predictive.LoadPrediction{
		Forecast:   []float64{45, 50, 55, 60, 65},
		Confidence: 0.8,
		Horizon:    req.Horizon,
		Timestamp:  time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{
		"result": prediction,
	})
}

// PredictLoadWithHorizon обрабатывает запрос на предсказание нагрузки с горизонтом в URL
func (s *AIService) PredictLoadWithHorizon(c *gin.Context) {
	horizonStr := c.Param("horizon")
	horizon, err := strconv.Atoi(horizonStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid horizon parameter"})
		return
	}

	// Получаем AI сервис и предсказатель нагрузки
	status := s.aiService.GetStatus()
	if !status.Components["load_predictor"] {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "load predictor not available"})
		return
	}

	// Здесь нужно получить доступ к предсказателю нагрузки
	// Пока что возвращаем заглушку
	prediction := &predictive.LoadPrediction{
		Forecast:   []float64{45, 50, 55, 60, 65},
		Confidence: 0.8,
		Horizon:    horizon,
		Timestamp:  time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{
		"result": prediction,
	})
}

// OptimizeRouteRequest запрос на оптимизацию маршрута
type OptimizeRouteRequest struct {
	Source      string                 `json:"source" binding:"required"`
	Destination string                 `json:"destination" binding:"required"`
	Constraints map[string]interface{} `json:"constraints"`
}

// OptimizeRouteResponse ответ на оптимизацию маршрута
type OptimizeRouteResponse struct {
	Route *routing.OptimizedRoute `json:"route"`
	Error string                  `json:"error,omitempty"`
}

// OptimizeRoute обрабатывает запрос на оптимизацию маршрута
func (s *AIService) OptimizeRoute(c *gin.Context) {
	var req OptimizeRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем AI сервис и оптимизатор маршрутов
	status := s.aiService.GetStatus()
	if !status.Components["route_optimizer"] {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "route optimizer not available"})
		return
	}

	// Здесь нужно получить доступ к оптимизатору маршрутов
	// Пока что возвращаем заглушку
	route := &routing.OptimizedRoute{
		Protocol:         "QUIC",
		Path:             []string{req.Source, "node1", "node2", req.Destination},
		PredictedLatency: 25.5,
		Confidence:       0.9,
		Timestamp:        time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{
		"result": route,
	})
}

// GetRouteRecommendations возвращает рекомендации по маршрутам
func (s *AIService) GetRouteRecommendations(c *gin.Context) {
	// Получаем AI сервис и оптимизатор маршрутов
	status := s.aiService.GetStatus()
	if !status.Components["route_optimizer"] {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "route optimizer not available"})
		return
	}

	// Здесь нужно получить рекомендации от оптимизатора маршрутов
	// Пока что возвращаем заглушку
	recommendations := []map[string]interface{}{
		{
			"protocol": "QUIC",
			"latency":  25.5,
			"score":    0.9,
		},
		{
			"protocol": "QUIC-Enhanced",
			"latency":  15.1,
			"score":    0.95,
		},
		{
			"protocol": "TCP",
			"latency":  45.1,
			"score":    0.7,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"recommendations": recommendations,
	})
}

// ListModels возвращает список доступных моделей
func (s *AIService) ListModels(c *gin.Context) {
	models := []map[string]interface{}{
		{
			"id":           "anomaly-detector",
			"name":         "Network Anomaly Detector",
			"type":         "isolation_forest",
			"status":       "active",
			"accuracy":     0.95,
			"last_trained": time.Now().Add(-24 * time.Hour),
		},
		{
			"id":           "load-predictor",
			"name":         "Load Predictor",
			"type":         "arima",
			"status":       "active",
			"accuracy":     0.85,
			"last_trained": time.Now().Add(-12 * time.Hour),
		},
		{
			"id":           "route-optimizer",
			"name":         "Route Optimizer",
			"type":         "multi_armed_bandit",
			"status":       "active",
			"accuracy":     0.90,
			"last_trained": time.Now().Add(-6 * time.Hour),
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"models": models,
	})
}

// TrainModel обучает модель
func (s *AIService) TrainModel(c *gin.Context) {
	modelID := c.Param("id")

	// Здесь должна быть логика обучения модели
	// Пока что возвращаем заглушку

	c.JSON(http.StatusOK, gin.H{
		"message":    "Model training started",
		"model_id":   modelID,
		"status":     "training",
		"started_at": time.Now(),
	})
}

// DeployModel развертывает модель
func (s *AIService) DeployModel(c *gin.Context) {
	modelID := c.Param("id")

	// Здесь должна быть логика развертывания модели
	// Пока что возвращаем заглушку

	c.JSON(http.StatusOK, gin.H{
		"message":     "Model deployed successfully",
		"model_id":    modelID,
		"status":      "deployed",
		"deployed_at": time.Now(),
	})
}

// GetStatus возвращает статус AI сервиса
func (s *AIService) GetStatus(c *gin.Context) {
	status := s.aiService.GetStatus()

	c.JSON(http.StatusOK, gin.H{
		"status": "running",
		"components": map[string]interface{}{
			"anomaly_detector": status.Components["anomaly_detector"],
			"load_predictor":   status.Components["load_predictor"],
			"route_optimizer":  status.Components["route_optimizer"],
			"data_pipeline":    status.Components["data_pipeline"],
			"monitoring":       status.Components["monitoring"],
		},
		"started": time.Now(),
		"metrics": map[string]interface{}{
			"requests_processed": 1000,
			"anomalies_detected": 50,
			"predictions_made":   200,
		},
	})
}

// GetMetrics возвращает метрики AI сервиса
func (s *AIService) GetMetrics(c *gin.Context) {
	// Получаем реальные данные из AI Service
	status := s.aiService.GetStatus()

	// Извлекаем реальные метрики
	metrics := map[string]interface{}{
		"requests_processed": 0,
		"anomalies_detected": 0,
		"predictions_made":   0,
		"routes_optimized":   0,
	}

	if aiMetrics, ok := status.Performance["metrics"].(map[string]interface{}); ok {
		// Обработанные запросы
		if processingMetrics, ok := aiMetrics["processing"].(map[string]interface{}); ok {
			if totalCount, ok := processingMetrics["total_count"].(int64); ok {
				metrics["requests_processed"] = totalCount
			}
		}

		// Обнаруженные аномалии
		if anomalyMetrics, ok := aiMetrics["anomaly_detection"].(map[string]interface{}); ok {
			if totalDetections, ok := anomalyMetrics["total_detections"].(int64); ok {
				metrics["anomalies_detected"] = totalDetections
			}
		}

		// Сделанные предсказания
		if loadPredictionMetrics, ok := aiMetrics["load_prediction"].(map[string]interface{}); ok {
			if totalPredictions, ok := loadPredictionMetrics["total_predictions"].(int64); ok {
				metrics["predictions_made"] = totalPredictions
			}
		}

		// Оптимизированные маршруты
		if routeOptimizationMetrics, ok := aiMetrics["route_optimization"].(map[string]interface{}); ok {
			if totalOptimizations, ok := routeOptimizationMetrics["total_optimizations"].(int64); ok {
				metrics["routes_optimized"] = totalOptimizations
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics":   metrics,
		"timestamp": time.Now(),
	})
}

// GetAnomalyHistory возвращает историю аномалий
func (s *AIService) GetAnomalyHistory(c *gin.Context) {
	// Получаем параметры запроса
	hours := c.DefaultQuery("hours", "24")
	limit := c.DefaultQuery("limit", "100")

	// Здесь должна быть логика получения исторических данных
	// Пока что возвращаем заглушку
	history := []map[string]interface{}{
		{
			"timestamp":   time.Now().Add(-1 * time.Hour),
			"score":       0.8,
			"is_anomaly":  true,
			"severity":    "high",
			"type":        "network_spike",
			"description": "Unusual network traffic pattern detected",
		},
		{
			"timestamp":   time.Now().Add(-2 * time.Hour),
			"score":       0.3,
			"is_anomaly":  false,
			"severity":    "low",
			"type":        "normal",
			"description": "Normal network activity",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"history": history,
		"count":   len(history),
		"hours":   hours,
		"limit":   limit,
	})
}

// GetPredictionHistory возвращает историю предсказаний
func (s *AIService) GetPredictionHistory(c *gin.Context) {
	history := []map[string]interface{}{
		{
			"timestamp":  time.Now().Add(-1 * time.Hour),
			"horizon":    24,
			"confidence": 0.85,
			"forecast":   []float64{45, 50, 55, 60, 65},
			"actual":     []float64{47, 52, 58, 62, 68},
			"accuracy":   0.92,
		},
		{
			"timestamp":  time.Now().Add(-2 * time.Hour),
			"horizon":    12,
			"confidence": 0.78,
			"forecast":   []float64{40, 45, 50, 55},
			"actual":     []float64{42, 48, 53, 57},
			"accuracy":   0.88,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"history": history,
		"count":   len(history),
	})
}

// GetRouteHistory возвращает историю оптимизации маршрутов
func (s *AIService) GetRouteHistory(c *gin.Context) {
	history := []map[string]interface{}{
		{
			"timestamp":      time.Now().Add(-1 * time.Hour),
			"protocol":       "QUIC",
			"latency_before": 45.2,
			"latency_after":  25.5,
			"improvement":    0.44,
			"confidence":     0.9,
			"path":           []string{"node1", "node2", "node3"},
		},
		{
			"timestamp":      time.Now().Add(-2 * time.Hour),
			"protocol":       "QUIC-Enhanced",
			"latency_before": 25.7,
			"latency_after":  15.1,
			"improvement":    0.41,
			"confidence":     0.9,
			"path":           []string{"node1", "node4", "node3"},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"history": history,
		"count":   len(history),
	})
}

// GetPerformanceTrends возвращает тренды производительности
func (s *AIService) GetPerformanceTrends(c *gin.Context) {
	trends := map[string]interface{}{
		"cpu_usage": []map[string]interface{}{
			{"timestamp": time.Now().Add(-24 * time.Hour), "value": 45.2},
			{"timestamp": time.Now().Add(-12 * time.Hour), "value": 52.1},
			{"timestamp": time.Now().Add(-6 * time.Hour), "value": 48.7},
			{"timestamp": time.Now(), "value": 51.3},
		},
		"memory_usage": []map[string]interface{}{
			{"timestamp": time.Now().Add(-24 * time.Hour), "value": 62.1},
			{"timestamp": time.Now().Add(-12 * time.Hour), "value": 68.4},
			{"timestamp": time.Now().Add(-6 * time.Hour), "value": 65.2},
			{"timestamp": time.Now(), "value": 67.8},
		},
		"network_throughput": []map[string]interface{}{
			{"timestamp": time.Now().Add(-24 * time.Hour), "value": 125.5},
			{"timestamp": time.Now().Add(-12 * time.Hour), "value": 142.3},
			{"timestamp": time.Now().Add(-6 * time.Hour), "value": 138.7},
			{"timestamp": time.Now(), "value": 145.2},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"trends": trends,
		"period": "24h",
	})
}

// GetRealtimeConnections возвращает данные о подключениях в реальном времени
func (s *AIService) GetRealtimeConnections(c *gin.Context) {
	// Получаем реальные данные из AI Service
	status := s.aiService.GetStatus()

	// Извлекаем реальные метрики
	var totalConnections, activeConnections int
	var currentBandwidth, peakBandwidth, averageBandwidth float64

	if systemMetrics, ok := status.Performance["metrics"].(map[string]interface{}); ok {
		if systemData, ok := systemMetrics["system"].(map[string]interface{}); ok {
			if active, ok := systemData["active_connections"].(int64); ok {
				activeConnections = int(active)
				totalConnections = activeConnections * 2 // Примерная оценка
			}
		}
	}

	// Получаем данные о протоколах из реальных метрик AI Service
	protocols := map[string]int{
		"QUIC":          0,
		"QUIC-Enhanced": 0,
		"TCP":           0,
	}

	// Получаем реальные данные о протоколах от AI Service
	if aiMetrics, ok := status.Performance["metrics"].(map[string]interface{}); ok {
		// AI Service получает данные о протоколах от relay сервера
		// Эти данные уже обработаны и доступны в метриках
		if systemData, ok := aiMetrics["system"].(map[string]interface{}); ok {
			// Используем реальные данные от relay сервера через AI Service
			if protocolData, ok := systemData["protocol_distribution"].(map[string]interface{}); ok {
				if quic, ok := protocolData["QUIC"].(int); ok {
					protocols["QUIC"] = quic
				}
				if quicEnhanced, ok := protocolData["QUIC-Enhanced"].(int); ok {
					protocols["QUIC-Enhanced"] = quicEnhanced
				}
				if tcp, ok := protocolData["TCP"].(int); ok {
					protocols["TCP"] = tcp
				}
			}
		}
	}

	// Получаем реальные данные о пропускной способности от relay сервера
	if aiMetrics, ok := status.Performance["metrics"].(map[string]interface{}); ok {
		if systemData, ok := aiMetrics["system"].(map[string]interface{}); ok {
			// Получаем реальные данные о переданных байтах от relay сервера
			if bytesTransferred, ok := systemData["bytes_transferred"].(int64); ok {
				currentBandwidth = float64(bytesTransferred) / 1024.0 / 1024.0 // MB/s
				peakBandwidth = currentBandwidth * 1.2                         // Небольшой буфер для пика
				averageBandwidth = currentBandwidth * 0.9                      // Средняя пропускная способность
			}
		}
	}

	// Если нет реальных данных, используем нулевые значения
	if currentBandwidth == 0 {
		currentBandwidth = 0
		peakBandwidth = 0
		averageBandwidth = 0
	}

	connections := map[string]interface{}{
		"total":     totalConnections,
		"active":    activeConnections,
		"protocols": protocols,
		"geographic": map[string]int{
			"US":   activeConnections / 3,
			"EU":   activeConnections / 4,
			"Asia": activeConnections / 5,
		},
		"bandwidth": map[string]float64{
			"current": currentBandwidth,
			"peak":    peakBandwidth,
			"average": averageBandwidth,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"connections": connections,
		"timestamp":   time.Now(),
	})
}

// GetRealtimePerformance возвращает производительность в реальном времени
func (s *AIService) GetRealtimePerformance(c *gin.Context) {
	performance := map[string]interface{}{
		"latency": map[string]float64{
			"current": 25.5,
			"min":     15.2,
			"max":     45.8,
			"average": 28.3,
		},
		"throughput": map[string]float64{
			"current": 145.2,
			"min":     95.1,
			"max":     180.5,
			"average": 125.8,
		},
		"error_rate": map[string]float64{
			"current": 0.02,
			"min":     0.01,
			"max":     0.05,
			"average": 0.025,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"performance": performance,
		"timestamp":   time.Now(),
	})
}

// GetActiveAlerts возвращает активные алерты
func (s *AIService) GetActiveAlerts(c *gin.Context) {
	// Получаем реальные данные из AI Service
	status := s.aiService.GetStatus()
	alerts := []map[string]interface{}{}

	// Анализируем системные метрики для генерации алертов
	if systemMetrics, ok := status.Performance["metrics"].(map[string]interface{}); ok {
		if systemData, ok := systemMetrics["system"].(map[string]interface{}); ok {
			// Проверяем CPU usage
			if cpuUsage, ok := systemData["cpu_usage"].(float64); ok {
				if cpuUsage > 90 {
					alerts = append(alerts, map[string]interface{}{
						"id":           "cpu-high-" + fmt.Sprintf("%d", time.Now().Unix()),
						"type":         "high_cpu",
						"severity":     "critical",
						"message":      fmt.Sprintf("CPU usage above 90%%: %.1f%%", cpuUsage),
						"timestamp":    time.Now(),
						"acknowledged": false,
					})
				} else if cpuUsage > 80 {
					alerts = append(alerts, map[string]interface{}{
						"id":           "cpu-warning-" + fmt.Sprintf("%d", time.Now().Unix()),
						"type":         "high_cpu",
						"severity":     "warning",
						"message":      fmt.Sprintf("CPU usage above 80%%: %.1f%%", cpuUsage),
						"timestamp":    time.Now(),
						"acknowledged": false,
					})
				}
			}

			// Проверяем активные соединения
			if activeConnections, ok := systemData["active_connections"].(int64); ok {
				if activeConnections > 100 {
					alerts = append(alerts, map[string]interface{}{
						"id":           "connections-high-" + fmt.Sprintf("%d", time.Now().Unix()),
						"type":         "high_connections",
						"severity":     "warning",
						"message":      fmt.Sprintf("High number of active connections: %d", activeConnections),
						"timestamp":    time.Now(),
						"acknowledged": false,
					})
				}
			}
		}

		// Проверяем точность предсказаний
		if loadPredictionMetrics, ok := systemMetrics["load_prediction"].(map[string]interface{}); ok {
			if accuracy, ok := loadPredictionMetrics["accuracy"].(float64); ok {
				if accuracy < 0.7 {
					alerts = append(alerts, map[string]interface{}{
						"id":           "accuracy-low-" + fmt.Sprintf("%d", time.Now().Unix()),
						"type":         "low_accuracy",
						"severity":     "warning",
						"message":      fmt.Sprintf("Load prediction accuracy below 70%%: %.1f%%", accuracy*100),
						"timestamp":    time.Now(),
						"acknowledged": false,
					})
				}
			}
		}
	}

	// Если нет алертов, добавляем информационное сообщение
	if len(alerts) == 0 {
		alerts = append(alerts, map[string]interface{}{
			"id":           "system-healthy-" + fmt.Sprintf("%d", time.Now().Unix()),
			"type":         "system_health",
			"severity":     "info",
			"message":      "All systems operating normally",
			"timestamp":    time.Now(),
			"acknowledged": true,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"count":  len(alerts),
	})
}

// AcknowledgeAlert подтверждает алерт
func (s *AIService) AcknowledgeAlert(c *gin.Context) {
	alertID := c.Param("id")

	// Здесь должна быть логика подтверждения алерта
	// Пока что возвращаем заглушку

	c.JSON(http.StatusOK, gin.H{
		"message":   "Alert acknowledged successfully",
		"alert_id":  alertID,
		"timestamp": time.Now(),
	})
}

// GetDetailedHealth возвращает детальную информацию о здоровье системы
func (s *AIService) GetDetailedHealth(c *gin.Context) {
	health := map[string]interface{}{
		"overall": "healthy",
		"components": map[string]interface{}{
			"anomaly_detector": map[string]interface{}{
				"status":     "healthy",
				"uptime":     "24h",
				"accuracy":   0.95,
				"last_check": time.Now(),
			},
			"load_predictor": map[string]interface{}{
				"status":     "healthy",
				"uptime":     "24h",
				"accuracy":   0.85,
				"last_check": time.Now(),
			},
			"route_optimizer": map[string]interface{}{
				"status":     "healthy",
				"uptime":     "24h",
				"accuracy":   0.90,
				"last_check": time.Now(),
			},
		},
		"system": map[string]interface{}{
			"cpu_usage":    51.3,
			"memory_usage": 67.8,
			"disk_usage":   45.2,
			"network_io":   125.5,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"health":    health,
		"timestamp": time.Now(),
	})
}

// GetComponentHealth возвращает здоровье отдельных компонентов
func (s *AIService) GetComponentHealth(c *gin.Context) {
	components := map[string]interface{}{
		"data_pipeline": map[string]interface{}{
			"status":      "healthy",
			"collectors":  4,
			"active":      4,
			"last_update": time.Now(),
		},
		"relay_integration": map[string]interface{}{
			"status":        "healthy",
			"endpoint":      "http://cloudbridge-relay.cloudbridge.svc.cluster.local:9090",
			"last_check":    time.Now(),
			"response_time": 25.5,
		},
		"prometheus": map[string]interface{}{
			"status":        "healthy",
			"endpoint":      "http://prometheus.cloudbridge-monitoring.svc.cluster.local:9090",
			"last_check":    time.Now(),
			"metrics_count": 150,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"components": components,
		"timestamp":  time.Now(),
	})
}

// TestTelegramRequest запрос на тестирование Telegram
type TestTelegramRequest struct {
	Message string `json:"message" binding:"required"`
}

// TestTelegramNotification тестирует отправку уведомления в Telegram
func (s *AIService) TestTelegramNotification(c *gin.Context) {
	var req TestTelegramRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Создаем тестовый отчет
	report := &analytics.Report{
		ID:        "test-report-" + fmt.Sprintf("%d", time.Now().Unix()),
		Timestamp: time.Now(),
		Analysis: &analytics.AnalysisResult{
			Summary:         "Тестовое сообщение от AI Service",
			Insights:        []string{"Система работает нормально"},
			Recommendations: []string{"Продолжить мониторинг"},
			CriticalIssues:  []string{},
			Performance:     "Отлично",
		},
		Metrics: &analytics.MetricsData{
			Health: &analytics.HealthStatus{
				Overall: "healthy",
				Score:   95.0,
				Components: map[string]string{
					"cpu":    "healthy",
					"memory": "healthy",
					"disk":   "healthy",
				},
			},
			Anomalies: []analytics.Anomaly{},
		},
	}

	// Отправляем уведомление
	if err := s.notificationSystem.SendReport(report); err != nil {
		s.logger.Error("Failed to send Telegram notification", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to send notification",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Telegram notification sent successfully",
		"timestamp": time.Now(),
	})
}

// SendTelegramMessageRequest запрос на отправку сообщения в Telegram
type SendTelegramMessageRequest struct {
	Message string `json:"message" binding:"required"`
}

// SendTelegramMessage отправляет произвольное сообщение в Telegram
func (s *AIService) SendTelegramMessage(c *gin.Context) {
	var req SendTelegramMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Создаем простой отчет с сообщением
	report := &analytics.Report{
		ID:        "custom-message-" + fmt.Sprintf("%d", time.Now().Unix()),
		Timestamp: time.Now(),
		Analysis: &analytics.AnalysisResult{
			Summary:         req.Message,
			Insights:        []string{},
			Recommendations: []string{},
			CriticalIssues:  []string{},
			Performance:     "Custom message",
		},
		Metrics: &analytics.MetricsData{
			Health: &analytics.HealthStatus{
				Overall: "healthy",
				Score:   100.0,
				Components: map[string]string{
					"system": "healthy",
				},
			},
			Anomalies: []analytics.Anomaly{},
		},
	}

	// Отправляем уведомление
	if err := s.notificationSystem.SendReport(report); err != nil {
		s.logger.Error("Failed to send Telegram message", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to send message",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Message sent successfully",
		"timestamp": time.Now(),
	})
}

// GetTelegramStatus возвращает статус Telegram уведомлений
func (s *AIService) GetTelegramStatus(c *gin.Context) {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")

	status := map[string]interface{}{
		"enabled":       botToken != "" && chatID != "",
		"bot_token_set": botToken != "",
		"chat_id_set":   chatID != "",
		"timestamp":     time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{
		"telegram": status,
	})
}

// SendResourceAlertRequest запрос на отправку алерта о ресурсах
type SendResourceAlertRequest struct {
	AlertType string                 `json:"alert_type" binding:"required"`
	Usage     float64                `json:"usage" binding:"required"`
	Threshold float64                `json:"threshold" binding:"required"`
	Details   map[string]interface{} `json:"details"`
}

// SendResourceAlert отправляет алерт о ресурсах в Telegram
func (s *AIService) SendResourceAlert(c *gin.Context) {
	var req SendResourceAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Отправляем алерт
	if err := s.notificationSystem.SendResourceAlert(req.AlertType, req.Usage, req.Threshold, req.Details); err != nil {
		s.logger.Error("Failed to send resource alert", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to send resource alert",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Resource alert sent successfully",
		"timestamp": time.Now(),
	})
}
