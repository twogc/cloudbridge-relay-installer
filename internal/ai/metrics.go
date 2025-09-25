package ai

import (
	"sync"
	"time"
)

// AIMetrics represents AI-specific metrics
type AIMetrics struct {
	mu sync.RWMutex

	// Anomaly detection metrics
	anomalyDetections     int64
	anomalyFalsePositives int64
	anomalyFalseNegatives int64
	anomalyAccuracy       float64

	// Load prediction metrics
	loadPredictions        int64
	loadPredictionErrors   int64
	loadPredictionAccuracy float64

	// Route optimization metrics
	routeOptimizations      int64
	routeOptimizationGains  []float64
	averageOptimizationGain float64

	// Processing metrics
	processingTime        time.Duration
	totalProcessingTime   time.Duration
	processingCount       int64
	averageProcessingTime time.Duration

	// System metrics
	memoryUsage       int64
	cpuUsage          float64
	activeConnections int64
}

// NewAIMetrics creates a new AI metrics instance
func NewAIMetrics() *AIMetrics {
	return &AIMetrics{
		routeOptimizationGains: make([]float64, 0),
	}
}

// RecordAnomalyDetection records an anomaly detection
func (m *AIMetrics) RecordAnomalyDetection(isAnomaly bool, isCorrect bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.anomalyDetections++

	if isAnomaly && !isCorrect {
		m.anomalyFalsePositives++
	} else if !isAnomaly && !isCorrect {
		m.anomalyFalseNegatives++
	}

	// Calculate accuracy
	if m.anomalyDetections > 0 {
		correctDetections := m.anomalyDetections - m.anomalyFalsePositives - m.anomalyFalseNegatives
		m.anomalyAccuracy = float64(correctDetections) / float64(m.anomalyDetections)
	}
}

// RecordAnomalyDetections records multiple anomaly detections
func (m *AIMetrics) RecordAnomalyDetections(count int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.anomalyDetections += int64(count)
}

// RecordLoadPrediction records a load prediction
func (m *AIMetrics) RecordLoadPrediction(actual float64, predicted float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.loadPredictions++

	// Calculate prediction error
	error := abs(actual-predicted) / actual
	if error > 0.1 { // 10% error threshold
		m.loadPredictionErrors++
	}

	// Calculate accuracy
	if m.loadPredictions > 0 {
		correctPredictions := m.loadPredictions - m.loadPredictionErrors
		m.loadPredictionAccuracy = float64(correctPredictions) / float64(m.loadPredictions)
	}
}

// RecordLoadPredictions records multiple load predictions
func (m *AIMetrics) RecordLoadPredictions(count int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.loadPredictions += int64(count)
}

// RecordRouteOptimization records a route optimization
func (m *AIMetrics) RecordRouteOptimization(gain float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.routeOptimizations++
	m.routeOptimizationGains = append(m.routeOptimizationGains, gain)

	// Keep only last 1000 gains for average calculation
	if len(m.routeOptimizationGains) > 1000 {
		m.routeOptimizationGains = m.routeOptimizationGains[1:]
	}

	// Calculate average gain
	totalGain := 0.0
	for _, g := range m.routeOptimizationGains {
		totalGain += g
	}
	m.averageOptimizationGain = totalGain / float64(len(m.routeOptimizationGains))
}

// RecordRouteOptimizations records multiple route optimizations
func (m *AIMetrics) RecordRouteOptimizations(count int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.routeOptimizations += int64(count)
}

// RecordProcessingTime records processing time
func (m *AIMetrics) RecordProcessingTime(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.processingTime = duration
	m.totalProcessingTime += duration
	m.processingCount++
	m.averageProcessingTime = m.totalProcessingTime / time.Duration(m.processingCount)
}

// UpdateSystemMetrics updates system metrics
func (m *AIMetrics) UpdateSystemMetrics(memoryUsage int64, cpuUsage float64, activeConnections int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.memoryUsage = memoryUsage
	m.cpuUsage = cpuUsage
	m.activeConnections = activeConnections
}

