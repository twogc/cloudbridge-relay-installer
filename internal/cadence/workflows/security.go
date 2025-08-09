package workflows

import (
	"time"

	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// SecurityComplianceWorkflowInput входные параметры для проверки соответствия безопасности
type SecurityComplianceWorkflowInput struct {
	ComplianceFramework string                 `json:"compliance_framework"` // iso27001, soc2, gdpr, pci_dss, nist
	AuditScope          []string               `json:"audit_scope"`          // network, data, access, infrastructure
	ComplianceLevel     string                 `json:"compliance_level"`     // basic, standard, advanced, critical
	Configuration       map[string]interface{} `json:"configuration"`
}

// SecurityComplianceWorkflowResult результат проверки соответствия безопасности
type SecurityComplianceWorkflowResult struct {
	ComplianceFramework string                 `json:"compliance_framework"`
	ComplianceScore     float64                `json:"compliance_score"`
	PassedControls      []string               `json:"passed_controls"`
	FailedControls      []string               `json:"failed_controls"`
	ComplianceResults   map[string]interface{} `json:"compliance_results"`
	Recommendations     []string               `json:"recommendations"`
	Status              string                 `json:"status"`
	StartedAt           time.Time              `json:"started_at"`
	CompletedAt         time.Time              `json:"completed_at"`
	Message             string                 `json:"message"`
}

// AuditLoggingWorkflowInput входные параметры для аудита логирования
type AuditLoggingWorkflowInput struct {
	LogSources        []string               `json:"log_sources"`        // system, application, security, network
	RetentionPeriod   string                 `json:"retention_period"`   // 1y, 7y, permanent
	AuditLevel        string                 `json:"audit_level"`        // minimal, standard, comprehensive
	AnalysisType      string                 `json:"analysis_type"`      // realtime, batch, forensic
	Configuration     map[string]interface{} `json:"configuration"`
}

// AuditLoggingWorkflowResult результат аудита логирования
type AuditLoggingWorkflowResult struct {
	LogSources         []string               `json:"log_sources"`
	ProcessedLogs      int64                  `json:"processed_logs"`
	AuditFindings      []string               `json:"audit_findings"`
	SecurityEvents     []string               `json:"security_events"`
	ComplianceStatus   string                 `json:"compliance_status"`
	LoggingResults     map[string]interface{} `json:"logging_results"`
	Status             string                 `json:"status"`
	StartedAt          time.Time              `json:"started_at"`
	CompletedAt        time.Time              `json:"completed_at"`
	Message            string                 `json:"message"`
}

// IncidentResponseWorkflowInput входные параметры для реагирования на инциденты
type IncidentResponseWorkflowInput struct {
	IncidentType      string                 `json:"incident_type"`      // security_breach, data_leak, ddos, malware
	Severity          string                 `json:"severity"`           // low, medium, high, critical
	ResponsePlan      string                 `json:"response_plan"`      // automated, manual, hybrid
	AffectedSystems   []string               `json:"affected_systems"`
	Configuration     map[string]interface{} `json:"configuration"`
}

// IncidentResponseWorkflowResult результат реагирования на инциденты
type IncidentResponseWorkflowResult struct {
	IncidentType       string                 `json:"incident_type"`
	Severity           string                 `json:"severity"`
	ResponseActions    []string               `json:"response_actions"`
	ContainmentStatus  string                 `json:"containment_status"`
	RecoveryStatus     string                 `json:"recovery_status"`
	IncidentResults    map[string]interface{} `json:"incident_results"`
	LessonsLearned     []string               `json:"lessons_learned"`
	Status             string                 `json:"status"`
	StartedAt          time.Time              `json:"started_at"`
	CompletedAt        time.Time              `json:"completed_at"`
	Message            string                 `json:"message"`
}

// BackupRecoveryWorkflowInput входные параметры для резервного копирования и восстановления
type BackupRecoveryWorkflowInput struct {
	BackupType        string                 `json:"backup_type"`        // full, incremental, differential, snapshot
	BackupScope       []string               `json:"backup_scope"`       // databases, files, configurations, keys
	ScheduleType      string                 `json:"schedule_type"`      // immediate, scheduled, continuous
	RetentionPolicy   string                 `json:"retention_policy"`   // 30d, 1y, 7y, permanent
	Configuration     map[string]interface{} `json:"configuration"`
}

// BackupRecoveryWorkflowResult результат резервного копирования и восстановления
type BackupRecoveryWorkflowResult struct {
	BackupType        string                 `json:"backup_type"`
	BackupSize        int64                  `json:"backup_size"`
	BackupLocation    string                 `json:"backup_location"`
	BackupStatus      string                 `json:"backup_status"`
	RecoveryTestStatus string                `json:"recovery_test_status"`
	BackupResults     map[string]interface{} `json:"backup_results"`
	Status            string                 `json:"status"`
	StartedAt         time.Time              `json:"started_at"`
	CompletedAt       time.Time              `json:"completed_at"`
	Message           string                 `json:"message"`
}

// SecurityComplianceWorkflow workflow для проверки соответствия требованиям безопасности
func SecurityComplianceWorkflow(ctx workflow.Context, input SecurityComplianceWorkflowInput) (*SecurityComplianceWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Запуск SecurityComplianceWorkflow",
		zap.String("compliance_framework", input.ComplianceFramework),
		zap.Strings("audit_scope", input.AuditScope),
		zap.String("compliance_level", input.ComplianceLevel),
	)

	result := &SecurityComplianceWorkflowResult{
		ComplianceFramework: input.ComplianceFramework,
		Status:              "auditing",
		StartedAt:           workflow.Now(ctx),
		PassedControls:      make([]string, 0),
		FailedControls:      make([]string, 0),
		ComplianceResults:   make(map[string]interface{}),
		Recommendations:     make([]string, 0),
	}

	// Настройки для activities
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Minute,
		HeartbeatTimeout:    3 * time.Minute,
		RetryPolicy: &workflow.RetryPolicy{
			InitialInterval:    2 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    1 * time.Minute,
			MaximumAttempts:    2,
		},
	}

	// 1. Получение требований соответствия для выбранного фреймворка
	logger.Info("Получение требований соответствия")
	
	var frameworkRequirements map[string]interface{}
	err := workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"GetComplianceRequirementsActivity",
		input.ComplianceFramework,
		input.ComplianceLevel,
		input.Configuration,
	).Get(ctx, &frameworkRequirements)

	if err != nil {
		logger.Error("Ошибка получения требований соответствия", zap.Error(err))
		result.Status = "failed"
		result.Message = "Не удалось получить требования соответствия: " + err.Error()
		result.CompletedAt = workflow.Now(ctx)
		return result, err
	}

	result.ComplianceResults["framework_requirements"] = frameworkRequirements

	// 2. Аудит каждой области безопасности
	for _, auditArea := range input.AuditScope {
		logger.Info("Аудит области безопасности", zap.String("area", auditArea))
		
		var areaAuditResult map[string]interface{}
		err := workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, activityOptions),
			"AuditSecurityAreaActivity",
			auditArea,
			input.ComplianceFramework,
			frameworkRequirements,
			input.Configuration,
		).Get(ctx, &areaAuditResult)

		if err != nil {
			logger.Warn("Ошибка аудита области",
				zap.String("area", auditArea),
				zap.Error(err),
			)
			// Продолжаем аудит других областей
		} else {
			result.ComplianceResults[auditArea] = areaAuditResult
			
			// Извлекаем пройденные и непройденные контроли
			if passedControls, ok := areaAuditResult["passed_controls"].([]interface{}); ok {
				for _, control := range passedControls {
					if controlStr, ok := control.(string); ok {
						result.PassedControls = append(result.PassedControls, controlStr)
					}
				}
			}
			
			if failedControls, ok := areaAuditResult["failed_controls"].([]interface{}); ok {
				for _, control := range failedControls {
					if controlStr, ok := control.(string); ok {
						result.FailedControls = append(result.FailedControls, controlStr)
					}
				}
			}
			
			logger.Info("Аудит области завершен",
				zap.String("area", auditArea),
				zap.Any("result", areaAuditResult),
			)
		}
	}

	// 3. Расчет общего индекса соответствия
	logger.Info("Расчет индекса соответствия")
	
	var complianceScoreResult map[string]interface{}
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"CalculateComplianceScoreActivity",
		result.PassedControls,
		result.FailedControls,
		input.ComplianceFramework,
		result.ComplianceResults,
	).Get(ctx, &complianceScoreResult)

	if err != nil {
		logger.Warn("Ошибка расчета индекса соответствия", zap.Error(err))
		result.ComplianceScore = 0.0
	} else {
		if score, ok := complianceScoreResult["compliance_score"].(float64); ok {
			result.ComplianceScore = score
		}
		result.ComplianceResults["score_calculation"] = complianceScoreResult
	}

	// 4. Генерация рекомендаций по улучшению соответствия
	logger.Info("Генерация рекомендаций")
	
	var recommendationsResult map[string]interface{}
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"GenerateComplianceRecommendationsActivity",
		result.FailedControls,
		result.ComplianceScore,
		input.ComplianceFramework,
		result.ComplianceResults,
	).Get(ctx, &recommendationsResult)

	if err != nil {
		logger.Warn("Ошибка генерации рекомендаций", zap.Error(err))
	} else {
		if recommendations, ok := recommendationsResult["recommendations"].([]interface{}); ok {
			for _, rec := range recommendations {
				if recStr, ok := rec.(string); ok {
					result.Recommendations = append(result.Recommendations, recStr)
				}
			}
		}
		result.ComplianceResults["recommendations"] = recommendationsResult
	}

	// 5. Создание отчета о соответствии
	logger.Info("Создание отчета о соответствии")
	
	var reportResult map[string]interface{}
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"GenerateComplianceReportActivity",
		input.ComplianceFramework,
		result.ComplianceScore,
		result.PassedControls,
		result.FailedControls,
		result.Recommendations,
	).Get(ctx, &reportResult)

	if err != nil {
		logger.Warn("Ошибка создания отчета", zap.Error(err))
	} else {
		result.ComplianceResults["report"] = reportResult
	}

	// Определяем итоговый статус
	if result.ComplianceScore >= 0.9 {
		result.Status = "compliant"
		result.Message = "Высокий уровень соответствия"
	} else if result.ComplianceScore >= 0.7 {
		result.Status = "mostly_compliant"
		result.Message = "Большинство требований выполнено"
	} else if result.ComplianceScore >= 0.5 {
		result.Status = "partially_compliant"
		result.Message = "Частичное соответствие требованиям"
	} else {
		result.Status = "non_compliant"
		result.Message = "Низкий уровень соответствия"
	}

	result.CompletedAt = workflow.Now(ctx)

	logger.Info("SecurityComplianceWorkflow завершен",
		zap.String("compliance_framework", input.ComplianceFramework),
		zap.String("status", result.Status),
		zap.Float64("compliance_score", result.ComplianceScore),
		zap.Int("passed_controls", len(result.PassedControls)),
		zap.Int("failed_controls", len(result.FailedControls)),
		zap.Int("recommendations", len(result.Recommendations)),
		zap.Duration("duration", result.CompletedAt.Sub(result.StartedAt)),
	)

	return result, nil
}

