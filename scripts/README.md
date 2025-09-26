# 🛠 CloudBridge Relay Scripts

Набор утилит для тестирования, мониторинга и обновления CloudBridge Relay.

## 📋 Доступные скрипты

### 🚀 `test-quick.sh` - Быстрое тестирование (5 минут)

Автоматизированный тест всего kubectl proxy функционала в minikube.

```bash
./scripts/test-quick.sh
```

**Что тестирует:**
- ✅ Базовые kubectl команды (get, list, describe)
- ✅ WebSocket streams (exec, logs, port-forward)
- ✅ Prometheus метрики
- ✅ Health endpoints
- ✅ Создание и cleanup тестовых ресурсов

**Требования:**
- Minikube запущен
- Docker доступен
- kubectl >= 1.29

### 🏥 `health-check.sh` - Проверка состояния системы

Комплексная диагностика работающей системы.

```bash
./scripts/health-check.sh
```

**Что проверяет:**
- ✅ Подключение к Kubernetes
- ✅ Статус подов и сервисов  
- ✅ HTTP endpoints (health, ready, metrics)
- ✅ kubectl proxy функциональность
- ✅ Использование ресурсов
- ✅ Ошибки в логах

**Конфигурация через переменные:**
```bash
export NAMESPACE="cloudbridge"
export DEPLOYMENT_NAME="cloudbridge-relay"
export RELAY_URL="http://localhost:8444"
export METRICS_URL="http://localhost:5551"  
export KUBECONFIG_PATH="~/.kube/config-relay"
./scripts/health-check.sh
```

### 🔄 `update-production.sh` - Обновление production системы

Безопасное rolling update с резервным копированием и проверками.

```bash
# Обновление до latest версии
./scripts/update-production.sh

# Обновление до конкретной версии
IMAGE_TAG=v1.2.0 ./scripts/update-production.sh

# Автоматическое подтверждение (для CI/CD)
AUTO_CONFIRM=true IMAGE_TAG=v1.2.0 ./scripts/update-production.sh
```

**Конфигурация:**
```bash
export NAMESPACE="cloudbridge"
export DEPLOYMENT_NAME="cloudbridge-relay"
export IMAGE_TAG="v1.2.0"
export IMAGE_REPO="ghcr.io/twogc/cloudbridge-relay"
export BACKUP_DIR="./backups"
export TIMEOUT="300s"
```

**Что делает:**
1. ✅ Создает резервные копии (deployment, configmap, service)
2. ✅ Проверяет состояние системы перед обновлением
3. ✅ Выполняет rolling update
4. ✅ Мониторит процесс обновления
5. ✅ Проверяет работоспособность после обновления
6. ✅ Автоматический откат при ошибках

## 🎯 Типовые сценарии использования

### Разработка и тестирование

```bash
# 1. Быстрая проверка изменений
./scripts/test-quick.sh

# 2. Полная проверка состояния  
./scripts/health-check.sh

# 3. Полные E2E тесты
export RELAY_URL="http://$(minikube ip):30444"
export METRICS_URL="http://$(minikube ip):30551"
./test/e2e/run_tests.sh
```

### Production мониторинг

```bash
# Ежедневный health check
./scripts/health-check.sh

# Проверка после инцидента
NAMESPACE=production DEPLOYMENT_NAME=cloudbridge-relay \
  ./scripts/health-check.sh

# Мониторинг метрик
curl -s http://your-relay:5551/metrics | grep kube_proxy_requests_total
```

### Production обновления

```bash
# Плановое обновление
IMAGE_TAG=v1.3.0 ./scripts/update-production.sh

# Экстренное обновление с автоподтверждением
AUTO_CONFIRM=true IMAGE_TAG=hotfix-v1.2.1 \
  NAMESPACE=production ./scripts/update-production.sh

# Откат к предыдущей версии
kubectl rollout undo deployment/cloudbridge-relay -n production
```

## 🔧 Расширение скриптов

### Добавление custom проверок

```bash
# В health-check.sh добавить:
check "Custom API endpoint" "curl -sf http://your-api/custom"
check "Database connectivity" "timeout 5s nc -z db-host 5432"
```

### Интеграция с мониторингом

```bash
# Отправка результатов в Slack/Discord
if [ $ISSUES -gt 0 ]; then
    curl -X POST -H 'Content-type: application/json' \
        --data '{"text":"CloudBridge Relay issues detected: '$ISSUES'"}' \
        $SLACK_WEBHOOK_URL
fi

# Отправка метрик в Prometheus Pushgateway  
echo "cloudbridge_health_issues $ISSUES" | \
    curl --data-binary @- http://pushgateway:9091/metrics/job/health-check
```

### CI/CD интеграция

```yaml
# GitHub Actions example
- name: Health Check
  run: |
    export KUBECONFIG=${{ secrets.KUBECONFIG }}
    export NAMESPACE=production
    ./scripts/health-check.sh

- name: Deploy Update
  if: github.ref == 'refs/heads/main'
  run: |
    export AUTO_CONFIRM=true
    export IMAGE_TAG=${{ github.sha }}
    ./scripts/update-production.sh
```

## 📊 Отчеты и логи

### Где найти логи скриптов

```bash
# Test logs
ls /tmp/test_*.log

# Health check в syslog
journalctl -t health-check

# Update logs
ls ./backups/
```

### Интерпретация результатов

**test-quick.sh:**
- ✅ Зеленые галочки = все работает
- ❌ Красные крестики = критические ошибки
- ⚠ Желтые треугольники = предупреждения

**health-check.sh:**
- Exit code 0 = система здорова
- Exit code 1 = обнаружены проблемы
- Подсчет проблем в переменной `$ISSUES`

**update-production.sh:**
- Exit code 0 = обновление успешно
- Exit code 1 = обновление провалено, откат выполнен
- Backups сохранены в `$BACKUP_DIR/`

---

## 🎉 Готовые шаблоны команд

### Разработчик

```bash
# Ежедневная проверка
./scripts/test-quick.sh && echo "✅ Dev environment OK"

# Перед коммитом
./scripts/health-check.sh && go test ./...
```

### DevOps инженер

```bash
# Мониторинг production
watch -n 30 './scripts/health-check.sh'

# Плановое обновление
./scripts/update-production.sh 2>&1 | tee update-$(date +%Y%m%d).log
```

### Site Reliability Engineer

```bash
# Проверка после алерта
NAMESPACE=production ./scripts/health-check.sh

# Диагностика проблем
kubectl logs deployment/cloudbridge-relay -n production --tail=100 | grep ERROR
curl -s http://production-relay:5551/metrics | grep -E "(up|kube_proxy_requests_total)"
```

**Все скрипты готовы к использованию!** 🚀
