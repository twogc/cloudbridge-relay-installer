package relay

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/config"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai"
	"go.uber.org/zap"
)

// ServerAIIntegration представляет интеграцию AI Middleware с основным сервером
type ServerAIIntegration struct {
	aiMiddleware *AIMiddleware
	aiService    AIServiceInterface
	logger       *zap.Logger
	enabled      bool
}

// NewServerAIIntegration создает новую интеграцию AI с сервером
func NewServerAIIntegration(cfg *config.Config, logger *zap.Logger) (*ServerAIIntegration, error) {
	if !cfg.AI.Enabled {
		return &ServerAIIntegration{
			enabled: false,
			logger:  logger,
		}, nil
	}

	// Инициализируем AI Service
	aiConfig := &ai.Config{
		AI: ai.AI{
			Enabled:               cfg.AI.Enabled,
			Mode:                  "production",
			LogLevel:              "info",
			MaxConcurrentRequests: 100,
			RequestTimeout:        30 * time.Second,
			CacheTTL:              5 * time.Minute,
		},
	}
	aiService := ai.NewService(aiConfig, logger)
	if err := aiService.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize AI service: %v", err)
	}

	// Инициализируем AI Middleware
	aiMiddleware := NewAIMiddleware(aiService, logger)
	if err := aiMiddleware.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize AI middleware: %v", err)
	}

	logger.Info("AI Server Integration initialized successfully",
		zap.Bool("enabled", true),
		zap.Bool("ai_middleware", aiMiddleware.IsEnabled()),
	)

	return &ServerAIIntegration{
		aiMiddleware: aiMiddleware,
		aiService:    aiService,
		logger:       logger,
		enabled:      true,
	}, nil
}

// IsEnabled возвращает true, если AI интеграция включена
func (sai *ServerAIIntegration) IsEnabled() bool {
	return sai.enabled
}

// GetAIMiddleware возвращает AI Middleware
func (sai *ServerAIIntegration) GetAIMiddleware() *AIMiddleware {
	return sai.aiMiddleware
}

// GetAIService возвращает AI Service
func (sai *ServerAIIntegration) GetAIService() AIServiceInterface {
	return sai.aiService
}

// ProcessConnection обрабатывает соединение через AI Middleware
func (sai *ServerAIIntegration) ProcessConnection(conn net.Conn) error {
	if !sai.enabled || sai.aiMiddleware == nil {
		return nil // AI не включен, пропускаем
	}

	if !sai.aiMiddleware.IsEnabled() {
		return nil // AI Middleware не активен
	}

	// Обрабатываем соединение через AI Middleware
	if err := sai.aiMiddleware.ProcessConnection(context.Background(), conn); err != nil {
		sai.logger.Warn("AI middleware processing failed",
			zap.Error(err),
			zap.String("remote_addr", conn.RemoteAddr().String()),
		)
		// Не возвращаем ошибку, чтобы не блокировать соединение
		return nil
	}

	return nil
}

// GetMetrics возвращает метрики AI интеграции
func (sai *ServerAIIntegration) GetMetrics() map[string]interface{} {
	if !sai.enabled {
		return map[string]interface{}{
			"enabled": false,
		}
	}

	metrics := map[string]interface{}{
		"enabled": true,
	}

	if sai.aiMiddleware != nil {
		aiMetrics := sai.aiMiddleware.GetMetrics()
		for k, v := range aiMetrics {
			metrics[k] = v
		}
	}

	if sai.aiService != nil {
		status := sai.aiService.GetStatus()
		metrics["ai_service_status"] = status
	}

	return metrics
}

// Shutdown корректно завершает работу AI интеграции
func (sai *ServerAIIntegration) Shutdown() error {
	if !sai.enabled {
		return nil
	}

	sai.logger.Info("Shutting down AI Server Integration")

	if sai.aiService != nil {
		if err := sai.aiService.Stop(); err != nil {
			sai.logger.Error("Failed to stop AI service", zap.Error(err))
		}
	}

	sai.logger.Info("AI Server Integration shutdown completed")
	return nil
}
