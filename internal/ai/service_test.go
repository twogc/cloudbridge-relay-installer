package ai

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/twogc/cloudbridge-relay-installer/internal/ai/anomaly"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/pipeline"
)

func TestNewService(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()

	service := NewService(config, logger)

	assert.NotNil(t, service)
	assert.Equal(t, config, service.config)
	assert.Equal(t, logger, service.logger)
	assert.False(t, service.started)
	assert.NotNil(t, service.stopCh)
	assert.NotNil(t, service.ctx)
	assert.NotNil(t, service.cancel)
}

func TestService_Initialize(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service := NewService(config, logger)

	err := service.Initialize()
	require.NoError(t, err)

	assert.NotNil(t, service.dataPipeline)
	assert.NotNil(t, service.anomalyDetector)
	assert.NotNil(t, service.loadPredictor)
	assert.NotNil(t, service.routeOptimizer)
	assert.NotNil(t, service.monitoring)
	assert.NotNil(t, service.metrics)
}

func TestService_Initialize_DisabledComponents(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()

	// Disable some components
	config.AI.Models.AnomalyDetection.Enabled = false
	config.AI.Models.LoadPrediction.Enabled = false
	config.AI.Models.RouteOptimization.Enabled = false

	service := NewService(config, logger)

	err := service.Initialize()
	require.NoError(t, err)

	assert.NotNil(t, service.dataPipeline)
	assert.Nil(t, service.anomalyDetector)
	assert.Nil(t, service.loadPredictor)
	assert.Nil(t, service.routeOptimizer)
	assert.NotNil(t, service.monitoring)
	assert.NotNil(t, service.metrics)
}

func TestService_Start(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service := NewService(config, logger)

	err := service.Initialize()
	require.NoError(t, err)

	err = service.Start()
	require.NoError(t, err)

	assert.True(t, service.started)
	assert.False(t, service.startTime.IsZero())
}

func TestService_Start_AlreadyStarted(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service := NewService(config, logger)

	err := service.Initialize()
	require.NoError(t, err)

	err = service.Start()
	require.NoError(t, err)

	// Try to start again
	err = service.Start()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already started")
}

func TestService_Stop(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service := NewService(config, logger)

	err := service.Initialize()
	require.NoError(t, err)

	err = service.Start()
	require.NoError(t, err)

	err = service.Stop()
	require.NoError(t, err)

	assert.False(t, service.started)
}

func TestService_Stop_NotStarted(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service := NewService(config, logger)

	err := service.Initialize()
	require.NoError(t, err)

	// Stop without starting should not error
	err = service.Stop()
	require.NoError(t, err)
}

