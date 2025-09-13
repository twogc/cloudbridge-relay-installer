package pipeline

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestRealFeatureExtractor_Process(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewRealFeatureExtractor(logger)

	testData := []DataPoint{
		{
			Timestamp: time.Now(),
			Source:    "test",
			Type:      "cpu",
			Data: map[string]interface{}{
				"cpu_usage": 75.5,
				"load_avg":  1.2,
				"temperature": 45.0,
			},
		},
		{
			Timestamp: time.Now().Add(-1 * time.Minute),
			Source:    "test",
			Type:      "memory",
			Data: map[string]interface{}{
				"memory_usage": 60.0,
				"free_memory":  40.0,
			},
		},
	}

	processedData, err := extractor.Process(testData)
	if err != nil {
		t.Fatalf("Failed to process data: %v", err)
	}

	if len(processedData) != len(testData) {
		t.Errorf("Expected %d processed points, got %d", len(testData), len(processedData))
	}

	// Check that features were extracted
	for _, point := range processedData {
		if point.Data == nil {
			t.Error("Expected features to be extracted")
			continue
		}

		// Check that features field exists
		features, ok := point.Data["features"]
		if !ok {
			t.Error("Expected features field to be present")
			continue
		}

		featuresMap, ok := features.(map[string]interface{})
		if !ok {
			t.Error("Expected features to be a map")
			continue
		}

		// Check for temporal features
		if _, ok := featuresMap["hour_of_day"]; !ok {
			t.Error("Expected hour_of_day feature")
		}

		if _, ok := featuresMap["day_of_week"]; !ok {
			t.Error("Expected day_of_week feature")
		}

		if _, ok := featuresMap["is_weekend"]; !ok {
			t.Error("Expected is_weekend feature")
		}
	}
}

func TestRealDataNormalizer_Process(t *testing.T) {
	logger := zap.NewNop()
	normalizer := NewRealDataNormalizer(logger)

	testData := []DataPoint{
		{
			Timestamp: time.Now(),
			Source:    "test",
			Type:      "cpu",
			Data: map[string]interface{}{
				"cpu_usage": 75.5,
				"load_avg":  1.2,
			},
		},
		{
			Timestamp: time.Now().Add(-1 * time.Minute),
			Source:    "test",
			Type:      "cpu",
			Data: map[string]interface{}{
				"cpu_usage": 85.0,
				"load_avg":  2.1,
			},
		},
		{
			Timestamp: time.Now().Add(-2 * time.Minute),
			Source:    "test",
			Type:      "cpu",
			Data: map[string]interface{}{
				"cpu_usage": 65.0,
				"load_avg":  0.8,
			},
		},
	}

	normalizedData, err := normalizer.Process(testData)
	if err != nil {
		t.Fatalf("Failed to normalize data: %v", err)
	}

	if len(normalizedData) != len(testData) {
		t.Errorf("Expected %d normalized points, got %d", len(testData), len(normalizedData))
	}

	// Check that normalization features were added
	for _, point := range normalizedData {
		if point.Data == nil {
			t.Error("Expected normalized data")
			continue
		}

		// Check for normalization features
		if _, ok := point.Data["cpu_usage_minmax"]; !ok {
			t.Error("Expected minmax normalization for cpu_usage")
		}

		if _, ok := point.Data["cpu_usage_zscore"]; !ok {
			t.Error("Expected zscore normalization for cpu_usage")
		}

		if _, ok := point.Data["cpu_usage_robust"]; !ok {
			t.Error("Expected robust normalization for cpu_usage")
		}
	}
}

func TestDataAggregator_Process(t *testing.T) {
	logger := zap.NewNop()
	aggregator := NewDataAggregator(logger, 1*time.Minute)

	// Create test data spanning 3 minutes
	baseTime := time.Now().Truncate(time.Minute)
	testData := []DataPoint{
		{
			Timestamp: baseTime,
			Source:    "test",
			Type:      "cpu",
			Data: map[string]interface{}{
				"cpu_usage": 75.5,
				"load_avg":  1.2,
			},
		},
		{
			Timestamp: baseTime.Add(30 * time.Second),
			Source:    "test",
			Type:      "cpu",
			Data: map[string]interface{}{
				"cpu_usage": 85.0,
				"load_avg":  2.1,
			},
		},
		{
			Timestamp: baseTime.Add(1 * time.Minute),
			Source:    "test",
			Type:      "cpu",
			Data: map[string]interface{}{
				"cpu_usage": 65.0,
				"load_avg":  0.8,
			},
		},
		{
			Timestamp: baseTime.Add(1*time.Minute + 30*time.Second),
			Source:    "test",
			Type:      "cpu",
			Data: map[string]interface{}{
				"cpu_usage": 70.0,
				"load_avg":  1.5,
			},
		},
	}

	aggregatedData, err := aggregator.Process(testData)
	if err != nil {
		t.Fatalf("Failed to aggregate data: %v", err)
	}

	// Should have 2 aggregated points (2 minutes of data)
	if len(aggregatedData) != 2 {
		t.Errorf("Expected 2 aggregated points, got %d", len(aggregatedData))
	}

	// Check that aggregation features were added
	for _, point := range aggregatedData {
		if point.Data == nil {
			t.Error("Expected aggregated data")
			continue
		}

		// Check for aggregation features
		if _, ok := point.Data["cpu_usage_count"]; !ok {
			t.Error("Expected count aggregation for cpu_usage")
		}

		if _, ok := point.Data["cpu_usage_mean"]; !ok {
			t.Error("Expected mean aggregation for cpu_usage")
		}

		if _, ok := point.Data["cpu_usage_min"]; !ok {
			t.Error("Expected min aggregation for cpu_usage")
		}

		if _, ok := point.Data["cpu_usage_max"]; !ok {
			t.Error("Expected max aggregation for cpu_usage")
		}
	}
}

func TestFeatureExtractor_StatisticalFeatures(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewRealFeatureExtractor(logger)

	// Test statistical calculations
	values := []float64{1.0, 2.0, 3.0, 4.0, 5.0}

	// Test mean
	mean := extractor.mean(values)
	expectedMean := 3.0
	if mean != expectedMean {
		t.Errorf("Expected mean %f, got %f", expectedMean, mean)
	}

	// Test median
	median := extractor.median(values)
	expectedMedian := 3.0
	if median != expectedMedian {
		t.Errorf("Expected median %f, got %f", expectedMedian, median)
	}

	// Test quantile
	q25 := extractor.quantile(values, 0.25)
	expectedQ25 := 2.0
	if q25 != expectedQ25 {
		t.Errorf("Expected 25th percentile %f, got %f", expectedQ25, q25)
	}
}

func TestDataNormalizer_Structure(t *testing.T) {
	logger := zap.NewNop()
	normalizer := NewRealDataNormalizer(logger)

	if normalizer == nil {
		t.Fatal("Expected normalizer to be created")
	}

	if normalizer.logger != logger {
		t.Error("Expected logger to be set")
	}
}
