# 🚀 **Production-Ready WAF Plan для CloudBridge Relay**

## **Обзор с учетом Production Redlines**

### **Реалистичные SLO и метрики:**
- **p95 end-to-end** (Ingress→Relay→Backend): 150–300 ms
- **p95 WAF-overhead**: ≤ 3–8 ms (внутри кластера)
- **FPR WAF** (ложные срабатывания): ≤ 0.5% в канареечных выкатах
- **Uptime**: 99.9% (при ≥3 репликах), 99.99% (при мультизоне)

### **Архитектура с L3/L4 защитой:**
```
Internet → L4 DDoS Protection (Cloud/ISP)
          → eBPF XDP (Node-level)
          → Traefik (L7 Ingress)
          + Coraza WAF (OWASP CRS)
          + CrowdSec (Bot Protection)
       → ORY Oathkeeper (Identity-aware proxy)
       → KrakenD (API Gateway)
       → CloudBridge Relay
       → Backend Services
```

##  **1. Реалистичные цели и SLO**

### **Метрики производительности:**
```yaml
# k8s/monitoring/slo-definitions.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: slo-definitions
  namespace: monitoring
data:
  slo.yaml: |
    service_level_objectives:
      # End-to-end latency (реалистично для интернета)
      end_to_end_latency:
        p95: "300ms"
        p99: "500ms"
        target: "150ms"
      
      # WAF overhead (внутри кластера)
      waf_overhead:
        p95: "8ms"
        p99: "15ms"
        target: "3ms"
      
      # Availability
      availability:
        target: "99.9%"  # 99.99% только при мультизоне
        measurement_window: "30d"
      
      # WAF False Positive Rate
      waf_false_positive_rate:
        target: "0.5%"
        measurement_window: "24h"
        canary_period: "72h"
      
      # Error rates
      error_rate:
        target: "0.1%"
        p95: "0.5%"
        measurement_window: "5m"
```

### **Prometheus Alerts:**
```yaml
# k8s/monitoring/alerts.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-alerts
  namespace: monitoring
data:
  alerts.yaml: |
    groups:
    - name: waf.rules
      rules:
      - alert: WAFBlockRateHigh
        expr: rate(coraza_waf_blocked_requests_total[5m]) / rate(traefik_http_requests_total[5m]) > 0.1
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "WAF block rate is high (possible FPR)"
          description: "WAF is blocking {{ $value | humanizePercentage }} of requests"
      
      - alert: WAFRulesEvalLatencyHigh
        expr: histogram_quantile(0.95, rate(coraza_waf_rule_eval_duration_seconds_bucket[5m])) > 0.02
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "WAF rule evaluation latency is high"
          description: "p95 WAF rule evaluation latency is {{ $value }}s"
      
      - alert: Upstream5xxSpike
        expr: rate(traefik_service_requests_total{code=~"5.."}[5m]) > 0.01
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Upstream 5xx error spike detected"
          description: "5xx error rate is {{ $value | humanizePercentage }}"
```

## 🛡️ **2. L3/L4 защита и eBPF**

### **Node-level eBPF protection:**
```yaml
# k8s/security/ebpf-daemonset.yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: ebpf-xdp-protection
  namespace: kube-system
spec:
  selector:
    matchLabels:
      app: ebpf-xdp-protection
  template:
    metadata:
      labels:
        app: ebpf-xdp-protection
    spec:
      hostNetwork: true
      containers:
      - name: ebpf-xdp
        image: cilium/ebpf-xdp:latest
        securityContext:
          privileged: true
          capabilities:
            add:
            - NET_ADMIN
            - SYS_ADMIN
        env:
        - name: XDP_MODE
          value: "native"
        - name: RATE_LIMIT
          value: "1000"  # packets per second per IP
        volumeMounts:
        - name: bpf-maps
          mountPath: /sys/fs/bpf
        - name: lib-modules
          mountPath: /lib/modules
        resources:
          requests:
            memory: "64Mi"
            cpu: "100m"
          limits:
            memory: "128Mi"
            cpu: "200m"
      volumes:
      - name: bpf-maps
        hostPath:
          path: /sys/fs/bpf
      - name: lib-modules
        hostPath:
          path: /lib/modules
      tolerations:
      - operator: Exists
```

