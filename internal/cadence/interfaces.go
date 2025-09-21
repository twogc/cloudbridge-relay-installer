package cadence

import (
	"context"
)

// CadenceClientInterface defines the interface for Cadence client operations
type CadenceClientInterface interface {
	ExecuteWorkflow(ctx context.Context, workflowID string, input interface{}) (interface{}, error)
	GetWorkflowStatus(ctx context.Context, workflowID string) (map[string]interface{}, error)
	CancelWorkflow(ctx context.Context, workflowID string) error
	GetWorkflowHistory(ctx context.Context, workflowID string) ([]map[string]interface{}, error)
}
