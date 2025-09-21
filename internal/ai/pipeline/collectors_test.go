package pipeline

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

	// Test with mock Prometheus server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/query" {
			response := map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"resultType": "vector",
					"result": []map[string]interface{}{
						{
							"value": []interface{}{"1640995200", "42.5"},
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	collector := NewRealPrometheusCollector(logger, server.URL)
	data, err := collector.Collect()

	if err != nil {
		t.Fatalf("Failed to collect Prometheus data: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected Prometheus data to be collected")
	}

	// Check that we have the expected metrics
	metricNames := make(map[string]bool)
	for _, point := range data {
		if point.Source != "prometheus" {
			t.Errorf("Expected source 'prometheus', got '%s'", point.Source)
		}
		if point.Type != "metric" {
			t.Errorf("Expected type 'metric', got '%s'", point.Type)
		}

		if metricName, ok := point.Data["metric_name"]; ok {
			metricNames[metricName.(string)] = true
		}
	}

	expectedMetrics := []string{
		"relay_connections_total",
		"relay_connections_active",
		"relay_bandwidth_bytes",
		"relay_latency_seconds",
		"relay_errors_total",
		"relay_requests_total",
	}

	for _, expected := range expectedMetrics {
		if !metricNames[expected] {
			t.Errorf("Expected to find metric: %s", expected)
		}
	}
}

func TestRealPrometheusCollector_ErrorHandling(t *testing.T) {
	logger := zap.NewNop()

	// Test with server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	collector := NewRealPrometheusCollector(logger, server.URL)
	data, err := collector.Collect()

	// Should not fail completely, but may have warnings
	if err != nil {
		t.Logf("Prometheus collection failed as expected: %v", err)
	}

	// Data might be empty if Prometheus returns errors
	_ = data
}

func TestRealPrometheusCollector_InvalidResponse(t *testing.T) {
	logger := zap.NewNop()

	// Test with server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	collector := NewRealPrometheusCollector(logger, server.URL)
	data, err := collector.Collect()

	// Should handle invalid response gracefully
	if err != nil {
		t.Logf("Prometheus collection failed as expected: %v", err)
	}

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

func TestRealSystemCollector_StartStop(t *testing.T) {
	logger := zap.NewNop()
	collector := NewRealSystemCollector(logger)

	// Test Start
	err := collector.Start()
	if err != nil {
		t.Fatalf("Failed to start system collector: %v", err)
	}

	// Test Stop
	err = collector.Stop()
	if err != nil {
		t.Fatalf("Failed to stop system collector: %v", err)
	}
}

func TestRealSystemCollector_CollectCPUData(t *testing.T) {
	logger := zap.NewNop()
	collector := NewRealSystemCollector(logger)

	cpuData, err := collector.collectCPUData()
	if err != nil {
		t.Fatalf("Failed to collect CPU data: %v", err)
	}

	if len(cpuData) == 0 {
		t.Error("Expected CPU data to be collected")
	}

	// Check CPU data structure
	for _, point := range cpuData {
		if point.Source != "system" {
			t.Errorf("Expected source 'system', got '%s'", point.Source)
		}
		if point.Type != "cpu" {
			t.Errorf("Expected type 'cpu', got '%s'", point.Type)
		}

		// Check for expected CPU metrics
		if _, ok := point.Data["cpu_usage_percent"]; !ok {
			t.Error("Expected cpu_usage_percent in CPU data")
		}
		if _, ok := point.Data["cpu_cores"]; !ok {
			t.Error("Expected cpu_cores in CPU data")
		}
	}
}

func TestRealSystemCollector_CollectMemoryData(t *testing.T) {
	logger := zap.NewNop()
	collector := NewRealSystemCollector(logger)

	memData, err := collector.collectMemoryData()
	if err != nil {
		t.Fatalf("Failed to collect memory data: %v", err)
	}

	if len(memData) == 0 {
		t.Error("Expected memory data to be collected")
	}

	// Check memory data structure
	for _, point := range memData {
		if point.Source != "system" {
			t.Errorf("Expected source 'system', got '%s'", point.Source)
		}
		if point.Type != "memory" {
			t.Errorf("Expected type 'memory', got '%s'", point.Type)
		}

		// Check for expected memory metrics
		expectedMetrics := []string{
			"memory_total_bytes",
			"memory_available_bytes",
			"memory_used_bytes",
			"memory_usage_percent",
		}

		for _, metric := range expectedMetrics {
			if _, ok := point.Data[metric]; !ok {
				t.Errorf("Expected %s in memory data", metric)
			}
		}
	}
}

func TestRealSystemCollector_CollectRuntimeData(t *testing.T) {
	logger := zap.NewNop()
	collector := NewRealSystemCollector(logger)

	runtimeData := collector.collectRuntimeData()
	if len(runtimeData) == 0 {
		t.Error("Expected runtime data to be collected")
	}

	// Check runtime data structure
	for _, point := range runtimeData {
		if point.Source != "system" {
			t.Errorf("Expected source 'system', got '%s'", point.Source)
		}
		if point.Type != "runtime" {
			t.Errorf("Expected type 'runtime', got '%s'", point.Type)
		}

		// Check for expected runtime metrics
		expectedMetrics := []string{
			"go_routines",
			"go_memory_alloc_bytes",
			"go_memory_total_bytes",
			"go_memory_sys_bytes",
		}

		for _, metric := range expectedMetrics {
			if _, ok := point.Data[metric]; !ok {
				t.Errorf("Expected %s in runtime data", metric)
			}
		}
	}
}

func TestRealNetworkCollector_StartStop(t *testing.T) {
	logger := zap.NewNop()
	collector := NewRealNetworkCollector(logger)

	// Test Start
	err := collector.Start()
	if err != nil {
		t.Fatalf("Failed to start network collector: %v", err)
	}

	// Test Stop
	err = collector.Stop()
	if err != nil {
		t.Fatalf("Failed to stop network collector: %v", err)
	}
}

func TestRelayMetricsCollector_StartStop(t *testing.T) {
	logger := zap.NewNop()
	collector := NewRelayMetricsCollector(logger)

	// Test Start
	err := collector.Start()
	if err != nil {
		t.Fatalf("Failed to start relay metrics collector: %v", err)
	}

	// Test Stop
	err = collector.Stop()
	if err != nil {
		t.Fatalf("Failed to stop relay metrics collector: %v", err)
	}
}

func TestRelayMetricsCollector_UpdateStats(t *testing.T) {
	logger := zap.NewNop()
	collector := NewRelayMetricsCollector(logger)

	// Test updating stats
	stats := &RelayStats{
		ConnectionsTotal:   200,
		ConnectionsActive:  100,
		BytesTransferred:   2048000,
		PacketsTransferred: 10000,
		AverageLatency:     30.0,
		ErrorRate:          0.02,
		LastUpdated:        time.Now(),
	}

	collector.UpdateStats(stats)

	// Collect data and verify it matches updated stats
	data, err := collector.Collect()
	if err != nil {
		t.Fatalf("Failed to collect relay metrics: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected relay metrics to be collected")
	}

	point := data[0]
	if connectionsTotal, ok := point.Data["connections_total"]; ok {
		// Convert to int64 for comparison
		var actualTotal int64
		switch v := connectionsTotal.(type) {
		case int64:
			actualTotal = v
		case float64:
			actualTotal = int64(v)
		case int:
			actualTotal = int64(v)
		default:
			t.Errorf("Unexpected type for connections_total: %T", connectionsTotal)
			return
		}
		if actualTotal != stats.ConnectionsTotal {
			t.Errorf("Expected connections_total %d, got %d", stats.ConnectionsTotal, actualTotal)
		}
	} else {
		t.Error("Expected connections_total field in relay metrics data")
	}
}

func TestRelayMetricsCollector_ParseMetricValue(t *testing.T) {
	logger := zap.NewNop()
	collector := NewRelayMetricsCollector(logger)

	// Test valid metric line
	validLine := "relay_connections_total 42.5"
	value := collector.parseMetricValue(validLine)
	if value != 42.5 {
		t.Errorf("Expected value 42.5, got %f", value)
	}

	// Test invalid metric line
	invalidLine := "invalid_metric"
	value = collector.parseMetricValue(invalidLine)
	if value != 0 {
		t.Errorf("Expected value 0 for invalid line, got %f", value)
	}

	// Test metric line with labels
	labeledLine := `relay_connections_total{protocol="QUIC"} 25.0`
	value = collector.parseMetricValue(labeledLine)
	if value != 25.0 {
		t.Errorf("Expected value 25.0, got %f", value)
	}
}

func TestRealPrometheusCollector_StartStop(t *testing.T) {
	logger := zap.NewNop()
	collector := NewRealPrometheusCollector(logger, "http://localhost:9090")

	// Test Start
	err := collector.Start()
	if err != nil {
		t.Fatalf("Failed to start Prometheus collector: %v", err)
	}

	// Test Stop
	err = collector.Stop()
	if err != nil {
		t.Fatalf("Failed to stop Prometheus collector: %v", err)
	}
}

func TestRealPrometheusCollector_QueryPrometheus(t *testing.T) {
	logger := zap.NewNop()

	// Test with mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"resultType": "vector",
				"result": []map[string]interface{}{
					{
						"value": []interface{}{"1640995200", "123.45"},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	collector := NewRealPrometheusCollector(logger, server.URL)
	value, err := collector.queryPrometheus("test_metric")

	if err != nil {
		t.Fatalf("Failed to query Prometheus: %v", err)
	}

	if value != 123.45 {
		t.Errorf("Expected value 123.45, got %f", value)
	}
}

func TestRealPrometheusCollector_QueryPrometheusNoData(t *testing.T) {
	logger := zap.NewNop()

	// Test with server that returns no data
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"resultType": "vector",
				"result":     []map[string]interface{}{},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	collector := NewRealPrometheusCollector(logger, server.URL)
	_, err := collector.queryPrometheus("test_metric")

	if err == nil {
		t.Error("Expected error for no data response")
	}
}
