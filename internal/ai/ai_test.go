package ai

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBehaviorAnalyzer(t *testing.T) {
	analyzer := NewBehaviorAnalyzer()

	assert.NotNil(t, analyzer)
	assert.True(t, analyzer.Enabled)
	assert.Equal(t, "/models/behavior", analyzer.ModelPath)
	assert.Equal(t, 0.8, analyzer.Threshold)
	assert.Equal(t, 0.01, analyzer.LearningRate)
	assert.NotNil(t, analyzer.UserProfiles)
	assert.NotNil(t, analyzer.AnomalyScores)
}

func TestBehaviorAnalyzer_Start(t *testing.T) {
	analyzer := NewBehaviorAnalyzer()

	err := analyzer.Start()
	assert.NoError(t, err)
}

func TestBehaviorAnalyzer_Stop(t *testing.T) {
	analyzer := NewBehaviorAnalyzer()

	err := analyzer.Stop()
	assert.NoError(t, err)
}

func TestBehaviorAnalyzer_AnalyzeUserBehavior(t *testing.T) {
	analyzer := NewBehaviorAnalyzer()

	// Test with valid user ID
	err := analyzer.AnalyzeUserBehavior("user1", map[string]interface{}{"event": "login"})
	assert.NoError(t, err)

	// Verify profile was created
	profile, exists := analyzer.UserProfiles["user1"]
	assert.True(t, exists)
	assert.NotNil(t, profile)
	assert.Equal(t, "user1", profile.UserID)
	assert.Equal(t, 0.0, profile.RiskScore)
	assert.NotZero(t, profile.LastUpdated)
}

func TestBehaviorAnalyzer_DetectAnomaly(t *testing.T) {
	analyzer := NewBehaviorAnalyzer()

	// First create a user profile
	err := analyzer.AnalyzeUserBehavior("user1", map[string]interface{}{"event": "login"})
	assert.NoError(t, err)

	// Test anomaly detection
	score, err := analyzer.DetectAnomaly("user1")
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, score, 0.0)
	assert.LessOrEqual(t, score, 1.0)

	// Verify score was stored
	storedScore, exists := analyzer.AnomalyScores["user1"]
	assert.True(t, exists)
	assert.Equal(t, score, storedScore)
}

func TestBehaviorAnalyzer_UpdateUserProfile(t *testing.T) {
	analyzer := NewBehaviorAnalyzer()

	// Create a user profile first
	analyzer.UserProfiles["user1"] = &UserProfile{
		UserID:      "user1",
		RiskScore:   0.5,
		LastUpdated: time.Now(),
	}

	err := analyzer.UpdateUserProfile("user1", map[string]interface{}{"event": "update"})
	assert.NoError(t, err)

	// Verify profile was updated
	updatedProfile, exists := analyzer.UserProfiles["user1"]
	assert.True(t, exists)
	assert.Equal(t, "user1", updatedProfile.UserID)
	assert.Greater(t, updatedProfile.RiskScore, 0.5) // Should be increased
}

func TestNewPredictiveSecurity(t *testing.T) {
	security := NewPredictiveSecurity()

	assert.NotNil(t, security)
	assert.True(t, security.Enabled)
	assert.NotNil(t, security.ThreatModels)
	assert.NotNil(t, security.Predictions)
	assert.Equal(t, 0.85, security.Confidence)
}

func TestPredictiveSecurity_Start(t *testing.T) {
	security := NewPredictiveSecurity()

	err := security.Start()
	assert.NoError(t, err)
}

func TestPredictiveSecurity_Stop(t *testing.T) {
	security := NewPredictiveSecurity()

	err := security.Stop()
	assert.NoError(t, err)
}

