package anomaly

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/ai/pipeline"
	"go.uber.org/zap"
)

// RealAnomalyDetector использует реальные данные для обнаружения аномалий
type RealAnomalyDetector struct {
	logger          *zap.Logger
	dataCollector   *pipeline.RealDataCollector
	isolationForest *RealIsolationForest
	threshold       float64
	learningRate    float64
	maxSamples      int
	contamination   float64

	// Данные для обучения
	samples     [][]float64
	labels      []bool
	sampleCount int

	// Метрики
	detectionCount int64
	falsePositives int64
	falseNegatives int64
	accuracy       float64

	// Базовые значения для сравнения
	baselineValues map[string]BaselineStats

	// Thread safety
	mu sync.RWMutex
}

// BaselineStats базовые статистики для метрик
type BaselineStats struct {
	Mean         float64
	StdDev       float64
	Min          float64
	Max          float64
	Percentile95 float64
	Percentile99 float64
	LastUpdate   time.Time
	SampleCount  int
}

// RealIsolationForest улучшенная реализация Isolation Forest
type RealIsolationForest struct {
	trees        []*RealIsolationTree
	threshold    float64
	featureCount int
	maxSamples   int
}

// RealIsolationTree улучшенное дерево изоляции
type RealIsolationTree struct {
	root        *RealNode
	heightLimit int
	sampleSize  int
}

// RealNode улучшенный узел дерева
type RealNode struct {
	left        *RealNode
	right       *RealNode
	feature     int
	threshold   float64
	isLeaf      bool
	height      int
	sampleCount int
}

// NewRealAnomalyDetector создает новый детектор аномалий на основе реальных данных
func NewRealAnomalyDetector(logger *zap.Logger, dataCollector *pipeline.RealDataCollector) *RealAnomalyDetector {
	return &RealAnomalyDetector{
		logger:         logger,
		dataCollector:  dataCollector,
		threshold:      0.5,
		learningRate:   0.01,
		maxSamples:     1000,
		contamination:  0.1,
		baselineValues: make(map[string]BaselineStats),
	}
}

// Initialize инициализирует детектор аномалий
func (rad *RealAnomalyDetector) Initialize() error {
	rad.logger.Info("Initializing Real Anomaly Detector...")

	// Создаем Isolation Forest
	rad.isolationForest = NewRealIsolationForest(100, rad.contamination, 10)

	// Собираем базовые данные для обучения
	if err := rad.collectBaselineData(); err != nil {
		rad.logger.Warn("Failed to collect baseline data", zap.Error(err))
	}

	// Инициализируем базовые значения
	rad.initializeBaselineValues()

	rad.logger.Info("Real Anomaly Detector initialized successfully")
	return nil
}

// DetectAnomaly обнаруживает аномалии в реальных данных
func (rad *RealAnomalyDetector) DetectAnomaly(metrics *NetworkMetrics) (*AnomalyResult, error) {
	rad.mu.Lock()
	defer rad.mu.Unlock()

	// Собираем текущие данные
	currentData, err := rad.collectCurrentData()
	if err != nil {
		return nil, fmt.Errorf("failed to collect current data: %w", err)
	}

	// Извлекаем признаки
	features := rad.extractFeatures(currentData)

	// Проверяем аномалии с помощью Isolation Forest
	anomalyScore, err := rad.isolationForest.Predict(features)
	if err != nil {
		return nil, fmt.Errorf("failed to predict anomaly: %w", err)
	}

	// Проверяем аномалии с помощью статистических методов
	statisticalScore := rad.detectStatisticalAnomalies(features)

	// Проверяем аномалии с помощью базовых значений
	baselineScore := rad.detectBaselineAnomalies(features)

	// Комбинируем результаты
	combinedScore := (anomalyScore + statisticalScore + baselineScore) / 3.0

	// Определяем, является ли это аномалией
	isAnomaly := combinedScore > rad.threshold

	// Определяем тип аномалии
	anomalyType := rad.determineAnomalyType(features, combinedScore)

	// Вычисляем уверенность
	confidence := rad.calculateConfidence(combinedScore, len(rad.samples))

	result := &AnomalyResult{
		IsAnomaly:   isAnomaly,
		Score:       combinedScore,
		Confidence:  confidence,
		AnomalyType: anomalyType,
		Timestamp:   time.Now(),
	}

	// Обновляем метрики
	if isAnomaly {
		rad.detectionCount++
		rad.logger.Warn("Anomaly detected",
			zap.Float64("score", combinedScore),
			zap.String("type", anomalyType),
			zap.Float64("confidence", confidence))
	}

	// Добавляем в обучающие данные
	rad.addToTrainingData(features, isAnomaly)

	// Периодически переобучаем модель
	if rad.sampleCount%100 == 0 {
		go rad.retrainModel()
	}

	return result, nil
}

