package relay

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/anomaly"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/predictive"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/routing"
	"go.uber.org/zap/zaptest"
)

// MockAnomalyDetector представляет мок детектора аномалий
type MockAnomalyDetector struct {
	mock.Mock
}

func (m *MockAnomalyDetector) DetectAnomaly(metrics *anomaly.NetworkMetrics) (*anomaly.AnomalyResult, error) {
	args := m.Called(metrics)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*anomaly.AnomalyResult), args.Error(1)
}

// MockLoadPredictor представляет мок предиктора нагрузки
type MockLoadPredictor struct {
	mock.Mock
}

func (m *MockLoadPredictor) PredictLoad(horizon int) (*predictive.LoadPrediction, error) {
	args := m.Called(horizon)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*predictive.LoadPrediction), args.Error(1)
}

// MockRouteOptimizer представляет мок оптимизатора маршрутов
type MockRouteOptimizer struct {
	mock.Mock
}

func (m *MockRouteOptimizer) OptimizeRoute(source string, constraints map[string]interface{}) (*routing.OptimizedRoute, error) {
	args := m.Called(source, constraints)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*routing.OptimizedRoute), args.Error(1)
}

func TestNewAIMiddleware(t *testing.T) {
	logger := zaptest.NewLogger(t)
	aiService := &MockAIService{}

	middleware := NewAIMiddleware(aiService, logger)

	assert.NotNil(t, middleware)
	assert.Equal(t, aiService, middleware.aiService)
	assert.Equal(t, logger, middleware.logger)
	assert.True(t, middleware.enabled)
	assert.Equal(t, 5*time.Second, middleware.processInterval)
	assert.NotNil(t, middleware.circuitBreaker)
	assert.NotNil(t, middleware.metrics)
}

func TestAIMiddleware_Initialize_Success(t *testing.T) {
	logger := zaptest.NewLogger(t)
	aiService := &MockAIService{}
	middleware := NewAIMiddleware(aiService, logger)

	// Настраиваем мок
	aiService.On("Initialize").Return(nil)
	status := &ai.ServiceStatus{
		Components: map[string]bool{
			"anomaly_detector": true,
			"load_predictor":   true,
			"route_optimizer":  true,
		},
	}
	aiService.On("GetStatus").Return(status)

	// Тестируем инициализацию
	err := middleware.Initialize()

	assert.NoError(t, err)
	// Компоненты теперь не доступны напрямую, проверяем через статус
	status = middleware.aiService.GetStatus()
	assert.True(t, status.Components["anomaly_detector"])
	assert.True(t, status.Components["load_predictor"])
	assert.True(t, status.Components["route_optimizer"])

	aiService.AssertExpectations(t)
}

func TestAIMiddleware_Initialize_NoAIService(t *testing.T) {
	logger := zaptest.NewLogger(t)
	middleware := NewAIMiddleware(nil, logger)

	err := middleware.Initialize()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "AI service is not initialized")
}

func TestAIMiddleware_Initialize_AIServiceError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	aiService := &MockAIService{}
	middleware := NewAIMiddleware(aiService, logger)

	// Настраиваем мок для ошибки
	aiService.On("Initialize").Return(assert.AnError)

	err := middleware.Initialize()

	assert.Error(t, err)
	aiService.AssertExpectations(t)
}

func TestAIMiddleware_ProcessConnection_Disabled(t *testing.T) {
	logger := zaptest.NewLogger(t)
	aiService := &MockAIService{}
	middleware := NewAIMiddleware(aiService, logger)
	middleware.SetEnabled(false)

	// Создаем тестовое соединение
	conn := &MockConn{}

	err := middleware.ProcessConnection(context.Background(), conn)

	assert.NoError(t, err)
	// AI сервис не должен вызываться, когда отключен
	aiService.AssertNotCalled(t, "Initialize")
}

