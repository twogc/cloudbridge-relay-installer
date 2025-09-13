package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestMonitoringEndpoints_GetSystemMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	endpoints := NewMonitoringEndpoints(logger)
	router := gin.New()
	router.GET("/monitoring/system", endpoints.GetSystemMetrics)

	req := httptest.NewRequest("GET", "/monitoring/system", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])

	metrics := response["metrics"].(map[string]interface{})
	assert.Contains(t, metrics, "cpu")
	assert.Contains(t, metrics, "memory")
	assert.Contains(t, metrics, "disk")
	assert.Contains(t, metrics, "network")
	assert.Contains(t, metrics, "runtime")
}

func TestMonitoringEndpoints_GetHealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	endpoints := NewMonitoringEndpoints(logger)
	router := gin.New()
	router.GET("/monitoring/health", endpoints.GetHealthCheck)

	req := httptest.NewRequest("GET", "/monitoring/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, []string{"healthy", "warning", "unhealthy"}, response["status"])
	assert.Equal(t, "cloudbridge-relay", response["service"])
	assert.Contains(t, response, "components")
	assert.Contains(t, response, "system")
}

func TestMonitoringEndpoints_GetPrometheusMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	endpoints := NewMonitoringEndpoints(logger)
	router := gin.New()
	router.GET("/monitoring/prometheus", endpoints.GetPrometheusMetrics)

	req := httptest.NewRequest("GET", "/monitoring/prometheus", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/plain; version=0.0.4; charset=utf-8", w.Header().Get("Content-Type"))

	// Проверяем, что ответ содержит метрики в формате Prometheus
	body := w.Body.String()
	assert.Contains(t, body, "cloudbridge_memory_usage_percent")
	assert.Contains(t, body, "cloudbridge_disk_usage_percent")
	assert.Contains(t, body, "cloudbridge_goroutines_total")
	assert.Contains(t, body, "cloudbridge_heap_alloc_bytes")
}

func TestMonitoringEndpoints_GetServiceStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	endpoints := NewMonitoringEndpoints(logger)
	router := gin.New()
	router.GET("/monitoring/services", endpoints.GetServiceStatus)

	req := httptest.NewRequest("GET", "/monitoring/services", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])

	services := response["services"].(map[string]interface{})
	assert.Contains(t, services, "api_server")
	assert.Contains(t, services, "relay_server")
	assert.Contains(t, services, "ai_service")
	assert.Contains(t, services, "data_pipeline")

	// Проверяем структуру каждого сервиса
	apiServer := services["api_server"].(map[string]interface{})
	assert.Equal(t, "healthy", apiServer["status"])
	assert.Contains(t, apiServer, "version")
	assert.Contains(t, apiServer, "endpoints")
}
