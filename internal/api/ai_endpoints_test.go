package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/anomaly"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/predictive"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/routing"
	"go.uber.org/zap"
)

// TestMockAIService мок для тестирования AIServiceInterface
type TestMockAIService struct {
	mock.Mock
}

func (m *TestMockAIService) GetStatus() *ai.ServiceStatus {
	args := m.Called()
	return args.Get(0).(*ai.ServiceStatus)
}

func (m *TestMockAIService) DetectAnomaly(ctx context.Context, metrics *anomaly.NetworkMetrics) (*anomaly.AnomalyResult, error) {
	args := m.Called(ctx, metrics)
	return args.Get(0).(*anomaly.AnomalyResult), args.Error(1)
}

func (m *TestMockAIService) PredictLoad(ctx context.Context, horizon int) (*predictive.LoadPrediction, error) {
	args := m.Called(ctx, horizon)
	return args.Get(0).(*predictive.LoadPrediction), args.Error(1)
}

func (m *TestMockAIService) OptimizeRoute(ctx context.Context, source, destination string, constraints map[string]interface{}) (*routing.OptimizedRoute, error) {
	args := m.Called(ctx, source, destination, constraints)
	return args.Get(0).(*routing.OptimizedRoute), args.Error(1)
}

func TestAIEndpoints_GetAIStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	t.Run("success", func(t *testing.T) {
		mockAI := new(TestMockAIService)
		status := &ai.ServiceStatus{
			Status: "running",
			Components: map[string]bool{
				"anomaly_detection": true,
				"load_prediction":   true,
			},
		}
		mockAI.On("GetStatus").Return(status)

		endpoints := NewAIEndpoints(mockAI, logger)
		router := gin.New()
		router.GET("/ai/status", endpoints.GetAIStatus)

		req := httptest.NewRequest("GET", "/ai/status", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response AIStatusResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "healthy", response.Status)
		assert.NotNil(t, response.Components)

		mockAI.AssertExpectations(t)
	})

	t.Run("service unavailable", func(t *testing.T) {
		endpoints := NewAIEndpoints(nil, logger)
		router := gin.New()
		router.GET("/ai/status", endpoints.GetAIStatus)

		req := httptest.NewRequest("GET", "/ai/status", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}

func TestAIEndpoints_DetectAnomaly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	t.Run("success", func(t *testing.T) {
		mockAI := new(TestMockAIService)
		result := &anomaly.AnomalyResult{
			IsAnomaly:   true,
			Score:       0.95,
			AnomalyType: "high_cpu_usage",
			Severity:    "high",
		}
		mockAI.On("DetectAnomaly", mock.Anything, mock.Anything).Return(result, nil)

		endpoints := NewAIEndpoints(mockAI, logger)
		router := gin.New()
		router.POST("/ai/detect-anomaly", endpoints.DetectAnomaly)

		requestBody := AnomalyDetectionRequest{
			Metrics: &anomaly.NetworkMetrics{
				BytesPerSecond:       1000,
				ConnectionsPerSecond: 10,
				PacketLossRate:       0.01,
			},
		}
		jsonBody, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/ai/detect-anomaly", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response["status"])

		mockAI.AssertExpectations(t)
	})

	t.Run("invalid request", func(t *testing.T) {
		endpoints := NewAIEndpoints(nil, logger)
		router := gin.New()
		router.POST("/ai/detect-anomaly", endpoints.DetectAnomaly)

		req := httptest.NewRequest("POST", "/ai/detect-anomaly", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}

func TestAIEndpoints_PredictLoad(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	t.Run("success", func(t *testing.T) {
		mockAI := new(TestMockAIService)
		prediction := &predictive.LoadPrediction{
			Forecast:   []float64{0.5, 0.6, 0.7},
			Confidence: 0.85,
		}
		mockAI.On("PredictLoad", mock.Anything, 24).Return(prediction, nil)

		endpoints := NewAIEndpoints(mockAI, logger)
		router := gin.New()
		router.GET("/ai/predict-load", endpoints.PredictLoad)

		req := httptest.NewRequest("GET", "/ai/predict-load?horizon=24", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response["status"])

		mockAI.AssertExpectations(t)
	})

	t.Run("invalid horizon", func(t *testing.T) {
		endpoints := NewAIEndpoints(nil, logger)
		router := gin.New()
		router.GET("/ai/predict-load", endpoints.PredictLoad)

		req := httptest.NewRequest("GET", "/ai/predict-load?horizon=invalid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}

func TestAIEndpoints_OptimizeRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	t.Run("success", func(t *testing.T) {
		mockAI := new(TestMockAIService)
		route := &routing.OptimizedRoute{
			Path:             []string{"A", "B", "C"},
			PredictedLatency: 100.5,
			Confidence:       0.9,
			Protocol:         "QUIC",
		}
		mockAI.On("OptimizeRoute", mock.Anything, "A", "C", mock.Anything).Return(route, nil)

		endpoints := NewAIEndpoints(mockAI, logger)
		router := gin.New()
		router.POST("/ai/optimize-route", endpoints.OptimizeRoute)

		requestBody := RouteOptimizationRequest{
			Source:      "A",
			Destination: "C",
			Constraints: map[string]interface{}{
				"max_latency": 200,
			},
		}
		jsonBody, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/ai/optimize-route", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response["status"])

		mockAI.AssertExpectations(t)
	})
}

func TestAIEndpoints_GetAIHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	t.Run("healthy", func(t *testing.T) {
		mockAI := new(TestMockAIService)
		status := &ai.ServiceStatus{
			Status: "running",
			Components: map[string]bool{
				"anomaly_detection": true,
			},
		}
		mockAI.On("GetStatus").Return(status)

		endpoints := NewAIEndpoints(mockAI, logger)
		router := gin.New()
		router.GET("/ai/health", endpoints.GetAIHealth)

		req := httptest.NewRequest("GET", "/ai/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "healthy", response["status"])

		mockAI.AssertExpectations(t)
	})

	t.Run("unhealthy", func(t *testing.T) {
		mockAI := new(TestMockAIService)
		status := &ai.ServiceStatus{
			Status: "stopped",
			Components: map[string]bool{
				"anomaly_detection": false,
			},
		}
		mockAI.On("GetStatus").Return(status)

		endpoints := NewAIEndpoints(mockAI, logger)
		router := gin.New()
		router.GET("/ai/health", endpoints.GetAIHealth)

		req := httptest.NewRequest("GET", "/ai/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "unhealthy", response["status"])

		mockAI.AssertExpectations(t)
	})
}
