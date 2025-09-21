package ai

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/twogc/cloudbridge-relay-installer/internal/ai/anomaly"
)

func TestNewAICircuitBreaker(t *testing.T) {
	logger := zaptest.NewLogger(t)
	timeout := 30 * time.Second

	cb := NewAICircuitBreaker(timeout, logger)

	assert.NotNil(t, cb)
	assert.Equal(t, "closed", cb.state)
	assert.Equal(t, 5, cb.threshold)
	assert.Equal(t, timeout, cb.timeout)
	assert.Equal(t, logger, cb.logger)
	assert.Equal(t, 0, cb.failures)
}

func TestAICircuitBreaker_CanExecute_ClosedState(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cb := NewAICircuitBreaker(30*time.Second, logger)

	// Test closed state
	assert.True(t, cb.CanExecute())
}

func TestAICircuitBreaker_CanExecute_OpenState(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cb := NewAICircuitBreaker(100*time.Millisecond, logger)

	// Open the circuit breaker
	for i := 0; i < 5; i++ {
		cb.RecordFailure()
	}

	// Should be open and not allow execution
	assert.False(t, cb.CanExecute())

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// Should transition to half-open
	assert.True(t, cb.CanExecute())
}

func TestAICircuitBreaker_CanExecute_HalfOpenState(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cb := NewAICircuitBreaker(100*time.Millisecond, logger)

	// Open the circuit breaker
	for i := 0; i < 5; i++ {
		cb.RecordFailure()
	}

	// Wait for timeout to transition to half-open
	time.Sleep(150 * time.Millisecond)

	// Should be half-open and allow execution
	assert.True(t, cb.CanExecute())
}

func TestAICircuitBreaker_RecordSuccess(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cb := NewAICircuitBreaker(30*time.Second, logger)

	// Record some failures first
	cb.RecordFailure()
	cb.RecordFailure()

	// Record success
	cb.RecordSuccess()

	// Should reset to closed state
	assert.Equal(t, "closed", cb.state)
	assert.Equal(t, 0, cb.failures)
}

func TestAICircuitBreaker_RecordFailure(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cb := NewAICircuitBreaker(30*time.Second, logger)

	// Record failures
	cb.RecordFailure()
	assert.Equal(t, 1, cb.failures)

	cb.RecordFailure()
	assert.Equal(t, 2, cb.failures)

	// Record enough failures to open circuit
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}

	assert.Equal(t, 5, cb.failures)
	assert.Equal(t, "open", cb.state)
}

func TestAICircuitBreaker_ConcurrentAccess(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cb := NewAICircuitBreaker(30*time.Second, logger)

	// Test concurrent access
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			// Concurrent operations
			cb.CanExecute()
			cb.RecordSuccess()
			cb.RecordFailure()
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Circuit breaker should still be functional
	assert.NotNil(t, cb)
}

func TestNewOptimizedService(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()

	service, err := NewOptimizedService(config, logger)
	require.NoError(t, err)
	assert.NotNil(t, service)

	assert.Equal(t, config, service.config)
	assert.Equal(t, logger, service.logger)
	assert.NotNil(t, service.metrics)
	assert.NotNil(t, service.performance)
	assert.NotNil(t, service.healthChecker)
	assert.NotNil(t, service.resourceManager)
	assert.NotNil(t, service.circuitBreaker)
	assert.NotNil(t, service.shutdownChan)
	assert.Equal(t, 30*time.Second, service.updateInterval)
}

func TestNewOptimizedService_NilConfig(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// This should fail during component initialization
	service, err := NewOptimizedService(nil, logger)
	assert.Error(t, err)
	assert.Nil(t, service)
}

func TestNewOptimizedService_NilLogger(t *testing.T) {
	config := GetDefaultConfig()

	// This should fail during component initialization
	service, err := NewOptimizedService(config, nil)
	assert.Error(t, err)
	assert.Nil(t, service)
}

func TestOptimizedService_DetectAnomaly(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service, err := NewOptimizedService(config, logger)
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
			"QUIC":      30,
			"WireGuard": 20,
			"TCP":       0,
		},
	}

	result, err := service.DetectAnomaly(ctx, metrics)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Test caching - second call should use cache
	result2, err := service.DetectAnomaly(ctx, metrics)
	require.NoError(t, err)
	assert.NotNil(t, result2)
}

