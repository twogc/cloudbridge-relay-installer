package ingress

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/balancer"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	"github.com/twogc/cloudbridge-relay-installer/internal/metrics"
	"go.uber.org/zap"
)

// IngressIntegration интегрирует ingress систему с основным balancer
type IngressIntegration struct {
	config     *IngressConfig
	ruleEngine *IngressRuleEngine
	balancer   *balancer.Balancer
	logger     *logger.Logger
	metrics    *metrics.Metrics

	// HTTP сервер для ingress
	httpServer *http.Server
	mu         sync.RWMutex
}

// NewIngressIntegration создает новую интеграцию ingress
func NewIngressIntegration(config *IngressConfig, balancer *balancer.Balancer, logger *logger.Logger, metrics *metrics.Metrics) (*IngressIntegration, error) {
	if config == nil {
		return nil, fmt.Errorf("ingress config is required")
	}

	if balancer == nil {
		return nil, fmt.Errorf("balancer is required")
	}

	// Создаем rule engine
	ruleEngine, err := NewIngressRuleEngine(config, logger, metrics)
	if err != nil {
		return nil, fmt.Errorf("failed to create rule engine: %w", err)
	}

	integration := &IngressIntegration{
		config:     config,
		ruleEngine: ruleEngine,
		balancer:   balancer,
		logger:     logger,
		metrics:    metrics,
	}

	// Создаем HTTP сервер если включен
	if config.Enabled {
		if err := integration.setupHTTPServer(); err != nil {
			return nil, fmt.Errorf("failed to setup HTTP server: %w", err)
		}
	}

	return integration, nil
}

// setupHTTPServer настраивает HTTP сервер для ingress
func (i *IngressIntegration) setupHTTPServer() error {
	mux := http.NewServeMux()

	// Основной обработчик для всех запросов
	mux.HandleFunc("/", i.handleRequest)

	// Health check endpoint
	mux.HandleFunc("/health", i.handleHealth)

	// Metrics endpoint
	if i.metrics != nil {
		// TODO: Добавить метод Handler в metrics
		mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			if _, err := w.Write([]byte("# Metrics endpoint - TODO: implement\n")); err != nil {
				i.logger.Error("Failed to write metrics response", zap.Error(err))
			}
		})
	}

	i.httpServer = &http.Server{
		Addr:    ":8080", // TODO: сделать конфигурируемым
		Handler: mux,
	}

	i.logger.Info("Ingress HTTP server configured",
		zap.String("addr", i.httpServer.Addr),
		zap.Bool("bastion_mode", i.config.BastionMode),
	)

	return nil
}

// handleRequest обрабатывает входящие HTTP запросы
func (i *IngressIntegration) handleRequest(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Проверяем IP access control
	if i.config.IPAccessControl.Enabled {
		if !i.ruleEngine.ipPolicy.Allow(r.RemoteAddr) {
			i.logger.Warn("Access denied by IP policy",
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("host", r.Host),
				zap.String("path", r.URL.Path),
			)

			if i.metrics != nil {
				i.metrics.IncErrors("ip_access_denied")
			}

			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}
	}

	// Находим подходящее правило
	rule, _ := i.ruleEngine.FindMatchingRule(r.Host, r.URL.Path)
	if rule == nil {
		i.logger.Error("Failed to find matching rule",
			zap.String("host", r.Host),
			zap.String("path", r.URL.Path),
		)

		if i.metrics != nil {
			i.metrics.IncErrors("no_matching_rule")
		}

		http.Error(w, "No matching rule found", http.StatusNotFound)
		return
	}

	// Обрабатываем запрос через rule engine
	err := i.ruleEngine.HandleRequest(w, r)

	// Обновляем метрики
	duration := time.Since(startTime)
	if i.metrics != nil {
		i.metrics.ObserveMessageLatency("ingress_request", duration)
		if err != nil {
			i.metrics.IncErrors("ingress_error")
		} else {
			i.metrics.IncConnections()
		}
	}

	if err != nil {
		i.logger.Error("Failed to handle request",
			zap.String("host", r.Host),
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)

		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleHealth обрабатывает health check запросы
func (i *IngressIntegration) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(`{"status":"healthy","service":"ingress"}`)); err != nil {
		i.logger.Error("Failed to write health response", zap.Error(err))
	}
}

// Start запускает ingress интеграцию
func (i *IngressIntegration) Start(ctx context.Context) error {
	if !i.config.Enabled {
		i.logger.Info("Ingress integration is disabled")
		return nil
	}

	if i.httpServer == nil {
		return fmt.Errorf("HTTP server not configured")
	}

	i.logger.Info("Starting ingress integration")

	// Запускаем HTTP сервер в горутине
	go func() {
		if err := i.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			i.logger.Error("HTTP server error", zap.Error(err))
		}
	}()

	i.logger.Info("Ingress integration started successfully")
	return nil
}

// Stop останавливает ingress интеграцию
func (i *IngressIntegration) Stop(ctx context.Context) error {
	if !i.config.Enabled {
		return nil
	}

	i.logger.Info("Stopping ingress integration")

	// Останавливаем HTTP сервер
	if i.httpServer != nil {
		if err := i.httpServer.Shutdown(ctx); err != nil {
			i.logger.Error("Failed to shutdown HTTP server", zap.Error(err))
			return err
		}
	}

	// Останавливаем rule engine
	if i.ruleEngine != nil {
		if err := i.ruleEngine.Stop(ctx); err != nil {
			i.logger.Error("Failed to stop rule engine", zap.Error(err))
			return err
		}
	}

	i.logger.Info("Ingress integration stopped successfully")
	return nil
}

// GetStats возвращает статистику ingress системы
func (i *IngressIntegration) GetStats() map[string]interface{} {
	stats := map[string]interface{}{
		"enabled":      i.config.Enabled,
		"bastion_mode": i.config.BastionMode,
	}

	if i.ruleEngine != nil {
		stats["rules_count"] = len(i.ruleEngine.rules)
		stats["ip_policy_enabled"] = i.config.IPAccessControl.Enabled
	}

	return stats
}
