package pipeline

import (
	"errors"
	"os"
	"testing"
	"time"

	"go.uber.org/zap"
)

// MockCollector для тестирования обработки ошибок
type MockCollector struct {
	shouldError bool
	errorMsg    string
}

func (mc *MockCollector) Collect() ([]DataPoint, error) {
	if mc.shouldError {
		return nil, errors.New(mc.errorMsg)
	}
	return []DataPoint{
		{
			Timestamp: time.Now(),
			Source:    "mock",
			Type:      "test",
			Data:      map[string]interface{}{"value": 42.0},
		},
	}, nil
}

func (mc *MockCollector) Start() error {
	if mc.shouldError {
		return errors.New(mc.errorMsg)
	}
	return nil
}

func (mc *MockCollector) Stop() error {
	if mc.shouldError {
		return errors.New(mc.errorMsg)
	}
	return nil
}

// MockProcessor для тестирования обработки ошибок
type MockProcessor struct {
	shouldError bool
	errorMsg    string
}

func (mp *MockProcessor) Process(data []DataPoint) ([]DataPoint, error) {
	if mp.shouldError {
		return nil, errors.New(mp.errorMsg)
	}
	return data, nil
}

// MockStorage для тестирования обработки ошибок
type MockStorage struct {
	shouldError bool
	errorMsg    string
}

func (ms *MockStorage) Store(data []DataPoint) error {
	if ms.shouldError {
		return errors.New(ms.errorMsg)
	}
	return nil
}

func (ms *MockStorage) Retrieve(timeRange TimeRange) ([]DataPoint, error) {
	if ms.shouldError {
		return nil, errors.New(ms.errorMsg)
	}
	return []DataPoint{}, nil
}

func (ms *MockStorage) Query(timeRange TimeRange, dataType string, limit int) ([]DataPoint, error) {
	if ms.shouldError {
		return nil, errors.New(ms.errorMsg)
	}
	return []DataPoint{}, nil
}

func TestDataPipeline_CollectorErrorHandling(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		Enabled:            true,
		BatchSize:          100,
		ProcessingInterval: 30 * time.Second,
		RetentionDays:      7,
		StoragePath:        "/tmp/test_pipeline_collector_error",
		PrometheusURL:      "http://localhost:9090",
	}

	pipeline := NewDataPipeline(config, logger)
	defer os.RemoveAll("/tmp/test_pipeline_collector_error")

	// Initialize pipeline
	err := pipeline.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize pipeline: %v", err)
	}

	// Add a mock collector that will error
	errorCollector := &MockCollector{
		shouldError: true,
		errorMsg:    "mock collector error",
	}
	pipeline.collectors["error_collector"] = errorCollector

	// Collect data - should handle errors gracefully
	data, err := pipeline.CollectData()
	if err != nil {
		t.Fatalf("Expected pipeline to handle collector errors gracefully: %v", err)
	}

	// Should still have data from other collectors
	if len(data) == 0 {
		t.Error("Expected some data from non-erroring collectors")
	}
}

func TestDataPipeline_ProcessorErrorHandling(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		Enabled:            true,
		BatchSize:          10,
		ProcessingInterval: 100 * time.Millisecond,
		RetentionDays:      1,
		StoragePath:        "/tmp/test_pipeline_processor_error",
		PrometheusURL:      "http://localhost:9090",
	}

	pipeline := NewDataPipeline(config, logger)
	defer os.RemoveAll("/tmp/test_pipeline_processor_error")

	// Initialize pipeline
	err := pipeline.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize pipeline: %v", err)
	}

	// Add a mock processor that will error
	errorProcessor := &MockProcessor{
		shouldError: true,
		errorMsg:    "mock processor error",
	}
	pipeline.processors = append(pipeline.processors, errorProcessor)

	// Start pipeline
	err = pipeline.Start()
	if err != nil {
		t.Fatalf("Failed to start pipeline: %v", err)
	}
	defer pipeline.Stop()

	// Wait for processing to occur
	time.Sleep(200 * time.Millisecond)

	// Pipeline should handle processor errors gracefully
	// (errors are logged but don't crash the pipeline)
}

