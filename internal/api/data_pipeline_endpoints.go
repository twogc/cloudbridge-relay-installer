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
	pipeline       *pipeline.DataPipeline
	logger         *zap.Logger
	inputValidator *InputValidator
}

// NewDataPipelineEndpoints создает новый DataPipelineEndpoints
func NewDataPipelineEndpoints(pipeline *pipeline.DataPipeline, logger *zap.Logger) *DataPipelineEndpoints {
	return &DataPipelineEndpoints{
		pipeline:       pipeline,
		logger:         logger,
		inputValidator: NewInputValidator(logger),
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

	// Валидируем имя коллектора
	validationResult := dpe.inputValidator.ValidateCollectorName(collectorName)
	if !validationResult.Valid {
		dpe.inputValidator.HandleValidationError(c, validationResult)
		return
	}

	// Используем валидированное имя коллектора
	collectorName = validationResult.Data.(string)

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

	// Валидируем запрос
	validationResult := dpe.inputValidator.ValidateDataQueryRequest(&req, nil)
	if !validationResult.Valid {
		dpe.inputValidator.HandleValidationError(c, validationResult)
		return
	}

	// Используем валидированные данные
	req = *validationResult.Data.(*DataQueryRequest)

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

	// Получаем реальные данные из pipeline
	data, err := dpe.pipeline.QueryData(timeRange, req.DataType, req.Limit)
	if err != nil {
		dpe.logger.Error("Failed to query data from pipeline", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":    "error",
			"message":   "Failed to query data from pipeline",
			"timestamp": time.Now(),
		})
		return
	}

	// Преобразуем данные в нужный формат
	responseData := make([]map[string]interface{}, 0, len(data))
	for _, dataPoint := range data {
		responseData = append(responseData, map[string]interface{}{
			"timestamp": dataPoint.Timestamp,
			"value":     dataPoint.Data,
			"type":      req.DataType,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"data_type":  req.DataType,
		"time_range": timeRange,
		"data":       responseData,
		"count":      len(responseData),
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

	// Получаем реальные метрики из pipeline
	metrics, err := dpe.pipeline.GetMetrics()
	if err != nil {
		dpe.logger.Error("Failed to get pipeline metrics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":    "error",
			"message":   "Failed to get pipeline metrics",
			"timestamp": time.Now(),
		})
		return
	}

	// Добавляем timestamp
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
