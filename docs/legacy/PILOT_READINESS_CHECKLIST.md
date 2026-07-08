# CloudBridge Relay - Pilot Readiness Checklist

**Дата создания:** 19 октября 2025
**Последнее обновление:** 19 октября 2025
**Текущая готовность:** 85% → Цель: 100%
**Цель:** Подготовка к пилотному проекту

---

## ОБЩИЙ СТАТУС

| Фаза | Задач | Выполнено | Прогресс | Срок |
|------|-------|-----------|----------|------|
| **Фаза 1: Критические исправления** | 8 | 0 | 0% | 1-2 дня |
| **Фаза 2: Production Hardening** | 6 | 0 | 0% | 3-5 дней |
| **Фаза 3: Maintainability** | 5 | 0 | 0% | 2-3 дня |
| **ИТОГО** | 19 | 0 | 0% | 6-10 дней |

---

# 🔴 ФАЗА 1: КРИТИЧЕСКИЕ ИСПРАВЛЕНИЯ (1-2 дня)

**Цель:** Устранить блокеры развертывания
**Критерий успеха:** Все pods запускаются без ошибок

---

## Task 1.1: Create Missing PersistentVolumeClaim

**Приоритет:** 🔴 CRITICAL
**Время:** 30 минут
**Ответственный:** _________
**Статус:** ⬜ TODO

### Проблема
- Deployment `cloudbridge-relay-deployment.yaml:197` ссылается на `cloudbridge-logs-pvc`
- PVC не существует → Pod не запустится
- Error: `persistentvolumeclaim "cloudbridge-logs-pvc" not found`

### Решение

**1. Создать файл PVC:**
```bash
touch /Users/maxlanies/Git/2GC/cloudbridge-relay-installer/k8s/persistentvolumeclaims/cloudbridge-logs-pvc.yaml
```

**2. Добавить содержимое:**
```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: cloudbridge-logs-pvc
  namespace: cloudbridge
  labels:
    app.kubernetes.io/name: cloudbridge-relay
    app.kubernetes.io/component: logging
    app.kubernetes.io/part-of: cloudbridge-relay
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: fast-ssd  # или standard для dev
  resources:
    requests:
      storage: 10Gi
```

**3. Проверить deployment:**
```bash
kubectl apply -f k8s/persistentvolumeclaims/cloudbridge-logs-pvc.yaml --dry-run=client -o yaml
```

**4. Применить изменения:**
```bash
kubectl apply -f k8s/persistentvolumeclaims/cloudbridge-logs-pvc.yaml
```

**5. Проверить статус:**
```bash
kubectl get pvc cloudbridge-logs-pvc -n cloudbridge
# Ожидаемый статус: Bound
```

### Критерии приемки
- [ ] Файл `k8s/persistentvolumeclaims/cloudbridge-logs-pvc.yaml` создан
- [ ] PVC применен в кластер
- [ ] Status: Bound
- [ ] cloudbridge-relay pod успешно монтирует volume
- [ ] No error logs related to PVC

### Проверка
```bash
# Проверить монтирование
kubectl describe pod -n cloudbridge -l app.kubernetes.io/name=cloudbridge-relay | grep -A 5 "Mounts:"

# Проверить что нет ошибок
kubectl logs -n cloudbridge -l app.kubernetes.io/name=cloudbridge-relay --tail=50 | grep -i error
```

**Результат:** od запускается, logs записываются в PVC

---

## Task 1.2: Remove Hardcoded Secrets from P2P Deployment

**Приоритет:** 🔴 CRITICAL (Security)
**Время:** 2 часа
**Ответственный:** _________
**Статус:** ⬜ TODO

### Проблема
- Файл: `k8s/deployments/cloudbridge-relay-p2p-deployment.yaml:58-63`
- JWT secrets в plain text:
  ```yaml
  - name: JWT_SECRET
    value: "85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE="
  - name: ZITADEL_CLIENT_SECRET
    value: "your-client-secret-here"
  ```
- **Risk:** Secrets exposed в Git, version control history

### Решение

**Вариант A: Kubernetes Secrets (Quick)**

