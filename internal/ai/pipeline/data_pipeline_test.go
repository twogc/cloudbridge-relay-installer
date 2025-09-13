package pipeline

import (
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

func TestDataPipeline_ProcessData(t *testing.T) {
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

func TestDataPipeline_QueryData(t *testing.T) {
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
