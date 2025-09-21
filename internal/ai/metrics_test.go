package ai

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAIMetrics(t *testing.T) {
	metrics := NewAIMetrics()

	assert.NotNil(t, metrics)
	assert.Equal(t, int64(0), metrics.anomalyDetections)
	assert.Equal(t, int64(0), metrics.anomalyFalsePositives)
	assert.Equal(t, int64(0), metrics.anomalyFalseNegatives)
	assert.Equal(t, float64(0), metrics.anomalyAccuracy)
	assert.Equal(t, int64(0), metrics.loadPredictions)
	assert.Equal(t, int64(0), metrics.loadPredictionErrors)
	assert.Equal(t, float64(0), metrics.loadPredictionAccuracy)
	assert.Equal(t, int64(0), metrics.routeOptimizations)
	assert.NotNil(t, metrics.routeOptimizationGains)
	assert.Equal(t, float64(0), metrics.averageOptimizationGain)
	assert.Equal(t, time.Duration(0), metrics.processingTime)
	assert.Equal(t, time.Duration(0), metrics.totalProcessingTime)
	assert.Equal(t, int64(0), metrics.processingCount)
	assert.Equal(t, time.Duration(0), metrics.averageProcessingTime)
	assert.Equal(t, int64(0), metrics.memoryUsage)
	assert.Equal(t, float64(0), metrics.cpuUsage)
	assert.Equal(t, int64(0), metrics.activeConnections)
}

func TestAIMetrics_RecordAnomalyDetection(t *testing.T) {
	metrics := NewAIMetrics()

	// Test correct anomaly detection
	metrics.RecordAnomalyDetection(true, true)
	assert.Equal(t, int64(1), metrics.anomalyDetections)
	assert.Equal(t, int64(0), metrics.anomalyFalsePositives)
	assert.Equal(t, int64(0), metrics.anomalyFalseNegatives)
	assert.Equal(t, float64(1), metrics.anomalyAccuracy)

	// Test false positive
	metrics.RecordAnomalyDetection(true, false)
	assert.Equal(t, int64(2), metrics.anomalyDetections)
	assert.Equal(t, int64(1), metrics.anomalyFalsePositives)
	assert.Equal(t, int64(0), metrics.anomalyFalseNegatives)
	assert.Equal(t, float64(0.5), metrics.anomalyAccuracy)

	// Test false negative
	metrics.RecordAnomalyDetection(false, false)
	assert.Equal(t, int64(3), metrics.anomalyDetections)
	assert.Equal(t, int64(1), metrics.anomalyFalsePositives)
	assert.Equal(t, int64(1), metrics.anomalyFalseNegatives)
	assert.Equal(t, float64(1.0/3.0), metrics.anomalyAccuracy)

	// Test correct normal detection
	metrics.RecordAnomalyDetection(false, true)
	assert.Equal(t, int64(4), metrics.anomalyDetections)
	assert.Equal(t, int64(1), metrics.anomalyFalsePositives)
	assert.Equal(t, int64(1), metrics.anomalyFalseNegatives)
	assert.Equal(t, float64(0.5), metrics.anomalyAccuracy)
}

func TestAIMetrics_RecordAnomalyDetections(t *testing.T) {
	metrics := NewAIMetrics()

	// Test recording multiple detections
	metrics.RecordAnomalyDetections(5)
	assert.Equal(t, int64(5), metrics.anomalyDetections)

	// Test recording more detections
	metrics.RecordAnomalyDetections(3)
	assert.Equal(t, int64(8), metrics.anomalyDetections)

	// Test recording zero detections
	metrics.RecordAnomalyDetections(0)
	assert.Equal(t, int64(8), metrics.anomalyDetections)

	// Test recording negative detections (should still work)
	metrics.RecordAnomalyDetections(-2)
	assert.Equal(t, int64(6), metrics.anomalyDetections)
}

