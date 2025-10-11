package workflows

import (
	"time"

	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// P2PQUICTunnelWorkflowInput входные параметры для управления P2P QUIC туннелями
type P2PQUICTunnelWorkflowInput struct {
	TenantID      string                 `json:"tenant_id"`
	PeerID        string                 `json:"peer_id"`
	Operation     string                 `json:"operation"` // create, update, delete, monitor
	TunnelConfig  map[string]interface{} `json:"tunnel_config"`
	NetworkConfig map[string]interface{} `json:"network_config"`
}

// P2PQUICTunnelWorkflowResult результат управления P2P QUIC туннелями
type P2PQUICTunnelWorkflowResult struct {
	TenantID    string                 `json:"tenant_id"`
	PeerID      string                 `json:"peer_id"`
	Operation   string                 `json:"operation"`
	Status      string                 `json:"status"`
	TunnelID    string                 `json:"tunnel_id"`
	Result      map[string]interface{} `json:"result"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt time.Time              `json:"completed_at"`
	Message     string                 `json:"message"`
}

// P2PQUICTunnelWorkflow workflow для управления P2P QUIC туннелями
func P2PQUICTunnelWorkflow(ctx workflow.Context, input P2PQUICTunnelWorkflowInput) (*P2PQUICTunnelWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Запуск P2PQUICTunnelWorkflow",
		zap.String("tenant_id", input.TenantID),
		zap.String("peer_id", input.PeerID),
		zap.String("operation", input.Operation),
	)

	// Настройка ActivityOptions для всех activities в этом workflow
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    5 * time.Minute,
		RetryPolicy: &workflow.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	result := &P2PQUICTunnelWorkflowResult{
		TenantID:  input.TenantID,
		PeerID:    input.PeerID,
		Operation: input.Operation,
		Status:    "running",
		StartedAt: workflow.Now(ctx),
		Result:    make(map[string]interface{}),
	}

	// Настройки для activities
	activityOptions := workflow.ActivityOptions{
		ScheduleToStartTimeout: 1 * time.Minute,
		StartToCloseTimeout:    5 * time.Minute,
		HeartbeatTimeout:       30 * time.Second,
		RetryPolicy: &workflow.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    3,
		},
	}

	switch input.Operation {
	case "create":
		// Создание P2P QUIC туннеля
		logger.Info("Создание P2P QUIC туннеля",
			zap.String("tenant_id", input.TenantID),
			zap.String("peer_id", input.PeerID),
		)

		var tunnelResult map[string]interface{}
		err := workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, activityOptions),
			"CreateP2PQUICTunnelActivity",
			input.TenantID,
			input.PeerID,
			input.TunnelConfig,
		).Get(ctx, &tunnelResult)

		if err != nil {
			logger.Error("Ошибка создания туннеля", zap.Error(err))
			result.Status = "failed"
			result.Message = "Ошибка создания туннеля: " + err.Error()
			result.CompletedAt = workflow.Now(ctx)
			return result, err
		}

		result.TunnelID = tunnelResult["tunnel_id"].(string)
		result.Result = tunnelResult

	case "update":
		// Обновление P2P QUIC туннеля
		logger.Info("Обновление P2P QUIC туннеля",
			zap.String("tenant_id", input.TenantID),
			zap.String("peer_id", input.PeerID),
		)

		var updateResult map[string]interface{}
		err := workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, activityOptions),
			"UpdateP2PQUICTunnelActivity",
			input.TenantID,
			input.PeerID,
			input.TunnelConfig,
		).Get(ctx, &updateResult)

		if err != nil {
			logger.Error("Ошибка обновления туннеля", zap.Error(err))
			result.Status = "failed"
			result.Message = "Ошибка обновления туннеля: " + err.Error()
			result.CompletedAt = workflow.Now(ctx)
			return result, err
		}

		result.Result = updateResult

	case "delete":
		// Удаление P2P QUIC туннеля
		logger.Info("Удаление P2P QUIC туннеля",
			zap.String("tenant_id", input.TenantID),
			zap.String("peer_id", input.PeerID),
		)

		var deleteResult map[string]interface{}
		err := workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, activityOptions),
			"DeleteP2PQUICTunnelActivity",
			input.TenantID,
			input.PeerID,
		).Get(ctx, &deleteResult)

		if err != nil {
			logger.Error("Ошибка удаления туннеля", zap.Error(err))
			result.Status = "failed"
			result.Message = "Ошибка удаления туннеля: " + err.Error()
			result.CompletedAt = workflow.Now(ctx)
			return result, err
		}

		result.Result = deleteResult

	case "monitor":
		// Мониторинг P2P QUIC туннеля
		logger.Info("Мониторинг P2P QUIC туннеля",
			zap.String("tenant_id", input.TenantID),
			zap.String("peer_id", input.PeerID),
		)

		// Долгоживущий workflow для мониторинга
		for {
			var monitorResult map[string]interface{}
			err := workflow.ExecuteActivity(
				workflow.WithActivityOptions(ctx, activityOptions),
				"MonitorP2PQUICTunnelActivity",
				input.TenantID,
				input.PeerID,
			).Get(ctx, &monitorResult)

			if err != nil {
				logger.Warn("Ошибка мониторинга туннеля", zap.Error(err))
			} else {
				result.Result = monitorResult
				logger.Debug("Результат мониторинга туннеля",
					zap.String("tenant_id", input.TenantID),
					zap.String("peer_id", input.PeerID),
					zap.Any("result", monitorResult),
				)
			}

			// Ждем 30 секунд до следующей проверки
			err = workflow.Sleep(ctx, 30*time.Second)
			if err != nil {
				// Workflow отменен
				logger.Info("Мониторинг туннеля остановлен")
				break
			}
		}

		result.Status = "completed"
		result.Message = "Мониторинг туннеля завершен"

	default:
		logger.Error("Неизвестная операция", zap.String("operation", input.Operation))
		result.Status = "failed"
		result.Message = "Неизвестная операция: " + input.Operation
		result.CompletedAt = workflow.Now(ctx)
		return result, nil
	}

	if result.Status != "completed" {
		result.Status = "completed"
		result.Message = "P2P QUIC туннель успешно обработан"
	}
	result.CompletedAt = workflow.Now(ctx)

	logger.Info("P2PQUICTunnelWorkflow завершен",
		zap.String("tenant_id", input.TenantID),
		zap.String("peer_id", input.PeerID),
		zap.String("operation", input.Operation),
		zap.String("status", result.Status),
		zap.Duration("duration", result.CompletedAt.Sub(result.StartedAt)),
	)

	return result, nil
}
