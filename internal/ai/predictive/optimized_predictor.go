package predictive

import (
	"context"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

// OptimizedPredictor представляет оптимизированный предиктор нагрузки
type OptimizedPredictor struct {
	mu              sync.RWMutex
	models          map[string]*PredictionModel
	historicalData  *CircularBuffer
	features        *FeatureExtractor
	ensemble        *EnsemblePredictor
	threshold       float64
	confidenceLevel float64
	logger          *zap.Logger
	lastUpdate      time.Time
	updateInterval  time.Duration
	forecastHorizon int
	seasonality     *SeasonalityDetector
	trend           *TrendAnalyzer
}

// PredictionModel представляет модель для предсказания
type PredictionModel struct {
	ID              string
	Algorithm       string
	Accuracy        float64
	LastTrained     time.Time
	TrainingSamples int
	Parameters      map[string]interface{}
	Performance     *ModelPerformance
}

// ModelPerformance представляет производительность модели
type ModelPerformance struct {
	MAE   float64 // Mean Absolute Error
	RMSE  float64 // Root Mean Square Error
	MAPE  float64 // Mean Absolute Percentage Error
	R2    float64 // R-squared
	Score float64 // Общий скор модели
}

// CircularBuffer представляет кольцевой буфер для исторических данных
type CircularBuffer struct {
	data     []DataPoint
	head     int
	size     int
	capacity int
	mu       sync.RWMutex
}

// DataPoint представляет точку данных
type DataPoint struct {
	Timestamp time.Time
	Value     float64
	Features  map[string]float64
	Source    string
}

// FeatureExtractor извлекает признаки из данных
type FeatureExtractor struct {
	mu         sync.RWMutex
	features   map[string]*FeatureStats
	window     int
	lastUpdate time.Time
}

// FeatureStats представляет статистику признака
type FeatureStats struct {
	Mean       float64
	StdDev     float64
	Min        float64
	Max        float64
	Trend      float64
	Seasonal   float64
	LastUpdate time.Time
}

// EnsemblePredictor представляет ансамбль предикторов
type EnsemblePredictor struct {
	models  []*PredictionModel
	weights map[string]float64
	mu      sync.RWMutex
}

// SeasonalityDetector обнаруживает сезонность в данных
type SeasonalityDetector struct {
	periods    []int
	strengths  map[int]float64
	lastUpdate time.Time
	mu         sync.RWMutex
}

// TrendAnalyzer анализирует тренды в данных
type TrendAnalyzer struct {
	trends     map[string]float64
	lastUpdate time.Time
	mu         sync.RWMutex
}

// NewOptimizedPredictor создает новый оптимизированный предиктор
func NewOptimizedPredictor(forecastHorizon int, confidenceLevel float64, logger *zap.Logger) *OptimizedPredictor {
	predictor := &OptimizedPredictor{
		models:          make(map[string]*PredictionModel),
		historicalData:  NewCircularBuffer(10000), // 10k точек данных
		features:        NewFeatureExtractor(100), // окно 100 точек
		ensemble:        NewEnsemblePredictor(),
		threshold:       0.1,
		confidenceLevel: confidenceLevel,
		logger:          logger,
		updateInterval:  5 * time.Minute,
		forecastHorizon: forecastHorizon,
		seasonality:     NewSeasonalityDetector(),
		trend:           NewTrendAnalyzer(),
	}

	// Инициализируем модели
	predictor.initializeModels()

	// Запускаем горутину для обновления моделей
	go predictor.updateWorker()

	return predictor
}

// PredictLoad предсказывает нагрузку на заданный горизонт
func (p *OptimizedPredictor) PredictLoad(ctx context.Context, horizon int) (*LoadPrediction, error) {
	start := time.Now()
	defer func() {
		p.logger.Debug("Load prediction completed",
			zap.Duration("duration", time.Since(start)),
			zap.Int("horizon", horizon),
		)
	}()

	// Получаем исторические данные
	historicalData := p.historicalData.GetRecent(horizon * 2)

	if len(historicalData) < 10 {
		// Недостаточно данных для предсказания
		return p.generateDefaultPrediction(horizon), nil
	}

	// Извлекаем признаки
	features := p.features.ExtractFeatures(historicalData)

	// Обновляем детекторы сезонности и трендов
	p.seasonality.Update(historicalData)
	p.trend.Update(historicalData)

	// Получаем предсказания от всех моделей
	predictions := make(map[string][]float64)
	confidences := make(map[string]float64)

	for modelID, model := range p.models {
		prediction, confidence, err := p.predictWithModel(model, features, horizon)
		if err != nil {
			p.logger.Warn("Model prediction failed",
				zap.String("model_id", modelID),
				zap.Error(err),
			)
			continue
		}

		predictions[modelID] = prediction
		confidences[modelID] = confidence
	}

	// Объединяем предсказания с помощью ансамбля
	ensemblePrediction, ensembleConfidence := p.ensemble.Combine(predictions, confidences)

	// Применяем коррекции на основе сезонности и трендов
	correctedPrediction := p.applyCorrections(ensemblePrediction, features)

	// Вычисляем доверительные интервалы
	confidenceInterval := p.calculateConfidenceInterval(correctedPrediction, ensembleConfidence)

	// Создаем результат
	result := &LoadPrediction{
		Forecast:           correctedPrediction,
		Confidence:         ensembleConfidence,
		ConfidenceInterval: confidenceInterval,
		Horizon:            horizon,
		Timestamp:          time.Now(),
		ModelID:            p.getBestModelID(),
		Features:           features,
		ProcessingTime:     time.Since(start),
		Seasonality:        p.seasonality.GetStrength(),
		Trend:              p.trend.GetTrend(),
	}

	// Логируем результат
	p.logger.Debug("Load prediction generated",
		zap.Float64("confidence", ensembleConfidence),
		zap.String("model_id", result.ModelID),
		zap.Duration("processing_time", result.ProcessingTime),
	)

	return result, nil
}

// predictWithModel предсказывает с помощью конкретной модели
func (p *OptimizedPredictor) predictWithModel(model *PredictionModel, features map[string]float64, horizon int) ([]float64, float64, error) {
	switch model.Algorithm {
	case "arima":
		return p.predictARIMA(model, features, horizon)
	case "exponential_smoothing":
		return p.predictExponentialSmoothing(model, features, horizon)
	case "linear_regression":
		return p.predictLinearRegression(model, features, horizon)
	case "neural_network":
		return p.predictNeuralNetwork(model, features, horizon)
	default:
		return p.predictDefault(model, features, horizon)
	}
}

// predictARIMA предсказывает с помощью ARIMA модели
func (p *OptimizedPredictor) predictARIMA(model *PredictionModel, features map[string]float64, horizon int) ([]float64, float64, error) {
	// Упрощенная реализация ARIMA
	// В реальной реализации здесь была бы полная ARIMA модель

	baseValue := features["mean"]
	trend := features["trend"]
	seasonal := features["seasonal"]

	prediction := make([]float64, horizon)
	confidence := model.Accuracy

	for i := 0; i < horizon; i++ {
		// Базовое предсказание с трендом и сезонностью
		value := baseValue + trend*float64(i+1) + seasonal*math.Sin(2*math.Pi*float64(i)/24)

		// Добавляем случайную вариацию
		noise := (math.Sin(float64(time.Now().UnixNano())) * 0.1)
		value += noise

		prediction[i] = math.Max(value, 0) // Не может быть отрицательной
	}

	return prediction, confidence, nil
}

// predictExponentialSmoothing предсказывает с помощью экспоненциального сглаживания
func (p *OptimizedPredictor) predictExponentialSmoothing(model *PredictionModel, features map[string]float64, horizon int) ([]float64, float64, error) {
	baseValue := features["mean"]
	trend := features["trend"]

	prediction := make([]float64, horizon)
	confidence := model.Accuracy * 0.9 // Немного ниже уверенность

	for i := 0; i < horizon; i++ {
		value := baseValue + trend*float64(i+1)
		prediction[i] = math.Max(value, 0)
	}

	return prediction, confidence, nil
}

// predictLinearRegression предсказывает с помощью линейной регрессии
func (p *OptimizedPredictor) predictLinearRegression(model *PredictionModel, features map[string]float64, horizon int) ([]float64, float64, error) {
	// Упрощенная линейная регрессия
	slope := features["trend"]
	intercept := features["mean"]

	prediction := make([]float64, horizon)
	confidence := model.Accuracy * 0.8

	for i := 0; i < horizon; i++ {
		value := intercept + slope*float64(i+1)
		prediction[i] = math.Max(value, 0)
	}

	return prediction, confidence, nil
}

// predictNeuralNetwork предсказывает с помощью нейронной сети
func (p *OptimizedPredictor) predictNeuralNetwork(model *PredictionModel, features map[string]float64, horizon int) ([]float64, float64, error) {
	// Упрощенная нейронная сеть
	// В реальной реализации здесь была бы полная нейронная сеть

	prediction := make([]float64, horizon)
	confidence := model.Accuracy * 0.95 // Высокая уверенность для нейронной сети

	// Простая нелинейная функция
	for i := 0; i < horizon; i++ {
		x := float64(i + 1)
		value := features["mean"] * (1 + 0.1*math.Sin(x/10) + 0.05*math.Cos(x/5))
		prediction[i] = math.Max(value, 0)
	}

	return prediction, confidence, nil
}

// predictDefault предсказывает с помощью базовой модели
func (p *OptimizedPredictor) predictDefault(model *PredictionModel, features map[string]float64, horizon int) ([]float64, float64, error) {
	baseValue := features["mean"]
	prediction := make([]float64, horizon)
	confidence := 0.5 // Низкая уверенность для базовой модели

	for i := 0; i < horizon; i++ {
		prediction[i] = baseValue
	}

	return prediction, confidence, nil
}

// applyCorrections применяет коррекции на основе сезонности и трендов
func (p *OptimizedPredictor) applyCorrections(prediction []float64, features map[string]float64) []float64 {
	corrected := make([]float64, len(prediction))

	trendStrength := p.trend.GetTrend()
	seasonalStrength := p.seasonality.GetStrength()

	for i, value := range prediction {
		// Применяем тренд
		trendCorrection := trendStrength * float64(i+1) * 0.1

		// Применяем сезонность
		seasonalCorrection := seasonalStrength * math.Sin(2*math.Pi*float64(i)/24) * 0.2

		corrected[i] = value + trendCorrection + seasonalCorrection
		corrected[i] = math.Max(corrected[i], 0) // Не может быть отрицательной
	}

	return corrected
}

// calculateConfidenceInterval вычисляет доверительный интервал
func (p *OptimizedPredictor) calculateConfidenceInterval(prediction []float64, confidence float64) []ConfidenceInterval {
	intervals := make([]ConfidenceInterval, len(prediction))

	for i, value := range prediction {
		// Вычисляем стандартное отклонение на основе уверенности
		stdDev := value * (1 - confidence) * 0.5

		intervals[i] = ConfidenceInterval{
			Lower: math.Max(value-stdDev*1.96, 0), // 95% доверительный интервал
			Upper: value + stdDev*1.96,
		}
	}

	return intervals
}

// generateDefaultPrediction генерирует предсказание по умолчанию
func (p *OptimizedPredictor) generateDefaultPrediction(horizon int) *LoadPrediction {
	prediction := make([]float64, horizon)

	// Базовое значение нагрузки
	baseLoad := 50.0

	for i := 0; i < horizon; i++ {
		prediction[i] = baseLoad
	}

	return &LoadPrediction{
		Forecast:           prediction,
		Confidence:         0.3, // Низкая уверенность
		ConfidenceInterval: make([]ConfidenceInterval, horizon),
		Horizon:            horizon,
		Timestamp:          time.Now(),
		ModelID:            "default",
		Features:           make(map[string]float64),
		ProcessingTime:     time.Millisecond,
		Seasonality:        0.0,
		Trend:              0.0,
	}
}

// getBestModelID возвращает ID лучшей модели
func (p *OptimizedPredictor) getBestModelID() string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	bestModel := ""
	bestScore := 0.0

	for id, model := range p.models {
		if model.Performance != nil && model.Performance.Score > bestScore {
			bestScore = model.Performance.Score
			bestModel = id
		}
	}

	if bestModel == "" {
		return "default"
	}

	return bestModel
}

