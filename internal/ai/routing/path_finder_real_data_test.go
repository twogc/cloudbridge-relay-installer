package routing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"
)

func TestNewPathFinderWithRealData(t *testing.T) {
	logger := zaptest.NewLogger(t)
	prometheusURL := "http://localhost:9090"
	cloudbridgeURL := "http://localhost:8080"

	pf := NewPathFinder(logger, prometheusURL, cloudbridgeURL)

	if pf == nil {
		t.Fatal("Expected PathFinder to be created, got nil")
	}

	if pf.metricsCollector == nil {
		t.Fatal("Expected metrics collector to be initialized")
	}

	if pf.topology == nil {
		t.Fatal("Expected topology to be initialized")
	}
}

func TestPathFinderInitializeWithRealData(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Создаем mock серверы
	prometheusServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	defer prometheusServer.Close()

	relayServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"bandwidth": 1000.0,
			"latency": 25.5,
			"error_rate": 0.01
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer relayServer.Close()

	pf := NewPathFinder(logger, prometheusServer.URL, relayServer.URL)

	err := pf.Initialize()
	if err != nil {
		t.Fatalf("Expected no error during initialization, got %v", err)
	}

	// Проверяем, что топология инициализирована
	if len(pf.topology.Nodes) == 0 {
		t.Error("Expected topology nodes to be initialized")
	}

	if len(pf.topology.Edges) == 0 {
		t.Error("Expected topology edges to be initialized")
	}

	// Проверяем, что метрики инициализированы
	allMetrics := pf.metricsCollector.GetAllNodeMetrics()
	if len(allMetrics) == 0 {
		t.Error("Expected node metrics to be initialized")
	}
}

func TestFindOptimalPathWithRealMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Создаем mock серверы
	prometheusServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"status": "success",
			"data": {
				"resultType": "vector",
				"result": [
					{
						"value": [1640995200, "50"]
					}
				]
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer prometheusServer.Close()

	relayServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"bandwidth": 1000.0,
			"latency": 20.0,
			"error_rate": 0.005
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer relayServer.Close()

	pf := NewPathFinder(logger, prometheusServer.URL, relayServer.URL)
	err := pf.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize PathFinder: %v", err)
	}

	// Собираем метрики для узлов
	err = pf.metricsCollector.CollectMetrics()
	if err != nil {
		t.Fatalf("Failed to collect metrics: %v", err)
	}

	// Тестируем поиск пути
	path, err := pf.FindOptimalPath("node1", "node2", "QUIC")
	if err != nil {
		t.Fatalf("Expected no error finding path, got %v", err)
	}

	if len(path) == 0 {
		t.Error("Expected non-empty path")
	}

	// Проверяем, что путь начинается с источника и заканчивается назначением
	if path[0] != "node1" {
		t.Errorf("Expected path to start with node1, got %s", path[0])
	}

	if path[len(path)-1] != "node2" {
		t.Errorf("Expected path to end with node2, got %s", path[len(path)-1])
	}
}

func TestCalculateNodeLoadWithRealMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Создаем mock серверы с разными значениями метрик
	prometheusServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"status": "success",
			"data": {
				"resultType": "vector",
				"result": [
					{
						"value": [1640995200, "200"]
					}
				]
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer prometheusServer.Close()

	relayServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"bandwidth": 2000.0,
			"latency": 15.0,
			"error_rate": 0.02
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer relayServer.Close()

	pf := NewPathFinder(logger, prometheusServer.URL, relayServer.URL)
	err := pf.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize PathFinder: %v", err)
	}

	// Собираем метрики
	err = pf.metricsCollector.CollectMetrics()
	if err != nil {
		t.Fatalf("Failed to collect metrics: %v", err)
	}

	// Тестируем расчет загрузки узла
	load := pf.calculateNodeLoad("node1")

	// Загрузка должна быть рассчитана на основе реальных метрик
	if load < 0 {
		t.Errorf("Expected non-negative load, got %d", load)
	}

	if load > 1000 {
		t.Errorf("Expected load <= 1000, got %d", load)
	}

	// Проверяем, что загрузка учитывает реальные метрики
	// С 200 активными соединениями, загрузка должна быть значительной
	if load < 50 {
		t.Errorf("Expected higher load with 200 connections, got %d", load)
	}
}

func TestCalculateEdgeLatencyWithRealMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Создаем mock серверы
	prometheusServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	defer prometheusServer.Close()

	relayServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"bandwidth": 1000.0,
			"latency": 30.0,
			"error_rate": 0.01
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer relayServer.Close()

	pf := NewPathFinder(logger, prometheusServer.URL, relayServer.URL)
	err := pf.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize PathFinder: %v", err)
	}

	// Собираем метрики
	err = pf.metricsCollector.CollectMetrics()
	if err != nil {
		t.Fatalf("Failed to collect metrics: %v", err)
	}

	// Создаем тестовое ребро
	edge := &Edge{
		From:      "node1",
		To:        "node2",
		Latency:   50.0,
		Bandwidth: 1000,
		Protocol:  "QUIC",
	}

	// Тестируем расчет латентности ребра
	latency := pf.calculateEdgeLatency(edge)

	// Латентность должна быть рассчитана на основе реальных метрик
	if latency <= 0 {
		t.Errorf("Expected positive latency, got %f", latency)
	}

	// Для QUIC протокола латентность должна быть уменьшена (коэффициент 0.8)
	expectedMinLatency := 50.0 * 0.8 // Базовый коэффициент для QUIC
	if latency < expectedMinLatency {
		t.Errorf("Expected latency >= %f for QUIC, got %f", expectedMinLatency, latency)
	}
}

func TestPathFinderUpdateWithRealMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Создаем mock серверы
	prometheusServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	}))
	defer prometheusServer.Close()

	relayServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"bandwidth": 1500.0,
			"latency": 20.0,
			"error_rate": 0.005
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer relayServer.Close()

	pf := NewPathFinder(logger, prometheusServer.URL, relayServer.URL)
	err := pf.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize PathFinder: %v", err)
	}

	// Получаем начальные значения загрузки узлов
	initialLoads := make(map[string]int)
	for nodeID := range pf.topology.Nodes {
		initialLoads[nodeID] = pf.topology.Nodes[nodeID].Load
	}

	// Выполняем обновление
	pf.Update()

	// Проверяем, что загрузка узлов обновилась
	for nodeID, node := range pf.topology.Nodes {
		if node.Load == initialLoads[nodeID] {
			t.Errorf("Expected load for node %s to be updated", nodeID)
		}
	}

	// Проверяем, что латентность ребер обновилась
	for _, edges := range pf.topology.Edges {
		for _, edge := range edges {
			if edge.Latency <= 0 {
				t.Errorf("Expected positive latency for edge %s->%s, got %f", edge.From, edge.To, edge.Latency)
			}
		}
	}
}

func TestPathFinderStartMetricsCollection(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Создаем mock серверы
	prometheusServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	defer prometheusServer.Close()

	relayServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"bandwidth": 1000.0,
			"latency": 25.0,
			"error_rate": 0.01
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer relayServer.Close()

	pf := NewPathFinder(logger, prometheusServer.URL, relayServer.URL)
	err := pf.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize PathFinder: %v", err)
	}

	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Запускаем сбор метрик
	done := make(chan struct{})
	go func() {
		pf.StartMetricsCollection(ctx)
		close(done)
	}()

	// Ждем завершения
	select {
	case <-done:
		// Ожидаемо завершилось по таймауту
	case <-time.After(200 * time.Millisecond):
		t.Error("Expected metrics collection to stop after context timeout")
	}
}

func TestPathFinderGetMetricsCollector(t *testing.T) {
	logger := zaptest.NewLogger(t)
	pf := NewPathFinder(logger, "http://localhost:9090", "http://localhost:8080")

	collector := pf.GetMetricsCollector()
	if collector == nil {
		t.Fatal("Expected metrics collector to be returned")
	}

	// Проверяем, что это тот же коллектор
	if collector != pf.metricsCollector {
		t.Error("Expected same metrics collector instance")
	}
}

func TestPathFinderWithHighLoadNodes(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Создаем mock серверы с высокими значениями метрик
	prometheusServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"status": "success",
			"data": {
				"resultType": "vector",
				"result": [
					{
						"value": [1640995200, "1000"]
					}
				]
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer prometheusServer.Close()

	relayServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"bandwidth": 500.0,
			"latency": 100.0,
			"error_rate": 0.1
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer relayServer.Close()

	pf := NewPathFinder(logger, prometheusServer.URL, relayServer.URL)
	err := pf.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize PathFinder: %v", err)
	}

	// Собираем метрики
	err = pf.metricsCollector.CollectMetrics()
	if err != nil {
		t.Fatalf("Failed to collect metrics: %v", err)
	}

	// Тестируем расчет загрузки при высокой нагрузке
	load := pf.calculateNodeLoad("node1")

	// При высокой нагрузке загрузка должна быть близка к максимуму
	if load < 800 {
		t.Errorf("Expected high load with 1000 connections, got %d", load)
	}

	// Но не должна превышать максимум
	if load > 1000 {
		t.Errorf("Expected load <= 1000, got %d", load)
	}
}

func TestPathFinderWithLowLoadNodes(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Создаем mock серверы с низкими значениями метрик
	prometheusServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"status": "success",
			"data": {
				"resultType": "vector",
				"result": [
					{
						"value": [1640995200, "10"]
					}
				]
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer prometheusServer.Close()

	relayServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"bandwidth": 2000.0,
			"latency": 5.0,
			"error_rate": 0.001
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer relayServer.Close()

	pf := NewPathFinder(logger, prometheusServer.URL, relayServer.URL)
	err := pf.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize PathFinder: %v", err)
	}

	// Собираем метрики
	err = pf.metricsCollector.CollectMetrics()
	if err != nil {
		t.Fatalf("Failed to collect metrics: %v", err)
	}

	// Тестируем расчет загрузки при низкой нагрузке
	load := pf.calculateNodeLoad("node1")

	// При низкой нагрузке загрузка должна быть небольшой
	if load > 200 {
		t.Errorf("Expected low load with 10 connections, got %d", load)
	}

	// Но должна быть неотрицательной
	if load < 0 {
		t.Errorf("Expected non-negative load, got %d", load)
	}
}
