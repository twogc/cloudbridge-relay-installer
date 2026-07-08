# 🧪 Manual Testing Plan - CloudBridge Relay MVP

**Дата**: 2025-10-13
**Цель**: Проверить все критические пути перед production deployment
**Время**: 3-4 часа
**MVP Готовность**: 90% → 95%

---

## 📋 Testing Checklist

### ✅ Test Suite 1: Service Startup & Health Checks (30 минут)
- [ ] 1.1: Relay server starts without errors
- [ ] 1.2: Health endpoint responds (GET /health)
- [ ] 1.3: Ready endpoint shows all services (GET /ready)
- [ ] 1.4: Metrics endpoint accessible (GET /metrics)
- [ ] 1.5: P2P API server starts on correct port
- [ ] 1.6: Compatibility layer loaded in health check

### ✅ Test Suite 2: Zitadel JWKS Integration (45 минут)
- [ ] 2.1: JWKS endpoint discovery works
- [ ] 2.2: JWKS fetched from Zitadel successfully
- [ ] 2.3: JWKS cached with correct TTL (1 hour)
- [ ] 2.4: Public key retrieval by keyID
- [ ] 2.5: JWT token validation with JWKS
- [ ] 2.6: Token introspection endpoint works
- [ ] 2.7: Cache refresh after TTL expiration
- [ ] 2.8: Error handling for invalid tokens

### ✅ Test Suite 3: API Compatibility Layer (45 минут)
- [ ] 3.1: Legacy path works: POST /api/v1/tenants/:tid/peers/register
- [ ] 3.2: Legacy path works: GET /api/v1/tenants/:tid/peers
- [ ] 3.3: Legacy path works: GET /api/v1/tenants/:tid/peers/:pid
- [ ] 3.4: Legacy path works: PUT /api/v1/tenants/:tid/peers/:pid/status
- [ ] 3.5: Legacy path works: GET /api/v1/tenants/:tid/peers/:pid/wireguard-config
- [ ] 3.6: tenant_id extracted and added to context
- [ ] 3.7: Compatibility metrics collected
- [ ] 3.8: Compatibility logs generated (debug & info)

### ✅ Test Suite 4: TURN Server (30 минут)
- [ ] 4.1: TURN server deployment script runs successfully
- [ ] 4.2: TURN pods running (2 replicas)
- [ ] 4.3: LoadBalancer service has external IP
- [ ] 4.4: UDP port 3478 accessible
- [ ] 4.5: TCP port 3478 accessible
- [ ] 4.6: TLS port 5349 accessible
- [ ] 4.7: Prometheus metrics endpoint works (port 9641)
- [ ] 4.8: TURN credentials work

### ✅ Test Suite 5: NAT Traversal Integration (45 минут)
- [ ] 5.1: ICE candidates generation works
- [ ] 5.2: STUN binding requests succeed
- [ ] 5.3: TURN allocation succeeds
- [ ] 5.4: Relay through TURN works
- [ ] 5.5: NAT type detection works
- [ ] 5.6: Direct P2P connection (when possible)
- [ ] 5.7: Fallback to TURN when direct fails
- [ ] 5.8: Connection metrics reported

### ✅ Test Suite 6: Multi-Tenancy & Isolation (30 минут)
- [ ] 6.1: Multiple tenants can register
- [ ] 6.2: Tenant isolation enforced
- [ ] 6.3: Per-tenant feature flags work
- [ ] 6.4: Tenant-specific IPAM allocation
- [ ] 6.5: Cross-tenant requests blocked

### ✅ Test Suite 7: Monitoring & Observability (15 минут)
- [ ] 7.1: Prometheus metrics scraped successfully
- [ ] 7.2: Logs structured and readable
- [ ] 7.3: Error logs contain useful context
- [ ] 7.4: Metrics match actual system state
- [ ] 7.5: Health checks reflect real status

---

## 🔧 Test Environment Setup

### Prerequisites