**1. Создать Secret:**
```bash
# Генерировать новый JWT secret
JWT_SECRET=$(openssl rand -base64 32)
ZITADEL_CLIENT_SECRET="<получить из Zitadel admin UI>"

kubectl create secret generic cloudbridge-relay-secrets \
  -n cloudbridge \
  --from-literal=jwt-secret="$JWT_SECRET" \
  --from-literal=zitadel-client-secret="$ZITADEL_CLIENT_SECRET" \
  --from-literal=turn-password="<TURN_PASSWORD>" \
  --dry-run=client -o yaml > k8s/secrets/cloudbridge-relay-secrets.yaml
```

**2. Зашифровать Secret (опционально):**
```bash
# Использовать kubeseal для sealed secrets
kubeseal < k8s/secrets/cloudbridge-relay-secrets.yaml \
  > k8s/secrets/cloudbridge-relay-secrets-sealed.yaml
```

**3. Обновить deployment:**
```yaml
# k8s/deployments/cloudbridge-relay-p2p-deployment.yaml
env:
  - name: JWT_SECRET
    valueFrom:
      secretKeyRef:
        name: cloudbridge-relay-secrets
        key: jwt-secret
  - name: ZITADEL_CLIENT_SECRET
    valueFrom:
      secretKeyRef:
        name: cloudbridge-relay-secrets
        key: zitadel-client-secret
```

**Вариант B: External Secrets Operator (Recommended)**

**1. Проверить что External Secrets установлен:**
```bash
kubectl get pods -n external-secrets-system
```

**2. Создать ExternalSecret:**
```yaml
# k8s/security/external-secrets/cloudbridge-relay-external-secret.yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: cloudbridge-relay-secrets
  namespace: cloudbridge
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: vault-backend  # или другой provider
    kind: ClusterSecretStore
  target:
    name: cloudbridge-relay-secrets
    creationPolicy: Owner
  data:
    - secretKey: jwt-secret
      remoteRef:
        key: cloudbridge/jwt-secret
    - secretKey: zitadel-client-secret
      remoteRef:
        key: cloudbridge/zitadel-client-secret
```

**3. Применить ExternalSecret:**
```bash
kubectl apply -f k8s/security/external-secrets/cloudbridge-relay-external-secret.yaml
```

### Критерии приемки
- [ ] Secrets удалены из YAML files
- [ ] Kubernetes Secret создан (или ExternalSecret)
- [ ] Deployment обновлен для использования secretKeyRef
- [ ] Pod успешно читает secrets
- [ ] Secrets не видны в `git log` (history cleaned if needed)
- [ ] `.gitignore` обновлен для исключения `k8s/secrets/*.yaml`

### Проверка
```bash
# Проверить что secrets НЕ в plain text
git log -p k8s/deployments/cloudbridge-relay-p2p-deployment.yaml | grep -i "jwt_secret"

# Проверить что Secret существует
kubectl get secret cloudbridge-relay-secrets -n cloudbridge -o yaml

# Проверить что pod читает secrets
kubectl exec -n cloudbridge deployment/cloudbridge-relay-p2p -- env | grep JWT_SECRET
```

**Результат:** o secrets in Git, все secrets в Kubernetes Secrets/Vault

---

## 🐛 Task 1.3: Fix Invalid Image Tag with Shell Expansion

**Приоритет:** 🔴 CRITICAL
**Время:** 15 минут
**Ответственный:** _________
**Статус:** ⬜ TODO

### Проблема
- Файл: `k8s/deployments/cloudbridge-relay-p2p-deployment.yaml:29`
- Invalid image tag:
  ```yaml
  image: cloudbridge-relay:enhanced-$(date +%Y%m%d-%H%M%S)
  ```
- Shell expansion не работает в Kubernetes YAML
- Error: `ImagePullBackOff`, image not found

### Решение

**1. Определить стратегию tagging:**

**Опция A: Fixed version tag (Recommended)**
```yaml
image: cr.yandex/crpblcmklr11njvr6ikg/cloudbridge-relay:v1.0.0-p2p
```

**Опция B: Git commit SHA**
```yaml
image: cr.yandex/crpblcmklr11njvr6ikg/cloudbridge-relay:$(git rev-parse --short HEAD)
# Заменить через CI/CD pipeline
```

**Опция C: Environment variable**
```yaml
# В deployment
image: cr.yandex/crpblcmklr11njvr6ikg/cloudbridge-relay:${IMAGE_TAG}

# При deploy
IMAGE_TAG=v1.0.0 envsubst < deployment.yaml | kubectl apply -f -
```

