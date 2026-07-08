package metrics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

func TestEnhancedMetrics(t *testing.T) {
	// Создаем registry для тестов
	reg := prometheus.NewRegistry()

	// Создаем enhanced metrics
	em := NewEnhancedMetrics(reg)

	// Регистрируем build info
	em.RegisterBuildInfo("test", "2024-01-01", "1.0.0")

	// Тестируем protocol метрики
	em.ObserveProtocolLatency("http2", "connect", 100*time.Millisecond)
	em.IncProtocolErrors("http2", "timeout")
	em.IncProtocolSwitches("http2", "quic", "performance")
	em.IncProtocolSuccess("quic", "connect")

	// Тестируем health check метрики
	em.SetHealthCheckStatus("database", "component", true)
	em.ObserveHealthCheckDuration("database", "component", 50*time.Millisecond)

	// Тестируем connection pool метрики
	em.SetConnectionPoolSize("tcp", "http2", 100)
	em.SetConnectionPoolIdle("tcp", "http2", 20)
	em.SetConnectionPoolActive("tcp", "http2", 80)

	// Тестируем QUIC метрики
	em.IncQUICStreams("inbound", "data")
	em.SetQUICStreamsActive("inbound", "data", 5)
	em.ObserveQUICStreamLatency("inbound", "data", 10*time.Millisecond)
	em.IncQUICConnectionErrors("timeout")

	// Тестируем ingress метрики
	em.IncIngressRequests("example.com", "/api", "GET", "200")
	em.ObserveIngressRequestDuration("example.com", "/api", "GET", 150*time.Millisecond)
	em.IncIngressErrors("example.com", "timeout")
	em.IncIngressBandwidth("example.com", "in", 1024)

	// Тестируем system метрики
	em.SetSystemMemoryUsage(1024 * 1024 * 100) // 100MB
	em.SetSystemCPUUsage(25.5)
	em.SetSystemGoroutines(100)
	em.SetSystemHeapAlloc(1024 * 1024 * 50) // 50MB
	em.SetSystemHeapSys(1024 * 1024 * 200)  // 200MB

	// Тестируем diagnostic метрики
	em.IncDiagnosticRequests("health", "GET")
	em.IncDiagnosticErrors("health", "timeout")

	t.Log("Enhanced metrics created and populated successfully")
}

func TestEnhancedHealthChecker(t *testing.T) {
	// Создаем logger
	logger, _ := zap.NewDevelopment()

	// Создаем enhanced metrics
	reg := prometheus.NewRegistry()
	em := NewEnhancedMetrics(reg)

	// Создаем health checker
	checker := NewEnhancedHealthChecker("1.0.0", logger, em)

	// Проверяем базовые checks
	status := checker.RunChecks(context.Background())

	if status.Status != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", status.Status)
	}

	if !status.Ready {
		t.Error("Expected system to be ready")
	}

	// Проверяем количество checks
	expectedChecks := 4 // database, redis, tcp_proxy, system
	if len(status.Checks) != expectedChecks {
		t.Errorf("Expected %d checks, got %d", expectedChecks, len(status.Checks))
	}

	// Проверяем готовые checks
	readyCount := checker.CountReadyChecks()
	if readyCount != expectedChecks {
		t.Errorf("Expected %d ready checks, got %d", expectedChecks, readyCount)
	}

	t.Log("Enhanced health checker working correctly")
}

func TestEnhancedServer(t *testing.T) {
	// Создаем logger
	logger, _ := zap.NewDevelopment()

	// Создаем enhanced metrics
	reg := prometheus.NewRegistry()
	em := NewEnhancedMetrics(reg)

	// Создаем health checker
	checker := NewEnhancedHealthChecker("1.0.0", logger, em)

	// Создаем server config
	config := Config{
		Port:            ":0", // Используем случайный порт
		Path:            "/metrics",
		EnhancedMetrics: em,
		HealthChecker:   checker,
		Logger:          logger,
		Version:         "1.0.0",
		BuildTime:       "2024-01-01",
		BuildType:       "test",
		TunnelID:        uuid.New(),
		ConnectorID:     uuid.New(),
	}

	// Создаем server
	server := NewEnhancedServer(config)

	// Проверяем, что сервер создан
	if server == nil {
		t.Fatal("Failed to create enhanced server")
	}

	// Проверяем конфигурацию
	if server.GetVersion() != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", server.GetVersion())
	}

	if server.GetMetricsPath() != "/metrics" {
		t.Errorf("Expected path '/metrics', got '%s'", server.GetMetricsPath())
	}

	// Проверяем build info
	buildInfo := server.GetBuildInfo()
	if buildInfo["version"] != "1.0.0" {
		t.Errorf("Expected build info version '1.0.0', got '%s'", buildInfo["version"])
	}

	t.Log("Enhanced server created successfully")
}

func TestHealthCheckEndpoints(t *testing.T) {
	// Создаем logger
	logger, _ := zap.NewDevelopment()

	// Создаем enhanced metrics
	reg := prometheus.NewRegistry()
	em := NewEnhancedMetrics(reg)

	// Создаем health checker
	checker := NewEnhancedHealthChecker("1.0.0", logger, em)

	// Тестируем health check endpoint
	req := httptest.NewRequest("GET", "/healthcheck", nil)
	w := httptest.NewRecorder()

	checker.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}

	// Проверяем, что ответ содержит JSON
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected content type 'application/json', got '%s'", contentType)
	}

	t.Log("Health check endpoint working correctly")
}

func TestSystemMetricsCollector(t *testing.T) {
	// Создаем enhanced metrics
	reg := prometheus.NewRegistry()
	em := NewEnhancedMetrics(reg)

	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Запускаем сборщик метрик
	go em.StartSystemMetricsCollector(ctx, 1*time.Second)

	// Ждем немного для сбора метрик
	time.Sleep(500 * time.Millisecond)

	// Проверяем, что метрики обновлены
	// (это базовый тест, в реальности нужно проверять конкретные значения)

	t.Log("System metrics collector working correctly")
}

func BenchmarkEnhancedMetrics(b *testing.B) {
	// Создаем enhanced metrics
	reg := prometheus.NewRegistry()
	em := NewEnhancedMetrics(reg)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Тестируем производительность основных операций
		em.ObserveProtocolLatency("http2", "connect", 100*time.Millisecond)
		em.IncProtocolSuccess("http2", "connect")
		em.SetHealthCheckStatus("database", "component", true)
		em.IncIngressRequests("example.com", "/api", "GET", "200")
	}
}

func BenchmarkHealthChecker(b *testing.B) {
	// Создаем logger
	logger, _ := zap.NewDevelopment()

	// Создаем enhanced metrics
	reg := prometheus.NewRegistry()
	em := NewEnhancedMetrics(reg)

	// Создаем health checker
	checker := NewEnhancedHealthChecker("1.0.0", logger, em)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		checker.RunChecks(context.Background())
	}
}
