# 🐳 **CloudBridge Relay - Инструкция по сборке Minikube**

## 🎯 **Цель**

Создать полноценную среду разработки и тестирования CloudBridge Relay в Minikube с поддержкой всех протоколов (QUIC, ICE/STUN/TURN, MASQUE) и мультитунельности.

## 📋 **Предварительные требования**

### **Системные требования:**
- **OS:** Ubuntu 20.04+ / CentOS 8+ / macOS 10.15+
- **RAM:** Минимум 4GB, рекомендуется 8GB+
- **CPU:** Минимум 2 ядра, рекомендуется 4+
- **Disk:** Минимум 20GB свободного места
- **Docker:** Версия 20.10+

### **Проверка системы:**
```bash
# Проверить версию ОС
cat /etc/os-release

# Проверить RAM
free -h

# Проверить CPU
nproc

# Проверить Docker
docker --version
docker info
```

## 🚀 **Установка Minikube**

### **1. Установка kubectl:**

```bash
# Скачать kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"

# Установить kubectl
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Проверить установку
kubectl version --client
```

### **2. Установка Minikube:**

```bash
# Скачать Minikube
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64

# Установить Minikube
sudo install minikube-linux-amd64 /usr/local/bin/minikube

# Проверить установку
minikube version
```

### **3. Запуск Minikube:**

```bash
# Запустить Minikube с оптимальными настройками
minikube start \
  --driver=docker \
  --memory=4096 \
  --cpus=2 \
  --disk-size=20g \
  --kubernetes-version=v1.28.0 \
  --addons=ingress \
  --addons=metrics-server

# Проверить статус
minikube status

# Проверить кластер
kubectl get nodes
kubectl get pods -A
```

## 🏗️ **Сборка CloudBridge Relay**

### **1. Подготовка проекта:**

```bash
# Перейти в директорию проекта
cd /home/ubuntu/cloudbridge-relay-installer

# Проверить структуру проекта
ls -la

# Проверить Dockerfile
cat Dockerfile
```

### **2. Сборка Docker образа:**

```bash
# Собрать образ с новым тегом
docker build -t cloudbridge-relay:quic-ice-v2 .

# Проверить собранный образ
docker images | grep cloudbridge-relay

# Проверить размер образа
docker images cloudbridge-relay:quic-ice-v2
```

### **3. Загрузка образа в Minikube:**

```bash
# Загрузить образ в Minikube
minikube image load cloudbridge-relay:quic-ice-v2

# Проверить загруженные образы
minikube image ls | grep cloudbridge-relay

# Альтернативный способ - использовать registry Minikube
eval $(minikube docker-env)
docker build -t cloudbridge-relay:quic-ice-v2 .
```

## ⚙️ **Конфигурация Kubernetes**

### **1. Создание namespace:**

```bash
# Создать namespace для CloudBridge
kubectl create namespace cloudbridge

# Проверить namespace
kubectl get namespaces | grep cloudbridge
```

### **2. Применение конфигураций:**

```bash
# Применить deployment
kubectl apply -f k8s/deployment.yaml -n cloudbridge

# Применить service
kubectl apply -f k8s/cloudbridge-relay-external-service.yaml -n cloudbridge

# Проверить статус
kubectl get pods -n cloudbridge
kubectl get services -n cloudbridge
```

### **3. Проверка конфигурации:**

```bash
# Проверить deployment
kubectl describe deployment cloudbridge-relay -n cloudbridge

# Проверить service
kubectl describe service cloudbridge-relay-external -n cloudbridge

# Проверить pods
kubectl get pods -n cloudbridge -o wide
```

## 🔧 **Настройка внешнего доступа**

### **1. Получение IP адресов:**

```bash
# Получить IP Minikube
minikube ip

# Получить IP LoadBalancer
kubectl get services -n cloudbridge -o wide

# Получить NodePort
kubectl get services -n cloudbridge
```

### **2. Настройка портов:**

```bash
# Проверить открытые порты
kubectl get services -n cloudbridge

# Пример вывода:
# NAME                          TYPE           CLUSTER-IP      EXTERNAL-IP   PORT(S)
# cloudbridge-relay-external    LoadBalancer   10.96.0.100     <pending>     8080:30091/TCP,8082:30882/TCP,19302:30258/UDP,9090:31284/UDP,8443:30443/TCP,9092:30467/UDP
```

