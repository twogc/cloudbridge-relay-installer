package workflows

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"go.uber.org/cadence/testsuite"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// WorkflowTestSuite представляет набор unit тестов для workflows
type WorkflowTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

// SetupTest выполняется перед каждым тестом
func (s *WorkflowTestSuite) SetupTest() {
	logger, _ := zap.NewDevelopment()
	s.SetLogger(logger)
}

// TestModuleInitializationWorkflow тестирует workflow инициализации модулей
func (s *WorkflowTestSuite) TestModuleInitializationWorkflow() {
	env := s.NewTestWorkflowEnvironment()
	
	// Мокаем activities
	env.OnActivity("CheckDependenciesActivity", mock.Anything).Return("all_dependencies_available", nil)
	env.OnActivity("InitializeModuleActivity", mock.Anything, mock.Anything).Return("module_test_initialized", nil)
	env.OnActivity("CheckModuleHealthActivity", mock.Anything).Return("healthy", nil)
	
	// Входные данные
	input := ModuleInitializationInput{
		ModuleName:    "test",
		Configuration: map[string]interface{}{"key": "value"},
		Dependencies:  []string{"dep1", "dep2"},
	}
	
	// Выполнение workflow
	env.ExecuteWorkflow(ModuleInitializationWorkflow, input)
	
	// Проверка результата
	assert.True(s.T(), env.IsWorkflowCompleted())
	assert.NoError(s.T(), env.GetWorkflowError())
	
	var result ModuleInitializationResult
	err := env.GetWorkflowResult(&result)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "test", result.ModuleName)
	assert.Equal(s.T(), "completed", result.Status)
}

// TestHealthCheckWorkflow тестирует workflow проверки здоровья
func (s *WorkflowTestSuite) TestHealthCheckWorkflow() {
	env := s.NewTestWorkflowEnvironment()
	
	// Мокаем activity
	env.OnActivity("PerformHealthCheckActivity", mock.Anything, mock.Anything, mock.Anything).Return(
		map[string]interface{}{
			"status": "healthy",
			"score":  0.95,
		}, nil)
	
	// Установка короткого таймаута для тестов
	env.SetTestTimeout(time.Second * 30)
	
	// Входные данные
	input := HealthCheckInput{
		Component: "test-component",
		Checks:    []string{"core"},
		Config:    map[string]interface{}{},
	}
	
	// Запуск workflow в отдельной горутине
	go func() {
		env.ExecuteWorkflow(HealthCheckWorkflow, input)
	}()
	
	// Ожидание запуска workflow
	time.Sleep(100 * time.Millisecond)
	
	// Workflow должен выполняться непрерывно, поэтому проверяем что он запущен
	assert.False(s.T(), env.IsWorkflowCompleted())
	
	// Отправляем сигнал остановки
	env.SignalWorkflow("control-command", "shutdown")
	
	// Ожидание завершения с таймаутом
	done := make(chan bool)
	go func() {
		for !env.IsWorkflowCompleted() && env.GetWorkflowError() == nil {
			time.Sleep(10 * time.Millisecond)
		}
		done <- true
	}()
	
	select {
	case <-done:
		// Теперь workflow должен завершиться
		assert.True(s.T(), env.IsWorkflowCompleted())
		assert.NoError(s.T(), env.GetWorkflowError())
		
		var result HealthCheckResult
		err := env.GetWorkflowResult(&result)
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), "test-component", result.Component)
		assert.Equal(s.T(), "stopped", result.Status)
	case <-time.After(10 * time.Second):
		s.T().Fatal("HealthCheckWorkflow не завершился в ожидаемое время")
	}
}