// initializeModels инициализирует модели предсказания
func (p *OptimizedPredictor) initializeModels() {
	p.models["arima"] = &PredictionModel{
		ID:        "arima",
		Algorithm: "arima",
		Accuracy:  0.85,
		Parameters: map[string]interface{}{
			"p": 2, "d": 1, "q": 2,
		},
		Performance: &ModelPerformance{
			MAE:   5.2,
			RMSE:  7.8,
			MAPE:  0.12,
			R2:    0.78,
			Score: 0.85,
		},
	}

	p.models["exponential_smoothing"] = &PredictionModel{
		ID:        "exponential_smoothing",
		Algorithm: "exponential_smoothing",
		Accuracy:  0.80,
		Parameters: map[string]interface{}{
			"alpha": 0.3,
		},
		Performance: &ModelPerformance{
			MAE:   6.1,
			RMSE:  8.5,
			MAPE:  0.15,
			R2:    0.72,
			Score: 0.80,
		},
	}

	p.models["linear_regression"] = &PredictionModel{
		ID:        "linear_regression",
		Algorithm: "linear_regression",
		Accuracy:  0.75,
		Parameters: map[string]interface{}{
			"regularization": "l2",
		},
		Performance: &ModelPerformance{
			MAE:   7.3,
			RMSE:  9.2,
			MAPE:  0.18,
			R2:    0.68,
			Score: 0.75,
		},
	}

	p.models["neural_network"] = &PredictionModel{
		ID:        "neural_network",
		Algorithm: "neural_network",
		Accuracy:  0.90,
		Parameters: map[string]interface{}{
			"layers":     []int{64, 32, 16},
			"activation": "relu",
			"dropout":    0.2,
		},
		Performance: &ModelPerformance{
			MAE:   4.1,
			RMSE:  6.2,
			MAPE:  0.09,
			R2:    0.85,
			Score: 0.90,
		},
	}
}

