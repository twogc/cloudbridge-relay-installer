package main

import (
	"fmt"
	"net"

	"github.com/twogc/cloudbridge-relay-installer/config"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	"github.com/twogc/cloudbridge-relay-installer/internal/relay"
	"go.uber.org/zap"
)

// AIComponents содержит все AI компоненты сервера
type AIComponents struct {
	Enabled            bool
	BehaviorAnalyzer   *ai.BehaviorAnalyzer
	PredictiveSecurity *ai.PredictiveSecurity
	AutomatedResponse  *ai.AutomatedResponse
	Monitoring         *ai.AIMonitoring
	Middleware         *relay.AIMiddleware
	ServerIntegration  *relay.ServerAIIntegration
}

// InitializeAIComponents инициализирует все AI компоненты
func InitializeAIComponents(cfg *config.Config, logger *logger.Logger) (*AIComponents, error) {
	aiComponents := &AIComponents{
		Enabled: cfg.AI.Enabled,
	}

	if !cfg.AI.Enabled {
		logger.Info("AI Analytics disabled")
		return aiComponents, nil
	}

	// Инициализируем базовые AI компоненты
	aiComponents.BehaviorAnalyzer = ai.NewBehaviorAnalyzer()
	aiComponents.PredictiveSecurity = ai.NewPredictiveSecurity()
	aiComponents.AutomatedResponse = ai.NewAutomatedResponse()
	aiComponents.Monitoring = ai.NewAIMonitoring()

	// Инициализируем AI Service Integration
	serverIntegration, err := relay.NewServerAIIntegration(cfg, logger.GetZapLogger())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize AI server integration: %v", err)
	}
	aiComponents.ServerIntegration = serverIntegration

	// Получаем AI Middleware из интеграции
	if serverIntegration.IsEnabled() {
		aiComponents.Middleware = serverIntegration.GetAIMiddleware()
	}

	logger.Info("AI Analytics initialized successfully",
		zap.Bool("behavior_analysis", aiComponents.BehaviorAnalyzer.Enabled),
		zap.Bool("predictive_security", aiComponents.PredictiveSecurity.Enabled),
		zap.Bool("automated_response", aiComponents.AutomatedResponse.Enabled),
		zap.Bool("ai_middleware", aiComponents.Middleware != nil && aiComponents.Middleware.IsEnabled()),
		zap.Bool("server_integration", aiComponents.ServerIntegration.IsEnabled()),
	)

	return aiComponents, nil
}

// ProcessConnectionWithAI обрабатывает соединение через AI Middleware
func (ai *AIComponents) ProcessConnectionWithAI(conn net.Conn) error {
	if !ai.Enabled || ai.ServerIntegration == nil {
		return nil // AI не включен
	}

	return ai.ServerIntegration.ProcessConnection(conn)
}

// GetAIMetrics возвращает метрики AI компонентов
func (ai *AIComponents) GetAIMetrics() map[string]interface{} {
	if !ai.Enabled {
		return map[string]interface{}{
			"enabled": false,
		}
	}

	metrics := map[string]interface{}{
		"enabled": true,
	}

	if ai.ServerIntegration != nil {
		integrationMetrics := ai.ServerIntegration.GetMetrics()
		for k, v := range integrationMetrics {
			metrics[k] = v
		}
	}

	return metrics
}

// ShutdownAI корректно завершает работу AI компонентов
func (ai *AIComponents) ShutdownAI() error {
	if !ai.Enabled || ai.ServerIntegration == nil {
		return nil
	}

	return ai.ServerIntegration.Shutdown()
}
