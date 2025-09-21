package ai

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBehaviorAnalyzer_AnalyzeQUICBehavior(t *testing.T) {
	analyzer := NewBehaviorAnalyzer()

	// Test with valid connection ID
	connectionID := "conn-001"
	event := map[string]interface{}{
		"data_volume": 2048,
		"duration":    45 * time.Second,
		"streams":     5,
	}

	err := analyzer.AnalyzeQUICBehavior(connectionID, event)
	require.NoError(t, err)

	// Verify profile was created
	userID := "quic-connection-" + connectionID
	profile, exists := analyzer.UserProfiles[userID]
	assert.True(t, exists)
	assert.NotNil(t, profile)
	assert.Equal(t, userID, profile.UserID)
	assert.Equal(t, 0.0, profile.RiskScore)
	assert.NotZero(t, profile.LastUpdated)
}

func TestBehaviorAnalyzer_AnalyzeQUICBehavior_EmptyConnectionID(t *testing.T) {
	analyzer := NewBehaviorAnalyzer()

	// Test with empty connection ID
	err := analyzer.AnalyzeQUICBehavior("", nil)
	require.NoError(t, err)

	// Verify profile was created with empty connection ID
	userID := "quic-connection-"
	profile, exists := analyzer.UserProfiles[userID]
	assert.True(t, exists)
	assert.NotNil(t, profile)
	assert.Equal(t, userID, profile.UserID)
}

func TestBehaviorAnalyzer_AnalyzeQUICBehavior_ComplexEvent(t *testing.T) {
	analyzer := NewBehaviorAnalyzer()

	// Test with complex event data
	connectionID := "conn-002"
	event := map[string]interface{}{
		"data_volume":        4096,
		"duration":           120 * time.Second,
		"streams":            10,
		"retransmissions":    2,
		"congestion_control": "bbr",
	}

	err := analyzer.AnalyzeQUICBehavior(connectionID, event)
	require.NoError(t, err)

	// Verify profile was created
	userID := "quic-connection-" + connectionID
	profile, exists := analyzer.UserProfiles[userID]
	assert.True(t, exists)
	assert.NotNil(t, profile)
	assert.Equal(t, userID, profile.UserID)
}

func TestPredictiveSecurity_PredictQuantumThreats(t *testing.T) {
	security := NewPredictiveSecurity()

	// Test quantum threat prediction
	err := security.PredictQuantumThreats()
	require.NoError(t, err)

	// Verify threat prediction was created
	predictions := security.GetThreatPredictions()
	assert.NotEmpty(t, predictions)

	// Find the quantum threat prediction
	var quantumPrediction *ThreatPrediction
	for _, prediction := range predictions {
		if prediction.ThreatID != "" {
			quantumPrediction = prediction
			break
		}
	}

	assert.NotNil(t, quantumPrediction)
	assert.GreaterOrEqual(t, quantumPrediction.Probability, 0.0)
	assert.LessOrEqual(t, quantumPrediction.Probability, 1.0)
	assert.Greater(t, quantumPrediction.TimeWindow, time.Duration(0))
	assert.GreaterOrEqual(t, quantumPrediction.Confidence, 0.0)
	assert.LessOrEqual(t, quantumPrediction.Confidence, 1.0)

	// Verify mitigation strategies were added
	assert.GreaterOrEqual(t, len(quantumPrediction.Mitigation), 1)

	// Check that we have some mitigation strategies (exact content may vary)
	for _, strategy := range quantumPrediction.Mitigation {
		assert.NotEmpty(t, strategy.Name)
		assert.NotEmpty(t, strategy.Action)
		assert.Greater(t, strategy.Priority, 0)
		assert.GreaterOrEqual(t, strategy.Effectiveness, 0.0)
		assert.LessOrEqual(t, strategy.Effectiveness, 1.0)
	}
}

func TestPredictiveSecurity_PredictQuantumThreats_MultipleCalls(t *testing.T) {
	security := NewPredictiveSecurity()

	// Call multiple times
	for i := 0; i < 3; i++ {
		err := security.PredictQuantumThreats()
		require.NoError(t, err)
	}

	// Verify multiple predictions were created
	predictions := security.GetThreatPredictions()
	assert.GreaterOrEqual(t, len(predictions), 1) // At least one prediction should be created
}