**2. Обновить deployment:**
```bash
# Edit file
vim k8s/deployments/cloudbridge-relay-p2p-deployment.yaml

# Заменить line 29:
# image: cloudbridge-relay:enhanced-$(date +%Y%m%d-%H%M%S)
# →
# image: cr.yandex/crpblcmklr11njvr6ikg/cloudbridge-relay:v1.0.0-p2p
```

**3. Проверить что image существует:**
```bash
# Если используете Yandex Container Registry
yc container image list --registry-id crpblcmklr11njvr6ikg

# Если локальный registry
curl -X GET http://localhost:5000/v2/cloudbridge-relay/tags/list
```

**4. Build and push correct image:**
```bash
# Build image
docker build -t cr.yandex/crpblcmklr11njvr6ikg/cloudbridge-relay:v1.0.0-p2p \
  -f docker/Dockerfile .

# Push to registry
docker push cr.yandex/crpblcmklr11njvr6ikg/cloudbridge-relay:v1.0.0-p2p
```

### Критерии приемки
- [ ] Invalid shell expansion удален из YAML
- [ ] Image tag указывает на существующий image
- [ ] Image доступен в registry
- [ ] Pod успешно pulls image
- [ ] No ImagePullBackOff errors

### Проверка
```bash
# Проверить deployment
kubectl describe deployment cloudbridge-relay-p2p -n cloudbridge | grep Image:

# Проверить pod
kubectl get pods -n cloudbridge -l app=cloudbridge-relay-p2p
# Status должен быть Running, не ImagePullBackOff

# Проверить events
kubectl get events -n cloudbridge --sort-by='.lastTimestamp' | grep -i image
```

**Результат:** od запускается с правильным image

---

## 💥 Task 1.4: Fix Failing Zitadel Integration Test

**Приоритет:** 🔴 CRITICAL
**Время:** 1 час
**Ответственный:** _________
**Статус:** ⬜ TODO

### Проблема
- Файл: `cmd/relay/comprehensive_integration_test.go:108`
- Test: `TestComprehensiveZitadelIntegration`
- Error:
  ```
  panic: runtime error: invalid memory address or nil pointer dereference
  Error: zitadel client not initialized
  ```
- Root cause: Zitadel client не инициализирован в test setup

### Решение

**1. Проверить текущий test:**
```bash
cd /Users/maxlanies/Git/2GC/cloudbridge-relay-installer
go test -v ./cmd/relay -run TestComprehensiveZitadelIntegration
```

**2. Открыть test file:**
```bash
vim cmd/relay/comprehensive_integration_test.go
```

**3. Добавить mock Zitadel client:**
```go
// cmd/relay/comprehensive_integration_test.go

func TestComprehensiveZitadelIntegration(t *testing.T) {
    logger := zaptest.NewLogger(t)

    // ДОБАВИТЬ: Mock Zitadel client initialization
    mockConfig := &zitadel.Config{
        Issuer:       "https://zitadel.local",
        ClientID:     "test-client-id",
        ClientSecret: "test-client-secret",
    }

    zitadelClient, err := zitadel.NewClient(mockConfig, logger)
    if err != nil {
        t.Fatalf("Failed to create mock Zitadel client: %v", err)
    }

    // ДОБАВИТЬ: Initialize tenant manager with client
    tenantMgr := tenant.NewManager(logger)
    tenantMgr.SetZitadelClient(zitadelClient)  // Ensure client is set

    // Rest of test...
    tenant, err := tenantMgr.CreateTenant("test-tenant", "test.local")
    require.NoError(t, err)
    require.NotNil(t, tenant)
}
```

**4. Или: Skip test если Zitadel не доступен:**
```go
func TestComprehensiveZitadelIntegration(t *testing.T) {
    // ДОБАВИТЬ: Skip if no Zitadel endpoint
    if os.Getenv("ZITADEL_ENDPOINT") == "" {
        t.Skip("Skipping Zitadel integration test: ZITADEL_ENDPOINT not set")
    }

    // ... rest of test
}
```

**5. Запустить test снова:**
```bash
go test -v ./cmd/relay -run TestComprehensiveZitadelIntegration
```

### Критерии приемки
- [ ] Test `TestComprehensiveZitadelIntegration` проходит без panic
- [ ] Zitadel client корректно инициализирован в test
- [ ] No nil pointer dereference
- [ ] All integration tests pass: `go test ./cmd/relay -v`

