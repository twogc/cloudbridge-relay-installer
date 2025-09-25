package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/ai/anomaly"
	"go.uber.org/zap/zaptest"
)

func TestErrorHandlingWithUnavailableServices(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()

	// Создаем серверы, которые возвращают ошибки
	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer errorServer.Close()

	// Обновляем конфигурацию с URL серверов, которые возвращают ошибки
	config.AI.DataSources.PrometheusURL = errorServer.URL
	config.AI.DataSources.CloudBridgeURL = errorServer.URL
	config.AI.DataSources.CadenceURL = errorServer.URL

	service := NewRealService(config, logger)

	// Тестируем инициализацию с недоступными сервисами
	err := service.Initialize()
	if err != nil {
		t.Fatalf("Expected no error during initialization even with unavailable services, got %v", err)
	}

	// Тестируем запуск сервиса
	err = service.Start()
	if err != nil {
		t.Fatalf("Expected no error starting service even with unavailable services, got %v", err)
	}

	// Ждем немного, чтобы горутины запустились
	time.Sleep(100 * time.Millisecond)

	// Тестируем операции с недоступными сервисами
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

	// Обнаружение аномалий должно работать даже с недоступными сервисами
	result, err := service.DetectAnomaly(metrics)
	if err != nil {
		t.Fatalf("Expected no error detecting anomaly even with unavailable services, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected anomaly result to be returned")
	}

	// Прогнозирование нагрузки должно работать
	prediction, err := service.PredictLoad(24)
	if err != nil {
		t.Fatalf("Expected no error predicting load even with unavailable services, got %v", err)
	}

	if prediction == nil {
		t.Fatal("Expected load prediction to be returned")
	}

	// Оптимизация маршрутов должна работать
	route, err := service.OptimizeRoute("node1", "node2", map[string]interface{}{
		"protocol": "QUIC",
	})
	if err != nil {
		t.Fatalf("Expected no error optimizing route even with unavailable services, got %v", err)
	}

	if route == nil {
		t.Fatal("Expected optimized route to be returned")
	}

	// Останавливаем сервис
	err = service.Stop()
	if err != nil {
		t.Fatalf("Expected no error stopping service, got %v", err)
	}
}

func TestErrorHandlingWithNetworkTimeout(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()

	// Создаем сервер, который не отвечает (имитация таймаута)
	timeoutServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Не отвечаем, чтобы вызвать таймаут
		time.Sleep(2 * time.Second)
	}))
	defer timeoutServer.Close()

	// Обновляем конфигурацию
	config.AI.DataSources.PrometheusURL = timeoutServer.URL
	config.AI.DataSources.CloudBridgeURL = timeoutServer.URL
	config.AI.DataSources.CadenceURL = timeoutServer.URL

	service := NewRealService(config, logger)

	// Инициализируем сервис
	err := service.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize service: %v", err)
	}

	// Тестируем операции с таймаутом сети
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

	// Обнаружение аномалий должно работать с fallback данными
	result, err := service.DetectAnomaly(metrics)
	if err != nil {
		t.Fatalf("Expected no error detecting anomaly with network timeout, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected anomaly result to be returned")
	}

	// Прогнозирование нагрузки должно работать с fallback данными
	prediction, err := service.PredictLoad(12)
	if err != nil {
		t.Fatalf("Expected no error predicting load with network timeout, got %v", err)
	}

	if prediction == nil {
		t.Fatal("Expected load prediction to be returned")
	}
}

func TestErrorHandlingWithInvalidData(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()

	// Создаем серверы, которые возвращают некорректные данные
	invalidDataServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Возвращаем некорректный JSON
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json data"))
	}))
	defer invalidDataServer.Close()

	// Обновляем конфигурацию
	config.AI.DataSources.PrometheusURL = invalidDataServer.URL
	config.AI.DataSources.CloudBridgeURL = invalidDataServer.URL
	config.AI.DataSources.CadenceURL = invalidDataServer.URL

	service := NewRealService(config, logger)

	// Инициализируем сервис
	err := service.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize service: %v", err)
	}

	// Тестируем операции с некорректными данными
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

	// Обнаружение аномалий должно работать с fallback данными
	result, err := service.DetectAnomaly(metrics)
	if err != nil {
		t.Fatalf("Expected no error detecting anomaly with invalid data, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected anomaly result to be returned")
	}

	// Прогнозирование нагрузки должно работать с fallback данными
	prediction, err := service.PredictLoad(12)
	if err != nil {
		t.Fatalf("Expected no error predicting load with invalid data, got %v", err)
	}

	if prediction == nil {
		t.Fatal("Expected load prediction to be returned")
	}
}

func TestErrorHandlingWithPartialServiceFailure(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()

	// Создаем серверы: один работает, другой возвращает ошибку
	workingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	defer workingServer.Close()

	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer errorServer.Close()

	// Обновляем конфигурацию: Prometheus работает, остальные возвращают ошибки
	config.AI.DataSources.PrometheusURL = workingServer.URL
	config.AI.DataSources.CloudBridgeURL = errorServer.URL
	config.AI.DataSources.CadenceURL = errorServer.URL

	service := NewRealService(config, logger)

	// Инициализируем сервис
	err := service.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize service: %v", err)
	}

	// Тестируем операции с частичным отказом сервисов
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

	// Обнаружение аномалий должно работать с частичными данными
	result, err := service.DetectAnomaly(metrics)
	if err != nil {
		t.Fatalf("Expected no error detecting anomaly with partial service failure, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected anomaly result to be returned")
	}

	// Прогнозирование нагрузки должно работать с частичными данными
	prediction, err := service.PredictLoad(12)
	if err != nil {
		t.Fatalf("Expected no error predicting load with partial service failure, got %v", err)
	}

	if prediction == nil {
		t.Fatal("Expected load prediction to be returned")
	}
}

