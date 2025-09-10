# Kubernetes Package CloudBridge Relay

## 1. Helm Chart Structure

```
cloudbridge-relay/
├── Chart.yaml
├── values.yaml
├── values-production.yaml
├── values-development.yaml
├── templates/
│   ├── deployment/
│   │   ├── relay-server.yaml
│   │   ├── ai-service.yaml
│   │   ├── cadence-service.yaml
│   │   └── keycloak.yaml
│   ├── service/
│   │   ├── relay-service.yaml
│   │   ├── ai-service.yaml
│   │   ├── cadence-service.yaml
│   │   └── keycloak-service.yaml
│   ├── ingress/
│   │   ├── relay-ingress.yaml
│   │   └── management-ingress.yaml
│   ├── configmap/
│   │   ├── relay-config.yaml
│   │   ├── ai-config.yaml
│   │   └── cadence-config.yaml
│   ├── secret/
│   │   ├── tls-secrets.yaml
│   │   ├── database-secrets.yaml
│   │   └── keycloak-secrets.yaml
│   ├── hpa/
│   │   ├── relay-hpa.yaml
│   │   └── ai-hpa.yaml
│   ├── networkpolicy/
│   │   ├── relay-network-policy.yaml
│   │   ├── ai-network-policy.yaml
│   │   └── management-network-policy.yaml
│   ├── psa/
│   │   ├── relay-psa.yaml
│   │   └── ai-psa.yaml
│   ├── rbac/
│   │   ├── service-account.yaml
│   │   ├── role.yaml
│   │   └── rolebinding.yaml
│   └── pvc/
│       ├── relay-storage.yaml
│       ├── ai-models-storage.yaml
│       └── cadence-storage.yaml
└── tests/
    ├── test-connectivity.yaml
    └── test-performance.yaml
```

## 2. Chart.yaml

```yaml
apiVersion: v2
name: cloudbridge-relay
description: CloudBridge Relay - High-performance corporate network platform
type: application
version: 2.0.0
appVersion: "2.0.0"
home: https://cloudbridge.io
sources:
  - https://github.com/cloudbridge/relay-installer
maintainers:
  - name: CloudBridge Team
    email: team@cloudbridge.io
keywords:
  - networking
  - vpn
  - security
  - quic
  - wireguard
  - ai
dependencies:
  - name: keycloak
    version: "18.0.0"
    repository: "https://charts.bitnami.com/bitnami"
    condition: keycloak.enabled
  - name: prometheus
    version: "25.0.0"
    repository: "https://prometheus-community.github.io/helm-charts"
    condition: monitoring.prometheus.enabled
  - name: grafana
    version: "7.0.0"
    repository: "https://grafana.github.io/helm-charts"
    condition: monitoring.grafana.enabled
```

## 3. values.yaml

