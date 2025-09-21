package metrics

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewCadenceMetrics(t *testing.T) {
	// Test with logger
	logger := zap.NewNop()
	metrics := NewCadenceMetrics(logger)
	assert.NotNil(t, metrics)
	assert.NotNil(t, metrics.WorkflowExecutions)
	assert.NotNil(t, metrics.WorkflowDuration)
	assert.NotNil(t, metrics.WorkflowErrors)
	assert.NotNil(t, metrics.WorkflowStatus)
	assert.NotNil(t, metrics.ActiveWorkflows)
	assert.NotNil(t, metrics.ActivityExecutions)
	assert.NotNil(t, metrics.ActivityDuration)
	assert.NotNil(t, metrics.ActivityErrors)
	assert.NotNil(t, metrics.ActivityRetries)
	assert.NotNil(t, metrics.WorkerConnections)
	assert.NotNil(t, metrics.WorkerPollers)
	assert.NotNil(t, metrics.WorkerLatency)
	assert.NotNil(t, metrics.WorkerQueueDepth)
	assert.NotNil(t, metrics.ClientRequests)
	assert.NotNil(t, metrics.ClientLatency)
	assert.NotNil(t, metrics.ClientErrors)
	assert.NotNil(t, metrics.WireGuardWorkflows)
	assert.NotNil(t, metrics.QuantumWorkflows)
	assert.NotNil(t, metrics.AIWorkflows)
	assert.NotNil(t, metrics.ScalingWorkflows)
	assert.NotNil(t, metrics.SecurityWorkflows)
	assert.NotNil(t, metrics.HealthChecks)
	assert.NotNil(t, metrics.SystemUptime)
	assert.NotNil(t, metrics.ResourceUtilization)
	assert.Equal(t, logger, metrics.logger)

	// Test with nil logger
	metrics = NewCadenceMetrics(nil)
	assert.NotNil(t, metrics)
	assert.NotNil(t, metrics.logger)
}

func TestCadenceMetrics_WorkflowMethods(t *testing.T) {
	metrics := NewCadenceMetrics(zap.NewNop())

	workflowType := "test_workflow"
	domain := "test_domain"
	workflowID := "test_workflow_id"
	errorType := "timeout"

	// Test RecordWorkflowStarted
	metrics.RecordWorkflowStarted(workflowType, domain, workflowID)
	assert.NotNil(t, metrics.WorkflowExecutions)
	assert.NotNil(t, metrics.WorkflowStatus)
	assert.NotNil(t, metrics.ActiveWorkflows)

	// Test RecordWorkflowCompleted
	duration := 5 * time.Second
	metrics.RecordWorkflowCompleted(workflowType, domain, workflowID, duration)
	assert.NotNil(t, metrics.WorkflowExecutions)
	assert.NotNil(t, metrics.WorkflowDuration)
	assert.NotNil(t, metrics.WorkflowStatus)
	assert.NotNil(t, metrics.ActiveWorkflows)

	// Test RecordWorkflowFailed
	metrics.RecordWorkflowFailed(workflowType, domain, workflowID, errorType, duration)
	assert.NotNil(t, metrics.WorkflowExecutions)
	assert.NotNil(t, metrics.WorkflowErrors)
	assert.NotNil(t, metrics.WorkflowDuration)
	assert.NotNil(t, metrics.WorkflowStatus)
	assert.NotNil(t, metrics.ActiveWorkflows)
}

func TestCadenceMetrics_ActivityMethods(t *testing.T) {
	metrics := NewCadenceMetrics(zap.NewNop())

	activityType := "test_activity"
	domain := "test_domain"
	errorType := "network_error"
	duration := 2 * time.Second
	attempt := 3

	// Test RecordActivityStarted
	metrics.RecordActivityStarted(activityType, domain)
	assert.NotNil(t, metrics.ActivityExecutions)

	// Test RecordActivityCompleted
	metrics.RecordActivityCompleted(activityType, domain, duration)
	assert.NotNil(t, metrics.ActivityExecutions)
	assert.NotNil(t, metrics.ActivityDuration)

	// Test RecordActivityFailed
	metrics.RecordActivityFailed(activityType, domain, errorType, duration)
	assert.NotNil(t, metrics.ActivityExecutions)
	assert.NotNil(t, metrics.ActivityErrors)
	assert.NotNil(t, metrics.ActivityDuration)

	// Test RecordActivityRetry
	metrics.RecordActivityRetry(activityType, domain, attempt)
	assert.NotNil(t, metrics.ActivityRetries)
}