func TestAIMetrics_RecordLoadPrediction(t *testing.T) {
	metrics := NewAIMetrics()

	// Test accurate prediction (within 10% error)
	metrics.RecordLoadPrediction(100.0, 95.0) // 5% error
	assert.Equal(t, int64(1), metrics.loadPredictions)
	assert.Equal(t, int64(0), metrics.loadPredictionErrors)
	assert.Equal(t, float64(1), metrics.loadPredictionAccuracy)

	// Test inaccurate prediction (more than 10% error)
	metrics.RecordLoadPrediction(100.0, 80.0) // 20% error
	assert.Equal(t, int64(2), metrics.loadPredictions)
	assert.Equal(t, int64(1), metrics.loadPredictionErrors)
	assert.Equal(t, float64(0.5), metrics.loadPredictionAccuracy)

	// Test edge case with zero actual value
	metrics.RecordLoadPrediction(0.0, 10.0)
	assert.Equal(t, int64(3), metrics.loadPredictions)
	assert.Equal(t, int64(2), metrics.loadPredictionErrors)
	assert.Equal(t, float64(1.0/3.0), metrics.loadPredictionAccuracy)

	// Test exact prediction
	metrics.RecordLoadPrediction(100.0, 100.0) // 0% error
	assert.Equal(t, int64(4), metrics.loadPredictions)
	assert.Equal(t, int64(2), metrics.loadPredictionErrors)
	assert.Equal(t, float64(0.5), metrics.loadPredictionAccuracy)
}

func TestAIMetrics_RecordLoadPredictions(t *testing.T) {
	metrics := NewAIMetrics()

	// Test recording multiple predictions
	metrics.RecordLoadPredictions(10)
	assert.Equal(t, int64(10), metrics.loadPredictions)

	// Test recording more predictions
	metrics.RecordLoadPredictions(5)
	assert.Equal(t, int64(15), metrics.loadPredictions)

	// Test recording zero predictions
	metrics.RecordLoadPredictions(0)
	assert.Equal(t, int64(15), metrics.loadPredictions)
}

func TestAIMetrics_RecordRouteOptimization(t *testing.T) {
	metrics := NewAIMetrics()

	// Test recording single optimization
	metrics.RecordRouteOptimization(0.15)
	assert.Equal(t, int64(1), metrics.routeOptimizations)
	assert.Len(t, metrics.routeOptimizationGains, 1)
	assert.Equal(t, 0.15, metrics.routeOptimizationGains[0])
	assert.Equal(t, 0.15, metrics.averageOptimizationGain)

	// Test recording multiple optimizations
	metrics.RecordRouteOptimization(0.25)
	metrics.RecordRouteOptimization(0.35)
	assert.Equal(t, int64(3), metrics.routeOptimizations)
	assert.Len(t, metrics.routeOptimizationGains, 3)
	assert.Equal(t, 0.25, metrics.averageOptimizationGain) // (0.15 + 0.25 + 0.35) / 3

	// Test recording zero gain
	metrics.RecordRouteOptimization(0.0)
	assert.Equal(t, int64(4), metrics.routeOptimizations)
	assert.Len(t, metrics.routeOptimizationGains, 4)
	assert.Equal(t, 0.1875, metrics.averageOptimizationGain) // (0.15 + 0.25 + 0.35 + 0.0) / 4

	// Test recording negative gain
	metrics.RecordRouteOptimization(-0.1)
	assert.Equal(t, int64(5), metrics.routeOptimizations)
	assert.Len(t, metrics.routeOptimizationGains, 5)
	assert.Equal(t, 0.13, metrics.averageOptimizationGain) // (0.15 + 0.25 + 0.35 + 0.0 - 0.1) / 5
}

func TestAIMetrics_RecordRouteOptimization_Limit(t *testing.T) {
	metrics := NewAIMetrics()

	// Test that gains are limited to 1000 entries
	for i := 0; i < 1001; i++ {
		metrics.RecordRouteOptimization(float64(i))
	}

	assert.Equal(t, int64(1001), metrics.routeOptimizations)
	assert.Len(t, metrics.routeOptimizationGains, 1000)                 // Should be limited to 1000
	assert.Equal(t, float64(1), metrics.routeOptimizationGains[0])      // First entry should be 1 (not 0)
	assert.Equal(t, float64(1000), metrics.routeOptimizationGains[999]) // Last entry should be 1000
}

