# 🚀 Установка kube-prometheus-stack в VK Cloud для CloudBridge Relay

## 📋 Подготовка к установке

### 1. Создание namespace для мониторинга
```bash
# Создать namespace для мониторинга
kubectl --insecure-skip-tls-verify create namespace cloudbridge-monitoring
```

### 2. Подготовка DNS записей
```bash
# Добавить в /etc/hosts или настроить DNS
# Заменить EXTERNAL_IP на реальный IP адрес кластера
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

### Шаг 2: Установка kube-prometheus-stack
1. Найти аддон "kube-prometheus-stack"
2. Нажать "Установить"
3. В поле "Values" вставить содержимое файла `configs/kube-prometheus-stack-values.yaml`
4. Нажать "Установить"

## 🔧 Ключевые настройки

### ✅ Основные параметры:
- **nameOverride**: `cloudbridge-prometheus`
- **fullnameOverride**: `cloudbridge-prometheus`
- **adminPassword**: `CloudBridge2025!`

### ✅ Ресурсы:
- **Grafana**: 500m CPU / 1Gi Memory
- **Prometheus**: 1000m CPU / 4Gi Memory
- **Alertmanager**: 500m CPU / 512Mi Memory

### ✅ Хранилище:
- **Grafana**: 5Gi (csi-ceph-hdd-me1)
- **Prometheus**: 20Gi (csi-ceph-ssd-me1)
- **Alertmanager**: 2Gi (csi-ceph-hdd-me1)

### ✅ Ingress:
- **Grafana**: `https://grafana.cloudbridge.local/grafana/`
- **Prometheus**: `https://prometheus.cloudbridge.local/prometheus/`
- **Alertmanager**: `https://alertmanager.cloudbridge.local/alertmanager/`

## 📊 Проверка установки

### 1. Проверка подов
```bash
# Проверить статус подов
kubectl --insecure-skip-tls-verify get pods -n cloudbridge-monitoring

# Ожидаемый результат:
# NAME                                                     READY   STATUS    RESTARTS   AGE
# cloudbridge-prometheus-grafana-xxx                      1/1     Running   0          2m
# cloudbridge-prometheus-kube-state-metrics-xxx           1/1     Running   0          2m
# cloudbridge-prometheus-prometheus-node-exporter-xxx     1/1     Running   0          2m
# cloudbridge-prometheus-prometheus-xxx                   1/1     Running   0          2m
# cloudbridge-prometheus-alertmanager-xxx                 1/1     Running   0          2m
# cloudbridge-prometheus-operator-xxx                     1/1     Running   0          2m
```

### 2. Проверка сервисов
```bash
# Проверить сервисы
kubectl --insecure-skip-tls-verify get svc -n cloudbridge-monitoring

# Ожидаемый результат:
# NAME                                    TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)
# cloudbridge-prometheus-grafana          ClusterIP   10.254.xxx.xxx   <none>        80/TCP
# cloudbridge-prometheus-prometheus       ClusterIP   10.254.xxx.xxx   <none>        9090/TCP
# cloudbridge-prometheus-alertmanager     ClusterIP   10.254.xxx.xxx   <none>        9093/TCP
```

### 3. Проверка Ingress
```bash
# Проверить Ingress
kubectl --insecure-skip-tls-verify get ingress -n cloudbridge-monitoring

# Ожидаемый результат:
# NAME                                    CLASS   HOSTS                           ADDRESS   PORTS
# cloudbridge-prometheus-grafana         nginx   grafana.cloudbridge.local                80, 443
# cloudbridge-prometheus-prometheus      nginx   prometheus.cloudbridge.local            80, 443
# cloudbridge-prometheus-alertmanager    nginx   alertmanager.cloudbridge.local          80, 443
```

## 🌐 Доступ к сервисам

### 1. Grafana
```bash
# URL: https://grafana.cloudbridge.local/grafana/
# Логин: admin
# Пароль: CloudBridge2025!

# Проверка доступа
curl -k https://grafana.cloudbridge.local/grafana/api/health
```