### **L4 rate limiting:**
```yaml
# k8s/security/iptables-rules.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: iptables-rules
  namespace: kube-system
data:
  iptables.sh: |
    #!/bin/bash
    # L4 rate limiting rules
    iptables -A INPUT -p tcp --dport 80 -m limit --limit 1000/sec --limit-burst 2000 -j ACCEPT
    iptables -A INPUT -p tcp --dport 443 -m limit --limit 1000/sec --limit-burst 2000 -j ACCEPT
    iptables -A INPUT -p tcp --dport 80 -j DROP
    iptables -A INPUT -p tcp --dport 443 -j DROP
    
    # SYN flood protection
    iptables -A INPUT -p tcp --syn -m limit --limit 100/sec --limit-burst 200 -j ACCEPT
    iptables -A INPUT -p tcp --syn -j DROP
```

## 📝 **3. Policy-as-Code и канареечные выкаты**

### **Структура репозитория waf-policies:**
```
waf-policies/
├── .github/
│   ├── workflows/
│   │   ├── validate-policies.yml
│   │   ├── canary-deploy.yml
│   │   └── negative-tests.yml
│   └── PULL_REQUEST_TEMPLATE.md
├── policies/
│   ├── coraza/
│   │   ├── crs-overrides/
│   │   │   ├── staging/
│   │   │   ├── canary/
│   │   │   └── production/
│   │   └── custom-rules/
│   ├── oathkeeper/
│   │   ├── access-rules/
│   │   └── policies/
│   └── krakend/
│       ├── endpoints/
│       └── middleware/
├── tests/
│   ├── negative/
│   │   ├── bola-tests.yml
│   │   ├── ssrf-tests.yml
│   │   └── path-traversal-tests.yml
│   └── positive/
│       └── legitimate-traffic.yml
├── scripts/
│   ├── validate-policies.sh
│   ├── canary-deploy.sh
│   └── rollback.sh
└── docs/
    ├── policy-guidelines.md
    └── deployment-process.md
```

### **CI/CD Pipeline:**
```yaml
# .github/workflows/validate-policies.yml
name: Validate WAF Policies
on:
  pull_request:
    paths:
      - 'policies/**'
      - 'tests/**'

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Validate Coraza Rules
      run: |
        docker run --rm -v $PWD:/policies coraza/coraza-waf:latest \
          --config=/policies/policies/coraza/coraza.conf \
          --validate-only
    
    - name: Validate Oathkeeper Rules
      run: |
        docker run --rm -v $PWD:/policies ory/oathkeeper:latest \
          rules validate /policies/policies/oathkeeper/access-rules/
    
    - name: Run Negative Tests
      run: |
        ./scripts/run-negative-tests.sh
    
    - name: Check Policy Coverage
      run: |
        ./scripts/check-coverage.sh
```

### **Канареечный деплой:**
```yaml
# .github/workflows/canary-deploy.yml
name: Canary Deploy WAF Policies
on:
  push:
    branches: [main]
    paths:
      - 'policies/**'

jobs:
  canary:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Deploy to Staging
      run: |
        kubectl apply -f policies/coraza/crs-overrides/staging/ -n traefik-system
        kubectl apply -f policies/oathkeeper/access-rules/ -n oathkeeper-system
    
    - name: Wait for Staging Validation
      run: |
        ./scripts/wait-for-staging-validation.sh
    
    - name: Deploy Canary (5% traffic)
      run: |
        ./scripts/canary-deploy.sh --percentage=5
    
    - name: Monitor Canary Metrics
      run: |
        ./scripts/monitor-canary.sh --duration=30m
    
    - name: Deploy to Production (if canary successful)
      if: success()
      run: |
        ./scripts/full-deploy.sh
```

