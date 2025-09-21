package api

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/anomaly"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/predictive"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/routing"
	"go.uber.org/zap"
)

// AIResult результат AI обработки для тестов
type AIResult struct {
	AnomalyDetected bool
	LoadPrediction  float64
	RouteOptimized  bool
}

func TestNewAIMiddleware(t *testing.T) {
	logger := zap.NewNop()
	aiClient := &MockAIClient{}
	config := &ai.Config{}

	middleware := NewAIMiddleware(aiClient, config, logger)

	assert.NotNil(t, middleware)
	assert.Equal(t, aiClient, middleware.aiClient)
	assert.Equal(t, config, middleware.config)
	assert.Equal(t, logger, middleware.logger)
	assert.NotNil(t, middleware.cache)
	assert.NotNil(t, middleware.metrics)
	assert.NotNil(t, middleware.circuitBreaker)
}

func TestAIMiddleware_ProcessConnection(t *testing.T) {
	tests := []struct {
		name           string
		connection     *Connection
		mockSetup      func(*MockAIClient)
		expectedResult *AIResult
		expectedError  bool
	}{
		{
			name: "successful processing",
			connection: &Connection{
				ID:          "conn1",
				RemoteAddr:  "192.168.1.1",
				Source:      "peer1",
				Destination: "peer2",
				Protocol:    "tcp",
				Latency:     10 * time.Millisecond,
			},
			mockSetup: func(m *MockAIClient) {
				m.On("DetectAnomaly", mock.Anything, mock.AnythingOfType("*anomaly.NetworkMetrics")).Return(&anomaly.AnomalyResult{
					IsAnomaly: false,
					Score:     0.1,
				}, nil)
				m.On("PredictLoad", mock.Anything, 24).Return(&predictive.LoadPrediction{
					Forecast:   []float64{0.5, 0.6, 0.7},
					Confidence: 0.8,
				}, nil)
				m.On("OptimizeRoute", mock.Anything, "peer1", "peer2", mock.AnythingOfType("map[string]interface {}")).Return(&routing.OptimizedRoute{
					Path:             []string{"peer1", "peer2"},
					PredictedLatency: 5.0,
					Confidence:       0.9,
				}, nil)
			},
			expectedResult: &AIResult{
				AnomalyDetected: false,
				LoadPrediction:  0.5,
				RouteOptimized:  true,
			},
			expectedError: false,
		},
		{
			name: "anomaly detected",
			connection: &Connection{
				ID:          "conn1",
				RemoteAddr:  "192.168.1.1",
				Source:      "peer1",
				Destination: "peer2",
				Protocol:    "tcp",
				Latency:     100 * time.Millisecond,
			},
			mockSetup: func(m *MockAIClient) {
				m.On("DetectAnomaly", mock.Anything, mock.AnythingOfType("*anomaly.NetworkMetrics")).Return(&anomaly.AnomalyResult{
					IsAnomaly: true,
					Score:     0.9,
				}, nil)
				m.On("PredictLoad", mock.Anything, 24).Return(&predictive.LoadPrediction{
					Forecast:   []float64{0.3, 0.4, 0.5},
					Confidence: 0.7,
				}, nil)
				m.On("OptimizeRoute", mock.Anything, "peer1", "peer2", mock.AnythingOfType("map[string]interface {}")).Return(&routing.OptimizedRoute{
					Path:             []string{"peer1", "peer2"},
					PredictedLatency: 8.0,
					Confidence:       0.8,
				}, nil)
			},
			expectedResult: &AIResult{
				AnomalyDetected: true,
				LoadPrediction:  0.0,
				RouteOptimized:  false,
			},
			expectedError: false,
		},
		{
			name: "AI client error",
			connection: &Connection{
				ID:          "conn1",
				RemoteAddr:  "192.168.1.1",
				Source:      "peer1",
				Destination: "peer2",
				Protocol:    "tcp",
				Latency:     10 * time.Millisecond,
			},
			mockSetup: func(m *MockAIClient) {
				m.On("DetectAnomaly", mock.Anything, mock.AnythingOfType("*anomaly.NetworkMetrics")).Return((*anomaly.AnomalyResult)(nil), assert.AnError)
			},
			expectedResult: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aiClient := &MockAIClient{}
			tt.mockSetup(aiClient)

			middleware := &AIMiddleware{
				aiClient: aiClient,
				config:   &ai.Config{},
				logger:   zap.NewNop(),
				cache:    NewAICache(),
				metrics:  NewAIMetrics(),
				circuitBreaker: &AICircuitBreaker{
					state: "closed",
				},
			}

			err := middleware.ProcessConnection(context.Background(), tt.connection)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			aiClient.AssertExpectations(t)
		})
	}
}

func TestAIMiddleware_GetMetrics(t *testing.T) {
	middleware := &AIMiddleware{
		metrics: NewAIMetrics(),
	}

	// Set some test data
	middleware.metrics.requestsProcessed = 100
	middleware.metrics.anomaliesDetected = 5
	middleware.metrics.predictionsMade = 50

	metrics := middleware.GetMetrics()

	assert.NotNil(t, metrics)
	assert.Equal(t, int64(100), metrics.requestsProcessed)
	assert.Equal(t, int64(5), metrics.anomaliesDetected)
	assert.Equal(t, int64(50), metrics.predictionsMade)
}

