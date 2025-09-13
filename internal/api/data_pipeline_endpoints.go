package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/pipeline"
	"go.uber.org/zap"
)

// DataPipelineEndpoints предоставляет Data Pipeline API endpoints
type DataPipelineEndpoints struct {
	pipeline *pipeline.DataPipeline
	logger   *zap.Logger
}

// NewDataPipelineEndpoints создает новый DataPipelineEndpoints
func NewDataPipelineEndpoints(pipeline *pipeline.DataPipeline, logger *zap.Logger) *DataPipelineEndpoints {
	return &DataPipelineEndpoints{
		pipeline: pipeline,
		logger:   logger,
	}
}

// PipelineStatusResponse ответ для статуса pipeline
type PipelineStatusResponse struct {
	Status     string                 `json:"status"`
	Collectors map[string]interface{} `json:"collectors"`
	Processors map[string]interface{} `json:"processors"`
	Storage    map[string]interface{} `json:"storage"`
	Timestamp  time.Time              `json:"timestamp"`
}

// DataQueryRequest запрос для запроса данных
type DataQueryRequest struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	DataType  string    `json:"data_type"`
	Limit     int       `json:"limit"`
}

// GetPipelineStatus возвращает статус Data Pipeline
func (dpe *DataPipelineEndpoints) GetPipelineStatus(c *gin.Context) {
	dpe.logger.Info("Getting data pipeline status")

	if dpe.pipeline == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "error",
			"message":   "Data pipeline is not available",
			"timestamp": time.Now(),
		})
		return
	}

	// Получаем статус коллекторов (заглушка)
	collectors := make(map[string]interface{})
	collectors["prometheus"] = map[string]interface{}{
		"status": "running",
		"type":   "prometheus",
	}
	collectors["system"] = map[string]interface{}{
		"status": "running",
		"type":   "system",
	}

	// Получаем статус процессоров (заглушка)
	processors := make(map[string]interface{})
	processors["feature_extractor"] = map[string]interface{}{
		"status": "running",
		"type":   "feature_extractor",
	}
	processors["data_normalizer"] = map[string]interface{}{
		"status": "running",
		"type":   "data_normalizer",
	}

	// Получаем статус хранилища (заглушка)
	storage := make(map[string]interface{})
	storage["status"] = "running"
	storage["type"] = "hybrid_storage"

	response := PipelineStatusResponse{
		Status:     "healthy",
		Collectors: collectors,
		Processors: processors,
		Storage:    storage,
		Timestamp:  time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// GetCollectorData возвращает данные от коллектора
func (dpe *DataPipelineEndpoints) GetCollectorData(c *gin.Context) {
	collectorName := c.Param("collector")
	if collectorName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":    "error",
			"message":   "Collector name is required",
			"timestamp": time.Now(),
		})
		return
	}

	dpe.logger.Info("Getting collector data", zap.String("collector", collectorName))

	if dpe.pipeline == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "error",
			"message":   "Data pipeline is not available",
			"timestamp": time.Now(),
		})
		return
	}

	// Заглушка для получения данных коллектора
	data := map[string]interface{}{
		"collector": collectorName,
		"timestamp": time.Now(),
		"data": map[string]interface{}{
			"cpu_usage":    45.2,
			"memory_usage": 67.8,
			"disk_usage":   23.1,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"collector": collectorName,
		"data":      data,
		"timestamp": time.Now(),
	})
}

