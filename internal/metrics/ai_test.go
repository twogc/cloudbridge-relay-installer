package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAIMetrics(t *testing.T) {
	metrics := NewAIMetrics()
	assert.NotNil(t, metrics)
}

func TestAIMetrics_BehaviorAnalysis(t *testing.T) {
	metrics := NewAIMetrics()

	// Test increment behavior analysis
	metrics.IncrementBehaviorAnalysis("user1", "login", "success")
	metrics.IncrementBehaviorAnalysis("user1", "login", "success")
	metrics.IncrementBehaviorAnalysis("user2", "logout", "success")

	// Test observe anomaly score
	metrics.ObserveAnomalyScore("user1", "login", 0.8)
	metrics.ObserveAnomalyScore("user1", "login", 0.6)
	metrics.ObserveAnomalyScore("user2", "logout", 0.9)

	// Test increment anomaly detections
	metrics.IncrementAnomalyDetections("user1", "unusual_login_time", "high")
	metrics.IncrementAnomalyDetections("user1", "unusual_login_time", "high")
	metrics.IncrementAnomalyDetections("user2", "suspicious_activity", "medium")

	// Test update user profiles
	metrics.UpdateUserProfiles("active", 100)
	metrics.UpdateUserProfiles("inactive", 50)

	// Test update model accuracy
	metrics.UpdateModelAccuracy("v1.0", 0.95)
	metrics.UpdateModelAccuracy("v2.0", 0.98)

	// Test observe model training duration
	metrics.ObserveModelTrainingDuration("v1.0", 3600.0) // 1 hour
	metrics.ObserveModelTrainingDuration("v2.0", 7200.0) // 2 hours

	// Test observe model inference duration
	metrics.ObserveModelInferenceDuration("v1.0", 0.1)
	metrics.ObserveModelInferenceDuration("v2.0", 0.05)
}

func TestAIMetrics_PredictiveSecurity(t *testing.T) {
	metrics := NewAIMetrics()

	// Test increment predictive threats
	metrics.IncrementPredictiveThreats("ddos", "high", "critical")
	metrics.IncrementPredictiveThreats("ddos", "high", "critical")
	metrics.IncrementPredictiveThreats("brute_force", "medium", "high")

	// Test update predictive accuracy
	metrics.UpdatePredictiveAccuracy("ddos", "v1.0", 0.92)
	metrics.UpdatePredictiveAccuracy("brute_force", "v1.0", 0.88)

	// Test increment false positives
	metrics.IncrementFalsePositives("ddos", "v1.0")
	metrics.IncrementFalsePositives("ddos", "v1.0")
	metrics.IncrementFalsePositives("brute_force", "v1.0")

	// Test increment false negatives
	metrics.IncrementFalseNegatives("ddos", "v1.0")
	metrics.IncrementFalseNegatives("brute_force", "v1.0")

	// Test observe predictive model training duration
	metrics.ObservePredictiveModelTrainingDuration("ddos", "v1.0", 1800.0)        // 30 minutes
	metrics.ObservePredictiveModelTrainingDuration("brute_force", "v1.0", 2400.0) // 40 minutes

	// Test observe predictive model inference duration
	metrics.ObservePredictiveModelInferenceDuration("ddos", "v1.0", 0.2)
	metrics.ObservePredictiveModelInferenceDuration("brute_force", "v1.0", 0.15)
}

func TestAIMetrics_AutomatedResponse(t *testing.T) {
	metrics := NewAIMetrics()

	// Test increment automated responses
	metrics.IncrementAutomatedResponses("block", "ddos", "success")
	metrics.IncrementAutomatedResponses("block", "ddos", "success")
	metrics.IncrementAutomatedResponses("alert", "brute_force", "success")

	// Test observe automated response duration
	metrics.ObserveAutomatedResponseDuration("block", 0.5)
	metrics.ObserveAutomatedResponseDuration("alert", 0.1)

	// Test update automated response success
	metrics.UpdateAutomatedResponseSuccess("block", 0.95)
	metrics.UpdateAutomatedResponseSuccess("alert", 0.98)

	// Test update automated response rules
	metrics.UpdateAutomatedResponseRules("active", 10)
	metrics.UpdateAutomatedResponseRules("inactive", 5)

	// Test increment automated response errors
	metrics.IncrementAutomatedResponseErrors("block", "timeout")
	metrics.IncrementAutomatedResponseErrors("alert", "network_error")
}

