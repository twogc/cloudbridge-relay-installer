package api

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewAIMetrics(t *testing.T) {
	metrics := NewAIMetrics()

	assert.NotNil(t, metrics)

	// Test initial values through GetStats
	stats := metrics.GetStats()
	assert.Equal(t, int64(0), stats["requests_processed"])
	assert.Equal(t, int64(0), stats["anomalies_detected"])
	assert.Equal(t, int64(0), stats["predictions_made"])
	assert.Equal(t, int64(0), stats["cache_hits"])
	assert.Equal(t, int64(0), stats["cache_misses"])
	assert.Equal(t, int64(0), stats["errors"])
	assert.Equal(t, "0s", stats["total_processing_time"])
}

func TestAIMetrics_IncrementRequestsProcessed(t *testing.T) {
	metrics := NewAIMetrics()

	metrics.IncrementRequestsProcessed()
	stats := metrics.GetStats()
	assert.Equal(t, int64(1), stats["requests_processed"])

	metrics.IncrementRequestsProcessed()
	stats = metrics.GetStats()
	assert.Equal(t, int64(2), stats["requests_processed"])
}

func TestAIMetrics_IncrementAnomaliesDetected(t *testing.T) {
	metrics := NewAIMetrics()

	metrics.IncrementAnomaliesDetected()
	stats := metrics.GetStats()
	assert.Equal(t, int64(1), stats["anomalies_detected"])

	metrics.IncrementAnomaliesDetected()
	stats = metrics.GetStats()
	assert.Equal(t, int64(2), stats["anomalies_detected"])
}

func TestAIMetrics_IncrementPredictionsMade(t *testing.T) {
	metrics := NewAIMetrics()

	metrics.IncrementPredictionsMade()
	stats := metrics.GetStats()
	assert.Equal(t, int64(1), stats["predictions_made"])

	metrics.IncrementPredictionsMade()
	stats = metrics.GetStats()
	assert.Equal(t, int64(2), stats["predictions_made"])
}

func TestAIMetrics_IncrementCacheHits(t *testing.T) {
	metrics := NewAIMetrics()

	metrics.IncrementCacheHits()
	stats := metrics.GetStats()
	assert.Equal(t, int64(1), stats["cache_hits"])

	metrics.IncrementCacheHits()
	stats = metrics.GetStats()
	assert.Equal(t, int64(2), stats["cache_hits"])
}

func TestAIMetrics_IncrementCacheMisses(t *testing.T) {
	metrics := NewAIMetrics()

	metrics.IncrementCacheMisses()
	stats := metrics.GetStats()
	assert.Equal(t, int64(1), stats["cache_misses"])

	metrics.IncrementCacheMisses()
	stats = metrics.GetStats()
	assert.Equal(t, int64(2), stats["cache_misses"])
}

func TestAIMetrics_IncrementErrors(t *testing.T) {
	metrics := NewAIMetrics()

	metrics.IncrementErrors()
	stats := metrics.GetStats()
	assert.Equal(t, int64(1), stats["errors"])

	metrics.IncrementErrors()
	stats = metrics.GetStats()
	assert.Equal(t, int64(2), stats["errors"])
}

func TestAIMetrics_RecordProcessingTime(t *testing.T) {
	metrics := NewAIMetrics()

	duration := 100 * time.Millisecond
	metrics.RecordProcessingTime(duration)
	stats := metrics.GetStats()
	assert.Equal(t, "100ms", stats["total_processing_time"])

	metrics.RecordProcessingTime(duration)
	stats = metrics.GetStats()
	assert.Equal(t, "200ms", stats["total_processing_time"])
}

func TestAIMetrics_GetStats(t *testing.T) {
	metrics := NewAIMetrics()

	// Set some test data using increment methods
	for i := int64(0); i < 100; i++ {
		metrics.IncrementRequestsProcessed()
	}
	for i := int64(0); i < 5; i++ {
		metrics.IncrementAnomaliesDetected()
	}
	for i := int64(0); i < 50; i++ {
		metrics.IncrementPredictionsMade()
	}
	for i := int64(0); i < 80; i++ {
		metrics.IncrementCacheHits()
	}
	for i := int64(0); i < 20; i++ {
		metrics.IncrementCacheMisses()
	}
	for i := int64(0); i < 2; i++ {
		metrics.IncrementErrors()
	}
	metrics.RecordProcessingTime(5 * time.Second)

	stats := metrics.GetStats()

	assert.Equal(t, int64(100), stats["requests_processed"])
	assert.Equal(t, int64(5), stats["anomalies_detected"])
	assert.Equal(t, int64(50), stats["predictions_made"])
	assert.Equal(t, int64(80), stats["cache_hits"])
	assert.Equal(t, int64(20), stats["cache_misses"])
	assert.Equal(t, int64(2), stats["errors"])
	assert.Equal(t, "5s", stats["total_processing_time"])

	// Test calculated fields
	assert.Equal(t, float64(80), stats["cache_hit_rate"])
	assert.Equal(t, "50ms", stats["average_processing_time"])
}