// TestPeerLifecycleWorkflow тестирует workflow управления пирами
func (s *WorkflowTestSuite) TestPeerLifecycleWorkflow() {
	env := s.NewTestWorkflowEnvironment()
	
	// Мокаем activities для операции добавления пира
	env.OnActivity("ValidatePeerConfigActivity", mock.Anything).Return(
		map[string]interface{}{"valid": true, "status": "valid"}, nil)
	env.OnActivity("CreateWireGuardInterfaceActivity", mock.Anything, mock.Anything).Return(
		map[string]interface{}{
			"interface_name": "wg-test",
			"assigned_ip":    "10.0.0.10/24",
			"status":         "created",
		}, nil)
	env.OnActivity("RegisterPeerActivity", mock.Anything, mock.Anything).Return(
		map[string]interface{}{"registered": true, "status": "registered"}, nil)
	
	// Входные данные
	input := PeerLifecycleWorkflowInput{
		PeerID:    "test-peer-123",
		Operation: "add",
		PeerConfig: map[string]interface{}{
			"public_key":  "test-key",
			"allowed_ips": []string{"10.0.0.0/24"},
		},
		NetworkConfig: map[string]interface{}{
			"network_id": "test-net",
		},
	}
	
	// Выполнение workflow
	env.ExecuteWorkflow(PeerLifecycleWorkflow, input)
	
	// Проверка результата
	assert.True(s.T(), env.IsWorkflowCompleted())
	assert.NoError(s.T(), env.GetWorkflowError())
	
	var result PeerLifecycleWorkflowResult
	err := env.GetWorkflowResult(&result)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "test-peer-123", result.PeerID)
	assert.Equal(s.T(), "add", result.Operation)
	assert.Equal(s.T(), "completed", result.Status)
}

// TestQuantumKeyRotationWorkflow тестирует workflow ротации квантовых ключей
func (s *WorkflowTestSuite) TestQuantumKeyRotationWorkflow() {
	env := s.NewTestWorkflowEnvironment()
	
	// Мокаем activities
	env.OnActivity("CheckKeyRotationNeedActivity", mock.Anything, mock.Anything).Return(
		map[string]interface{}{"rotation_needed": true}, nil)
	env.OnActivity("GenerateQuantumKeyActivity", mock.Anything, mock.Anything).Return(
		map[string]interface{}{
			"new_key_id": "new-key-123",
			"algorithm":  "kyber",
			"generated":  true,
		}, nil)
	env.OnActivity("ValidateQuantumKeyActivity", mock.Anything).Return(
		map[string]interface{}{"valid": true}, nil)
	env.OnActivity("ReplaceQuantumKeyActivity", mock.Anything, mock.Anything).Return(
		map[string]interface{}{"replaced": true}, nil)
	env.OnActivity("SecureDeleteQuantumKeyActivity", mock.Anything).Return(
		map[string]interface{}{"deleted": true}, nil)
	env.OnActivity("UpdateKeyReferencesActivity", mock.Anything, mock.Anything).Return(
		map[string]interface{}{"updated": true}, nil)
	
	// Входные данные
	input := QuantumKeyRotationWorkflowInput{
		KeyID:          "old-key-123",
		Algorithm:      "kyber",
		RotationPolicy: "time_based",
		Configuration:  map[string]interface{}{},
		ForceRotation:  true,
	}
	
	// Выполнение workflow
	env.ExecuteWorkflow(QuantumKeyRotationWorkflow, input)
	
	// Проверка результата
	assert.True(s.T(), env.IsWorkflowCompleted())
	assert.NoError(s.T(), env.GetWorkflowError())
	
	var result QuantumKeyRotationWorkflowResult
	err := env.GetWorkflowResult(&result)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "old-key-123", result.OldKeyID)
	assert.Equal(s.T(), "new-key-123", result.NewKeyID)
	assert.Equal(s.T(), "kyber", result.Algorithm)
	assert.Equal(s.T(), "completed", result.Status)
}