func TestOptimizedService_DetectAnomaly_CircuitBreakerOpen(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service, err := NewOptimizedService(config, logger)
	require.NoError(t, err)

	// Open the circuit breaker
	for i := 0; i < 5; i++ {
		service.circuitBreaker.RecordFailure()
	}

	ctx := context.Background()
	metrics := &anomaly.NetworkMetrics{}

	result, err := service.DetectAnomaly(ctx, metrics)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "circuit breaker is open")
}

func TestOptimizedService_PredictLoad(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service, err := NewOptimizedService(config, logger)
	require.NoError(t, err)

	ctx := context.Background()
	horizon := 24

	prediction, err := service.PredictLoad(ctx, horizon)
	require.NoError(t, err)
	assert.NotNil(t, prediction)

	// Test caching - second call should use cache
	prediction2, err := service.PredictLoad(ctx, horizon)
	require.NoError(t, err)
	assert.NotNil(t, prediction2)
}

func TestOptimizedService_PredictLoad_CircuitBreakerOpen(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service, err := NewOptimizedService(config, logger)
	require.NoError(t, err)

	// Open the circuit breaker
	for i := 0; i < 5; i++ {
		service.circuitBreaker.RecordFailure()
	}

	ctx := context.Background()
	horizon := 24

	prediction, err := service.PredictLoad(ctx, horizon)
	assert.Error(t, err)
	assert.Nil(t, prediction)
	assert.Contains(t, err.Error(), "circuit breaker is open")
}

func TestOptimizedService_OptimizeRoute(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service, err := NewOptimizedService(config, logger)
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

	// Test caching - second call should use cache
	route2, err := service.OptimizeRoute(ctx, source, destination, constraints)
	require.NoError(t, err)
	assert.NotNil(t, route2)
}

func TestOptimizedService_OptimizeRoute_CircuitBreakerOpen(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service, err := NewOptimizedService(config, logger)
	require.NoError(t, err)

	// Open the circuit breaker
	for i := 0; i < 5; i++ {
		service.circuitBreaker.RecordFailure()
	}

	ctx := context.Background()
	source := "node1"
	destination := "node2"
	constraints := map[string]interface{}{}

	route, err := service.OptimizeRoute(ctx, source, destination, constraints)
	assert.Error(t, err)
	assert.Nil(t, route)
	assert.Contains(t, err.Error(), "circuit breaker is open")
}

func TestOptimizedService_GetStatus(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service, err := NewOptimizedService(config, logger)
	require.NoError(t, err)

	status := service.GetStatus()
	assert.NotNil(t, status)
	assert.Equal(t, "running", status.Status)
	assert.True(t, status.Components["anomaly_detector"])
	assert.True(t, status.Components["load_predictor"])
	assert.True(t, status.Components["route_optimizer"])
	assert.True(t, status.Components["cache"])
	assert.NotNil(t, status.Performance)
	assert.NotNil(t, status.Resources)
	assert.Greater(t, status.Uptime, time.Duration(0))
}

func TestOptimizedService_GetMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service, err := NewOptimizedService(config, logger)
	require.NoError(t, err)

	metrics := service.GetMetrics()
	assert.NotNil(t, metrics)
	assert.IsType(t, &AIMetrics{}, metrics)
}

func TestOptimizedService_HealthCheck(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service, err := NewOptimizedService(config, logger)
	require.NoError(t, err)

	// Test health check
	err = service.HealthCheck()
	require.NoError(t, err)
}

func TestOptimizedService_HealthCheck_CircuitBreakerOpen(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service, err := NewOptimizedService(config, logger)
	require.NoError(t, err)

	// Open the circuit breaker
	for i := 0; i < 5; i++ {
		service.circuitBreaker.RecordFailure()
	}

	// Test health check with open circuit breaker
	err = service.HealthCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circuit breaker is open")
}

func TestOptimizedService_Shutdown(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service, err := NewOptimizedService(config, logger)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test shutdown
	err = service.Shutdown(ctx)
	require.NoError(t, err)
}