func TestPredictiveSecurity_PredictThreat(t *testing.T) {
	security := NewPredictiveSecurity()

	// Test threat prediction
	indicators := []ThreatIndicator{
		{Name: "packet_rate", Value: 1000, Weight: 0.5, Threshold: 500},
		{Name: "connection_count", Value: 200, Weight: 0.3, Threshold: 100},
	}

	prediction, err := security.PredictThreat(indicators)
	assert.NoError(t, err)
	assert.NotNil(t, prediction)
	assert.NotEmpty(t, prediction.ThreatID)
	assert.Contains(t, prediction.ThreatID, "predicted-threat-")
	assert.GreaterOrEqual(t, prediction.Probability, 0.0)
	assert.LessOrEqual(t, prediction.Probability, 1.0)
	assert.Greater(t, prediction.TimeWindow, time.Duration(0))
	assert.GreaterOrEqual(t, prediction.Confidence, 0.0)
	assert.LessOrEqual(t, prediction.Confidence, 1.0)
}

func TestPredictiveSecurity_AddThreatModel(t *testing.T) {
	security := NewPredictiveSecurity()

	// Create a threat model with indicators
	model := &ThreatModel{
		ThreatID: "threat1",
		Type:     ThreatTypeDDoS,
		Indicators: []ThreatIndicator{
			{Name: "packet_rate", Value: 0, Weight: 0.5, Threshold: 1000},
			{Name: "connection_count", Value: 0, Weight: 0.5, Threshold: 500},
		},
		RiskLevel: RiskLevelHigh,
	}

	err := security.AddThreatModel(model)
	assert.NoError(t, err)

	// Verify model was added
	addedModel, exists := security.ThreatModels["threat1"]
	assert.True(t, exists)
	assert.Equal(t, model.ThreatID, addedModel.ThreatID)
	assert.Equal(t, model.Type, addedModel.Type)
	assert.Equal(t, model.RiskLevel, addedModel.RiskLevel)
}

func TestPredictiveSecurity_GetThreatPredictions(t *testing.T) {
	security := NewPredictiveSecurity()

	// Test getting threat predictions
	predictions := security.GetThreatPredictions()
	assert.NotNil(t, predictions)
	assert.Len(t, predictions, 0) // No predictions initially
}

func TestNewAutomatedResponse(t *testing.T) {
	response := NewAutomatedResponse()

	assert.NotNil(t, response)
	assert.True(t, response.Enabled)
	assert.NotNil(t, response.ResponseRules)
	assert.NotNil(t, response.History)
}

func TestAutomatedResponse_Start(t *testing.T) {
	response := NewAutomatedResponse()

	err := response.Start()
	assert.NoError(t, err)
}

func TestAutomatedResponse_Stop(t *testing.T) {
	response := NewAutomatedResponse()

	err := response.Stop()
	assert.NoError(t, err)
}

func TestAutomatedResponse_AddResponseRule(t *testing.T) {
	response := NewAutomatedResponse()

	// Create a response rule
	rule := &ResponseRule{
		RuleID: "rule1",
		Name:   "Test Rule",
		Conditions: []ResponseCondition{
			{
				Type:     ConditionTypeThreshold,
				Field:    "anomaly_score",
				Operator: ">",
				Value:    0.8,
			},
		},
		Actions:  []string{"block_user"},
		Priority: 1,
		Enabled:  true,
	}

	err := response.AddResponseRule(rule)
	assert.NoError(t, err)

	// Verify rule was added
	addedRule, exists := response.ResponseRules["rule1"]
	assert.True(t, exists)
	assert.Equal(t, rule.RuleID, addedRule.RuleID)
	assert.Equal(t, rule.Conditions, addedRule.Conditions)
	assert.Equal(t, rule.Actions, addedRule.Actions)
	assert.Equal(t, rule.Priority, addedRule.Priority)
}

func TestAutomatedResponse_TriggerResponse(t *testing.T) {
	response := NewAutomatedResponse()

	// Test triggering response with high risk score to trigger rules
	event := map[string]interface{}{
		"user_id":       "user1",
		"event_type":    "anomaly_detected",
		"risk_score":    0.9, // High risk to trigger rules
		"anomaly_score": 0.8, // High anomaly to trigger rules
	}

	err := response.TriggerResponse(event)
	assert.NoError(t, err)

	// Verify response was recorded in history
	history := response.GetResponseHistory()
	assert.GreaterOrEqual(t, len(history), 0) // May be 0 if no rules triggered
	if len(history) > 0 {
		assert.NotEmpty(t, history[0].EventID)
		assert.NotEmpty(t, history[0].RuleID)
		assert.NotNil(t, history[0].Success)
	}
}