// AuditLoggingWorkflow workflow для аудита системы логирования
func AuditLoggingWorkflow(ctx workflow.Context, input AuditLoggingWorkflowInput) (*AuditLoggingWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Запуск AuditLoggingWorkflow",
		zap.Strings("log_sources", input.LogSources),
		zap.String("retention_period", input.RetentionPeriod),
		zap.String("audit_level", input.AuditLevel),
		zap.String("analysis_type", input.AnalysisType),
	)

	result := &AuditLoggingWorkflowResult{
		LogSources:       input.LogSources,
		Status:           "auditing",
		StartedAt:        workflow.Now(ctx),
		AuditFindings:    make([]string, 0),
		SecurityEvents:   make([]string, 0),
		LoggingResults:   make(map[string]interface{}),
	}

	// Настройки для activities
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 25 * time.Minute,
		HeartbeatTimeout:    2 * time.Minute,
		RetryPolicy: &workflow.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    1 * time.Minute,
			MaximumAttempts:    3,
		},
	}

	// 1. Сбор логов из всех источников
	logger.Info("Сбор логов из источников")
	
	var logCollectionResult map[string]interface{}
	err := workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"CollectAuditLogsActivity",
		input.LogSources,
		input.RetentionPeriod,
		input.Configuration,
	).Get(ctx, &logCollectionResult)

	if err != nil {
		logger.Error("Ошибка сбора логов", zap.Error(err))
		result.Status = "failed"
		result.Message = "Не удалось собрать логи: " + err.Error()
		result.CompletedAt = workflow.Now(ctx)
		return result, err
	}

	if processedLogs, ok := logCollectionResult["processed_logs"].(float64); ok {
		result.ProcessedLogs = int64(processedLogs)
	}
	result.LoggingResults["log_collection"] = logCollectionResult

	// 2. Анализ логов в зависимости от типа анализа
	switch input.AnalysisType {
	case "realtime":
		// Реалтайм анализ логов
		logger.Info("Реалтайм анализ логов")
		
		var realtimeResult map[string]interface{}
		err = workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, activityOptions),
			"PerformRealtimeLogAnalysisActivity",
			logCollectionResult,
			input.AuditLevel,
		).Get(ctx, &realtimeResult)

		if err != nil {
			logger.Warn("Ошибка реалтайм анализа", zap.Error(err))
		} else {
			result.LoggingResults["realtime_analysis"] = realtimeResult
		}

	case "batch":
		// Пакетный анализ логов
		logger.Info("Пакетный анализ логов")
		
		var batchResult map[string]interface{}
		err = workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, activityOptions),
			"PerformBatchLogAnalysisActivity",
			logCollectionResult,
			input.AuditLevel,
		).Get(ctx, &batchResult)

		if err != nil {
			logger.Warn("Ошибка пакетного анализа", zap.Error(err))
		} else {
			result.LoggingResults["batch_analysis"] = batchResult
		}

	case "forensic":
		// Форензический анализ логов
		logger.Info("Форензический анализ логов")
		
		var forensicResult map[string]interface{}
		err = workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, activityOptions),
			"PerformForensicLogAnalysisActivity",
			logCollectionResult,
			input.AuditLevel,
		).Get(ctx, &forensicResult)

		if err != nil {
			logger.Warn("Ошибка форензического анализа", zap.Error(err))
		} else {
			result.LoggingResults["forensic_analysis"] = forensicResult
		}
	}

	// 3. Извлечение событий безопасности
	logger.Info("Извлечение событий безопасности")
	
	var securityEventsResult map[string]interface{}
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"ExtractSecurityEventsActivity",
		result.LoggingResults,
		input.AuditLevel,
	).Get(ctx, &securityEventsResult)

	if err != nil {
		logger.Warn("Ошибка извлечения событий безопасности", zap.Error(err))
	} else {
		if events, ok := securityEventsResult["security_events"].([]interface{}); ok {
			for _, event := range events {
				if eventStr, ok := event.(string); ok {
					result.SecurityEvents = append(result.SecurityEvents, eventStr)
				}
			}
		}
		result.LoggingResults["security_events"] = securityEventsResult
	}

	// 4. Аудит соответствия требованиям логирования
	logger.Info("Аудит соответствия требованиям логирования")
	
	var complianceAuditResult map[string]interface{}
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"AuditLoggingComplianceActivity",
		result.LoggingResults,
		input.RetentionPeriod,
		input.AuditLevel,
	).Get(ctx, &complianceAuditResult)

	if err != nil {
		logger.Warn("Ошибка аудита соответствия", zap.Error(err))
		result.ComplianceStatus = "unknown"
	} else {
		if status, ok := complianceAuditResult["compliance_status"].(string); ok {
			result.ComplianceStatus = status
		}
		
		if findings, ok := complianceAuditResult["audit_findings"].([]interface{}); ok {
			for _, finding := range findings {
				if findingStr, ok := finding.(string); ok {
					result.AuditFindings = append(result.AuditFindings, findingStr)
				}
			}
		}
		
		result.LoggingResults["compliance_audit"] = complianceAuditResult
	}

	// 5. Генерация отчета аудита
	logger.Info("Генерация отчета аудита логирования")
	
	var reportResult map[string]interface{}
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"GenerateLoggingAuditReportActivity",
		result.ProcessedLogs,
		result.SecurityEvents,
		result.AuditFindings,
		result.ComplianceStatus,
	).Get(ctx, &reportResult)

	if err != nil {
		logger.Warn("Ошибка генерации отчета", zap.Error(err))
	} else {
		result.LoggingResults["audit_report"] = reportResult
	}

	result.Status = "completed"
	result.Message = "Аудит логирования завершен"
	result.CompletedAt = workflow.Now(ctx)

	logger.Info("AuditLoggingWorkflow завершен",
		zap.String("analysis_type", input.AnalysisType),
		zap.String("status", result.Status),
		zap.String("compliance_status", result.ComplianceStatus),
		zap.Int64("processed_logs", result.ProcessedLogs),
		zap.Int("security_events", len(result.SecurityEvents)),
		zap.Int("audit_findings", len(result.AuditFindings)),
		zap.Duration("duration", result.CompletedAt.Sub(result.StartedAt)),
	)

	return result, nil
}

