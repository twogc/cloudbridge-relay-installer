# 🎯 Финальная установка Docker Registry

## ✅ Все готово!
- **Anycast IP**: `212.233.88.74`
- **Плавающий IP**: `90.156.150.25`
- **Конфигурация**: Обновлена и готова

## 🚀 Установка Docker Registry (2 минуты)

### Шаг 1: Установить через VK Cloud Console
1. **VK Cloud Console** → **Kubernetes** → **kubernetes-cluster-8926** → **Аддоны**
2. **Найти** "docker-registry" → **Установить**
3. **Название приложения**: `cloudbridge-registry`
4. **Пространство имен**: `kube-system`
5. **Values**: Скопировать содержимое `configs/docker-registry-vkcloud.yaml`
6. **Установить аддон**

### Шаг 2: Проверить установку
```bash
export KUBECONFIG=~/.kube/kubernetes-cluster-8926_kubeconfig.yaml

# Проверить поды registry
kubectl --insecure-skip-tls-verify get pods -n kube-system | grep registry

# Проверить сервис registry
kubectl --insecure-skip-tls-verify get services -n kube-system | grep registry
```

### Шаг 3: Загрузить образы
```bash
# Автоматическая загрузка образов
./scripts/upload-images-to-registry.sh 90.156.150.25
```

### Шаг 4: Развернуть микросервисы
```bash
# Развернуть реальные сервисы
kubectl --insecure-skip-tls-verify apply -f k8s/auth-service-real-deployment.yaml
kubectl --insecure-skip-tls-verify apply -f k8s/metrics-service-real-deployment.yaml

# Проверить статус
kubectl --insecure-skip-tls-verify get pods -n cloudbridge
```

## 🔐 Учетные данные
- **Registry**: `90.156.150.25:5000`
- **Логин**: `cloudbridge`
- **Пароль**: `password`

## 📊 Проверка работы
```bash
# Проверить registry
curl -u cloudbridge:password http://90.156.150.25:5000/v2/_catalog

# Проверить поды
kubectl --insecure-skip-tls-verify get pods -n cloudbridge

# Проверить сервисы
kubectl --insecure-skip-tls-verify get services -n cloudbridge
```

## 🎉 Результат
После выполнения всех шагов у вас будет:
- ✅ Приватный Docker Registry на `90.156.150.25:5000`
- ✅ CloudBridge образы загружены в registry
- ✅ Auth Service развернут в Kubernetes
- ✅ Metrics Service развернут в Kubernetes
- ✅ Микросервисы готовы к работе

---

**⏱️ Время установки**: 2 минуты  
**🎯 Статус**: ГОТОВ К УСТАНОВКЕ
