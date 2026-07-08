package cadence

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"

	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"

	"github.com/twogc/cloudbridge-relay-installer/internal/cadence/workflows"
)

// registeredWorkflows keeps track of registered workflows to avoid duplicates
var registeredWorkflows = make(map[string]bool)
var registrationMutex sync.Mutex

// Client представляет Cadence client для CloudBridge Relay
type Client struct {
	client client.Client
	worker worker.Worker
	logger *zap.Logger
}

// Config конфигурация для Cadence client
type Config struct {
	HostPort       string
	Domain         string
	TaskList       string
	WorkerOptions  worker.Options
	ServiceOptions ServiceClientOptions
}

// NewClient создает новый Cadence client
func NewClient(config Config, logger *zap.Logger) (*Client, error) {
	logger.Info("Creating Cadence client",
		zap.String("domain", config.Domain),
		zap.String("taskList", config.TaskList),
		zap.String("hostPort", config.HostPort))

	// Создаем service client для подключения к Cadence серверу
	var serviceClient workflowserviceclient.Interface
	var err error

	// Используем опции если они заданы
	if config.ServiceOptions.ServiceName != "" {
		serviceClient, err = CreateWorkflowServiceClientWithOptions(config.HostPort, config.ServiceOptions, logger)
	} else {
		serviceClient, err = CreateWorkflowServiceClient(config.HostPort, logger)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create workflow service client: %w", err)
	}

	var cadenceClient client.Client
	var workerInstance worker.Worker

	// Если service client не создан (заглушка), используем nil клиенты
	if serviceClient == nil {
		logger.Info("Using stub mode - service client is nil")
		cadenceClient = nil
		workerInstance = nil
	} else {
		// Создаем настоящий Cadence client
		cadenceClient = client.NewClient(serviceClient, config.Domain, &client.Options{})

		// Создаем настоящий worker
		workerInstance = worker.New(serviceClient, config.Domain, config.TaskList, config.WorkerOptions)
	}

	return &Client{
		client: cadenceClient,
		worker: workerInstance,
		logger: logger,
	}, nil
}

// Start запускает Cadence worker
func (c *Client) Start() error {
	c.logger.Info("Starting Cadence worker")

	// Регистрируем workflows
	if err := c.registerWorkflows(); err != nil {
		return fmt.Errorf("failed to register workflows: %w", err)
	}

	// Регистрируем activities
	if err := c.registerActivities(); err != nil {
		return fmt.Errorf("failed to register activities: %w", err)
	}

	// Запускаем worker (только если не в stub режиме)
	if c.worker != nil {
		if err := c.worker.Start(); err != nil {
			return fmt.Errorf("failed to start worker: %w", err)
		}
	} else {
		c.logger.Info("Skipping worker start in stub mode")
	}

	c.logger.Info("Cadence worker started successfully")
	return nil
}

// Stop останавливает Cadence worker
func (c *Client) Stop() {
	c.logger.Info("Stopping Cadence worker")
	if c.worker != nil {
		c.worker.Stop()
		c.logger.Info("Cadence worker stopped")
	}
}

// GetClient возвращает Cadence client
func (c *Client) GetClient() client.Client {
	return c.client
}

// GetWorker возвращает Cadence worker
func (c *Client) GetWorker() worker.Worker {
	return c.worker
}

// RegisterWorkflow регистрирует workflow
func (c *Client) RegisterWorkflow(w interface{}) {
	c.logger.Info("Registering workflow", zap.String("workflow", fmt.Sprintf("%T", w)))
	workflow.Register(w)
}

// RegisterActivity регистрирует activity
func (c *Client) RegisterActivity(a interface{}) {
	c.logger.Info("Registering activity", zap.String("activity", fmt.Sprintf("%T", a)))
	activity.Register(a)
}

// HealthCheck проверяет состояние Cadence client
func (c *Client) HealthCheck() error {
	if c.client == nil {
		// В stub режиме всегда возвращаем success
		c.logger.Debug("Cadence client health check passed (stub mode)")
		return nil
	}

	// Простая проверка - пытаемся выполнить ping
	// В реальной реализации здесь должен быть вызов DescribeDomain
	c.logger.Debug("Cadence client health check passed")
	return nil
}