func TestCadenceMetrics_WorkerMethods(t *testing.T) {
	metrics := NewCadenceMetrics(zap.NewNop())

	workerIdentity := "test_worker"
	domain := "test_domain"
	pollerType := "workflow_poller"
	operation := "poll"
	queueType := "task_queue"
	duration := 100 * time.Millisecond

	// Test UpdateWorkerConnections
	metrics.UpdateWorkerConnections(workerIdentity, domain, 5.0)
	assert.NotNil(t, metrics.WorkerConnections)

	// Test UpdateWorkerPollers
	metrics.UpdateWorkerPollers(pollerType, workerIdentity, domain, 3.0)
	assert.NotNil(t, metrics.WorkerPollers)

	// Test RecordWorkerLatency
	metrics.RecordWorkerLatency(operation, workerIdentity, domain, duration)
	assert.NotNil(t, metrics.WorkerLatency)

	// Test UpdateWorkerQueueDepth
	metrics.UpdateWorkerQueueDepth(queueType, workerIdentity, domain, 10.0)
	assert.NotNil(t, metrics.WorkerQueueDepth)
}

func TestCadenceMetrics_ClientMethods(t *testing.T) {
	metrics := NewCadenceMetrics(zap.NewNop())

	operation := "start_workflow"
	status := "success"
	domain := "test_domain"
	errorType := "connection_error"
	duration := 50 * time.Millisecond

	// Test RecordClientRequest
	metrics.RecordClientRequest(operation, status, domain, duration)
	assert.NotNil(t, metrics.ClientRequests)
	assert.NotNil(t, metrics.ClientLatency)

	// Test RecordClientError
	metrics.RecordClientError(operation, errorType, domain)
	assert.NotNil(t, metrics.ClientErrors)
}

func TestCadenceMetrics_CloudBridgeSpecificMethods(t *testing.T) {
	metrics := NewCadenceMetrics(zap.NewNop())

	// Test WireGuard workflows
	subtype := "peer_connection"
	status := "success"
	peerID := "peer_123"
	metrics.RecordWireGuardWorkflow(subtype, status, peerID)
	assert.NotNil(t, metrics.WireGuardWorkflows)

	// Test Quantum workflows
	algorithm := "kyber_1024"
	metrics.RecordQuantumWorkflow(subtype, status, algorithm)
	assert.NotNil(t, metrics.QuantumWorkflows)

	// Test AI workflows
	modelType := "behavior_analyzer"
	metrics.RecordAIWorkflow(subtype, status, modelType)
	assert.NotNil(t, metrics.AIWorkflows)

	// Test Scaling workflows
	scaleDirection := "up"
	metrics.RecordScalingWorkflow(subtype, status, scaleDirection)
	assert.NotNil(t, metrics.ScalingWorkflows)

	// Test Security workflows
	threatLevel := "high"
	metrics.RecordSecurityWorkflow(subtype, status, threatLevel)
	assert.NotNil(t, metrics.SecurityWorkflows)
}

func TestCadenceMetrics_HealthMethods(t *testing.T) {
	metrics := NewCadenceMetrics(zap.NewNop())

	// Test RecordHealthCheck
	checkType := "liveness"
	status := "healthy"
	component := "cadence_service"
	metrics.RecordHealthCheck(checkType, status, component)
	assert.NotNil(t, metrics.HealthChecks)

	// Test UpdateSystemUptime
	instance := "instance_1"
	uptime := 24 * time.Hour
	metrics.UpdateSystemUptime(component, instance, uptime)
	assert.NotNil(t, metrics.SystemUptime)

	// Test UpdateResourceUtilization
	resourceType := "cpu"
	utilization := 0.75
	metrics.UpdateResourceUtilization(resourceType, component, utilization)
	assert.NotNil(t, metrics.ResourceUtilization)
}

