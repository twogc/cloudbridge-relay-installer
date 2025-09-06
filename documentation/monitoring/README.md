# Monitoring Documentation

## 📁 Файлы в этой папке

- **[MONITORING_GUIDE.md](./MONITORING_GUIDE.md)** - Руководство по мониторингу

## 📊 Система мониторинга

### Компоненты мониторинга
- **Prometheus** - Сбор метрик
- **Grafana** - Визуализация данных
- **Alertmanager** - Управление алертами
- **Blackbox Exporter** - Мониторинг endpoints

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

## 🚀 Развертывание мониторинга

### 1. Запуск мониторинга
```bash
# Развертывание мониторинга
./deploy-monitoring.sh

# Проверка статуса
kubectl get pods -n cloudbridge-monitoring
```

### 2. Доступ к дашбордам
```bash
# Port forward для Grafana
kubectl port-forward service/grafana 3000:3000 -n cloudbridge-monitoring

# Port forward для Prometheus
kubectl port-forward service/prometheus 9090:9090 -n cloudbridge-monitoring
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



