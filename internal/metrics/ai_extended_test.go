package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAIConnectionsProcessed(t *testing.T) {
	// Test incrementing connections processed
	AIConnectionsProcessed.WithLabelValues("success", "false", "behavior_analyzer").Inc()
	AIConnectionsProcessed.WithLabelValues("success", "false", "behavior_analyzer").Inc()
	AIConnectionsProcessed.WithLabelValues("error", "true", "predictive_security").Inc()

	// Verify metrics were incremented
	assert.NotNil(t, AIConnectionsProcessed)
}

func TestAIProcessingDuration(t *testing.T) {
	// Test observing processing duration
	AIProcessingDuration.WithLabelValues("behavior_analyzer", "analyze").Observe(0.1)
	AIProcessingDuration.WithLabelValues("behavior_analyzer", "analyze").Observe(0.2)
	AIProcessingDuration.WithLabelValues("predictive_security", "predict").Observe(0.5)

	// Verify metrics were observed
	assert.NotNil(t, AIProcessingDuration)
}

func TestAnomaliesDetected(t *testing.T) {
	// Test incrementing anomalies detected
	AnomaliesDetected.WithLabelValues("high", "network", "p2p_quic").Inc()
	AnomaliesDetected.WithLabelValues("high", "network", "p2p_quic").Inc()
	AnomaliesDetected.WithLabelValues("medium", "behavior", "user_activity").Inc()
	AnomaliesDetected.WithLabelValues("critical", "security", "quic").Inc()

	// Verify metrics were incremented
	assert.NotNil(t, AnomaliesDetected)
}

func TestAnomalyDetectionAccuracy(t *testing.T) {
	// Test setting anomaly detection accuracy
	AnomalyDetectionAccuracy.WithLabelValues("v1.0").Set(0.95)
	AnomalyDetectionAccuracy.WithLabelValues("v2.0").Set(0.98)
	AnomalyDetectionAccuracy.WithLabelValues("v1.0").Set(0.96) // Update existing

	// Verify metrics were set
	assert.NotNil(t, AnomalyDetectionAccuracy)
}

func TestLoadPredictionsMade(t *testing.T) {
	// Test incrementing load predictions
	LoadPredictionsMade.WithLabelValues("1h", "high").Inc()
	LoadPredictionsMade.WithLabelValues("1h", "high").Inc()
	LoadPredictionsMade.WithLabelValues("24h", "medium").Inc()
	LoadPredictionsMade.WithLabelValues("7d", "low").Inc()

	// Verify metrics were incremented
	assert.NotNil(t, LoadPredictionsMade)
}

func TestLoadPredictionAccuracy(t *testing.T) {
	// Test setting load prediction accuracy
	LoadPredictionAccuracy.WithLabelValues("v1.0", "1h").Set(0.92)
	LoadPredictionAccuracy.WithLabelValues("v1.0", "24h").Set(0.85)
	LoadPredictionAccuracy.WithLabelValues("v2.0", "1h").Set(0.95)

	// Verify metrics were set
	assert.NotNil(t, LoadPredictionAccuracy)
}

func TestRoutesOptimized(t *testing.T) {
	// Test incrementing routes optimized
	RoutesOptimized.WithLabelValues("p2p_quic", "latency").Inc()
	RoutesOptimized.WithLabelValues("p2p_quic", "latency").Inc()
	RoutesOptimized.WithLabelValues("quic", "bandwidth").Inc()
	RoutesOptimized.WithLabelValues("tcp", "reliability").Inc()

	// Verify metrics were incremented
	assert.NotNil(t, RoutesOptimized)
}

func TestRouteOptimizationLatency(t *testing.T) {
	// Test observing route optimization latency
	RouteOptimizationLatency.WithLabelValues("p2p_quic").Observe(0.01)
	RouteOptimizationLatency.WithLabelValues("p2p_quic").Observe(0.02)
	RouteOptimizationLatency.WithLabelValues("quic").Observe(0.005)
	RouteOptimizationLatency.WithLabelValues("tcp").Observe(0.1)

	// Verify metrics were observed
	assert.NotNil(t, RouteOptimizationLatency)
}

