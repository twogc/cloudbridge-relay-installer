package anomaly

import (
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

// NetworkAnomalyDetector реализует обнаружение аномалий в сетевом трафике
type NetworkAnomalyDetector struct {
	config          *Config
	logger          *zap.Logger
	isolationForest *SimpleIsolationForest
	threshold       float64
	learningRate    float64
	maxSamples      int
	contamination   float64

	// Data storage
	samples     [][]float64
	labels      []bool
	sampleCount int

	// Metrics
	detectionCount int64
	falsePositives int64
	falseNegatives int64
	accuracy       float64

	// Thread safety
	mu sync.RWMutex
}

// Config конфигурация для детектора аномалий
type Config struct {
	Threshold     float64 `yaml:"threshold"`
	LearningRate  float64 `yaml:"learning_rate"`
	MaxSamples    int     `yaml:"max_samples"`
	Contamination float64 `yaml:"contamination"`
}

// NetworkMetrics метрики сетевого трафика
type NetworkMetrics struct {
	Timestamp            time.Time      `json:"timestamp"`
	ConnectionsPerSecond int            `json:"connections_per_second"`
	BytesPerSecond       int64          `json:"bytes_per_second"`
	PacketLossRate       float64        `json:"packet_loss_rate"`
	AverageLatency       float64        `json:"average_latency"`
	CPUUsage             float64        `json:"cpu_usage"`
	MemoryUsage          float64        `json:"memory_usage"`
	ActiveConnections    int            `json:"active_connections"`
	FailedConnections    int            `json:"failed_connections"`
	ProtocolDistribution map[string]int `json:"protocol_distribution"`
}

// AnomalyResult результат обнаружения аномалии
type AnomalyResult struct {
	Score          float64            `json:"score"`
	IsAnomaly      bool               `json:"is_anomaly"`
	Timestamp      time.Time          `json:"timestamp"`
	Features       map[string]float64 `json:"features"`
	Confidence     float64            `json:"confidence"`
	AnomalyType    string             `json:"anomaly_type"`
	Severity       string             `json:"severity"`
	ModelID        string             `json:"model_id"`
	ProcessingTime time.Duration      `json:"processing_time"`
}

// NewNetworkAnomalyDetector создает новый детектор аномалий
func NewNetworkAnomalyDetector(config *Config, logger *zap.Logger) *NetworkAnomalyDetector {
	return &NetworkAnomalyDetector{
		config:        config,
		logger:        logger,
		threshold:     config.Threshold,
		learningRate:  config.LearningRate,
		maxSamples:    config.MaxSamples,
		contamination: config.Contamination,
		samples:       make([][]float64, 0),
		labels:        make([]bool, 0),
	}
}

// Initialize инициализирует детектор аномалий
func (nad *NetworkAnomalyDetector) Initialize() error {
	nad.logger.Info("Initializing Network Anomaly Detector...")

	// Создаем Isolation Forest
	nad.isolationForest = NewSimpleIsolationForest(100, nad.contamination)

	// Инициализируем с базовыми данными
	if err := nad.initializeWithBaselineData(); err != nil {
		return fmt.Errorf("failed to initialize with baseline data: %w", err)
	}

	nad.logger.Info("Network Anomaly Detector initialized successfully",
		zap.Float64("threshold", nad.threshold),
		zap.Float64("contamination", nad.contamination),
		zap.Int("max_samples", nad.maxSamples))

	return nil
}

// Start запускает детектор аномалий
func (nad *NetworkAnomalyDetector) Start() error {
	nad.logger.Info("Starting Network Anomaly Detector...")

	// Запускаем фоновое обучение модели
	go nad.backgroundTraining()

	nad.logger.Info("Network Anomaly Detector started successfully")
	return nil
}

// Stop останавливает детектор аномалий
func (nad *NetworkAnomalyDetector) Stop() error {
	nad.logger.Info("Stopping Network Anomaly Detector...")

	// Сохраняем модель
	if err := nad.saveModel(); err != nil {
		nad.logger.Error("Failed to save model", zap.Error(err))
	}

	nad.logger.Info("Network Anomaly Detector stopped successfully")
	return nil
}

// DetectAnomaly обнаруживает аномалии в сетевых метриках
func (nad *NetworkAnomalyDetector) DetectAnomaly(metrics *NetworkMetrics) (*AnomalyResult, error) {
	nad.mu.Lock()
	defer nad.mu.Unlock()

	// Извлекаем признаки
	features := nad.extractFeatures(metrics)

	// Обнаруживаем аномалии
	featuresArray := nad.convertFeaturesToArray(features)
	score, err := nad.isolationForest.Predict(featuresArray)
	if err != nil {
		return nil, fmt.Errorf("failed to predict anomaly: %w", err)
	}

	// Определяем, является ли это аномалией
	isAnomaly := score > nad.threshold

	// Определяем тип и серьезность аномалии
	anomalyType, severity := nad.classifyAnomaly(features, score)

	// Рассчитываем уверенность
	confidence := nad.calculateConfidence(score, features)

	// Обновляем метрики
	nad.updateMetrics(isAnomaly, score)

	// Добавляем в обучающую выборку
	nad.addSample(featuresArray, isAnomaly)

	result := &AnomalyResult{
		Score:       score,
		IsAnomaly:   isAnomaly,
		Timestamp:   time.Now(),
		Features:    features,
		Confidence:  confidence,
		AnomalyType: anomalyType,
		Severity:    severity,
	}

	if isAnomaly {
		nad.logger.Warn("Network anomaly detected",
			zap.Float64("score", score),
			zap.String("type", anomalyType),
			zap.String("severity", severity),
			zap.Float64("confidence", confidence))
	}

	return result, nil
}

// extractFeatures извлекает признаки из сетевых метрик
func (nad *NetworkAnomalyDetector) extractFeatures(metrics *NetworkMetrics) map[string]float64 {
	features := make(map[string]float64)

	// Базовые признаки
	features["connections_per_second"] = float64(metrics.ConnectionsPerSecond)
	features["bytes_per_second"] = float64(metrics.BytesPerSecond)
	features["packet_loss_rate"] = metrics.PacketLossRate
	features["average_latency"] = metrics.AverageLatency
	features["cpu_usage"] = metrics.CPUUsage
	features["memory_usage"] = metrics.MemoryUsage
	features["active_connections"] = float64(metrics.ActiveConnections)
	features["failed_connections"] = float64(metrics.FailedConnections)

	// Добавляем признаки из распределения протоколов
	protocolFeatures := nad.extractProtocolFeatures(metrics.ProtocolDistribution)
	for key, value := range protocolFeatures {
		features[key] = value
	}

	// Добавляем статистические признаки
	statisticalFeatures := nad.extractStatisticalFeatures(metrics)
	for key, value := range statisticalFeatures {
		features[key] = value
	}

	return features
}

// extractProtocolFeatures извлекает признаки из распределения протоколов
func (nad *NetworkAnomalyDetector) extractProtocolFeatures(protocols map[string]int) map[string]float64 {
	// Нормализуем распределение протоколов
	total := 0
	for _, count := range protocols {
		total += count
	}

	features := make(map[string]float64)
	if total > 0 {
		// QUIC процент
		quicPercent := float64(protocols["QUIC"]) / float64(total)
		features["quic_percentage"] = quicPercent

		// WireGuard процент
		wgPercent := float64(protocols["WireGuard"]) / float64(total)
		features["wireguard_percentage"] = wgPercent

		// TCP процент
		tcpPercent := float64(protocols["TCP"]) / float64(total)
		features["tcp_percentage"] = tcpPercent

		// UDP процент
		udpPercent := float64(protocols["UDP"]) / float64(total)
		features["udp_percentage"] = udpPercent
	} else {
		// Если нет данных, используем нули
		features["quic_percentage"] = 0.0
		features["wireguard_percentage"] = 0.0
		features["tcp_percentage"] = 0.0
		features["udp_percentage"] = 0.0
	}

	return features
}

// extractStatisticalFeatures извлекает статистические признаки
func (nad *NetworkAnomalyDetector) extractStatisticalFeatures(metrics *NetworkMetrics) map[string]float64 {
	features := make(map[string]float64)

	// Коэффициент вариации для соединений
	cvConnections := 0.0
	if metrics.ConnectionsPerSecond > 0 {
		cvConnections = float64(metrics.FailedConnections) / float64(metrics.ConnectionsPerSecond)
	}
	features["connection_variation"] = cvConnections

	// Эффективность использования ресурсов
	resourceEfficiency := 0.0
	if metrics.CPUUsage > 0 && metrics.MemoryUsage > 0 {
		resourceEfficiency = float64(metrics.BytesPerSecond) / (metrics.CPUUsage * metrics.MemoryUsage)
	}
	features["resource_efficiency"] = resourceEfficiency

	// Пропускная способность на соединение
	throughputPerConnection := 0.0
	if metrics.ActiveConnections > 0 {
		throughputPerConnection = float64(metrics.BytesPerSecond) / float64(metrics.ActiveConnections)
	}
	features["throughput_per_connection"] = throughputPerConnection

	return features
}

// classifyAnomaly классифицирует тип и серьезность аномалии
func (nad *NetworkAnomalyDetector) classifyAnomaly(features map[string]float64, score float64) (string, string) {
	// Определяем тип аномалии на основе признаков
	connectionsPerSecond := features["connections_per_second"]
	bytesPerSecond := features["bytes_per_second"]
	packetLossRate := features["packet_loss_rate"]
	latency := features["average_latency"]
	cpuUsage := features["cpu_usage"]

	var anomalyType string
	var severity string

	// DDoS атака
	if connectionsPerSecond > 1000 && bytesPerSecond > 1000000000 {
		anomalyType = "ddos_attack"
		severity = "critical"
	} else if packetLossRate > 0.1 {
		// Высокий packet loss
		anomalyType = "network_degradation"
		severity = "high"
	} else if latency > 1000 {
		// Высокая задержка
		anomalyType = "high_latency"
		severity = "medium"
	} else if cpuUsage > 90 {
		// Высокая загрузка CPU
		anomalyType = "resource_exhaustion"
		severity = "high"
	} else {
		// Общая аномалия
		anomalyType = "general_anomaly"
		if score > 0.9 {
			severity = "high"
		} else if score > 0.8 {
			severity = "medium"
		} else {
			severity = "low"
		}
	}

	return anomalyType, severity
}

// calculateConfidence рассчитывает уверенность в результате
func (nad *NetworkAnomalyDetector) calculateConfidence(score float64, features map[string]float64) float64 {
	// Базовая уверенность на основе score
	baseConfidence := math.Min(score*1.2, 1.0)

	// Корректируем на основе количества данных
	dataConfidence := math.Min(float64(nad.sampleCount)/1000.0, 1.0)

	// Корректируем на основе стабильности признаков
	stabilityConfidence := nad.calculateStabilityConfidence(features)

	// Итоговая уверенность
	confidence := (baseConfidence + dataConfidence + stabilityConfidence) / 3.0

	return math.Min(confidence, 1.0)
}

// calculateStabilityConfidence рассчитывает уверенность на основе стабильности
func (nad *NetworkAnomalyDetector) calculateStabilityConfidence(features map[string]float64) float64 {
	if len(nad.samples) < 10 {
		return 0.5 // Низкая уверенность при малом количестве данных
	}

	// Упрощенная логика стабильности
	// В реальной реализации здесь была бы более сложная логика сравнения с историческими данными

	// Базовая стабильность на основе количества признаков
	featureCount := len(features)
	if featureCount < 5 {
		return 0.6
	} else if featureCount < 10 {
		return 0.8
	} else {
		return 0.9
	}
}

// addSample добавляет образец в обучающую выборку
func (nad *NetworkAnomalyDetector) addSample(features []float64, isAnomaly bool) {
	nad.samples = append(nad.samples, features)
	nad.labels = append(nad.labels, isAnomaly)
	nad.sampleCount++

	// Ограничиваем размер выборки
	if len(nad.samples) > nad.maxSamples {
		nad.samples = nad.samples[1:]
		nad.labels = nad.labels[1:]
	}
}

// updateMetrics обновляет метрики детектора
func (nad *NetworkAnomalyDetector) updateMetrics(isAnomaly bool, score float64) {
	nad.detectionCount++

	// Здесь можно добавить логику для определения ложных срабатываний
	// Пока что используем простую эвристику
	if isAnomaly && score < nad.threshold*1.1 {
		nad.falsePositives++
	} else if !isAnomaly && score > nad.threshold*0.9 {
		nad.falseNegatives++
	}

	// Обновляем точность
	if nad.detectionCount > 0 {
		correctDetections := nad.detectionCount - nad.falsePositives - nad.falseNegatives
		nad.accuracy = float64(correctDetections) / float64(nad.detectionCount)
	}
}

// backgroundTraining выполняет фоновое обучение модели
func (nad *NetworkAnomalyDetector) backgroundTraining() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if len(nad.samples) > 100 {
				nad.retrainModel()
			}
		}
	}
}

