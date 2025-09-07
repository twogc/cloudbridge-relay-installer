# Monitoring Documentation

## 📁 Файлы в этой папке

- **[MONITORING_GUIDE.md](./MONITORING_GUIDE.md)** - Руководство по мониторингу
- **[PROMETHEUS_OPTIMIZATION_GUIDE.md](./PROMETHEUS_OPTIMIZATION_GUIDE.md)** - Руководство по оптимизации Prometheus
- **[UI_MONITORING_GUIDE.md](./UI_MONITORING_GUIDE.md)** - Руководство по мониторингу через UI

## 📊 Система мониторинга

### Компоненты мониторинга
- **Prometheus Optimized** - Сбор метрик с умной оптимизацией (94% экономия места)
- **Grafana** - Визуализация данных
- **Alertmanager** - Управление алертами
- **Blackbox Exporter** - Мониторинг endpoints

### 🚀 **Новые возможности (v2.0):**
- **Умная оптимизация пространства** - 94% экономия дискового пространства
- **Автоматическая очистка** - каждые 6 часов
- **Оптимизированные интервалы** - 4-20x реже scraping
- **Фильтрация метрик** - удаление избыточных данных
- **UI мониторинг** - удобные интерфейсы для Cadence и Minikube
- **Автоматический запуск** - скрипт для управления всеми UI

### Метрики
- **Application Metrics** - Метрики приложения
- **System Metrics** - Системные метрики
- **Network Metrics** - Сетевые метрики
- **Security Metrics** - Метрики безопасности

### Дашборды
- **CloudBridge Overview** - Общий обзор системы
- **Performance Dashboard** - Производительность
- **Security Dashboard** - Безопасность
- **Infrastructure Dashboard** - Инфраструктура

## 🚀 Развертывание оптимизированного мониторинга

### 1. Запуск оптимизированного мониторинга
```bash
# Развертывание оптимизированного Prometheus
kubectl apply -f k8s/storageclasses/prometheus-optimized-storage.yaml
kubectl apply -f k8s/persistentvolumes/prometheus-optimized-pv.yaml
kubectl apply -f k8s/persistentvolumeclaims/prometheus-optimized-pvc.yaml
kubectl apply -f k8s/configmaps/prometheus-optimized-config.yaml
kubectl apply -f k8s/deployments/prometheus-optimized-deployment.yaml
kubectl apply -f k8s/services/prometheus-optimized-service.yaml
kubectl apply -f k8s/cronjobs/prometheus-cleanup-cronjob.yaml

# Проверка статуса
kubectl get pods -n cloudbridge-monitoring

# Ручная очистка (если нужно)
./scripts/prometheus-cleanup.sh
```

### 2. Доступ к дашбордам
```bash
# Автоматический запуск всех UI интерфейсов
./scripts/monitoring-ui.sh start

# Открытие всех интерфейсов в браузере
./scripts/monitoring-ui.sh open

# Или ручной port forward
kubectl port-forward service/grafana 3000:3000 -n cloudbridge-monitoring
kubectl port-forward service/prometheus-optimized-service 9090:9090 -n cloudbridge-monitoring
kubectl port-forward service/cadence-service 7934:7934 -n cloudbridge
```

### 3. Анализ производительности
```bash
# Запуск анализа производительности
./performance-analysis.sh

# Security scanning
./security/scanning/security-scan.sh
```

## 📈 Ключевые метрики

### Application Metrics
- Request rate
- Response time
- Error rate
- Throughput

### System Metrics
- CPU usage
- Memory usage
- Disk usage
- Network I/O

### Security Metrics
- Failed login attempts
- Security policy violations
- Network anomalies
- Vulnerability scans



