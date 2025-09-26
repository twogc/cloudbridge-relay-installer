package e2e

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// MetricsTestSuite tests Prometheus metrics collection and reporting
type MetricsTestSuite struct {
	suite.Suite
	metricsURL     string
	relayURL       string
	clusterID      string
	kubeconfigPath string
	testNamespace  string
}

func (suite *MetricsTestSuite) SetupSuite() {
	suite.metricsURL = getEnvOrDefault("METRICS_URL", "http://localhost:5551")
	suite.relayURL = getEnvOrDefault("RELAY_URL", "http://localhost:8444")
	suite.clusterID = getEnvOrDefault("CLUSTER_ID", "minikube")
	suite.testNamespace = "metrics-test"
	suite.kubeconfigPath = "/tmp/kubeconfig-metrics-test"

	// Create kubeconfig
	suite.createKubeconfig()

	// Create test namespace and resources
	suite.setupTestResources()
}

func (suite *MetricsTestSuite) TearDownSuite() {
	// Cleanup
	suite.runKubectl("delete", "namespace", suite.testNamespace, "--ignore-not-found=true")
}

func (suite *MetricsTestSuite) createKubeconfig() {
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
	require.NoError(suite.T(), err)
}

func (suite *MetricsTestSuite) setupTestResources() {
	// Create namespace
	suite.runKubectl("create", "namespace", suite.testNamespace, "--dry-run=client")
	suite.runKubectl("create", "namespace", suite.testNamespace)

	// Create test pod for exec/logs testing
	suite.runKubectl("run", "metrics-test-pod", "--image=busybox:1.35",
		"-n", suite.testNamespace, "--", "sleep", "3600")

	// Wait for pod to be ready
	time.Sleep(10 * time.Second)
}

// Test basic kube proxy metrics
func (suite *MetricsTestSuite) TestKubeProxyMetrics() {
	// Generate some requests first
	suite.generateMetrics()

	// Wait for metrics to be updated
	time.Sleep(3 * time.Second)

	metrics, err := suite.fetchMetrics()
	require.NoError(suite.T(), err, "Failed to fetch metrics")

	expectedMetrics := []string{
		"kube_proxy_requests_total",
		"kube_proxy_request_duration_seconds",
		"kube_upstream_errors_total",
	}

	for _, expectedMetric := range expectedMetrics {
		suite.Run(expectedMetric, func() {
			assert.Contains(suite.T(), metrics, expectedMetric,
				"Expected metric %s not found", expectedMetric)
		})
	}

	// Check cluster_id label is present
	clusterLabel := fmt.Sprintf(`cluster_id="%s"`, suite.clusterID)
	assert.Contains(suite.T(), metrics, clusterLabel,
		"Cluster ID label not found in metrics")

	// Check for specific verb/resource combinations
	expectedLabels := []string{
		`verb="get"`,
		`resource="pods"`,
		`code="200"`,
	}

	for _, label := range expectedLabels {
		assert.Contains(suite.T(), metrics, label,
			"Expected label %s not found in metrics", label)
	}
}

// Test WebSocket metrics
func (suite *MetricsTestSuite) TestWebSocketMetrics() {
	// Generate WebSocket traffic
	suite.runKubectl("exec", "metrics-test-pod", "-n", suite.testNamespace,
		"--", "echo", "websocket-test")

	suite.runKubectl("logs", "metrics-test-pod", "-n", suite.testNamespace)

	// Wait for metrics to update
	time.Sleep(3 * time.Second)

	metrics, err := suite.fetchMetrics()
	require.NoError(suite.T(), err)

	expectedWSMetrics := []string{
		"kube_ws_handshakes_total",
		"kube_ws_upgrades_total",
		"kube_ws_active_streams",
		"kube_ws_errors_total",
	}

	for _, expectedMetric := range expectedWSMetrics {
		suite.Run(expectedMetric, func() {
			assert.Contains(suite.T(), metrics, expectedMetric,
				"Expected WebSocket metric %s not found", expectedMetric)
		})
	}

	// Check for WebSocket-specific labels
	expectedWSLabels := []string{
		`subprotocol="v4.channel.k8s.io"`,
		`code="101"`,
	}

	for _, label := range expectedWSLabels {
		if strings.Contains(metrics, label) {
			suite.T().Logf("Found WebSocket label: %s", label)
		}
	}
}

// Test metrics accuracy and consistency
func (suite *MetricsTestSuite) TestMetricsAccuracy() {
	// Reset baseline by fetching initial metrics
	initialMetrics, err := suite.fetchMetrics()
	require.NoError(suite.T(), err)

	initialRequestCount := suite.extractMetricValue(initialMetrics,
		"kube_proxy_requests_total", map[string]string{
			"cluster_id": suite.clusterID,
			"verb":       "get",
			"resource":   "pods",
			"code":       "200",
		})

	// Make exactly 5 requests
	const numRequests = 5
	for i := 0; i < numRequests; i++ {
		suite.runKubectl("get", "pods", "-n", suite.testNamespace)
		time.Sleep(500 * time.Millisecond) // Small delay between requests
	}

	// Wait for metrics to update
	time.Sleep(3 * time.Second)

	finalMetrics, err := suite.fetchMetrics()
	require.NoError(suite.T(), err)

	finalRequestCount := suite.extractMetricValue(finalMetrics,
		"kube_proxy_requests_total", map[string]string{
			"cluster_id": suite.clusterID,
			"verb":       "get",
			"resource":   "pods",
			"code":       "200",
		})

	// Verify the counter increased by at least the number of requests made
	increase := finalRequestCount - initialRequestCount
	assert.GreaterOrEqual(suite.T(), int(increase), numRequests,
		"Request counter should have increased by at least %d, got %d",
		numRequests, increase)
}

