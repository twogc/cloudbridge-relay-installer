# Kyverno Deployment Guide

**Date**: 2025-10-13
**Status**: Ready for deployment
**Priority**: 🔴 **CRITICAL - Week 1**

---

## Executive Summary

Deploy **Kyverno policy engine** to enforce supply chain security policies:
- ✅ Image signature verification (cosign)
- ✅ SBOM attestation requirements
- ✅ Vulnerability scan enforcement
- ✅ Allowed registries whitelist
- ✅ Non-root containers
- ✅ Immutable tags

**Deployment Strategy**: **Gradual rollout** (Audit → Enforce)

---

## Prerequisites

### 1. Kyverno Installation

```bash
# Add Helm repository
helm repo add kyverno https://kyverno.github.io/kyverno/
helm repo update

# Create namespace
kubectl create namespace kyverno

# Install Kyverno
helm install kyverno kyverno/kyverno \
  --namespace kyverno \
  --version 3.1.4 \
  --set replicaCount=3 \
  --set createSelfSignedCert=true \
  --set resources.limits.memory=512Mi \
  --set resources.requests.memory=256Mi

# Verify installation
kubectl get pods -n kyverno
kubectl get validatingwebhookconfigurations
```

### 2. Generate Cosign Keys

```bash
# Generate signing key pair
cosign generate-key-pair

# This creates:
# - cosign.key (private key - keep secret!)
# - cosign.pub (public key - use in policies)

# Store private key in Kubernetes secret
kubectl create secret generic cosign-signing-key \
  --from-file=cosign.key \
  --namespace=cloudbridge

# Store public key for verification
kubectl create secret generic cosign-pub-key \
  --from-file=cosign.pub \
  --namespace=cloudbridge

# Display public key for policies
cat cosign.pub
```

---

## Deployment Phases

### Phase 1: Audit Mode (Week 1, Days 1-2)

**Goal**: Collect data on policy violations without blocking deployments

```bash
# Deploy policies in AUDIT mode
kubectl apply -f k8s/policies/kyverno/require-image-signature.yaml
kubectl apply -f k8s/policies/kyverno/require-sbom.yaml

# All policies start with validationFailureAction: Audit
```

**Monitor violations**:
```bash
# View policy reports
kubectl get policyreport -A

# View cluster-wide reports
kubectl get clusterpolicyreport

# Detailed report
kubectl describe policyreport polr-ns-cloudbridge -n cloudbridge
```

**Expected findings**:
- Unsigned images: ~50-80% initially
- Missing SBOM: ~90% initially
- :latest tags: ~20-30%
- Non-approved registries: ~10-15%

### Phase 2: Gradual Enforcement (Week 1, Days 3-5)

**Goal**: Enforce critical policies, audit others

**Step 1: Update policies to Enforce**

```bash
# Edit policies to change validationFailureAction: Audit → Enforce
kubectl edit clusterpolicy require-image-signature
kubectl edit clusterpolicy disallow-latest-tag
kubectl edit clusterpolicy allowed-registries
```

**Step 2: Test enforcement**

```bash
# Try deploying unsigned image (should fail)
kubectl run test-unsigned \
  --image=nginx:latest \
  --namespace=cloudbridge

# Expected: Error from Kyverno admission webhook
```

**Step 3: Sign existing images**

```bash
# Sign all CloudBridge images
for image in $(kubectl get pods -n cloudbridge -o jsonpath='{.items[*].spec.containers[*].image}'); do
  echo "Signing: $image"
  cosign sign --key cosign.key $image
done
```

### Phase 3: Full Enforcement (Week 1, Days 6-7)

**Goal**: All policies enforced, zero unsigned images

```bash
# Update ALL policies to Enforce
for policy in k8s/policies/kyverno/*.yaml; do
  # Change Audit → Enforce in all policies
  sed -i 's/validationFailureAction: Audit/validationFailureAction: Enforce/g' $policy
done

# Apply updated policies
kubectl apply -f k8s/policies/kyverno/
```

