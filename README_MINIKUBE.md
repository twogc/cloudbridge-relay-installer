# 🚀 Быстрое развертывание в Minikube

## ⚡ Автоматическое развертывание

```bash
# Запуск полного развертывания одной командой
./scripts/deploy-to-minikube.sh
```

## 🔧 Ручное развертывание

### 1. Сборка и загрузка образов
```bash
# Сборка образов
docker build -t cloudbridge/relay-server:latest -f Dockerfile .
docker build -t cloudbridge/ai-service:latest -f Dockerfile.ai-service .

# Загрузка в Minikube
minikube image load cloudbridge/relay-server:latest
minikube image load cloudbridge/ai-service:latest
```

### 2. Включение AI
```bash
kubectl patch configmap cloudbridge-config -n cloudbridge --type merge -p '{
  "data": {
    "config.json": "{\"ai\":{\"enabled\":true,\"behavior_analysis\":{\"enabled\":true},\"predictive_security\":{\"enabled\":true},\"automated_response\":{\"enabled\":true},\"monitoring\":{\"enabled\":true}}}"
  }
}'
```

### 3. Обновление deployment
```bash
kubectl rollout restart deployment/cloudbridge-relay -n cloudbridge
kubectl rollout status deployment/cloudbridge-relay -n cloudbridge
```

### 4. Тестирование
```bash
# Port-forward
kubectl port-forward -n cloudbridge service/p2p-api-nodeport 8082:8082 &
kubectl port-forward -n cloudbridge service/ai-service-nodeport 8083:8083 &

# Тесты
curl http://localhost:8082/health
curl http://localhost:8083/health
curl http://localhost:8083/api/v1/ai/models
```

## 🧪 Только тестирование
```bash
./scripts/deploy-to-minikube.sh --test-only
```

## 📦 Только сборка
```bash
./scripts/deploy-to-minikube.sh --build-only
```

## 🔍 Проверка статуса
```bash
kubectl get pods -n cloudbridge
kubectl logs -n cloudbridge deployment/cloudbridge-relay | grep "AI Analytics"
```

## 📊 Доступ к сервисам
- **P2P API**: http://localhost:8082
- **AI Service**: http://localhost:8083
- **Grafana**: `kubectl port-forward -n cloudbridge-monitoring service/grafana 3000:3000`

## 🛠️ Устранение неполадок
```bash
# Остановка port-forward
pkill -f "kubectl port-forward"

# Перезапуск deployment
kubectl rollout restart deployment/cloudbridge-relay -n cloudbridge

# Проверка событий
kubectl get events -n cloudbridge --sort-by='.lastTimestamp'
```

---
**Полная инструкция**: [docs/MINIKUBE_DEPLOYMENT_GUIDE.md](docs/MINIKUBE_DEPLOYMENT_GUIDE.md)

