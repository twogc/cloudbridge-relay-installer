package relay

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai"
	"go.uber.org/zap/zaptest"
)

func TestNewRelayAIIntegration(t *testing.T) {
	logger := zaptest.NewLogger(t)
	aiService := &MockAIService{}
	config := GetDefaultAIIntegrationConfig()

	integration := NewRelayAIIntegration(aiService, config, logger)

	assert.NotNil(t, integration)
	assert.Equal(t, aiService, integration.aiService)
	assert.Equal(t, logger, integration.logger)
	assert.Equal(t, config.Enabled, integration.enabled)
	assert.Equal(t, config, integration.config)
}

func TestRelayAIIntegration_Initialize_Disabled(t *testing.T) {
	logger := zaptest.NewLogger(t)
	aiService := &MockAIService{}
	config := GetDefaultAIIntegrationConfig()
	config.Enabled = false

	integration := NewRelayAIIntegration(aiService, config, logger)

	err := integration.Initialize()

	assert.NoError(t, err)
	assert.False(t, integration.enabled)
	assert.Nil(t, integration.aiMiddleware)
}

func TestRelayAIIntegration_Initialize_NoAIService(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultAIIntegrationConfig()

	integration := NewRelayAIIntegration(nil, config, logger)

	err := integration.Initialize()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "AI service is not initialized")
}

func TestRelayAIIntegration_Initialize_Success(t *testing.T) {
	logger := zaptest.NewLogger(t)
	aiService := &MockAIService{}
	config := GetDefaultAIIntegrationConfig()

	// Настраиваем мок
	aiService.On("Initialize").Return(nil)
	status := &ai.ServiceStatus{
		Components: map[string]bool{
			"anomaly_detector": true,
			"load_predictor":   true,
			"route_optimizer":  true,
		},
	}
	aiService.On("GetStatus").Return(status)

	integration := NewRelayAIIntegration(aiService, config, logger)

	err := integration.Initialize()

	assert.NoError(t, err)
	assert.True(t, integration.enabled)
	assert.NotNil(t, integration.aiMiddleware)

	aiService.AssertExpectations(t)
}

func TestRelayAIIntegration_Initialize_AIServiceError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	aiService := &MockAIService{}
	config := GetDefaultAIIntegrationConfig()

	// Настраиваем мок для ошибки
	aiService.On("Initialize").Return(assert.AnError)

	integration := NewRelayAIIntegration(aiService, config, logger)

	err := integration.Initialize()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize AI middleware")

	aiService.AssertExpectations(t)
}

func TestRelayAIIntegration_ProcessConnection_Disabled(t *testing.T) {
	logger := zaptest.NewLogger(t)
	aiService := &MockAIService{}
	config := GetDefaultAIIntegrationConfig()
	config.Enabled = false

	integration := NewRelayAIIntegration(aiService, config, logger)
	integration.Initialize()

	conn := &MockConn{}

	err := integration.ProcessConnection(context.Background(), conn)

	assert.NoError(t, err)
	// AI сервис не должен вызываться, когда отключен
	aiService.AssertNotCalled(t, "Initialize")
}

func TestRelayAIIntegration_ProcessConnection_NotInitialized(t *testing.T) {
	logger := zaptest.NewLogger(t)
	aiService := &MockAIService{}
	config := GetDefaultAIIntegrationConfig()

	integration := NewRelayAIIntegration(aiService, config, logger)
	// Не вызываем Initialize()

	conn := &MockConn{}

	err := integration.ProcessConnection(context.Background(), conn)

	assert.NoError(t, err)
	// AI сервис не должен вызываться, когда не инициализирован
	aiService.AssertNotCalled(t, "Initialize")
}

func TestRelayAIIntegration_ProcessConnection_Success(t *testing.T) {
	logger := zaptest.NewLogger(t)
	aiService := &MockAIService{}
	config := GetDefaultAIIntegrationConfig()

	// Настраиваем мок
	aiService.On("Initialize").Return(nil)
	status := &ai.ServiceStatus{
		Components: map[string]bool{
			"anomaly_detector": true,
			"load_predictor":   true,
			"route_optimizer":  true,
		},
	}
	aiService.On("GetStatus").Return(status)

	integration := NewRelayAIIntegration(aiService, config, logger)
	integration.Initialize()

	conn := &MockConn{}
	addr := &MockAddr{network: "tcp", address: "192.168.1.100:12345"}
	conn.On("RemoteAddr").Return(addr)

	err := integration.ProcessConnection(context.Background(), conn)

	assert.NoError(t, err)

	aiService.AssertExpectations(t)
}

func TestRelayAIIntegration_SetEnabled(t *testing.T) {
	logger := zaptest.NewLogger(t)
	aiService := &MockAIService{}
	config := GetDefaultAIIntegrationConfig()

	integration := NewRelayAIIntegration(aiService, config, logger)

	assert.True(t, integration.IsEnabled())

	integration.SetEnabled(false)
	assert.False(t, integration.IsEnabled())

	integration.SetEnabled(true)
	assert.True(t, integration.IsEnabled())
}

func TestRelayAIIntegration_GetMetrics_NotInitialized(t *testing.T) {
	logger := zaptest.NewLogger(t)
	aiService := &MockAIService{}
	config := GetDefaultAIIntegrationConfig()

	integration := NewRelayAIIntegration(aiService, config, logger)

	metrics := integration.GetMetrics()

	assert.NotNil(t, metrics)
	assert.True(t, metrics["enabled"].(bool))
	assert.Equal(t, "not_initialized", metrics["status"])
}