## 🔐 **4. Управление секретами**

### **External Secrets Operator:**
```yaml
# k8s/secrets/external-secrets.yaml
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: vault-secret-store
  namespace: oathkeeper-system
spec:
  provider:
    vault:
      server: "https://vault.cloudbridge.local"
      path: "secret"
      version: "v2"
      auth:
        kubernetes:
          mountPath: "kubernetes"
          role: "oathkeeper"
---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: zitadel-client-secret
  namespace: oathkeeper-system
spec:
  refreshInterval: "1h"
  secretStoreRef:
    name: vault-secret-store
    kind: SecretStore
  target:
    name: zitadel-client-secret
    creationPolicy: Owner
  data:
  - secretKey: client-secret
    remoteRef:
      key: "zitadel/cloudbridge-relay"
      property: "client_secret"
---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: jwks-cache-key
  namespace: oathkeeper-system
spec:
  refreshInterval: "15m"  # JWKS rotation
  secretStoreRef:
    name: vault-secret-store
    kind: SecretStore
  target:
    name: jwks-cache-key
    creationPolicy: Owner
  data:
  - secretKey: cache-key
    remoteRef:
      key: "zitadel/jwks"
      property: "cache_key"
```

### **JWKS Pinning и Refresh Jitter:**
```yaml
# k8s/oathkeeper/jwks-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: oathkeeper-jwks-config
  namespace: oathkeeper-system
data:
  jwks-config.yaml: |
    authenticators:
      oidc:
        config:
          providers:
            - id: zitadel
              issuer: https://zitadel.local
              client_id: cloudbridge-relay
              client_secret: ${ZITADEL_CLIENT_SECRET}
              scope:
                - openid
                - profile
                - email
                - cloudbridge-relay
              # JWKS configuration
              jwks_url: https://zitadel.local/.well-known/jwks.json
              jwks_cache_ttl: 15m
              jwks_refresh_interval: 10m
              jwks_refresh_jitter: 2m  # ±2 minutes jitter
              jwks_pinning:
                enabled: true
                fingerprint: "sha256:ABC123..."
              # ACR/AMR support
              acr_values: ["1", "2"]  # MFA levels
              amr_required: ["mfa"]
              # Audience and issuer validation
              audience: ["cloudbridge-relay"]
              issuer_validation: true
              # Token validation
              token_validation:
                clock_skew: 30s
                max_age: 1h
```

## 📊 **5. Observability для аудита ISO/270xx**

### **Обязательные логи:**
```yaml
# k8s/logging/fluent-bit-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluent-bit-config
  namespace: logging
data:
  fluent-bit.conf: |
    [SERVICE]
        Flush         1
        Log_Level     info
        Daemon        off
        Parsers_File  parsers.conf
        HTTP_Server   On
        HTTP_Listen   0.0.0.0
        HTTP_Port     2020

    [INPUT]
        Name              tail
        Path              /var/log/traefik/*.log
        Parser            traefik
        Tag               traefik.*
        Refresh_Interval  5
        Mem_Buf_Limit     50MB
        Skip_Long_Lines   On

    [INPUT]
        Name              tail
        Path              /var/log/coraza/*.log
        Parser            coraza
        Tag               coraza.*
        Refresh_Interval  5
        Mem_Buf_Limit     50MB
        Skip_Long_Lines   On

    [INPUT]
        Name              tail
        Path              /var/log/oathkeeper/*.log
        Parser            oathkeeper
        Tag               oathkeeper.*
        Refresh_Interval  5
        Mem_Buf_Limit     50MB
        Skip_Long_Lines   On

    [FILTER]
        Name                kubernetes
        Match               traefik.*
        Kube_URL            https://kubernetes.default.svc:443
        Kube_CA_File        /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
        Kube_Token_File     /var/run/secrets/kubernetes.io/serviceaccount/token
        Kube_Tag_Prefix     traefik.
        Merge_Log           On
        Keep_Log            Off
        K8S-Logging.Parser  On
        K8S-Logging.Exclude Off

    [OUTPUT]
        Name  opensearch
        Match *
        Host  opensearch.logging.svc.cluster.local
        Port  9200
        Index cloudbridge-waf-logs
        Type  _doc
        # Retention policy for audit logs (365 days)
        Retry_Limit 3
        Buffer_Size 5MB
        # Security headers
        HTTP_User admin
        HTTP_Passwd ${OPENSEARCH_PASSWORD}
        tls On
        tls.verify Off
```

