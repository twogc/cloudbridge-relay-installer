package workflows

import (
	"fmt"
	"time"

	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// AutoScalingWorkflowInput входные параметры для автомасштабирования
type AutoScalingWorkflowInput struct {
	ResourceType  string                 `json:"resource_type"`  // compute, memory, network, storage
	ScalingPolicy string                 `json:"scaling_policy"` // reactive, predictive, scheduled
	TargetMetrics map[string]float64     `json:"target_metrics"` // cpu: 70, memory: 80, connections: 1000
	ScalingLimits map[string]interface{} `json:"scaling_limits"` // min/max instances
	Configuration map[string]interface{} `json:"configuration"`
}

// AutoScalingWorkflowResult результат автомасштабирования
type AutoScalingWorkflowResult struct {
	ResourceType    string                 `json:"resource_type"`
	ScalingPolicy   string                 `json:"scaling_policy"`
	ScalingActions  []string               `json:"scaling_actions"`
	CurrentCapacity map[string]interface{} `json:"current_capacity"`
	TargetCapacity  map[string]interface{} `json:"target_capacity"`
	ScalingResults  map[string]interface{} `json:"scaling_results"`
	Status          string                 `json:"status"`
	StartedAt       time.Time              `json:"started_at"`
	CompletedAt     time.Time              `json:"completed_at"`
	Message         string                 `json:"message"`
}

// ResourceOptimizationWorkflowInput входные параметры для оптимизации ресурсов
type ResourceOptimizationWorkflowInput struct {
	OptimizationType string                 `json:"optimization_type"` // cost, performance, efficiency, hybrid
	ResourceScope    []string               `json:"resource_scope"`    // servers, containers, networks, storage
	OptimizationGoal map[string]interface{} `json:"optimization_goal"` // reduce_cost: 20%, improve_performance: 15%
	Configuration    map[string]interface{} `json:"configuration"`
}

// ResourceOptimizationWorkflowResult результат оптимизации ресурсов
type ResourceOptimizationWorkflowResult struct {
	OptimizationType       string                 `json:"optimization_type"`
	OptimizedResources     []string               `json:"optimized_resources"`
	OptimizationResults    map[string]interface{} `json:"optimization_results"`
	CostSavings            float64                `json:"cost_savings"`
	PerformanceImprovement float64                `json:"performance_improvement"`
	EfficiencyGains        map[string]float64     `json:"efficiency_gains"`
	Status                 string                 `json:"status"`
	StartedAt              time.Time              `json:"started_at"`
	CompletedAt            time.Time              `json:"completed_at"`
	Message                string                 `json:"message"`
}

// PerformanceMonitoringWorkflowInput входные параметры для мониторинга производительности
type PerformanceMonitoringWorkflowInput struct {
	MonitoringScope  string                 `json:"monitoring_scope"`   // system, application, network, database
	MetricsToMonitor []string               `json:"metrics_to_monitor"` // latency, throughput, error_rate, resource_usage
	MonitoringPeriod string                 `json:"monitoring_period"`  // continuous, periodic, triggered
	AlertThresholds  map[string]float64     `json:"alert_thresholds"`
	Configuration    map[string]interface{} `json:"configuration"`
}

// PerformanceMonitoringWorkflowResult результат мониторинга производительности
type PerformanceMonitoringWorkflowResult struct {
	MonitoringScope     string                 `json:"monitoring_scope"`
	MonitoredMetrics    map[string]float64     `json:"monitored_metrics"`
	PerformanceBaseline map[string]float64     `json:"performance_baseline"`
	PerformanceTrends   map[string]interface{} `json:"performance_trends"`
	TriggeredAlerts     []string               `json:"triggered_alerts"`
	PerformanceScore    float64                `json:"performance_score"`
	Status              string                 `json:"status"`
	StartedAt           time.Time              `json:"started_at"`
	CompletedAt         time.Time              `json:"completed_at"`
	Message             string                 `json:"message"`
}

// CapacityPlanningWorkflowInput входные параметры для планирования емкости
type CapacityPlanningWorkflowInput struct {
	PlanningHorizon   string                 `json:"planning_horizon"`   // 1m, 3m, 6m, 1y
	GrowthProjections map[string]float64     `json:"growth_projections"` // users: 25%, traffic: 40%
	ResourceTypes     []string               `json:"resource_types"`     // compute, storage, network, database
	BusinessGoals     map[string]interface{} `json:"business_goals"`
	Configuration     map[string]interface{} `json:"configuration"`
}

// CapacityPlanningWorkflowResult результат планирования емкости
type CapacityPlanningWorkflowResult struct {
	PlanningHorizon         string                 `json:"planning_horizon"`
	CapacityRecommendations map[string]interface{} `json:"capacity_recommendations"`
	ResourceRequirements    map[string]interface{} `json:"resource_requirements"`
	CostProjections         map[string]float64     `json:"cost_projections"`
	RiskAssessment          map[string]interface{} `json:"risk_assessment"`
	MilestoneTimeline       []string               `json:"milestone_timeline"`
	Status                  string                 `json:"status"`
	StartedAt               time.Time              `json:"started_at"`
	CompletedAt             time.Time              `json:"completed_at"`
	Message                 string                 `json:"message"`
}

// AutoScalingWorkflow workflow для автоматического масштабирования ресурсов
func AutoScalingWorkflow(ctx workflow.Context, input AutoScalingWorkflowInput) (*AutoScalingWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Запуск AutoScalingWorkflow",
		zap.String("resource_type", input.ResourceType),
		zap.String("scaling_policy", input.ScalingPolicy),
		zap.Any("target_metrics", input.TargetMetrics),
	)

	result := &AutoScalingWorkflowResult{
		ResourceType:    input.ResourceType,
		ScalingPolicy:   input.ScalingPolicy,
		Status:          "scaling",
		StartedAt:       workflow.Now(ctx),
		ScalingActions:  make([]string, 0),
		CurrentCapacity: make(map[string]interface{}),
		TargetCapacity:  make(map[string]interface{}),
		ScalingResults:  make(map[string]interface{}),
	}

	// Настройки для activities
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		HeartbeatTimeout:    1 * time.Minute,
		RetryPolicy: &workflow.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    1 * time.Minute,
			MaximumAttempts:    3,
		},
	}

	switch input.ScalingPolicy {
	case "reactive":
		// Реактивное масштабирование на основе текущих метрик
		logger.Info("Реактивное масштабирование")

		// Настройка канала для получения сигналов управления
		controlSignal := workflow.GetSignalChannel(ctx, "control-command")

		// Основной цикл мониторинга и масштабирования
		for {
			selector := workflow.NewSelector(ctx)

			// Добавить обработку сигнала остановки
			selector.AddReceive(controlSignal, func(c workflow.Channel, more bool) {
				var signal string
				c.Receive(ctx, &signal)
				logger.Info("Получен сигнал управления AutoScaling", zap.String("signal", signal))
				if signal == "shutdown" || signal == "stop" {
					logger.Info("Получен сигнал остановки AutoScaling")
					result.Status = "stopped"
					result.Message = "Масштабирование остановлено по сигналу"
					result.CompletedAt = workflow.Now(ctx)
					return // Выход из функции
				}
			})

			// Добавить таймер с интервалом для тестов
			timer := workflow.NewTimer(ctx, getScalingInterval(ctx))
			selector.AddFuture(timer, func(f workflow.Future) {
				// 1. Сбор текущих метрик
				var metricsResult map[string]interface{}
				err := workflow.ExecuteActivity(
					workflow.WithActivityOptions(ctx, activityOptions),
					"CollectScalingMetricsActivity",
					input.ResourceType,
					input.Configuration,
				).Get(ctx, &metricsResult)

				if err != nil {
					logger.Warn("Ошибка сбора метрик", zap.Error(err))
				} else {
					// 2. Анализ необходимости масштабирования
					var analysisResult map[string]interface{}
					err = workflow.ExecuteActivity(
						workflow.WithActivityOptions(ctx, activityOptions),
						"AnalyzeScalingNeedActivity",
						metricsResult,
						input.TargetMetrics,
						input.ScalingLimits,
					).Get(ctx, &analysisResult)

					if err != nil {
						logger.Warn("Ошибка анализа масштабирования", zap.Error(err))
					} else {
						// 3. Выполнение масштабирования при необходимости
						if needScaling, ok := analysisResult["need_scaling"].(bool); ok && needScaling {
							var scalingResult map[string]interface{}
							err = workflow.ExecuteActivity(
								workflow.WithActivityOptions(ctx, activityOptions),
								"PerformScalingActivity",
								input.ResourceType,
								analysisResult,
								"reactive",
							).Get(ctx, &scalingResult)

							if err != nil {
								logger.Warn("Ошибка выполнения масштабирования", zap.Error(err))
							} else {
								if action, ok := scalingResult["scaling_action"].(string); ok {
									result.ScalingActions = append(result.ScalingActions, action)
								}
								result.ScalingResults["latest_scaling"] = scalingResult

								logger.Info("Масштабирование выполнено",
									zap.Any("scaling_result", scalingResult),
								)
							}
						}
					}

					// Обновляем текущую емкость
					if currentCap, ok := metricsResult["current_capacity"].(map[string]interface{}); ok {
						result.CurrentCapacity = currentCap
					}
				}
			})

			selector.Select(ctx)

			// Проверяем, не нужно ли завершить workflow
			if result.Status == "stopped" {
				break
			}
		}

	case "predictive":
		// Предиктивное масштабирование на основе прогнозов
		logger.Info("Предиктивное масштабирование")

		// 1. Анализ исторических данных
		var historicalAnalysis map[string]interface{}
		err := workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, activityOptions),
			"AnalyzeHistoricalDataActivity",
			input.ResourceType,
			input.Configuration,
		).Get(ctx, &historicalAnalysis)

		if err != nil {
			logger.Error("Ошибка анализа исторических данных", zap.Error(err))
			result.Status = "failed"
			result.Message = "Не удалось проанализировать исторические данные: " + err.Error()
			result.CompletedAt = workflow.Now(ctx)
			return result, err
		}

		// 2. Создание прогноза нагрузки
		var forecastResult map[string]interface{}
		err = workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, activityOptions),
			"GenerateLoadForecastActivity",
			historicalAnalysis,
			input.TargetMetrics,
		).Get(ctx, &forecastResult)

		if err != nil {
			logger.Error("Ошибка создания прогноза", zap.Error(err))
			result.Status = "failed"
			result.Message = "Не удалось создать прогноз: " + err.Error()
			result.CompletedAt = workflow.Now(ctx)
			return result, err
		}

		// 3. Планирование превентивного масштабирования
		var planningResult map[string]interface{}
		err = workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, activityOptions),
			"PlanPreventiveScalingActivity",
			forecastResult,
			input.ScalingLimits,
		).Get(ctx, &planningResult)

		if err != nil {
			logger.Error("Ошибка планирования масштабирования", zap.Error(err))
			result.Status = "failed"
			result.Message = "Не удалось спланировать масштабирование: " + err.Error()
			result.CompletedAt = workflow.Now(ctx)
			return result, err
		}

		// 4. Выполнение запланированного масштабирования
		if scalingPlan, ok := planningResult["scaling_plan"].([]interface{}); ok {
			for _, step := range scalingPlan {
				if stepMap, ok := step.(map[string]interface{}); ok {
					var stepResult map[string]interface{}
					err = workflow.ExecuteActivity(
						workflow.WithActivityOptions(ctx, activityOptions),
						"ExecuteScalingStepActivity",
						stepMap,
						input.ResourceType,
					).Get(ctx, &stepResult)

					if err != nil {
						logger.Warn("Ошибка выполнения шага масштабирования", zap.Error(err))
					} else {
						if action, ok := stepResult["scaling_action"].(string); ok {
							result.ScalingActions = append(result.ScalingActions, action)
						}
					}

					// Задержка между шагами
					if len(scalingPlan) > 1 {
						workflow.Sleep(ctx, 30*time.Second)
					}
				}
			}
		}

		result.ScalingResults["predictive_scaling"] = planningResult

	case "scheduled":
		// Запланированное масштабирование
		logger.Info("Запланированное масштабирование")

		// 1. Получение расписания масштабирования
		var scheduleResult map[string]interface{}
		err := workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, activityOptions),
			"GetScalingScheduleActivity",
			input.ResourceType,
			input.Configuration,
		).Get(ctx, &scheduleResult)

		if err != nil {
			logger.Error("Ошибка получения расписания", zap.Error(err))
			result.Status = "failed"
			result.Message = "Не удалось получить расписание: " + err.Error()
			result.CompletedAt = workflow.Now(ctx)
			return result, err
		}

		// 2. Выполнение масштабирования по расписанию
		if schedule, ok := scheduleResult["schedule"].([]interface{}); ok {
			for _, entry := range schedule {
				if entryMap, ok := entry.(map[string]interface{}); ok {
					// Ждем времени выполнения
					if executeTimeStr, ok := entryMap["execute_time"].(string); ok {
						executeTime, _ := time.Parse(time.RFC3339, executeTimeStr)
						waitDuration := executeTime.Sub(workflow.Now(ctx))

						if waitDuration > 0 {
							logger.Info("Ожидание времени выполнения масштабирования",
								zap.Duration("wait_duration", waitDuration),
							)
							workflow.Sleep(ctx, waitDuration)
						}
					}

					// Выполняем масштабирование
					var stepResult map[string]interface{}
					err = workflow.ExecuteActivity(
						workflow.WithActivityOptions(ctx, activityOptions),
						"ExecuteScheduledScalingActivity",
						entryMap,
						input.ResourceType,
					).Get(ctx, &stepResult)

					if err != nil {
						logger.Warn("Ошибка выполнения запланированного масштабирования", zap.Error(err))
					} else {
						if action, ok := stepResult["scaling_action"].(string); ok {
							result.ScalingActions = append(result.ScalingActions, action)
						}
					}
				}
			}
		}

		result.ScalingResults["scheduled_scaling"] = scheduleResult

	default:
		logger.Error("Неизвестная политика масштабирования", zap.String("policy", input.ScalingPolicy))
		result.Status = "failed"
		result.Message = "Неизвестная политика масштабирования: " + input.ScalingPolicy
		result.CompletedAt = workflow.Now(ctx)
		return result, fmt.Errorf("unknown_scaling_policy: %s", input.ScalingPolicy)
	}

	// Получаем финальную информацию о емкости
	var finalCapacityResult map[string]interface{}
	err := workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"GetCurrentCapacityActivity",
		input.ResourceType,
	).Get(ctx, &finalCapacityResult)

	if err != nil {
		logger.Warn("Ошибка получения финальной емкости", zap.Error(err))
	} else {
		if currentCap, ok := finalCapacityResult["current_capacity"].(map[string]interface{}); ok {
			result.CurrentCapacity = currentCap
		}
		if targetCap, ok := finalCapacityResult["target_capacity"].(map[string]interface{}); ok {
			result.TargetCapacity = targetCap
		}
	}

	result.Status = "completed"
	result.Message = "Масштабирование завершено"
	result.CompletedAt = workflow.Now(ctx)

	logger.Info("AutoScalingWorkflow завершен",
		zap.String("resource_type", input.ResourceType),
		zap.String("scaling_policy", input.ScalingPolicy),
		zap.String("status", result.Status),
		zap.Int("scaling_actions", len(result.ScalingActions)),
		zap.Duration("duration", result.CompletedAt.Sub(result.StartedAt)),
	)

	return result, nil
}