func TestAIMetrics_Monitoring(t *testing.T) {
	metrics := NewAIMetrics()

	// Test increment monitoring alerts
	metrics.IncrementMonitoringAlerts("anomaly", "high", "active")
	metrics.IncrementMonitoringAlerts("anomaly", "high", "active")
	metrics.IncrementMonitoringAlerts("threat", "critical", "resolved")

	// Test update monitoring metrics
	metrics.UpdateMonitoringMetrics("cpu_usage", "behavior_analyzer", 75.5)
	metrics.UpdateMonitoringMetrics("memory_usage", "predictive_security", 1024.0)

	// Test update monitoring dashboard status
	metrics.UpdateMonitoringDashboardStatus("http://localhost:8080/dashboard", true)
	metrics.UpdateMonitoringDashboardStatus("http://localhost:8080/dashboard", false)

	// Test increment monitoring data points
	metrics.IncrementMonitoringDataPoints("behavior", "user_events")
	metrics.IncrementMonitoringDataPoints("behavior", "user_events")
	metrics.IncrementMonitoringDataPoints("threat", "network_traffic")

	// Test observe monitoring data processing duration
	metrics.ObserveMonitoringDataProcessingDuration("behavior", 0.1)
	metrics.ObserveMonitoringDataProcessingDuration("threat", 0.2)
}

func TestAIMetrics_Integration(t *testing.T) {
	metrics := NewAIMetrics()

	// Test WireGuard integration
	metrics.IncrementIntegrationWireGuardAnalysis("behavior", "success")
	metrics.IncrementIntegrationWireGuardAnalysis("behavior", "success")
	metrics.IncrementIntegrationWireGuardThreats("ddos", "block")

	// Test QUIC integration
	metrics.IncrementIntegrationQUICAnalysis("performance", "success")
	metrics.IncrementIntegrationQUICAnalysis("performance", "success")
	metrics.IncrementIntegrationQUICThreats("eavesdropping", "encrypt")

	// Test Quantum integration
	metrics.IncrementIntegrationQuantumAnalysis("threat_assessment", "success")
	metrics.IncrementIntegrationQuantumAnalysis("threat_assessment", "success")
	metrics.IncrementIntegrationQuantumThreats("quantum_attack", "post_quantum_crypto")
}

func TestAIMetrics_System(t *testing.T) {
	metrics := NewAIMetrics()

	// Test update system status
	metrics.UpdateSystemStatus("behavior_analyzer", true)
	metrics.UpdateSystemStatus("predictive_security", true)
	metrics.UpdateSystemStatus("automated_response", false)

	// Test update system memory usage
	metrics.UpdateSystemMemoryUsage("behavior_analyzer", 1024*1024*100)   // 100MB
	metrics.UpdateSystemMemoryUsage("predictive_security", 1024*1024*200) // 200MB

	// Test update system CPU usage
	metrics.UpdateSystemCPUUsage("behavior_analyzer", 25.5)
	metrics.UpdateSystemCPUUsage("predictive_security", 45.0)

	// Test observe system model load time
	metrics.ObserveSystemModelLoadTime("behavior", "v1.0", 5.0)
	metrics.ObserveSystemModelLoadTime("predictive", "v1.0", 10.0)

	// Test increment system errors
	metrics.IncrementSystemErrors("behavior_analyzer", "model_load_failed")
	metrics.IncrementSystemErrors("predictive_security", "inference_timeout")
}

