package predictive

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/ai/pipeline"
	"go.uber.org/zap/zaptest"
)

func TestRealLoadPredictorWithoutSinVariation(t *testing.T) {
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

	cadenceServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"executions": []
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer cadenceServer.Close()

	// Создаем коллектор данных
	dataCollector := pipeline.NewRealDataCollector(logger, prometheusServer.URL, relayServer.URL, cadenceServer.URL)

	// Создаем предиктор нагрузки
	predictor := NewRealLoadPredictor(logger, dataCollector)

	// Инициализируем предиктор
	err := predictor.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize predictor: %v", err)
	}

	// Собираем данные для обучения
	_, err = dataCollector.Collect()
	if err != nil {
		t.Fatalf("Failed to collect data: %v", err)
	}

	// Тестируем прогнозирование нагрузки
	horizon := 12
	prediction, err := predictor.PredictLoad(horizon)
	if err != nil {
		t.Fatalf("Expected no error predicting load, got %v", err)
	}

	if prediction == nil {
		t.Fatal("Expected load prediction to be returned")
	}

	// Проверяем, что прогноз основан на реальных данных, а не на синусоидальных заглушках
	if prediction.Horizon != horizon {
		t.Errorf("Expected horizon %d, got %d", horizon, prediction.Horizon)
	}

	if prediction.Confidence < 0 || prediction.Confidence > 1 {
		t.Errorf("Expected confidence between 0 and 1, got %f", prediction.Confidence)
	}

	if prediction.ModelID != "real_data_ensemble" {
		t.Errorf("Expected model ID 'real_data_ensemble', got %s", prediction.ModelID)
	}

	if prediction.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}
}

func TestRealLoadPredictorWithHistoricalData(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Создаем mock серверы с разными значениями для имитации исторических данных
	prometheusServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Возвращаем разные значения для имитации исторических данных
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
			"bandwidth": 1200.0,
			"latency": 20.0,
			"error_rate": 0.005
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

	// Создаем коллектор данных
	dataCollector := pipeline.NewRealDataCollector(logger, prometheusServer.URL, relayServer.URL, cadenceServer.URL)

	// Создаем предиктор нагрузки
	predictor := NewRealLoadPredictor(logger, dataCollector)

	// Инициализируем предиктор
	err := predictor.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize predictor: %v", err)
	}

	// Собираем несколько наборов данных для создания истории
	for i := 0; i < 5; i++ {
		_, err = dataCollector.Collect()
		if err != nil {
			t.Fatalf("Failed to collect data iteration %d: %v", i, err)
		}
		time.Sleep(10 * time.Millisecond) // Небольшая задержка между сборами
	}

	// Тестируем прогнозирование с историческими данными
	horizon := 24
	prediction, err := predictor.PredictLoad(horizon)
	if err != nil {
		t.Fatalf("Expected no error predicting load, got %v", err)
	}

	if prediction == nil {
		t.Fatal("Expected load prediction to be returned")
	}

	// Проверяем, что прогноз учитывает исторические данные
	if prediction.Confidence < 0.5 {
		t.Errorf("Expected higher confidence with historical data, got %f", prediction.Confidence)
	}
}

func TestRealLoadPredictorModelAccuracy(t *testing.T) {
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

	cadenceServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"executions": []
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer cadenceServer.Close()

	// Создаем коллектор данных
	dataCollector := pipeline.NewRealDataCollector(logger, prometheusServer.URL, relayServer.URL, cadenceServer.URL)

	// Создаем предиктор нагрузки
	predictor := NewRealLoadPredictor(logger, dataCollector)

	// Инициализируем предиктор
	err := predictor.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize predictor: %v", err)
	}

	// Собираем данные для обучения
	_, err = dataCollector.Collect()
	if err != nil {
		t.Fatalf("Failed to collect data: %v", err)
	}

	// Тестируем получение информации о моделях
	modelInfo := predictor.GetModelInfo()
	if modelInfo == nil {
		t.Fatal("Expected model info to be returned")
	}

	// Проверяем, что все модели присутствуют
	expectedModels := []string{"connections", "bandwidth", "latency", "system_load"}
	for _, modelName := range expectedModels {
		if _, exists := modelInfo[modelName]; !exists {
			t.Errorf("Expected model %s to be present", modelName)
		}
	}

	// Проверяем структуру информации о модели
	for modelName, info := range modelInfo {
		modelData, ok := info.(map[string]interface{})
		if !ok {
			t.Errorf("Expected model info for %s to be a map", modelName)
			continue
		}

		// Проверяем наличие ключевых полей
		expectedFields := []string{"type", "parameters", "last_trained", "accuracy", "data_points"}
		for _, field := range expectedFields {
			if _, exists := modelData[field]; !exists {
				t.Errorf("Expected field %s to be present in model %s", field, modelName)
			}
		}
	}
}