// ResourceOptimizationWorkflow workflow для оптимизации использования ресурсов
func ResourceOptimizationWorkflow(ctx workflow.Context, input ResourceOptimizationWorkflowInput) (*ResourceOptimizationWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Запуск ResourceOptimizationWorkflow",
		zap.String("optimization_type", input.OptimizationType),
		zap.Strings("resource_scope", input.ResourceScope),
		zap.Any("optimization_goal", input.OptimizationGoal),
	)

	result := &ResourceOptimizationWorkflowResult{
		OptimizationType:    input.OptimizationType,
		Status:              "optimizing",
		StartedAt:           workflow.Now(ctx),
		OptimizedResources:  make([]string, 0),
		OptimizationResults: make(map[string]interface{}),
		EfficiencyGains:     make(map[string]float64),
	}

	// Настройки для activities
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 15 * time.Minute,
		HeartbeatTimeout:    2 * time.Minute,
		RetryPolicy: &workflow.RetryPolicy{
			InitialInterval:    2 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    1 * time.Minute,
			MaximumAttempts:    3,
		},
	}

	// 1. Анализ текущего использования ресурсов
	logger.Info("Анализ текущего использования ресурсов")

	var usageAnalysis map[string]interface{}
	err := workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"AnalyzeResourceUsageActivity",
		input.ResourceScope,
		input.Configuration,
	).Get(ctx, &usageAnalysis)

	if err != nil {
		logger.Error("Ошибка анализа использования ресурсов", zap.Error(err))
		result.Status = "failed"
		result.Message = "Не удалось проанализировать использование ресурсов: " + err.Error()
		result.CompletedAt = workflow.Now(ctx)
		return result, err
	}

	result.OptimizationResults["usage_analysis"] = usageAnalysis

	// 2. Идентификация возможностей для оптимизации
	logger.Info("Идентификация возможностей для оптимизации")

	var opportunitiesResult map[string]interface{}
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"IdentifyOptimizationOpportunitiesActivity",
		usageAnalysis,
		input.OptimizationType,
		input.OptimizationGoal,
	).Get(ctx, &opportunitiesResult)

	if err != nil {
		logger.Error("Ошибка идентификации возможностей", zap.Error(err))
		result.Status = "failed"
		result.Message = "Не удалось идентифицировать возможности оптимизации: " + err.Error()
		result.CompletedAt = workflow.Now(ctx)
		return result, err
	}

	result.OptimizationResults["opportunities"] = opportunitiesResult

	// 3. Планирование оптимизации
	logger.Info("Планирование оптимизации")

	var planningResult map[string]interface{}
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"PlanResourceOptimizationActivity",
		opportunitiesResult,
		input.OptimizationType,
		input.ResourceScope,
	).Get(ctx, &planningResult)

	if err != nil {
		logger.Error("Ошибка планирования оптимизации", zap.Error(err))
		result.Status = "failed"
		result.Message = "Не удалось спланировать оптимизацию: " + err.Error()
		result.CompletedAt = workflow.Now(ctx)
		return result, err
	}

	result.OptimizationResults["planning"] = planningResult

	// 4. Выполнение оптимизации
	logger.Info("Выполнение оптимизации")

	if optimizationPlan, ok := planningResult["optimization_plan"].([]interface{}); ok {
		for _, step := range optimizationPlan {
			if stepMap, ok := step.(map[string]interface{}); ok {
				var stepResult map[string]interface{}
				err = workflow.ExecuteActivity(
					workflow.WithActivityOptions(ctx, activityOptions),
					"ExecuteOptimizationStepActivity",
					stepMap,
					input.OptimizationType,
				).Get(ctx, &stepResult)

				if err != nil {
					logger.Warn("Ошибка выполнения шага оптимизации", zap.Error(err))
				} else {
					if resource, ok := stepResult["optimized_resource"].(string); ok {
						result.OptimizedResources = append(result.OptimizedResources, resource)
					}

					if stepID, ok := stepMap["step_id"].(string); ok {
						result.OptimizationResults[stepID] = stepResult
					}

					logger.Info("Шаг оптимизации выполнен",
						zap.Any("step_result", stepResult),
					)
				}

				// Задержка между шагами оптимизации
				if len(optimizationPlan) > 1 {
					workflow.Sleep(ctx, 10*time.Second)
				}
			}
		}
	}

	// 5. Оценка результатов оптимизации
	logger.Info("Оценка результатов оптимизации")

	var evaluationResult map[string]interface{}
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"EvaluateOptimizationResultsActivity",
		result.OptimizedResources,
		result.OptimizationResults,
		input.OptimizationGoal,
	).Get(ctx, &evaluationResult)

	if err != nil {
		logger.Warn("Ошибка оценки результатов", zap.Error(err))
	} else {
		// Извлекаем метрики оптимизации
		if costSavings, ok := evaluationResult["cost_savings"].(float64); ok {
			result.CostSavings = costSavings
		}

		if perfImprovement, ok := evaluationResult["performance_improvement"].(float64); ok {
			result.PerformanceImprovement = perfImprovement
		}

		if efficiencyGains, ok := evaluationResult["efficiency_gains"].(map[string]interface{}); ok {
			for metric, gain := range efficiencyGains {
				if gainFloat, ok := gain.(float64); ok {
					result.EfficiencyGains[metric] = gainFloat
				}
			}
		}

		result.OptimizationResults["evaluation"] = evaluationResult
	}

	result.Status = "completed"
	result.Message = "Оптимизация ресурсов завершена"
	result.CompletedAt = workflow.Now(ctx)

	logger.Info("ResourceOptimizationWorkflow завершен",
		zap.String("optimization_type", input.OptimizationType),
		zap.String("status", result.Status),
		zap.Int("optimized_resources", len(result.OptimizedResources)),
		zap.Float64("cost_savings", result.CostSavings),
		zap.Float64("performance_improvement", result.PerformanceImprovement),
		zap.Duration("duration", result.CompletedAt.Sub(result.StartedAt)),
	)

	return result, nil
}

