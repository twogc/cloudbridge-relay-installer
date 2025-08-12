package workflows

import (
	"fmt"
	"time"

	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// CloudBridgeMainWorkflowCorrected - исправленная версия главного workflow
// с соблюдением всех принципов Cadence
func CloudBridgeMainWorkflowCorrected(ctx workflow.Context, input CloudBridgeMainWorkflowInput) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("CloudBridgeMainWorkflow started (corrected version)",
		zap.String("domain", input.Domain),
		zap.String("start_time", input.StartTime.Format(time.RFC3339)),
	)



	// Этап 1: Инициализация модулей
	logger.Info("Executing ModuleInitializationWorkflow...")
	err := workflow.ExecuteChildWorkflow(ctx, ModuleInitializationWorkflow, ModuleInitializationInput{
		ModuleName:    "core",
		Configuration: input.Configuration,
		Dependencies:  []string{"database", "metrics"},
	}).Get(ctx, nil)

	if err != nil {
		logger.Error("ModuleInitializationWorkflow failed", zap.Error(err))
		return fmt.Errorf("module initialization failed: %w", err)
	}
	logger.Info("ModuleInitializationWorkflow completed successfully")

	// Этап 2: Запуск долгоживущих child workflows
	logger.Info("Starting long-running child workflows...")

	// Список workflows для запуска (детерминированный порядок)
	workflowsToStart := []struct {
		name     string
		workflow interface{}
		input    interface{}
	}{
		{
			name:     "HealthCheckWorkflow",
			workflow: HealthCheckWorkflow,
			input:    HealthCheckInput{Component: "system", Checks: []string{"core", "network"}},
		},
		{
			name:     "PeerLifecycleWorkflow",
			workflow: PeerLifecycleWorkflow,
			input:    PeerLifecycleWorkflowInput{Operation: "monitor"},
		},
		{
			name:     "QuantumKeyRotationWorkflow",
			workflow: QuantumKeyRotationWorkflow,
			input:    QuantumKeyRotationWorkflowInput{RotationPolicy: "time_based"},
		},
		{
			name:     "ThreatAnalysisWorkflow",
			workflow: ThreatAnalysisWorkflow,
			input:    ThreatAnalysisWorkflowInput{AnalysisType: "continuous"},
		},
	}

	// Запускаем child workflows параллельно с использованием workflow.Go
	for _, wf := range workflowsToStart {
		// Важно: создаем копию переменной для замыкания
		workflowName := wf.name
		workflowFunc := wf.workflow
		workflowInput := wf.input

		workflow.Go(ctx, func(ctx workflow.Context) {
			logger.Info("Starting child workflow", zap.String("workflow", workflowName))
			err := workflow.ExecuteChildWorkflow(ctx, workflowFunc, workflowInput).Get(ctx, nil)
			if err != nil {
				logger.Error("Child workflow failed",
					zap.String("workflow", workflowName),
					zap.Error(err),
				)
				// В production здесь может быть логика перезапуска
			} else {
				logger.Info("Child workflow completed", zap.String("workflow", workflowName))
			}
		})
	}

	// Этап 3: Основной control loop с signal handling
	logger.Info("Entering main control loop...")

	// Настройка signal channels
	shutdownChannel := workflow.GetSignalChannel(ctx, "control-command")
	configUpdateChannel := workflow.GetSignalChannel(ctx, "config-update")

	// Главный цикл управления
	for {
		selector := workflow.NewSelector(ctx)

		// Периодический heartbeat (каждые 24 часа)
		heartbeatTimer := workflow.NewTimer(ctx, 24*time.Hour)
		selector.AddFuture(heartbeatTimer, func(f workflow.Future) {
			logger.Info("CloudBridgeMainWorkflow daily heartbeat",
				zap.String("uptime", workflow.Now(ctx).Sub(input.StartTime).String()),
			)

			// Выполняем периодические maintenance tasks
			workflow.Go(ctx, func(ctx workflow.Context) {
				err := performMaintenanceTasks(ctx, input.Configuration)
				if err != nil {
					logger.Error("Maintenance tasks failed", zap.Error(err))
				}
			})
		})

		// Обработка команд управления
		selector.AddReceive(shutdownChannel, func(c workflow.Channel, more bool) {
			var command string
			c.Receive(ctx, &command)
			
			logger.Info("Received control command", zap.String("command", command))
			
			switch command {
			case "shutdown", "stop":
				logger.Info("Graceful shutdown initiated")
				// Здесь можно добавить логику graceful shutdown дочерних workflows
				return // Выход из workflow
			case "restart":
				logger.Info("Restart requested - signaling child workflows")
				// Можно отправить сигналы дочерним workflows
			case "health-check":
				logger.Info("Health check requested")
				// Запускаем health check activity
				workflow.Go(ctx, func(ctx workflow.Context) {
					performHealthCheckTask(ctx)
				})
			default:
				logger.Warn("Unknown control command", zap.String("command", command))
			}
		})

		// Обработка обновления конфигурации
		selector.AddReceive(configUpdateChannel, func(c workflow.Channel, more bool) {
			var newConfig map[string]interface{}
			c.Receive(ctx, &newConfig)
			
			logger.Info("Configuration update received")
			
			// Обновляем конфигурацию (детерминированно)
			workflow.Go(ctx, func(ctx workflow.Context) {
				err := applyConfigurationUpdate(ctx, newConfig)
				if err != nil {
					logger.Error("Configuration update failed", zap.Error(err))
				} else {
					logger.Info("Configuration updated successfully")
				}
			})
		})

		// Выполняем selector (блокирующий вызов до получения события)
		selector.Select(ctx)

		// Проверяем, не нужно ли завершиться
		// (эта проверка выполнится после каждого события)
		if ctx.Err() != nil {
			logger.Info("Workflow context cancelled, exiting")
			break
		}
	}

	logger.Info("CloudBridgeMainWorkflow finished gracefully")
	return nil
}