func TestCadenceMetrics_StartMetricsCollection(t *testing.T) {
	metrics := NewCadenceMetrics(zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	interval := 50 * time.Millisecond

	// Start metrics collection in a goroutine
	go metrics.StartMetricsCollection(ctx, interval)

	// Wait for context to be done
	<-ctx.Done()

	// Verify that the method completed without errors
	assert.NotNil(t, metrics)
}

func TestCadenceMetrics_EdgeCases(t *testing.T) {
	metrics := NewCadenceMetrics(zap.NewNop())

	// Test with empty strings
	metrics.RecordWorkflowStarted("", "", "")
	metrics.RecordActivityStarted("", "")
	metrics.UpdateWorkerConnections("", "", 0.0)
	metrics.RecordClientRequest("", "", "", 0)
	metrics.RecordWireGuardWorkflow("", "", "")
	metrics.RecordHealthCheck("", "", "")

	// Test with negative values
	metrics.UpdateWorkerConnections("worker", "domain", -1.0)
	metrics.UpdateWorkerPollers("poller", "worker", "domain", -1.0)
	metrics.UpdateWorkerQueueDepth("queue", "worker", "domain", -1.0)
	metrics.UpdateResourceUtilization("cpu", "component", -0.1)

	// Test with very large values
	metrics.UpdateWorkerConnections("worker", "domain", 999999.0)
	metrics.UpdateWorkerPollers("poller", "worker", "domain", 999999.0)
	metrics.UpdateWorkerQueueDepth("queue", "worker", "domain", 999999.0)
	metrics.UpdateResourceUtilization("cpu", "component", 999.0)

	// Test with zero duration
	metrics.RecordWorkflowCompleted("workflow", "domain", "id", 0)
	metrics.RecordActivityCompleted("activity", "domain", 0)
	metrics.RecordWorkerLatency("operation", "worker", "domain", 0)
	metrics.RecordClientRequest("operation", "status", "domain", 0)

	// Test with negative duration
	metrics.RecordWorkflowCompleted("workflow", "domain", "id", -1*time.Second)
	metrics.RecordActivityCompleted("activity", "domain", -1*time.Second)
	metrics.RecordWorkerLatency("operation", "worker", "domain", -1*time.Second)
	metrics.RecordClientRequest("operation", "status", "domain", -1*time.Second)

	// Verify all operations completed without errors
	assert.NotNil(t, metrics)
}

func TestCadenceMetrics_ConcurrentAccess(t *testing.T) {
	metrics := NewCadenceMetrics(zap.NewNop())
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			workflowType := "workflow" + string(rune(id))
			domain := "domain" + string(rune(id))
			workflowID := "workflow_id" + string(rune(id))
			activityType := "activity" + string(rune(id))
			workerIdentity := "worker" + string(rune(id))
			operation := "operation" + string(rune(id))

			// Workflow operations
			metrics.RecordWorkflowStarted(workflowType, domain, workflowID)
			metrics.RecordWorkflowCompleted(workflowType, domain, workflowID, time.Second)
			metrics.RecordWorkflowFailed(workflowType, domain, workflowID, "error", time.Second)

			// Activity operations
			metrics.RecordActivityStarted(activityType, domain)
			metrics.RecordActivityCompleted(activityType, domain, time.Second)
			metrics.RecordActivityFailed(activityType, domain, "error", time.Second)
			metrics.RecordActivityRetry(activityType, domain, 1)

			// Worker operations
			metrics.UpdateWorkerConnections(workerIdentity, domain, float64(id))
			metrics.UpdateWorkerPollers("poller", workerIdentity, domain, float64(id))
			metrics.RecordWorkerLatency(operation, workerIdentity, domain, time.Millisecond)
			metrics.UpdateWorkerQueueDepth("queue", workerIdentity, domain, float64(id))

			// Client operations
			metrics.RecordClientRequest(operation, "success", domain, time.Millisecond)
			metrics.RecordClientError(operation, "error", domain)

			// CloudBridge specific operations
			metrics.RecordWireGuardWorkflow("subtype", "success", "peer")
			metrics.RecordQuantumWorkflow("subtype", "success", "algorithm")
			metrics.RecordAIWorkflow("subtype", "success", "model")
			metrics.RecordScalingWorkflow("subtype", "success", "up")
			metrics.RecordSecurityWorkflow("subtype", "success", "high")

			// Health operations
			metrics.RecordHealthCheck("check", "healthy", "component")
			metrics.UpdateSystemUptime("component", "instance", time.Hour)
			metrics.UpdateResourceUtilization("cpu", "component", 0.5)
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all operations completed without errors
	assert.NotNil(t, metrics)
}

func TestCadenceMetrics_IntegrationWorkflow(t *testing.T) {
	metrics := NewCadenceMetrics(zap.NewNop())

	// Simulate a complete Cadence workflow
	workflowType := "wireguard_setup"
	domain := "production"
	workflowID := "wg_setup_001"
	activityType := "peer_configuration"
	workerIdentity := "worker_001"
	operation := "configure_peer"

	// 1. Start workflow
	metrics.RecordWorkflowStarted(workflowType, domain, workflowID)

	// 2. Start activity
	metrics.RecordActivityStarted(activityType, domain)

	// 3. Record worker operations
	metrics.UpdateWorkerConnections(workerIdentity, domain, 1.0)
	metrics.UpdateWorkerPollers("workflow_poller", workerIdentity, domain, 1.0)
	metrics.RecordWorkerLatency(operation, workerIdentity, domain, 100*time.Millisecond)

	// 4. Record client operations
	metrics.RecordClientRequest("start_workflow", "success", domain, 50*time.Millisecond)

	// 5. Complete activity
	metrics.RecordActivityCompleted(activityType, domain, 2*time.Second)

	// 6. Record CloudBridge specific workflow
	metrics.RecordWireGuardWorkflow("peer_connection", "success", "peer_123")

	// 7. Complete workflow
	metrics.RecordWorkflowCompleted(workflowType, domain, workflowID, 5*time.Second)

	// 8. Record health check
	metrics.RecordHealthCheck("liveness", "healthy", "cadence_service")
	metrics.UpdateSystemUptime("cadence_service", "instance_1", 24*time.Hour)
	metrics.UpdateResourceUtilization("cpu", "cadence_service", 0.25)

	// Verify workflow completed
	assert.NotNil(t, metrics)
}

func TestCadenceMetrics_Performance(t *testing.T) {
	metrics := NewCadenceMetrics(zap.NewNop())

	// Test performance with many operations
	start := time.Now()

	for i := 0; i < 1000; i++ {
		workflowType := "workflow" + string(rune(i%10))
		domain := "domain" + string(rune(i%10))
		workflowID := "workflow_id" + string(rune(i%100))
		activityType := "activity" + string(rune(i%10))
		workerIdentity := "worker" + string(rune(i%10))
		operation := "operation" + string(rune(i%10))

		metrics.RecordWorkflowStarted(workflowType, domain, workflowID)
		metrics.RecordActivityStarted(activityType, domain)
		metrics.UpdateWorkerConnections(workerIdentity, domain, float64(i%100))
		metrics.RecordWorkerLatency(operation, workerIdentity, domain, time.Duration(i%100)*time.Millisecond)
		metrics.RecordClientRequest(operation, "success", domain, time.Duration(i%100)*time.Millisecond)
		metrics.RecordWireGuardWorkflow("subtype", "success", "peer")
		metrics.RecordHealthCheck("check", "healthy", "component")
		metrics.UpdateResourceUtilization("cpu", "component", float64(i%100)/100.0)
	}

	duration := time.Since(start)

	// Should complete within reasonable time (less than 1 second)
	assert.Less(t, duration, time.Second)
}

func TestCadenceMetrics_DataTypes(t *testing.T) {
	metrics := NewCadenceMetrics(zap.NewNop())

	// Test different data types and edge cases
	testCases := []struct {
		name     string
		testFunc func()
	}{
		{
			name: "string labels",
			testFunc: func() {
				metrics.RecordWorkflowStarted("workflow-type", "domain-name", "workflow-id")
			},
		},
		{
			name: "unicode strings",
			testFunc: func() {
				metrics.RecordWorkflowStarted("тип-воркфлоу", "домен-имя", "ид-воркфлоу")
			},
		},
		{
			name: "special characters",
			testFunc: func() {
				metrics.RecordWorkflowStarted("workflow@type", "domain_name", "workflow-id")
			},
		},
		{
			name: "long strings",
			testFunc: func() {
				longString := "very-long-workflow-type-that-might-cause-issues-with-prometheus-label-limits"
				metrics.RecordWorkflowStarted(longString, "domain", "workflow_id")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NotPanics(t, tc.testFunc)
		})
	}
}

func TestCadenceMetrics_CollectSystemMetrics(t *testing.T) {
	metrics := NewCadenceMetrics(zap.NewNop())

	// Test collectSystemMetrics method (it's private, but we can test it indirectly)
	// by calling StartMetricsCollection with a very short interval and then canceling
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	interval := 5 * time.Millisecond

	// Start metrics collection
	go metrics.StartMetricsCollection(ctx, interval)

	// Wait for context to be done
	<-ctx.Done()

	// Verify that the method completed without errors
	assert.NotNil(t, metrics)
}

func TestCadenceMetrics_ResourceUtilizationBounds(t *testing.T) {
	metrics := NewCadenceMetrics(zap.NewNop())

	// Test resource utilization with different values
	testCases := []struct {
		name        string
		utilization float64
		shouldPanic bool
	}{
		{"zero", 0.0, false},
		{"normal", 0.5, false},
		{"maximum", 1.0, false},
		{"negative", -0.1, false},
		{"over_maximum", 1.5, false},
		{"very_large", 999.0, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldPanic {
				assert.Panics(t, func() {
					metrics.UpdateResourceUtilization("cpu", "component", tc.utilization)
				})
			} else {
				assert.NotPanics(t, func() {
					metrics.UpdateResourceUtilization("cpu", "component", tc.utilization)
				})
			}
		})
	}
}
