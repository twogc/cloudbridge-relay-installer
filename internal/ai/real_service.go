package ai

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/ai/anomaly"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/cadence"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/pipeline"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/predictive"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/routing"
	"go.uber.org/zap"
)

// RealService представляет улучшенный AI сервис с реальными данными
type RealService struct {
	config *Config
	logger *zap.Logger

	// Реальные компоненты
	realDataCollector   *pipeline.RealDataCollector
	realAnomalyDetector *anomaly.RealAnomalyDetector
	realLoadPredictor   *predictive.RealLoadPredictor
	cadenceIntegration  *cadence.CadenceIntegration

	// Компоненты маршрутизации с реальными данными
	pathFinder       *routing.PathFinder
	latencyPredictor *routing.LatencyPredictor

	// Мониторинг и метрики
	monitoring *AIMonitoring
	metrics    *AIMetrics

	// Состояние сервиса
	started   bool
	startTime time.Time
	stopCh    chan struct{}
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewRealService создает новый AI сервис с реальными данными
func NewRealService(config *Config, logger *zap.Logger) *RealService {
	ctx, cancel := context.WithCancel(context.Background())

	return &RealService{
		config:  config,
		logger:  logger,
		metrics: NewAIMetrics(),
		stopCh:  make(chan struct{}),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Initialize инициализирует AI сервис с реальными компонентами
func (rs *RealService) Initialize() error {
	rs.logger.Info("Initializing Real AI service...")

	// Инициализируем коллектор реальных данных
	rs.realDataCollector = pipeline.NewRealDataCollector(
		rs.logger,
		rs.config.AI.DataSources.PrometheusURL,
		rs.config.AI.DataSources.CloudBridgeURL,
		rs.config.AI.DataSources.CadenceURL,
	)

	// Инициализируем детектор аномалий с реальными данными
	rs.realAnomalyDetector = anomaly.NewRealAnomalyDetector(rs.logger, rs.realDataCollector)
	if err := rs.realAnomalyDetector.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize real anomaly detector: %w", err)
	}

	// Инициализируем предиктор нагрузки с реальными данными
	rs.realLoadPredictor = predictive.NewRealLoadPredictor(rs.logger, rs.realDataCollector)
	if err := rs.realLoadPredictor.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize real load predictor: %w", err)
	}

	// Инициализируем интеграцию с Cadence
	rs.cadenceIntegration = cadence.NewCadenceIntegration(
		rs.logger,
		rs.config.AI.DataSources.CadenceURL,
		rs.config.AI.Cadence.Domain,
	)

	// Инициализируем компоненты маршрутизации с реальными данными
	rs.pathFinder = routing.NewPathFinder(
		rs.logger,
		rs.config.AI.DataSources.PrometheusURL,
		rs.config.AI.DataSources.CloudBridgeURL,
	)
	if err := rs.pathFinder.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize path finder: %w", err)
	}

	rs.latencyPredictor = routing.NewLatencyPredictor(
		rs.logger,
		rs.config.AI.DataSources.PrometheusURL,
		rs.config.AI.DataSources.CloudBridgeURL,
	)
	if err := rs.latencyPredictor.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize latency predictor: %w", err)
	}

	// Инициализируем мониторинг
	rs.monitoring = NewAIMonitoring()

	rs.logger.Info("Real AI service initialized successfully")
	return nil
}

// Start запускает AI сервис
func (rs *RealService) Start() error {
	if rs.started {
		return fmt.Errorf("service already started")
	}

	rs.logger.Info("Starting Real AI service...")
	rs.startTime = time.Now()
	rs.started = true

	// Запускаем периодический сбор данных
	rs.wg.Add(1)
	go func() {
		defer rs.wg.Done()
		rs.realDataCollector.StartPeriodicCollection(rs.ctx)
	}()

	// Запускаем мониторинг
	rs.wg.Add(1)
	go func() {
		defer rs.wg.Done()
		rs.monitoring.Start()
	}()

	// Запускаем периодическое переобучение моделей
	rs.wg.Add(1)
	go func() {
		defer rs.wg.Done()
		rs.periodicModelRetraining()
	}()

	// Запускаем интеграцию с Cadence
	rs.wg.Add(1)
	go func() {
		defer rs.wg.Done()
		rs.cadenceIntegrationLoop()
	}()

	// Запускаем сбор метрик для маршрутизации
	rs.wg.Add(1)
	go func() {
		defer rs.wg.Done()
		rs.pathFinder.StartMetricsCollection(rs.ctx)
	}()

	rs.wg.Add(1)
	go func() {
		defer rs.wg.Done()
		rs.latencyPredictor.StartMetricsCollection(rs.ctx)
	}()

	rs.logger.Info("Real AI service started successfully")
	return nil
}

