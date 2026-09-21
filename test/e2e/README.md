# E2E Tests for CloudBridge Relay Kubectl Proxy

Comprehensive end-to-end tests for CloudBridge Relay kubectl proxy functionality.

## Overview

These tests verify the complete kubectl proxy integration including:

- **Basic kubectl commands** (get, list, describe, watch)
- **WebSocket streams** (exec, attach, logs, port-forward)
- **Authorization policies** (tenant/user/group-based access control)
- **Prometheus metrics** (request counters, duration, WebSocket metrics)
- **Integration scenarios** (real workloads, concurrent operations)

## Prerequisites

1. **CloudBridge Relay running** with kubectl proxy enabled
2. **kubectl >= 1.29** (for WebSocket support)
3. **Go >= 1.21** (for test execution)
4. **Accessible Kubernetes cluster** (minikube, kind, or real cluster)

## Quick Start

```bash
# 1. Start CloudBridge Relay with minikube config
./cloudbridge-relay --config=configs/minikube-example.json

# 2. Run all E2E tests
./test/e2e/run_tests.sh

# 3. Or run specific test suites
cd test/e2e
go test -v -run TestKubeProxyE2E # Basic kubectl functionality
go test -v -run TestAuthorizationE2E # Authorization policies
go test -v -run TestMetricsE2E # Prometheus metrics
go test -v -run TestIntegration # Full integration tests
```

## Test Configuration

Set environment variables to customize test behavior:

```bash
export RELAY_URL="http://localhost:8444" # CloudBridge Relay URL
export METRICS_URL="http://localhost:5551" # Prometheus metrics endpoint
export CLUSTER_ID="minikube" # Cluster ID in relay config
export TEST_NAMESPACE="e2e-test" # Kubernetes namespace for tests
export KUBECTL_REMOTE_COMMAND_WEBSOCKETS=1 # Enable WebSocket for exec/attach
export KUBECTL_PORT_FORWARD_WEBSOCKETS=1 # Enable WebSocket for port-forward
```

## Test Suites

### 1. Basic Kubectl Commands (`kube_proxy_test.go`)

Tests fundamental kubectl operations through the relay:

- **Resource access**: get, list, describe pods/services/deployments
- **Cross-namespace operations**: accessing multiple namespaces
- **Error handling**: invalid resources, cluster not found
- **Performance**: concurrent requests, response times

```bash
go test -v -run TestKubeProxyE2E/TestBasicKubectlCommands
```

### 2. WebSocket Streams (`kube_proxy_test.go`)

Tests real-time kubectl operations that require WebSocket upgrades:

- **Exec**: `kubectl exec -it pod -- /bin/bash`
- **Logs**: `kubectl logs -f deployment/app`
- **Port-forward**: `kubectl port-forward svc/app 8080:80`
- **Attach**: `kubectl attach pod`

```bash
go test -v -run TestKubeProxyE2E/TestWebSocketStreams
```

### 3. Authorization Policies (`auth_test.go`)

Tests fine-grained access control based on tenant/user/group:

- **Tenant-based access**: Different policies per tenant
- **Resource-based access**: Allow/deny specific resources
- **Namespace isolation**: Restrict access by namespace
- **Verb-based control**: Read-only vs full access
- **WebSocket authorization**: Exec/portforward permissions

```bash
go test -v -run TestAuthorizationE2E
```

### 4. Prometheus Metrics (`metrics_test.go`)

Tests metrics collection and accuracy:

- **Request metrics**: `kube_proxy_requests_total`, `kube_proxy_request_duration_seconds`
- **WebSocket metrics**: `kube_ws_handshakes_total`, `kube_ws_active_streams`
- **Error metrics**: `kube_upstream_errors_total`
- **Label accuracy**: cluster_id, verb, resource, code labels
- **Cardinality testing**: Reasonable number of metric series

```bash
go test -v -run TestMetricsE2E
```

### 5. Integration Tests (`integration_test.go`)

Tests complete application scenarios:

- **Application lifecycle**: Deploy, scale, configure applications
- **Real workloads**: Nginx web server, busybox backend
- **Concurrent operations**: Multiple users, simultaneous requests
- **Full pipeline**: Authorization → Proxy → WebSocket → Metrics

```bash
go test -v -run TestIntegration
```

## Test Structure

```
test/e2e/
├── run_tests.sh # Main test runner script
├── kube_proxy_test.go # Basic kubectl + WebSocket tests
├── auth_test.go # Authorization policy tests
├── metrics_test.go # Prometheus metrics tests
├── integration_test.go # Full integration scenarios
├── go.mod # Go dependencies
└── README.md # This file
```

