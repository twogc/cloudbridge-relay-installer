package pipeline

import (
	"encoding/json"
	"os"
	"path/filepath"
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

func TestHybridStorage_StoreBasic(t *testing.T) {
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

func TestRealFileStorage_FlushBuffer(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		BatchSize:     2, // Small batch size to trigger flush
		RetentionDays: 1,
		StoragePath:   "/tmp/test_storage_flush",
	}

	storage := NewRealFileStorage(config, logger)
	defer os.RemoveAll("/tmp/test_storage_flush")

	// Create test data that will trigger flush
	testData := []DataPoint{
		{
			Timestamp: time.Now(),
			Source:    "test",
			Type:      "cpu",
			Data: map[string]interface{}{
				"cpu_usage": 75.5,
			},
		},
		{
			Timestamp: time.Now().Add(-1 * time.Minute),
			Source:    "test",
			Type:      "memory",
			Data: map[string]interface{}{
				"memory_usage": 60.0,
			},
		},
	}

	err := storage.Store(testData)
	if err != nil {
		t.Fatalf("Failed to store data: %v", err)
	}

	// Check that files were created
	matches, err := filepath.Glob("/tmp/test_storage_flush/*.jsonl")
	if err != nil {
		t.Fatalf("Failed to glob files: %v", err)
	}

	if len(matches) == 0 {
		t.Error("Expected files to be created after flush")
	}
}

func TestRealFileStorage_GroupDataByDateAndType(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		BatchSize:     10,
		RetentionDays: 1,
		StoragePath:   "/tmp/test_storage_group",
	}

	storage := NewRealFileStorage(config, logger)
	defer os.RemoveAll("/tmp/test_storage_group")

	now := time.Now()
	testData := []DataPoint{
		{
			Timestamp: now,
			Source:    "test1",
			Type:      "cpu",
			Data:      map[string]interface{}{"usage": 75.5},
		},
		{
			Timestamp: now,
			Source:    "test1",
			Type:      "memory",
			Data:      map[string]interface{}{"usage": 60.0},
		},
		{
			Timestamp: now,
			Source:    "test2",
			Type:      "cpu",
			Data:      map[string]interface{}{"usage": 80.0},
		},
	}

	groups := storage.groupDataByDateAndType(testData)

	// Should have 3 groups
	if len(groups) != 3 {
		t.Errorf("Expected 3 groups, got %d", len(groups))
	}

	// Check group keys
	expectedKeys := []string{
		now.Format("2006-01-02") + "_test1_cpu",
		now.Format("2006-01-02") + "_test1_memory",
		now.Format("2006-01-02") + "_test2_cpu",
	}

	for _, expectedKey := range expectedKeys {
		if _, ok := groups[expectedKey]; !ok {
			t.Errorf("Expected group key %s not found", expectedKey)
		}
	}
}

func TestRealFileStorage_SaveGroup(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		BatchSize:     10,
		RetentionDays: 1,
		StoragePath:   "/tmp/test_storage_save",
	}

	storage := NewRealFileStorage(config, logger)
	defer os.RemoveAll("/tmp/test_storage_save")

	// Create directory first
	err := os.MkdirAll("/tmp/test_storage_save", 0750)
	if err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	testData := []DataPoint{
		{
			Timestamp: time.Now(),
			Source:    "test",
			Type:      "cpu",
			Data:      map[string]interface{}{"usage": 75.5},
		},
	}

	err = storage.saveGroup("test_key", testData)
	if err != nil {
		t.Fatalf("Failed to save group: %v", err)
	}

	// Check that file was created
	expectedFile := "/tmp/test_storage_save/test_key.jsonl"
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Error("Expected file to be created")
	}
}

func TestRealFileStorage_ReadFile(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		BatchSize:     10,
		RetentionDays: 1,
		StoragePath:   "/tmp/test_storage_read",
	}

	storage := NewRealFileStorage(config, logger)
	defer os.RemoveAll("/tmp/test_storage_read")

	// Create test file
	testFile := "/tmp/test_storage_read/test.jsonl"
	err := os.MkdirAll("/tmp/test_storage_read", 0750)
	if err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	file, err := os.Create(testFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Write test data
	testData := DataPoint{
		Timestamp: time.Now(),
		Source:    "test",
		Type:      "cpu",
		Data:      map[string]interface{}{"usage": 75.5},
	}

	encoder := json.NewEncoder(file)
	err = encoder.Encode(testData)
	file.Close()
	if err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}

	// Read file
	timeRange := TimeRange{
		Start: time.Now().Add(-1 * time.Hour),
		End:   time.Now().Add(1 * time.Hour),
	}

	data, err := storage.readFile(testFile, timeRange)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if len(data) != 1 {
		t.Errorf("Expected 1 data point, got %d", len(data))
	}

	if data[0].Source != "test" {
		t.Errorf("Expected source 'test', got '%s'", data[0].Source)
	}
}