// collectCurrentData собирает текущие данные
func (rad *RealAnomalyDetector) collectCurrentData() (map[string]interface{}, error) {
	dataPoints, err := rad.dataCollector.Collect()
	if err != nil {
		return nil, err
	}

	currentData := make(map[string]interface{})

	for _, dp := range dataPoints {
		switch dp.Source {
		case "prometheus":
			if metricName, ok := dp.Data["metric_name"].(string); ok {
				if value, ok := dp.Data["value"].(float64); ok {
					currentData[metricName] = value
				}
			}
		case "cloudbridge-relay":
			if dp.Type == "stats" {
				for key, value := range dp.Data {
					currentData["relay_"+key] = value
				}
			}
		case "system":
			if dp.Type == "system_metrics" {
				for key, value := range dp.Data {
					currentData["system_"+key] = value
				}
			}
		}
	}

	return currentData, nil
}

// extractFeatures извлекает признаки из данных
func (rad *RealAnomalyDetector) extractFeatures(data map[string]interface{}) []float64 {
	features := make([]float64, 0, 20)

	// Извлекаем числовые признаки
	featureKeys := []string{
		"relay_active_connections",
		"go_goroutines",
		"go_memstats_alloc_bytes",
		"process_cpu_seconds_total",
		"process_resident_memory_bytes",
		"system_cpu_percent",
		"system_memory_used_percent",
		"relay_bandwidth",
		"relay_latency",
		"relay_response_time",
		"relay_error_rate",
	}

	for _, key := range featureKeys {
		if value, ok := data[key]; ok {
			if floatValue, ok := value.(float64); ok {
				features = append(features, floatValue)
			} else if intValue, ok := value.(int); ok {
				features = append(features, float64(intValue))
			} else {
				features = append(features, 0.0)
			}
		} else {
			features = append(features, 0.0)
		}
	}

	// Добавляем производные признаки
	features = append(features, rad.calculateDerivedFeatures(data)...)

	return features
}

// calculateDerivedFeatures вычисляет производные признаки
func (rad *RealAnomalyDetector) calculateDerivedFeatures(data map[string]interface{}) []float64 {
	derived := make([]float64, 0, 5)

	// Отношение CPU к памяти
	if cpu, ok := data["system_cpu_percent"].(float64); ok {
		if mem, ok := data["system_memory_used_percent"].(float64); ok {
			if mem > 0 {
				derived = append(derived, cpu/mem)
			} else {
				derived = append(derived, 0.0)
			}
		}
	}

	// Отношение подключений к латентности
	if connections, ok := data["relay_active_connections"].(float64); ok {
		if latency, ok := data["relay_latency"].(float64); ok {
			if latency > 0 {
				derived = append(derived, connections/latency)
			} else {
				derived = append(derived, 0.0)
			}
		}
	}

	// Отношение пропускной способности к ошибкам
	if bandwidth, ok := data["relay_bandwidth"].(float64); ok {
		if errorRate, ok := data["relay_error_rate"].(float64); ok {
			if errorRate > 0 {
				derived = append(derived, bandwidth/errorRate)
			} else {
				derived = append(derived, bandwidth)
			}
		}
	}

	// Время суток (циклический признак)
	hour := float64(time.Now().Hour())
	derived = append(derived, math.Sin(2*math.Pi*hour/24.0))
	derived = append(derived, math.Cos(2*math.Pi*hour/24.0))

	return derived
}

