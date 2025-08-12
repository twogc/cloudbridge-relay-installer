package workflows

import (
	"fmt"
	"time"

	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// QuantumKeyRotationWorkflowInput входные параметры для ротации квантовых ключей
type QuantumKeyRotationWorkflowInput struct {
	KeyID           string                 `json:"key_id"`
	Algorithm       string                 `json:"algorithm"`       // kyber, dilithium, falcon, sphincs_plus
	RotationPolicy  string                 `json:"rotation_policy"` // time_based, usage_based, event_based
	Configuration   map[string]interface{} `json:"configuration"`
	ForceRotation   bool                   `json:"force_rotation"`
}

// QuantumKeyRotationWorkflowResult результат ротации квантовых ключей
type QuantumKeyRotationWorkflowResult struct {
	OldKeyID      string                 `json:"old_key_id"`
	NewKeyID      string                 `json:"new_key_id"`
	Algorithm     string                 `json:"algorithm"`
	Status        string                 `json:"status"`
	RotationInfo  map[string]interface{} `json:"rotation_info"`
	StartedAt     time.Time              `json:"started_at"`
	CompletedAt   time.Time              `json:"completed_at"`
	Message       string                 `json:"message"`
}

// KeyValidationWorkflowInput входные параметры для валидации квантовых ключей
type KeyValidationWorkflowInput struct {
	KeyIDs         []string               `json:"key_ids"`
	ValidationType string                 `json:"validation_type"` // integrity, security, compliance, performance
	Configuration  map[string]interface{} `json:"configuration"`
}

// KeyValidationWorkflowResult результат валидации квантовых ключей
type KeyValidationWorkflowResult struct {
	ValidatedKeys   []string               `json:"validated_keys"`
	InvalidKeys     []string               `json:"invalid_keys"`
	ValidationResults map[string]interface{} `json:"validation_results"`
	Status          string                 `json:"status"`
	StartedAt       time.Time              `json:"started_at"`
	CompletedAt     time.Time              `json:"completed_at"`
	Message         string                 `json:"message"`
}

// AlgorithmMigrationWorkflowInput входные параметры для миграции алгоритмов
type AlgorithmMigrationWorkflowInput struct {
	FromAlgorithm  string                 `json:"from_algorithm"`
	ToAlgorithm    string                 `json:"to_algorithm"`
	MigrationMode  string                 `json:"migration_mode"` // gradual, immediate, scheduled
	AffectedKeys   []string               `json:"affected_keys"`
	Configuration  map[string]interface{} `json:"configuration"`
}

// AlgorithmMigrationWorkflowResult результат миграции алгоритмов
type AlgorithmMigrationWorkflowResult struct {
	FromAlgorithm    string                 `json:"from_algorithm"`
	ToAlgorithm      string                 `json:"to_algorithm"`
	MigratedKeys     []string               `json:"migrated_keys"`
	FailedKeys       []string               `json:"failed_keys"`
	MigrationResults map[string]interface{} `json:"migration_results"`
	Status           string                 `json:"status"`
	StartedAt        time.Time              `json:"started_at"`
	CompletedAt      time.Time              `json:"completed_at"`
	Message          string                 `json:"message"`
}

// SecurityAuditWorkflowInput входные параметры для аудита безопасности
type SecurityAuditWorkflowInput struct {
	AuditScope    string                 `json:"audit_scope"`    // keys, algorithms, protocols, compliance
	AuditLevel    string                 `json:"audit_level"`    // basic, comprehensive, deep
	Components    []string               `json:"components"`
	Configuration map[string]interface{} `json:"configuration"`
}

// SecurityAuditWorkflowResult результат аудита безопасности
type SecurityAuditWorkflowResult struct {
	AuditScope       string                 `json:"audit_scope"`
	AuditLevel       string                 `json:"audit_level"`
	AuditResults     map[string]interface{} `json:"audit_results"`
	SecurityScore    float64                `json:"security_score"`
	Vulnerabilities  []string               `json:"vulnerabilities"`
	Recommendations  []string               `json:"recommendations"`
	Status           string                 `json:"status"`
	StartedAt        time.Time              `json:"started_at"`
	CompletedAt      time.Time              `json:"completed_at"`
	Message          string                 `json:"message"`
}

// QuantumKeyRotationWorkflow workflow для ротации квантовых ключей
func QuantumKeyRotationWorkflow(ctx workflow.Context, input QuantumKeyRotationWorkflowInput) (*QuantumKeyRotationWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Запуск QuantumKeyRotationWorkflow",
		zap.String("key_id", input.KeyID),
		zap.String("algorithm", input.Algorithm),
		zap.String("rotation_policy", input.RotationPolicy),
		zap.Bool("force_rotation", input.ForceRotation),
	)

	result := &QuantumKeyRotationWorkflowResult{
		OldKeyID:     input.KeyID,
		Algorithm:    input.Algorithm,
		Status:       "rotating",
		StartedAt:    workflow.Now(ctx),
		RotationInfo: make(map[string]interface{}),
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

	// 1. Проверка необходимости ротации (если не принудительная)
	if !input.ForceRotation {
		logger.Info("Проверка необходимости ротации ключа")
		
		var rotationCheck map[string]interface{}
		err := workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, activityOptions),
			"CheckKeyRotationNeedActivity",
			input.KeyID,
			input.RotationPolicy,
			input.Configuration,
		).Get(ctx, &rotationCheck)

		if err != nil {
			logger.Error("Ошибка проверки необходимости ротации", zap.Error(err))
			result.Status = "failed"
			result.Message = "Не удалось проверить необходимость ротации: " + err.Error()
			result.CompletedAt = workflow.Now(ctx)
			return result, err
		}

		// Если ротация не нужна, завершаем workflow
		if needRotation, ok := rotationCheck["need_rotation"].(bool); ok && !needRotation {
			result.Status = "skipped"
			result.Message = "Ротация ключа не требуется"
			result.CompletedAt = workflow.Now(ctx)
			logger.Info("Ротация ключа не требуется", zap.String("key_id", input.KeyID))
			return result, nil
		}

		result.RotationInfo["rotation_check"] = rotationCheck
	}

	// 2. Генерация нового квантового ключа
	logger.Info("Генерация нового квантового ключа", zap.String("algorithm", input.Algorithm))
	
	var newKeyResult map[string]interface{}
	err := workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"GenerateQuantumKeyActivity",
		input.Algorithm,
		input.Configuration,
	).Get(ctx, &newKeyResult)

	if err != nil {
		logger.Error("Ошибка генерации нового ключа", zap.Error(err))
		result.Status = "failed"
		result.Message = "Не удалось сгенерировать новый ключ: " + err.Error()
		result.CompletedAt = workflow.Now(ctx)
		return result, err
	}

	if newKeyID, ok := newKeyResult["key_id"].(string); ok {
		result.NewKeyID = newKeyID
	}
	result.RotationInfo["new_key_generation"] = newKeyResult

	// 3. Валидация нового ключа
	logger.Info("Валидация нового ключа", zap.String("new_key_id", result.NewKeyID))
	
	var validationResult map[string]interface{}
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"ValidateQuantumKeyActivity",
		result.NewKeyID,
		input.Algorithm,
	).Get(ctx, &validationResult)

	if err != nil {
		logger.Error("Ошибка валидации нового ключа", zap.Error(err))
		result.Status = "failed"
		result.Message = "Валидация нового ключа не прошла: " + err.Error()
		result.CompletedAt = workflow.Now(ctx)
		return result, err
	}

	result.RotationInfo["key_validation"] = validationResult

	// 4. Постепенная замена старого ключа новым
	logger.Info("Замена старого ключа новым")
	
	var replacementResult map[string]interface{}
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"ReplaceQuantumKeyActivity",
		input.KeyID,
		result.NewKeyID,
		input.Algorithm,
		input.Configuration,
	).Get(ctx, &replacementResult)

	if err != nil {
		logger.Error("Ошибка замены ключа", zap.Error(err))
		result.Status = "failed"
		result.Message = "Не удалось заменить ключ: " + err.Error()
		result.CompletedAt = workflow.Now(ctx)
		return result, err
	}

	result.RotationInfo["key_replacement"] = replacementResult

	// 5. Безопасное удаление старого ключа
	logger.Info("Безопасное удаление старого ключа", zap.String("old_key_id", input.KeyID))
	
	var deletionResult map[string]interface{}
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"SecureDeleteQuantumKeyActivity",
		input.KeyID,
	).Get(ctx, &deletionResult)

	if err != nil {
		logger.Warn("Ошибка удаления старого ключа", zap.Error(err))
		// Не критичная ошибка, продолжаем
	}

	result.RotationInfo["key_deletion"] = deletionResult

	// 6. Обновление ссылок на ключ в системе
	logger.Info("Обновление ссылок на ключ в системе")
	
	var updateResult map[string]interface{}
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"UpdateKeyReferencesActivity",
		input.KeyID,
		result.NewKeyID,
		input.Algorithm,
	).Get(ctx, &updateResult)

	if err != nil {
		logger.Error("Ошибка обновления ссылок на ключ", zap.Error(err))
		result.Status = "failed"
		result.Message = "Не удалось обновить ссылки на ключ: " + err.Error()
		result.CompletedAt = workflow.Now(ctx)
		return result, err
	}

	result.RotationInfo["reference_update"] = updateResult

	result.Status = "completed"
	result.Message = "Ротация квантового ключа успешно завершена"
	result.CompletedAt = workflow.Now(ctx)

	logger.Info("QuantumKeyRotationWorkflow завершен",
		zap.String("old_key_id", result.OldKeyID),
		zap.String("new_key_id", result.NewKeyID),
		zap.String("algorithm", result.Algorithm),
		zap.String("status", result.Status),
		zap.Duration("duration", result.CompletedAt.Sub(result.StartedAt)),
	)

	return result, nil
}