### **3. Тестирование доступа:**

```bash
# Тестировать HTTP API
curl http://$(minikube ip):30091/health

# Тестировать P2P API
curl http://$(minikube ip):30882/health

# Тестировать STUN (UDP)
echo "test" | nc -u $(minikube ip) 30258
```

## 🚀 **Запуск и тестирование**

### **1. Проверка логов:**

```bash
# Получить имя pod
POD_NAME=$(kubectl get pods -n cloudbridge -o jsonpath='{.items[0].metadata.name}')

# Проверить логи
kubectl logs $POD_NAME -n cloudbridge

# Следить за логами в реальном времени
kubectl logs -f $POD_NAME -n cloudbridge
```

### **2. Port forwarding для тестирования:**

```bash
# Port forwarding для HTTP API
kubectl port-forward service/cloudbridge-relay-external 8080:8080 -n cloudbridge &

# Port forwarding для P2P API
kubectl port-forward service/cloudbridge-relay-external 8082:8082 -n cloudbridge &

# Проверить процессы port forwarding
ps aux | grep kubectl
```

### **3. Тестирование API:**

```bash
# Тестировать HTTP API
curl http://localhost:8080/health

# Тестировать P2P API
curl http://localhost:8082/health

# Тестировать с JWT токеном
curl -X POST "http://localhost:8082/api/v1/tenants/tenant-216420165/peers/register" \
  -H "Authorization: Bearer $(cat official-token.txt)" \
  -H "Content-Type: application/json" \
  -d '{"public_key": "test-key", "allowed_ips": ["10.0.0.1/32"]}'
```

## 🔄 **Обновление и перезапуск**

### **1. Обновление образа:**

```bash
# Пересобрать образ
docker build -t cloudbridge-relay:quic-ice-v3 .

# Загрузить в Minikube
minikube image load cloudbridge-relay:quic-ice-v3

# Обновить deployment
kubectl set image deployment/cloudbridge-relay cloudbridge-relay=cloudbridge-relay:quic-ice-v3 -n cloudbridge

# Проверить обновление
kubectl rollout status deployment/cloudbridge-relay -n cloudbridge
```

### **2. Перезапуск сервиса:**

```bash
# Перезапустить deployment
kubectl rollout restart deployment/cloudbridge-relay -n cloudbridge

# Проверить статус
kubectl get pods -n cloudbridge

# Проверить логи после перезапуска
kubectl logs -f deployment/cloudbridge-relay -n cloudbridge
```

## 🐛 **Отладка и диагностика**

### **1. Проверка ресурсов:**

```bash
# Проверить использование ресурсов
kubectl top pods -n cloudbridge
kubectl top nodes

# Проверить события
kubectl get events -n cloudbridge --sort-by='.lastTimestamp'
```

### **2. Проверка сети:**

```bash
# Проверить сетевые политики
kubectl get networkpolicies -n cloudbridge

# Проверить endpoints
kubectl get endpoints -n cloudbridge

# Проверить ingress
kubectl get ingress -n cloudbridge
```

### **3. Проверка конфигурации:**

```bash
# Проверить ConfigMap
kubectl get configmaps -n cloudbridge
kubectl describe configmap cloudbridge-config -n cloudbridge

# Проверить Secrets
kubectl get secrets -n cloudbridge
kubectl describe secret cloudbridge-secrets -n cloudbridge
```

## 📊 **Мониторинг и метрики**

### **1. Включение метрик:**

```bash
# Проверить metrics-server
kubectl get pods -n kube-system | grep metrics-server

# Если не установлен, установить
minikube addons enable metrics-server

# Проверить метрики
kubectl top pods -n cloudbridge
```

### **2. Логирование:**

```bash
# Настроить логирование
kubectl logs -f deployment/cloudbridge-relay -n cloudbridge --tail=100

# Сохранить логи в файл
kubectl logs deployment/cloudbridge-relay -n cloudbridge > cloudbridge.log
```

## 🧪 **Тестирование P2P соединений**

### **1. Подготовка тестовых клиентов:**