// detectStatisticalAnomalies обнаруживает аномалии с помощью статистических методов
func (rad *RealAnomalyDetector) detectStatisticalAnomalies(features []float64) float64 {
	if len(rad.samples) < 10 {
		return 0.5 // Нейтральный скор при недостатке данных
	}

	anomalyScore := 0.0
	featureCount := 0

	for i, feature := range features {
		if i >= len(rad.samples[0]) {
			break
		}

		// Вычисляем статистики для признака
		values := make([]float64, 0, len(rad.samples))
		for _, sample := range rad.samples {
			if i < len(sample) {
				values = append(values, sample[i])
			}
		}

		if len(values) < 2 {
			continue
		}

		mean := rad.calculateMean(values)
		stdDev := rad.calculateStdDev(values, mean)

		if stdDev > 0 {
			// Z-score
			zScore := math.Abs(feature-mean) / stdDev

			// Преобразуем Z-score в скор аномалии (0-1)
			featureScore := math.Min(1.0, zScore/3.0) // 3 сигмы = 99.7%
			anomalyScore += featureScore
			featureCount++
		}
	}

	if featureCount > 0 {
		return anomalyScore / float64(featureCount)
	}

	return 0.5
}

// detectBaselineAnomalies обнаруживает аномалии с помощью базовых значений
func (rad *RealAnomalyDetector) detectBaselineAnomalies(features []float64) float64 {
	anomalyScore := 0.0
	featureCount := 0

	featureKeys := []string{
		"relay_active_connections",
		"go_goroutines",
		"go_memstats_alloc_bytes",
		"process_cpu_seconds_total",
		"process_resident_memory_bytes",
		"system_cpu_percent",
		"system_memory_used_percent",
		"relay_bandwidth",
		"relay_latency",
		"relay_response_time",
		"relay_error_rate",
	}

	for i, feature := range features {
		if i >= len(featureKeys) {
			break
		}

		key := featureKeys[i]
		if baseline, ok := rad.baselineValues[key]; ok {
			// Проверяем отклонение от базовых значений
			if baseline.StdDev > 0 {
				zScore := math.Abs(feature-baseline.Mean) / baseline.StdDev
				featureScore := math.Min(1.0, zScore/2.0) // 2 сигмы = 95%
				anomalyScore += featureScore
				featureCount++
			}
		}
	}

	if featureCount > 0 {
		return anomalyScore / float64(featureCount)
	}

	return 0.5
}

// determineAnomalyType определяет тип аномалии
func (rad *RealAnomalyDetector) determineAnomalyType(features []float64, score float64) string {
	if score < 0.3 {
		return "normal"
	} else if score < 0.6 {
		return "minor_anomaly"
	} else if score < 0.8 {
		return "moderate_anomaly"
	} else {
		return "major_anomaly"
	}
}

// calculateConfidence вычисляет уверенность в результате
func (rad *RealAnomalyDetector) calculateConfidence(score float64, sampleCount int) float64 {
	// Уверенность растет с количеством образцов
	dataConfidence := math.Min(1.0, float64(sampleCount)/100.0)

	// Уверенность выше для экстремальных значений
	scoreConfidence := math.Abs(score-0.5) * 2.0

	// Общая уверенность
	confidence := (dataConfidence + scoreConfidence) / 2.0

	return math.Min(1.0, confidence)
}

// addToTrainingData добавляет данные в обучающую выборку
func (rad *RealAnomalyDetector) addToTrainingData(features []float64, isAnomaly bool) {
	rad.samples = append(rad.samples, features)
	rad.labels = append(rad.labels, isAnomaly)
	rad.sampleCount++

	// Ограничиваем размер выборки
	if len(rad.samples) > rad.maxSamples {
		rad.samples = rad.samples[1:]
		rad.labels = rad.labels[1:]
	}

	// Обновляем базовые значения
	rad.updateBaselineValues(features)
}

