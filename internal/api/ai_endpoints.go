package api

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/anomaly"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/predictive"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/routing"
	"go.uber.org/zap"
)

// AIEndpoints предоставляет AI API endpoints
type AIEndpoints struct {
	aiService AIServiceInterface
	logger    *zap.Logger
}

// NewAIEndpoints создает новый AIEndpoints
func NewAIEndpoints(aiService AIServiceInterface, logger *zap.Logger) *AIEndpoints {
	return &AIEndpoints{
		aiService: aiService,
		logger:    logger,
	}
}

// AIServiceInterface определяет интерфейс для AI сервиса
type AIServiceInterface interface {
	GetStatus() *ai.ServiceStatus
	DetectAnomaly(ctx context.Context, metrics *anomaly.NetworkMetrics) (*anomaly.AnomalyResult, error)
	PredictLoad(ctx context.Context, horizon int) (*predictive.LoadPrediction, error)
	OptimizeRoute(ctx context.Context, source, destination string, constraints map[string]interface{}) (*routing.OptimizedRoute, error)
}

// AIStatusResponse ответ для AI статуса
type AIStatusResponse struct {
	Status     string                 `json:"status"`
	Components map[string]interface{} `json:"components"`
	Timestamp  time.Time              `json:"timestamp"`
}

// AnomalyDetectionRequest запрос для детекции аномалий
type AnomalyDetectionRequest struct {
	Metrics *anomaly.NetworkMetrics `json:"metrics"`
}

// LoadPredictionRequest запрос для предсказания нагрузки
type LoadPredictionRequest struct {
	Horizon int `json:"horizon"`
}

// RouteOptimizationRequest запрос для оптимизации маршрута
type RouteOptimizationRequest struct {
	Source      string                 `json:"source"`
	Destination string                 `json:"destination"`
	Constraints map[string]interface{} `json:"constraints"`
}

// GetAIStatus возвращает статус AI сервиса
func (ae *AIEndpoints) GetAIStatus(c *gin.Context) {
	ae.logger.Info("Getting AI service status")

	if ae.aiService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "error",
			"message":   "AI service is not available",
			"timestamp": time.Now(),
		})
		return
	}

	status := ae.aiService.GetStatus()
	if status == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "error",
			"message":   "AI service is not available",
			"timestamp": time.Now(),
		})
		return
	}

	// Конвертируем map[string]bool в map[string]interface{}
	components := make(map[string]interface{})
	for k, v := range status.Components {
		components[k] = v
	}

	response := AIStatusResponse{
		Status:     "healthy",
		Components: components,
		Timestamp:  time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// DetectAnomaly детектирует аномалии в сетевых метриках
func (ae *AIEndpoints) DetectAnomaly(c *gin.Context) {
	if ae.aiService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "error",
			"message":   "AI service is not available",
			"timestamp": time.Now(),
		})
		return
	}

	var req AnomalyDetectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ae.logger.Error("Invalid anomaly detection request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"status":    "error",
			"message":   "Invalid request format",
			"timestamp": time.Now(),
		})
		return
	}

	if req.Metrics == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":    "error",
			"message":   "Metrics are required",
			"timestamp": time.Now(),
		})
		return
	}

	ae.logger.Info("Detecting anomaly", zap.Any("metrics", req.Metrics))

	result, err := ae.aiService.DetectAnomaly(c.Request.Context(), req.Metrics)
	if err != nil {
		ae.logger.Error("Anomaly detection failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":    "error",
			"message":   "Anomaly detection failed",
			"timestamp": time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"result":    result,
		"timestamp": time.Now(),
	})
}

// PredictLoad предсказывает нагрузку
func (ae *AIEndpoints) PredictLoad(c *gin.Context) {
	if ae.aiService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "error",
			"message":   "AI service is not available",
			"timestamp": time.Now(),
		})
		return
	}

	horizonStr := c.DefaultQuery("horizon", "24")
	horizon, err := strconv.Atoi(horizonStr)
	if err != nil {
		ae.logger.Error("Invalid horizon parameter", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"status":    "error",
			"message":   "Invalid horizon parameter",
			"timestamp": time.Now(),
		})
		return
	}

	ae.logger.Info("Predicting load", zap.Int("horizon", horizon))

	prediction, err := ae.aiService.PredictLoad(c.Request.Context(), horizon)
	if err != nil {
		ae.logger.Error("Load prediction failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":    "error",
			"message":   "Load prediction failed",
			"timestamp": time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"prediction": prediction,
		"timestamp":  time.Now(),
	})
}

// OptimizeRoute оптимизирует маршрут
func (ae *AIEndpoints) OptimizeRoute(c *gin.Context) {
	if ae.aiService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "error",
			"message":   "AI service is not available",
			"timestamp": time.Now(),
		})
		return
	}

	var req RouteOptimizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ae.logger.Error("Invalid route optimization request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"status":    "error",
			"message":   "Invalid request format",
			"timestamp": time.Now(),
		})
		return
	}

	if req.Source == "" || req.Destination == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":    "error",
			"message":   "Source and destination are required",
			"timestamp": time.Now(),
		})
		return
	}

	ae.logger.Info("Optimizing route",
		zap.String("source", req.Source),
		zap.String("destination", req.Destination),
		zap.Any("constraints", req.Constraints))

	route, err := ae.aiService.OptimizeRoute(c.Request.Context(), req.Source, req.Destination, req.Constraints)
	if err != nil {
		ae.logger.Error("Route optimization failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":    "error",
			"message":   "Route optimization failed",
			"timestamp": time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"route":     route,
		"timestamp": time.Now(),
	})
}

// GetAIMetrics возвращает метрики AI сервиса
func (ae *AIEndpoints) GetAIMetrics(c *gin.Context) {
	ae.logger.Info("Getting AI metrics")

	if ae.aiService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "error",
			"message":   "AI service is not available",
			"timestamp": time.Now(),
		})
		return
	}

	status := ae.aiService.GetStatus()
	if status == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "error",
			"message":   "AI service is not available",
			"timestamp": time.Now(),
		})
		return
	}

	// Извлекаем метрики из статуса
	metrics := make(map[string]interface{})
	// TODO: добавить метрики когда они будут доступны в ServiceStatus

	// Добавляем общую информацию
	metrics["service_status"] = status.Status
	metrics["components_count"] = len(status.Components)
	metrics["timestamp"] = time.Now()

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"metrics": metrics,
	})
}

// GetAIHealth возвращает health check для AI сервиса
func (ae *AIEndpoints) GetAIHealth(c *gin.Context) {
	ae.logger.Info("Getting AI health")

	if ae.aiService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "unhealthy",
			"message":   "AI service is not available",
			"timestamp": time.Now(),
		})
		return
	}

	status := ae.aiService.GetStatus()
	if status == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "unhealthy",
			"message":   "AI service is not available",
			"timestamp": time.Now(),
		})
		return
	}

	healthStatus := "healthy"
	if status.Status != "running" {
		healthStatus = "unhealthy"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     healthStatus,
		"service":    "ai-service",
		"components": status.Components,
		"timestamp":  time.Now(),
	})
}
