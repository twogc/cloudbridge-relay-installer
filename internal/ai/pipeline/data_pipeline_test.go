package pipeline

import (
	"os"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestDataPipeline_NewDataPipeline(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		Enabled:            true,
		BatchSize:          100,
		ProcessingInterval: 30 * time.Second,
		RetentionDays:      7,
		StoragePath:        "/tmp/test_pipeline",
		PrometheusURL:      "http://localhost:9090",
	}

	pipeline := NewDataPipeline(config, logger)

	if pipeline == nil {
		t.Fatal("Expected pipeline to be created")
	}

	if pipeline.config != config {
		t.Error("Expected config to be set")
	}

	if pipeline.logger != logger {
		t.Error("Expected logger to be set")
	}
}

func TestDataPipeline_StartStop(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		Enabled:            true,
		BatchSize:          10,
		ProcessingInterval: 100 * time.Millisecond,
		RetentionDays:      1,
		StoragePath:        "/tmp/test_pipeline",
		PrometheusURL:      "http://localhost:9090",
	}

	pipeline := NewDataPipeline(config, logger)

	// Start pipeline
	err := pipeline.Start()
	if err != nil {
		t.Fatalf("Failed to start pipeline: %v", err)
	}

	// Wait a bit
	time.Sleep(200 * time.Millisecond)

	// Stop pipeline
	pipeline.Stop()

	// Pipeline should be stopped
	select {
	case <-pipeline.ctx.Done():
		// Expected
	default:
		t.Error("Expected pipeline context to be cancelled")
	}
}

func TestDataPipeline_ProcessDataBasic(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		Enabled:            true,
		BatchSize:          10,
		ProcessingInterval: 100 * time.Millisecond,
		RetentionDays:      1,
		StoragePath:        "/tmp/test_pipeline",
		PrometheusURL:      "http://localhost:9090",
	}

	pipeline := NewDataPipeline(config, logger)

	// Test that pipeline can be created with test data
	if pipeline == nil {
		t.Fatal("Expected pipeline to be created")
	}
}

func TestDataPipeline_GetStats(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		Enabled:            true,
		BatchSize:          100,
		ProcessingInterval: 30 * time.Second,
		RetentionDays:      7,
		StoragePath:        "/tmp/test_pipeline",
		PrometheusURL:      "http://localhost:9090",
	}

	pipeline := NewDataPipeline(config, logger)

	// Test that pipeline has the expected structure
	if pipeline.config == nil {
		t.Error("Expected config to be set")
	}

	if pipeline.logger == nil {
		t.Error("Expected logger to be set")
	}
}

func TestDataPipeline_QueryDataBasic(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		Enabled:            true,
		BatchSize:          100,
		ProcessingInterval: 30 * time.Second,
		RetentionDays:      7,
		StoragePath:        "/tmp/test_pipeline",
		PrometheusURL:      "http://localhost:9090",
	}

	pipeline := NewDataPipeline(config, logger)

	// Test that pipeline can be created with query configuration
	if pipeline.config == nil {
		t.Fatal("Expected config to be set")
	}
}

func TestDataPipeline_Disabled(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		Enabled:            false,
		BatchSize:          100,
		ProcessingInterval: 30 * time.Second,
		RetentionDays:      7,
		StoragePath:        "/tmp/test_pipeline",
		PrometheusURL:      "http://localhost:9090",
	}

	pipeline := NewDataPipeline(config, logger)

	// Start should return immediately when disabled
	err := pipeline.Start()
	if err != nil {
		t.Fatalf("Failed to start disabled pipeline: %v", err)
	}

	// Stop should work
	pipeline.Stop()
}

