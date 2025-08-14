# 🎯 Финальные шаги миграции CloudBridge

## ✅ Что сделано:
- ✅ Тестовые сервисы удалены
- ✅ Конфигурация Docker Registry готова
- ✅ Образы собраны и готовы к загрузке
- ✅ Скрипт загрузки готов

## 🚀 Следующие шаги (5 минут):

### Шаг 1: Установить Docker Registry
1. **VK Cloud Console** → **Kubernetes** → **kubernetes-cluster-8926** → **Аддоны**
2. **Найти** "docker-registry" → **Установить**
3. **Название приложения**: `cloudbridge-registry`
4. **Пространство имен**: `kube-system`
5. **Values**: Скопировать содержимое `configs/docker-registry-vkcloud.yaml`
6. **Установить аддон**

### Шаг 2: Проверить установку registry
```bash
export KUBECONFIG=~/.kube/kubernetes-cluster-8926_kubeconfig.yaml

# Проверить поды registry
kubectl --insecure-skip-tls-verify get pods -n kube-system | grep registry

# Проверить сервис registry
kubectl --insecure-skip-tls-verify get services -n kube-system | grep registry
```

### Шаг 3: Загрузить образы в registry
```bash
# Автоматическая загрузка образов
./scripts/upload-images-to-registry.sh 90.156.150.25
```

### Шаг 4: Развернуть реальные микросервисы
```bash
# Развернуть Auth Service
kubectl --insecure-skip-tls-verify apply -f k8s/auth-service-real-deployment.yaml

# Развернуть Metrics Service
kubectl --insecure-skip-tls-verify apply -f k8s/metrics-service-real-deployment.yaml

# Проверить статус
kubectl --insecure-skip-tls-verify get pods -n cloudbridge
```

### Шаг 5: Проверить работу сервисов
```bash
# Проверить поды
kubectl --insecure-skip-tls-verify get pods -n cloudbridge

# Проверить сервисы
kubectl --insecure-skip-tls-verify get services -n cloudbridge

# Проверить логи Auth Service
kubectl --insecure-skip-tls-verify logs -n cloudbridge -l app=auth-service-real

# Проверить логи Metrics Service
kubectl --insecure-skip-tls-verify logs -n cloudbridge -l app=metrics-service-real
```

## 🔐 Учетные данные
- **Registry**: `90.156.150.25:5000`
- **Логин**: `cloudbridge`
- **Пароль**: `password`

## 📊 Ожидаемый результат
После выполнения всех шагов:
- ✅ Docker Registry работает на `90.156.150.25:5000`
- ✅ CloudBridge образы загружены в registry
- ✅ Auth Service развернут и работает
- ✅ Metrics Service развернут и работает
- ✅ Микросервисы готовы к использованию

## 🚨 Если что-то не работает

### Registry недоступен:
```bash
# Проверить статус registry
kubectl --insecure-skip-tls-verify get pods -n kube-system | grep registry
kubectl --insecure-skip-tls-verify logs -n kube-system <registry-pod-name>
```

### Образы не загружаются:
```bash
# Проверить доступность registry
curl -u cloudbridge:password http://90.156.150.25:5000/v2/_catalog
```

### Поды не запускаются:
```bash
# Проверить логи подов
kubectl --insecure-skip-tls-verify describe pods -n cloudbridge
kubectl --insecure-skip-tls-verify logs -n cloudbridge <pod-name>
```

---

**⏱️ Время до завершения**: 5 минут  
**🎯 Статус**: ГОТОВ К ФИНАЛЬНОЙ МИГРАЦИИ
