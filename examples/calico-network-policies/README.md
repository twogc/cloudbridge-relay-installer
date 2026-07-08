# CloudBridge "Железобетонная" Network Security

🏆 **Enterprise-grade multi-tenant network isolation using Calico Network Policies**

## Overview

This directory contains production-ready Calico Network Policies that implement **"железобетонную" (ironclad)** security for CloudBridge multi-tenant environments. The policies follow NSA/CISA Kubernetes hardening guidelines and implement defense-in-depth security.

## Security Features

### **Implemented Protections:**

- **Default-deny inter-tenant communication** - Zero-trust baseline
- **Per-tenant isolation** - Dedicated network policies per tenant
- **IMDS/Metadata server protection** - Blocks AWS/GCP/Azure metadata access
- **kube-apiserver access control** - Prevents unauthorized API access
- **DNS/System service exceptions** - Allows necessary system communication
- **Network policy monitoring** - Comprehensive logging and metrics

### **Defense Layers:**

1. **Tier-based policy ordering** - Correct precedence for security rules
2. **Comprehensive metadata blocking** - Multi-cloud IMDS protection
3. **Fine-grained RBAC** - Role-based access control
4. **Automated testing** - Validation of isolation effectiveness

## 📁 Files Description

| File | Purpose |
|------|---------|
| `01-security-tier.yaml` | Creates dedicated security tier for policy ordering |
| `02-baseline-deny-all.yaml` | Default-deny policy for inter-tenant traffic |
| `03-per-tenant-allow.yaml` | Per-tenant allow policies (template included) |
| `04-system-exceptions.yaml` | DNS and system service access exceptions |
| `05-metadata-protection.yaml` | IMDS and kube-apiserver blocking |
| `06-rbac-security.yaml` | Advanced RBAC and security features |
| `deploy-ironclad-security.sh` | **Main deployment script** |
| `test-tenant-isolation.sh` | **Automated testing suite** |

## Quick Start

### 1. Prerequisites

```bash
# Ensure Calico is installed
kubectl get crd globalnetworkpolicies.projectcalico.org

# Ensure kubectl access to cluster
kubectl cluster-info
```

### 2. Deploy Security Policies

```bash
# Make scripts executable
chmod +x deploy-ironclad-security.sh test-tenant-isolation.sh

# Deploy all security policies
./deploy-ironclad-security.sh
```

### 3. Test Isolation

```bash
# Run automated isolation tests
./test-tenant-isolation.sh
```

Expected output:
```
ALL TESTS PASSED - ЖЕЛЕЗОБЕТОННАЯ ISOLATION CONFIRMED! 🏆

Security validation results:
  Inter-tenant ping blocked
  Inter-tenant HTTP blocked
  Metadata services blocked
  DNS resolution working
  Intra-tenant communication working
```

## 🏗️ Manual Configuration

### Adding New Tenants

For each new tenant, create a dedicated allow policy:

```yaml
apiVersion: projectcalico.org/v3
kind: GlobalNetworkPolicy
metadata:
  name: tenant-YOUR_TENANT_ID-allow-intra
spec:
  tier: security
  order: 200
  selector: tenant.cloudbridge.io/id == "tenant-YOUR_TENANT_ID"
  types: [Ingress, Egress]
  ingress:
  - action: Allow
    source:
      selector: tenant.cloudbridge.io/id == "tenant-YOUR_TENANT_ID"
  egress:
  - action: Allow
    destination:
      selector: tenant.cloudbridge.io/id == "tenant-YOUR_TENANT_ID"
```

### Labeling Pods

Label your tenant pods appropriately:

```bash
# Label CloudBridge relay pods
kubectl label pods -n cloudbridge -l app=cloudbridge-relay \
  tenant.cloudbridge.io/id=system --overwrite

# Label tenant-specific pods
kubectl label pods -n tenant-namespace --all \
  tenant.cloudbridge.io/id=tenant-XXXXX --overwrite
```

## Monitoring & Troubleshooting

### Check Policy Status

```bash
# List all GlobalNetworkPolicies
kubectl get globalnetworkpolicies

# Check specific policy
kubectl describe globalnetworkpolicy deny-all-cross-tenant-baseline
```

### Monitor Policy Hits

```bash
# Felix logs for policy decisions
kubectl logs -n calico-system -l k8s-app=calico-felix | grep -i policy

# Check denied connections
kubectl logs -n calico-system -l k8s-app=calico-felix | grep "denied"
```

### Test Specific Scenarios

```bash
# Test ping between tenants (should fail)
kubectl exec pod-tenant-A -- ping pod-tenant-B-ip

# Test HTTP between tenants (should fail)
kubectl exec pod-tenant-A -- curl http://svc-tenant-B

# Test DNS resolution (should work)
kubectl exec pod-tenant-A -- nslookup kubernetes.default.svc.cluster.local

# Test IMDS access (should fail)
kubectl exec pod-tenant-A -- curl http://169.254.169.254/
```

## ⚙️ Advanced Configuration

### Custom RBAC Rules

Modify `06-rbac-security.yaml` to add custom RBAC rules:

```yaml
# Example: Allow specific pods to access kube-apiserver
apiVersion: projectcalico.org/v3  
kind: NetworkPolicy
metadata:
  name: allow-apiserver-for-operators
  namespace: your-namespace
spec:
  selector: role == "operator"
  types: [Egress]
  egress:
  - action: Allow
    destination:
      services:
        name: kubernetes
        namespace: default
```

### Rate Limiting

Add rate limiting annotations to policies (if supported):

```yaml
metadata:
  annotations:
    "calico.io/rate-limit": "100req/min"
```

### Multi-Cloud Metadata Protection

The policies include protection for multiple cloud providers:

- **AWS**: 169.254.169.254 (IMDS v1/v2)
- **GCP**: 169.254.169.254 (metadata server)
- **Azure**: 169.254.169.254 + 168.63.129.16 (WireServer)

## Performance Considerations

- **Policy order matters** - Lower numbers = higher precedence
- **Selector efficiency** - Use specific selectors for better performance
- **Tier usage** - Reduces policy evaluation overhead
- **Regular cleanup** - Remove unused policies

## Common Issues

### 1. Policies Not Applied

```bash
# Check Calico controller status
kubectl get pods -n calico-system

# Verify CRD installation
kubectl get crd | grep calico
```

### 2. DNS Resolution Broken

Verify system exceptions are applied:
```bash
kubectl get globalnetworkpolicy allow-system-services -o yaml
```

### 3. Tier Not Supported

Some Calico installations don't support Tiers. The policies include fallback configurations without tiers.

## 📚 References

- [Calico Network Policies Documentation](https://docs.projectcalico.org/security/policy)
- [NSA/CISA Kubernetes Hardening Guide](https://media.defense.gov/2022/Aug/29/2003066362/-1/-1/0/CTR_KUBERNETES_HARDENING_GUIDANCE_1.2_20220829.PDF)
- [Kubernetes Network Policies](https://kubernetes.io/docs/concepts/services-networking/network-policies/)
- [Calico Policy Tiers](https://docs.tigera.io/calico/latest/network-policy/policy-tiers/tiered-policy)

## Compliance

These policies meet or exceed requirements for:

- **NSA/CISA Kubernetes Hardening**
- **CIS Kubernetes Benchmark**
- **NIST Cybersecurity Framework**
- **Zero Trust Architecture principles**

---

**Status**: Production Ready  
**Last Updated**: September 27, 2025  
**Tested With**: Calico v3.x, Kubernetes 1.25+