func TestAIMetrics_EdgeCases(t *testing.T) {
	metrics := NewAIMetrics()

	// Test with empty strings
	metrics.IncrementBehaviorAnalysis("", "", "")
	metrics.ObserveAnomalyScore("", "", 0.0)
	metrics.IncrementAnomalyDetections("", "", "")
	metrics.UpdateUserProfiles("", 0)
	metrics.UpdateModelAccuracy("", 0.0)

	// Test with negative values
	metrics.ObserveAnomalyScore("user1", "login", -0.1)
	metrics.UpdateUserProfiles("active", -10)
	metrics.UpdateModelAccuracy("v1.0", -0.1)

	// Test with very large values
	metrics.ObserveAnomalyScore("user1", "login", 999.0)
	metrics.UpdateUserProfiles("active", 999999)
	metrics.UpdateModelAccuracy("v1.0", 999.0)

	// Test with zero duration
	metrics.ObserveModelTrainingDuration("v1.0", 0.0)
	metrics.ObserveModelInferenceDuration("v1.0", 0.0)
	metrics.ObserveAutomatedResponseDuration("block", 0.0)
	metrics.ObserveMonitoringDataProcessingDuration("behavior", 0.0)
	metrics.ObserveSystemModelLoadTime("behavior", "v1.0", 0.0)
}

func TestAIMetrics_ConcurrentAccess(t *testing.T) {
	metrics := NewAIMetrics()
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			userID := "user" + string(rune(id))
			metrics.IncrementBehaviorAnalysis(userID, "login", "success")
			metrics.ObserveAnomalyScore(userID, "login", 0.5)
			metrics.IncrementAnomalyDetections(userID, "anomaly", "medium")
			metrics.UpdateUserProfiles("active", id)
			metrics.UpdateModelAccuracy("v1.0", 0.9)
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify that all operations completed without errors
	assert.NotNil(t, metrics)
}

func TestAIMetrics_IntegrationWorkflow(t *testing.T) {
	metrics := NewAIMetrics()

	// Simulate a complete AI workflow
	userID := "test-user"
	threatType := "ddos"
	responseType := "block"

	// 1. Behavior analysis
	metrics.IncrementBehaviorAnalysis(userID, "login", "success")
	metrics.ObserveAnomalyScore(userID, "login", 0.8)

	// 2. Anomaly detection
	metrics.IncrementAnomalyDetections(userID, "unusual_pattern", "high")

	// 3. Predictive threat detection
	metrics.IncrementPredictiveThreats(threatType, "high", "critical")

	// 4. Automated response
	metrics.IncrementAutomatedResponses(responseType, threatType, "success")
	metrics.ObserveAutomatedResponseDuration(responseType, 0.5)

	// 5. Monitoring
	metrics.IncrementMonitoringAlerts("threat", "critical", "active")
	metrics.UpdateMonitoringMetrics("response_time", "automated_response", 0.5)

	// 6. System metrics
	metrics.UpdateSystemStatus("automated_response", true)
	metrics.UpdateSystemCPUUsage("automated_response", 15.0)

	// Verify workflow completed
	assert.NotNil(t, metrics)
}

func TestAIMetrics_Performance(t *testing.T) {
	metrics := NewAIMetrics()

	// Test performance with many operations
	start := time.Now()

	for i := 0; i < 1000; i++ {
		userID := "user" + string(rune(i%100))
		metrics.IncrementBehaviorAnalysis(userID, "login", "success")
		metrics.ObserveAnomalyScore(userID, "login", float64(i%100)/100.0)
		metrics.IncrementAnomalyDetections(userID, "anomaly", "medium")
	}

	duration := time.Since(start)

	// Should complete within reasonable time (less than 1 second)
	assert.Less(t, duration, time.Second)
}

func TestAIMetrics_DataTypes(t *testing.T) {
	metrics := NewAIMetrics()

	// Test different data types and edge cases
	testCases := []struct {
		name     string
		testFunc func()
	}{
		{
			name: "string labels",
			testFunc: func() {
				metrics.IncrementBehaviorAnalysis("user-123", "login-action", "success-status")
			},
		},
		{
			name: "unicode strings",
			testFunc: func() {
				metrics.IncrementBehaviorAnalysis("пользователь-123", "действие-входа", "статус-успех")
			},
		},
		{
			name: "special characters",
			testFunc: func() {
				metrics.IncrementBehaviorAnalysis("user@domain.com", "action-with-dash", "status_with_underscore")
			},
		},
		{
			name: "long strings",
			testFunc: func() {
				longString := "very-long-user-id-that-might-cause-issues-with-prometheus-label-limits"
				metrics.IncrementBehaviorAnalysis(longString, "long-action-name", "long-status-name")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NotPanics(t, tc.testFunc)
		})
	}
}