### Проверка
```bash
# Run specific test
go test -v ./cmd/relay -run TestComprehensiveZitadelIntegration

# Run all relay tests
go test -v ./cmd/relay

# Check for any remaining failures
go test ./... 2>&1 | grep -i fail
```

**Результат:** ll tests pass, no panics

---

## ⚕️ Task 1.5: Add Health Checks to cadence-web-deployment

**Приоритет:** 🟠 HIGH
**Время:** 30 минут
**Ответственный:** _________
**Статус:** ⬜ TODO

### Проблема
- Файл: `k8s/deployments/cadence-web-deployment.yaml`
- No readiness/liveness probes defined
- Kubernetes не может определить pod health
- Traffic может быть отправлен на unhealthy pod

### Решение

**1. Определить health endpoints:**
```bash
# Проверить какие endpoints доступны в cadence-web
# Обычно: /, /health, /ready
```

**2. Обновить deployment:**
```yaml
# k8s/deployments/cadence-web-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cadence-web
spec:
  template:
    spec:
      containers:
      - name: cadence-web
        image: ubercadence/web:v3.30.2  # Pin version!
        ports:
        - containerPort: 8088
          name: http

        # ДОБАВИТЬ: Readiness probe
        readinessProbe:
          httpGet:
            path: /
            port: 8088
            scheme: HTTP
          initialDelaySeconds: 10
          periodSeconds: 10
          timeoutSeconds: 5
          successThreshold: 1
          failureThreshold: 3

        # ДОБАВИТЬ: Liveness probe
        livenessProbe:
          httpGet:
            path: /
            port: 8088
            scheme: HTTP
          initialDelaySeconds: 30
          periodSeconds: 15
          timeoutSeconds: 5
          successThreshold: 1
          failureThreshold: 3

        # ДОБАВИТЬ: Startup probe (optional)
        startupProbe:
          httpGet:
            path: /
            port: 8088
          initialDelaySeconds: 0
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 30  # 5 min max startup time
```

**3. Применить изменения:**
```bash
kubectl apply -f k8s/deployments/cadence-web-deployment.yaml
```

### Критерии приемки
- [ ] Readiness probe добавлен
- [ ] Liveness probe добавлен
- [ ] Probes используют правильный endpoint
- [ ] Pod переходит в Ready state
- [ ] Kubernetes отмечает pod как healthy

### Проверка
```bash
# Проверить probe configuration
kubectl describe deployment cadence-web -n cloudbridge | grep -A 10 "Liveness\|Readiness"

# Проверить pod status
kubectl get pods -n cloudbridge -l app=cadence-web
# Status должен быть Running с READY 1/1

# Проверить probe failures
kubectl get events -n cloudbridge | grep -i "liveness\|readiness"
```

**Результат:** adence-web pod healthy с working probes

---

## 🔓 Task 1.6: Uncomment Health Checks in P2P Deployment

**Приоритет:** 🟠 HIGH
**Время:** 15 минут
**Ответственный:** _________
**Статус:** ⬜ TODO

### Проблема
- Файл: `k8s/deployments/cloudbridge-relay-p2p-deployment.yaml:81`
- Health checks закомментированы:
  ```yaml
  # livenessProbe:
  #   httpGet:
  #     path: /health
  #     port: 8080
  ```
- Kubernetes не может определить pod health

### Решение

**1. Открыть файл:**
```bash
vim k8s/deployments/cloudbridge-relay-p2p-deployment.yaml
```

**2. Раскомментировать health checks:**
```yaml
# Line ~81-95
livenessProbe:
  httpGet:
    path: /health
    port: 8080
    scheme: HTTP
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /ready
    port: 8080
    scheme: HTTP
  initialDelaySeconds: 10
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 3
```

**3. Проверить что endpoints существуют:**
```bash
# После deployment
kubectl port-forward -n cloudbridge deployment/cloudbridge-relay-p2p 8080:8080

# В другом терминале
curl http://localhost:8080/health
curl http://localhost:8080/ready
```

**4. Применить изменения:**
```bash
kubectl apply -f k8s/deployments/cloudbridge-relay-p2p-deployment.yaml
```

### Критерии приемки
- [ ] Health checks раскомментированы
- [ ] Endpoints `/health` и `/ready` возвращают 200 OK
- [ ] Pod переходит в Ready state
- [ ] No probe failures в events