// QueryData запрашивает данные из pipeline
func (dpe *DataPipelineEndpoints) QueryData(c *gin.Context) {
	var req DataQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dpe.logger.Error("Invalid data query request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"status":    "error",
			"message":   "Invalid request format",
			"timestamp": time.Now(),
		})
		return
	}

	// Устанавливаем значения по умолчанию
	if req.EndTime.IsZero() {
		req.EndTime = time.Now()
	}
	if req.StartTime.IsZero() {
		req.StartTime = req.EndTime.Add(-24 * time.Hour) // Последние 24 часа
	}
	if req.Limit == 0 {
		req.Limit = 1000
	}

	dpe.logger.Info("Querying data",
		zap.Time("start_time", req.StartTime),
		zap.Time("end_time", req.EndTime),
		zap.String("data_type", req.DataType),
		zap.Int("limit", req.Limit))

	if dpe.pipeline == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "error",
			"message":   "Data pipeline is not available",
			"timestamp": time.Now(),
		})
		return
	}

	// Создаем time range
	timeRange := pipeline.TimeRange{
		Start: req.StartTime,
		End:   req.EndTime,
	}

	// Заглушка для запроса данных
	data := []map[string]interface{}{
		{
			"timestamp": time.Now().Add(-1 * time.Hour),
			"value":     42.5,
			"type":      req.DataType,
		},
		{
			"timestamp": time.Now().Add(-30 * time.Minute),
			"value":     38.2,
			"type":      req.DataType,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"data_type":  req.DataType,
		"time_range": timeRange,
		"data":       data,
		"count":      len(data),
		"timestamp":  time.Now(),
	})
}

// GetPipelineMetrics возвращает метрики Data Pipeline
func (dpe *DataPipelineEndpoints) GetPipelineMetrics(c *gin.Context) {
	dpe.logger.Info("Getting data pipeline metrics")

	if dpe.pipeline == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "error",
			"message":   "Data pipeline is not available",
			"timestamp": time.Now(),
		})
		return
	}

	metrics := make(map[string]interface{})

	// Метрики коллекторов (заглушка)
	collectorMetrics := make(map[string]interface{})
	collectorMetrics["prometheus"] = map[string]interface{}{
		"status": "running",
		"type":   "prometheus",
	}
	collectorMetrics["system"] = map[string]interface{}{
		"status": "running",
		"type":   "system",
	}
	metrics["collectors"] = collectorMetrics

	// Метрики процессоров (заглушка)
	processorMetrics := make(map[string]interface{})
	processorMetrics["feature_extractor"] = map[string]interface{}{
		"status": "running",
		"type":   "feature_extractor",
	}
	processorMetrics["data_normalizer"] = map[string]interface{}{
		"status": "running",
		"type":   "data_normalizer",
	}
	metrics["processors"] = processorMetrics

	// Метрики хранилища (заглушка)
	metrics["storage"] = map[string]interface{}{
		"status": "running",
		"type":   "hybrid_storage",
	}

	// Общие метрики
	metrics["pipeline_status"] = "healthy"
	metrics["timestamp"] = time.Now()

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"metrics": metrics,
	})
}

// StartPipeline запускает Data Pipeline
func (dpe *DataPipelineEndpoints) StartPipeline(c *gin.Context) {
	dpe.logger.Info("Starting data pipeline")

	if dpe.pipeline == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "error",
			"message":   "Data pipeline is not available",
			"timestamp": time.Now(),
		})
		return
	}

	err := dpe.pipeline.Start()
	if err != nil {
		dpe.logger.Error("Failed to start pipeline", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":    "error",
			"message":   "Failed to start pipeline",
			"timestamp": time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"message":   "Data pipeline started",
		"timestamp": time.Now(),
	})
}

// StopPipeline останавливает Data Pipeline
func (dpe *DataPipelineEndpoints) StopPipeline(c *gin.Context) {
	dpe.logger.Info("Stopping data pipeline")

	if dpe.pipeline == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "error",
			"message":   "Data pipeline is not available",
			"timestamp": time.Now(),
		})
		return
	}

	err := dpe.pipeline.Stop()
	if err != nil {
		dpe.logger.Error("Failed to stop pipeline", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":    "error",
			"message":   "Failed to stop pipeline",
			"timestamp": time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"message":   "Data pipeline stopped",
		"timestamp": time.Now(),
	})
}

// GetDataTypes возвращает доступные типы данных
func (dpe *DataPipelineEndpoints) GetDataTypes(c *gin.Context) {
	dpe.logger.Info("Getting available data types")

	dataTypes := []string{
		"prometheus_metrics",
		"system_metrics",
		"network_metrics",
		"relay_metrics",
		"processed_data",
		"aggregated_data",
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"data_types": dataTypes,
		"timestamp":  time.Now(),
	})
}