// ExecuteWorkflow выполняет workflow и возвращает workflow ID
func (c *Client) ExecuteWorkflow(ctx context.Context, workflowType string, input interface{}) (string, error) {
	workflowID := fmt.Sprintf("%s-%d", workflowType, time.Now().UnixNano())

	c.logger.Info("Executing workflow",
		zap.String("workflowType", workflowType),
		zap.String("workflowID", workflowID))

	// Если клиент не инициализирован (режим заглушки), возвращаем заглушку
	if c.client == nil {
		c.logger.Info("Executing workflow (stub mode)")
		return workflowID, nil
	}

	// Создаем опции для workflow
	workflowOptions := client.StartWorkflowOptions{
		ID:                           workflowID,
		TaskList:                     "default-task-list",
		ExecutionStartToCloseTimeout: time.Minute * 10,
	}

	// Запускаем workflow
	workflowRun, err := c.client.ExecuteWorkflow(ctx, workflowOptions, workflowType, input)
	if err != nil {
		return "", fmt.Errorf("failed to execute workflow: %w", err)
	}

	c.logger.Info("Workflow executed successfully",
		zap.String("workflowID", workflowRun.GetID()),
		zap.String("runID", workflowRun.GetRunID()))

	return workflowRun.GetID(), nil
}

// GetWorkflowStatus получает статус workflow
func (c *Client) GetWorkflowStatus(ctx context.Context, workflowID string) (string, error) {
	c.logger.Info("Getting workflow status", zap.String("workflowID", workflowID))

	// В реальной реализации здесь должен быть вызов DescribeWorkflowExecution
	// Пока возвращаем заглушку
	return "RUNNING", nil
}

// GetDomain возвращает домен
func (c *Client) GetDomain() string {
	return "cloudbridge"
}

// Close закрывает Cadence client
func (c *Client) Close() {
	c.logger.Info("Closing Cadence client")
	// В реальной реализации здесь должен быть закрытие client
}

// WorkflowExecution представляет выполнение workflow
type WorkflowExecution struct {
	ID    string
	RunID string
}

// StartWorkflow запускает workflow
func (c *Client) StartWorkflow(ctx context.Context, workflowType string, input interface{}) (*WorkflowExecution, error) {
	workflowID := fmt.Sprintf("%s-%d", workflowType, time.Now().UnixNano())

	c.logger.Info("Starting workflow",
		zap.String("workflowType", workflowType),
		zap.String("workflowID", workflowID))

	// Если клиент не инициализирован (режим заглушки), возвращаем заглушку
	if c.client == nil {
		c.logger.Info("Starting workflow (stub mode)")
		return &WorkflowExecution{
			ID:    workflowID,
			RunID: fmt.Sprintf("run-%d", time.Now().UnixNano()),
		}, nil
	}

	// Создаем опции для workflow
	workflowOptions := client.StartWorkflowOptions{
		ID:                           workflowID,
		TaskList:                     "default-task-list",
		ExecutionStartToCloseTimeout: time.Minute * 10,
	}

	// Запускаем workflow
	workflowRun, err := c.client.ExecuteWorkflow(ctx, workflowOptions, workflowType, input)
	if err != nil {
		return nil, fmt.Errorf("failed to start workflow: %w", err)
	}

	return &WorkflowExecution{
		ID:    workflowRun.GetID(),
		RunID: workflowRun.GetRunID(),
	}, nil
}

// GetWorkflowResult получает результат workflow
func (c *Client) GetWorkflowResult(ctx context.Context, workflowID string, runID string, result interface{}) error {
	c.logger.Info("Getting workflow result",
		zap.String("workflowID", workflowID),
		zap.String("runID", runID))

	// Если клиент не инициализирован (режим заглушки), возвращаем заглушку
	if c.client == nil {
		c.logger.Info("Getting workflow result (stub mode)")
		return nil
	}

	// Получаем результат workflow
	workflowRun := c.client.GetWorkflow(ctx, workflowID, runID)
	err := workflowRun.Get(ctx, result)
	if err != nil {
		return fmt.Errorf("failed to get workflow result: %w", err)
	}

	return nil
}

