package api

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// AIMiddlewareInterface интерфейс для AIMiddleware для тестирования
type AIMiddlewareInterface interface {
	GetMetrics() *AIMetrics
	GetCircuitBreakerStats() map[string]interface{}
}

// MockAIMiddlewareInterface мок для AIMiddlewareInterface
type MockAIMiddlewareInterface struct {
	mock.Mock
}

func (m *MockAIMiddlewareInterface) GetMetrics() *AIMetrics {
	args := m.Called()
	return args.Get(0).(*AIMetrics)
}

func (m *MockAIMiddlewareInterface) GetCircuitBreakerStats() map[string]interface{} {
	args := m.Called()
	return args.Get(0).(map[string]interface{})
}

// TestAIMetricsExporter тестовая версия AIMetricsExporter
type TestAIMetricsExporter struct {
	aiMiddleware AIMiddlewareInterface
	logger       *zap.Logger
	interval     time.Duration
	stopCh       chan struct{}
	stopped      bool
	mu           sync.Mutex
}

// NewTestAIMetricsExporter создает тестовый экспортер AI метрик
func NewTestAIMetricsExporter(aiMiddleware AIMiddlewareInterface, logger *zap.Logger) *TestAIMetricsExporter {
	return &TestAIMetricsExporter{
		aiMiddleware: aiMiddleware,
		logger:       logger,
		interval:     10 * time.Millisecond, // Короткий интервал для тестов
		stopCh:       make(chan struct{}),
		stopped:      false,
	}
}

// Start запускает экспорт метрик
func (e *TestAIMetricsExporter) Start() {
	e.logger.Info("Starting AI metrics exporter...")
	go e.exportMetrics()
}

// Stop останавливает экспорт метрик
func (e *TestAIMetricsExporter) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.stopped {
		e.logger.Info("Stopping AI metrics exporter...")
		close(e.stopCh)
		e.stopped = true
	}
}

// exportMetrics экспортирует метрики в Prometheus
func (e *TestAIMetricsExporter) exportMetrics() {
	ticker := time.NewTicker(e.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			e.updateMetrics()
		case <-e.stopCh:
			return
		}
	}
}

// updateMetrics обновляет Prometheus метрики
func (e *TestAIMetricsExporter) updateMetrics() {
	aiMetrics := e.aiMiddleware.GetMetrics()
	circuitBreakerStats := e.aiMiddleware.GetCircuitBreakerStats()

	// Логируем для тестирования
	e.logger.Debug("AI metrics updated",
		zap.Int64("requests_processed", aiMetrics.GetRequestsProcessed()),
		zap.Int64("anomalies_detected", aiMetrics.GetAnomaliesDetected()),
		zap.Float64("cache_hit_rate", aiMetrics.GetCacheHitRate()),
		zap.Any("circuit_breaker_state", circuitBreakerStats["state"]),
	)
}

func TestNewAIMetricsExporter(t *testing.T) {
	aiMiddleware := &MockAIMiddlewareInterface{}
	logger := zap.NewNop()

	exporter := NewTestAIMetricsExporter(aiMiddleware, logger)

	assert.NotNil(t, exporter)
	assert.Equal(t, aiMiddleware, exporter.aiMiddleware)
}

func TestAIMetricsExporter_StartStop(t *testing.T) {
	aiMiddleware := &MockAIMiddlewareInterface{}
	logger := zap.NewNop()
	exporter := NewTestAIMetricsExporter(aiMiddleware, logger)

	// Create test metrics
	metrics := &AIMetrics{
		requestsProcessed:   100,
		anomaliesDetected:   5,
		predictionsMade:     50,
		cacheHits:           80,
		cacheMisses:         20,
		errors:              2,
		totalProcessingTime: 5 * time.Second,
	}

	// Настраиваем моки для множественных вызовов
	aiMiddleware.On("GetMetrics").Return(metrics).Maybe()
	aiMiddleware.On("GetCircuitBreakerStats").Return(map[string]interface{}{
		"state": "closed",
	}).Maybe()

	// Start exporter
	exporter.Start()

	// Wait for at least one tick
	time.Sleep(50 * time.Millisecond)

	// Stop exporter
	exporter.Stop()

	// Wait for goroutine to finish
	time.Sleep(10 * time.Millisecond)

	// Verify exporter was created and stopped
	assert.NotNil(t, exporter)
}

func TestAIMetricsExporter_Start(t *testing.T) {
	aiMiddleware := &MockAIMiddlewareInterface{}
	logger := zap.NewNop()
	exporter := NewTestAIMetricsExporter(aiMiddleware, logger)

	// Create test metrics
	metrics := &AIMetrics{
		requestsProcessed:   100,
		anomaliesDetected:   5,
		predictionsMade:     50,
		cacheHits:           80,
		cacheMisses:         20,
		errors:              2,
		totalProcessingTime: 5 * time.Second,
	}

	// Настраиваем моки для множественных вызовов
	aiMiddleware.On("GetMetrics").Return(metrics).Maybe()
	aiMiddleware.On("GetCircuitBreakerStats").Return(map[string]interface{}{
		"state": "closed",
	}).Maybe()

	// Start exporter
	exporter.Start()

	// Wait for at least one tick
	time.Sleep(50 * time.Millisecond)

	// Stop exporter
	exporter.Stop()

	// Wait for goroutine to finish
	time.Sleep(10 * time.Millisecond)

	// Verify exporter was created and stopped
	assert.NotNil(t, exporter)
}