// Stop останавливает AI сервис
func (rs *RealService) Stop() error {
	if !rs.started {
		return fmt.Errorf("service not started")
	}

	rs.logger.Info("Stopping Real AI service...")
	rs.started = false

	// Отменяем контекст
	rs.cancel()

	// Ждем завершения всех горутин
	rs.wg.Wait()

	rs.logger.Info("Real AI service stopped")
	return nil
}

// DetectAnomaly обнаруживает аномалии с использованием реальных данных
func (rs *RealService) DetectAnomaly(metrics *anomaly.NetworkMetrics) (*anomaly.AnomalyResult, error) {
	rs.logger.Debug("Detecting anomalies with real data")

	result, err := rs.realAnomalyDetector.DetectAnomaly(metrics)
	if err != nil {
		rs.metrics.IncrementAnomalyDetectionErrors()
		return nil, fmt.Errorf("failed to detect anomaly: %w", err)
	}

	if result.IsAnomaly {
		rs.metrics.IncrementAnomalyDetections()
		rs.logger.Warn("Anomaly detected",
			zap.Float64("score", result.Score),
			zap.String("type", result.AnomalyType),
			zap.Float64("confidence", result.Confidence))
	}

	return result, nil
}

// PredictLoad прогнозирует нагрузку с использованием реальных данных
func (rs *RealService) PredictLoad(horizon int) (*predictive.LoadPrediction, error) {
	rs.logger.Debug("Predicting load with real data", zap.Int("horizon", horizon))

	prediction, err := rs.realLoadPredictor.PredictLoad(horizon)
	if err != nil {
		rs.metrics.IncrementLoadPredictionErrors()
		return nil, fmt.Errorf("failed to predict load: %w", err)
	}

	rs.metrics.IncrementLoadPredictions()
	rs.logger.Info("Load prediction completed",
		zap.Int("horizon", horizon),
		zap.Float64("confidence", prediction.Confidence))

	return prediction, nil
}

// OptimizeRoute оптимизирует маршрут с использованием реальных данных
func (rs *RealService) OptimizeRoute(source, destination string, constraints map[string]interface{}) (*routing.OptimizedRoute, error) {
	rs.logger.Debug("Optimizing route",
		zap.String("source", source),
		zap.String("destination", destination),
		zap.Any("constraints", constraints))

	// Определяем протокол из ограничений
	protocol := "QUIC" // По умолчанию
	if p, ok := constraints["protocol"].(string); ok {
		protocol = p
	}

	// Находим оптимальный путь
	path, err := rs.pathFinder.FindOptimalPath(source, destination, protocol)
	if err != nil {
		rs.metrics.IncrementRouteOptimizationErrors()
		return nil, fmt.Errorf("failed to find optimal path: %w", err)
	}

	// Предсказываем латентность
	predictedLatency, err := rs.latencyPredictor.Predict(source, destination, protocol)
	if err != nil {
		rs.logger.Warn("Failed to predict latency, using default", zap.Error(err))
		predictedLatency = 50.0 // Значение по умолчанию
	}

	// Вычисляем уверенность на основе качества данных
	confidence := rs.calculateRouteConfidence(source, destination)

	route := &routing.OptimizedRoute{
		Protocol:         protocol,
		PredictedLatency: predictedLatency,
		Confidence:       confidence,
		Timestamp:        time.Now(),
		Path:             path,
	}

	rs.metrics.IncrementRouteOptimizations()
	rs.logger.Info("Route optimization completed",
		zap.String("protocol", route.Protocol),
		zap.Float64("latency", route.PredictedLatency),
		zap.Float64("confidence", route.Confidence),
		zap.Strings("path", route.Path))

	return route, nil
}

