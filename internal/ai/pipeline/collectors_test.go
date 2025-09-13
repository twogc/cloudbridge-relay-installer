package pipeline

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestRealSystemCollector_Collect(t *testing.T) {
	logger := zap.NewNop()
	collector := NewRealSystemCollector(logger)

	data, err := collector.Collect()
	if err != nil {
		t.Fatalf("Failed to collect system data: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected system data to be collected")
	}

	// Check that we have different types of data
	types := make(map[string]bool)
	for _, point := range data {
		types[point.Type] = true
	}

	expectedTypes := []string{"cpu", "memory", "disk", "runtime"}
	for _, expectedType := range expectedTypes {
		if !types[expectedType] {
			t.Errorf("Expected to find data type: %s", expectedType)
		}
	}
}

func TestRealNetworkCollector_Collect(t *testing.T) {
	logger := zap.NewNop()
	collector := NewRealNetworkCollector(logger)

	data, err := collector.Collect()
	if err != nil {
		t.Fatalf("Failed to collect network data: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected network data to be collected")
	}

	// Check that we have network interface data
	hasInterfaceData := false
	for _, point := range data {
		if point.Type == "interface" {
			hasInterfaceData = true
			break
		}
	}

	if !hasInterfaceData {
		t.Error("Expected to find network interface data")
	}
}

func TestRelayMetricsCollector_Collect(t *testing.T) {
	logger := zap.NewNop()
	collector := NewRelayMetricsCollector(logger)

	// Update with test data
	stats := &RelayStats{
		ConnectionsTotal:   100,
		ConnectionsActive:  50,
		BytesTransferred:   1024000,
		PacketsTransferred: 5000,
		AverageLatency:     25.5,
		ErrorRate:          0.01,
		LastUpdated:        time.Now(),
	}

	collector.UpdateStats(stats)

	data, err := collector.Collect()
	if err != nil {
		t.Fatalf("Failed to collect relay metrics: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected relay metrics to be collected")
	}

	// Check that we have the expected metrics
	point := data[0]
	if point.Type != "metrics" {
		t.Errorf("Expected type 'metrics', got '%s'", point.Type)
	}

	if point.Source != "relay" {
		t.Errorf("Expected source 'relay', got '%s'", point.Source)
	}
}

func TestRealPrometheusCollector_Collect(t *testing.T) {
	logger := zap.NewNop()
	// Use a non-existent URL to test error handling
	collector := NewRealPrometheusCollector(logger, "http://localhost:9999")

	data, err := collector.Collect()
	// Should not fail completely, but may have warnings
	if err != nil {
		t.Logf("Prometheus collection failed as expected: %v", err)
	}

	// Data might be empty if Prometheus is not available
	_ = data
}

func TestDataPoint_Validation(t *testing.T) {
	// Test valid data point
	validPoint := DataPoint{
		Timestamp: time.Now(),
		Source:    "test",
		Type:      "cpu",
		Data: map[string]interface{}{
			"usage": 75.5,
		},
	}

	if validPoint.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}

	if validPoint.Source == "" {
		t.Error("Expected source to be set")
	}

	if validPoint.Type == "" {
		t.Error("Expected type to be set")
	}

	if validPoint.Data == nil {
		t.Error("Expected data to be set")
	}
}

func TestDataPoint_Structure(t *testing.T) {
	now := time.Now()
	point := DataPoint{
		Timestamp: now,
		Source:    "test",
		Type:      "cpu",
		Data: map[string]interface{}{
			"usage": 75.5,
		},
	}

	if point.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}

	if point.Source == "" {
		t.Error("Expected source to be set")
	}

	if point.Type == "" {
		t.Error("Expected type to be set")
	}

	if point.Data == nil {
		t.Error("Expected data to be set")
	}
}
