# 📚 Документация по развертыванию CloudBridge Relay в VK Cloud

## 📁 Структура документации

```
├── VK_CLOUD_DEPLOYMENT_GUIDE.md     # Подробное руководство по развертыванию
├── QUICK_DEPLOYMENT_CHECKLIST.md    # Быстрая шпаргалка
├── k8s-manifests/                   # Готовые Kubernetes манифесты
│   ├── 00-namespace.yaml            # Namespace
│   ├── 01-pvc.yaml                  # PersistentVolumeClaim
│   ├── 02-secrets.yaml              # Секреты
│   ├── 03-configmap.yaml            # ConfigMap
│   ├── 04-registry.yaml             # Docker Registry
│   ├── 05-deployments.yaml          # Deployments
│   ├── 06-services.yaml             # Services
│   ├── 07-ingress.yaml              # Ingress
│   └── 08-servicemonitor.yaml       # ServiceMonitor
├── deploy.sh                        # Скрипт автоматического развертывания
└── README_DEPLOYMENT.md             # Этот файл
```

## 🚀 Быстрый старт

### Автоматическое развертывание (рекомендуется)

```bash
# 1. Клонирование репозитория
git clone <repository-url>
cd cloudbridge-relay-installer

# 2. Настройка окружения
source openstack_auth.sh
export KUBECONFIG=~/.kube/config-vkcloud

# 3. Запуск автоматического развертывания
./deploy.sh
```

### Ручное развертывание

```bash
# 1. Создание namespace
kubectl apply -f k8s-manifests/00-namespace.yaml

# 2. Создание PVC
kubectl apply -f k8s-manifests/01-pvc.yaml

# 3. Создание секретов (обновите значения!)
kubectl apply -f k8s-manifests/02-secrets.yaml

# 4. Создание ConfigMap
kubectl apply -f k8s-manifests/03-configmap.yaml

# 5. Развертывание registry
kubectl apply -f k8s-manifests/04-registry.yaml

# 6. Ожидание готовности registry
kubectl wait --for=condition=Available deployment/docker-registry -n cloudbridge --timeout=300s

# 7. Загрузка образов
kubectl port-forward service/docker-registry 5000:5000 -n cloudbridge &
docker tag cloudbridge-relay:latest localhost:5000/cloudbridge-relay:latest
docker push localhost:5000/cloudbridge-relay:latest

# 8. Развертывание сервисов
kubectl apply -f k8s-manifests/05-deployments.yaml
kubectl apply -f k8s-manifests/06-services.yaml

# 9. Создание Ingress
kubectl apply -f k8s-manifests/07-ingress.yaml

# 10. Создание ServiceMonitor
kubectl apply -f k8s-manifests/08-servicemonitor.yaml
```

## 📋 Предварительные требования

### Системные требования
- Ubuntu 18.04+ или аналогичная ОС
- Docker 20.10+
- kubectl 1.28+
- Helm 3.12+
- Terraform 1.5+

### VK Cloud требования
- Аккаунт в VK Cloud
- Проект с доступом к Kubernetes
- OpenStack RC файл v3
- Достаточные квоты для создания кластера

## 🔧 Настройка

### 1. Настройка OpenStack CLI

```bash
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

### 2. Создание кластера Kubernetes

```bash
# Переход в директорию Terraform
cd terraform-k8s

# Инициализация Terraform
terraform init

# Планирование развертывания
terraform plan

# Развертывание кластера
terraform apply -auto-approve

# Настройка kubeconfig
export KUBECONFIG=~/.kube/config-vkcloud
kubectl get nodes
```

### 3. Установка аддонов

```bash
# Добавление Helm репозиториев
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo add jetstack https://charts.jetstack.io
helm repo update

# Установка Prometheus Stack
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring --create-namespace

# Установка NGINX Ingress
helm install ingress-nginx ingress-nginx/ingress-nginx \
  --namespace ingress-nginx --create-namespace

# Установка cert-manager
helm install cert-manager jetstack/cert-manager \
  --namespace cert-manager --create-namespace --set installCRDs=true
```

## 📊 Мониторинг

### Доступ к сервисам мониторинга

```bash
# Prometheus
kubectl port-forward service/prometheus-kube-prometheus-prometheus 9090:9090 -n monitoring &

# Grafana
kubectl port-forward service/prometheus-grafana 3000:80 -n monitoring &

# Jaeger
kubectl port-forward service/jaeger-query 16686:16686 -n jaeger &
```

### Полезные команды

```bash
# Статус всех ресурсов
kubectl get all -n cloudbridge