### **Аудит-логи с ретеншном:**
```yaml
# k8s/logging/opensearch-index-template.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: opensearch-index-template
  namespace: logging
data:
  index-template.json: |
    {
      "index_patterns": ["cloudbridge-waf-logs-*"],
      "template": {
        "settings": {
          "number_of_shards": 3,
          "number_of_replicas": 1,
          "index.lifecycle.name": "waf-logs-policy",
          "index.lifecycle.rollover_alias": "cloudbridge-waf-logs"
        },
        "mappings": {
          "properties": {
            "@timestamp": {
              "type": "date"
            },
            "trace_id": {
              "type": "keyword"
            },
            "client_ip": {
              "type": "ip"
            },
            "user_agent": {
              "type": "text"
            },
            "request_method": {
              "type": "keyword"
            },
            "request_url": {
              "type": "keyword"
            },
            "response_status": {
              "type": "integer"
            },
            "waf_action": {
              "type": "keyword"
            },
            "waf_rule_id": {
              "type": "keyword"
            },
            "oathkeeper_subject": {
              "type": "keyword"
            },
            "oathkeeper_policy": {
              "type": "keyword"
            }
          }
        }
      }
    }
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: opensearch-ilm-policy
  namespace: logging
data:
  ilm-policy.json: |
    {
      "policy": {
        "phases": {
          "hot": {
            "actions": {
              "rollover": {
                "max_size": "50GB",
                "max_age": "7d"
              }
            }
          },
          "warm": {
            "min_age": "7d",
            "actions": {
              "allocate": {
                "number_of_replicas": 0
              }
            }
          },
          "delete": {
            "min_age": "365d"
          }
        }
      }
    }
```

##  **6. Fail-open / Fail-closed стратегии**

### **Конфигурация по типам API:**
```yaml
# k8s/traefik/fail-strategies.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: traefik-fail-strategies
  namespace: traefik-system
data:
  traefik.yml: |
    # Public API - fail-open
    http:
      middlewares:
        waf-public:
          plugin:
            coraza:
              config:
                fail_open: true
                timeout: 5s
                max_request_size: 10MB
        
        # Admin API - fail-closed
        waf-admin:
          plugin:
            coraza:
              config:
                fail_open: false
                timeout: 10s
                max_request_size: 1MB
        
        # Internal API - fail-closed with alert
        waf-internal:
          plugin:
            coraza:
              config:
                fail_open: false
                timeout: 15s
                max_request_size: 5MB
                alert_on_fail: true
```

### **Oathkeeper fail strategies:**
```yaml
# k8s/oathkeeper/fail-strategies.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: oathkeeper-fail-strategies
  namespace: oathkeeper-system
data:
  access-rules-public.yaml: |
    id: public-api-fail-open
    upstream:
      url: http://cloudbridge-relay.cloudbridge.svc.cluster.local:8080
    match:
      url: https://api.cloudbridge.local/public<**>
    authenticators:
      - handler: oidc
        config:
          provider: zitadel
          fail_open: true  # Allow on auth failure
    authorizer:
      handler: remote
      config:
        fail_open: true
    mutators:
      - handler: header
  access-rules-admin.yaml: |
    id: admin-api-fail-closed
    upstream:
      url: http://cloudbridge-relay.cloudbridge.svc.cluster.local:8080
    match:
      url: https://admin.cloudbridge.local<**>
    authenticators:
      - handler: oidc
        config:
          provider: zitadel
          fail_open: false  # Block on auth failure
    authorizer:
      handler: remote
      config:
        fail_open: false
    mutators:
      - handler: header
```

