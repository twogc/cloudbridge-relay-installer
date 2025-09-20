package workflows

import (
	"time"

	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// AIOptimizedRoutingWorkflowInput входные параметры для AI-оптимизированной маршрутизации
type AIOptimizedRoutingWorkflowInput struct {
	TenantID          string                 `json:"tenant_id"`
	SourcePeerID      string                 `json:"source_peer_id"`
	TargetPeerID      string                 `json:"target_peer_id"`
	DataSize          int64                  `json:"data_size"`
	Priority          string                 `json:"priority"` // low, medium, high, critical
	NetworkConditions map[string]interface{} `json:"network_conditions"`
}

// AIOptimizedRoutingWorkflowResult результат AI-оптимизированной маршрутизации
type AIOptimizedRoutingWorkflowResult struct {
	TenantID            string                 `json:"tenant_id"`
	SourcePeerID        string                 `json:"source_peer_id"`
	TargetPeerID        string                 `json:"target_peer_id"`
	Status              string                 `json:"status"`
	OptimalRoute        map[string]interface{} `json:"optimal_route"`
	PredictedLatency    time.Duration          `json:"predicted_latency"`
	PredictedThroughput int64                  `json:"predicted_throughput"`
	Confidence          float64                `json:"confidence"`
	AIRecommendations   []string               `json:"ai_recommendations"`
	Result              map[string]interface{} `json:"result"`
	StartedAt           time.Time              `json:"started_at"`
	CompletedAt         time.Time              `json:"completed_at"`
	Message             string                 `json:"message"`
}

// AIOptimizedRoutingWorkflow workflow для AI-оптимизированной маршрутизации
func AIOptimizedRoutingWorkflow(ctx workflow.Context, input AIOptimizedRoutingWorkflowInput) (*AIOptimizedRoutingWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Запуск AIOptimizedRoutingWorkflow",
		zap.String("tenant_id", input.TenantID),
		zap.String("source_peer", input.SourcePeerID),
		zap.String("target_peer", input.TargetPeerID),
		zap.String("priority", input.Priority),
		zap.Int64("data_size", input.DataSize),
	)

	result := &AIOptimizedRoutingWorkflowResult{
		TenantID:     input.TenantID,
		SourcePeerID: input.SourcePeerID,
		TargetPeerID: input.TargetPeerID,
		Status:       "running",
		StartedAt:    workflow.Now(ctx),
		Result:       make(map[string]interface{}),
	}

	// Настройки для activities
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 3 * time.Minute,
		HeartbeatTimeout:    30 * time.Second,
		RetryPolicy: &workflow.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    3,
		},
	}

	// 1. Анализ сетевых условий через AI
	logger.Info("Анализ сетевых условий через AI")
	var networkAnalysisResult map[string]interface{}
	err := workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"AnalyzeNetworkConditionsActivity",
		input.TenantID,
		input.NetworkConditions,
	).Get(ctx, &networkAnalysisResult)

	if err != nil {
		logger.Error("Ошибка анализа сетевых условий", zap.Error(err))
		result.Status = "failed"
		result.Message = "Ошибка анализа сетевых условий: " + err.Error()
		result.CompletedAt = workflow.Now(ctx)
		return result, nil
	}

	// 2. ML-предсказание оптимального маршрута
	logger.Info("ML-предсказание оптимального маршрута")
	var routingPredictionResult map[string]interface{}
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"PredictOptimalRouteActivity",
		input.TenantID,
		input.SourcePeerID,
		input.TargetPeerID,
		input.DataSize,
		input.Priority,
		networkAnalysisResult,
	).Get(ctx, &routingPredictionResult)

	if err != nil {
		logger.Error("Ошибка предсказания маршрута", zap.Error(err))
		result.Status = "failed"
		result.Message = "Ошибка предсказания маршрута: " + err.Error()
		result.CompletedAt = workflow.Now(ctx)
		return result, nil
	}

	// 3. Генерация AI рекомендаций
	logger.Info("Генерация AI рекомендаций")
	var recommendationsResult map[string]interface{}
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"GenerateAIRecommendationsActivity",
		input.TenantID,
		routingPredictionResult,
		networkAnalysisResult,
	).Get(ctx, &recommendationsResult)

	if err != nil {
		logger.Warn("Ошибка генерации рекомендаций", zap.Error(err))
		// Продолжаем даже если рекомендации не удалось сгенерировать
	}

	// Обработка результатов
	if route, ok := routingPredictionResult["route"].(map[string]interface{}); ok {
		result.OptimalRoute = route
	}
	if latency, ok := routingPredictionResult["latency"].(int64); ok {
		result.PredictedLatency = time.Duration(latency) * time.Millisecond
	}
	if throughput, ok := routingPredictionResult["throughput"].(int64); ok {
		result.PredictedThroughput = throughput
	}
	if confidence, ok := routingPredictionResult["confidence"].(float64); ok {
		result.Confidence = confidence
	}
	if recommendations, ok := recommendationsResult["recommendations"].([]interface{}); ok {
		for _, rec := range recommendations {
			if recStr, ok := rec.(string); ok {
				result.AIRecommendations = append(result.AIRecommendations, recStr)
			}
		}
	}

	result.Result = routingPredictionResult
	result.Status = "completed"
	result.CompletedAt = workflow.Now(ctx)
	result.Message = "AI-оптимизированная маршрутизация завершена"

	logger.Info("AIOptimizedRoutingWorkflow завершен",
		zap.String("tenant_id", input.TenantID),
		zap.String("source_peer", input.SourcePeerID),
		zap.String("target_peer", input.TargetPeerID),
		zap.String("status", result.Status),
		zap.Duration("predicted_latency", result.PredictedLatency),
		zap.Int64("predicted_throughput", result.PredictedThroughput),
		zap.Float64("confidence", result.Confidence),
		zap.Int("recommendations_count", len(result.AIRecommendations)),
		zap.Duration("duration", result.CompletedAt.Sub(result.StartedAt)),
	)

	return result, nil
}