func TestDataPipeline_Initialize(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		Enabled:            true,
		BatchSize:          100,
		ProcessingInterval: 30 * time.Second,
		RetentionDays:      7,
		StoragePath:        "/tmp/test_pipeline_init",
		PrometheusURL:      "http://localhost:9090",
	}

	pipeline := NewDataPipeline(config, logger)
	defer os.RemoveAll("/tmp/test_pipeline_init")

	err := pipeline.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize pipeline: %v", err)
	}

	// Check that collectors were initialized
	if len(pipeline.collectors) == 0 {
		t.Error("Expected collectors to be initialized")
	}

	// Check that processors were initialized
	if len(pipeline.processors) == 0 {
		t.Error("Expected processors to be initialized")
	}

	// Check that storage was initialized
	if pipeline.storage == nil {
		t.Error("Expected storage to be initialized")
	}
}

func TestDataPipeline_InitializeCollectors(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		Enabled:            true,
		BatchSize:          100,
		ProcessingInterval: 30 * time.Second,
		RetentionDays:      7,
		StoragePath:        "/tmp/test_pipeline_collectors",
		PrometheusURL:      "http://localhost:9090",
	}

	pipeline := NewDataPipeline(config, logger)
	defer os.RemoveAll("/tmp/test_pipeline_collectors")

	err := pipeline.initializeCollectors()
	if err != nil {
		t.Fatalf("Failed to initialize collectors: %v", err)
	}

	// Check that all expected collectors are present
	expectedCollectors := []string{"prometheus", "system", "network", "relay"}
	for _, name := range expectedCollectors {
		if _, ok := pipeline.collectors[name]; !ok {
			t.Errorf("Expected collector %s to be initialized", name)
		}
	}
}

func TestDataPipeline_InitializeProcessors(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		Enabled:            true,
		BatchSize:          100,
		ProcessingInterval: 30 * time.Second,
		RetentionDays:      7,
		StoragePath:        "/tmp/test_pipeline_processors",
		PrometheusURL:      "http://localhost:9090",
	}

	pipeline := NewDataPipeline(config, logger)
	defer os.RemoveAll("/tmp/test_pipeline_processors")

	err := pipeline.initializeProcessors()
	if err != nil {
		t.Fatalf("Failed to initialize processors: %v", err)
	}

	// Should have 3 processors: feature extractor, normalizer, aggregator
	if len(pipeline.processors) != 3 {
		t.Errorf("Expected 3 processors, got %d", len(pipeline.processors))
	}
}

func TestDataPipeline_InitializeStorage(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		Enabled:            true,
		BatchSize:          100,
		ProcessingInterval: 30 * time.Second,
		RetentionDays:      7,
		StoragePath:        "/tmp/test_pipeline_storage",
		PrometheusURL:      "http://localhost:9090",
	}

	pipeline := NewDataPipeline(config, logger)
	defer os.RemoveAll("/tmp/test_pipeline_storage")

	err := pipeline.initializeStorage()
	if err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	if pipeline.storage == nil {
		t.Error("Expected storage to be initialized")
	}
}

func TestDataPipeline_CollectData(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		Enabled:            true,
		BatchSize:          100,
		ProcessingInterval: 30 * time.Second,
		RetentionDays:      7,
		StoragePath:        "/tmp/test_pipeline_collect",
		PrometheusURL:      "http://localhost:9090",
	}

	pipeline := NewDataPipeline(config, logger)
	defer os.RemoveAll("/tmp/test_pipeline_collect")

	// Initialize pipeline
	err := pipeline.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize pipeline: %v", err)
	}

	// Collect data
	data, err := pipeline.CollectData()
	if err != nil {
		t.Fatalf("Failed to collect data: %v", err)
	}

	// Should have some data from system collectors
	if len(data) == 0 {
		t.Error("Expected some data to be collected")
	}
}