**Verify enforcement**:
```bash
# Check policy status
kubectl get clusterpolicy

# View admission metrics
kubectl get --raw /metrics | grep kyverno_policy
```

---

## Policy Configuration

### 1. Image Signature Policy

**File**: `k8s/policies/kyverno/require-image-signature.yaml`

**Update with your cosign public key**:

```bash
# Extract public key
COSIGN_PUB_KEY=$(cat cosign.pub)

# Update policy
cat > /tmp/update-policy.yaml <<EOF
spec:
  rules:
    - name: require-signature-cloudbridge-images
      verifyImages:
        - attestors:
            - entries:
                - keys:
                    publicKeys: |-
$(cat cosign.pub | sed 's/^/                      /')
EOF

# Apply to existing policy
kubectl patch clusterpolicy require-image-signature \
  --type merge \
  --patch-file /tmp/update-policy.yaml
```

### 2. SBOM Attestation Policy

**File**: `k8s/policies/kyverno/require-sbom.yaml`

**Configuration**:
- Accepts SPDX 2.3 or CycloneDX 1.4/1.5
- Requires minimum 1 package documented
- Verifies attestation signature

**Generate SBOM for image**:
```bash
# Generate SBOM
syft packages \
  cr.yandex/crpblcmklr11njvr6ikg/cloudbridge-relay:v1.0.0 \
  -o spdx-json > sbom.spdx.json

# Attest SBOM
cosign attest --predicate sbom.spdx.json \
  --key cosign.key \
  --type spdx \
  cr.yandex/crpblcmklr11njvr6ikg/cloudbridge-relay:v1.0.0
```

### 3. Vulnerability Scan Policy

**File**: `k8s/policies/kyverno/require-sbom.yaml` (vulnerability section)

**Configuration**:
- **Critical**: 0 allowed (BLOCK)
- **High**: ≤5 allowed (with justification)
- **Medium/Low**: Unlimited (AUDIT)

**Scan and attest**:
```bash
# Scan with grype
grype cr.yandex/crpblcmklr11njvr6ikg/cloudbridge-relay:v1.0.0 \
  -o json > vulnerabilities.json

# Attest scan results
cosign attest --predicate vulnerabilities.json \
  --key cosign.key \
  --type https://cosign.sigstore.dev/attestation/vuln/v1 \
  cr.yandex/crpblcmklr11njvr6ikg/cloudbridge-relay:v1.0.0
```

---

## Policy Exceptions

### Temporary Bypass (Emergency)

```yaml
# Add label to pod to bypass policies
apiVersion: v1
kind: Pod
metadata:
  name: emergency-pod
  namespace: cloudbridge
  labels:
    kyverno.io/exclude: "true"  # Bypass ALL policies
spec:
  containers:
    - name: app
      image: unsigned-image:latest
```

⚠️ **WARNING**: Only use in emergencies. Requires approval.

### Namespace Exclusions

```bash
# Exclude namespace from policies
kubectl label namespace cloudbridge-dev \
  policy.kyverno.io/exclude=enabled
```

### Per-Policy Exceptions

```yaml
# Allow root for specific app
apiVersion: v1
kind: Pod
metadata:
  labels:
    security.cloudbridge.2gc.ru/allow-root: "true"
spec:
  containers:
    - name: privileged-app
      securityContext:
        runAsUser: 0
```

---

## Monitoring

### Prometheus Metrics

```prometheus
# Policy violations
sum by (policy_name, policy_action) (
  kyverno_policy_results_total{policy_action="block"}
)

# Admission latency
histogram_quantile(0.95,
  rate(kyverno_admission_requests_duration_seconds_bucket[5m])
)

# Policy evaluation time
rate(kyverno_policy_execution_duration_seconds_sum[5m])
```

### Alerts

