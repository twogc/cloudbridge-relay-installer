package pipeline

import (
	"math"
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
				"cpu_usage":   75.5,
				"load_avg":    1.2,
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

func TestRealFeatureExtractor_ExtractNumericFeatures(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewRealFeatureExtractor(logger)

	testData := map[string]interface{}{
		"cpu_usage":    75.5,
		"memory_usage": 60.0,
		"load_avg":     1.2,
		"temperature":  45.0,
	}

	features := extractor.extractNumericFeatures(testData)

	// Check that numeric features are extracted
	expectedFeatures := []string{"cpu_usage", "memory_usage", "load_avg", "temperature"}
	for _, feature := range expectedFeatures {
		if _, ok := features[feature]; !ok {
			t.Errorf("Expected feature %s to be extracted", feature)
		}
	}

	// Check that log features are added for positive values
	if _, ok := features["cpu_usage_log"]; !ok {
		t.Error("Expected log feature for cpu_usage")
	}
}

func TestRealFeatureExtractor_ExtractStatisticalFeatures(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewRealFeatureExtractor(logger)

	testData := map[string]interface{}{
		"metric1": 10.0,
		"metric2": 20.0,
		"metric3": 30.0,
		"metric4": 40.0,
		"metric5": 50.0,
	}

	features := extractor.extractStatisticalFeatures(testData)

	// Check that statistical features are extracted
	expectedStats := []string{"mean", "median", "std", "min", "max", "range", "variance"}
	for _, stat := range expectedStats {
		if _, ok := features[stat]; !ok {
			t.Errorf("Expected statistical feature %s to be extracted", stat)
		}
	}

	// Check that mean is calculated correctly
	if mean, ok := features["mean"].(float64); ok {
		expectedMean := 30.0
		if math.Abs(mean-expectedMean) > 0.001 {
			t.Errorf("Expected mean %f, got %f", expectedMean, mean)
		}
	}
}

func TestRealFeatureExtractor_ExtractCategoricalFeatures(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewRealFeatureExtractor(logger)

	testData := map[string]interface{}{
		"hostname":     "server-01",
		"status":       "active",
		"is_healthy":   true,
		"empty_string": "",
	}

	features := extractor.extractCategoricalFeatures(testData)

	// Check string features
	if length, ok := features["hostname_length"].(float64); ok {
		if length != 9.0 {
			t.Errorf("Expected hostname_length 9.0, got %f", length)
		}
	}

	if isEmpty, ok := features["empty_string_is_empty"].(float64); ok {
		if isEmpty != 1.0 {
			t.Errorf("Expected empty_string_is_empty 1.0, got %f", isEmpty)
		}
	}

	// Check boolean features
	if isHealthy, ok := features["is_healthy"].(float64); ok {
		if isHealthy != 1.0 {
			t.Errorf("Expected is_healthy 1.0, got %f", isHealthy)
		}
	}
}

func TestRealFeatureExtractor_ExtractTemporalFeatures(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewRealFeatureExtractor(logger)

	testTime := time.Date(2023, 12, 25, 14, 30, 0, 0, time.UTC)
	testData := map[string]interface{}{
		"timestamp":     testTime.Format(time.RFC3339),
		"created_at":    testTime.Add(-1 * time.Hour).Format(time.RFC3339),
		"regular_field": "not_a_time",
	}

	features := extractor.extractTemporalFeatures(testData)

	// Check that temporal features are extracted
	if hour, ok := features["timestamp_hour"].(float64); ok {
		if hour != 14.0 {
			t.Errorf("Expected timestamp_hour 14.0, got %f", hour)
		}
	}

	if minute, ok := features["timestamp_minute"].(float64); ok {
		if minute != 30.0 {
			t.Errorf("Expected timestamp_minute 30.0, got %f", minute)
		}
	}

	if weekday, ok := features["timestamp_weekday"].(float64); ok {
		if weekday != 1.0 { // Monday
			t.Errorf("Expected timestamp_weekday 1.0, got %f", weekday)
		}
	}
}

func TestRealFeatureExtractor_StatisticalCalculations(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewRealFeatureExtractor(logger)

	// Test with known values
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

	// Test standard deviation
	std := extractor.std(values)
	expectedStd := math.Sqrt(2.0) // sqrt of variance
	if math.Abs(std-expectedStd) > 0.001 {
		t.Errorf("Expected std %f, got %f", expectedStd, std)
	}

	// Test min/max
	min := extractor.min(values)
	if min != 1.0 {
		t.Errorf("Expected min 1.0, got %f", min)
	}

	max := extractor.max(values)
	if max != 5.0 {
		t.Errorf("Expected max 5.0, got %f", max)
	}

	// Test variance
	variance := extractor.variance(values)
	expectedVariance := 2.0
	if math.Abs(variance-expectedVariance) > 0.001 {
		t.Errorf("Expected variance %f, got %f", expectedVariance, variance)
	}
}

func TestRealFeatureExtractor_QuantileCalculation(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewRealFeatureExtractor(logger)

	values := []float64{1.0, 2.0, 3.0, 4.0, 5.0}

	// Test 25th percentile
	q25 := extractor.quantile(values, 0.25)
	expectedQ25 := 2.0
	if q25 != expectedQ25 {
		t.Errorf("Expected 25th percentile %f, got %f", expectedQ25, q25)
	}

	// Test 75th percentile
	q75 := extractor.quantile(values, 0.75)
	expectedQ75 := 4.0
	if q75 != expectedQ75 {
		t.Errorf("Expected 75th percentile %f, got %f", expectedQ75, q75)
	}

	// Test median (50th percentile)
	q50 := extractor.quantile(values, 0.5)
	expectedQ50 := 3.0
	if q50 != expectedQ50 {
		t.Errorf("Expected 50th percentile %f, got %f", expectedQ50, q50)
	}
}

func TestRealFeatureExtractor_SkewnessAndKurtosis(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewRealFeatureExtractor(logger)

	// Test with symmetric data (should have skewness close to 0)
	symmetricValues := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	skewness := extractor.skewness(symmetricValues)
	if math.Abs(skewness) > 0.1 {
		t.Errorf("Expected skewness close to 0 for symmetric data, got %f", skewness)
	}

	// Test with asymmetric data
	asymmetricValues := []float64{1.0, 1.0, 1.0, 5.0, 10.0}
	skewness = extractor.skewness(asymmetricValues)
	if skewness <= 0 {
		t.Errorf("Expected positive skewness for right-skewed data, got %f", skewness)
	}

	// Test kurtosis
	kurtosis := extractor.kurtosis(symmetricValues)
	// Kurtosis should be reasonable (not NaN or infinite)
	if math.IsNaN(kurtosis) || math.IsInf(kurtosis, 0) {
		t.Errorf("Expected finite kurtosis, got %f", kurtosis)
	}
}

func TestRealFeatureExtractor_HelperFunctions(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewRealFeatureExtractor(logger)

	// Test hasNumbers
	if !extractor.hasNumbers("abc123def") {
		t.Error("Expected hasNumbers to return true for string with numbers")
	}
	if extractor.hasNumbers("abcdef") {
		t.Error("Expected hasNumbers to return false for string without numbers")
	}

	// Test hasSpecialChars
	if !extractor.hasSpecialChars("abc@def") {
		t.Error("Expected hasSpecialChars to return true for string with special chars")
	}
	if extractor.hasSpecialChars("abcdef123") {
		t.Error("Expected hasSpecialChars to return false for alphanumeric string")
	}

	// Test isWeekend
	weekend := time.Date(2023, 12, 24, 0, 0, 0, 0, time.UTC) // Sunday
	weekday := time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC) // Monday

	if extractor.isWeekend(weekend) != 1.0 {
		t.Error("Expected isWeekend to return 1.0 for weekend")
	}
	if extractor.isWeekend(weekday) != 0.0 {
		t.Error("Expected isWeekend to return 0.0 for weekday")
	}

	// Test boolToFloat
	if extractor.boolToFloat(true) != 1.0 {
		t.Error("Expected boolToFloat(true) to return 1.0")
	}
	if extractor.boolToFloat(false) != 0.0 {
		t.Error("Expected boolToFloat(false) to return 0.0")
	}
}

