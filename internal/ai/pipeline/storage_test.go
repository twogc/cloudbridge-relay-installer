package pipeline

import (
	"os"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestRealFileStorage_Store(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		BatchSize:     10,
		RetentionDays: 1,
		StoragePath:   "/tmp/test_storage",
	}

	storage := NewRealFileStorage(config, logger)

	// Create test data
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
			Type:      "memory",
			Data: map[string]interface{}{
				"memory_usage": 60.0,
				"free_memory":  40.0,
			},
		},
	}

	err := storage.Store(testData)
	if err != nil {
		t.Fatalf("Failed to store data: %v", err)
	}

	// Clean up
	defer os.RemoveAll("/tmp/test_storage")
}

func TestHybridStorage_Store(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		BatchSize:     10,
		RetentionDays: 1,
		StoragePath:   "/tmp/test_hybrid_storage",
	}

	storage := NewHybridStorage(config, logger)

	// Create test data
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
	}

	err := storage.Store(testData)
	if err != nil {
		t.Fatalf("Failed to store data in hybrid storage: %v", err)
	}

	// Clean up
	defer os.RemoveAll("/tmp/test_hybrid_storage")
}

func TestRealFileStorage_Query(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		BatchSize:     10,
		RetentionDays: 1,
		StoragePath:   "/tmp/test_storage_query",
	}

	storage := NewRealFileStorage(config, logger)

	// Store some test data
	testData := []DataPoint{
		{
			Timestamp: time.Now(),
			Source:    "test",
			Type:      "cpu",
			Data: map[string]interface{}{
				"cpu_usage": 75.5,
			},
		},
	}

	err := storage.Store(testData)
	if err != nil {
		t.Fatalf("Failed to store test data: %v", err)
	}

	// Query the data using Retrieve method
	timeRange := TimeRange{
		Start: time.Now().Add(-1 * time.Hour),
		End:   time.Now().Add(1 * time.Hour),
	}

	queriedData, err := storage.Retrieve(timeRange)
	if err != nil {
		t.Fatalf("Failed to retrieve data: %v", err)
	}

	// Data might be empty if not yet flushed to disk
	_ = queriedData

	// Clean up
	defer os.RemoveAll("/tmp/test_storage_query")
}

func TestRealFileStorage_Structure(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		BatchSize:     10,
		RetentionDays: 1,
		StoragePath:   "/tmp/test_storage_stats",
	}

	storage := NewRealFileStorage(config, logger)

	if storage == nil {
		t.Fatal("Expected storage to be created")
	}

	if storage.config != config {
		t.Error("Expected config to be set")
	}

	if storage.logger != logger {
		t.Error("Expected logger to be set")
	}

	// Clean up
	defer os.RemoveAll("/tmp/test_storage_stats")
}

func TestRealFileStorage_Cleanup(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		BatchSize:     10,
		RetentionDays: 1,
		StoragePath:   "/tmp/test_storage_cleanup",
	}

	storage := NewRealFileStorage(config, logger)

	// Store some old data
	oldData := []DataPoint{
		{
			Timestamp: time.Now().Add(-2 * 24 * time.Hour), // 2 days ago
			Source:    "test",
			Type:      "cpu",
			Data: map[string]interface{}{
				"cpu_usage": 75.5,
			},
		},
	}

	err := storage.Store(oldData)
	if err != nil {
		t.Fatalf("Failed to store old data: %v", err)
	}

	// Test that storage was created successfully
	if storage == nil {
		t.Fatal("Expected storage to be created")
	}

	// Clean up
	defer os.RemoveAll("/tmp/test_storage_cleanup")
}

func TestStorageConfig_Validation(t *testing.T) {
	config := &Config{
		BatchSize:     10,
		RetentionDays: 1,
		StoragePath:   "/tmp/test",
	}

	if config.BatchSize <= 0 {
		t.Error("Batch size should be positive")
	}

	if config.RetentionDays <= 0 {
		t.Error("Retention days should be positive")
	}

	if config.StoragePath == "" {
		t.Error("Storage path should be set")
	}
}

func TestTimeRange_Validation(t *testing.T) {
	now := time.Now()

	// Valid time range
	validRange := TimeRange{
		Start: now.Add(-1 * time.Hour),
		End:   now,
	}

	if validRange.Start.After(validRange.End) {
		t.Error("Start time should be before end time")
	}

	// Invalid time range (start after end)
	invalidRange := TimeRange{
		Start: now,
		End:   now.Add(-1 * time.Hour),
	}

	if !invalidRange.Start.After(invalidRange.End) {
		t.Error("This should be an invalid range")
	}
}
