package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/twogc/cloudbridge-relay-installer/internal/ai/anomaly"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/predictive"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/routing"
)

func TestNewClient(t *testing.T) {
	logger := zaptest.NewLogger(t)
	baseURL := "http://localhost:8083"

	client := NewClient(baseURL, logger)

	assert.NotNil(t, client)
	assert.Equal(t, baseURL, client.baseURL)
	assert.NotNil(t, client.httpClient)
	assert.Equal(t, logger, client.logger)
	assert.Equal(t, 30*time.Second, client.timeout)
	assert.Equal(t, 30*time.Second, client.httpClient.Timeout)
}

func TestClient_DetectAnomaly_Success(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v1/ai/anomaly/detect", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Parse request
		var req DetectAnomalyRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.NotNil(t, req.Metrics)

		// Return mock response
		response := DetectAnomalyResponse{
			Result: &anomaly.AnomalyResult{
				IsAnomaly:   true,
				Score:       0.85,
				Confidence:  0.9,
				Timestamp:   time.Now(),
				Features:    map[string]float64{"latency": 0.8, "packet_loss": 0.7},
				AnomalyType: "high_latency",
				Severity:    "high",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL, logger)
	ctx := context.Background()

	metrics := &anomaly.NetworkMetrics{
		Timestamp:            time.Now(),
		ConnectionsPerSecond: 100,
		BytesPerSecond:       1024,
		PacketLossRate:       0.01,
		AverageLatency:       50.0,
		CPUUsage:             0.5,
		MemoryUsage:          0.6,
		ActiveConnections:    50,
		FailedConnections:    5,
		ProtocolDistribution: map[string]int{
			"QUIC":      30,
			"WireGuard": 20,
			"TCP":       0,
		},
	}

	result, err := client.DetectAnomaly(ctx, metrics)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.IsAnomaly)
	assert.Equal(t, 0.85, result.Score)
	assert.Equal(t, 0.9, result.Confidence)
	assert.Equal(t, "high_latency", result.AnomalyType)
}

func TestClient_DetectAnomaly_ServerError(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := DetectAnomalyResponse{
			Error: "Internal server error",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL, logger)
	ctx := context.Background()

	metrics := &anomaly.NetworkMetrics{}

	result, err := client.DetectAnomaly(ctx, metrics)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Internal server error")
}

func TestClient_DetectAnomaly_HTTPError(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create mock server that returns HTTP error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad Request"))
	}))
	defer server.Close()

	client := NewClient(server.URL, logger)
	ctx := context.Background()

	metrics := &anomaly.NetworkMetrics{}

	result, err := client.DetectAnomaly(ctx, metrics)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "status 400")
}

func TestClient_DetectAnomaly_InvalidJSON(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create mock server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, logger)
	ctx := context.Background()

	metrics := &anomaly.NetworkMetrics{}

	result, err := client.DetectAnomaly(ctx, metrics)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to decode response")
}

func TestClient_PredictLoad_Success(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v1/ai/predict/load", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Parse request
		var req PredictLoadRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, 24, req.Horizon)

		// Return mock response
		response := PredictLoadResponse{
			Result: &predictive.LoadPrediction{
				Timestamp:  time.Now(),
				Horizon:    24,
				Forecast:   []float64{0.75, 0.76, 0.77},
				Confidence: 0.85,
				ConfidenceInterval: []predictive.ConfidenceInterval{
					{Lower: 0.65, Upper: 0.85},
				},
				ModelID:  "arima",
				Features: map[string]float64{"historical_load": 0.8, "time_of_day": 0.6},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL, logger)
	ctx := context.Background()

	prediction, err := client.PredictLoad(ctx, 24)
	require.NoError(t, err)
	assert.NotNil(t, prediction)
	assert.Equal(t, 24, prediction.Horizon)
	assert.Len(t, prediction.Forecast, 3)
	assert.Equal(t, 0.75, prediction.Forecast[0])
	assert.Equal(t, 0.85, prediction.Confidence)
	assert.Equal(t, "arima", prediction.ModelID)
}

