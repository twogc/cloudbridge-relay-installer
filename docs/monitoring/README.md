# 📊 Мониторинг и аналитика

Эта папка содержит документацию по настройке и использованию систем мониторинга CloudBridge Relay.

## 📋 Содержание

### 🤖 Machine Learning мониторинг
- **ML_Monitoring_and_Grafana.md** - Интеграция ML с Grafana
- **ml_anomaly_detection.md** - Обнаружение аномалий

## 🎯 Компоненты мониторинга

### 📈 Prometheus Stack
- **Prometheus**: Сбор метрик
- **Grafana**: Визуализация и дашборды
- **AlertManager**: Управление алертами

### 🔍 Метрики микросервисов
- **Auth Service**: метрики аутентификации
- **Metrics Service**: бизнес-метрики
- **Системные метрики**: CPU, память, сеть

### 🚨 Алерты
- **Критические алерты**: недоступность сервисов
- **Предупреждения**: высокое потребление ресурсов
- **Бизнес-алерты**: аномалии в работе

## 🚀 Быстрый старт

1. Установить Prometheus Stack: `../installation/QUICK_INSTALL_PROMETHEUS.md`
2. Настроить ML мониторинг: `ML_Monitoring_and_Grafana.md`
3. Настроить обнаружение аномалий: `ml_anomaly_detection.md`

## 📊 Доступ к дашбордам

- **Grafana**: `http://<IP>:3000` (admin/CloudBridge2025!)
- **Prometheus**: `http://<IP>:9090`
- **AlertManager**: `http://<IP>:9093`
