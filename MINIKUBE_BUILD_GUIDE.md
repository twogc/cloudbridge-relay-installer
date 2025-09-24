# 🛡️ **CloudBridge Relay - Безопасная сборка в Minikube**

## ⚠️ **ВАЖНО: Проблемы при обычной сборке**

При сборке Docker образов в Minikube **падают другие поды** (Zitadel → 503 ошибки) из-за перегрузки ресурсов.

## ✅ **Решение: Безопасная сборка**

### **1. Запуск Minikube с достаточными ресурсами:**
```bash
# Остановить если запущен
minikube stop

# Запустить с увеличенными ресурсами
minikube start --driver=docker --memory=8192 --cpus=4 --disk-size=50g

# Проверить статус
minikube status
```

### **2. Проверка нагрузки перед сборкой:**
```bash
# Проверить текущую нагрузку (должна быть < 2.0)
uptime

# Если нагрузка высокая - ждать снижения
watch uptime
```

## 🏗️ **Безопасная сборка**

### **Метод 1: Автоматический скрипт (рекомендуется)**
```bash
# Использовать готовый безопасный скрипт
/home/ubuntu/safe-build.sh

# Что делает скрипт:
# - Проверяет нагрузку системы
# - Ждет снижения нагрузки если нужно
# - Собирает с ограничением ресурсов (2GB RAM, 2 CPU)
# - Создает уникальный тег автоматически
# - Обновляет deployment
```

### **Метод 2: Ручная безопасная сборка**
```bash
# 1. Проверить нагрузку
uptime  # Load average должен быть < 2.0

# 2. Настроить Docker для Minikube
eval $(minikube docker-env)

# 3. Создать уникальный тег
TAG="build-$(date +%Y%m%d-%H%M%S)"
echo "Тег: $TAG"

# 4. Собрать с ограничениями ресурсов
cd /home/ubuntu/cloudbridge-relay-installer
docker build --memory=2g --cpus=2 -t cloudbridge-relay:$TAG .

# 5. Обновить deployment
kubectl set image deployment/cloudbridge-relay cloudbridge-relay=cloudbridge-relay:$TAG -n cloudbridge
```

## 🔍 **Мониторинг во время сборки**

```bash
# Мониторить нагрузку в реальном времени
watch -n 5 'uptime && kubectl get pods -n zitadel | grep -v Running || echo "Все поды работают"'

# Проверить статус после сборки
kubectl get pods --all-namespaces | grep -v Running
```

## 🚀 **Тестирование**

```bash
# Проверить что поды запущены
kubectl get pods -n cloudbridge

# Port forwarding для тестов
kubectl port-forward service/cloudbridge-relay-external 8080:8080 -n cloudbridge &

# Тест API
curl http://localhost:8080/health
```

## ⚠️ **Признаки проблем**

```bash
# Проверить нагрузку (должна быть < 5.0)
uptime

# Проверить падающие поды
kubectl get pods --all-namespaces | grep -E "(Error|CrashLoop|ImagePull)"

# Если Zitadel падает - перезапустить сборку позже
kubectl get pods -n zitadel
```

## **Быстрый чек-лист безопасной сборки:**
- [ ] Minikube запущен с 8GB+ RAM
- [ ] Load average < 2.0 перед сборкой  
- [ ] Используется `/home/ubuntu/safe-build.sh`
- [ ] Мониторинг во время сборки
- [ ] Все поды остались работать

---

**🛡️ Результат: Безопасная сборка без падения других подов!**