// TestThreatAnalysisWorkflow тестирует AI workflow анализа угроз
func (s *WorkflowTestSuite) TestThreatAnalysisWorkflow() {
	env := s.NewTestWorkflowEnvironment()
	
	// Мокаем activities
	env.OnActivity("CollectThreatDataActivity", mock.Anything, mock.Anything).Return(
		map[string]interface{}{
			"data_collected": 1000,
			"sources":        []string{"logs", "network"},
		}, nil)
	env.OnActivity("PreprocessThreatDataActivity", mock.Anything).Return(
		map[string]interface{}{"preprocessed": true}, nil)
	env.OnActivity("PerformRealtimeThreatAnalysisActivity", mock.Anything, mock.Anything).Return(
		map[string]interface{}{
			"threats_detected": []string{"ddos"},
			"risk_level":       "medium",
			"confidence":       0.85,
		}, nil)
	
	// Входные данные
	input := ThreatAnalysisWorkflowInput{
		AnalysisType: "realtime",
		DataSources:  []string{"logs", "network"},
		TimeWindow:   "1h",
		ThreatModels: []string{"ddos", "intrusion"},
		Configuration: map[string]interface{}{
			"ai_model": "v1.0",
		},
	}
	
	// Выполнение workflow
	env.ExecuteWorkflow(ThreatAnalysisWorkflow, input)
	
	// Проверка результата
	assert.True(s.T(), env.IsWorkflowCompleted())
	assert.NoError(s.T(), env.GetWorkflowError())
	
	var result ThreatAnalysisWorkflowResult
	err := env.GetWorkflowResult(&result)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "realtime", result.AnalysisType)
	assert.Equal(s.T(), "threats_detected", result.Status)
	assert.Equal(s.T(), "medium", result.RiskLevel)
	assert.Contains(s.T(), result.DetectedThreats, "ddos")
}

// TestAutoScalingWorkflow тестирует workflow автомасштабирования
func (s *WorkflowTestSuite) TestAutoScalingWorkflow() {
	env := s.NewTestWorkflowEnvironment()
	
	// Мокаем activities для реактивного масштабирования
	env.OnActivity("CollectScalingMetricsActivity", mock.Anything, mock.Anything).Return(
		map[string]interface{}{
			"current_capacity": map[string]interface{}{"cpu": 30, "memory": 25},
			"metrics_collected": true,
		}, nil)
	env.OnActivity("AnalyzeScalingNeedActivity", mock.Anything, mock.Anything, mock.Anything).Return(
		map[string]interface{}{
			"need_scaling":    false,
			"scaling_action": "no_scaling_needed",
			"reason":         "metrics_within_threshold",
		}, nil)
	env.OnActivity("GetCurrentCapacityActivity", mock.Anything).Return(
		map[string]interface{}{
			"current_capacity": map[string]interface{}{"instances": 2},
			"target_capacity":  map[string]interface{}{"instances": 2},
		}, nil)
	
	// Установка короткого таймаута для тестов
	env.SetTestTimeout(time.Second * 30)
	
	// Входные данные
	input := AutoScalingWorkflowInput{
		ResourceType:  "compute",
		ScalingPolicy: "reactive",
		TargetMetrics: map[string]float64{"cpu": 70, "memory": 80},
		ScalingLimits: map[string]interface{}{"min_instances": 1, "max_instances": 5},
		Configuration: map[string]interface{}{},
	}
	
	// Запуск workflow в отдельной горутине
	go func() {
		env.ExecuteWorkflow(AutoScalingWorkflow, input)
	}()
	
	// Ожидание запуска workflow
	time.Sleep(100 * time.Millisecond)
	
	// Отправляем сигнал остановки
	env.SignalWorkflow("control-command", "shutdown")
	
	// Ожидание завершения с таймаутом
	done := make(chan bool)
	go func() {
		for !env.IsWorkflowCompleted() && env.GetWorkflowError() == nil {
			time.Sleep(10 * time.Millisecond)
		}
		done <- true
	}()
	
	select {
	case <-done:
		assert.True(s.T(), env.IsWorkflowCompleted())
		assert.NoError(s.T(), env.GetWorkflowError())
		
		var result AutoScalingWorkflowResult
		err := env.GetWorkflowResult(&result)
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), "compute", result.ResourceType)
		assert.Equal(s.T(), "reactive", result.ScalingPolicy)
		assert.Equal(s.T(), "stopped", result.Status)
	case <-time.After(10 * time.Second):
		s.T().Fatal("AutoScalingWorkflow не завершился в ожидаемое время")
	}
}