// KeyValidationWorkflow workflow для валидации квантовых ключей
func KeyValidationWorkflow(ctx workflow.Context, input KeyValidationWorkflowInput) (*KeyValidationWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Запуск KeyValidationWorkflow",
		zap.Strings("key_ids", input.KeyIDs),
		zap.String("validation_type", input.ValidationType),
	)

	result := &KeyValidationWorkflowResult{
		Status:            "validating",
		StartedAt:         workflow.Now(ctx),
		ValidatedKeys:     make([]string, 0),
		InvalidKeys:       make([]string, 0),
		ValidationResults: make(map[string]interface{}),
	}

	// Настройки для activities
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		HeartbeatTimeout:    1 * time.Minute,
		RetryPolicy: &workflow.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    2,
		},
	}

	// Валидируем ключи параллельно
	futures := make([]workflow.Future, len(input.KeyIDs))
	for i, keyID := range input.KeyIDs {
		futures[i] = workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, activityOptions),
			"PerformKeyValidationActivity",
			keyID,
			input.ValidationType,
			input.Configuration,
		)
	}

	// Собираем результаты валидации
	for i, future := range futures {
		keyID := input.KeyIDs[i]
		
		var validationResult map[string]interface{}
		err := future.Get(ctx, &validationResult)
		
		if err != nil {
			logger.Warn("Ошибка валидации ключа",
				zap.String("key_id", keyID),
				zap.Error(err),
			)
			result.InvalidKeys = append(result.InvalidKeys, keyID)
			result.ValidationResults[keyID] = map[string]interface{}{
				"status": "failed",
				"error":  err.Error(),
			}
		} else {
			if valid, ok := validationResult["valid"].(bool); ok && valid {
				result.ValidatedKeys = append(result.ValidatedKeys, keyID)
				logger.Info("Ключ прошел валидацию", zap.String("key_id", keyID))
			} else {
				result.InvalidKeys = append(result.InvalidKeys, keyID)
				logger.Warn("Ключ не прошел валидацию", zap.String("key_id", keyID))
			}
			
			result.ValidationResults[keyID] = validationResult
		}
	}

	// Генерация общего отчета о валидации
	logger.Info("Генерация отчета о валидации")
	
	var reportResult map[string]interface{}
	err := workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"GenerateValidationReportActivity",
		result.ValidatedKeys,
		result.InvalidKeys,
		result.ValidationResults,
		input.ValidationType,
	).Get(ctx, &reportResult)

	if err != nil {
		logger.Warn("Ошибка генерации отчета", zap.Error(err))
	} else {
		result.ValidationResults["report"] = reportResult
	}

	// Определяем итоговый статус
	if len(result.InvalidKeys) == 0 {
		result.Status = "all_valid"
		result.Message = "Все ключи прошли валидацию"
	} else if len(result.ValidatedKeys) > 0 {
		result.Status = "partial"
		result.Message = "Часть ключей прошла валидацию"
	} else {
		result.Status = "all_invalid"
		result.Message = "Ни один ключ не прошел валидацию"
	}

	result.CompletedAt = workflow.Now(ctx)

	logger.Info("KeyValidationWorkflow завершен",
		zap.String("validation_type", input.ValidationType),
		zap.String("status", result.Status),
		zap.Int("validated", len(result.ValidatedKeys)),
		zap.Int("invalid", len(result.InvalidKeys)),
		zap.Duration("duration", result.CompletedAt.Sub(result.StartedAt)),
	)

	return result, nil
}