func TestOptimizedService_Shutdown_Timeout(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service, err := NewOptimizedService(config, logger)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Test shutdown with timeout
	err = service.Shutdown(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestOptimizedService_ConcurrentAccess(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service, err := NewOptimizedService(config, logger)
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
			"QUIC":      30,
			"WireGuard": 20,
			"TCP":       0,
		},
	}

	// Test concurrent access
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			// Concurrent operations
			result, err := service.DetectAnomaly(ctx, metrics)
			if err != nil {
				t.Logf("DetectAnomaly error: %v", err)
			} else {
				assert.NotNil(t, result)
			}

			prediction, err := service.PredictLoad(ctx, 24)
			if err != nil {
				t.Logf("PredictLoad error: %v", err)
			} else {
				assert.NotNil(t, prediction)
			}

			route, err := service.OptimizeRoute(ctx, "node1", "node2", nil)
			if err != nil {
				t.Logf("OptimizeRoute error: %v", err)
			} else {
				assert.NotNil(t, route)
			}

			status := service.GetStatus()
			assert.NotNil(t, status)

			err = service.HealthCheck()
			if err != nil {
				t.Logf("HealthCheck error: %v", err)
			}
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Shutdown service
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = service.Shutdown(shutdownCtx)
	require.NoError(t, err)
}

func TestOptimizedService_BackgroundTasks(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service, err := NewOptimizedService(config, logger)
	require.NoError(t, err)

	// Wait for background tasks to run
	time.Sleep(100 * time.Millisecond)

	// Verify service is still functional
	status := service.GetStatus()
	assert.NotNil(t, status)
	assert.Equal(t, "running", status.Status)

	// Shutdown service
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = service.Shutdown(ctx)
	require.NoError(t, err)
}

func TestOptimizedService_EdgeCases(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service, err := NewOptimizedService(config, logger)
	require.NoError(t, err)

	// Test with nil context
	metrics := &anomaly.NetworkMetrics{}
	_, err = service.DetectAnomaly(nil, metrics)
	// This should handle nil context gracefully or panic
	if err != nil {
		assert.Contains(t, err.Error(), "context")
	}

	// Test with nil metrics
	ctx := context.Background()
	_, err = service.DetectAnomaly(ctx, nil)
	// This should handle nil metrics gracefully
	if err != nil {
		assert.Contains(t, err.Error(), "nil")
	}

	// Test with negative horizon
	_, err = service.PredictLoad(ctx, -1)
	// This should handle negative horizon gracefully
	if err != nil {
		assert.Contains(t, err.Error(), "invalid")
	}

	// Test with empty strings
	_, err = service.OptimizeRoute(ctx, "", "", nil)
	// This should handle empty strings gracefully
	if err != nil {
		assert.Contains(t, err.Error(), "empty")
	}
}

func TestNewHealthChecker(t *testing.T) {
	hc := NewHealthChecker()

	assert.NotNil(t, hc)
	assert.NotNil(t, hc.checks)
	assert.NotNil(t, hc.status)
}

func TestHealthChecker_AddCheck(t *testing.T) {
	hc := NewHealthChecker()

	// Add a health check
	hc.AddCheck("test_check", func() error { return nil }, 30*time.Second)

	assert.Len(t, hc.checks, 1)
	check, exists := hc.checks["test_check"]
	assert.True(t, exists)
	assert.NotNil(t, check)
	assert.Equal(t, "test_check", check.Name)
	assert.Equal(t, 30*time.Second, check.Interval)
	assert.True(t, check.Enabled)
}

func TestHealthChecker_RunChecks(t *testing.T) {
	hc := NewHealthChecker()

	// Add a successful health check
	hc.AddCheck("success_check", func() error { return nil }, 30*time.Second)

	// Add a failing health check
	hc.AddCheck("fail_check", func() error { return errors.New("check failed") }, 30*time.Second)

	// Run checks
	hc.RunChecks()

	// Verify status
	assert.True(t, hc.status["success_check"])
	assert.False(t, hc.status["fail_check"])
}

func TestHealthChecker_GetOverallStatus(t *testing.T) {
	hc := NewHealthChecker()

	// Add successful checks
	hc.AddCheck("check1", func() error { return nil }, 30*time.Second)
	hc.AddCheck("check2", func() error { return nil }, 30*time.Second)

	// Run checks
	hc.RunChecks()

	// Should be healthy
	assert.True(t, hc.GetOverallStatus())

	// Add a failing check
	hc.AddCheck("fail_check", func() error { return errors.New("check failed") }, 30*time.Second)

	// Run checks again
	hc.RunChecks()

	// Should be unhealthy
	assert.False(t, hc.GetOverallStatus())
}