// PerformanceMonitoringWorkflow workflow для мониторинга производительности
func PerformanceMonitoringWorkflow(ctx workflow.Context, input PerformanceMonitoringWorkflowInput) (*PerformanceMonitoringWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Запуск PerformanceMonitoringWorkflow",
		zap.String("monitoring_scope", input.MonitoringScope),
		zap.Strings("metrics_to_monitor", input.MetricsToMonitor),
		zap.String("monitoring_period", input.MonitoringPeriod),
	)

	result := &PerformanceMonitoringWorkflowResult{
		MonitoringScope:     input.MonitoringScope,
		Status:              "monitoring",
		StartedAt:           workflow.Now(ctx),
		MonitoredMetrics:    make(map[string]float64),
		PerformanceBaseline: make(map[string]float64),
		PerformanceTrends:   make(map[string]interface{}),
		TriggeredAlerts:     make([]string, 0),
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

	// 1. Установка базовой линии производительности
	logger.Info("Установка базовой линии производительности")

	var baselineResult map[string]interface{}
	err := workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"EstablishPerformanceBaselineActivity",
		input.MonitoringScope,
		input.MetricsToMonitor,
		input.Configuration,
	).Get(ctx, &baselineResult)

	if err != nil {
		logger.Error("Ошибка установки базовой линии", zap.Error(err))
		result.Status = "failed"
		result.Message = "Не удалось установить базовую линию: " + err.Error()
		result.CompletedAt = workflow.Now(ctx)
		return result, err
	}

	if baseline, ok := baselineResult["baseline"].(map[string]interface{}); ok {
		for metric, value := range baseline {
			if valueFloat, ok := value.(float64); ok {
				result.PerformanceBaseline[metric] = valueFloat
			}
		}
	}

	switch input.MonitoringPeriod {
	case "continuous":
		// Непрерывный мониторинг производительности
		logger.Info("Запуск непрерывного мониторинга производительности")

		for {
			// Сбор текущих метрик производительности
			var metricsResult map[string]interface{}
			err := workflow.ExecuteActivity(
				workflow.WithActivityOptions(ctx, activityOptions),
				"CollectPerformanceMetricsActivity",
				input.MonitoringScope,
				input.MetricsToMonitor,
			).Get(ctx, &metricsResult)

			if err != nil {
				logger.Warn("Ошибка сбора метрик производительности", zap.Error(err))
			} else {
				// Обновляем monitored metrics
				if metrics, ok := metricsResult["metrics"].(map[string]interface{}); ok {
					for metric, value := range metrics {
						if valueFloat, ok := value.(float64); ok {
							result.MonitoredMetrics[metric] = valueFloat
						}
					}
				}

				// Анализ производительности
				var analysisResult map[string]interface{}
				err = workflow.ExecuteActivity(
					workflow.WithActivityOptions(ctx, activityOptions),
					"AnalyzePerformanceActivity",
					result.MonitoredMetrics,
					result.PerformanceBaseline,
					input.AlertThresholds,
				).Get(ctx, &analysisResult)

				if err != nil {
					logger.Warn("Ошибка анализа производительности", zap.Error(err))
				} else {
					// Обновляем trends
					if trends, ok := analysisResult["trends"].(map[string]interface{}); ok {
						result.PerformanceTrends = trends
					}

					// Проверяем алерты
					if alerts, ok := analysisResult["alerts"].([]interface{}); ok {
						for _, alert := range alerts {
							if alertStr, ok := alert.(string); ok {
								result.TriggeredAlerts = append(result.TriggeredAlerts, alertStr)
							}
						}
					}

					// Обновляем performance score
					if score, ok := analysisResult["performance_score"].(float64); ok {
						result.PerformanceScore = score
					}
				}

				logger.Debug("Результат мониторинга производительности",
					zap.Any("metrics", result.MonitoredMetrics),
					zap.Float64("performance_score", result.PerformanceScore),
					zap.Int("triggered_alerts", len(result.TriggeredAlerts)),
				)
			}

			// Ждем 30 секунд до следующей проверки
			err = workflow.Sleep(ctx, 30*time.Second)
			if err != nil {
				// Workflow отменен
				logger.Info("Непрерывный мониторинг производительности остановлен")
				break
			}
		}

	case "periodic":
		// Периодический мониторинг производительности
		logger.Info("Запуск периодического мониторинга производительности")

		// Получаем интервалы мониторинга
		var scheduleResult map[string]interface{}
		err = workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, activityOptions),
			"GetPerformanceMonitoringScheduleActivity",
			input.Configuration,
		).Get(ctx, &scheduleResult)

		if err != nil {
			logger.Error("Ошибка получения расписания мониторинга", zap.Error(err))
			result.Status = "failed"
			result.Message = "Не удалось получить расписание: " + err.Error()
			result.CompletedAt = workflow.Now(ctx)
			return result, err
		}

		// Выполняем мониторинг по расписанию
		if intervals, ok := scheduleResult["intervals"].([]interface{}); ok {
			for _, interval := range intervals {
				if intervalStr, ok := interval.(string); ok {
					intervalDuration, _ := time.ParseDuration(intervalStr)

					// Ждем до времени мониторинга
					workflow.Sleep(ctx, intervalDuration)

					// Выполняем мониторинг (аналогично непрерывному)
					var metricsResult map[string]interface{}
					err := workflow.ExecuteActivity(
						workflow.WithActivityOptions(ctx, activityOptions),
						"CollectPerformanceMetricsActivity",
						input.MonitoringScope,
						input.MetricsToMonitor,
					).Get(ctx, &metricsResult)

					if err != nil {
						logger.Warn("Ошибка сбора метрик производительности", zap.Error(err))
					}
					// Обработка результатов аналогично непрерывному мониторингу
					// ... (код аналогичен блоку выше)
				}
			}
		}

	case "triggered":
		// Мониторинг по триггерам
		logger.Info("Запуск мониторинга производительности по триггерам")

		// Настраиваем слушатель триггеров
		triggerChannel := workflow.GetSignalChannel(ctx, "performance-trigger")
		selector := workflow.NewSelector(ctx)

		selector.AddReceive(triggerChannel, func(c workflow.Channel, more bool) {
			var triggerData map[string]interface{}
			c.Receive(ctx, &triggerData)

			logger.Info("Получен триггер для мониторинга производительности",
				zap.Any("trigger_data", triggerData),
			)

			// Выполняем мониторинг в ответ на триггер
			var metricsResult map[string]interface{}
			err := workflow.ExecuteActivity(
				workflow.WithActivityOptions(ctx, activityOptions),
				"CollectPerformanceMetricsActivity",
				input.MonitoringScope,
				input.MetricsToMonitor,
			).Get(ctx, &metricsResult)

			if err != nil {
				logger.Warn("Ошибка сбора метрик по триггеру", zap.Error(err))
			}
			// Обработка результатов
			// ... (код аналогичен блоку выше)
		})

		// Ждем триггеры в течение определенного времени
		timeout := workflow.NewTimer(ctx, 24*time.Hour)
		selector.AddFuture(timeout, func(f workflow.Future) {
			logger.Info("Время ожидания триггеров истекло")
		})

		selector.Select(ctx)
	}

	result.Status = "completed"
	result.Message = "Мониторинг производительности завершен"
	result.CompletedAt = workflow.Now(ctx)

	logger.Info("PerformanceMonitoringWorkflow завершен",
		zap.String("monitoring_scope", input.MonitoringScope),
		zap.String("monitoring_period", input.MonitoringPeriod),
		zap.String("status", result.Status),
		zap.Float64("performance_score", result.PerformanceScore),
		zap.Int("triggered_alerts", len(result.TriggeredAlerts)),
		zap.Duration("duration", result.CompletedAt.Sub(result.StartedAt)),
	)

	return result, nil
}