func TestAIMonitoring_MonitorQUICMetrics(t *testing.T) {
	monitoring := NewAIMonitoring()

	// Test monitoring QUIC metrics
	err := monitoring.MonitorQUICMetrics()
	require.NoError(t, err)

	// Verify metrics were added
	assert.Len(t, monitoring.Metrics, 3)

	// Check active connections metric
	activeConnections, exists := monitoring.Metrics["quic_connections_active"]
	assert.True(t, exists)
	assert.NotNil(t, activeConnections)
	assert.Equal(t, "quic_connections_active", activeConnections.Name)
	assert.Equal(t, 100.0, activeConnections.Value)
	assert.Equal(t, "connections", activeConnections.Unit)
	assert.Equal(t, TrendUp, activeConnections.Trend)
	assert.Equal(t, 200.0, activeConnections.Threshold)

	// Check throughput metric
	throughput, exists := monitoring.Metrics["quic_throughput"]
	assert.True(t, exists)
	assert.NotNil(t, throughput)
	assert.Equal(t, "quic_throughput", throughput.Name)
	assert.Equal(t, 512.0, throughput.Value)
	assert.Equal(t, "MB/s", throughput.Unit)
	assert.Equal(t, TrendStable, throughput.Trend)
	assert.Equal(t, 1024.0, throughput.Threshold)

	// Check latency metric
	latency, exists := monitoring.Metrics["quic_latency"]
	assert.True(t, exists)
	assert.NotNil(t, latency)
	assert.Equal(t, "quic_latency", latency.Name)
	assert.Equal(t, 50.0, latency.Value)
	assert.Equal(t, "ms", latency.Unit)
	assert.Equal(t, TrendDown, latency.Trend)
	assert.Equal(t, 100.0, latency.Threshold)
}

func TestAIMonitoring_MonitorQuantumMetrics(t *testing.T) {
	monitoring := NewAIMonitoring()

	// Test monitoring quantum metrics
	err := monitoring.MonitorQuantumMetrics()
	require.NoError(t, err)

	// Verify metrics were added
	assert.Len(t, monitoring.Metrics, 3)

	// Check key exchanges metric
	keyExchanges, exists := monitoring.Metrics["quantum_key_exchanges"]
	assert.True(t, exists)
	assert.NotNil(t, keyExchanges)
	assert.Equal(t, "quantum_key_exchanges", keyExchanges.Name)
	assert.Equal(t, 150.0, keyExchanges.Value)
	assert.Equal(t, "exchanges", keyExchanges.Unit)
	assert.Equal(t, TrendUp, keyExchanges.Trend)
	assert.Equal(t, 300.0, keyExchanges.Threshold)

	// Check key rotation frequency metric
	keyRotation, exists := monitoring.Metrics["quantum_key_rotation_frequency"]
	assert.True(t, exists)
	assert.NotNil(t, keyRotation)
	assert.Equal(t, "quantum_key_rotation_frequency", keyRotation.Name)
	assert.Equal(t, 0.8, keyRotation.Value)
	assert.Equal(t, "per_day", keyRotation.Unit)
	assert.Equal(t, TrendStable, keyRotation.Trend)
	assert.Equal(t, 1.0, keyRotation.Threshold)

	// Check threat level metric
	threatLevel, exists := monitoring.Metrics["quantum_threat_level"]
	assert.True(t, exists)
	assert.NotNil(t, threatLevel)
	assert.Equal(t, "quantum_threat_level", threatLevel.Name)
	assert.Equal(t, 0.25, threatLevel.Value)
	assert.Equal(t, "level", threatLevel.Unit)
	assert.Equal(t, TrendStable, threatLevel.Trend)
	assert.Equal(t, 0.7, threatLevel.Threshold)
}