// AlgorithmMigrationWorkflow workflow для миграции между квантовыми алгоритмами
func AlgorithmMigrationWorkflow(ctx workflow.Context, input AlgorithmMigrationWorkflowInput) (*AlgorithmMigrationWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Запуск AlgorithmMigrationWorkflow",
		zap.String("from_algorithm", input.FromAlgorithm),
		zap.String("to_algorithm", input.ToAlgorithm),
		zap.String("migration_mode", input.MigrationMode),
		zap.Int("affected_keys", len(input.AffectedKeys)),
	)

	result := &AlgorithmMigrationWorkflowResult{
		FromAlgorithm:    input.FromAlgorithm,
		ToAlgorithm:      input.ToAlgorithm,
		Status:           "migrating",
		StartedAt:        workflow.Now(ctx),
		MigratedKeys:     make([]string, 0),
		FailedKeys:       make([]string, 0),
		MigrationResults: make(map[string]interface{}),
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

	// 1. Проверка совместимости алгоритмов
	logger.Info("Проверка совместимости алгоритмов")
	
	var compatibilityResult map[string]interface{}
	err := workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"CheckAlgorithmCompatibilityActivity",
		input.FromAlgorithm,
		input.ToAlgorithm,
		input.Configuration,
	).Get(ctx, &compatibilityResult)

	if err != nil {
		logger.Error("Ошибка проверки совместимости", zap.Error(err))
		result.Status = "failed"
		result.Message = "Алгоритмы несовместимы: " + err.Error()
		result.CompletedAt = workflow.Now(ctx)
		return result, err
	}

	result.MigrationResults["compatibility_check"] = compatibilityResult

	// 2. Планирование миграции
	logger.Info("Планирование миграции", zap.String("mode", input.MigrationMode))
	
	var migrationPlan map[string]interface{}
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"PlanAlgorithmMigrationActivity",
		input.FromAlgorithm,
		input.ToAlgorithm,
		input.MigrationMode,
		input.AffectedKeys,
		input.Configuration,
	).Get(ctx, &migrationPlan)

	if err != nil {
		logger.Error("Ошибка планирования миграции", zap.Error(err))
		result.Status = "failed"
		result.Message = "Не удалось спланировать миграцию: " + err.Error()
		result.CompletedAt = workflow.Now(ctx)
		return result, err
	}

	result.MigrationResults["migration_plan"] = migrationPlan

	// 3. Выполнение миграции в зависимости от режима
	switch input.MigrationMode {
	case "immediate":
		// Немедленная миграция всех ключей параллельно
		logger.Info("Немедленная миграция ключей")
		
		futures := make([]workflow.Future, len(input.AffectedKeys))
		for i, keyID := range input.AffectedKeys {
			futures[i] = workflow.ExecuteActivity(
				workflow.WithActivityOptions(ctx, activityOptions),
				"MigrateKeyAlgorithmActivity",
				keyID,
				input.FromAlgorithm,
				input.ToAlgorithm,
				"immediate",
				input.Configuration,
			)
		}

		// Собираем результаты
		for i, future := range futures {
			keyID := input.AffectedKeys[i]
			
			var migrationResult map[string]interface{}
			err := future.Get(ctx, &migrationResult)
			
			if err != nil {
				logger.Warn("Ошибка миграции ключа",
					zap.String("key_id", keyID),
					zap.Error(err),
				)
				result.FailedKeys = append(result.FailedKeys, keyID)
			} else {
				result.MigratedKeys = append(result.MigratedKeys, keyID)
				logger.Info("Ключ успешно мигрирован", zap.String("key_id", keyID))
			}
			
			result.MigrationResults[keyID] = migrationResult
		}

	case "gradual":
		// Постепенная миграция с задержками
		logger.Info("Постепенная миграция ключей")
		
		for _, keyID := range input.AffectedKeys {
			var migrationResult map[string]interface{}
			err := workflow.ExecuteActivity(
				workflow.WithActivityOptions(ctx, activityOptions),
				"MigrateKeyAlgorithmActivity",
				keyID,
				input.FromAlgorithm,
				input.ToAlgorithm,
				"gradual",
				input.Configuration,
			).Get(ctx, &migrationResult)
			
			if err != nil {
				logger.Warn("Ошибка миграции ключа",
					zap.String("key_id", keyID),
					zap.Error(err),
				)
				result.FailedKeys = append(result.FailedKeys, keyID)
			} else {
				result.MigratedKeys = append(result.MigratedKeys, keyID)
				logger.Info("Ключ успешно мигрирован", zap.String("key_id", keyID))
			}
			
			result.MigrationResults[keyID] = migrationResult

			// Задержка между миграциями
			if len(input.AffectedKeys) > 1 {
				workflow.Sleep(ctx, 30*time.Second)
			}
		}

	case "scheduled":
		// Запланированная миграция
		logger.Info("Запланированная миграция ключей")
		
		// Ждем оптимального времени для миграции
		var scheduleResult map[string]interface{}
		err = workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, activityOptions),
			"ScheduleMigrationActivity",
			input.AffectedKeys,
			input.Configuration,
		).Get(ctx, &scheduleResult)

		if err != nil {
			logger.Error("Ошибка планирования времени миграции", zap.Error(err))
			result.Status = "failed"
			result.Message = "Не удалось запланировать миграцию: " + err.Error()
			result.CompletedAt = workflow.Now(ctx)
			return result, err
		}

		// Ждем запланированного времени
		if scheduleTime, ok := scheduleResult["schedule_time"].(string); ok {
			scheduledTime, _ := time.Parse(time.RFC3339, scheduleTime)
			waitDuration := scheduledTime.Sub(workflow.Now(ctx))
			if waitDuration > 0 {
				logger.Info("Ожидание запланированного времени миграции",
					zap.Duration("wait_duration", waitDuration),
				)
				workflow.Sleep(ctx, waitDuration)
			}
		}

		// Выполняем миграцию
		for _, keyID := range input.AffectedKeys {
			var migrationResult map[string]interface{}
			err := workflow.ExecuteActivity(
				workflow.WithActivityOptions(ctx, activityOptions),
				"MigrateKeyAlgorithmActivity",
				keyID,
				input.FromAlgorithm,
				input.ToAlgorithm,
				"scheduled",
				input.Configuration,
			).Get(ctx, &migrationResult)
			
			if err != nil {
				logger.Warn("Ошибка миграции ключа",
					zap.String("key_id", keyID),
					zap.Error(err),
				)
				result.FailedKeys = append(result.FailedKeys, keyID)
			} else {
				result.MigratedKeys = append(result.MigratedKeys, keyID)
				logger.Info("Ключ успешно мигрирован", zap.String("key_id", keyID))
			}
			
			result.MigrationResults[keyID] = migrationResult
		}

	default:
		logger.Error("Неизвестный режим миграции", zap.String("mode", input.MigrationMode))
		result.Status = "failed"
		result.Message = "Неизвестный режим миграции: " + input.MigrationMode
		result.CompletedAt = workflow.Now(ctx)
		return result, fmt.Errorf("unknown_migration_mode: %s", input.MigrationMode)
	}

	// 4. Постмиграционная валидация
	logger.Info("Постмиграционная валидация")
	
	if len(result.MigratedKeys) > 0 {
		var validationResult map[string]interface{}
		err = workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, activityOptions),
			"ValidateMigrationActivity",
			result.MigratedKeys,
			input.ToAlgorithm,
		).Get(ctx, &validationResult)

		if err != nil {
			logger.Warn("Ошибка постмиграционной валидации", zap.Error(err))
		} else {
			result.MigrationResults["post_migration_validation"] = validationResult
		}
	}

	// Определяем итоговый статус
	if len(result.FailedKeys) == 0 {
		result.Status = "completed"
		result.Message = "Все ключи успешно мигрированы"
	} else if len(result.MigratedKeys) > 0 {
		result.Status = "partial"
		result.Message = "Часть ключей мигрирована"
	} else {
		result.Status = "failed"
		result.Message = "Ни один ключ не был мигрирован"
	}

	result.CompletedAt = workflow.Now(ctx)

	logger.Info("AlgorithmMigrationWorkflow завершен",
		zap.String("from_algorithm", input.FromAlgorithm),
		zap.String("to_algorithm", input.ToAlgorithm),
		zap.String("migration_mode", input.MigrationMode),
		zap.String("status", result.Status),
		zap.Int("migrated", len(result.MigratedKeys)),
		zap.Int("failed", len(result.FailedKeys)),
		zap.Duration("duration", result.CompletedAt.Sub(result.StartedAt)),
	)

	return result, nil
}

