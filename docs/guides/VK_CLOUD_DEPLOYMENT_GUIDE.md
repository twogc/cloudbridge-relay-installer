# Руководство по развертыванию CloudBridge Relay в VK Cloud Kubernetes

## Обзор

Данное руководство описывает пошаговый процесс развертывания проекта CloudBridge Relay в кластере Kubernetes VK Cloud. Проект включает в себя:

- **CloudBridge Relay** - основной сервис
- **AI Service** - сервис искусственного интеллекта
- **Cadence** - workflow engine
- **PostgreSQL** - база данных
- **Cassandra** - NoSQL база данных
- **WireGuard** - VPN сервис

## Предварительные требования

### Системные требования:
- Ubuntu 18.04+ или аналогичная ОС
- Docker 20.10+
- kubectl 1.28+
- Helm 3.12+
- Terraform 1.5+

### VK Cloud требования:
- Аккаунт в VK Cloud
- Проект с доступом к Kubernetes
- OpenStack RC файл v3
- Достаточные квоты для создания кластера

## Содержание

1. [Подготовка окружения](#1-подготовка-окружения)
2. [Создание кластера Kubernetes](#2-создание-кластера-kubernetes)
3. [Установка аддонов](#3-установка-аддонов)
4. [Создание Docker образов](#4-создание-docker-образов)
5. [Настройка Registry](#5-настройка-registry)
6. [Развертывание сервисов](#6-развертывание-сервисов)
7. [Настройка мониторинга](#7-настройка-мониторинга)
8. [Тестирование](#8-тестирование)
9. [Устранение неполадок](#9-устранение-неполадок)

---

## 1. Подготовка окружения

### 1.1 Установка необходимых инструментов

```bash
# Обновление системы
sudo apt update && sudo apt upgrade -y

# Установка Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# Настройка зеркал Docker Hub (для России)
cat << EOF | sudo tee -a /etc/docker/daemon.json
{
  "registry-mirrors": [
    "https://dockerhub.timeweb.cloud",
    "https://huecker.io",
    "https://mirror.gcr.io",
    "https://c.163.com",
    "https://registry.docker-cn.com",
    "https://daocloud.io",
    "https://cr.yandex/mirror",
    "https://noohub.ru"
  ]
}
EOF

# Перезапуск Docker
sudo systemctl restart docker

# Проверка настройки зеркал
docker info | grep -A 10 "Registry Mirrors"

# Установка kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Установка Helm
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Установка Terraform
wget -O- https://apt.releases.hashicorp.com/gpg | gpg --dearmor | sudo tee /usr/share/keyrings/hashicorp-archive-keyring.gpg
echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/hashicorp.list
sudo apt update && sudo apt install terraform
```

### 1.2 Настройка OpenStack CLI

```bash
# Установка OpenStack CLI
sudo apt install python3-openstackclient -y

# Создание файла аутентификации
cat > openstack_auth.sh << 'EOF'
#!/usr/bin/env bash

export OS_AUTH_URL="https://infra.mail.ru:35357/v3/"
export OS_PROJECT_ID="YOUR_PROJECT_ID"
export OS_REGION_NAME="RegionOne"
unset OS_PROJECT_NAME
unset OS_PROJECT_DOMAIN_ID
unset OS_TENANT_ID
unset OS_TENANT_NAME

export OS_USERNAME="YOUR_USERNAME"
export OS_USER_DOMAIN_NAME="users"
export OS_PASSWORD="YOUR_PASSWORD"

export OS_INTERFACE=public
export OS_IDENTITY_API_VERSION=3
EOF

chmod +x openstack_auth.sh
```

### 1.3 Настройка Terraform

```bash
# Создание конфигурации Terraform mirror
cat > ~/.terraformrc << 'EOF'
provider_installation {
  network_mirror {
    url = "https://terraform-mirror.mcs.mail.ru"
    include = ["registry.terraform.io/*/*"]
  }
  direct {
    exclude = ["registry.terraform.io/*/*"]
  }
}
EOF
```

---

## 2. Создание кластера Kubernetes

### 2.1 Подготовка Terraform конфигурации

```bash
# Создание директории для Terraform
mkdir -p terraform-k8s
cd terraform-k8s

# Создание provider.tf
cat > provider.tf << 'EOF'
terraform {
  required_providers {
    vkcs = {
      source = "vk-cs/vkcs"
      version = "~> 0.8.0"
    }
  }
}

provider "vkcs" {
  username   = "YOUR_USERNAME"
  password   = "YOUR_PASSWORD"
  project_id = "YOUR_PROJECT_ID"
  region     = "RegionOne"
  auth_url   = "https://infra.mail.ru:35357/v3/"
}
EOF

# Создание main.tf
cat > k8s-cluster.tf << 'EOF'
# Получение информации о доступных ресурсах
data "vkcs_compute_flavor" "k8s-master-flavor" {
  name = "STD3-4-8"  # 4 CPU, 8GB RAM
}

data "vkcs_kubernetes_clustertemplates" "k8s-template-list" {}

data "vkcs_kubernetes_clustertemplate" "k8s-template" {
  version = "1.28"
}

data "vkcs_networking_network" "k8s-network" {
  name = "cloudbridge-network"
}

data "vkcs_networking_subnet" "k8s-subnet" {
  name = "cloudbridge-subnet"
  network_id = data.vkcs_networking_network.k8s-network.id
}

# Создание кластера Kubernetes
resource "vkcs_kubernetes_cluster" "k8s-cluster" {
  name                = "k8s-cluster"
  cluster_type        = "standard"
  cluster_template_id = data.vkcs_kubernetes_clustertemplate.k8s-template.id
  master_flavor       = data.vkcs_compute_flavor.k8s-master-flavor.id
  master_count        = 1
  network_id          = data.vkcs_networking_network.k8s-network.id
  subnet_id           = data.vkcs_networking_subnet.k8s-subnet.id
  availability_zone   = "ME1"
  floating_ip_enabled = true
}

# Создание группы узлов
resource "vkcs_kubernetes_node_group" "worker-nodes" {
  cluster_id         = vkcs_kubernetes_cluster.k8s-cluster.id
  name               = "worker-nodes"
  flavor_id          = data.vkcs_compute_flavor.k8s-master-flavor.id
  node_count         = 2
  availability_zones = ["ME1"]
}

# Вывод kubeconfig
output "k8s_config" {
  value = vkcs_kubernetes_cluster.k8s-cluster.k8s_config
  sensitive = true
}
EOF
```

### 2.2 Развертывание кластера

```bash
# Инициализация Terraform
terraform init

# Планирование развертывания
terraform plan

# Развертывание кластера
terraform apply -auto-approve

# Получение kubeconfig
terraform output -raw k8s_config > ~/.kube/config-vkcloud
export KUBECONFIG=~/.kube/config-vkcloud

# Проверка подключения
kubectl get nodes
```

---

## 3. Установка аддонов

### 3.1 Установка Helm репозиториев

```bash
# Добавление репозиториев
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo add grafana https://grafana.github.io/helm-charts
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo add jetstack https://charts.jetstack.io
helm repo add argo https://argoproj.github.io/argo-helm
helm repo add vmware-tanzu https://vmware-tanzu.github.io/helm-charts
helm repo add jaegertracing https://jaegertracing.github.io/helm-charts

# Обновление репозиториев
helm repo update
```

### 3.2 Установка Prometheus Stack

```bash
# Создание namespace для мониторинга
kubectl create namespace monitoring

# Установка kube-prometheus-stack
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --version 78.2.1 \
  --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false \
  --set prometheus.prometheusSpec.podMonitorSelectorNilUsesHelmValues=false \
  --set prometheus.prometheusSpec.ruleSelectorNilUsesHelmValues=false
```

### 3.3 Установка NGINX Ingress

```bash
# Создание namespace для ingress
kubectl create namespace ingress-nginx

# Установка NGINX Ingress
helm install ingress-nginx ingress-nginx/ingress-nginx \
  --namespace ingress-nginx \
  --version 4.13.3 \
  --set controller.service.type=LoadBalancer \
  --set controller.service.externalTrafficPolicy=Local
```

### 3.4 Установка cert-manager

```bash
# Создание namespace для cert-manager
kubectl create namespace cert-manager

# Установка cert-manager
helm install cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --version v1.18.2 \
  --set installCRDs=true
```

### 3.5 Установка ArgoCD

```bash
# Создание namespace для ArgoCD
kubectl create namespace argocd

# Установка ArgoCD
helm install argocd argo/argo-cd \
  --namespace argocd \
  --version 3.1.7
```

### 3.6 Установка Velero

```bash
# Создание namespace для Velero
kubectl create namespace velero

# Установка Velero
helm install velero vmware-tanzu/velero \
  --namespace velero \
  --version 1.17.0
```

### 3.7 Установка Jaeger

```bash
# Создание namespace для Jaeger
kubectl create namespace jaeger

# Установка Jaeger
helm install jaeger jaegertracing/jaeger \
  --namespace jaeger \
  --version 2.11.0
```

---

## 4. Создание Docker образов

### 4.1 Подготовка Dockerfile

```bash
# Создание Dockerfile для CloudBridge Relay
cat > docker/Dockerfile << 'EOF'
# Multi-stage build for CloudBridge Relay with Enhanced Components
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk update && apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cloudbridge-relay ./cmd/relay

# Final stage
FROM alpine:3.18

# Install runtime dependencies including WireGuard
RUN apk update && apk --no-cache add ca-certificates tzdata wireguard-tools

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/cloudbridge-relay .

# Copy config files
COPY --from=builder /app/configs ./configs

# Create directories for enhanced components
RUN mkdir -p /app/data /app/certs && \
    chown -R appuser:appgroup /app/data /app/certs

# Change ownership
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose ports
EXPOSE 9091 8082 8443 5553/udp 19302/udp 9090/udp 9092/udp

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9091/health || exit 1

# Run the application
CMD ["./cloudbridge-relay"]
EOF
```

### 4.2 Сборка образов

```bash
# Сборка образа CloudBridge Relay
docker build -f docker/Dockerfile -t cloudbridge-relay:latest .

# Сборка образа AI Service
docker build -f /tmp/ai-service.Dockerfile -t ai-service:latest .

# Сборка образа Cadence
docker build -f /tmp/cadence.Dockerfile -t cadence:latest .

# Проверка созданных образов
docker images | grep -E "(cloudbridge-relay|ai-service|cadence)"
```

---

## 5. Настройка Registry

### 5.1 Создание Docker Registry в кластере

```bash
# Создание namespace для registry
kubectl create namespace cloudbridge

# Создание PVC для registry
cat > registry-pvc.yaml << 'EOF'
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: registry-pvc
  namespace: cloudbridge
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
  storageClassName: csi-ceph-hdd-me1
EOF

kubectl apply -f registry-pvc.yaml

# Создание deployment для registry
cat > registry-deployment.yaml << 'EOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: docker-registry
  namespace: cloudbridge
  labels:
    app: docker-registry
spec:
  replicas: 1
  selector:
    matchLabels:
      app: docker-registry
  template:
    metadata:
      labels:
        app: docker-registry
    spec:
      containers:
      - name: registry
        image: registry:2
        ports:
        - containerPort: 5000
        env:
        - name: REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY
          value: /var/lib/registry
        volumeMounts:
        - name: registry-storage
          mountPath: /var/lib/registry
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "200m"
      volumes:
      - name: registry-storage
        persistentVolumeClaim:
          claimName: registry-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: docker-registry
  namespace: cloudbridge
spec:
  selector:
    app: docker-registry
  ports:
  - port: 5000
    targetPort: 5000
  type: ClusterIP
EOF

kubectl apply -f registry-deployment.yaml
```

### 5.2 Настройка HTTPS Registry (опционально)

```bash
# Создание TLS сертификата
kubectl create secret tls registry-tls \
  --cert=certs/edge.2gc.ru-fullchain.pem \
  --key=certs/edge.2gc.ru-privkey.pem \
  --namespace=cloudbridge

# Создание HTTPS registry deployment
cat > registry-https-deployment.yaml << 'EOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: docker-registry-https
  namespace: cloudbridge
  labels:
    app: docker-registry-https
spec:
  replicas: 1
  selector:
    matchLabels:
      app: docker-registry-https
  template:
    metadata:
      labels:
        app: docker-registry-https
    spec:
      containers:
      - name: registry
        image: registry:2
        ports:
        - containerPort: 5000
        env:
        - name: REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY
          value: /var/lib/registry
        - name: REGISTRY_HTTP_TLS_CERTIFICATE
          value: /certs/tls.crt
        - name: REGISTRY_HTTP_TLS_KEY
          value: /certs/tls.key
        volumeMounts:
        - name: registry-storage
          mountPath: /var/lib/registry
        - name: tls-certs
          mountPath: /certs
          readOnly: true
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "200m"
      volumes:
      - name: registry-storage
        persistentVolumeClaim:
          claimName: registry-pvc
      - name: tls-certs
        secret:
          secretName: registry-tls
---
apiVersion: v1
kind: Service
metadata:
  name: docker-registry-https
  namespace: cloudbridge
spec:
  selector:
    app: docker-registry-https
  ports:
  - port: 5000
    targetPort: 5000
  type: ClusterIP
EOF

kubectl apply -f registry-https-deployment.yaml
```

### 5.3 Загрузка образов в registry

```bash
# Получение IP адреса registry
REGISTRY_IP=$(kubectl get service docker-registry -n cloudbridge -o jsonpath='{.spec.clusterIP}')

# Настройка port-forward для доступа к registry
kubectl port-forward service/docker-registry 5000:5000 -n cloudbridge &

# Тегирование образов для registry
docker tag cloudbridge-relay:latest localhost:5000/cloudbridge-relay:latest
docker tag ai-service:latest localhost:5000/ai-service:latest
docker tag cadence:latest localhost:5000/cadence:latest

# Загрузка образов в registry
docker push localhost:5000/cloudbridge-relay:latest
docker push localhost:5000/ai-service:latest
docker push localhost:5000/cadence:latest
```

---

## 6. Развертывание сервисов

### 6.1 Создание namespace и базовых ресурсов

```bash
# Создание namespace
kubectl create namespace cloudbridge

# Создание PVC для логов
cat > cloudbridge-logs-pvc.yaml << 'EOF'
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: cloudbridge-logs-pvc
  namespace: cloudbridge
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
  storageClassName: csi-ceph-hdd-me1
EOF

kubectl apply -f cloudbridge-logs-pvc.yaml

# Создание PVC для Cadence
cat > cadence-pvc.yaml << 'EOF'
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: cadence-pvc
  namespace: cloudbridge
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 5Gi
  storageClassName: csi-ceph-hdd-me1
EOF

kubectl apply -f cadence-pvc.yaml
```

### 6.2 Создание секретов

```bash
# Создание секрета для CloudBridge
kubectl create secret generic cloudbridge-secrets \
  --from-literal=cassandra-username=cassandra \
  --from-literal=cassandra-password=cassandra \
  --from-literal=api-key=your-api-key \
  --from-literal=jwt-secret=your-jwt-secret \
  --namespace=cloudbridge

# Создание секрета для WireGuard
kubectl create secret generic wireguard-keys \
  --from-literal=private-key=your-private-key \
  --from-literal=public-key=your-public-key \
  --namespace=cloudbridge

# Создание секрета для TLS
kubectl create secret tls cloudbridge-tls \
  --cert=certs/edge.2gc.ru-fullchain.pem \
  --key=certs/edge.2gc.ru-privkey.pem \
  --namespace=cloudbridge

# Создание секрета для registry
kubectl create secret docker-registry k8s-registry-creds \
  --docker-server=http://localhost:5000 \
  --docker-username=admin \
  --docker-password=admin \
  --namespace=cloudbridge
```

### 6.3 Создание ConfigMap

```bash
# Создание ConfigMap для CloudBridge
cat > cloudbridge-config.yaml << 'EOF'
apiVersion: v1
kind: ConfigMap
metadata:
  name: cloudbridge-config
  namespace: cloudbridge
data:
  config.json: |
    {
      "log_level": "info",
      "gin_mode": "release",
      "tz": "UTC",
      "wireguard_enabled": true,
      "grpc_enabled": true,
      "grpc_port": 8444,
      "grpc_tls_enabled": true,
      "grpc_cert_file": "/app/certs/tls.crt",
      "grpc_key_file": "/app/certs/tls.key",
      "prometheus_url": "http://prometheus-server.monitoring.svc.cluster.local:9090"
    }
EOF

kubectl apply -f cloudbridge-config.yaml
```

### 6.4 Развертывание основных сервисов

```bash
# Создание deployment для CloudBridge Relay
cat > cloudbridge-relay-deployment.yaml << 'EOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cloudbridge-relay
  namespace: cloudbridge
  labels:
    app.kubernetes.io/name: cloudbridge-relay
    app.kubernetes.io/version: "1.0.0"
    app.kubernetes.io/component: relay
    app.kubernetes.io/part-of: cloudbridge-relay
    app.kubernetes.io/managed-by: kubectl
spec:
  replicas: 3
  selector:
    matchLabels:
      app.kubernetes.io/name: cloudbridge-relay
      app.kubernetes.io/component: relay
  template:
    metadata:
      labels:
        app.kubernetes.io/name: cloudbridge-relay
        app.kubernetes.io/component: relay
        app.kubernetes.io/part-of: cloudbridge-relay
        app.kubernetes.io/version: "1.0.0"
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8084"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: cloudbridge-relay
      imagePullSecrets:
      - name: k8s-registry-creds
      containers:
      - name: cloudbridge-relay
        image: localhost:5000/cloudbridge-relay:latest
        ports:
        - name: http
          containerPort: 9090
          protocol: TCP
        - name: quic
          containerPort: 9090
          protocol: UDP
        - name: wireguard
          containerPort: 51820
          protocol: UDP
        - name: metrics
          containerPort: 8084
          protocol: TCP
        - name: grpc
          containerPort: 8444
          protocol: TCP
        env:
        - name: CONFIG_PATH
          value: "/app/configs/config.json"
        - name: LOG_LEVEL
          value: "info"
        - name: GIN_MODE
          value: "release"
        - name: TZ
          value: "UTC"
        - name: CASSANDRA_HOSTS
          value: "cassandra-service"
        - name: CASSANDRA_PORT
          value: "9042"
        - name: CASSANDRA_USERNAME
          valueFrom:
            secretKeyRef:
              name: cloudbridge-secrets
              key: cassandra-username
        - name: CASSANDRA_PASSWORD
          valueFrom:
            secretKeyRef:
              name: cloudbridge-secrets
              key: cassandra-password
        - name: CADENCE_HOST
          value: "cadence-service"
        - name: CADENCE_PORT
          value: "7933"
        - name: WIREGUARD_ENABLED
          value: "true"
        - name: WIREGUARD_PRIVATE_KEY
          valueFrom:
            secretKeyRef:
              name: wireguard-keys
              key: private-key
        - name: WIREGUARD_PUBLIC_KEY
          valueFrom:
            secretKeyRef:
              name: wireguard-keys
              key: public-key
        - name: API_KEY
          valueFrom:
            secretKeyRef:
              name: cloudbridge-secrets
              key: api-key
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: cloudbridge-secrets
              key: jwt-secret
        - name: PROMETHEUS_URL
          value: "http://prometheus-server.monitoring.svc.cluster.local:9090"
        - name: RELAY_JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: cloudbridge-secrets
              key: jwt-secret
        - name: GRPC_ENABLED
          value: "true"
        - name: GRPC_PORT
          value: "8444"
        - name: GRPC_TLS_ENABLED
          value: "true"
        - name: GRPC_CERT_FILE
          value: "/app/certs/tls.crt"
        - name: GRPC_KEY_FILE
          value: "/app/certs/tls.key"
        volumeMounts:
        - name: config
          mountPath: /app/configs
          readOnly: true
        - name: logs
          mountPath: /app/logs
        - name: wireguard-config
          mountPath: /etc/wireguard
          readOnly: true
        - name: tls-certs
          mountPath: /app/certs
          readOnly: true
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 9090
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /health
            port: 9090
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 3
      volumes:
      - name: config
        configMap:
          name: cloudbridge-config
      - name: logs
        persistentVolumeClaim:
          claimName: cloudbridge-logs-pvc
      - name: wireguard-config
        secret:
          secretName: wireguard-keys
      - name: tls-certs
        secret:
          secretName: cloudbridge-tls
      nodeSelector:
        kubernetes.io/os: linux
      tolerations:
      - key: node-role.kubernetes.io/master
        operator: Exists
        effect: NoSchedule
EOF

kubectl apply -f cloudbridge-relay-deployment.yaml
```

### 6.5 Создание сервисов

```bash
# Создание сервисов для CloudBridge
cat > cloudbridge-services.yaml << 'EOF'
apiVersion: v1
kind: Service
metadata:
  name: cloudbridge-relay-service
  namespace: cloudbridge
spec:
  selector:
    app.kubernetes.io/name: cloudbridge-relay
  ports:
  - name: http
    port: 9090
    targetPort: 9090
    protocol: TCP
  - name: quic
    port: 9090
    targetPort: 9090
    protocol: UDP
  - name: wireguard
    port: 51820
    targetPort: 51820
    protocol: UDP
  - name: grpc
    port: 8444
    targetPort: 8444
    protocol: TCP
  type: ClusterIP
---
apiVersion: v1
kind: Service
metadata:
  name: cassandra-service
  namespace: cloudbridge
spec:
  selector:
    app: cassandra
  ports:
  - port: 9042
    targetPort: 9042
  type: ClusterIP
---
apiVersion: v1
kind: Service
metadata:
  name: cadence-service
  namespace: cloudbridge
spec:
  selector:
    app: cadence
  ports:
  - port: 7933
    targetPort: 7933
  type: ClusterIP
EOF

kubectl apply -f cloudbridge-services.yaml
```

---

## 7. Настройка мониторинга

### 7.1 Проверка статуса Prometheus

```bash
# Проверка статуса Prometheus
kubectl get pods -n monitoring

# Получение доступа к Prometheus
kubectl port-forward service/prometheus-kube-prometheus-prometheus 9090:9090 -n monitoring &
```

### 7.2 Проверка статуса Grafana

```bash
# Получение пароля администратора Grafana
kubectl get secret prometheus-grafana -n monitoring -o jsonpath="{.data.admin-password}" | base64 --decode

# Получение доступа к Grafana
kubectl port-forward service/prometheus-grafana 3000:80 -n monitoring &
```

### 7.3 Настройка ServiceMonitor

```bash
# Создание ServiceMonitor для CloudBridge
cat > cloudbridge-servicemonitor.yaml << 'EOF'
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: cloudbridge-relay
  namespace: cloudbridge
  labels:
    app: cloudbridge-relay
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: cloudbridge-relay
  endpoints:
  - port: metrics
    path: /metrics
    interval: 30s
EOF

kubectl apply -f cloudbridge-servicemonitor.yaml
```

---

## 8. Тестирование

### 8.1 Проверка статуса подов

```bash
# Проверка статуса всех подов
kubectl get pods -n cloudbridge

# Проверка логов CloudBridge Relay
kubectl logs -l app.kubernetes.io/name=cloudbridge-relay -n cloudbridge

# Проверка событий
kubectl get events -n cloudbridge --sort-by='.lastTimestamp'
```

### 8.2 Проверка сервисов

```bash
# Проверка сервисов
kubectl get services -n cloudbridge

# Проверка endpoints
kubectl get endpoints -n cloudbridge

# Тестирование подключения к сервисам
kubectl run test-pod --image=busybox --rm -it --restart=Never -n cloudbridge -- wget -qO- http://cloudbridge-relay-service:9090/health
```

### 8.3 Проверка мониторинга

```bash
# Проверка метрик в Prometheus
curl http://localhost:9090/api/v1/targets

# Проверка дашбордов в Grafana
# Откройте http://localhost:3000 в браузере
# Логин: admin
# Пароль: (полученный ранее)
```

---

## 9. Устранение неполадок

### 9.1 Проблемы с образами

```bash
# Проверка доступности registry
kubectl run test-pod --image=busybox --rm -it --restart=Never -n cloudbridge -- wget -qO- http://docker-registry:5000/v2/

# Проверка секретов для registry
kubectl get secret k8s-registry-creds -n cloudbridge -o yaml

# Пересоздание секрета
kubectl delete secret k8s-registry-creds -n cloudbridge
kubectl create secret docker-registry k8s-registry-creds \
  --docker-server=http://docker-registry:5000 \
  --docker-username=admin \
  --docker-password=admin \
  --namespace=cloudbridge
```

### 9.2 Проблемы с PVC

```bash
# Проверка статуса PVC
kubectl get pvc -n cloudbridge

# Проверка доступных storage classes
kubectl get storageclass

# Пересоздание PVC с правильным storage class
kubectl delete pvc cloudbridge-logs-pvc -n cloudbridge
kubectl apply -f cloudbridge-logs-pvc.yaml
```

### 9.3 Проблемы с сетью

```bash
# Проверка DNS разрешения
kubectl run test-pod --image=busybox --rm -it --restart=Never -n cloudbridge -- nslookup kubernetes.default

# Проверка подключения к сервисам
kubectl run test-pod --image=busybox --rm -it --restart=Never -n cloudbridge -- wget -qO- http://cloudbridge-relay-service:9090/health
```

### 9.4 Проблемы с мониторингом

```bash
# Проверка статуса Prometheus
kubectl get pods -n monitoring

# Проверка логов Prometheus
kubectl logs -l app.kubernetes.io/name=prometheus -n monitoring

# Перезапуск Prometheus
kubectl rollout restart deployment prometheus-kube-prometheus-prometheus -n monitoring
```

---

## Мониторинг и логирование

### Доступ к сервисам:

- **Prometheus**: `http://localhost:9090`
- **Grafana**: `http://localhost:3000`
- **Jaeger**: `http://localhost:16686`
- **ArgoCD**: `http://localhost:8080`

### Полезные команды:

```bash
# Просмотр всех ресурсов
kubectl get all -n cloudbridge

# Просмотр логов
kubectl logs -f -l app.kubernetes.io/name=cloudbridge-relay -n cloudbridge

# Масштабирование
kubectl scale deployment cloudbridge-relay --replicas=5 -n cloudbridge

# Обновление образов
kubectl set image deployment/cloudbridge-relay cloudbridge-relay=localhost:5000/cloudbridge-relay:v2.0 -n cloudbridge
```

---

## Заключение

Данное руководство описывает полный процесс развертывания CloudBridge Relay в кластере Kubernetes VK Cloud. Следуя этим шагам, вы сможете:

1. Создать кластер Kubernetes в VK Cloud
2. становить все необходимые аддоны
3. оздать и загрузить Docker образы
4. астроить registry для хранения образов
5. азвернуть все сервисы CloudBridge Relay
6. астроить мониторинг и логирование
7. ротестировать работу системы

### Следующие шаги:

- Настройка CI/CD пайплайнов
- Настройка автоматического масштабирования
- Настройка резервного копирования
- Оптимизация производительности

**Удачного развертывания!**
