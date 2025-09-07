package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	_ "net/http/pprof" //nosec G108 - профилирование для производительности
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/cadence"
	"github.com/twogc/cloudbridge-relay-installer/internal/cadence/activities"
	"github.com/twogc/cloudbridge-relay-installer/internal/cadence/workflows"
	"github.com/twogc/cloudbridge-relay-installer/internal/metrics"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// CadenceWorker представляет Cadence worker для CloudBridge Relay
type CadenceWorker struct {
	client         *cadence.Client
	worker         worker.Worker
	config         *cadence.CadenceConfig
	logger         *zap.Logger
	cadenceMetrics *metrics.CadenceMetrics

	// Activities
	commonActivities    *activities.CommonActivityContext
	wireguardActivities *activities.WireGuardActivities

	// Управление жизненным циклом
	ctx    context.Context
	cancel context.CancelFunc

	// Статус
	isRunning bool
}

// NewCadenceWorker создает новый Cadence worker
func NewCadenceWorker(client *cadence.Client, config *cadence.CadenceConfig, logger *zap.Logger, cadenceMetrics *metrics.CadenceMetrics) (*CadenceWorker, error) {
	if client == nil {
		return nil, fmt.Errorf("cadence клиент не может быть nil")
	}

	if config == nil {
		return nil, fmt.Errorf("конфигурация не может быть nil")
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	if cadenceMetrics == nil {
		cadenceMetrics = metrics.NewCadenceMetrics(logger)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Создаем activities
	activityConfig := map[string]interface{}{
		"domain":       config.Domain,
		"service_name": config.ServiceName,
		"environment":  "production", // Должно браться из конфигурации
	}

	commonActivities := activities.NewCommonActivityContext(logger, activityConfig)
	wireguardActivities := activities.NewWireGuardActivities(logger, activityConfig)

	return &CadenceWorker{
		client:              client,
		config:              config,
		logger:              logger,
		cadenceMetrics:      cadenceMetrics,
		commonActivities:    commonActivities,
		wireguardActivities: wireguardActivities,
		ctx:                 ctx,
		cancel:              cancel,
	}, nil
}

// Start запускает Cadence worker
func (cw *CadenceWorker) Start() error {
	cw.logger.Info("Запуск Cadence worker",
		zap.String("domain", cw.config.Domain),
		zap.String("task_list", cw.config.Worker.TaskListName),
		zap.String("identity", cw.config.Worker.Identity),
	)

	// Получаем worker от клиента (может быть nil для заглушек)
	cw.worker = cw.client.GetWorker()
	if cw.worker == nil {
		cw.logger.Info("Cadence worker не доступен (заглушка режим)")
		// В режиме заглушек продолжаем без реального worker
	}

	// Регистрируем workflows
	if err := cw.registerWorkflows(); err != nil {
		return fmt.Errorf("не удалось зарегистрировать workflows: %w", err)
	}

	// Регистрируем activities
	if err := cw.registerActivities(); err != nil {
		return fmt.Errorf("не удалось зарегистрировать activities: %w", err)
	}

	// Запускаем профилирование в debug режиме
	if err := cw.enableProfiling(); err != nil {
		cw.logger.Warn("Не удалось запустить профилирование", zap.Error(err))
	}

	// Запускаем health check endpoint
	if err := cw.enableHealthEndpoint(); err != nil {
		cw.logger.Warn("Не удалось запустить health endpoint", zap.Error(err))
	}

	// Запускаем worker в отдельной горутине (если доступен)
	if cw.worker != nil {
		go func() {
			cw.logger.Info("Запуск Cadence worker loop")
			if err := cw.worker.Start(); err != nil {
				cw.logger.Error("Ошибка при запуске Cadence worker", zap.Error(err))
			}
		}()
	} else {
		cw.logger.Info("Cadence worker loop пропущен (заглушка режим)")
	}

	cw.isRunning = true
	cw.logger.Info("Cadence worker успешно запущен")

	return nil
}

// Stop останавливает Cadence worker
func (cw *CadenceWorker) Stop() error {
	if !cw.isRunning {
		return nil
	}

	cw.logger.Info("Остановка Cadence worker")

	// Останавливаем worker
	if cw.worker != nil {
		cw.worker.Stop()
		cw.logger.Info("Cadence worker остановлен")
	}

	// Отменяем контекст
	cw.cancel()

	cw.isRunning = false
	cw.logger.Info("Cadence worker успешно остановлен")

	return nil
}

// IsRunning возвращает статус worker'а
func (cw *CadenceWorker) IsRunning() bool {
	return cw.isRunning
}

// registerWorkflows регистрирует все workflows
func (cw *CadenceWorker) registerWorkflows() error {
	cw.logger.Info("Регистрация Cadence workflows")

	// Основные workflows
	workflow.Register(workflows.CloudBridgeMainWorkflow)
	workflow.Register(workflows.ModuleInitializationWorkflow)
	workflow.Register(workflows.HealthCheckWorkflow)

	// WireGuard workflows
	workflow.Register(workflows.PeerLifecycleWorkflow)
	workflow.Register(workflows.MeshTopologyWorkflow)
	workflow.Register(workflows.ConnectionRecoveryWorkflow)
	workflow.Register(workflows.LoadBalancingWorkflow)

	// Quantum workflows
	workflow.Register(workflows.QuantumKeyRotationWorkflow)
	workflow.Register(workflows.KeyValidationWorkflow)
	workflow.Register(workflows.AlgorithmMigrationWorkflow)
	workflow.Register(workflows.SecurityAuditWorkflow)

	// AI workflows
	workflow.Register(workflows.ThreatAnalysisWorkflow)
	workflow.Register(workflows.BehaviorMonitoringWorkflow)
	workflow.Register(workflows.AutomatedResponseWorkflow)
	workflow.Register(workflows.ModelTrainingWorkflow)

	// Scaling workflows
	workflow.Register(workflows.AutoScalingWorkflow)
	workflow.Register(workflows.ResourceOptimizationWorkflow)
	workflow.Register(workflows.PerformanceMonitoringWorkflow)
	workflow.Register(workflows.CapacityPlanningWorkflow)

	// Security workflows
	workflow.Register(workflows.SecurityComplianceWorkflow)
	workflow.Register(workflows.AuditLoggingWorkflow)
	workflow.Register(workflows.IncidentResponseWorkflow)
	workflow.Register(workflows.BackupRecoveryWorkflow)

	cw.logger.Info("Все workflows зарегистрированы",
		zap.Int("workflow_count", 20), // Обновить при добавлении новых workflows
	)

	return nil
}

// registerActivities регистрирует все activities
func (cw *CadenceWorker) registerActivities() error {
	cw.logger.Info("Регистрация Cadence activities")

	// Common activities
	activity.Register(cw.commonActivities.CheckDependenciesActivity)
	activity.Register(cw.commonActivities.InitializeModuleActivity)
	activity.Register(cw.commonActivities.CheckModuleHealthActivity)
	activity.Register(cw.commonActivities.PerformHealthCheckActivity)
	activity.Register(cw.commonActivities.ValidateConfigurationActivity)
	activity.Register(cw.commonActivities.LogEventActivity)
	activity.Register(cw.commonActivities.SendNotificationActivity)
	activity.Register(cw.commonActivities.GetSystemMetricsActivity)
	activity.Register(cw.commonActivities.ExecuteCommandActivity)
	activity.Register(cw.commonActivities.WaitForConditionActivity)

	// WireGuard activities
	activity.Register(cw.wireguardActivities.ValidatePeerConfigActivity)
	activity.Register(cw.wireguardActivities.CreateWireGuardInterfaceActivity)
	activity.Register(cw.wireguardActivities.RegisterPeerActivity)
	activity.Register(cw.wireguardActivities.DisconnectPeerActivity)
	activity.Register(cw.wireguardActivities.RemoveWireGuardInterfaceActivity)
	activity.Register(cw.wireguardActivities.UnregisterPeerActivity)
	activity.Register(cw.wireguardActivities.UpdateWireGuardConfigActivity)
	activity.Register(cw.wireguardActivities.MonitorPeerActivity)
	activity.Register(cw.wireguardActivities.AnalyzeMeshTopologyActivity)
	activity.Register(cw.wireguardActivities.PlanMeshTopologyActivity)
	activity.Register(cw.wireguardActivities.BuildMeshTopologyActivity)
	activity.Register(cw.wireguardActivities.ValidateMeshTopologyActivity)

	// Здесь должны быть зарегистрированы activities для:
	// - Quantum криптографии
	// - AI/ML операций
	// - Автомасштабирования
	// - Безопасности
	// Пока регистрируем placeholder activities

	// Quantum activities (placeholder)
	activity.RegisterWithOptions(cw.createQuantumActivity("CheckKeyRotationNeedActivity"), activity.RegisterOptions{Name: "CheckKeyRotationNeedActivity"})
	activity.RegisterWithOptions(cw.createQuantumActivity("GenerateQuantumKeyActivity"), activity.RegisterOptions{Name: "GenerateQuantumKeyActivity"})
	activity.RegisterWithOptions(cw.createQuantumActivity("ValidateQuantumKeyActivity"), activity.RegisterOptions{Name: "ValidateQuantumKeyActivity"})
	activity.RegisterWithOptions(cw.createQuantumActivity("ReplaceQuantumKeyActivity"), activity.RegisterOptions{Name: "ReplaceQuantumKeyActivity"})
	activity.RegisterWithOptions(cw.createQuantumActivity("SecureDeleteQuantumKeyActivity"), activity.RegisterOptions{Name: "SecureDeleteQuantumKeyActivity"})
	activity.RegisterWithOptions(cw.createQuantumActivity("UpdateKeyReferencesActivity"), activity.RegisterOptions{Name: "UpdateKeyReferencesActivity"})

	// AI activities (placeholder)
	activity.RegisterWithOptions(cw.createAIActivity("CollectThreatDataActivity"), activity.RegisterOptions{Name: "CollectThreatDataActivity"})
	activity.RegisterWithOptions(cw.createAIActivity("PreprocessThreatDataActivity"), activity.RegisterOptions{Name: "PreprocessThreatDataActivity"})
	activity.RegisterWithOptions(cw.createAIActivity("PerformRealtimeThreatAnalysisActivity"), activity.RegisterOptions{Name: "PerformRealtimeThreatAnalysisActivity"})
	activity.RegisterWithOptions(cw.createAIActivity("AnalyzeUserBehaviorActivity"), activity.RegisterOptions{Name: "AnalyzeUserBehaviorActivity"})

	// Scaling activities (placeholder)
	activity.RegisterWithOptions(cw.createScalingActivity("CollectScalingMetricsActivity"), activity.RegisterOptions{Name: "CollectScalingMetricsActivity"})
	activity.RegisterWithOptions(cw.createScalingActivity("AnalyzeScalingNeedActivity"), activity.RegisterOptions{Name: "AnalyzeScalingNeedActivity"})
	activity.RegisterWithOptions(cw.createScalingActivity("PerformScalingActivity"), activity.RegisterOptions{Name: "PerformScalingActivity"})

	// Security activities (placeholder)
	activity.RegisterWithOptions(cw.createSecurityActivity("GetComplianceRequirementsActivity"), activity.RegisterOptions{Name: "GetComplianceRequirementsActivity"})
	activity.RegisterWithOptions(cw.createSecurityActivity("AuditSecurityAreaActivity"), activity.RegisterOptions{Name: "AuditSecurityAreaActivity"})
	activity.RegisterWithOptions(cw.createSecurityActivity("CalculateComplianceScoreActivity"), activity.RegisterOptions{Name: "CalculateComplianceScoreActivity"})

	cw.logger.Info("Все activities зарегистрированы",
		zap.Int("activity_count", 35), // Примерное количество
	)

	return nil
}

// createQuantumActivity создает placeholder activity для Quantum операций
func (cw *CadenceWorker) createQuantumActivity(activityName string) interface{} {
	return func(ctx context.Context, args ...interface{}) (map[string]interface{}, error) {
		cw.logger.Info("Выполнение Quantum activity",
			zap.String("activity", activityName),
			zap.Int("args_count", len(args)),
		)

		// Записываем метрику
		cw.cadenceMetrics.RecordActivityStarted(activityName, cw.config.Domain)

		// Симуляция выполнения quantum операции
		result := map[string]interface{}{
			"activity_name": activityName,
			"status":        "completed",
			"timestamp":     "2024-01-01T00:00:00Z", // Placeholder
			"quantum_safe":  true,
		}

		// Записываем завершение
		cw.cadenceMetrics.RecordActivityCompleted(activityName, cw.config.Domain, 0) // duration будет вычислен в реальной реализации

		return result, nil
	}
}

// createAIActivity создает placeholder activity для AI операций
func (cw *CadenceWorker) createAIActivity(activityName string) interface{} {
	return func(ctx context.Context, args ...interface{}) (map[string]interface{}, error) {
		cw.logger.Info("Выполнение AI activity",
			zap.String("activity", activityName),
			zap.Int("args_count", len(args)),
		)

		// Записываем метрику
		cw.cadenceMetrics.RecordActivityStarted(activityName, cw.config.Domain)

		// Симуляция выполнения AI операции
		result := map[string]interface{}{
			"activity_name":    activityName,
			"status":           "completed",
			"timestamp":        "2024-01-01T00:00:00Z", // Placeholder
			"ai_model_used":    "cloudbridge-ai-v1.0",
			"confidence_score": 0.95,
		}

		// Записываем завершение
		cw.cadenceMetrics.RecordActivityCompleted(activityName, cw.config.Domain, 0)

		return result, nil
	}
}

// createScalingActivity создает placeholder activity для Scaling операций
func (cw *CadenceWorker) createScalingActivity(activityName string) interface{} {
	return func(ctx context.Context, args ...interface{}) (map[string]interface{}, error) {
		cw.logger.Info("Выполнение Scaling activity",
			zap.String("activity", activityName),
			zap.Int("args_count", len(args)),
		)

		// Записываем метрику
		cw.cadenceMetrics.RecordActivityStarted(activityName, cw.config.Domain)

		// Симуляция выполнения scaling операции
		result := map[string]interface{}{
			"activity_name":    activityName,
			"status":           "completed",
			"timestamp":        "2024-01-01T00:00:00Z", // Placeholder
			"scaling_decision": "scale_up",
			"resource_delta":   "+2",
		}

		// Записываем завершение
		cw.cadenceMetrics.RecordActivityCompleted(activityName, cw.config.Domain, 0)

		return result, nil
	}
}

// createSecurityActivity создает placeholder activity для Security операций
func (cw *CadenceWorker) createSecurityActivity(activityName string) interface{} {
	return func(ctx context.Context, args ...interface{}) (map[string]interface{}, error) {
		cw.logger.Info("Выполнение Security activity",
			zap.String("activity", activityName),
			zap.Int("args_count", len(args)),
		)

		// Записываем метрику
		cw.cadenceMetrics.RecordActivityStarted(activityName, cw.config.Domain)

		// Симуляция выполнения security операции
		result := map[string]interface{}{
			"activity_name":  activityName,
			"status":         "completed",
			"timestamp":      "2024-01-01T00:00:00Z", // Placeholder
			"security_level": "high",
			"compliance_met": true,
		}

		// Записываем завершение
		cw.cadenceMetrics.RecordActivityCompleted(activityName, cw.config.Domain, 0)

		return result, nil
	}
}

// StartMainWorkflow запускает основной CloudBridge workflow
func (cw *CadenceWorker) StartMainWorkflow() error {
	if !cw.isRunning {
		return fmt.Errorf("cadence worker не запущен")
	}

	cw.logger.Info("Запуск основного CloudBridge workflow")

	// Подготавливаем входные данные для основного workflow
	input := workflows.CloudBridgeMainWorkflowInput{
		Domain:    cw.config.Domain,
		StartTime: time.Now(),
		Configuration: map[string]interface{}{
			"service_name":    cw.config.ServiceName,
			"worker_identity": cw.config.Worker.Identity,
			"task_list":       cw.config.Worker.TaskListName,
		},
	}

	// Запускаем workflow
	workflowID, err := cw.client.ExecuteWorkflow(cw.ctx, "CloudBridgeMainWorkflow", input)
	if err != nil {
		cw.logger.Error("Не удалось запустить основной workflow", zap.Error(err))
		return err
	}

	cw.logger.Info("Основной CloudBridge workflow запущен",
		zap.String("workflow_id", workflowID),
	)

	// Записываем метрику
	cw.cadenceMetrics.RecordWorkflowStarted("CloudBridgeMainWorkflow", cw.config.Domain, workflowID)

	return nil
}

// GetWorkflowStatus получает статус workflow
func (cw *CadenceWorker) GetWorkflowStatus(workflowID, runID string) (map[string]interface{}, error) {
	if !cw.isRunning {
		return nil, fmt.Errorf("cadence worker не запущен")
	}

	// Здесь должна быть логика получения статуса workflow
	// Пока возвращаем placeholder
	status := map[string]interface{}{
		"workflow_id": workflowID,
		"run_id":      runID,
		"status":      "running",
		"started_at":  time.Now().Format(time.RFC3339),
	}

	return status, nil
}

// GetWorkerStats получает статистику worker'а
func (cw *CadenceWorker) GetWorkerStats() map[string]interface{} {
	stats := map[string]interface{}{
		"is_running":           cw.isRunning,
		"domain":               cw.config.Domain,
		"task_list":            cw.config.Worker.TaskListName,
		"identity":             cw.config.Worker.Identity,
		"uptime":               time.Since(time.Now()).String(), // Placeholder
		"processed_workflows":  0,                               // Будет вычисляться в реальной реализации
		"processed_activities": 0,                               // Будет вычисляться в реальной реализации
	}

	return stats
}

// enableProfiling запускает pprof сервер для профилирования производительности
func (cw *CadenceWorker) enableProfiling() error {
	// Проверяем, включено ли профилирование в конфигурации
	if !cw.shouldEnableProfiling() {
		cw.logger.Debug("Профилирование отключено в конфигурации")
		return nil
	}

	profilingPort := cw.getProfilingPort()
	cw.logger.Info("Запуск pprof сервера для профилирования",
		zap.String("port", profilingPort),
		zap.String("endpoint", "/debug/pprof/"),
	)

	go func() {
		// Создаем отдельный HTTP сервер для профилирования
		mux := http.NewServeMux()

		// pprof endpoints уже автоматически регистрируются через import _ "net/http/pprof"
		// но мы можем добавить кастомные endpoints
		mux.HandleFunc("/debug/cadence/stats", cw.handleCadenceStats)
		mux.HandleFunc("/debug/cadence/workflows", cw.handleWorkflowsDebug)
		mux.HandleFunc("/debug/cadence/activities", cw.handleActivitiesDebug)

		server := &http.Server{
			Addr:    profilingPort,
			Handler: mux,
		}

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			cw.logger.Error("Ошибка pprof сервера", zap.Error(err))
		}
	}()

	return nil
}

// enableHealthEndpoint запускает health check endpoint
func (cw *CadenceWorker) enableHealthEndpoint() error {
	// Проверяем, включены ли health endpoints в конфигурации
	if !cw.shouldEnableHealth() {
		cw.logger.Debug("Health endpoints отключены в конфигурации")
		return nil
	}

	healthPort := cw.getHealthPort()
	cw.logger.Info("Запуск health check endpoint",
		zap.String("port", healthPort),
		zap.String("endpoint", "/health"),
	)

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/health", cw.handleHealthCheck)
		mux.HandleFunc("/readiness", cw.handleReadinessCheck)
		mux.HandleFunc("/liveness", cw.handleLivenessCheck)

		server := &http.Server{
			Addr:         healthPort,
			Handler:      mux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  15 * time.Second,
		}

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			cw.logger.Error("Ошибка health check сервера", zap.Error(err))
		}
	}()

	return nil
}

