package api

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewInputValidator(t *testing.T) {
	logger := zap.NewNop()

	validator := NewInputValidator(logger)

	assert.NotNil(t, validator)
}

func TestInputValidator_ValidateJSON(t *testing.T) {
	logger := zap.NewNop()
	validator := NewInputValidator(logger)

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid JSON",
			input:    `{"key": "value"}`,
			expected: true,
		},
		{
			name:     "invalid JSON",
			input:    `{"key": "value"`,
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateJSON(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInputValidator_ValidateEmail(t *testing.T) {
	logger := zap.NewNop()
	validator := NewInputValidator(logger)

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid email",
			input:    "test@example.com",
			expected: true,
		},
		{
			name:     "invalid email",
			input:    "invalid-email",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateEmail(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInputValidator_ValidateURL(t *testing.T) {
	logger := zap.NewNop()
	validator := NewInputValidator(logger)

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid HTTP URL",
			input:    "http://example.com",
			expected: true,
		},
		{
			name:     "valid HTTPS URL",
			input:    "https://example.com",
			expected: true,
		},
		{
			name:     "invalid URL",
			input:    "not-a-url",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateURL(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInputValidator_SanitizeString(t *testing.T) {
	logger := zap.NewNop()
	validator := NewInputValidator(logger)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal string",
			input:    "hello world",
			expected: "helloworld",
		},
		{
			name:     "string with HTML",
			input:    "<script>alert('xss')</script>",
			expected: "scriptalertxssscript",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "string with special characters",
			input:    "test@#$%^&*()",
			expected: "test",
		},
		{
			name:     "string with allowed characters",
			input:    "test-123_abc",
			expected: "test-123_abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.SanitizeString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInputValidator_ValidateDataQueryRequest(t *testing.T) {
	logger := zap.NewNop()
	validator := NewInputValidator(logger)

	now := time.Now()
	validReq := &DataQueryRequest{
		StartTime: now.Add(-1 * time.Hour),
		EndTime:   now,
		DataType:  "metrics",
		Limit:     100,
	}

	tests := []struct {
		name     string
		request  *DataQueryRequest
		rules    *DataQueryValidationRules
		expected bool
	}{
		{
			name:     "valid request with default rules",
			request:  validReq,
			rules:    nil,
			expected: true,
		},
		{
			name:    "valid request with custom rules",
			request: validReq,
			rules: &DataQueryValidationRules{
				MaxTimeRange:     2 * time.Hour,
				MinTimeRange:     30 * time.Minute,
				MaxLimit:         200,
				MinLimit:         1,
				AllowedDataTypes: []string{"metrics", "logs"},
				MaxQuerySize:     2048,
			},
			expected: true,
		},
		{
			name: "invalid time range - start time in future",
			request: &DataQueryRequest{
				StartTime: now.Add(1 * time.Hour),
				EndTime:   now.Add(2 * time.Hour),
				DataType:  "metrics",
				Limit:     100,
			},
			rules:    nil,
			expected: false,
		},
		{
			name: "invalid time range - end time before start time",
			request: &DataQueryRequest{
				StartTime: now,
				EndTime:   now.Add(-1 * time.Hour),
				DataType:  "metrics",
				Limit:     100,
			},
			rules:    nil,
			expected: false,
		},
		{
			name: "invalid limit - too high",
			request: &DataQueryRequest{
				StartTime: now.Add(-1 * time.Hour),
				EndTime:   now,
				DataType:  "metrics",
				Limit:     20000,
			},
			rules:    nil,
			expected: false,
		},
		{
			name: "invalid data type",
			request: &DataQueryRequest{
				StartTime: now.Add(-1 * time.Hour),
				EndTime:   now,
				DataType:  "invalid_type",
				Limit:     100,
			},
			rules:    nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateDataQueryRequest(tt.request, tt.rules)
			assert.Equal(t, tt.expected, result.Valid)
		})
	}
}

func TestInputValidator_ValidateCollectorName(t *testing.T) {
	logger := zap.NewNop()
	validator := NewInputValidator(logger)

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid collector name - prometheus",
			input:    "prometheus",
			expected: true,
		},
		{
			name:     "valid collector name - system",
			input:    "system",
			expected: true,
		},
		{
			name:     "valid collector name - network",
			input:    "network",
			expected: true,
		},
		{
			name:     "valid collector name - case insensitive",
			input:    "PROMETHEUS",
			expected: true,
		},
		{
			name:     "invalid collector name",
			input:    "invalid_collector",
			expected: false,
		},
		{
			name:     "empty collector name",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateCollectorName(tt.input)
			assert.Equal(t, tt.expected, result.Valid)
		})
	}
}

func TestInputValidator_ValidatePipelineAction(t *testing.T) {
	logger := zap.NewNop()
	validator := NewInputValidator(logger)

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid action - start",
			input:    "start",
			expected: true,
		},
		{
			name:     "valid action - stop",
			input:    "stop",
			expected: true,
		},
		{
			name:     "valid action - restart",
			input:    "restart",
			expected: true,
		},
		{
			name:     "valid action - status",
			input:    "status",
			expected: true,
		},
		{
			name:     "valid action - case insensitive",
			input:    "START",
			expected: true,
		},
		{
			name:     "invalid action",
			input:    "invalid_action",
			expected: false,
		},
		{
			name:     "empty action",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidatePipelineAction(tt.input)
			assert.Equal(t, tt.expected, result.Valid)
		})
	}
}