```bash
# 1. Check Go version
go version  # Should be 1.21+

# 2. Check Kubernetes cluster
kubectl cluster-info
kubectl get nodes

# 3. Check required tools
which curl jq kubectl

# 4. Set environment variables
export KUBECONFIG=~/.kube/config
export NAMESPACE=cloudbridge
export JWT_SECRET="your-test-jwt-secret"
```

### Build the Project

```bash
# Navigate to project
cd /Users/maxlanies/Git/2GC/cloudbridge-relay-installer

# Build relay server
go build -o bin/relay ./cmd/relay

# Check for compilation errors
echo "Exit code: $?"
```

### Configuration Files

```bash
# Check config exists
ls -lh configs/config.json

# Validate JSON syntax
jq . configs/config.json > /dev/null && echo "✅ Valid JSON" || echo "❌ Invalid JSON"
```

---

## 📝 Test Execution Guide

### Test Suite 1: Service Startup & Health Checks

#### 1.1: Relay Server Starts Without Errors

```bash
# Start relay server (in background or separate terminal)
./bin/relay -config configs/config.json -log-level debug 2>&1 | tee logs/relay-startup.log &
RELAY_PID=$!

# Wait for startup
sleep 5

# Check if process running
ps -p $RELAY_PID && echo "✅ Relay running" || echo "❌ Relay crashed"

# Check logs for errors
grep -i "error\|fatal\|panic" logs/relay-startup.log && echo "❌ Errors found" || echo "✅ No errors"
```

**Expected Output**:
```
✅ Relay running
✅ No errors
```

**Actual Output**:
```
[Record your actual output here]
```

---

#### 1.2: Health Endpoint Responds

```bash
# Test health endpoint
curl -s http://localhost:8080/health | jq .

# Check status code
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health
```

**Expected Output**:
```json
{
  "status": "healthy",
  "timestamp": "2025-10-13T...",
  "version": "1.0.0",
  "compatibility_layer": {
    "name": "client_compatibility_layer",
    "status": "healthy",
    "version": "1.0.0"
  }
}
HTTP Status: 200
```

**Actual Output**:
```
[Record your actual output here]
```

---

#### 1.3: Ready Endpoint Shows All Services

```bash
# Test ready endpoint
curl -s http://localhost:8080/ready | jq .
```

**Expected Output**:
```json
{
  "status": "ready",
  "services": {
    "http": true,
    "p2p_quic": true,
    "quic": false
  }
}
```

**Actual Output**:
```
[Record your actual output here]
```

---

#### 1.4: Metrics Endpoint Accessible

```bash
# Test metrics endpoint (Prometheus format)
curl -s http://localhost:9090/metrics | head -20

# Count metrics
curl -s http://localhost:9090/metrics | grep -c "^relay_"
```

**Expected Output**:
```
# HELP relay_connections_by_status_total Total number of connections by status
# TYPE relay_connections_by_status_total counter
relay_connections_by_status_total{status="accepted"} 0
...
(Should have 20+ relay_* metrics)
```

**Actual Output**:
```
[Record your actual output here]
```

---

### Test Suite 2: Zitadel JWKS Integration

**Note**: These tests require a running Zitadel instance. If Zitadel is not available, we'll use mock tests.

#### 2.1: JWKS Endpoint Discovery Works

```bash
# Check Zitadel configuration
cat configs/config.json | jq .auth.zitadel

# Test discovery endpoint (if Zitadel running)
ZITADEL_DOMAIN="https://auth.2gc.ru"  # Update with actual domain
curl -s "$ZITADEL_DOMAIN/.well-known/openid-configuration" | jq .jwks_uri
```

**Expected Output**:
```
"https://auth.2gc.ru/oauth/v2/keys"
```

**Actual Output**:
```
[Record your actual output here]
```

---

#### 2.2: JWKS Fetched Successfully

```bash
# Fetch JWKS
JWKS_URI="https://auth.2gc.ru/oauth/v2/keys"  # From discovery
curl -s "$JWKS_URI" | jq .

# Check for keys
curl -s "$JWKS_URI" | jq '.keys | length'
```

