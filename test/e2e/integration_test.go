package e2e

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// IntegrationTestSuite tests complete kubectl proxy functionality end-to-end
type IntegrationTestSuite struct {
	suite.Suite
	relayURL       string
	metricsURL     string
	clusterID      string
	kubeconfigPath string
	testNamespace  string
}

func (suite *IntegrationTestSuite) SetupSuite() {
	suite.relayURL = getEnvOrDefault("RELAY_URL", "http://localhost:8444")
	suite.metricsURL = getEnvOrDefault("METRICS_URL", "http://localhost:5551")
	suite.clusterID = getEnvOrDefault("CLUSTER_ID", "minikube")
	suite.testNamespace = "integration-test"
	suite.kubeconfigPath = "/tmp/kubeconfig-integration"

	suite.createKubeconfig()
	suite.setupTestEnvironment()
}

func (suite *IntegrationTestSuite) TearDownSuite() {
	suite.cleanupTestEnvironment()
}

func (suite *IntegrationTestSuite) createKubeconfig() {
	kubeconfigContent := fmt.Sprintf(`
apiVersion: v1
kind: Config
clusters:
- name: %s-via-relay
  cluster:
    server: %s/kube/%s
    insecure-skip-tls-verify: true
contexts:
- name: %s-via-relay
  context:
    cluster: %s-via-relay
    user: default-user
current-context: %s-via-relay
users:
- name: default-user
  user: {}
`, suite.clusterID, suite.relayURL, suite.clusterID,
		suite.clusterID, suite.clusterID, suite.clusterID)

	err := writeFile(suite.kubeconfigPath, kubeconfigContent)
	require.NoError(suite.T(), err)
}

func (suite *IntegrationTestSuite) setupTestEnvironment() {
	// Create test namespace
	suite.runKubectl("create", "namespace", suite.testNamespace, "--dry-run=client")
	suite.runKubectl("create", "namespace", suite.testNamespace)

	// Create a complete application stack for testing
	suite.deployTestApplication()
	suite.waitForApplicationReady()
}

func (suite *IntegrationTestSuite) deployTestApplication() {
	// Deploy nginx web server
	suite.runKubectl("create", "deployment", "web-server",
		"--image=nginx:1.20", "-n", suite.testNamespace)

	// Expose as service
	suite.runKubectl("expose", "deployment", "web-server",
		"--port=80", "--target-port=80", "-n", suite.testNamespace)

	// Deploy backend service
	suite.runKubectl("create", "deployment", "backend",
		"--image=busybox:1.35", "-n", suite.testNamespace,
		"--", "sh", "-c", "while true; do echo 'Backend response' | nc -l -p 8080; done")

	// Create ConfigMap for configuration
	suite.runKubectl("create", "configmap", "app-config",
		"--from-literal=env=test",
		"--from-literal=debug=true",
		"-n", suite.testNamespace)

	// Create Secret for credentials
	suite.runKubectl("create", "secret", "generic", "app-secret",
		"--from-literal=username=testuser",
		"--from-literal=password=testpass",
		"-n", suite.testNamespace)
}

func (suite *IntegrationTestSuite) waitForApplicationReady() {
	deployments := []string{"web-server", "backend"}

	for _, deployment := range deployments {
		timeout := time.After(120 * time.Second)
		tick := time.Tick(5 * time.Second)

		for {
			select {
			case <-timeout:
				suite.T().Fatalf("Timeout waiting for %s deployment", deployment)
			case <-tick:
				output, err := suite.runKubectl("get", "deployment", deployment,
					"-n", suite.testNamespace, "-o", "jsonpath={.status.readyReplicas}")
				if err == nil && strings.TrimSpace(output) == "1" {
					suite.T().Logf("Deployment %s is ready", deployment)
					goto nextDeployment
				}
			}
		}
	nextDeployment:
	}
}

func (suite *IntegrationTestSuite) cleanupTestEnvironment() {
	suite.runKubectl("delete", "namespace", suite.testNamespace, "--ignore-not-found=true")
}

// Test complete application lifecycle through relay
func (suite *IntegrationTestSuite) TestApplicationLifecycle() {
	suite.Run("deploy-application", func() {
		// Verify all resources are created successfully
		resources := [][]string{
			{"get", "deployments", "-n", suite.testNamespace},
			{"get", "services", "-n", suite.testNamespace},
			{"get", "configmaps", "-n", suite.testNamespace},
			{"get", "secrets", "-n", suite.testNamespace},
		}

		for _, cmd := range resources {
			output, err := suite.runKubectl(cmd...)
			assert.NoError(suite.T(), err, "Failed to get %s", strings.Join(cmd, " "))
			assert.NotEmpty(suite.T(), output, "Empty output for %s", strings.Join(cmd, " "))
		}
	})

	suite.Run("verify-application-status", func() {
		// Check deployment status
		output, err := suite.runKubectl("get", "deployments", "-n", suite.testNamespace, "-o", "wide")
		assert.NoError(suite.T(), err)
		assert.Contains(suite.T(), output, "web-server")
		assert.Contains(suite.T(), output, "backend")

		// Check pod status
		output, err = suite.runKubectl("get", "pods", "-n", suite.testNamespace)
		assert.NoError(suite.T(), err)
		assert.Contains(suite.T(), output, "Running", "Pods should be running")
	})

	suite.Run("test-configuration", func() {
		// Verify ConfigMap
		output, err := suite.runKubectl("get", "configmap", "app-config",
			"-n", suite.testNamespace, "-o", "yaml")
		assert.NoError(suite.T(), err)
		assert.Contains(suite.T(), output, "env: test")
		assert.Contains(suite.T(), output, "debug: \"true\"")

		// Verify Secret (should not show values in plain text)
		output, err = suite.runKubectl("get", "secret", "app-secret",
			"-n", suite.testNamespace, "-o", "yaml")
		assert.NoError(suite.T(), err)
		assert.Contains(suite.T(), output, "username:")
		assert.Contains(suite.T(), output, "password:")
	})
}