```yaml
# k8s/monitoring/alerts/kyverno-alerts.yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: kyverno-alerts
  namespace: cloudbridge
spec:
  groups:
    - name: kyverno
      interval: 30s
      rules:
        - alert: KyvernoHighViolationRate
          expr: |
            sum(rate(kyverno_policy_results_total{policy_action="block"}[5m])) > 10
          for: 5m
          labels:
            severity: warning
          annotations:
            summary: "High rate of policy violations"
            description: "More than 10 violations/min in last 5 minutes"

        - alert: KyvernoWebhookDown
          expr: |
            absent(up{job="kyverno"})
          for: 2m
          labels:
            severity: critical
          annotations:
            summary: "Kyverno webhook is down"
            description: "Policy enforcement is disabled"
```

### Dashboards

```bash
# Import Kyverno Grafana dashboard
kubectl apply -f https://raw.githubusercontent.com/kyverno/kyverno/main/charts/kyverno/dashboards/kyverno-dashboard.json
```

**Key panels**:
- Policy violations over time
- Admission latency p95/p99
- Policy evaluation success rate
- Top violating images
- Namespace compliance score

---

## CI/CD Integration

### GitHub Actions

```yaml
# .github/workflows/sign-and-attest.yaml
name: Sign and Attest Image

on:
  push:
    branches: [main]

jobs:
  sign:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      id-token: write  # For keyless signing
      packages: write

    steps:
      - uses: actions/checkout@v4

      - name: Install cosign
        uses: sigstore/cosign-installer@v3

      - name: Install syft
        uses: anchore/sbom-action/download-syft@v0

      - name: Login to registry
        uses: docker/login-action@v3
        with:
          registry: cr.yandex
          username: ${{ secrets.REGISTRY_USERNAME }}
          password: ${{ secrets.REGISTRY_PASSWORD }}

      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          push: true
          tags: cr.yandex/crpblcmklr11njvr6ikg/cloudbridge-relay:${{ github.sha }}

      - name: Sign image
        run: |
          cosign sign --yes \
            --key env://COSIGN_KEY \
            cr.yandex/crpblcmklr11njvr6ikg/cloudbridge-relay:${{ github.sha }}
        env:
          COSIGN_KEY: ${{ secrets.COSIGN_PRIVATE_KEY }}
          COSIGN_PASSWORD: ${{ secrets.COSIGN_PASSWORD }}

      - name: Generate SBOM
        run: |
          syft packages \
            cr.yandex/crpblcmklr11njvr6ikg/cloudbridge-relay:${{ github.sha }} \
            -o spdx-json > sbom.spdx.json

      - name: Attest SBOM
        run: |
          cosign attest --yes \
            --predicate sbom.spdx.json \
            --type spdx \
            --key env://COSIGN_KEY \
            cr.yandex/crpblcmklr11njvr6ikg/cloudbridge-relay:${{ github.sha }}
        env:
          COSIGN_KEY: ${{ secrets.COSIGN_PRIVATE_KEY }}
          COSIGN_PASSWORD: ${{ secrets.COSIGN_PASSWORD }}

      - name: Scan vulnerabilities
        run: |
          grype cr.yandex/crpblcmklr11njvr6ikg/cloudbridge-relay:${{ github.sha }} \
            -o json > vulnerabilities.json

      - name: Attest vulnerabilities
        run: |
          cosign attest --yes \
            --predicate vulnerabilities.json \
            --type https://cosign.sigstore.dev/attestation/vuln/v1 \
            --key env://COSIGN_KEY \
            cr.yandex/crpblcmklr11njvr6ikg/cloudbridge-relay:${{ github.sha }}
        env:
          COSIGN_KEY: ${{ secrets.COSIGN_PRIVATE_KEY }}
          COSIGN_PASSWORD: ${{ secrets.COSIGN_PASSWORD }}
```

---

## Troubleshooting

