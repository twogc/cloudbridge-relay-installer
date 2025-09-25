package routing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"
)

func TestNewNetworkMetricsCollector(t *testing.T) {
	logger := zaptest.NewLogger(t)
	prometheusURL := "http://localhost:9090"
	cloudbridgeURL := "http://localhost:8080"

	collector := NewNetworkMetricsCollector(logger, prometheusURL, cloudbridgeURL)

	if collector == nil {
		t.Fatal("Expected collector to be created, got nil")
	}

	if collector.prometheusURL != prometheusURL {
		t.Errorf("Expected prometheus URL %s, got %s", prometheusURL, collector.prometheusURL)
	}

	if collector.cloudbridgeURL != cloudbridgeURL {
		t.Errorf("Expected cloudbridge URL %s, got %s", cloudbridgeURL, collector.cloudbridgeURL)
	}

	if collector.updateInterval != 30*time.Second {
		t.Errorf("Expected update interval 30s, got %v", collector.updateInterval)
	}
}

func TestInitializeNodes(t *testing.T) {
	logger := zaptest.NewLogger(t)
	collector := NewNetworkMetricsCollector(logger, "http://localhost:9090", "http://localhost:8080")

	nodes := []*Node{
		{ID: "node1", Location: "US-East", Capacity: 1000},
		{ID: "node2", Location: "US-West", Capacity: 1000},
		{ID: "node3", Location: "EU-Central", Capacity: 1000},
	}

	collector.InitializeNodes(nodes)

	if len(collector.metrics) != 3 {
		t.Errorf("Expected 3 nodes, got %d", len(collector.metrics))
	}

	for _, node := range nodes {
		metrics, exists := collector.metrics[node.ID]
		if !exists {
			t.Errorf("Expected metrics for node %s", node.ID)
			continue
		}

		if metrics.NodeID != node.ID {
			t.Errorf("Expected node ID %s, got %s", node.ID, metrics.NodeID)
		}

		if metrics.Location != node.Location {
			t.Errorf("Expected location %s, got %s", node.Location, metrics.Location)
		}

		if metrics.CPUUsage != 0.0 {
			t.Errorf("Expected initial CPU usage 0.0, got %f", metrics.CPUUsage)
		}
	}
}

func TestCollectMetricsWithMockServers(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Создаем mock серверы
	prometheusServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/query" {
			response := `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"value": [1640995200, "150"]
						}
					]
				}
			}`
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		}
	}))
	defer prometheusServer.Close()

	relayServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/stats" {
			response := `{
				"bandwidth": 1000.0,
				"latency": 25.5,
				"error_rate": 0.01
			}`
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		}
	}))
	defer relayServer.Close()

	collector := NewNetworkMetricsCollector(logger, prometheusServer.URL, relayServer.URL)

	// Инициализируем узлы
	nodes := []*Node{
		{ID: "node1", Location: "US-East", Capacity: 1000},
	}
	collector.InitializeNodes(nodes)

	// Собираем метрики
	err := collector.CollectMetrics()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Проверяем, что метрики обновились
	metrics, exists := collector.GetNodeMetrics("node1")
	if !exists {
		t.Fatal("Expected metrics for node1")
	}

	if metrics.ActiveConns != 150 {
		t.Errorf("Expected active connections 150, got %d", metrics.ActiveConns)
	}

	if metrics.Bandwidth != 1000.0 {
		t.Errorf("Expected bandwidth 1000.0, got %f", metrics.Bandwidth)
	}

	if metrics.Latency != 25.5 {
		t.Errorf("Expected latency 25.5, got %f", metrics.Latency)
	}

	if metrics.ErrorRate != 0.01 {
		t.Errorf("Expected error rate 0.01, got %f", metrics.ErrorRate)
	}
}