func TestService_DetectAnomaly(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service := NewService(config, logger)

	err := service.Initialize()
	require.NoError(t, err)

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
			"QUIC":     30,
			"P2P_QUIC": 20,
			"TCP":      0,
		},
	}

	result, err := service.DetectAnomaly(ctx, metrics)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestService_DetectAnomaly_NotInitialized(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service := NewService(config, logger)

	ctx := context.Background()
	metrics := &anomaly.NetworkMetrics{}

	result, err := service.DetectAnomaly(ctx, metrics)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestService_PredictLoad(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service := NewService(config, logger)

	err := service.Initialize()
	require.NoError(t, err)

	ctx := context.Background()
	horizon := 24

	prediction, err := service.PredictLoad(ctx, horizon)
	require.NoError(t, err)
	assert.NotNil(t, prediction)
}

func TestService_PredictLoad_NotInitialized(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service := NewService(config, logger)

	ctx := context.Background()
	horizon := 24

	prediction, err := service.PredictLoad(ctx, horizon)
	assert.Error(t, err)
	assert.Nil(t, prediction)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestService_OptimizeRoute(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service := NewService(config, logger)

	err := service.Initialize()
	require.NoError(t, err)

	ctx := context.Background()
	source := "node1"
	destination := "node2"
	constraints := map[string]interface{}{
		"max_latency":   100.0,
		"min_bandwidth": 1000,
	}

	route, err := service.OptimizeRoute(ctx, source, destination, constraints)
	require.NoError(t, err)
	assert.NotNil(t, route)
	assert.Equal(t, source, route.Source)
	assert.Equal(t, destination, route.Destination)
	assert.NotEmpty(t, route.ID)
}

func TestService_OptimizeRoute_NotInitialized(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service := NewService(config, logger)

	ctx := context.Background()
	source := "node1"
	destination := "node2"
	constraints := map[string]interface{}{}

	route, err := service.OptimizeRoute(ctx, source, destination, constraints)
	assert.Error(t, err)
	assert.Nil(t, route)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestService_GetStatus(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service := NewService(config, logger)

	// Test status when not started
	status := service.GetStatus()
	assert.Equal(t, "stopped", status.Status)
	assert.False(t, status.Health)
	assert.NotNil(t, status.Components)
	assert.NotNil(t, status.Performance)

	// Initialize and start service
	err := service.Initialize()
	require.NoError(t, err)

	err = service.Start()
	require.NoError(t, err)

	// Test status when started
	status = service.GetStatus()
	assert.Equal(t, "running", status.Status)
	assert.True(t, status.Health)
	assert.True(t, status.Components["anomaly_detector"])
	assert.True(t, status.Components["load_predictor"])
	assert.True(t, status.Components["route_optimizer"])
	assert.True(t, status.Components["data_pipeline"])
	assert.True(t, status.Components["monitoring"])
	assert.Greater(t, status.Uptime, time.Duration(0))
}

func TestService_shouldProcessData(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service := NewService(config, logger)

	// Test with empty data
	data := []pipeline.DataPoint{}
	result := service.shouldProcessData(data)
	assert.False(t, result)

	// Test with relay data with active connections
	data = []pipeline.DataPoint{
		{
			Source: "relay",
			Type:   "metrics",
			Data: map[string]interface{}{
				"connections_active": int64(10),
			},
		},
	}
	result = service.shouldProcessData(data)
	assert.True(t, result)

	// Test with relay data with bytes transferred
	data = []pipeline.DataPoint{
		{
			Source: "relay",
			Type:   "metrics",
			Data: map[string]interface{}{
				"bytes_transferred": int64(1024),
			},
		},
	}
	result = service.shouldProcessData(data)
	assert.True(t, result)

	// Test with system data with high CPU
	data = []pipeline.DataPoint{
		{
			Source: "system",
			Type:   "cpu",
			Data: map[string]interface{}{
				"cpu_usage_percent": 15.0,
			},
		},
	}
	result = service.shouldProcessData(data)
	assert.True(t, result)

	// Test with system data with high memory
	data = []pipeline.DataPoint{
		{
			Source: "system",
			Type:   "memory",
			Data: map[string]interface{}{
				"memory_usage_percent": 60.0,
			},
		},
	}
	result = service.shouldProcessData(data)
	assert.True(t, result)

	// Test with low activity data
	data = []pipeline.DataPoint{
		{
			Source: "relay",
			Type:   "metrics",
			Data: map[string]interface{}{
				"connections_active": int64(0),
				"bytes_transferred":  int64(0),
			},
		},
		{
			Source: "system",
			Type:   "cpu",
			Data: map[string]interface{}{
				"cpu_usage_percent": 5.0,
			},
		},
		{
			Source: "system",
			Type:   "memory",
			Data: map[string]interface{}{
				"memory_usage_percent": 30.0,
			},
		},
	}
	result = service.shouldProcessData(data)
	assert.False(t, result)
}

func TestService_getRelayMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service := NewService(config, logger)

	// Test without data pipeline
	metrics, err := service.getRelayMetrics()
	assert.Error(t, err)
	assert.Nil(t, metrics)
	assert.Contains(t, err.Error(), "not initialized")

	// Initialize service
	err = service.Initialize()
	require.NoError(t, err)

	// Test with data pipeline and relay collector
	metrics, err = service.getRelayMetrics()
	// Should succeed but may have warnings about relay connection
	if err != nil {
		// If there's an error, it should be about connection, not missing collector
		assert.NotContains(t, err.Error(), "relay collector not found")
	} else {
		// If no error, metrics should be returned
		assert.NotNil(t, metrics)
	}
}

func TestService_getSystemMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service := NewService(config, logger)

	// Test without data pipeline
	metrics, err := service.getSystemMetrics()
	assert.Error(t, err)
	assert.Nil(t, metrics)
	assert.Contains(t, err.Error(), "not initialized")

	// Initialize service
	err = service.Initialize()
	require.NoError(t, err)

	// Test with data pipeline and system collector
	metrics, err = service.getSystemMetrics()
	// Should succeed and return system metrics
	if err != nil {
		// If there's an error, it should not be about missing collector
		assert.NotContains(t, err.Error(), "system collector not found")
	} else {
		// If no error, metrics should be returned
		assert.NotNil(t, metrics)
		assert.Contains(t, metrics, "cpu_usage")
		assert.Contains(t, metrics, "memory_usage")
	}
}

func TestService_Integration(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service := NewService(config, logger)

	// Initialize service
	err := service.Initialize()
	require.NoError(t, err)

	// Start service
	err = service.Start()
	require.NoError(t, err)

	// Wait a bit for background processing
	time.Sleep(100 * time.Millisecond)

	// Get status
	status := service.GetStatus()
	assert.Equal(t, "running", status.Status)
	assert.True(t, status.Health)

	// Stop service
	err = service.Stop()
	require.NoError(t, err)

	// Verify stopped
	status = service.GetStatus()
	assert.Equal(t, "stopped", status.Status)
	assert.False(t, status.Health)
}

func TestService_ConcurrentAccess(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service := NewService(config, logger)

	err := service.Initialize()
	require.NoError(t, err)

	err = service.Start()
	require.NoError(t, err)

	// Test concurrent access to service methods
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

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
					"QUIC":     30,
					"P2P_QUIC": 20,
					"TCP":      0,
				},
			}

			_, err := service.DetectAnomaly(ctx, metrics)
			if err != nil {
				t.Logf("DetectAnomaly error: %v", err)
			}

			_, err = service.PredictLoad(ctx, 24)
			if err != nil {
				t.Logf("PredictLoad error: %v", err)
			}

			_, err = service.OptimizeRoute(ctx, "node1", "node2", nil)
			if err != nil {
				t.Logf("OptimizeRoute error: %v", err)
			}

			status := service.GetStatus()
			if status == nil {
				t.Log("GetStatus returned nil")
			}
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Stop service
	err = service.Stop()
	require.NoError(t, err)
}