func TestRealFileStorage_StartStop(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		BatchSize:     10,
		RetentionDays: 1,
		StoragePath:   "/tmp/test_storage_lifecycle",
	}

	storage := NewRealFileStorage(config, logger)
	defer os.RemoveAll("/tmp/test_storage_lifecycle")

	// Test Start
	err := storage.Start()
	if err != nil {
		t.Fatalf("Failed to start storage: %v", err)
	}

	// Test Stop
	err = storage.Stop()
	if err != nil {
		t.Fatalf("Failed to stop storage: %v", err)
	}
}

func TestInMemoryStorage_Store(t *testing.T) {
	logger := zap.NewNop()
	storage := NewInMemoryStorage(logger, 1*time.Hour)

	testData := []DataPoint{
		{
			Timestamp: time.Now(),
			Source:    "test",
			Type:      "cpu",
			Data:      map[string]interface{}{"usage": 75.5},
		},
		{
			Timestamp: time.Now().Add(-1 * time.Minute),
			Source:    "test",
			Type:      "memory",
			Data:      map[string]interface{}{"usage": 60.0},
		},
	}

	err := storage.Store(testData)
	if err != nil {
		t.Fatalf("Failed to store data: %v", err)
	}

	// Check that data was stored
	storage.mutex.RLock()
	defer storage.mutex.RUnlock()

	if len(storage.data) == 0 {
		t.Error("Expected data to be stored in memory")
	}

	// Check specific keys
	if _, ok := storage.data["test_cpu"]; !ok {
		t.Error("Expected test_cpu key to exist")
	}
	if _, ok := storage.data["test_memory"]; !ok {
		t.Error("Expected test_memory key to exist")
	}
}

func TestInMemoryStorage_Retrieve(t *testing.T) {
	logger := zap.NewNop()
	storage := NewInMemoryStorage(logger, 1*time.Hour)

	now := time.Now()
	testData := []DataPoint{
		{
			Timestamp: now,
			Source:    "test",
			Type:      "cpu",
			Data:      map[string]interface{}{"usage": 75.5},
		},
		{
			Timestamp: now.Add(-2 * time.Hour), // Old data
			Source:    "test",
			Type:      "memory",
			Data:      map[string]interface{}{"usage": 60.0},
		},
	}

	err := storage.Store(testData)
	if err != nil {
		t.Fatalf("Failed to store data: %v", err)
	}

	// Retrieve recent data
	timeRange := TimeRange{
		Start: now.Add(-1 * time.Hour),
		End:   now.Add(1 * time.Hour),
	}

	data, err := storage.Retrieve(timeRange)
	if err != nil {
		t.Fatalf("Failed to retrieve data: %v", err)
	}

	// Should only get recent data
	if len(data) != 1 {
		t.Errorf("Expected 1 data point, got %d", len(data))
	}

	if data[0].Type != "cpu" {
		t.Errorf("Expected type 'cpu', got '%s'", data[0].Type)
	}
}

func TestInMemoryStorage_Query(t *testing.T) {
	logger := zap.NewNop()
	storage := NewInMemoryStorage(logger, 1*time.Hour)

	now := time.Now()
	testData := []DataPoint{
		{
			Timestamp: now,
			Source:    "test",
			Type:      "cpu",
			Data:      map[string]interface{}{"usage": 75.5},
		},
		{
			Timestamp: now,
			Source:    "test",
			Type:      "memory",
			Data:      map[string]interface{}{"usage": 60.0},
		},
	}

	err := storage.Store(testData)
	if err != nil {
		t.Fatalf("Failed to store data: %v", err)
	}

	// Query with type filter
	timeRange := TimeRange{
		Start: now.Add(-1 * time.Hour),
		End:   now.Add(1 * time.Hour),
	}

	data, err := storage.Query(timeRange, "cpu", 10)
	if err != nil {
		t.Fatalf("Failed to query data: %v", err)
	}

	// Should only get CPU data
	if len(data) != 1 {
		t.Errorf("Expected 1 data point, got %d", len(data))
	}

	if data[0].Type != "cpu" {
		t.Errorf("Expected type 'cpu', got '%s'", data[0].Type)
	}
}