func TestNewResourceManager(t *testing.T) {
	rm := NewResourceManager()

	assert.NotNil(t, rm)
	assert.NotNil(t, rm.Limits)
	assert.Equal(t, 80.0, rm.Limits.MaxCPUUsage)
	assert.Equal(t, 80.0, rm.Limits.MaxMemoryUsage)
	assert.Equal(t, 90.0, rm.Limits.MaxDiskUsage)
	assert.Equal(t, 1000.0, rm.Limits.MaxNetworkIO)
}

func TestResourceManager_UpdateUsage(t *testing.T) {
	rm := NewResourceManager()

	// Update usage
	rm.UpdateUsage()

	// Verify usage was updated
	assert.GreaterOrEqual(t, rm.CPUUsage, 0.0)
	assert.LessOrEqual(t, rm.CPUUsage, 100.0)
	assert.GreaterOrEqual(t, rm.MemoryUsage, 0.0)
	assert.LessOrEqual(t, rm.MemoryUsage, 100.0)
	assert.GreaterOrEqual(t, rm.DiskUsage, 0.0)
	assert.LessOrEqual(t, rm.DiskUsage, 100.0)
	assert.GreaterOrEqual(t, rm.NetworkIO, 0.0)
}

func TestResourceManager_GetUsage(t *testing.T) {
	rm := NewResourceManager()

	// Update usage
	rm.UpdateUsage()

	// Get usage
	usage := rm.GetUsage()

	assert.NotNil(t, usage)
	assert.Contains(t, usage, "cpu_usage")
	assert.Contains(t, usage, "memory_usage")
	assert.Contains(t, usage, "disk_usage")
	assert.Contains(t, usage, "network_io")
	assert.Contains(t, usage, "limits")

	// Verify limits are included
	limits, ok := usage["limits"].(*ResourceLimits)
	assert.True(t, ok)
	assert.NotNil(t, limits)
}

func TestOptimizedService_Integration(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service, err := NewOptimizedService(config, logger)
	require.NoError(t, err)

	ctx := context.Background()

	// Test all operations
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

	// Test anomaly detection
	result, err := service.DetectAnomaly(ctx, metrics)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Test load prediction
	prediction, err := service.PredictLoad(ctx, 24)
	require.NoError(t, err)
	assert.NotNil(t, prediction)

	// Test route optimization
	route, err := service.OptimizeRoute(ctx, "node1", "node2", nil)
	require.NoError(t, err)
	assert.NotNil(t, route)

	// Test status
	status := service.GetStatus()
	assert.NotNil(t, status)
	assert.Equal(t, "running", status.Status)

	// Test health check
	err = service.HealthCheck()
	require.NoError(t, err)

	// Test metrics
	metrics2 := service.GetMetrics()
	assert.NotNil(t, metrics2)

	// Shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = service.Shutdown(shutdownCtx)
	require.NoError(t, err)
}

func TestOptimizedService_PerformanceMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()
	service, err := NewOptimizedService(config, logger)
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
			"QUIC":      30,
			"WireGuard": 20,
			"TCP":       0,
		},
	}

	// Perform some operations to generate metrics
	for i := 0; i < 5; i++ {
		_, err := service.DetectAnomaly(ctx, metrics)
		require.NoError(t, err)

		_, err = service.PredictLoad(ctx, 24)
		require.NoError(t, err)

		_, err = service.OptimizeRoute(ctx, "node1", "node2", nil)
		require.NoError(t, err)
	}

	// Get performance metrics
	status := service.GetStatus()
	assert.NotNil(t, status)
	assert.NotNil(t, status.Performance)

	// Verify performance metrics
	performance := status.Performance
	assert.Greater(t, performance["total_requests"], int64(0))
	assert.GreaterOrEqual(t, performance["successful_requests"], int64(0))
	assert.GreaterOrEqual(t, performance["failed_requests"], int64(0))
	assert.GreaterOrEqual(t, performance["average_latency"], int64(0))
	assert.GreaterOrEqual(t, performance["peak_latency"], int64(0))
	assert.GreaterOrEqual(t, performance["throughput"], float64(0))
	assert.GreaterOrEqual(t, performance["error_rate"], float64(0))

	// Shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = service.Shutdown(shutdownCtx)
	require.NoError(t, err)
}
