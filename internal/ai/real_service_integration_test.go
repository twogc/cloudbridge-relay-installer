package ai

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/ai/anomaly"
	"go.uber.org/zap/zaptest"
)

func TestNewRealService(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()

	service := NewRealService(config, logger)

	if service == nil {
		t.Fatal("Expected RealService to be created, got nil")
	}

	if service.config != config {
		t.Error("Expected config to be set")
	}

	if service.logger != logger {
		t.Error("Expected logger to be set")
	}

	if service.metrics == nil {
		t.Fatal("Expected metrics to be initialized")
	}

	if service.started {
		t.Error("Expected service to not be started initially")
	}
}

func TestRealServiceInitializeWithMockServers(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()

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

	cadenceServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"executions": []
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer cadenceServer.Close()

	// Обновляем конфигурацию с URL mock серверов
	config.AI.DataSources.PrometheusURL = prometheusServer.URL
	config.AI.DataSources.CloudBridgeURL = relayServer.URL
	config.AI.DataSources.CadenceURL = cadenceServer.URL

	service := NewRealService(config, logger)

	// Тестируем инициализацию
	err := service.Initialize()
	if err != nil {
		t.Fatalf("Expected no error during initialization, got %v", err)
	}

	// Проверяем, что все компоненты инициализированы
	if service.realDataCollector == nil {
		t.Error("Expected real data collector to be initialized")
	}

	if service.realAnomalyDetector == nil {
		t.Error("Expected real anomaly detector to be initialized")
	}

	if service.realLoadPredictor == nil {
		t.Error("Expected real load predictor to be initialized")
	}

	if service.cadenceIntegration == nil {
		t.Error("Expected cadence integration to be initialized")
	}

	if service.pathFinder == nil {
		t.Error("Expected path finder to be initialized")
	}

	if service.latencyPredictor == nil {
		t.Error("Expected latency predictor to be initialized")
	}

	if service.monitoring == nil {
		t.Error("Expected monitoring to be initialized")
	}
}

func TestRealServiceStartAndStop(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()

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

	cadenceServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"executions": []
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer cadenceServer.Close()

	// Обновляем конфигурацию
	config.AI.DataSources.PrometheusURL = prometheusServer.URL
	config.AI.DataSources.CloudBridgeURL = relayServer.URL
	config.AI.DataSources.CadenceURL = cadenceServer.URL

	service := NewRealService(config, logger)

	// Инициализируем сервис
	err := service.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize service: %v", err)
	}

	// Тестируем запуск сервиса
	err = service.Start()
	if err != nil {
		t.Fatalf("Expected no error starting service, got %v", err)
	}

	if !service.started {
		t.Error("Expected service to be started")
	}

	// Ждем немного, чтобы горутины запустились
	time.Sleep(100 * time.Millisecond)

	// Тестируем остановку сервиса
	err = service.Stop()
	if err != nil {
		t.Fatalf("Expected no error stopping service, got %v", err)
	}

	if service.started {
		t.Error("Expected service to be stopped")
	}
}

func TestRealServiceDetectAnomaly(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()

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

	cadenceServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"executions": []
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer cadenceServer.Close()

	// Обновляем конфигурацию
	config.AI.DataSources.PrometheusURL = prometheusServer.URL
	config.AI.DataSources.CloudBridgeURL = relayServer.URL
	config.AI.DataSources.CadenceURL = cadenceServer.URL

	service := NewRealService(config, logger)

	// Инициализируем сервис
	err := service.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize service: %v", err)
	}

	// Создаем тестовые метрики
	metrics := &anomaly.NetworkMetrics{
		Timestamp:            time.Now(),
		ConnectionsPerSecond: 100,
		BytesPerSecond:       1000000,
		PacketLossRate:       0.02,
		AverageLatency:       200.0,
		CPUUsage:             85.0,
		MemoryUsage:          90.0,
		ActiveConnections:    1000,
		FailedConnections:    50,
		ProtocolDistribution: map[string]int{
			"QUIC":     60,
			"P2P_QUIC": 20,
			"TCP":      15,
			"UDP":      5,
		},
	}

	// Тестируем обнаружение аномалий
	result, err := service.DetectAnomaly(metrics)
	if err != nil {
		t.Fatalf("Expected no error detecting anomaly, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected anomaly result to be returned")
	}

	// Проверяем структуру результата
	if result.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}

	if result.Score < 0 || result.Score > 1 {
		t.Errorf("Expected score between 0 and 1, got %f", result.Score)
	}

	if result.Confidence < 0 || result.Confidence > 1 {
		t.Errorf("Expected confidence between 0 and 1, got %f", result.Confidence)
	}
}

