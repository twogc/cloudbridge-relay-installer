package e2e

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// KubeProxyE2ETestSuite contains all E2E tests for kubectl proxy functionality
type KubeProxyE2ETestSuite struct {
	suite.Suite
	relayURL       string
	kubeconfigPath string
	metricsURL     string
	testNamespace  string
	clusterID      string
}

// SetupSuite runs before all tests in the suite
func (suite *KubeProxyE2ETestSuite) SetupSuite() {
	// Default values - can be overridden via environment variables
	suite.relayURL = getEnvOrDefault("RELAY_URL", "http://localhost:8444")
	suite.metricsURL = getEnvOrDefault("METRICS_URL", "http://localhost:5551")
	suite.clusterID = getEnvOrDefault("CLUSTER_ID", "minikube")
	suite.testNamespace = getEnvOrDefault("TEST_NAMESPACE", "e2e-test")
	suite.kubeconfigPath = "/tmp/kubeconfig-relay-test"

	// Create kubeconfig for relay
	suite.createTestKubeconfig()

	// Set up test environment
	suite.setupTestEnvironment()

	// Wait for relay to be ready
	suite.waitForRelayReady()
}

// TearDownSuite runs after all tests in the suite
func (suite *KubeProxyE2ETestSuite) TearDownSuite() {
	// Cleanup test resources
	suite.cleanupTestEnvironment()
}

