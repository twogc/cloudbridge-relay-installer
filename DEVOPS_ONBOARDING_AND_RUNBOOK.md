## DevOps Onboarding and Runbook

### 1) Purpose
- **Goal**: Enable a DevOps engineer to quickly understand the CloudBridge Relay project, review docs/code, deploy it, and run tests.
- **Audience**: DevOps/SRE/Platform engineers supporting development and operations.

### 2) High-level Overview
- **Component**: CloudBridge Relay (P2P relay/coordination service built with Go + Gin).
- **Key features**:
  - Peer registration (relay-assisted mode supported)
  - Heartbeat/status updates
  - Peer discovery (with whitelist support from JWT claims)
  - Logging via zap
- **Auth**: JWT (HS256 fallback). Claims include `tenant_id`, `server_id/sub`, `connection_type`, `peer_whitelist`, etc.
- **State**: In-memory PeerStore (non-persistent; data resets on pod restart).
- **Packaging**: Docker image; deployed to Kubernetes (tested on Minikube).

### 3) Repository Layout (important paths)
- `cmd/relay/` — service entrypoint
- `internal/api/` — Gin handlers, router, middleware
- `internal/token/` — P2P JWT claims and validation
- `internal/keycloak/` — optional token validation helpers
- `configs/` — runtime configuration (e.g., `config.json`)
- `scripts/` — helper scripts (`test-discovery-bypass.sh`, `generate-test-jwt.go`)
- `Dockerfile` / `Dockerfile.quick` — images
- `documentation/` — aux docs

### 4) Documentation to Review
- Root reports: `CRITICAL_P2P_API_FIX_REPORT.md`, `CRITICAL_P2P_API_FIXED_FINAL.md`, `DOCUMENTATION_CLEANUP_REPORT.md`, `DISCOVERY_ISSUE_ANALYSIS.md`, `QUICK_FIX_ATTEMPT_REPORT.md`.
- Source docs (read in this order):
  1. `internal/api/peer_handlers.go` (Register, Heartbeat, Discover)
  2. `internal/api/router.go` (routes)
  3. `internal/token/p2p_claims.go` (claims + validation)
  4. `internal/token/validator.go` (ValidateP2PToken)
  5. `internal/api/auth_middleware.go` (Authorization header parsing)

### 5) Prerequisites
- Linux/macOS with:
  - Go 1.21+ (local builds)
  - Docker (for images)
  - kubectl + Minikube (local k8s)
  - jq, curl, make (optional but helpful)
- Network access to pull base images (or use `Dockerfile.quick`).
- JWT secret configured (Kubernetes Secret or env var) and TLS where applicable.

### 6) Local Build and Run (without Docker)
```bash
# From repo root
go build -o bin/relay ./cmd/relay

# Example run (adjust config path if needed)
JWT_SECRET="<your-hs256-secret>" \
PORT=8080 \
./bin/relay -config ./configs/config.json

# Health check
curl -s http://localhost:8080/health | jq .
```

### 7) Docker Build
- Standard (may need network to fetch Alpine pkgs):
```bash
docker build -t cloudbridge-relay:local .
```
- Fast path (uses prebuilt local binary):
```bash
go build -o relay-server-test ./cmd/relay
docker build -f Dockerfile.quick -t cloudbridge-relay:quick .
```

### 8) Minikube Setup and Image Use
```bash
# Start Minikube
minikube start

# Point Docker to Minikube's daemon so images are available to k8s
eval "$(minikube docker-env)"

# Build an image inside Minikube Docker
go build -o relay-server-test ./cmd/relay
docker build -f Dockerfile.quick -t cloudbridge-relay:quick .
```