func TestRealServicePredictLoad(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()

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

	cadenceServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"executions": []
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer cadenceServer.Close()

	// Обновляем конфигурацию
	config.AI.DataSources.PrometheusURL = prometheusServer.URL
	config.AI.DataSources.CloudBridgeURL = relayServer.URL
	config.AI.DataSources.CadenceURL = cadenceServer.URL

	service := NewRealService(config, logger)

	// Инициализируем сервис
	err := service.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize service: %v", err)
	}

	// Тестируем прогнозирование нагрузки
	horizon := 24
	prediction, err := service.PredictLoad(horizon)
	if err != nil {
		t.Fatalf("Expected no error predicting load, got %v", err)
	}

	if prediction == nil {
		t.Fatal("Expected load prediction to be returned")
	}

	// Проверяем структуру прогноза
	if prediction.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}

	if prediction.Horizon != horizon {
		t.Errorf("Expected horizon %d, got %d", horizon, prediction.Horizon)
	}

	if prediction.Confidence < 0 || prediction.Confidence > 1 {
		t.Errorf("Expected confidence between 0 and 1, got %f", prediction.Confidence)
	}

	if prediction.ModelID == "" {
		t.Error("Expected model ID to be set")
	}
}

func TestRealServiceOptimizeRoute(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()

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

	cadenceServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"executions": []
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer cadenceServer.Close()

	// Обновляем конфигурацию
	config.AI.DataSources.PrometheusURL = prometheusServer.URL
	config.AI.DataSources.CloudBridgeURL = relayServer.URL
	config.AI.DataSources.CadenceURL = cadenceServer.URL

	service := NewRealService(config, logger)

	// Инициализируем сервис
	err := service.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize service: %v", err)
	}

	// Тестируем оптимизацию маршрута
	source := "node1"
	destination := "node2"
	constraints := map[string]interface{}{
		"protocol":      "QUIC",
		"max_latency":   50,
		"min_bandwidth": 1000,
	}

	route, err := service.OptimizeRoute(source, destination, constraints)
	if err != nil {
		t.Fatalf("Expected no error optimizing route, got %v", err)
	}

	if route == nil {
		t.Fatal("Expected optimized route to be returned")
	}

	// Проверяем структуру маршрута
	if route.Protocol != "QUIC" {
		t.Errorf("Expected protocol QUIC, got %s", route.Protocol)
	}

	if route.PredictedLatency <= 0 {
		t.Errorf("Expected positive predicted latency, got %f", route.PredictedLatency)
	}

	if route.Confidence < 0 || route.Confidence > 1 {
		t.Errorf("Expected confidence between 0 and 1, got %f", route.Confidence)
	}

	if route.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}

	if len(route.Path) == 0 {
		t.Error("Expected non-empty path")
	}

	// Проверяем, что путь начинается с источника и заканчивается назначением
	if route.Path[0] != source {
		t.Errorf("Expected path to start with %s, got %s", source, route.Path[0])
	}

	if route.Path[len(route.Path)-1] != destination {
		t.Errorf("Expected path to end with %s, got %s", destination, route.Path[len(route.Path)-1])
	}
}

