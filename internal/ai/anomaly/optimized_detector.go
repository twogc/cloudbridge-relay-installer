package anomaly

import (
	"context"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

// OptimizedDetector представляет оптимизированный детектор аномалий
type OptimizedDetector struct {
	mu                sync.RWMutex
	models            map[string]*AnomalyModel
	baseline          *BaselineModel
	threshold         float64
	learningRate      float64
	maxSamples        int
	contamination     float64
	logger            *zap.Logger
	lastUpdate        time.Time
	updateInterval    time.Duration
	featureWeights    map[string]float64
	adaptiveThreshold bool
}

// AnomalyModel представляет модель для детекции аномалий
type AnomalyModel struct {
	ID              string
	Algorithm       string
	Threshold       float64
	LastTrained     time.Time
	TrainingSamples int
	Accuracy        float64
	Features        []string
	Weights         map[string]float64
	Statistics      *FeatureStatistics
}

// BaselineModel представляет базовую модель для сравнения
type BaselineModel struct {
	Features    map[string]*FeatureBaseline
	LastUpdate  time.Time
	SampleCount int
}

// FeatureBaseline представляет базовые значения для признака
type FeatureBaseline struct {
	Mean        float64
	StdDev      float64
	Min         float64
	Max         float64
	Percentiles map[int]float64
	LastUpdate  time.Time
}

// FeatureStatistics представляет статистику признаков
type FeatureStatistics struct {
	Mean     float64
	StdDev   float64
	Min      float64
	Max      float64
	Skewness float64
	Kurtosis float64
}

// NewOptimizedDetector создает новый оптимизированный детектор аномалий
func NewOptimizedDetector(threshold float64, learningRate float64, maxSamples int, contamination float64, logger *zap.Logger) *OptimizedDetector {
	detector := &OptimizedDetector{
		models:            make(map[string]*AnomalyModel),
		threshold:         threshold,
		learningRate:      learningRate,
		maxSamples:        maxSamples,
		contamination:     contamination,
		logger:            logger,
		updateInterval:    5 * time.Minute,
		featureWeights:    make(map[string]float64),
		adaptiveThreshold: true,
	}

	// Инициализируем базовую модель
	detector.baseline = &BaselineModel{
		Features: make(map[string]*FeatureBaseline),
	}

	// Инициализируем веса признаков
	detector.initializeFeatureWeights()

	// Запускаем горутину для периодического обновления моделей
	go detector.updateWorker()

	return detector
}

// DetectAnomaly обнаруживает аномалии в сетевых метриках
func (d *OptimizedDetector) DetectAnomaly(ctx context.Context, metrics *NetworkMetrics) (*AnomalyResult, error) {
	start := time.Now()
	defer func() {
		d.logger.Debug("Anomaly detection completed",
			zap.Duration("duration", time.Since(start)),
			zap.String("remote_addr", metrics.Timestamp.Format(time.RFC3339)),
		)
	}()

	// Извлекаем признаки
	features := d.extractFeatures(metrics)

	// Вычисляем аномальный скор
	score := d.calculateAnomalyScore(features)

	// Определяем, является ли это аномалией
	isAnomaly := score > d.getAdaptiveThreshold()

	// Определяем тип аномалии
	anomalyType := d.classifyAnomaly(features, score)

	// Определяем серьезность
	severity := d.determineSeverity(score, anomalyType)

	// Вычисляем уверенность
	confidence := d.calculateConfidence(score, features)

	// Обновляем базовую модель
	d.updateBaseline(features)

	// Создаем результат
	result := &AnomalyResult{
		IsAnomaly:      isAnomaly,
		Score:          score,
		Confidence:     confidence,
		AnomalyType:    anomalyType,
		Severity:       severity,
		Timestamp:      time.Now(),
		Features:       features,
		ModelID:        d.getBestModelID(),
		ProcessingTime: time.Since(start),
	}

	// Логируем результат
	if isAnomaly {
		d.logger.Warn("Anomaly detected",
			zap.Float64("score", score),
			zap.String("type", anomalyType),
			zap.String("severity", severity),
			zap.Float64("confidence", confidence),
		)
	}

	return result, nil
}

// extractFeatures извлекает признаки из метрик
func (d *OptimizedDetector) extractFeatures(metrics *NetworkMetrics) map[string]float64 {
	features := make(map[string]float64)

	// Базовые сетевые признаки
	features["connections_per_second"] = float64(metrics.ConnectionsPerSecond)
	features["bytes_per_second"] = float64(metrics.BytesPerSecond)
	features["packet_loss_rate"] = metrics.PacketLossRate
	features["average_latency"] = metrics.AverageLatency
	features["cpu_usage"] = metrics.CPUUsage
	features["memory_usage"] = metrics.MemoryUsage
	features["active_connections"] = float64(metrics.ActiveConnections)
	features["failed_connections"] = float64(metrics.FailedConnections)

	// Вычисляемые признаки
	features["connection_success_rate"] = 1.0 - (float64(metrics.FailedConnections) / float64(metrics.ActiveConnections+1))
	features["bandwidth_utilization"] = d.calculateBandwidthUtilization(metrics)
	features["latency_variance"] = d.calculateLatencyVariance(metrics)
	features["protocol_diversity"] = d.calculateProtocolDiversity(metrics)

	// Временные признаки
	hour := float64(time.Now().Hour())
	features["hour_of_day"] = hour
	features["is_business_hours"] = d.isBusinessHours(hour)
	features["is_weekend"] = d.isWeekend()

	return features
}

// calculateAnomalyScore вычисляет аномальный скор
func (d *OptimizedDetector) calculateAnomalyScore(features map[string]float64) float64 {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var totalScore float64
	var totalWeight float64

	for feature, value := range features {
		weight, exists := d.featureWeights[feature]
		if !exists {
			weight = 1.0
		}

		// Получаем базовые значения для признака
		baseline, exists := d.baseline.Features[feature]
		if !exists {
			// Если нет базовых значений, используем простую эвристику
			score := d.calculateSimpleScore(feature, value)
			totalScore += score * weight
			totalWeight += weight
			continue
		}

		// Вычисляем отклонение от базовых значений
		deviation := d.calculateDeviation(value, baseline)
		score := d.normalizeScore(deviation)

		totalScore += score * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0.0
	}

	return totalScore / totalWeight
}

// calculateDeviation вычисляет отклонение от базовых значений
func (d *OptimizedDetector) calculateDeviation(value float64, baseline *FeatureBaseline) float64 {
	if baseline.StdDev == 0 {
		return 0.0
	}

	// Z-score
	zScore := (value - baseline.Mean) / baseline.StdDev

	// Возвращаем абсолютное значение Z-score
	return math.Abs(zScore)
}

// normalizeScore нормализует скор в диапазон [0, 1]
func (d *OptimizedDetector) normalizeScore(deviation float64) float64 {
	// Используем сигмоидальную функцию для нормализации
	return 1.0 / (1.0 + math.Exp(-deviation))
}

// calculateSimpleScore вычисляет простой скор для признака без базовых значений
func (d *OptimizedDetector) calculateSimpleScore(feature string, value float64) float64 {
	switch feature {
	case "packet_loss_rate":
		if value > 0.05 { // 5% потери пакетов
			return 0.8
		}
		return value * 10
	case "average_latency":
		if value > 200 { // 200ms латентность
			return 0.9
		}
		return value / 200
	case "cpu_usage":
		if value > 90 { // 90% CPU
			return 0.9
		}
		return value / 100
	case "memory_usage":
		if value > 90 { // 90% памяти
			return 0.9
		}
		return value / 100
	default:
		return 0.0
	}
}

// classifyAnomaly классифицирует тип аномалии
func (d *OptimizedDetector) classifyAnomaly(features map[string]float64, score float64) string {
	// Анализируем признаки для определения типа аномалии
	if features["packet_loss_rate"] > 0.1 {
		return "network_degradation"
	}
	if features["average_latency"] > 500 {
		return "high_latency"
	}
	if features["cpu_usage"] > 95 {
		return "resource_exhaustion"
	}
	if features["memory_usage"] > 95 {
		return "memory_pressure"
	}
	if features["failed_connections"] > float64(features["active_connections"])*0.1 {
		return "connection_failures"
	}
	if features["bytes_per_second"] > 1000000000 { // 1GB/s
		return "traffic_spike"
	}

	return "unknown"
}

// determineSeverity определяет серьезность аномалии
func (d *OptimizedDetector) determineSeverity(score float64, anomalyType string) string {
	if score > 0.9 {
		return "critical"
	}
	if score > 0.7 {
		return "high"
	}
	if score > 0.5 {
		return "medium"
	}
	return "low"
}

// calculateConfidence вычисляет уверенность в результате
func (d *OptimizedDetector) calculateConfidence(score float64, features map[string]float64) float64 {
	// Базовая уверенность на основе скора
	baseConfidence := math.Min(score*1.2, 1.0)

	// Корректировка на основе количества признаков
	featureCount := float64(len(features))
	featureConfidence := math.Min(featureCount/10.0, 1.0)

	// Корректировка на основе качества данных
	dataQuality := d.assessDataQuality(features)

	// Итоговая уверенность
	confidence := (baseConfidence + featureConfidence + dataQuality) / 3.0
	return math.Min(confidence, 1.0)
}

// assessDataQuality оценивает качество данных
func (d *OptimizedDetector) assessDataQuality(features map[string]float64) float64 {
	// Проверяем наличие всех ключевых признаков
	requiredFeatures := []string{"connections_per_second", "bytes_per_second", "packet_loss_rate", "average_latency"}
	missingFeatures := 0

	for _, feature := range requiredFeatures {
		if _, exists := features[feature]; !exists {
			missingFeatures++
		}
	}

	// Качество данных снижается при отсутствии признаков
	quality := 1.0 - (float64(missingFeatures) / float64(len(requiredFeatures)))
	return math.Max(quality, 0.0)
}

// updateBaseline обновляет базовую модель
func (d *OptimizedDetector) updateBaseline(features map[string]float64) {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	d.baseline.LastUpdate = now
	d.baseline.SampleCount++

	for feature, value := range features {
		baseline, exists := d.baseline.Features[feature]
		if !exists {
			baseline = &FeatureBaseline{
				Percentiles: make(map[int]float64),
			}
			d.baseline.Features[feature] = baseline
		}

		// Обновляем статистики (упрощенный алгоритм)
		if d.baseline.SampleCount == 1 {
			baseline.Mean = value
			baseline.Min = value
			baseline.Max = value
			baseline.StdDev = 0
		} else {
			// Экспоненциальное скользящее среднее
			alpha := d.learningRate
			baseline.Mean = alpha*value + (1-alpha)*baseline.Mean

			// Обновляем min/max
			if value < baseline.Min {
				baseline.Min = value
			}
			if value > baseline.Max {
				baseline.Max = value
			}
		}

		baseline.LastUpdate = now
	}
}

// getAdaptiveThreshold возвращает адаптивный порог
func (d *OptimizedDetector) getAdaptiveThreshold() float64 {
	if !d.adaptiveThreshold {
		return d.threshold
	}

	// Адаптивный порог на основе времени суток и нагрузки
	hour := time.Now().Hour()

	// Ночью порог выше (меньше ложных срабатываний)
	if hour >= 22 || hour <= 6 {
		return d.threshold * 1.2
	}

	// В рабочее время порог ниже (больше чувствительности)
	if hour >= 9 && hour <= 17 {
		return d.threshold * 0.8
	}

	return d.threshold
}

// getBestModelID возвращает ID лучшей модели
func (d *OptimizedDetector) getBestModelID() string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	bestModel := ""
	bestAccuracy := 0.0

	for id, model := range d.models {
		if model.Accuracy > bestAccuracy {
			bestAccuracy = model.Accuracy
			bestModel = id
		}
	}

	if bestModel == "" {
		return "baseline"
	}

	return bestModel
}