func TestInputValidator_DefaultDataQueryValidationRules(t *testing.T) {
	rules := DefaultDataQueryValidationRules()

	assert.NotNil(t, rules)
	assert.Equal(t, 30*24*time.Hour, rules.MaxTimeRange)
	assert.Equal(t, 1*time.Minute, rules.MinTimeRange)
	assert.Equal(t, 10000, rules.MaxLimit)
	assert.Equal(t, 1, rules.MinLimit)
	assert.Equal(t, 1024*1024, rules.MaxQuerySize)
	assert.Contains(t, rules.AllowedDataTypes, "metrics")
	assert.Contains(t, rules.AllowedDataTypes, "logs")
	assert.Contains(t, rules.AllowedDataTypes, "traces")
}

func TestInputValidator_ValidateTimeRange(t *testing.T) {
	logger := zap.NewNop()
	validator := NewInputValidator(logger)

	now := time.Now()
	rules := DefaultDataQueryValidationRules()

	tests := []struct {
		name      string
		startTime time.Time
		endTime   time.Time
		expected  bool
	}{
		{
			name:      "valid time range",
			startTime: now.Add(-1 * time.Hour),
			endTime:   now,
			expected:  true,
		},
		{
			name:      "zero start time",
			startTime: time.Time{},
			endTime:   now,
			expected:  false,
		},
		{
			name:      "zero end time",
			startTime: now.Add(-1 * time.Hour),
			endTime:   time.Time{},
			expected:  false,
		},
		{
			name:      "start time after end time",
			startTime: now,
			endTime:   now.Add(-1 * time.Hour),
			expected:  false,
		},
		{
			name:      "time range too short",
			startTime: now.Add(-30 * time.Second),
			endTime:   now,
			expected:  false,
		},
		{
			name:      "time range too long",
			startTime: now.Add(-31 * 24 * time.Hour),
			endTime:   now,
			expected:  false,
		},
		{
			name:      "start time in future",
			startTime: now.Add(1 * time.Hour),
			endTime:   now.Add(2 * time.Hour),
			expected:  false,
		},
		{
			name:      "end time in future",
			startTime: now.Add(-1 * time.Hour),
			endTime:   now.Add(1 * time.Hour),
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.validateTimeRange(tt.startTime, tt.endTime, rules)
			if tt.expected {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

func TestInputValidator_ValidateLimit(t *testing.T) {
	logger := zap.NewNop()
	validator := NewInputValidator(logger)

	rules := DefaultDataQueryValidationRules()

	tests := []struct {
		name     string
		limit    int
		expected bool
	}{
		{
			name:     "valid limit",
			limit:    100,
			expected: true,
		},
		{
			name:     "minimum limit",
			limit:    1,
			expected: true,
		},
		{
			name:     "maximum limit",
			limit:    10000,
			expected: true,
		},
		{
			name:     "limit too low",
			limit:    0,
			expected: false,
		},
		{
			name:     "limit too high",
			limit:    10001,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.validateLimit(tt.limit, rules)
			if tt.expected {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

func TestInputValidator_ValidateDataType(t *testing.T) {
	logger := zap.NewNop()
	validator := NewInputValidator(logger)

	rules := DefaultDataQueryValidationRules()

	tests := []struct {
		name     string
		dataType string
		expected bool
	}{
		{
			name:     "valid data type - metrics",
			dataType: "metrics",
			expected: true,
		},
		{
			name:     "valid data type - logs",
			dataType: "logs",
			expected: true,
		},
		{
			name:     "valid data type - case insensitive",
			dataType: "METRICS",
			expected: true,
		},
		{
			name:     "invalid data type",
			dataType: "invalid_type",
			expected: false,
		},
		{
			name:     "empty data type",
			dataType: "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.validateDataType(tt.dataType, rules)
			if tt.expected {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

func TestInputValidator_SanitizeDataQueryRequest(t *testing.T) {
	logger := zap.NewNop()
	validator := NewInputValidator(logger)

	now := time.Now()
	req := &DataQueryRequest{
		StartTime: now.Add(-1 * time.Hour),
		EndTime:   now,
		DataType:  "metrics",
		Limit:     100,
	}

	result := validator.sanitizeDataQueryRequest(req)

	assert.NotNil(t, result)
	assert.Equal(t, req.StartTime, result.StartTime)
	assert.Equal(t, req.EndTime, result.EndTime)
	assert.Equal(t, "metrics", result.DataType)
	assert.Equal(t, 100, result.Limit)
}

func TestInputValidator_SanitizeDataQueryRequest_Defaults(t *testing.T) {
	logger := zap.NewNop()
	validator := NewInputValidator(logger)

	req := &DataQueryRequest{
		DataType: "metrics",
		Limit:    0,
	}

	result := validator.sanitizeDataQueryRequest(req)

	assert.NotNil(t, result)
	assert.False(t, result.StartTime.IsZero())
	assert.False(t, result.EndTime.IsZero())
	assert.Equal(t, "metrics", result.DataType)
	assert.Equal(t, 1000, result.Limit) // default limit
}

func TestInputValidator_IsValidCollectorName(t *testing.T) {
	logger := zap.NewNop()
	validator := NewInputValidator(logger)

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid collector - prometheus",
			input:    "prometheus",
			expected: true,
		},
		{
			name:     "valid collector - system",
			input:    "system",
			expected: true,
		},
		{
			name:     "valid collector - network",
			input:    "network",
			expected: true,
		},
		{
			name:     "valid collector - relay",
			input:    "relay",
			expected: true,
		},
		{
			name:     "valid collector - application",
			input:    "application",
			expected: true,
		},
		{
			name:     "valid collector - database",
			input:    "database",
			expected: true,
		},
		{
			name:     "valid collector - cache",
			input:    "cache",
			expected: true,
		},
		{
			name:     "invalid collector",
			input:    "invalid",
			expected: false,
		},
		{
			name:     "case insensitive",
			input:    "PROMETHEUS",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.isValidCollectorName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