# Логи CloudBridge Relay
kubectl logs -f -l app.kubernetes.io/name=cloudbridge-relay -n cloudbridge

# События
kubectl get events -n cloudbridge --sort-by='.lastTimestamp'

# Масштабирование
kubectl scale deployment cloudbridge-relay --replicas=5 -n cloudbridge

# Обновление образа
kubectl set image deployment/cloudbridge-relay cloudbridge-relay=localhost:5000/cloudbridge-relay:v2.0 -n cloudbridge
```

## 🚨 Устранение неполадок

### Проблема: Pod в состоянии ImagePullBackOff

```bash
# Проверка секрета для registry
kubectl get secret k8s-registry-creds -n cloudbridge -o yaml

# Пересоздание секрета
kubectl delete secret k8s-registry-creds -n cloudbridge
kubectl create secret docker-registry k8s-registry-creds \
  --docker-server=http://docker-registry:5000 \
  --docker-username=admin \
  --docker-password=admin \
  --namespace=cloudbridge
```

### Проблема: PVC в состоянии Pending

```bash
# Проверка доступных storage classes
kubectl get storageclass

# Обновление PVC с правильным storage class
kubectl patch pvc cloudbridge-logs-pvc -n cloudbridge -p '{"spec":{"storageClassName":"csi-ceph-hdd-me1"}}'
```

### Проблема: Registry недоступен

```bash
# Проверка статуса registry
kubectl get pods -n cloudbridge | grep registry

# Перезапуск registry
kubectl rollout restart deployment docker-registry -n cloudbridge

# Проверка подключения
kubectl run test-pod --image=busybox --rm -it --restart=Never -n cloudbridge -- wget -qO- http://docker-registry:5000/v2/
```

## 📈 Масштабирование

### Горизонтальное масштабирование

```bash
# Увеличение количества реплик
kubectl scale deployment cloudbridge-relay --replicas=5 -n cloudbridge

# Автоматическое масштабирование
kubectl autoscale deployment cloudbridge-relay --min=2 --max=10 --cpu-percent=70 -n cloudbridge
```

### Вертикальное масштабирование

```bash
# Обновление ресурсов в deployment
kubectl patch deployment cloudbridge-relay -n cloudbridge -p '{"spec":{"template":{"spec":{"containers":[{"name":"cloudbridge-relay","resources":{"requests":{"memory":"1Gi","cpu":"500m"},"limits":{"memory":"2Gi","cpu":"1000m"}}}]}}}}'
```

## 🔄 Обновление

### Обновление образов

```bash
# Сборка нового образа
docker build -f docker/Dockerfile -t cloudbridge-relay:v2.0 .

# Загрузка в registry
kubectl port-forward service/docker-registry 5000:5000 -n cloudbridge &
docker tag cloudbridge-relay:v2.0 localhost:5000/cloudbridge-relay:v2.0
docker push localhost:5000/cloudbridge-relay:v2.0

# Обновление deployment
kubectl set image deployment/cloudbridge-relay cloudbridge-relay=localhost:5000/cloudbridge-relay:v2.0 -n cloudbridge
```

### Откат изменений

```bash
# Просмотр истории развертывания
kubectl rollout history deployment/cloudbridge-relay -n cloudbridge

# Откат к предыдущей версии
kubectl rollout undo deployment/cloudbridge-relay -n cloudbridge

# Откат к конкретной версии
kubectl rollout undo deployment/cloudbridge-relay --to-revision=2 -n cloudbridge
```

## 🗑️ Удаление

### Удаление сервисов

```bash
# Удаление всех ресурсов CloudBridge
kubectl delete namespace cloudbridge

# Удаление мониторинга
helm uninstall prometheus -n monitoring
helm uninstall ingress-nginx -n ingress-nginx
helm uninstall cert-manager -n cert-manager
```

### Удаление кластера

```bash
# Удаление кластера через Terraform
cd terraform-k8s
terraform destroy -auto-approve
```

## 📞 Поддержка

При возникновении проблем:

1. Проверьте логи: `kubectl logs -f -l app.kubernetes.io/name=cloudbridge-relay -n cloudbridge`
2. Проверьте события: `kubectl get events -n cloudbridge --sort-by='.lastTimestamp'`
3. Проверьте статус ресурсов: `kubectl get all -n cloudbridge`
4. Обратитесь к документации VK Cloud
5. Создайте issue в репозитории проекта

---

**Удачного развертывания! 🚀**