```yaml
# Global configuration
global:
  imageRegistry: "registry.cloudbridge.io"
  imagePullSecrets:
    - name: cloudbridge-registry-secret
  storageClass: "fast-ssd"
  domain: "cloudbridge.local"
  tls:
    enabled: true
    certManager: true

# Relay Server configuration
relay:
  enabled: true
  replicaCount: 3
  image:
    repository: cloudbridge/relay-server
    tag: "2.0.0"
    pullPolicy: IfNotPresent
  service:
    type: LoadBalancer
    ports:
      tcp: 9090
      udp: 51820
      http: 9091
  resources:
    requests:
      cpu: 1000m
      memory: 2Gi
    limits:
      cpu: 2000m
      memory: 4Gi
  autoscaling:
    enabled: true
    minReplicas: 3
    maxReplicas: 10
    targetCPUUtilizationPercentage: 70
    targetMemoryUtilizationPercentage: 80
  nodeSelector:
    node-type: "relay"
  tolerations:
    - key: "relay-workload"
      operator: "Equal"
      value: "true"
      effect: "NoSchedule"

# AI Service configuration
ai:
  enabled: true
  replicaCount: 2
  image:
    repository: cloudbridge/ai-service
    tag: "2.0.0"
    pullPolicy: IfNotPresent
  service:
    type: ClusterIP
    port: 8080
  resources:
    requests:
      cpu: 500m
      memory: 1Gi
    limits:
      cpu: 1000m
      memory: 2Gi
  autoscaling:
    enabled: true
    minReplicas: 2
    maxReplicas: 5
    targetCPUUtilizationPercentage: 80
  models:
    storageSize: 10Gi
    storageClass: "fast-ssd"

# Cadence Service configuration
cadence:
  enabled: true
  replicaCount: 3
  image:
    repository: cloudbridge/cadence-service
    tag: "2.0.0"
    pullPolicy: IfNotPresent
  service:
    type: ClusterIP
    port: 7933
  resources:
    requests:
      cpu: 500m
      memory: 1Gi
    limits:
      cpu: 1000m
      memory: 2Gi
  persistence:
    enabled: true
    size: 20Gi
    storageClass: "fast-ssd"

# Keycloak configuration
keycloak:
  enabled: true
  auth:
    adminUser: admin
    adminPassword: "changeme"
  postgresql:
    enabled: true
    auth:
      postgresPassword: "changeme"
      database: keycloak

# Monitoring configuration
monitoring:
  prometheus:
    enabled: true
    server:
      persistentVolume:
        enabled: true
        size: 50Gi
  grafana:
    enabled: true
    adminPassword: "changeme"
    persistence:
      enabled: true
      size: 10Gi

# Security configuration
security:
  podSecurityStandards:
    enforce: "restricted"
    audit: "restricted"
    warn: "restricted"
  networkPolicies:
    enabled: true
  serviceAccount:
    create: true
    annotations: {}
  rbac:
    create: true

# Ingress configuration
ingress:
  enabled: true
  className: "nginx"
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
  hosts:
    - host: relay.cloudbridge.local
      paths:
        - path: /
          pathType: Prefix
    - host: management.cloudbridge.local
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: cloudbridge-tls
      hosts:
        - relay.cloudbridge.local
        - management.cloudbridge.local
```

## 4. Relay Server Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "cloudbridge-relay.fullname" . }}-relay
  labels:
    {{- include "cloudbridge-relay.labels" . | nindent 4 }}
    component: relay-server
spec:
  replicas: {{ .Values.relay.replicaCount }}
  selector:
    matchLabels:
      {{- include "cloudbridge-relay.selectorLabels" . | nindent 6 }}
      component: relay-server
  template:
    metadata:
      labels:
        {{- include "cloudbridge-relay.selectorLabels" . | nindent 8 }}
        component: relay-server
      annotations:
        checksum/config: {{ include (print $.Template.BasePath "/configmap/relay-config.yaml") . | sha256sum }}
    spec:
      serviceAccountName: {{ include "cloudbridge-relay.serviceAccountName" . }}
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        runAsGroup: 1000
        fsGroup: 1000
      containers:
        - name: relay-server
          image: "{{ .Values.global.imageRegistry }}/{{ .Values.relay.image.repository }}:{{ .Values.relay.image.tag }}"
          imagePullPolicy: {{ .Values.relay.image.pullPolicy }}
          ports:
            - name: tcp
              containerPort: 9090
              protocol: TCP
            - name: udp
              containerPort: 51820
              protocol: UDP
            - name: http
              containerPort: 9091
              protocol: TCP
          env:
            - name: RELAY_CONFIG_PATH
              value: "/etc/relay/config.yaml"
            - name: LOG_LEVEL
              value: "info"
            - name: METRICS_ENABLED
              value: "true"
          volumeMounts:
            - name: config
              mountPath: /etc/relay
              readOnly: true
            - name: tls-certs
              mountPath: /etc/tls
              readOnly: true
            - name: storage
              mountPath: /var/lib/relay
          resources:
            {{- toYaml .Values.relay.resources | nindent 12 }}
          livenessProbe:
            httpGet:
              path: /health
              port: http
            initialDelaySeconds: 30
            periodSeconds: 10
            timeoutSeconds: 5
            failureThreshold: 3
          readinessProbe:
            httpGet:
              path: /ready
              port: http
            initialDelaySeconds: 5
            periodSeconds: 5
            timeoutSeconds: 3
            failureThreshold: 3
          securityContext:
            allowPrivilegeEscalation: false
            readOnlyRootFilesystem: true
            capabilities:
              drop:
                - ALL
      volumes:
        - name: config
          configMap:
            name: {{ include "cloudbridge-relay.fullname" . }}-relay-config
        - name: tls-certs
          secret:
            secretName: {{ include "cloudbridge-relay.fullname" . }}-tls
        - name: storage
          persistentVolumeClaim:
            claimName: {{ include "cloudbridge-relay.fullname" . }}-relay-storage
      nodeSelector:
        {{- toYaml .Values.relay.nodeSelector | nindent 8 }}
      tolerations:
        {{- toYaml .Values.relay.tolerations | nindent 8 }}
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
            - weight: 100
              podAffinityTerm:
                labelSelector:
                  matchExpressions:
                    - key: component
                      operator: In
                      values:
                        - relay-server
                topologyKey: kubernetes.io/hostname