// IncidentResponseWorkflow workflow для реагирования на инциденты безопасности
func IncidentResponseWorkflow(ctx workflow.Context, input IncidentResponseWorkflowInput) (*IncidentResponseWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Запуск IncidentResponseWorkflow",
		zap.String("incident_type", input.IncidentType),
		zap.String("severity", input.Severity),
		zap.String("response_plan", input.ResponsePlan),
		zap.Strings("affected_systems", input.AffectedSystems),
	)

	result := &IncidentResponseWorkflowResult{
		IncidentType:    input.IncidentType,
		Severity:        input.Severity,
		Status:          "responding",
		StartedAt:       workflow.Now(ctx),
		ResponseActions: make([]string, 0),
		IncidentResults: make(map[string]interface{}),
		LessonsLearned:  make([]string, 0),
	}

	// Настройки для activities (инциденты требуют быстрого реагирования)
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 15 * time.Minute,
		HeartbeatTimeout:    1 * time.Minute,
		RetryPolicy: &workflow.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 1.5,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    3,
		},
	}

	// 1. Анализ инцидента и определение плана реагирования
	logger.Info("Анализ инцидента")
	
	var incidentAnalysisResult map[string]interface{}
	err := workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"AnalyzeIncidentActivity",
		input.IncidentType,
		input.Severity,
		input.AffectedSystems,
		input.Configuration,
	).Get(ctx, &incidentAnalysisResult)

	if err != nil {
		logger.Error("Ошибка анализа инцидента", zap.Error(err))
		result.Status = "failed"
		result.Message = "Не удалось проанализировать инцидент: " + err.Error()
		result.CompletedAt = workflow.Now(ctx)
		return result, err
	}

	result.IncidentResults["incident_analysis"] = incidentAnalysisResult

	// 2. Фаза сдерживания (Containment)
	logger.Info("Фаза сдерживания инцидента")
	
	var containmentResult map[string]interface{}
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"ContainIncidentActivity",
		input.IncidentType,
		input.Severity,
		input.AffectedSystems,
		input.ResponsePlan,
	).Get(ctx, &containmentResult)

	if err != nil {
		logger.Error("Ошибка сдерживания инцидента", zap.Error(err))
		result.ContainmentStatus = "failed"
	} else {
		if status, ok := containmentResult["containment_status"].(string); ok {
			result.ContainmentStatus = status
		}
		
		if actions, ok := containmentResult["containment_actions"].([]interface{}); ok {
			for _, action := range actions {
				if actionStr, ok := action.(string); ok {
					result.ResponseActions = append(result.ResponseActions, actionStr)
				}
			}
		}
		
		result.IncidentResults["containment"] = containmentResult
		
		logger.Info("Сдерживание инцидента завершено",
			zap.String("status", result.ContainmentStatus),
		)
	}

	// 3. Фаза искоренения (Eradication)
	if result.ContainmentStatus == "successful" || result.ContainmentStatus == "partial" {
		logger.Info("Фаза искоренения инцидента")
		
		var eradicationResult map[string]interface{}
		err = workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, activityOptions),
			"EradicateIncidentActivity",
			input.IncidentType,
			containmentResult,
			input.AffectedSystems,
		).Get(ctx, &eradicationResult)

		if err != nil {
			logger.Warn("Ошибка искоренения инцидента", zap.Error(err))
		} else {
			if actions, ok := eradicationResult["eradication_actions"].([]interface{}); ok {
				for _, action := range actions {
					if actionStr, ok := action.(string); ok {
						result.ResponseActions = append(result.ResponseActions, actionStr)
					}
				}
			}
			
			result.IncidentResults["eradication"] = eradicationResult
			
			logger.Info("Искоренение инцидента завершено")
		}
	}

	// 4. Фаза восстановления (Recovery)
	if result.ContainmentStatus != "failed" {
		logger.Info("Фаза восстановления после инцидента")
		
		var recoveryResult map[string]interface{}
		err = workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, activityOptions),
			"RecoverFromIncidentActivity",
			input.IncidentType,
			input.AffectedSystems,
			result.IncidentResults,
		).Get(ctx, &recoveryResult)

		if err != nil {
			logger.Warn("Ошибка восстановления", zap.Error(err))
			result.RecoveryStatus = "failed"
		} else {
			if status, ok := recoveryResult["recovery_status"].(string); ok {
				result.RecoveryStatus = status
			}
			
			if actions, ok := recoveryResult["recovery_actions"].([]interface{}); ok {
				for _, action := range actions {
					if actionStr, ok := action.(string); ok {
						result.ResponseActions = append(result.ResponseActions, actionStr)
					}
				}
			}
			
			result.IncidentResults["recovery"] = recoveryResult
			
			logger.Info("Восстановление после инцидента завершено",
				zap.String("status", result.RecoveryStatus),
			)
		}
	}

	// 5. Фаза извлечения уроков (Lessons Learned)
	logger.Info("Анализ уроков и создание рекомендаций")
	
	var lessonsLearnedResult map[string]interface{}
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"AnalyzeLessonsLearnedActivity",
		input.IncidentType,
		input.Severity,
		result.IncidentResults,
		result.ResponseActions,
	).Get(ctx, &lessonsLearnedResult)

	if err != nil {
		logger.Warn("Ошибка анализа уроков", zap.Error(err))
	} else {
		if lessons, ok := lessonsLearnedResult["lessons_learned"].([]interface{}); ok {
			for _, lesson := range lessons {
				if lessonStr, ok := lesson.(string); ok {
					result.LessonsLearned = append(result.LessonsLearned, lessonStr)
				}
			}
		}
		
		result.IncidentResults["lessons_learned"] = lessonsLearnedResult
	}

	// 6. Создание отчета об инциденте
	logger.Info("Создание отчета об инциденте")
	
	var reportResult map[string]interface{}
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"GenerateIncidentReportActivity",
		input.IncidentType,
		input.Severity,
		result.ContainmentStatus,
		result.RecoveryStatus,
		result.ResponseActions,
		result.LessonsLearned,
	).Get(ctx, &reportResult)

	if err != nil {
		logger.Warn("Ошибка создания отчета", zap.Error(err))
	} else {
		result.IncidentResults["incident_report"] = reportResult
	}

	// Определяем итоговый статус
	if result.ContainmentStatus == "successful" && result.RecoveryStatus == "successful" {
		result.Status = "resolved"
		result.Message = "Инцидент успешно разрешен"
	} else if result.ContainmentStatus == "successful" || result.ContainmentStatus == "partial" {
		result.Status = "contained"
		result.Message = "Инцидент сдержан, восстановление в процессе"
	} else {
		result.Status = "active"
		result.Message = "Инцидент требует дальнейшего реагирования"
	}

	result.CompletedAt = workflow.Now(ctx)

	logger.Info("IncidentResponseWorkflow завершен",
		zap.String("incident_type", input.IncidentType),
		zap.String("severity", input.Severity),
		zap.String("status", result.Status),
		zap.String("containment_status", result.ContainmentStatus),
		zap.String("recovery_status", result.RecoveryStatus),
		zap.Int("response_actions", len(result.ResponseActions)),
		zap.Int("lessons_learned", len(result.LessonsLearned)),
		zap.Duration("duration", result.CompletedAt.Sub(result.StartedAt)),
	)

	return result, nil
}

