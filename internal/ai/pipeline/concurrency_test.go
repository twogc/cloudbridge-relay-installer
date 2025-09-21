package pipeline

import (
	"os"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestDataPipeline_ConcurrentStartStop(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		Enabled:            true,
		BatchSize:          100,
		ProcessingInterval: 30 * time.Second,
		RetentionDays:      7,
		StoragePath:        "/tmp/test_pipeline_concurrent_start_stop",
		PrometheusURL:      "http://localhost:9090",
	}

	pipeline := NewDataPipeline(config, logger)
	defer os.RemoveAll("/tmp/test_pipeline_concurrent_start_stop")

	// Initialize pipeline
	err := pipeline.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize pipeline: %v", err)
	}

	var wg sync.WaitGroup
	errors := make(chan error, 10)

	// Start multiple goroutines trying to start the pipeline
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := pipeline.Start()
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		if err != nil {
			t.Logf("Expected some start errors due to concurrency: %v", err)
		}
	}

	// Stop pipeline once
	err = pipeline.Stop()
	if err != nil {
		t.Fatalf("Failed to stop pipeline: %v", err)
	}
}

func TestDataPipeline_ConcurrentDataCollection(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		Enabled:            true,
		BatchSize:          100,
		ProcessingInterval: 30 * time.Second,
		RetentionDays:      7,
		StoragePath:        "/tmp/test_pipeline_concurrent_collection",
		PrometheusURL:      "http://localhost:9090",
	}

	pipeline := NewDataPipeline(config, logger)
	defer os.RemoveAll("/tmp/test_pipeline_concurrent_collection")

	// Initialize pipeline
	err := pipeline.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize pipeline: %v", err)
	}

	var wg sync.WaitGroup
	results := make(chan []DataPoint, 10)

	// Start multiple goroutines collecting data concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			data, err := pipeline.CollectData()
			if err != nil {
				t.Errorf("Failed to collect data: %v", err)
				return
			}
			results <- data
		}()
	}

	wg.Wait()
	close(results)

	// Check that all collections succeeded
	count := 0
	for data := range results {
		if len(data) == 0 {
			t.Error("Expected some data to be collected")
		}
		count++
	}

	if count != 10 {
		t.Errorf("Expected 10 data collections, got %d", count)
	}
}

func TestDataPipeline_ConcurrentMetricsAccess(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		Enabled:            true,
		BatchSize:          100,
		ProcessingInterval: 30 * time.Second,
		RetentionDays:      7,
		StoragePath:        "/tmp/test_pipeline_concurrent_metrics",
		PrometheusURL:      "http://localhost:9090",
	}

	pipeline := NewDataPipeline(config, logger)
	defer os.RemoveAll("/tmp/test_pipeline_concurrent_metrics")

	// Initialize pipeline
	err := pipeline.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize pipeline: %v", err)
	}

	var wg sync.WaitGroup
	results := make(chan map[string]interface{}, 20)

	// Start multiple goroutines accessing metrics concurrently
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			metrics, err := pipeline.GetMetrics()
			if err != nil {
				t.Errorf("Failed to get metrics: %v", err)
				return
			}
			results <- metrics
		}()
	}

	wg.Wait()
	close(results)

	// Check that all metric accesses succeeded
	count := 0
	for metrics := range results {
		if metrics == nil {
			t.Error("Expected non-nil metrics")
		}
		count++
	}

	if count != 20 {
		t.Errorf("Expected 20 metric accesses, got %d", count)
	}
}

func TestDataPipeline_ConcurrentQueryAccess(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		Enabled:            true,
		BatchSize:          100,
		ProcessingInterval: 30 * time.Second,
		RetentionDays:      7,
		StoragePath:        "/tmp/test_pipeline_concurrent_query",
		PrometheusURL:      "http://localhost:9090",
	}

	pipeline := NewDataPipeline(config, logger)
	defer os.RemoveAll("/tmp/test_pipeline_concurrent_query")

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

	var wg sync.WaitGroup
	results := make(chan []DataPoint, 15)

	// Start multiple goroutines querying data concurrently
	for i := 0; i < 15; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			timeRange := TimeRange{
				Start: time.Now().Add(-1 * time.Hour),
				End:   time.Now(),
			}
			data, err := pipeline.QueryData(timeRange, "cpu", 10)
			if err != nil {
				t.Errorf("Failed to query data: %v", err)
				return
			}
			results <- data
		}()
	}

	wg.Wait()
	close(results)

	// Check that all queries succeeded
	count := 0
	for data := range results {
		// Data might be empty, but query should succeed
		_ = data
		count++
	}

	if count != 15 {
		t.Errorf("Expected 15 queries, got %d", count)
	}
}

