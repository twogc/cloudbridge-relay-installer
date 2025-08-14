# 🚀 Установка Docker Registry - ПРЯМО СЕЙЧАС

## ✅ Anycast IP готов
- **Anycast IP**: `212.233.88.74`
- **Статус**: Готов к использованию

## 📋 Следующие шаги (5 минут)

### 1. Создать плавающий IP
1. **VK Cloud Console** → **Сети** → **Плавающие IP**
2. **Создать плавающий IP**
3. **Скопировать полученный IP** (например: `95.163.250.129`)

### 2. Обновить конфигурацию
Откройте `configs/docker-registry-vkcloud-template.yaml` и замените:
```yaml
loadBalancerIP: "<ВАШ_ПЛАВАЮЩИЙ_IP>"  # ← Вставьте ваш IP
```

### 3. Установить Docker Registry
1. **VK Cloud Console** → **Kubernetes** → **kubernetes-cluster-8926** → **Аддоны**
2. **Найти** "docker-registry" → **Установить**
3. **Название**: `cloudbridge-registry`
4. **Namespace**: `kube-system`
5. **Values**: Вставить обновленную конфигурацию
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

---

**⏱️ Время до завершения**: 5 минут  
**🎯 Результат**: CloudBridge микросервисы в продакшене
