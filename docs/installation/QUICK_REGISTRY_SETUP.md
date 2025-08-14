# ⚡ Быстрая установка Docker Registry в VK Cloud

## 🎯 Цель
Установить Docker Registry и загрузить CloudBridge образы за 10 минут.

## 📋 Быстрые шаги

### 1. Создать плавающий IP
1. **VK Cloud Console** → **Сети** → **Плавающие IP**
2. **Создать плавающий IP**
3. **Скопировать IP адрес** (например: `95.163.250.128`)

### 2. Обновить конфигурацию
Откройте `configs/docker-registry-vkcloud.yaml` и замените:
```yaml
service:
  loadBalancerIP: "95.163.250.128"  # ← Вставьте ваш IP
```

### 3. Установить Docker Registry
1. **VK Cloud Console** → **Kubernetes** → **kubernetes-cluster-8926** → **Аддоны**
2. **Найти** "docker-registry" → **Установить**
3. **Название**: `cloudbridge-registry`
4. **Namespace**: `kube-system`
5. **Values**: Вставить содержимое `configs/docker-registry-vkcloud.yaml`
6. **Установить**

### 4. Загрузить образы
```bash
# Заменить <плавающий_IP> на ваш IP
./scripts/upload-images-to-registry.sh <плавающий_IP>
```

### 5. Развернуть микросервисы
```bash
export KUBECONFIG=~/.kube/kubernetes-cluster-8926_kubeconfig.yaml

# Развернуть реальные сервисы
kubectl --insecure-skip-tls-verify apply -f k8s/auth-service-real-deployment.yaml
kubectl --insecure-skip-tls-verify apply -f k8s/metrics-service-real-deployment.yaml

# Проверить статус
kubectl --insecure-skip-tls-verify get pods -n cloudbridge
```

## 🔐 Учетные данные
- **Registry**: `<плавающий_IP>:5000`
- **Логин**: `cloudbridge`
- **Пароль**: `password`

## 📊 Проверка
```bash
# Проверить registry
curl -u cloudbridge:password http://<плавающий_IP>:5000/v2/_catalog

# Проверить поды
kubectl --insecure-skip-tls-verify get pods -n cloudbridge
```

## 🚨 Если не работает
1. **Проверить плавающий IP** - убедитесь, что IP создан и указан правильно
2. **Проверить registry** - `kubectl get pods -n kube-system | grep registry`
3. **Проверить логи** - `kubectl logs -n kube-system <registry-pod-name>`

---

**⏱️ Время установки**: 10 минут  
**🎯 Результат**: CloudBridge образы в приватном registry