// SecurityAuditWorkflow workflow для аудита безопасности квантовой криптографии
func SecurityAuditWorkflow(ctx workflow.Context, input SecurityAuditWorkflowInput) (*SecurityAuditWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Запуск SecurityAuditWorkflow",
		zap.String("audit_scope", input.AuditScope),
		zap.String("audit_level", input.AuditLevel),
		zap.Strings("components", input.Components),
	)

	result := &SecurityAuditWorkflowResult{
		AuditScope:      input.AuditScope,
		AuditLevel:      input.AuditLevel,
		Status:          "auditing",
		StartedAt:       workflow.Now(ctx),
		AuditResults:    make(map[string]interface{}),
		Vulnerabilities: make([]string, 0),
		Recommendations: make([]string, 0),
	}

	// Настройки для activities
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Minute,
		HeartbeatTimeout:    2 * time.Minute,
		RetryPolicy: &workflow.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    1 * time.Minute,
			MaximumAttempts:    2,
		},
	}

	// 1. Аудит ключей (если в scope)
	if input.AuditScope == "keys" || input.AuditScope == "comprehensive" {
		logger.Info("Аудит квантовых ключей")
		
		var keyAuditResult map[string]interface{}
		err := workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, activityOptions),
			"AuditQuantumKeysActivity",
			input.AuditLevel,
			input.Components,
			input.Configuration,
		).Get(ctx, &keyAuditResult)

		if err != nil {
			logger.Warn("Ошибка аудита ключей", zap.Error(err))
		} else {
			result.AuditResults["keys"] = keyAuditResult
		}
	}

	// 2. Аудит алгоритмов (если в scope)
	if input.AuditScope == "algorithms" || input.AuditScope == "comprehensive" {
		logger.Info("Аудит квантовых алгоритмов")
		
		var algorithmAuditResult map[string]interface{}
		err := workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, activityOptions),
			"AuditQuantumAlgorithmsActivity",
			input.AuditLevel,
			input.Components,
			input.Configuration,
		).Get(ctx, &algorithmAuditResult)

		if err != nil {
			logger.Warn("Ошибка аудита алгоритмов", zap.Error(err))
		} else {
			result.AuditResults["algorithms"] = algorithmAuditResult
		}
	}

	// 3. Аудит протоколов (если в scope)
	if input.AuditScope == "protocols" || input.AuditScope == "comprehensive" {
		logger.Info("Аудит квантовых протоколов")
		
		var protocolAuditResult map[string]interface{}
		err := workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, activityOptions),
			"AuditQuantumProtocolsActivity",
			input.AuditLevel,
			input.Components,
			input.Configuration,
		).Get(ctx, &protocolAuditResult)

		if err != nil {
			logger.Warn("Ошибка аудита протоколов", zap.Error(err))
		} else {
			result.AuditResults["protocols"] = protocolAuditResult
		}
	}

	// 4. Аудит соответствия стандартам (если в scope)
	if input.AuditScope == "compliance" || input.AuditScope == "comprehensive" {
		logger.Info("Аудит соответствия стандартам")
		
		var complianceAuditResult map[string]interface{}
		err := workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, activityOptions),
			"AuditComplianceActivity",
			input.AuditLevel,
			input.Components,
			input.Configuration,
		).Get(ctx, &complianceAuditResult)

		if err != nil {
			logger.Warn("Ошибка аудита соответствия", zap.Error(err))
		} else {
			result.AuditResults["compliance"] = complianceAuditResult
		}
	}

	// 5. Анализ уязвимостей
	logger.Info("Анализ уязвимостей")
	
	var vulnerabilityResult map[string]interface{}
	err := workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"AnalyzeVulnerabilitiesActivity",
		result.AuditResults,
		input.AuditLevel,
	).Get(ctx, &vulnerabilityResult)

	if err != nil {
		logger.Warn("Ошибка анализа уязвимостей", zap.Error(err))
	} else {
		result.AuditResults["vulnerabilities"] = vulnerabilityResult
		
		if vulns, ok := vulnerabilityResult["vulnerabilities"].([]interface{}); ok {
			for _, vuln := range vulns {
				if vulnStr, ok := vuln.(string); ok {
					result.Vulnerabilities = append(result.Vulnerabilities, vulnStr)
				}
			}
		}
	}

	// 6. Генерация рекомендаций
	logger.Info("Генерация рекомендаций по безопасности")
	
	var recommendationResult map[string]interface{}
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"GenerateSecurityRecommendationsActivity",
		result.AuditResults,
		result.Vulnerabilities,
	).Get(ctx, &recommendationResult)

	if err != nil {
		logger.Warn("Ошибка генерации рекомендаций", zap.Error(err))
	} else {
		result.AuditResults["recommendations"] = recommendationResult
		
		if recs, ok := recommendationResult["recommendations"].([]interface{}); ok {
			for _, rec := range recs {
				if recStr, ok := rec.(string); ok {
					result.Recommendations = append(result.Recommendations, recStr)
				}
			}
		}
	}

	// 7. Расчет общего security score
	logger.Info("Расчет общего security score")
	
	var scoreResult map[string]interface{}
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		"CalculateSecurityScoreActivity",
		result.AuditResults,
		len(result.Vulnerabilities),
	).Get(ctx, &scoreResult)

	if err != nil {
		logger.Warn("Ошибка расчета security score", zap.Error(err))
		result.SecurityScore = 0.0
	} else {
		if score, ok := scoreResult["security_score"].(float64); ok {
			result.SecurityScore = score
		}
		result.AuditResults["security_score"] = scoreResult
	}

	// Определяем итоговый статус
	if result.SecurityScore >= 0.9 {
		result.Status = "excellent"
		result.Message = "Отличный уровень безопасности"
	} else if result.SecurityScore >= 0.7 {
		result.Status = "good"
		result.Message = "Хороший уровень безопасности"
	} else if result.SecurityScore >= 0.5 {
		result.Status = "moderate"
		result.Message = "Умеренный уровень безопасности"
	} else {
		result.Status = "poor"
		result.Message = "Низкий уровень безопасности"
	}

	result.CompletedAt = workflow.Now(ctx)

	logger.Info("SecurityAuditWorkflow завершен",
		zap.String("audit_scope", input.AuditScope),
		zap.String("audit_level", input.AuditLevel),
		zap.String("status", result.Status),
		zap.Float64("security_score", result.SecurityScore),
		zap.Int("vulnerabilities", len(result.Vulnerabilities)),
		zap.Int("recommendations", len(result.Recommendations)),
		zap.Duration("duration", result.CompletedAt.Sub(result.StartedAt)),
	)

	return result, nil
}
