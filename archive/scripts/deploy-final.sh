#!/bin/bash
# CloudBridge Relay - Финальное развертывание по MINIKUBE_BUILD_GUIDE.md

set -e

echo "Финальное развертывание CloudBridge Relay по MINIKUBE_BUILD_GUIDE.md..."
echo ""

# Проверка предварительных условий
echo "Проверка предварительных условий..."
if ! command -v kubectl &> /dev/null; then
    echo "kubectl не найден. Установите kubectl."
    exit 1
fi

if ! command -v minikube &> /dev/null; then
    echo "minikube не найден. Установите minikube."
    exit 1
fi

if ! command -v docker &> /dev/null; then
    echo "docker не найден. Установите docker."
    exit 1
fi

# Проверка статуса Minikube
if ! minikube status &> /dev/null; then
    echo "Minikube не запущен. Запустите: minikube start --driver=docker --memory=4096 --cpus=2"
    exit 1
fi

echo "Предварительные условия выполнены"

# Переход в директорию проекта
cd /home/ubuntu/cloudbridge-relay-installer

# 1. Создание уникального тега (согласно MINIKUBE_BUILD_GUIDE.md)
echo "🏷️ Создание уникального тега образа..."
TAG="turn-pool-$(date +%Y%m%d-%H%M%S)"
echo "Используем тег: $TAG"

# 2. Сборка Docker образа (согласно MINIKUBE_BUILD_GUIDE.md)
echo "🔨 Сборка Docker образа..."
docker build -t cloudbridge-relay:$TAG .

# Проверка собранного образа
echo "Проверка собранного образа..."
docker images | grep cloudbridge-relay

# 3. Загрузка в Minikube (согласно MINIKUBE_BUILD_GUIDE.md)
echo "⬆️ Загрузка образа в Minikube..."
minikube image load cloudbridge-relay:$TAG

# Проверка загруженных образов
echo "Проверка загруженных образов..."
minikube image ls | grep cloudbridge-relay

# 4. Создание namespace
echo "📁 Создание namespace..."
kubectl create namespace cloudbridge --dry-run=client -o yaml | kubectl apply -f -

# 5. Развертывание ConfigMaps и Secrets
echo "⚙️ Развертывание ConfigMaps и Secrets..."
kubectl apply -f k8s/configmaps/
kubectl apply -f k8s/ai-telegram-config.yaml
kubectl apply -f k8s/ai-telegram-secret.yaml

# Создание базовых secrets
kubectl create secret generic cloudbridge-secrets -n cloudbridge \
  --from-literal=cassandra-username=cassandra \
  --from-literal=cassandra-password=cassandra \
  --from-literal=api-key=test-api-key \
  --from-literal=jwt-secret=85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE= \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl create secret generic wireguard-keys -n cloudbridge \
  --from-literal=private-key=test-private-key \
  --from-literal=public-key=test-public-key \
  --dry-run=client -o yaml | kubectl apply -f -

# 6. Развертывание Cassandra
echo "🗄️ Развертывание Cassandra..."
cat > /tmp/cassandra-simple.yaml << 'EOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cassandra
  namespace: cloudbridge
spec:
  replicas: 1
  selector:
    matchLabels:
      app: cassandra
  template:
    metadata:
      labels:
        app: cassandra
    spec:
      containers:
      - name: cassandra
        image: cassandra:3.11
        ports:
        - containerPort: 9042
          name: cql
        - containerPort: 9160
          name: thrift
        env:
        - name: CASSANDRA_CLUSTER_NAME
          value: "cloudbridge-cluster"
        - name: CASSANDRA_DC
          value: "datacenter1"
        - name: CASSANDRA_RACK
          value: "rack1"
        - name: HEAP_NEWSIZE
          value: "128M"
        - name: MAX_HEAP_SIZE
          value: "512M"
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "500m"
        readinessProbe:
          exec:
            command:
            - /bin/bash
            - -c
            - "cqlsh -e 'describe keyspaces'"
          initialDelaySeconds: 60
          periodSeconds: 30
          timeoutSeconds: 10
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
  - name: cql
    port: 9042
    targetPort: 9042
  - name: thrift
    port: 9160
    targetPort: 9160