// Test metrics performance under load
func (suite *MetricsTestSuite) TestMetricsUnderLoad() {
	// Generate load
	const numRequests = 20
	const concurrency = 5

	requestChan := make(chan int, numRequests)
	doneChan := make(chan bool, concurrency)

	// Fill request channel
	for i := 0; i < numRequests; i++ {
		requestChan <- i
	}
	close(requestChan)

	// Start concurrent workers
	for w := 0; w < concurrency; w++ {
		go func(workerID int) {
			for range requestChan {
				suite.runKubectl("get", "pods", "-n", suite.testNamespace)
			}
			doneChan <- true
		}(w)
	}

	// Wait for all workers to complete
	for w := 0; w < concurrency; w++ {
		<-doneChan
	}

	// Wait for metrics to stabilize
	time.Sleep(5 * time.Second)

	// Verify metrics endpoint is still responsive
	start := time.Now()
	metrics, err := suite.fetchMetrics()
	duration := time.Since(start)

	assert.NoError(suite.T(), err, "Metrics endpoint should remain responsive under load")
	assert.Less(suite.T(), duration, 5*time.Second, "Metrics fetch should be fast")
	assert.Contains(suite.T(), metrics, "kube_proxy_requests_total", "Metrics should contain request counts")

	// Verify reasonable metric values
	requestCount := suite.extractMetricValue(metrics, "kube_proxy_requests_total",
		map[string]string{"cluster_id": suite.clusterID})
	assert.Greater(suite.T(), requestCount, float64(0), "Should have some recorded requests")
}

// Test metrics cardinality and memory usage
func (suite *MetricsTestSuite) TestMetricsCardinality() {
	// Generate requests with different label combinations
	operations := [][]string{
		{"get", "pods"},
		{"get", "services"},
		{"get", "configmaps"},
		{"describe", "pods"},
		{"describe", "services"},
	}

	namespaces := []string{
		suite.testNamespace,
		"default",
		"kube-system",
	}

	// Generate diverse requests
	for _, op := range operations {
		for _, ns := range namespaces {
			args := append(op, "-n", ns)
			suite.runKubectl(args...)
			time.Sleep(100 * time.Millisecond)
		}
	}

	time.Sleep(3 * time.Second)

	metrics, err := suite.fetchMetrics()
	require.NoError(suite.T(), err)

	// Count unique metric series for kube_proxy_requests_total
	requestMetrics := suite.extractAllMetricSeries(metrics, "kube_proxy_requests_total")

	// Should have multiple series but not excessive cardinality
	assert.Greater(suite.T(), len(requestMetrics), 5,
		"Should have multiple request metric series")
	assert.Less(suite.T(), len(requestMetrics), 100,
		"Cardinality should be reasonable (< 100 series)")

	suite.T().Logf("Found %d unique request metric series", len(requestMetrics))
}

// Helper methods

func (suite *MetricsTestSuite) generateMetrics() {
	// Generate various types of requests to create metrics
	requests := [][]string{
		{"get", "pods", "-n", suite.testNamespace},
		{"get", "services", "-n", suite.testNamespace},
		{"get", "namespaces"},
		{"describe", "pod", "metrics-test-pod", "-n", suite.testNamespace},
	}

	for _, req := range requests {
		suite.runKubectl(req...)
		time.Sleep(200 * time.Millisecond)
	}
}

func (suite *MetricsTestSuite) fetchMetrics() (string, error) {
	resp, err := http.Get(suite.metricsURL + "/metrics")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("metrics endpoint returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (suite *MetricsTestSuite) extractMetricValue(metricsText, metricName string, labels map[string]string) float64 {
	scanner := bufio.NewScanner(strings.NewReader(metricsText))

	for scanner.Scan() {
		line := scanner.Text()

		// Skip comments and empty lines
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}

		// Check if line starts with the metric name
		if strings.HasPrefix(line, metricName) {
			// Check if all required labels are present
			allLabelsMatch := true
			for key, value := range labels {
				expectedLabel := fmt.Sprintf(`%s="%s"`, key, value)
				if !strings.Contains(line, expectedLabel) {
					allLabelsMatch = false
					break
				}
			}

			if allLabelsMatch {
				// Extract the value (last part after space)
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					if val, err := strconv.ParseFloat(parts[len(parts)-1], 64); err == nil {
						return val
					}
				}
			}
		}
	}

	return 0
}

func (suite *MetricsTestSuite) extractAllMetricSeries(metricsText, metricName string) []string {
	var series []string
	scanner := bufio.NewScanner(strings.NewReader(metricsText))

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, metricName) && !strings.HasPrefix(line, "#") {
			series = append(series, line)
		}
	}

	return series
}

func (suite *MetricsTestSuite) runKubectl(args ...string) (string, error) {
	return runKubectlWithConfig(suite.kubeconfigPath, args...)
}

func TestMetricsE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping metrics E2E tests in short mode")
	}

	suite.Run(t, new(MetricsTestSuite))
}
