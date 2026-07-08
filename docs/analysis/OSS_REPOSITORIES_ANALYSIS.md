# OSS Repositories Analysis and Standards Compliance

**Date**: 2025-10-13
**Status**: **CRITICAL ANALYSIS COMPLETE**
**Priority**: 🔴 **HIGH - Production Security**

## Executive Summary

Анализ `oss-repositories/` показывает, что **многие стандарты уже реализованы** в upstream проектах:
- *pion/turn**: RFC 5766 short-lived credentials (TURN REST API draft-uberti-behave-turn-rest-00)
- *coturn**: Production-grade TURN server с full RFC 5766 compliance
- *pion/ice**: Complete ICE implementation с candidate gathering, connectivity checks
- ⚠️ **CloudBridge custom code**: Partial reinvention вместо использования pion standards

## Critical Finding: Code Duplication

### Issue

Созданный код в `internal/nat/turn_credentials.go` **дублирует функциональность** из `pion/turn`:

| Feature | Custom Code | Pion/Turn | Recommendation |
|---------|-------------|-----------|----------------|
| HMAC-SHA1 credentials | turn_credentials.go` | `l_cred.go` | **Use pion/turn** |
| Time-boxed username format | <timestamp>:<peer_id>` | `<imestamp>:<user>` | **Compatible - OK** |
| Credential validation | ustom | `LngTermTURNRESTAuthHandler` | **Integrate pion** |
| Quota management | ustom TURNQuotaManager |  Not in pion | **Keep custom** |

### Recommendation

**Action Plan**:
1. eep `TURNQuotaManager` (custom business logic)
2. **Refactor** `TURNCredentialManager` to **wrap** pion/turn functions
3. Use `pion/turn.GenerateLongTermTURNRESTCredentials()` directly
4. Use `pion/turn.LongTermTURNRESTAuthHandler()` for validation

---

## OSS Repository Inventory

### 1. Network Stack (`oss-repositories/network/`)

#### 1.1 ICE/STUN/TURN

**Directory**: `network/ice-stun-turn/`

##### pion-ice
- **Version**: Latest (git cloned)
- **License**: MIT
- **Status**: ntegrated
- **Standards**: RFC 8445 (ICE), RFC 5389 (STUN)
- **Key Files**:
  - `agent.go`: ICE agent implementation
  - `gather.go`: Candidate gathering
  - `selection.go`: Candidate pair selection
- **Integration**: `internal/nat/pion_ice_adapter.go`

##### pion-turn
- **Version**: v3.0.0+
- **License**: MIT (REUSE compliant)
- **Status**: ⚠️ **PARTIALLY INTEGRATED**
- **Standards**:
  - RFC 5766 (TURN)
  - draft-uberti-behave-turn-rest-00 (TURN REST API)
- **Key Files**:
  - `server.go`: TURN server implementation
  - `client.go`: TURN client
  - `lt_cred.go`: **Short-lived credentials** (RFC 5766 Section 10.2)
  - `relay_address_generator_*.go`: Relay allocation strategies
- **Critical Functions**:
  ```go
  // EXACTLY what we need!
  func GenerateLongTermTURNRESTCredentials(
      sharedSecret string,
      user string,
      duration time.Duration
  ) (username string, password string, error)

  // Server-side validation
  func LongTermTURNRESTAuthHandler(
      sharedSecret string,
      logger logging.LeveledLogger
  ) AuthHandler
  ```

##### coturn
- **Version**: 4.6.2
- **License**: BSD-3-Clause
- **Status**: **Available for deployment**
- **Language**: C
- **Standards**: Full RFC 5766 + extensions
- **Features**:
  - Long-term & short-term credentials
  - PostgreSQL/MySQL/Redis/MongoDB backends
  - Prometheus metrics (`--prometheus`)
  - STUN/TURN/TURNS (TLS)
  - Bandwidth limiting
  - Quota management
  - REST API support
- **Config**: `turnserver.conf`
- **Key Options**:
  ```bash
  listening-port=3478
  tls-listening-port=5349
  lt-cred-mech  # Long-term credential mechanism
  use-auth-secret  # TURN REST API mode
  static-auth-secret=<SECRET>
  realm=cloudbridge.2gc.ru
  ```

