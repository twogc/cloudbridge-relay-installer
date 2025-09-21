package api

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/twogc/cloudbridge-relay-installer/internal/ai"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/anomaly"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/predictive"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/routing"
)

// MockAIService мок для AIServiceInterface
type MockAIService struct {
	mock.Mock
}

func (m *MockAIService) GetStatus() *ai.ServiceStatus {
	args := m.Called()
	return args.Get(0).(*ai.ServiceStatus)
}

func (m *MockAIService) DetectAnomaly(ctx context.Context, metrics *anomaly.NetworkMetrics) (*anomaly.AnomalyResult, error) {
	args := m.Called(ctx, metrics)
	return args.Get(0).(*anomaly.AnomalyResult), args.Error(1)
}

func (m *MockAIService) PredictLoad(ctx context.Context, horizon int) (*predictive.LoadPrediction, error) {
	args := m.Called(ctx, horizon)
	return args.Get(0).(*predictive.LoadPrediction), args.Error(1)
}

func (m *MockAIService) OptimizeRoute(ctx context.Context, source, destination string, constraints map[string]interface{}) (*routing.OptimizedRoute, error) {
	args := m.Called(ctx, source, destination, constraints)
	return args.Get(0).(*routing.OptimizedRoute), args.Error(1)
}

func TestNewAIEndpoints(t *testing.T) {
	logger := zap.NewNop()
	mockService := &MockAIService{}

	endpoints := NewAIEndpoints(mockService, logger)

	assert.NotNil(t, endpoints)
	assert.Equal(t, mockService, endpoints.aiService)
	assert.Equal(t, logger, endpoints.logger)
}

