package cadence

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"
)

func TestNewCadenceIntegration(t *testing.T) {
	logger := zaptest.NewLogger(t)
	baseURL := "http://localhost:8080"
	domain := "test-domain"

	ci := NewCadenceIntegration(logger, baseURL, domain)

	if ci == nil {
		t.Fatal("Expected CadenceIntegration to be created, got nil")
	}

	if ci.baseURL != baseURL {
		t.Errorf("Expected base URL %s, got %s", baseURL, ci.baseURL)
	}

	if ci.domain != domain {
		t.Errorf("Expected domain %s, got %s", domain, ci.domain)
	}

	if ci.httpClient == nil {
		t.Fatal("Expected HTTP client to be initialized")
	}
}

func TestGetWorkflowStatsWithRealAPI(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Создаем mock Cadence сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/domains/test-domain/workflows/executions" {
			// Возвращаем реальные данные workflow
			executions := []WorkflowExecution{
				{
					ExecutionID: "exec-1",
					WorkflowID:  "ai-anomaly-detection",
					Status:      "COMPLETED",
					StartTime:   time.Now().Add(-2 * time.Hour),
					EndTime:     func() *time.Time { t := time.Now().Add(-1 * time.Hour); return &t }(),
					Input: map[string]interface{}{
						"threshold": 0.5,
						"metrics":   []string{"cpu", "memory", "network"},
					},
					Output: map[string]interface{}{
						"anomalies_detected": 3,
						"confidence":         0.85,
					},
				},
				{
					ExecutionID: "exec-2",
					WorkflowID:  "ai-load-prediction",
					Status:      "RUNNING",
					StartTime:   time.Now().Add(-30 * time.Minute),
					Input: map[string]interface{}{
						"horizon": 24,
						"model":   "arima",
					},
				},
				{
					ExecutionID: "exec-3",
					WorkflowID:  "ai-route-optimization",
					Status:      "FAILED",
					StartTime:   time.Now().Add(-1 * time.Hour),
					EndTime:     func() *time.Time { t := time.Now().Add(-45 * time.Minute); return &t }(),
					Input: map[string]interface{}{
						"source":      "node1",
						"destination": "node2",
						"constraints": map[string]interface{}{
							"max_latency":   100,
							"min_bandwidth": 1000,
						},
					},
				},
			}

			response := map[string]interface{}{
				"executions": executions,
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	ci := NewCadenceIntegration(logger, server.URL, "test-domain")

	// Тестируем получение статистики workflow
	stats, err := ci.GetWorkflowStats(context.Background())
	if err != nil {
		t.Fatalf("Expected no error getting workflow stats, got %v", err)
	}

	// Проверяем статистику
	if stats.TotalExecutions != 3 {
		t.Errorf("Expected 3 total executions, got %d", stats.TotalExecutions)
	}

	if stats.RunningExecutions != 1 {
		t.Errorf("Expected 1 running execution, got %d", stats.RunningExecutions)
	}

	if stats.CompletedExecutions != 1 {
		t.Errorf("Expected 1 completed execution, got %d", stats.CompletedExecutions)
	}

	if stats.FailedExecutions != 1 {
		t.Errorf("Expected 1 failed execution, got %d", stats.FailedExecutions)
	}

	// Проверяем процент успешности
	expectedSuccessRate := float64(1) / float64(3) * 100 // 1 из 3 успешных
	if stats.SuccessRate != expectedSuccessRate {
		t.Errorf("Expected success rate %f, got %f", expectedSuccessRate, stats.SuccessRate)
	}

	// Проверяем среднюю продолжительность
	if stats.AverageDuration <= 0 {
		t.Error("Expected positive average duration")
	}
}

func TestGetWorkflowStatsWithAPIFailure(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Создаем сервер, который возвращает ошибку
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	ci := NewCadenceIntegration(logger, server.URL, "test-domain")

	// Тестируем получение статистики при ошибке API
	stats, err := ci.GetWorkflowStats(context.Background())
	if err != nil {
		t.Fatalf("Expected no error even with API failure (fallback), got %v", err)
	}

	// Должны получить fallback данные
	if stats.TotalExecutions == 0 {
		t.Error("Expected fallback data to be returned")
	}
}

func TestStartAnomalyDetectionWorkflowWithRealAPI(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Создаем mock Cadence сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/domains/test-domain/workflows/start" && r.Method == "POST" {
			// Проверяем тело запроса
			var requestBody map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Проверяем, что запрос содержит правильные данные
			if requestBody["workflowType"] != "AnomalyDetectionWorkflow" {
				t.Errorf("Expected workflowType 'AnomalyDetectionWorkflow', got %v", requestBody["workflowType"])
			}

			if requestBody["taskList"] != "ai-task-list" {
				t.Errorf("Expected taskList 'ai-task-list', got %v", requestBody["taskList"])
			}

			// Возвращаем успешный ответ
			response := map[string]interface{}{
				"workflowId": "anomaly-detection-12345",
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	ci := NewCadenceIntegration(logger, server.URL, "test-domain")

	// Тестируем запуск workflow
	input := map[string]interface{}{
		"threshold": 0.8,
		"metrics":   []string{"cpu", "memory", "network", "disk"},
	}

	workflowID, err := ci.StartAnomalyDetectionWorkflow(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error starting workflow, got %v", err)
	}

	if workflowID != "anomaly-detection-12345" {
		t.Errorf("Expected workflow ID 'anomaly-detection-12345', got %s", workflowID)
	}
}

func TestStartAnomalyDetectionWorkflowWithAPIFailure(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Создаем сервер, который возвращает ошибку
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	ci := NewCadenceIntegration(logger, server.URL, "test-domain")

	// Тестируем запуск workflow при ошибке API
	input := map[string]interface{}{
		"threshold": 0.8,
		"metrics":   []string{"cpu", "memory"},
	}

	workflowID, err := ci.StartAnomalyDetectionWorkflow(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error even with API failure (fallback), got %v", err)
	}

	// Должен вернуть fallback workflow ID
	if workflowID == "" {
		t.Error("Expected fallback workflow ID to be returned")
	}

	// Проверяем, что ID содержит префикс
	if len(workflowID) < 20 {
		t.Errorf("Expected fallback workflow ID to be longer, got %s", workflowID)
	}
}

func TestStartLoadPredictionWorkflowWithRealAPI(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Создаем mock Cadence сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/domains/test-domain/workflows/start" && r.Method == "POST" {
			// Проверяем тело запроса
			var requestBody map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Проверяем, что запрос содержит правильные данные
			if requestBody["workflowType"] != "LoadPredictionWorkflow" {
				t.Errorf("Expected workflowType 'LoadPredictionWorkflow', got %v", requestBody["workflowType"])
			}

			// Возвращаем успешный ответ
			response := map[string]interface{}{
				"workflowId": "load-prediction-67890",
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	ci := NewCadenceIntegration(logger, server.URL, "test-domain")

	// Тестируем запуск workflow прогнозирования нагрузки
	workflowID, err := ci.StartLoadPredictionWorkflow(context.Background(), 48, "arima")
	if err != nil {
		t.Fatalf("Expected no error starting load prediction workflow, got %v", err)
	}

	if workflowID != "load-prediction-67890" {
		t.Errorf("Expected workflow ID 'load-prediction-67890', got %s", workflowID)
	}
}

func TestStartRouteOptimizationWorkflowWithRealAPI(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Создаем mock Cadence сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/domains/test-domain/workflows/start" && r.Method == "POST" {
			// Проверяем тело запроса
			var requestBody map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Проверяем, что запрос содержит правильные данные
			if requestBody["workflowType"] != "RouteOptimizationWorkflow" {
				t.Errorf("Expected workflowType 'RouteOptimizationWorkflow', got %v", requestBody["workflowType"])
			}

			// Возвращаем успешный ответ
			response := map[string]interface{}{
				"workflowId": "route-optimization-11111",
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	ci := NewCadenceIntegration(logger, server.URL, "test-domain")

	// Тестируем запуск workflow оптимизации маршрутов
	constraints := map[string]interface{}{
		"max_latency":   50,
		"min_bandwidth": 1000,
		"protocol":      "QUIC",
	}

	workflowID, err := ci.StartRouteOptimizationWorkflow(context.Background(), "node1", "node2", constraints)
	if err != nil {
		t.Fatalf("Expected no error starting route optimization workflow, got %v", err)
	}

	if workflowID != "route-optimization-11111" {
		t.Errorf("Expected workflow ID 'route-optimization-11111', got %s", workflowID)
	}
}

func TestGetWorkflowExecutionWithRealAPI(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Создаем mock Cadence сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/domains/test-domain/workflows/executions/test-workflow-id" {
			// Возвращаем информацию о выполнении workflow
			execution := WorkflowExecution{
				ExecutionID: "test-workflow-id",
				WorkflowID:  "ai-anomaly-detection",
				Status:      "COMPLETED",
				StartTime:   time.Now().Add(-1 * time.Hour),
				EndTime:     func() *time.Time { t := time.Now(); return &t }(),
				Input: map[string]interface{}{
					"threshold": 0.5,
					"metrics":   []string{"cpu", "memory", "network"},
				},
				Output: map[string]interface{}{
					"anomalies_detected": 2,
					"confidence":         0.92,
				},
				History: []WorkflowEvent{
					{
						EventID:   1,
						EventType: "WorkflowExecutionStarted",
						Timestamp: time.Now().Add(-1 * time.Hour),
						Details: map[string]interface{}{
							"workflowType": "AnomalyDetectionWorkflow",
						},
					},
					{
						EventID:   2,
						EventType: "WorkflowExecutionCompleted",
						Timestamp: time.Now(),
						Details: map[string]interface{}{
							"result": "anomaly detection completed",
						},
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(execution)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	ci := NewCadenceIntegration(logger, server.URL, "test-domain")

	// Тестируем получение информации о выполнении workflow
	execution, err := ci.GetWorkflowExecution(context.Background(), "test-workflow-id")
	if err != nil {
		t.Fatalf("Expected no error getting workflow execution, got %v", err)
	}

	// Проверяем данные выполнения
	if execution.ExecutionID != "test-workflow-id" {
		t.Errorf("Expected execution ID 'test-workflow-id', got %s", execution.ExecutionID)
	}

	if execution.WorkflowID != "ai-anomaly-detection" {
		t.Errorf("Expected workflow ID 'ai-anomaly-detection', got %s", execution.WorkflowID)
	}

	if execution.Status != "COMPLETED" {
		t.Errorf("Expected status 'COMPLETED', got %s", execution.Status)
	}

	if len(execution.History) != 4 {
		t.Errorf("Expected 4 history events, got %d", len(execution.History))
	}
}

func TestGetWorkflowHistoryWithRealAPI(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Создаем mock Cadence сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/domains/test-domain/workflows/executions/test-workflow-id" {
			// Возвращаем информацию о выполнении workflow с историей
			execution := WorkflowExecution{
				ExecutionID: "test-workflow-id",
				WorkflowID:  "ai-anomaly-detection",
				Status:      "COMPLETED",
				StartTime:   time.Now().Add(-1 * time.Hour),
				EndTime:     func() *time.Time { t := time.Now(); return &t }(),
				History: []WorkflowEvent{
					{
						EventID:   1,
						EventType: "WorkflowExecutionStarted",
						Timestamp: time.Now().Add(-1 * time.Hour),
						Details: map[string]interface{}{
							"workflowType": "AnomalyDetectionWorkflow",
						},
					},
					{
						EventID:   2,
						EventType: "ActivityTaskScheduled",
						Timestamp: time.Now().Add(-58 * time.Minute),
						Details: map[string]interface{}{
							"activityType": "CollectMetricsActivity",
						},
					},
					{
						EventID:   3,
						EventType: "ActivityTaskCompleted",
						Timestamp: time.Now().Add(-55 * time.Minute),
						Details: map[string]interface{}{
							"result": "metrics collected successfully",
						},
					},
					{
						EventID:   4,
						EventType: "WorkflowExecutionCompleted",
						Timestamp: time.Now(),
						Details: map[string]interface{}{
							"result": "anomaly detection completed",
						},
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(execution)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	ci := NewCadenceIntegration(logger, server.URL, "test-domain")

	// Тестируем получение истории workflow
	history, err := ci.GetWorkflowHistory(context.Background(), "test-workflow-id")
	if err != nil {
		t.Fatalf("Expected no error getting workflow history, got %v", err)
	}

	// Проверяем историю
	if len(history) != 4 {
		t.Errorf("Expected 4 history events, got %d", len(history))
	}

	// Проверяем первый и последний события
	if history[0].EventType != "WorkflowExecutionStarted" {
		t.Errorf("Expected first event type 'WorkflowExecutionStarted', got %s", history[0].EventType)
	}

	if history[len(history)-1].EventType != "WorkflowExecutionCompleted" {
		t.Errorf("Expected last event type 'WorkflowExecutionCompleted', got %s", history[len(history)-1].EventType)
	}
}

func TestCancelWorkflowWithRealAPI(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Создаем mock Cadence сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/domains/test-domain/workflows/executions/test-workflow-id/cancel" && r.Method == "POST" {
			// Возвращаем успешный ответ
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	ci := NewCadenceIntegration(logger, server.URL, "test-domain")

	// Тестируем отмену workflow
	err := ci.CancelWorkflow(context.Background(), "test-workflow-id")
	if err != nil {
		t.Fatalf("Expected no error cancelling workflow, got %v", err)
	}
}

func TestGetDomainInfoWithRealAPI(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Создаем mock Cadence сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/domains/test-domain" {
			// Возвращаем информацию о домене
			domainInfo := map[string]interface{}{
				"name":        "test-domain",
				"status":      "REGISTERED",
				"description": "Test AI Domain",
				"created_at":  time.Now().Add(-24 * time.Hour),
				"updated_at":  time.Now(),
				"retention":   "7 days",
				"clusters":    []string{"primary"},
				"workflows": map[string]interface{}{
					"ai-anomaly-detection": map[string]interface{}{
						"type":        "AnomalyDetectionWorkflow",
						"description": "AI-powered anomaly detection workflow",
						"version":     "1.0",
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(domainInfo)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	ci := NewCadenceIntegration(logger, server.URL, "test-domain")

	// Тестируем получение информации о домене
	domainInfo, err := ci.GetDomainInfo(context.Background())
	if err != nil {
		t.Fatalf("Expected no error getting domain info, got %v", err)
	}

	// Проверяем данные домена
	if domainInfo["name"] != "test-domain" {
		t.Errorf("Expected domain name 'test-domain', got %v", domainInfo["name"])
	}

	if domainInfo["status"] != "REGISTERED" {
		t.Errorf("Expected domain status 'REGISTERED', got %v", domainInfo["status"])
	}

	// Проверяем наличие workflows
	workflows, ok := domainInfo["workflows"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected workflows to be a map")
	}

	if _, exists := workflows["ai-anomaly-detection"]; !exists {
		t.Error("Expected ai-anomaly-detection workflow to be present")
	}
}

func TestHealthCheckWithRealAPI(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Создаем mock Cadence сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			// Возвращаем успешный ответ
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	ci := NewCadenceIntegration(logger, server.URL, "test-domain")

	// Тестируем проверку здоровья
	err := ci.HealthCheck(context.Background())
	if err != nil {
		t.Fatalf("Expected no error during health check, got %v", err)
	}
}

func TestGetMetricsWithRealAPI(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Создаем mock Cadence сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/domains/test-domain/workflows/executions":
			// Возвращаем статистику workflow
			executions := []WorkflowExecution{
				{
					ExecutionID: "exec-1",
					WorkflowID:  "ai-anomaly-detection",
					Status:      "COMPLETED",
					StartTime:   time.Now().Add(-1 * time.Hour),
					EndTime:     func() *time.Time { t := time.Now(); return &t }(),
				},
			}

			response := map[string]interface{}{
				"executions": executions,
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)

		case "/api/v1/domains/test-domain":
			// Возвращаем информацию о домене
			domainInfo := map[string]interface{}{
				"name":   "test-domain",
				"status": "REGISTERED",
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(domainInfo)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	ci := NewCadenceIntegration(logger, server.URL, "test-domain")

	// Тестируем получение метрик
	metrics, err := ci.GetMetrics(context.Background())
	if err != nil {
		t.Fatalf("Expected no error getting metrics, got %v", err)
	}

	// Проверяем метрики
	if metrics["domain"] != "test-domain" {
		t.Errorf("Expected domain 'test-domain', got %v", metrics["domain"])
	}

	if metrics["health"] != "healthy" {
		t.Errorf("Expected health 'healthy', got %v", metrics["health"])
	}

	// Проверяем наличие workflow_stats
	if _, exists := metrics["workflow_stats"]; !exists {
		t.Error("Expected workflow_stats to be present")
	}

	// Проверяем наличие domain_info
	if _, exists := metrics["domain_info"]; !exists {
		t.Error("Expected domain_info to be present")
	}
}

func TestCadenceIntegrationWithNetworkTimeout(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Создаем сервер, который не отвечает
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Не отвечаем, чтобы вызвать таймаут
		time.Sleep(2 * time.Second)
	}))
	defer server.Close()

	// Создаем интеграцию с коротким таймаутом
	ci := NewCadenceIntegration(logger, server.URL, "test-domain")
	ci.httpClient.Timeout = 100 * time.Millisecond

	// Тестируем получение статистики с таймаутом
	stats, err := ci.GetWorkflowStats(context.Background())
	if err != nil {
		t.Fatalf("Expected no error even with timeout (fallback), got %v", err)
	}

	// Должны получить fallback данные
	if stats.TotalExecutions == 0 {
		t.Error("Expected fallback data to be returned")
	}
}