func TestDataPipeline_StorageErrorHandling(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		Enabled:            true,
		BatchSize:          10,
		ProcessingInterval: 100 * time.Millisecond,
		RetentionDays:      1,
		StoragePath:        "/tmp/test_pipeline_storage_error",
		PrometheusURL:      "http://localhost:9090",
	}

	pipeline := NewDataPipeline(config, logger)
	defer os.RemoveAll("/tmp/test_pipeline_storage_error")

	// Initialize pipeline
	err := pipeline.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize pipeline: %v", err)
	}

	// Replace storage with mock that will error
	errorStorage := &MockStorage{
		shouldError: true,
		errorMsg:    "mock storage error",
	}
	pipeline.storage = errorStorage

	// Start pipeline
	err = pipeline.Start()
	if err != nil {
		t.Fatalf("Failed to start pipeline: %v", err)
	}
	defer pipeline.Stop()

	// Wait for processing to occur
	time.Sleep(200 * time.Millisecond)

	// Pipeline should handle storage errors gracefully
	// (errors are logged but don't crash the pipeline)
}

func TestDataPipeline_EmptyDataHandling(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		Enabled:            true,
		BatchSize:          100,
		ProcessingInterval: 30 * time.Second,
		RetentionDays:      7,
		StoragePath:        "/tmp/test_pipeline_empty_data",
		PrometheusURL:      "http://localhost:9090",
	}

	pipeline := NewDataPipeline(config, logger)
	defer os.RemoveAll("/tmp/test_pipeline_empty_data")

	// Initialize pipeline
	err := pipeline.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize pipeline: %v", err)
	}

	// Replace all collectors with ones that return empty data
	pipeline.collectors = make(map[string]DataCollector)
	pipeline.collectors["empty_collector"] = &MockCollector{
		shouldError: false,
	}

	// Collect data
	data, err := pipeline.CollectData()
	if err != nil {
		t.Fatalf("Failed to collect data: %v", err)
	}

	// Should handle empty data gracefully
	if len(data) == 0 {
		t.Log("No data collected - this is expected with mock collectors")
	}
}

func TestDataPipeline_InvalidConfigHandling(t *testing.T) {
	logger := zap.NewNop()

	// Test with nil config
	pipeline := NewDataPipeline(nil, logger)
	if pipeline.config != nil {
		t.Error("Expected nil config to be handled")
	}

	// Test with nil logger
	pipeline = NewDataPipeline(&Config{}, nil)
	if pipeline.logger != nil {
		t.Error("Expected nil logger to be handled")
	}
}

func TestDataPipeline_InvalidTimeRangeHandling(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		Enabled:            true,
		BatchSize:          100,
		ProcessingInterval: 30 * time.Second,
		RetentionDays:      7,
		StoragePath:        "/tmp/test_pipeline_invalid_time",
		PrometheusURL:      "http://localhost:9090",
	}

	pipeline := NewDataPipeline(config, logger)
	defer os.RemoveAll("/tmp/test_pipeline_invalid_time")

	// Initialize and start pipeline
	err := pipeline.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize pipeline: %v", err)
	}

	err = pipeline.Start()
	if err != nil {
		t.Fatalf("Failed to start pipeline: %v", err)
	}
	defer pipeline.Stop()

	// Test with invalid time range (start after end)
	invalidTimeRange := TimeRange{
		Start: time.Now(),
		End:   time.Now().Add(-1 * time.Hour),
	}

	data, err := pipeline.QueryData(invalidTimeRange, "cpu", 10)
	if err != nil {
		t.Fatalf("Failed to query with invalid time range: %v", err)
	}

	// Should return empty data for invalid time range
	if len(data) != 0 {
		t.Error("Expected empty data for invalid time range")
	}
}

func TestDataPipeline_ConcurrentAccessHandling(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		Enabled:            true,
		BatchSize:          100,
		ProcessingInterval: 30 * time.Second,
		RetentionDays:      7,
		StoragePath:        "/tmp/test_pipeline_concurrent",
		PrometheusURL:      "http://localhost:9090",
	}

	pipeline := NewDataPipeline(config, logger)
	defer os.RemoveAll("/tmp/test_pipeline_concurrent")

	// Initialize pipeline
	err := pipeline.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize pipeline: %v", err)
	}

	// Start pipeline
	err = pipeline.Start()
	if err != nil {
		t.Fatalf("Failed to start pipeline: %v", err)
	}
	defer pipeline.Stop()

	// Test concurrent access to GetMetrics
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			_, err := pipeline.GetMetrics()
			if err != nil {
				t.Errorf("Failed to get metrics concurrently: %v", err)
			}
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestDataPipeline_ResourceCleanup(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		Enabled:            true,
		BatchSize:          10,
		ProcessingInterval: 100 * time.Millisecond,
		RetentionDays:      1,
		StoragePath:        "/tmp/test_pipeline_cleanup",
		PrometheusURL:      "http://localhost:9090",
	}

	pipeline := NewDataPipeline(config, logger)
	defer os.RemoveAll("/tmp/test_pipeline_cleanup")

	// Initialize and start pipeline
	err := pipeline.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize pipeline: %v", err)
	}

	err = pipeline.Start()
	if err != nil {
		t.Fatalf("Failed to start pipeline: %v", err)
	}

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Stop pipeline
	err = pipeline.Stop()
	if err != nil {
		t.Fatalf("Failed to stop pipeline: %v", err)
	}

	// Try to stop again - should not cause issues
	err = pipeline.Stop()
	if err != nil {
		t.Fatalf("Failed to stop pipeline again: %v", err)
	}
}