### 9) Kubernetes Deployment (example)
- Ensure a namespace:
```bash
kubectl create namespace cloudbridge --dry-run=client -o yaml | kubectl apply -f -
```
- Create secret with JWT secret (HS256 fallback):
```bash
kubectl -n cloudbridge create secret generic cloudbridge-secrets \
  --from-literal=RELAY_JWT_SECRET='eozy96a8+j125pOpIhCyytge1rR0MTiG4wBi/J9zpew=' \
  --dry-run=client -o yaml | kubectl apply -f -
```
- Apply Deployment/Service (adjust image/tag and ports as needed):
```bash
cat <<'YAML' | kubectl -n cloudbridge apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cloudbridge-relay
spec:
  replicas: 3
  selector:
    matchLabels:
      app.kubernetes.io/name: cloudbridge-relay
  template:
    metadata:
      labels:
        app.kubernetes.io/name: cloudbridge-relay
    spec:
      containers:
        - name: cloudbridge-relay
          image: cloudbridge-relay:quick
          imagePullPolicy: IfNotPresent
          env:
            - name: RELAY_JWT_SECRET
              valueFrom:
                secretKeyRef:
                  name: cloudbridge-secrets
                  key: RELAY_JWT_SECRET
          ports:
            - containerPort: 8080
          readinessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 10
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 10
            periodSeconds: 20
---
apiVersion: v1
kind: Service
metadata:
  name: cloudbridge-relay
spec:
  selector:
    app.kubernetes.io/name: cloudbridge-relay
  ports:
    - name: http
      port: 8080
      targetPort: 8080
      protocol: TCP
  type: ClusterIP
YAML
```

### 10) Port-Forward for Local Testing
```bash
kubectl -n cloudbridge port-forward deploy/cloudbridge-relay 8080:8080
# Then use http://localhost:8080
```

### 11) Testing the API
- Health:
```bash
curl -s http://localhost:8080/health | jq .
```
- Generate a valid test JWT with corrected whitelist:
```bash
go run scripts/generate-test-jwt.go
# Copy the printed token into $JWT
JWT="<pasted token>"
```
- Register peer:
```bash
curl -k -s -X POST "http://localhost:8080/api/v1/tenants/<tenant_id>/peers/register" \
  -H "Authorization: Bearer $JWT" \
  -H "Content-Type: application/json" \
  -d '{"public_key":"test-$(date +%s)","allowed_ips":["10.0.0.0/24"]}' | jq .
```
- Heartbeat:
```bash
curl -k -s -X PUT "http://localhost:8080/api/v1/tenants/<tenant_id>/peers/<peer_id>/status" \
  -H "Authorization: Bearer $JWT" \
  -H "Content-Type: application/json" \
  -d '{"relay_session_id":"<from_registration>","status":"online"}' | jq .
```
- Discovery:
```bash
curl -k -s "http://localhost:8080/api/v1/tenants/<tenant_id>/peers/discover" \
  -H "Authorization: Bearer $JWT" | jq .
```

### 12) Common Issues and Fixes
- 401 Invalid token/authorization format
  - Ensure header is exactly `Authorization: Bearer <token>` (case-sensitive on `Bearer ` prefix in code).
  - Ensure server’s HS256 secret matches token signing secret (`RELAY_JWT_SECRET`).
- 404 on heartbeat
  - Use `PUT` method; ensure `<peer_id>` matches stored peer ID.
  - Auto-registration on heartbeat exists, but initial registration recommended.
- Discovery returns `peers: null`
  - Likely whitelist mismatch in JWT. Include your peer ID in `peer_whitelist`.
  - Debug fallback exists to log/inspect registered peers.
- ImagePullBackOff
  - Use `eval "$(minikube docker-env)"` and build image locally with `IfNotPresent`.
- CrashLoopBackOff with quick image
  - Ensure binary is copied and executable in the image (`Dockerfile.quick` adds `chmod +x`).

### 13) Logs and Diagnostics
```bash
kubectl -n cloudbridge get pods
kubectl -n cloudbridge logs deploy/cloudbridge-relay -c cloudbridge-relay --tail=200
kubectl -n cloudbridge describe deploy/cloudbridge-relay
```

### 14) Configuration
- Env vars:
  - `RELAY_JWT_SECRET` — HS256 fallback secret
  - `PORT` — API port (default 8080)
- Config file: `configs/config.json` (network and runtime options if used).

### 15) CI/CD (baseline recommendations)
- Build-and-push image to registry on main merges
- Template manifests via Helm/Kustomize with env-specific values
- Smoke tests: health, register, heartbeat, discover
- Gate deploy if smoke tests fail

### 16) Security Notes
- Rotate `RELAY_JWT_SECRET` regularly; store only in secrets manager
- Prefer asymmetric JWT validation (JWKS) in production
- Limit service exposure; use mTLS or private networks where possible

### 17) Next Steps for DevOps
- Read the docs in section 4, skim code paths in section 3
- Stand up local Minikube and deploy via sections 8–11
- Validate with provided scripts in `scripts/`
- Report any discrepancies or missing instructions back to the team