func TestAIMetrics_RecordRouteOptimizations(t *testing.T) {
	metrics := NewAIMetrics()

	// Test recording multiple optimizations
	metrics.RecordRouteOptimizations(5)
	assert.Equal(t, int64(5), metrics.routeOptimizations)

	// Test recording more optimizations
	metrics.RecordRouteOptimizations(3)
	assert.Equal(t, int64(8), metrics.routeOptimizations)

	// Test recording zero optimizations
	metrics.RecordRouteOptimizations(0)
	assert.Equal(t, int64(8), metrics.routeOptimizations)
}

func TestAIMetrics_RecordProcessingTime(t *testing.T) {
	metrics := NewAIMetrics()

	// Test recording processing time
	duration1 := 100 * time.Millisecond
	metrics.RecordProcessingTime(duration1)
	assert.Equal(t, duration1, metrics.processingTime)
	assert.Equal(t, duration1, metrics.totalProcessingTime)
	assert.Equal(t, int64(1), metrics.processingCount)
	assert.Equal(t, duration1, metrics.averageProcessingTime)

	// Test recording another processing time
	duration2 := 200 * time.Millisecond
	metrics.RecordProcessingTime(duration2)
	assert.Equal(t, duration2, metrics.processingTime)
	assert.Equal(t, duration1+duration2, metrics.totalProcessingTime)
	assert.Equal(t, int64(2), metrics.processingCount)
	assert.Equal(t, 150*time.Millisecond, metrics.averageProcessingTime)

	// Test recording zero processing time
	duration3 := time.Duration(0)
	metrics.RecordProcessingTime(duration3)
	assert.Equal(t, duration3, metrics.processingTime)
	assert.Equal(t, duration1+duration2, metrics.totalProcessingTime)
	assert.Equal(t, int64(3), metrics.processingCount)
	assert.Equal(t, 100*time.Millisecond, metrics.averageProcessingTime) // (100 + 200 + 0) / 3
}

func TestAIMetrics_UpdateSystemMetrics(t *testing.T) {
	metrics := NewAIMetrics()

	// Test updating system metrics
	memoryUsage := int64(1024 * 1024 * 1024) // 1GB
	cpuUsage := 75.5
	activeConnections := int64(150)

	metrics.UpdateSystemMetrics(memoryUsage, cpuUsage, activeConnections)
	assert.Equal(t, memoryUsage, metrics.memoryUsage)
	assert.Equal(t, cpuUsage, metrics.cpuUsage)
	assert.Equal(t, activeConnections, metrics.activeConnections)

	// Test updating with zero values
	metrics.UpdateSystemMetrics(0, 0.0, 0)
	assert.Equal(t, int64(0), metrics.memoryUsage)
	assert.Equal(t, float64(0), metrics.cpuUsage)
	assert.Equal(t, int64(0), metrics.activeConnections)

	// Test updating with negative values
	metrics.UpdateSystemMetrics(-100, -10.0, -5)
	assert.Equal(t, int64(-100), metrics.memoryUsage)
	assert.Equal(t, float64(-10.0), metrics.cpuUsage)
	assert.Equal(t, int64(-5), metrics.activeConnections)
}

func TestAIMetrics_GetMetrics(t *testing.T) {
	metrics := NewAIMetrics()

	// Set up some test data
	metrics.RecordAnomalyDetection(true, true)
	metrics.RecordAnomalyDetection(true, false) // false positive
	metrics.RecordLoadPrediction(100.0, 95.0)
	metrics.RecordLoadPrediction(100.0, 80.0) // error
	metrics.RecordRouteOptimization(0.15)
	metrics.RecordProcessingTime(100 * time.Millisecond)
	metrics.UpdateSystemMetrics(1024*1024*1024, 75.5, 150)

	// Get all metrics
	allMetrics := metrics.GetMetrics()
	assert.NotNil(t, allMetrics)

	// Test anomaly detection metrics
	anomalyMetrics, exists := allMetrics["anomaly_detection"].(map[string]interface{})
	require.True(t, exists)
	assert.Equal(t, int64(2), anomalyMetrics["total_detections"])
	assert.Equal(t, int64(1), anomalyMetrics["false_positives"])
	assert.Equal(t, int64(0), anomalyMetrics["false_negatives"])
	assert.Equal(t, float64(0.5), anomalyMetrics["accuracy"])

	// Test load prediction metrics
	loadMetrics, exists := allMetrics["load_prediction"].(map[string]interface{})
	require.True(t, exists)
	assert.Equal(t, int64(2), loadMetrics["total_predictions"])
	assert.Equal(t, int64(1), loadMetrics["prediction_errors"])
	assert.Equal(t, float64(0.5), loadMetrics["accuracy"])

	// Test route optimization metrics
	routeMetrics, exists := allMetrics["route_optimization"].(map[string]interface{})
	require.True(t, exists)
	assert.Equal(t, int64(1), routeMetrics["total_optimizations"])
	assert.Equal(t, 0.15, routeMetrics["average_gain"])

	// Test processing metrics
	processingMetrics, exists := allMetrics["processing"].(map[string]interface{})
	require.True(t, exists)
	assert.Equal(t, int64(100), processingMetrics["current_time"]) // 100ms in milliseconds
	assert.Equal(t, int64(100), processingMetrics["average_time"])
	assert.Equal(t, int64(1), processingMetrics["total_count"])

	// Test system metrics
	systemMetrics, exists := allMetrics["system"].(map[string]interface{})
	require.True(t, exists)
	assert.Equal(t, int64(1024*1024*1024), systemMetrics["memory_usage"])
	assert.Equal(t, 75.5, systemMetrics["cpu_usage"])
	assert.Equal(t, int64(150), systemMetrics["active_connections"])
}

