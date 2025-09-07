# Поэтапный план запуска и устранения ошибок CloudBridge Relay

## 📊 Текущее состояние системы (2025-09-07)

### ✅ Работающие сервисы:
- **CloudBridge Relay**: 3/3 пода работают стабильно
- **Cassandra**: 1/1 под работает (20Gi PV)
- **Prometheus**: 1/1 под работает (мониторинг)
- **Grafana**: 1/1 под работает (визуализация)

### ❌ Проблемные сервисы:
- **Cadence**: 2/2 пода в состоянии CrashLoopBackOff (76-77 перезапусков)

---

## 🎯 Поэтапный план устранения ошибок

### Этап 1: Диагностика проблемы с Cadence
**Время выполнения**: 10-15 минут

#### 1.1 Анализ логов Cadence
```bash
# Посмотреть логи последнего пода
kubectl logs cadence-849fb9c4f4-f6958 -n cloudbridge --tail=50

# Посмотреть описание пода для понимания причин
kubectl describe pod cadence-849fb9c4f4-f6958 -n cloudbridge
```

#### 1.2 Проверка подключения к Cassandra
```bash
# Проверить доступность Cassandra
kubectl exec -it cassandra-6dc4499998-5qgl9 -n cloudbridge -- cqlsh -e "DESCRIBE KEYSPACES;"

# Проверить сетевую связность
kubectl exec -it cadence-849fb9c4f4-f6958 -n cloudbridge -- nslookup cassandra-service.cloudbridge.svc.cluster.local
```

#### 1.3 Проверка конфигурации
```bash
# Проверить ConfigMap
kubectl get configmap cadence-config -n cloudbridge -o yaml

# Проверить переменные окружения
kubectl exec -it cadence-849fb9c4f4-f6958 -n cloudbridge -- env | grep CASSANDRA
```

---

### Этап 2: Исправление конфигурации Cadence
**Время выполнения**: 20-30 минут

#### 2.1 Создание правильного ConfigMap
```bash
# Создать новый ConfigMap с правильной конфигурацией
kubectl apply -f k8s/cadence/configmap.yaml
```

#### 2.2 Обновление deployment
```bash
# Применить обновленную конфигурацию
kubectl apply -f k8s/deployments/cadence-deployment.yaml

# Перезапустить deployment
kubectl rollout restart deployment/cadence -n cloudbridge
```

#### 2.3 Альтернативное решение - использование переменных окружения
Если ConfigMap не работает, использовать только переменные окружения:
```bash
# Удалить монтирование ConfigMap
# Добавить все необходимые переменные окружения в deployment
```

---

### Этап 3: Тестирование подключения Cadence к Cassandra
**Время выполнения**: 10-15 минут

#### 3.1 Проверка статуса подов
```bash
# Мониторинг статуса подов
kubectl get pods -n cloudbridge -w

# Проверка логов после исправления
kubectl logs -f deployment/cadence -n cloudbridge
```

#### 3.2 Тестирование подключения
```bash
# Проверить, что Cadence может подключиться к Cassandra
kubectl exec -it <cadence-pod> -n cloudbridge -- cqlsh cassandra-service.cloudbridge.svc.cluster.local -e "SELECT * FROM system_schema.keyspaces;"
```

---

### Этап 4: Проверка всех сервисов
**Время выполнения**: 15-20 минут

#### 4.1 Проверка основных сервисов
```bash
# Проверить статус всех подов
kubectl get pods -n cloudbridge
kubectl get pods -n cloudbridge-monitoring

# Проверить сервисы
kubectl get services -n cloudbridge
kubectl get services -n cloudbridge-monitoring
```

#### 4.2 Проверка Persistent Volumes
```bash
# Проверить PV и PVC
kubectl get pv
kubectl get pvc -n cloudbridge
```

---

### Этап 5: Тестирование доступности через port-forward
**Время выполнения**: 10-15 минут

#### 5.1 CloudBridge Relay
```bash
# Запустить port-forward для CloudBridge
kubectl port-forward -n cloudbridge service/cloudbridge-relay-service 8080:8080 &

# Тестировать API
curl http://localhost:8080/health
```

#### 5.2 Prometheus
```bash
# Запустить port-forward для Prometheus
kubectl port-forward -n cloudbridge-monitoring service/prometheus 9090:9090 &

# Проверить доступность
curl http://localhost:9090/api/v1/query?query=up
```