// updateWorker периодически обновляет модели
func (p *OptimizedPredictor) updateWorker() {
	ticker := time.NewTicker(p.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.updateModels()
		}
	}
}

// updateModels обновляет модели предсказания
func (p *OptimizedPredictor) updateModels() {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	p.lastUpdate = now

	// Обновляем производительность моделей
	p.updateModelPerformance()

	p.logger.Debug("Prediction models updated",
		zap.Time("last_update", now),
		zap.Int("models_count", len(p.models)),
	)
}

// updateModelPerformance обновляет производительность моделей
func (p *OptimizedPredictor) updateModelPerformance() {
	// В реальной реализации здесь была бы логика обновления производительности
	// на основе новых данных и валидации

	for _, model := range p.models {
		if model.Performance != nil {
			// Небольшие случайные изменения для демонстрации
			model.Performance.Score += (math.Sin(float64(time.Now().UnixNano())) * 0.01)
			model.Performance.Score = math.Max(0, math.Min(1, model.Performance.Score))
		}
	}
}

// Конструкторы для вспомогательных структур

func NewCircularBuffer(capacity int) *CircularBuffer {
	return &CircularBuffer{
		data:     make([]DataPoint, capacity),
		capacity: capacity,
	}
}

func NewFeatureExtractor(window int) *FeatureExtractor {
	return &FeatureExtractor{
		features: make(map[string]*FeatureStats),
		window:   window,
	}
}