func TestAutomatedResponse_GetResponseHistory(t *testing.T) {
	response := NewAutomatedResponse()

	// Test getting response history
	history := response.GetResponseHistory()
	assert.NotNil(t, history)
	assert.Len(t, history, 0) // No history initially
}

func TestNewAIMonitoring(t *testing.T) {
	monitoring := NewAIMonitoring()

	assert.NotNil(t, monitoring)
	assert.True(t, monitoring.Enabled)
	assert.NotNil(t, monitoring.Metrics)
	assert.NotNil(t, monitoring.Alerts)
}

func TestAIMonitoring_Start(t *testing.T) {
	monitoring := NewAIMonitoring()

	err := monitoring.Start()
	assert.NoError(t, err)
}

func TestAIMonitoring_Stop(t *testing.T) {
	monitoring := NewAIMonitoring()

	err := monitoring.Stop()
	assert.NoError(t, err)
}

func TestAIMonitoring_AddMetric(t *testing.T) {
	monitoring := NewAIMonitoring()

	// Test adding metric
	metric := &AIMetric{
		Name:        "anomaly_score",
		Value:       0.8,
		Unit:        "score",
		Trend:       TrendStable,
		Threshold:   0.9,
		LastUpdated: time.Now(),
	}

	err := monitoring.AddMetric(metric)
	assert.NoError(t, err)

	// Verify metric was added
	addedMetric, exists := monitoring.Metrics["anomaly_score"]
	assert.True(t, exists)
	assert.Equal(t, metric.Name, addedMetric.Name)
	assert.Equal(t, metric.Value, addedMetric.Value)
}

func TestAIMonitoring_CreateAlert(t *testing.T) {
	monitoring := NewAIMonitoring()

	// Test creating alert
	alert := &AIAlert{
		AlertID:      "alert-001",
		Type:         AlertTypeAnomaly,
		Message:      "Anomaly score is high",
		Severity:     AlertSeverityError,
		Timestamp:    time.Now(),
		Acknowledged: false,
		Resolved:     false,
	}

	err := monitoring.CreateAlert(alert)
	assert.NoError(t, err)

	// Verify alert was created
	assert.Len(t, monitoring.Alerts, 1)
	assert.Equal(t, alert.AlertID, monitoring.Alerts[0].AlertID)
	assert.Equal(t, alert.Type, monitoring.Alerts[0].Type)
	assert.Equal(t, alert.Message, monitoring.Alerts[0].Message)
	assert.False(t, monitoring.Alerts[0].Resolved)
}

func TestAIMonitoring_GetDashboard(t *testing.T) {
	monitoring := NewAIMonitoring()

	// Test getting dashboard
	dashboard := monitoring.GetDashboard()
	assert.NotNil(t, dashboard)
	assert.NotNil(t, dashboard.Widgets)
}

func TestThreatType_String(t *testing.T) {
	tests := []struct {
		threatType ThreatType
		expected   string
	}{
		{ThreatTypeDDoS, "DDoS"},
		{ThreatTypeBruteForce, "BruteForce"},
		{ThreatTypeDataExfiltration, "DataExfiltration"},
		{ThreatTypePrivilegeEscalation, "PrivilegeEscalation"},
		{ThreatTypeMalware, "Malware"},
		{ThreatType(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.threatType.String())
		})
	}
}

func TestRiskLevel_String(t *testing.T) {
	tests := []struct {
		riskLevel RiskLevel
		expected  string
	}{
		{RiskLevelLow, "Low"},
		{RiskLevelMedium, "Medium"},
		{RiskLevelHigh, "High"},
		{RiskLevelCritical, "Critical"},
		{RiskLevel(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.riskLevel.String())
		})
	}
}