func TestCollectMetricsWithServerErrors(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Создаем сервер, который возвращает ошибки
	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer errorServer.Close()

	collector := NewNetworkMetricsCollector(logger, errorServer.URL, errorServer.URL)

	// Инициализируем узлы
	nodes := []*Node{
		{ID: "node1", Location: "US-East", Capacity: 1000},
	}
	collector.InitializeNodes(nodes)

	// Собираем метрики - должно работать даже с ошибками серверов
	err := collector.CollectMetrics()
	if err != nil {
		t.Fatalf("Expected no error even with server errors, got %v", err)
	}

	// Проверяем, что метрики все еще доступны (с fallback значениями)
	metrics, exists := collector.GetNodeMetrics("node1")
	if !exists {
		t.Fatal("Expected metrics for node1 even with server errors")
	}

	// Метрики должны иметь значения по умолчанию
	if metrics.NodeID != "node1" {
		t.Errorf("Expected node ID node1, got %s", metrics.NodeID)
	}
}

func TestGetAllNodeMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t)
	collector := NewNetworkMetricsCollector(logger, "http://localhost:9090", "http://localhost:8080")

	nodes := []*Node{
		{ID: "node1", Location: "US-East", Capacity: 1000},
		{ID: "node2", Location: "US-West", Capacity: 1000},
	}
	collector.InitializeNodes(nodes)

	allMetrics := collector.GetAllNodeMetrics()

	if len(allMetrics) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(allMetrics))
	}

	for _, node := range nodes {
		if _, exists := allMetrics[node.ID]; !exists {
			t.Errorf("Expected metrics for node %s", node.ID)
		}
	}
}

func TestGetNodeMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t)
	collector := NewNetworkMetricsCollector(logger, "http://localhost:9090", "http://localhost:8080")

	nodes := []*Node{
		{ID: "node1", Location: "US-East", Capacity: 1000},
	}
	collector.InitializeNodes(nodes)

	// Тест существующего узла
	metrics, exists := collector.GetNodeMetrics("node1")
	if !exists {
		t.Fatal("Expected metrics for node1")
	}
	if metrics.NodeID != "node1" {
		t.Errorf("Expected node ID node1, got %s", metrics.NodeID)
	}

	// Тест несуществующего узла
	_, exists = collector.GetNodeMetrics("nonexistent")
	if exists {
		t.Error("Expected no metrics for nonexistent node")
	}
}

func TestStartPeriodicCollection(t *testing.T) {
	logger := zaptest.NewLogger(t)
	collector := NewNetworkMetricsCollector(logger, "http://localhost:9090", "http://localhost:8080")

	// Инициализируем узлы
	nodes := []*Node{
		{ID: "node1", Location: "US-East", Capacity: 1000},
	}
	collector.InitializeNodes(nodes)

	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Запускаем периодический сбор
	done := make(chan struct{})
	go func() {
		collector.StartPeriodicCollection(ctx)
		close(done)
	}()

	// Ждем завершения
	select {
	case <-done:
		// Ожидаемо завершилось по таймауту
	case <-time.After(200 * time.Millisecond):
		t.Error("Expected periodic collection to stop after context timeout")
	}
}

func TestGetLastUpdate(t *testing.T) {
	logger := zaptest.NewLogger(t)
	collector := NewNetworkMetricsCollector(logger, "http://localhost:9090", "http://localhost:8080")

	// Изначально время обновления должно быть нулевым
	lastUpdate := collector.GetLastUpdate()
	if !lastUpdate.IsZero() {
		t.Errorf("Expected zero time initially, got %v", lastUpdate)
	}

	// Инициализируем узлы и собираем метрики
	nodes := []*Node{
		{ID: "node1", Location: "US-East", Capacity: 1000},
	}
	collector.InitializeNodes(nodes)

	// Создаем mock сервер для успешного сбора
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"status": "success",
			"data": {
				"resultType": "vector",
				"result": [
					{
						"value": [1640995200, "100"]
					}
				]
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	// Обновляем URL и собираем метрики
	collector.prometheusURL = server.URL
	collector.cloudbridgeURL = server.URL

	err := collector.CollectMetrics()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Теперь время обновления должно быть установлено
	lastUpdate = collector.GetLastUpdate()
	if lastUpdate.IsZero() {
		t.Error("Expected non-zero update time after collecting metrics")
	}

	// Проверяем, что время обновления недавнее
	if time.Since(lastUpdate) > time.Minute {
		t.Errorf("Expected recent update time, got %v", lastUpdate)
	}
}

func TestGetUpdateInterval(t *testing.T) {
	logger := zaptest.NewLogger(t)
	collector := NewNetworkMetricsCollector(logger, "http://localhost:9090", "http://localhost:8080")

	interval := collector.GetUpdateInterval()
	expected := 30 * time.Second

	if interval != expected {
		t.Errorf("Expected update interval %v, got %v", expected, interval)
	}
}

func TestCollectMetricsConcurrency(t *testing.T) {
	logger := zaptest.NewLogger(t)
	collector := NewNetworkMetricsCollector(logger, "http://localhost:9090", "http://localhost:8080")

	nodes := []*Node{
		{ID: "node1", Location: "US-East", Capacity: 1000},
		{ID: "node2", Location: "US-West", Capacity: 1000},
	}
	collector.InitializeNodes(nodes)

	// Запускаем несколько горутин для сбора метрик одновременно
	done := make(chan struct{})
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			collector.CollectMetrics()
		}()
	}

	// Ждем завершения всех горутин
	for i := 0; i < numGoroutines; i++ {
		select {
		case <-done:
			// Горутина завершилась
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for goroutines to complete")
		}
	}

	// Проверяем, что метрики все еще доступны
	allMetrics := collector.GetAllNodeMetrics()
	if len(allMetrics) != 2 {
		t.Errorf("Expected 2 nodes after concurrent access, got %d", len(allMetrics))
	}
}