func TestAIEndpoints_GetAIStatus(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*MockAIService)
		expectedStatus int
	}{
		{
			name: "successful status retrieval",
			mockSetup: func(m *MockAIService) {
				m.On("GetStatus").Return(&ai.ServiceStatus{
					Status: "active",
					Components: map[string]bool{
						"anomaly_detector": true,
						"load_predictor":   true,
					},
					Health: true,
					Uptime: time.Hour,
				})
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockAIService{}
			tt.mockSetup(mockService)

			endpoints := &AIEndpoints{
				aiService: mockService,
				logger:    zap.NewNop(),
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/ai/status", nil)

			endpoints.GetAIStatus(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestAIEndpoints_DetectAnomaly(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		mockSetup      func(*MockAIService)
		expectedStatus int
	}{
		{
			name:        "successful anomaly detection",
			requestBody: `{"metrics": {"timestamp": "2023-01-01T00:00:00Z", "connections_per_second": 100, "bytes_per_second": 1000, "packet_loss_rate": 0.01, "average_latency": 50.0, "cpu_usage": 75.0, "memory_usage": 60.0, "active_connections": 50, "failed_connections": 2, "protocol_distribution": {"tcp": 80, "udp": 20}}}`,
			mockSetup: func(m *MockAIService) {
				m.On("DetectAnomaly", mock.Anything, mock.AnythingOfType("*anomaly.NetworkMetrics")).Return(&anomaly.AnomalyResult{
					IsAnomaly:   true,
					AnomalyType: "high_cpu",
					Confidence:  0.95,
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid JSON",
			requestBody:    `{"invalid": json}`,
			mockSetup:      func(m *MockAIService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing metrics",
			requestBody:    `{}`,
			mockSetup:      func(m *MockAIService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "service error",
			requestBody: `{"metrics": {"timestamp": "2023-01-01T00:00:00Z", "connections_per_second": 100}}`,
			mockSetup: func(m *MockAIService) {
				m.On("DetectAnomaly", mock.Anything, mock.AnythingOfType("*anomaly.NetworkMetrics")).Return((*anomaly.AnomalyResult)(nil), assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockAIService{}
			tt.mockSetup(mockService)

			endpoints := &AIEndpoints{
				aiService: mockService,
				logger:    zap.NewNop(),
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/ai/anomaly", bytes.NewBufferString(tt.requestBody))
			c.Request.Header.Set("Content-Type", "application/json")

			endpoints.DetectAnomaly(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestAIEndpoints_PredictLoad(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		mockSetup      func(*MockAIService)
		expectedStatus int
	}{
		{
			name: "successful load prediction with default horizon",
			url:  "/ai/predict",
			mockSetup: func(m *MockAIService) {
				m.On("PredictLoad", mock.Anything, 24).Return(&predictive.LoadPrediction{
					Forecast:   []float64{22.3, 25.1, 28.7, 30.2, 32.1},
					Confidence: 0.88,
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "successful load prediction with custom horizon",
			url:  "/ai/predict?horizon=5",
			mockSetup: func(m *MockAIService) {
				m.On("PredictLoad", mock.Anything, 5).Return(&predictive.LoadPrediction{
					Forecast:   []float64{22.3, 25.1, 28.7, 30.2, 32.1},
					Confidence: 0.88,
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid horizon parameter",
			url:            "/ai/predict?horizon=invalid",
			mockSetup:      func(m *MockAIService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			url:  "/ai/predict?horizon=3",
			mockSetup: func(m *MockAIService) {
				m.On("PredictLoad", mock.Anything, 3).Return((*predictive.LoadPrediction)(nil), assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockAIService{}
			tt.mockSetup(mockService)

			endpoints := &AIEndpoints{
				aiService: mockService,
				logger:    zap.NewNop(),
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", tt.url, nil)

			endpoints.PredictLoad(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestAIEndpoints_OptimizeRoute(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		mockSetup      func(*MockAIService)
		expectedStatus int
	}{
		{
			name:        "successful route optimization",
			requestBody: `{"source": "A", "destination": "B", "constraints": {"max_latency": 100}}`,
			mockSetup: func(m *MockAIService) {
				m.On("OptimizeRoute", mock.Anything, "A", "B", mock.Anything).Return(&routing.OptimizedRoute{
					Protocol:         "tcp",
					Path:             []string{"A", "C", "B"},
					PredictedLatency: 85.5,
					Confidence:       0.92,
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid JSON",
			requestBody:    `{"invalid": json}`,
			mockSetup:      func(m *MockAIService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing source",
			requestBody:    `{"destination": "B"}`,
			mockSetup:      func(m *MockAIService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing destination",
			requestBody:    `{"source": "A"}`,
			mockSetup:      func(m *MockAIService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "service error",
			requestBody: `{"source": "A", "destination": "B"}`,
			mockSetup: func(m *MockAIService) {
				m.On("OptimizeRoute", mock.Anything, "A", "B", mock.Anything).Return((*routing.OptimizedRoute)(nil), assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockAIService{}
			tt.mockSetup(mockService)

			endpoints := &AIEndpoints{
				aiService: mockService,
				logger:    zap.NewNop(),
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/ai/optimize", bytes.NewBufferString(tt.requestBody))
			c.Request.Header.Set("Content-Type", "application/json")

			endpoints.OptimizeRoute(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestAIEndpoints_GetAIMetrics(t *testing.T) {
	endpoints := &AIEndpoints{
		aiService: nil,
		logger:    zap.NewNop(),
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/ai/metrics", nil)

	endpoints.GetAIMetrics(c)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestAIEndpoints_GetAIHealth(t *testing.T) {
	endpoints := &AIEndpoints{
		aiService: nil,
		logger:    zap.NewNop(),
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/ai/health", nil)

	endpoints.GetAIHealth(c)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestAIEndpoints_NilService(t *testing.T) {
	endpoints := &AIEndpoints{
		aiService: nil,
		logger:    zap.NewNop(),
	}

	tests := []struct {
		name           string
		method         string
		url            string
		requestBody    string
		expectedStatus int
	}{
		{
			name:           "GetAIStatus with nil service",
			method:         "GET",
			url:            "/ai/status",
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name:           "DetectAnomaly with nil service",
			method:         "POST",
			url:            "/ai/anomaly",
			requestBody:    `{"metrics": {"timestamp": "2023-01-01T00:00:00Z", "connections_per_second": 100}}`,
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name:           "PredictLoad with nil service",
			method:         "POST",
			url:            "/ai/predict",
			requestBody:    `{"horizon": 5}`,
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name:           "OptimizeRoute with nil service",
			method:         "POST",
			url:            "/ai/optimize",
			requestBody:    `{"source": "A", "destination": "B"}`,
			expectedStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			var body *bytes.Buffer
			if tt.requestBody != "" {
				body = bytes.NewBufferString(tt.requestBody)
				c.Request = httptest.NewRequest(tt.method, tt.url, body)
				c.Request.Header.Set("Content-Type", "application/json")
			} else {
				c.Request = httptest.NewRequest(tt.method, tt.url, nil)
			}

			switch tt.url {
			case "/ai/status":
				endpoints.GetAIStatus(c)
			case "/ai/anomaly":
				endpoints.DetectAnomaly(c)
			case "/ai/predict":
				endpoints.PredictLoad(c)
			case "/ai/optimize":
				endpoints.OptimizeRoute(c)
			}

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
