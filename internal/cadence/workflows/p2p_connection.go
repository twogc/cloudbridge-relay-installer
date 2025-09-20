package workflows

import (
	"time"

	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// P2PConnectionOrchestrationWorkflowInput входные параметры для оркестрации P2P соединений
type P2PConnectionOrchestrationWorkflowInput struct {
	TenantID          string                 `json:"tenant_id"`
	SourcePeerID      string                 `json:"source_peer_id"`
	TargetPeerID      string                 `json:"target_peer_id"`
	ConnectionType    string                 `json:"connection_type"` // direct, relay, masque
	Protocol          string                 `json:"protocol"`        // quic, websocket, tcp
	Configuration     map[string]interface{} `json:"configuration"`
}

// P2PConnectionOrchestrationWorkflowResult результат оркестрации P2P соединений
type P2PConnectionOrchestrationWorkflowResult struct {
	TenantID          string                 `json:"tenant_id"`
	SourcePeerID      string                 `json:"source_peer_id"`
	TargetPeerID      string                 `json:"target_peer_id"`
	ConnectionID      string                 `json:"connection_id"`
	Status            string                 `json:"status"`
	ConnectionType    string                 `json:"connection_type"`
	Protocol          string                 `json:"protocol"`
	Latency           time.Duration          `json:"latency"`
	Throughput        int64                  `json:"throughput"`
	Result            map[string]interface{} `json:"result"`
	StartedAt         time.Time              `json:"started_at"`
	CompletedAt       time.Time              `json:"completed_at"`
	Message           string                 `json:"message"`
}

// P2PConnectionOrchestrationWorkflow workflow для оркестрации P2P соединений
func P2PConnectionOrchestrationWorkflow(ctx workflow.Context, input P2PConnectionOrchestrationWorkflowInput) (*P2PConnectionOrchestrationWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Запуск P2PConnectionOrchestrationWorkflow",
		zap.String("tenant_id", input.TenantID),
		zap.String("source_peer", input.SourcePeerID),
		zap.String("target_peer", input.TargetPeerID),
		zap.String("connection_type", input.ConnectionType),
		zap.String("protocol", input.Protocol),
	)

	result := &P2PConnectionOrchestrationWorkflowResult{
		TenantID:       input.TenantID,
		SourcePeerID:   input.SourcePeerID,
		TargetPeerID:   input.TargetPeerID,
		ConnectionType: input.ConnectionType,
		Protocol:       input.Protocol,
		Status:         "running",
		StartedAt:      workflow.Now(ctx),
		Result:         make(map[string]interface{}),
	}

	// Настройки для activities
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		HeartbeatTimeout:    30 * time.Second,
		RetryPolicy: &workflow.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    3,
		},
	}

	// 1. Проверка доступности пиров
	logger.Info("Проверка доступности пиров")
	var peerCheckResult map[string]interface{}
	err := workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"CheckPeersAvailabilityActivity",
		input.TenantID,
		[]string{input.SourcePeerID, input.TargetPeerID},
	).Get(ctx, &peerCheckResult)

	if err != nil {
		logger.Error("Ошибка проверки доступности пиров", zap.Error(err))
		result.Status = "failed"
		result.Message = "Ошибка проверки доступности пиров: " + err.Error()
		result.CompletedAt = workflow.Now(ctx)
		return result, err
	}

	// Проверяем, что оба пира доступны
	if available, ok := peerCheckResult["available"].(bool); !ok || !available {
		logger.Error("Один или оба пира недоступны", zap.Any("result", peerCheckResult))
		result.Status = "failed"
		result.Message = "Один или оба пира недоступны"
		result.CompletedAt = workflow.Now(ctx)
		return result, nil
	}

	// 2. Выбор оптимального типа соединения
	logger.Info("Выбор оптимального типа соединения")
	var connectionTypeResult map[string]interface{}
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"SelectOptimalConnectionTypeActivity",
		input.TenantID,
		input.SourcePeerID,
		input.TargetPeerID,
		input.Configuration,
	).Get(ctx, &connectionTypeResult)

	if err != nil {
		logger.Error("Ошибка выбора типа соединения", zap.Error(err))
		result.Status = "failed"
		result.Message = "Ошибка выбора типа соединения: " + err.Error()
		result.CompletedAt = workflow.Now(ctx)
		return result, err
	}

	// Обновляем тип соединения на основе рекомендации
	if recommendedType, ok := connectionTypeResult["recommended_type"].(string); ok {
		result.ConnectionType = recommendedType
		logger.Info("Тип соединения обновлен на основе рекомендации", 
			zap.String("recommended_type", recommendedType),
		)
	}

	// 3. Установка P2P соединения
	logger.Info("Установка P2P соединения")
	var connectionResult map[string]interface{}
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"EstablishP2PConnectionActivity",
		input.TenantID,
		input.SourcePeerID,
		input.TargetPeerID,
		result.ConnectionType,
		input.Protocol,
	).Get(ctx, &connectionResult)

	if err != nil {
		logger.Error("Ошибка установки соединения", zap.Error(err))
		result.Status = "failed"
		result.Message = "Ошибка установки соединения: " + err.Error()
		result.CompletedAt = workflow.Now(ctx)
		return result, err
	}

	// 4. Валидация соединения
	logger.Info("Валидация P2P соединения")
	var validationResult map[string]interface{}
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"ValidateP2PConnectionActivity",
		input.TenantID,
		connectionResult["connection_id"].(string),
	).Get(ctx, &validationResult)

	if err != nil {
		logger.Warn("Ошибка валидации соединения", zap.Error(err))
		// Продолжаем даже если валидация не удалась
	}

	// Обработка результатов
	result.ConnectionID = connectionResult["connection_id"].(string)
	if latency, ok := connectionResult["latency"].(int64); ok {
		result.Latency = time.Duration(latency) * time.Millisecond
	}
	if throughput, ok := connectionResult["throughput"].(int64); ok {
		result.Throughput = throughput
	}
	result.Result = connectionResult

	result.Status = "completed"
	result.CompletedAt = workflow.Now(ctx)
	result.Message = "P2P соединение успешно установлено"

	logger.Info("P2PConnectionOrchestrationWorkflow завершен",
		zap.String("tenant_id", input.TenantID),
		zap.String("source_peer", input.SourcePeerID),
		zap.String("target_peer", input.TargetPeerID),
		zap.String("connection_id", result.ConnectionID),
		zap.String("status", result.Status),
		zap.Duration("latency", result.Latency),
		zap.Int64("throughput", result.Throughput),
		zap.Duration("duration", result.CompletedAt.Sub(result.StartedAt)),
	)

	return result, nil
}