**Expected Output**:
```json
{
  "keys": [
    {
      "kty": "RSA",
      "use": "sig",
      "kid": "...",
      "alg": "RS256",
      "n": "...",
      "e": "AQAB"
    }
  ]
}
Keys count: 1 or more
```

**Actual Output**:
```
[Record your actual output here]
```

---

#### 2.5: JWT Token Validation with JWKS

**Manual Test** (requires valid JWT token):

```bash
# Get a test JWT token from Zitadel
# This would normally be done through Zitadel login flow
TEST_TOKEN="eyJ..."  # Replace with actual token

# Test API with token
curl -s http://localhost:8080/api/v1/tenants/test-tenant/peers/register \
  -H "Authorization: Bearer $TEST_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "peer_id": "test-peer-1",
    "public_key": "test-key"
  }' | jq .
```

**Expected Output**:
```json
{
  "success": true,
  "peer_id": "test-peer-1",
  "message": "Peer registered successfully"
}
```

**Actual Output**:
```
[Record your actual output here]
```

---

### Test Suite 3: API Compatibility Layer

#### 3.1: Legacy Path - Peer Registration

```bash
# Test legacy path: POST /api/v1/tenants/:tid/peers/register
curl -s -X POST http://localhost:8080/api/v1/tenants/test-tenant/peers/register \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TEST_TOKEN" \
  -d '{
    "peer_id": "peer-legacy-1",
    "public_key": "legacy-test-key",
    "endpoint": "192.168.1.100:51820"
  }' | jq .

# Check relay logs for compatibility layer message
grep "Client compatibility request" logs/relay-startup.log | tail -1
```

**Expected Output**:
```json
{
  "success": true,
  "peer_id": "peer-legacy-1"
}
```

**Log Expected**:
```
INFO Client compatibility request method=POST path=/api/v1/tenants/test-tenant/peers/register
```

**Actual Output**:
```
[Record your actual output here]
```

---

#### 3.2: Legacy Path - Peer Discovery

```bash
# Test legacy path: GET /api/v1/tenants/:tid/peers
curl -s http://localhost:8080/api/v1/tenants/test-tenant/peers \
  -H "Authorization: Bearer $TEST_TOKEN" | jq .
```

**Expected Output**:
```json
{
  "success": true,
  "peers": [
    {
      "peer_id": "peer-legacy-1",
      "status": "online",
      "last_seen": "2025-10-13T..."
    }
  ],
  "total": 1
}
```

**Actual Output**:
```
[Record your actual output here]
```

---

#### 3.6: tenant_id Extracted to Context

**Check Logs**:

```bash
# Should see tenant_id in debug logs
grep "tenant_id" logs/relay-startup.log | grep -i "compatibility\|translation" | tail -5
```

**Expected Output**:
```
DEBUG API path translation legacy_path=/api/v1/tenants/test-tenant/peers/register tenant_id=test-tenant
```

**Actual Output**:
```
[Record your actual output here]
```

---

#### 3.7: Compatibility Metrics Collected

```bash
# Check metrics for compatibility layer
curl -s http://localhost:9090/metrics | grep -i "compatibility"

# Should see custom metrics if implemented
```

**Expected Output**:
```
(Metrics related to compatibility layer usage)
```

**Actual Output**:
```
[Record your actual output here]
```

---

### Test Suite 4: TURN Server

#### 4.1: TURN Server Deployment

```bash
# Deploy TURN server
./scripts/deploy-turn-server.sh

# Check deployment status
kubectl get pods -n cloudbridge -l app=turn-server
kubectl get svc -n cloudbridge -l app=turn-server
```

**Expected Output**:
```
NAME                           READY   STATUS    RESTARTS   AGE
turn-server-7b8f9d4c5-abc12   1/1     Running   0          30s
turn-server-7b8f9d4c5-def34   1/1     Running   0          30s

NAME          TYPE           CLUSTER-IP      EXTERNAL-IP     PORT(S)
turn-server   LoadBalancer   10.96.123.45    203.0.113.10    3478:30001/UDP,3478:30002/TCP...
```

