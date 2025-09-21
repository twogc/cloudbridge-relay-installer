package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// DataPipelineInterface интерфейс для DataPipeline для тестирования
type DataPipelineInterface interface {
	GetStatus() (map[string]interface{}, error)
	GetCollectorData() ([]map[string]interface{}, error)
	QueryData(query string) ([]map[string]interface{}, error)
	GetMetrics() (map[string]interface{}, error)
	Start() error
	Stop() error
}

// InputValidatorInterface интерфейс для InputValidator для тестирования
type InputValidatorInterface interface {
	ValidateQuery(query string) error
}

// TestDataPipelineEndpoints тестовая версия DataPipelineEndpoints
type TestDataPipelineEndpoints struct {
	pipeline       DataPipelineInterface
	logger         *zap.Logger
	inputValidator InputValidatorInterface
}

// NewTestDataPipelineEndpoints создает тестовый DataPipelineEndpoints
func NewTestDataPipelineEndpoints(pipeline DataPipelineInterface, logger *zap.Logger) *TestDataPipelineEndpoints {
	return &TestDataPipelineEndpoints{
		pipeline:       pipeline,
		logger:         logger,
		inputValidator: &MockInputValidator{},
	}
}

// GetPipelineStatus тестовая версия GetPipelineStatus
func (dpe *TestDataPipelineEndpoints) GetPipelineStatus(c *gin.Context) {
	status, err := dpe.pipeline.GetStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, status)
}

// GetCollectorData тестовая версия GetCollectorData
func (dpe *TestDataPipelineEndpoints) GetCollectorData(c *gin.Context) {
	data, err := dpe.pipeline.GetCollectorData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, data)
}