**Decision Point**:
- **Option A**: Use **embedded Go TURN** (`internal/nat/turn_server_enhanced.go`) + pion/turn for auth
- **Option B**: Deploy **coturn** as sidecar + API-only relay credentials
- **Option C**: **Hybrid**: coturn for production relay + Go API for credential vending

**Recommendation**: **Option C (Hybrid)**
- Coturn handles actual TURN relay (battle-tested, C performance)
- Go API handles credential generation, quota, tenant isolation
- Best security: short-lived credentials, per-tenant quotas

#### 1.2 QUIC (`network/quic/`)

**Directory**: `network/quic/quic-go/`

- **Version**: v0.48.2
- **License**: MIT
- **Status**: ully integrated
- **Features**:
  - BBRv2 congestion control
  - ACK Frequency (RFC 9???)
  - FEC (Forward Error Correction)
  - QUIC v1 (RFC 9000)
  - HTTP/3 (RFC 9114)

**Critical Issue**: Экспериментальные фичи без feature flags
- ⚠️ **BBRv2**: May cause issues with older stacks
- ⚠️ **ACK Frequency**: Not widely supported
- ⚠️ **FEC**: Custom implementation, interop risk

**Recommendation**: Add feature flags per tenant

#### 1.3 MASQUE (`network/masque/`)

- **Status**: ntegrated
- **Standards**: RFC 9298 (CONNECT-UDP)
- **Files**: `internal/masque/`

---

### 2. Kubernetes (`oss-repositories/k8s/`)

#### 2.1 MetalLB
- **Version**: v0.14.8
- **License**: Apache 2.0
- **Status**: ⚠️ **NOT DEPLOYED**
- **Purpose**: Bare-metal LoadBalancer
- **Modes**:
  - L2 mode (ARP/NDP)
  - BGP mode (production-ready)

**Gap**: VK Cloud LoadBalancer используется, но нет fallback для bare-metal

#### 2.2 kube-vip
- **Version**: v0.8.5
- **License**: Apache 2.0
- **Status**: ⚠️ **NOT DEPLOYED**
- **Purpose**: HA VIP for control plane / services

**Recommendation**: Deploy для multi-master HA

---

### 3. Supply Chain Security (`oss-repositories/supply-chain/`)

#### 3.1 cosign
- **Version**: v2.4.1
- **License**: Apache 2.0
- **Status**: cripts ready, ⚠️ **NO ENFORCEMENT**
- **Files**:
  - `scripts/setup/setup-cosign-keys.sh`
  - `.github/workflows/sign-images.yaml`

**Critical Gap**: Kyverno/OPA policies **NOT deployed**

#### 3.2 syft
- **Version**: v1.16.0
- **License**: Apache 2.0
- **Status**: BOM generation ready
- **Files**: `scripts/security/generate-sbom.sh`

#### 3.3 grype
- **Version**: v0.82.1
- **License**: Apache 2.0
- **Status**: ulnerability scanning ready

**Critical Gap**: No policy enforcement for High/Critical CVEs

---

### 4. Authentication (`oss-repositories/auth/`)

#### 4.1 Zitadel SDK
- **Status**: ully integrated
- **Files**: `internal/zitadel/`, `internal/auth/`
- **Standards**: OAuth 2.0, OIDC, JWT

---

### 5. Observability (`oss-repositories/observability/`)

#### 5.1 qlog
- **Status**: ntegrated
- **Standards**: draft-ietf-quic-qlog
- **Gap**: ⚠️ No sampling, no PII redaction

#### 5.2 qvis
- **Status**: isualization ready
- **Gap**: ⚠️ Access control not configured

---

## Standards Compliance Matrix

| Component | Standard | OSS Repo | Integration Status | Gap |
|-----------|----------|----------|-------------------|-----|
| TURN credentials | RFC 5766 §10.2 | pion/turn | ⚠️ **DUPLICATED** | Use pion functions |
| TURN server | RFC 5766 | coturn | ⚠️ **NOT DEPLOYED** | Deploy production |
| ICE | RFC 8445 | pion/ice | ntegrated | None |
| QUIC | RFC 9000 | quic-go | ntegrated | Feature flags needed |
| HTTP/3 | RFC 9114 | quic-go | ntegrated | None |
| MASQUE | RFC 9298 | masque-go | ntegrated | None |
| JWT | RFC 7519 | go-jose | ntegrated | None |
| OIDC | OpenID Connect | zitadel-sdk | ntegrated | None |
| Image signing | Sigstore | cosign | ⚠️ **NO ENFORCEMENT** | Deploy Kyverno |
| SBOM | SPDX/CycloneDX | syft | enerated | Enforce in CI/CD |
| qlog | QLOG draft | qlog | ⚠️ **NO SAMPLING** | Add sampling |