func TestRealDataNormalizer_UpdateStats(t *testing.T) {
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

	normalizer.updateStats(testData)

	// Check that stats were updated
	normalizer.mutex.RLock()
	defer normalizer.mutex.RUnlock()

	if len(normalizer.stats) == 0 {
		t.Error("Expected stats to be updated")
	}

	// Check specific stats
	if cpuStats, ok := normalizer.stats["cpu_usage"]; ok {
		if cpuStats.Min != 65.0 {
			t.Errorf("Expected cpu_usage min 65.0, got %f", cpuStats.Min)
		}
		if cpuStats.Max != 85.0 {
			t.Errorf("Expected cpu_usage max 85.0, got %f", cpuStats.Max)
		}
		if cpuStats.Count != 3 {
			t.Errorf("Expected cpu_usage count 3, got %d", cpuStats.Count)
		}
	}
}

func TestRealDataNormalizer_ToFloat64(t *testing.T) {
	logger := zap.NewNop()
	normalizer := NewRealDataNormalizer(logger)

	// Test float64
	if val, ok := normalizer.toFloat64(42.5); !ok || val != 42.5 {
		t.Error("Expected toFloat64 to handle float64")
	}

	// Test int
	if val, ok := normalizer.toFloat64(42); !ok || val != 42.0 {
		t.Error("Expected toFloat64 to handle int")
	}

	// Test int64
	if val, ok := normalizer.toFloat64(int64(42)); !ok || val != 42.0 {
		t.Error("Expected toFloat64 to handle int64")
	}

	// Test string
	if val, ok := normalizer.toFloat64("42.5"); !ok || val != 42.5 {
		t.Error("Expected toFloat64 to handle string")
	}

	// Test invalid string
	if _, ok := normalizer.toFloat64("invalid"); ok {
		t.Error("Expected toFloat64 to return false for invalid string")
	}

	// Test unsupported type
	if _, ok := normalizer.toFloat64(true); ok {
		t.Error("Expected toFloat64 to return false for unsupported type")
	}
}