```bash
# Создать тестовый JWT токен
cat > test-token.txt << 'EOF'
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6ImZhbGxiYWNrLWtleSJ9.eyJwcm90b2NvbF90eXBlIjoicDJwLW1lc2giLCJzY29wZSI6InAycC1tZXNoLWNsYWltcyIsIm9yZ19pZCI6InRlbmFudC0yMTY0MjAxNjUiLCJ0ZW5hbnRfaWQiOiJ0ZW5hbnQtMjE2NDIwMTY1Iiwic2VydmVyX2lkIjoic2VydmVyLTE3NTgyMTkxOTU3MDMiLCJjb25uZWN0aW9uX3R5cGUiOiJxdWljLWljZSIsIm1heF9wZWVycyI6IjEwIiwicGVybWlzc2lvbnMiOlsibWVzaF9qb2luIiwibWVzaF9tYW5hZ2UiXSwibmV0d29ya19jb25maWciOnsic3VibmV0IjoiMTAuMC4wLjAvMjQiLCJnYXRld2F5IjoiMTAuMC4wLjEiLCJkbnMiOlsiOC44LjguOCIsIjEuMS4xLjEiXSwibXR1IjoxNDIwLCJmaXJld2FsbF9ydWxlcyI6WyJhbGxvd19zc2giLCJhbGxvd19odHRwIl0sImVuYWJsZV9pcHY2IjpmYWxzZX0sInF1aWNfY29uZmlnIjp7ImhhbmRzaGFrZV90aW1lb3V0IjoiMTBzIiwiaWRsZV90aW1lb3V0IjoiMzBzIiwibWF4X3N0cmVhbXMiOjEwMCwiY29ubmVjdGlvbl9taWdyYXRpb24iOnRydWUsIm11bHRpcGxleGluZyI6dHJ1ZX0sImljZV9jb25maWciOnsic3R1bl9zZXJ2ZXJzIjpbImVkZ2UuMmdjLnJ1OjE5MzAyIl0sInR1cm5fc2VydmVycyI6W10sImljZV90aW1lb3V0IjoiMzBzIiwiY29ubmVjdGlvbl90aW1lb3V0IjoiMTBzIiwiZ2F0aGVyaW5nX3RpbWVvdXQiOiI1cyIsImNhbmRpZGF0ZV90eXBlcyI6WyJob3N0Iiwic3JmbHgiLCJyZWxheSJdfSwibWFzcXVlX2NvbmZpZyI6eyJlbmFibGVkIjp0cnVlLCJwcm94eV91cmwiOiJodHRwczovL2VkZ2UuMmdjLnJ1Ojg0NDMiLCJhbHBuX3Byb3RvY29scyI6WyJoMyIsImgzLTI5Il0sImZhbGxiYWNrX2VuYWJsZWQiOnRydWV9LCJtZXNoX2NvbmZpZyI6eyJuZXR3b3JrX2lkIjoibWVzaC1uZXR3b3JrLTAwMSIsInN1Ym5ldCI6IjEwLjAuMC4wLzE2IiwicmVnaXN0cnlfdXJsIjoiaHR0cHM6Ly9lZGdlLjJnYy5ydTo4MDgwIiwiaGVhcnRiZWF0X2ludGVydmFsIjoiMzBzIiwibWF4X3BlZXJzIjoxMCwicm91dGluZ19zdHJhdGVneSI6InBlcmZvcm1hbmNlX29wdGltYWwiLCJlbmFibGVfYXV0b19kaXNjb3ZlcnkiOnRydWUsInRydXN0X2xldmVsIjoiYmFzaWMifSwicGVlcl93aGl0ZWxpc3QiOlsicGVlci0wMDEiLCJwZWVyLTAwMiIsInBlZXItMDAzIl0sInJlbGF5X2VuZHBvaW50cyI6eyJodHRwX2FwaSI6ImVkZ2UuMmdjLnJ1OjgwODAiLCJwMnBfYXBpIjoiZWRnZS4yZ2MucnU6ODA4MiIsInF1aWNfdHJhbnNwb3J0IjoiZWRnZS4yZ2MucnU6OTA5MCIsInN0dW5fc2VydmVyIjoiZWRnZS4yZ2MucnU6MTkzMDIiLCJtYXNxdWVfcHJveHkiOiJlZGdlLjJnYy5ydTo4NDQzIiwiZW5oYW5jZWRfcXVpYyI6ImVkZ2UuMmdjLnJ1OjkwOTIifSwiaWF0IjoxNzU4MjE5MTk4LCJpc3MiOiJodHRwczovL2F1dGguMmdjLnJ1L3JlYWxtcy9jbG91ZGJyaWRnZSIsImF1ZCI6ImFjY291bnQiLCJzdWIiOiJzZXJ2ZXItY2xpZW50LXNlcnZlci0xNzU4MjE5MTk1NzAzIiwianRpIjoiand0XzE3NTgyMTkxOTgxMDlfbXoxdDhyc2RhIn0.d-o1yXtZOlnyv53Uhg5zEI3l3IRbQV27VdnEQIn5w_4
EOF
```