### 2. Prometheus
```bash
# URL: https://prometheus.cloudbridge.local/prometheus/
# Проверка доступа
curl -k https://prometheus.cloudbridge.local/prometheus/api/v1/status/config
```

### 3. Alertmanager
```bash
# URL: https://alertmanager.cloudbridge.local/alertmanager/
# Проверка доступа
curl -k https://alertmanager.cloudbridge.local/alertmanager/api/v1/status
```

## 📈 Настройка дашбордов

### 1. Импорт дашбордов в Grafana
```bash
# Создать ConfigMap с дашбордами
kubectl --insecure-skip-tls-verify apply -f k8s/monitoring/grafana-dashboards.yaml
```

### 2. Основные дашборды:
- **CloudBridge Overview**: Общий обзор всех сервисов
- **Auth Service**: Метрики аутентификации
- **Metrics Service**: Метрики и производительность
- **Kubernetes Cluster**: Состояние кластера

## 🔔 Настройка алертов

### 1. Создание правил алертов
```bash
# Применить правила алертов
kubectl --insecure-skip-tls-verify apply -f k8s/monitoring/prometheus-rules.yaml
```

### 2. Основные алерты:
- **Критические**: Сервисы недоступны
- **Предупреждения**: Высокая нагрузка
- **Информационные**: Обновления и события

## 🔧 Интеграция с CloudBridge микросервисами

### 1. Добавление ServiceMonitors
```bash
# Применить ServiceMonitors для микросервисов
kubectl --insecure-skip-tls-verify apply -f k8s/monitoring/service-monitors.yaml
```

### 2. Проверка мониторинга
```bash
# Проверить targets в Prometheus
curl -k https://prometheus.cloudbridge.local/prometheus/api/v1/targets

# Проверить метрики
curl -k https://prometheus.cloudbridge.local/prometheus/api/v1/query?query=up
```

## 🚨 Устранение неполадок

### 1. Поды не запускаются
```bash
# Проверить логи
kubectl --insecure-skip-tls-verify logs -f deployment/cloudbridge-prometheus-operator -n cloudbridge-monitoring

# Проверить события
kubectl --insecure-skip-tls-verify get events -n cloudbridge-monitoring
```

### 2. Ingress не работает
```bash
# Проверить Ingress Controller
kubectl --insecure-skip-tls-verify get pods -n ingress-nginx

# Проверить сертификаты
kubectl --insecure-skip-tls-verify get certificates -n cloudbridge-monitoring
```

### 3. Метрики не собираются
```bash
# Проверить ServiceMonitors
kubectl --insecure-skip-tls-verify get servicemonitors -n cloudbridge-monitoring

# Проверить endpoints
kubectl --insecure-skip-tls-verify get endpoints -n cloudbridge
```

## 📊 Полезные команды

### Мониторинг ресурсов
```bash
# Использование CPU и Memory
kubectl --insecure-skip-tls-verify top pods -n cloudbridge-monitoring

# Использование диска
kubectl --insecure-skip-tls-verify exec -it cloudbridge-prometheus-prometheus-0 -n cloudbridge-monitoring -- df -h
```

### Резервное копирование
```bash
# Экспорт конфигурации
kubectl --insecure-skip-tls-verify get configmap -n cloudbridge-monitoring -o yaml > prometheus-backup.yaml

# Экспорт правил алертов
kubectl --insecure-skip-tls-verify get prometheusrules -n cloudbridge-monitoring -o yaml > alerts-backup.yaml
```

## 🎯 Следующие шаги

### 1. Настройка уведомлений
- Настроить email уведомления в Alertmanager
- Настроить Slack/Telegram интеграцию
- Настроить PagerDuty для критических алертов

### 2. Оптимизация
- Настроить retention policy для метрик
- Оптимизировать запросы Prometheus
- Настроить горизонтальное масштабирование

### 3. Безопасность
- Настроить RBAC для доступа к мониторингу
- Настроить network policies
- Настроить audit logging

---

**🎉 Готово!** Система мониторинга CloudBridge Relay установлена и настроена в VK Cloud!

