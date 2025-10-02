# CloudBridge Cloudflare Integration Testing

Этот набор скриптов предназначен для тестирования интеграции CloudBridge Relay с Cloudflare без нарушения работы основной системы.

## 🎯 Цель тестирования

Проверить работоспособность и производительность соединения между российским Relay-сервером и Cloudflare Magic WAN, используя:

1. **cloudflared** (QUIC/TLS 1.3, L7) — вариант «маскировки» трафика
2. **IPsec/GRE** → Magic WAN Anycast IPv4 (для будущих тестов)

## 📁 Структура файлов

```
scripts/cloudflare-testing/
├── deploy-testing.sh          # Развертывание тестового окружения
├── setup-cloudflare.sh        # Настройка Cloudflare туннеля
├── test-connectivity.sh       # Тестирование подключения
└── README.md                  # Этот файл

k8s/cloudflare-testing/
├── namespace.yaml             # Namespace для тестирования
├── cloudbridge-relay-testing-deployment.yaml
├── cloudbridge-testing-configmap.yaml
├── cloudflared-deployment.yaml
├── cloudflared-configmap.yaml
└── cloudbridge-testing-services.yaml
```

## 🚀 Быстрый старт

### 1. Развертывание тестового окружения

```bash
# Создает namespace и развертывает все компоненты
./scripts/cloudflare-testing/deploy-testing.sh
```

### 2. Настройка Cloudflare туннеля

```bash
# Настраивает Cloudflare туннель (требует аутентификации)
./scripts/cloudflare-testing/setup-cloudflare.sh
```

### 3. Тестирование подключения

```bash
# Запускает серию тестов подключения
./scripts/cloudflare-testing/test-connectivity.sh
```

## 🔧 Конфигурация

### Тестовые порты

Чтобы не конфликтовать с production системой, используются отдельные порты:

| Сервис | Production | Testing |
|--------|------------|---------|
| Main Server | 5550 | 5551 |
| Metrics | 5551 | 5552 |
| P2P API | 5552 | 5553 |
| HTTP API | 5553 | 5554 |
| HTTPS API | 5554 | 5555 |
| QUIC | 9090 | 9091 |
| STUN | 19302 | 19303 |

### Переменные окружения

```bash
# Cloudflare настройки
CLOUDFLARE_TUNNEL_ID="your-tunnel-id"
CLOUDFLARE_ACCOUNT_TAG="your-account-tag"
CLOUDFLARE_TUNNEL_SECRET="your-tunnel-secret"

# JWT секреты
RELAY_JWT_SECRET="85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE="
JWT_SECRET="85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE="
```

## 🧪 Тесты

### Базовые тесты подключения

1. **Health Check** - проверка состояния сервисов
2. **Metrics** - доступность метрик
3. **API** - тестирование API endpoints
4. **DNS** - разрешение DNS запросов
5. **External** - внешняя связность
6. **QUIC** - тестирование QUIC протокола

### Нагрузочные тесты

```bash
# Тест производительности
kubectl exec -n cloudbridge-testing deployment/cloudbridge-relay-testing -- \
  iperf3 -c <cloudflare-endpoint> -u -b 200M -t 60

# Тест с большими пакетами (проверка 16KB cap)
kubectl exec -n cloudbridge-testing deployment/cloudbridge-relay-testing -- \
  curl --http3 --limit-rate 10M https://speedtest.relay-host/largefile
```

### Мониторинг

```bash
# Просмотр логов
kubectl logs -n cloudbridge-testing deployment/cloudbridge-relay-testing -f
kubectl logs -n cloudbridge-testing deployment/cloudflared-testing -f

# Метрики
kubectl port-forward -n cloudbridge-testing svc/cloudbridge-relay-testing 5552:5552
curl http://localhost:5552/metrics

# Статус подов
kubectl get pods -n cloudbridge-testing
kubectl top pods -n cloudbridge-testing
```

## 🔍 Диагностика

### Проверка статуса

```bash
# Статус всех компонентов
kubectl get all -n cloudbridge-testing

# Проверка сервисов
kubectl get svc -n cloudbridge-testing

# Проверка ConfigMaps
kubectl get configmap -n cloudbridge-testing

# Проверка секретов
kubectl get secret -n cloudbridge-testing
```

### Логи и отладка

```bash
# Логи CloudBridge relay
kubectl logs -n cloudbridge-testing deployment/cloudbridge-relay-testing

# Логи cloudflared
kubectl logs -n cloudbridge-testing deployment/cloudflared-testing

# Описание подов для диагностики
kubectl describe pod -n cloudbridge-testing -l app=cloudbridge-relay-testing
kubectl describe pod -n cloudbridge-testing -l app=cloudflared-testing
```

## 🧹 Очистка

### Удаление тестового окружения

```bash
# Удаление namespace (удалит все ресурсы)
kubectl delete namespace cloudbridge-testing

# Или удаление отдельных компонентов
kubectl delete -f k8s/cloudflare-testing/
```

## 📊 Мониторинг и метрики

### Prometheus метрики

Тестовое окружение интегрируется с существующим Prometheus:

```yaml
# prometheus-config.yaml
- job_name: 'cloudbridge-testing'
  static_configs:
  - targets: ['cloudbridge-relay-testing.cloudbridge-testing:5552']
  scrape_interval: 30s
  metrics_path: /metrics
```

### Grafana дашборды

Создайте отдельный дашборд для тестирования:

```json
{
  "dashboard": {
    "title": "CloudBridge Testing - Cloudflare Integration",
    "panels": [
      {
        "title": "Cloudflare Tunnel Status",
        "targets": [
          {
            "expr": "cloudflared_tunnel_up{instance=\"cloudbridge-testing\"}"
          }
        ]
      }
    ]
  }
}
```

## 🚨 Troubleshooting

### Частые проблемы

1. **Туннель не подключается**
   - Проверьте credentials в секрете
   - Убедитесь, что tunnel ID правильный
   - Проверьте логи cloudflared

2. **Нет внешней связности**
   - Проверьте DNS настройки
   - Убедитесь, что firewall не блокирует трафик
   - Проверьте Cloudflare dashboard

3. **Высокое использование CPU**
   - Проверьте настройки QUIC
   - Убедитесь, что включены аппаратные ускорения
   - Проверьте нагрузку на сеть

### Отладка

```bash
# Проверка сетевых соединений
kubectl exec -n cloudbridge-testing deployment/cloudbridge-relay-testing -- \
  netstat -tulpn

# Проверка процессов
kubectl exec -n cloudbridge-testing deployment/cloudbridge-relay-testing -- \
  ps aux

# Проверка сетевых интерфейсов
kubectl exec -n cloudbridge-testing deployment/cloudbridge-relay-testing -- \
  ip addr show
```

## 📈 Критерии успеха

1. **Connectivity**: Relay успешно соединяется с Cloudflare
2. **Throughput**: Для нагрузки 200 Mbps CPU < 80% и packet loss < 1%
3. **Latency**: RTT к Cloudflare < 60 ms
4. **No 16KB cap**: Нет резкого падения при больших пакетах
5. **Fallback**: Корректное переключение при блокировке UDP
6. **Monitoring**: Все метрики поступают в Prometheus

## 🔄 Интеграция с production

После успешного тестирования:

1. Обновите production конфигурацию
2. Примените изменения через blue-green deployment
3. Мониторьте метрики в production
4. Подготовьте план отката

## 📞 Поддержка

При возникновении проблем:

1. Проверьте логи всех компонентов
2. Убедитесь в правильности конфигурации
3. Проверьте сетевую связность
4. Обратитесь к документации Cloudflare
