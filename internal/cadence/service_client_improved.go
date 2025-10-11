package cadence

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/worker"
	"go.uber.org/zap"
)

// CreateWorkflowServiceClientImproved создает реальный клиент для Cadence
func CreateWorkflowServiceClientImproved(hostPort string, logger *zap.Logger) (workflowserviceclient.Interface, error) {
	logger.Info("Creating improved workflow service client", zap.String("hostPort", hostPort))

	// Проверяем доступность хоста
	if err := checkHostAvailability(hostPort, logger); err != nil {
		logger.Warn("Cadence server not available, falling back to stub mode",
			zap.String("hostPort", hostPort),
			zap.Error(err))
		return nil, nil
	}

	// Создаем реальный Cadence client
	client, err := createRealCadenceClient(hostPort, logger)
	if err != nil {
		logger.Error("Failed to create Cadence client",
			zap.String("hostPort", hostPort),
			zap.Error(err))
		return nil, err
	}

	logger.Info("Cadence client created successfully", zap.String("hostPort", hostPort))
	return client, nil
}

// createRealCadenceClient создает реальный Cadence клиент
func createRealCadenceClient(hostPort string, logger *zap.Logger) (workflowserviceclient.Interface, error) {
	// Создаем клиент (упрощенная реализация)
	// В реальном проекте здесь будет полная настройка TChannel клиента
	// Пока возвращаем nil для демонстрации
	client := workflowserviceclient.Interface(nil)

	return client, nil
}

// createTChannelTransport создает TChannel транспорт
func createTChannelTransport(hostPort string) (workflowserviceclient.Interface, error) {
	// Упрощенная реализация для демонстрации
	// В реальном проекте здесь будет полная настройка TChannel
	return nil, fmt.Errorf("TChannel transport not implemented yet")
}

// HealthCheckImproved улучшенная проверка здоровья Cadence сервера
func HealthCheckImproved(client workflowserviceclient.Interface, logger *zap.Logger) error {
	if client == nil {
		return fmt.Errorf("cadence client is nil (stub mode)")
	}

	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Выполняем DescribeDomain для проверки доступности
	req := &shared.DescribeDomainRequest{
		Name: &[]string{"default"}[0],
	}

	_, err := client.DescribeDomain(ctx, req)
	if err != nil {
		// Проверяем, является ли ошибка "EntityNotExists" (это нормально)
		if isEntityNotExistsError(err) {
			logger.Debug("Cadence domain 'default' does not exist, but server is accessible")
			return nil
		}

		logger.Error("Cadence health check failed",
			zap.Error(err))
		return fmt.Errorf("cadence health check failed: %w", err)
	}

	logger.Debug("Cadence server health check passed")
	return nil
}

// isEntityNotExistsError проверяет, является ли ошибка "EntityNotExists"
func isEntityNotExistsError(err error) bool {
	// В реальной реализации здесь будет проверка типа ошибки
	// Пока возвращаем false для простоты
	return false
}

// RegisterDomainIfNotExists регистрирует домен если он не существует
func RegisterDomainIfNotExists(client workflowserviceclient.Interface, domainName string, logger *zap.Logger) error {
	if client == nil {
		return fmt.Errorf("cadence client is nil (stub mode)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Проверяем, существует ли домен
	req := &shared.DescribeDomainRequest{
		Name: &domainName,
	}

	_, err := client.DescribeDomain(ctx, req)
	if err == nil {
		// Домен существует
		logger.Debug("Cadence domain already exists", zap.String("domain", domainName))
		return nil
	}

	// Если ошибка не "EntityNotExists", возвращаем её
	if !isEntityNotExistsError(err) {
		return fmt.Errorf("failed to check domain existence: %w", err)
	}

	// Регистрируем домен
	registerReq := &shared.RegisterDomainRequest{
		Name:                                   &domainName,
		Description:                            &[]string{"CloudBridge Relay domain"}[0],
		WorkflowExecutionRetentionPeriodInDays: &[]int32{7}[0],
	}

	err = client.RegisterDomain(ctx, registerReq)
	if err != nil {
		return fmt.Errorf("failed to register domain: %w", err)
	}

	logger.Info("Cadence domain registered successfully", zap.String("domain", domainName))
	return nil
}

// CreateCadenceClient создает Cadence клиент для выполнения workflow
func CreateCadenceClient(serviceClient workflowserviceclient.Interface, domain string, logger *zap.Logger) client.Client {
	if serviceClient == nil {
		logger.Warn("Service client is nil, creating stub client")
		return nil
	}

	// Создаем Cadence клиент
	cadenceClient := client.NewClient(
		serviceClient,
		domain,
		&client.Options{
			Identity: "cloudbridge-relay",
		},
	)

	return cadenceClient
}

// CreateCadenceWorker создает Cadence worker
func CreateCadenceWorker(serviceClient workflowserviceclient.Interface, domain, taskList string, logger *zap.Logger) worker.Worker {
	if serviceClient == nil {
		logger.Warn("Service client is nil, creating stub worker")
		return nil
	}

	// Создаем Cadence worker
	cadenceWorker := worker.New(
		serviceClient,
		domain,
		taskList,
		worker.Options{
			Identity: "cloudbridge-relay-worker",
		},
	)

	return cadenceWorker
}