## 🔄 **7. Релизы и откат**

### **Feature flags:**
```yaml
# k8s/feature-flags/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: feature-flags
  namespace: cloudbridge
data:
  flags.yaml: |
    waf:
      enabled: true
      version: "v1.2.0"
      canary_percentage: 0
    
    oathkeeper:
      enabled: true
      version: "v0.40.0"
      canary_percentage: 0
    
    krakend:
      enabled: true
      version: "v2.4.0"
      canary_percentage: 0
    
    # Rollback flags
    rollback:
      waf_previous_version: "v1.1.0"
      oathkeeper_previous_version: "v0.39.0"
      krakend_previous_version: "v2.3.0"
```

### **Blue/Green Ingress:**
```yaml
# k8s/ingress/blue-green-ingress.yaml
apiVersion: networking.k8s.io/v1
kind: IngressClass
metadata:
  name: traefik-blue
  annotations:
    ingressclass.kubernetes.io/is-default-class: "true"
spec:
  controller: traefik.io/ingress-controller
---
apiVersion: networking.k8s.io/v1
kind: IngressClass
metadata:
  name: traefik-green
  annotations:
    ingressclass.kubernetes.io/is-default-class: "false"
spec:
  controller: traefik.io/ingress-controller
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: cloudbridge-blue
  namespace: cloudbridge
  annotations:
    kubernetes.io/ingress.class: traefik-blue
    traefik.ingress.kubernetes.io/router.entrypoints: websecure
    traefik.ingress.kubernetes.io/router.tls: "true"
spec:
  tls:
  - hosts:
    - api.cloudbridge.local
    secretName: cloudbridge-tls
  rules:
  - host: api.cloudbridge.local
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: cloudbridge-relay
            port:
              number: 8080
```

## 🔐 **8. Совместимость с Zitadel**

### **ACR/AMR поддержка:**
```yaml
# k8s/oathkeeper/zitadel-compatibility.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: oathkeeper-zitadel-config
  namespace: oathkeeper-system
data:
  zitadel-config.yaml: |
    authenticators:
      oidc:
        config:
          providers:
            - id: zitadel
              issuer: https://zitadel.local
              client_id: cloudbridge-relay
              client_secret: ${ZITADEL_CLIENT_SECRET}
              scope:
                - openid
                - profile
                - email
                - cloudbridge-relay
              # ACR (Authentication Context Class Reference)
              acr_values: ["1", "2"]  # 1=basic, 2=MFA
              acr_validation:
                required: ["1"]  # Minimum ACR level
                preferred: ["2"] # Preferred ACR level
              # AMR (Authentication Methods References)
              amr_required: ["pwd", "mfa"]  # Password + MFA
              amr_validation: true
              # Audience and authorization party
              audience: ["cloudbridge-relay"]
              azp_validation: true  # Authorized party validation
              # Token validation
              token_validation:
                clock_skew: 30s
                max_age: 1h
                leeway: 10s
              # Claims mapping
              claims_mapping:
                subject: "sub"
                email: "email"
                name: "name"
                roles: "cloudbridge_roles"
                tenant: "cloudbridge_tenant"
```

### **Machine-to-Machine:**
```yaml
# k8s/oathkeeper/m2m-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: oathkeeper-m2m-config
  namespace: oathkeeper-system
data:
  m2m-access-rules.yaml: |
    id: machine-to-machine
    upstream:
      url: http://cloudbridge-relay.cloudbridge.svc.cluster.local:8080
    match:
      url: https://api.cloudbridge.local/m2m<**>
    authenticators:
      - handler: oidc
        config:
          provider: zitadel
          grant_type: "client_credentials"
          scope: ["cloudbridge-relay:m2m"]
    authorizer:
      handler: remote
      config:
        http:
          url: http://oathkeeper-authorizer:8080/authorize
          headers:
            X-M2M-Auth: "true"
    mutators:
      - handler: header
        config:
          headers:
            X-Client-ID: "{{ print .Extra.client_id }}"
            X-Client-Scope: "{{ print .Extra.scope }}"
            X-M2M-Auth: "true"
```