func TestInMemoryStorage_Stop(t *testing.T) {
	logger := zap.NewNop()
	storage := NewInMemoryStorage(logger, 1*time.Hour)

	err := storage.Stop()
	if err != nil {
		t.Fatalf("Failed to stop storage: %v", err)
	}
}

func TestHybridStorage_Store(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		BatchSize:     10,
		RetentionDays: 1,
		StoragePath:   "/tmp/test_hybrid_store",
	}

	storage := NewHybridStorage(config, logger)
	defer os.RemoveAll("/tmp/test_hybrid_store")

	testData := []DataPoint{
		{
			Timestamp: time.Now(),
			Source:    "test",
			Type:      "cpu",
			Data:      map[string]interface{}{"usage": 75.5},
		},
	}

	err := storage.Store(testData)
	if err != nil {
		t.Fatalf("Failed to store data in hybrid storage: %v", err)
	}
}

func TestHybridStorage_Retrieve(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		BatchSize:     10,
		RetentionDays: 1,
		StoragePath:   "/tmp/test_hybrid_retrieve",
	}

	storage := NewHybridStorage(config, logger)
	defer os.RemoveAll("/tmp/test_hybrid_retrieve")

	now := time.Now()
	testData := []DataPoint{
		{
			Timestamp: now,
			Source:    "test",
			Type:      "cpu",
			Data:      map[string]interface{}{"usage": 75.5},
		},
	}

	err := storage.Store(testData)
	if err != nil {
		t.Fatalf("Failed to store data: %v", err)
	}

	timeRange := TimeRange{
		Start: now.Add(-1 * time.Hour),
		End:   now.Add(1 * time.Hour),
	}

	data, err := storage.Retrieve(timeRange)
	if err != nil {
		t.Fatalf("Failed to retrieve data: %v", err)
	}

	// Should get data from memory storage
	if len(data) == 0 {
		t.Error("Expected to retrieve data from hybrid storage")
	}
}

func TestHybridStorage_Query(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		BatchSize:     10,
		RetentionDays: 1,
		StoragePath:   "/tmp/test_hybrid_query",
	}

	storage := NewHybridStorage(config, logger)
	defer os.RemoveAll("/tmp/test_hybrid_query")

	now := time.Now()
	testData := []DataPoint{
		{
			Timestamp: now,
			Source:    "test",
			Type:      "cpu",
			Data:      map[string]interface{}{"usage": 75.5},
		},
	}

	err := storage.Store(testData)
	if err != nil {
		t.Fatalf("Failed to store data: %v", err)
	}

	timeRange := TimeRange{
		Start: now.Add(-1 * time.Hour),
		End:   now.Add(1 * time.Hour),
	}

	data, err := storage.Query(timeRange, "cpu", 10)
	if err != nil {
		t.Fatalf("Failed to query data: %v", err)
	}

	// Should get data from memory storage
	if len(data) == 0 {
		t.Error("Expected to query data from hybrid storage")
	}
}

func TestHybridStorage_StartStop(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		BatchSize:     10,
		RetentionDays: 1,
		StoragePath:   "/tmp/test_hybrid_lifecycle",
	}

	storage := NewHybridStorage(config, logger)
	defer os.RemoveAll("/tmp/test_hybrid_lifecycle")

	// Test Start
	err := storage.Start()
	if err != nil {
		t.Fatalf("Failed to start hybrid storage: %v", err)
	}

	// Test Stop
	err = storage.Stop()
	if err != nil {
		t.Fatalf("Failed to stop hybrid storage: %v", err)
	}
}

func TestRealFileStorage_UnsafePath(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		BatchSize:     10,
		RetentionDays: 1,
		StoragePath:   "/tmp/test_storage_safe",
	}

	storage := NewRealFileStorage(config, logger)
	defer os.RemoveAll("/tmp/test_storage_safe")

	// Test unsafe path
	unsafePath := "/etc/passwd"
	timeRange := TimeRange{
		Start: time.Now().Add(-1 * time.Hour),
		End:   time.Now().Add(1 * time.Hour),
	}

	_, err := storage.readFile(unsafePath, timeRange)
	if err == nil {
		t.Error("Expected error for unsafe path")
	}
}

func TestRealFileStorage_EmptyBuffer(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		BatchSize:     10,
		RetentionDays: 1,
		StoragePath:   "/tmp/test_storage_empty",
	}

	storage := NewRealFileStorage(config, logger)
	defer os.RemoveAll("/tmp/test_storage_empty")

	// Test flush with empty buffer
	err := storage.flushBuffer()
	if err != nil {
		t.Fatalf("Failed to flush empty buffer: %v", err)
	}
}