**Actual Output**:
```
[Record your actual output here]
```

---

#### 4.3: LoadBalancer External IP

```bash
# Get external IP
TURN_EXTERNAL_IP=$(kubectl get svc turn-server -n cloudbridge -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
echo "TURN External IP: $TURN_EXTERNAL_IP"

# Update config if needed
# (Script should do this automatically)
```

**Expected Output**:
```
TURN External IP: 203.0.113.10 (actual cloud provider IP)
```

**Actual Output**:
```
[Record your actual output here]
```

---

#### 4.4-4.6: TURN Port Accessibility

```bash
# Test UDP port 3478
nc -vzu $TURN_EXTERNAL_IP 3478 && echo "✅ UDP 3478 open" || echo "❌ UDP 3478 closed"

# Test TCP port 3478
nc -vz $TURN_EXTERNAL_IP 3478 && echo "✅ TCP 3478 open" || echo "❌ TCP 3478 closed"

# Test TLS port 5349
openssl s_client -connect $TURN_EXTERNAL_IP:5349 < /dev/null 2>&1 | grep "CONNECTED"
```

**Expected Output**:
```
✅ UDP 3478 open
✅ TCP 3478 open
CONNECTED(00000003)
```

**Actual Output**:
```
[Record your actual output here]
```

---

#### 4.7: Prometheus Metrics Endpoint

```bash
# Port-forward to metrics endpoint
kubectl port-forward -n cloudbridge svc/turn-server 9641:9641 &
PF_PID=$!

sleep 2

# Fetch metrics
curl -s http://localhost:9641/metrics | head -20

# Cleanup
kill $PF_PID
```

**Expected Output**:
```
# HELP turn_allocations_total Total number of TURN allocations
# TYPE turn_allocations_total counter
turn_allocations_total 0
...
```

**Actual Output**:
```
[Record your actual output here]
```

---

### Test Suite 5: NAT Traversal Integration

**Note**: These tests require two peers to test actual NAT traversal. For manual testing, we'll verify the components work individually.

#### 5.1: ICE Candidates Generation

```bash
# Test ICE credentials endpoint
curl -s -X POST http://localhost:8080/api/v1/tenants/test-tenant/ice/credentials \
  -H "Authorization: Bearer $TEST_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "peer_id": "test-peer-1"
  }' | jq .
```

**Expected Output**:
```json
{
  "success": true,
  "ice_servers": [
    {
      "urls": ["turn:203.0.113.10:3478"],
      "username": "cloudbridge",
      "credential": "..."
    },
    {
      "urls": ["stun:stun.2gc.ru:3478"]
    }
  ],
  "ttl": 86400
}
```

**Actual Output**:
```
[Record your actual output here]
```

---

#### 5.2: STUN Binding Requests

**Manual test using `stunclient` (if available)**:

```bash
# Test STUN server
# Install: apt-get install stun-client
stunclient stun.2gc.ru 3478
```

**Expected Output**:
```
Binding test: success
Local address: 192.168.1.100:12345
Mapped address: 203.0.113.50:12345
```

**Actual Output**:
```
[Record your actual output here or skip if tool not available]
```

---

### Test Suite 6: Multi-Tenancy & Isolation

#### 6.1-6.2: Multiple Tenants and Isolation

```bash
# Register peer for tenant A
curl -s -X POST http://localhost:8080/api/v1/tenants/tenant-a/peers/register \
  -H "Content-Type: application/json" \
  -d '{"peer_id": "peer-a1", "public_key": "key-a1"}' | jq .

# Register peer for tenant B
curl -s -X POST http://localhost:8080/api/v1/tenants/tenant-b/peers/register \
  -H "Content-Type: application/json" \
  -d '{"peer_id": "peer-b1", "public_key": "key-b1"}' | jq .

# Try to access tenant-b peer from tenant-a context (should fail or be isolated)
curl -s http://localhost:8080/api/v1/tenants/tenant-a/peers/peer-b1 \
  -H "Authorization: Bearer $TENANT_A_TOKEN" | jq .
```