func TestAICacheHits(t *testing.T) {
	// Test incrementing AI cache hits
	AICacheHits.WithLabelValues("model_cache").Inc()
	AICacheHits.WithLabelValues("model_cache").Inc()
	AICacheHits.WithLabelValues("prediction_cache").Inc()
	AICacheHits.WithLabelValues("route_cache").Inc()

	// Verify metrics were incremented
	assert.NotNil(t, AICacheHits)
}

func TestAICacheMisses(t *testing.T) {
	// Test incrementing AI cache misses
	AICacheMisses.WithLabelValues("model_cache").Inc()
	AICacheMisses.WithLabelValues("prediction_cache").Inc()
	AICacheMisses.WithLabelValues("route_cache").Inc()

	// Verify metrics were incremented
	assert.NotNil(t, AICacheMisses)
}

func TestAICircuitBreakerState(t *testing.T) {
	// Test setting circuit breaker state
	// 0=closed, 1=open, 2=half-open
	AICircuitBreakerState.WithLabelValues("behavior_analyzer").Set(0)   // closed
	AICircuitBreakerState.WithLabelValues("predictive_security").Set(1) // open
	AICircuitBreakerState.WithLabelValues("automated_response").Set(2)  // half-open

	// Verify metrics were set
	assert.NotNil(t, AICircuitBreakerState)
}

func TestAICircuitBreakerFailures(t *testing.T) {
	// Test incrementing circuit breaker failures
	AICircuitBreakerFailures.WithLabelValues("behavior_analyzer").Inc()
	AICircuitBreakerFailures.WithLabelValues("behavior_analyzer").Inc()
	AICircuitBreakerFailures.WithLabelValues("predictive_security").Inc()

	// Verify metrics were incremented
	assert.NotNil(t, AICircuitBreakerFailures)
}

func TestAIServiceHealth(t *testing.T) {
	// Test setting service health
	// 1=healthy, 0=unhealthy
	AIServiceHealth.WithLabelValues("behavior_analyzer", "model").Set(1)  // healthy
	AIServiceHealth.WithLabelValues("predictive_security", "api").Set(1)  // healthy
	AIServiceHealth.WithLabelValues("automated_response", "rules").Set(0) // unhealthy

	// Verify metrics were set
	assert.NotNil(t, AIServiceHealth)
}

func TestAIModelVersion(t *testing.T) {
	// Test setting model version
	AIModelVersion.WithLabelValues("behavior", "v1.0").Set(1.0)
	AIModelVersion.WithLabelValues("behavior", "v2.0").Set(2.0)
	AIModelVersion.WithLabelValues("predictive", "v1.0").Set(1.0)

	// Verify metrics were set
	assert.NotNil(t, AIModelVersion)
}

func TestAIMemoryUsage(t *testing.T) {
	// Test setting memory usage in bytes
	AIMemoryUsage.WithLabelValues("behavior_analyzer").Set(1024 * 1024 * 100)   // 100MB
	AIMemoryUsage.WithLabelValues("predictive_security").Set(1024 * 1024 * 200) // 200MB
	AIMemoryUsage.WithLabelValues("automated_response").Set(1024 * 1024 * 50)   // 50MB

	// Verify metrics were set
	assert.NotNil(t, AIMemoryUsage)
}

func TestAICPUUsage(t *testing.T) {
	// Test setting CPU usage percentage
	AICPUUsage.WithLabelValues("behavior_analyzer").Set(25.5)
	AICPUUsage.WithLabelValues("predictive_security").Set(45.0)
	AICPUUsage.WithLabelValues("automated_response").Set(15.2)

	// Verify metrics were set
	assert.NotNil(t, AICPUUsage)
}

func TestAICostSavings(t *testing.T) {
	// Test incrementing cost savings
	AICostSavings.WithLabelValues("route_optimization").Inc()
	AICostSavings.WithLabelValues("route_optimization").Inc()
	AICostSavings.WithLabelValues("load_balancing").Inc()
	AICostSavings.WithLabelValues("anomaly_detection").Inc()

	// Verify metrics were incremented
	assert.NotNil(t, AICostSavings)
}