func TestAIMetrics_GetStats_ZeroValues(t *testing.T) {
	metrics := NewAIMetrics()

	stats := metrics.GetStats()

	assert.Equal(t, int64(0), stats["requests_processed"])
	assert.Equal(t, int64(0), stats["anomalies_detected"])
	assert.Equal(t, int64(0), stats["predictions_made"])
	assert.Equal(t, int64(0), stats["cache_hits"])
	assert.Equal(t, int64(0), stats["cache_misses"])
	assert.Equal(t, int64(0), stats["errors"])
	assert.Equal(t, "0s", stats["total_processing_time"])

	// Test calculated fields with zero values
	assert.Equal(t, float64(0), stats["cache_hit_rate"])
	assert.Equal(t, "0s", stats["average_processing_time"])
}

func TestAIMetrics_ConcurrentAccess(t *testing.T) {
	metrics := NewAIMetrics()

	// Test concurrent access
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			metrics.IncrementRequestsProcessed()
			metrics.IncrementAnomaliesDetected()
			metrics.IncrementPredictionsMade()
			metrics.IncrementCacheHits()
			metrics.IncrementCacheMisses()
			metrics.IncrementErrors()
			metrics.RecordProcessingTime(10 * time.Millisecond)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify final values
	stats := metrics.GetStats()
	assert.Equal(t, int64(10), stats["requests_processed"])
	assert.Equal(t, int64(10), stats["anomalies_detected"])
	assert.Equal(t, int64(10), stats["predictions_made"])
	assert.Equal(t, int64(10), stats["cache_hits"])
	assert.Equal(t, int64(10), stats["cache_misses"])
	assert.Equal(t, int64(10), stats["errors"])
	assert.Equal(t, "100ms", stats["total_processing_time"])
}

func TestAIMetrics_Reset(t *testing.T) {
	metrics := NewAIMetrics()

	// Set some test data using increment methods
	for i := int64(0); i < 100; i++ {
		metrics.IncrementRequestsProcessed()
	}
	for i := int64(0); i < 5; i++ {
		metrics.IncrementAnomaliesDetected()
	}
	for i := int64(0); i < 50; i++ {
		metrics.IncrementPredictionsMade()
	}
	for i := int64(0); i < 80; i++ {
		metrics.IncrementCacheHits()
	}
	for i := int64(0); i < 20; i++ {
		metrics.IncrementCacheMisses()
	}
	for i := int64(0); i < 2; i++ {
		metrics.IncrementErrors()
	}
	metrics.RecordProcessingTime(5 * time.Second)

	// Reset metrics
	metrics.Reset()

	// Verify all values are reset
	stats := metrics.GetStats()
	assert.Equal(t, int64(0), stats["requests_processed"])
	assert.Equal(t, int64(0), stats["anomalies_detected"])
	assert.Equal(t, int64(0), stats["predictions_made"])
	assert.Equal(t, int64(0), stats["cache_hits"])
	assert.Equal(t, int64(0), stats["cache_misses"])
	assert.Equal(t, int64(0), stats["errors"])
	assert.Equal(t, "0s", stats["total_processing_time"])
}

func TestAIMetrics_GetStats_EdgeCases(t *testing.T) {
	metrics := NewAIMetrics()

	// Test with very small values
	metrics.IncrementRequestsProcessed()
	metrics.IncrementAnomaliesDetected()
	metrics.IncrementPredictionsMade()
	metrics.IncrementCacheHits()
	metrics.IncrementCacheMisses()
	metrics.IncrementErrors()
	metrics.RecordProcessingTime(1 * time.Nanosecond)

	stats := metrics.GetStats()

	assert.Equal(t, float64(50), stats["cache_hit_rate"]) // 1 hit out of 2 total (1 hit + 1 miss) = 50%
	assert.Equal(t, "1ns", stats["average_processing_time"])
}

func TestAIMetrics_GetStats_LargeValues(t *testing.T) {
	metrics := NewAIMetrics()

	// Test with large values using increment methods
	for i := int64(0); i < 1000000; i++ {
		metrics.IncrementRequestsProcessed()
	}
	for i := int64(0); i < 1000; i++ {
		metrics.IncrementAnomaliesDetected()
	}
	for i := int64(0); i < 500000; i++ {
		metrics.IncrementPredictionsMade()
	}
	for i := int64(0); i < 800000; i++ {
		metrics.IncrementCacheHits()
	}
	for i := int64(0); i < 200000; i++ {
		metrics.IncrementCacheMisses()
	}
	for i := int64(0); i < 100; i++ {
		metrics.IncrementErrors()
	}
	metrics.RecordProcessingTime(1000 * time.Second)

	stats := metrics.GetStats()

	assert.Equal(t, float64(80), stats["cache_hit_rate"]) // 800000 hits out of 1000000 total = 80%
	assert.Equal(t, "1ms", stats["average_processing_time"])
}