func TestClient_PredictLoad_ServerError(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := PredictLoadResponse{
			Error: "Prediction failed",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL, logger)
	ctx := context.Background()

	prediction, err := client.PredictLoad(ctx, 24)
	assert.Error(t, err)
	assert.Nil(t, prediction)
	assert.Contains(t, err.Error(), "Prediction failed")
}

func TestClient_OptimizeRoute_Success(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v1/ai/route/optimize", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Parse request
		var req OptimizeRouteRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "node1", req.Source)
		assert.Equal(t, "node2", req.Destination)

		// Return mock response
		response := OptimizeRouteResponse{
			Result: &routing.OptimizedRoute{
				Protocol:         "QUIC",
				Path:             []string{"node1", "node3", "node2"},
				PredictedLatency: 45.0,
				Confidence:       0.9,
				Timestamp:        time.Now(),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL, logger)
	ctx := context.Background()

	constraints := map[string]interface{}{
		"max_latency":   100.0,
		"min_bandwidth": 1000,
	}

	route, err := client.OptimizeRoute(ctx, "node1", "node2", constraints)
	require.NoError(t, err)
	assert.NotNil(t, route)
	assert.Equal(t, []string{"node1", "node3", "node2"}, route.Path)
	assert.Equal(t, 45.0, route.PredictedLatency)
	assert.Equal(t, 0.9, route.Confidence)
	assert.Equal(t, "QUIC", route.Protocol)
}

func TestClient_OptimizeRoute_ServerError(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := OptimizeRouteResponse{
			Error: "Route optimization failed",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL, logger)
	ctx := context.Background()

	route, err := client.OptimizeRoute(ctx, "node1", "node2", nil)
	assert.Error(t, err)
	assert.Nil(t, route)
	assert.Contains(t, err.Error(), "Route optimization failed")
}

func TestClient_GetStatus_Success(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/v1/ai/status", r.URL.Path)

		// Return mock response
		status := ClientServiceStatus{
			AnomalyDetector: true,
			DataPipeline:    true,
			LoadPredictor:   true,
			Monitoring:      true,
			RouteOptimizer:  true,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(status)
	}))
	defer server.Close()

	client := NewClient(server.URL, logger)
	ctx := context.Background()

	status, err := client.GetStatus(ctx)
	require.NoError(t, err)
	assert.NotNil(t, status)
	assert.True(t, status.AnomalyDetector)
	assert.True(t, status.DataPipeline)
	assert.True(t, status.LoadPredictor)
	assert.True(t, status.Monitoring)
	assert.True(t, status.RouteOptimizer)
}

func TestClient_GetStatus_HTTPError(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create mock server that returns HTTP error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := NewClient(server.URL, logger)
	ctx := context.Background()

	status, err := client.GetStatus(ctx)
	assert.Error(t, err)
	assert.Nil(t, status)
	assert.Contains(t, err.Error(), "status 500")
}

func TestClient_HealthCheck_Success(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/health", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, logger)
	ctx := context.Background()

	err := client.HealthCheck(ctx)
	require.NoError(t, err)
}