---

## Recommendations by Priority

### 🔴 Priority 1: Security (Immediate)

1. **TURN Credentials Integration**
   -  **STOP** using custom `turn_credentials.go`
   - *USE** `pion/turn.GenerateLongTermTURNRESTCredentials()`
   - *KEEP** `TURNQuotaManager` for business logic
   - **Timeline**: Today
   - **File**: `internal/nat/turn_credentials_pion.go`

2. **Deploy Kyverno Policies**
   - Enforce image signatures
   - Require SBOM attestation
   - Deny unsigned images
   - **Timeline**: Week 1
   - **Files**: `k8s/policies/kyverno/`

3. **Deploy coturn Production**
   - Use `oss-repositories/network/ice-stun-turn/coturn/`
   - Config: `use-auth-secret` + shared secret
   - 2+ replicas with LoadBalancer
   - **Timeline**: Week 1

### 🟡 Priority 2: Reliability (Week 2)

4. **Deploy MetalLB/kube-vip**
   - BGP mode for production
   - L2 mode for lab
   - **Timeline**: Week 2

5. **Add QUIC Feature Flags**
   - Per-tenant BBRv2 toggle
   - ACK Frequency opt-in
   - FEC opt-in
   - **Timeline**: Week 2

6. **qlog Sampling**
   - 1-5% sampling rate
   - PII redaction
   - TTL 7-14 days
   - **Timeline**: Week 2

### 🟢 Priority 3: Operational Excellence (Week 3-4)

7. **External Secrets Rollout**
   - Blue/green activation
   - Fail-close mode
   - **Timeline**: Week 3

8. **API Contract Tests (Pact)**
   - Consumer-driven tests
   - Version compatibility matrix
   - **Timeline**: Week 3

9. **SLO Dashboards**
   - NAT success ≥95%
   - TURN ratio ≤35%
   - P95 RTT <50ms
   - **Timeline**: Week 4

---

## License Compliance Report