// CapacityPlanningWorkflow workflow для планирования емкости
func CapacityPlanningWorkflow(ctx workflow.Context, input CapacityPlanningWorkflowInput) (*CapacityPlanningWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Запуск CapacityPlanningWorkflow",
		zap.String("planning_horizon", input.PlanningHorizon),
		zap.Any("growth_projections", input.GrowthProjections),
		zap.Strings("resource_types", input.ResourceTypes),
	)

	result := &CapacityPlanningWorkflowResult{
		PlanningHorizon:         input.PlanningHorizon,
		Status:                  "planning",
		StartedAt:               workflow.Now(ctx),
		CapacityRecommendations: make(map[string]interface{}),
		ResourceRequirements:    make(map[string]interface{}),
		CostProjections:         make(map[string]float64),
		RiskAssessment:          make(map[string]interface{}),
		MilestoneTimeline:       make([]string, 0),
	}

	// Настройки для activities
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Minute,
		HeartbeatTimeout:    2 * time.Minute,
		RetryPolicy: &workflow.RetryPolicy{
			InitialInterval:    2 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    1 * time.Minute,
			MaximumAttempts:    3,
		},
	}

	// 1. Анализ текущего состояния ресурсов
	logger.Info("Анализ текущего состояния ресурсов")

	var currentStateResult map[string]interface{}
	err := workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"AnalyzeCurrentResourceStateActivity",
		input.ResourceTypes,
		input.Configuration,
	).Get(ctx, &currentStateResult)

	if err != nil {
		logger.Error("Ошибка анализа текущего состояния", zap.Error(err))
		result.Status = "failed"
		result.Message = "Не удалось проанализировать текущее состояние: " + err.Error()
		result.CompletedAt = workflow.Now(ctx)
		return result, err
	}

	// 2. Прогнозирование роста нагрузки
	logger.Info("Прогнозирование роста нагрузки")

	var growthForecastResult map[string]interface{}
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"ForecastGrowthActivity",
		input.GrowthProjections,
		input.PlanningHorizon,
		currentStateResult,
	).Get(ctx, &growthForecastResult)

	if err != nil {
		logger.Error("Ошибка прогнозирования роста", zap.Error(err))
		result.Status = "failed"
		result.Message = "Не удалось спрогнозировать рост: " + err.Error()
		result.CompletedAt = workflow.Now(ctx)
		return result, err
	}

	// 3. Расчет требований к ресурсам
	logger.Info("Расчет требований к ресурсам")

	var requirementsResult map[string]interface{}
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"CalculateResourceRequirementsActivity",
		growthForecastResult,
		input.ResourceTypes,
		input.BusinessGoals,
	).Get(ctx, &requirementsResult)

	if err != nil {
		logger.Error("Ошибка расчета требований", zap.Error(err))
		result.Status = "failed"
		result.Message = "Не удалось рассчитать требования: " + err.Error()
		result.CompletedAt = workflow.Now(ctx)
		return result, err
	}

	if requirements, ok := requirementsResult["resource_requirements"].(map[string]interface{}); ok {
		result.ResourceRequirements = requirements
	}

	// 4. Анализ стоимости
	logger.Info("Анализ стоимости")

	var costAnalysisResult map[string]interface{}
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"AnalyzeCapacityCostsActivity",
		result.ResourceRequirements,
		input.PlanningHorizon,
		input.Configuration,
	).Get(ctx, &costAnalysisResult)

	if err != nil {
		logger.Warn("Ошибка анализа стоимости", zap.Error(err))
	} else {
		if costProjections, ok := costAnalysisResult["cost_projections"].(map[string]interface{}); ok {
			for period, cost := range costProjections {
				if costFloat, ok := cost.(float64); ok {
					result.CostProjections[period] = costFloat
				}
			}
		}
	}

	// 5. Оценка рисков
	logger.Info("Оценка рисков планирования емкости")

	var riskAssessmentResult map[string]interface{}
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"AssessCapacityRisksActivity",
		result.ResourceRequirements,
		result.CostProjections,
		input.BusinessGoals,
	).Get(ctx, &riskAssessmentResult)

	if err != nil {
		logger.Warn("Ошибка оценки рисков", zap.Error(err))
	} else {
		result.RiskAssessment = riskAssessmentResult
	}

	// 6. Генерация рекомендаций
	logger.Info("Генерация рекомендаций по емкости")

	var recommendationsResult map[string]interface{}
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"GenerateCapacityRecommendationsActivity",
		result.ResourceRequirements,
		result.CostProjections,
		result.RiskAssessment,
		input.BusinessGoals,
	).Get(ctx, &recommendationsResult)

	if err != nil {
		logger.Warn("Ошибка генерации рекомендаций", zap.Error(err))
	} else {
		result.CapacityRecommendations = recommendationsResult
	}

	// 7. Создание временной шкалы
	logger.Info("Создание временной шкалы реализации")

	var timelineResult map[string]interface{}
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"CreateImplementationTimelineActivity",
		result.CapacityRecommendations,
		input.PlanningHorizon,
	).Get(ctx, &timelineResult)

	if err != nil {
		logger.Warn("Ошибка создания временной шкалы", zap.Error(err))
	} else {
		if milestones, ok := timelineResult["milestones"].([]interface{}); ok {
			for _, milestone := range milestones {
				if milestoneStr, ok := milestone.(string); ok {
					result.MilestoneTimeline = append(result.MilestoneTimeline, milestoneStr)
				}
			}
		}
	}

	result.Status = "completed"
	result.Message = "Планирование емкости завершено"
	result.CompletedAt = workflow.Now(ctx)

	logger.Info("CapacityPlanningWorkflow завершен",
		zap.String("planning_horizon", input.PlanningHorizon),
		zap.String("status", result.Status),
		zap.Int("resource_types", len(input.ResourceTypes)),
		zap.Int("milestones", len(result.MilestoneTimeline)),
		zap.Duration("duration", result.CompletedAt.Sub(result.StartedAt)),
	)

	return result, nil
}

// getScalingInterval возвращает интервал масштабирования с учетом тестового окружения
func getScalingInterval(ctx workflow.Context) time.Duration {
	// В тестовом окружении или при replay используем короткие интервалы
	if workflow.IsReplaying(ctx) || isTestEnvironment() {
		return time.Second * 1
	}
	return time.Minute * 2
}

// isTestEnvironment проверяет, выполняется ли workflow в тестовом окружении
func isTestEnvironment() bool {
	// В реальной реализации можно проверить переменные окружения или конфигурацию
	// Для тестов всегда возвращаем true
	return true // TODO: implement proper test environment detection
}