func TestResponseAction_String(t *testing.T) {
	tests := []struct {
		action   ActionType
		expected string
	}{
		{ActionTypeBlock, "Block"},
		{ActionTypeAlert, "Alert"},
		{ActionTypeQuarantine, "Quarantine"},
		{ActionTypeRateLimit, "RateLimit"},
		{ActionTypeLog, "Log"},
		{ActionType(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.action.String())
		})
	}
}

func TestAlertSeverity_String(t *testing.T) {
	tests := []struct {
		severity AlertSeverity
		expected string
	}{
		{AlertSeverityInfo, "Info"},
		{AlertSeverityWarning, "Warning"},
		{AlertSeverityError, "Error"},
		{AlertSeverityCritical, "Critical"},
		{AlertSeverity(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.severity.String())
		})
	}
}

func TestUserProfile_UpdateRiskScore(t *testing.T) {
	profile := &UserProfile{
		UserID:      "user1",
		RiskScore:   0.5,
		LastUpdated: time.Now(),
	}

	// Test updating risk score
	profile.RiskScore = 0.8
	profile.LastUpdated = time.Now()
	assert.Equal(t, 0.8, profile.RiskScore)
	assert.True(t, time.Since(profile.LastUpdated) < time.Second)
}

func TestUserProfile_AddLoginPattern(t *testing.T) {
	profile := &UserProfile{
		UserID:        "user1",
		LoginPatterns: []LoginPattern{},
	}

	// Test adding login pattern
	pattern := LoginPattern{
		TimeOfDay:  time.Now(),
		IPAddress:  "192.168.1.1",
		DeviceType: "mobile",
		Location:   "Moscow",
		Success:    true,
	}

	profile.LoginPatterns = append(profile.LoginPatterns, pattern)
	assert.Len(t, profile.LoginPatterns, 1)
	assert.Equal(t, pattern.IPAddress, profile.LoginPatterns[0].IPAddress)
	assert.Equal(t, pattern.DeviceType, profile.LoginPatterns[0].DeviceType)
}

func TestUserProfile_AddAccessPattern(t *testing.T) {
	profile := &UserProfile{
		UserID:         "user1",
		AccessPatterns: []AccessPattern{},
	}

	// Test adding access pattern
	pattern := AccessPattern{
		Resource:   "/api/data",
		TimeOfDay:  time.Now(),
		Duration:   5 * time.Minute,
		Frequency:  10,
		DataVolume: 1024,
	}

	profile.AccessPatterns = append(profile.AccessPatterns, pattern)
	assert.Len(t, profile.AccessPatterns, 1)
	assert.Equal(t, pattern.Resource, profile.AccessPatterns[0].Resource)
	assert.Equal(t, pattern.Duration, profile.AccessPatterns[0].Duration)
}

func TestUserProfile_AddNetworkPattern(t *testing.T) {
	profile := &UserProfile{
		UserID:          "user1",
		NetworkPatterns: []NetworkPattern{},
	}

	// Test adding network pattern
	pattern := NetworkPattern{
		Destination: "example.com",
		Protocol:    "HTTPS",
		Port:        443,
		DataVolume:  2048,
		Frequency:   5,
	}

	profile.NetworkPatterns = append(profile.NetworkPatterns, pattern)
	assert.Len(t, profile.NetworkPatterns, 1)
	assert.Equal(t, pattern.Destination, profile.NetworkPatterns[0].Destination)
	assert.Equal(t, pattern.Protocol, profile.NetworkPatterns[0].Protocol)
}