func TestRealSystemCollector_ConcurrentAccess(t *testing.T) {
	logger := zap.NewNop()
	collector := NewRealSystemCollector(logger)

	var wg sync.WaitGroup
	results := make(chan []DataPoint, 10)

	// Start multiple goroutines collecting system data concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			data, err := collector.Collect()
			if err != nil {
				t.Errorf("Failed to collect system data: %v", err)
				return
			}
			results <- data
		}()
	}

	wg.Wait()
	close(results)

	// Check that all collections succeeded
	count := 0
	for data := range results {
		if len(data) == 0 {
			t.Error("Expected some system data to be collected")
		}
		count++
	}

	if count != 10 {
		t.Errorf("Expected 10 system data collections, got %d", count)
	}
}

func TestRealNetworkCollector_ConcurrentAccess(t *testing.T) {
	logger := zap.NewNop()
	collector := NewRealNetworkCollector(logger)

	var wg sync.WaitGroup
	results := make(chan []DataPoint, 10)

	// Start multiple goroutines collecting network data concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			data, err := collector.Collect()
			if err != nil {
				t.Errorf("Failed to collect network data: %v", err)
				return
			}
			results <- data
		}()
	}

	wg.Wait()
	close(results)

	// Check that all collections succeeded
	count := 0
	for data := range results {
		if len(data) == 0 {
			t.Error("Expected some network data to be collected")
		}
		count++
	}

	if count != 10 {
		t.Errorf("Expected 10 network data collections, got %d", count)
	}
}

func TestRelayMetricsCollector_ConcurrentAccess(t *testing.T) {
	logger := zap.NewNop()
	collector := NewRelayMetricsCollector(logger)

	var wg sync.WaitGroup
	results := make(chan []DataPoint, 10)

	// Start multiple goroutines collecting relay metrics concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			data, err := collector.Collect()
			if err != nil {
				t.Errorf("Failed to collect relay metrics: %v", err)
				return
			}
			results <- data
		}()
	}

	wg.Wait()
	close(results)

	// Check that all collections succeeded
	count := 0
	for data := range results {
		if len(data) == 0 {
			t.Error("Expected some relay metrics to be collected")
		}
		count++
	}

	if count != 10 {
		t.Errorf("Expected 10 relay metrics collections, got %d", count)
	}
}

func TestRealDataNormalizer_ConcurrentAccess(t *testing.T) {
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
	}

	var wg sync.WaitGroup
	results := make(chan []DataPoint, 10)

	// Start multiple goroutines processing data concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			data, err := normalizer.Process(testData)
			if err != nil {
				t.Errorf("Failed to process data: %v", err)
				return
			}
			results <- data
		}()
	}

	wg.Wait()
	close(results)

	// Check that all processing succeeded
	count := 0
	for data := range results {
		if len(data) != len(testData) {
			t.Errorf("Expected %d processed points, got %d", len(testData), len(data))
		}
		count++
	}

	if count != 10 {
		t.Errorf("Expected 10 data processing operations, got %d", count)
	}
}

func TestDataAggregator_ConcurrentAccess(t *testing.T) {
	logger := zap.NewNop()
	aggregator := NewDataAggregator(logger, 1*time.Minute)

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
	}

	var wg sync.WaitGroup
	results := make(chan []DataPoint, 10)

	// Start multiple goroutines aggregating data concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			data, err := aggregator.Process(testData)
			if err != nil {
				t.Errorf("Failed to aggregate data: %v", err)
				return
			}
			results <- data
		}()
	}

	wg.Wait()
	close(results)

	// Check that all aggregations succeeded
	count := 0
	for data := range results {
		if len(data) != 1 {
			t.Errorf("Expected 1 aggregated point, got %d", len(data))
		}
		count++
	}

	if count != 10 {
		t.Errorf("Expected 10 data aggregation operations, got %d", count)
	}
}