func TestAIMetrics_GetAnomalyMetrics(t *testing.T) {
	metrics := NewAIMetrics()

	// Set up test data
	metrics.RecordAnomalyDetection(true, true)
	metrics.RecordAnomalyDetection(true, false)  // false positive
	metrics.RecordAnomalyDetection(false, false) // false negative

	anomalyMetrics := metrics.GetAnomalyMetrics()
	assert.NotNil(t, anomalyMetrics)
	assert.Equal(t, int64(3), anomalyMetrics["total_detections"])
	assert.Equal(t, int64(1), anomalyMetrics["false_positives"])
	assert.Equal(t, int64(1), anomalyMetrics["false_negatives"])
	assert.Equal(t, float64(1.0/3.0), anomalyMetrics["accuracy"])
}

func TestAIMetrics_GetLoadPredictionMetrics(t *testing.T) {
	metrics := NewAIMetrics()

	// Set up test data
	metrics.RecordLoadPrediction(100.0, 95.0) // accurate
	metrics.RecordLoadPrediction(100.0, 80.0) // error

	loadMetrics := metrics.GetLoadPredictionMetrics()
	assert.NotNil(t, loadMetrics)
	assert.Equal(t, int64(2), loadMetrics["total_predictions"])
	assert.Equal(t, int64(1), loadMetrics["prediction_errors"])
	assert.Equal(t, float64(0.5), loadMetrics["accuracy"])
}

func TestAIMetrics_GetRouteOptimizationMetrics(t *testing.T) {
	metrics := NewAIMetrics()

	// Set up test data
	metrics.RecordRouteOptimization(0.15)
	metrics.RecordRouteOptimization(0.25)

	routeMetrics := metrics.GetRouteOptimizationMetrics()
	assert.NotNil(t, routeMetrics)
	assert.Equal(t, int64(2), routeMetrics["total_optimizations"])
	assert.Equal(t, 0.2, routeMetrics["average_gain"]) // (0.15 + 0.25) / 2
}

func TestAIMetrics_GetProcessingMetrics(t *testing.T) {
	metrics := NewAIMetrics()

	// Set up test data
	metrics.RecordProcessingTime(100 * time.Millisecond)
	metrics.RecordProcessingTime(200 * time.Millisecond)

	processingMetrics := metrics.GetProcessingMetrics()
	assert.NotNil(t, processingMetrics)
	assert.Equal(t, int64(200), processingMetrics["current_time"]) // 200ms in milliseconds
	assert.Equal(t, int64(150), processingMetrics["average_time"]) // (100 + 200) / 2
	assert.Equal(t, int64(2), processingMetrics["total_count"])
}

func TestAIMetrics_GetSystemMetrics(t *testing.T) {
	metrics := NewAIMetrics()

	// Set up test data
	metrics.UpdateSystemMetrics(1024*1024*1024, 75.5, 150)

	systemMetrics := metrics.GetSystemMetrics()
	assert.NotNil(t, systemMetrics)
	assert.Equal(t, int64(1024*1024*1024), systemMetrics["memory_usage"])
	assert.Equal(t, 75.5, systemMetrics["cpu_usage"])
	assert.Equal(t, int64(150), systemMetrics["active_connections"])
}

