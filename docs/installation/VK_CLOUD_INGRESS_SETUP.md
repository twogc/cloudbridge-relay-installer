# 🚀 Установка ingress-nginx в VK Cloud для CloudBridge Relay

## 📋 Подготовка к установке

### 1. Проверка доступности floating IP
```bash
# Проверить доступные floating IP в VK Cloud
# Если нет доступных IP, запросить увеличение квоты
```

### 2. Подготовка DNS записей
```bash
# Добавить в /etc/hosts или настроить DNS
# Заменить EXTERNAL_IP на реальный IP адрес кластера
EXTERNAL_IP auth.cloudbridge.local
EXTERNAL_IP metrics.cloudbridge.local
EXTERNAL_IP api.cloudbridge.local
EXTERNAL_IP grafana.cloudbridge.local
EXTERNAL_IP prometheus.cloudbridge.local
EXTERNAL_IP alertmanager.cloudbridge.local
```

## 🛠️ Установка в VK Cloud Console

### Шаг 1: Переход к кластеру
1. Открыть VK Cloud Console
2. Перейти к разделу "Kubernetes"
3. Выбрать кластер `kubernetes-cluster-8926`
4. Перейти на вкладку "Аддоны"

### Шаг 2: Установка ingress-nginx
1. Найти аддон "ingress-nginx"
2. Нажать "Установить"
3. В поле "Values" вставить содержимое файла `configs/ingress-nginx-values.yaml`
4. **ВАЖНО**: В поле `loadBalancerIP` указать доступный floating IP
5. Нажать "Установить"

## 🔧 Ключевые настройки

### ✅ Основные параметры:
- **replicaCount**: `2.0` (высокая доступность)
- **autoscaling**: `enabled: true` (автоматическое масштабирование)
- **loadBalancerIP**: Указать доступный floating IP

### ✅ Ресурсы:
- **Controller**: 200m-1000m CPU / 180Mi-1Gi Memory
- **Default Backend**: 20m CPU / 40Mi Memory

### ✅ Автомасштабирование:
- **minReplicas**: 2
- **maxReplicas**: 5
- **targetCPUUtilizationPercentage**: 70
- **targetMemoryUtilizationPercentage**: 80

### ✅ SSL настройки:
- **ssl-protocols**: TLSv1.2 TLSv1.3
- **ssl-ciphers**: ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384
- **ssl-redirect**: true
- **force-ssl-redirect**: true

## 📊 Проверка установки

### 1. Проверка подов
```bash
# Проверить статус подов
kubectl --insecure-skip-tls-verify get pods -n ingress-nginx

# Ожидаемый результат:
# NAME                                        READY   STATUS    RESTARTS   AGE
# ingress-nginx-controller-xxx                1/1     Running   0          2m
# ingress-nginx-controller-xxx                1/1     Running   0          2m
# ingress-nginx-defaultbackend-xxx            1/1     Running   0          2m
# ingress-nginx-defaultbackend-xxx            1/1     Running   0          2m
```

### 2. Проверка сервисов
```bash
# Проверить сервисы
kubectl --insecure-skip-tls-verify get svc -n ingress-nginx

# Ожидаемый результат:
# NAME                                 TYPE           CLUSTER-IP       EXTERNAL-IP   PORT(S)
# ingress-nginx-controller             LoadBalancer   10.254.xxx.xxx   EXTERNAL_IP   80:xxxxx/TCP,443:xxxxx/TCP
# ingress-nginx-controller-admission   ClusterIP      10.254.xxx.xxx   <none>        443/TCP
# ingress-nginx-defaultbackend         ClusterIP      10.254.xxx.xxx   <none>        80/TCP
```

### 3. Проверка LoadBalancer
```bash
# Проверить внешний IP
kubectl --insecure-skip-tls-verify get svc ingress-nginx-controller -n ingress-nginx

# Должен быть EXTERNAL_IP в статусе
```

## 🌐 Тестирование доступа

### 1. Базовый тест
```bash
# Тест HTTP
curl -I http://EXTERNAL_IP

# Ожидаемый результат: 404 (default backend)
# HTTP/1.1 404 Not Found
```

### 2. Тест HTTPS
```bash
# Тест HTTPS (будет редирект на HTTP пока нет сертификатов)
curl -I -k https://EXTERNAL_IP

# Ожидаемый результат: 404 (default backend)
```

### 3. Тест с DNS
```bash
# После настройки DNS
curl -I http://auth.cloudbridge.local
curl -I http://metrics.cloudbridge.local
curl -I http://api.cloudbridge.local
```

## 🔧 Интеграция с CloudBridge микросервисами

### 1. Обновление Ingress правил
```bash
# Применить обновленные Ingress правила
kubectl --insecure-skip-tls-verify apply -f k8s/cloudbridge-ingress-ssl.yaml
```

