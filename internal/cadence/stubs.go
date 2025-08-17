//go:build !cadence
// +build !cadence

package cadence

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// Stubs for Cadence types when real Cadence is not available
// This allows the code to compile without actual Cadence dependencies

// Client представляет заглушку для Cadence клиента
type Client struct {
	logger  *zap.Logger
	config  *CadenceConfig
	running bool
}

// NewClient создает новый заглушечный Cadence клиент
func NewClient(config *CadenceConfig, logger *zap.Logger) (*Client, error) {
	if config == nil {
		return nil, fmt.Errorf("конфигурация не может быть nil")
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	return &Client{
		logger: logger,
		config: config,
	}, nil
}

// Start запускает заглушечный клиент
func (c *Client) Start() error {
	c.logger.Info("Запуск заглушечного Cadence клиента")
	c.running = true
	return nil
}

// Stop останавливает заглушечный клиент
func (c *Client) Stop() error {
	c.logger.Info("Остановка заглушечного Cadence клиента")
	c.running = false
	return nil
}

// IsRunning возвращает статус клиента
func (c *Client) IsRunning() bool {
	return c.running
}

// StartWorkflow запускает workflow (заглушка)
func (c *Client) StartWorkflow(ctx context.Context, workflowType string, input interface{}) (*WorkflowExecution, error) {
	c.logger.Info("Запуск workflow (заглушка)",
		zap.String("workflow_type", workflowType),
		zap.Any("input", input),
	)

	return &WorkflowExecution{
		ID:    fmt.Sprintf("workflow_%d", time.Now().Unix()),
		RunID: fmt.Sprintf("run_%d", time.Now().UnixNano()),
	}, nil
}

// GetWorkflowResult получает результат workflow (заглушка)
func (c *Client) GetWorkflowResult(ctx context.Context, workflowID, runID string, result interface{}) error {
	c.logger.Info("Получение результата workflow (заглушка)",
		zap.String("workflow_id", workflowID),
		zap.String("run_id", runID),
	)

	// Симуляция результата
	time.Sleep(100 * time.Millisecond)
	return nil
}

// SignalWorkflow отправляет сигнал workflow (заглушка)
func (c *Client) SignalWorkflow(ctx context.Context, workflowID, runID, signalName string, signalInput interface{}) error {
	c.logger.Info("Отправка сигнала workflow (заглушка)",
		zap.String("workflow_id", workflowID),
		zap.String("run_id", runID),
		zap.String("signal_name", signalName),
	)

	return nil
}

// GetWorker возвращает заглушку worker
func (c *Client) GetWorker() interface{} {
	return &WorkerStub{logger: c.logger}
}

// WorkerStub представляет заглушку для Cadence worker
type WorkerStub struct {
	logger  *zap.Logger
	running bool
}

// Start запускает заглушечный worker
func (w *WorkerStub) Start() error {
	w.logger.Info("Запуск заглушечного Cadence worker")
	w.running = true
	return nil
}

// Stop останавливает заглушечный worker
func (w *WorkerStub) Stop() {
	w.logger.Info("Остановка заглушечного Cadence worker")
	w.running = false
}