### Проверка
```bash
# Проверить configuration
kubectl describe deployment cloudbridge-relay-p2p -n cloudbridge | grep -A 5 "Liveness\|Readiness"

# Проверить pod
kubectl get pods -n cloudbridge -l app=cloudbridge-relay-p2p
# READY должен быть 1/1

# Проверить logs
kubectl logs -n cloudbridge deployment/cloudbridge-relay-p2p | grep -i "health"
```

**Результат:** 2P deployment имеет working health checks

---

##  Task 1.7: Fix Port Conflicts

**Приоритет:** 🟡 MEDIUM
**Время:** 1 час
**Ответственный:** _________
**Статус:** ⬜ TODO

### Проблема
- **Conflict 1:** Port 9090 используется для:
  - CloudBridge relay HTTP API
  - Prometheus metrics endpoint
- **Conflict 2:** Port 7934 используется для:
  - Cadence web UI
  - Prometheus metrics

### Решение

**1. Проанализировать текущее использование:**
```bash
# Найти все использования port 9090
grep -r "9090" k8s/deployments/ k8s/services/

# Найти все использования port 7934
grep -r "7934" k8s/deployments/ k8s/services/
```

**2. Разделить ports:**

**Для CloudBridge Relay:**
```yaml
# k8s/deployments/cloudbridge-relay-deployment.yaml
ports:
  - containerPort: 8080
    name: http-api
    protocol: TCP
  - containerPort: 9090
    name: quic  # Keep for QUIC
    protocol: UDP
  - containerPort: 9091  # CHANGE: New metrics port
    name: metrics
    protocol: TCP
```

**Для Cadence:**
```yaml
# k8s/deployments/cadence-web-deployment.yaml
ports:
  - containerPort: 8088  # Web UI
    name: http
  - containerPort: 9092  # CHANGE: Separate metrics port
    name: metrics
```

**3. Обновить Services:**
```yaml
# k8s/services/cloudbridge-services.yaml
---
apiVersion: v1
kind: Service
metadata:
  name: cloudbridge-relay-metrics
spec:
  ports:
  - port: 9091  # Changed
    targetPort: metrics
    name: metrics
```

**4. Обновить ServiceMonitor:**
```yaml
# k8s/monitoring/prometheus-servicemonitors.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: cloudbridge-relay
spec:
  endpoints:
  - port: metrics  # Will use port 9091
    path: /metrics
```

### Критерии приемки
- [ ] Port conflicts resolved
- [ ] Каждый service использует unique port
- [ ] Services успешно bind to ports
- [ ] Prometheus scraping работает
- [ ] No "address already in use" errors

### Проверка
```bash
# Проверить port bindings
kubectl get svc -n cloudbridge -o wide

# Проверить что Prometheus scraping работает
kubectl port-forward -n cloudbridge svc/prometheus-optimized 9090:9090
# Open http://localhost:9090/targets

# Проверить logs для port errors
kubectl logs -n cloudbridge deployment/cloudbridge-relay | grep -i "address already in use"
```

**Результат:** ll services на unique ports, no conflicts

---

## Task 1.8: Add Resource Limits to All Containers

**Приоритет:** 🟡 MEDIUM
**Время:** 2 часа
**Ответственный:** _________
**Статус:** ⬜ TODO

### Проблема
- Containers без resource limits:
  - `cadence-web-deployment.yaml` - no limits at all
  - Init containers в нескольких deployments
  - Sidecar `prometheus-proxy`
- Risk: OOM kills, resource starvation

### Решение

**1. Определить baseline resource requirements:**
```bash
# Проверить текущее использование
kubectl top pods -n cloudbridge
kubectl top nodes
```

**2. Обновить cadence-web-deployment:**
```yaml
# k8s/deployments/cadence-web-deployment.yaml
spec:
  containers:
  - name: cadence-web
    image: ubercadence/web:v3.30.2
    resources:
      requests:
        cpu: 100m
        memory: 128Mi
      limits:
        cpu: 500m
        memory: 512Mi
```

**3. Обновить init containers:**
```yaml
# Example: cloudbridge-relay-deployment.yaml
initContainers:
- name: wait-for-postgres
  image: postgres:14-alpine
  resources:
    requests:
      cpu: 50m
      memory: 64Mi
    limits:
      cpu: 100m
      memory: 128Mi
```

