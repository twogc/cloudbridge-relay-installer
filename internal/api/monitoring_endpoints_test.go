package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewMonitoringEndpoints(t *testing.T) {
	logger := zap.NewNop()

	endpoints := NewMonitoringEndpoints(logger)

	assert.NotNil(t, endpoints)
}

func TestMonitoringEndpoints_GetSystemMetrics(t *testing.T) {
	logger := zap.NewNop()
	endpoints := NewMonitoringEndpoints(logger)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/monitoring/system", nil)

	endpoints.GetSystemMetrics(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMonitoringEndpoints_GetHealthCheck(t *testing.T) {
	logger := zap.NewNop()
	endpoints := NewMonitoringEndpoints(logger)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/monitoring/health", nil)

	endpoints.GetHealthCheck(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMonitoringEndpoints_GetPrometheusMetrics(t *testing.T) {
	logger := zap.NewNop()
	endpoints := NewMonitoringEndpoints(logger)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/monitoring/prometheus", nil)

	endpoints.GetPrometheusMetrics(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMonitoringEndpoints_GetServiceStatus(t *testing.T) {
	logger := zap.NewNop()
	endpoints := NewMonitoringEndpoints(logger)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/monitoring/services", nil)

	endpoints.GetServiceStatus(c)

	assert.Equal(t, http.StatusOK, w.Code)
}
