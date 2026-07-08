package cadence

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestNewClient(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := Config{
		HostPort: "localhost:7933",
		Domain:   "test-domain",
		TaskList: "test-task-list",
	}

	client, err := NewClient(config, logger)

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.Production())
	assert.True(t, client.IsStubMode() || client.Production() != nil)
}

func TestClient_Start(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := Config{
		HostPort: "localhost:7933",
		Domain:   "test-domain",
		TaskList: "test-task-list",
	}

	client, err := NewClient(config, logger)
	require.NoError(t, err)

	// Test starting client
	err = client.Start()
	assert.NoError(t, err)
}

func TestClient_Stop(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := Config{
		HostPort: "localhost:7933",
		Domain:   "test-domain",
		TaskList: "test-task-list",
	}

	client, err := NewClient(config, logger)
	assert.NoError(t, err)

	// Test stopping client (Stop() doesn't return error)
	client.Stop()
}

func TestClient_ExecuteWorkflow(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := Config{
		HostPort: "localhost:7933",
		Domain:   "test-domain",
		TaskList: "test-task-list",
	}

	client, err := NewClient(config, logger)
	assert.NoError(t, err)

	ctx := context.Background()
	workflowType := "test-workflow"
	input := map[string]interface{}{
		"param1": "value1",
		"param2": 42,
	}

	// Test executing workflow
	workflowID, err := client.ExecuteWorkflow(ctx, workflowType, input)
	assert.NoError(t, err)
	assert.NotEmpty(t, workflowID)
	assert.Contains(t, workflowID, workflowType)
}

func TestClient_GetWorkflowStatus(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := Config{
		HostPort: "localhost:7933",
		Domain:   "test-domain",
		TaskList: "test-task-list",
	}

	client, err := NewClient(config, logger)
	assert.NoError(t, err)

	ctx := context.Background()
	workflowID := "test-workflow"

	// Test getting workflow status
	status, err := client.GetWorkflowStatus(ctx, workflowID)
	assert.NoError(t, err)
	assert.Equal(t, "RUNNING", status)
}

func TestClient_StartWorkflow(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := Config{
		HostPort: "localhost:7933",
		Domain:   "test-domain",
		TaskList: "test-task-list",
	}

	client, err := NewClient(config, logger)
	assert.NoError(t, err)

	ctx := context.Background()
	workflowType := "test-workflow"
	input := map[string]interface{}{
		"param1": "value1",
	}

	// Test starting workflow
	execution, err := client.StartWorkflow(ctx, workflowType, input)
	assert.NoError(t, err)
	assert.NotNil(t, execution)
	assert.NotEmpty(t, execution.ID)
	assert.NotEmpty(t, execution.RunID)
}

func TestClient_GetWorkflowResult(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := Config{
		HostPort: "localhost:7933",
		Domain:   "test-domain",
		TaskList: "test-task-list",
	}

	client, err := NewClient(config, logger)
	assert.NoError(t, err)

	ctx := context.Background()
	workflowID := "test-workflow"
	runID := "test-run"

	// Test getting workflow result as map
	var result map[string]interface{}
	err = client.GetWorkflowResult(ctx, workflowID, runID, &result)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, workflowID, result["workflow_id"])
	assert.Equal(t, "completed", result["status"])
}

func TestClient_GetWorker(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := Config{
		HostPort: "localhost:7933",
		Domain:   "test-domain",
		TaskList: "test-task-list",
	}

	client, err := NewClient(config, logger)
	require.NoError(t, err)

	// Test getting worker
	worker := client.GetWorker()
	assert.Nil(t, worker) // Should be nil in stub implementation
}

func TestClient_GetClient(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := Config{
		HostPort: "localhost:7933",
		Domain:   "test-domain",
		TaskList: "test-task-list",
	}

	client, err := NewClient(config, logger)
	require.NoError(t, err)

	// Test getting client
	cadenceClient := client.GetClient()
	assert.Nil(t, cadenceClient) // Should be nil in stub implementation
}