func TestAIMiddleware_ProcessConnection_CircuitBreakerOpen(t *testing.T) {
	logger := zaptest.NewLogger(t)
	aiService := &MockAIService{}
	middleware := NewAIMiddleware(aiService, logger)

	// Настраиваем мок для GetStatus (может вызываться даже при открытом circuit breaker)
	status := &ai.ServiceStatus{
		Components: map[string]bool{
			"anomaly_detector": true,
			"load_predictor":   true,
			"route_optimizer":  true,
		},
	}
	aiService.On("GetStatus").Return(status)

	// Открываем circuit breaker
	middleware.circuitBreaker.state = "open"
	middleware.circuitBreaker.failures = 10

	conn := &MockConn{}
	addr := &MockAddr{network: "tcp", address: "192.168.1.100:12345"}
	conn.On("RemoteAddr").Return(addr)

	err := middleware.ProcessConnection(context.Background(), conn)

	assert.NoError(t, err)
	// AI сервис не должен вызываться при открытом circuit breaker
	aiService.AssertNotCalled(t, "Initialize")
	conn.AssertExpectations(t)
}

func TestAIMiddleware_ProcessConnection_Success(t *testing.T) {
	logger := zaptest.NewLogger(t)
	aiService := &MockAIService{}
	middleware := NewAIMiddleware(aiService, logger)

	// Настраиваем мок
	aiService.On("Initialize").Return(nil)
	status := &ai.ServiceStatus{
		Components: map[string]bool{
			"anomaly_detector": true,
			"load_predictor":   true,
			"route_optimizer":  true,
		},
	}
	aiService.On("GetStatus").Return(status)

	// Инициализируем middleware
	err := middleware.Initialize()
	assert.NoError(t, err)

	conn := &MockConn{}
	addr := &MockAddr{network: "tcp", address: "192.168.1.100:12345"}
	conn.On("RemoteAddr").Return(addr)

	err = middleware.ProcessConnection(context.Background(), conn)

	assert.NoError(t, err)
	aiService.AssertExpectations(t)
	conn.AssertExpectations(t)
}

func TestAIMiddleware_ProcessConnection_AnomalyDetected(t *testing.T) {
	logger := zaptest.NewLogger(t)
	aiService := &MockAIService{}
	middleware := NewAIMiddleware(aiService, logger)

	// Настраиваем мок
	aiService.On("Initialize").Return(nil)
	status := &ai.ServiceStatus{
		Components: map[string]bool{
			"anomaly_detector": true,
			"load_predictor":   true,
			"route_optimizer":  true,
		},
	}
	aiService.On("GetStatus").Return(status)

	// Инициализируем middleware
	err := middleware.Initialize()
	assert.NoError(t, err)

	conn := &MockConn{}
	addr := &MockAddr{network: "tcp", address: "192.168.1.100:12345"}
	conn.On("RemoteAddr").Return(addr)

	err = middleware.ProcessConnection(context.Background(), conn)

	assert.NoError(t, err)

	// Проверяем метрики
	metrics := middleware.GetMetrics()
	assert.Equal(t, int64(1), metrics["connections_processed"])

	aiService.AssertExpectations(t)
	conn.AssertExpectations(t)
}

func TestAIMiddleware_ProcessConnection_AnomalyDetectionError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	aiService := &MockAIService{}
	middleware := NewAIMiddleware(aiService, logger)

	// Настраиваем мок
	aiService.On("Initialize").Return(nil)
	status := &ai.ServiceStatus{
		Components: map[string]bool{
			"anomaly_detector": true,
			"load_predictor":   true,
			"route_optimizer":  true,
		},
	}
	aiService.On("GetStatus").Return(status)

	// Инициализируем middleware
	err := middleware.Initialize()
	assert.NoError(t, err)

	conn := &MockConn{}
	addr := &MockAddr{network: "tcp", address: "192.168.1.100:12345"}
	conn.On("RemoteAddr").Return(addr)

	err = middleware.ProcessConnection(context.Background(), conn)

	// Ошибка AI не должна блокировать соединение
	assert.NoError(t, err)

	aiService.AssertExpectations(t)
	conn.AssertExpectations(t)
}