func TestRealServiceGetStatus(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()

	service := NewRealService(config, logger)

	// Тестируем получение статуса до инициализации
	status := service.GetStatus()
	if status == nil {
		t.Fatal("Expected status to be returned")
	}

	// До инициализации все компоненты должны быть false
	if status.AnomalyDetector {
		t.Error("Expected anomaly detector to be false before initialization")
	}

	if status.LoadPredictor {
		t.Error("Expected load predictor to be false before initialization")
	}

	if status.RouteOptimizer {
		t.Error("Expected route optimizer to be false before initialization")
	}

	// Инициализируем сервис
	err := service.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize service: %v", err)
	}

	// Тестируем получение статуса после инициализации
	status = service.GetStatus()
	if status == nil {
		t.Fatal("Expected status to be returned")
	}

	// После инициализации все компоненты должны быть true
	if !status.AnomalyDetector {
		t.Error("Expected anomaly detector to be true after initialization")
	}

	if !status.LoadPredictor {
		t.Error("Expected load predictor to be true after initialization")
	}

	if !status.RouteOptimizer {
		t.Error("Expected route optimizer to be true after initialization")
	}

	if !status.Monitoring {
		t.Error("Expected monitoring to be true after initialization")
	}

	if !status.DataPipeline {
		t.Error("Expected data pipeline to be true after initialization")
	}
}

func TestRealServiceHealthCheck(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()

	service := NewRealService(config, logger)

	// Тестируем проверку здоровья до инициализации
	err := service.HealthCheck()
	if err == nil {
		t.Error("Expected error for health check before initialization")
	}

	// Инициализируем сервис
	err = service.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize service: %v", err)
	}

	// Тестируем проверку здоровья после инициализации
	err = service.HealthCheck()
	if err != nil {
		t.Fatalf("Expected no error for health check after initialization, got %v", err)
	}
}

func TestRealServiceGetRealDataInfo(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()

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

	cadenceServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"executions": []
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer cadenceServer.Close()

	// Обновляем конфигурацию
	config.AI.DataSources.PrometheusURL = prometheusServer.URL
	config.AI.DataSources.CloudBridgeURL = relayServer.URL
	config.AI.DataSources.CadenceURL = cadenceServer.URL

	service := NewRealService(config, logger)

	// Тестируем получение информации о данных до инициализации
	info := service.GetRealDataInfo()
	if info == nil {
		t.Fatal("Expected data info to be returned")
	}

	if _, exists := info["error"]; !exists {
		t.Error("Expected error to be present before initialization")
	}

	// Инициализируем сервис
	err := service.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize service: %v", err)
	}

	// Тестируем получение информации о данных после инициализации
	info = service.GetRealDataInfo()
	if info == nil {
		t.Fatal("Expected data info to be returned")
	}

	// Проверяем наличие ключевых полей
	expectedFields := []string{"last_data_points", "last_update", "data_sources", "update_interval", "sample_data"}
	for _, field := range expectedFields {
		if _, exists := info[field]; !exists {
			t.Errorf("Expected field %s to be present", field)
		}
	}
}

func TestRealServiceStartAnomalyDetectionWorkflow(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()

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

	cadenceServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"workflowId": "anomaly-detection-12345"
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer cadenceServer.Close()

	// Обновляем конфигурацию
	config.AI.DataSources.PrometheusURL = prometheusServer.URL
	config.AI.DataSources.CloudBridgeURL = relayServer.URL
	config.AI.DataSources.CadenceURL = cadenceServer.URL

	service := NewRealService(config, logger)

	// Инициализируем сервис
	err := service.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize service: %v", err)
	}

	// Тестируем запуск workflow обнаружения аномалий
	input := map[string]interface{}{
		"threshold": 0.8,
		"metrics":   []string{"cpu", "memory", "network"},
	}

	workflowID, err := service.StartAnomalyDetectionWorkflow(input)
	if err != nil {
		t.Fatalf("Expected no error starting anomaly detection workflow, got %v", err)
	}

	if workflowID == "" {
		t.Error("Expected workflow ID to be returned")
	}
}

func TestRealServiceConcurrentOperations(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()

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

	cadenceServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"executions": []
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer cadenceServer.Close()

	// Обновляем конфигурацию
	config.AI.DataSources.PrometheusURL = prometheusServer.URL
	config.AI.DataSources.CloudBridgeURL = relayServer.URL
	config.AI.DataSources.CadenceURL = cadenceServer.URL

	service := NewRealService(config, logger)

	// Инициализируем сервис
	err := service.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize service: %v", err)
	}

	// Тестируем параллельные операции
	done := make(chan struct{})
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer func() { done <- struct{}{} }()

			// Тестируем разные операции
			service.GetStatus()
			service.HealthCheck()
			service.GetRealDataInfo()
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
}