EOF

kubectl apply -f /tmp/cassandra-simple.yaml

# 7. Развертывание CloudBridge Relay с новым тегом (согласно MINIKUBE_BUILD_GUIDE.md)
echo "Развертывание CloudBridge Relay с тегом $TAG..."
cat > /tmp/cloudbridge-relay-deployment.yaml << EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cloudbridge-relay
  namespace: cloudbridge
spec:
  replicas: 2
  selector:
    matchLabels:
      app: cloudbridge-relay
  template:
    metadata:
      labels:
        app: cloudbridge-relay
    spec:
      containers:
      - name: cloudbridge-relay
        image: cloudbridge-relay:$TAG
        imagePullPolicy: Never
        ports:
        - containerPort: 5550
          name: main-server
        - containerPort: 5552
          name: p2p-api
        - containerPort: 5553
          name: http-api
        - containerPort: 8443
          name: https-masque
        - containerPort: 9090
          name: quic-transport
        - containerPort: 9092
          name: enhanced-quic
        - containerPort: 19302
          name: stun-server
        - containerPort: 5554
          name: udp-transport
        env:
        - name: SERVER_HOST
          value: "0.0.0.0"
        - name: SERVER_PORT
          value: "5550"
        - name: LOG_LEVEL
          value: "info"
        - name: P2P_ENABLED
          value: "true"
        - name: ICE_ENABLED
          value: "true"
        - name: MASQUE_ENABLED
          value: "true"
        - name: RELAY_JWT_SECRET
          value: "85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE="
        - name: TURN_PASSWORD
          value: "cloudbridge-turn-secret"
        - name: TURN_BACKUP_PASSWORD
          value: "cloudbridge-backup-secret"
        - name: CONFIG_PATH
          value: "/app/configs/config.json"
        resources:
          limits:
            cpu: 500m
            memory: 512Mi
          requests:
            cpu: 250m
            memory: 256Mi
        volumeMounts:
        - name: config
          mountPath: /app/configs
          readOnly: true
        command: ["./cloudbridge-relay"]
        args: ["--config", "/app/configs/config.json"]
      volumes:
      - name: config
        configMap:
          name: cloudbridge-relay-config
EOF

kubectl apply -f /tmp/cloudbridge-relay-deployment.yaml

# 8. Развертывание AI Service
echo " Развертывание AI Service..."
cat > /tmp/ai-service-deployment.yaml << 'EOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ai-service
  namespace: cloudbridge
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ai-service
  template:
    metadata:
      labels:
        app: ai-service
    spec:
      containers:
      - name: ai-service
        image: ai-service:latest
        imagePullPolicy: Never
        ports:
        - name: http
          containerPort: 8083
        - name: metrics
          containerPort: 8084
        env:
        - name: AI_PORT
          value: "8083"
        - name: LOG_LEVEL
          value: "info"
        - name: GIN_MODE
          value: "release"
        - name: CADENCE_HOST
          value: "cadence-service"
        - name: CADENCE_PORT
          value: "7933"
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
            port: 8083
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8083
          initialDelaySeconds: 5
          periodSeconds: 5
        volumeMounts:
        - name: ai-config
          mountPath: /app/config
          readOnly: true
      volumes:
      - name: ai-config
        configMap:
          name: ai-telegram-config
---
apiVersion: v1
kind: Service
metadata:
  name: ai-service
  namespace: cloudbridge
spec:
  selector:
    app: ai-service
  ports:
  - name: http
    port: 8083
    targetPort: 8083
  - name: metrics
    port: 8084
    targetPort: 8084
EOF

kubectl apply -f /tmp/ai-service-deployment.yaml