func TestInMemoryStorage_ConcurrentAccess(t *testing.T) {
	logger := zap.NewNop()
	storage := NewInMemoryStorage(logger, 1*time.Hour)

	testData := []DataPoint{
		{
			Timestamp: time.Now(),
			Source:    "test",
			Type:      "cpu",
			Data:      map[string]interface{}{"usage": 75.5},
		},
	}

	var wg sync.WaitGroup

	// Start multiple goroutines storing data concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := storage.Store(testData)
			if err != nil {
				t.Errorf("Failed to store data: %v", err)
			}
		}()
	}

	// Start multiple goroutines retrieving data concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			timeRange := TimeRange{
				Start: time.Now().Add(-1 * time.Hour),
				End:   time.Now().Add(1 * time.Hour),
			}
			_, err := storage.Retrieve(timeRange)
			if err != nil {
				t.Errorf("Failed to retrieve data: %v", err)
			}
		}()
	}

	wg.Wait()

	// Clean up
	err := storage.Stop()
	if err != nil {
		t.Errorf("Failed to stop storage: %v", err)
	}
}

func TestRealFileStorage_ConcurrentAccess(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		BatchSize:     10,
		RetentionDays: 1,
		StoragePath:   "/tmp/test_storage_concurrent",
	}

	storage := NewRealFileStorage(config, logger)
	defer os.RemoveAll("/tmp/test_storage_concurrent")

	testData := []DataPoint{
		{
			Timestamp: time.Now(),
			Source:    "test",
			Type:      "cpu",
			Data:      map[string]interface{}{"usage": 75.5},
		},
	}

	var wg sync.WaitGroup

	// Start multiple goroutines storing data concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := storage.Store(testData)
			if err != nil {
				t.Errorf("Failed to store data: %v", err)
			}
		}()
	}

	// Start multiple goroutines retrieving data concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			timeRange := TimeRange{
				Start: time.Now().Add(-1 * time.Hour),
				End:   time.Now().Add(1 * time.Hour),
			}
			_, err := storage.Retrieve(timeRange)
			if err != nil {
				t.Errorf("Failed to retrieve data: %v", err)
			}
		}()
	}

	wg.Wait()

	// Clean up
	err := storage.Stop()
	if err != nil {
		t.Errorf("Failed to stop storage: %v", err)
	}
}

func TestHybridStorage_ConcurrentAccess(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		BatchSize:     10,
		RetentionDays: 1,
		StoragePath:   "/tmp/test_hybrid_concurrent",
	}

	storage := NewHybridStorage(config, logger)
	defer os.RemoveAll("/tmp/test_hybrid_concurrent")

	testData := []DataPoint{
		{
			Timestamp: time.Now(),
			Source:    "test",
			Type:      "cpu",
			Data:      map[string]interface{}{"usage": 75.5},
		},
	}

	var wg sync.WaitGroup

	// Start multiple goroutines storing data concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := storage.Store(testData)
			if err != nil {
				t.Errorf("Failed to store data: %v", err)
			}
		}()
	}

	// Start multiple goroutines retrieving data concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			timeRange := TimeRange{
				Start: time.Now().Add(-1 * time.Hour),
				End:   time.Now().Add(1 * time.Hour),
			}
			_, err := storage.Retrieve(timeRange)
			if err != nil {
				t.Errorf("Failed to retrieve data: %v", err)
			}
		}()
	}

	// Start multiple goroutines querying data concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			timeRange := TimeRange{
				Start: time.Now().Add(-1 * time.Hour),
				End:   time.Now().Add(1 * time.Hour),
			}
			_, err := storage.Query(timeRange, "cpu", 10)
			if err != nil {
				t.Errorf("Failed to query data: %v", err)
			}
		}()
	}

	wg.Wait()

	// Clean up
	err := storage.Stop()
	if err != nil {
		t.Errorf("Failed to stop storage: %v", err)
	}
}