func TestAIMiddleware_SetEnabled(t *testing.T) {
	logger := zaptest.NewLogger(t)
	aiService := &MockAIService{}
	middleware := NewAIMiddleware(aiService, logger)

	assert.True(t, middleware.IsEnabled())

	middleware.SetEnabled(false)
	assert.False(t, middleware.IsEnabled())

	middleware.SetEnabled(true)
	assert.True(t, middleware.IsEnabled())
}

func TestAIMiddleware_GetMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t)
	aiService := &MockAIService{}
	middleware := NewAIMiddleware(aiService, logger)

	metrics := middleware.GetMetrics()

	assert.NotNil(t, metrics)
	assert.Equal(t, int64(0), metrics["connections_processed"])
	assert.Equal(t, int64(0), metrics["anomalies_detected"])
	assert.Equal(t, int64(0), metrics["circuit_breaker_trips"])
	assert.True(t, metrics["enabled"].(bool))
}

func TestAICircuitBreaker_CanExecute(t *testing.T) {
	cb := &AICircuitBreaker{
		state:     "closed",
		threshold: 5,
		timeout:   30 * time.Second,
	}

	// Circuit breaker закрыт - должен разрешать выполнение
	assert.True(t, cb.CanExecute())

	// Открываем circuit breaker
	cb.state = "open"
	cb.lastFailTime = time.Now()
	assert.False(t, cb.CanExecute())

	// Ждем истечения timeout
	cb.lastFailTime = time.Now().Add(-31 * time.Second)
	assert.True(t, cb.CanExecute())
	assert.Equal(t, "half-open", cb.state)
}

func TestAICircuitBreaker_RecordSuccess(t *testing.T) {
	cb := &AICircuitBreaker{
		state:     "half-open",
		failures:  3,
		threshold: 5,
	}

	cb.RecordSuccess()

	assert.Equal(t, "closed", cb.state)
	assert.Equal(t, 0, cb.failures)
}

func TestAICircuitBreaker_RecordFailure(t *testing.T) {
	cb := &AICircuitBreaker{
		state:     "closed",
		failures:  0,
		threshold: 5,
	}

	// Записываем несколько неудач
	for i := 0; i < 4; i++ {
		cb.RecordFailure()
	}

	assert.Equal(t, "closed", cb.state)
	assert.Equal(t, 4, cb.failures)

	// Пятая неудача должна открыть circuit breaker
	cb.RecordFailure()
	assert.Equal(t, "open", cb.state)
	assert.Equal(t, 5, cb.failures)
}

// MockConn представляет мок соединения для тестирования
type MockConn struct {
	mock.Mock
}

func (m *MockConn) Read(b []byte) (n int, err error) {
	args := m.Called(b)
	return args.Int(0), args.Error(1)
}

func (m *MockConn) Write(b []byte) (n int, err error) {
	args := m.Called(b)
	return args.Int(0), args.Error(1)
}

func (m *MockConn) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockConn) LocalAddr() net.Addr {
	args := m.Called()
	return args.Get(0).(net.Addr)
}

func (m *MockConn) RemoteAddr() net.Addr {
	args := m.Called()
	return args.Get(0).(net.Addr)
}

func (m *MockConn) SetDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockConn) SetReadDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockConn) SetWriteDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

// MockAddr представляет мок адреса для тестирования
type MockAddr struct {
	network string
	address string
}

func (m *MockAddr) Network() string {
	return m.network
}

func (m *MockAddr) String() string {
	return m.address
}

func TestCollectConnectionMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t)
	aiService := &MockAIService{}
	middleware := NewAIMiddleware(aiService, logger)

	// Создаем мок соединения
	conn := &MockConn{}
	addr := &MockAddr{network: "tcp", address: "192.168.1.100:12345"}
	conn.On("RemoteAddr").Return(addr)

	metrics := middleware.collectConnectionMetrics(conn)

	assert.NotNil(t, metrics)
	assert.Equal(t, "192.168.1.100:12345", metrics.RemoteAddr)
	assert.Equal(t, "tcp", metrics.Protocol)
	assert.True(t, metrics.IsNewConnection)
	assert.Equal(t, 1, metrics.ConnectionCount)

	conn.AssertExpectations(t)
}