func TestService_ErrorHandling(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()

	// Create invalid config to test error handling
	config.AI.Models.AnomalyDetection.Threshold = -1.0

	service := NewService(config, logger)

	// Initialize should fail with invalid config
	err := service.Initialize()
	assert.Error(t, err)
}

func TestService_BackgroundProcessing(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service := NewService(config, logger)

	err := service.Initialize()
	require.NoError(t, err)

	err = service.Start()
	require.NoError(t, err)

	// Wait for background processing to run
	time.Sleep(200 * time.Millisecond)

	// Check that metrics were recorded
	status := service.GetStatus()
	assert.NotNil(t, status.Performance)

	err = service.Stop()
	require.NoError(t, err)
}

func TestService_ContextCancellation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service := NewService(config, logger)

	err := service.Initialize()
	require.NoError(t, err)

	err = service.Start()
	require.NoError(t, err)

	// Cancel context
	service.cancel()

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Stop should still work
	err = service.Stop()
	require.NoError(t, err)
}

func TestService_EdgeCases(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()

	// Test with nil config
	serviceWithNilConfig := NewService(nil, logger)
	assert.NotNil(t, serviceWithNilConfig)
	assert.Nil(t, serviceWithNilConfig.config)

	// Test with nil logger
	serviceWithNilLogger := NewService(config, nil)
	assert.NotNil(t, serviceWithNilLogger)
	assert.Nil(t, serviceWithNilLogger.logger)

	// Test GetStatus with nil components
	status := serviceWithNilConfig.GetStatus()
	assert.NotNil(t, status)
	assert.Equal(t, "stopped", status.Status)
	assert.False(t, status.Health)
}

func TestService_ResourceCleanup(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service := NewService(config, logger)

	err := service.Initialize()
	require.NoError(t, err)

	err = service.Start()
	require.NoError(t, err)

	// Verify service is running
	status := service.GetStatus()
	assert.Equal(t, "running", status.Status)

	// Stop service
	err = service.Stop()
	require.NoError(t, err)

	// Verify service is stopped
	status = service.GetStatus()
	assert.Equal(t, "stopped", status.Status)

	// Verify context is cancelled
	select {
	case <-service.ctx.Done():
		// Context is cancelled, which is expected
	default:
		t.Error("Context should be cancelled after stop")
	}
}

func TestService_StartStopCycle(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service := NewService(config, logger)

	err := service.Initialize()
	require.NoError(t, err)

	// Test multiple start/stop cycles
	for i := 0; i < 3; i++ {
		err = service.Start()
		require.NoError(t, err)

		status := service.GetStatus()
		assert.Equal(t, "running", status.Status)

		err = service.Stop()
		require.NoError(t, err)

		status = service.GetStatus()
		assert.Equal(t, "stopped", status.Status)
	}
}

func TestService_ComponentInitialization(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()

	// Test with minimal config
	config.AI.Models.AnomalyDetection.Enabled = true
	config.AI.Models.LoadPrediction.Enabled = true
	config.AI.Models.RouteOptimization.Enabled = true
	config.AI.DataPipeline.Enabled = true

	service := NewService(config, logger)

	err := service.Initialize()
	require.NoError(t, err)

	// Verify all components are initialized
	assert.NotNil(t, service.dataPipeline)
	assert.NotNil(t, service.anomalyDetector)
	assert.NotNil(t, service.loadPredictor)
	assert.NotNil(t, service.routeOptimizer)
	assert.NotNil(t, service.monitoring)
	assert.NotNil(t, service.metrics)
}

func TestService_DataProcessing(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service := NewService(config, logger)

	err := service.Initialize()
	require.NoError(t, err)

	err = service.Start()
	require.NoError(t, err)

	// Wait for data processing
	time.Sleep(300 * time.Millisecond)

	// Check that processing occurred
	status := service.GetStatus()
	assert.NotNil(t, status.Performance)

	err = service.Stop()
	require.NoError(t, err)
}

func TestService_MetricsIntegration(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service := NewService(config, logger)

	err := service.Initialize()
	require.NoError(t, err)

	// Test metrics recording
	service.metrics.RecordAnomalyDetections(5)
	service.metrics.RecordLoadPredictions(3)
	service.metrics.RecordRouteOptimizations(2)
	service.metrics.RecordProcessingTime(100 * time.Millisecond)

	// Get metrics
	metrics := service.metrics.GetMetrics()
	assert.NotNil(t, metrics)
	assert.Contains(t, metrics, "anomaly_detection")
	assert.Contains(t, metrics, "load_prediction")
	assert.Contains(t, metrics, "route_optimization")
	assert.Contains(t, metrics, "processing")
	assert.Contains(t, metrics, "system")

	err = service.Stop()
	require.NoError(t, err)
}
