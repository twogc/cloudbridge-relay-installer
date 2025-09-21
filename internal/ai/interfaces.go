package ai

import (
	"context"

	"github.com/twogc/cloudbridge-relay-installer/internal/ai/anomaly"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/predictive"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/routing"
)

// AIClientInterface defines the interface for AI client operations
type AIClientInterface interface {
	GetStatus(ctx context.Context) (*ClientServiceStatus, error)
	DetectAnomaly(ctx context.Context, metrics *anomaly.NetworkMetrics) (*anomaly.AnomalyResult, error)
	PredictLoad(ctx context.Context, horizon int) (*predictive.LoadPrediction, error)
	OptimizeRoute(ctx context.Context, source, destination string, constraints map[string]interface{}) (*routing.OptimizedRoute, error)
	HealthCheck(ctx context.Context) error
}
