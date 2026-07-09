# CloudBridge Relay - Развертывание в VK Cloud Kubernetes

> **SoT (Source of Truth) for pilot / lab / L0** — this repository.  
> Companion client: `cloudbridge-client` (`scripts/all-smoke.sh`).  
> `cloudbridge-scalable-relay` is multi-region **WIP**, not deployable main (see strategy `RELAY_SOT_POLICY.md`).

## Обзор проекта

CloudBridge Relay - это высокопроизводительная система для ретрансляции и маршрутизации сетевого трафика с поддержкой:

- **P2P соединения** - прямое взаимодействие между клиентами
- **WireGuard VPN** - безопасные туннели
- **AI/ML сервисы** - интеллектуальная обработка трафика
- **Workflow Engine** - автоматизация процессов
- **Мониторинг** - полная наблюдаемость системы

## Быстрый старт

### Автоматическое развертывание (рекомендуется)

```bash
# 1. Клонирование репозитория
git clone <repository-url>
cd cloudbridge-relay-installer

# 2. Настройка зеркал Docker Hub (для России)
sudo ./scripts/deploy/setup-docker-mirrors.sh

# 3. Настройка окружения
source scripts/deploy/openstack_auth.sh
export KUBECONFIG=~/.kube/config-vkcloud

# 4. Запуск автоматического развертывания
./scripts/deploy/deploy.sh
```

### Ручное развертывание

```bash
# 1. Создание кластера
cd terraform-k8s
terraform init
terraform apply -auto-approve

# 2. Установка аддонов
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus prometheus-community/kube-prometheus-stack --namespace monitoring --create-namespace

# 3. Развертывание сервисов
kubectl apply -f k8s-manifests/
```

## 📚 Документация

См. **[docs/README.md](docs/README.md)** — навигация по всей документации. Для агентов: **[openwiki/quickstart.md](openwiki/quickstart.md)**.

### Основные руководства

- **[docs/guides/VK_CLOUD_DEPLOYMENT_GUIDE.md](docs/guides/VK_CLOUD_DEPLOYMENT_GUIDE.md)** — развертывание в VK Cloud
- **[docs/guides/QUICK_DEPLOYMENT_CHECKLIST.md](docs/guides/QUICK_DEPLOYMENT_CHECKLIST.md)** — быстрая шпаргалка
- **[docs/guides/DOCKER_MIRRORS_SETUP.md](docs/guides/DOCKER_MIRRORS_SETUP.md)** — зеркала Docker Hub
- **[docs/guides/README_DEPLOYMENT.md](docs/guides/README_DEPLOYMENT.md)** — обзор deployment-документации

### Готовые манифесты

```
k8s-manifests/
├── 00-namespace.yaml            # Namespace
├── 01-pvc.yaml                  # PersistentVolumeClaim
├── 02-secrets.yaml              # Секреты
├── 03-configmap.yaml            # ConfigMap
├── 04-registry.yaml             # Docker Registry
├── 05-deployments.yaml          # Deployments
├── 06-services.yaml             # Services
├── 07-ingress.yaml              # Ingress
└── 08-servicemonitor.yaml       # ServiceMonitor
```

### Автоматические скрипты

- **`scripts/deploy/deploy.sh`** — полное автоматическое развертывание
- **`scripts/deploy/setup-docker-mirrors.sh`** — настройка зеркал Docker Hub

## 🛠️ Предварительные требования

### Системные требования
- Ubuntu 18.04+ или аналогичная ОС
- Docker 20.10+ с настроенными зеркалами
- kubectl 1.28+
- Helm 3.12+
- Terraform 1.5+

### VK Cloud требования
- Аккаунт в VK Cloud
- Проект с доступом к Kubernetes
- OpenStack RC файл v3
- Достаточные квоты для создания кластера

## 🔧 Настройка

### 1. Настройка зеркал Docker Hub (для России)

```bash
# Автоматическая настройка
sudo ./scripts/deploy/setup-docker-mirrors.sh

# Проверка статуса зеркал
./scripts/deploy/setup-docker-mirrors.sh --status

# Тестирование
./scripts/deploy/setup-docker-mirrors.sh --test
```

### 2. Настройка OpenStack CLI

```bash
# Создание файла аутентификации
cat > scripts/deploy/openstack_auth.sh << 'EOF'
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

chmod +x scripts/deploy/openstack_auth.sh
```

### 3. Создание кластера Kubernetes

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

## Мониторинг

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

## Устранение неполадок

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

### Проблема: Docker Hub недоступен

```bash
# Проверка зеркал
./scripts/deploy/setup-docker-mirrors.sh --status

# Перезапуск Docker
sudo systemctl restart docker

# Тестирование загрузки
docker pull alpine:latest
```

## Масштабирование

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

## 🏗️ Архитектура

### Компоненты системы

```
┌─────────────────────────────────────────────────────────────┐
│                    CloudBridge Relay                       │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │   P2P API   │  │  WireGuard  │  │   HTTP/3    │        │
│  │   Server    │  │   Server    │  │   Server    │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │ AI Service  │  │   Cadence   │  │ PostgreSQL │        │
│  │             │  │ Workflow   │  │             │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │ Prometheus  │  │   Grafana   │  │   Jaeger   │        │
│  │             │  │             │  │             │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────────────────────────────────────────────┘
```

### Сетевые порты

> Полная таблица: [openwiki/port-scheme.md](openwiki/port-scheme.md).

| Сервис | Порт | Протокол | Описание |
|--------|------|----------|----------|
| CloudBridge Relay | 9090 | TCP | HTTP API (`http_api.port`) |
| CloudBridge Relay | 9090 | UDP | QUIC transport |
| CloudBridge Relay | 5552 | TCP | P2P REST API |
| CloudBridge Relay | **5551** | TCP | Prometheus metrics |
| CloudBridge Relay | 51820 | UDP | WireGuard |
| AI Service | **8083** | TCP | HTTP + `/metrics` |
| AI Worker | **8085** | TCP | Health / metrics |
| IPAM | **8080** | TCP | IP allocation |
| Cadence | 7933–7939 | gRPC | Workflow API |
| PostgreSQL | 5432 | TCP | База данных |
| Cassandra | 9042 | TCP | NoSQL база |

## 📞 Поддержка

При возникновении проблем:

1. Проверьте логи: `kubectl logs -f -l app.kubernetes.io/name=cloudbridge-relay -n cloudbridge`
2. Проверьте события: `kubectl get events -n cloudbridge --sort-by='.lastTimestamp'`
3. Проверьте статус ресурсов: `kubectl get all -n cloudbridge`
4. Проверьте зеркала Docker: `./scripts/deploy/setup-docker-mirrors.sh --status`
5. Обратитесь к документации VK Cloud
6. Создайте issue в репозитории проекта

## Лицензия

Этот проект распространяется под лицензией MIT. См. файл [LICENSE](LICENSE) для подробностей.

## 🤝 Вклад в проект

Мы приветствуем вклад в развитие проекта! Пожалуйста, ознакомьтесь с [CONTRIBUTING.md](CONTRIBUTING.md) для получения информации о том, как внести свой вклад.

---

**Удачного развертывания! 🚀**