// performMaintenanceTasks - выполняет периодические maintenance задачи
func performMaintenanceTasks(ctx workflow.Context, config map[string]interface{}) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting maintenance tasks")

	// Настройки для maintenance activities
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		HeartbeatTimeout:    time.Minute,
		RetryPolicy: &workflow.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Список maintenance задач (детерминированный порядок)
	maintenanceTasks := []string{
		"CleanupLogsActivity",
		"UpdateMetricsActivity", 
		"CheckDiskSpaceActivity",
		"ValidateConfigurationActivity",
	}

	for _, task := range maintenanceTasks {
		logger.Info("Executing maintenance task", zap.String("task", task))
		
		err := workflow.ExecuteActivity(ctx, task, config).Get(ctx, nil)
		if err != nil {
			logger.Error("Maintenance task failed",
				zap.String("task", task),
				zap.Error(err),
			)
			// Продолжаем выполнение остальных задач
		} else {
			logger.Info("Maintenance task completed", zap.String("task", task))
		}
	}

	logger.Info("Maintenance tasks completed")
	return nil
}

// performHealthCheckTask - выполняет проверку здоровья системы
func performHealthCheckTask(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting health check task")

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		HeartbeatTimeout:    30 * time.Second,
		RetryPolicy: &workflow.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    2,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Детерминированный список компонентов для проверки
	components := []string{"database", "redis", "wireguard", "quantum", "ai"}

	for _, component := range components {
		var healthStatus map[string]interface{}
		err := workflow.ExecuteActivity(ctx, "CheckComponentHealthActivity", component).Get(ctx, &healthStatus)
		
		if err != nil {
			logger.Error("Health check failed",
				zap.String("component", component),
				zap.Error(err),
			)
		} else {
			status := "unknown"
			if s, ok := healthStatus["status"].(string); ok {
				status = s
			}
			
			logger.Info("Component health check completed",
				zap.String("component", component),
				zap.String("status", status),
			)
		}
	}

	return nil
}

// applyConfigurationUpdate - применяет обновление конфигурации
func applyConfigurationUpdate(ctx workflow.Context, newConfig map[string]interface{}) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Applying configuration update")

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		HeartbeatTimeout:    30 * time.Second,
		RetryPolicy: &workflow.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Валидация новой конфигурации
	var validationResult map[string]interface{}
	err := workflow.ExecuteActivity(ctx, "ValidateConfigurationActivity", newConfig).Get(ctx, &validationResult)
	if err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	isValid := false
	if v, ok := validationResult["valid"].(bool); ok {
		isValid = v
	}

	if !isValid {
		return fmt.Errorf("configuration validation failed: invalid configuration")
	}

	// Применение конфигурации
	err = workflow.ExecuteActivity(ctx, "ApplyConfigurationActivity", newConfig).Get(ctx, nil)
	if err != nil {
		return fmt.Errorf("configuration application failed: %w", err)
	}

	// Уведомление дочерних workflows об обновлении конфигурации
	// (через signals, если необходимо)

	logger.Info("Configuration update applied successfully")
	return nil
}

// CloudBridgeMainWorkflowWithContinueAsNew - версия с Continue-As-New
// для предотвращения роста истории workflow
func CloudBridgeMainWorkflowWithContinueAsNew(ctx workflow.Context, input CloudBridgeMainWorkflowInput) error {
	logger := workflow.GetLogger(ctx)
	
	// Проверяем, не пора ли сделать Continue-As-New
	// (например, каждые 30 дней или при достижении определенного размера истории)
	if shouldContinueAsNew(ctx, input) {
		logger.Info("Performing Continue-As-New to prevent history growth")
		
		// Подготавливаем новый input с обновленным временем
		newInput := input
		newInput.StartTime = workflow.Now(ctx)
		
		return workflow.NewContinueAsNewError(ctx, CloudBridgeMainWorkflowWithContinueAsNew, newInput)
	}
	
	// Основная логика workflow (аналогично CloudBridgeMainWorkflowCorrected)
	return CloudBridgeMainWorkflowCorrected(ctx, input)
}

// shouldContinueAsNew определяет, нужно ли выполнить Continue-As-New
func shouldContinueAsNew(ctx workflow.Context, input CloudBridgeMainWorkflowInput) bool {
	// Проверяем время работы workflow
	uptime := workflow.Now(ctx).Sub(input.StartTime)
	
	// Continue-As-New каждые 30 дней
	if uptime > 30*24*time.Hour {
		return true
	}
	
	// Можно добавить другие условия:
	// - размер истории workflow
	// - количество событий
	// - наличие специального signal
	
	return false
}