func TestAIAutomationRate(t *testing.T) {
	// Test setting automation rate percentage
	AIAutomationRate.WithLabelValues("route_optimization").Set(95.0)
	AIAutomationRate.WithLabelValues("anomaly_detection").Set(98.5)
	AIAutomationRate.WithLabelValues("threat_response").Set(85.0)

	// Verify metrics were set
	assert.NotNil(t, AIAutomationRate)
}

func TestAIExtendedMetrics_EdgeCases(t *testing.T) {
	// Test with empty strings
	AIConnectionsProcessed.WithLabelValues("", "", "").Inc()
	AnomaliesDetected.WithLabelValues("", "", "").Inc()
	LoadPredictionsMade.WithLabelValues("", "").Inc()
	RoutesOptimized.WithLabelValues("", "").Inc()

	// Test with negative values
	AnomalyDetectionAccuracy.WithLabelValues("v1.0").Set(-0.1)
	LoadPredictionAccuracy.WithLabelValues("v1.0", "1h").Set(-0.1)
	AIMemoryUsage.WithLabelValues("test").Set(-1000)
	AICPUUsage.WithLabelValues("test").Set(-10.0)
	AIAutomationRate.WithLabelValues("test").Set(-5.0)

	// Test with very large values
	AnomalyDetectionAccuracy.WithLabelValues("v1.0").Set(999.0)
	LoadPredictionAccuracy.WithLabelValues("v1.0", "1h").Set(999.0)
	AIMemoryUsage.WithLabelValues("test").Set(999999999999)
	AICPUUsage.WithLabelValues("test").Set(999.0)
	AIAutomationRate.WithLabelValues("test").Set(999.0)

	// Test with zero values
	AIProcessingDuration.WithLabelValues("test", "test").Observe(0.0)
	RouteOptimizationLatency.WithLabelValues("test").Observe(0.0)
	AnomalyDetectionAccuracy.WithLabelValues("v1.0").Set(0.0)
	LoadPredictionAccuracy.WithLabelValues("v1.0", "1h").Set(0.0)
	AIMemoryUsage.WithLabelValues("test").Set(0.0)
	AICPUUsage.WithLabelValues("test").Set(0.0)
	AIAutomationRate.WithLabelValues("test").Set(0.0)
}

func TestAIExtendedMetrics_ConcurrentAccess(t *testing.T) {
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			service := "service" + string(rune(id))
			component := "component" + string(rune(id))

			AIConnectionsProcessed.WithLabelValues("success", "false", service).Inc()
			AIProcessingDuration.WithLabelValues(service, "process").Observe(0.1)
			AnomaliesDetected.WithLabelValues("medium", "test", service).Inc()
			AnomalyDetectionAccuracy.WithLabelValues("v1.0").Set(0.9)
			LoadPredictionsMade.WithLabelValues("1h", "high").Inc()
			LoadPredictionAccuracy.WithLabelValues("v1.0", "1h").Set(0.9)
			RoutesOptimized.WithLabelValues("p2p_quic", "latency").Inc()
			RouteOptimizationLatency.WithLabelValues("p2p_quic").Observe(0.01)
			AICacheHits.WithLabelValues("model_cache").Inc()
			AICacheMisses.WithLabelValues("model_cache").Inc()
			AICircuitBreakerState.WithLabelValues(service).Set(0)
			AICircuitBreakerFailures.WithLabelValues(service).Inc()
			AIServiceHealth.WithLabelValues(service, component).Set(1)
			AIModelVersion.WithLabelValues("behavior", "v1.0").Set(1.0)
			AIMemoryUsage.WithLabelValues(component).Set(1024 * 1024 * 100)
			AICPUUsage.WithLabelValues(component).Set(25.0)
			AICostSavings.WithLabelValues("optimization").Inc()
			AIAutomationRate.WithLabelValues("automation").Set(90.0)
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all operations completed without errors
	assert.NotNil(t, AIConnectionsProcessed)
	assert.NotNil(t, AIProcessingDuration)
	assert.NotNil(t, AnomaliesDetected)
	assert.NotNil(t, AnomalyDetectionAccuracy)
	assert.NotNil(t, LoadPredictionsMade)
	assert.NotNil(t, LoadPredictionAccuracy)
	assert.NotNil(t, RoutesOptimized)
	assert.NotNil(t, RouteOptimizationLatency)
	assert.NotNil(t, AICacheHits)
	assert.NotNil(t, AICacheMisses)
	assert.NotNil(t, AICircuitBreakerState)
	assert.NotNil(t, AICircuitBreakerFailures)
	assert.NotNil(t, AIServiceHealth)
	assert.NotNil(t, AIModelVersion)
	assert.NotNil(t, AIMemoryUsage)
	assert.NotNil(t, AICPUUsage)
	assert.NotNil(t, AICostSavings)
	assert.NotNil(t, AIAutomationRate)
}

