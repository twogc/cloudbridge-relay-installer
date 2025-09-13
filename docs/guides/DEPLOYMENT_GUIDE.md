# CloudBridge Relay - Руководство по развертыванию

## 📋 Обзор

CloudBridge Relay - это серверная часть системы CloudBridge, которая обеспечивает:
- **HTTP Health Checks** на порту 80
- **Main TCP API** на порту 8080
- **Metrics (Prometheus)** на порту 8081
- **WireGuard P2P Mesh** на порту 51820/UDP
- **QUIC Protocol** на порту 9090/UDP

## 🚀 Быстрый старт

### 1. Локальное развертывание

```bash
# Клонирование репозитория
git clone <repository-url>
cd cloudbridge-relay-installer

# Сборка и запуск
make build
make docker-build
make docker-run

# Тестирование
make test-ports
```

### 2. Kubernetes развертывание

```bash
# Обычное развертывание
make deploy-k8s

# Развертывание с Anycast
make deploy-k8s-anycast

# Тестирование
make test-k8s
```

## 🔧 Конфигурация

### Основные порты

| Порт | Протокол | Назначение | Статус |
|------|----------|------------|--------|
| 80 | TCP | HTTP Health Checks | ✅ Работает |
| 8080 | TCP | Main API | ✅ Работает |
| 8081 | TCP | Metrics (Prometheus) | ✅ Работает |
| 51820 | UDP | WireGuard P2P Mesh | ✅ Работает |
| 9090 | UDP | QUIC Protocol | ✅ Работает |

### IP адреса

- **Anycast IP**: `83.166.235.45`
- **LoadBalancer IP**: `109.120.191.126`

## 🐳 Docker развертывание

### 1. Сборка образа

```bash
make docker-build
```

### 2. Запуск контейнера

```bash
make docker-run
```

### 3. Проверка статуса

```bash
make status
```

### 4. Остановка

```bash
make docker-stop
```

## ☸️ Kubernetes развертывание

### 1. Обычное развертывание

```bash
# Применить манифесты
kubectl apply -f ../terraform-relay/k8s/manifests/cloudbridge-manifests.yaml

# Проверить статус
kubectl get pods -n cloudbridge-relay
kubectl get services -n cloudbridge-relay
```

### 2. Anycast развертывание

```bash
# Применить anycast манифесты
kubectl apply -f ../terraform-relay/k8s/manifests/cloudbridge-anycast-manifests.yaml

# Проверить статус
kubectl get pods -n cloudbridge-relay
```

### 3. Мониторинг

```bash
# Логи подов
kubectl logs -n cloudbridge-relay -l app=cloudbridge-relay -f

# Метрики
kubectl port-forward -n cloudbridge-relay svc/cloudbridge-relay 8081:8081
curl http://localhost:8081/metrics
```

## 🧪 Тестирование

### 1. Локальное тестирование

```bash
# Тест всех портов
make test-ports

# Тест удаленного сервера
make test-ports-remote

# Тест Anycast IP
make test-anycast
```

### 2. Ручное тестирование

```bash
# HTTP Health Check
curl http://localhost:80/health
curl http://localhost:80/ready

# TCP порты
nc -zv localhost 8080
nc -zv localhost 8081

# UDP порты
nc -zu localhost 51820
nc -zu localhost 9090
```

### 3. Тестирование с клиентом

```bash
# Запуск клиента
cd ../cloudbridge-client
./cloudbridge-client --config config.yaml
```

## 🔍 Диагностика

### 1. Проверка логов

```bash
# Docker
docker-compose logs cloudbridge-relay

# Kubernetes
kubectl logs -n cloudbridge-relay -l app=cloudbridge-relay
```

### 2. Проверка портов

```bash
# Локально
netstat -tulpn | grep cloudbridge

# В контейнере
docker exec cloudbridge-relay netstat -tulpn
```

### 3. Проверка конфигурации

```bash
# Проверить конфиг
cat configs/config.json

# Проверить переменные окружения
docker exec cloudbridge-relay env | grep CLOUDBRIDGE
```

## 🛠️ Устранение неполадок

### Проблема: Порт 51820/UDP не отвечает