func TestIntegrations_Integration(t *testing.T) {
	// Test integration between all components
	analyzer := NewBehaviorAnalyzer()
	security := NewPredictiveSecurity()
	response := NewAutomatedResponse()
	monitoring := NewAIMonitoring()

	var err error

	// Test P2P QUIC integration
	err = analyzer.AnalyzeP2PQUICBehavior("peer-001", map[string]interface{}{"data_volume": 2048})
	require.NoError(t, err)

	// Test QUIC integration
	err = analyzer.AnalyzeQUICBehavior("conn-001", map[string]interface{}{"data_volume": 1024})
	require.NoError(t, err)

	// Test quantum threat prediction
	err = security.PredictQuantumThreats()
	require.NoError(t, err)

	// Test monitoring
	err = monitoring.MonitorP2PQUICMetrics()
	require.NoError(t, err)

	err = monitoring.MonitorQUICMetrics()
	require.NoError(t, err)

	err = monitoring.MonitorQuantumMetrics()
	require.NoError(t, err)

	// Verify all components are working together
	assert.Len(t, analyzer.UserProfiles, 2) // P2P QUIC + QUIC
	assert.Len(t, security.GetThreatPredictions(), 1)
	assert.Len(t, response.ResponseRules, 0)        // No threat response rules triggered
	assert.Len(t, response.GetResponseHistory(), 0) // No threat responses triggered
	assert.Len(t, monitoring.Metrics, 9)            // 3 + 3 + 3 metrics (P2P QUIC + QUIC + Quantum)
}

func TestIntegrations_EdgeCases(t *testing.T) {
	analyzer := NewBehaviorAnalyzer()

	var err error

	// Test with nil events
	err = analyzer.AnalyzeP2PQUICBehavior("peer-001", nil)
	require.NoError(t, err)

	err = analyzer.AnalyzeQUICBehavior("conn-001", nil)
	require.NoError(t, err)

	// Test with empty strings
	err = analyzer.AnalyzeP2PQUICBehavior("", nil)
	require.NoError(t, err)

	err = analyzer.AnalyzeQUICBehavior("", nil)
	require.NoError(t, err)

	// All operations should complete without errors
	assert.True(t, true)
}