### **2. Тестирование регистрации пиров:**

```bash
# Тест 1: Регистрация первого пира
curl -X POST "http://localhost:8082/api/v1/tenants/tenant-216420165/peers/register" \
  -H "Authorization: Bearer $(cat test-token.txt)" \
  -H "Content-Type: application/json" \
  -d '{
    "public_key": "test-peer-1-key",
    "allowed_ips": ["10.100.77.10/32"],
    "peer_info": {
      "name": "Test Peer 1",
      "location": "Test Location 1"
    }
  }'

# Тест 2: Регистрация второго пира
curl -X POST "http://localhost:8082/api/v1/tenants/tenant-216420165/peers/register" \
  -H "Authorization: Bearer $(cat test-token.txt)" \
  -H "Content-Type: application/json" \
  -d '{
    "public_key": "test-peer-2-key",
    "allowed_ips": ["10.100.77.11/32"],
    "peer_info": {
      "name": "Test Peer 2",
      "location": "Test Location 2"
    }
  }'
```

### **3. Тестирование P2P соединений:**

```bash
# Тест QUIC соединения
echo "Hello from Peer 1" | nc -u localhost 31284

# Тест STUN сервера
python3 -c "
import socket
import struct

def test_stun():
    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    sock.settimeout(3.0)
    
    try:
        transaction_id = b'123456789012'
        magic_cookie = 0x2112A442
        stun_request = struct.pack('>HHI', 0x0001, 0x0000, magic_cookie) + transaction_id
        
        sock.sendto(stun_request, ('localhost', 30258))
        response, addr = sock.recvfrom(1024)
        
        msg_type = struct.unpack('>H', response[0:2])[0]
        if msg_type == 0x0101:
            print('✅ STUN Server - WORKING')
        else:
            print(f'⚠️  STUN Server - Unexpected response: 0x{msg_type:04x}')
        
    except Exception as e:
        print(f'❌ STUN Server - Error: {e}')
    finally:
        sock.close()

test_stun()
"
```

## 🧹 **Очистка и остановка**

### **1. Остановка Minikube:**

```bash
# Остановить Minikube
minikube stop

# Удалить Minikube (если нужно)
minikube delete
```

### **2. Очистка ресурсов:**

```bash
# Удалить namespace
kubectl delete namespace cloudbridge

# Очистить Docker образы
docker rmi cloudbridge-relay:quic-ice-v2
docker rmi cloudbridge-relay:quic-ice-v3

# Очистить Minikube образы
minikube image rm cloudbridge-relay:quic-ice-v2
minikube image rm cloudbridge-relay:quic-ice-v3
```

## 📋 **Чек-лист готовности**

### **✅ Проверки перед запуском:**

- [ ] Minikube установлен и запущен
- [ ] kubectl настроен и работает
- [ ] Docker образ собран успешно
- [ ] Образ загружен в Minikube
- [ ] Kubernetes конфигурации применены
- [ ] Pods запущены и работают
- [ ] Services созданы и доступны
- [ ] Port forwarding настроен
- [ ] API endpoints отвечают
- [ ] JWT токены работают
- [ ] P2P регистрация работает
- [ ] STUN сервер отвечает
- [ ] QUIC порты доступны

### **✅ Проверки после запуска:**

- [ ] Логи не содержат ошибок
- [ ] Метрики показывают нормальные значения
- [ ] Память и CPU в пределах нормы
- [ ] Сетевые соединения стабильны
- [ ] Heartbeat система работает
- [ ] P2P маршрутизация функционирует
- [ ] Мультитунельность изолирована
- [ ] Graceful shutdown работает

---

**🎯 Результат: Полноценная среда разработки CloudBridge Relay в Minikube с поддержкой всех протоколов и мультитунельности!**