// BackupRecoveryWorkflow workflow для резервного копирования и восстановления
func BackupRecoveryWorkflow(ctx workflow.Context, input BackupRecoveryWorkflowInput) (*BackupRecoveryWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Запуск BackupRecoveryWorkflow",
		zap.String("backup_type", input.BackupType),
		zap.Strings("backup_scope", input.BackupScope),
		zap.String("schedule_type", input.ScheduleType),
		zap.String("retention_policy", input.RetentionPolicy),
	)

	result := &BackupRecoveryWorkflowResult{
		BackupType:    input.BackupType,
		Status:        "backing_up",
		StartedAt:     workflow.Now(ctx),
		BackupResults: make(map[string]interface{}),
	}

	// Настройки для activities (бэкапы могут занимать много времени)
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Hour,
		HeartbeatTimeout:    5 * time.Minute,
		RetryPolicy: &workflow.RetryPolicy{
			InitialInterval:    5 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    2 * time.Minute,
			MaximumAttempts:    2,
		},
	}

	switch input.ScheduleType {
	case "immediate":
		// Немедленное выполнение резервного копирования
		logger.Info("Немедленное выполнение резервного копирования")
		
		// 1. Подготовка к резервному копированию
		var preparationResult map[string]interface{}
		err := workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, activityOptions),
			"PrepareBackupActivity",
			input.BackupType,
			input.BackupScope,
			input.Configuration,
		).Get(ctx, &preparationResult)

		if err != nil {
			logger.Error("Ошибка подготовки резервного копирования", zap.Error(err))
			result.Status = "failed"
			result.Message = "Не удалось подготовить резервное копирование: " + err.Error()
			result.CompletedAt = workflow.Now(ctx)
			return result, err
		}

		result.BackupResults["preparation"] = preparationResult

		// 2. Выполнение резервного копирования
		var backupResult map[string]interface{}
		err = workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, activityOptions),
			"PerformBackupActivity",
			input.BackupType,
			input.BackupScope,
			preparationResult,
		).Get(ctx, &backupResult)

		if err != nil {
			logger.Error("Ошибка выполнения резервного копирования", zap.Error(err))
			result.Status = "failed"
			result.Message = "Не удалось выполнить резервное копирование: " + err.Error()
			result.CompletedAt = workflow.Now(ctx)
			return result, err
		}

		// Извлекаем информацию о резервной копии
		if backupSize, ok := backupResult["backup_size"].(float64); ok {
			result.BackupSize = int64(backupSize)
		}
		
		if backupLocation, ok := backupResult["backup_location"].(string); ok {
			result.BackupLocation = backupLocation
		}
		
		if backupStatus, ok := backupResult["backup_status"].(string); ok {
			result.BackupStatus = backupStatus
		}

		result.BackupResults["backup_execution"] = backupResult

	case "scheduled":
		// Запланированное резервное копирование
		logger.Info("Запланированное резервное копирование")
		
		// Получаем расписание резервного копирования
		var scheduleResult map[string]interface{}
		err := workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, activityOptions),
			"GetBackupScheduleActivity",
			input.Configuration,
		).Get(ctx, &scheduleResult)

		if err != nil {
			logger.Error("Ошибка получения расписания", zap.Error(err))
			result.Status = "failed"
			result.Message = "Не удалось получить расписание: " + err.Error()
			result.CompletedAt = workflow.Now(ctx)
			return result, err
		}

		// Ждем времени выполнения
		if scheduleTime, ok := scheduleResult["next_backup_time"].(string); ok {
			scheduledTime, _ := time.Parse(time.RFC3339, scheduleTime)
			waitDuration := scheduledTime.Sub(workflow.Now(ctx))
			
			if waitDuration > 0 {
				logger.Info("Ожидание времени выполнения резервного копирования",
					zap.Duration("wait_duration", waitDuration),
				)
				workflow.Sleep(ctx, waitDuration)
			}
		}

		// Выполняем резервное копирование (аналогично immediate)
		// ... (код аналогичен блоку immediate выше)

	case "continuous":
		// Непрерывное резервное копирование
		logger.Info("Непрерывное резервное копирование")
		
		for {
			// Выполняем инкрементальное резервное копирование
			var incrementalResult map[string]interface{}
			err := workflow.ExecuteActivity(
				workflow.WithActivityOptions(ctx, activityOptions),
				"PerformIncrementalBackupActivity",
				input.BackupScope,
				input.Configuration,
			).Get(ctx, &incrementalResult)

			if err != nil {
				logger.Warn("Ошибка инкрементального резервного копирования", zap.Error(err))
			} else {
				// Обновляем результаты
				if backupSize, ok := incrementalResult["backup_size"].(float64); ok {
					result.BackupSize += int64(backupSize)
				}
				
				result.BackupResults["latest_incremental"] = incrementalResult
				
				logger.Debug("Инкрементальное резервное копирование выполнено",
					zap.Any("result", incrementalResult),
				)
			}

			// Ждем интервал между резервными копиями (например, 1 час)
			err = workflow.Sleep(ctx, 1*time.Hour)
			if err != nil {
				// Workflow отменен
				logger.Info("Непрерывное резервное копирование остановлено")
				break
			}
		}
	}

	// 3. Проверка целостности резервной копии
	if result.BackupStatus == "successful" || result.BackupStatus == "completed" {
		logger.Info("Проверка целостности резервной копии")
		
		var integrityResult map[string]interface{}
		err := workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, activityOptions),
			"VerifyBackupIntegrityActivity",
			result.BackupLocation,
			input.BackupType,
		).Get(ctx, &integrityResult)

		if err != nil {
			logger.Warn("Ошибка проверки целостности", zap.Error(err))
		} else {
			result.BackupResults["integrity_check"] = integrityResult
		}
	}

	// 4. Тест восстановления
	if result.BackupStatus == "successful" || result.BackupStatus == "completed" {
		logger.Info("Тест восстановления из резервной копии")
		
		var recoveryTestResult map[string]interface{}
		err := workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, activityOptions),
			"TestBackupRecoveryActivity",
			result.BackupLocation,
			input.BackupType,
			input.BackupScope,
		).Get(ctx, &recoveryTestResult)

		if err != nil {
			logger.Warn("Ошибка теста восстановления", zap.Error(err))
			result.RecoveryTestStatus = "failed"
		} else {
			if testStatus, ok := recoveryTestResult["recovery_test_status"].(string); ok {
				result.RecoveryTestStatus = testStatus
			}
			result.BackupResults["recovery_test"] = recoveryTestResult
		}
	}

	// 5. Применение политики хранения
	logger.Info("Применение политики хранения")
	
	var retentionResult map[string]interface{}
	err := workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"ApplyRetentionPolicyActivity",
		input.RetentionPolicy,
		input.BackupScope,
		input.Configuration,
	).Get(ctx, &retentionResult)

	if err != nil {
		logger.Warn("Ошибка применения политики хранения", zap.Error(err))
	} else {
		result.BackupResults["retention_policy"] = retentionResult
	}

	result.Status = "completed"
	result.Message = "Резервное копирование завершено"
	result.CompletedAt = workflow.Now(ctx)

	logger.Info("BackupRecoveryWorkflow завершен",
		zap.String("backup_type", input.BackupType),
		zap.String("schedule_type", input.ScheduleType),
		zap.String("status", result.Status),
		zap.String("backup_status", result.BackupStatus),
		zap.String("recovery_test_status", result.RecoveryTestStatus),
		zap.Int64("backup_size", result.BackupSize),
		zap.String("backup_location", result.BackupLocation),
		zap.Duration("duration", result.CompletedAt.Sub(result.StartedAt)),
	)

	return result, nil
}
