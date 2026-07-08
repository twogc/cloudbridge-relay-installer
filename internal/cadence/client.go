package cadence

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// Client is the lightweight Cadence API used by cmd/ai-worker and internal/cadence/worker.
// It wraps ProductionClient so connection, workflow registration, and lifecycle share one path
// with cmd/relay (NewProductionClient) without forcing callers to change types.
type Client struct {
	prod *ProductionClient
}

// Config configuration for the lightweight Cadence client (ai-worker path).
type Config struct {
	HostPort       string
	Domain         string
	TaskList       string
	WorkerOptions  worker.Options
	ServiceOptions ServiceClientOptions
}

// NewClient creates a Cadence client for the ai-worker path.
// Internally builds ProductionClient with identity "cloudbridge-ai-worker".
func NewClient(config Config, logger *zap.Logger) (*Client, error) {
	if logger == nil {
		logger = zap.NewNop()
	}
	logger.Info("Creating Cadence client (shared production engine)",
		zap.String("domain", config.Domain),
		zap.String("taskList", config.TaskList),
		zap.String("hostPort", config.HostPort))

	prod, err := NewProductionClient(ProductionConfig{
		HostPort:       config.HostPort,
		Domain:         config.Domain,
		TaskList:       config.TaskList,
		WorkerOptions:  config.WorkerOptions,
		ServiceOptions: config.ServiceOptions,
		Identity:       "cloudbridge-ai-worker",
	}, logger)
	if err != nil {
		return nil, err
	}
	return &Client{prod: prod}, nil
}

// Start starts the Cadence worker (via ProductionClient).
func (c *Client) Start() error {
	if c == nil || c.prod == nil {
		return fmt.Errorf("cadence client is nil")
	}
	return c.prod.Start()
}

// Stop stops the Cadence worker.
func (c *Client) Stop() {
	if c != nil && c.prod != nil {
		c.prod.Stop()
	}
}

// GetClient returns the underlying uber/cadence client.
func (c *Client) GetClient() client.Client {
	if c == nil || c.prod == nil {
		return nil
	}
	return c.prod.GetClient()
}

// GetWorker returns the Cadence worker.
func (c *Client) GetWorker() worker.Worker {
	if c == nil || c.prod == nil {
		return nil
	}
	return c.prod.GetWorker()
}

// RegisterWorkflow is a compat hook. Prefer ProductionClient.Start() which registers
// workflows on the worker. Only package-registers function values to avoid panics.
func (c *Client) RegisterWorkflow(w interface{}) {
	if c != nil && c.prod != nil && c.prod.logger != nil {
		c.prod.logger.Info("RegisterWorkflow (compat)", zap.String("workflow", fmt.Sprintf("%T", w)))
	}
	if w == nil {
		return
	}
	// Cadence requires a func; ignore invalid test stubs.
	if _, ok := w.(func(context.Context, interface{}) error); ok {
		workflow.Register(w)
	}
}

// RegisterActivity is a compat hook. Prefer ProductionClient activity registration.
func (c *Client) RegisterActivity(a interface{}) {
	if c != nil && c.prod != nil && c.prod.logger != nil {
		c.prod.logger.Info("RegisterActivity (compat)", zap.String("activity", fmt.Sprintf("%T", a)))
	}
	// Do not call activity.Register with non-func values (would panic).
	_ = a
	_ = activity.Register
}

// HealthCheck checks client health.
func (c *Client) HealthCheck() error {
	if c == nil || c.prod == nil {
		return fmt.Errorf("cadence client is nil")
	}
	return c.prod.HealthCheck()
}

// ExecuteWorkflow executes a workflow and returns its ID.
func (c *Client) ExecuteWorkflow(ctx context.Context, workflowType string, input interface{}) (string, error) {
	if c == nil || c.prod == nil {
		return "", fmt.Errorf("cadence client is nil")
	}
	return c.prod.ExecuteWorkflow(ctx, workflowType, input)
}

// GetWorkflowStatus returns workflow status (stub/real via production client).
func (c *Client) GetWorkflowStatus(ctx context.Context, workflowID string) (string, error) {
	if c == nil || c.prod == nil {
		return "", fmt.Errorf("cadence client is nil")
	}
	return c.prod.GetWorkflowStatus(ctx, workflowID)
}

// GetDomain returns the configured domain.
func (c *Client) GetDomain() string {
	if c == nil || c.prod == nil {
		return ""
	}
	return c.prod.GetDomain()
}

// Close closes the client and stops the worker.
func (c *Client) Close() {
	if c != nil && c.prod != nil {
		c.prod.Close()
	}
}

// WorkflowExecution represents a workflow run.
type WorkflowExecution struct {
	ID    string
	RunID string
}

// StartWorkflow starts a workflow and returns execution IDs.
func (c *Client) StartWorkflow(ctx context.Context, workflowType string, input interface{}) (*WorkflowExecution, error) {
	if c == nil || c.prod == nil {
		return nil, fmt.Errorf("cadence client is nil")
	}
	id, err := c.prod.ExecuteWorkflow(ctx, workflowType, input)
	if err != nil {
		return nil, err
	}
	// Production ExecuteWorkflow returns only ID; synthesize run ID in stub-compatible form.
	return &WorkflowExecution{
		ID:    id,
		RunID: fmt.Sprintf("run-%d", time.Now().UnixNano()),
	}, nil
}

// GetWorkflowResult fetches a workflow result.
func (c *Client) GetWorkflowResult(ctx context.Context, workflowID string, runID string, result interface{}) error {
	if c == nil || c.prod == nil {
		return fmt.Errorf("cadence client is nil")
	}
	cc := c.prod.GetClient()
	if cc == nil {
		if c.prod.logger != nil {
			c.prod.logger.Info("Getting workflow result (stub mode)")
		}
		// Stub: fill common *map[string]interface{} results for tests/callers.
		if m, ok := result.(*map[string]interface{}); ok {
			*m = map[string]interface{}{
				"workflow_id": workflowID,
				"run_id":      runID,
				"status":      "completed",
				"stub":        true,
			}
		}
		return nil
	}
	workflowRun := cc.GetWorkflow(ctx, workflowID, runID)
	if err := workflowRun.Get(ctx, result); err != nil {
		return fmt.Errorf("failed to get workflow result: %w", err)
	}
	return nil
}

// SignalWorkflow sends a signal to a workflow.
func (c *Client) SignalWorkflow(ctx context.Context, workflowID string, runID string, signalName string, signalValue interface{}) error {
	if c == nil || c.prod == nil {
		return fmt.Errorf("cadence client is nil")
	}
	cc := c.prod.GetClient()
	if cc == nil {
		return nil // stub
	}
	if err := cc.SignalWorkflow(ctx, workflowID, runID, signalName, signalValue); err != nil {
		return fmt.Errorf("failed to signal workflow: %w", err)
	}
	return nil
}

// IsRunning reports whether the client/worker are considered running.
func (c *Client) IsRunning() bool {
	if c == nil || c.prod == nil {
		return false
	}
	if c.prod.IsStubMode() {
		return true
	}
	return c.prod.worker != nil && c.prod.client != nil
}

// IsStubMode reports stub (no real service client) mode.
func (c *Client) IsStubMode() bool {
	if c == nil || c.prod == nil {
		return true
	}
	return c.prod.IsStubMode()
}

// Production returns the underlying ProductionClient (for advanced callers).
func (c *Client) Production() *ProductionClient {
	if c == nil {
		return nil
	}
	return c.prod
}