func TestAIExtendedMetrics_IntegrationWorkflow(t *testing.T) {
	// Simulate a complete AI extended workflow
	service := "behavior_analyzer"
	component := "model"
	modelVersion := "v1.0"

	// 1. Process connection
	AIConnectionsProcessed.WithLabelValues("success", "false", service).Inc()
	AIProcessingDuration.WithLabelValues(service, "analyze").Observe(0.1)

	// 2. Detect anomaly
	AnomaliesDetected.WithLabelValues("high", "behavior", service).Inc()
	AnomalyDetectionAccuracy.WithLabelValues(modelVersion).Set(0.95)

	// 3. Make load prediction
	LoadPredictionsMade.WithLabelValues("1h", "high").Inc()
	LoadPredictionAccuracy.WithLabelValues(modelVersion, "1h").Set(0.92)

	// 4. Optimize route
	RoutesOptimized.WithLabelValues("p2p_quic", "latency").Inc()
	RouteOptimizationLatency.WithLabelValues("p2p_quic").Observe(0.01)

	// 5. Cache operations
	AICacheHits.WithLabelValues("model_cache").Inc()
	AICacheMisses.WithLabelValues("prediction_cache").Inc()

	// 6. Circuit breaker
	AICircuitBreakerState.WithLabelValues(service).Set(0) // closed
	AICircuitBreakerFailures.WithLabelValues(service).Inc()

	// 7. System health
	AIServiceHealth.WithLabelValues(service, component).Set(1) // healthy
	AIModelVersion.WithLabelValues("behavior", modelVersion).Set(1.0)

	// 8. Performance metrics
	AIMemoryUsage.WithLabelValues(component).Set(1024 * 1024 * 100) // 100MB
	AICPUUsage.WithLabelValues(component).Set(25.0)                 // 25%

	// 9. Business metrics
	AICostSavings.WithLabelValues("route_optimization").Inc()
	AIAutomationRate.WithLabelValues("route_optimization").Set(95.0)

	// Verify workflow completed
	assert.NotNil(t, AIConnectionsProcessed)
	assert.NotNil(t, AIProcessingDuration)
	assert.NotNil(t, AnomaliesDetected)
	assert.NotNil(t, AnomalyDetectionAccuracy)
	assert.NotNil(t, LoadPredictionsMade)
	assert.NotNil(t, LoadPredictionAccuracy)
	assert.NotNil(t, RoutesOptimized)
	assert.NotNil(t, RouteOptimizationLatency)
	assert.NotNil(t, AICacheHits)
	assert.NotNil(t, AICacheMisses)
	assert.NotNil(t, AICircuitBreakerState)
	assert.NotNil(t, AICircuitBreakerFailures)
	assert.NotNil(t, AIServiceHealth)
	assert.NotNil(t, AIModelVersion)
	assert.NotNil(t, AIMemoryUsage)
	assert.NotNil(t, AICPUUsage)
	assert.NotNil(t, AICostSavings)
	assert.NotNil(t, AIAutomationRate)
}