Each test file contains:
- **Suite setup/teardown**: Test environment preparation
- **Test cases**: Individual test scenarios with assertions
- **Helper methods**: Common utilities (kubectl execution, kubeconfig creation)

## Example Test Scenarios

### Successful kubectl exec through relay:
```bash
kubectl --kubeconfig /tmp/kubeconfig-relay exec deployment/nginx -- echo "hello"
# → Relay receives request
# → Checks authorization (tenant_id: default, verb: exec, resource: pods)
# → Upgrades to WebSocket with v4.channel.k8s.io subprotocol
# → Proxies to Kubernetes API server
# → Returns "hello" output
# → Records metrics: kube_proxy_requests_total{verb="exec"} +1
```

### Authorization denial:
```bash
kubectl --kubeconfig /tmp/kubeconfig-relay get pods -n kube-system
# → Relay checks policy for tenant "readonly-viewer"
# → Finds deny rule for kube-system namespace
# → Returns HTTP 403 Forbidden
# → Records audit: {"allowed": false, "reason": "denied by explicit deny rule"}
```

### Metrics verification:
```bash
curl http://localhost:5551/metrics | grep kube_proxy_requests_total
# kube_proxy_requests_total{cluster_id="minikube",verb="get",resource="pods",code="200"} 15
# kube_proxy_requests_total{cluster_id="minikube",verb="exec",resource="pods",code="200"} 3
```

## Running Tests in CI/CD

The tests are designed to run in automated environments:

```yaml
# .github/workflows/e2e-tests.yml
name: E2E Tests
on: [push, pull_request]
jobs:
  e2e:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: '1.21'

      # Setup minikube + CloudBridge Relay
      - name: Start minikube
        run: |
          minikube start
          eval $(minikube docker-env)
          docker build -t cloudbridge-relay:local .

      - name: Deploy relay
        run: |
          kubectl apply -f k8s-minikube-deployment.yaml
          kubectl wait --for=condition=ready pod -l app=cloudbridge-relay --timeout=300s

      # Run E2E tests
      - name: Run E2E Tests
        run: ./test/e2e/run_tests.sh
        env:
          RELAY_URL: "http://$(minikube ip):30444"
          METRICS_URL: "http://$(minikube ip):30551"
```

## Troubleshooting

### Common Issues

**Test failures with "connection refused":**
```bash
# Check if relay is running and accessible
curl http://localhost:8444/health
kubectl logs deployment/cloudbridge-relay
```

**WebSocket tests failing:**
```bash
# Ensure kubectl >= 1.29 and WebSocket env vars set
kubectl version --client
export KUBECTL_REMOTE_COMMAND_WEBSOCKETS=1
export KUBECTL_PORT_FORWARD_WEBSOCKETS=1
```

**Authorization tests not working:**
```bash
# Check relay logs for authorization decisions
kubectl logs deployment/cloudbridge-relay | grep -i "kube-proxy-audit"
```

**Metrics tests failing:**
```bash
# Verify metrics endpoint is accessible
curl http://localhost:5551/metrics | grep kube_proxy
```

### Test Logs

Test logs are saved to `/tmp/test_*.log` for debugging:

```bash
# View specific test failures
cat /tmp/test_basic_kubectl_commands.log
cat /tmp/test_authorization.log
cat /tmp/test_metrics.log
```

### Manual Verification

You can also run kubectl commands manually to verify functionality:

```bash
# Create test kubeconfig
export KUBECONFIG=~/.kube/config-relay

# Test basic commands
kubectl get pods --all-namespaces
kubectl logs deployment/nginx-test -n e2e-test

# Test WebSocket streams
kubectl exec -it deployment/busybox-test -n e2e-test -- /bin/sh
kubectl port-forward svc/nginx-test 8080:80 -n e2e-test
```

---

## Test Coverage

The E2E test suite provides comprehensive coverage of CloudBridge Relay kubectl proxy functionality:

- **Protocol compatibility**: HTTP/1.1, HTTP/2, WebSocket upgrades
- **Security**: mTLS to API servers, authorization policies, audit logging
- **Performance**: Concurrent requests, low latency, proper resource usage
- **Reliability**: Error handling, recovery, timeout management
- **Observability**: Complete metrics and audit trail

Run these tests regularly to ensure kubectl proxy functionality remains robust and reliable!