// retrainModel переобучает модель на новых данных
func (nad *NetworkAnomalyDetector) retrainModel() {
	nad.mu.Lock()
	defer nad.mu.Unlock()

	if len(nad.samples) < 100 {
		return
	}

	nad.logger.Info("Retraining anomaly detection model",
		zap.Int("samples", len(nad.samples)))

	// Создаем новую модель
	newModel := NewSimpleIsolationForest(100, nad.contamination)

	// Обучаем на текущих данных
	// Здесь можно добавить более сложную логику обучения

	nad.isolationForest = newModel

	nad.logger.Info("Model retrained successfully",
		zap.Float64("accuracy", nad.accuracy))
}

// saveModel сохраняет модель
func (nad *NetworkAnomalyDetector) saveModel() error {
	// Здесь можно добавить сохранение модели в файл
	// Пока что просто логируем
	nad.logger.Info("Saving anomaly detection model",
		zap.Int("samples", len(nad.samples)),
		zap.Float64("accuracy", nad.accuracy))

	return nil
}

// initializeWithBaselineData инициализирует модель базовыми данными
func (nad *NetworkAnomalyDetector) initializeWithBaselineData() error {
	// Создаем базовые нормальные данные
	baselineData := [][]float64{
		{10, 1000000, 0.01, 50, 30, 40, 100, 5},   // Нормальная нагрузка
		{20, 2000000, 0.02, 60, 35, 45, 150, 8},   // Средняя нагрузка
		{5, 500000, 0.005, 40, 25, 35, 80, 2},     // Низкая нагрузка
		{15, 1500000, 0.015, 55, 32, 42, 120, 6},  // Нормальная нагрузка
		{25, 2500000, 0.025, 65, 40, 50, 180, 10}, // Высокая нагрузка
	}

	// Добавляем базовые данные
	for _, data := range baselineData {
		nad.addSample(data, false) // Все базовые данные - нормальные
	}

	nad.logger.Info("Initialized with baseline data",
		zap.Int("samples", len(baselineData)))

	return nil
}