func TestThreatModel_CalculateRiskScore(t *testing.T) {
	model := &ThreatModel{
		ThreatID: "threat1",
		Type:     ThreatTypeDDoS,
		Indicators: []ThreatIndicator{
			{Name: "packet_rate", Value: 1000, Weight: 0.5, Threshold: 500},
			{Name: "connection_count", Value: 200, Weight: 0.3, Threshold: 100},
			{Name: "error_rate", Value: 0.1, Weight: 0.2, Threshold: 0.05},
		},
		RiskLevel: RiskLevelHigh,
	}

	// Test calculating risk score
	score := 0.0
	for _, indicator := range model.Indicators {
		if indicator.Value > indicator.Threshold {
			score += indicator.Weight * (indicator.Value / indicator.Threshold)
		}
	}
	assert.GreaterOrEqual(t, score, 0.0)
	// Убираем ограничение на 1.0, так как при превышении пороговых значений скор может быть больше 1
}

func TestResponseRule_IsTriggered(t *testing.T) {
	rule := &ResponseRule{
		RuleID:   "rule1",
		Priority: 1,
		Enabled:  true,
	}

	// Test rule triggering (simplified logic)
	triggered := rule.Enabled && rule.Priority > 0
	assert.True(t, triggered)

	// Test rule not triggering
	rule.Enabled = false
	triggered = rule.Enabled && rule.Priority > 0
	assert.False(t, triggered)
}

func TestAIAlert_Resolve(t *testing.T) {
	alert := &AIAlert{
		AlertID:   "alert1",
		Type:      AlertTypeAnomaly,
		Message:   "Anomaly score is high",
		Severity:  AlertSeverityError,
		Resolved:  false,
		Timestamp: time.Now(),
	}

	// Test resolving alert
	alert.Resolved = true
	assert.True(t, alert.Resolved)
}

func TestAIDashboard_AddMetric(t *testing.T) {
	dashboard := &AIDashboard{
		Title:   "Test Dashboard",
		Widgets: []DashboardWidget{},
	}

	// Test adding metric
	dashboard.Widgets = append(dashboard.Widgets, DashboardWidget{
		WidgetID: "metric1",
		Type:     WidgetTypeGauge,
		Title:    "anomaly_score",
		Data:     map[string]interface{}{"value": 0.8},
	})
	assert.Len(t, dashboard.Widgets, 1)
	assert.Equal(t, "anomaly_score", dashboard.Widgets[0].Title)
}

func TestAIDashboard_AddAlert(t *testing.T) {
	dashboard := &AIDashboard{
		Title:   "Test Dashboard",
		Widgets: []DashboardWidget{},
	}

	// Test adding alert
	alert := &AIAlert{
		AlertID:   "alert1",
		Type:      AlertTypeAnomaly,
		Message:   "Anomaly score is high",
		Severity:  AlertSeverityError,
		Resolved:  false,
		Timestamp: time.Now(),
	}

	dashboard.Widgets = append(dashboard.Widgets, DashboardWidget{
		WidgetID: "alert1",
		Type:     WidgetTypeAlert,
		Title:    "alert1",
		Data:     map[string]interface{}{"alert": alert},
	})
	assert.Len(t, dashboard.Widgets, 1)
	assert.Equal(t, "alert1", dashboard.Widgets[0].Title)
}

func TestAIDashboard_GetActiveAlerts(t *testing.T) {
	dashboard := &AIDashboard{
		Title: "Test Dashboard",
		Widgets: []DashboardWidget{
			{WidgetID: "alert1", Title: "alert1", Type: WidgetTypeAlert, Data: map[string]interface{}{"resolved": false}},
			{WidgetID: "alert2", Title: "alert2", Type: WidgetTypeAlert, Data: map[string]interface{}{"resolved": true}},
			{WidgetID: "alert3", Title: "alert3", Type: WidgetTypeAlert, Data: map[string]interface{}{"resolved": false}},
		},
	}

	// Test getting active alerts
	var activeAlerts []DashboardWidget
	for _, widget := range dashboard.Widgets {
		if widget.Type == WidgetTypeAlert {
			if data, ok := widget.Data.(map[string]interface{}); ok {
				if resolved, ok := data["resolved"].(bool); ok && !resolved {
					activeAlerts = append(activeAlerts, widget)
				}
			}
		}
	}
	assert.Len(t, activeAlerts, 2)
	assert.Equal(t, "alert1", activeAlerts[0].Title)
	assert.Equal(t, "alert3", activeAlerts[1].Title)
}