```

## 5. Horizontal Pod Autoscaler

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: {{ include "cloudbridge-relay.fullname" . }}-relay-hpa
  labels:
    {{- include "cloudbridge-relay.labels" . | nindent 4 }}
    component: relay-server
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: {{ include "cloudbridge-relay.fullname" . }}-relay
  minReplicas: {{ .Values.relay.autoscaling.minReplicas }}
  maxReplicas: {{ .Values.relay.autoscaling.maxReplicas }}
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: {{ .Values.relay.autoscaling.targetCPUUtilizationPercentage }}
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: {{ .Values.relay.autoscaling.targetMemoryUtilizationPercentage }}
    - type: Pods
      pods:
        metric:
          name: cloudbridge_connections_active
        target:
          type: AverageValue
          averageValue: "1000"
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
        - type: Percent
          value: 10
          periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
        - type: Percent
          value: 50
          periodSeconds: 60
        - type: Pods
          value: 2
          periodSeconds: 60
      selectPolicy: Max
```

## 6. Network Policies

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: {{ include "cloudbridge-relay.fullname" . }}-relay-netpol
  labels:
    {{- include "cloudbridge-relay.labels" . | nindent 4 }}
    component: relay-server
spec:
  podSelector:
    matchLabels:
      {{- include "cloudbridge-relay.selectorLabels" . | nindent 6 }}
      component: relay-server
  policyTypes:
    - Ingress
    - Egress
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              name: {{ .Release.Namespace }}
        - podSelector:
            matchLabels:
              component: ai-service
        - podSelector:
            matchLabels:
              component: cadence-service
      ports:
        - protocol: TCP
          port: 9090
        - protocol: UDP
          port: 51820
        - protocol: TCP
          port: 9091
    - from:
        - namespaceSelector: {}
      ports:
        - protocol: TCP
          port: 9090
        - protocol: UDP
          port: 51820
  egress:
    - to:
        - podSelector:
            matchLabels:
              component: ai-service
      ports:
        - protocol: TCP
          port: 8080
    - to:
        - podSelector:
            matchLabels:
              component: cadence-service
      ports:
        - protocol: TCP
          port: 7933
    - to:
        - podSelector:
            matchLabels:
              app.kubernetes.io/name: keycloak
      ports:
        - protocol: TCP
          port: 8080
    - to: []
      ports:
        - protocol: TCP
          port: 53
        - protocol: UDP
          port: 53
        - protocol: TCP
          port: 443
        - protocol: TCP
          port: 80