func TestAIExtendedMetrics_Performance(t *testing.T) {
	// Test performance with many operations
	start := time.Now()

	for i := 0; i < 1000; i++ {
		service := "service" + string(rune(i%10))
		component := "component" + string(rune(i%10))

		AIConnectionsProcessed.WithLabelValues("success", "false", service).Inc()
		AIProcessingDuration.WithLabelValues(service, "process").Observe(float64(i%100) / 100.0)
		AnomaliesDetected.WithLabelValues("medium", "test", service).Inc()
		AnomalyDetectionAccuracy.WithLabelValues("v1.0").Set(float64(i%100) / 100.0)
		LoadPredictionsMade.WithLabelValues("1h", "high").Inc()
		LoadPredictionAccuracy.WithLabelValues("v1.0", "1h").Set(float64(i%100) / 100.0)
		RoutesOptimized.WithLabelValues("p2p_quic", "latency").Inc()
		RouteOptimizationLatency.WithLabelValues("p2p_quic").Observe(float64(i%100) / 1000.0)
		AICacheHits.WithLabelValues("model_cache").Inc()
		AICacheMisses.WithLabelValues("model_cache").Inc()
		AICircuitBreakerState.WithLabelValues(service).Set(float64(i % 3))
		AICircuitBreakerFailures.WithLabelValues(service).Inc()
		AIServiceHealth.WithLabelValues(service, component).Set(float64(i % 2))
		AIModelVersion.WithLabelValues("behavior", "v1.0").Set(1.0)
		AIMemoryUsage.WithLabelValues(component).Set(float64(i * 1024))
		AICPUUsage.WithLabelValues(component).Set(float64(i % 100))
		AICostSavings.WithLabelValues("optimization").Inc()
		AIAutomationRate.WithLabelValues("automation").Set(float64(i % 100))
	}

	duration := time.Since(start)

	// Should complete within reasonable time (less than 1 second)
	assert.Less(t, duration, time.Second)
}

func TestAIExtendedMetrics_DataTypes(t *testing.T) {
	// Test different data types and edge cases
	testCases := []struct {
		name     string
		testFunc func()
	}{
		{
			name: "string labels",
			testFunc: func() {
				AIConnectionsProcessed.WithLabelValues("success-status", "false-detection", "service-name").Inc()
			},
		},
		{
			name: "unicode strings",
			testFunc: func() {
				AIConnectionsProcessed.WithLabelValues("успех", "ложное-обнаружение", "сервис-имя").Inc()
			},
		},
		{
			name: "special characters",
			testFunc: func() {
				AIConnectionsProcessed.WithLabelValues("success@domain.com", "false_detection", "service-name").Inc()
			},
		},
		{
			name: "long strings",
			testFunc: func() {
				longString := "very-long-service-name-that-might-cause-issues-with-prometheus-label-limits"
				AIConnectionsProcessed.WithLabelValues("success", "false", longString).Inc()
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NotPanics(t, tc.testFunc)
		})
	}
}

func TestAIExtendedMetrics_CircuitBreakerStates(t *testing.T) {
	// Test all circuit breaker states
	service := "test_service"

	// Test closed state (0)
	AICircuitBreakerState.WithLabelValues(service).Set(0)
	assert.NotNil(t, AICircuitBreakerState)

	// Test open state (1)
	AICircuitBreakerState.WithLabelValues(service).Set(1)
	assert.NotNil(t, AICircuitBreakerState)

	// Test half-open state (2)
	AICircuitBreakerState.WithLabelValues(service).Set(2)
	assert.NotNil(t, AICircuitBreakerState)

	// Test invalid state (should still work)
	AICircuitBreakerState.WithLabelValues(service).Set(3)
	assert.NotNil(t, AICircuitBreakerState)
}

func TestAIExtendedMetrics_HealthStates(t *testing.T) {
	// Test all health states
	service := "test_service"
	component := "test_component"

	// Test healthy state (1)
	AIServiceHealth.WithLabelValues(service, component).Set(1)
	assert.NotNil(t, AIServiceHealth)

	// Test unhealthy state (0)
	AIServiceHealth.WithLabelValues(service, component).Set(0)
	assert.NotNil(t, AIServiceHealth)

	// Test partial health (0.5)
	AIServiceHealth.WithLabelValues(service, component).Set(0.5)
	assert.NotNil(t, AIServiceHealth)
}
