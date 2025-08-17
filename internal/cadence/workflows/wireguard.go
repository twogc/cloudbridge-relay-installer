package workflows

import (
	"fmt"
	"time"

	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// PeerLifecycleWorkflowInput входные параметры для управления жизненным циклом пиров
type PeerLifecycleWorkflowInput struct {
	PeerID        string                 `json:"peer_id"`
	Operation     string                 `json:"operation"` // add, remove, update, monitor
	PeerConfig    map[string]interface{} `json:"peer_config"`
	NetworkConfig map[string]interface{} `json:"network_config"`
}

// PeerLifecycleWorkflowResult результат управления жизненным циклом пиров
type PeerLifecycleWorkflowResult struct {
	PeerID      string                 `json:"peer_id"`
	Status      string                 `json:"status"`
	Operation   string                 `json:"operation"`
	Result      map[string]interface{} `json:"result"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt time.Time              `json:"completed_at"`
	Message     string                 `json:"message"`
}

// MeshTopologyWorkflowInput входные параметры для управления mesh топологией
type MeshTopologyWorkflowInput struct {
	NetworkID     string                 `json:"network_id"`
	Peers         []string               `json:"peers"`
	TopologyType  string                 `json:"topology_type"` // full_mesh, hub_spoke, hybrid
	Configuration map[string]interface{} `json:"configuration"`
}

// MeshTopologyWorkflowResult результат управления mesh топологией
type MeshTopologyWorkflowResult struct {
	NetworkID      string                 `json:"network_id"`
	Status         string                 `json:"status"`
	TopologyType   string                 `json:"topology_type"`
	ConnectedPeers int                    `json:"connected_peers"`
	FailedPeers    []string               `json:"failed_peers"`
	Topology       map[string]interface{} `json:"topology"`
	StartedAt      time.Time              `json:"started_at"`
	CompletedAt    time.Time              `json:"completed_at"`
	Message        string                 `json:"message"`
}

// ConnectionRecoveryWorkflowInput входные параметры для восстановления соединений
type ConnectionRecoveryWorkflowInput struct {
	FailedConnections []string               `json:"failed_connections"`
	RecoveryStrategy  string                 `json:"recovery_strategy"` // immediate, gradual, scheduled
	MaxRetries        int                    `json:"max_retries"`
	Configuration     map[string]interface{} `json:"configuration"`
}

// ConnectionRecoveryWorkflowResult результат восстановления соединений
type ConnectionRecoveryWorkflowResult struct {
	Status               string                 `json:"status"`
	RecoveredConnections []string               `json:"recovered_connections"`
	FailedConnections    []string               `json:"failed_connections"`
	RecoveryStrategy     string                 `json:"recovery_strategy"`
	Attempts             int                    `json:"attempts"`
	Results              map[string]interface{} `json:"results"`
	StartedAt            time.Time              `json:"started_at"`
	CompletedAt          time.Time              `json:"completed_at"`
	Message              string                 `json:"message"`
}

// LoadBalancingWorkflowInput входные параметры для балансировки нагрузки
type LoadBalancingWorkflowInput struct {
	LoadBalancerID string                 `json:"load_balancer_id"`
	Peers          []string               `json:"peers"`
	Strategy       string                 `json:"strategy"` // round_robin, least_connections, weighted, geographic
	Configuration  map[string]interface{} `json:"configuration"`
}

// LoadBalancingWorkflowResult результат балансировки нагрузки
type LoadBalancingWorkflowResult struct {
	LoadBalancerID   string                 `json:"load_balancer_id"`
	Status           string                 `json:"status"`
	Strategy         string                 `json:"strategy"`
	ActivePeers      []string               `json:"active_peers"`
	LoadDistribution map[string]float64     `json:"load_distribution"`
	Configuration    map[string]interface{} `json:"configuration"`
	StartedAt        time.Time              `json:"started_at"`
	CompletedAt      time.Time              `json:"completed_at"`
	Message          string                 `json:"message"`
}

// PeerLifecycleWorkflow workflow для управления жизненным циклом WireGuard пиров
func PeerLifecycleWorkflow(ctx workflow.Context, input PeerLifecycleWorkflowInput) (*PeerLifecycleWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Запуск PeerLifecycleWorkflow",
		zap.String("peer_id", input.PeerID),
		zap.String("operation", input.Operation),
	)

	result := &PeerLifecycleWorkflowResult{
		PeerID:    input.PeerID,
		Operation: input.Operation,
		Status:    "running",
		StartedAt: workflow.Now(ctx),
		Result:    make(map[string]interface{}),
	}

	// Настройки для activities
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		HeartbeatTimeout:    30 * time.Second,
		RetryPolicy: &workflow.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    3,
		},
	}

	switch input.Operation {
	case "add":
		// Добавление нового пира
		logger.Info("Добавление нового пира", zap.String("peer_id", input.PeerID))

		// 1. Валидация конфигурации пира
		var validationResult map[string]interface{}
		err := workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, activityOptions),
			"ValidatePeerConfigActivity",
			input.PeerConfig,
		).Get(ctx, &validationResult)

		if err != nil {
			logger.Error("Ошибка валидации конфигурации пира", zap.Error(err))
			result.Status = "failed"
			result.Message = "Валидация конфигурации не прошла: " + err.Error()
			result.CompletedAt = workflow.Now(ctx)
			return result, err
		}

		// 2. Создание WireGuard интерфейса для пира
		var interfaceResult map[string]interface{}
		err = workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, activityOptions),
			"CreateWireGuardInterfaceActivity",
			input.PeerID,
			input.PeerConfig,
		).Get(ctx, &interfaceResult)

		if err != nil {
			logger.Error("Ошибка создания WireGuard интерфейса", zap.Error(err))
			result.Status = "failed"
			result.Message = "Не удалось создать WireGuard интерфейс: " + err.Error()
			result.CompletedAt = workflow.Now(ctx)
			return result, err
		}

		// 3. Регистрация пира в системе
		var registrationResult map[string]interface{}
		err = workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, activityOptions),
			"RegisterPeerActivity",
			input.PeerID,
			interfaceResult,
		).Get(ctx, &registrationResult)

		if err != nil {
			logger.Error("Ошибка регистрации пира", zap.Error(err))
			result.Status = "failed"
			result.Message = "Не удалось зарегистрировать пира: " + err.Error()
			result.CompletedAt = workflow.Now(ctx)
			return result, err
		}

		result.Result = registrationResult
		result.Status = "completed"
		result.Message = "Пир успешно добавлен"

	case "remove":
		// Удаление пира
		logger.Info("Удаление пира", zap.String("peer_id", input.PeerID))

		// 1. Отключение пира от сети
		var disconnectResult map[string]interface{}
		err := workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, activityOptions),
			"DisconnectPeerActivity",
			input.PeerID,
		).Get(ctx, &disconnectResult)

		if err != nil {
			logger.Warn("Ошибка отключения пира", zap.Error(err))
			// Продолжаем удаление даже если отключение не удалось
		}

		// 2. Удаление WireGuard интерфейса
		var removeResult map[string]interface{}
		err = workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, activityOptions),
			"RemoveWireGuardInterfaceActivity",
			input.PeerID,
		).Get(ctx, &removeResult)

		if err != nil {
			logger.Error("Ошибка удаления WireGuard интерфейса", zap.Error(err))
			result.Status = "failed"
			result.Message = "Не удалось удалить WireGuard интерфейс: " + err.Error()
			result.CompletedAt = workflow.Now(ctx)
			return result, err
		}

		// 3. Удаление пира из системы
		var unregisterResult map[string]interface{}
		err = workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, activityOptions),
			"UnregisterPeerActivity",
			input.PeerID,
		).Get(ctx, &unregisterResult)

		if err != nil {
			logger.Error("Ошибка удаления пира из системы", zap.Error(err))
			result.Status = "failed"
			result.Message = "Не удалось удалить пира из системы: " + err.Error()
			result.CompletedAt = workflow.Now(ctx)
			return result, err
		}

		result.Result = unregisterResult
		result.Status = "completed"
		result.Message = "Пир успешно удален"

	case "update":
		// Обновление конфигурации пира
		logger.Info("Обновление конфигурации пира", zap.String("peer_id", input.PeerID))

		// 1. Валидация новой конфигурации
		var validationResult map[string]interface{}
		err := workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, activityOptions),
			"ValidatePeerConfigActivity",
			input.PeerConfig,
		).Get(ctx, &validationResult)

		if err != nil {
			logger.Error("Ошибка валидации новой конфигурации", zap.Error(err))
			result.Status = "failed"
			result.Message = "Валидация новой конфигурации не прошла: " + err.Error()
			result.CompletedAt = workflow.Now(ctx)
			return result, err
		}

		// 2. Обновление конфигурации WireGuard
		var updateResult map[string]interface{}
		err = workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, activityOptions),
			"UpdateWireGuardConfigActivity",
			input.PeerID,
			input.PeerConfig,
		).Get(ctx, &updateResult)

		if err != nil {
			logger.Error("Ошибка обновления WireGuard конфигурации", zap.Error(err))
			result.Status = "failed"
			result.Message = "Не удалось обновить WireGuard конфигурацию: " + err.Error()
			result.CompletedAt = workflow.Now(ctx)
			return result, err
		}

		result.Result = updateResult
		result.Status = "completed"
		result.Message = "Конфигурация пира успешно обновлена"

	case "monitor":
		// Мониторинг пира (долгоживущий workflow)
		logger.Info("Запуск мониторинга пира", zap.String("peer_id", input.PeerID))

		// Основной цикл мониторинга
		for {
			// Проверка состояния пира
			var monitorResult map[string]interface{}
			err := workflow.ExecuteActivity(
				workflow.WithActivityOptions(ctx, activityOptions),
				"MonitorPeerActivity",
				input.PeerID,
			).Get(ctx, &monitorResult)

			if err != nil {
				logger.Warn("Ошибка мониторинга пира", zap.Error(err))
			} else {
				result.Result = monitorResult
				logger.Debug("Результат мониторинга пира",
					zap.String("peer_id", input.PeerID),
					zap.Any("result", monitorResult),
				)
			}

			// Ждем 1 минуту до следующей проверки
			err = workflow.Sleep(ctx, 1*time.Minute)
			if err != nil {
				// Workflow отменен
				logger.Info("Мониторинг пира остановлен")
				break
			}
		}

		result.Status = "completed"
		result.Message = "Мониторинг пира завершен"

	default:
		logger.Error("Неизвестная операция", zap.String("operation", input.Operation))
		result.Status = "failed"
		result.Message = "Неизвестная операция: " + input.Operation
		result.CompletedAt = workflow.Now(ctx)
		return result, fmt.Errorf("unknown_operation: %s", input.Operation)
	}

	result.CompletedAt = workflow.Now(ctx)

	logger.Info("PeerLifecycleWorkflow завершен",
		zap.String("peer_id", input.PeerID),
		zap.String("operation", input.Operation),
		zap.String("status", result.Status),
		zap.Duration("duration", result.CompletedAt.Sub(result.StartedAt)),
	)

	return result, nil
}

// MeshTopologyWorkflow workflow для управления mesh топологией
func MeshTopologyWorkflow(ctx workflow.Context, input MeshTopologyWorkflowInput) (*MeshTopologyWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Запуск MeshTopologyWorkflow",
		zap.String("network_id", input.NetworkID),
		zap.String("topology_type", input.TopologyType),
		zap.Int("peers_count", len(input.Peers)),
	)

	result := &MeshTopologyWorkflowResult{
		NetworkID:    input.NetworkID,
		TopologyType: input.TopologyType,
		Status:       "building",
		StartedAt:    workflow.Now(ctx),
		FailedPeers:  make([]string, 0),
		Topology:     make(map[string]interface{}),
	}

	// Настройки для activities
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 15 * time.Minute,
		HeartbeatTimeout:    1 * time.Minute,
		RetryPolicy: &workflow.RetryPolicy{
			InitialInterval:    2 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    1 * time.Minute,
			MaximumAttempts:    3,
		},
	}

	// 1. Анализ существующей топологии
	logger.Info("Анализ существующей топологии")

	var topologyAnalysis map[string]interface{}
	err := workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"AnalyzeMeshTopologyActivity",
		input.NetworkID,
		input.Peers,
	).Get(ctx, &topologyAnalysis)

	if err != nil {
		logger.Error("Ошибка анализа топологии", zap.Error(err))
		result.Status = "failed"
		result.Message = "Не удалось проанализировать топологию: " + err.Error()
		result.CompletedAt = workflow.Now(ctx)
		return result, err
	}

	// 2. Планирование новой топологии
	logger.Info("Планирование новой топологии", zap.String("type", input.TopologyType))

	var topologyPlan map[string]interface{}
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"PlanMeshTopologyActivity",
		input.TopologyType,
		input.Peers,
		topologyAnalysis,
		input.Configuration,
	).Get(ctx, &topologyPlan)

	if err != nil {
		logger.Error("Ошибка планирования топологии", zap.Error(err))
		result.Status = "failed"
		result.Message = "Не удалось спланировать топологию: " + err.Error()
		result.CompletedAt = workflow.Now(ctx)
		return result, err
	}

	// 3. Построение топологии
	logger.Info("Построение mesh топологии")

	var buildResult map[string]interface{}
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"BuildMeshTopologyActivity",
		input.NetworkID,
		topologyPlan,
	).Get(ctx, &buildResult)

	if err != nil {
		logger.Error("Ошибка построения топологии", zap.Error(err))
		result.Status = "failed"
		result.Message = "Не удалось построить топологию: " + err.Error()
		result.CompletedAt = workflow.Now(ctx)
		return result, err
	}

	// 4. Валидация и тестирование топологии
	logger.Info("Валидация новой топологии")

	var validationResult map[string]interface{}
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"ValidateMeshTopologyActivity",
		input.NetworkID,
		buildResult,
	).Get(ctx, &validationResult)

	if err != nil {
		logger.Warn("Ошибка валидации топологии", zap.Error(err))
		// Продолжаем даже если валидация не удалась
	}

	// Обработка результатов
	if connectedPeers, ok := buildResult["connected_peers"].(float64); ok {
		result.ConnectedPeers = int(connectedPeers)
	}

	if failedPeers, ok := buildResult["failed_peers"].([]interface{}); ok {
		for _, peer := range failedPeers {
			if peerStr, ok := peer.(string); ok {
				result.FailedPeers = append(result.FailedPeers, peerStr)
			}
		}
	}

	result.Topology = buildResult
	result.Status = "completed"
	result.Message = "Mesh топология успешно построена"
	result.CompletedAt = workflow.Now(ctx)

	logger.Info("MeshTopologyWorkflow завершен",
		zap.String("network_id", input.NetworkID),
		zap.String("topology_type", input.TopologyType),
		zap.String("status", result.Status),
		zap.Int("connected_peers", result.ConnectedPeers),
		zap.Int("failed_peers", len(result.FailedPeers)),
		zap.Duration("duration", result.CompletedAt.Sub(result.StartedAt)),
	)

	return result, nil
}

// ConnectionRecoveryWorkflow workflow для восстановления соединений
func ConnectionRecoveryWorkflow(ctx workflow.Context, input ConnectionRecoveryWorkflowInput) (*ConnectionRecoveryWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Запуск ConnectionRecoveryWorkflow",
		zap.Strings("failed_connections", input.FailedConnections),
		zap.String("strategy", input.RecoveryStrategy),
		zap.Int("max_retries", input.MaxRetries),
	)

	result := &ConnectionRecoveryWorkflowResult{
		Status:               "recovering",
		RecoveryStrategy:     input.RecoveryStrategy,
		StartedAt:            workflow.Now(ctx),
		RecoveredConnections: make([]string, 0),
		FailedConnections:    make([]string, 0),
		Results:              make(map[string]interface{}),
	}

	// Настройки для activities
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		HeartbeatTimeout:    30 * time.Second,
		RetryPolicy: &workflow.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 1.5,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    int32(input.MaxRetries),
		},
	}

	switch input.RecoveryStrategy {
	case "immediate":
		// Немедленное восстановление всех соединений параллельно
		logger.Info("Немедленное восстановление соединений")

		futures := make([]workflow.Future, len(input.FailedConnections))
		for i, connectionID := range input.FailedConnections {
			futures[i] = workflow.ExecuteActivity(
				workflow.WithActivityOptions(ctx, activityOptions),
				"RecoverConnectionActivity",
				connectionID,
				"immediate",
				input.Configuration,
			)
		}

		// Ждем завершения всех восстановлений
		for i, future := range futures {
			var recoveryResult map[string]interface{}
			err := future.Get(ctx, &recoveryResult)

			connectionID := input.FailedConnections[i]
			if err != nil {
				logger.Warn("Не удалось восстановить соединение",
					zap.String("connection_id", connectionID),
					zap.Error(err),
				)
				result.FailedConnections = append(result.FailedConnections, connectionID)
			} else {
				logger.Info("Соединение восстановлено",
					zap.String("connection_id", connectionID),
				)
				result.RecoveredConnections = append(result.RecoveredConnections, connectionID)
			}

			result.Results[connectionID] = recoveryResult
		}

	case "gradual":
		// Постепенное восстановление соединений с задержками
		logger.Info("Постепенное восстановление соединений")

		for _, connectionID := range input.FailedConnections {
			var recoveryResult map[string]interface{}
			err := workflow.ExecuteActivity(
				workflow.WithActivityOptions(ctx, activityOptions),
				"RecoverConnectionActivity",
				connectionID,
				"gradual",
				input.Configuration,
			).Get(ctx, &recoveryResult)

			if err != nil {
				logger.Warn("Не удалось восстановить соединение",
					zap.String("connection_id", connectionID),
					zap.Error(err),
				)
				result.FailedConnections = append(result.FailedConnections, connectionID)
			} else {
				logger.Info("Соединение восстановлено",
					zap.String("connection_id", connectionID),
				)
				result.RecoveredConnections = append(result.RecoveredConnections, connectionID)
			}

			result.Results[connectionID] = recoveryResult
			result.Attempts++

			// Задержка между восстановлениями
			if len(input.FailedConnections) > 1 {
				workflow.Sleep(ctx, 10*time.Second)
			}
		}

	case "scheduled":
		// Запланированное восстановление (например, в период низкой нагрузки)
		logger.Info("Запланированное восстановление соединений")

		// Ждем оптимального времени для восстановления
		var scheduleResult map[string]interface{}
		err := workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, activityOptions),
			"ScheduleRecoveryActivity",
			input.FailedConnections,
			input.Configuration,
		).Get(ctx, &scheduleResult)

		if err != nil {
			logger.Error("Ошибка планирования восстановления", zap.Error(err))
			result.Status = "failed"
			result.Message = "Не удалось запланировать восстановление: " + err.Error()
			result.CompletedAt = workflow.Now(ctx)
			return result, err
		}

		// Ждем запланированного времени
		if scheduleTime, ok := scheduleResult["schedule_time"].(string); ok {
			scheduledTime, _ := time.Parse(time.RFC3339, scheduleTime)
			waitDuration := scheduledTime.Sub(workflow.Now(ctx))
			if waitDuration > 0 {
				logger.Info("Ожидание запланированного времени восстановления",
					zap.Duration("wait_duration", waitDuration),
				)
				workflow.Sleep(ctx, waitDuration)
			}
		}

		// Выполняем восстановление
		for _, connectionID := range input.FailedConnections {
			var recoveryResult map[string]interface{}
			err := workflow.ExecuteActivity(
				workflow.WithActivityOptions(ctx, activityOptions),
				"RecoverConnectionActivity",
				connectionID,
				"scheduled",
				input.Configuration,
			).Get(ctx, &recoveryResult)

			if err != nil {
				logger.Warn("Не удалось восстановить соединение",
					zap.String("connection_id", connectionID),
					zap.Error(err),
				)
				result.FailedConnections = append(result.FailedConnections, connectionID)
			} else {
				logger.Info("Соединение восстановлено",
					zap.String("connection_id", connectionID),
				)
				result.RecoveredConnections = append(result.RecoveredConnections, connectionID)
			}

			result.Results[connectionID] = recoveryResult
			result.Attempts++
		}

	default:
		logger.Error("Неизвестная стратегия восстановления", zap.String("strategy", input.RecoveryStrategy))
		result.Status = "failed"
		result.Message = "Неизвестная стратегия восстановления: " + input.RecoveryStrategy
		result.CompletedAt = workflow.Now(ctx)
		return result, fmt.Errorf("unknown_strategy: %s", input.RecoveryStrategy)
	}

	// Определяем итоговый статус
	if len(result.FailedConnections) == 0 {
		result.Status = "completed"
		result.Message = "Все соединения успешно восстановлены"
	} else if len(result.RecoveredConnections) > 0 {
		result.Status = "partial"
		result.Message = "Часть соединений восстановлена"
	} else {
		result.Status = "failed"
		result.Message = "Не удалось восстановить ни одного соединения"
	}

	result.CompletedAt = workflow.Now(ctx)

	logger.Info("ConnectionRecoveryWorkflow завершен",
		zap.String("strategy", input.RecoveryStrategy),
		zap.String("status", result.Status),
		zap.Int("recovered", len(result.RecoveredConnections)),
		zap.Int("failed", len(result.FailedConnections)),
		zap.Int("attempts", result.Attempts),
		zap.Duration("duration", result.CompletedAt.Sub(result.StartedAt)),
	)

	return result, nil
}

// LoadBalancingWorkflow workflow для балансировки нагрузки
func LoadBalancingWorkflow(ctx workflow.Context, input LoadBalancingWorkflowInput) (*LoadBalancingWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Запуск LoadBalancingWorkflow",
		zap.String("load_balancer_id", input.LoadBalancerID),
		zap.String("strategy", input.Strategy),
		zap.Int("peers_count", len(input.Peers)),
	)

	result := &LoadBalancingWorkflowResult{
		LoadBalancerID:   input.LoadBalancerID,
		Strategy:         input.Strategy,
		Status:           "configuring",
		StartedAt:        workflow.Now(ctx),
		ActivePeers:      make([]string, 0),
		LoadDistribution: make(map[string]float64),
		Configuration:    input.Configuration,
	}

	// Настройки для activities
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		HeartbeatTimeout:    1 * time.Minute,
		RetryPolicy: &workflow.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    3,
		},
	}

	// 1. Анализ текущей нагрузки на пиры
	logger.Info("Анализ текущей нагрузки")

	var loadAnalysis map[string]interface{}
	err := workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"AnalyzeLoadActivity",
		input.Peers,
		input.LoadBalancerID,
	).Get(ctx, &loadAnalysis)

	if err != nil {
		logger.Error("Ошибка анализа нагрузки", zap.Error(err))
		result.Status = "failed"
		result.Message = "Не удалось проанализировать нагрузку: " + err.Error()
		result.CompletedAt = workflow.Now(ctx)
		return result, err
	}

	// 2. Планирование балансировки
	logger.Info("Планирование балансировки", zap.String("strategy", input.Strategy))

	var balancingPlan map[string]interface{}
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"PlanLoadBalancingActivity",
		input.Strategy,
		input.Peers,
		loadAnalysis,
		input.Configuration,
	).Get(ctx, &balancingPlan)

	if err != nil {
		logger.Error("Ошибка планирования балансировки", zap.Error(err))
		result.Status = "failed"
		result.Message = "Не удалось спланировать балансировку: " + err.Error()
		result.CompletedAt = workflow.Now(ctx)
		return result, err
	}

	// 3. Применение балансировки
	logger.Info("Применение балансировки нагрузки")

	var applyResult map[string]interface{}
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"ApplyLoadBalancingActivity",
		input.LoadBalancerID,
		balancingPlan,
	).Get(ctx, &applyResult)

	if err != nil {
		logger.Error("Ошибка применения балансировки", zap.Error(err))
		result.Status = "failed"
		result.Message = "Не удалось применить балансировку: " + err.Error()
		result.CompletedAt = workflow.Now(ctx)
		return result, err
	}

	// 4. Мониторинг балансировки (краткосрочный)
	logger.Info("Мониторинг балансировки")

	for i := 0; i < 5; i++ { // 5 проверок с интервалом 30 секунд
		var monitorResult map[string]interface{}
		err = workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, activityOptions),
			"MonitorLoadBalancingActivity",
			input.LoadBalancerID,
		).Get(ctx, &monitorResult)

		if err != nil {
			logger.Warn("Ошибка мониторинга балансировки", zap.Error(err))
		} else {
			logger.Debug("Результат мониторинга балансировки",
				zap.Any("result", monitorResult),
			)
		}

		if i < 4 { // Не ждем после последней проверки
			workflow.Sleep(ctx, 30*time.Second)
		}
	}

	// Обработка результатов
	if activePeers, ok := applyResult["active_peers"].([]interface{}); ok {
		for _, peer := range activePeers {
			if peerStr, ok := peer.(string); ok {
				result.ActivePeers = append(result.ActivePeers, peerStr)
			}
		}
	}

	if loadDist, ok := applyResult["load_distribution"].(map[string]interface{}); ok {
		for peer, load := range loadDist {
			if loadFloat, ok := load.(float64); ok {
				result.LoadDistribution[peer] = loadFloat
			}
		}
	}

	result.Status = "completed"
	result.Message = "Балансировка нагрузки успешно настроена"
	result.CompletedAt = workflow.Now(ctx)

	logger.Info("LoadBalancingWorkflow завершен",
		zap.String("load_balancer_id", input.LoadBalancerID),
		zap.String("strategy", input.Strategy),
		zap.String("status", result.Status),
		zap.Int("active_peers", len(result.ActivePeers)),
		zap.Duration("duration", result.CompletedAt.Sub(result.StartedAt)),
	)

	return result, nil
}