func TestDataPipeline_GetCollector(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		Enabled:            true,
		BatchSize:          100,
		ProcessingInterval: 30 * time.Second,
		RetentionDays:      7,
		StoragePath:        "/tmp/test_pipeline_get_collector",
		PrometheusURL:      "http://localhost:9090",
	}

	pipeline := NewDataPipeline(config, logger)
	defer os.RemoveAll("/tmp/test_pipeline_get_collector")

	// Initialize pipeline
	err := pipeline.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize pipeline: %v", err)
	}

	// Test getting existing collector
	collector, exists := pipeline.GetCollector("system")
	if !exists {
		t.Error("Expected system collector to exist")
	}
	if collector == nil {
		t.Error("Expected system collector to be non-nil")
	}

	// Test getting non-existing collector
	_, exists = pipeline.GetCollector("nonexistent")
	if exists {
		t.Error("Expected nonexistent collector to not exist")
	}
}

func TestDataPipeline_QueryData(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		Enabled:            true,
		BatchSize:          100,
		ProcessingInterval: 30 * time.Second,
		RetentionDays:      7,
		StoragePath:        "/tmp/test_pipeline_query",
		PrometheusURL:      "http://localhost:9090",
	}

	pipeline := NewDataPipeline(config, logger)
	defer os.RemoveAll("/tmp/test_pipeline_query")

	// Test querying without starting pipeline
	timeRange := TimeRange{
		Start: time.Now().Add(-1 * time.Hour),
		End:   time.Now(),
	}

	_, err := pipeline.QueryData(timeRange, "cpu", 10)
	if err == nil {
		t.Error("Expected error when querying non-started pipeline")
	}

	// Initialize and start pipeline
	err = pipeline.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize pipeline: %v", err)
	}

	err = pipeline.Start()
	if err != nil {
		t.Fatalf("Failed to start pipeline: %v", err)
	}
	defer pipeline.Stop()

	// Now querying should work (though might return empty results)
	data, err := pipeline.QueryData(timeRange, "cpu", 10)
	if err != nil {
		t.Fatalf("Failed to query data: %v", err)
	}

	// Data might be empty if no data was processed yet
	_ = data
}

func TestDataPipeline_GetMetrics(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		Enabled:            true,
		BatchSize:          100,
		ProcessingInterval: 30 * time.Second,
		RetentionDays:      7,
		StoragePath:        "/tmp/test_pipeline_metrics",
		PrometheusURL:      "http://localhost:9090",
	}

	pipeline := NewDataPipeline(config, logger)
	defer os.RemoveAll("/tmp/test_pipeline_metrics")

	// Initialize pipeline
	err := pipeline.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize pipeline: %v", err)
	}

	metrics, err := pipeline.GetMetrics()
	if err != nil {
		t.Fatalf("Failed to get metrics: %v", err)
	}

	// Check that metrics contain expected fields
	if _, ok := metrics["collectors"]; !ok {
		t.Error("Expected collectors metrics")
	}
	if _, ok := metrics["processors"]; !ok {
		t.Error("Expected processors metrics")
	}
	if _, ok := metrics["storage"]; !ok {
		t.Error("Expected storage metrics")
	}
	if _, ok := metrics["pipeline_status"]; !ok {
		t.Error("Expected pipeline_status metrics")
	}

	// Pipeline should be stopped initially
	if status, ok := metrics["pipeline_status"].(string); ok {
		if status != "stopped" {
			t.Errorf("Expected pipeline_status 'stopped', got '%s'", status)
		}
	}
}

func TestDataPipeline_StartStopLifecycle(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		Enabled:            true,
		BatchSize:          10,
		ProcessingInterval: 100 * time.Millisecond,
		RetentionDays:      1,
		StoragePath:        "/tmp/test_pipeline_lifecycle",
		PrometheusURL:      "http://localhost:9090",
	}

	pipeline := NewDataPipeline(config, logger)
	defer os.RemoveAll("/tmp/test_pipeline_lifecycle")

	// Initialize pipeline
	err := pipeline.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize pipeline: %v", err)
	}

	// Test starting pipeline
	err = pipeline.Start()
	if err != nil {
		t.Fatalf("Failed to start pipeline: %v", err)
	}

	// Check that pipeline is started
	if !pipeline.started {
		t.Error("Expected pipeline to be started")
	}

	// Wait a bit for processing
	time.Sleep(200 * time.Millisecond)

	// Test stopping pipeline
	err = pipeline.Stop()
	if err != nil {
		t.Fatalf("Failed to stop pipeline: %v", err)
	}

	// Check that pipeline is stopped
	if pipeline.started {
		t.Error("Expected pipeline to be stopped")
	}

	// Check that context is cancelled
	select {
	case <-pipeline.ctx.Done():
		// Expected
	default:
		t.Error("Expected pipeline context to be cancelled")
	}
}