func TestClient_HealthCheck(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := Config{
		HostPort: "localhost:7933",
		Domain:   "test-domain",
		TaskList: "test-task-list",
	}

	client, err := NewClient(config, logger)
	assert.NoError(t, err)

	// Test health check
	err = client.HealthCheck()
	assert.NoError(t, err)
}

func TestClient_GetDomain(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := Config{
		HostPort: "localhost:7933",
		Domain:   "test-domain",
		TaskList: "test-task-list",
	}

	client, err := NewClient(config, logger)
	assert.NoError(t, err)

	// Test getting domain (from Config via ProductionClient)
	domain := client.GetDomain()
	assert.Equal(t, "test-domain", domain)
}

func TestClient_IsRunning(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := Config{
		HostPort: "localhost:7933",
		Domain:   "test-domain",
		TaskList: "test-task-list",
	}

	client, err := NewClient(config, logger)
	assert.NoError(t, err)

	// Test checking if running
	running := client.IsRunning()
	assert.True(t, running)
}

func TestClient_ConcurrentAccess(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := Config{
		HostPort: "localhost:7933",
		Domain:   "test-domain",
		TaskList: "test-task-list",
	}

	client, err := NewClient(config, logger)
	assert.NoError(t, err)

	// Test concurrent access
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(workflowType string) {
			defer func() { done <- true }()

			ctx := context.Background()
			input := map[string]interface{}{
				"param": workflowType,
			}

			workflowID, err := client.ExecuteWorkflow(ctx, workflowType, input)
			if err == nil {
				client.GetWorkflowStatus(ctx, workflowID)
			}
			_ = workflowID
		}("workflow-" + string(rune(i)))
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify client is still functional
	assert.True(t, client.IsRunning())
	assert.Equal(t, "test-domain", client.GetDomain())
}

func TestClient_EdgeCases(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := Config{
		HostPort: "localhost:7933",
		Domain:   "test-domain",
		TaskList: "test-task-list",
	}

	client, err := NewClient(config, logger)
	assert.NoError(t, err)

	ctx := context.Background()

	t.Run("empty workflow type", func(t *testing.T) {
		workflowID, err := client.ExecuteWorkflow(ctx, "", map[string]interface{}{})
		assert.NoError(t, err) // Stub implementation doesn't validate
		assert.NotEmpty(t, workflowID)
	})

	t.Run("nil input", func(t *testing.T) {
		workflowID, err := client.ExecuteWorkflow(ctx, "test-workflow", nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, workflowID)
	})

	t.Run("cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		workflowID, err := client.ExecuteWorkflow(ctx, "test-workflow", map[string]interface{}{})
		assert.NoError(t, err) // Stub implementation doesn't check context
		assert.NotEmpty(t, workflowID)
	})

	t.Run("timeout context", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		time.Sleep(1 * time.Millisecond) // Ensure timeout

		workflowID, err := client.ExecuteWorkflow(ctx, "test-workflow", map[string]interface{}{})
		assert.NoError(t, err) // Stub implementation doesn't check context
		assert.NotEmpty(t, workflowID)
	})
}

func TestClient_WorkflowLifecycle(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := Config{
		HostPort: "localhost:7933",
		Domain:   "test-domain",
		TaskList: "test-task-list",
	}

	client, err := NewClient(config, logger)
	assert.NoError(t, err)

	ctx := context.Background()
	workflowType := "lifecycle-test-workflow"
	input := map[string]interface{}{
		"test_param": "test_value",
	}

	// Start workflow
	execution, err := client.StartWorkflow(ctx, workflowType, input)
	assert.NoError(t, err)
	assert.NotNil(t, execution)
	assert.NotEmpty(t, execution.ID)
	assert.NotEmpty(t, execution.RunID)

	// Execute workflow
	workflowID, err := client.ExecuteWorkflow(ctx, workflowType, input)
	assert.NoError(t, err)
	assert.NotEmpty(t, workflowID)

	// Check status
	status, err := client.GetWorkflowStatus(ctx, workflowID)
	assert.NoError(t, err)
	assert.Equal(t, "RUNNING", status)

	// Get workflow result
	var result map[string]interface{}
	err = client.GetWorkflowResult(ctx, workflowID, execution.RunID, &result)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, workflowID, result["workflow_id"])
}

