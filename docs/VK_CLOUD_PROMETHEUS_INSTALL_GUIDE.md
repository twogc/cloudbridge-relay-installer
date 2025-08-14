# 🚀 Установка Prometheus Stack в VK Cloud для CloudBridge Relay

## 📋 Обзор

Это руководство описывает установку Prometheus Stack аддона в VK Cloud для мониторинга CloudBridge Relay микросервисов.

## 🎯 Что будет установлено

- **Prometheus** - система мониторинга и сбора метрик
- **Grafana** - веб-интерфейс для визуализации метрик
- **Alertmanager** - система алертов и уведомлений
- **ServiceMonitors** - автоматическое обнаружение CloudBridge сервисов
- **PrometheusRules** - правила алертов для CloudBridge

## 📋 Предварительные требования

1. **Kubernetes кластер** в VK Cloud
2. **kubectl** настроен для подключения к кластеру
3. **kauthproxy** установлен (для доступа к веб-интерфейсам)

## 🔧 Шаг 1: Подготовка кластера

### Проверка подключения к кластеру
```bash
kubectl cluster-info
kubectl get nodes
```

### Создание namespace для мониторинга
```bash
kubectl create namespace cloudbridge-monitoring
```

## 🚀 Шаг 2: Установка аддона в консоли VK Cloud