func TestAIMetrics_Reset(t *testing.T) {
	metrics := NewAIMetrics()

	// Set up test data
	metrics.RecordAnomalyDetection(true, true)
	metrics.RecordLoadPrediction(100.0, 95.0)
	metrics.RecordRouteOptimization(0.15)
	metrics.RecordProcessingTime(100 * time.Millisecond)
	metrics.UpdateSystemMetrics(1024*1024*1024, 75.5, 150)

	// Verify data is set
	assert.Equal(t, int64(1), metrics.anomalyDetections)
	assert.Equal(t, int64(1), metrics.loadPredictions)
	assert.Equal(t, int64(1), metrics.routeOptimizations)
	assert.Equal(t, time.Duration(100*time.Millisecond), metrics.processingTime)
	assert.Equal(t, int64(1024*1024*1024), metrics.memoryUsage)

	// Reset metrics
	metrics.Reset()

	// Verify all data is reset
	assert.Equal(t, int64(0), metrics.anomalyDetections)
	assert.Equal(t, int64(0), metrics.anomalyFalsePositives)
	assert.Equal(t, int64(0), metrics.anomalyFalseNegatives)
	assert.Equal(t, float64(0), metrics.anomalyAccuracy)
	assert.Equal(t, int64(0), metrics.loadPredictions)
	assert.Equal(t, int64(0), metrics.loadPredictionErrors)
	assert.Equal(t, float64(0), metrics.loadPredictionAccuracy)
	assert.Equal(t, int64(0), metrics.routeOptimizations)
	assert.Len(t, metrics.routeOptimizationGains, 0)
	assert.Equal(t, float64(0), metrics.averageOptimizationGain)
	assert.Equal(t, time.Duration(0), metrics.processingTime)
	assert.Equal(t, time.Duration(0), metrics.totalProcessingTime)
	assert.Equal(t, int64(0), metrics.processingCount)
	assert.Equal(t, time.Duration(0), metrics.averageProcessingTime)
	assert.Equal(t, int64(0), metrics.memoryUsage)
	assert.Equal(t, float64(0), metrics.cpuUsage)
	assert.Equal(t, int64(0), metrics.activeConnections)
}

func TestAIMetrics_ConcurrentAccess(t *testing.T) {
	metrics := NewAIMetrics()

	// Test concurrent access
	var wg sync.WaitGroup
	numGoroutines := 100

	// Start multiple goroutines that modify metrics
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			metrics.RecordAnomalyDetection(true, true)
			metrics.RecordLoadPrediction(100.0, 95.0)
			metrics.RecordRouteOptimization(0.15)
			metrics.RecordProcessingTime(100 * time.Millisecond)
			metrics.UpdateSystemMetrics(1024, 50.0, 10)
		}()
	}

	wg.Wait()

	// Verify final counts
	assert.Equal(t, int64(numGoroutines), metrics.anomalyDetections)
	assert.Equal(t, int64(numGoroutines), metrics.loadPredictions)
	assert.Equal(t, int64(numGoroutines), metrics.routeOptimizations)
	assert.Equal(t, int64(numGoroutines), metrics.processingCount)
	assert.Equal(t, int64(1024), metrics.memoryUsage)
	assert.Equal(t, float64(50.0), metrics.cpuUsage)
	assert.Equal(t, int64(10), metrics.activeConnections)
}

func TestAIMetrics_EdgeCases(t *testing.T) {
	metrics := NewAIMetrics()

	// Test with extreme values
	metrics.RecordAnomalyDetection(true, true)
	metrics.RecordLoadPrediction(0.0, 100.0)  // Division by zero case
	metrics.RecordRouteOptimization(1000.0)   // Very large gain
	metrics.RecordProcessingTime(time.Hour)   // Very long processing time
	metrics.UpdateSystemMetrics(-1, -1.0, -1) // Negative values

	// Verify metrics handle edge cases gracefully
	assert.Equal(t, int64(1), metrics.anomalyDetections)
	assert.Equal(t, int64(1), metrics.loadPredictions)
	assert.Equal(t, int64(1), metrics.loadPredictionErrors) // Should count as error due to division by zero
	assert.Equal(t, int64(1), metrics.routeOptimizations)
	assert.Equal(t, 1000.0, metrics.averageOptimizationGain)
	assert.Equal(t, time.Hour, metrics.processingTime)
	assert.Equal(t, int64(-1), metrics.memoryUsage)
	assert.Equal(t, float64(-1.0), metrics.cpuUsage)
	assert.Equal(t, int64(-1), metrics.activeConnections)
}

