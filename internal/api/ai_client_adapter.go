package api

import (
	"context"
	"fmt"

	"github.com/twogc/cloudbridge-relay-installer/internal/ai"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/anomaly"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/predictive"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/routing"
	"go.uber.org/zap"
)

// AIClientAdapter адаптер для ai.Client, реализующий AIServiceInterface
type AIClientAdapter struct {
	client *ai.Client
	logger *zap.Logger
}

// NewAIClientAdapter создает новый адаптер для AI клиента
func NewAIClientAdapter(client *ai.Client, logger *zap.Logger) *AIClientAdapter {
	return &AIClientAdapter{
		client: client,
		logger: logger,
	}
}

// GetStatus возвращает статус AI сервиса
func (a *AIClientAdapter) GetStatus() *ai.ServiceStatus {
	ctx := context.Background()

	// Получаем статус от AI клиента
	clientStatus, err := a.client.GetStatus(ctx)
	if err != nil {
		a.logger.Error("Failed to get AI service status", zap.Error(err))
		// Возвращаем статус с ошибкой
		return &ai.ServiceStatus{
			Status: "error",
			Components: map[string]bool{
				"anomaly_detection":  false,
				"load_prediction":    false,
				"route_optimization": false,
			},
		}
	}

	// Преобразуем ClientServiceStatus в ServiceStatus
	status := "running"
	components := map[string]bool{
		"anomaly_detection":  clientStatus.AnomalyDetector,
		"load_prediction":    clientStatus.LoadPredictor,
		"route_optimization": clientStatus.RouteOptimizer,
		"data_pipeline":      clientStatus.DataPipeline,
		"monitoring":         clientStatus.Monitoring,
	}

	// Проверяем, все ли компоненты работают
	allHealthy := true
	for _, healthy := range components {
		if !healthy {
			allHealthy = false
			break
		}
	}

	if !allHealthy {
		status = "degraded"
	}

	return &ai.ServiceStatus{
		Status:     status,
		Components: components,
	}
}

// DetectAnomaly обнаруживает аномалии в сетевых метриках
func (a *AIClientAdapter) DetectAnomaly(ctx context.Context, metrics *anomaly.NetworkMetrics) (*anomaly.AnomalyResult, error) {
	a.logger.Debug("Detecting anomaly",
		zap.Int64("bytes_per_second", metrics.BytesPerSecond),
		zap.Int("connections_per_second", metrics.ConnectionsPerSecond),
		zap.Float64("packet_loss_rate", metrics.PacketLossRate),
	)

	result, err := a.client.DetectAnomaly(ctx, metrics)
	if err != nil {
		a.logger.Error("Failed to detect anomaly", zap.Error(err))
		return nil, fmt.Errorf("anomaly detection failed: %w", err)
	}

	a.logger.Debug("Anomaly detection completed",
		zap.Bool("is_anomaly", result.IsAnomaly),
		zap.Float64("score", result.Score),
		zap.String("anomaly_type", result.AnomalyType),
	)

	return result, nil
}

// PredictLoad предсказывает нагрузку на заданный горизонт
func (a *AIClientAdapter) PredictLoad(ctx context.Context, horizon int) (*predictive.LoadPrediction, error) {
	a.logger.Debug("Predicting load", zap.Int("horizon", horizon))

	result, err := a.client.PredictLoad(ctx, horizon)
	if err != nil {
		a.logger.Error("Failed to predict load", zap.Error(err))
		return nil, fmt.Errorf("load prediction failed: %w", err)
	}

	a.logger.Debug("Load prediction completed",
		zap.Float64("confidence", result.Confidence),
		zap.Int("forecast_length", len(result.Forecast)),
	)

	return result, nil
}

// OptimizeRoute оптимизирует маршрут между источником и назначением
func (a *AIClientAdapter) OptimizeRoute(ctx context.Context, source, destination string, constraints map[string]interface{}) (*routing.OptimizedRoute, error) {
	a.logger.Debug("Optimizing route",
		zap.String("source", source),
		zap.String("destination", destination),
		zap.Any("constraints", constraints),
	)

	result, err := a.client.OptimizeRoute(ctx, source, destination, constraints)
	if err != nil {
		a.logger.Error("Failed to optimize route", zap.Error(err))
		return nil, fmt.Errorf("route optimization failed: %w", err)
	}

	a.logger.Debug("Route optimization completed",
		zap.Float64("predicted_latency", result.PredictedLatency),
		zap.Float64("confidence", result.Confidence),
		zap.String("protocol", result.Protocol),
	)

	return result, nil
}

// HealthCheck проверяет здоровье AI сервиса
func (a *AIClientAdapter) HealthCheck(ctx context.Context) error {
	return a.client.HealthCheck(ctx)
}