# 9. Развертывание Services (согласно MINIKUBE_BUILD_GUIDE.md)
echo "Развертывание Services..."
kubectl apply -f k8s/cloudbridge-relay-external-service.yaml -n cloudbridge

# Создание дополнительных сервисов
cat > /tmp/additional-services.yaml << 'EOF'
apiVersion: v1
kind: Service
metadata:
  name: cloudbridge-relay
  namespace: cloudbridge
spec:
  selector:
    app: cloudbridge-relay
  ports:
  - name: main-server
    port: 5550
    targetPort: 5550
  - name: p2p-api
    port: 5552
    targetPort: 5552
  - name: http-api
    port: 5553
    targetPort: 5553
  - name: quic-transport
    port: 9090
    targetPort: 9090
  - name: stun-server
    port: 19302
    targetPort: 19302
  type: ClusterIP
---
apiVersion: v1
kind: Service
metadata:
  name: cloudbridge-relay-nodeport
  namespace: cloudbridge
spec:
  selector:
    app: cloudbridge-relay
  ports:
  - name: http-api
    port: 5553
    targetPort: 5553
    nodePort: 30553
  - name: p2p-api
    port: 5552
    targetPort: 5552
    nodePort: 30552
  - name: quic-transport
    port: 9090
    targetPort: 9090
    nodePort: 30990
  type: NodePort
EOF

kubectl apply -f /tmp/additional-services.yaml

# 10. Инициализация Cassandra
echo " Инициализация Cassandra для Cadence..."
echo "Ожидание готовности Cassandra..."
kubectl wait --for=condition=ready pod -l app=cassandra -n cloudbridge --timeout=300s

echo "Создание keyspaces для Cadence..."
kubectl exec -n cloudbridge deployment/cassandra -- cqlsh -e "CREATE KEYSPACE IF NOT EXISTS cadence WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': 1};"
kubectl exec -n cloudbridge deployment/cassandra -- cqlsh -e "CREATE KEYSPACE IF NOT EXISTS cadence_visibility WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': 1};"

echo "🗃️ Инициализация схемы Cadence..."
kubectl exec -n cloudbridge deployment/cassandra -- cqlsh -e "USE cadence; CREATE TABLE IF NOT EXISTS schema_version (version int, dirty boolean, PRIMARY KEY (version));"
kubectl exec -n cloudbridge deployment/cassandra -- cqlsh -e "USE cadence; INSERT INTO schema_version (version, dirty) VALUES (0, false) IF NOT EXISTS;"

# 11. Проверка статуса
echo "Развертывание завершено!"
echo ""
echo "Статус подов:"
kubectl get pods -n cloudbridge

echo ""
echo "Доступные сервисы:"
kubectl get svc -n cloudbridge

echo ""
echo "Команды для доступа (согласно MINIKUBE_BUILD_GUIDE.md):"
echo "  CloudBridge Relay: kubectl port-forward svc/cloudbridge-relay 9090:9090 -n cloudbridge"
echo "  AI Service:        kubectl port-forward svc/ai-service 8083:8083 -n cloudbridge"
echo "  Cassandra:         kubectl port-forward svc/cassandra-service 9042:9042 -n cloudbridge"

echo ""
echo "🧪 Тестирование (согласно MINIKUBE_BUILD_GUIDE.md):"
echo "  Health check:      curl http://localhost:9090/health"
echo "  AI Service:        curl http://localhost:8083/health"

echo ""
echo "🏷️ Использованный тег образа: $TAG"
echo "CloudBridge Relay с TURN Pool успешно развернут по MINIKUBE_BUILD_GUIDE.md!"

# 12. Проверка TURN Pool
echo ""
echo "🔄 Проверка TURN Pool..."
sleep 10
kubectl logs -n cloudbridge deployment/cloudbridge-relay | grep -i "turn\|pool" | tail -5 || echo "TURN Pool логи не найдены"