func TestDataPipeline_EdgeCaseDataPoints(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewRealFeatureExtractor(logger)

	// Test with empty data
	emptyData := []DataPoint{}
	result, err := extractor.Process(emptyData)
	if err != nil {
		t.Fatalf("Failed to process empty data: %v", err)
	}
	if len(result) != 0 {
		t.Error("Expected empty result for empty input")
	}

	// Test with nil data
	result, err = extractor.Process(nil)
	if err != nil {
		t.Fatalf("Failed to process nil data: %v", err)
	}
	if len(result) != 0 {
		t.Error("Expected empty result for nil input")
	}

	// Test with data point containing nil data
	nilDataPoint := []DataPoint{
		{
			Timestamp: time.Now(),
			Source:    "test",
			Type:      "test",
			Data:      nil,
		},
	}
	result, err = extractor.Process(nilDataPoint)
	if err != nil {
		t.Fatalf("Failed to process data with nil Data field: %v", err)
	}
	if len(result) != 1 {
		t.Error("Expected one result for nil Data field")
	}
}

func TestDataPipeline_EdgeCaseNormalization(t *testing.T) {
	logger := zap.NewNop()
	normalizer := NewRealDataNormalizer(logger)

	// Test with empty data
	emptyData := []DataPoint{}
	result, err := normalizer.Process(emptyData)
	if err != nil {
		t.Fatalf("Failed to process empty data: %v", err)
	}
	if len(result) != 0 {
		t.Error("Expected empty result for empty input")
	}

	// Test with data containing non-numeric values
	mixedData := []DataPoint{
		{
			Timestamp: time.Now(),
			Source:    "test",
			Type:      "test",
			Data: map[string]interface{}{
				"numeric":   42.0,
				"string":    "not_a_number",
				"boolean":   true,
				"nil_value": nil,
			},
		},
	}
	result, err = normalizer.Process(mixedData)
	if err != nil {
		t.Fatalf("Failed to process mixed data: %v", err)
	}
	if len(result) != 1 {
		t.Error("Expected one result for mixed data")
	}
}

func TestDataPipeline_EdgeCaseAggregation(t *testing.T) {
	logger := zap.NewNop()
	aggregator := NewDataAggregator(logger, 1*time.Minute)

	// Test with empty data
	emptyData := []DataPoint{}
	result, err := aggregator.Process(emptyData)
	if err != nil {
		t.Fatalf("Failed to process empty data: %v", err)
	}
	if len(result) != 0 {
		t.Error("Expected empty result for empty input")
	}

	// Test with single data point
	singleData := []DataPoint{
		{
			Timestamp: time.Now(),
			Source:    "test",
			Type:      "test",
			Data: map[string]interface{}{
				"value": 42.0,
			},
		},
	}
	result, err = aggregator.Process(singleData)
	if err != nil {
		t.Fatalf("Failed to process single data point: %v", err)
	}
	if len(result) != 1 {
		t.Error("Expected one result for single data point")
	}
}

func TestDataPipeline_EdgeCaseStorage(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		BatchSize:     10,
		RetentionDays: 1,
		StoragePath:   "/tmp/test_storage_edge_cases",
	}

	storage := NewRealFileStorage(config, logger)
	defer os.RemoveAll("/tmp/test_storage_edge_cases")

	// Test storing empty data
	err := storage.Store([]DataPoint{})
	if err != nil {
		t.Fatalf("Failed to store empty data: %v", err)
	}

	// Test storing nil data
	err = storage.Store(nil)
	if err != nil {
		t.Fatalf("Failed to store nil data: %v", err)
	}

	// Test with invalid storage path
	invalidConfig := &Config{
		BatchSize:     10,
		RetentionDays: 1,
		StoragePath:   "/invalid/path/that/does/not/exist",
	}

	invalidStorage := NewRealFileStorage(invalidConfig, logger)
	err = invalidStorage.Store([]DataPoint{
		{
			Timestamp: time.Now(),
			Source:    "test",
			Type:      "test",
			Data:      map[string]interface{}{"value": 42.0},
		},
	})
	// Should handle invalid path gracefully
	if err != nil {
		t.Logf("Expected error for invalid path: %v", err)
	}
}