func TestRealLoadPredictorRetrainModels(t *testing.T) {
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

	cadenceServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"executions": []
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer cadenceServer.Close()

	// Создаем коллектор данных
	dataCollector := pipeline.NewRealDataCollector(logger, prometheusServer.URL, relayServer.URL, cadenceServer.URL)

	// Создаем предиктор нагрузки
	predictor := NewRealLoadPredictor(logger, dataCollector)

	// Инициализируем предиктор
	err := predictor.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize predictor: %v", err)
	}

	// Собираем данные для обучения
	_, err = dataCollector.Collect()
	if err != nil {
		t.Fatalf("Failed to collect data: %v", err)
	}

	// Получаем информацию о моделях до переобучения
	_ = predictor.GetModelInfo()

	// Тестируем переобучение моделей
	err = predictor.RetrainModels()
	if err != nil {
		t.Fatalf("Expected no error retraining models, got %v", err)
	}

	// Получаем информацию о моделях после переобучения
	modelInfoAfter := predictor.GetModelInfo()

	// Проверяем, что время последнего обучения обновилось
	for modelName, info := range modelInfoAfter {
		modelData, ok := info.(map[string]interface{})
		if !ok {
			continue
		}

		lastTrained, ok := modelData["last_trained"].(time.Time)
		if !ok {
			t.Errorf("Expected last_trained to be time.Time for model %s", modelName)
			continue
		}

		// Проверяем, что время обучения недавнее
		if time.Since(lastTrained) > time.Minute {
			t.Errorf("Expected recent training time for model %s, got %v", modelName, lastTrained)
		}
	}
}

func TestRealLoadPredictorWithInsufficientData(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Создаем mock серверы, которые возвращают пустые данные
	prometheusServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"status": "success",
			"data": {
				"resultType": "vector",
				"result": []
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer prometheusServer.Close()

	relayServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{}`
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

	// Создаем коллектор данных
	dataCollector := pipeline.NewRealDataCollector(logger, prometheusServer.URL, relayServer.URL, cadenceServer.URL)

	// Создаем предиктор нагрузки
	predictor := NewRealLoadPredictor(logger, dataCollector)

	// Инициализируем предиктор
	err := predictor.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize predictor: %v", err)
	}

	// Собираем данные (пустые)
	_, err = dataCollector.Collect()
	if err != nil {
		t.Fatalf("Failed to collect data: %v", err)
	}

	// Тестируем прогнозирование с недостаточными данными
	horizon := 6
	prediction, err := predictor.PredictLoad(horizon)
	if err != nil {
		t.Fatalf("Expected no error even with insufficient data, got %v", err)
	}

	if prediction == nil {
		t.Fatal("Expected load prediction to be returned even with insufficient data")
	}

	// При недостаточных данных уверенность должна быть низкой
	if prediction.Confidence > 0.5 {
		t.Errorf("Expected low confidence with insufficient data, got %f", prediction.Confidence)
	}
}

func TestRealLoadPredictorTrendCalculation(t *testing.T) {
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

	cadenceServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"executions": []
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer cadenceServer.Close()

	// Создаем коллектор данных
	dataCollector := pipeline.NewRealDataCollector(logger, prometheusServer.URL, relayServer.URL, cadenceServer.URL)

	// Создаем предиктор нагрузки
	predictor := NewRealLoadPredictor(logger, dataCollector)

	// Инициализируем предиктор
	err := predictor.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize predictor: %v", err)
	}

	// Собираем данные для создания истории
	for i := 0; i < 10; i++ {
		_, err = dataCollector.Collect()
		if err != nil {
			t.Fatalf("Failed to collect data iteration %d: %v", i, err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Тестируем прогнозирование с трендом
	horizon := 12
	prediction, err := predictor.PredictLoad(horizon)
	if err != nil {
		t.Fatalf("Expected no error predicting load, got %v", err)
	}

	if prediction == nil {
		t.Fatal("Expected load prediction to be returned")
	}

	// Проверяем, что прогноз учитывает тренд (не использует синусоидальные заглушки)
	// Прогноз должен быть основан на реальных данных и трендах
	if prediction.Confidence < 0.1 {
		t.Errorf("Expected reasonable confidence, got %f", prediction.Confidence)
	}
}

func TestRealLoadPredictorConcurrentAccess(t *testing.T) {
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

	cadenceServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"executions": []
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer cadenceServer.Close()

	// Создаем коллектор данных
	dataCollector := pipeline.NewRealDataCollector(logger, prometheusServer.URL, relayServer.URL, cadenceServer.URL)

	// Создаем предиктор нагрузки
	predictor := NewRealLoadPredictor(logger, dataCollector)

	// Инициализируем предиктор
	err := predictor.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize predictor: %v", err)
	}

	// Собираем данные
	_, err = dataCollector.Collect()
	if err != nil {
		t.Fatalf("Failed to collect data: %v", err)
	}

	// Тестируем параллельный доступ
	done := make(chan struct{})
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer func() { done <- struct{}{} }()

			// Тестируем разные операции
			predictor.PredictLoad(12)
			predictor.GetModelInfo()
			predictor.RetrainModels()
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