### 2.1 Откройте консоль VK Cloud
1. Войдите в [консоль VK Cloud](https://mcs.mail.ru/)
2. Перейдите в раздел **Kubernetes**
3. Выберите ваш кластер
4. Перейдите на вкладку **Аддоны**

### 2.2 Добавьте аддон kube-prometheus-stack
1. Нажмите **Добавить аддон**
2. Выберите **kube-prometheus-stack**
3. Заполните поля:
   - **Название приложения**: `cloudbridge-prometheus`
   - **Пространство имен**: `cloudbridge-monitoring`
   - **Пароль Grafana**: `CloudBridge2025!`

### 2.3 Вставьте конфигурацию
Скопируйте содержимое файла `configs/kube-prometheus-stack-minimal.yaml` в поле **Код настройки аддона**.

### 2.4 Установите аддон
Нажмите кнопку **Установить аддон** и дождитесь завершения установки.

## 🔍 Шаг 3: Проверка установки

### Проверка подов
```bash
kubectl get pods -n cloudbridge-monitoring
```

Ожидаемый вывод:
```
NAME                                                     READY   STATUS    RESTARTS   AGE
prometheus-grafana-7c4f8b8b8b-xxxxx                      2/2     Running   0          5m
prometheus-kube-state-metrics-7c4f8b8b8b-xxxxx          1/1     Running   0          5m
prometheus-operator-7c4f8b8b8b-xxxxx                    1/1     Running   0          5m
prometheus-prometheus-node-exporter-xxxxx               1/1     Running   0          5m
prometheus-prometheus-0                                  1/1     Running   0          5m
prometheus-alertmanager-0                                1/1     Running   0          5m
```

### Проверка сервисов
```bash
kubectl get services -n cloudbridge-monitoring
```

### Проверка ServiceMonitors
```bash
kubectl get servicemonitors -n cloudbridge-monitoring
```

## 🌐 Шаг 4: Доступ к веб-интерфейсам

### 4.1 Установка kauthproxy (если не установлен)

#### Для Linux:
```bash
# Скачайте kauthproxy
wget https://github.com/kauthproxy/kauthproxy/releases/latest/download/kauthproxy_linux_amd64.zip

# Распакуйте архив
unzip kauthproxy_linux_amd64.zip

# Переместите в PATH
sudo mv kauthproxy /usr/local/bin/
sudo chmod +x /usr/local/bin/kauthproxy
```

#### Для macOS:
```bash
# Скачайте kauthproxy
wget https://github.com/kauthproxy/kauthproxy/releases/latest/download/kauthproxy_darwin_amd64.zip

# Распакуйте архив
unzip kauthproxy_darwin_amd64.zip

# Переместите в PATH
sudo mv kauthproxy /usr/local/bin/
sudo chmod +x /usr/local/bin/kauthproxy
```

### 4.2 Доступ к Grafana

#### Способ 1: Через kauthproxy (рекомендуется)
```bash
kauthproxy -n cloudbridge-monitoring https://prometheus-grafana.svc
```

#### Способ 2: Через port-forward
```bash
kubectl -n cloudbridge-monitoring port-forward service/prometheus-grafana 8001:80
```

Затем откройте в браузере: http://127.0.0.1:8001

**Логин**: `admin`  
**Пароль**: `CloudBridge2025!`

### 4.3 Доступ к Prometheus

#### Способ 1: Через kauthproxy
```bash
kauthproxy -n cloudbridge-monitoring https://prometheus-kube-prometheus-prometheus.svc
```

#### Способ 2: Через port-forward
```bash
kubectl -n cloudbridge-monitoring port-forward service/prometheus-kube-prometheus-prometheus 9090:9090
```

Затем откройте в браузере: http://127.0.0.1:9090

### 4.4 Доступ к Alertmanager

#### Способ 1: Через kauthproxy
```bash
kauthproxy -n cloudbridge-monitoring https://prometheus-kube-prometheus-alertmanager.svc
```

#### Способ 2: Через port-forward
```bash
kubectl -n cloudbridge-monitoring port-forward service/prometheus-kube-prometheus-alertmanager 9093:9093
```

Затем откройте в браузере: http://127.0.0.1:9093

## 📊 Шаг 5: Настройка дашбордов Grafana

### 5.1 Импорт дашбордов

После входа в Grafana импортируйте следующие дашборды:

1. **Kubernetes Cluster Monitoring** (ID: 315)
2. **Kubernetes Pods** (ID: 6417)
3. **Kubernetes Services** (ID: 6418)

### 5.2 Создание дашборда для CloudBridge

Создайте новый дашборд для CloudBridge сервисов:

```json
{
  "dashboard": {
    "title": "CloudBridge Relay Dashboard",
    "panels": [
      {
        "title": "Auth Service - HTTP Requests",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total{app=\"auth-service\"}[5m])",
            "legendFormat": "{{method}} {{status}}"
          }
        ]
      },
      {
        "title": "QUIC Connections",
        "type": "stat",
        "targets": [
          {
            "expr": "enhanced_quic_connections_total",
            "legendFormat": "Active Connections"
          }
        ]
      },
      {
        "title": "WireGuard Peers",
        "type": "stat",
        "targets": [
          {
            "expr": "wireguard_peers_total",
            "legendFormat": "Active Peers"
          }
        ]
      }
    ]
  }
}
```

## 🔔 Шаг 6: Настройка алертов

### 6.1 Проверка правил алертов
```bash
kubectl get prometheusrules -n cloudbridge-monitoring
```

### 6.2 Настройка уведомлений в Alertmanager

1. Откройте Alertmanager
2. Перейдите в **Configuration**
3. Добавьте конфигурацию для уведомлений:

```yaml
global:
  smtp_smarthost: 'smtp.gmail.com:587'
  smtp_from: 'alertmanager@cloudbridge.com'
  smtp_auth_username: 'your-email@gmail.com'
  smtp_auth_password: 'your-app-password'

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'team-cloudbridge'

receivers:
- name: 'team-cloudbridge'
  email_configs:
  - to: 'team@cloudbridge.com'
    send_resolved: true
```

## 🧪 Шаг 7: Тестирование мониторинга

### 7.1 Проверка метрик CloudBridge сервисов

Убедитесь, что CloudBridge сервисы развернуты и экспортируют метрики:

```bash
# Проверка Auth Service
kubectl -n cloudbridge port-forward service/auth-service 8080:8080
curl http://localhost:8080/metrics

# Проверка Metrics Service
kubectl -n cloudbridge port-forward service/metrics-service 8084:8080
curl http://localhost:8084/metrics
```

### 7.2 Проверка ServiceMonitors

```bash
kubectl describe servicemonitor cloudbridge-auth-service -n cloudbridge-monitoring
kubectl describe servicemonitor cloudbridge-metrics-service -n cloudbridge-monitoring
```

### 7.3 Проверка алертов

```bash
# Проверка активных алертов
kubectl -n cloudbridge-monitoring port-forward service/prometheus-kube-prometheus-prometheus 9090:9090
```

Откройте http://127.0.0.1:9090 и перейдите в раздел **Alerts**.

## 🔧 Шаг 8: Оптимизация и настройка

### 8.1 Настройка retention и storage

Отредактируйте конфигурацию Prometheus:

```bash
kubectl edit prometheus prometheus -n cloudbridge-monitoring
```

Установите retention в 30 дней и увеличьте storage при необходимости.

### 8.2 Настройка ресурсов

При необходимости увеличьте ресурсы для компонентов:

```bash
kubectl edit deployment prometheus-grafana -n cloudbridge-monitoring
kubectl edit prometheus prometheus -n cloudbridge-monitoring
```

### 8.3 Настройка backup

Настройте регулярное резервное копирование данных Grafana:

```bash
# Создание backup скрипта
cat > backup-grafana.sh << 'EOF'
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
kubectl -n cloudbridge-monitoring exec deployment/prometheus-grafana -- tar czf /tmp/grafana-backup-$DATE.tar.gz /var/lib/grafana
kubectl -n cloudbridge-monitoring cp cloudbridge-monitoring/prometheus-grafana-0:/tmp/grafana-backup-$DATE.tar.gz ./grafana-backup-$DATE.tar.gz
EOF

chmod +x backup-grafana.sh
```

## 🚨 Устранение неполадок

### Проблема: Grafana не запускается
```bash
kubectl logs deployment/prometheus-grafana -n cloudbridge-monitoring
kubectl describe pod -l app.kubernetes.io/name=grafana -n cloudbridge-monitoring
```

### Проблема: Prometheus не собирает метрики
```bash
kubectl logs prometheus/prometheus -n cloudbridge-monitoring
kubectl get servicemonitors -n cloudbridge-monitoring
```

### Проблема: Алерты не работают
```bash
kubectl logs alertmanager/prometheus-kube-prometheus-alertmanager -n cloudbridge-monitoring
kubectl get prometheusrules -n cloudbridge-monitoring
```

### Проблема: Недостаточно ресурсов
```bash
kubectl top pods -n cloudbridge-monitoring
kubectl describe nodes
```

## 📈 Мониторинг производительности

### Ключевые метрики для отслеживания:

1. **CPU и Memory** - использование ресурсов компонентами
2. **Storage** - использование дискового пространства
3. **Network** - сетевой трафик
4. **Scrape duration** - время сбора метрик
5. **Alert firing rate** - частота срабатывания алертов

### Рекомендуемые дашборды:

- **Prometheus Overview** - общий обзор Prometheus
- **Grafana Overview** - общий обзор Grafana
- **Alertmanager Overview** - общий обзор Alertmanager
- **CloudBridge Services** - метрики CloudBridge сервисов

## 🎯 Заключение

После выполнения всех шагов у вас будет полностью настроенная система мониторинга для CloudBridge Relay с:

- ✅ Автоматическим сбором метрик
- ✅ Визуализацией в Grafana
- ✅ Системой алертов
- ✅ Мониторингом CloudBridge сервисов
- ✅ Готовыми дашбордами

**Следующий шаг**: Настройка внешнего доступа через Ingress и SSL сертификаты.

---

**Статус**: ✅ **ГОТОВ К ИСПОЛЬЗОВАНИЮ**  
**Время установки**: 30-60 минут  
**Сложность**: Средняя