### Issue: Policy not enforcing

**Check webhook configuration**:
```bash
kubectl get validatingwebhookconfigurations
kubectl describe validatingwebhookconfigurations kyverno-resource-validating-webhook-cfg
```

**Verify Kyverno is running**:
```bash
kubectl get pods -n kyverno
kubectl logs -n kyverno deployment/kyverno -f
```

### Issue: Signature verification fails

**Check public key**:
```bash
# Verify policy has correct public key
kubectl get clusterpolicy require-image-signature -o yaml | grep "BEGIN PUBLIC KEY" -A 10

# Manually verify signature
cosign verify --key cosign.pub cr.yandex/crpblcmklr11njvr6ikg/cloudbridge-relay:v1.0.0
```

### Issue: SBOM attestation not found

**Check attestations**:
```bash
# List attestations
cosign tree cr.yandex/crpblcmklr11njvr6ikg/cloudbridge-relay:v1.0.0

# Verify specific attestation
cosign verify-attestation \
  --key cosign.pub \
  --type spdx \
  cr.yandex/crpblcmklr11njvr6ikg/cloudbridge-relay:v1.0.0
```

### Issue: High admission latency

**Check resource limits**:
```bash
# Increase Kyverno resources
helm upgrade kyverno kyverno/kyverno \
  --namespace kyverno \
  --reuse-values \
  --set resources.limits.memory=1Gi \
  --set resources.requests.memory=512Mi
```

---

## Rollback Procedure

### Emergency Disable

```bash
# Disable ALL policies (emergency only)
kubectl patch clusterpolicy require-image-signature \
  --type merge \
  --patch '{"spec":{"validationFailureAction":"Audit"}}'

kubectl patch clusterpolicy require-sbom-attestation \
  --type merge \
  --patch '{"spec":{"validationFailureAction":"Audit"}}'

# Or delete policies entirely
kubectl delete clusterpolicy --all
```

### Gradual Rollback

```bash
# Move back to Audit mode
for policy in $(kubectl get clusterpolicy -o name); do
  kubectl patch $policy \
    --type merge \
    --patch '{"spec":{"validationFailureAction":"Audit"}}'
done
```

---

## Success Criteria

**Week 1 Goals**:
- [x] Kyverno installed and running
- [ ] All policies deployed in Audit mode
- [ ] Policy reports generated
- [ ] Image signing process documented
- [ ] CI/CD integration complete

**Week 2 Goals**:
- [ ] Critical policies enforced (signatures, registries)
- [ ] 100% of images signed
- [ ] SBOM attestations for all images
- [ ] Zero policy violations in production

**Week 4 Goals**:
- [ ] All policies enforced
- [ ] Automated compliance reporting
- [ ] SLO: <1% policy violations
- [ ] Monitoring dashboards operational

---

## Next Steps

1. **Deploy Kyverno** (30 min)
   ```bash
   helm install kyverno kyverno/kyverno --namespace kyverno
   ```

2. **Apply policies in Audit mode** (15 min)
   ```bash
   kubectl apply -f k8s/policies/kyverno/
   ```

3. **Monitor for 48 hours** (2 days)
   ```bash
   kubectl get policyreport -A --watch
   ```

4. **Sign existing images** (4 hours)
   ```bash
   ./scripts/security/sign-all-images.sh
   ```

5. **Enable enforcement** (1 hour)
   ```bash
   ./scripts/security/enable-policy-enforcement.sh
   ```

---

## References

- **Kyverno Docs**: https://kyverno.io/docs/
- **Cosign**: https://docs.sigstore.dev/cosign/overview/
- **SLSA**: https://slsa.dev/
- **SPDX**: https://spdx.dev/
- **CycloneDX**: https://cyclonedx.org/

---

**Status**: Ready for deployment
**Estimated Time**: 8 hours (Week 1)
**Risk Level**: Medium (gradual rollout mitigates)