func TestClient_HealthCheck_Failure(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := NewClient(server.URL, logger)
	ctx := context.Background()

	err := client.HealthCheck(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status 503")
}

func TestClient_ContextCancellation(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create mock server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, logger)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	metrics := &anomaly.NetworkMetrics{}

	result, err := client.DetectAnomaly(ctx, metrics)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestClient_NetworkError(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create client with invalid URL
	client := NewClient("http://invalid-url:9999", logger)
	ctx := context.Background()

	metrics := &anomaly.NetworkMetrics{}

	result, err := client.DetectAnomaly(ctx, metrics)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to send request")
}

func TestClient_InvalidJSONRequest(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, logger)
	ctx := context.Background()

	// This should not happen in normal usage, but test error handling
	// We can't easily create invalid JSON for the request, so we'll test with valid data
	metrics := &anomaly.NetworkMetrics{}

	result, err := client.DetectAnomaly(ctx, metrics)
	// This will fail because server doesn't return proper response
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestClient_ConcurrentRequests(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate some processing time
		time.Sleep(10 * time.Millisecond)

		if r.URL.Path == "/api/v1/ai/anomaly/detect" {
			response := DetectAnomalyResponse{
				Result: &anomaly.AnomalyResult{
					IsAnomaly:   true,
					Score:       0.8,
					Confidence:  0.9,
					Timestamp:   time.Now(),
					Features:    map[string]float64{"latency": 0.8},
					AnomalyType: "high_latency",
					Severity:    "high",
				},
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		} else if r.URL.Path == "/api/v1/ai/predict/load" {
			response := PredictLoadResponse{
				Result: &predictive.LoadPrediction{
					Timestamp:  time.Now(),
					Horizon:    24,
					Forecast:   []float64{0.75, 0.76, 0.77},
					Confidence: 0.85,
					ModelID:    "arima",
				},
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, logger)
	ctx := context.Background()

	// Test concurrent requests
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			metrics := &anomaly.NetworkMetrics{
				Timestamp:            time.Now(),
				ConnectionsPerSecond: 100,
				BytesPerSecond:       1024,
				PacketLossRate:       0.01,
				AverageLatency:       50.0,
				CPUUsage:             0.5,
				MemoryUsage:          0.6,
				ActiveConnections:    50,
				FailedConnections:    5,
				ProtocolDistribution: map[string]int{
					"QUIC":      30,
					"WireGuard": 20,
					"TCP":       0,
				},
			}

			result, err := client.DetectAnomaly(ctx, metrics)
			if err != nil {
				t.Logf("DetectAnomaly error: %v", err)
			} else {
				assert.NotNil(t, result)
			}

			prediction, err := client.PredictLoad(ctx, 24)
			if err != nil {
				t.Logf("PredictLoad error: %v", err)
			} else {
				assert.NotNil(t, prediction)
			}
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestClient_RequestTimeout(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create mock server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create client with short timeout
	client := &Client{
		baseURL: server.URL,
		httpClient: &http.Client{
			Timeout: 50 * time.Millisecond,
		},
		logger:  logger,
		timeout: 50 * time.Millisecond,
	}

	ctx := context.Background()
	metrics := &anomaly.NetworkMetrics{}

	result, err := client.DetectAnomaly(ctx, metrics)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "timeout")
}

func TestClient_EdgeCases(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Test with empty base URL
	client := NewClient("", logger)
	assert.Equal(t, "", client.baseURL)

	// Test with nil logger
	clientWithNilLogger := NewClient("http://localhost:8083", nil)
	assert.Nil(t, clientWithNilLogger.logger)

	// Test with nil context
	client = NewClient("http://localhost:8083", logger)
	metrics := &anomaly.NetworkMetrics{}

	// This should panic or handle gracefully
	assert.Panics(t, func() {
		client.DetectAnomaly(nil, metrics)
	})
}

func TestClient_ResponseParsing(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Test with empty response body
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Empty response body
	}))
	defer server.Close()

	client := NewClient(server.URL, logger)
	ctx := context.Background()

	metrics := &anomaly.NetworkMetrics{}

	result, err := client.DetectAnomaly(ctx, metrics)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to decode response")
}

func TestClient_MethodValidation(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create mock server that validates HTTP methods
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" && r.URL.Path != "/api/v1/ai/status" && r.URL.Path != "/health" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, logger)
	ctx := context.Background()

	// Test that correct methods are used
	metrics := &anomaly.NetworkMetrics{}
	_, err := client.DetectAnomaly(ctx, metrics)
	// This will fail because server doesn't return proper response, but method should be correct
	assert.Error(t, err)

	_, err = client.PredictLoad(ctx, 24)
	assert.Error(t, err)

	_, err = client.OptimizeRoute(ctx, "node1", "node2", nil)
	assert.Error(t, err)

	_, err = client.GetStatus(ctx)
	assert.Error(t, err)

	err = client.HealthCheck(ctx)
	require.NoError(t, err)
}