**Решение:**
1. Проверить, что WireGuard включен в конфигурации
2. Проверить права доступа к /dev/net/tun
3. Проверить firewall правила

```bash
# Проверить WireGuard интерфейс
sudo wg show

# Проверить права
ls -la /dev/net/tun

# Проверить firewall
sudo iptables -L | grep 51820
```

### Проблема: Порт 9090/UDP не отвечает

**Решение:**
1. Проверить, что QUIC включен в конфигурации
2. Проверить UDP порты

```bash
# Проверить UDP порты
sudo netstat -ulpn | grep 9090

# Проверить конфигурацию QUIC
grep -A 5 "quic" configs/config.json
```

### Проблема: Health checks не проходят

**Решение:**
1. Проверить HTTP сервер на порту 80
2. Проверить endpoints /health и /ready

```bash
# Проверить HTTP сервер
curl -v http://localhost:80/health
curl -v http://localhost:80/ready

# Проверить логи
docker logs cloudbridge-relay | grep "HTTP server"
```

### Проблема: Anycast IP не работает

**Решение:**
1. Проверить LoadBalancer конфигурацию
2. Проверить health checks от Anycast IP
3. Проверить сетевые правила

```bash
# Проверить LoadBalancer
kubectl get svc -n cloudbridge-relay cloudbridge-relay-lb

# Проверить Anycast health check
curl http://83.166.235.45:80/health
```

## 📊 Мониторинг

### 1. Prometheus метрики

```bash
# Доступ к метрикам
curl http://localhost:8081/metrics

# Основные метрики
- relay_connections_total
- relay_auth_attempts_total
- relay_tunnels_created_total
- relay_active_connections
```

### 2. Grafana дашборд

```bash
# Доступ к Grafana
http://localhost:3000
# Логин: admin
# Пароль: admin
```

### 3. Логирование

```bash
# Структурированные логи в JSON формате
docker logs cloudbridge-relay | jq '.'

# Фильтрация по уровню
docker logs cloudbridge-relay | grep "ERROR"
```

## 🔒 Безопасность

### 1. Сетевая безопасность

- Все порты защищены firewall
- UDP порты открыты только для необходимых протоколов
- HTTP endpoints защищены rate limiting

### 2. Аутентификация

- JWT токены для API доступа
- Keycloak интеграция для OAuth 2.0
- WireGuard ключи для P2P соединений

### 3. Мониторинг безопасности

```bash
# Проверка безопасности
make security-check

# Аудит логов
docker logs cloudbridge-relay | grep -i "security\|auth\|failed"
```

## 📈 Масштабирование

### 1. Горизонтальное масштабирование

```bash
# Увеличить количество реплик
kubectl scale deployment cloudbridge-relay --replicas=5 -n cloudbridge-relay

# Автоматическое масштабирование
kubectl autoscale deployment cloudbridge-relay --min=2 --max=10 --cpu-percent=80 -n cloudbridge-relay
```

### 2. Вертикальное масштабирование

```yaml
resources:
  limits:
    cpu: "4000m"
    memory: "4Gi"
  requests:
    cpu: "1000m"
    memory: "1Gi"
```

## 🚀 Продакшн развертывание

### 1. Полная сборка

```bash
make prod
```

### 2. Развертывание с Anycast

```bash
make build-and-deploy-anycast
```

### 3. Проверка готовности

```bash
make test-anycast
make test-k8s
```

## 📞 Поддержка

### Полезные команды

```bash
# Справка
make help

# Статус всех сервисов
make status

# Полная очистка
make clean-all

# Справка по портам
make ports-help
```

### Логи и отладка

```bash
# Подробные логи
docker logs cloudbridge-relay -f

# Отладка конфигурации
docker exec cloudbridge-relay cat /app/configs/config.json

# Проверка процессов
docker exec cloudbridge-relay ps aux
```

---

## ✅ Чек-лист готовности

- [ ] Docker образ собран и протестирован
- [ ] Все порты открыты и отвечают
- [ ] Health checks проходят
- [ ] Anycast IP работает
- [ ] Kubernetes развертывание успешно
- [ ] Мониторинг настроен
- [ ] Логирование работает
- [ ] Безопасность проверена
- [ ] Масштабирование протестировано
- [ ] Документация обновлена