// calculateRouteConfidence вычисляет уверенность в маршруте на основе качества данных
func (rs *RealService) calculateRouteConfidence(source, destination string) float64 {
	confidence := 0.5 // Базовая уверенность

	// Проверяем наличие метрик для узлов
	if rs.pathFinder != nil {
		sourceMetrics, sourceExists := rs.pathFinder.GetMetricsCollector().GetNodeMetrics(source)
		destMetrics, destExists := rs.pathFinder.GetMetricsCollector().GetNodeMetrics(destination)

		if sourceExists && destExists {
			// Увеличиваем уверенность при наличии актуальных метрик
			confidence += 0.3

			// Дополнительная уверенность за актуальность данных
			now := time.Now()
			if now.Sub(sourceMetrics.LastUpdate) < 5*time.Minute {
				confidence += 0.1
			}
			if now.Sub(destMetrics.LastUpdate) < 5*time.Minute {
				confidence += 0.1
			}
		}
	}

	// Ограничиваем уверенность
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// GetStatus возвращает статус AI сервиса
func (rs *RealService) GetStatus() *ClientServiceStatus {
	status := &ClientServiceStatus{
		AnomalyDetector: rs.realAnomalyDetector != nil,
		LoadPredictor:   rs.realLoadPredictor != nil,
		RouteOptimizer:  rs.pathFinder != nil && rs.latencyPredictor != nil,
		Monitoring:      rs.monitoring != nil,
		DataPipeline:    rs.realDataCollector != nil,
	}

	// Дополнительная информация будет доступна через отдельные методы

	return status
}

// HealthCheck проверяет здоровье AI сервиса
func (rs *RealService) HealthCheck() error {
	if !rs.started {
		return fmt.Errorf("service not started")
	}

	// Проверяем компоненты
	if rs.realDataCollector == nil {
		return fmt.Errorf("data collector not initialized")
	}

	if rs.realAnomalyDetector == nil {
		return fmt.Errorf("anomaly detector not initialized")
	}

	if rs.realLoadPredictor == nil {
		return fmt.Errorf("load predictor not initialized")
	}

	if rs.pathFinder == nil {
		return fmt.Errorf("path finder not initialized")
	}

	if rs.latencyPredictor == nil {
		return fmt.Errorf("latency predictor not initialized")
	}

	// Проверяем Cadence
	if rs.cadenceIntegration != nil {
		if err := rs.cadenceIntegration.HealthCheck(rs.ctx); err != nil {
			rs.logger.Warn("Cadence health check failed", zap.Error(err))
		}
	}

	return nil
}

// periodicModelRetraining периодически переобучает модели
func (rs *RealService) periodicModelRetraining() {
	ticker := time.NewTicker(1 * time.Hour) // Переобучаем каждый час
	defer ticker.Stop()

	rs.logger.Info("Starting periodic model retraining")

	for {
		select {
		case <-rs.ctx.Done():
			rs.logger.Info("Stopping periodic model retraining")
			return
		case <-ticker.C:
			rs.logger.Info("Starting model retraining cycle")

			// Переобучаем предиктор нагрузки
			if rs.realLoadPredictor != nil {
				if err := rs.realLoadPredictor.RetrainModels(); err != nil {
					rs.logger.Error("Failed to retrain load predictor", zap.Error(err))
				} else {
					rs.logger.Info("Load predictor retrained successfully")
				}
			}

			// Переобучаем детектор аномалий
			if rs.realAnomalyDetector != nil {
				// Детектор аномалий переобучается автоматически при добавлении новых данных
				rs.logger.Info("Anomaly detector retrained automatically")
			}

			rs.logger.Info("Model retraining cycle completed")
		}
	}
}

// cadenceIntegrationLoop периодически синхронизируется с Cadence
func (rs *RealService) cadenceIntegrationLoop() {
	ticker := time.NewTicker(5 * time.Minute) // Синхронизируемся каждые 5 минут
	defer ticker.Stop()

	rs.logger.Info("Starting Cadence integration loop")

	for {
		select {
		case <-rs.ctx.Done():
			rs.logger.Info("Stopping Cadence integration loop")
			return
		case <-ticker.C:
			rs.logger.Debug("Syncing with Cadence")

			// Получаем статистику workflow
			if stats, err := rs.cadenceIntegration.GetWorkflowStats(rs.ctx); err != nil {
				rs.logger.Warn("Failed to get Cadence workflow stats", zap.Error(err))
			} else {
				rs.logger.Debug("Cadence workflow stats",
					zap.Int64("total", stats.TotalExecutions),
					zap.Int64("running", stats.RunningExecutions),
					zap.Float64("success_rate", stats.SuccessRate))
			}

			// Получаем метрики Cadence
			if metrics, err := rs.cadenceIntegration.GetMetrics(rs.ctx); err != nil {
				rs.logger.Warn("Failed to get Cadence metrics", zap.Error(err))
			} else {
				rs.logger.Debug("Cadence metrics retrieved", zap.Any("metrics", metrics))
			}
		}
	}
}

// GetRealDataInfo возвращает информацию о реальных данных
func (rs *RealService) GetRealDataInfo() map[string]interface{} {
	if rs.realDataCollector == nil {
		return map[string]interface{}{
			"error": "data collector not initialized",
		}
	}

	lastDataPoints := rs.realDataCollector.GetLastDataPoints()
	lastUpdate := rs.realDataCollector.GetLastUpdate()

	// Анализируем источники данных
	sources := make(map[string]int)
	for _, dp := range lastDataPoints {
		sources[dp.Source]++
	}

	return map[string]interface{}{
		"last_data_points": len(lastDataPoints),
		"last_update":      lastUpdate,
		"data_sources":     sources,
		"update_interval":  rs.realDataCollector.GetUpdateInterval().Seconds(),
		"sample_data":      rs.getSampleDataPoints(lastDataPoints),
	}
}

// getSampleDataPoints возвращает примеры точек данных
func (rs *RealService) getSampleDataPoints(dataPoints []pipeline.DataPoint) []map[string]interface{} {
	samples := make([]map[string]interface{}, 0, 5)

	for i, dp := range dataPoints {
		if i >= 5 { // Ограничиваем количество примеров
			break
		}

		sample := map[string]interface{}{
			"timestamp": dp.Timestamp,
			"source":    dp.Source,
			"type":      dp.Type,
			"data":      dp.Data,
		}
		samples = append(samples, sample)
	}

	return samples
}

// StartAnomalyDetectionWorkflow запускает workflow обнаружения аномалий в Cadence
func (rs *RealService) StartAnomalyDetectionWorkflow(input map[string]interface{}) (string, error) {
	if rs.cadenceIntegration == nil {
		return "", fmt.Errorf("cadence integration not initialized")
	}

	workflowID, err := rs.cadenceIntegration.StartAnomalyDetectionWorkflow(rs.ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to start anomaly detection workflow: %w", err)
	}

	rs.logger.Info("Anomaly detection workflow started", zap.String("workflow_id", workflowID))
	return workflowID, nil
}

// StartLoadPredictionWorkflow запускает workflow прогнозирования нагрузки в Cadence
func (rs *RealService) StartLoadPredictionWorkflow(horizon int, model string) (string, error) {
	if rs.cadenceIntegration == nil {
		return "", fmt.Errorf("cadence integration not initialized")
	}

	workflowID, err := rs.cadenceIntegration.StartLoadPredictionWorkflow(rs.ctx, horizon, model)
	if err != nil {
		return "", fmt.Errorf("failed to start load prediction workflow: %w", err)
	}

	rs.logger.Info("Load prediction workflow started", zap.String("workflow_id", workflowID))
	return workflowID, nil
}

// GetWorkflowExecution получает информацию о выполнении workflow
func (rs *RealService) GetWorkflowExecution(workflowID string) (*cadence.WorkflowExecution, error) {
	if rs.cadenceIntegration == nil {
		return nil, fmt.Errorf("cadence integration not initialized")
	}

	execution, err := rs.cadenceIntegration.GetWorkflowExecution(rs.ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow execution: %w", err)
	}

	return execution, nil
}
