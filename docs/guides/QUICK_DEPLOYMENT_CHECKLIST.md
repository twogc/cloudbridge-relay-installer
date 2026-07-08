# ⚡ Быстрая шпаргалка по развертыванию CloudBridge Relay в VK Cloud

## Быстрый старт (5 минут)

### 1. Подготовка окружения
```bash
# Установка инструментов
curl -fsSL https://get.docker.com | sh
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Настройка зеркал Docker Hub (для России)
sudo ./setup-docker-mirrors.sh

# Настройка OpenStack
source openstack_auth.sh
```

### 2. Создание кластера
```bash
cd terraform-k8s
terraform init
terraform apply -auto-approve
export KUBECONFIG=~/.kube/config-vkcloud
kubectl get nodes
```

### 3. Установка аддонов
```bash
# Helm репозитории
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo add jetstack https://charts.jetstack.io
helm repo update

# Установка компонентов
helm install prometheus prometheus-community/kube-prometheus-stack --namespace monitoring --create-namespace
helm install ingress-nginx ingress-nginx/ingress-nginx --namespace ingress-nginx --create-namespace
helm install cert-manager jetstack/cert-manager --namespace cert-manager --create-namespace --set installCRDs=true
```

### 4. Создание образов
```bash
# Сборка образов
docker build -f docker/Dockerfile -t cloudbridge-relay:latest .
docker build -f /tmp/ai-service.Dockerfile -t ai-service:latest .
docker build -f /tmp/cadence.Dockerfile -t cadence:latest .
```

### 5. Настройка Registry
```bash
# Создание namespace
kubectl create namespace cloudbridge

# Создание PVC
kubectl apply -f registry-pvc.yaml

# Развертывание registry
kubectl apply -f registry-deployment.yaml

# Port-forward для доступа
kubectl port-forward service/docker-registry 5000:5000 -n cloudbridge &

# Загрузка образов
docker tag cloudbridge-relay:latest localhost:5000/cloudbridge-relay:latest
docker tag ai-service:latest localhost:5000/ai-service:latest
docker tag cadence:latest localhost:5000/cadence:latest

docker push localhost:5000/cloudbridge-relay:latest
docker push localhost:5000/ai-service:latest
docker push localhost:5000/cadence:latest
```

### 6. Развертывание сервисов
```bash
# Создание секретов
kubectl create secret docker-registry k8s-registry-creds --docker-server=http://localhost:5000 --docker-username=admin --docker-password=admin --namespace=cloudbridge

# Развертывание манифестов
kubectl apply -f k8s/
```

## Чек-лист проверки

### Кластер готов
- [ ] `kubectl get nodes` - 3 узла в состоянии Ready
- [ ] `kubectl get pods -n monitoring` - Prometheus Stack запущен
- [ ] `kubectl get pods -n ingress-nginx` - NGINX Ingress запущен

### egistry работает
- [ ] `kubectl get pods -n cloudbridge | grep registry` - Running
- [ ] `kubectl port-forward service/docker-registry 5000:5000 -n cloudbridge &`
- [ ] `curl http://localhost:5000/v2/` - возвращает `{}`

### бразы загружены
- [ ] `docker push localhost:5000/cloudbridge-relay:latest` - успешно
- [ ] `docker push localhost:5000/ai-service:latest` - успешно
- [ ] `docker push localhost:5000/cadence:latest` - успешно

### ервисы запущены
- [ ] `kubectl get pods -n cloudbridge` - все поды в состоянии Running
- [ ] `kubectl get services -n cloudbridge` - сервисы созданы
- [ ] `kubectl logs -l app.kubernetes.io/name=cloudbridge-relay -n cloudbridge` - логи без ошибок

##  Быстрые команды

### Проверка статуса
```bash
# Общий статус
kubectl get all -n cloudbridge

# Логи CloudBridge Relay
kubectl logs -f -l app.kubernetes.io/name=cloudbridge-relay -n cloudbridge

# События
kubectl get events -n cloudbridge --sort-by='.lastTimestamp'
```

### Масштабирование
```bash
# Увеличить количество реплик
kubectl scale deployment cloudbridge-relay --replicas=5 -n cloudbridge

# Автоматическое масштабирование
kubectl autoscale deployment cloudbridge-relay --min=2 --max=10 --cpu-percent=70 -n cloudbridge
```

### Обновление
```bash
# Обновление образа
kubectl set image deployment/cloudbridge-relay cloudbridge-relay=localhost:5000/cloudbridge-relay:v2.0 -n cloudbridge

# Откат
kubectl rollout undo deployment/cloudbridge-relay -n cloudbridge
```

### Мониторинг
```bash
# Prometheus
kubectl port-forward service/prometheus-kube-prometheus-prometheus 9090:9090 -n monitoring &

# Grafana
kubectl port-forward service/prometheus-grafana 3000:80 -n monitoring &

# Jaeger
kubectl port-forward service/jaeger-query 16686:16686 -n jaeger &
```

## 🚨 Устранение неполадок

### Проблема: Pod в состоянии ImagePullBackOff
```bash
# Проверка секрета
kubectl get secret k8s-registry-creds -n cloudbridge -o yaml

# Пересоздание секрета
kubectl delete secret k8s-registry-creds -n cloudbridge
kubectl create secret docker-registry k8s-registry-creds --docker-server=http://docker-registry:5000 --docker-username=admin --docker-password=admin --namespace=cloudbridge
```

### Проблема: PVC в состоянии Pending
```bash
# Проверка storage classes
kubectl get storageclass

# Обновление PVC
kubectl patch pvc cloudbridge-logs-pvc -n cloudbridge -p '{"spec":{"storageClassName":"csi-ceph-hdd-me1"}}'
```

### Проблема: Registry недоступен
```bash
# Проверка registry
kubectl get pods -n cloudbridge | grep registry

# Перезапуск registry
kubectl rollout restart deployment docker-registry -n cloudbridge

# Проверка подключения
kubectl run test-pod --image=busybox --rm -it --restart=Never -n cloudbridge -- wget -qO- http://docker-registry:5000/v2/
```

## Полезные ссылки

- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin / полученный пароль)
- **Jaeger**: http://localhost:16686
- **ArgoCD**: http://localhost:8080

## Следующие шаги

1. **Настройка CI/CD** - автоматическое развертывание
2. **Мониторинг** - настройка алертов
3. **Безопасность** - настройка RBAC и Network Policies
4. **Резервное копирование** - настройка Velero
5. **Масштабирование** - настройка HPA и VPA

---

**Удачного развертывания!**
