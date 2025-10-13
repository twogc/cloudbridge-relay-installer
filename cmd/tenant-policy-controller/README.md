# Tenant Policy Controller

Kubernetes controller that automatically manages network isolation for multi-tenant CloudBridge relay deployments using Calico NetworkPolicies.

## Overview

This controller watches for Pods with the `cloudbridge.io/tenant` label and automatically creates Calico NetworkPolicies to enforce tenant isolation. Pods from different tenants cannot communicate with each other, ensuring secure multi-tenancy.

## Features

- **Automatic Network Policy Creation**: Creates Calico NetworkPolicies for each tenant
- **Tenant Isolation**: Blocks inter-tenant communication at the network layer
- **Leader Election**: Supports multiple replicas with automatic leader election
- **Calico Integration**: Uses Calico CRDs for advanced network policy features
- **Production Ready**: Includes metrics, health checks, and profiling support

## Requirements

### Calico CNI
This controller **requires Calico CNI** to be installed in the Kubernetes cluster:

```bash
# Install Calico (example for Kubernetes 1.28+)
kubectl apply -f https://raw.githubusercontent.com/projectcalico/calico/v3.27.0/manifests/calico.yaml
```

Verify Calico is running:
```bash
kubectl get pods -n kube-system | grep calico
```

### RBAC Permissions
The controller requires ClusterRole permissions for:
- **Pods**: `get`, `list`, `watch` (to discover tenant pods)
- **Calico GlobalNetworkPolicies**: Full CRUD (to manage global policies)
- **Calico NetworkPolicies**: Full CRUD (to manage tenant policies)
- **Leases**: Full CRUD (for leader election)

## Configuration

### Command-line Flags

```bash
tenant-policy-controller [flags]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--metrics-bind-address` | `:8080` | Metrics endpoint bind address |
| `--health-probe-bind-address` | `:8081` | Health/ready endpoint bind address |
| `--leader-elect` | `true` | Enable leader election for HA |
| `--watch-namespace` | `""` | Restrict to single namespace (empty = all) |
| `--concurrency` | `2` | Max concurrent reconciles |
| `--cache-resync` | `0` | Cache resync period (0 = disabled) |
| `--pprof-bind-address` | `""` | pprof address (empty = disabled) |
| `--zap-devel` | `false` | Development logging mode |
| `--zap-log-level` | `info` | Log level (debug, info, error) |

### Environment Variables

The controller uses in-cluster Kubernetes configuration by default. For out-of-cluster development:

```bash
export KUBECONFIG=~/.kube/config
```

## How It Works

### 1. Pod Discovery
The controller watches all Pods with the label:
```yaml
metadata:
  labels:
    cloudbridge.io/tenant: <tenant-id>
```

### 2. Policy Creation
For each unique tenant ID, the controller creates a Calico NetworkPolicy:

```yaml
apiVersion: crd.projectcalico.org/v3
kind: NetworkPolicy
metadata:
  name: tenant-isolation-<tenant-id>
  namespace: <pod-namespace>
spec:
  selector: cloudbridge.io/tenant == '<tenant-id>'
  types:
  - Ingress
  - Egress
  ingress:
  - action: Allow
    source:
      selector: cloudbridge.io/tenant == '<tenant-id>'
  egress:
  - action: Allow
    destination:
      selector: cloudbridge.io/tenant == '<tenant-id>'
  - action: Allow  # DNS
    protocol: UDP
    destination:
      ports: [53]
```

### 3. Isolation Enforcement
- ✅ Pods with same `tenant` label can communicate
- ❌ Pods with different `tenant` labels are blocked
- ✅ DNS queries (UDP port 53) are always allowed
- ✅ Policies are automatically updated when pods change

## Deployment

### Using Kubernetes Manifest

The controller is deployed using [k8s/tenant-policy-controller.yaml](../../k8s/tenant-policy-controller.yaml):

```bash
# Apply deployment
kubectl apply -f k8s/tenant-policy-controller.yaml

# Verify deployment
kubectl get deployment -n cloudbridge-system tenant-policy-controller
kubectl get pods -n cloudbridge-system -l app=tenant-policy-controller
```

### High Availability

The controller supports leader election for HA deployments:

```yaml
spec:
  replicas: 3  # Multiple replicas
```

Only one replica will be active (leader), others will be on standby.