// QueryData тестовая версия QueryData
func (dpe *TestDataPipelineEndpoints) QueryData(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	data, err := dpe.pipeline.QueryData(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetPipelineMetrics тестовая версия GetPipelineMetrics
func (dpe *TestDataPipelineEndpoints) GetPipelineMetrics(c *gin.Context) {
	metrics, err := dpe.pipeline.GetMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, metrics)
}

// StartPipeline тестовая версия StartPipeline
func (dpe *TestDataPipelineEndpoints) StartPipeline(c *gin.Context) {
	err := dpe.pipeline.Start()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Pipeline started successfully"})
}

// StopPipeline тестовая версия StopPipeline
func (dpe *TestDataPipelineEndpoints) StopPipeline(c *gin.Context) {
	err := dpe.pipeline.Stop()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Pipeline stopped successfully"})
}

func TestNewDataPipelineEndpoints(t *testing.T) {
	logger := zap.NewNop()
	mockPipeline := &MockDataPipeline{}

	endpoints := NewTestDataPipelineEndpoints(mockPipeline, logger)

	assert.NotNil(t, endpoints)
}

func TestDataPipelineEndpoints_GetStatus(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*MockDataPipeline)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful status retrieval",
			mockSetup: func(m *MockDataPipeline) {
				m.On("GetStatus").Return(map[string]interface{}{
					"running": true,
					"status":  "active",
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"running":true,"status":"active"}`,
		},
		{
			name: "pipeline error",
			mockSetup: func(m *MockDataPipeline) {
				m.On("GetStatus").Return((map[string]interface{})(nil), assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Internal server error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			mockPipeline := &MockDataPipeline{}
			tt.mockSetup(mockPipeline)

			endpoints := &TestDataPipelineEndpoints{
				pipeline: mockPipeline,
				logger:   zap.NewNop(),
			}

			endpoints.GetPipelineStatus(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
			mockPipeline.AssertExpectations(t)
		})
	}
}

func TestDataPipelineEndpoints_GetCollectorData(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*MockDataPipeline)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful data retrieval",
			mockSetup: func(m *MockDataPipeline) {
				m.On("GetCollectorData").Return([]map[string]interface{}{
					{"id": "collector1", "data": "test data"},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"id":"collector1","data":"test data"}]`,
		},
		{
			name: "pipeline error",
			mockSetup: func(m *MockDataPipeline) {
				m.On("GetCollectorData").Return(([]map[string]interface{})(nil), assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Internal server error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			mockPipeline := &MockDataPipeline{}
			tt.mockSetup(mockPipeline)

			endpoints := &TestDataPipelineEndpoints{
				pipeline: mockPipeline,
				logger:   zap.NewNop(),
			}

			endpoints.GetCollectorData(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
			mockPipeline.AssertExpectations(t)
		})
	}
}

func TestDataPipelineEndpoints_QueryData(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		mockSetup      func(*MockDataPipeline)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:  "successful query",
			query: "SELECT * FROM data",
			mockSetup: func(m *MockDataPipeline) {
				m.On("QueryData", "SELECT * FROM data").Return([]map[string]interface{}{
					{"timestamp": "2023-01-01", "value": 100},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"timestamp":"2023-01-01","value":100}]`,
		},
		{
			name:  "pipeline error",
			query: "SELECT * FROM data",
			mockSetup: func(m *MockDataPipeline) {
				m.On("QueryData", "SELECT * FROM data").Return(([]map[string]interface{})(nil), assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Internal server error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/query?q="+url.QueryEscape(tt.query), nil)

			mockPipeline := &MockDataPipeline{}
			tt.mockSetup(mockPipeline)

			endpoints := &TestDataPipelineEndpoints{
				pipeline: mockPipeline,
				logger:   zap.NewNop(),
			}

			endpoints.QueryData(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
			mockPipeline.AssertExpectations(t)
		})
	}
}

func TestDataPipelineEndpoints_GetMetrics(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*MockDataPipeline)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful metrics retrieval",
			mockSetup: func(m *MockDataPipeline) {
				m.On("GetMetrics").Return(map[string]interface{}{
					"total_records":     1000,
					"processed_records": 950,
					"failed_records":    50,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"total_records":1000,"processed_records":950,"failed_records":50}`,
		},
		{
			name: "pipeline error",
			mockSetup: func(m *MockDataPipeline) {
				m.On("GetMetrics").Return((map[string]interface{})(nil), assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Internal server error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			mockPipeline := &MockDataPipeline{}
			tt.mockSetup(mockPipeline)

			endpoints := &TestDataPipelineEndpoints{
				pipeline: mockPipeline,
				logger:   zap.NewNop(),
			}

			endpoints.GetPipelineMetrics(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
			mockPipeline.AssertExpectations(t)
		})
	}
}

func TestDataPipelineEndpoints_StartPipeline(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*MockDataPipeline)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful start",
			mockSetup: func(m *MockDataPipeline) {
				m.On("Start").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"Pipeline started successfully"}`,
		},
		{
			name: "start error",
			mockSetup: func(m *MockDataPipeline) {
				m.On("Start").Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Internal server error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			mockPipeline := &MockDataPipeline{}
			tt.mockSetup(mockPipeline)

			endpoints := &TestDataPipelineEndpoints{
				pipeline: mockPipeline,
				logger:   zap.NewNop(),
			}

			endpoints.StartPipeline(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
			mockPipeline.AssertExpectations(t)
		})
	}
}

func TestDataPipelineEndpoints_StopPipeline(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*MockDataPipeline)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful stop",
			mockSetup: func(m *MockDataPipeline) {
				m.On("Stop").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"Pipeline stopped successfully"}`,
		},
		{
			name: "stop error",
			mockSetup: func(m *MockDataPipeline) {
				m.On("Stop").Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Internal server error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			mockPipeline := &MockDataPipeline{}
			tt.mockSetup(mockPipeline)

			endpoints := &TestDataPipelineEndpoints{
				pipeline: mockPipeline,
				logger:   zap.NewNop(),
			}

			endpoints.StopPipeline(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
			mockPipeline.AssertExpectations(t)
		})
	}
}

func TestDataPipelineEndpoints_InvalidInput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Отправляем невалидный JSON
	invalidJSON := `{"invalid": json}`
	c.Request = httptest.NewRequest("POST", "/api/v1/data/query", strings.NewReader(invalidJSON))
	c.Request.Header.Set("Content-Type", "application/json")

	mockPipeline := &MockDataPipeline{}
	validator := &MockInputValidator{}

	endpoints := &TestDataPipelineEndpoints{
		pipeline:       mockPipeline,
		inputValidator: validator,
		logger:         zap.NewNop(),
	}

	endpoints.QueryData(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