// GetMetrics returns all metrics
func (m *AIMetrics) GetMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"anomaly_detection": map[string]interface{}{
			"total_detections": m.anomalyDetections,
			"false_positives":  m.anomalyFalsePositives,
			"false_negatives":  m.anomalyFalseNegatives,
			"accuracy":         m.anomalyAccuracy,
		},
		"load_prediction": map[string]interface{}{
			"total_predictions": m.loadPredictions,
			"prediction_errors": m.loadPredictionErrors,
			"accuracy":          m.loadPredictionAccuracy,
		},
		"route_optimization": map[string]interface{}{
			"total_optimizations": m.routeOptimizations,
			"average_gain":        m.averageOptimizationGain,
		},
		"processing": map[string]interface{}{
			"current_time": m.processingTime.Milliseconds(),
			"average_time": m.averageProcessingTime.Milliseconds(),
			"total_count":  m.processingCount,
		},
		"system": map[string]interface{}{
			"memory_usage":       m.memoryUsage,
			"cpu_usage":          m.cpuUsage,
			"active_connections": m.activeConnections,
		},
	}
}

// GetAnomalyMetrics returns anomaly detection metrics
func (m *AIMetrics) GetAnomalyMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"total_detections": m.anomalyDetections,
		"false_positives":  m.anomalyFalsePositives,
		"false_negatives":  m.anomalyFalseNegatives,
		"accuracy":         m.anomalyAccuracy,
	}
}

// GetLoadPredictionMetrics returns load prediction metrics
func (m *AIMetrics) GetLoadPredictionMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"total_predictions": m.loadPredictions,
		"prediction_errors": m.loadPredictionErrors,
		"accuracy":          m.loadPredictionAccuracy,
	}
}

// GetRouteOptimizationMetrics returns route optimization metrics
func (m *AIMetrics) GetRouteOptimizationMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"total_optimizations": m.routeOptimizations,
		"average_gain":        m.averageOptimizationGain,
	}
}

// GetProcessingMetrics returns processing metrics
func (m *AIMetrics) GetProcessingMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"current_time": m.processingTime.Milliseconds(),
		"average_time": m.averageProcessingTime.Milliseconds(),
		"total_count":  m.processingCount,
	}
}

// GetSystemMetrics returns system metrics
func (m *AIMetrics) GetSystemMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"memory_usage":       m.memoryUsage,
		"cpu_usage":          m.cpuUsage,
		"active_connections": m.activeConnections,
	}
}

// Reset resets all metrics
// IncrementAnomalyDetections увеличивает счетчик обнаружений аномалий
func (m *AIMetrics) IncrementAnomalyDetections() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.anomalyDetections++
}

// IncrementAnomalyDetectionErrors увеличивает счетчик ошибок обнаружения аномалий
func (m *AIMetrics) IncrementAnomalyDetectionErrors() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.anomalyFalsePositives++
}

// IncrementLoadPredictions увеличивает счетчик прогнозов нагрузки
func (m *AIMetrics) IncrementLoadPredictions() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.loadPredictions++
}

// IncrementLoadPredictionErrors увеличивает счетчик ошибок прогнозирования нагрузки
func (m *AIMetrics) IncrementLoadPredictionErrors() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.loadPredictionErrors++
}

// IncrementRouteOptimizations увеличивает счетчик оптимизаций маршрутов
func (m *AIMetrics) IncrementRouteOptimizations() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.routeOptimizations++
}

// IncrementRouteOptimizationErrors увеличивает счетчик ошибок оптимизации маршрутов
func (m *AIMetrics) IncrementRouteOptimizationErrors() {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Можно добавить отдельное поле для ошибок оптимизации маршрутов
	// Пока используем существующую логику
}

func (m *AIMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.anomalyDetections = 0
	m.anomalyFalsePositives = 0
	m.anomalyFalseNegatives = 0
	m.anomalyAccuracy = 0

	m.loadPredictions = 0
	m.loadPredictionErrors = 0
	m.loadPredictionAccuracy = 0

	m.routeOptimizations = 0
	m.routeOptimizationGains = make([]float64, 0)
	m.averageOptimizationGain = 0

	m.processingTime = 0
	m.totalProcessingTime = 0
	m.processingCount = 0
	m.averageProcessingTime = 0

	m.memoryUsage = 0
	m.cpuUsage = 0
	m.activeConnections = 0
}

// abs returns absolute value of float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
