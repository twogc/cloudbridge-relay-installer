package api

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// InputValidator валидирует и санитизирует входные данные
type InputValidator struct {
	logger *zap.Logger
}

// NewInputValidator создает новый валидатор входных данных
func NewInputValidator(logger *zap.Logger) *InputValidator {
	return &InputValidator{
		logger: logger,
	}
}

// ValidationError представляет ошибку валидации
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// ValidationResult представляет результат валидации
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors,omitempty"`
	Data   interface{}       `json:"data,omitempty"`
}

// DataQueryValidationRules правила валидации для DataQueryRequest
type DataQueryValidationRules struct {
	MaxTimeRange     time.Duration `json:"max_time_range"`
	MinTimeRange     time.Duration `json:"min_time_range"`
	MaxLimit         int           `json:"max_limit"`
	MinLimit         int           `json:"min_limit"`
	AllowedDataTypes []string      `json:"allowed_data_types"`
	MaxQuerySize     int           `json:"max_query_size"`
}

// DefaultDataQueryValidationRules возвращает правила валидации по умолчанию
func DefaultDataQueryValidationRules() *DataQueryValidationRules {
	return &DataQueryValidationRules{
		MaxTimeRange: 30 * 24 * time.Hour, // 30 дней
		MinTimeRange: 1 * time.Minute,     // 1 минута
		MaxLimit:     10000,               // максимум 10,000 записей
		MinLimit:     1,                   // минимум 1 запись
		AllowedDataTypes: []string{
			"metrics",
			"logs",
			"traces",
			"events",
			"network",
			"system",
			"application",
		},
		MaxQuerySize: 1024 * 1024, // 1MB
	}
}

// ValidateDataQueryRequest валидирует запрос на запрос данных
func (iv *InputValidator) ValidateDataQueryRequest(req *DataQueryRequest, rules *DataQueryValidationRules) *ValidationResult {
	if rules == nil {
		rules = DefaultDataQueryValidationRules()
	}

	var errors []ValidationError

	// Валидация временного диапазона
	if err := iv.validateTimeRange(req.StartTime, req.EndTime, rules); err != nil {
		errors = append(errors, *err)
	}

	// Валидация лимита
	if err := iv.validateLimit(req.Limit, rules); err != nil {
		errors = append(errors, *err)
	}

	// Валидация типа данных
	if err := iv.validateDataType(req.DataType, rules); err != nil {
		errors = append(errors, *err)
	}

	// Санитизация данных
	sanitizedReq := iv.sanitizeDataQueryRequest(req)

	return &ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
		Data:   sanitizedReq,
	}
}

// validateTimeRange валидирует временной диапазон
func (iv *InputValidator) validateTimeRange(startTime, endTime time.Time, rules *DataQueryValidationRules) *ValidationError {
	// Проверяем, что времена не нулевые
	if startTime.IsZero() {
		return &ValidationError{
			Field:   "start_time",
			Message: "Start time is required",
			Code:    "REQUIRED",
		}
	}

	if endTime.IsZero() {
		return &ValidationError{
			Field:   "end_time",
			Message: "End time is required",
			Code:    "REQUIRED",
		}
	}

	// Проверяем, что start_time < end_time
	if startTime.After(endTime) {
		return &ValidationError{
			Field:   "time_range",
			Message: "Start time must be before end time",
			Code:    "INVALID_RANGE",
		}
	}

	// Проверяем минимальный временной диапазон
	duration := endTime.Sub(startTime)
	if duration < rules.MinTimeRange {
		return &ValidationError{
			Field:   "time_range",
			Message: fmt.Sprintf("Time range must be at least %v", rules.MinTimeRange),
			Code:    "MIN_RANGE_EXCEEDED",
		}
	}

	// Проверяем максимальный временной диапазон
	if duration > rules.MaxTimeRange {
		return &ValidationError{
			Field:   "time_range",
			Message: fmt.Sprintf("Time range must not exceed %v", rules.MaxTimeRange),
			Code:    "MAX_RANGE_EXCEEDED",
		}
	}

	// Проверяем, что времена не в будущем
	now := time.Now()
	if startTime.After(now) {
		return &ValidationError{
			Field:   "start_time",
			Message: "Start time cannot be in the future",
			Code:    "FUTURE_TIME",
		}
	}

	if endTime.After(now) {
		return &ValidationError{
			Field:   "end_time",
			Message: "End time cannot be in the future",
			Code:    "FUTURE_TIME",
		}
	}

	return nil
}

// validateLimit валидирует лимит записей
func (iv *InputValidator) validateLimit(limit int, rules *DataQueryValidationRules) *ValidationError {
	if limit < rules.MinLimit {
		return &ValidationError{
			Field:   "limit",
			Message: fmt.Sprintf("Limit must be at least %d", rules.MinLimit),
			Code:    "MIN_LIMIT_EXCEEDED",
		}
	}

	if limit > rules.MaxLimit {
		return &ValidationError{
			Field:   "limit",
			Message: fmt.Sprintf("Limit must not exceed %d", rules.MaxLimit),
			Code:    "MAX_LIMIT_EXCEEDED",
		}
	}

	return nil
}