### Approved Licenses (Compatible with Proprietary)

 **MIT**: quic-go, pion/*, zitadel-sdk, gobreaker
 **Apache 2.0**: cosign, syft, grype, MetalLB, kube-vip
 **BSD-3-Clause**: coturn

### Action Items

1. enerate `THIRD_PARTY_NOTICES.md`
2. dd SPDX identifiers to all files
3.  **TODO**: Setup automated license scanning (scancode/ort)

**Script**:
```bash
# Generate THIRD_PARTY_NOTICES
go run github.com/google/go-licenses save ./... --save_path=licenses/
```

---

## Exit Criteria for OSS Integration

Each OSS component считается "fully integrated" только если:

- [x] *License approved** and documented
- [ ]  **Imported as Go module** (NOT cloned)
- [x] *Tests passing** in CI/CD
- [ ]  **Security scanning** (grype/trivy) on dependencies
- [ ]  **Runbook** created (upgrade, rollback, troubleshoot)
- [x] *ADR** documented (why this library?)
- [ ]  **SLO defined** and monitored

**Current Status**: 3/7 criteria met (43%)

---

## Refactoring Plan: TURN Credentials

### Phase 1: Wrapper (Today)

Create `internal/nat/turn_credentials_pion.go`:

```go
package nat

import (
    "time"

    pionTurn "github.com/pion/turn/v3"
    "go.uber.org/zap"
)

// TURNCredentialService wraps pion/turn with quota management
type TURNCredentialService struct {
    logger       *zap.Logger
    sharedSecret string
    quotaManager *TURNQuotaManager
}

func NewTURNCredentialService(sharedSecret string, logger *zap.Logger) *TURNCredentialService {
    return &TURNCredentialService{
        logger:       logger,
        sharedSecret: sharedSecret,
        quotaManager: NewTURNQuotaManager(logger),
    }
}

// GenerateCredential uses pion/turn standard implementation
func (tcs *TURNCredentialService) GenerateCredential(
    tenantID string,
    peerID string,
    ttl time.Duration,
) (*TURNCredentialResponse, error) {
    // Check quota first
    if err := tcs.quotaManager.CheckQuota(tenantID, peerID); err != nil {
        return nil, err
    }

    // Use pion/turn standard function
    username, password, err := pionTurn.GenerateLongTermTURNRESTCredentials(
        tcs.sharedSecret,
        peerID,
        ttl,
    )
    if err != nil {
        return nil, err
    }

    // Increment quota
    tcs.quotaManager.IncrementAllocation(tenantID, peerID)

    // Return in our format
    return &TURNCredentialResponse{
        Username:   username,
        Password:   password,
        TTL:        int(ttl.Seconds()),
        ExpiresAt:  time.Now().Add(ttl),
        URIs:       tcs.getTURNURIs(),
        Realm:      "cloudbridge.2gc.ru",
        QuotaUsed:  tcs.quotaManager.GetTenantQuota(tenantID).ActiveAllocations,
        QuotaLimit: tcs.quotaManager.GetTenantQuota(tenantID).QuotaLimit,
    }, nil
}
```

### Phase 2: Server Integration (Week 1)

Update `internal/nat/turn_server_enhanced.go` to use pion auth handler:

```go
import pionTurn "github.com/pion/turn/v3"

func (ts *EnhancedTURNServer) authenticate(message *STUNMessage, clientAddr *net.UDPAddr) bool {
    // Use pion/turn standard validation
    authHandler := pionTurn.LongTermTURNRESTAuthHandler(
        ts.sharedSecret,
        ts.logger,
    )

    key, ok := authHandler(username, ts.realm, clientAddr)
    return ok && key != nil
}
```

### Phase 3: Deprecation (Week 2)

- Mark `turn_credentials.go` as deprecated
- Redirect all calls to `turn_credentials_pion.go`
- Add migration guide

### Phase 4: Removal (Week 4)

- Delete old implementation
- Update all imports
- Release notes: "TURN credentials now use pion/turn standards"

---

## Metrics for OSS Integration Quality

### Current Score: 62/100

| Metric | Weight | Score | Max | Notes |
|--------|--------|-------|-----|-------|
| Standards compliance | 30% | 21 | 30 | Using pion, but duplicating |
| License compliance | 20% | 20 | 20 | All licenses approved |
| Security enforcement | 25% | 5 | 25 | No Kyverno, no policy |
| Dependency management | 15% | 8 | 15 | Cloned vs modules |
| Documentation | 10% | 8 | 10 | Good ADRs, missing runbooks |

**Target**: ≥85/100 for production readiness

---

## Action Items Summary

### Today (Immediate)
- [ ] Create `turn_credentials_pion.go` wrapper
- [ ] Update `ice_credentials.go` handler to use wrapper
- [ ] Run integration tests
- [ ] Document migration path

### Week 1
- [ ] Deploy Kyverno policies
- [ ] Deploy coturn production (2+ replicas)
- [ ] Configure coturn with shared secret
- [ ] Test end-to-end TURN flow
- [ ] Create runbooks

### Week 2
- [ ] Convert cloned repos to Go modules
- [ ] Setup license scanning CI/CD
- [ ] Add QUIC feature flags
- [ ] Deploy MetalLB/kube-vip staging
- [ ] qlog sampling implementation

### Week 3-4
- [ ] External Secrets blue/green
- [ ] API contract tests (Pact)
- [ ] SLO dashboards
- [ ] Chaos testing
- [ ] Production validation

---

## Conclusion

**Key Findings**:
1. *Good**: OSS repos contain standards-compliant implementations
2. ⚠️ **Issue**: Custom code duplicates pion/turn functionality
3. 🔴 **Critical**: No security policy enforcement (Kyverno)
4. 🔴 **Critical**: coturn not deployed in production

**Immediate Action**: Refactor TURN credentials to use pion/turn standards TODAY.

**Timeline**: 4-week critical path to close all gaps.

---

**Status**: Document complete, ready for review
**Next Steps**: Begin Phase 1 refactoring (turn_credentials_pion.go)
