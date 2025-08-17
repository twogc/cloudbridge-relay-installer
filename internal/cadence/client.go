package cadence

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/transport/tchannel"
	"go.uber.org/zap"
)

// WorkflowExecution представляет выполнение workflow
type WorkflowExecution struct {
	ID    string
	RunID string
}

// CadenceClient представляет Cadence клиент для CloudBridge Relay
type CadenceClient struct {
	// Конфигурация
	config *CadenceConfig
	logger *zap.Logger

	// Cadence компоненты
	workflowClient client.Client
	worker         worker.Worker
	dispatcher     *yarpc.Dispatcher

	// Управление жизненным циклом
	ctx    context.Context
	cancel context.CancelFunc

	// Статус
	isRunning bool
}

// NewCadenceClient создает новый Cadence клиент
func NewCadenceClient(config *CadenceConfig, logger *zap.Logger) (*CadenceClient, error) {
	if config == nil {
		return nil, fmt.Errorf("конфигурация не может быть nil")
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &CadenceClient{
		config: config,
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// Start запускает Cadence клиент и worker
func (c *CadenceClient) Start() error {
	c.logger.Info("Запуск Cadence клиента",
		zap.String("domain", c.config.Domain),
		zap.String("hostPort", c.config.Client.HostPort),
	)

	// Создаем YARPC dispatcher
	if err := c.createDispatcher(); err != nil {
		return fmt.Errorf("не удалось создать YARPC dispatcher: %w", err)
	}

	// Запускаем dispatcher
	if err := c.dispatcher.Start(); err != nil {
		return fmt.Errorf("не удалось запустить YARPC dispatcher: %w", err)
	}

	// Создаем Cadence клиент
	if err := c.createWorkflowClient(); err != nil {
		return fmt.Errorf("не удалось создать Cadence клиент: %w", err)
	}

	// Создаем и запускаем worker
	if err := c.createWorker(); err != nil {
		return fmt.Errorf("не удалось создать Cadence worker: %w", err)
	}

	// Запускаем worker в отдельной горутине
	go func() {
		c.logger.Info("Запуск Cadence worker")
		if err := c.worker.Start(); err != nil {
			c.logger.Error("Ошибка при запуске Cadence worker", zap.Error(err))
		}
	}()

	c.isRunning = true
	c.logger.Info("Cadence клиент успешно запущен")

	return nil
}

// Stop останавливает Cadence клиент
func (c *CadenceClient) Stop() error {
	if !c.isRunning {
		return nil
	}

	c.logger.Info("Остановка Cadence клиента")

	// Останавливаем worker
	if c.worker != nil {
		c.worker.Stop()
		c.logger.Info("Cadence worker остановлен")
	}

	// Workflow клиент не требует явного закрытия
	if c.workflowClient != nil {
		c.logger.Info("Cadence workflow клиент освобожден")
	}

	// Останавливаем dispatcher
	if c.dispatcher != nil {
		if err := c.dispatcher.Stop(); err != nil {
			c.logger.Error("Ошибка при остановке YARPC dispatcher", zap.Error(err))
		} else {
			c.logger.Info("YARPC dispatcher остановлен")
		}
	}

	// Отменяем контекст
	c.cancel()

	c.isRunning = false
	c.logger.Info("Cadence клиент успешно остановлен")

	return nil
}

// GetWorkflowClient возвращает Cadence workflow клиент
func (c *CadenceClient) GetWorkflowClient() client.Client {
	return c.workflowClient
}

// GetWorker возвращает Cadence worker
func (c *CadenceClient) GetWorker() worker.Worker {
	return c.worker
}

// StartWorkflow запускает новый workflow
func (c *CadenceClient) StartWorkflow(ctx context.Context, workflowType string, input interface{}) (*workflow.Execution, error) {
	if !c.isRunning {
		return nil, fmt.Errorf("Cadence клиент не запущен")
	}

	options := client.StartWorkflowOptions{
		ID:                           fmt.Sprintf("%s-%d", workflowType, time.Now().Unix()),
		TaskList:                     c.config.Worker.TaskListName,
		ExecutionStartToCloseTimeout: c.config.GetExecutionTimeout(workflowType),
	}

	execution, err := c.workflowClient.StartWorkflow(ctx, options, workflowType, input)
	if err != nil {
		c.logger.Error("Не удалось запустить workflow",
			zap.String("workflowType", workflowType),
			zap.Error(err),
		)
		return nil, err
	}

	c.logger.Info("Workflow успешно запущен",
		zap.String("workflowType", workflowType),
		zap.String("workflowID", execution.ID),
		zap.String("runID", execution.RunID),
	)

	return execution, nil
}

// GetWorkflowResult получает результат выполнения workflow
func (c *CadenceClient) GetWorkflowResult(ctx context.Context, workflowID, runID string, result interface{}) error {
	if !c.isRunning {
		return fmt.Errorf("Cadence клиент не запущен")
	}

	run := c.workflowClient.GetWorkflow(ctx, workflowID, runID)
	return run.Get(ctx, result)
}

// TerminateWorkflow принудительно завершает workflow
func (c *CadenceClient) TerminateWorkflow(ctx context.Context, workflowID, runID, reason string) error {
	if !c.isRunning {
		return fmt.Errorf("Cadence клиент не запущен")
	}

	err := c.workflowClient.TerminateWorkflow(ctx, workflowID, runID, reason, nil)
	if err != nil {
		c.logger.Error("Не удалось завершить workflow",
			zap.String("workflowID", workflowID),
			zap.String("runID", runID),
			zap.String("reason", reason),
			zap.Error(err),
		)
		return err
	}

	c.logger.Info("Workflow успешно завершен",
		zap.String("workflowID", workflowID),
		zap.String("runID", runID),
		zap.String("reason", reason),
	)

	return nil
}

// QueryWorkflow выполняет запрос к workflow
func (c *CadenceClient) QueryWorkflow(ctx context.Context, workflowID, runID, queryType string, args ...interface{}) (interface{}, error) {
	if !c.isRunning {
		return nil, fmt.Errorf("Cadence клиент не запущен")
	}

	response, err := c.workflowClient.QueryWorkflow(ctx, workflowID, runID, queryType, args...)
	if err != nil {
		c.logger.Error("Не удалось выполнить запрос к workflow",
			zap.String("workflowID", workflowID),
			zap.String("runID", runID),
			zap.String("queryType", queryType),
			zap.Error(err),
		)
		return nil, err
	}

	c.logger.Debug("Запрос к workflow выполнен успешно",
		zap.String("workflowID", workflowID),
		zap.String("runID", runID),
		zap.String("queryType", queryType),
	)

	return response, nil
}

// SignalWorkflow отправляет сигнал workflow
func (c *CadenceClient) SignalWorkflow(ctx context.Context, workflowID, runID, signalName string, arg interface{}) error {
	if !c.isRunning {
		return fmt.Errorf("Cadence клиент не запущен")
	}

	err := c.workflowClient.SignalWorkflow(ctx, workflowID, runID, signalName, arg)
	if err != nil {
		c.logger.Error("Не удалось отправить сигнал workflow",
			zap.String("workflowID", workflowID),
			zap.String("runID", runID),
			zap.String("signalName", signalName),
			zap.Error(err),
		)
		return err
	}

	c.logger.Debug("Сигнал workflow отправлен успешно",
		zap.String("workflowID", workflowID),
		zap.String("runID", runID),
		zap.String("signalName", signalName),
	)

	return nil
}

// IsRunning возвращает статус клиента
func (c *CadenceClient) IsRunning() bool {
	return c.isRunning
}

// GetConfig возвращает конфигурацию клиента
func (c *CadenceClient) GetConfig() *CadenceConfig {
	return c.config
}

// createDispatcher создает YARPC dispatcher для подключения к Cadence
func (c *CadenceClient) createDispatcher() error {
	ch, err := tchannel.NewChannelTransport(tchannel.ServiceName(c.config.ServiceName))
	if err != nil {
		return fmt.Errorf("не удалось создать TChannel транспорт: %w", err)
	}

	dispatcher := yarpc.NewDispatcher(yarpc.Config{
		Name: c.config.ServiceName,
		Outbounds: yarpc.Outbounds{
			"cadence-frontend": {Unary: ch.NewSingleOutbound(c.config.Client.HostPort)},
		},
	})

	c.dispatcher = dispatcher
	return nil
}

// createWorkflowClient создает Cadence workflow клиент
func (c *CadenceClient) createWorkflowClient() error {
	// Создаем service client
	serviceClient := workflowserviceclient.New(c.dispatcher.ClientConfig("cadence-frontend"))

	workflowClient := client.NewClient(serviceClient, c.config.Client.Domain, &client.Options{})

	c.workflowClient = workflowClient
	return nil
}

// createWorker создает Cadence worker
func (c *CadenceClient) createWorker() error {
	// Создаем service client для worker
	serviceClient := workflowserviceclient.New(c.dispatcher.ClientConfig("cadence-frontend"))

	workerInstance := worker.New(
		serviceClient,
		c.config.Client.Domain,
		c.config.Worker.TaskListName,
		worker.Options{
			MaxConcurrentActivityExecutionSize:      c.config.Worker.MaxConcurrentActivityExecutionSize,
			MaxConcurrentLocalActivityExecutionSize: c.config.Worker.MaxConcurrentLocalActivityExecutionSize,
			MaxConcurrentActivityTaskPollers:        10,
			MaxConcurrentDecisionTaskPollers:        10,
			Identity:                                c.config.Worker.Identity,
		},
	)

	c.worker = workerInstance
	return nil
}