// initializeFeatureWeights инициализирует веса признаков
func (d *OptimizedDetector) initializeFeatureWeights() {
	// Веса основаны на важности признаков для детекции аномалий
	d.featureWeights = map[string]float64{
		"packet_loss_rate":        1.5,
		"average_latency":         1.3,
		"failed_connections":      1.2,
		"cpu_usage":               1.1,
		"memory_usage":            1.1,
		"connection_success_rate": 1.0,
		"bytes_per_second":        0.9,
		"connections_per_second":  0.8,
		"active_connections":      0.7,
		"bandwidth_utilization":   0.6,
		"latency_variance":        0.5,
		"protocol_diversity":      0.4,
		"hour_of_day":             0.3,
		"is_business_hours":       0.2,
		"is_weekend":              0.1,
	}
}

// updateWorker периодически обновляет модели
func (d *OptimizedDetector) updateWorker() {
	ticker := time.NewTicker(d.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			d.updateModels()
		}
	}
}

// updateModels обновляет модели детекции
func (d *OptimizedDetector) updateModels() {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	d.lastUpdate = now

	// Обновляем веса признаков на основе производительности
	d.updateFeatureWeights()

	d.logger.Debug("Anomaly detection models updated",
		zap.Time("last_update", now),
		zap.Int("baseline_samples", d.baseline.SampleCount),
	)
}