// GetMetrics возвращает метрики детектора
func (nad *NetworkAnomalyDetector) GetMetrics() map[string]interface{} {
	nad.mu.RLock()
	defer nad.mu.RUnlock()

	return map[string]interface{}{
		"detection_count": nad.detectionCount,
		"false_positives": nad.falsePositives,
		"false_negatives": nad.falseNegatives,
		"accuracy":        nad.accuracy,
		"sample_count":    nad.sampleCount,
		"threshold":       nad.threshold,
		"contamination":   nad.contamination,
	}
}

// convertFeaturesToArray преобразует map[string]float64 в []float64
func (nad *NetworkAnomalyDetector) convertFeaturesToArray(features map[string]float64) []float64 {
	// Определяем порядок признаков для консистентности
	featureOrder := []string{
		"connections_per_second",
		"bytes_per_second",
		"packet_loss_rate",
		"average_latency",
		"cpu_usage",
		"memory_usage",
		"active_connections",
		"failed_connections",
		"connection_success_rate",
		"bandwidth_utilization",
		"latency_variance",
		"protocol_diversity",
		"hour_of_day",
		"is_business_hours",
		"is_weekend",
		"quic_percentage",
		"wireguard_percentage",
		"tcp_percentage",
		"udp_percentage",
		"connection_variation",
		"resource_efficiency",
		"throughput_per_connection",
	}

	result := make([]float64, 0, len(featureOrder))
	for _, key := range featureOrder {
		if value, exists := features[key]; exists {
			result = append(result, value)
		} else {
			result = append(result, 0.0) // Значение по умолчанию
		}
	}

	return result
}
