package routing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"
)

func TestNewLatencyPredictorWithRealData(t *testing.T) {
	logger := zaptest.NewLogger(t)
	prometheusURL := "http://localhost:9090"
	cloudbridgeURL := "http://localhost:8080"

	lp := NewLatencyPredictor(logger, prometheusURL, cloudbridgeURL)

	if lp == nil {
		t.Fatal("Expected LatencyPredictor to be created, got nil")
	}

	if lp.metricsCollector == nil {
		t.Fatal("Expected metrics collector to be initialized")
	}

	// Модели инициализируются в методе Initialize, а не в конструкторе
	// Проверяем, что объект создан корректно
	if lp.logger == nil {
		t.Fatal("Expected logger to be initialized")
	}
}

func TestLatencyPredictorInitializeWithRealData(t *testing.T) {
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

	lp := NewLatencyPredictor(logger, prometheusServer.URL, relayServer.URL)

	err := lp.Initialize()
	if err != nil {
		t.Fatalf("Expected no error during initialization, got %v", err)
	}

	// Проверяем, что модели инициализированы
	expectedProtocols := []string{"QUIC", "P2P_QUIC", "TCP", "UDP"}
	for _, protocol := range expectedProtocols {
		if _, exists := lp.models[protocol]; !exists {
			t.Errorf("Expected model for protocol %s to be initialized", protocol)
		}
	}
}

func TestPredictWithRealMetrics(t *testing.T) {
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
			"bandwidth": 1000.0,
			"latency": 30.0,
			"error_rate": 0.02
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer relayServer.Close()

	lp := NewLatencyPredictor(logger, prometheusServer.URL, relayServer.URL)
	err := lp.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize LatencyPredictor: %v", err)
	}

	// Собираем метрики
	err = lp.metricsCollector.CollectMetrics()
	if err != nil {
		t.Fatalf("Failed to collect metrics: %v", err)
	}

	// Тестируем предсказание латентности
	latency, err := lp.Predict("node1", "node2", "QUIC")
	if err != nil {
		t.Fatalf("Expected no error predicting latency, got %v", err)
	}

	// Латентность должна быть положительной
	if latency <= 0 {
		t.Errorf("Expected positive latency, got %f", latency)
	}

	// Латентность должна быть в разумных пределах
	if latency > 1000 {
		t.Errorf("Expected latency <= 1000ms, got %f", latency)
	}
}

func TestGetLoadBasedLatencyWithRealMetrics(t *testing.T) {
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
			"latency": 20.0,
			"error_rate": 0.005
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer relayServer.Close()

	lp := NewLatencyPredictor(logger, prometheusServer.URL, relayServer.URL)
	err := lp.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize LatencyPredictor: %v", err)
	}

	// Собираем метрики
	err = lp.metricsCollector.CollectMetrics()
	if err != nil {
		t.Fatalf("Failed to collect metrics: %v", err)
	}

	// Тестируем расчет латентности на основе нагрузки
	loadLatency := lp.getLoadBasedLatency()

	// Латентность на основе нагрузки должна быть неотрицательной
	if loadLatency < 0 {
		t.Errorf("Expected non-negative load-based latency, got %f", loadLatency)
	}

	// При высокой нагрузке латентность должна быть значительной
	if loadLatency > 20 {
		t.Errorf("Expected load-based latency <= 20ms, got %f", loadLatency)
	}
}

func TestCalculateRealVariationWithRealMetrics(t *testing.T) {
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

	lp := NewLatencyPredictor(logger, prometheusServer.URL, relayServer.URL)
	err := lp.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize LatencyPredictor: %v", err)
	}

	// Собираем метрики
	err = lp.metricsCollector.CollectMetrics()
	if err != nil {
		t.Fatalf("Failed to collect metrics: %v", err)
	}

	// Создаем тестовую модель
	model := &LatencyModel{
		Protocol:    "QUIC",
		BaseLatency: 10.0,
		Variation:   2.0,
		LastUpdate:  time.Now(),
	}

	// Тестируем расчет реальной вариации
	variation := lp.calculateRealVariation("node1", "node2", model)

	// Вариация должна быть неотрицательной
	if variation < 0 {
		t.Errorf("Expected non-negative variation, got %f", variation)
	}

	// Вариация должна быть пропорциональна базовой вариации модели
	if variation > model.Variation {
		t.Errorf("Expected variation <= model variation (%f), got %f", model.Variation, variation)
	}
}

func TestCalculateNodeStability(t *testing.T) {
	logger := zaptest.NewLogger(t)
	lp := NewLatencyPredictor(logger, "http://localhost:9090", "http://localhost:8080")

	// Тестируем стабильный узел
	stableMetrics := &NodeMetrics{
		NodeID:      "stable-node",
		CPUUsage:    20.0,  // Низкая загрузка
		MemoryUsage: 30.0,  // Низкая загрузка
		NetworkLoad: 10.0,  // Низкая загрузка
		ErrorRate:   0.001, // Низкая ошибка
		LastUpdate:  time.Now(),
	}

	stability := lp.calculateNodeStability(stableMetrics)

	// Стабильный узел должен иметь низкую нестабильность
	if stability > 0.5 {
		t.Errorf("Expected low instability for stable node, got %f", stability)
	}

	// Тестируем нестабильный узел
	unstableMetrics := &NodeMetrics{
		NodeID:      "unstable-node",
		CPUUsage:    90.0, // Высокая загрузка
		MemoryUsage: 85.0, // Высокая загрузка
		NetworkLoad: 95.0, // Высокая загрузка
		ErrorRate:   0.1,  // Высокая ошибка
		LastUpdate:  time.Now(),
	}

	stability = lp.calculateNodeStability(unstableMetrics)

	// Нестабильный узел должен иметь высокую нестабильность
	if stability < 0.5 {
		t.Errorf("Expected high instability for unstable node, got %f", stability)
	}
}

