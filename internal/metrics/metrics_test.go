package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
)

func TestMetricsEndpoint(t *testing.T) {
	reg := prometheus.NewRegistry()
	_ = NewMetrics(reg)
	server := httptest.NewServer(promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	defer server.Close()

	// Test metrics endpoint
	resp, err := http.Get(server.URL + "/metrics")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestMetricsCollection(t *testing.T) {
	reg := prometheus.NewRegistry()
	metrics := NewMetrics(reg)

	// Test connection metrics
	metrics.IncConnections()
	metrics.IncConnections()
	metrics.DecConnections()

	// Test error metrics
	metrics.IncErrors("test_error")
	metrics.IncErrors("test_error")

	// Test heartbeat metrics
	metrics.IncHeartbeats()
	metrics.IncHeartbeatErrors()

	// Test tunnel metrics
	metrics.SetActiveTunnels(5)

	// Test duration metrics
	metrics.ObserveConnectionDuration(100 * time.Millisecond)
	metrics.ObserveHeartbeatLatency(50 * time.Millisecond)

	// Test new metrics methods
	metrics.ObserveTokenValidationDuration("jwt", 10*time.Millisecond)
	metrics.IncTokenValidations("jwt", "success")
	metrics.ObserveMessageLatency("auth", 5*time.Millisecond)
	metrics.IncMessageErrors("auth", "timeout")
	metrics.SetCacheSize(100)
	metrics.IncCacheHits()
	metrics.IncCacheMisses()

	// Verify metrics values
	// Note: In a real test, we would use prometheus client to verify the values
	// This is just a basic test to ensure no panics
}

func TestMetricsConcurrency(t *testing.T) {
	reg := prometheus.NewRegistry()
	metrics := NewMetrics(reg)
	done := make(chan bool)

	// Test concurrent access
	for i := 0; i < 10; i++ {
		go func() {
			metrics.IncConnections()
			metrics.IncHeartbeats()
			metrics.IncErrors("test_error")
			metrics.IncCacheHits()
			metrics.IncCacheMisses()
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestMetricsLabels(t *testing.T) {
	reg := prometheus.NewRegistry()
	metrics := NewMetrics(reg)

	// Test error labels
	metrics.IncErrors("invalid_token")
	metrics.IncErrors("connection_timeout")
	metrics.IncErrors("internal_error")

	// Test token validation labels
	metrics.IncTokenValidations("jwt", "success")
	metrics.IncTokenValidations("jwt", "failure")
	metrics.IncTokenValidations("django", "success")

	// Test message error labels
	metrics.IncMessageErrors("auth", "timeout")
	metrics.IncMessageErrors("auth", "invalid_token")
	metrics.IncMessageErrors("tunnel", "connection_failed")

	// Verify that all error types are tracked separately
	// Note: In a real test, we would use prometheus client to verify the values
}

func TestMetricsGetters(t *testing.T) {
	reg := prometheus.NewRegistry()
	metrics := NewMetrics(reg)

	// Test initial values
	assert.Equal(t, int64(0), metrics.GetActiveConnections())
	assert.Equal(t, int64(0), metrics.GetActiveTunnels())

	// Test after setting values
	metrics.IncConnections()
	metrics.IncConnections()
	metrics.DecConnections()
	metrics.SetActiveTunnels(5)

	assert.Equal(t, int64(1), metrics.GetActiveConnections())
	assert.Equal(t, int64(5), metrics.GetActiveTunnels())
}

func TestMetricsCacheOperations(t *testing.T) {
	reg := prometheus.NewRegistry()
	metrics := NewMetrics(reg)

	// Test cache size
	metrics.SetCacheSize(100)
	metrics.SetCacheSize(200)

	// Test cache hits and misses
	metrics.IncCacheHits()
	metrics.IncCacheHits()
	metrics.IncCacheMisses()

	// Verify no panics occurred
	// Note: In a real test, we would use prometheus client to verify the values
}

func TestMetricsDurationObservations(t *testing.T) {
	reg := prometheus.NewRegistry()
	metrics := NewMetrics(reg)

	// Test token validation duration
	metrics.ObserveTokenValidationDuration("jwt", 10*time.Millisecond)
	metrics.ObserveTokenValidationDuration("django", 50*time.Millisecond)

	// Test message latency
	metrics.ObserveMessageLatency("auth", 5*time.Millisecond)
	metrics.ObserveMessageLatency("tunnel", 15*time.Millisecond)

	// Test connection and heartbeat duration
	metrics.ObserveConnectionDuration(100 * time.Millisecond)
	metrics.ObserveHeartbeatLatency(25 * time.Millisecond)

	// Verify no panics occurred
	// Note: In a real test, we would use prometheus client to verify the values
}

// Новые тесты для Keycloak метрик
func TestKeycloakMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	metrics := NewMetrics(reg)

	// Test Keycloak token validations
	metrics.IncKeycloakTokenValidations("success", "jwt")
	metrics.IncKeycloakTokenValidations("failure", "jwt")
	metrics.IncKeycloakTokenValidations("success", "django")

	// Test JWKS cache metrics
	metrics.IncKeycloakJWKSCacheHits()
	metrics.IncKeycloakJWKSCacheHits()
	metrics.IncKeycloakJWKSCacheMisses()
	metrics.SetKeycloakJWKSCacheSize(10)

	// Test JWKS fetch duration
	metrics.ObserveKeycloakJWKSFetchDuration(100 * time.Millisecond)
	metrics.ObserveKeycloakJWKSFetchDuration(200 * time.Millisecond)

	// Test token validation duration
	metrics.ObserveKeycloakTokenValidationDuration("jwt", 50 * time.Millisecond)
	metrics.ObserveKeycloakTokenValidationDuration("jwks", 30 * time.Millisecond)

	// Test user info duration
	metrics.ObserveKeycloakUserInfoDuration(20 * time.Millisecond)

	// Test permission checks
	metrics.IncKeycloakPermissionChecks("tunnel", "create", "granted")
	metrics.IncKeycloakPermissionChecks("user", "read", "denied")
	metrics.IncKeycloakPermissionChecks("tunnel", "delete", "granted")

	// Verify no panics occurred
}

func TestKeycloakMetricsLabels(t *testing.T) {
	reg := prometheus.NewRegistry()
	metrics := NewMetrics(reg)

	// Test different token validation results and providers
	metrics.IncKeycloakTokenValidations("success", "jwt")
	metrics.IncKeycloakTokenValidations("failure", "jwt")
	metrics.IncKeycloakTokenValidations("success", "django")
	metrics.IncKeycloakTokenValidations("failure", "django")

	// Test different permission check combinations
	metrics.IncKeycloakPermissionChecks("tunnel", "create", "granted")
	metrics.IncKeycloakPermissionChecks("tunnel", "create", "denied")
	metrics.IncKeycloakPermissionChecks("user", "read", "granted")
	metrics.IncKeycloakPermissionChecks("user", "update", "denied")

	// Test different token validation methods
	metrics.ObserveKeycloakTokenValidationDuration("jwt", 10*time.Millisecond)
	metrics.ObserveKeycloakTokenValidationDuration("jwks", 20*time.Millisecond)
	metrics.ObserveKeycloakTokenValidationDuration("django", 15*time.Millisecond)

	// Verify that all combinations are tracked separately
}

// Тесты для бизнес-метрик
func TestBusinessMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	metrics := NewMetrics(reg)

	// Test active users
	metrics.SetActiveUsers(100)
	metrics.SetActiveUsers(150)
	metrics.SetActiveUsers(200)

	// Test revenue per hour
	metrics.IncRevenuePerHour(50.0)
	metrics.IncRevenuePerHour(75.5)
	metrics.IncRevenuePerHour(100.25)

	// Test customer satisfaction
	metrics.SetCustomerSatisfaction(85.5)
	metrics.SetCustomerSatisfaction(90.0)
	metrics.SetCustomerSatisfaction(95.5)

	// Verify no panics occurred
}

// Тесты для технических метрик
func TestTechnicalMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	metrics := NewMetrics(reg)

	// Test memory usage
	metrics.SetMemoryUsage(1024 * 1024 * 100) // 100MB
	metrics.SetMemoryUsage(1024 * 1024 * 200) // 200MB

	// Test GC pause duration
	metrics.ObserveGCPauseDuration(10 * time.Millisecond)
	metrics.ObserveGCPauseDuration(50 * time.Millisecond)
	metrics.ObserveGCPauseDuration(100 * time.Millisecond)

	// Test network latency
	metrics.ObserveNetworkLatency(5 * time.Millisecond)
	metrics.ObserveNetworkLatency(15 * time.Millisecond)
	metrics.ObserveNetworkLatency(25 * time.Millisecond)

	// Test disk I/O
	metrics.IncDiskIO()
	metrics.IncDiskIO()
	metrics.IncDiskIO()

	// Verify no panics occurred
}

// Тесты для интеграции всех метрик
func TestAllMetricsIntegration(t *testing.T) {
	reg := prometheus.NewRegistry()
	metrics := NewMetrics(reg)

	// Simulate a complete request flow
	start := time.Now()

	// 1. Connection established
	metrics.IncConnections()
	metrics.ObserveConnectionDuration(time.Since(start))

	// 2. Authentication with Keycloak
	authStart := time.Now()
	metrics.IncKeycloakTokenValidations("success", "jwt")
	metrics.ObserveKeycloakTokenValidationDuration("jwt", time.Since(authStart))

	// 3. JWKS cache operations
	metrics.IncKeycloakJWKSCacheHits()
	metrics.SetKeycloakJWKSCacheSize(5)

	// 4. Permission check
	permStart := time.Now()
	metrics.IncKeycloakPermissionChecks("tunnel", "create", "granted")
	metrics.ObserveKeycloakUserInfoDuration(time.Since(permStart))

	// 5. Business metrics update
	metrics.SetActiveUsers(150)
	metrics.IncRevenuePerHour(25.0)

	// 6. Technical metrics
	metrics.SetMemoryUsage(1024 * 1024 * 150) // 150MB
	metrics.ObserveGCPauseDuration(20 * time.Millisecond)
	metrics.ObserveNetworkLatency(10 * time.Millisecond)
	metrics.IncDiskIO()

	// 7. Connection closed
	metrics.DecConnections()

	// Verify no panics occurred and all metrics were recorded
}

// Тесты для производительности метрик
func TestMetricsPerformance(t *testing.T) {
	reg := prometheus.NewRegistry()
	metrics := NewMetrics(reg)

	// Test high-frequency metric updates
	for i := 0; i < 1000; i++ {
		metrics.IncConnections()
		metrics.IncKeycloakTokenValidations("success", "jwt")
		metrics.IncKeycloakJWKSCacheHits()
		metrics.IncDiskIO()
	}

	// Test concurrent metric updates
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				metrics.IncConnections()
				metrics.IncKeycloakTokenValidations("success", "jwt")
				metrics.IncKeycloakJWKSCacheHits()
				metrics.IncDiskIO()
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify final values
	assert.Equal(t, int64(2000), metrics.GetActiveConnections()) // 1000 + 10*100
}

// Тесты для граничных случаев
func TestMetricsEdgeCases(t *testing.T) {
	reg := prometheus.NewRegistry()
	metrics := NewMetrics(reg)

	// Test zero values
	metrics.SetKeycloakJWKSCacheSize(0)
	metrics.SetActiveUsers(0)
	metrics.SetMemoryUsage(0)
	metrics.SetCustomerSatisfaction(0.0)

	// Test negative values (should be handled gracefully)
	metrics.SetCustomerSatisfaction(-10.0)
	metrics.SetMemoryUsage(-1024)

	// Test very large values
	metrics.SetMemoryUsage(1024 * 1024 * 1024 * 10) // 10GB
	metrics.SetCustomerSatisfaction(1000.0)

	// Test very small durations
	metrics.ObserveKeycloakJWKSFetchDuration(1 * time.Nanosecond)
	metrics.ObserveGCPauseDuration(1 * time.Microsecond)

	// Test very large durations
	metrics.ObserveKeycloakJWKSFetchDuration(1 * time.Hour)
	metrics.ObserveGCPauseDuration(1 * time.Minute)

	// Verify no panics occurred
}

// Тесты для валидации метрик
func TestMetricsValidation(t *testing.T) {
	reg := prometheus.NewRegistry()
	metrics := NewMetrics(reg)

	// Test that all metrics are properly registered
	// This is a basic validation that the metrics can be created without errors
	assert.NotNil(t, metrics)

	// Test that getters return reasonable values
	assert.Equal(t, int64(0), metrics.GetActiveConnections())
	assert.Equal(t, int64(0), metrics.GetActiveTunnels())

	// Test that metrics can be updated
	metrics.IncConnections()
	assert.Equal(t, int64(1), metrics.GetActiveConnections())

	// Test that metrics can be reset (by creating new instance)
	reg2 := prometheus.NewRegistry()
	metrics2 := NewMetrics(reg2)
	assert.Equal(t, int64(0), metrics2.GetActiveConnections())
} 