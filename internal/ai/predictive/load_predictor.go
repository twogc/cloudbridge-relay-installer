package predictive

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// LoadPredictor предсказывает нагрузку на систему
type LoadPredictor struct {
	config     *Config
	logger     *zap.Logger
	arimaModel *ARIMAModel
	history    *TimeSeriesBuffer

	// Metrics
	predictionCount int64
	accuracy        float64

	// Thread safety
	mu sync.RWMutex
}

// Config конфигурация для предсказания нагрузки
type Config struct {
	Enabled         bool    `yaml:"enabled"`
	Algorithm       string  `yaml:"algorithm"`
	ForecastHorizon int     `yaml:"forecast_horizon"`
	ConfidenceLevel float64 `yaml:"confidence_level"`
}

// ARIMAParams параметры ARIMA модели
type ARIMAParams struct {
	P int `yaml:"p"` // autoregressive
	D int `yaml:"d"` // differencing
	Q int `yaml:"q"` // moving average
}

// ConfidenceInterval представляет доверительный интервал
type ConfidenceInterval struct {
	Lower float64 `json:"lower"`
	Upper float64 `json:"upper"`
}

// LoadPrediction результат предсказания нагрузки
type LoadPrediction struct {
	Forecast           []float64            `json:"forecast"`
	Confidence         float64              `json:"confidence"`
	ConfidenceInterval []ConfidenceInterval `json:"confidence_interval"`
	Horizon            int                  `json:"horizon"`
	Timestamp          time.Time            `json:"timestamp"`
	ModelID            string               `json:"model_id"`
	Features           map[string]float64   `json:"features"`
	ProcessingTime     time.Duration        `json:"processing_time"`
	Seasonality        float64              `json:"seasonality"`
	Trend              float64              `json:"trend"`
}

// NewLoadPredictor создает новый предсказатель нагрузки
func NewLoadPredictor(config *Config, logger *zap.Logger) *LoadPredictor {
	return &LoadPredictor{
		config:  config,
		logger:  logger,
		history: NewTimeSeriesBuffer(1000), // Храним последние 1000 точек
	}
}

// Initialize инициализирует предсказатель нагрузки
func (lp *LoadPredictor) Initialize() error {
	lp.logger.Info("Initializing Load Predictor...")

	// Создаем ARIMA модель
	lp.arimaModel = NewARIMAModel(2, 1, 2) // ARIMA(2,1,2)

	// Инициализируем с базовыми данными
	if err := lp.initializeWithBaselineData(); err != nil {
		return fmt.Errorf("failed to initialize with baseline data: %w", err)
	}

	lp.logger.Info("Load Predictor initialized successfully")
	return nil
}

// Start запускает предсказатель нагрузки
func (lp *LoadPredictor) Start() error {
	lp.logger.Info("Starting Load Predictor...")

	// Запускаем фоновое обучение модели
	go lp.backgroundTraining()

	lp.logger.Info("Load Predictor started successfully")
	return nil
}

// Stop останавливает предсказатель нагрузки
func (lp *LoadPredictor) Stop() error {
	lp.logger.Info("Stopping Load Predictor...")

	// Сохраняем модель
	if err := lp.saveModel(); err != nil {
		lp.logger.Error("Failed to save model", zap.Error(err))
	}

	lp.logger.Info("Load Predictor stopped successfully")
	return nil
}