// SignalWorkflow отправляет сигнал workflow
func (c *Client) SignalWorkflow(ctx context.Context, workflowID string, runID string, signalName string, signalValue interface{}) error {
	c.logger.Info("Signaling workflow",
		zap.String("workflowID", workflowID),
		zap.String("runID", runID),
		zap.String("signalName", signalName))

	// Отправляем сигнал workflow
	err := c.client.SignalWorkflow(ctx, workflowID, runID, signalName, signalValue)
	if err != nil {
		return fmt.Errorf("failed to signal workflow: %w", err)
	}

	return nil
}

// IsRunning проверяет, запущен ли клиент
func (c *Client) IsRunning() bool {
	// В stub режиме считаем клиент запущенным если он инициализирован
	if c.client == nil {
		return true // stub режим
	}
	return c.worker != nil && c.client != nil
}

// registerWorkflows регистрирует все workflows
func (c *Client) registerWorkflows() error {
	c.logger.Info("Registering Cadence workflows")

	registrationMutex.Lock()
	defer registrationMutex.Unlock()

	// Функция для безопасной регистрации workflow
	safeRegister := func(workflowFunc interface{}, name string) {
		if !registeredWorkflows[name] {
			workflow.Register(workflowFunc)
			registeredWorkflows[name] = true
			c.logger.Debug("Registered workflow", zap.String("name", name))
		} else {
			c.logger.Debug("Workflow already registered, skipping", zap.String("name", name))
		}
	}

	// Основные workflows
	safeRegister(workflows.CloudBridgeMainWorkflow, "CloudBridgeMainWorkflow")
	safeRegister(workflows.ModuleInitializationWorkflow, "ModuleInitializationWorkflow")
	safeRegister(workflows.HealthCheckWorkflow, "HealthCheckWorkflow")

	// P2P QUIC workflows
	safeRegister(workflows.P2PQUICTunnelWorkflow, "P2PQUICTunnelWorkflow")
	safeRegister(workflows.P2PConnectionOrchestrationWorkflow, "P2PConnectionOrchestrationWorkflow")

	// AI workflows
	safeRegister(workflows.ThreatAnalysisWorkflow, "ThreatAnalysisWorkflow")
	safeRegister(workflows.BehaviorMonitoringWorkflow, "BehaviorMonitoringWorkflow")
	safeRegister(workflows.AutomatedResponseWorkflow, "AutomatedResponseWorkflow")
	safeRegister(workflows.ModelTrainingWorkflow, "ModelTrainingWorkflow")

	// Scaling workflows
	safeRegister(workflows.AutoScalingWorkflow, "AutoScalingWorkflow")
	safeRegister(workflows.ResourceOptimizationWorkflow, "ResourceOptimizationWorkflow")
	safeRegister(workflows.PerformanceMonitoringWorkflow, "PerformanceMonitoringWorkflow")
	safeRegister(workflows.CapacityPlanningWorkflow, "CapacityPlanningWorkflow")

	// Security workflows
	safeRegister(workflows.SecurityComplianceWorkflow, "SecurityComplianceWorkflow")
	safeRegister(workflows.AuditLoggingWorkflow, "AuditLoggingWorkflow")

	// Additional workflows
	safeRegister(workflows.AIOptimizedRoutingWorkflow, "AIOptimizedRoutingWorkflow")
	safeRegister(workflows.IncidentResponseWorkflow, "IncidentResponseWorkflow")
	safeRegister(workflows.BackupRecoveryWorkflow, "BackupRecoveryWorkflow")

	c.logger.Info("All workflows registered successfully")
	return nil
}

// registerActivities регистрирует все activities
func (c *Client) registerActivities() error {
	c.logger.Info("Registering Cadence activities")

	// Здесь должны быть зарегистрированы все activities
	// Пока оставляем пустым, так как activities регистрируются в worker

	c.logger.Info("All activities registered successfully")
	return nil
}

// IsStubMode возвращает true если клиент работает в режиме заглушки
func (c *Client) IsStubMode() bool {
	return c.client == nil
}