func TestClient_ErrorHandling(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := Config{
		HostPort: "localhost:7933",
		Domain:   "test-domain",
		TaskList: "test-task-list",
	}

	client, err := NewClient(config, logger)
	assert.NoError(t, err)

	ctx := context.Background()

	t.Run("invalid workflow ID", func(t *testing.T) {
		status, err := client.GetWorkflowStatus(ctx, "invalid-workflow-id")
		assert.NoError(t, err)
		assert.Equal(t, "RUNNING", status) // Stub always returns RUNNING
	})

	t.Run("get result for non-existent workflow", func(t *testing.T) {
		var result map[string]interface{}
		err := client.GetWorkflowResult(ctx, "non-existent-workflow", "non-existent-run", &result)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "non-existent-workflow", result["workflow_id"])
	})

	t.Run("signal non-existent workflow", func(t *testing.T) {
		err := client.SignalWorkflow(ctx, "non-existent-workflow", "non-existent-run", "test-signal", "test-value")
		assert.NoError(t, err)
	})
}

func TestClient_Configuration(t *testing.T) {
	logger := zaptest.NewLogger(t)

	t.Run("default config", func(t *testing.T) {
		config := Config{
			HostPort: "localhost:7933",
			Domain:   "test-domain",
			TaskList: "test-task-list",
		}

		client, err := NewClient(config, logger)
		assert.NoError(t, err)
		assert.NotNil(t, client)

		domain := client.GetDomain()
		assert.Equal(t, "test-domain", domain) // from Config.Domain via ProductionClient
	})

	t.Run("custom config", func(t *testing.T) {
		config := Config{
			HostPort: "cadence.example.com:7933",
			Domain:   "production-domain",
			TaskList: "production-task-list",
		}

		client, err := NewClient(config, logger)
		assert.NoError(t, err)
		assert.NotNil(t, client)

		domain := client.GetDomain()
		assert.Equal(t, "production-domain", domain)
	})
}

func TestClient_StartStopCycle(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := Config{
		HostPort: "localhost:7933",
		Domain:   "test-domain",
		TaskList: "test-task-list",
	}

	client, err := NewClient(config, logger)
	assert.NoError(t, err)

	// Test multiple start/stop cycles
	for i := 0; i < 3; i++ {
		err = client.Start()
		assert.NoError(t, err)

		// Verify client is running after start
		assert.True(t, client.IsRunning())

		client.Stop() // Stop doesn't return error
	}
}

func TestClient_RegisterWorkflow(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := Config{
		HostPort: "localhost:7933",
		Domain:   "test-domain",
		TaskList: "test-task-list",
	}

	client, err := NewClient(config, logger)
	assert.NoError(t, err)

	// Test registering workflow
	workflow := struct {
		Name string
	}{Name: "test-workflow"}

	client.RegisterWorkflow(workflow)
	// No return value to check, just verify no panic
}

func TestClient_RegisterActivity(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := Config{
		HostPort: "localhost:7933",
		Domain:   "test-domain",
		TaskList: "test-task-list",
	}

	client, err := NewClient(config, logger)
	assert.NoError(t, err)

	// Test registering activity
	activity := struct {
		Name string
	}{Name: "test-activity"}

	client.RegisterActivity(activity)
	// No return value to check, just verify no panic
}

func TestClient_SignalWorkflow(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := Config{
		HostPort: "localhost:7933",
		Domain:   "test-domain",
		TaskList: "test-task-list",
	}

	client, err := NewClient(config, logger)
	assert.NoError(t, err)

	ctx := context.Background()
	workflowID := "test-workflow"
	runID := "test-run"
	signalName := "test-signal"
	signalValue := "test-value"

	// Test signaling workflow
	err = client.SignalWorkflow(ctx, workflowID, runID, signalName, signalValue)
	assert.NoError(t, err)
}

func TestClient_Close(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := Config{
		HostPort: "localhost:7933",
		Domain:   "test-domain",
		TaskList: "test-task-list",
	}

	client, err := NewClient(config, logger)
	assert.NoError(t, err)

	// Test closing client
	client.Close()
	// No return value to check, just verify no panic
}