// TestSecurityComplianceWorkflow тестирует workflow проверки соответствия
func (s *WorkflowTestSuite) TestSecurityComplianceWorkflow() {
	env := s.NewTestWorkflowEnvironment()
	
	// Мокаем activities
	env.OnActivity("GetComplianceRequirementsActivity", mock.Anything, mock.Anything, mock.Anything).Return(
		map[string]interface{}{
			"requirements": []string{"firewall", "encryption", "access_control"},
		}, nil)
	env.OnActivity("AuditSecurityAreaActivity", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		map[string]interface{}{
			"passed_controls": []string{"firewall", "encryption"},
			"failed_controls": []string{"access_control"},
		}, nil)
	env.OnActivity("CalculateComplianceScoreActivity", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		map[string]interface{}{
			"compliance_score": 0.75,
		}, nil)
	env.OnActivity("GenerateComplianceRecommendationsActivity", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		map[string]interface{}{
			"recommendations": []string{"Improve access control policies"},
		}, nil)
	env.OnActivity("GenerateComplianceReportActivity", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		map[string]interface{}{
			"report_generated": true,
		}, nil)
	
	// Входные данные
	input := SecurityComplianceWorkflowInput{
		ComplianceFramework: "iso27001",
		AuditScope:          []string{"network", "data"},
		ComplianceLevel:     "standard",
		Configuration:       map[string]interface{}{},
	}
	
	// Выполнение workflow
	env.ExecuteWorkflow(SecurityComplianceWorkflow, input)
	
	// Проверка результата
	assert.True(s.T(), env.IsWorkflowCompleted())
	assert.NoError(s.T(), env.GetWorkflowError())
	
	var result SecurityComplianceWorkflowResult
	err := env.GetWorkflowResult(&result)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "iso27001", result.ComplianceFramework)
	assert.Equal(s.T(), 0.75, result.ComplianceScore)
	assert.Contains(s.T(), result.PassedControls, "firewall")
	assert.Contains(s.T(), result.FailedControls, "access_control")
}

// TestWorkflowWithTimeout тестирует workflow с таймаутом
func (s *WorkflowTestSuite) TestWorkflowWithTimeout() {
	env := s.NewTestWorkflowEnvironment()
	
	// Настройка короткого таймаута
	env.SetTestTimeout(time.Second * 5)
	
	// Мокаем activity которая будет выполняться долго
	env.OnActivity("LongRunningActivity", mock.Anything).After(time.Minute).Return("result", nil)
	
	// Простой workflow для тестирования таймаута
	testWorkflow := func(ctx workflow.Context) error {
		return workflow.ExecuteActivity(ctx, "LongRunningActivity", nil).Get(ctx, nil)
	}
	
	// Выполнение workflow
	env.ExecuteWorkflow(testWorkflow)
	
	// Workflow должен завершиться по таймауту
	assert.True(s.T(), env.IsWorkflowCompleted())
	assert.Error(s.T(), env.GetWorkflowError())
}

// TestWorkflowWithRetry тестирует workflow с retry политикой
func (s *WorkflowTestSuite) TestWorkflowWithRetry() {
	env := s.NewTestWorkflowEnvironment()
	
	// Мокаем activity которая сначала падает, потом работает
	env.OnActivity("UnreliableActivity", mock.Anything).Return("", assert.AnError).Once()
	env.OnActivity("UnreliableActivity", mock.Anything).Return("success", nil).Once()
	
	// Простой workflow с retry
	testWorkflow := func(ctx workflow.Context) (string, error) {
		options := workflow.ActivityOptions{
			StartToCloseTimeout: time.Minute,
			RetryPolicy: &workflow.RetryPolicy{
				InitialInterval:    time.Second,
				BackoffCoefficient: 2.0,
				MaximumInterval:    time.Minute,
				MaximumAttempts:    3,
			},
		}
		ctx = workflow.WithActivityOptions(ctx, options)
		
		var result string
		err := workflow.ExecuteActivity(ctx, "UnreliableActivity", nil).Get(ctx, &result)
		return result, err
	}
	
	// Выполнение workflow
	env.ExecuteWorkflow(testWorkflow)
	
	// Workflow должен завершиться успешно после retry
	assert.True(s.T(), env.IsWorkflowCompleted())
	assert.NoError(s.T(), env.GetWorkflowError())
	
	var result string
	err := env.GetWorkflowResult(&result)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "success", result)
}

// TestWorkflowTestSuite запускает все unit тесты для workflows
func TestWorkflowTestSuite(t *testing.T) {
	suite.Run(t, new(WorkflowTestSuite))
}