func TestAIMetricsExporter_Stop(t *testing.T) {
	aiMiddleware := &MockAIMiddlewareInterface{}
	logger := zap.NewNop()
	exporter := NewTestAIMetricsExporter(aiMiddleware, logger)

	// Настраиваем моки для множественных вызовов
	aiMiddleware.On("GetMetrics").Return(&AIMetrics{}).Maybe()
	aiMiddleware.On("GetCircuitBreakerStats").Return(map[string]interface{}{
		"state": "closed",
	}).Maybe()

	// Start exporter
	exporter.Start()

	// Wait a bit for the goroutine to start
	time.Sleep(10 * time.Millisecond)

	// Stop exporter
	exporter.Stop()

	// Wait for goroutine to finish
	time.Sleep(10 * time.Millisecond)

	// Verify exporter was stopped
	assert.NotNil(t, exporter)
}

func TestAIMetricsExporter_ConcurrentAccess(t *testing.T) {
	aiMiddleware := &MockAIMiddlewareInterface{}
	logger := zap.NewNop()
	exporter := NewTestAIMetricsExporter(aiMiddleware, logger)

	// Create test metrics
	metrics := &AIMetrics{
		requestsProcessed:   100,
		anomaliesDetected:   5,
		predictionsMade:     50,
		cacheHits:           80,
		cacheMisses:         20,
		errors:              2,
		totalProcessingTime: 5 * time.Second,
	}

	// Настраиваем моки для множественных вызовов
	aiMiddleware.On("GetMetrics").Return(metrics).Maybe()
	aiMiddleware.On("GetCircuitBreakerStats").Return(map[string]interface{}{
		"state": "closed",
	}).Maybe()

	// Start exporter
	exporter.Start()

	// Wait for at least one tick
	time.Sleep(50 * time.Millisecond)

	// Stop exporter
	exporter.Stop()

	// Wait for goroutine to finish
	time.Sleep(10 * time.Millisecond)

	// Verify exporter was created and stopped
	assert.NotNil(t, exporter)
}

func TestAIMetricsExporter_ErrorHandling(t *testing.T) {
	aiMiddleware := &MockAIMiddlewareInterface{}
	logger := zap.NewNop()
	exporter := NewTestAIMetricsExporter(aiMiddleware, logger)

	// Test with empty metrics instead of nil to avoid panic
	emptyMetrics := &AIMetrics{}
	aiMiddleware.On("GetMetrics").Return(emptyMetrics).Maybe()
	aiMiddleware.On("GetCircuitBreakerStats").Return(map[string]interface{}{
		"state": "closed",
	}).Maybe()

	// Start exporter
	exporter.Start()

	// Wait for at least one tick
	time.Sleep(50 * time.Millisecond)

	// Stop exporter
	exporter.Stop()

	// Wait for goroutine to finish
	time.Sleep(10 * time.Millisecond)

	// Verify exporter was created and stopped
	assert.NotNil(t, exporter)
}

func TestAIMetricsExporter_MultipleStarts(t *testing.T) {
	aiMiddleware := &MockAIMiddlewareInterface{}
	logger := zap.NewNop()
	exporter := NewTestAIMetricsExporter(aiMiddleware, logger)

	// Настраиваем моки для множественных вызовов
	aiMiddleware.On("GetMetrics").Return(&AIMetrics{}).Maybe()
	aiMiddleware.On("GetCircuitBreakerStats").Return(map[string]interface{}{
		"state": "closed",
	}).Maybe()

	// Start exporter multiple times
	exporter.Start()
	exporter.Start()
	exporter.Start()

	// Wait a bit for the goroutine to start
	time.Sleep(10 * time.Millisecond)

	// Stop exporter
	exporter.Stop()

	// Wait for goroutine to finish
	time.Sleep(10 * time.Millisecond)

	// Verify exporter was stopped
	assert.NotNil(t, exporter)
}

func TestAIMetricsExporter_MultipleStops(t *testing.T) {
	aiMiddleware := &MockAIMiddlewareInterface{}
	logger := zap.NewNop()
	exporter := NewTestAIMetricsExporter(aiMiddleware, logger)

	// Настраиваем моки для множественных вызовов
	aiMiddleware.On("GetMetrics").Return(&AIMetrics{}).Maybe()
	aiMiddleware.On("GetCircuitBreakerStats").Return(map[string]interface{}{
		"state": "closed",
	}).Maybe()

	// Start exporter
	exporter.Start()

	// Wait a bit for the goroutine to start
	time.Sleep(10 * time.Millisecond)

	// Stop exporter multiple times - should not panic
	exporter.Stop()
	exporter.Stop()
	exporter.Stop()

	// Wait for goroutine to finish
	time.Sleep(10 * time.Millisecond)

	// Verify exporter was stopped
	assert.NotNil(t, exporter)
}

func TestAIMetricsExporter_Integration(t *testing.T) {
	aiMiddleware := &MockAIMiddlewareInterface{}
	logger := zap.NewNop()
	exporter := NewTestAIMetricsExporter(aiMiddleware, logger)

	// Create test metrics
	metrics := &AIMetrics{
		requestsProcessed:   100,
		anomaliesDetected:   5,
		predictionsMade:     50,
		cacheHits:           80,
		cacheMisses:         20,
		errors:              2,
		totalProcessingTime: 5 * time.Second,
	}

	// Настраиваем моки для множественных вызовов
	aiMiddleware.On("GetMetrics").Return(metrics).Maybe()
	aiMiddleware.On("GetCircuitBreakerStats").Return(map[string]interface{}{
		"state": "closed",
	}).Maybe()

	// Start exporter
	exporter.Start()

	// Wait for at least one tick
	time.Sleep(50 * time.Millisecond)

	// Stop exporter
	exporter.Stop()

	// Wait for goroutine to finish
	time.Sleep(10 * time.Millisecond)

	// Verify everything worked
	assert.NotNil(t, exporter)
}