// shouldEnableProfiling проверяет, должно ли быть включено профилирование
func (cw *CadenceWorker) shouldEnableProfiling() bool {
	return cw.config.Profiling.Enabled
}

// getProfilingPort возвращает порт для профилирования
func (cw *CadenceWorker) getProfilingPort() string {
	return cw.config.Profiling.Port
}

// getHealthPort возвращает порт для health checks
func (cw *CadenceWorker) getHealthPort() string {
	return cw.config.Health.Port
}

// shouldEnableHealth проверяет, должны ли быть включены health endpoints
func (cw *CadenceWorker) shouldEnableHealth() bool {
	return cw.config.Health.Enabled
}

// handleCadenceStats обрабатывает запрос статистики Cadence
func (cw *CadenceWorker) handleCadenceStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	stats := cw.GetWorkerStats()

	// Добавляем дополнительную информацию о Cadence
	stats["cadence_version"] = "1.2.9"  // Версия из go.mod
	stats["registered_workflows"] = 20  // Количество зарегистрированных workflows
	stats["registered_activities"] = 35 // Количество зарегистрированных activities

	if err := json.NewEncoder(w).Encode(stats); err != nil {
		cw.logger.Error("Ошибка при кодировании статистики", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleWorkflowsDebug обрабатывает debug информацию о workflows
func (cw *CadenceWorker) handleWorkflowsDebug(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	workflows := map[string]interface{}{
		"main_workflows": []string{
			"CloudBridgeMainWorkflow",
			"ModuleInitializationWorkflow",
			"HealthCheckWorkflow",
		},
		"wireguard_workflows": []string{
			"PeerLifecycleWorkflow",
			"MeshTopologyWorkflow",
			"ConnectionRecoveryWorkflow",
			"LoadBalancingWorkflow",
		},
		"quantum_workflows": []string{
			"QuantumKeyRotationWorkflow",
			"KeyValidationWorkflow",
			"AlgorithmMigrationWorkflow",
			"SecurityAuditWorkflow",
		},
		"ai_workflows": []string{
			"ThreatAnalysisWorkflow",
			"BehaviorMonitoringWorkflow",
			"AutomatedResponseWorkflow",
			"ModelTrainingWorkflow",
		},
		"scaling_workflows": []string{
			"AutoScalingWorkflow",
			"ResourceOptimizationWorkflow",
			"PerformanceMonitoringWorkflow",
			"CapacityPlanningWorkflow",
		},
		"security_workflows": []string{
			"SecurityComplianceWorkflow",
			"AuditLoggingWorkflow",
			"IncidentResponseWorkflow",
			"BackupRecoveryWorkflow",
		},
	}

	if err := json.NewEncoder(w).Encode(workflows); err != nil {
		cw.logger.Error("Ошибка при кодировании workflows debug", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleActivitiesDebug обрабатывает debug информацию о activities
func (cw *CadenceWorker) handleActivitiesDebug(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	activities := map[string]interface{}{
		"common_activities": []string{
			"CheckDependenciesActivity",
			"InitializeModuleActivity",
			"CheckModuleHealthActivity",
			"PerformHealthCheckActivity",
		},
		"wireguard_activities": []string{
			"ValidatePeerConfigActivity",
			"CreateWireGuardInterfaceActivity",
			"RegisterPeerActivity",
			"DisconnectPeerActivity",
		},
		"quantum_activities": []string{
			"CheckKeyRotationNeedActivity",
			"GenerateQuantumKeyActivity",
			"ValidateQuantumKeyActivity",
			"ReplaceQuantumKeyActivity",
		},
		"ai_activities": []string{
			"CollectThreatDataActivity",
			"PreprocessThreatDataActivity",
			"PerformRealtimeThreatAnalysisActivity",
			"AnalyzeUserBehaviorActivity",
		},
	}

	if err := json.NewEncoder(w).Encode(activities); err != nil {
		cw.logger.Error("Ошибка при кодировании activities debug", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleHealthCheck обрабатывает общий health check
func (cw *CadenceWorker) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	health := map[string]interface{}{
		"status":         "healthy",
		"worker_running": cw.isRunning,
		"domain":         cw.config.Domain,
		"timestamp":      time.Now().Format(time.RFC3339),
		"checks": map[string]string{
			"cadence_worker": func() string {
				if cw.isRunning {
					return "healthy"
				}
				return "unhealthy"
			}(),
			"cadence_client": "healthy", // В реальной реализации нужна проверка соединения
		},
	}

	status := http.StatusOK
	if !cw.isRunning {
		status = http.StatusServiceUnavailable
		health["status"] = "unhealthy"
	}

	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(health); err != nil {
		cw.logger.Error("Ошибка при кодировании health check", zap.Error(err))
	}
}

// handleReadinessCheck обрабатывает readiness probe
func (cw *CadenceWorker) handleReadinessCheck(w http.ResponseWriter, r *http.Request) {
	if cw.isRunning && cw.worker != nil {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("ready")); err != nil {
			cw.logger.Error("Failed to write readiness response", zap.Error(err))
		}
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		if _, err := w.Write([]byte("not ready")); err != nil {
			cw.logger.Error("Failed to write readiness unavailable response", zap.Error(err))
		}
	}
}

// handleLivenessCheck обрабатывает liveness probe
func (cw *CadenceWorker) handleLivenessCheck(w http.ResponseWriter, r *http.Request) {
	// Простая проверка - если процесс отвечает, значит он живой
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("alive")); err != nil {
		cw.logger.Error("Failed to write liveness response", zap.Error(err))
	}
}