func TestNodeMetricsStructure(t *testing.T) {
	logger := zaptest.NewLogger(t)
	collector := NewNetworkMetricsCollector(logger, "http://localhost:9090", "http://localhost:8080")

	nodes := []*Node{
		{ID: "test-node", Location: "Test-Location", Capacity: 2000},
	}
	collector.InitializeNodes(nodes)

	metrics, exists := collector.GetNodeMetrics("test-node")
	if !exists {
		t.Fatal("Expected metrics for test-node")
	}

	// Проверяем все поля структуры
	expectedFields := map[string]interface{}{
		"NodeID":      "test-node",
		"Location":    "Test-Location",
		"CPUUsage":    0.0,
		"MemoryUsage": 0.0,
		"NetworkLoad": 0.0,
		"ActiveConns": 0,
		"Bandwidth":   0.0,
		"Latency":     0.0,
		"ErrorRate":   0.0,
		"Health":      "unknown",
	}

	if metrics.NodeID != expectedFields["NodeID"] {
		t.Errorf("Expected NodeID %v, got %v", expectedFields["NodeID"], metrics.NodeID)
	}

	if metrics.Location != expectedFields["Location"] {
		t.Errorf("Expected Location %v, got %v", expectedFields["Location"], metrics.Location)
	}

	if metrics.CPUUsage != expectedFields["CPUUsage"] {
		t.Errorf("Expected CPUUsage %v, got %v", expectedFields["CPUUsage"], metrics.CPUUsage)
	}

	if metrics.MemoryUsage != expectedFields["MemoryUsage"] {
		t.Errorf("Expected MemoryUsage %v, got %v", expectedFields["MemoryUsage"], metrics.MemoryUsage)
	}

	if metrics.NetworkLoad != expectedFields["NetworkLoad"] {
		t.Errorf("Expected NetworkLoad %v, got %v", expectedFields["NetworkLoad"], metrics.NetworkLoad)
	}

	if metrics.ActiveConns != expectedFields["ActiveConns"] {
		t.Errorf("Expected ActiveConns %v, got %v", expectedFields["ActiveConns"], metrics.ActiveConns)
	}

	if metrics.Bandwidth != expectedFields["Bandwidth"] {
		t.Errorf("Expected Bandwidth %v, got %v", expectedFields["Bandwidth"], metrics.Bandwidth)
	}

	if metrics.Latency != expectedFields["Latency"] {
		t.Errorf("Expected Latency %v, got %v", expectedFields["Latency"], metrics.Latency)
	}

	if metrics.ErrorRate != expectedFields["ErrorRate"] {
		t.Errorf("Expected ErrorRate %v, got %v", expectedFields["ErrorRate"], metrics.ErrorRate)
	}

	if metrics.Health != expectedFields["Health"] {
		t.Errorf("Expected Health %v, got %v", expectedFields["Health"], metrics.Health)
	}

	// Проверяем, что LastUpdate установлено
	if metrics.LastUpdate.IsZero() {
		t.Error("Expected LastUpdate to be set")
	}
}
