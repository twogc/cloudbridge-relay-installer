package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestNewGRPCMetrics(t *testing.T) {
	// Создаем новый регистр для тестов
	reg := prometheus.NewRegistry()

	// Создаем gRPC метрики
	grpcMetrics := NewGRPCMetrics(reg)

	// Проверяем, что метрики созданы
	assert.NotNil(t, grpcMetrics)
	assert.NotNil(t, grpcMetrics.connectionsTotal)
	assert.NotNil(t, grpcMetrics.connectionsActive)
	assert.NotNil(t, grpcMetrics.requestsTotal)
	assert.NotNil(t, grpcMetrics.streamsTotal)
	assert.NotNil(t, grpcMetrics.handshakeDuration)
}

func TestGRPCConnectionMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	grpcMetrics := NewGRPCMetrics(reg)

	// Тестируем инкремент соединений
	grpcMetrics.IncConnections()
	grpcMetrics.IncConnections()

	// Тестируем декремент соединений
	grpcMetrics.DecConnections()

	// Тестируем ошибки соединений
	grpcMetrics.IncConnectionsFailed("handshake_failed", "192.168.1.1")
	grpcMetrics.IncConnectionsFailed("timeout", "192.168.1.2")

	// Тестируем отброшенные соединения
	grpcMetrics.IncConnectionsDropped("client_disconnect", "192.168.1.1")
	grpcMetrics.IncConnectionsDropped("server_error", "192.168.1.2")

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestGRPCRequestMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	grpcMetrics := NewGRPCMetrics(reg)

	// Тестируем запросы
	grpcMetrics.IncRequests("GetTunnel", "TunnelService", "OK")
	grpcMetrics.IncRequests("GetTunnel", "TunnelService", "INTERNAL_ERROR")
	grpcMetrics.IncRequests("CreateTunnel", "TunnelService", "OK")

	// Тестируем запросы в полете
	grpcMetrics.IncRequestsInFlight("GetTunnel", "TunnelService")
	grpcMetrics.IncRequestsInFlight("GetTunnel", "TunnelService")
	grpcMetrics.DecRequestsInFlight("GetTunnel", "TunnelService")

	// Тестируем длительность запросов
	grpcMetrics.ObserveRequestDuration("GetTunnel", "TunnelService", 0.1)
	grpcMetrics.ObserveRequestDuration("CreateTunnel", "TunnelService", 0.5)

	// Тестируем размер запросов и ответов
	grpcMetrics.ObserveRequestSize("GetTunnel", "TunnelService", 1024)
	grpcMetrics.ObserveResponseSize("GetTunnel", "TunnelService", 2048)

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestGRPCMethodMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	grpcMetrics := NewGRPCMetrics(reg)

	// Тестируем вызовы методов
	grpcMetrics.IncMethodCalls("GetTunnel", "TunnelService", "success")
	grpcMetrics.IncMethodCalls("GetTunnel", "TunnelService", "error")
	grpcMetrics.IncMethodCalls("CreateTunnel", "TunnelService", "success")

	// Тестируем длительность методов
	grpcMetrics.ObserveMethodDuration("GetTunnel", "TunnelService", 0.05)
	grpcMetrics.ObserveMethodDuration("CreateTunnel", "TunnelService", 0.2)

	// Тестируем ошибки методов
	grpcMetrics.IncMethodErrors("GetTunnel", "TunnelService", "not_found")
	grpcMetrics.IncMethodErrors("CreateTunnel", "TunnelService", "validation_error")

	// Тестируем пропускную способность методов
	grpcMetrics.IncMethodThroughput("GetTunnel", "TunnelService", "incoming", 10)
	grpcMetrics.IncMethodThroughput("GetTunnel", "TunnelService", "outgoing", 5)

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestGRPCStreamMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	grpcMetrics := NewGRPCMetrics(reg)

	// Тестируем инкремент потоков
	grpcMetrics.IncStreams()
	grpcMetrics.IncStreams()

	// Тестируем декремент потоков
	grpcMetrics.DecStreams()

	// Тестируем создание потоков
	grpcMetrics.IncStreamsCreated("StreamTunnel", "TunnelService")
	grpcMetrics.IncStreamsCreated("StreamHeartbeat", "HeartbeatService")

	// Тестируем закрытие потоков
	grpcMetrics.IncStreamsClosed("StreamTunnel", "TunnelService", "normal")
	grpcMetrics.IncStreamsClosed("StreamHeartbeat", "HeartbeatService", "error")

	// Тестируем ошибки потоков
	grpcMetrics.IncStreamsErrors("StreamTunnel", "TunnelService", "connection_lost")
	grpcMetrics.IncStreamsErrors("StreamHeartbeat", "HeartbeatService", "timeout")

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestGRPCPerformanceMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	grpcMetrics := NewGRPCMetrics(reg)

	// Тестируем длительность handshake
	grpcMetrics.ObserveHandshakeDuration(0.1)
	grpcMetrics.ObserveHandshakeDuration(0.2)

	// Тестируем латентность соединения
	grpcMetrics.ObserveConnectionLatency("192.168.1.1", 0.05)
	grpcMetrics.ObserveConnectionLatency("192.168.1.2", 0.08)

	// Тестируем пропускную способность
	grpcMetrics.IncThroughputBytes("GetTunnel", "TunnelService", "incoming", 1024)
	grpcMetrics.IncThroughputBytes("GetTunnel", "TunnelService", "outgoing", 2048)
	grpcMetrics.IncThroughputMessages("GetTunnel", "TunnelService", "incoming", 10)
	grpcMetrics.IncThroughputMessages("GetTunnel", "TunnelService", "outgoing", 20)

	// Тестируем утилизацию полосы пропускания
	grpcMetrics.SetBandwidthUtilization("GetTunnel", "TunnelService", 75.5)
	grpcMetrics.SetBandwidthUtilization("CreateTunnel", "TunnelService", 45.2)

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestGRPCConnectionQualityMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	grpcMetrics := NewGRPCMetrics(reg)

	// Тестируем события keepalive
	grpcMetrics.IncKeepaliveEvents("192.168.1.1", "ping")
	grpcMetrics.IncKeepaliveEvents("192.168.1.1", "pong")
	grpcMetrics.IncKeepaliveEvents("192.168.1.2", "ping")

	// Тестируем ошибки keepalive
	grpcMetrics.IncKeepaliveFailures("192.168.1.1", "timeout")
	grpcMetrics.IncKeepaliveFailures("192.168.1.2", "connection_lost")

	// Тестируем время работы соединения
	grpcMetrics.SetConnectionUptime("192.168.1.1", 3600.0)
	grpcMetrics.SetConnectionUptime("192.168.1.2", 7200.0)

	// Тестируем последнюю активность
	grpcMetrics.SetLastActivity("192.168.1.1", 1640995200.0)
	grpcMetrics.SetLastActivity("192.168.1.2", 1640995300.0)

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestGRPCErrorMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	grpcMetrics := NewGRPCMetrics(reg)

	// Тестируем ошибки сервера
	grpcMetrics.IncServerErrors("internal_error", "GetTunnel", "TunnelService")
	grpcMetrics.IncServerErrors("timeout", "CreateTunnel", "TunnelService")

	// Тестируем ошибки клиента
	grpcMetrics.IncClientErrors("invalid_request", "192.168.1.1")
	grpcMetrics.IncClientErrors("authentication_failed", "192.168.1.2")

	// Тестируем события таймаута
	grpcMetrics.IncTimeoutEvents("request_timeout", "GetTunnel", "TunnelService")
	grpcMetrics.IncTimeoutEvents("connection_timeout", "CreateTunnel", "TunnelService")

	// Тестируем попытки переподключения
	grpcMetrics.IncReconnectionAttempts("192.168.1.1")
	grpcMetrics.IncReconnectionAttempts("192.168.1.2")

	// Тестируем события circuit breaker
	grpcMetrics.IncCircuitBreakerEvents("GetTunnel", "TunnelService", "opened")
	grpcMetrics.IncCircuitBreakerEvents("CreateTunnel", "TunnelService", "closed")

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestGRPCSecurityMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	grpcMetrics := NewGRPCMetrics(reg)

	// Тестируем ошибки аутентификации
	grpcMetrics.IncAuthenticationFailures("192.168.1.1", "invalid_token")
	grpcMetrics.IncAuthenticationFailures("192.168.1.2", "expired_token")

	// Тестируем ошибки авторизации
	grpcMetrics.IncAuthorizationFailures("192.168.1.1", "GetTunnel", "TunnelService")
	grpcMetrics.IncAuthorizationFailures("192.168.1.2", "CreateTunnel", "TunnelService")

	// Тестируем нарушения rate limit
	grpcMetrics.IncRateLimitViolations("192.168.1.1", "requests_per_second")
	grpcMetrics.IncRateLimitViolations("192.168.1.2", "concurrent_connections")

	// Тестируем нарушения RBAC
	grpcMetrics.IncRBACViolations("192.168.1.1", "GetTunnel", "TunnelService", "insufficient_permissions")
	grpcMetrics.IncRBACViolations("192.168.1.2", "CreateTunnel", "TunnelService", "role_denied")

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestGRPCMiddlewareMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	grpcMetrics := NewGRPCMetrics(reg)

	// Тестируем длительность middleware
	grpcMetrics.ObserveMiddlewareDuration("auth", "GetTunnel", "TunnelService", 0.01)
	grpcMetrics.ObserveMiddlewareDuration("logging", "CreateTunnel", "TunnelService", 0.005)

	// Тестируем ошибки middleware
	grpcMetrics.IncMiddlewareErrors("auth", "token_validation_failed")
	grpcMetrics.IncMiddlewareErrors("rate_limit", "quota_exceeded")

	// Тестируем вызовы interceptor'ов
	grpcMetrics.IncInterceptorCalls("auth_interceptor", "GetTunnel", "TunnelService")
	grpcMetrics.IncInterceptorCalls("logging_interceptor", "CreateTunnel", "TunnelService")

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestGRPCTenantMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	grpcMetrics := NewGRPCMetrics(reg)

	// Тестируем запросы тенантов
	grpcMetrics.IncTenantRequests("tenant1", "GetTunnel", "TunnelService")
	grpcMetrics.IncTenantRequests("tenant2", "CreateTunnel", "TunnelService")

	// Тестируем соединения тенантов
	grpcMetrics.SetTenantConnections("tenant1", 5)
	grpcMetrics.SetTenantConnections("tenant2", 3)

	// Тестируем полосу пропускания тенантов
	grpcMetrics.IncTenantBandwidth("tenant1", "incoming", 1024)
	grpcMetrics.IncTenantBandwidth("tenant1", "outgoing", 2048)
	grpcMetrics.IncTenantBandwidth("tenant2", "incoming", 512)

	// Тестируем ошибки тенантов
	grpcMetrics.IncTenantErrors("tenant1", "quota_exceeded")
	grpcMetrics.IncTenantErrors("tenant2", "authentication_failed")

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestGRPCResourceMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	grpcMetrics := NewGRPCMetrics(reg)

	// Тестируем использование памяти
	grpcMetrics.SetMemoryUsage("server", 1024*1024)
	grpcMetrics.SetMemoryUsage("connections", 512*1024)

	// Тестируем использование CPU
	grpcMetrics.SetCPUUsage("server", 25.5)
	grpcMetrics.SetCPUUsage("connections", 15.2)

	// Тестируем количество горутин
	grpcMetrics.SetGoroutineCount("server", 10)
	grpcMetrics.SetGoroutineCount("connections", 5)

	// Тестируем количество файловых дескрипторов
	grpcMetrics.SetFileDescriptorCount(20)
	grpcMetrics.SetFileDescriptorCount(15)

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestGRPCConfigurationMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	grpcMetrics := NewGRPCMetrics(reg)

	// Тестируем статус сервера
	grpcMetrics.SetServerStatus(true)
	grpcMetrics.SetServerStatus(false)

	// Тестируем версию сервера
	grpcMetrics.SetServerVersion("1.0.0")
	grpcMetrics.SetServerVersion("1.1.0")

	// Тестируем обновления конфигурации
	grpcMetrics.IncConfigUpdates("server_config")
	grpcMetrics.IncConfigUpdates("middleware_config")

	// Тестируем ошибки конфигурации
	grpcMetrics.IncConfigErrors("invalid_config")
	grpcMetrics.IncConfigErrors("parse_error")

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestGRPCMetricsIntegration(t *testing.T) {
	// Создаем основной Metrics с gRPC метриками
	reg := prometheus.NewRegistry()
	mainMetrics := NewMetrics(reg)

	// Проверяем, что gRPC метрики доступны
	grpcMetrics := mainMetrics.GetGRPCMetrics()
	assert.NotNil(t, grpcMetrics)

	// Тестируем методы основного Metrics
	mainMetrics.IncGRPCConnections()
	mainMetrics.DecGRPCConnections()
	mainMetrics.IncGRPCConnectionsFailed("test_reason", "192.168.1.1")
	mainMetrics.IncGRPCRequests("GetTunnel", "TunnelService", "OK")
	mainMetrics.IncGRPCRequestsInFlight("GetTunnel", "TunnelService")
	mainMetrics.DecGRPCRequestsInFlight("GetTunnel", "TunnelService")
	mainMetrics.ObserveGRPCRequestDuration("GetTunnel", "TunnelService", 0.1)
	mainMetrics.ObserveGRPCRequestSize("GetTunnel", "TunnelService", 1024)
	mainMetrics.ObserveGRPCResponseSize("GetTunnel", "TunnelService", 2048)
	mainMetrics.IncGRPCMethodCalls("GetTunnel", "TunnelService", "success")
	mainMetrics.ObserveGRPCMethodDuration("GetTunnel", "TunnelService", 0.05)
	mainMetrics.IncGRPCMethodErrors("GetTunnel", "TunnelService", "not_found")
	mainMetrics.IncGRPCMethodThroughput("GetTunnel", "TunnelService", "incoming", 10)
	mainMetrics.IncGRPCStreams()
	mainMetrics.DecGRPCStreams()
	mainMetrics.IncGRPCStreamsCreated("StreamTunnel", "TunnelService")
	mainMetrics.IncGRPCStreamsClosed("StreamTunnel", "TunnelService", "normal")
	mainMetrics.IncGRPCStreamsErrors("StreamTunnel", "TunnelService", "timeout")
	mainMetrics.ObserveGRPCHandshakeDuration(0.1)
	mainMetrics.ObserveGRPCConnectionLatency("192.168.1.1", 0.05)
	mainMetrics.IncGRPCThroughputBytes("GetTunnel", "TunnelService", "incoming", 1024)
	mainMetrics.IncGRPCThroughputMessages("GetTunnel", "TunnelService", "incoming", 10)
	mainMetrics.SetGRPCBandwidthUtilization("GetTunnel", "TunnelService", 75.5)
	mainMetrics.IncGRPCKeepaliveEvents("192.168.1.1", "ping")
	mainMetrics.IncGRPCKeepaliveFailures("192.168.1.1", "timeout")
	mainMetrics.SetGRPCConnectionUptime("192.168.1.1", 3600.0)
	mainMetrics.SetGRPCLastActivity("192.168.1.1", 1640995200.0)
	mainMetrics.IncGRPCServerErrors("internal_error", "GetTunnel", "TunnelService")
	mainMetrics.IncGRPCClientErrors("invalid_request", "192.168.1.1")
	mainMetrics.IncGRPCTimeoutEvents("request_timeout", "GetTunnel", "TunnelService")
	mainMetrics.IncGRPCReconnectionAttempts("192.168.1.1")
	mainMetrics.IncGRPCCircuitBreakerEvents("GetTunnel", "TunnelService", "opened")
	mainMetrics.IncGRPCAuthenticationFailures("192.168.1.1", "invalid_token")
	mainMetrics.IncGRPCAuthorizationFailures("192.168.1.1", "GetTunnel", "TunnelService")
	mainMetrics.IncGRPCRateLimitViolations("192.168.1.1", "requests_per_second")
	mainMetrics.IncGRPCRBACViolations("192.168.1.1", "GetTunnel", "TunnelService", "insufficient_permissions")
	mainMetrics.ObserveGRPCMiddlewareDuration("auth", "GetTunnel", "TunnelService", 0.01)
	mainMetrics.IncGRPCMiddlewareErrors("auth", "token_validation_failed")
	mainMetrics.IncGRPCInterceptorCalls("auth_interceptor", "GetTunnel", "TunnelService")
	mainMetrics.IncGRPCTenantRequests("tenant1", "GetTunnel", "TunnelService")
	mainMetrics.SetGRPCTenantConnections("tenant1", 5)
	mainMetrics.IncGRPCTenantBandwidth("tenant1", "incoming", 1024)
	mainMetrics.IncGRPCTenantErrors("tenant1", "quota_exceeded")
	mainMetrics.SetGRPCMemoryUsage("server", 1024*1024)
	mainMetrics.SetGRPCCPUUsage("server", 25.5)
	mainMetrics.SetGRPCGoroutineCount("server", 10)
	mainMetrics.SetGRPCFileDescriptorCount(20)
	mainMetrics.SetGRPCServerStatus(true)
	mainMetrics.SetGRPCServerVersion("1.0.0")
	mainMetrics.IncGRPCConfigUpdates("server_config")
	mainMetrics.IncGRPCConfigErrors("invalid_config")

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}