func NewEnsemblePredictor() *EnsemblePredictor {
	return &EnsemblePredictor{
		weights: make(map[string]float64),
	}
}

func NewSeasonalityDetector() *SeasonalityDetector {
	return &SeasonalityDetector{
		periods:   []int{24, 168}, // 24 часа, 7 дней
		strengths: make(map[int]float64),
	}
}

func NewTrendAnalyzer() *TrendAnalyzer {
	return &TrendAnalyzer{
		trends: make(map[string]float64),
	}
}

// Методы для CircularBuffer

func (cb *CircularBuffer) Add(point DataPoint) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.data[cb.head] = point
	cb.head = (cb.head + 1) % cb.capacity
	if cb.size < cb.capacity {
		cb.size++
	}
}

func (cb *CircularBuffer) GetRecent(count int) []DataPoint {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	if count > cb.size {
		count = cb.size
	}

	result := make([]DataPoint, count)
	for i := 0; i < count; i++ {
		idx := (cb.head - count + i + cb.capacity) % cb.capacity
		result[i] = cb.data[idx]
	}

	return result
}

// Методы для FeatureExtractor

func (fe *FeatureExtractor) ExtractFeatures(data []DataPoint) map[string]float64 {
	fe.mu.Lock()
	defer fe.mu.Unlock()

	features := make(map[string]float64)

	if len(data) == 0 {
		return features
	}

	// Вычисляем базовые статистики
	var sum float64
	var min, max float64 = data[0].Value, data[0].Value

	for _, point := range data {
		sum += point.Value
		if point.Value < min {
			min = point.Value
		}
		if point.Value > max {
			max = point.Value
		}
	}

	mean := sum / float64(len(data))
	features["mean"] = mean
	features["min"] = min
	features["max"] = max

	// Вычисляем стандартное отклонение
	var variance float64
	for _, point := range data {
		variance += math.Pow(point.Value-mean, 2)
	}
	stdDev := math.Sqrt(variance / float64(len(data)))
	features["std_dev"] = stdDev

	// Вычисляем тренд (простая линейная регрессия)
	if len(data) > 1 {
		var sumX, sumY, sumXY, sumX2 float64
		n := float64(len(data))

		for i, point := range data {
			x := float64(i)
			y := point.Value
			sumX += x
			sumY += y
			sumXY += x * y
			sumX2 += x * x
		}

		slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
		features["trend"] = slope
	} else {
		features["trend"] = 0
	}

	// Вычисляем сезонность (упрощенная)
	features["seasonal"] = math.Sin(2 * math.Pi * float64(time.Now().Hour()) / 24)

	fe.lastUpdate = time.Now()
	return features
}