### 2. Проверка Ingress
```bash
# Проверить Ingress правила
kubectl --insecure-skip-tls-verify get ingress -n cloudbridge

# Ожидаемый результат:
# NAME                    CLASS   HOSTS                           ADDRESS   PORTS
# cloudbridge-ingress     nginx   auth.cloudbridge.local,metrics.cloudbridge.local   80, 443
```

### 3. Тестирование маршрутизации
```bash
# Тест Auth Service
curl -H "Host: auth.cloudbridge.local" http://EXTERNAL_IP/health

# Тест Metrics Service
curl -H "Host: metrics.cloudbridge.local" http://EXTERNAL_IP/health

# Тест API Gateway
curl -H "Host: api.cloudbridge.local" http://EXTERNAL_IP/auth/health
```

## 📈 Мониторинг ingress-nginx

### 1. Метрики Prometheus
```bash
# Проверить метрики ingress-nginx
curl http://EXTERNAL_IP:9913/metrics

# Основные метрики:
# - nginx_ingress_controller_nginx_process_connections
# - nginx_ingress_controller_nginx_process_requests
# - nginx_ingress_controller_nginx_process_resident_memory_bytes
```

### 2. Статистика
```bash
# Проверить статистику
curl http://EXTERNAL_IP:18080/nginx_status

# Ожидаемый результат:
# Active connections: 1
# server accepts handled requests
# 10 10 20
# Reading: 0 Writing: 1 Waiting: 0
```

## 🔔 Настройка алертов

### 1. Создание правил алертов для ingress-nginx
```yaml
# k8s/monitoring/ingress-alerts.yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: ingress-nginx-alerts
  namespace: cloudbridge-monitoring
spec:
  groups:
  - name: ingress-nginx
    rules:
    - alert: IngressNginxDown
      expr: up{job="ingress-nginx-controller"} == 0
      for: 1m
      labels:
        severity: critical
      annotations:
        summary: "Ingress Nginx Controller is down"
        description: "Ingress Nginx Controller has been down for more than 1 minute"
    
    - alert: IngressNginxHighErrorRate
      expr: rate(nginx_ingress_controller_nginx_process_requests_total{status=~"5.."}[5m]) > 0.1
      for: 2m
      labels:
        severity: warning
      annotations:
        summary: "High error rate in Ingress Nginx"
        description: "Error rate is {{ $value }} errors per second"
```

### 2. Применение алертов
```bash
# Применить правила алертов
kubectl --insecure-skip-tls-verify apply -f k8s/monitoring/ingress-alerts.yaml
```

## 🚨 Устранение неполадок

### 1. LoadBalancer не получает внешний IP
```bash
# Проверить события
kubectl --insecure-skip-tls-verify get events -n ingress-nginx

# Проверить квоту floating IP
# Обратиться в поддержку VK Cloud для увеличения квоты
```

### 2. Ingress не маршрутизирует трафик
```bash
# Проверить логи ingress-nginx
kubectl --insecure-skip-tls-verify logs -f deployment/ingress-nginx-controller -n ingress-nginx

# Проверить конфигурацию nginx
kubectl --insecure-skip-tls-verify exec -it ingress-nginx-controller-xxx -n ingress-nginx -- nginx -t
```

### 3. SSL сертификаты не работают
```bash
# Проверить сертификаты
kubectl --insecure-skip-tls-verify get certificates -n cloudbridge

# Проверить cert-manager
kubectl --insecure-skip-tls-verify get pods -n cert-manager
```

## 📊 Полезные команды

### Мониторинг производительности
```bash
# Использование ресурсов
kubectl --insecure-skip-tls-verify top pods -n ingress-nginx

# Логи в реальном времени
kubectl --insecure-skip-tls-verify logs -f deployment/ingress-nginx-controller -n ingress-nginx
```

### Конфигурация
```bash
# Проверить конфигурацию nginx
kubectl --insecure-skip-tls-verify exec -it ingress-nginx-controller-xxx -n ingress-nginx -- cat /etc/nginx/nginx.conf

# Перезагрузить конфигурацию
kubectl --insecure-skip-tls-verify exec -it ingress-nginx-controller-xxx -n ingress-nginx -- nginx -s reload
```

## 🎯 Следующие шаги

### 1. Настройка SSL сертификатов
- Установить cert-manager
- Настроить ClusterIssuer для Let's Encrypt
- Создать сертификаты для доменов

### 2. Оптимизация производительности
- Настроить кэширование
- Оптимизировать rate limiting
- Настроить gzip сжатие

### 3. Безопасность
- Настроить network policies
- Настроить WAF правила
- Настроить IP whitelist

---

**🎉 Готово!** Ingress Nginx установлен и настроен для CloudBridge Relay!

