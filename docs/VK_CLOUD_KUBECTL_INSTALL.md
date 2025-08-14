# 🚀 Установка Grafana через kubectl apply в VK Cloud

## 📋 Альтернативный способ установки

Поскольку аддоны не работают, попробуем установить Grafana напрямую через kubectl apply в консоли VK Cloud.

## 🎯 Манифест для копирования

Скопируйте **ВСЕ** содержимое ниже в консоль kubectl в VK Cloud:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: grafana-monitoring
  labels:
    name: grafana-monitoring
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: grafana
  namespace: grafana-monitoring
  labels:
    app: grafana
spec:
  replicas: 1
  selector:
    matchLabels:
      app: grafana
  template:
    metadata:
      labels:
        app: grafana
    spec:
      containers:
      - name: grafana
        image: grafana/grafana:10.1.5
        ports:
        - containerPort: 3000
        env:
        - name: GF_SECURITY_ADMIN_PASSWORD
          value: "admin123"
        - name: GF_SECURITY_ADMIN_USER
          value: "admin"
        resources:
          requests:
            cpu: 10m
            memory: 16Mi
          limits:
            cpu: 25m
            memory: 32Mi
---
apiVersion: v1
kind: Service
metadata:
  name: grafana-service
  namespace: grafana-monitoring
  labels:
    app: grafana
spec:
  type: ClusterIP
  ports:
  - port: 80
    targetPort: 3000
    protocol: TCP
  selector:
    app: grafana
```

## 📝 Шаги установки

### 1. Откройте консоль VK Cloud
- https://mcs.mail.ru/
- Войдите в аккаунт `info@2gc.ru`
- Выберите проект с кластером `kubernetes-cluster-6379`

### 2. Перейдите к кластеру
- Меню слева → **Kubernetes**
- Найдите `kubernetes-cluster-6379`
- Нажмите на него

### 3. Откройте kubectl консоль
- Вкладка **Ресурсы кластера**
- Найдите **kubectl консоль** или **Web Terminal**
- Нажмите на него

### 4. Вставьте и выполните команду
1. Вставьте манифест выше в консоль
2. Выполните команду:
```bash
kubectl apply -f -
```

### 5. Проверьте установку
```bash
kubectl get pods -n grafana-monitoring
kubectl get services -n grafana-monitoring
```

## 🔍 Проверка установки

### В kubectl консоли:
```bash
# Проверить поды
kubectl get pods -n grafana-monitoring

# Проверить сервисы
kubectl get services -n grafana-monitoring

# Проверить логи
kubectl logs -n grafana-monitoring deployment/grafana
```

### Статус пода должен быть **Running**

## 🌐 Доступ к Grafana

### Способ 1: Port-forward в kubectl консоли
```bash
kubectl -n grafana-monitoring port-forward service/grafana-service 3000:80
```

### Способ 2: Через LoadBalancer
Создайте LoadBalancer сервис:
```bash
kubectl patch service grafana-service -n grafana-monitoring -p '{"spec":{"type":"LoadBalancer"}}'
```

### Способ 3: Через NodePort
Создайте NodePort сервис:
```bash
kubectl patch service grafana-service -n grafana-monitoring -p '{"spec":{"type":"NodePort"}}'
```

## 🔑 Логин в Grafana
- **URL**: http://[IP]:3000 (при port-forward)
- **Логин**: `admin`
- **Пароль**: `admin123`

## ⚠️ Преимущества этого способа

✅ **Прямая установка** - минуем проблемы с аддонами  
✅ **Простой манифест** - только необходимые компоненты  
✅ **Минимальные ресурсы** - 10m CPU, 16Mi RAM  
✅ **Быстрая установка** - несколько секунд  

## 🚨 Если не работает

### Проблема: "ImagePullBackOff"
```bash
# Проверьте образ
kubectl describe pod -n grafana-monitoring -l app=grafana

# Попробуйте другой образ
kubectl set image deployment/grafana grafana=grafana/grafana:9.5.0 -n grafana-monitoring
```

### Проблема: "Pending"
```bash
# Проверьте ресурсы узла
kubectl describe node

# Уменьшите ресурсы
kubectl patch deployment grafana -n grafana-monitoring -p '{"spec":{"template":{"spec":{"containers":[{"name":"grafana","resources":{"requests":{"cpu":"5m","memory":"8Mi"}}}]}}}}'
```

## 📊 Следующие шаги

После успешной установки:
1. **Добавьте Prometheus** как источник данных
2. **Настройте дашборды** для CloudBridge
3. **Добавьте алерты** для критических событий

---

**Статус**: ✅ Альтернативный способ установки  
**Кластер**: kubernetes-cluster-6379  
**Время установки**: 1-2 минуты  
**Сложность**: Низкая