## **9. ISO 27001/27017/27018 документация**

### **ISMS Controls Mapping:**
```yaml
# docs/iso-controls-mapping.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: iso-controls-mapping
  namespace: documentation
data:
  controls-mapping.yaml: |
    iso_27001_controls:
      A.8_Asset_Management:
        - waf_rules_inventory
        - policy_versioning
        - asset_classification
      
      A.12_Operations_Security:
        - waf_monitoring
        - incident_response
        - change_management
      
      A.13_Communications_Security:
        - network_segmentation
        - tls_encryption
        - mfa_authentication
      
      A.12.4_Logging_and_Monitoring:
        - audit_logs
        - security_monitoring
        - log_retention
    
    compliance_requirements:
      iso_27017_cloud_security:
        - shared_responsibility_model
        - cloud_service_monitoring
        - data_location_controls
      
      iso_27018_privacy:
        - data_processing_records
        - consent_management
        - data_retention_policies
```

### **Policy Documents:**
```markdown
# docs/policies/change-management.md
# Change Management Policy for WAF Rules

## Purpose
This policy defines the process for managing changes to WAF rules and security policies.

## Scope
- Coraza WAF rules
- Oathkeeper access rules
- KrakenD endpoint configurations
- Security middleware configurations

## Process
1. **Planning**: All changes must be planned and documented
2. **Testing**: Changes must be tested in staging environment
3. **Approval**: Changes require approval from security team
4. **Deployment**: Canary deployment with monitoring
5. **Validation**: Post-deployment validation and monitoring

## Responsibilities
- **Security Team**: Rule design and approval
- **DevOps Team**: Deployment and monitoring
- **Compliance Team**: Audit and validation
```

## **10. Check-list PoC → Prod**

### **Pre-Production Checklist:**
```yaml
# scripts/production-readiness-checklist.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: production-readiness-checklist
  namespace: cloudbridge
data:
  checklist.yaml: |
    pre_production:
      canary_deployment:
        - percentage: 5-10%
        - duration: 30m
        - metrics_monitoring: true
        - rollback_plan: ready
      
      negative_testing:
        - bola_tests: passed
        - ssrf_tests: passed
        - path_traversal_tests: passed
        - rce_patterns: tested
        - file_upload_abuse: tested
        - header_smuggling: tested
      
      rule_hygiene:
        - paranoia_level: 1-2
        - custom_exceptions: configured
        - false_positive_rate: <0.5%
        - rule_coverage: documented
      
      rollback_rehearsal:
        - rollback_time: <5min
        - data_consistency: verified
        - service_availability: maintained
      
      audit_readiness:
        - logs_accessible: true
        - retention_policy: 365d
        - compliance_reports: generated
        - last_release_documentation: complete
```

## 🎯 **Заключение**

Этот production-ready план обеспечивает:

1. **🛡️ Реалистичные SLO** с учетом интернет-латентности
2. **⚡ L3/L4 защиту** с eBPF и rate limiting
3. **📝 Policy-as-Code** с канареечными выкатами
4. **🔐 Безопасное управление секретами** через Vault
5. **📊 Полную observability** для аудита ISO
6. ** Fail-open/fail-closed** стратегии
7. **🔄 Blue/Green релизы** с быстрым откатом
8. **🔐 Совместимость с Zitadel** ACR/AMR
9. **ISO compliance** документацию
10. **Production readiness** чек-лист

**Готовность к production: 95%**
**Время до production: 4-6 недель**
**Команда: 3-4 разработчика + 1 security engineer**