## Monitoring

### Metrics Endpoint

Prometheus metrics are exposed at `:8080/metrics`:

```bash
# Port-forward to access metrics
kubectl port-forward -n cloudbridge-system deployment/tenant-policy-controller 8080:8080

# View metrics
curl http://localhost:8080/metrics
```

### Health Checks

```bash
# Health check
curl http://localhost:8081/healthz

# Readiness check
curl http://localhost:8081/readyz
```

### Performance Profiling

Enable pprof for performance analysis:

```bash
tenant-policy-controller --pprof-bind-address=:6060
```

Access pprof endpoints:
```bash
curl http://localhost:6060/debug/pprof/
```

## Development

### Building

```bash
# Build binary
go build -o bin/tenant-policy-controller ./cmd/tenant-policy-controller

# Build Docker image
docker build -f docker/Dockerfile.tenant-policy-controller -t tenant-policy-controller:latest .
```

### Running Locally

```bash
# Run against current kubeconfig cluster
go run ./cmd/tenant-policy-controller/main.go \
  --leader-elect=false \
  --zap-devel=true \
  --zap-log-level=debug
```

### Testing

```bash
# Run unit tests
go test ./internal/controllers/... -v

# Run with race detector
go test -race ./internal/controllers/...
```

## Troubleshooting

### Controller Not Creating Policies

1. **Check Calico Installation**:
   ```bash
   kubectl get crd | grep calico
   # Should show: networkpolicies.crd.projectcalico.org
   ```

2. **Check Pod Labels**:
   ```bash
   kubectl get pods --show-labels | grep cloudbridge.io/tenant
   ```

3. **Check Controller Logs**:
   ```bash
   kubectl logs -n cloudbridge-system deployment/tenant-policy-controller
   ```

4. **Check RBAC Permissions**:
   ```bash
   kubectl auth can-i create networkpolicies.crd.projectcalico.org --as=system:serviceaccount:cloudbridge-system:tenant-policy-controller
   ```

### Multiple Policies Created

If you see duplicate policies:
- Check that leader election is enabled (`--leader-elect=true`)
- Verify only one pod is active (check leader election logs)

### Performance Issues

For clusters with many tenants:
- Increase `--concurrency` (default: 2)
- Enable `--pprof-bind-address` for profiling
- Consider `--watch-namespace` to limit scope

## Architecture

```
┌─────────────────────────────────────────┐
│  Kubernetes API Server                  │
│  - Pods (watch)                         │
│  - Calico NetworkPolicies (CRUD)        │
└─────────────────┬───────────────────────┘
                  │
                  ↓
┌─────────────────────────────────────────┐
│  Tenant Policy Controller               │
│  ┌───────────────────────────────────┐  │
│  │  Manager (controller-runtime)     │  │
│  │  - Leader Election                │  │
│  │  - Cache & Indexing               │  │
│  │  - Event Distribution             │  │
│  └───────────────┬───────────────────┘  │
│                  ↓                       │
│  ┌───────────────────────────────────┐  │
│  │  TenantPolicyReconciler           │  │
│  │  - Watch Pods (tenant label)      │  │
│  │  - Create/Update NetworkPolicies  │  │
│  │  - Handle Deletions               │  │
│  └───────────────────────────────────┘  │
└─────────────────────────────────────────┘
```

## Related Files

- **Controller Implementation**: [internal/controllers/tenant_policy_controller.go](../../internal/controllers/tenant_policy_controller.go)
- **Controller Tests**: [internal/controllers/tenant_policy_controller_test.go](../../internal/controllers/tenant_policy_controller_test.go)
- **Kubernetes Manifest**: [k8s/tenant-policy-controller.yaml](../../k8s/tenant-policy-controller.yaml)
- **Dockerfile**: [docker/Dockerfile.tenant-policy-controller](../../docker/Dockerfile.tenant-policy-controller)
- **Analysis**: [TENANT_POLICY_CONTROLLER_ANALYSIS.md](../../TENANT_POLICY_CONTROLLER_ANALYSIS.md)

## References

- [Calico Documentation](https://docs.tigera.io/calico/latest/about)
- [Kubernetes Network Policies](https://kubernetes.io/docs/concepts/services-networking/network-policies/)
- [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime)

---

Last Updated: 2025-10-13
Status: Production-ready, requires Calico CNI