func TestDataAggregator_ToFloat64(t *testing.T) {
	logger := zap.NewNop()
	aggregator := NewDataAggregator(logger, 1*time.Minute)

	// Test float64
	if val, ok := aggregator.toFloat64(42.5); !ok || val != 42.5 {
		t.Error("Expected toFloat64 to handle float64")
	}

	// Test int
	if val, ok := aggregator.toFloat64(42); !ok || val != 42.0 {
		t.Error("Expected toFloat64 to handle int")
	}

	// Test int64
	if val, ok := aggregator.toFloat64(int64(42)); !ok || val != 42.0 {
		t.Error("Expected toFloat64 to handle int64")
	}

	// Test string
	if val, ok := aggregator.toFloat64("42.5"); !ok || val != 42.5 {
		t.Error("Expected toFloat64 to handle string")
	}

	// Test invalid string
	if _, ok := aggregator.toFloat64("invalid"); ok {
		t.Error("Expected toFloat64 to return false for invalid string")
	}
}

func TestDataAggregator_EmptyData(t *testing.T) {
	logger := zap.NewNop()
	aggregator := NewDataAggregator(logger, 1*time.Minute)

	// Test with empty data
	emptyData := []DataPoint{}
	result, err := aggregator.Process(emptyData)
	if err != nil {
		t.Fatalf("Failed to process empty data: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected empty result for empty input, got %d points", len(result))
	}
}

func TestDataAggregator_SingleWindow(t *testing.T) {
	logger := zap.NewNop()
	aggregator := NewDataAggregator(logger, 1*time.Minute)

	// Test with data in single window
	baseTime := time.Now().Truncate(time.Minute)
	testData := []DataPoint{
		{
			Timestamp: baseTime,
			Source:    "test",
			Type:      "cpu",
			Data: map[string]interface{}{
				"cpu_usage": 75.5,
			},
		},
		{
			Timestamp: baseTime.Add(30 * time.Second),
			Source:    "test",
			Type:      "cpu",
			Data: map[string]interface{}{
				"cpu_usage": 85.0,
			},
		},
	}

	result, err := aggregator.Process(testData)
	if err != nil {
		t.Fatalf("Failed to process single window data: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 aggregated point, got %d", len(result))
	}

	// Check aggregation results
	point := result[0]
	if count, ok := point.Data["cpu_usage_count"].(float64); ok {
		if count != 2.0 {
			t.Errorf("Expected cpu_usage_count 2.0, got %f", count)
		}
	}

	if mean, ok := point.Data["cpu_usage_mean"].(float64); ok {
		expectedMean := 80.25
		if math.Abs(mean-expectedMean) > 0.001 {
			t.Errorf("Expected cpu_usage_mean %f, got %f", expectedMean, mean)
		}
	}
}

func TestDataAggregator_StatisticalFunctions(t *testing.T) {
	logger := zap.NewNop()
	aggregator := NewDataAggregator(logger, 1*time.Minute)

	values := []float64{1.0, 2.0, 3.0, 4.0, 5.0}

	// Test sum
	sum := aggregator.sum(values)
	expectedSum := 15.0
	if sum != expectedSum {
		t.Errorf("Expected sum %f, got %f", expectedSum, sum)
	}

	// Test mean
	mean := aggregator.mean(values)
	expectedMean := 3.0
	if mean != expectedMean {
		t.Errorf("Expected mean %f, got %f", expectedMean, mean)
	}

	// Test min
	min := aggregator.min(values)
	if min != 1.0 {
		t.Errorf("Expected min 1.0, got %f", min)
	}

	// Test max
	max := aggregator.max(values)
	if max != 5.0 {
		t.Errorf("Expected max 5.0, got %f", max)
	}

	// Test std
	std := aggregator.std(values)
	expectedStd := math.Sqrt(2.0)
	if math.Abs(std-expectedStd) > 0.001 {
		t.Errorf("Expected std %f, got %f", expectedStd, std)
	}

	// Test median
	median := aggregator.median(values)
	if median != 3.0 {
		t.Errorf("Expected median 3.0, got %f", median)
	}
}