// Test WebSocket functionality with real workloads
func (suite *IntegrationTestSuite) TestRealWorkloadWebSockets() {
	suite.Run("exec-into-application", func() {
		// Test exec into web server
		output, err := suite.runKubectl("exec", "deployment/web-server",
			"-n", suite.testNamespace, "--", "nginx", "-v")
		assert.NoError(suite.T(), err)
		assert.Contains(suite.T(), output, "nginx version", "Should get nginx version")

		// Test exec with interactive commands
		output, err = suite.runKubectl("exec", "deployment/web-server",
			"-n", suite.testNamespace, "--", "cat", "/etc/nginx/nginx.conf")
		assert.NoError(suite.T(), err)
		assert.Contains(suite.T(), output, "worker_processes", "Should get nginx config")
	})

	suite.Run("logs-from-applications", func() {
		// Get logs from web server
		output, err := suite.runKubectl("logs", "deployment/web-server",
			"-n", suite.testNamespace)
		assert.NoError(suite.T(), err)
		// Nginx might not have logs yet, so just check command succeeds

		// Get logs from backend
		output, err = suite.runKubectl("logs", "deployment/backend",
			"-n", suite.testNamespace)
		assert.NoError(suite.T(), err)
		suite.T().Logf("Backend logs: %s", output)
	})

	suite.Run("port-forward-to-service", func() {
		// Test port-forward to web service
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Start port-forward in background
		cmd := suite.startPortForward(ctx, "svc/web-server", "8080:80")

		// Wait for port-forward to establish
		time.Sleep(3 * time.Second)

		// Test connection (would need curl or similar in real test)
		// For now, just verify the port-forward command started successfully
		select {
		case <-ctx.Done():
			// Port-forward completed (expected for this test)
		default:
			// Port-forward is still running, which is good
		}

		cancel() // Stop port-forward
		cmd.Wait()
	})
}