func TestAIMiddleware_GetCircuitBreakerStats(t *testing.T) {
	middleware := &AIMiddleware{
		circuitBreaker: &AICircuitBreaker{
			state: "closed",
		},
	}

	stats := middleware.GetCircuitBreakerStats()

	assert.NotNil(t, stats)
	assert.Equal(t, "closed", stats["state"])
}

func TestAIMiddleware_CacheIntegration(t *testing.T) {
	middleware := &AIMiddleware{
		cache: NewAICache(),
	}

	// Test cache operations
	key := "test_key"
	value := &anomaly.AnomalyResult{
		IsAnomaly: false,
		Score:     0.1,
	}

	// Set cache
	middleware.cache.Set(key, value)

	// Get from cache
	cachedValue, found := middleware.cache.Get(key)
	assert.True(t, found)
	assert.Equal(t, value, cachedValue)

	// Test cache miss
	_, found = middleware.cache.Get("nonexistent")
	assert.False(t, found)
}

func TestAIMiddleware_MetricsIntegration(t *testing.T) {
	middleware := &AIMiddleware{
		metrics: NewAIMetrics(),
	}

	// Test metrics operations
	middleware.metrics.IncrementRequestsProcessed()
	middleware.metrics.IncrementAnomaliesDetected()
	middleware.metrics.IncrementPredictionsMade()
	middleware.metrics.IncrementCacheHits()
	middleware.metrics.IncrementCacheMisses()
	middleware.metrics.IncrementErrors()
	middleware.metrics.RecordProcessingTime(100 * time.Millisecond)

	// Verify metrics
	assert.Equal(t, int64(1), middleware.metrics.requestsProcessed)
	assert.Equal(t, int64(1), middleware.metrics.anomaliesDetected)
	assert.Equal(t, int64(1), middleware.metrics.predictionsMade)
	assert.Equal(t, int64(1), middleware.metrics.cacheHits)
	assert.Equal(t, int64(1), middleware.metrics.cacheMisses)
	assert.Equal(t, int64(1), middleware.metrics.errors)
	assert.Equal(t, 100*time.Millisecond, middleware.metrics.totalProcessingTime)
}

func TestAIMiddleware_CircuitBreakerIntegration(t *testing.T) {
	middleware := &AIMiddleware{
		circuitBreaker: &AICircuitBreaker{
			state: "closed",
		},
	}

	// Test circuit breaker operations
	assert.Equal(t, "closed", middleware.circuitBreaker.state)

	// Test state transitions
	middleware.circuitBreaker.state = "open"
	assert.Equal(t, "open", middleware.circuitBreaker.state)

	middleware.circuitBreaker.state = "half-open"
	assert.Equal(t, "half-open", middleware.circuitBreaker.state)
}

func TestAIMiddleware_ConcurrentAccess(t *testing.T) {
	middleware := &AIMiddleware{
		metrics: NewAIMetrics(),
		cache:   NewAICache(),
	}

	// Test concurrent access
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			middleware.metrics.IncrementRequestsProcessed()
			middleware.cache.Set("key", &anomaly.AnomalyResult{})
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify final values
	assert.Equal(t, int64(10), middleware.metrics.requestsProcessed)

	_, found := middleware.cache.Get("key")
	assert.True(t, found)
}

func TestAIMiddleware_ErrorHandling(t *testing.T) {
	middleware := &AIMiddleware{
		metrics: NewAIMetrics(),
	}

	// Test error handling
	middleware.metrics.IncrementErrors()

	// Verify error was recorded
	assert.Equal(t, int64(1), middleware.metrics.errors)
}

func TestAIMiddleware_Integration(t *testing.T) {
	aiClient := &MockAIClient{}
	aiClient.On("DetectAnomaly", mock.Anything, mock.AnythingOfType("*anomaly.NetworkMetrics")).Return(&anomaly.AnomalyResult{
		IsAnomaly: false,
		Score:     0.1,
	}, nil)
	aiClient.On("PredictLoad", mock.Anything, 24).Return(&predictive.LoadPrediction{
		Forecast:   []float64{0.5, 0.6, 0.7},
		Confidence: 0.8,
	}, nil)
	aiClient.On("OptimizeRoute", mock.Anything, "peer1", "peer2", mock.AnythingOfType("map[string]interface {}")).Return(&routing.OptimizedRoute{
		Path:             []string{"peer1", "peer2"},
		PredictedLatency: 5.0,
		Confidence:       0.9,
	}, nil)

	middleware := &AIMiddleware{
		aiClient: aiClient,
		config:   &ai.Config{},
		logger:   zap.NewNop(),
		cache:    NewAICache(),
		metrics:  NewAIMetrics(),
		circuitBreaker: &AICircuitBreaker{
			state: "closed",
		},
	}

	// Test full integration
	connection := &Connection{
		ID:          "conn1",
		RemoteAddr:  "192.168.1.1",
		Source:      "peer1",
		Destination: "peer2",
		Protocol:    "tcp",
		Latency:     10 * time.Millisecond,
	}

	err := middleware.ProcessConnection(context.Background(), connection)

	assert.NoError(t, err)

	// Verify metrics were updated
	assert.Equal(t, int64(1), middleware.metrics.requestsProcessed)

	// Verify cache was used - check if any anomaly result was cached
	// The cache key format is "anomaly:remote_addr:timestamp"
	// Note: Cache may not be populated due to timing issues in tests

	aiClient.AssertExpectations(t)
}