func TestPredictWithDifferentProtocols(t *testing.T) {
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

	lp := NewLatencyPredictor(logger, prometheusServer.URL, relayServer.URL)
	err := lp.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize LatencyPredictor: %v", err)
	}

	// Собираем метрики
	err = lp.metricsCollector.CollectMetrics()
	if err != nil {
		t.Fatalf("Failed to collect metrics: %v", err)
	}

	// Тестируем разные протоколы
	protocols := []string{"QUIC", "P2P_QUIC", "TCP", "UDP"}
	latencies := make(map[string]float64)

	for _, protocol := range protocols {
		latency, err := lp.Predict("node1", "node2", protocol)
		if err != nil {
			t.Fatalf("Expected no error predicting latency for %s, got %v", protocol, err)
		}

		latencies[protocol] = latency

		// Латентность должна быть положительной
		if latency <= 0 {
			t.Errorf("Expected positive latency for %s, got %f", protocol, latency)
		}
	}

	// QUIC должен быть быстрее TCP
	if latencies["QUIC"] >= latencies["TCP"] {
		t.Errorf("Expected QUIC to be faster than TCP: QUIC=%f, TCP=%f", latencies["QUIC"], latencies["TCP"])
	}

	// P2P_QUIC должен быть быстрее TCP
	if latencies["P2P_QUIC"] >= latencies["TCP"] {
		t.Errorf("Expected P2P_QUIC to be faster than TCP: P2P_QUIC=%f, TCP=%f", latencies["P2P_QUIC"], latencies["TCP"])
	}
}

func TestLatencyPredictorUpdateWithRealMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Создаем mock серверы
	prometheusServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"status": "success",
			"data": {
				"resultType": "vector",
				"result": [
					{
						"value": [1640995200, "120"]
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
			"bandwidth": 1200.0,
			"latency": 22.0,
			"error_rate": 0.008
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer relayServer.Close()

	lp := NewLatencyPredictor(logger, prometheusServer.URL, relayServer.URL)
	err := lp.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize LatencyPredictor: %v", err)
	}

	// Получаем начальные значения моделей
	initialModels := make(map[string]*LatencyModel)
	for protocol, model := range lp.models {
		initialModels[protocol] = &LatencyModel{
			Protocol:    model.Protocol,
			BaseLatency: model.BaseLatency,
			Variation:   model.Variation,
			LastUpdate:  model.LastUpdate,
		}
	}

	// Выполняем обновление
	lp.Update()

	// Проверяем, что модели обновились
	for protocol, model := range lp.models {
		initialModel := initialModels[protocol]
		if model.LastUpdate.Before(initialModel.LastUpdate) || model.LastUpdate.Equal(initialModel.LastUpdate) {
			t.Errorf("Expected model for protocol %s to be updated", protocol)
		}
	}
}

func TestLatencyPredictorStartMetricsCollection(t *testing.T) {
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

	lp := NewLatencyPredictor(logger, prometheusServer.URL, relayServer.URL)
	err := lp.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize LatencyPredictor: %v", err)
	}

	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Запускаем сбор метрик
	done := make(chan struct{})
	go func() {
		lp.StartMetricsCollection(ctx)
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

func TestLatencyPredictorGetMetricsCollector(t *testing.T) {
	logger := zaptest.NewLogger(t)
	lp := NewLatencyPredictor(logger, "http://localhost:9090", "http://localhost:8080")

	collector := lp.GetMetricsCollector()
	if collector == nil {
		t.Fatal("Expected metrics collector to be returned")
	}

	// Проверяем, что это тот же коллектор
	if collector != lp.metricsCollector {
		t.Error("Expected same metrics collector instance")
	}
}

func TestPredictWithUnknownProtocol(t *testing.T) {
	logger := zaptest.NewLogger(t)
	lp := NewLatencyPredictor(logger, "http://localhost:9090", "http://localhost:8080")

	err := lp.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize LatencyPredictor: %v", err)
	}

	// Тестируем неизвестный протокол
	_, err = lp.Predict("node1", "node2", "UNKNOWN_PROTOCOL")
	if err == nil {
		t.Error("Expected error for unknown protocol")
	}

	expectedError := "unknown protocol: UNKNOWN_PROTOCOL"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestPredictWithHighLoadNetwork(t *testing.T) {
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

	lp := NewLatencyPredictor(logger, prometheusServer.URL, relayServer.URL)
	err := lp.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize LatencyPredictor: %v", err)
	}

	// Собираем метрики
	err = lp.metricsCollector.CollectMetrics()
	if err != nil {
		t.Fatalf("Failed to collect metrics: %v", err)
	}

	// Тестируем предсказание при высокой нагрузке
	latency, err := lp.Predict("node1", "node2", "TCP")
	if err != nil {
		t.Fatalf("Expected no error predicting latency, got %v", err)
	}

	// При высокой нагрузке латентность должна быть значительной
	if latency < 50 {
		t.Errorf("Expected higher latency with high load, got %f", latency)
	}
}