**Expected Output**:
```json
{
  "success": false,
  "error": "Peer not found in tenant scope"
}
```

**Actual Output**:
```
[Record your actual output here]
```

---

### Test Suite 7: Monitoring & Observability

#### 7.1: Prometheus Metrics Scraped

```bash
# Get all relay metrics
curl -s http://localhost:9090/metrics | grep "^relay_" | head -20

# Count unique metric names
curl -s http://localhost:9090/metrics | grep "^relay_" | cut -d'{' -f1 | sort -u | wc -l
```

**Expected Output**:
```
relay_connections_by_status_total{status="accepted"} 5
relay_connections_by_status_total{status="active"} 3
...
(Should have 15-25 unique relay_* metrics)
```

**Actual Output**:
```
[Record your actual output here]
```

---

#### 7.2: Logs Structured and Readable

```bash
# Check log format (should be JSON or structured)
tail -20 logs/relay-startup.log | head -5

# Check for required fields
grep -o '"level":\|"msg":\|"timestamp":' logs/relay-startup.log | sort -u
```

**Expected Output**:
```
{"level":"info","msg":"Starting relay server","timestamp":"2025-10-13T..."}
...
"level":
"msg":
"timestamp":
```

**Actual Output**:
```
[Record your actual output here]
```

---

## 📊 Test Results Summary

### Overall Test Coverage

| Test Suite | Tests Total | Passed | Failed | Skipped | Pass Rate |
|------------|-------------|--------|--------|---------|-----------|
| 1. Service Startup | 6 | - | - | - | -% |
| 2. Zitadel JWKS | 8 | - | - | - | -% |
| 3. API Compatibility | 8 | - | - | - | -% |
| 4. TURN Server | 8 | - | - | - | -% |
| 5. NAT Traversal | 8 | - | - | - | -% |
| 6. Multi-Tenancy | 5 | - | - | - | -% |
| 7. Observability | 5 | - | - | - | -% |
| **TOTAL** | **48** | **-** | **-** | **-** | **-%** |

---

## 🐛 Bugs Found

### Critical Bugs
1. **Bug ID**: CRIT-001
   - **Description**: [Describe critical bug]
   - **Test Suite**: [Which test suite]
   - **Severity**: Critical
   - **Steps to Reproduce**: [Steps]
   - **Expected**: [Expected behavior]
   - **Actual**: [Actual behavior]
   - **Fix Required**: Yes/No

### High Priority Bugs
1. **Bug ID**: HIGH-001
   - **Description**: [Describe bug]
   - **Test Suite**: [Which test suite]
   - **Severity**: High
   - **Fix Required**: Yes/No

### Medium/Low Priority Issues
1. **Bug ID**: MED-001
   - **Description**: [Describe issue]
   - **Severity**: Medium
   - **Can Defer**: Yes/No

---

## ✅ Testing Completion Checklist

- [ ] All test suites executed
- [ ] Test results documented
- [ ] Bugs logged with severity
- [ ] Critical bugs have fix plan
- [ ] Non-critical bugs triaged
- [ ] Test summary updated
- [ ] Pass rate calculated
- [ ] Ready for bug fix phase

---

## 🚀 Next Steps After Testing

### If Pass Rate > 90%
✅ **Proceed to Production Deployment**
- Fix only critical/high bugs
- Document known issues
- Deploy to production
- Monitor closely

### If Pass Rate 70-90%
⚠️ **Bug Fix Phase Required**
- Fix critical bugs immediately
- Fix high priority bugs
- Re-test affected areas
- Then proceed to deployment

### If Pass Rate < 70%
❌ **Significant Issues Found**
- Full bug fix sprint needed
- Re-architecture may be required
- Full re-test after fixes
- Delay production deployment

---

**Testing Started**: [Date/Time]
**Testing Completed**: [Date/Time]
**Total Time**: [Hours]
**Tester**: Claude Code
**Final Status**: [PASS/FAIL/PARTIAL]