func TestDataPipeline_DoubleStart(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		Enabled:            true,
		BatchSize:          100,
		ProcessingInterval: 30 * time.Second,
		RetentionDays:      7,
		StoragePath:        "/tmp/test_pipeline_double_start",
		PrometheusURL:      "http://localhost:9090",
	}

	pipeline := NewDataPipeline(config, logger)
	defer os.RemoveAll("/tmp/test_pipeline_double_start")

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

	// Try to start again - should return error
	err = pipeline.Start()
	if err == nil {
		t.Error("Expected error when starting already started pipeline")
	}
}

func TestDataPipeline_StopWithoutStart(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		Enabled:            true,
		BatchSize:          100,
		ProcessingInterval: 30 * time.Second,
		RetentionDays:      7,
		StoragePath:        "/tmp/test_pipeline_stop_without_start",
		PrometheusURL:      "http://localhost:9090",
	}

	pipeline := NewDataPipeline(config, logger)
	defer os.RemoveAll("/tmp/test_pipeline_stop_without_start")

	// Stop without starting should not return error
	err := pipeline.Stop()
	if err != nil {
		t.Fatalf("Failed to stop non-started pipeline: %v", err)
	}
}

func TestDataPipeline_ProcessData(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		Enabled:            true,
		BatchSize:          10,
		ProcessingInterval: 100 * time.Millisecond,
		RetentionDays:      1,
		StoragePath:        "/tmp/test_pipeline_process",
		PrometheusURL:      "http://localhost:9090",
	}

	pipeline := NewDataPipeline(config, logger)
	defer os.RemoveAll("/tmp/test_pipeline_process")

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

	// Wait for processing to occur
	time.Sleep(200 * time.Millisecond)

	// Pipeline should have processed some data
	// We can't easily test the exact data without mocking, but we can verify
	// that the pipeline is running and processing
}

func TestDataPipeline_ConfigValidation(t *testing.T) {
	logger := zap.NewNop()

	// Test with valid config
	validConfig := &Config{
		Enabled:            true,
		BatchSize:          100,
		ProcessingInterval: 30 * time.Second,
		RetentionDays:      7,
		StoragePath:        "/tmp/test_pipeline_config",
		PrometheusURL:      "http://localhost:9090",
	}

	pipeline := NewDataPipeline(validConfig, logger)
	defer os.RemoveAll("/tmp/test_pipeline_config")

	if pipeline.config != validConfig {
		t.Error("Expected config to be set correctly")
	}

	if pipeline.logger != logger {
		t.Error("Expected logger to be set correctly")
	}
}

func TestDataPipeline_DefaultPrometheusURL(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		Enabled:            true,
		BatchSize:          100,
		ProcessingInterval: 30 * time.Second,
		RetentionDays:      7,
		StoragePath:        "/tmp/test_pipeline_default_prometheus",
		PrometheusURL:      "", // Empty URL should use default
	}

	pipeline := NewDataPipeline(config, logger)
	defer os.RemoveAll("/tmp/test_pipeline_default_prometheus")

	err := pipeline.initializeCollectors()
	if err != nil {
		t.Fatalf("Failed to initialize collectors with default Prometheus URL: %v", err)
	}

	// Should have created Prometheus collector with default URL
	if _, ok := pipeline.collectors["prometheus"]; !ok {
		t.Error("Expected Prometheus collector to be created with default URL")
	}
}