```

## 7. Pod Security Standards

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: {{ .Release.Namespace }}
  labels:
    pod-security.kubernetes.io/enforce: {{ .Values.security.podSecurityStandards.enforce }}
    pod-security.kubernetes.io/audit: {{ .Values.security.podSecurityStandards.audit }}
    pod-security.kubernetes.io/warn: {{ .Values.security.podSecurityStandards.warn }}
---
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: {{ include "cloudbridge-relay.fullname" . }}-psp
  labels:
    {{- include "cloudbridge-relay.labels" . | nindent 4 }}
spec:
  privileged: false
  allowPrivilegeEscalation: false
  requiredDropCapabilities:
    - ALL
  volumes:
    - 'configMap'
    - 'emptyDir'
    - 'projected'
    - 'secret'
    - 'downwardAPI'
    - 'persistentVolumeClaim'
  runAsUser:
    rule: 'MustRunAsNonRoot'
  seLinux:
    rule: 'RunAsAny'
  fsGroup:
    rule: 'RunAsAny'
  readOnlyRootFilesystem: true
```

## 8. Service Account and RBAC

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{ include "cloudbridge-relay.serviceAccountName" . }}
  labels:
    {{- include "cloudbridge-relay.labels" . | nindent 4 }}
  annotations:
    {{- toYaml .Values.security.serviceAccount.annotations | nindent 4 }}
automountServiceAccountToken: true
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: {{ include "cloudbridge-relay.fullname" . }}-role
  labels:
    {{- include "cloudbridge-relay.labels" . | nindent 4 }}
rules:
  - apiGroups: [""]
    resources: ["configmaps", "secrets"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["apps"]
    resources: ["deployments"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["autoscaling"]
    resources: ["horizontalpodautoscalers"]
    verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: {{ include "cloudbridge-relay.fullname" . }}-rolebinding
  labels:
    {{- include "cloudbridge-relay.labels" . | nindent 4 }}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: {{ include "cloudbridge-relay.fullname" . }}-role
subjects:
  - kind: ServiceAccount
    name: {{ include "cloudbridge-relay.serviceAccountName" . }}
    namespace: {{ .Release.Namespace }}
```

## 9. Persistent Volume Claims

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: {{ include "cloudbridge-relay.fullname" . }}-relay-storage
  labels:
    {{- include "cloudbridge-relay.labels" . | nindent 4 }}
    component: relay-server
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: {{ .Values.relay.storage.size | default "10Gi" }}
  storageClassName: {{ .Values.global.storageClass }}
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: {{ include "cloudbridge-relay.fullname" . }}-ai-models-storage
  labels:
    {{- include "cloudbridge-relay.labels" . | nindent 4 }}
    component: ai-service
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: {{ .Values.ai.models.storageSize }}
  storageClassName: {{ .Values.global.storageClass }}
```

## 10. Installation Commands

```bash
# Add CloudBridge Helm repository
helm repo add cloudbridge https://charts.cloudbridge.io
helm repo update

# Install CloudBridge Relay
helm install cloudbridge-relay cloudbridge/cloudbridge-relay \
  --namespace cloudbridge \
  --create-namespace \
  --values values-production.yaml \
  --set global.domain=cloudbridge.io \
  --set global.tls.enabled=true \
  --set relay.replicaCount=5 \
  --set ai.replicaCount=3 \
  --set monitoring.prometheus.enabled=true \
  --set monitoring.grafana.enabled=true

# Upgrade existing installation
helm upgrade cloudbridge-relay cloudbridge/cloudbridge-relay \
  --namespace cloudbridge \
  --values values-production.yaml

# Uninstall
helm uninstall cloudbridge-relay --namespace cloudbridge
```

---

*Этот Kubernetes пакет предоставляет полную конфигурацию для развертывания CloudBridge Relay в production среде, включая Helm charts, манифесты, HPA, Network Policies и Pod Security Standards.*
