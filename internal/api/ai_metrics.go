package api

import (
	"sync"
	"time"
)

// AIMetrics собирает метрики AI обработки
type AIMetrics struct {
	requestsProcessed int64
	anomaliesDetected int64
	predictionsMade   int64
	routesOptimized   int64
	processingTime    time.Duration
	cacheHits         int64
	cacheMisses       int64
	errors            int64
	mu                sync.RWMutex
}

// NewAIMetrics создает новые AI метрики
func NewAIMetrics() *AIMetrics {
	return &AIMetrics{}
}

// IncrementRequestsProcessed увеличивает счетчик обработанных запросов
func (m *AIMetrics) IncrementRequestsProcessed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requestsProcessed++
}

// IncrementAnomaliesDetected увеличивает счетчик обнаруженных аномалий
func (m *AIMetrics) IncrementAnomaliesDetected() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.anomaliesDetected++
}

// IncrementPredictionsMade увеличивает счетчик сделанных предсказаний
func (m *AIMetrics) IncrementPredictionsMade() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.predictionsMade++
}

// IncrementRoutesOptimized увеличивает счетчик оптимизированных маршрутов
func (m *AIMetrics) IncrementRoutesOptimized() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.routesOptimized++
}

// IncrementCacheHits увеличивает счетчик попаданий в кэш
func (m *AIMetrics) IncrementCacheHits() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cacheHits++
}

// IncrementCacheMisses увеличивает счетчик промахов кэша
func (m *AIMetrics) IncrementCacheMisses() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cacheMisses++
}

// IncrementErrors увеличивает счетчик ошибок
func (m *AIMetrics) IncrementErrors() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors++
}

// RecordProcessingTime записывает время обработки
func (m *AIMetrics) RecordProcessingTime(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.processingTime += duration
}

// GetRequestsProcessed возвращает количество обработанных запросов
func (m *AIMetrics) GetRequestsProcessed() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.requestsProcessed
}

// GetAnomaliesDetected возвращает количество обнаруженных аномалий
func (m *AIMetrics) GetAnomaliesDetected() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.anomaliesDetected
}

// GetPredictionsMade возвращает количество сделанных предсказаний
func (m *AIMetrics) GetPredictionsMade() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.predictionsMade
}

// GetRoutesOptimized возвращает количество оптимизированных маршрутов
func (m *AIMetrics) GetRoutesOptimized() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.routesOptimized
}

// GetCacheHits возвращает количество попаданий в кэш
func (m *AIMetrics) GetCacheHits() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cacheHits
}

// GetCacheMisses возвращает количество промахов кэша
func (m *AIMetrics) GetCacheMisses() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cacheMisses
}

// GetErrors возвращает количество ошибок
func (m *AIMetrics) GetErrors() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.errors
}

// GetProcessingTime возвращает общее время обработки
func (m *AIMetrics) GetProcessingTime() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.processingTime
}

// GetAverageProcessingTime возвращает среднее время обработки
func (m *AIMetrics) GetAverageProcessingTime() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.requestsProcessed == 0 {
		return 0
	}

	return m.processingTime / time.Duration(m.requestsProcessed)
}

// GetCacheHitRate возвращает процент попаданий в кэш
func (m *AIMetrics) GetCacheHitRate() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	total := m.cacheHits + m.cacheMisses
	if total == 0 {
		return 0
	}

	return float64(m.cacheHits) / float64(total) * 100
}

// GetStats возвращает все метрики в виде карты
func (m *AIMetrics) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"requests_processed":      m.requestsProcessed,
		"anomalies_detected":      m.anomaliesDetected,
		"predictions_made":        m.predictionsMade,
		"routes_optimized":        m.routesOptimized,
		"cache_hits":              m.cacheHits,
		"cache_misses":            m.cacheMisses,
		"cache_hit_rate":          m.GetCacheHitRate(),
		"errors":                  m.errors,
		"total_processing_time":   m.processingTime.String(),
		"average_processing_time": m.GetAverageProcessingTime().String(),
	}
}

// Reset сбрасывает все метрики
func (m *AIMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.requestsProcessed = 0
	m.anomaliesDetected = 0
	m.predictionsMade = 0
	m.routesOptimized = 0
	m.cacheHits = 0
	m.cacheMisses = 0
	m.errors = 0
	m.processingTime = 0
}