func (suite *KubeProxyE2ETestSuite) createTestKubeconfig() {
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
    user: test-user
current-context: %s-via-relay
users:
- name: test-user
  user: {}
`, suite.clusterID, suite.relayURL, suite.clusterID,
		suite.clusterID, suite.clusterID, suite.clusterID)

	err := writeFile(suite.kubeconfigPath, kubeconfigContent)
	require.NoError(suite.T(), err, "Failed to create test kubeconfig")
}

func (suite *KubeProxyE2ETestSuite) setupTestEnvironment() {
	// Create test namespace
	_, err := suite.runKubectl("create", "namespace", suite.testNamespace, "--dry-run=client")
	if err == nil {
		suite.runKubectl("create", "namespace", suite.testNamespace)
	}

	// Create test deployment
	suite.runKubectl("create", "deployment", "nginx-test",
		"--image=nginx:1.20", "-n", suite.testNamespace)

	// Create test service
	suite.runKubectl("expose", "deployment", "nginx-test",
		"--port=80", "--target-port=80", "-n", suite.testNamespace)

	// Create busybox for exec tests
	suite.runKubectl("create", "deployment", "busybox-test",
		"--image=busybox:1.35", "-n", suite.testNamespace,
		"--", "sleep", "3600")

	// Wait for pods to be ready
	suite.waitForDeploymentReady("nginx-test")
	suite.waitForDeploymentReady("busybox-test")
}

func (suite *KubeProxyE2ETestSuite) cleanupTestEnvironment() {
	// Delete test namespace (cascades to all resources)
	suite.runKubectl("delete", "namespace", suite.testNamespace, "--ignore-not-found=true")
}

func (suite *KubeProxyE2ETestSuite) waitForRelayReady() {
	timeout := time.After(60 * time.Second)
	tick := time.Tick(2 * time.Second)

	for {
		select {
		case <-timeout:
			suite.T().Fatal("Timeout waiting for relay to be ready")
		case <-tick:
			resp, err := http.Get(suite.relayURL + "/health")
			if err == nil && resp.StatusCode == 200 {
				resp.Body.Close()
				return
			}
			if resp != nil {
				resp.Body.Close()
			}
		}
	}
}

func (suite *KubeProxyE2ETestSuite) waitForDeploymentReady(deploymentName string) {
	timeout := time.After(120 * time.Second)
	tick := time.Tick(5 * time.Second)

	for {
		select {
		case <-timeout:
			suite.T().Fatalf("Timeout waiting for deployment %s to be ready", deploymentName)
		case <-tick:
			output, err := suite.runKubectl("get", "deployment", deploymentName,
				"-n", suite.testNamespace, "-o", "jsonpath={.status.readyReplicas}")
			if err == nil && strings.TrimSpace(output) == "1" {
				return
			}
		}
	}
}

// Test basic kubectl commands through relay
func (suite *KubeProxyE2ETestSuite) TestBasicKubectlCommands() {
	testCases := []struct {
		name     string
		args     []string
		contains string
	}{
		{
			name:     "get pods",
			args:     []string{"get", "pods", "-n", suite.testNamespace},
			contains: "nginx-test",
		},
		{
			name:     "get services",
			args:     []string{"get", "services", "-n", suite.testNamespace},
			contains: "nginx-test",
		},
		{
			name:     "get namespaces",
			args:     []string{"get", "namespaces"},
			contains: suite.testNamespace,
		},
		{
			name:     "describe deployment",
			args:     []string{"describe", "deployment", "nginx-test", "-n", suite.testNamespace},
			contains: "nginx:1.20",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			output, err := suite.runKubectl(tc.args...)
			assert.NoError(suite.T(), err, "kubectl command failed")
			assert.Contains(suite.T(), output, tc.contains, "Expected content not found in output")
		})
	}
}

// Test WebSocket streams functionality
func (suite *KubeProxyE2ETestSuite) TestWebSocketStreams() {
	suite.Run("logs", func() {
		// Test pod logs
		output, err := suite.runKubectl("logs", "deployment/nginx-test", "-n", suite.testNamespace)
		assert.NoError(suite.T(), err, "kubectl logs failed")
		suite.T().Logf("Nginx logs output: %s", output)
		// Nginx logs should contain some output or be empty (but not error)
	})

	suite.Run("exec", func() {
		// Test exec command
		output, err := suite.runKubectl("exec", "deployment/busybox-test",
			"-n", suite.testNamespace, "--", "echo", "hello-from-relay")
		assert.NoError(suite.T(), err, "kubectl exec failed")
		assert.Contains(suite.T(), output, "hello-from-relay", "Exec output not correct")
	})

	suite.Run("exec-whoami", func() {
		// Test exec whoami
		output, err := suite.runKubectl("exec", "deployment/busybox-test",
			"-n", suite.testNamespace, "--", "whoami")
		assert.NoError(suite.T(), err, "kubectl exec whoami failed")
		assert.Contains(suite.T(), output, "root", "whoami output not correct")
	})
}

// Test authorization policies
func (suite *KubeProxyE2ETestSuite) TestAuthorizationPolicies() {
	suite.Run("allowed-namespace", func() {
		// Test access to allowed namespace (test namespace)
		_, err := suite.runKubectl("get", "pods", "-n", suite.testNamespace)
		assert.NoError(suite.T(), err, "Access to allowed namespace should succeed")
	})

	suite.Run("system-namespace", func() {
		// Test access to kube-system (should work with minikube-admin policy)
		_, err := suite.runKubectl("get", "pods", "-n", "kube-system")
		// This might succeed or fail depending on the policy configuration
		// In minikube-example.json we allow all namespaces
		suite.T().Logf("Access to kube-system: %v", err)
	})

	suite.Run("create-resource", func() {
		// Test creating a resource (should be allowed with minikube-admin policy)
		output, err := suite.runKubectl("create", "configmap", "test-cm",
			"--from-literal=key=value", "-n", suite.testNamespace, "--dry-run=client")
		assert.NoError(suite.T(), err, "Create resource should be allowed")
		assert.Contains(suite.T(), output, "configmap/test-cm", "ConfigMap creation output incorrect")
	})
}

// Test metrics collection
func (suite *KubeProxyE2ETestSuite) TestMetricsCollection() {
	// Make some requests first to generate metrics
	suite.runKubectl("get", "pods", "-n", suite.testNamespace)
	suite.runKubectl("get", "services", "-n", suite.testNamespace)

	// Wait a moment for metrics to be updated
	time.Sleep(2 * time.Second)

	resp, err := http.Get(suite.metricsURL + "/metrics")
	require.NoError(suite.T(), err, "Failed to fetch metrics")
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(suite.T(), err, "Failed to read metrics response")

	metricsText := string(body)

	// Check for kube proxy metrics
	expectedMetrics := []string{
		"kube_proxy_requests_total",
		"kube_proxy_request_duration_seconds",
		"kube_ws_handshakes_total",
		"kube_ws_active_streams",
	}

	for _, metric := range expectedMetrics {
		assert.Contains(suite.T(), metricsText, metric,
			"Expected metric %s not found", metric)
	}

	// Check for cluster_id label
	assert.Contains(suite.T(), metricsText, fmt.Sprintf(`cluster_id="%s"`, suite.clusterID),
		"Cluster ID label not found in metrics")
}

// Test audit logging functionality
func (suite *KubeProxyE2ETestSuite) TestAuditLogging() {
	// This test would typically check log output
	// For now, we just verify that requests are processed
	// In a real test, you'd capture and verify audit log entries

	suite.Run("audit-entries", func() {
		// Make some requests that should generate audit entries
		suite.runKubectl("get", "pods", "-n", suite.testNamespace)
		suite.runKubectl("get", "services", "-n", suite.testNamespace)

		// Note: In a real implementation, you would:
		// 1. Configure log capturing
		// 2. Parse log entries
		// 3. Verify audit fields (tenant_id, user_id, verb, resource, allowed, etc.)

		suite.T().Log("Audit logging test completed - manual verification required")
	})
}

// Test error handling
func (suite *KubeProxyE2ETestSuite) TestErrorHandling() {
	suite.Run("invalid-cluster", func() {
		// Test with non-existent cluster ID
		invalidConfig := fmt.Sprintf(`
apiVersion: v1
kind: Config
clusters:
- name: invalid-cluster
  cluster:
    server: %s/kube/invalid-cluster
    insecure-skip-tls-verify: true
contexts:
- name: invalid-cluster
  context:
    cluster: invalid-cluster
    user: test-user
current-context: invalid-cluster
users:
- name: test-user
  user: {}
`, suite.relayURL)

		invalidKubeconfigPath := "/tmp/kubeconfig-invalid"
		err := writeFile(invalidKubeconfigPath, invalidConfig)
		require.NoError(suite.T(), err)

		// This should fail with cluster not found
		output, err := runKubectlWithConfig(invalidKubeconfigPath, "get", "pods")
		assert.Error(suite.T(), err, "Request to invalid cluster should fail")
		assert.Contains(suite.T(), output, "cluster not found",
			"Error message should indicate cluster not found")
	})

	suite.Run("invalid-resource", func() {
		// Test with non-existent resource
		output, err := suite.runKubectl("get", "invalid-resource-type")
		assert.Error(suite.T(), err, "Request for invalid resource should fail")
		suite.T().Logf("Invalid resource error: %s", output)
	})
}

// Test performance and load
func (suite *KubeProxyE2ETestSuite) TestPerformance() {
	suite.Run("concurrent-requests", func() {
		// Test multiple concurrent requests
		const numRequests = 10
		results := make(chan error, numRequests)

		for i := 0; i < numRequests; i++ {
			go func(id int) {
				_, err := suite.runKubectl("get", "pods", "-n", suite.testNamespace,
					"--request-timeout=30s")
				results <- err
			}(i)
		}

		// Collect results
		successCount := 0
		for i := 0; i < numRequests; i++ {
			err := <-results
			if err == nil {
				successCount++
			} else {
				suite.T().Logf("Request %d failed: %v", i, err)
			}
		}

		// Expect at least 80% success rate
		successRate := float64(successCount) / float64(numRequests)
		assert.GreaterOrEqual(suite.T(), successRate, 0.8,
			"Success rate too low: %d/%d (%.1f%%)", successCount, numRequests, successRate*100)
	})
}

// Helper functions

func (suite *KubeProxyE2ETestSuite) runKubectl(args ...string) (string, error) {
	return runKubectlWithConfig(suite.kubeconfigPath, args...)
}

func runKubectlWithConfig(kubeconfigPath string, args ...string) (string, error) {
	// Set WebSocket environment variables
	cmd := exec.Command("kubectl", args...)
	cmd.Env = append(cmd.Env,
		"KUBECTL_REMOTE_COMMAND_WEBSOCKETS=1",
		"KUBECTL_PORT_FORWARD_WEBSOCKETS=1",
		"KUBECONFIG="+kubeconfigPath,
	)

	output, err := cmd.CombinedOutput()
	return string(output), err
}

func writeFile(path, content string) error {
	return exec.Command("sh", "-c", fmt.Sprintf("cat > %s << 'EOF'\n%s\nEOF", path, content)).Run()
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := exec.Command("sh", "-c", "echo $"+key).String(); value != "" {
		return strings.TrimSpace(value)
	}
	return defaultValue
}

// Test runner function
func TestKubeProxyE2E(t *testing.T) {
	// Skip if not running E2E tests
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	suite.Run(t, new(KubeProxyE2ETestSuite))
}
