# Tenant Policy Controller - Duplication Analysis

## Summary
There are **TWO** versions of tenant-policy-controller:
1. **cmd/tenant-policy-controller** (119 lines) - Full-featured version with Calico integration
2. **cmd/tenant-policy-controller-simple** (78 lines) - Simplified version without Calico

Both use the same underlying controller: `internal/controllers/TenantPolicyReconciler`

## Comparison

### 1. cmd/tenant-policy-controller/main.go (119 lines)

**Features**:
- Calico CRD integration (`calicov1.AddToScheme`)
- dvanced configuration options:
  - `--watch-namespace` - Restrict to single namespace
  - `--concurrency` - Max concurrent reconciles (default: 2)
  - `--cache-resync` - Cache resync period
  - `--pprof-bind-address` - Performance profiling
- eader election enabled by default (`--leader-elect` defaults to true)
- roper health checks using `healthz.Ping`
- roduction-ready configuration

**Dependencies**:
```go
import (
    calicov1 "github.com/projectcalico/api/pkg/apis/projectcalico/v3"
    controllers "github.com/twogc/cloudbridge-relay-installer/internal/controllers"
    // ... kubernetes libraries
)
```

**Configuration**:
```go
flag.StringVar(&watchNamespace, "watch-namespace", "", "...")
flag.IntVar(&concurrency, "concurrency", 2, "...")
flag.DurationVar(&resync, "cache-resync", 0, "...")
flag.StringVar(&pprofAddr, "pprof-bind-address", "", "...")
flag.BoolVar(&enableLeaderElection, "leader-elect", true, "...")
```

### 2. cmd/tenant-policy-controller-simple/main.go (78 lines)

**Features**:
-  NO Calico CRD integration
-  Minimal configuration (only metrics/probe addresses)
-  Leader election **disabled** by default (`--leader-elect` defaults to false)
- Uses deprecated health check: `mgr.GetWebhookServer().StartedChecker()` (no webhook server!)
-  Missing production features

**Problem with Health Checks**:
```go
// Lines 63-66, 68-71
// WRONG: Using WebhookServer checker when no webhook server exists
if err := mgr.AddHealthzCheck("healthz", mgr.GetWebhookServer().StartedChecker()); err != nil {
    setupLog.Error(err, "unable to set up health check")
    os.Exit(1)
}
```

This will likely cause issues since no webhook server is configured!

## Key Differences Table

| Feature | tenant-policy-controller | tenant-policy-controller-simple |
|---------|--------------------------|----------------------------------|
| Lines of code | 119 | 78 |
| Calico CRDs | es |  No |
| Leader election default | nabled |  Disabled |
| Concurrency control | es (--concurrency) |  No |
| Watch namespace | es (--watch-namespace) |  No |
| Cache resync | es (--cache-resync) |  No |
| pprof profiling | es (--pprof-bind-address) |  No |
| Health checks | orrect (healthz.Ping) |  Wrong (WebhookServer) |
| Production ready | es |  No |

## Deployment Status

### Kubernetes Manifest
**File**: [k8s/tenant-policy-controller.yaml](k8s/tenant-policy-controller.yaml)

**Uses**: Full-featured version (tenant-policy-controller)
```yaml
containers:
- name: tenant-policy-controller
  image: tenant-policy-controller:tenant-controller-20251011-020400
```

**Permissions**: ClusterRole with access to:
- Pods (get, list, watch)
- Calico GlobalNetworkPolicies (full CRUD)
- Calico NetworkPolicies (full CRUD)
- Leases for leader election (full CRUD)

## Shared Controller

Both versions use the same controller implementation:
**File**: [internal/controllers/tenant_policy_controller.go](internal/controllers/tenant_policy_controller.go)

```go
type TenantPolicyReconciler struct {
    Client client.Client
    Scheme *runtime.Scheme
    Log    logr.Logger
}
```

This controller is responsible for:
- Watching Pods with `cloudbridge.io/tenant` label
- Creating/updating Calico NetworkPolicies for tenant isolation
- Managing multi-tenant network isolation

## Analysis

### Why Two Versions Exist?

**Hypothesis**:
- **tenant-policy-controller-simple** was created first as a prototype
- **tenant-policy-controller** was later developed with production features
- Simple version was kept for testing/development but never removed

### Evidence:
1. Simple version has **wrong health checks** (uses WebhookServer when none exists)
2. Simple version **disables leader election** (not production-ready)
3. Full version has **Calico CRDs** registered (required for actual functionality)
4. Kubernetes manifest uses **full version**, not simple

## Problems with Simple Version

### Critical Issues:
1. **Broken Health Checks**:
   ```go
   mgr.GetWebhookServer().StartedChecker()  // No webhook server exists!
   ```
   This will fail or return incorrect status

2. **No Leader Election**:
   - Default: `--leader-elect=false`
   - Multiple replicas will conflict
   - Not safe for production

3. **No Calico CRDs**:
   - Cannot manage Calico NetworkPolicies
   - Controller will fail when trying to create policies

4. **Limited Configuration**:
   - No concurrency control
   - No namespace restriction
   - No performance profiling

## Recommendation

### EEP: cmd/tenant-policy-controller (Full Version)

**Reasons**:
- Production-ready configuration
- Calico CRD integration (required for functionality)
- Proper health checks
- Leader election enabled
- Advanced configuration options
- Already deployed in k8s/tenant-policy-controller.yaml

###  REMOVE: cmd/tenant-policy-controller-simple

**Reasons**:
- Broken health checks
- Missing Calico CRDs (won't work!)
- Not production-ready
- Not used in any deployment
- Adds confusion

## Action Plan

### 1. Delete Simple Version
```bash
rm -rf cmd/tenant-policy-controller-simple/
```

### 2. Update Documentation
Document that tenant-policy-controller is the only version:
- Add README.md in cmd/tenant-policy-controller/
- Explain Calico integration requirement
- Document configuration options

### 3. Verify Dockerfile
Check that docker/Dockerfile.tenant-policy-controller builds the correct version

### 4. Verify Deployment
Confirm k8s/tenant-policy-controller.yaml is using the correct image

## Dockerfile Verification

**File**: [docker/Dockerfile.tenant-policy-controller](docker/Dockerfile.tenant-policy-controller)

**Line 20**:
```dockerfile
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o tenant-policy-controller ./cmd/tenant-policy-controller
```

 **Correctly builds the full version** (cmd/tenant-policy-controller)

**Dockerfile Features**:
- Multi-stage build (golang:1.25.12-alpine → alpine:latest)
- on-root user (tenant-controller:1001)
- inimal dependencies (ca-certificates, tzdata)
- ealth check configured
- ecurity best practices

## Conclusion

### Current State
- *cmd/tenant-policy-controller** - Used by Dockerfile and k8s manifest
-  **cmd/tenant-policy-controller-simple** - Unused, broken, should be removed

### Final Recommendation

**REMOVE cmd/tenant-policy-controller-simple** - It serves no purpose and has critical bugs:

```bash
# Remove unused simple version
rm -rf cmd/tenant-policy-controller-simple/

# Add README to document the full version
touch cmd/tenant-policy-controller/README.md
```

### Benefits of Removal
1. **Eliminate Confusion**: One controller, clear purpose
2. **Remove Broken Code**: Simple version has buggy health checks
3. **Simplify Maintenance**: One version to maintain and test
4. **Improve Documentation**: Focus on production-ready version

## Next Steps

1. elete cmd/tenant-policy-controller-simple/
2. reate README.md for cmd/tenant-policy-controller/
3. ocument Calico integration requirements
4. pdate PROJECT_ARCHITECTURE_ANALYSIS.md