#### 5.3 Grafana
```bash
# Запустить port-forward для Grafana
kubectl port-forward -n cloudbridge-monitoring service/grafana 3000:3000 &

# Открыть в браузере: http://localhost:3000
# Логин: admin, Пароль: admin123
```

#### 5.4 Cadence (после исправления)
```bash
# Запустить port-forward для Cadence
kubectl port-forward -n cloudbridge service/cadence-service 7933:7933 &

# Проверить доступность
curl http://localhost:7933/health
```

---

### Этап 6: Настройка мониторинга и дашбордов
**Время выполнения**: 20-30 минут

#### 6.1 Проверка подключения Prometheus к Grafana
```bash
# Проверить, что Grafana видит Prometheus как источник данных
# В Grafana: Configuration -> Data Sources -> Prometheus
```

#### 6.2 Создание базовых дашбордов
- Дашборд для CloudBridge Relay
- Дашборд для Cassandra
- Дашборд для Cadence
- Дашборд для системных метрик

#### 6.3 Настройка алертов
- Алерты на недоступность сервисов
- Алерты на высокое использование ресурсов
- Алерты на ошибки в логах

---

## 🔧 Команды для быстрого исправления

### Если Cadence не может подключиться к Cassandra:
```bash
# 1. Проверить сетевую связность
kubectl exec -it cassandra-6dc4499998-5qgl9 -n cloudbridge -- cqlsh -e "SELECT * FROM system.local;"

# 2. Проверить keyspaces
kubectl exec -it cassandra-6dc4499998-5qgl9 -n cloudbridge -- cqlsh -e "DESCRIBE KEYSPACES;"

# 3. Пересоздать keyspaces если нужно
kubectl exec -it cassandra-6dc4499998-5qgl9 -n cloudbridge -- cqlsh -e "CREATE KEYSPACE IF NOT EXISTS cadence WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};"
kubectl exec -it cassandra-6dc4499998-5qgl9 -n cloudbridge -- cqlsh -e "CREATE KEYSPACE IF NOT EXISTS cadence_visibility WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};"
```

### Если проблемы с правами доступа:
```bash
# Проверить права на PV
kubectl describe pv
kubectl describe pvc cadence-pvc -n cloudbridge

# Пересоздать PV если нужно
kubectl delete pvc cadence-pvc -n cloudbridge
kubectl apply -f k8s/persistent-volumes/cloudbridge-pv.yaml
```

---

## 📋 Чек-лист для проверки

### После каждого этапа проверить:
- [ ] Все поды в состоянии Running
- [ ] Нет ошибок в логах
- [ ] Сервисы доступны через port-forward
- [ ] Метрики собираются в Prometheus
- [ ] Дашборды работают в Grafana

### Финальная проверка:
- [ ] CloudBridge Relay: 3/3 пода работают
- [ ] Cassandra: 1/1 под работает
- [ ] Cadence: 1/1 под работает
- [ ] Prometheus: 1/1 под работает
- [ ] Grafana: 1/1 под работает
- [ ] Все сервисы доступны через port-forward
- [ ] Мониторинг настроен и работает

---

## 🚨 Экстренные действия

### Если система полностью не работает:
```bash
# 1. Перезапустить все deployments
kubectl rollout restart deployment/cloudbridge-relay -n cloudbridge
kubectl rollout restart deployment/cassandra -n cloudbridge
kubectl rollout restart deployment/cadence -n cloudbridge
kubectl rollout restart deployment/prometheus -n cloudbridge-monitoring
kubectl rollout restart deployment/grafana -n cloudbridge-monitoring

# 2. Проверить ресурсы кластера
kubectl top nodes
kubectl top pods -n cloudbridge
kubectl top pods -n cloudbridge-monitoring

# 3. Очистить проблемные поды
kubectl delete pods --all -n cloudbridge
kubectl delete pods --all -n cloudbridge-monitoring
```

---

## 📞 Контакты для поддержки

- **Документация**: `DEVELOPER_TASKS.md`, `STATUS_REPORT.md`
- **Логи**: `kubectl logs <pod-name> -n <namespace>`
- **Мониторинг**: Grafana на порту 3000
- **Метрики**: Prometheus на порту 9090

---

*Создано: 2025-09-07*  
*Статус: Готов к выполнению*