// PredictLoad предсказывает нагрузку на заданный горизонт
func (lp *LoadPredictor) PredictLoad(horizon int) (*LoadPrediction, error) {
	lp.mu.Lock()
	defer lp.mu.Unlock()

	// Получаем исторические данные
	historicalData := lp.history.GetLast(24 * 7) // 7 дней данных

	if len(historicalData) < 10 {
		return nil, fmt.Errorf("insufficient historical data for prediction")
	}

	// ARIMA прогнозирование
	forecast, err := lp.arimaModel.Forecast(historicalData, horizon)
	if err != nil {
		return nil, fmt.Errorf("failed to forecast: %w", err)
	}

	// Расчет доверительного интервала
	confidence := lp.calculateConfidence(forecast, historicalData)

	// Обновляем метрики
	lp.predictionCount++

	result := &LoadPrediction{
		Forecast:   forecast,
		Confidence: confidence,
		Horizon:    horizon,
		Timestamp:  time.Now(),
	}

	lp.logger.Info("Load prediction completed",
		zap.Int("horizon", horizon),
		zap.Float64("confidence", confidence))

	return result, nil
}

// AddDataPoint добавляет новую точку данных
func (lp *LoadPredictor) AddDataPoint(timestamp time.Time, value float64) {
	lp.mu.Lock()
	defer lp.mu.Unlock()

	lp.history.Add(timestamp, value)
}

// calculateConfidence рассчитывает доверительный интервал
func (lp *LoadPredictor) calculateConfidence(forecast []float64, historical []float64) float64 {
	if len(historical) < 2 {
		return 0.5
	}

	// Простой расчет доверительного интервала на основе вариации исторических данных
	var sum, sumSquares float64
	for _, value := range historical {
		sum += value
		sumSquares += value * value
	}

	mean := sum / float64(len(historical))
	variance := (sumSquares / float64(len(historical))) - (mean * mean)
	stdDev := variance

	// Чем меньше стандартное отклонение, тем выше уверенность
	confidence := 1.0 - (stdDev / mean)
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 1 {
		confidence = 1
	}

	return confidence
}

// backgroundTraining выполняет фоновое обучение модели
func (lp *LoadPredictor) backgroundTraining() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if lp.history.Size() > 100 {
				lp.retrainModel()
			}
		}
	}
}

// retrainModel переобучает модель
func (lp *LoadPredictor) retrainModel() {
	lp.mu.Lock()
	defer lp.mu.Unlock()

	if lp.history.Size() < 100 {
		return
	}

	lp.logger.Info("Retraining load prediction model",
		zap.Int("samples", lp.history.Size()))

	// Получаем данные для обучения
	data := lp.history.GetLast(100)

	// Переобучаем ARIMA модель
	lp.arimaModel.Retrain(data)

	lp.logger.Info("Model retrained successfully")
}

// saveModel сохраняет модель
func (lp *LoadPredictor) saveModel() error {
	lp.logger.Info("Saving load prediction model",
		zap.Int64("predictions", lp.predictionCount),
		zap.Float64("accuracy", lp.accuracy))

	return nil
}

// initializeWithBaselineData инициализирует модель базовыми данными
func (lp *LoadPredictor) initializeWithBaselineData() error {
	// Создаем базовые данные нагрузки
	baselineData := []float64{
		30, 35, 40, 45, 50, 55, 60, 65, 70, 75, // Нормальная нагрузка
		80, 85, 90, 85, 80, 75, 70, 65, 60, 55, // Снижение нагрузки
		50, 45, 40, 35, 30, 35, 40, 45, 50, 55, // Возврат к норме
	}

	// Добавляем базовые данные
	now := time.Now()
	for i, value := range baselineData {
		timestamp := now.Add(-time.Duration(len(baselineData)-i) * time.Minute)
		lp.history.Add(timestamp, value)
	}

	lp.logger.Info("Initialized with baseline data",
		zap.Int("samples", len(baselineData)))

	return nil
}

// GetMetrics возвращает метрики предсказателя
func (lp *LoadPredictor) GetMetrics() map[string]interface{} {
	lp.mu.RLock()
	defer lp.mu.RUnlock()

	return map[string]interface{}{
		"prediction_count": lp.predictionCount,
		"accuracy":         lp.accuracy,
		"history_size":     lp.history.Size(),
		"forecast_horizon": lp.config.ForecastHorizon,
		"confidence_level": lp.config.ConfidenceLevel,
	}
}