func TestRelayAIIntegration_GetMetrics_Initialized(t *testing.T) {
	logger := zaptest.NewLogger(t)
	aiService := &MockAIService{}
	config := GetDefaultAIIntegrationConfig()

	// Настраиваем мок
	aiService.On("Initialize").Return(nil)
	status := &ai.ServiceStatus{
		Components: map[string]bool{
			"anomaly_detector": true,
			"load_predictor":   true,
			"route_optimizer":  true,
		},
	}
	aiService.On("GetStatus").Return(status)

	integration := NewRelayAIIntegration(aiService, config, logger)
	integration.Initialize()

	metrics := integration.GetMetrics()

	assert.NotNil(t, metrics)
	assert.True(t, metrics["integration_enabled"].(bool))
	assert.NotNil(t, metrics["config"])

	aiService.AssertExpectations(t)
}

func TestRelayAIIntegration_GetHealthStatus(t *testing.T) {
	logger := zaptest.NewLogger(t)
	aiService := &MockAIService{}
	config := GetDefaultAIIntegrationConfig()

	integration := NewRelayAIIntegration(aiService, config, logger)

	status := integration.GetHealthStatus()

	assert.NotNil(t, status)
	assert.True(t, status["enabled"].(bool))
	assert.False(t, status["initialized"].(bool))
	assert.NotNil(t, status["timestamp"])
}

func TestRelayAIIntegration_Shutdown_Disabled(t *testing.T) {
	logger := zaptest.NewLogger(t)
	aiService := &MockAIService{}
	config := GetDefaultAIIntegrationConfig()
	config.Enabled = false

	integration := NewRelayAIIntegration(aiService, config, logger)

	err := integration.Shutdown(context.Background())

	assert.NoError(t, err)
	// AI сервис не должен вызываться, когда отключен
	aiService.AssertNotCalled(t, "Stop")
}

func TestRelayAIIntegration_Shutdown_NoAIService(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := GetDefaultAIIntegrationConfig()

	integration := NewRelayAIIntegration(nil, config, logger)

	err := integration.Shutdown(context.Background())

	assert.NoError(t, err)
}

func TestRelayAIIntegration_Shutdown_Success(t *testing.T) {
	logger := zaptest.NewLogger(t)
	aiService := &MockAIService{}
	config := GetDefaultAIIntegrationConfig()

	// Настраиваем мок
	aiService.On("Stop").Return(nil)

	integration := NewRelayAIIntegration(aiService, config, logger)

	err := integration.Shutdown(context.Background())

	assert.NoError(t, err)

	aiService.AssertExpectations(t)
}

func TestRelayAIIntegration_Shutdown_Error(t *testing.T) {
	logger := zaptest.NewLogger(t)
	aiService := &MockAIService{}
	config := GetDefaultAIIntegrationConfig()

	// Настраиваем мок для ошибки
	aiService.On("Stop").Return(assert.AnError)

	integration := NewRelayAIIntegration(aiService, config, logger)

	err := integration.Shutdown(context.Background())

	assert.Error(t, err)

	aiService.AssertExpectations(t)
}

func TestGetDefaultAIIntegrationConfig(t *testing.T) {
	config := GetDefaultAIIntegrationConfig()

	assert.NotNil(t, config)
	assert.True(t, config.Enabled)
	assert.Equal(t, 5*time.Second, config.ProcessingInterval)
	assert.Equal(t, 5, config.CircuitBreakerConfig.FailureThreshold)
	assert.Equal(t, 3, config.CircuitBreakerConfig.SuccessThreshold)
	assert.Equal(t, 30*time.Second, config.CircuitBreakerConfig.Timeout)
	assert.True(t, config.MetricsConfig.Enabled)
	assert.Equal(t, 30*time.Second, config.MetricsConfig.ExportInterval)
}

func TestAIIntegrationConfig_Validate_Success(t *testing.T) {
	config := GetDefaultAIIntegrationConfig()

	err := config.Validate()

	assert.NoError(t, err)
}

func TestAIIntegrationConfig_Validate_InvalidProcessingInterval(t *testing.T) {
	config := GetDefaultAIIntegrationConfig()
	config.ProcessingInterval = 0

	err := config.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "processing_interval must be positive")
}

func TestAIIntegrationConfig_Validate_InvalidFailureThreshold(t *testing.T) {
	config := GetDefaultAIIntegrationConfig()
	config.CircuitBreakerConfig.FailureThreshold = 0

	err := config.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failure_threshold must be positive")
}

func TestAIIntegrationConfig_Validate_InvalidSuccessThreshold(t *testing.T) {
	config := GetDefaultAIIntegrationConfig()
	config.CircuitBreakerConfig.SuccessThreshold = 0

	err := config.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "success_threshold must be positive")
}

func TestAIIntegrationConfig_Validate_InvalidTimeout(t *testing.T) {
	config := GetDefaultAIIntegrationConfig()
	config.CircuitBreakerConfig.Timeout = 0

	err := config.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout must be positive")
}

func TestAIIntegrationConfig_Validate_InvalidMetricsInterval(t *testing.T) {
	config := GetDefaultAIIntegrationConfig()
	config.MetricsConfig.Enabled = true
	config.MetricsConfig.ExportInterval = 0

	err := config.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "export_interval must be positive")
}

func TestAIIntegrationConfig_Validate_MetricsDisabled(t *testing.T) {
	config := GetDefaultAIIntegrationConfig()
	config.MetricsConfig.Enabled = false
	config.MetricsConfig.ExportInterval = 0 // Это должно быть валидно, когда метрики отключены

	err := config.Validate()

	assert.NoError(t, err)
}
