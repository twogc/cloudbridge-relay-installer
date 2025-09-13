package relay

import (
	"context"
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"
)

// RelayAIIntegration обеспечивает интеграцию AI с основным Relay Server
type RelayAIIntegration struct {
	aiMiddleware *AIMiddleware
	aiService    AIServiceInterface
	logger       *zap.Logger
	enabled      bool
	config       *AIIntegrationConfig
}

// AIIntegrationConfig содержит конфигурацию интеграции AI
type AIIntegrationConfig struct {
	Enabled              bool                 `json:"enabled"`
	ProcessingInterval   time.Duration        `json:"processing_interval"`
	CircuitBreakerConfig CircuitBreakerConfig `json:"circuit_breaker"`
	MetricsConfig        MetricsConfig        `json:"metrics"`
}

// CircuitBreakerConfig содержит конфигурацию circuit breaker
type CircuitBreakerConfig struct {
	FailureThreshold int           `json:"failure_threshold"`
	SuccessThreshold int           `json:"success_threshold"`
	Timeout          time.Duration `json:"timeout"`
}

// MetricsConfig содержит конфигурацию метрик
type MetricsConfig struct {
	Enabled        bool          `json:"enabled"`
	ExportInterval time.Duration `json:"export_interval"`
}

// NewRelayAIIntegration создает новую интеграцию AI с Relay Server
func NewRelayAIIntegration(aiService AIServiceInterface, config *AIIntegrationConfig, logger *zap.Logger) *RelayAIIntegration {
	return &RelayAIIntegration{
		aiService: aiService,
		logger:    logger,
		enabled:   config.Enabled,
		config:    config,
	}
}

// Initialize инициализирует интеграцию AI
func (r *RelayAIIntegration) Initialize() error {
	if !r.enabled {
		r.logger.Info("AI integration is disabled")
		return nil
	}

	if r.aiService == nil {
		return fmt.Errorf("AI service is not initialized")
	}

	// Создаем AI middleware
	r.aiMiddleware = NewAIMiddleware(r.aiService, r.logger)

	// Инициализируем middleware
	if err := r.aiMiddleware.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize AI middleware: %w", err)
	}

	r.logger.Info("AI integration initialized successfully",
		zap.Bool("enabled", r.enabled),
		zap.Duration("processing_interval", r.config.ProcessingInterval),
	)

	return nil
}

// ProcessConnection обрабатывает соединение через AI
func (r *RelayAIIntegration) ProcessConnection(ctx context.Context, conn net.Conn) error {
	if !r.enabled || r.aiMiddleware == nil {
		return nil // AI отключен или не инициализирован
	}

	// Обрабатываем соединение через AI middleware
	return r.aiMiddleware.ProcessConnection(ctx, conn)
}

// SetEnabled включает/выключает AI интеграцию
func (r *RelayAIIntegration) SetEnabled(enabled bool) {
	r.enabled = enabled
	if r.aiMiddleware != nil {
		r.aiMiddleware.SetEnabled(enabled)
	}
	r.logger.Info("AI integration enabled status changed", zap.Bool("enabled", enabled))
}

// IsEnabled возвращает статус AI интеграции
func (r *RelayAIIntegration) IsEnabled() bool {
	return r.enabled
}

// GetMetrics возвращает метрики AI интеграции
func (r *RelayAIIntegration) GetMetrics() map[string]interface{} {
	if r.aiMiddleware == nil {
		return map[string]interface{}{
			"enabled": r.enabled,
			"status":  "not_initialized",
		}
	}

	metrics := r.aiMiddleware.GetMetrics()
	metrics["integration_enabled"] = r.enabled
	metrics["config"] = r.config

	return metrics
}

// GetHealthStatus возвращает статус здоровья AI интеграции
func (r *RelayAIIntegration) GetHealthStatus() map[string]interface{} {
	status := map[string]interface{}{
		"enabled":     r.enabled,
		"initialized": r.aiMiddleware != nil,
		"timestamp":   time.Now().UTC(),
	}

	if r.aiMiddleware != nil {
		metrics := r.aiMiddleware.GetMetrics()
		status["ai_middleware"] = metrics
	}

	return status
}

// Shutdown корректно завершает работу AI интеграции
func (r *RelayAIIntegration) Shutdown(ctx context.Context) error {
	if !r.enabled || r.aiService == nil {
		return nil
	}

	r.logger.Info("Shutting down AI integration...")

	// Останавливаем AI сервис
	if err := r.aiService.Stop(); err != nil {
		r.logger.Error("Failed to stop AI service", zap.Error(err))
		return err
	}

	r.logger.Info("AI integration shutdown completed")
	return nil
}

// GetDefaultConfig возвращает конфигурацию по умолчанию
func GetDefaultAIIntegrationConfig() *AIIntegrationConfig {
	return &AIIntegrationConfig{
		Enabled:            true,
		ProcessingInterval: 5 * time.Second,
		CircuitBreakerConfig: CircuitBreakerConfig{
			FailureThreshold: 5,
			SuccessThreshold: 3,
			Timeout:          30 * time.Second,
		},
		MetricsConfig: MetricsConfig{
			Enabled:        true,
			ExportInterval: 30 * time.Second,
		},
	}
}

// ValidateConfig проверяет корректность конфигурации
func (c *AIIntegrationConfig) Validate() error {
	if c.ProcessingInterval <= 0 {
		return fmt.Errorf("processing_interval must be positive")
	}

	if c.CircuitBreakerConfig.FailureThreshold <= 0 {
		return fmt.Errorf("circuit_breaker.failure_threshold must be positive")
	}

	if c.CircuitBreakerConfig.SuccessThreshold <= 0 {
		return fmt.Errorf("circuit_breaker.success_threshold must be positive")
	}

	if c.CircuitBreakerConfig.Timeout <= 0 {
		return fmt.Errorf("circuit_breaker.timeout must be positive")
	}

	if c.MetricsConfig.Enabled && c.MetricsConfig.ExportInterval <= 0 {
		return fmt.Errorf("metrics.export_interval must be positive when metrics are enabled")
	}

	return nil
}