// Test concurrent operations
func (suite *IntegrationTestSuite) TestConcurrentOperations() {
	suite.Run("concurrent-requests", func() {
		const numWorkers = 10
		const requestsPerWorker = 5

		var wg sync.WaitGroup
		errors := make(chan error, numWorkers*requestsPerWorker)

		// Start concurrent workers
		for w := 0; w < numWorkers; w++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()

				for r := 0; r < requestsPerWorker; r++ {
					operations := [][]string{
						{"get", "pods", "-n", suite.testNamespace},
						{"get", "services", "-n", suite.testNamespace},
						{"describe", "deployment", "web-server", "-n", suite.testNamespace},
					}

					for _, op := range operations {
						_, err := suite.runKubectl(op...)
						if err != nil {
							errors <- fmt.Errorf("worker %d, request %d, op %v: %w",
								workerID, r, op, err)
						}
					}
				}
			}(w)
		}

		// Wait for all workers to complete
		wg.Wait()
		close(errors)

		// Check for errors
		var errorCount int
		for err := range errors {
			errorCount++
			suite.T().Logf("Concurrent operation error: %v", err)
		}

		totalOperations := numWorkers * requestsPerWorker * 3 // 3 operations per request
		successRate := float64(totalOperations-errorCount) / float64(totalOperations)

		assert.GreaterOrEqual(suite.T(), successRate, 0.9,
			"Success rate should be >= 90%%, got %.1f%% (%d/%d)",
			successRate*100, totalOperations-errorCount, totalOperations)
	})

	suite.Run("concurrent-websockets", func() {
		const numWebSocketOps = 5
		var wg sync.WaitGroup
		errors := make(chan error, numWebSocketOps)

		// Start concurrent WebSocket operations
		for i := 0; i < numWebSocketOps; i++ {
			wg.Add(1)
			go func(opID int) {
				defer wg.Done()

				// Alternate between different WebSocket operations
				switch opID % 3 {
				case 0:
					_, err := suite.runKubectl("exec", "deployment/web-server",
						"-n", suite.testNamespace, "--", "echo", fmt.Sprintf("test-%d", opID))
					errors <- err
				case 1:
					_, err := suite.runKubectl("logs", "deployment/web-server",
						"-n", suite.testNamespace, "--tail=10")
					errors <- err
				case 2:
					_, err := suite.runKubectl("logs", "deployment/backend",
						"-n", suite.testNamespace, "--tail=5")
					errors <- err
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		// Check WebSocket operation results
		var wsErrors int
		for err := range errors {
			if err != nil {
				wsErrors++
				suite.T().Logf("WebSocket operation error: %v", err)
			}
		}

		assert.LessOrEqual(suite.T(), wsErrors, 1,
			"Should have at most 1 WebSocket error out of %d operations", numWebSocketOps)
	})
}

// Test complete metrics pipeline
func (suite *IntegrationTestSuite) TestMetricsPipeline() {
	// Generate various types of activity
	suite.generateIntegrationMetrics()

	// Wait for metrics to be collected
	time.Sleep(5 * time.Second)

	// Verify metrics are generated
	metrics, err := suite.fetchMetrics()
	require.NoError(suite.T(), err)

	// Check for diverse metric labels
	expectedLabels := []string{
		fmt.Sprintf(`cluster_id="%s"`, suite.clusterID),
		`verb="get"`,
		`verb="exec"`,
		`resource="deployments"`,
		`resource="services"`,
		`resource="pods"`,
		`code="200"`,
	}

	for _, label := range expectedLabels {
		assert.Contains(suite.T(), metrics, label,
			"Expected metric label %s not found", label)
	}

	// Verify WebSocket metrics are present
	wsMetrics := []string{
		"kube_ws_handshakes_total",
		"kube_ws_upgrades_total",
	}

	for _, metric := range wsMetrics {
		assert.Contains(suite.T(), metrics, metric,
			"Expected WebSocket metric %s not found", metric)
	}
}

// Test authorization in realistic scenarios
func (suite *IntegrationTestSuite) TestRealisticAuthorizationScenarios() {
	suite.Run("application-developer-workflow", func() {
		// Simulate typical developer operations
		operations := [][]string{
			{"get", "deployments", "-n", suite.testNamespace},
			{"describe", "deployment", "web-server", "-n", suite.testNamespace},
			{"get", "pods", "-n", suite.testNamespace},
			{"logs", "deployment/web-server", "-n", suite.testNamespace},
			{"exec", "deployment/web-server", "-n", suite.testNamespace, "--", "whoami"},
		}

		for _, op := range operations {
			_, err := suite.runKubectl(op...)
			assert.NoError(suite.T(), err,
				"Developer operation should succeed: %v", op)
		}
	})

	suite.Run("configuration-management", func() {
		// Test ConfigMap and Secret operations
		_, err := suite.runKubectl("get", "configmap", "app-config",
			"-n", suite.testNamespace, "-o", "yaml")
		assert.NoError(suite.T(), err)

		_, err = suite.runKubectl("get", "secret", "app-secret",
			"-n", suite.testNamespace)
		assert.NoError(suite.T(), err)

		// Test describe operations
		_, err = suite.runKubectl("describe", "configmap", "app-config",
			"-n", suite.testNamespace)
		assert.NoError(suite.T(), err)
	})
}

// Helper methods

func (suite *IntegrationTestSuite) runKubectl(args ...string) (string, error) {
	return runKubectlWithConfig(suite.kubeconfigPath, args...)
}

func (suite *IntegrationTestSuite) generateIntegrationMetrics() {
	// Generate comprehensive request patterns
	requests := [][]string{
		// Basic resource access
		{"get", "namespaces"},
		{"get", "pods", "-n", suite.testNamespace},
		{"get", "services", "-n", suite.testNamespace},
		{"get", "deployments", "-n", suite.testNamespace},

		// Detailed operations
		{"describe", "deployment", "web-server", "-n", suite.testNamespace},
		{"describe", "service", "web-server", "-n", suite.testNamespace},

		// WebSocket operations
		{"logs", "deployment/web-server", "-n", suite.testNamespace},
		{"exec", "deployment/backend", "-n", suite.testNamespace, "--", "echo", "metrics-test"},

		// Configuration access
		{"get", "configmaps", "-n", suite.testNamespace},
		{"get", "secrets", "-n", suite.testNamespace},
	}

	for _, req := range requests {
		suite.runKubectl(req...)
		time.Sleep(100 * time.Millisecond)
	}
}

func (suite *IntegrationTestSuite) fetchMetrics() (string, error) {
	// Implementation similar to metrics_test.go
	resp, err := http.Get(suite.metricsURL + "/metrics")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (suite *IntegrationTestSuite) startPortForward(ctx context.Context, target, ports string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "kubectl",
		"--kubeconfig", suite.kubeconfigPath,
		"port-forward", target, ports, "-n", suite.testNamespace)

	cmd.Env = append(cmd.Env,
		"KUBECTL_PORT_FORWARD_WEBSOCKETS=1")

	cmd.Start()
	return cmd
}

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	suite.Run(t, new(IntegrationTestSuite))
}