// updateBaselineValues обновляет базовые значения
func (rad *RealAnomalyDetector) updateBaselineValues(features []float64) {
	featureKeys := []string{
		"relay_active_connections",
		"go_goroutines",
		"go_memstats_alloc_bytes",
		"process_cpu_seconds_total",
		"process_resident_memory_bytes",
		"system_cpu_percent",
		"system_memory_used_percent",
		"relay_bandwidth",
		"relay_latency",
		"relay_response_time",
		"relay_error_rate",
	}

	for i, feature := range features {
		if i >= len(featureKeys) {
			break
		}

		key := featureKeys[i]
		baseline := rad.baselineValues[key]

		// Обновляем статистики
		baseline.SampleCount++
		oldMean := baseline.Mean
		baseline.Mean = oldMean + (feature-oldMean)/float64(baseline.SampleCount)

		// Обновляем стандартное отклонение (упрощенная формула)
		if baseline.SampleCount > 1 {
			baseline.StdDev = math.Sqrt((baseline.StdDev*baseline.StdDev + (feature-oldMean)*(feature-baseline.Mean)) / float64(baseline.SampleCount-1))
		}

		// Обновляем min/max
		if baseline.SampleCount == 1 || feature < baseline.Min {
			baseline.Min = feature
		}
		if baseline.SampleCount == 1 || feature > baseline.Max {
			baseline.Max = feature
		}

		baseline.LastUpdate = time.Now()
		rad.baselineValues[key] = baseline
	}
}

// collectBaselineData собирает базовые данные для обучения
func (rad *RealAnomalyDetector) collectBaselineData() error {
	rad.logger.Info("Collecting baseline data for anomaly detection...")

	// Собираем данные в течение некоторого времени
	for i := 0; i < 10; i++ {
		data, err := rad.collectCurrentData()
		if err != nil {
			rad.logger.Warn("Failed to collect baseline data", zap.Int("iteration", i), zap.Error(err))
			continue
		}

		features := rad.extractFeatures(data)
		rad.addToTrainingData(features, false) // Считаем базовые данные нормальными

		time.Sleep(5 * time.Second)
	}

	rad.logger.Info("Baseline data collection completed", zap.Int("samples", len(rad.samples)))
	return nil
}

// initializeBaselineValues инициализирует базовые значения
func (rad *RealAnomalyDetector) initializeBaselineValues() {
	featureKeys := []string{
		"relay_active_connections",
		"go_goroutines",
		"go_memstats_alloc_bytes",
		"process_cpu_seconds_total",
		"process_resident_memory_bytes",
		"system_cpu_percent",
		"system_memory_used_percent",
		"relay_bandwidth",
		"relay_latency",
		"relay_response_time",
		"relay_error_rate",
	}

	for _, key := range featureKeys {
		rad.baselineValues[key] = BaselineStats{
			Mean:         0.0,
			StdDev:       1.0,
			Min:          math.Inf(1),
			Max:          math.Inf(-1),
			Percentile95: 0.0,
			Percentile99: 0.0,
			LastUpdate:   time.Now(),
			SampleCount:  0,
		}
	}
}

// retrainModel переобучает модель
func (rad *RealAnomalyDetector) retrainModel() {
	rad.mu.Lock()
	defer rad.mu.Unlock()

	if len(rad.samples) < 10 {
		return
	}

	rad.logger.Info("Retraining anomaly detection model...")

	// Пересоздаем Isolation Forest
	rad.isolationForest = NewRealIsolationForest(100, rad.contamination, len(rad.samples[0]))

	// Обучаем на текущих данных
	rad.isolationForest.Fit(rad.samples)

	rad.logger.Info("Model retraining completed", zap.Int("samples", len(rad.samples)))
}

// calculateMean вычисляет среднее значение
func (rad *RealAnomalyDetector) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}

	return sum / float64(len(values))
}

// calculateStdDev вычисляет стандартное отклонение
func (rad *RealAnomalyDetector) calculateStdDev(values []float64, mean float64) float64 {
	if len(values) < 2 {
		return 1.0
	}

	sumSquaredDiff := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquaredDiff += diff * diff
	}

	return math.Sqrt(sumSquaredDiff / float64(len(values)-1))
}

// GetStats возвращает статистику детектора
func (rad *RealAnomalyDetector) GetStats() map[string]interface{} {
	rad.mu.RLock()
	defer rad.mu.RUnlock()

	return map[string]interface{}{
		"detection_count": rad.detectionCount,
		"false_positives": rad.falsePositives,
		"false_negatives": rad.falseNegatives,
		"accuracy":        rad.accuracy,
		"sample_count":    rad.sampleCount,
		"threshold":       rad.threshold,
		"baseline_count":  len(rad.baselineValues),
	}
}