**4. Обновить sidecars:**
```yaml
# prometheus-proxy sidecar
- name: prometheus-proxy
  image: prom/prometheus:v2.45.0
  resources:
    requests:
      cpu: 50m
      memory: 64Mi
    limits:
      cpu: 200m
      memory: 256Mi
```

**5. Создать resource policy template:**
```yaml
# k8s/policies/resource-quotas.yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: cloudbridge-quota
  namespace: cloudbridge
spec:
  hard:
    requests.cpu: "10"
    requests.memory: 20Gi
    limits.cpu: "20"
    limits.memory: 40Gi
    persistentvolumeclaims: "10"
```

### Критерии приемки
- [ ] Все containers имеют resource requests
- [ ] Все containers имеют resource limits
- [ ] Limits установлены разумно (не слишком высоко/низко)
- [ ] No OOMKilled pods
- [ ] ResourceQuota создан для namespace

### Проверка
```bash
# Найти containers без limits
kubectl get deployments -n cloudbridge -o json | \
  jq '.items[] | select(.spec.template.spec.containers[].resources.limits == null) | .metadata.name'

# Проверить resource usage
kubectl top pods -n cloudbridge --containers

# Проверить ResourceQuota
kubectl describe resourcequota cloudbridge-quota -n cloudbridge
```

**Результат:** ll containers имеют resource limits, no OOM kills

---

## ФАЗА 1: ЧЕКЛИСТ ЗАВЕРШЕНИЯ

### Pre-Deployment Verification

- [ ] **Task 1.1:** PVC `cloudbridge-logs-pvc` создан и Bound
- [ ] **Task 1.2:** Secrets удалены из Git, используются K8s Secrets
- [ ] **Task 1.3:** Image tags исправлены, images pull успешно
- [ ] **Task 1.4:** Zitadel integration test проходит
- [ ] **Task 1.5:** Cadence-web имеет health checks
- [ ] **Task 1.6:** P2P deployment health checks раскомментированы
- [ ] **Task 1.7:** Port conflicts resolved
- [ ] **Task 1.8:** Resource limits добавлены всем containers

### Deployment Test

```bash
# 1. Apply all changes
kubectl apply -k k8s/base/

# 2. Check pod status
kubectl get pods -n cloudbridge

# Expected: All pods Running with READY 1/1 or 2/2

# 3. Check for errors
kubectl get events -n cloudbridge --sort-by='.lastTimestamp' | grep -i error

# 4. Verify services
kubectl get svc -n cloudbridge

# 5. Test health endpoints
kubectl port-forward -n cloudbridge svc/cloudbridge-relay 8080:8080
curl http://localhost:8080/health
curl http://localhost:8080/ready
```

### Success Criteria

 **Фаза 1 завершена когда:**

1. се 8 tasks выполнены
2. се pods в статусе Running
3. o ImagePullBackOff errors
4. o CrashLoopBackOff errors
5. ll health checks passing
6. ntegration tests проходят
7. o secrets в Git history
8. esource usage в пределах limits

### Rollback Plan

**Если что-то пошло не так:**

```bash
# 1. Rollback deployment
kubectl rollout undo deployment/cloudbridge-relay -n cloudbridge

# 2. Проверить previous revision
kubectl rollout history deployment/cloudbridge-relay -n cloudbridge

# 3. Restore from backup
kubectl apply -f k8s/deployments/cloudbridge-relay-deployment-backup-20251010-203045.yaml

# 4. Check logs
kubectl logs -n cloudbridge deployment/cloudbridge-relay --previous
```

---

## 📞 КОНТАКТЫ И ESCALATION

**Technical Lead:** _________
**DevOps Lead:** _________
**Security Lead:** _________

**Escalation Path:**
1. Developer → Technical Lead (30 min)
2. Technical Lead → DevOps Lead (1 hour)
3. DevOps Lead → CTO (2 hours)

---

## NOTES

- После завершения Фазы 1, обновить этот документ
- Commit changes: `git add . && git commit -m "Phase 1 complete"`
- Create tag: `git tag -a v1.0.0-phase1 -m "Phase 1: Critical fixes complete"`
- Proceed to Phase 2: Production Hardening

---

**Last Updated:** 2025-10-19
**Next Review:** После завершения Phase 1
**Document Owner:** CloudBridge DevOps Team