func TestErrorHandlingWithContextCancellation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()

	// Создаем серверы
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

	// Обновляем конфигурацию
	config.AI.DataSources.PrometheusURL = server.URL
	config.AI.DataSources.CloudBridgeURL = server.URL
	config.AI.DataSources.CadenceURL = server.URL

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

	// Ждем немного
	time.Sleep(100 * time.Millisecond)

	// Тестируем отмену контекста
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Создаем новый сервис с отмененным контекстом для тестирования
	serviceWithCancel := NewRealService(config, logger)
	err = serviceWithCancel.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize service with cancel: %v", err)
	}

	// Запускаем сервис с отмененным контекстом
	serviceWithCancel.ctx = ctx
	serviceWithCancel.cancel = cancel

	err = serviceWithCancel.Start()
	if err != nil {
		t.Fatalf("Failed to start service with cancel: %v", err)
	}

	// Ждем отмены контекста
	time.Sleep(100 * time.Millisecond)

	// Останавливаем основной сервис
	err = service.Stop()
	if err != nil {
		t.Fatalf("Failed to stop service: %v", err)
	}
}

func TestErrorHandlingWithInvalidConfiguration(t *testing.T) {
	_ = zaptest.NewLogger(t)

	// Создаем конфигурацию с некорректными значениями
	config := &Config{
		AI: AI{
			Enabled:               true,
			Mode:                  "invalid-mode", // Некорректный режим
			LogLevel:              "info",
			MaxConcurrentRequests: 100,
			RequestTimeout:        30 * time.Second,
			CacheTTL:              5 * time.Minute,
			Models: Models{
				AnomalyDetection: AnomalyDetection{
					Enabled:       true,
					Algorithm:     "isolation_forest",
					Threshold:     1.5, // Некорректное значение (должно быть 0-1)
					LearningRate:  0.01,
					MaxSamples:    10000,
					Contamination: 0.1,
				},
				LoadPrediction: LoadPrediction{
					Enabled:   true,
					Algorithm: "arima",
					ARIMAParams: ARIMAParams{
						P: 2,
						D: 1,
						Q: 2,
					},
					ForecastHorizon: 24,
					ConfidenceLevel: 1.5, // Некорректное значение (должно быть 0-1)
				},
				RouteOptimization: RouteOptimization{
					Enabled:         true,
					Algorithm:       "multi_armed_bandit",
					ExplorationRate: 0.1,
					LearningRate:    0.05,
					UpdateInterval:  time.Minute,
				},
			},
			DataPipeline: DataPipeline{
				Enabled:            true,
				BatchSize:          1000,
				ProcessingInterval: 60 * time.Second,
				RetentionDays:      30,
				Features: Features{
					Network:  []string{"bandwidth_utilization", "latency", "packet_loss", "connection_count", "protocol_distribution"},
					System:   []string{"cpu_usage", "memory_usage", "disk_io", "network_io"},
					Security: []string{"failed_connections", "suspicious_patterns", "rate_limiting", "authentication_failures"},
				},
			},
			Monitoring: Monitoring{
				Enabled:         true,
				MetricsInterval: 30 * time.Second,
				AlertThresholds: AlertThresholds{
					AnomalyScore:       0.9,
					PredictionAccuracy: 0.7,
					ProcessingLatency:  100 * time.Millisecond,
				},
			},
			Storage: Storage{
				Type:        "file",
				Path:        "/data/ai",
				MaxSize:     "1GB",
				Compression: true,
			},
			API: API{
				Enabled:     true,
				Port:        99999, // Некорректный порт
				CORSEnabled: true,
				RateLimit:   1000,
			},
			Development: Development{
				Debug:    false,
				MockData: false,
				TestMode: false,
			},
		},
	}

	// Тестируем валидацию конфигурации
	err := config.Validate()
	if err == nil {
		t.Error("Expected validation error for invalid configuration")
	}

	// Проверяем, что ошибка содержит информацию о некорректных полях
	errorMsg := err.Error()
	if errorMsg == "" {
		t.Error("Expected non-empty error message")
	}
}

func TestErrorHandlingWithNilInputs(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()

	service := NewRealService(config, logger)

	// Тестируем операции с nil входными данными
	_, err := service.DetectAnomaly(nil)
	if err == nil {
		t.Error("Expected error for nil metrics input")
	}

	// Тестируем операции до инициализации
	_, err = service.PredictLoad(12)
	if err == nil {
		t.Error("Expected error for prediction before initialization")
	}

	_, err = service.OptimizeRoute("", "", nil)
	if err == nil {
		t.Error("Expected error for route optimization with empty inputs")
	}

	// Тестируем получение статуса до инициализации
	status := service.GetStatus()
	if status == nil {
		t.Fatal("Expected status to be returned even before initialization")
	}

	// Тестируем проверку здоровья до инициализации
	err = service.HealthCheck()
	if err == nil {
		t.Error("Expected error for health check before initialization")
	}
}

func TestErrorHandlingWithServiceAlreadyStarted(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultConfig()

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

	// Пытаемся запустить сервис повторно
	err = service.Start()
	if err == nil {
		t.Error("Expected error for starting already started service")
	}

	// Останавливаем сервис
	err = service.Stop()
	if err != nil {
		t.Fatalf("Failed to stop service: %v", err)
	}

	// Пытаемся остановить сервис повторно
	err = service.Stop()
	if err == nil {
		t.Error("Expected error for stopping already stopped service")
	}
}