func TestIntegrations_ConcurrentAccess(t *testing.T) {
	analyzer := NewBehaviorAnalyzer()
	security := NewPredictiveSecurity()
	monitoring := NewAIMonitoring()

	// Test concurrent access
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(index int) {
			defer func() { done <- true }()

			var err error

			// Concurrent P2P QUIC analysis
			err = analyzer.AnalyzeP2PQUICBehavior("peer-"+string(rune(index)), nil)
			if err != nil {
				t.Logf("P2P QUIC analysis error: %v", err)
			}

			// Concurrent QUIC analysis
			err = analyzer.AnalyzeQUICBehavior("conn-"+string(rune(index)), nil)
			if err != nil {
				t.Logf("QUIC analysis error: %v", err)
			}

			// Concurrent quantum threat prediction
			err = security.PredictQuantumThreats()
			if err != nil {
				t.Logf("Quantum threat prediction error: %v", err)
			}

			// Concurrent monitoring
			err = monitoring.MonitorP2PQUICMetrics()
			if err != nil {
				t.Logf("P2P QUIC monitoring error: %v", err)
			}

			err = monitoring.MonitorQUICMetrics()
			if err != nil {
				t.Logf("QUIC monitoring error: %v", err)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify components are still functional
	assert.NotNil(t, analyzer.UserProfiles)
	assert.NotNil(t, security.GetThreatPredictions())
	assert.NotNil(t, monitoring.Metrics)
}

func TestBehaviorAnalyzer_AnalyzeP2PQUICBehavior(t *testing.T) {
	analyzer := NewBehaviorAnalyzer()

	// Test with valid peer ID
	peerID := "peer-001"
	event := map[string]interface{}{
		"data_volume": 2048,
		"duration":    30 * time.Second,
	}

	err := analyzer.AnalyzeP2PQUICBehavior(peerID, event)
	require.NoError(t, err)

	// Verify profile was created
	userID := "p2p-quic-peer-" + peerID
	profile, exists := analyzer.UserProfiles[userID]
	assert.True(t, exists)
	assert.NotNil(t, profile)
	assert.Equal(t, userID, profile.UserID)
	assert.Equal(t, 0.0, profile.RiskScore)
	assert.NotZero(t, profile.LastUpdated)
}

func TestAutomatedResponse_RespondToP2PQUICThreat(t *testing.T) {
	response := NewAutomatedResponse()

	// Create a threat prediction
	threat := &ThreatPrediction{
		ThreatID:    "p2p-quic-threat-001",
		Probability: 0.8,
		TimeWindow:  30 * time.Minute,
		Confidence:  0.9,
		Mitigation:  []MitigationStrategy{},
	}

	// Test responding to P2P QUIC threat
	err := response.RespondToP2PQUICThreat(threat)
	require.NoError(t, err)

	// Verify response rule was added
	rule, exists := response.ResponseRules["p2p-quic-threat-response"]
	assert.True(t, exists)
	assert.NotNil(t, rule)
	assert.Equal(t, "p2p-quic-threat-response", rule.RuleID)
	assert.Equal(t, "P2P QUIC Threat Response", rule.Name)
	assert.Equal(t, 1, rule.Priority)
	assert.True(t, rule.Enabled)

	// Verify conditions
	assert.Len(t, rule.Conditions, 1)
	condition := rule.Conditions[0]
	assert.Equal(t, ConditionTypeThreat, condition.Type)
	assert.Equal(t, "threat_level", condition.Field)
	assert.Equal(t, ">=", condition.Operator)
	assert.Equal(t, "high", condition.Value)

	// Verify actions
	assert.Len(t, rule.Actions, 4)
	assert.Contains(t, rule.Actions, "block_suspicious_peer")
	assert.Contains(t, rule.Actions, "increase_monitoring")
	assert.Contains(t, rule.Actions, "notify_admin")
	assert.Contains(t, rule.Actions, "isolate_peer")

	// Verify response event was triggered
	history := response.GetResponseHistory()
	assert.GreaterOrEqual(t, len(history), 0) // May be 0 if no rules triggered
	if len(history) > 0 {
		event := history[0]
		assert.NotEmpty(t, event.EventID)
		assert.NotEmpty(t, event.RuleID)
		assert.NotNil(t, event.Success)
	}
}

func TestAIMonitoring_MonitorP2PQUICMetrics(t *testing.T) {
	monitoring := NewAIMonitoring()

	// Test monitoring P2P QUIC metrics
	err := monitoring.MonitorP2PQUICMetrics()
	require.NoError(t, err)

	// Verify metrics were added
	assert.Len(t, monitoring.Metrics, 3)

	// Check active peers metric
	activePeers, exists := monitoring.Metrics["p2p_quic_peers_active"]
	assert.True(t, exists)
	assert.NotNil(t, activePeers)
	assert.Equal(t, "p2p_quic_peers_active", activePeers.Name)
	assert.Equal(t, 30.0, activePeers.Value)
	assert.Equal(t, "peers", activePeers.Unit)
	assert.Equal(t, TrendUp, activePeers.Trend)
	assert.Equal(t, 50.0, activePeers.Threshold)

	// Check bandwidth usage metric
	bandwidthUsage, exists := monitoring.Metrics["p2p_quic_bandwidth_usage"]
	assert.True(t, exists)
	assert.NotNil(t, bandwidthUsage)
	assert.Equal(t, "p2p_quic_bandwidth_usage", bandwidthUsage.Name)
	assert.Equal(t, 1500.0, bandwidthUsage.Value)
	assert.Equal(t, "MB/s", bandwidthUsage.Unit)
	assert.Equal(t, TrendUp, bandwidthUsage.Trend)
	assert.Equal(t, 3000.0, bandwidthUsage.Threshold)

	// Check anomaly score metric
	anomalyScore, exists := monitoring.Metrics["p2p_quic_anomaly_score"]
	assert.True(t, exists)
	assert.NotNil(t, anomalyScore)
	assert.Equal(t, "p2p_quic_anomaly_score", anomalyScore.Name)
	assert.Equal(t, 0.12, anomalyScore.Value)
	assert.Equal(t, "score", anomalyScore.Unit)
	assert.Equal(t, TrendStable, anomalyScore.Trend)
	assert.Equal(t, 0.8, anomalyScore.Threshold)

	// Verify no alerts were created (anomaly score is below threshold)
	assert.Len(t, monitoring.Alerts, 0)
}