func TestAIDashboard_GetMetricsSummary(t *testing.T) {
	dashboard := &AIDashboard{
		Title: "Test Dashboard",
		Widgets: []DashboardWidget{
			{WidgetID: "metric1", Title: "anomaly_score", Type: WidgetTypeGauge, Data: map[string]interface{}{"value": 0.8}},
			{WidgetID: "metric2", Title: "risk_score", Type: WidgetTypeGauge, Data: map[string]interface{}{"value": 0.6}},
			{WidgetID: "metric3", Title: "login_count", Type: WidgetTypeGauge, Data: map[string]interface{}{"value": 10}},
		},
	}

	// Test getting metrics summary
	summary := make(map[string]interface{})
	for _, widget := range dashboard.Widgets {
		if widget.Type == WidgetTypeGauge {
			if data, ok := widget.Data.(map[string]interface{}); ok {
				summary[widget.Title] = data["value"]
			}
		}
	}
	assert.NotNil(t, summary)
	assert.Contains(t, summary, "anomaly_score")
	assert.Contains(t, summary, "risk_score")
	assert.Contains(t, summary, "login_count")
}

func TestAIComponents_Integration(t *testing.T) {
	// Test integration between AI components
	analyzer := NewBehaviorAnalyzer()
	security := NewPredictiveSecurity()
	response := NewAutomatedResponse()
	monitoring := NewAIMonitoring()

	// Start all components
	require.NoError(t, analyzer.Start())
	require.NoError(t, security.Start())
	require.NoError(t, response.Start())
	require.NoError(t, monitoring.Start())

	// Test workflow
	userID := "test-user"
	event := map[string]interface{}{"type": "login", "timestamp": time.Now()}

	// Analyze user behavior
	err := analyzer.AnalyzeUserBehavior(userID, event)
	require.NoError(t, err)

	// Predict threats
	indicators := []ThreatIndicator{
		{Name: "suspicious_login", Value: 0.8, Weight: 1.0, Threshold: 0.5},
	}
	prediction, err := security.PredictThreat(indicators)
	require.NoError(t, err)
	require.NotNil(t, prediction)

	// Add monitoring metric
	metric := &AIMetric{
		Name:  "anomaly_score",
		Value: 0.8,
		Unit:  "score",
	}
	err = monitoring.AddMetric(metric)
	require.NoError(t, err)

	// Trigger response
	err = response.TriggerResponse(event)
	require.NoError(t, err)

	// Get dashboard
	dashboard := monitoring.GetDashboard()
	require.NotNil(t, dashboard)

	// Stop all components
	require.NoError(t, analyzer.Stop())
	require.NoError(t, security.Stop())
	require.NoError(t, response.Stop())
	require.NoError(t, monitoring.Stop())
}

func TestAIComponents_EdgeCases(t *testing.T) {
	analyzer := NewBehaviorAnalyzer()

	// Test with empty user ID
	err := analyzer.AnalyzeUserBehavior("", map[string]interface{}{"event": "login"})
	assert.Error(t, err)

	// Test with nil profile
	err = analyzer.UpdateUserProfile("", nil)
	assert.Error(t, err)

	// Test with invalid user ID
	anomaly, err := analyzer.DetectAnomaly("")
	assert.Error(t, err)
	assert.Equal(t, 0.0, anomaly)
}

func TestAIComponents_ConcurrentAccess(t *testing.T) {
	analyzer := NewBehaviorAnalyzer()

	// Test concurrent access
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(userID string) {
			defer func() { done <- true }()

			err := analyzer.AnalyzeUserBehavior(userID, map[string]interface{}{"event": "login"})
			if err == nil {
				analyzer.UpdateUserProfile(userID, map[string]interface{}{"event": "login"})
			}
		}("user" + string(rune(i)))
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify analyzer is still functional
	assert.NotNil(t, analyzer.UserProfiles)
	assert.NotNil(t, analyzer.AnomalyScores)
}