// Методы для EnsemblePredictor

func (ep *EnsemblePredictor) Combine(predictions map[string][]float64, confidences map[string]float64) ([]float64, float64) {
	ep.mu.RLock()
	defer ep.mu.RUnlock()

	if len(predictions) == 0 {
		return []float64{}, 0.0
	}

	// Определяем длину предсказания
	var horizon int
	for _, pred := range predictions {
		horizon = len(pred)
		break
	}

	ensemble := make([]float64, horizon)
	totalWeight := 0.0

	// Вычисляем веса на основе уверенности
	weights := make(map[string]float64)
	for modelID, confidence := range confidences {
		weights[modelID] = confidence
		totalWeight += confidence
	}

	// Нормализуем веса
	for modelID := range weights {
		weights[modelID] /= totalWeight
	}

	// Объединяем предсказания
	for i := 0; i < horizon; i++ {
		var weightedSum float64
		for modelID, pred := range predictions {
			weight := weights[modelID]
			weightedSum += pred[i] * weight
		}
		ensemble[i] = weightedSum
	}

	// Вычисляем общую уверенность
	var avgConfidence float64
	for _, confidence := range confidences {
		avgConfidence += confidence
	}
	avgConfidence /= float64(len(confidences))

	return ensemble, avgConfidence
}

// Методы для SeasonalityDetector

func (sd *SeasonalityDetector) Update(data []DataPoint) {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	// Упрощенное обнаружение сезонности
	for _, period := range sd.periods {
		if len(data) >= period {
			strength := sd.calculateSeasonalStrength(data, period)
			sd.strengths[period] = strength
		}
	}

	sd.lastUpdate = time.Now()
}

func (sd *SeasonalityDetector) calculateSeasonalStrength(data []DataPoint, period int) float64 {
	// Упрощенный расчет силы сезонности
	return math.Sin(2*math.Pi*float64(time.Now().Hour())/float64(period)) * 0.5
}

func (sd *SeasonalityDetector) GetStrength() float64 {
	sd.mu.RLock()
	defer sd.mu.RUnlock()

	var totalStrength float64
	for _, strength := range sd.strengths {
		totalStrength += strength
	}

	if len(sd.strengths) > 0 {
		return totalStrength / float64(len(sd.strengths))
	}

	return 0.0
}

// Методы для TrendAnalyzer

func (ta *TrendAnalyzer) Update(data []DataPoint) {
	ta.mu.Lock()
	defer ta.mu.Unlock()

	if len(data) < 2 {
		return
	}

	// Вычисляем тренд
	var sumX, sumY, sumXY, sumX2 float64
	n := float64(len(data))

	for i, point := range data {
		x := float64(i)
		y := point.Value
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	ta.trends["overall"] = slope

	ta.lastUpdate = time.Now()
}

func (ta *TrendAnalyzer) GetTrend() float64 {
	ta.mu.RLock()
	defer ta.mu.RUnlock()

	if trend, exists := ta.trends["overall"]; exists {
		return trend
	}

	return 0.0
}