func TestAIMetrics_AccuracyCalculations(t *testing.T) {
	metrics := NewAIMetrics()

	// Test anomaly accuracy calculation
	metrics.RecordAnomalyDetection(true, true)   // Correct
	metrics.RecordAnomalyDetection(true, false)  // False positive
	metrics.RecordAnomalyDetection(false, true)  // Correct
	metrics.RecordAnomalyDetection(false, false) // False negative

	anomalyMetrics := metrics.GetAnomalyMetrics()
	expectedAccuracy := float64(2) / float64(4) // 2 correct out of 4 total
	assert.Equal(t, expectedAccuracy, anomalyMetrics["accuracy"])

	// Test load prediction accuracy calculation
	metrics.RecordLoadPrediction(100.0, 95.0)  // Accurate (5% error)
	metrics.RecordLoadPrediction(100.0, 80.0)  // Inaccurate (20% error)
	metrics.RecordLoadPrediction(100.0, 105.0) // Accurate (5% error)

	loadMetrics := metrics.GetLoadPredictionMetrics()
	expectedLoadAccuracy := float64(2) / float64(3) // 2 accurate out of 3 total
	assert.Equal(t, expectedLoadAccuracy, loadMetrics["accuracy"])
}

func TestAIMetrics_ProcessingTimeCalculations(t *testing.T) {
	metrics := NewAIMetrics()

	// Test average processing time calculation
	metrics.RecordProcessingTime(100 * time.Millisecond)
	metrics.RecordProcessingTime(200 * time.Millisecond)
	metrics.RecordProcessingTime(300 * time.Millisecond)

	processingMetrics := metrics.GetProcessingMetrics()
	expectedAverage := int64(200) // (100 + 200 + 300) / 3 = 200ms
	assert.Equal(t, expectedAverage, processingMetrics["average_time"])
	assert.Equal(t, int64(300), processingMetrics["current_time"]) // Last recorded time
	assert.Equal(t, int64(3), processingMetrics["total_count"])
}

func TestAIMetrics_EmptyState(t *testing.T) {
	metrics := NewAIMetrics()

	// Test getting metrics from empty state
	allMetrics := metrics.GetMetrics()
	assert.NotNil(t, allMetrics)

	// All values should be zero or empty
	anomalyMetrics := allMetrics["anomaly_detection"].(map[string]interface{})
	assert.Equal(t, int64(0), anomalyMetrics["total_detections"])
	assert.Equal(t, int64(0), anomalyMetrics["false_positives"])
	assert.Equal(t, int64(0), anomalyMetrics["false_negatives"])
	assert.Equal(t, float64(0), anomalyMetrics["accuracy"])

	loadMetrics := allMetrics["load_prediction"].(map[string]interface{})
	assert.Equal(t, int64(0), loadMetrics["total_predictions"])
	assert.Equal(t, int64(0), loadMetrics["prediction_errors"])
	assert.Equal(t, float64(0), loadMetrics["accuracy"])

	routeMetrics := allMetrics["route_optimization"].(map[string]interface{})
	assert.Equal(t, int64(0), routeMetrics["total_optimizations"])
	assert.Equal(t, float64(0), routeMetrics["average_gain"])

	processingMetrics := allMetrics["processing"].(map[string]interface{})
	assert.Equal(t, int64(0), processingMetrics["current_time"])
	assert.Equal(t, int64(0), processingMetrics["average_time"])
	assert.Equal(t, int64(0), processingMetrics["total_count"])

	systemMetrics := allMetrics["system"].(map[string]interface{})
	assert.Equal(t, int64(0), systemMetrics["memory_usage"])
	assert.Equal(t, float64(0), systemMetrics["cpu_usage"])
	assert.Equal(t, int64(0), systemMetrics["active_connections"])
}