// updateFeatureWeights обновляет веса признаков
func (d *OptimizedDetector) updateFeatureWeights() {
	// Простая эвристика для обновления весов
	// В реальной реализации здесь была бы более сложная логика

	// Увеличиваем вес признаков, которые часто указывают на аномалии
	for feature, weight := range d.featureWeights {
		baseline, exists := d.baseline.Features[feature]
		if exists && baseline.StdDev > 0 {
			// Если признак имеет высокую вариативность, увеличиваем его вес
			if baseline.StdDev > baseline.Mean*0.5 {
				d.featureWeights[feature] = weight * 1.1
			}
		}
	}
}

// Вспомогательные методы

func (d *OptimizedDetector) calculateBandwidthUtilization(metrics *NetworkMetrics) float64 {
	// Упрощенный расчет утилизации полосы пропускания
	return float64(metrics.BytesPerSecond) / 1000000000.0 // Нормализация к 1GB/s
}

func (d *OptimizedDetector) calculateLatencyVariance(metrics *NetworkMetrics) float64 {
	// Упрощенный расчет вариации латентности
	return metrics.AverageLatency * 0.1 // 10% от средней латентности
}

func (d *OptimizedDetector) calculateProtocolDiversity(metrics *NetworkMetrics) float64 {
	// Расчет разнообразия протоколов
	protocolCount := len(metrics.ProtocolDistribution)
	return float64(protocolCount) / 10.0 // Нормализация к 10 протоколам
}

func (d *OptimizedDetector) isBusinessHours(hour float64) float64 {
	if hour >= 9 && hour <= 17 {
		return 1.0
	}
	return 0.0
}

func (d *OptimizedDetector) isWeekend() float64 {
	weekday := time.Now().Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return 1.0
	}
	return 0.0
}
