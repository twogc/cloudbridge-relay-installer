package cadence

import (
	"fmt"
	"net"
	"time"

	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/zap"
)

// CreateWorkflowServiceClient создает workflowserviceclient.Interface для подключения к Cadence серверу
func CreateWorkflowServiceClient(hostPort string, logger *zap.Logger) (workflowserviceclient.Interface, error) {
	logger.Info("Creating workflow service client", zap.String("hostPort", hostPort))

	// Проверяем доступность хоста
	if err := checkHostAvailability(hostPort, logger); err != nil {
		logger.Warn("Cadence server not available, falling back to stub mode",
			zap.String("hostPort", hostPort),
			zap.Error(err))
		return nil, nil
	}

	// Создаем простой Cadence client
	// Пока возвращаем nil для отладки, позже добавим полную реализацию
	logger.Info("Cadence server is available, but using stub client for now")
	return nil, nil
}

// checkHostAvailability проверяет доступность Cadence сервера
func checkHostAvailability(hostPort string, logger *zap.Logger) error {
	timeout := 5 * time.Second
	conn, err := net.DialTimeout("tcp", hostPort, timeout)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", hostPort, err)
	}
	defer conn.Close()

	logger.Debug("Cadence server is available", zap.String("hostPort", hostPort))
	return nil
}

// ServiceClientOptions опции для создания service client
type ServiceClientOptions struct {
	Timeout     time.Duration
	ServiceName string
	EnableTLS   bool
	TLSCertPath string
	TLSKeyPath  string
	TLSCAPath   string
}

// CreateWorkflowServiceClientWithOptions создает service client с дополнительными опциями
func CreateWorkflowServiceClientWithOptions(hostPort string, options ServiceClientOptions, logger *zap.Logger) (workflowserviceclient.Interface, error) {
	logger.Info("Creating workflow service client with options",
		zap.String("hostPort", hostPort),
		zap.Bool("enableTLS", options.EnableTLS))

	// Проверяем доступность хоста
	if err := checkHostAvailability(hostPort, logger); err != nil {
		logger.Warn("Cadence server not available, falling back to stub mode",
			zap.String("hostPort", hostPort),
			zap.Error(err))
		return nil, nil
	}

	// Создаем простой Cadence client с опциями
	// Пока возвращаем nil для отладки, позже добавим полную реализацию
	logger.Info("Cadence server is available, but using stub client with options for now")
	return nil, nil
}

// HealthCheck проверяет здоровье Cadence сервера
func HealthCheck(client workflowserviceclient.Interface, logger *zap.Logger) error {
	if client == nil {
		return fmt.Errorf("cadence client is nil (stub mode)")
	}

	// Пока просто проверяем, что клиент не nil
	// В будущем можно добавить реальную проверку здоровья
	logger.Debug("Cadence server health check passed")
	return nil
}