// validateDataType валидирует тип данных
func (iv *InputValidator) validateDataType(dataType string, rules *DataQueryValidationRules) *ValidationError {
	if dataType == "" {
		return &ValidationError{
			Field:   "data_type",
			Message: "Data type is required",
			Code:    "REQUIRED",
		}
	}

	// Проверяем, что тип данных разрешен
	allowed := false
	for _, allowedType := range rules.AllowedDataTypes {
		if strings.EqualFold(dataType, allowedType) {
			allowed = true
			break
		}
	}

	if !allowed {
		return &ValidationError{
			Field:   "data_type",
			Message: fmt.Sprintf("Data type '%s' is not allowed. Allowed types: %s", dataType, strings.Join(rules.AllowedDataTypes, ", ")),
			Code:    "INVALID_TYPE",
		}
	}

	return nil
}

// sanitizeDataQueryRequest санитизирует запрос на запрос данных
func (iv *InputValidator) sanitizeDataQueryRequest(req *DataQueryRequest) *DataQueryRequest {
	sanitized := &DataQueryRequest{
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		DataType:  iv.sanitizeString(req.DataType),
		Limit:     req.Limit,
	}

	// Устанавливаем значения по умолчанию
	if sanitized.EndTime.IsZero() {
		sanitized.EndTime = time.Now()
	}
	if sanitized.StartTime.IsZero() {
		sanitized.StartTime = sanitized.EndTime.Add(-24 * time.Hour)
	}
	if sanitized.Limit == 0 {
		sanitized.Limit = 1000
	}

	return sanitized
}

// sanitizeString санитизирует строку
func (iv *InputValidator) sanitizeString(s string) string {
	if s == "" {
		return s
	}

	// Удаляем лишние пробелы
	s = strings.TrimSpace(s)

	// Удаляем потенциально опасные символы
	// Разрешаем только буквы, цифры, дефисы и подчеркивания
	reg := regexp.MustCompile(`[^a-zA-Z0-9\-_]`)
	s = reg.ReplaceAllString(s, "")

	return s
}

// ValidateCollectorName валидирует имя коллектора
func (iv *InputValidator) ValidateCollectorName(collectorName string) *ValidationResult {
	var errors []ValidationError

	if collectorName == "" {
		errors = append(errors, ValidationError{
			Field:   "collector",
			Message: "Collector name is required",
			Code:    "REQUIRED",
		})
	} else {
		// Проверяем формат имени коллектора
		if !iv.isValidCollectorName(collectorName) {
			errors = append(errors, ValidationError{
				Field:   "collector",
				Message: "Invalid collector name format",
				Code:    "INVALID_FORMAT",
			})
		}
	}

	return &ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
		Data:   iv.sanitizeString(collectorName),
	}
}

// isValidCollectorName проверяет, является ли имя коллектора валидным
func (iv *InputValidator) isValidCollectorName(name string) bool {
	// Разрешенные имена коллекторов
	allowedCollectors := []string{
		"prometheus",
		"system",
		"network",
		"relay",
		"application",
		"database",
		"cache",
	}

	for _, allowed := range allowedCollectors {
		if strings.EqualFold(name, allowed) {
			return true
		}
	}

	return false
}

// ValidatePipelineAction валидирует действие с pipeline
func (iv *InputValidator) ValidatePipelineAction(action string) *ValidationResult {
	var errors []ValidationError

	if action == "" {
		errors = append(errors, ValidationError{
			Field:   "action",
			Message: "Action is required",
			Code:    "REQUIRED",
		})
	} else {
		// Разрешенные действия
		allowedActions := []string{"start", "stop", "restart", "status"}
		allowed := false
		for _, allowedAction := range allowedActions {
			if strings.EqualFold(action, allowedAction) {
				allowed = true
				break
			}
		}

		if !allowed {
			errors = append(errors, ValidationError{
				Field:   "action",
				Message: fmt.Sprintf("Action '%s' is not allowed. Allowed actions: %s", action, strings.Join(allowedActions, ", ")),
				Code:    "INVALID_ACTION",
			})
		}
	}

	return &ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
		Data:   iv.sanitizeString(action),
	}
}

// HandleValidationError обрабатывает ошибки валидации
func (iv *InputValidator) HandleValidationError(c *gin.Context, result *ValidationResult) {
	iv.logger.Warn("Validation failed",
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.Int("error_count", len(result.Errors)),
	)

	c.JSON(http.StatusBadRequest, gin.H{
		"status":    "error",
		"message":   "Validation failed",
		"errors":    result.Errors,
		"timestamp": time.Now(),
	})
}

// CreateValidationMiddleware создает middleware для валидации
func (iv *InputValidator) CreateValidationMiddleware(validationFunc func(*gin.Context) *ValidationResult) gin.HandlerFunc {
	return func(c *gin.Context) {
		result := validationFunc(c)
		if !result.Valid {
			iv.HandleValidationError(c, result)
			c.Abort()
			return
		}

		// Сохраняем валидированные данные в контексте
		if result.Data != nil {
			c.Set("validated_data", result.Data)
		}

		c.Next()
	}
}
