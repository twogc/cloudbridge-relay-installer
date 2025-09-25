package ai

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/ai/anomaly"
	"go.uber.org/zap/zaptest"
)

func BenchmarkRealServiceAnomalyDetection(b *testing.B) {
	logger := zaptest.NewLogger(b)
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
		b.Fatalf("Failed to initialize service: %v", err)
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

	// Запускаем бенчмарк
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.DetectAnomaly(metrics)
		if err != nil {
			b.Fatalf("Error detecting anomaly: %v", err)
		}
	}
}

func BenchmarkRealServiceLoadPrediction(b *testing.B) {
	logger := zaptest.NewLogger(b)
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
		b.Fatalf("Failed to initialize service: %v", err)
	}

	// Запускаем бенчмарк
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.PredictLoad(24)
		if err != nil {
			b.Fatalf("Error predicting load: %v", err)
		}
	}
}

func BenchmarkRealServiceRouteOptimization(b *testing.B) {
	logger := zaptest.NewLogger(b)
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
		b.Fatalf("Failed to initialize service: %v", err)
	}

	constraints := map[string]interface{}{
		"protocol":      "QUIC",
		"max_latency":   50,
		"min_bandwidth": 1000,
	}

	// Запускаем бенчмарк
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.OptimizeRoute("node1", "node2", constraints)
		if err != nil {
			b.Fatalf("Error optimizing route: %v", err)
		}
	}
}

func TestRealServiceConcurrentPerformance(t *testing.T) {
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

	constraints := map[string]interface{}{
		"protocol":      "QUIC",
		"max_latency":   50,
		"min_bandwidth": 1000,
	}

	// Тестируем производительность при параллельных операциях
	numGoroutines := 50
	operationsPerGoroutine := 10

	var wg sync.WaitGroup
	startTime := time.Now()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				// Тестируем разные операции
				service.DetectAnomaly(metrics)
				service.PredictLoad(12)
				service.OptimizeRoute("node1", "node2", constraints)
				service.GetStatus()
				service.HealthCheck()
			}
		}()
	}

	wg.Wait()
	duration := time.Since(startTime)

	t.Logf("Completed %d operations across %d goroutines in %v",
		numGoroutines*operationsPerGoroutine*5, numGoroutines, duration)
	t.Logf("Average time per operation: %v",
		duration/time.Duration(numGoroutines*operationsPerGoroutine*5))

	// Проверяем, что все операции завершились за разумное время
	if duration > 30*time.Second {
		t.Errorf("Operations took too long: %v", duration)
	}
}

func TestRealServiceMemoryUsage(t *testing.T) {
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

	// Запускаем сервис
	err = service.Start()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}

	// Ждем немного, чтобы горутины запустились
	time.Sleep(100 * time.Millisecond)

	// Создаем много операций для тестирования использования памяти
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

	constraints := map[string]interface{}{
		"protocol":      "QUIC",
		"max_latency":   50,
		"min_bandwidth": 1000,
	}

	// Выполняем много операций
	for i := 0; i < 1000; i++ {
		service.DetectAnomaly(metrics)
		service.PredictLoad(12)
		service.OptimizeRoute("node1", "node2", constraints)
	}

	// Проверяем, что сервис все еще работает
	status := service.GetStatus()
	if status == nil {
		t.Fatal("Expected status to be returned")
	}

	// Останавливаем сервис
	err = service.Stop()
	if err != nil {
		t.Fatalf("Failed to stop service: %v", err)
	}
}

func TestRealServiceLatencyUnderLoad(t *testing.T) {
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

	_ = map[string]interface{}{
		"protocol":      "QUIC",
		"max_latency":   50,
		"min_bandwidth": 1000,
	}

	// Тестируем латентность при нагрузке
	numOperations := 100
	var totalLatency time.Duration

	for i := 0; i < numOperations; i++ {
		start := time.Now()

		// Выполняем операцию обнаружения аномалий
		_, err := service.DetectAnomaly(metrics)
		if err != nil {
			t.Fatalf("Error detecting anomaly: %v", err)
		}

		latency := time.Since(start)
		totalLatency += latency

		// Проверяем, что латентность в разумных пределах
		if latency > 1*time.Second {
			t.Errorf("Operation took too long: %v", latency)
		}
	}

	averageLatency := totalLatency / time.Duration(numOperations)
	t.Logf("Average latency for anomaly detection: %v", averageLatency)

	// Проверяем, что средняя латентность приемлема
	if averageLatency > 100*time.Millisecond {
		t.Errorf("Average latency too high: %v", averageLatency)
	}
}

func TestRealServiceThroughput(t *testing.T) {
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

	constraints := map[string]interface{}{
		"protocol":      "QUIC",
		"max_latency":   50,
		"min_bandwidth": 1000,
	}

	// Тестируем пропускную способность
	testDuration := 5 * time.Second
	startTime := time.Now()
	operationCount := 0

	for time.Since(startTime) < testDuration {
		// Выполняем разные операции
		service.DetectAnomaly(metrics)
		service.PredictLoad(12)
		service.OptimizeRoute("node1", "node2", constraints)
		operationCount += 3
	}

	actualDuration := time.Since(startTime)
	throughput := float64(operationCount) / actualDuration.Seconds()

	t.Logf("Completed %d operations in %v", operationCount, actualDuration)
	t.Logf("Throughput: %.2f operations/second", throughput)

	// Проверяем, что пропускная способность приемлема
	if throughput < 10 {
		t.Errorf("Throughput too low: %.2f operations/second", throughput)
	}
}
